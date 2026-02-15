package gestores

import (
	"MSTransaccionesFinancieras/internal/infra/webhook"
	"MSTransaccionesFinancieras/internal/utils"
	"errors"
	"log"

	tigerbeetle "github.com/tigerbeetle/tigerbeetle-go"
	"github.com/tigerbeetle/tigerbeetle-go/pkg/types"
)

type GestorTransferencias struct {
	tbClient    tigerbeetle.Client
	notificador *webhook.Notificador
}

func NewGestorTransferencias(tbClient tigerbeetle.Client, notificador *webhook.Notificador) *GestorTransferencias {
	return &GestorTransferencias{
		tbClient:    tbClient,
		notificador: notificador,
	}
}

// Lógica de negocio que se ejecuta al recibir un lote del consumidor. TODO: agregar comentario completo
func (gt *GestorTransferencias) ProcesarLote(batch []types.Transfer) error {
	if gt.tbClient == nil {
		return errors.New("Conexión a TigerBeetle no inicializada")
	}

	results, err := gt.tbClient.CreateTransfers(batch)
	if err != nil {
		log.Printf("ERROR [GestorTransferencias.ProcesarLote]: Error de comunicación con TigerBeetle al enviar batch de %d transfers: %v", len(batch), err)
		return err
	}

	if len(results) > 0 {
		log.Printf("RESPUESTA TB: Fallo en %d de %d transfers. Detalle de resultados:", len(results), len(batch))
		for _, result := range results {
			if int(result.Index) < len(batch) {
				idTransferencia := utils.Uint128AStringDecimal(batch[result.Index].ID)
				log.Printf(" -> Transfer ID %s, Resultado: %s", idTransferencia, result.Result.String())
			} else {
				log.Printf(" -> Índice de resultado %d fuera de rango (tamaño batch %d)", result.Index, len(batch))
			}
		}
	} else {
		log.Printf("RESPUESTA TB: Batch de %d transfers procesado exitosamente.", len(batch))
	}

	// Llamada síncrona al notificador
	if err := gt.notificador.NotificarTransferencias(batch, results); err != nil {
		log.Printf("ERROR [GestorTransferencias.ProcesarLote]: Falló la notificación del Webhook: %v", err)
		return err
	}

	return nil
}
