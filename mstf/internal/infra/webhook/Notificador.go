package webhook

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"MSTransaccionesFinancieras/internal/config"
	"MSTransaccionesFinancieras/internal/models"

	"github.com/tigerbeetle/tigerbeetle-go/pkg/types"
)

type Notificador struct {
	cfg config.Config
}

func NewNotificador(cfg config.Config) *Notificador {
	return &Notificador{cfg: cfg}
}

func (n *Notificador) NotificarTransferencias(transfers []types.Transfer, results []types.TransferEventResult) error {
	resultadosTransferenciaMap := make(map[uint32]types.TransferEventResult)
	for _, res := range results {
		resultadosTransferenciaMap[res.Index] = res
	}
	notificaciones := make([]models.TransferenciaNotificada, 0, len(transfers))
	for i, t := range transfers {
		var result types.TransferEventResult
		if res, exists := resultadosTransferenciaMap[uint32(i)]; exists {
			result = res
		} else {
			result.Result = types.TransferOK
		}

		notificacion := models.NewTransferenciaNotificada(t, result)
		notificaciones = append(notificaciones, notificacion)
	}
	payload := models.LoteNotificado{
		CantidadProcesada: len(notificaciones),
		Transferencias:    notificaciones,
	}

	return n.llamarWebhook(payload)
}

func (n *Notificador) llamarWebhook(payload models.LoteNotificado) error {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		log.Printf("ERROR [Notificador.llamarWebhook]: Fallo al serializar payload: %v", err)
		return err
	}
	urlWebhook := n.cfg.URLWebhook
	// TODO: si no está config la url, definir comportamiento, de momento solo se loguea la advertencia
	if urlWebhook == "" {
		log.Printf("ADVERTENCIA Notificador: URLWebhook no configurada. Simulación de envío exitoso:\n%s", string(jsonPayload))
		return nil
	}
	log.Printf("Notificador: Enviando POST Webhook a %s. Lote de %d transfers.", urlWebhook, payload.CantidadProcesada)
	client := http.Client{
		Timeout: 15 * time.Second,
	}
	resp, err := client.Post(urlWebhook, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		log.Printf("ERROR [Notificador.llamarWebhook]: Fallo al llamar Webhook a %s: %v", urlWebhook, err)
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		log.Printf("Notificador: Webhook enviado exitosamente. Status: %s", resp.Status)
		return nil
	}
	log.Printf("ERROR [Notificador.llamarWebhook]: Webhook respondió con un error. Status: %s", resp.Status)
	return errors.New("webhook devolvió status no exitoso: " + resp.Status)
}
