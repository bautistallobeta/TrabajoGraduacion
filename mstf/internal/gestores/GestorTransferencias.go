package gestores

import (
	"MSTransaccionesFinancieras/internal/infra/persistence"
	"MSTransaccionesFinancieras/internal/infra/webhook"
	"MSTransaccionesFinancieras/internal/models"
	"MSTransaccionesFinancieras/internal/utils"
	"encoding/binary"
	"errors"
	"log"
	"strconv"

	"github.com/tigerbeetle/tigerbeetle-go/pkg/types"
)

type GestorTransferencias struct {
	notificador *webhook.Notificador
}

func NewGestorTransferencias(notificador *webhook.Notificador) *GestorTransferencias {
	return &GestorTransferencias{
		notificador: notificador,
	}
}

// BuscarAvanzado busca transferencias según los filtros especificados.
// Si IdsTransferencia tiene elementos, hace LookupTransfers directo e ignora el resto de filtros.
// Parámetros numéricos con valor 0 deshabilitan ese filtro en TigerBeetle.
// Estado: "F" para finalizadas, "R" para revertidas, "" para todas.
// MontoMin/MontoMax: filtrado client-side, 0 = sin límite.
// FechaInicio/FechaFin: nanosegundos epoch (Timestamp de TB, no UserData32).
// Los resultados se ordenan de más reciente a más antigua.
func (gt *GestorTransferencias) BuscarAvanzado(
	IdsTransferencia []types.Uint128,
	IdUsuarioFinal uint64,
	IdCategoria uint64,
	IdMoneda uint32,
	Estado string,
	MontoMin uint64,
	MontoMax uint64,
	FechaInicio uint64,
	FechaFin uint64,
	Limit uint32,
) ([]types.Transfer, error) {
	if persistence.ClienteTB == nil {
		return nil, errors.New("Conexión a TigerBeetle no inicializada")
	}

	// Camino A: lookup directo por IDs (ignora el resto de parámetros)
	if len(IdsTransferencia) > 0 {
		return persistence.ClienteTB.LookupTransfers(IdsTransferencia)
	}

	// Mapear estado a Code de TB
	var codigo uint16 = 0
	switch Estado {
	case "F":
		codigo = models.CodigoTransferenciaNormal
	case "R":
		codigo = models.CodigoTransferenciaReversion
	}

	resultados := make([]types.Transfer, 0)
	cursorTimestampMax := FechaFin // avanza hacia atrás en cada página (reversed=true)

	for {
		restantes := Limit - uint32(len(resultados))
		if restantes == 0 {
			break
		}

		filter := types.QueryFilter{
			UserData128:  types.ToUint128(IdUsuarioFinal),
			UserData64:   IdCategoria,
			UserData32:   0,
			Code:         codigo,
			Ledger:       IdMoneda,
			TimestampMin: FechaInicio,
			TimestampMax: cursorTimestampMax,
			Limit:        restantes,
			Flags:        types.QueryFilterFlags{Reversed: true}.ToUint32(),
		}

		log.Printf("GestorTransferencias.BuscarAvanzado: Ejecutando QueryTransfers (TimestampMax=%d, Limit=%d)", cursorTimestampMax, restantes)
		transfers, err := persistence.ClienteTB.QueryTransfers(filter)
		if err != nil {
			log.Printf("ERROR [GestorTransferencias.BuscarAvanzado]: QueryTransfers falló: %v", err)
			return nil, err
		}

		if len(transfers) == 0 {
			break
		}

		for _, t := range transfers {
			if t.Code == models.CodigoTransferenciaCierre {
				continue
			}
			if !pasaFiltroMonto(t, MontoMin, MontoMax) {
				continue
			}
			resultados = append(resultados, t)
			if uint32(len(resultados)) >= Limit {
				break
			}
		}

		if uint32(len(transfers)) < restantes {
			break
		}

		if uint32(len(resultados)) >= Limit {
			break
		}

		// el último elemento es el más antiguo del batch (reversed=true); avanzar cursor hacia atrás
		ultimoTimestamp := transfers[len(transfers)-1].Timestamp
		if ultimoTimestamp == 0 {
			break
		}
		cursorTimestampMax = ultimoTimestamp - 1

		if len(resultados) > 50000 {
			log.Printf("ADVERTENCIA [GestorTransferencias.BuscarAvanzado]: Límite de seguridad alcanzado (50.000 transferencias)")
			break
		}
	}

	log.Printf("GestorTransferencias.BuscarAvanzado: Total encontrado: %d transferencias", len(resultados))
	return resultados, nil
}

