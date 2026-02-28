package gestores

import (
	"MSTransaccionesFinancieras/internal/infra/persistence"
	"MSTransaccionesFinancieras/internal/models"
	"MSTransaccionesFinancieras/internal/utils"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/tigerbeetle/tigerbeetle-go/pkg/types"
)

type CuentaNueva struct {
	IdMoneda                      uint32
	IdUsuarioFinal                uint64
	Fecha                         string
	DebitosNoDebenExcederCreditos bool
}

type GestorCuentas struct {
}

func NewGestorCuentas() *GestorCuentas {
	return &GestorCuentas{}
}

// BuscarAvanzado busca cuentas según los filtros especificados.
// Si IdsCuenta tiene elementos, hace LookupAccounts directo e ignora el resto de filtros.
// Parámetros con valor 0 desactivan ese filtro en TigerBeetle.
// Estado: "A" para activas, "I" para inactivas/cerradas, "" para todas.
// Limit: máximo número de cuentas a retornar (0 = sin límite).
func (gc *GestorCuentas) BuscarAvanzado(
	IdsCuenta []types.Uint128,
	IdUsuarioFinal uint64,
	IdMoneda uint32,
	Estado string,
	Limit uint32,
) ([]types.Account, error) {

	if persistence.ClienteTB == nil {
		return nil, errors.New("Conexión a TigerBeetle no inicializada")
	}

	// Camino A: lookup directo por IDs (ignora el resto de parámetros)
	if len(IdsCuenta) > 0 {
		return persistence.ClienteTB.LookupAccounts(IdsCuenta)
	}

	resultados := make([]types.Account, 0)

	// Timestamp para paginación
	var timestampMin uint64 = 0

	batchSize := Limit

	// Iterar hasta alcanzar el límite total solicitado
	for {
		// Calcular cuántos resultados faltan para alcanzar el límite
		restantes := Limit - uint32(len(resultados))
		if restantes == 0 {
			log.Printf("GestorCuentas.BuscarAvanzado: Límite alcanzado (%d cuentas)", Limit)
			break
		}

		// Ajustar batch size para no buscar más resultados de lo necesario
		currentBatchSize := batchSize
		if restantes < currentBatchSize {
			currentBatchSize = restantes
		}

		filter := types.QueryFilter{
			UserData128:  types.ToUint128(0),
			UserData64:   IdUsuarioFinal,
			UserData32:   0,
			Code:         0,
			Ledger:       IdMoneda,
			TimestampMin: timestampMin,
			TimestampMax: 0,
			Limit:        currentBatchSize,
			Flags: types.QueryFilterFlags{
				Reversed: false,
			}.ToUint32(),
		}

		log.Printf("GestorCuentas.BuscarAvanzado: Ejecutando QueryAccounts (TimestampMin=%d, Limit=%d)", timestampMin, currentBatchSize)
		accounts, err := persistence.ClienteTB.QueryAccounts(filter)
		if err != nil {
			log.Printf("ERROR [GestorCuentas.BuscarAvanzado]: Fallo QueryAccounts: %v", err)
			return nil, err
		}

		log.Printf("GestorCuentas.BuscarAvanzado: Obtenidos %d resultados en esta iteración", len(accounts))

		if len(accounts) == 0 {
			break
		}

		resultados = append(resultados, accounts...)

		if uint32(len(accounts)) < currentBatchSize {
			log.Printf("GestorCuentas.BuscarAvanzado: Obtenidos %d < %d, no hay más resultados", len(accounts), currentBatchSize)
			break
		}
		ultimoTimestamp := accounts[len(accounts)-1].Timestamp
		timestampMin = ultimoTimestamp + 1

		log.Printf("GestorCuentas.BuscarAvanzado: Obtenidos %d == %d, continuando iteración", len(accounts), currentBatchSize)

		// (TO DO: mejorar paginación - traer siempre el max que se musetran en la pagina y avanzar con el timestamp del mas viejo en cada llamado)
		if len(resultados) > 50000 {
			log.Printf("ADVERTENCIA: Se alcanzó el límite de seguridad de 50,000 cuentas")
			break
		}
	}

	log.Printf("GestorCuentas.BuscarAvanzado: Total acumulado antes de filtrar por estado: %d cuentas", len(resultados))

	if Estado != "" {
		resultados = filtrarPorEstado(resultados, Estado)
		log.Printf("GestorCuentas.BuscarAvanzado: Después de filtrar por estado '%s': %d cuentas", Estado, len(resultados))
	}

	return resultados, nil
}

