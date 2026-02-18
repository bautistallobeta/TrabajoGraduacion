package controllers

import (
	"MSTransaccionesFinancieras/internal/gestores"
	kafka "MSTransaccionesFinancieras/internal/infra/kafkamstf"
	"MSTransaccionesFinancieras/internal/models"
	"net/http"

	"github.com/labstack/echo/v4"
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
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("Parametros incorrectos: "+err.Error()))
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
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("JSON inválido o tipo de datos incorrecto: "+err.Error()))
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
		return c.JSON(http.StatusInternalServerError, models.NewErrorRespuesta("Error al publicar en Kafka: "+err.Error()))
	}

	// 202 Accepted (está encolada en kafka pendiente de ser procesada)
	return c.JSON(http.StatusAccepted, map[string]interface{}{
		"Mensaje": "Transferencia aceptada y encolada en Kafka.",
		"Id":      req.IdTransferencia,
	})
}
