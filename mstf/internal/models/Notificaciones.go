package models

import (
	"MSTransaccionesFinancieras/internal/utils"
	"strconv"
	"time"

	"github.com/tigerbeetle/tigerbeetle-go/pkg/types"
)

// resultado final de una transferencia procesada.
type TransferenciaNotificada struct {
	IdTransferencia string    `json:"IdTransferencia"`
	IdUsuarioFinal  uint64    `json:"IdUsuarioFinal"`
	Monto           string    `json:"Monto"`
	IdMoneda        uint32    `json:"IdMoneda"`
	Tipo            string    `json:"Tipo"`
	Categoria       uint64    `json:"Categoria"`
	Estado          string    `json:"Estado"`
	Mensaje         string    `json:"Mensaje"`
	FechaProceso    time.Time `json:"FechaProceso"`
}

// struct que se envía a traves del Webhook
type LoteNotificado struct {
	CantidadProcesada int                       `json:"CantidadProcesada"`
	Transferencias    []TransferenciaNotificada `json:"Transferencias"`
}

// Crear una notif a partir de una Transferencia, su mensaje Kafka original y su resultado de TigerBeetle
func NewTransferenciaNotificada(transfer types.Transfer, kafkaMsg KafkaTransferencias, result types.TransferEventResult) TransferenciaNotificada {
	estado := "F"
	mensaje := result.Result.String()
	if result.Result != types.TransferOK {
		estado = "E"
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
		FechaProceso:    time.Now(),
	}
}

// Crear una notif para una transferencia rechazada por validación previa (no fue a TigerBeetle)
func NewTransferenciaNotificadaError(transfer types.Transfer, kafkaMsg KafkaTransferencias, mensajeError string) TransferenciaNotificada {
	return TransferenciaNotificada{
		IdTransferencia: utils.Uint128AStringDecimal(transfer.ID),
		IdUsuarioFinal:  kafkaMsg.IdUsuarioFinal,
		Monto:           utils.Uint128AStringDecimal(transfer.Amount),
		IdMoneda:        transfer.Ledger,
		Tipo:            kafkaMsg.Tipo,
		Categoria:       kafkaMsg.IdCategoria,
		Estado:          "E",
		Mensaje:         mensajeError,
		FechaProceso:    time.Now(),
	}
}

// Crear una notif para un mensaje de Kafka que falló en el parseo (no se pudo construir la Transfer de TB).
func NewTransferenciaNotificadaParseoError(kafkaMsg KafkaTransferencias, mensajeError string) TransferenciaNotificada {
	idTransferencia := kafkaMsg.IdTransferencia
	if idTransferencia == "" {
		idTransferencia = "0"
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
		FechaProceso:    time.Now(),
	}
}
