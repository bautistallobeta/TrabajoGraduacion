package models

import (
	"errors"

	"MSTransaccionesFinancieras/internal/infra/persistence"
	"MSTransaccionesFinancieras/internal/utils"

	"github.com/tigerbeetle/tigerbeetle-go/pkg/types"
)

// "wrapper" de Account de TB
type Cuenta struct {
	IdCuenta       string
	IdUsuarioFinal uint64
	IdLedger       uint32
	Creditos       string
	Debitos        string
	Tipo           uint64
	Estado         string
	FechaAlta      string
	FechaRegistro  string
}

// TODO: limite se tiene que obtener de db relacional
const LimiteHistorialBalances uint32 = 100

func (c *Cuenta) Dame() error {
	idCuentaCast, err := utils.ParsearUint128(c.IdCuenta)
	if err != nil {
		return errors.New("IdCuenta formato incorrecto")
	}
	if idCuentaCast == (types.Uint128{}) || idCuentaCast == types.ToUint128(0) {
		return errors.New("IdCuenta no puede ser nulo ni cero")
	}

	if persistence.ClienteTB == nil {
		return errors.New("Conexión a TigerBeetle no inicializada")
	}

	accounts, err := persistence.ClienteTB.LookupAccounts([]types.Uint128{idCuentaCast})
	if err != nil {
		return err
	}

	if len(accounts) == 0 {
		return errors.New("Cuenta no encontrada en TigerBeetle")
	}

	cuentaTB := accounts[0]

	c.IdUsuarioFinal = cuentaTB.UserData64
	c.IdLedger = cuentaTB.Ledger
	c.Tipo = uint64(cuentaTB.Code)
	c.Creditos = utils.Uint128AStringDecimal(cuentaTB.CreditsPosted)
	c.Debitos = utils.Uint128AStringDecimal(cuentaTB.DebitsPosted)

	// Leer FechaAlta desde UserData32
	if cuentaTB.UserData32 != 0 {
		fechaAlta, err := utils.UserData32AFecha(cuentaTB.UserData32)
		if err == nil {
			c.FechaAlta = fechaAlta
		}
	}

	// Leer FechaRegistro desde Timestamp de TB
	if cuentaTB.Timestamp != 0 {
		c.FechaRegistro = utils.TimestampAFecha(cuentaTB.Timestamp)
	}

	closedFlags := types.AccountFlags{Closed: true}.ToUint16()

	if (cuentaTB.Flags & closedFlags) != 0 {
		c.Estado = "I"
	} else {
		c.Estado = "A"
	}

	return nil
}

// TODO
func (c *Cuenta) Activar() (string, error) {
	c.Estado = "A"
	return "Activada", nil
}

// TODO
func (c *Cuenta) Desactivar() (string, error) {
	c.Estado = "I"
	return "Desactivada", nil
}

// TODO: comentario de método de clase
func (c *Cuenta) DameHistorialBalances(timestampMin uint64, timestampMax uint64, limite uint32) ([]types.AccountBalance, error) {
	if c.IdCuenta == "" {
		return nil, errors.New("IdCuenta no puede estar vacío")
	}

	idCuentaCast, err := utils.ParsearUint128(c.IdCuenta)
	if err != nil {
		return nil, errors.New("IdCuenta formato incorrecto")
	}

	if idCuentaCast == (types.Uint128{}) || idCuentaCast == types.ToUint128(0) {
		return nil, errors.New("IdCuenta no puede ser nulo ni cero")
	}

	if persistence.ClienteTB == nil {
		return nil, errors.New("Conexión a TigerBeetle no inicializada")
	}

	if limite <= 0 {
		limite = LimiteHistorialBalances
	}

	filtro := types.AccountFilter{
		AccountID:    idCuentaCast,
		TimestampMin: timestampMin,
		TimestampMax: timestampMax,
		Limit:        limite,
		Flags:        0,
	}

	balances, err := persistence.ClienteTB.GetAccountBalances(filtro)
	if err != nil {
		return nil, err
	}
	return balances, nil
}
