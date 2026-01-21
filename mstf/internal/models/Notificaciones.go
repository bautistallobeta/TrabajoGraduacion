package models

import (
	"MSTransaccionesFinancieras/internal/utils"
	"time"

	"github.com/tigerbeetle/tigerbeetle-go/pkg/types"
)

// resultado final de una transferencia procesada.
type TransferenciaNotificada struct {
	IdTransferencia string    `json:"id_transferencia"`
	IdCuentaDebito  string    `json:"id_cuenta_debito"`
	IdCuentaCredito string    `json:"id_cuenta_credito"`
	Monto           string    `json:"monto"`
	Ledger          uint32    `json:"ledger"`
	Categoria       uint64    `json:"categoria"`
	Tipo            uint16    `json:"tipo"`
	EstadoTB        string    `json:"estado_tb"`
	FechaProceso    time.Time `json:"fecha_proceso"`
}

// struct que se envía a traves del Webhook (TODO: eliminar este struct y que solamente se envíe un array de TransferenciaNotificada)
type LoteNotificado struct {
	CantidadProcesada int                       `json:"cantidad_procesada"`
	Transferencias    []TransferenciaNotificada `json:"transferencias"`
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
		FechaProceso:    time.Unix(int64(transfer.Timestamp), 0),
	}
}