// ProcesarLote procesa un lote de transferencias recibido del consumidor Kafka.
// Valida reglas de negocio antes de enviar a TigerBeetle.
// Las transferencias que fallan validación no van a TB pero sí se notifican con su error.
// FallidasParseo son transferencias que fallaron en el parseo del mensaje Kafka (también se notifican)
func (gt *GestorTransferencias) ProcesarLote(Batch []types.Transfer, KafkaMsgs []models.KafkaTransferencias, FallidasParseo []models.TransferenciaNotificada) error {
	var paraEnviar []types.Transfer
	var kafkaMsgsValidos []models.KafkaTransferencias
	fallidas := append([]models.TransferenciaNotificada{}, FallidasParseo...)

	// validar existencia y saldo de cuentas antes de ir a TigerBeetle
	erroresCuentas, err := gt.preValidarCuentas(Batch)
	if err != nil {
		// error de infraestructura (TB caído) → no notificar, dejar que procesarConRetry reintente
		log.Printf("ERROR [GestorTransferencias.ProcesarLote]: Error de infraestructura en preValidarCuentas: %v", err)
		return err
	}

	for i, t := range Batch {
		if erroresCuentas[i] != "" {
			log.Printf("[VALIDACIÓN] Transfer ID %s rechazada: %s", utils.Uint128AStringDecimal(t.ID), erroresCuentas[i])
			fallidas = append(fallidas, models.NewTransferenciaNotificadaError(t, KafkaMsgs[i], erroresCuentas[i]))
			continue
		}

		// validaciones de reglas de negocio (montos, moneda, reversión)
		var estadoError string
		if t.Code == models.CodigoTransferenciaReversion {
			var errInfra error
			estadoError, errInfra = gt.validarReversion(t, KafkaMsgs[i])
			if errInfra != nil {
				// error de infraestructura (TB caído) → no notificar, dejar que procesarConRetry reintente
				log.Printf("ERROR [GestorTransferencias.ProcesarLote]: Error de infraestructura en validarReversion: %v", errInfra)
				return errInfra
			}
		} else {
			estadoError = gt.validarTransferencia(t)
		}
		if estadoError != "" {
			log.Printf("[VALIDACIÓN] Transfer ID %s rechazada: %s", utils.Uint128AStringDecimal(t.ID), estadoError)
			fallidas = append(fallidas, models.NewTransferenciaNotificadaError(t, KafkaMsgs[i], estadoError))
		} else {
			paraEnviar = append(paraEnviar, t)
			kafkaMsgsValidos = append(kafkaMsgsValidos, KafkaMsgs[i])
		}
	}

	if len(fallidas) > 0 {
		log.Printf("VALIDACIÓN: %d de %d transfers rechazadas antes de TigerBeetle.", len(fallidas), len(Batch)+len(FallidasParseo))
	}

	var results []types.TransferEventResult
	if len(paraEnviar) > 0 {
		if persistence.ClienteTB == nil {
			return errors.New("Conexión a TigerBeetle no inicializada")
		}
		var err error
		results, err = persistence.ClienteTB.CreateTransfers(paraEnviar)
		if err != nil {
			log.Printf("ERROR [GestorTransferencias.ProcesarLote]: Error de comunicación con TigerBeetle al enviar batch de %d transfers: %v", len(paraEnviar), err)
			return err
		}

		if len(results) > 0 {
			log.Printf("RESPUESTA TB: Fallo en %d de %d transfers. Detalle de resultados:", len(results), len(paraEnviar))
			for _, result := range results {
				if int(result.Index) < len(paraEnviar) {
					idTransferencia := utils.Uint128AStringDecimal(paraEnviar[result.Index].ID)
					log.Printf(" -> Transfer ID %s, Resultado: %s", idTransferencia, result.Result.String())
				} else {
					log.Printf(" -> Índice de resultado %d fuera de rango (tamaño batch %d)", result.Index, len(paraEnviar))
				}
			}
		} else {
			log.Printf("RESPUESTA TB: Batch de %d transfers procesado exitosamente.", len(paraEnviar))
		}
	}

	// Notificar todo: resultados de TB + rechazadas
	if err := gt.notificador.NotificarTransferencias(paraEnviar, kafkaMsgsValidos, results, fallidas); err != nil {
		log.Printf("ERROR [GestorTransferencias.ProcesarLote]: Falló la notificación del Webhook: %v", err)
		return err
	}

	return nil
}

