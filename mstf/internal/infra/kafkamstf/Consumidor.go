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
			mensajesLote, transferenciasLote, kafkaMsgsLote := c.armarLoteDesdeKafka(ctx)

			if len(transferenciasLote) == 0 {
				continue
			}

			log.Printf("Lote de Kafka recibido. Procesando %d transferencias.", len(transferenciasLote))

			// Procesar el lote con TigerBeetle y notificar
			if err := c.procesador.ProcesarLote(transferenciasLote, kafkaMsgsLote); err != nil {
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
func (c *Consumidor) armarLoteDesdeKafka(ctx context.Context) ([]kafka.Message, []types.Transfer, []models.KafkaTransferencias) {
	mensajesLote := make([]kafka.Message, 0, c.config.TamanoLoteKafka)
	transferenciasLote := make([]types.Transfer, 0, c.config.TamanoLoteKafka)
	kafkaMsgsLote := make([]models.KafkaTransferencias, 0, c.config.TamanoLoteKafka)

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
		transfer, kafkaMsg, err := c.parseKafkaMessage(msg)
		if err != nil {
			log.Printf("ERROR [Consumidor.armarLoteDesdeKafka]: Mensaje Kafka inválido (Offset: %d): %v. Saltando mensaje", msg.Offset, err)
			if errCommit := c.reader.CommitMessages(ctx, msg); errCommit != nil {
				log.Printf("ERROR [Consumidor.armarLoteDesdeKafka]: No se pudo hacer commit del mensaje inválido: %v", errCommit)
			}
			continue
		}
		mensajesLote = append(mensajesLote, msg)
		transferenciasLote = append(transferenciasLote, transfer)
		kafkaMsgsLote = append(kafkaMsgsLote, kafkaMsg)
	}
	return mensajesLote, transferenciasLote, kafkaMsgsLote
}

// Mensaje de Kafka a Transfer de TigerBeetle.
// Construye DebitAccountID y CreditAccountID a partir de IdUsuarioFinal, IdMoneda y Tipo (I/E).
func (c *Consumidor) parseKafkaMessage(msg kafka.Message) (types.Transfer, models.KafkaTransferencias, error) {
	var kafkaMsg models.KafkaTransferencias

	if err := json.Unmarshal(msg.Value, &kafkaMsg); err != nil {
		return types.Transfer{}, kafkaMsg, errors.New("fallo al parsear JSON: " + err.Error())
	}
	if kafkaMsg.IdTransferencia == "" {
		return types.Transfer{}, kafkaMsg, errors.New("IdTransferencia está vacío")
	}
	if kafkaMsg.IdUsuarioFinal == 0 {
		return types.Transfer{}, kafkaMsg, errors.New("IdUsuarioFinal no puede ser cero")
	}
	if kafkaMsg.Monto == 0 {
		return types.Transfer{}, kafkaMsg, errors.New("monto no puede ser cero")
	}
	if kafkaMsg.Tipo != "I" && kafkaMsg.Tipo != "E" {
		return types.Transfer{}, kafkaMsg, errors.New("Tipo debe ser 'I' (ingreso) o 'E' (egreso)")
	}

	idTransferenciaCast, err := utils.ParsearUint128(kafkaMsg.IdTransferencia)
	if err != nil {
		return types.Transfer{}, kafkaMsg, errors.New("IdTransferencia formato incorrecto")
	}

	// Construir ID de cuenta usuario: concatenación de IdMoneda + IdUsuarioFinal
	idCuentaUsuarioStr := utils.ConcatenarIDString(uint64(kafkaMsg.IdMoneda), kafkaMsg.IdUsuarioFinal)
	idCuentaUsuario, err := utils.ParsearUint128(idCuentaUsuarioStr)
	if err != nil {
		return types.Transfer{}, kafkaMsg, errors.New("no se pudo construir ID de cuenta usuario")
	}

	// Obtener IdCuentaEmpresa de la moneda
	// TODO: eliminar hardcodeo de token
	moneda := &models.Monedas{IdMoneda: int(kafkaMsg.IdMoneda)}
	if _, err := moneda.Dame("cf904666e02a79cfd50b074ab3c360c0"); err != nil {
		return types.Transfer{}, kafkaMsg, errors.New("no se pudo obtener la moneda: " + err.Error())
	}
	if moneda.IdCuentaEmpresa == "" {
		return types.Transfer{}, kafkaMsg, errors.New("la moneda no tiene cuenta empresa configurada")
	}
	idCuentaEmpresa, err := utils.ParsearUint128(moneda.IdCuentaEmpresa)
	if err != nil {
		return types.Transfer{}, kafkaMsg, errors.New("IdCuentaEmpresa formato incorrecto: " + err.Error())
	}

	// Asignar débito/crédito según Tipo
	var debitAccountID, creditAccountID types.Uint128
	if kafkaMsg.Tipo == "E" {
		// Egreso: sale del usuario, entra a la empresa
		debitAccountID = idCuentaUsuario
		creditAccountID = idCuentaEmpresa
	} else {
		// Ingreso: sale de la empresa, entra al usuario
		debitAccountID = idCuentaEmpresa
		creditAccountID = idCuentaUsuario
	}

	timeStamp, _ := utils.FechaAUserData128(kafkaMsg.Fecha)
	transferencia := types.Transfer{
		ID:              idTransferenciaCast,
		DebitAccountID:  debitAccountID,
		CreditAccountID: creditAccountID,
		Amount:          types.ToUint128(kafkaMsg.Monto),
		Ledger:          kafkaMsg.IdMoneda,
		Code:            1, // Código hardcodeado (TODO: definir qué hacer con el code)
		UserData64:      kafkaMsg.IdCategoria,
		UserData128:     timeStamp,
	}
	return transferencia, kafkaMsg, nil
}
