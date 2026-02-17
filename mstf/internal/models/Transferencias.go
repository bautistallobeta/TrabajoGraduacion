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
	IdUsuarioFinal  uint64
	IdMoneda        uint32
	Monto           string
	Tipo            string
	Categoria       uint64
	Fecha           string
	Estado          string
}

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

	t.IdMoneda = transferenciaTB.Ledger
	t.Monto = utils.Uint128AStringDecimal(transferenciaTB.Amount)
	t.Categoria = transferenciaTB.UserData64
	t.Fecha = fecha
	t.Estado = "F"

	// Derivar Tipo e IdUsuarioFinal comparando DebitAccountID/CreditAccountID con la cuenta empresa
	// TODO: eliminar hardcodeo de token
	moneda := &Monedas{IdMoneda: int(transferenciaTB.Ledger)}
	if _, err := moneda.Dame("cf904666e02a79cfd50b074ab3c360c0"); err == nil && moneda.IdCuentaEmpresa != "" {
		idCuentaEmpresa, errParse := utils.ParsearUint128(moneda.IdCuentaEmpresa)
		if errParse == nil {
			// Identificar cuál cuenta es del usuario (la que no es empresa)
			var idCuentaUsuario types.Uint128
			if transferenciaTB.DebitAccountID == idCuentaEmpresa {
				t.Tipo = "I" // La empresa debita → el usuario recibe → Ingreso
				idCuentaUsuario = transferenciaTB.CreditAccountID
			} else if transferenciaTB.CreditAccountID == idCuentaEmpresa {
				t.Tipo = "E" // La empresa recibe crédito → el usuario paga → Egreso
				idCuentaUsuario = transferenciaTB.DebitAccountID
			}

			// Obtener IdUsuarioFinal del UserData64 de la cuenta usuario en TB
			cuentas, errLookup := persistence.ClienteTB.LookupAccounts([]types.Uint128{idCuentaUsuario})
			if errLookup == nil && len(cuentas) > 0 {
				t.IdUsuarioFinal = cuentas[0].UserData64
			}
		}
	}

	return nil
}

// TODO
func (t *Transferencias) Revertir() (string, error) {
	return "Reversión no implementada", errors.New("No implementado")
}
