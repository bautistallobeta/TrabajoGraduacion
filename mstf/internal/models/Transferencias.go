package models

import (
	"encoding/binary"
	"errors"

	"MSTransaccionesFinancieras/internal/infra/persistence"
	"MSTransaccionesFinancieras/internal/utils"

	"github.com/tigerbeetle/tigerbeetle-go/pkg/types"
)

const CodigoTransferenciaNormal uint16 = 1
const CodigoTransferenciaReversion uint16 = 2
const CodigoTransferenciaCierre uint16 = 3

// "wrapper" de Transfer de TB
type Transferencias struct {
	IdTransferencia         string
	IdCuentaDebito          string
	IdCuentaCredito         string
	IdUsuarioFinal          uint64
	IdMoneda                uint32
	Monto                   string
	Tipo                    string
	Categoria               uint64
	Fecha                   string
	FechaProceso            string
	Estado                  string
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

	t.IdCuentaDebito = utils.Uint128AStringDecimal(transferenciaTB.DebitAccountID)
	t.IdCuentaCredito = utils.Uint128AStringDecimal(transferenciaTB.CreditAccountID)
	t.IdMoneda = transferenciaTB.Ledger
	t.Monto = utils.Uint128ADecimalMoneda(transferenciaTB.Amount)
	t.Categoria = transferenciaTB.UserData64
	t.Fecha = fecha
	t.FechaProceso = utils.TimestampAFecha(transferenciaTB.Timestamp)
	code := transferenciaTB.Code

	if code == CodigoTransferenciaReversion {
		t.Estado = "R"
		t.IdTransferenciaOriginal = utils.Uint128AStringDecimal(transferenciaTB.UserData128)
	} else {
		t.Estado = "F"
	}

	// Derivar Tipo e IdUsuarioFinal comparando DebitAccountID/CreditAccountID con la cuenta empresa
	// TODO: eliminar hardcodeo de token
	moneda := &Monedas{IdMoneda: int(transferenciaTB.Ledger)}
	if _, err := moneda.Dame("cf904666e02a79cfd50b074ab3c360c0"); err == nil && moneda.IdCuentaEmpresa != "" {
		idCuentaEmpresa, errParse := utils.ParsearUint128(moneda.IdCuentaEmpresa)
		if errParse == nil {
			var idCuentaUsuario types.Uint128
			if code == CodigoTransferenciaReversion {
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

// PoblarDesdeTB llena el struct con los datos directamente disponibles en el Transfer de TB,
// sin realizar consultas adicionales a cuentas ni monedas.
// Tipo queda vacío para transfers normales (requiere lookup de cuenta empresa para derivarlo).
// IdUsuarioFinal se lee de UserData128 (solo disponible en transfers creadas tras el cambio que lo almacena ahí).
func (t *Transferencias) PoblarDesdeTB(tb types.Transfer) {
	t.IdTransferencia = utils.Uint128AStringDecimal(tb.ID)
	t.IdCuentaDebito = utils.Uint128AStringDecimal(tb.DebitAccountID)
	t.IdCuentaCredito = utils.Uint128AStringDecimal(tb.CreditAccountID)
	t.IdMoneda = tb.Ledger
	t.Monto = utils.Uint128AStringDecimal(tb.Amount)
	t.Categoria = tb.UserData64

	if tb.Code == CodigoTransferenciaReversion {
		t.Estado = "R"
		t.Tipo = "R"
		t.IdTransferenciaOriginal = utils.Uint128AStringDecimal(tb.UserData128)
	} else {
		t.Estado = "F"
		t.IdUsuarioFinal = binary.LittleEndian.Uint64(tb.UserData128[:8])
	}

	if tb.UserData32 > 0 {
		if fecha, err := utils.UserData32AFecha(tb.UserData32); err == nil {
			t.Fecha = fecha
		}
	}
	if tb.Timestamp != 0 {
		t.FechaProceso = utils.TimestampAFecha(tb.Timestamp)
	}
}