// Crear intenta crear una cuenta en TigerBeetle.
// Retorna (idCuenta, existe, error).
// existe=true indica que la cuenta ya existía con los mismos parámetros (idempotencia ante reintentos).
// Si Cuenta.IdUsuarioFinal es 0, se trata como cuenta empresa (DebitsMustNotExceedCredits=false).
func (gc *GestorCuentas) Crear(Cuenta models.Cuentas) (string, bool, error) {
	idMoneda := Cuenta.IdMoneda
	idUsuarioFinal := Cuenta.IdUsuarioFinal
	fechaAlta := Cuenta.Fecha
	debitosNoDebenExcederCreditos := Cuenta.IdUsuarioFinal != 0

	if persistence.ClienteTB == nil {
		return "", false, errors.New("Conexión a TigerBeetle no inicializada")
	}

	// Verificar que la moneda exista y esté activa
	moneda := &models.Monedas{IdMoneda: int(idMoneda)}
	if _, err := moneda.Dame(); err != nil {
		return "", false, errors.New("La moneda no existe o no está activa")
	}
	// Chequeo que se hace solo si la cuenta no es cuentaempresa
	if debitosNoDebenExcederCreditos && moneda.Estado != "A" {
		return "", false, errors.New("La moneda no existe o no está activa")
	}

	// idMoneda son los 64 bits mas significativos y idUsuarioFinal los menos significativos
	idCuenta := utils.ConcatenarIDString(uint64(idMoneda), idUsuarioFinal)
	if idCuenta == "" || idCuenta == "0" {
		return "", false, errors.New("IdCuenta no puede ser vacío o cero")
	}

	tbId, err := utils.ParsearUint128(idCuenta)
	if err != nil {
		return "", false, errors.New("IdCuenta formato incorrecto")
	}

	fechaAltaUint32, err := utils.FechaAUserData32(fechaAlta)
	if err != nil {
		return "", false, errors.New("Formato de Fecha inválido: " + err.Error())
	}

	cuentaTB := types.Account{
		ID:         tbId,
		Ledger:     idMoneda,
		Code:       1,
		UserData64: idUsuarioFinal,
		UserData32: fechaAltaUint32,
		Flags: types.AccountFlags{
			DebitsMustNotExceedCredits: debitosNoDebenExcederCreditos,
			History:                    true,
		}.ToUint16(),
	}

	results, err := persistence.ClienteTB.CreateAccounts([]types.Account{cuentaTB})
	if err != nil {
		return "", false, errors.New("error de comunicación con TigerBeetle")
	}
	if len(results) > 0 {
		// AccountExists significa que ya existe con los mismos parámetros: idempotencia ante reintentos
		if results[0].Result == types.AccountExists {
			return idCuenta, true, nil
		}
		return "", false, errors.New("fallo en la creación de la cuenta: " + results[0].Result.String())
	}
	return idCuenta, false, nil
}

