package models

import (
	"errors"
	"strconv"

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
	Fecha          string
	FechaProceso   string
}

const limiteHistorialBalancesPorDefecto uint32 = 100

// Instancia los datos de la cuenta leyendo desde TigerBeetle a partir de IdUsuarioFinal e IdMoneda
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
	c.Creditos = utils.Uint128ADecimalMoneda(cuentaTB.CreditsPosted)
	c.Debitos = utils.Uint128ADecimalMoneda(cuentaTB.DebitsPosted)

	// Leer Fecha desde UserData32
	if cuentaTB.UserData32 != 0 {
		fecha, err := utils.UserData32AFecha(cuentaTB.UserData32)
		if err == nil {
			c.Fecha = fecha
		}
	}

	// Leer FechaProceso desde Timestamp de TB
	if cuentaTB.Timestamp != 0 {
		c.FechaProceso = utils.TimestampAFecha(cuentaTB.Timestamp)
	}

	closedFlags := types.AccountFlags{Closed: true}.ToUint16()

	if (cuentaTB.Flags & closedFlags) != 0 {
		c.Estado = "I"
	} else {
		c.Estado = "A"
	}

	return nil
}

// Busca las transferencias asociadas a la cuenta (como débito o crédito) en un rango de fechas, con un límite máximo de resultados
// Ordena de la más reciente a la más antigua (por Timestamp de TB)
// - FechaInicio: timestamp mínimo (inclusive) de las transferencias a buscar
// - FechaFin: timestamp máximo (inclusive) de las transferencias a buscar
// - Limite: cantidad máxima de transferencias a retornar (si es 0, se usa un valor por defecto)
func (c *Cuentas) BuscarTransferenciasCuenta(FechaInicio uint64, FechaFin uint64, Limite uint32) ([]types.Transfer, error) {
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

	if Limite <= 0 {
		Limite = obtenerLimiteHistorialBalances()
	}

	filtro := types.AccountFilter{
		AccountID:    idCuentaCast,
		TimestampMin: FechaInicio,
		TimestampMax: FechaFin,
		Limit:        Limite,
		Flags:        7, // Debits(1) + Credits(2) + Reversed(4)
	}

	return persistence.ClienteTB.GetAccountTransfers(filtro)
}

// Busca el historial de balances de la cuenta en un rango de fechas, con un límite máximo de resultados
// Ordena de la más reciente a la más antigua (por Timestamp de TB)
// - FechaInicio: timestamp mínimo (inclusive) de los balances a buscar
// - FechaFin: timestamp máximo (inclusive) de los balances a buscar
// - Limite: cantidad máxima de balances a retornar (si es 0, se usa un valor por defecto)
func (c *Cuentas) DameHistorialBalances(FechaInicio uint64, FechaFin uint64, Limite uint32) ([]types.AccountBalance, error) {
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

	if Limite <= 0 {
		Limite = obtenerLimiteHistorialBalances()
	}

	filtro := types.AccountFilter{
		AccountID:    idCuentaCast,
		TimestampMin: FechaInicio,
		TimestampMax: FechaFin,
		Limit:        Limite,
		Flags:        3, // Debits(1) + Credits(2)
	}

	balances, err := persistence.ClienteTB.GetAccountBalances(filtro)
	if err != nil {
		return nil, err
	}
	return balances, nil
}

// --------------------------------------------------------------------------------
// Funciones aux
// --------------------------------------------------------------------------------
func obtenerLimiteHistorialBalances() uint32 {
	p := &Parametros{Parametro: "LIMITEHISTORIALBALANCE"}
	if _, err := p.Dame(); err != nil || p.Valor == "" {
		return limiteHistorialBalancesPorDefecto
	}
	val, err := strconv.ParseUint(p.Valor, 10, 32)
	if err != nil {
		return limiteHistorialBalancesPorDefecto
	}
	return uint32(val)
}
