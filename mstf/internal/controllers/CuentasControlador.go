package controllers

import (
	"MSTransaccionesFinancieras/internal/gestores"
	"MSTransaccionesFinancieras/internal/models"
	"MSTransaccionesFinancieras/internal/utils"
	"fmt"
	"log"
	"math/big"
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
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("Parámetros inválidos: "+utils.SanitizarError(err)))
	}

	if req.IdUsuarioFinal <= 0 || req.IdMoneda <= 0 {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("IdUsuarioFinal e IdMoneda son requeridos y deben ser mayores a cero"))
	}

	cuenta := &models.Cuentas{IdUsuarioFinal: req.IdUsuarioFinal, IdMoneda: req.IdMoneda}
	err := cuenta.Dame()
	if err != nil {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("Error al obtener cuenta: "+utils.SanitizarError(err)))
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
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("Parámetros inválidos: "+utils.SanitizarError(err)))
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
		return c.JSON(http.StatusInternalServerError, models.NewErrorRespuesta("Error al obtener historial: "+utils.SanitizarError(err)))
	}

	historial := make([]BalanceHistorial, 0, len(balances))
	cien := big.NewInt(100)
	for _, balance := range balances {
		creditosRaw := balance.CreditsPosted.BigInt()
		debitosRaw := balance.DebitsPosted.BigInt()
		balanceRaw := new(big.Int).Sub(&creditosRaw, &debitosRaw)

		signo := ""
		if balanceRaw.Sign() < 0 {
			signo = "-"
			balanceRaw.Neg(balanceRaw)
		}
		entero, resto := new(big.Int), new(big.Int)
		entero.DivMod(balanceRaw, cien, resto)

		historial = append(historial, BalanceHistorial{
			Debitos:   utils.Uint128ADecimalMoneda(balance.DebitsPosted),
			Creditos:  utils.Uint128ADecimalMoneda(balance.CreditsPosted),
			Balance:   fmt.Sprintf("%s%s.%02d", signo, entero.String(), resto.Int64()),
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

func (cc *CuentasControlador) DameTransferencias(c echo.Context) error {
	type Request struct {
		IdUsuarioFinal uint64 `param:"idusuariofinal"`
		IdMoneda       uint32 `param:"idmoneda"`
	}

	req := &Request{}

	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("Parámetros inválidos: "+utils.SanitizarError(err)))
	}

	if req.IdUsuarioFinal <= 0 || req.IdMoneda <= 0 {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("IdUsuarioFinal e IdMoneda son requeridos y deben ser mayores a cero"))
	}

	timestampMinStr := c.QueryParam("TimestampMin")
	timestampMaxStr := c.QueryParam("TimestampMax")
	limiteStr := c.QueryParam("Limite")
	var timestampMin uint64 = 0
	var timestampMax uint64 = 0
	var limite uint32 = 0

	if timestampMinStr != "" {
		parsed, err := strconv.ParseUint(timestampMinStr, 10, 64)
		if err != nil {
			return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("Parámetro 'TimestampMin' inválido"))
		}
		timestampMin = parsed
	}

	if timestampMaxStr != "" {
		parsed, err := strconv.ParseUint(timestampMaxStr, 10, 64)
		if err != nil {
			return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("Parámetro 'TimestampMax' inválido"))
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

	cuenta := &models.Cuentas{IdUsuarioFinal: req.IdUsuarioFinal, IdMoneda: req.IdMoneda}
	transferencias, err := cuenta.BuscarTransferenciasCuenta(timestampMin, timestampMax, limite)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.NewErrorRespuesta("Error al obtener transferencias: "+utils.SanitizarError(err)))
	}

	respuesta := make([]models.Transferencias, 0, len(transferencias))
	for _, tb := range transferencias {
		if tb.Code == models.CodigoTransferenciaCierre {
			continue
		}
		t := &models.Transferencias{}
		t.PoblarDesdeTB(tb)
		respuesta = append(respuesta, *t)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"Total":          len(respuesta),
		"Transferencias": respuesta,
	})
}

