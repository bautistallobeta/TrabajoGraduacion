package models

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

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
func (c *Cuentas) ListarTransferenciasCuenta(FechaInicio uint64, FechaFin uint64, Limite uint32) ([]types.Transfer, error) {
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
func (c *Cuentas) ListarHistorialBalances(FechaInicio uint64, FechaFin uint64, Limite uint32) ([]types.AccountBalance, error) {
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

// Cierra una cuenta en TigerBeetle creando un pending transfer con closing_debit.
// La cuenta empresa de la moneda actúa como cuenta crédito (monto 0, no se transfiere dinero).
// Idempotente: si la cuenta ya está cerrada, retorna nil.
func (c *Cuentas) Desactivar() error {
	idMoneda := c.IdMoneda
	idUsuarioFinal := c.IdUsuarioFinal

	if persistence.ClienteTB == nil {
		return errors.New("Conexión a TigerBeetle no inicializada")
	}

	moneda := &Monedas{IdMoneda: int(idMoneda)}
	if _, err := moneda.Dame(); err != nil {
		return errors.New("La moneda no existe o no está activa")
	}
	if moneda.Estado != "A" || moneda.IdCuentaEmpresa == "" {
		return errors.New("La moneda no existe o no está activa")
	}

	idCuentaEmpresa, err := utils.ParsearUint128(moneda.IdCuentaEmpresa)
	if err != nil {
		return errors.New("IdCuentaEmpresa formato incorrecto")
	}

	idCuentaStr := utils.ConcatenarIDString(uint64(idMoneda), idUsuarioFinal)
	idCuenta, err := utils.ParsearUint128(idCuentaStr)
	if err != nil {
		return errors.New("Error al construir IdCuenta")
	}

	if idCuenta == idCuentaEmpresa {
		return errors.New("No se puede desactivar la cuenta empresa")
	}

	accounts, err := persistence.ClienteTB.LookupAccounts([]types.Uint128{idCuenta})
	if err != nil {
		return errors.New("error de comunicación con TigerBeetle")
	}
	if len(accounts) == 0 {
		return errors.New("Cuenta no encontrada")
	}

	flagCerrada := types.AccountFlags{Closed: true}.ToUint16()
	if (accounts[0].Flags & flagCerrada) != 0 {
		log.Printf("GestorCuentas.Desactivar: cuenta %s ya estaba cerrada", idCuentaStr)
		return nil
	}

	// ID único: high=timestamp nanosegundos, low=idUsuarioFinal
	// No colisiona con transfers normales (64 MSBits=0000...0) ni reversiones (64 MSBits=0000...1)
	idCierreStr := utils.ConcatenarIDString(uint64(time.Now().UnixNano()), idUsuarioFinal)
	idCierreTB, err := utils.ParsearUint128(idCierreStr)
	if err != nil {
		return errors.New("Error al generar ID de closing transfer")
	}

	closingTransfer := types.Transfer{
		ID:              idCierreTB,
		DebitAccountID:  idCuenta,
		CreditAccountID: idCuentaEmpresa,
		Amount:          types.ToUint128(0),
		Ledger:          idMoneda,
		Code:            CodigoTransferenciaCierre,
		Flags: types.TransferFlags{
			Pending:      true,
			ClosingDebit: true,
		}.ToUint16(),
	}

	results, err := persistence.ClienteTB.CreateTransfers([]types.Transfer{closingTransfer})
	if err != nil {
		return errors.New("error de comunicación con TigerBeetle")
	}
	if len(results) > 0 {
		return fmt.Errorf("error al desactivar cuenta: %s", results[0].Result.String())
	}

	log.Printf("GestorCuentas.Desactivar: cuenta %s desactivada exitosamente", idCuentaStr)
	return nil
}

// Reabre una cuenta cerrada en TigerBeetle voidando el closing pending transfer.
// Busca la transfer de cierre como la más reciente en débito de la cuenta (la única posible
// tras el cierre, ya que TB rechaza nuevas transfers sobre cuentas cerradas).
// Idempotente: si la cuenta ya está activa, retorna nil.
func (c *Cuentas) Activar() error {
	idMoneda := c.IdMoneda
	idUsuarioFinal := c.IdUsuarioFinal

	if persistence.ClienteTB == nil {
		return errors.New("Conexión a TigerBeetle no inicializada")
	}

	idCuentaStr := utils.ConcatenarIDString(uint64(idMoneda), idUsuarioFinal)
	idCuenta, err := utils.ParsearUint128(idCuentaStr)
	if err != nil {
		return errors.New("Error al construir IdCuenta")
	}

	accounts, err := persistence.ClienteTB.LookupAccounts([]types.Uint128{idCuenta})
	if err != nil {
		return errors.New("error de comunicación con TigerBeetle")
	}
	if len(accounts) == 0 {
		return errors.New("Cuenta no encontrada")
	}

	flagCerrada := types.AccountFlags{Closed: true}.ToUint16()
	if (accounts[0].Flags & flagCerrada) == 0 {
		log.Printf("GestorCuentas.Activar: cuenta %s ya estaba activa", idCuentaStr)
		return nil
	}

	// La transfer más reciente en débito debe ser el closing pending transfer
	// (Aunque igualmente TB no admite nuevas transfers sobre cuentas cerradas)
	filtro := types.AccountFilter{
		AccountID: idCuenta,
		Limit:     1,
		Flags:     5, // Debits(1) + Reversed(4)
	}
	transfers, err := persistence.ClienteTB.GetAccountTransfers(filtro)
	if err != nil {
		return errors.New("error de comunicación con TigerBeetle")
	}
	if len(transfers) == 0 {
		return errors.New("No se encontró el closing transfer de la cuenta")
	}

	closingTransfer := transfers[0]
	flagPending := types.TransferFlags{Pending: true}.ToUint16()
	if (closingTransfer.Flags & flagPending) == 0 {
		return errors.New("La transfer de cierre no está en estado pendiente")
	}

	// ID único para el void transfer
	idVoidStr := utils.ConcatenarIDString(uint64(time.Now().UnixNano()), idUsuarioFinal)
	idVoidTB, err := utils.ParsearUint128(idVoidStr)
	if err != nil {
		return errors.New("Error al generar ID de void transfer")
	}

	voidTransfer := types.Transfer{
		ID:        idVoidTB,
		PendingID: closingTransfer.ID,
		Code:      CodigoTransferenciaCierre, // mismo code que el closing
		Flags:     types.TransferFlags{VoidPendingTransfer: true}.ToUint16(),
	}

	results, err := persistence.ClienteTB.CreateTransfers([]types.Transfer{voidTransfer})
	if err != nil {
		return errors.New("error de comunicación con TigerBeetle")
	}
	if len(results) > 0 {
		return fmt.Errorf("error al activar cuenta: %s", results[0].Result.String())
	}

	log.Printf("GestorCuentas.Activar: cuenta %s activada exitosamente", idCuentaStr)
	return nil
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
