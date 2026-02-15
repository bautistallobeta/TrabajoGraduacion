package models

import (
	"errors"

	"MSTransaccionesFinancieras/internal/infra/persistence"
	"MSTransaccionesFinancieras/internal/utils"

	"github.com/tigerbeetle/tigerbeetle-go/pkg/types"
)

// "wrapper" de Transfer de TB
type Transferencias struct {
	IdTransferencia string
	IdCuentaDebito  string
	IdCuentaCredito string
	IdLedger        uint32
	Monto           string
	Categoria       uint64
	Tipo            uint16
	Fecha           string
	Estado          string
}

// TODO: agregar comentario de método de clase
func (t *Transferencias) Dame() error {
	idTransferenciaCast, err := utils.ParsearUint128(t.IdTransferencia)
	if err != nil {
		return errors.New("IdTransferencia inválido: " + err.Error())
	}

	if idTransferenciaCast == (types.Uint128{}) || idTransferenciaCast == types.ToUint128(0) {
		return errors.New("IdTransferencia no puede ser nulo ni cero")
	}

	if persistence.ClienteTB == nil {
		return errors.New("Conexión a TigerBeetle no inicializada")
	}

	transfers, err := persistence.ClienteTB.LookupTransfers([]types.Uint128{idTransferenciaCast})
	if err != nil {
		return err
	}

	if len(transfers) == 0 {
		return errors.New("Transferencia no encontrada en TigerBeetle")
	}

	transferenciaTB := transfers[0]

	fecha, _ := utils.UserData128AFecha(transferenciaTB.UserData128)

	// Asignacipon de campos tb a modelo
	t.IdCuentaDebito = utils.Uint128AStringDecimal(transferenciaTB.DebitAccountID)
	t.IdCuentaCredito = utils.Uint128AStringDecimal(transferenciaTB.CreditAccountID)
	t.IdLedger = transferenciaTB.Ledger
	t.Monto = utils.Uint128AStringDecimal(transferenciaTB.Amount)
	t.Categoria = transferenciaTB.UserData64
	t.Tipo = uint16(transferenciaTB.Flags)
	t.Fecha = fecha
	t.Estado = "F"

	return nil
}

// TODO
func (t *Transferencias) Revertir() (string, error) {
	return "Reversión no implementada", errors.New("No implementado")
}