func (cc *CuentasControlador) Crear(c echo.Context) error {
	type crearCuentaRequest struct {
		IdUsuarioFinal uint64 `json:"IdUsuarioFinal"`
		IdMoneda       uint32 `json:"IdMoneda"`
		Fecha          string `json:"Fecha"`
	}

	req := &crearCuentaRequest{}

	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("JSON inválido o tipo de datos incorrecto: "+utils.SanitizarError(err)))
	}

	if req.IdUsuarioFinal <= 0 || req.IdMoneda <= 0 {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("Faltan campos requeridos: IdUsuarioFinal, IdMoneda"))
	}
	if req.Fecha == "" {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("Falta campo requerido: Fecha"))
	}

	// cuentas creadas vía APIREST: DebitsMustNotExceedCredits = true (IdUsuarioFinal > 0)
	idCuentaTBString, existe, err := cc.Gestor.Crear(models.Cuentas{IdMoneda: req.IdMoneda, IdUsuarioFinal: req.IdUsuarioFinal, Fecha: req.Fecha})
	log.Printf("\n\nCuentasControlador.Crear: Resultado de creación en GestorCuentas: mensaje='%s', existe=%v, error='%v'", idCuentaTBString, existe, err)
	if err != nil {
		return c.JSON(http.StatusConflict, models.NewErrorRespuesta("Error al crear cuenta: "+utils.SanitizarError(err)))
	}

	if existe {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"Mensaje": "Cuenta ya existente",
		})
	}
	return c.JSON(http.StatusCreated, map[string]interface{}{
		"Mensaje": "Cuenta creada exitosamente",
	})
}

func (cc *CuentasControlador) Desactivar(c echo.Context) error {
	type Request struct {
		IdUsuarioFinal uint64 `param:"idusuariofinal"`
		IdMoneda       uint32 `param:"idmoneda"`
	}

	req := &Request{}
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("Parámetros inválidos: "+utils.SanitizarError(err)))
	}
	if req.IdUsuarioFinal <= 0 || req.IdMoneda <= 0 {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("IdUsuarioFinal e IdMoneda son requeridos y deben ser mayores a cero"))
	}

	cuenta := models.Cuentas{IdMoneda: req.IdMoneda, IdUsuarioFinal: req.IdUsuarioFinal}
	cuenta.Dame()
	if err := cuenta.Desactivar(); err != nil {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("Error al desactivar cuenta: "+utils.SanitizarError(err)))
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"Mensaje": "Cuenta desactivada exitosamente",
	})
}

func (cc *CuentasControlador) Activar(c echo.Context) error {
	type Request struct {
		IdUsuarioFinal uint64 `param:"idusuariofinal"`
		IdMoneda       uint32 `param:"idmoneda"`
	}

	req := &Request{}
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("Parámetros inválidos: "+utils.SanitizarError(err)))
	}
	if req.IdUsuarioFinal <= 0 || req.IdMoneda <= 0 {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("IdUsuarioFinal e IdMoneda son requeridos y deben ser mayores a cero"))
	}
	cuenta := models.Cuentas{IdMoneda: req.IdMoneda, IdUsuarioFinal: req.IdUsuarioFinal}
	cuenta.Dame()
	if err := cuenta.Activar(); err != nil {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("Error al activar cuenta: "+utils.SanitizarError(err)))
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"Mensaje": "Cuenta activada exitosamente",
	})
}

