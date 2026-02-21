package kafkamstf

import (
	"MSTransaccionesFinancieras/internal/config"
	"MSTransaccionesFinancieras/internal/models"
	"context"
	"encoding/json"
	"log"

	"github.com/segmentio/kafka-go"
)

// EL PRODUCTOR SOLO EXISTE PARA PROBAR LA CREACIÓN DE TRANSFERENCIAS, NO ES PARTE DEL FLUJO NORMAL DEL MS
type ProductorKafka struct {
	writer *kafka.Writer
}

func InitProductor(cfg config.Config) (*ProductorKafka, error) {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(cfg.BrokersKafka...),
		Topic:        cfg.TopicKafka,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireAll,
		MaxAttempts:  3,
	}

	log.Println("Productor de Kafka conectado.")
	return &ProductorKafka{writer: writer}, nil
}

func (p *ProductorKafka) Close() {
	if p.writer != nil {
		if err := p.writer.Close(); err != nil {
			log.Printf("Error al cerrar Kafka writer (productor): %v", err)
		}
	}
}

func (p *ProductorKafka) PublicarTransferencia(ctx context.Context, msg models.KafkaTransferencias) error {
	// Serializar a json
	jsonValue, err := json.Marshal(msg)
	if err != nil {
		log.Printf("ERROR [Productor.PublicarTransferencia]: Fallo al serializar JSON: %v", err)
		return err
	}
	mensajeKafka := kafka.Message{
		Key:   []byte(msg.IdTransferencia),
		Value: jsonValue,
	}
	err = p.writer.WriteMessages(ctx, mensajeKafka)
	if err != nil {
		log.Printf("ERROR [Productor.PublicarTransferencia]: Fallo al escribir mensaje en Kafka: %v", err)
		return err
	}
	return nil
}
