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
	"MSTransaccionesFinancieras/internal/infra/persistence"
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
			mensajesLote, transferenciasLote, kafkaMsgsLote, fallidasParseo := c.armarLoteDesdeKafka(ctx)

			if len(transferenciasLote) == 0 && len(fallidasParseo) == 0 {
				continue
			}

			log.Printf("Lote de Kafka recibido. Procesando %d transferencias (%d fallidas en parseo).", len(transferenciasLote), len(fallidasParseo))

			// Procesar el lote con TigerBeetle y notificar
			if err := c.procesador.ProcesarLote(transferenciasLote, kafkaMsgsLote, fallidasParseo); err != nil {
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
func (c *Consumidor) armarLoteDesdeKafka(ctx context.Context) ([]kafka.Message, []types.Transfer, []models.KafkaTransferencias, []models.TransferenciaNotificada) {
	mensajesLote := make([]kafka.Message, 0, c.config.TamanoLoteKafka)
	transferenciasLote := make([]types.Transfer, 0, c.config.TamanoLoteKafka)
	kafkaMsgsLote := make([]models.KafkaTransferencias, 0, c.config.TamanoLoteKafka)
	var fallidasParseo []models.TransferenciaNotificada

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
			log.Printf("ERROR [Consumidor.armarLoteDesdeKafka]: Mensaje Kafka inválido (Offset: %d): %v. Se notificará en el webhook.", msg.Offset, err)
			fallidasParseo = append(fallidasParseo, models.NewTransferenciaNotificadaParseoError(kafkaMsg, err.Error()))
			mensajesLote = append(mensajesLote, msg)
			continue
		}
		mensajesLote = append(mensajesLote, msg)
		transferenciasLote = append(transferenciasLote, transfer)
		kafkaMsgsLote = append(kafkaMsgsLote, kafkaMsg)
	}
	return mensajesLote, transferenciasLote, kafkaMsgsLote, fallidasParseo
}

// Mensaje de Kafka a Transfer de TigerBeetle.
// Construye DebitAccountID y CreditAccountID a partir de IdUsuarioFinal, IdMoneda y Tipo (I/E).
func (c *Consumidor) parseKafkaMessage(msg kafka.Message) (types.Transfer, models.KafkaTransferencias, error) {
	var kafkaMsg models.KafkaTransferencias

	if err := json.Unmarshal(msg.Value, &kafkaMsg); err != nil {
		return types.Transfer{}, kafkaMsg, errors.New("Fallo al parsear JSON: " + err.Error())
	}
	if kafkaMsg.IdTransferencia == "" {
		return types.Transfer{}, kafkaMsg, errors.New("IdTransferencia está vacío")
	}
	if kafkaMsg.IdUsuarioFinal == 0 {
		return types.Transfer{}, kafkaMsg, errors.New("IdUsuarioFinal no puede ser cero")
	}
	if kafkaMsg.Tipo != "I" && kafkaMsg.Tipo != "E" && kafkaMsg.Tipo != "R" {
		return types.Transfer{}, kafkaMsg, errors.New("Tipo debe ser 'I' (ingreso), 'E' (egreso) o 'R' (reversión)")
	}
	if kafkaMsg.Tipo != "R" && kafkaMsg.Monto == 0 {
		return types.Transfer{}, kafkaMsg, errors.New("Monto no puede ser cero")
	}

	idTransferenciaCast, err := utils.ParsearUint128(kafkaMsg.IdTransferencia)
	if err != nil {
		return types.Transfer{}, kafkaMsg, errors.New("IdTransferencia formato incorrecto")
	}

	// Para Tipo="R", construir la transferencia de reversión a partir de la original
	if kafkaMsg.Tipo == "R" {
		return c.buildReversion(idTransferenciaCast, kafkaMsg)
	}

	// Flujo normal para I/E
	idCuentaUsuarioStr := utils.ConcatenarIDString(uint64(kafkaMsg.IdMoneda), kafkaMsg.IdUsuarioFinal)
	idCuentaUsuario, err := utils.ParsearUint128(idCuentaUsuarioStr)
	if err != nil {
		return types.Transfer{}, kafkaMsg, errors.New("No se pudo construir ID de cuenta usuario")
	}

	// Obtener IdCuentaEmpresa de la moneda
	// TODO: eliminar hardcodeo de token
	moneda := &models.Monedas{IdMoneda: int(kafkaMsg.IdMoneda)}
	if _, err := moneda.Dame("cf904666e02a79cfd50b074ab3c360c0"); err != nil {
		return types.Transfer{}, kafkaMsg, errors.New("La moneda no existe o no se encuentra activa")
	}
	if moneda.IdCuentaEmpresa == "" {
		return types.Transfer{}, kafkaMsg, errors.New("La moneda no existe o no se encuentra activa")
	}
	idCuentaEmpresa, err := utils.ParsearUint128(moneda.IdCuentaEmpresa)
	if err != nil {
		return types.Transfer{}, kafkaMsg, errors.New("IdCuentaEmpresa formato incorrecto: " + err.Error())
	}

	// Asignar débito/crédito según Tipo
	var debitAccountID, creditAccountID types.Uint128
	if kafkaMsg.Tipo == "E" {
		debitAccountID = idCuentaUsuario
		creditAccountID = idCuentaEmpresa
	} else {
		debitAccountID = idCuentaEmpresa
		creditAccountID = idCuentaUsuario
	}

	timeStampUint32, _ := utils.FechaAUserData32(kafkaMsg.Fecha)
	transferencia := types.Transfer{
		ID:              idTransferenciaCast,
		DebitAccountID:  debitAccountID,
		CreditAccountID: creditAccountID,
		Amount:          types.ToUint128(kafkaMsg.Monto),
		Ledger:          kafkaMsg.IdMoneda,
		Code:            models.CodigoTransferenciaNormal,
		UserData64:      kafkaMsg.IdCategoria,
		UserData32:      timeStampUint32,
	}
	return transferencia, kafkaMsg, nil
}

// construye una transferencia de reversión a partir de la original en TigerBeetle.
// (invierte las cuentas debit/credit, mismo monto, y guarda id original en userdata128)
func (c *Consumidor) buildReversion(idOriginal types.Uint128, kafkaMsg models.KafkaTransferencias) (types.Transfer, models.KafkaTransferencias, error) {
	if persistence.ClienteTB == nil {
		return types.Transfer{}, kafkaMsg, errors.New("Conexión a TigerBeetle no inicializada")
	}

	originals, err := persistence.ClienteTB.LookupTransfers([]types.Uint128{idOriginal})
	if err != nil {
		return types.Transfer{}, kafkaMsg, errors.New("Error al buscar transferencia original: " + err.Error())
	}
	if len(originals) == 0 {
		return types.Transfer{}, kafkaMsg, errors.New("No existe la transferencia a revertir")
	}

	original := originals[0]

	// bit 64 encendido sobre el ID original
	IdReversion := original.ID
	IdReversion[8] |= 0x01

	timeStampUint32, _ := utils.FechaAUserData32(kafkaMsg.Fecha)

	transferencia := types.Transfer{
		ID:              IdReversion,
		DebitAccountID:  original.CreditAccountID, // invertido
		CreditAccountID: original.DebitAccountID,  // invertido
		Amount:          original.Amount,
		Ledger:          original.Ledger,
		Code:            models.CodigoTransferenciaReversion,
		UserData128:     original.ID, // ref a la original
		UserData64:      original.UserData64,
		UserData32:      timeStampUint32,
	}
	return transferencia, kafkaMsg, nil
}