// Valida reglas de negocio sobre una transferencia antes de enviarla a TigerBeetle.
// Retorna "" si la transferencia es válida, o un string con el código de error.
func (gt *GestorTransferencias) validarTransferencia(t types.Transfer) string {
	monto := binary.LittleEndian.Uint64(t.Amount[:8])
	paramMax := &models.Parametros{Parametro: "MONTOMAXTRANSFER"}
	if _, err := paramMax.Dame(); err == nil {
		if max, err := strconv.ParseUint(paramMax.Valor, 10, 64); err == nil && monto > max {
			return "El monto excede el máximo permitido por transferencia"
		}
	}
	paramMin := &models.Parametros{Parametro: "MONTOMINTRANSFER"}
	if _, err := paramMin.Dame(); err == nil {
		if min, err := strconv.ParseUint(paramMin.Valor, 10, 64); err == nil && monto < min {
			return "El monto es inferior al mínimo permitido por transferencia"
		}
	}

	moneda := &models.Monedas{IdMoneda: int(t.Ledger)}
	if _, err := moneda.Dame(); err != nil {
		return "La moneda no existe o no está activa"
	}
	if moneda.Estado != "A" {
		return "La moneda no existe o no está activa"
	}

	return ""
}

// preValidarCuentas verifica, en una única llamada batch a TigerBeetle, que:
//   - la cuenta débito y la cuenta crédito existen,
//   - la cuenta débito no está cerrada (flag Closed),
//   - si la cuenta débito tiene el flag DebitsMustNotExceedCredits, el saldo
//     disponible (descontando los débitos virtuales ya aprobados en este batch)
//     es suficiente para cubrir el monto.
//
// Retorna (slice, nil): slice del mismo largo que batch ("" = válida, otro valor = error de negocio).
// Retorna (nil, error): error de infraestructura (TB caído) que debe reintentarse, no notificarse.
func (gt *GestorTransferencias) preValidarCuentas(batch []types.Transfer) ([]string, error) {
	errores := make([]string, len(batch))

	if len(batch) == 0 {
		return errores, nil
	}

	//IDs únicos de cuentas débito y crédito del batch
	idsSet := make(map[types.Uint128]struct{})
	for _, t := range batch {
		idsSet[t.DebitAccountID] = struct{}{}
		idsSet[t.CreditAccountID] = struct{}{}
	}
	ids := make([]types.Uint128, 0, len(idsSet))
	for id := range idsSet {
		ids = append(ids, id)
	}

	accounts, err := persistence.ClienteTB.LookupAccounts(ids)
	if err != nil {
		// error de infraestructura: el caller debe reintentar, no notificar como error de negocio
		return nil, err
	}

	// mapeo accountID → Account para lookup
	mapaAccounts := make(map[types.Uint128]types.Account, len(accounts))
	for _, a := range accounts {
		mapaAccounts[a.ID] = a
	}

	flagCerrada := types.AccountFlags{Closed: true}.ToUint16()
	flagDebitsMustNotExceedCredits := types.AccountFlags{DebitsMustNotExceedCredits: true}.ToUint16()

	// Acumulador de débitos virtuales aprobados en este batch por cuenta
	debitosVirtuales := make(map[types.Uint128]uint64)

	for i, t := range batch {
		debitAccount, existeDebit := mapaAccounts[t.DebitAccountID]
		if !existeDebit {
			errores[i] = "Cuenta no encontrada"
			continue
		}
		creditAccount, existeCredit := mapaAccounts[t.CreditAccountID]
		if !existeCredit {
			errores[i] = "Cuenta no encontrada"
			continue
		}
		if (debitAccount.Flags & flagCerrada) != 0 {
			errores[i] = "La cuenta está cerrada"
			continue
		}
		if (creditAccount.Flags & flagCerrada) != 0 {
			errores[i] = "La cuenta está cerrada"
			continue
		}

		if (debitAccount.Flags & flagDebitsMustNotExceedCredits) != 0 {
			// Leer lower 64 bits en LE
			creditsPosted := binary.LittleEndian.Uint64(debitAccount.CreditsPosted[:8])
			debitsPosted := binary.LittleEndian.Uint64(debitAccount.DebitsPosted[:8])
			monto := binary.LittleEndian.Uint64(t.Amount[:8])
			acumulado := debitosVirtuales[t.DebitAccountID]

			// balance disponible real menos lo comprometido en este batch
			balance := creditsPosted - debitsPosted
			if balance < acumulado+monto {
				errores[i] = "Saldo insuficiente en cuenta"
				continue
			}
			debitosVirtuales[t.DebitAccountID] += monto
		}
	}

	return errores, nil
}

