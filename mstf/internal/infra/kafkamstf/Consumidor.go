package kafkamstf

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"sync"

	"MSTransaccionesFinancieras/internal/config"
	"MSTransaccionesFinancieras/internal/gestores"
	"MSTransaccionesFinancieras/internal/models"
	"MSTransaccionesFinancieras/internal/utils"

	"github.com/segmentio/kafka-go"
	"github.com/tigerbeetle/tigerbeetle-go/pkg/types"
)

type Consumidor struct {
	reader     *kafka.Reader
	config     config.Config
	stopChan   chan struct{}
	wg         sync.WaitGroup
	procesador *gestores.GestorTransferencias
}

func NewConsumidor(cfg config.Config, procesador *gestores.GestorTransferencias) *Consumidor {
	lectorKafka := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  cfg.BrokersKafka,
		Topic:    cfg.TopicKafka,
		GroupID:  cfg.GroupIDKafka,
		MaxWait:  cfg.TimeoutLoteKafka,
		MinBytes: 1,
		MaxBytes: 10e6,
	})

	return &Consumidor{
		procesador: procesador,
		reader:     lectorKafka,
		config:     cfg,
		stopChan:   make(chan struct{}),
	}
}

// Start inicia el consumidor en una goroutine
func (c *Consumidor) Start() {
	c.wg.Add(1)
	go c.batchLoop()
	log.Printf("Consumidor iniciado. Kafka Topic: %s, GroupID: %s, Max Batch (App): %d, Max Wait (Kafka): %s",
		c.config.TopicKafka, c.config.GroupIDKafka, c.config.TamanoLoteKafka, c.config.TimeoutLoteKafka)
}

func (c *Consumidor) Close() {
	close(c.stopChan)
	c.wg.Wait()
	if err := c.reader.Close(); err != nil {
		log.Printf("Error al cerrar Kafka reader: %v", err)
	}
	log.Println("Consumidor detenido.")
}

// loop principal que procesa lotes de transferencias
func (c *Consumidor) batchLoop() {
	defer c.wg.Done()
	ctx := context.Background()

	for {
		select {
		case <-c.stopChan:
			log.Println("Deteniendo batchLoop...")
			return
		default:
			// Armar el lote desde Kafka
			mensajesLote, transferenciasLote := c.armarLoteDesdeKafka(ctx)

			if len(transferenciasLote) == 0 {
				continue
			}

			log.Printf("Lote de Kafka recibido. Procesando %d transferencias.", len(transferenciasLote))

			// Procesar el lote con TigerBeetle y notificar
			if err := c.procesador.ProcesarLote(transferenciasLote); err != nil {
				log.Printf("CRÍTICO [Consumidor.batchLoop]: Falló procesamiento del lote. NO se hará commit. El lote se reintentará: %v", err)
				continue
			}

			// Commit de offsets en Kafka (solo llega hasta acá si no falló la escritura en TB)
			log.Printf("Lote procesado exitosamente. Haciendo commit de %d offsets en Kafka.", len(mensajesLote))
			if err := c.reader.CommitMessages(ctx, mensajesLote...); err != nil {
				log.Printf("CRÍTICO [Consumidor.batchLoop]: No se pudo hacer commit de offsets: %v", err)
			}
		}
	}
}

// leer msj de kafka y armar lote de transferencias
func (c *Consumidor) armarLoteDesdeKafka(ctx context.Context) ([]kafka.Message, []types.Transfer) {
	mensajesLote := make([]kafka.Message, 0, c.config.TamanoLoteKafka)
	transferenciasLote := make([]types.Transfer, 0, c.config.TamanoLoteKafka)

	ctxLote, cancelarLote := context.WithTimeout(ctx, c.config.TimeoutLoteKafka)
	defer cancelarLote()

	for i := 0; i < c.config.TamanoLoteKafka; i++ {
		msg, err := c.reader.FetchMessage(ctxLote)
		if err != nil {
			if err == context.DeadlineExceeded || err == io.EOF {
				break
			}
			if err == context.Canceled && ctx.Err() == context.Canceled {
				break
			}
			log.Printf("ERROR [Consumidor.armarLoteDesdeKafka]: No se pudo fetch mensaje: %v", err)
			break
		}
		transfer, err := c.parseKafkaMessage(msg)
		if err != nil {
			log.Printf("ERROR [Consumidor.armarLoteDesdeKafka]: Mensaje Kafka inválido (Offset: %d): %v. Saltando mensaje", msg.Offset, err)
			if errCommit := c.reader.CommitMessages(ctx, msg); errCommit != nil {
				log.Printf("ERROR [Consumidor.armarLoteDesdeKafka]: No se pudo hacer commit del mensaje inválido: %v", errCommit)
			}
			continue
		}
		mensajesLote = append(mensajesLote, msg)
		transferenciasLote = append(transferenciasLote, transfer)
	}
	return mensajesLote, transferenciasLote
}

// Mensaje de Kafka a Transfer de TigerBeetle
func (c *Consumidor) parseKafkaMessage(msg kafka.Message) (types.Transfer, error) {
	var kafkaMsg models.KafkaTransferencias

	if err := json.Unmarshal(msg.Value, &kafkaMsg); err != nil {
		return types.Transfer{}, errors.New("fallo al parsear JSON: " + err.Error())
	}
	if kafkaMsg.IdTransferencia == "" || kafkaMsg.IdCuentaDebito == "" || kafkaMsg.IdCuentaCredito == "" {
		return types.Transfer{}, errors.New("IDs de transferencia o cuentas están vacíos")
	}
	if kafkaMsg.Monto == 0 {
		return types.Transfer{}, errors.New("monto no puede ser cero")
	}

	idTransferenciaCast, err := utils.ParsearUint128(kafkaMsg.IdTransferencia)
	if err != nil {
		return types.Transfer{}, errors.New("IdTransferencia formato incorrecto")
	}
	idDebitoCast, err := utils.ParsearUint128(kafkaMsg.IdCuentaDebito)
	if err != nil {
		return types.Transfer{}, errors.New("IdCuentaDebito formato incorrecto")
	}
	idCreditoCast, err := utils.ParsearUint128(kafkaMsg.IdCuentaCredito)
	if err != nil {
		return types.Transfer{}, errors.New("IdCuentaCredito formato incorrecto")
	}

	timeStamp, _ := utils.FechaAUserData128(kafkaMsg.Fecha)
	log.Printf("Fecha parseada a UserData128: %s -> %s", kafkaMsg.Fecha, timeStamp.String())
	transferencia := types.Transfer{
		ID:              idTransferenciaCast,
		DebitAccountID:  idDebitoCast,
		CreditAccountID: idCreditoCast,
		Amount:          types.ToUint128(kafkaMsg.Monto),
		Ledger:          kafkaMsg.Ledger,
		Code:            1, // Código hardcodeadeo (TODO: definir qué hacer con el code)
		UserData64:      kafkaMsg.IdCategoria,
		UserData128:     timeStamp,
	}
	return transferencia, nil
}
