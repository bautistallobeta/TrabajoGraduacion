package controllers

import (
	"MSTransaccionesFinancieras/internal/gestores"
	"MSTransaccionesFinancieras/internal/models"
	"MSTransaccionesFinancieras/internal/utils"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/tigerbeetle/tigerbeetle-go/pkg/types"
)

type CuentasControlador struct {
	Gestor *gestores.GestorCuentas
}

func NewCuentasControlador(gc *gestores.GestorCuentas) *CuentasControlador {
	return &CuentasControlador{Gestor: gc}
}

func (cc *CuentasControlador) DameCuenta(c echo.Context) error {
	type Request struct {
		IdCuenta string `param:"id_cuenta"`
	}

	req := &Request{}

	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("Parámetro 'id_cuenta' inválido: "+err.Error()))
	}

	if req.IdCuenta == "" {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("IdCuenta no puede ser vacío"))
	}

	cuenta := &models.Cuenta{IdCuenta: req.IdCuenta}
	if err := cuenta.Dame(); err != nil {
		if err != nil {
			return c.JSON(http.StatusNotFound, models.NewErrorRespuesta(err.Error()))
		}

		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("Error al obtener cuenta: "+err.Error()))
	}
	return c.JSON(http.StatusOK, cuenta)
}

func (cc *CuentasControlador) DameHistorialCuenta(c echo.Context) error {
	type Request struct {
		IdCuenta string `param:"id_cuenta"`
	}

	type BalanceHistorial struct {
		Debitos   string `json:"debitos"`
		Creditos  string `json:"creditos"`
		Balance   string `json:"balance"`
		Timestamp uint64 `json:"timestamp"`
	}

	type Response struct {
		IdCuenta  string             `json:"id_cuenta"`
		Total     int                `json:"total"`
		Historial []BalanceHistorial `json:"historial"`
	}

	req := &Request{}

	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("Parámetro 'id_cuenta' inválido: "+err.Error()))
	}

	if req.IdCuenta == "" {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("IdCuenta no puede ser vacío"))
	}

	// Parsear query params opcionales (0 = sin filrto)
	timestampMinStr := c.QueryParam("timestamp_min")
	timestampMaxStr := c.QueryParam("timestamp_max")
	limiteStr := c.QueryParam("limite")
	var timestampMin uint64 = 0
	var timestampMax uint64 = 0
	var limite uint32 = 0

	if timestampMinStr != "" {
		parsed, err := strconv.ParseUint(timestampMinStr, 10, 64)
		if err != nil {
			return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("Parámetro 'timestamp_min' inválido"))
		}
		timestampMin = parsed
	}

	if timestampMaxStr != "" {
		parsed, err := strconv.ParseUint(timestampMaxStr, 10, 64)
		if err != nil {
			return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("Parámetro 'timestamp_max' inválido"))
		}
		timestampMax = parsed
	}

	if limiteStr != "" {
		parsed, err := strconv.ParseUint(limiteStr, 10, 32)
		if err != nil {
			return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("Parámetro 'limite' inválido"))
		}
		limite = uint32(parsed)
	}

	//obtener historial
	cuenta := &models.Cuenta{IdCuenta: req.IdCuenta}
	balances, err := cuenta.DameHistorialBalances(timestampMin, timestampMax, limite)
	if err != nil {
		if err != nil {
			return c.JSON(http.StatusNotFound, models.NewErrorRespuesta(err.Error()))
		}
		return c.JSON(http.StatusInternalServerError, models.NewErrorRespuesta("Error al obtener historial: "+err.Error()))
	}

	// convertir AccountBalance a BalanceHistorial
	historial := make([]BalanceHistorial, 0, len(balances))
	for _, balance := range balances {
		creditos := balance.CreditsPosted.BigInt()
		debitos := balance.DebitsPosted.BigInt()
		balanceNeto := creditos.Sub(&creditos, &debitos)

		historial = append(historial, BalanceHistorial{
			Debitos:   utils.Uint128AStringDecimal(balance.DebitsPosted),
			Creditos:  utils.Uint128AStringDecimal(balance.CreditsPosted),
			Balance:   balanceNeto.String(),
			Timestamp: balance.Timestamp,
		})
	}

	respuesta := Response{
		IdCuenta:  req.IdCuenta,
		Total:     len(historial),
		Historial: historial,
	}

	return c.JSON(http.StatusOK, respuesta)
}

