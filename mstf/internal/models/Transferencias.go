package models

import (
	"errors"

	"MSTransaccionesFinancieras/internal/infra/persistence"
	"MSTransaccionesFinancieras/internal/utils"

	"github.com/tigerbeetle/tigerbeetle-go/pkg/types"
)

const CodigoTransferenciaNormal uint16 = 1
const CodigoTransferenciaReversion uint16 = 2

// "wrapper" de Transfer de TB
type Transferencias struct {
	IdTransferencia         string
	IdUsuarioFinal          uint64
	IdMoneda                uint32
	Monto                   string
	Tipo                    string
	Categoria               uint64
	Fecha                   string
	Estado                  string
	Code                    uint16
	IdTransferenciaOriginal string `json:",omitempty"`
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

	// TODO: corregir esto - quedó de cuando guardaba el itmestamp en userdata128 -  "transferencias nuevas usan UserData32 (segundos), viejas usan UserData128 (nanosegundos)"
	var fecha string
	if transferenciaTB.UserData32 > 0 {
		fecha, _ = utils.UserData32AFecha(transferenciaTB.UserData32)
	} else {
		fecha, _ = utils.UserData128AFecha(transferenciaTB.UserData128)
	}

	t.IdMoneda = transferenciaTB.Ledger
	t.Monto = utils.Uint128AStringDecimal(transferenciaTB.Amount)
	t.Categoria = transferenciaTB.UserData64
	t.Fecha = fecha
	t.Estado = "F"
	t.Code = transferenciaTB.Code

	if t.Code == CodigoTransferenciaReversion {
		t.IdTransferenciaOriginal = utils.Uint128AStringDecimal(transferenciaTB.UserData128)
	}

	// Derivar Tipo e IdUsuarioFinal comparando DebitAccountID/CreditAccountID con la cuenta empresa
	// TODO: eliminar hardcodeo de token
	moneda := &Monedas{IdMoneda: int(transferenciaTB.Ledger)}
	if _, err := moneda.Dame("cf904666e02a79cfd50b074ab3c360c0"); err == nil && moneda.IdCuentaEmpresa != "" {
		idCuentaEmpresa, errParse := utils.ParsearUint128(moneda.IdCuentaEmpresa)
		if errParse == nil {
			var idCuentaUsuario types.Uint128
			if t.Code == CodigoTransferenciaReversion {
				t.Tipo = "R"
				// En reversión las cuentas están invertidas
				if transferenciaTB.DebitAccountID == idCuentaEmpresa {
					idCuentaUsuario = transferenciaTB.CreditAccountID
				} else {
					idCuentaUsuario = transferenciaTB.DebitAccountID
				}
			} else {
				if transferenciaTB.DebitAccountID == idCuentaEmpresa {
					t.Tipo = "I"
					idCuentaUsuario = transferenciaTB.CreditAccountID
				} else if transferenciaTB.CreditAccountID == idCuentaEmpresa {
					t.Tipo = "E"
					idCuentaUsuario = transferenciaTB.DebitAccountID
				}
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
