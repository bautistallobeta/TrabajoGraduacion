package controllers

import (
	"MSTransaccionesFinancieras/internal/gestores"
	kafka "MSTransaccionesFinancieras/internal/infra/kafkamstf"
	"MSTransaccionesFinancieras/internal/models"
	"MSTransaccionesFinancieras/internal/utils"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/tigerbeetle/tigerbeetle-go/pkg/types"
)

type TransferenciasControlador struct {
	Gestor    *gestores.GestorTransferencias
	Productor *kafka.ProductorKafka
}

func NewTransferenciasControlador(gt *gestores.GestorTransferencias, pr *kafka.ProductorKafka) *TransferenciasControlador {
	return &TransferenciasControlador{Gestor: gt, Productor: pr}
}
func (tc *TransferenciasControlador) Dame(c echo.Context) error {
	type Request struct {
		IdTransferencia string `param:"IdTransferencia"`
	}
	req := &Request{}

	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("Parametros incorrectos: "+utils.SanitizarError(err)))
	}

	if req.IdTransferencia == "" {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("IdTransferencia no puede ser vacío"))
	}

	transferencia := &models.Transferencias{IdTransferencia: req.IdTransferencia}
	if err := transferencia.Dame(); err != nil {
		return c.JSON(http.StatusNotFound, models.NewErrorRespuesta("Transferencia no encontrada"))
	}
	return c.JSON(http.StatusOK, transferencia)
}

// Este método responde al POST /transferencias que se creó UNICAMENTE para probar el ms
func (tc *TransferenciasControlador) Crear(c echo.Context) error {
	req := new(models.KafkaTransferencias)

	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("JSON inválido o tipo de datos incorrecto: "+utils.SanitizarError(err)))
	}

	if req.IdTransferencia == "" {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("IdTransferencia es obligatorio"))
	}
	if req.IdUsuarioFinal == 0 {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("IdUsuarioFinal es obligatorio"))
	}
	if req.IdMoneda <= 0 {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("IdMoneda debe ser mayor a cero"))
	}
	if req.Tipo != "I" && req.Tipo != "E" && req.Tipo != "R" {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("Tipo debe ser 'I' (ingreso), 'E' (egreso) o 'R' (reversión)"))
	}
	if req.Tipo != "R" && req.Monto <= 0 {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("Monto debe ser mayor a cero"))
	}

	//Publish de la transferencia en Kafka
	err := tc.Productor.PublicarTransferencia(c.Request().Context(), *req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.NewErrorRespuesta("Error al publicar en Kafka: "+utils.SanitizarError(err)))
	}

	// 202 Accepted (está encolada en kafka pendiente de ser procesada)
	return c.JSON(http.StatusAccepted, map[string]interface{}{
		"Mensaje": "Transferencia aceptada y encolada en Kafka.",
		"Id":      req.IdTransferencia,
	})
}

func (tc *TransferenciasControlador) Buscar(c echo.Context) error {
	// IdsTransferencia: array de IDs (repeated query param). Si se recibe, camino directo LookupTransfers.
	idsStr := c.QueryParams()["IdsTransferencia"]
	var idsTransferencia []types.Uint128
	if len(idsStr) > 0 {
		idsTransferencia = make([]types.Uint128, 0, len(idsStr))
		for _, s := range idsStr {
			id, err := utils.ParsearUint128(s)
			if err != nil {
				return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("IdsTransferencia contiene un ID inválido: "+s))
			}
			idsTransferencia = append(idsTransferencia, id)
		}
	}

	var idUsuarioFinal uint64 = 0
	if s := c.QueryParam("IdUsuarioFinal"); s != "" {
		parsed, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("IdUsuarioFinal debe ser un número válido"))
		}
		idUsuarioFinal = parsed
	}

	var idCategoria uint64 = 0
	if s := c.QueryParam("IdCategoria"); s != "" {
		parsed, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("IdCategoria debe ser un número válido"))
		}
		idCategoria = parsed
	}

	var idMoneda uint32 = 0
	if s := c.QueryParam("IdMoneda"); s != "" {
		parsed, err := strconv.ParseUint(s, 10, 32)
		if err != nil {
			return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("IdMoneda debe ser un número válido"))
		}
		idMoneda = uint32(parsed)
	}

	estado := c.QueryParam("Estado")
	if estado != "" && estado != "F" && estado != "R" {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("Estado debe ser 'F' (finalizada), 'R' (revertida), o vacío"))
	}

	var montoMin uint64 = 0
	if s := c.QueryParam("MontoMin"); s != "" {
		parsed, err := strconv.ParseFloat(s, 64)
		if err != nil || parsed < 0 {
			return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("MontoMin debe ser un número válido"))
		}
		montoMin = utils.MontoDecimalAUnidadMinima(parsed)
	}

	var montoMax uint64 = 0
	if s := c.QueryParam("MontoMax"); s != "" {
		parsed, err := strconv.ParseFloat(s, 64)
		if err != nil || parsed < 0 {
			return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("MontoMax debe ser un número válido"))
		}
		montoMax = utils.MontoDecimalAUnidadMinima(parsed)
	}

	if montoMin != 0 && montoMax != 0 && montoMin > montoMax {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("MontoMin no puede ser mayor a MontoMax"))
	}

	var timestampMin uint64 = 0
	if s := c.QueryParam("FechaDesde"); s != "" {
		ts, err := utils.FechaATimestampNS(s)
		if err != nil {
			return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("FechaDesde inválida: "+utils.SanitizarError(err)))
		}
		timestampMin = ts
	}

	var timestampMax uint64 = 0
	if s := c.QueryParam("FechaHasta"); s != "" {
		ts, err := utils.FechaATimestampNS(s)
		if err != nil {
			return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("FechaHasta inválida: "+utils.SanitizarError(err)))
		}
		timestampMax = ts
	}

	var limite uint32 = 100
	if s := c.QueryParam("Limite"); s != "" {
		parsed, err := strconv.ParseUint(s, 10, 32)
		if err != nil {
			return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("Limite debe ser un número válido"))
		}
		if parsed > 500 {
			return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("Limite no puede ser mayor a 500"))
		}
		if parsed == 0 {
			return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("Limite debe ser mayor a 0"))
		}
		limite = uint32(parsed)
	}

	transfers, err := tc.Gestor.BuscarAvanzado(idsTransferencia, idUsuarioFinal, idCategoria, idMoneda, estado, montoMin, montoMax, timestampMin, timestampMax, limite)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.NewErrorRespuesta("Error al buscar transferencias: "+utils.SanitizarError(err)))
	}

	respuesta := make([]models.Transferencias, 0, len(transfers))
	for _, t := range transfers {
		var tr models.Transferencias
		tr.PoblarDesdeTB(t)
		respuesta = append(respuesta, tr)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"Total":          len(respuesta),
		"Transferencias": respuesta,
	})
}