func (cc *CuentasControlador) CrearCuenta(c echo.Context) error {
	type crearCuentaRequest struct {
		IdUsuarioFinal uint64 `json:"id_usuario_final"`
		IdLedger       uint32 `json:"id_ledger"`
		FechaAlta      string `json:"fecha_alta"`
	}

	req := &crearCuentaRequest{}

	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("JSON inválido o tipo de datos incorrecto: "+err.Error()))
	}

	if req.IdUsuarioFinal == 0 || req.IdLedger == 0 {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("Faltan campos requeridos: id_usuario_final, id_ledger"))
	}
	if req.FechaAlta == "" {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("Falta campo requerido: fecha_alta"))
	}

	// cuentas creadas vía API siempre tienen DebitsMustNotExceedCredits = true
	idCuentaTBString, err := cc.Gestor.Crear(req.IdLedger, req.IdUsuarioFinal, req.FechaAlta, true)

	if err != nil {
		return c.JSON(http.StatusConflict, models.NewErrorRespuesta("Error al crear cuenta: "+err.Error()))
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"status": "OK: Cuenta creada exitosamente",
		"id":     idCuentaTBString,
	})
}

func (cc *CuentasControlador) BuscarCuentas(c echo.Context) error {
	idUsuarioFinalStr := c.QueryParam("id_usuario_final")
	idLedgerStr := c.QueryParam("id_ledger")
	tipoStr := c.QueryParam("tipo")
	estado := c.QueryParam("estado")
	limitStr := c.QueryParam("limit")

	var idUsuarioFinal uint64 = 0
	if idUsuarioFinalStr != "" {
		parsed, err := strconv.ParseUint(idUsuarioFinalStr, 10, 64)
		if err != nil {
			return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("id_usuario_final debe ser un número válido"))
		}
		idUsuarioFinal = parsed
	}

	var idLedger uint32 = 0
	if idLedgerStr != "" {
		parsed, err := strconv.ParseUint(idLedgerStr, 10, 32)
		if err != nil {
			return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("id_ledger debe ser un número válido"))
		}
		idLedger = uint32(parsed)
	}

	var tipo uint16 = 0
	if tipoStr != "" {
		parsed, err := strconv.ParseUint(tipoStr, 10, 16)
		if err != nil {
			return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("tipo debe ser un número válido"))
		}
		tipo = uint16(parsed)
	}

	// solo  se acepta estado "A", "I", o vacío
	if estado != "" && estado != "A" && estado != "I" {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("estado debe ser 'A' (activo), 'I' (inactivo), o vacío"))
	}

	// Parsear limit con valor hardcodeado a  100 (TODO: valor que se tiene que obtener de la db relac (?))
	var limit uint32 = 100
	if limitStr != "" {
		parsed, err := strconv.ParseUint(limitStr, 10, 32)
		if err != nil {
			return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("limit debe ser un número válido"))
		}
		if parsed > 500 {
			return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("limit no puede ser mayor a 500"))
		}
		if parsed == 0 {
			return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("limit debe ser mayor a 0"))
		}
		limit = uint32(parsed)
	}

	cuentas, err := cc.Gestor.Buscar(idUsuarioFinal, idLedger, tipo, estado, limit)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.NewErrorRespuesta("Error al buscar cuentas: "+err.Error()))
	}

	// armar respuesta formateada
	respuesta := make([]models.Cuenta, 0, len(cuentas))
	for _, cuentaTB := range cuentas {
		cuenta := models.Cuenta{
			IdCuenta:       utils.Uint128AStringDecimal(cuentaTB.ID),
			IdUsuarioFinal: cuentaTB.UserData64,
			IdLedger:       cuentaTB.Ledger,
			Creditos:       utils.Uint128AStringDecimal(cuentaTB.CreditsPosted),
			Debitos:        utils.Uint128AStringDecimal(cuentaTB.DebitsPosted),
			Tipo:           uint64(cuentaTB.Code),
		}
		//Leer el estado
		closedFlags := types.AccountFlags{Closed: true}.ToUint16()
		if (cuentaTB.Flags & closedFlags) != 0 {
			cuenta.Estado = "I"
		} else {
			cuenta.Estado = "A"
		}
		// FechaAlta se lee de UserData32
		if cuentaTB.UserData32 != 0 {
			fechaAlta, err := utils.UserData32AFecha(cuentaTB.UserData32)
			if err == nil {
				cuenta.FechaAlta = fechaAlta
			}
		}
		// FechaRegistro se lee de Timestamp de TB
		if cuentaTB.Timestamp != 0 {
			cuenta.FechaRegistro = utils.TimestampAFecha(cuentaTB.Timestamp)
		}

		respuesta = append(respuesta, cuenta)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"total":   len(respuesta),
		"cuentas": respuesta,
	})
}
