package gestores

import (
	"MSTransaccionesFinancieras/internal/infra/persistence"
	"MSTransaccionesFinancieras/internal/models"
	"MSTransaccionesFinancieras/internal/utils"
	"errors"
	"fmt"
	"log"

	"github.com/tigerbeetle/tigerbeetle-go/pkg/types"
)

type CuentaNueva struct {
	IdMoneda                      uint32
	IdUsuarioFinal                uint64
	FechaAlta                     string
	DebitosNoDebenExcederCreditos bool
}

type GestorCuentas struct {
}

func NewGestorCuentas() *GestorCuentas {
	return &GestorCuentas{}
}

// Buscar cuentas según los filtros especificados.
// Parámetros con valor 0 desactivan ese filtro en TigerBeetle.
// estado: "A" para activas, "I" para inactivas/cerradas, "" para todas.
// limit: máximo número de cuentas a retornar (0 = sin límite).
func (gc *GestorCuentas) Buscar(
	idUsuarioFinal uint64,
	idMoneda uint32,
	estado string,
	limit uint32,
) ([]types.Account, error) {

	if persistence.ClienteTB == nil {
		return nil, errors.New("Conexión a TigerBeetle no inicializada")
	}

	resultados := make([]types.Account, 0)

	// Timestamp para paginación
	var timestampMin uint64 = 0

	batchSize := limit

	// Iterar hasta alcanzar el límite total solicitado
	for {
		// Calcular cuántos resultados faltan para alcanzar el límite
		restantes := limit - uint32(len(resultados))
		if restantes == 0 {
			log.Printf("GestorCuentas.Buscar: Límite alcanzado (%d cuentas)", limit)
			break
		}

		// Ajustar batch size para no buscar más resultados de lo necesario
		currentBatchSize := batchSize
		if restantes < currentBatchSize {
			currentBatchSize = restantes
		}

		filter := types.QueryFilter{
			UserData128:  types.ToUint128(0),
			UserData64:   idUsuarioFinal,
			UserData32:   0,
			Code:         0,
			Ledger:       idMoneda,
			TimestampMin: timestampMin,
			TimestampMax: 0,
			Limit:        currentBatchSize,
			Flags: types.QueryFilterFlags{
				Reversed: false,
			}.ToUint32(),
		}

		log.Printf("GestorCuentas.Buscar: Ejecutando QueryAccounts (TimestampMin=%d, Limit=%d)", timestampMin, currentBatchSize)
		accounts, err := persistence.ClienteTB.QueryAccounts(filter)
		if err != nil {
			log.Printf("ERROR [GestorCuentas.Buscar]: Fallo QueryAccounts: %v", err)
			return nil, err
		}

		log.Printf("GestorCuentas.Buscar: Obtenidos %d resultados en esta iteración", len(accounts))

		if len(accounts) == 0 {
			break
		}

		resultados = append(resultados, accounts...)

		if uint32(len(accounts)) < currentBatchSize {
			log.Printf("GestorCuentas.Buscar: Obtenidos %d < %d, no hay más resultados", len(accounts), currentBatchSize)
			break
		}
		ultimoTimestamp := accounts[len(accounts)-1].Timestamp
		timestampMin = ultimoTimestamp + 1

		log.Printf("GestorCuentas.Buscar: Obtenidos %d == %d, continuando iteración", len(accounts), currentBatchSize)

		// (TO DO: mejorar paginación - traer siempre el max que se musetran en la pagina y avanzar con el timestamp del mas viejo en cada llamado)
		if len(resultados) > 50000 {
			log.Printf("ADVERTENCIA: Se alcanzó el límite de seguridad de 50,000 cuentas")
			break
		}
	}

	log.Printf("GestorCuentas.Buscar: Total acumulado antes de filtrar por estado: %d cuentas", len(resultados))

	if estado != "" {
		resultados = filtrarPorEstado(resultados, estado)
		log.Printf("GestorCuentas.Buscar: Después de filtrar por estado '%s': %d cuentas", estado, len(resultados))
	}

	return resultados, nil
}

