package gestores

import (
	"MSTransaccionesFinancieras/internal/infra/persistence"
	"MSTransaccionesFinancieras/internal/models"
	"MSTransaccionesFinancieras/internal/utils"
	"errors"
	"log"

	tigerbeetle "github.com/tigerbeetle/tigerbeetle-go"
	"github.com/tigerbeetle/tigerbeetle-go/pkg/types"
)

type GestorCuentas struct {
}

func NewGestorCuentas(tbClient tigerbeetle.Client) *GestorCuentas {
	return &GestorCuentas{}
}

// Buscar cuentas según los filtros especificados.
// Parámetros con valor 0 desactivan ese filtro en TigerBeetle.
// estado: "A" para activas, "I" para inactivas/cerradas, "" para todas.
// limit: máximo número de cuentas a retornar (0 = sin límite).
func (gc *GestorCuentas) Buscar(
	idUsuarioFinal uint64,
	idLedger uint32,
	tipo uint16,
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
			Code:         tipo,
			Ledger:       idLedger,
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
func (gc *GestorCuentas) Crear(idLedger uint32, idUsuarioFinal uint64, fechaAlta string, debitosNoDebenExcederCreditos bool) (string, error) {
	if persistence.ClienteTB == nil {
		return "", errors.New("Conexión a TigerBeetle no inicializada")
	}

	// idLedger son los 64 bits mas significativa y idUsuarioFinal los menos significativa
	idCuenta := utils.ConcatenarIDString(uint64(idLedger), idUsuarioFinal)
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
		Ledger:     idLedger,
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
