package models

import (
	"MSTransaccionesFinancieras/internal/utils"
	"strconv"
	"time"

	"github.com/tigerbeetle/tigerbeetle-go/pkg/types"
)

// resultado final de una transferencia procesada.
type TransferenciaNotificada struct {
	IdTransferencia string `json:"IdTransferencia"`
	IdUsuarioFinal  uint64 `json:"IdUsuarioFinal"`
	Monto           string `json:"Monto"`
	IdMoneda        uint32 `json:"IdMoneda"`
	Tipo            string `json:"Tipo"`
	Categoria       uint64 `json:"Categoria"`
	Estado          string `json:"Estado"`
	Mensaje         string `json:"Mensaje"`
	Fecha           string `json:"Fecha"`
	FechaProceso    string `json:"FechaProceso"`
}

// struct que se envía a traves del Webhook
type LoteNotificado struct {
	CantidadProcesada int                       `json:"CantidadProcesada"`
	Transferencias    []TransferenciaNotificada `json:"Transferencias"`
}

// Crear una notif a partir de una Transferencia, su mensaje Kafka original y su resultado de TigerBeetle.
// Si TB devuelve TransferExists (code 46), la transfer se reporta como exitosa con mensaje
// "TransferOKAlreadyProcessed": ya fue procesada en un intento anterior (idempotencia ante reintentos).
func NewTransferenciaNotificada(transfer types.Transfer, kafkaMsg KafkaTransferencias, result types.TransferEventResult) TransferenciaNotificada {
	var estado, mensaje string
	switch result.Result {
	case types.TransferOK:
		estado = "F"
		mensaje = "TransferOK"
	case types.TransferExists:
		// Idempotencia: misma transfer enviada en un reintento. Ya fue registrada en TB exitosamente.
		estado = "F"
		mensaje = "TransferOKAlreadyProcessed"
	default:
		estado = "E"
		mensaje = result.Result.String()
	}

	// TODO: Fixear Fecha de la operación: decodificada desde UserData32 (segundos epoch almacenados al parsear).
	// Si UserData32 es 0 (fecha no pudo parsearse al construir la transfer), se usa el valor raw de Kafka.
	fecha := "-"
	if transfer.UserData32 > 0 {
		if f, err := utils.UserData32AFecha(transfer.UserData32); err == nil {
			fecha = f
		}
	} else if kafkaMsg.Fecha != "" {
		fecha = kafkaMsg.Fecha
	}

	// FechaProceso: timestamp de cuando TB procesó la transfer.
	fechaProceso := "-"
	if estado == "F" {
		fechaProceso = time.Now().UTC().Format("2006-01-02 15:04:05")
	}

	return TransferenciaNotificada{
		IdTransferencia: utils.Uint128AStringDecimal(transfer.ID),
		IdUsuarioFinal:  kafkaMsg.IdUsuarioFinal,
		Monto:           utils.Uint128AStringDecimal(transfer.Amount),
		IdMoneda:        transfer.Ledger,
		Tipo:            kafkaMsg.Tipo,
		Categoria:       transfer.UserData64,
		Estado:          estado,
		Mensaje:         mensaje,
		Fecha:           fecha,
		FechaProceso:    fechaProceso,
	}
}

// Crear una notif para una transferencia rechazada por validación previa (no fue a TigerBeetle).
func NewTransferenciaNotificadaError(transfer types.Transfer, kafkaMsg KafkaTransferencias, mensajeError string) TransferenciaNotificada {
	fecha := kafkaMsg.Fecha
	if fecha == "" {
		fecha = "-"
	}
	return TransferenciaNotificada{
		IdTransferencia: utils.Uint128AStringDecimal(transfer.ID),
		IdUsuarioFinal:  kafkaMsg.IdUsuarioFinal,
		Monto:           utils.Uint128AStringDecimal(transfer.Amount),
		IdMoneda:        transfer.Ledger,
		Tipo:            kafkaMsg.Tipo,
		Categoria:       kafkaMsg.IdCategoria,
		Estado:          "E",
		Mensaje:         mensajeError,
		Fecha:           fecha,
		FechaProceso:    "-",
	}
}

// Crear una notif para un mensaje de Kafka que falló en el parseo (no se pudo construir la Transfer de TB).
func NewTransferenciaNotificadaParseoError(kafkaMsg KafkaTransferencias, mensajeError string) TransferenciaNotificada {
	idTransferencia := kafkaMsg.IdTransferencia
	if idTransferencia == "" {
		idTransferencia = "0"
	}
	fecha := kafkaMsg.Fecha
	if fecha == "" {
		fecha = "-"
	}
	return TransferenciaNotificada{
		IdTransferencia: idTransferencia,
		IdUsuarioFinal:  kafkaMsg.IdUsuarioFinal,
		Monto:           strconv.FormatUint(kafkaMsg.Monto, 10),
		IdMoneda:        kafkaMsg.IdMoneda,
		Tipo:            kafkaMsg.Tipo,
		Categoria:       kafkaMsg.IdCategoria,
		Estado:          "E",
		Mensaje:         mensajeError,
		Fecha:           fecha,
		FechaProceso:    "-",
	}
}