func (cc *CuentasControlador) Buscar(c echo.Context) error {
	//  arrays "paralelos" para EL lookup directo
	idsUsuarioFinalStr := c.QueryParams()["IdsUsuarioFinal"]
	idsMonedaStr := c.QueryParams()["IdsMoneda"]

	var idsCuenta []types.Uint128
	if len(idsUsuarioFinalStr) > 0 || len(idsMonedaStr) > 0 {
		if len(idsUsuarioFinalStr) != len(idsMonedaStr) {
			return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("IdsUsuarioFinal e IdsMoneda deben tener la misma cantidad de elementos"))
		}
		idsCuenta = make([]types.Uint128, 0, len(idsUsuarioFinalStr))
		for i := range idsUsuarioFinalStr {
			usuarioFinal, err := strconv.ParseUint(idsUsuarioFinalStr[i], 10, 64)
			if err != nil {
				return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("IdsUsuarioFinal contiene un valor inválido: "+idsUsuarioFinalStr[i]))
			}
			moneda, err := strconv.ParseUint(idsMonedaStr[i], 10, 32)
			if err != nil {
				return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("IdsMoneda contiene un valor inválido: "+idsMonedaStr[i]))
			}
			idCuentaStr := utils.ConcatenarIDString(uint64(moneda), usuarioFinal)
			idCuenta, err := utils.ParsearUint128(idCuentaStr)
			if err != nil {
				return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("No se pudo construir el ID de cuenta para el par en índice "+strconv.Itoa(i)))
			}
			idsCuenta = append(idsCuenta, idCuenta)
		}
	}

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

	limiteMaximo := obtenerLimiteMaximoBuscarCuentas()
	var limit uint32 = obtenerLimiteBuscarCuentas()
	if limitStr != "" {
		parsed, err := strconv.ParseUint(limitStr, 10, 32)
		if err != nil {
			return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("Limite debe ser un número válido"))
		}
		if parsed > uint64(limiteMaximo) {
			return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta(fmt.Sprintf("Limite no puede ser mayor a %d", limiteMaximo)))
		}
		if parsed <= 0 {
			return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("Limite debe ser mayor a 0"))
		}
		limit = uint32(parsed)
	}

	cuentas, err := cc.Gestor.BuscarAvanzado(idsCuenta, idUsuarioFinal, idMoneda, estado, limit)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.NewErrorRespuesta("Error al buscar cuentas: "+utils.SanitizarError(err)))
	}

	// respuesta formateada
	respuesta := make([]models.Cuentas, 0, len(cuentas))
	for _, cuentaTB := range cuentas {
		cuenta := models.Cuentas{
			IdCuenta:       utils.Uint128AStringDecimal(cuentaTB.ID),
			IdUsuarioFinal: cuentaTB.UserData64,
			IdMoneda:       cuentaTB.Ledger,
			Creditos:       utils.Uint128ADecimalMoneda(cuentaTB.CreditsPosted),
			Debitos:        utils.Uint128ADecimalMoneda(cuentaTB.DebitsPosted),
		}
		//Obtencion del estado
		closedFlags := types.AccountFlags{Closed: true}.ToUint16()
		if (cuentaTB.Flags & closedFlags) != 0 {
			cuenta.Estado = "I"
		} else {
			cuenta.Estado = "A"
		}
		// Fecha se lee de UserData32
		if cuentaTB.UserData32 != 0 {
			fecha, err := utils.UserData32AFecha(cuentaTB.UserData32)
			if err == nil {
				cuenta.Fecha = fecha
			}
		}
		// FechaProceso se lee de Timestamp de TB
		if cuentaTB.Timestamp != 0 {
			cuenta.FechaProceso = utils.TimestampAFecha(cuentaTB.Timestamp)
		}

		respuesta = append(respuesta, cuenta)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"Total":   len(respuesta),
		"Cuentas": respuesta,
	})
}

func obtenerLimiteBuscarCuentas() uint32 {
	p := &models.Parametros{Parametro: "LIMITEBUSCARCUENTAS"}
	if _, err := p.Dame(); err != nil || p.Valor == "" {
		return 100
	}
	val, err := strconv.ParseUint(p.Valor, 10, 32)
	if err != nil {
		return 100
	}
	return uint32(val)
}

func obtenerLimiteMaximoBuscarCuentas() uint32 {
	p := &models.Parametros{Parametro: "LIMITEMAXIMOBUSCARCUENTAS"}
	if _, err := p.Dame(); err != nil || p.Valor == "" {
		return 500
	}
	val, err := strconv.ParseUint(p.Valor, 10, 32)
	if err != nil {
		return 500
	}
	return uint32(val)
}
