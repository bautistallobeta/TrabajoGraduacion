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

// Procesa un lote de transferencias recibido del consumidor Kafka.
// Valida reglas de negocio antes de enviar a TigerBeetle.
// Las transferencias que fallan validación no van a TB pero sí se notifican con su error.
func (gt *GestorTransferencias) ProcesarLote(batch []types.Transfer, kafkaMsgs []models.KafkaTransferencias) error {
	if persistence.ClienteTB == nil {
		return errors.New("Conexión a TigerBeetle no inicializada")
	}

	// Validación previa: separar transferencias válidas de las que fallan reglas de negocio
	var paraEnviar []types.Transfer
	var kafkaMsgsValidos []models.KafkaTransferencias
	var fallidas []models.TransferenciaNotificada
	for i, t := range batch {
		if estadoError := gt.validarTransferencia(t); estadoError != "" {
			log.Printf("[VALIDACIÓN] Transfer ID %s rechazada: %s", utils.Uint128AStringDecimal(t.ID), estadoError)
			fallidas = append(fallidas, models.NewTransferenciaNotificadaError(t, kafkaMsgs[i], estadoError))
		} else {
			paraEnviar = append(paraEnviar, t)
			kafkaMsgsValidos = append(kafkaMsgsValidos, kafkaMsgs[i])
		}
	}

	if len(fallidas) > 0 {
		log.Printf("VALIDACIÓN: %d de %d transfers rechazadas antes de TigerBeetle.", len(fallidas), len(batch))
	}

	// Enviar a TigerBeetle solo las válidas
	var results []types.TransferEventResult
	if len(paraEnviar) > 0 {
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

	// Notificar todo: resultados de TB + rechazadas por validación previa
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

	// Verificación de monto máximo por transferencia (parámetro MONTOMAXTRANSFER)
	// TODO: eliminar hardcodeo de token
	paramMax := &models.Parametros{Parametro: "MONTOMAXTRANSFER"}
	if _, err := paramMax.Dame("cf904666e02a79cfd50b074ab3c360c0"); err == nil {
		if max, err := strconv.ParseUint(paramMax.Valor, 10, 64); err == nil && monto > max {
			return "El monto excede el máximo permitido por transferencia"
		}
	}

	// Verificación de monto mínimo por transferencia (parámetro MONTOMINTRANSFER)
	// TODO: eliminar hardcodeo de token
	paramMin := &models.Parametros{Parametro: "MONTOMINTRANSFER"}
	if _, err := paramMin.Dame("cf904666e02a79cfd50b074ab3c360c0"); err == nil {
		if min, err := strconv.ParseUint(paramMin.Valor, 10, 64); err == nil && monto < min {
			return "El monto es inferior al mínimo permitido por transferencia"
		}
	}

	// Verificación de moneda existente y activa
	// TODO: eliminar hardcodeo de token
	moneda := &models.Monedas{IdMoneda: int(t.Ledger)}
	if _, err := moneda.Dame("cf904666e02a79cfd50b074ab3c360c0"); err != nil {
		return "La moneda no existe o no está activa"
	}
	if moneda.Estado != "A" {
		return "La moneda no existe o no está activa"
	}

	return ""
}
