package models

import (
	"errors"

	"MSTransaccionesFinancieras/internal/infra/persistence"
	"MSTransaccionesFinancieras/internal/utils"

	"github.com/tigerbeetle/tigerbeetle-go/pkg/types"
)

// "wrapper" de Account de TB
type Cuentas struct {
	IdCuenta       string
	IdUsuarioFinal uint64
	IdMoneda       uint32
	Creditos       string
	Debitos        string
	Estado         string
	FechaAlta      string
	FechaRegistro  string
}

// TODO: limite se tiene que obtener de db relacional
const LimiteHistorialBalances uint32 = 100

func (c *Cuentas) Dame() error {
	if c.IdUsuarioFinal <= 0 || c.IdMoneda <= 0 {
		return errors.New("IdUsuarioFinal e IdMoneda son requeridos y deben ser mayores a cero")
	}

	idCuentaStr := utils.ConcatenarIDString(uint64(c.IdMoneda), c.IdUsuarioFinal)
	idCuentaCast, err := utils.ParsearUint128(idCuentaStr)
	if err != nil {
		return errors.New("Error al construir IdCuenta: " + err.Error())
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

	c.IdCuenta = utils.Uint128AStringDecimal(cuentaTB.ID)
	c.IdUsuarioFinal = cuentaTB.UserData64
	c.IdMoneda = cuentaTB.Ledger
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
func (c *Cuentas) Activar() (string, error) {
	c.Estado = "A"
	return "Activada", nil
}

// TODO
func (c *Cuentas) Desactivar() (string, error) {
	c.Estado = "I"
	return "Desactivada", nil
}

func (c *Cuentas) DameHistorialBalances(timestampMin uint64, timestampMax uint64, limite uint32) ([]types.AccountBalance, error) {
	if c.IdUsuarioFinal <= 0 || c.IdMoneda <= 0 {
		return nil, errors.New("IdUsuarioFinal e IdMoneda son requeridos y deben ser mayores a cero")
	}

	idCuentaStr := utils.ConcatenarIDString(uint64(c.IdMoneda), c.IdUsuarioFinal)
	idCuentaCast, err := utils.ParsearUint128(idCuentaStr)
	if err != nil {
		return nil, errors.New("Error al construir IdCuenta: " + err.Error())
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