// CrearLote crea múltiples cuentas en TigerBeetle en un solo llamado.
// Recibe los mismos datos que Crear pero como array.
func (gc *GestorCuentas) CrearLote(Cuentas []CuentaNueva) ([]string, error) {
	if persistence.ClienteTB == nil {
		return nil, errors.New("Conexión a TigerBeetle no inicializada")
	}
	if len(Cuentas) == 0 {
		return []string{}, nil
	}

	// Verificar que la moneda de cada cuenta exista y esté activa
	for _, c := range Cuentas {
		moneda := &models.Monedas{IdMoneda: int(c.IdMoneda)}
		if _, err := moneda.Dame(); err != nil {
			return nil, fmt.Errorf("La moneda no existe o no está activa (IdMoneda=%d)", c.IdMoneda)
		}
		if moneda.Estado != "A" {
			return nil, fmt.Errorf("La moneda no existe o no está activa (IdMoneda=%d)", c.IdMoneda)
		}
	}

	cuentasTB := make([]types.Account, 0, len(Cuentas))
	ids := make([]string, 0, len(Cuentas))

	for _, c := range Cuentas {
		idCuenta := utils.ConcatenarIDString(uint64(c.IdMoneda), c.IdUsuarioFinal)
		if idCuenta == "" || idCuenta == "0" {
			return nil, fmt.Errorf("IdCuenta no puede ser vacío o cero (IdMoneda=%d, Usuario=%d)", c.IdMoneda, c.IdUsuarioFinal)
		}

		tbId, err := utils.ParsearUint128(idCuenta)
		if err != nil {
			return nil, fmt.Errorf("IdCuenta formato incorrecto (IdMoneda=%d, Usuario=%d)", c.IdMoneda, c.IdUsuarioFinal)
		}

		fechaAltaUint32, err := utils.FechaAUserData32(c.Fecha)
		if err != nil {
			return nil, fmt.Errorf("Formato de Fecha inválido para cuenta IdMoneda=%d: %v", c.IdMoneda, err)
		}

		cuentaTB := types.Account{
			ID:         tbId,
			Ledger:     c.IdMoneda,
			Code:       1,
			UserData64: c.IdUsuarioFinal,
			UserData32: fechaAltaUint32,
			Flags: types.AccountFlags{
				DebitsMustNotExceedCredits: c.DebitosNoDebenExcederCreditos,
				History:                    true,
			}.ToUint16(),
		}

		cuentasTB = append(cuentasTB, cuentaTB)
		ids = append(ids, idCuenta)
	}

	results, err := persistence.ClienteTB.CreateAccounts(cuentasTB)
	if err != nil {
		return nil, errors.New("error de comunicación con TigerBeetle")
	}
	fallosReales := 0
	for _, r := range results {
		// AccountExists significa que ya existe con los mismos parámetros: idempotencia ante reintentos
		if r.Result == types.AccountExists {
			continue
		}
		fallosReales++
		if int(r.Index) < len(ids) {
			log.Printf("ERROR [GestorCuentas.CrearLote]: Cuenta %s falló: %s", ids[r.Index], r.Result.String())
		}
	}
	if fallosReales > 0 {
		return nil, fmt.Errorf("fallo en la creación de %d de %d cuentas", fallosReales, len(Cuentas))
	}

	return ids, nil
}

// Desactivar cierra una cuenta en TigerBeetle creando un pending transfer con closing_debit.
// La cuenta empresa de la moneda actúa como cuenta crédito (monto 0, no se transfiere dinero).
// Idempotente: si la cuenta ya está cerrada, retorna nil.
func (gc *GestorCuentas) Desactivar(Cuenta models.Cuentas) error {
	idMoneda := Cuenta.IdMoneda
	idUsuarioFinal := Cuenta.IdUsuarioFinal

	if persistence.ClienteTB == nil {
		return errors.New("Conexión a TigerBeetle no inicializada")
	}

	moneda := &models.Monedas{IdMoneda: int(idMoneda)}
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
	// No colisiona con transfers normales (upper=0) ni reversiones (upper=1)
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
		Code:            models.CodigoTransferenciaCierre,
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

// Activar reabre una cuenta cerrada en TigerBeetle voidando el closing pending transfer.
// Busca la transfer de cierre como la más reciente en débito de la cuenta (la única posible
// tras el cierre, ya que TB rechaza nuevas transfers sobre cuentas cerradas).
// Idempotente: si la cuenta ya está activa, retorna nil.
func (gc *GestorCuentas) Activar(Cuenta models.Cuentas) error {
	idMoneda := Cuenta.IdMoneda
	idUsuarioFinal := Cuenta.IdUsuarioFinal

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
	// (TB no admite nuevas transfers sobre cuentas cerradas)
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
		Code:      models.CodigoTransferenciaCierre, // mismo code que el closing → filtrado en Buscar
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
// Funciones Aux
// --------------------------------------------------------------------------------

// filtra un slice de accounts por estado (A/I)
func filtrarPorEstado(accounts []types.Account, estado string) []types.Account {
	closedFlag := types.AccountFlags{Closed: true}.ToUint16()
	resultado := make([]types.Account, 0, len(accounts))

	for _, acc := range accounts {
		esCerrada := (acc.Flags & closedFlag) != 0

		if estado == "I" && esCerrada {
			//inactivas/cerradas
			resultado = append(resultado, acc)
		} else if estado == "A" && !esCerrada {
			//activas
			resultado = append(resultado, acc)
		}
	}
	return resultado
}
