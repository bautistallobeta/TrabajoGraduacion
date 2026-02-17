package models

import (
	"MSTransaccionesFinancieras/internal/utils"
	"time"

	"github.com/tigerbeetle/tigerbeetle-go/pkg/types"
)

// resultado final de una transferencia procesada.
type TransferenciaNotificada struct {
	IdTransferencia string    `json:"IdTransferencia"`
	IdUsuarioFinal  string    `json:"IdUsuarioFinal"`
	Monto           string    `json:"Monto"`
	IdMoneda        uint32    `json:"IdMoneda"`
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
		Categoria:       uint64(transfer.Code),
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
		Categoria:       uint64(transfer.Code),
		Estado:          "E",
		Mensaje:         mensajeError,
		FechaProceso:    time.Now(),
	}
}
