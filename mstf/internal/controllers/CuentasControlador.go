package controllers

import (
	"MSTransaccionesFinancieras/internal/gestores"
	"MSTransaccionesFinancieras/internal/models"
	"MSTransaccionesFinancieras/internal/utils"
	"log"
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

func (cc *CuentasControlador) Dame(c echo.Context) error {
	type Request struct {
		IdUsuarioFinal uint64 `param:"IdUsuarioFinal"`
		IdMoneda       uint32 `param:"IdMoneda"`
	}

	req := &Request{}

	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("Parámetros inválidos: "+err.Error()))
	}

	if req.IdUsuarioFinal <= 0 || req.IdMoneda <= 0 {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("IdUsuarioFinal e IdMoneda son requeridos y deben ser mayores a cero"))
	}

	cuenta := &models.Cuentas{IdUsuarioFinal: req.IdUsuarioFinal, IdMoneda: req.IdMoneda}
	err := cuenta.Dame()
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("Error al obtener cuenta: "+err.Error()))
	}
	return c.JSON(http.StatusOK, cuenta)
}

func (cc *CuentasControlador) DameHistorial(c echo.Context) error {
	type Request struct {
		IdUsuarioFinal uint64 `param:"IdUsuarioFinal"`
		IdMoneda       uint32 `param:"IdMoneda"`
	}

	type BalanceHistorial struct {
		Debitos   string `json:"Debitos"`
		Creditos  string `json:"Creditos"`
		Balance   string `json:"Balance"`
		Timestamp uint64 `json:"Timestamp"`
	}

	type Response struct {
		IdUsuarioFinal uint64             `json:"IdUsuarioFinal"`
		IdMoneda       uint32             `json:"IdMoneda"`
		Total          int                `json:"Total"`
		Historial      []BalanceHistorial `json:"Historial"`
	}

	req := &Request{}

	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("Parámetros inválidos: "+err.Error()))
	}

	if req.IdUsuarioFinal <= 0 || req.IdMoneda <= 0 {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("IdUsuarioFinal e IdMoneda son requeridos y deben ser mayores a cero"))
	}

	// Parsear query params opcionales (0 = sin filtro)
	timestampMinStr := c.QueryParam("TimeStampMin")
	timestampMaxStr := c.QueryParam("TimeStampMax")
	limiteStr := c.QueryParam("Limite")
	var timestampMin uint64 = 0
	var timestampMax uint64 = 0
	var limite uint32 = 0

	if timestampMinStr != "" {
		parsed, err := strconv.ParseUint(timestampMinStr, 10, 64)
		if err != nil {
			return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("Parámetro 'TimeStampMin' inválido"))
		}
		timestampMin = parsed
	}

	if timestampMaxStr != "" {
		parsed, err := strconv.ParseUint(timestampMaxStr, 10, 64)
		if err != nil {
			return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("Parámetro 'TimeStampMax' inválido"))
		}
		timestampMax = parsed
	}

	if limiteStr != "" {
		parsed, err := strconv.ParseUint(limiteStr, 10, 32)
		if err != nil {
			return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("Parámetro 'Limite' inválido"))
		}
		limite = uint32(parsed)
	}

	//obtener historial
	cuenta := &models.Cuentas{IdUsuarioFinal: req.IdUsuarioFinal, IdMoneda: req.IdMoneda}
	balances, err := cuenta.DameHistorialBalances(timestampMin, timestampMax, limite)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.NewErrorRespuesta("Error al obtener historial: "+err.Error()))
	}

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
		IdUsuarioFinal: req.IdUsuarioFinal,
		IdMoneda:       req.IdMoneda,
		Total:          len(historial),
		Historial:      historial,
	}

	return c.JSON(http.StatusOK, respuesta)
}

func (cc *CuentasControlador) Crear(c echo.Context) error {
	type crearCuentaRequest struct {
		IdUsuarioFinal uint64 `json:"IdUsuarioFinal"`
		IdMoneda       uint32 `json:"IdMoneda"`
		FechaAlta      string `json:"FechaAlta"`
	}

	req := &crearCuentaRequest{}

	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("JSON inválido o tipo de datos incorrecto: "+err.Error()))
	}

	if req.IdUsuarioFinal <= 0 || req.IdMoneda <= 0 {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("Faltan campos requeridos: IdUsuarioFinal, IdMoneda"))
	}
	if req.FechaAlta == "" {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("Falta campo requerido: FechaAlta"))
	}

	// cuentas creadas vía API siempre tienen DebitsMustNotExceedCredits = true
	idCuentaTBString, err := cc.Gestor.Crear(req.IdMoneda, req.IdUsuarioFinal, req.FechaAlta, true)
	log.Printf("\n\nCuentasControlador.Crear: Resultado de creación en GestorCuentas: mensaje='%s', error='%v'", idCuentaTBString, err)
	if err != nil {
		return c.JSON(http.StatusConflict, models.NewErrorRespuesta("Error al crear cuenta: "+err.Error()))
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"Mensaje": "Cuenta creada exitosamente",
	})
}

func (cc *CuentasControlador) Buscar(c echo.Context) error {
	idUsuarioFinalStr := c.QueryParam("IdUsuarioFinal")
	idMonedaStr := c.QueryParam("IdMoneda")
	estado := c.QueryParam("Estado")
	limitStr := c.QueryParam("Limite")

	var idUsuarioFinal uint64 = 0
	if idUsuarioFinalStr != "" {
		parsed, err := strconv.ParseUint(idUsuarioFinalStr, 10, 64)
		if err != nil {
			return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("IdUsuarioFinal debe ser un número válido"))
		}
		idUsuarioFinal = parsed
	}

	var idMoneda uint32 = 0
	if idMonedaStr != "" {
		parsed, err := strconv.ParseUint(idMonedaStr, 10, 32)
		if err != nil {
			return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("IdMoneda debe ser un número válido"))
		}
		idMoneda = uint32(parsed)
	}

	// solo  se acepta estado "A", "I", o vacío
	if estado != "" && estado != "A" && estado != "I" {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("Estado debe ser 'A' (activo), 'I' (inactivo), o vacío"))
	}

	// Parsear limit con valor hardcodeado a  100 (TODO: valor que se tiene que obtener de la db relac (?))
	var limit uint32 = 100
	if limitStr != "" {
		parsed, err := strconv.ParseUint(limitStr, 10, 32)
		if err != nil {
			return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("Limite debe ser un número válido"))
		}
		if parsed > 500 {
			return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("Limite no puede ser mayor a 500"))
		}
		if parsed <= 0 {
			return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("Limite debe ser mayor a 0"))
		}
		limit = uint32(parsed)
	}

	cuentas, err := cc.Gestor.Buscar(idUsuarioFinal, idMoneda, estado, limit)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.NewErrorRespuesta("Error al buscar cuentas: "+err.Error()))
	}

	// armar respuesta formateada
	respuesta := make([]models.Cuentas, 0, len(cuentas))
	for _, cuentaTB := range cuentas {
		cuenta := models.Cuentas{
			IdCuenta:       utils.Uint128AStringDecimal(cuentaTB.ID),
			IdUsuarioFinal: cuentaTB.UserData64,
			IdMoneda:       cuentaTB.Ledger,
			Creditos:       utils.Uint128AStringDecimal(cuentaTB.CreditsPosted),
			Debitos:        utils.Uint128AStringDecimal(cuentaTB.DebitsPosted),
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
		"Total":   len(respuesta),
		"Cuentas": respuesta,
	})
}
