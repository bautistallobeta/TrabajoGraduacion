package models

import (
	"MSTransaccionesFinancieras/internal/utils"
	"time"

	"github.com/tigerbeetle/tigerbeetle-go/pkg/types"
)

// resultado final de una transferencia procesada.
type TransferenciaNotificada struct {
	IdTransferencia string    `json:"IdTransferencia"`
	IdCuentaDebito  string    `json:"IdCuentaDebito"`
	IdCuentaCredito string    `json:"IdCuentaCredito"`
	Monto           string    `json:"Monto"`
	Ledger          uint32    `json:"Ledger"`
	Categoria       uint64    `json:"Categoria"`
	Tipo            uint16    `json:"Tipo"`
	EstadoTB        string    `json:"EstadoTB"`
	FechaProceso    time.Time `json:"FechaProceso"`
}

// struct que se envía a traves del Webhook (TODO: eliminar este struct y que solamente se envíe un array de TransferenciaNotificada)
type LoteNotificado struct {
	CantidadProcesada int                       `json:"CantidadProcesada"`
	Transferencias    []TransferenciaNotificada `json:"Transferencias"`
}

// Crear una notif a partir de una Transferencia y su resultado
func NewTransferenciaNotificada(transfer types.Transfer, result types.TransferEventResult) TransferenciaNotificada {
	return TransferenciaNotificada{
		IdTransferencia: utils.Uint128AStringDecimal(transfer.ID),
		IdCuentaDebito:  utils.Uint128AStringDecimal(transfer.DebitAccountID),
		IdCuentaCredito: utils.Uint128AStringDecimal(transfer.CreditAccountID),
		Monto:           utils.Uint128AStringDecimal(transfer.Amount),
		Ledger:          transfer.Ledger,
		Categoria:       uint64(transfer.Code),
		Tipo:            uint16(transfer.Flags),
		EstadoTB:        result.Result.String(),
		FechaProceso:    time.Now(),
	}
}