// TODO: agregar comentario de método de clase
func (gc *GestorCuentas) Crear(idMoneda uint32, idUsuarioFinal uint64, fechaAlta string, debitosNoDebenExcederCreditos bool) (string, error) {
	if persistence.ClienteTB == nil {
		return "", errors.New("Conexión a TigerBeetle no inicializada")
	}

	// Verificar que la moneda exista y esté activa
	// TODO: el token está hardcodeado, reemplazar cuando se resuelva cómo obtenerlo en esta capa
	moneda := &models.Monedas{IdMoneda: int(idMoneda)}
	if _, err := moneda.Dame("cf904666e02a79cfd50b074ab3c360c0"); err != nil {
		return "", errors.New("La moneda no existe o no está activa")
	}
	if moneda.Estado != "A" {
		return "", errors.New("La moneda no existe o no está activa")
	}

	// idMoneda son los 64 bits mas significativos y idUsuarioFinal los menos significativos
	idCuenta := utils.ConcatenarIDString(uint64(idMoneda), idUsuarioFinal)
	if idCuenta == "" || idCuenta == "0" {
		return "", errors.New("IdCuenta no puede ser vacío o cero")
	}

	tbId, err := utils.ParsearUint128(idCuenta)
	if err != nil {
		return "", errors.New("IdCuenta formato incorrecto")
	}

	fechaAltaUint32, err := utils.FechaAUserData32(fechaAlta)
	if err != nil {
		return "", errors.New("Formato de FechaAlta inválido: " + err.Error())
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
		return "", errors.New("error de comunicación con TigerBeetle")
	}
	if len(results) > 0 {
		return "", errors.New("fallo en la creación de la cuenta: " + results[0].Result.String())
	}
	return idCuenta, nil
}

// Crea múltiples cuentas en TigerBeetle en un solo llamado.
// Recibe los mismos datos que Crear pero como array.
func (gc *GestorCuentas) CrearLote(cuentas []CuentaNueva) ([]string, error) {
	if persistence.ClienteTB == nil {
		return nil, errors.New("Conexión a TigerBeetle no inicializada")
	}
	if len(cuentas) == 0 {
		return []string{}, nil
	}

	// Verificar que la moneda de cada cuenta exista y esté activa
	// TODO: el token está hardcodeado, reemplazar cuando se resuelva cómo obtenerlo en esta capa
	for _, c := range cuentas {
		moneda := &models.Monedas{IdMoneda: int(c.IdMoneda)}
		if _, err := moneda.Dame("cf904666e02a79cfd50b074ab3c360c0"); err != nil {
			return nil, fmt.Errorf("La moneda no existe o no está activa (IdMoneda=%d)", c.IdMoneda)
		}
		if moneda.Estado != "A" {
			return nil, fmt.Errorf("La moneda no existe o no está activa (IdMoneda=%d)", c.IdMoneda)
		}
	}

	cuentasTB := make([]types.Account, 0, len(cuentas))
	ids := make([]string, 0, len(cuentas))

	for _, c := range cuentas {
		idCuenta := utils.ConcatenarIDString(uint64(c.IdMoneda), c.IdUsuarioFinal)
		if idCuenta == "" || idCuenta == "0" {
			return nil, fmt.Errorf("IdCuenta no puede ser vacío o cero (IdMoneda=%d, Usuario=%d)", c.IdMoneda, c.IdUsuarioFinal)
		}

		tbId, err := utils.ParsearUint128(idCuenta)
		if err != nil {
			return nil, fmt.Errorf("IdCuenta formato incorrecto (IdMoneda=%d, Usuario=%d)", c.IdMoneda, c.IdUsuarioFinal)
		}

		fechaAltaUint32, err := utils.FechaAUserData32(c.FechaAlta)
		if err != nil {
			return nil, fmt.Errorf("Formato de FechaAlta inválido para cuenta IdMoneda=%d: %v", c.IdMoneda, err)
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
	if len(results) > 0 {
		for _, r := range results {
			if int(r.Index) < len(ids) {
				log.Printf("ERROR [GestorCuentas.CrearLote]: Cuenta %s falló: %s", ids[r.Index], r.Result.String())
			}
		}
		return nil, fmt.Errorf("fallo en la creación de %d de %d cuentas", len(results), len(cuentas))
	}

	return ids, nil
}

// TODO
func (gc *GestorCuentas) Borrar(id types.Uint128) (string, error) {
	return "Borrado lógico no implementado", errors.New("No implementado")
}

// TODO
func (gc *GestorCuentas) Modificar(cuenta models.Cuentas) (string, error) {
	return "Modificación no implementada", errors.New("No implementado")
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