// Valida que la transferencia a revertir sea la última de la cuenta.
// Retorna ("mensaje", nil) para errores de negocio, ("", error) para errores de infraestructura.
func (gt *GestorTransferencias) validarReversion(t types.Transfer, kafkaMsg models.KafkaTransferencias) (string, error) {
	idCuentaStr := utils.ConcatenarIDString(uint64(kafkaMsg.IdMoneda), kafkaMsg.IdUsuarioFinal)
	idCuenta, err := utils.ParsearUint128(idCuentaStr)
	if err != nil {
		return "No se pudo construir ID de cuenta usuario para validar reversión", nil
	}

	// Obtener la última transfer de la cuenta
	// Flags: bit 0 = Debits, bit 1 = Credits, bit 2 = Reversed → 0b111 = 7
	filtro := types.AccountFilter{
		AccountID: idCuenta,
		Limit:     1,
		Flags:     7, // Debits + Credits + Reversed
	}

	transfers, err := persistence.ClienteTB.GetAccountTransfers(filtro)
	if err != nil {
		// error de infraestructura: el caller debe reintentar, no notificar como error de negocio
		return "", err
	}
	if len(transfers) == 0 {
		return "No se encontraron transferencias para esta cuenta", nil
	}

	ultimaTransfer := transfers[0]

	if ultimaTransfer.Code == models.CodigoTransferenciaReversion {
		return "No se puede revertir una reversión", nil
	}

	if ultimaTransfer.ID != t.UserData128 {
		return "Solo se puede revertir la última transferencia de la cuenta", nil
	}

	return "", nil
}

// retorna false si el monto de la transfer está fuera del rango [montoMin, montoMax].
// Un valor 0 en alguno de los límites desactiva ese extremo del rango.
func pasaFiltroMonto(t types.Transfer, montoMin, montoMax uint64) bool {
	if montoMin == 0 && montoMax == 0 {
		return true
	}
	monto := binary.LittleEndian.Uint64(t.Amount[:8])
	if montoMin != 0 && monto < montoMin {
		return false
	}
	if montoMax != 0 && monto > montoMax {
		return false
	}
	return true
}
