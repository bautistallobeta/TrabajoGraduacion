package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/segmentio/kafka-go"
)

// Ctes
var (
	apiURL           = flag.String("api-url", "http://localhost:8080", "URL base del microservicio")
	apiKey           = flag.String("api-key", "CAMBIAR_ESTE_VALOR", "API Key para saltar auth de usuario")
	kafkaBroker      = flag.String("kafka-broker", "kafka:9092", "Dirección del broker de Kafka")
	kafkaTopic       = flag.String("kafka-topic", "transfers_pendientes", "Topic de Kafka")
	webhookPort      = flag.String("webhook-port", "9999", "Puerto para recibir webhooks")
	numMonedas       = flag.Int("monedas", 100, "Cantidad de monedas a crear")
	cuentasPorMoneda = flag.Int("cuentas", 1000, "Cantidad de cuentas a crear por cada moneda")
	totalTransfers   = flag.Int("transferencias", 10000000, "Cantidad total de transferencias a simular")
	dockerService    = flag.String("docker-service", "mstf", "Nombre del servicio del MS en docker-compose")
	autoDocker       = flag.Bool("auto-docker", false, "Si es true, el script hará stop/start del contenedor automáticamente")
)

type ReqMoneda struct {
	IdMoneda int `json:"IdMoneda"`
}

type ReqCuenta struct {
	IdUsuarioFinal uint64 `json:"IdUsuarioFinal"`
	IdMoneda       uint32 `json:"IdMoneda"`
	Fecha          string `json:"Fecha"`
}

// Estructura idem models.KafkaTransferencias
type KafkaMsg struct {
	IdTransferencia string  `json:"IdTransferencia"`
	IdUsuarioFinal  uint64  `json:"IdUsuarioFinal"`
	Monto           float64 `json:"Monto"`
	IdMoneda        uint32  `json:"IdMoneda"`
	Tipo            string  `json:"Tipo"`
	IdCategoria     uint64  `json:"IdCategoria"`
	Fecha           string  `json:"Fecha"`
}

type LoteWebhook struct {
	CantidadProcesada int `json:"CantidadProcesada"`
}

var (
	procesadas  atomic.Int64
	inicioTimer time.Time
	finTimer    time.Time
	primerLote  sync.Once
	doneChan    = make(chan bool)
)

func main() {
	flag.Parse()

	log.Println("=== INICIANDO TEST DE ESTRÉS ===")

	// FASE 1: SETUP
	log.Println("\n1) El ms DEBE estar encendido...")
	paresCuentas := setupDatos()

	if len(paresCuentas) == 0 {
		log.Fatal("No se pudieron crear cuentas válidas para la prueba.")
	}

	// FASE 2: DETENCIÓN Y LLENADO (LA REPRESA)
	log.Println("\n2) Llenado de kafka (apagar ms)...")
	manejarDocker("stop")

	llenarKafka(paresCuentas)

	// FASE 3: MEDICIÓN (APERTURA DE COMPUERTAS)
	log.Println("\n3) Levantar webhook (levantar de nuevo el ms)...")
	go iniciarServidorWebhook()

	// Pequeña pausa para asegurar que el server levantó
	time.Sleep(1 * time.Second)
	manejarDocker("start")

	log.Println("Esperando procesamiento de", *totalTransfers, "transferencias...")

	// Timeout de seguridad de 20 minutos
	select {
	case <-doneChan:
		calcularResultados()
	case <-time.After(20 * time.Minute):
		log.Println("TIMEOUT: Pasaron 20 minutos y no se completaron todas las transferencias.")
		calcularResultados()
	}
}

// --- FUNCIONES DE FASE 1 (API) ---

func setupDatos() []ReqCuenta {
	client := &http.Client{Timeout: 5 * time.Second}
	var cuentasValidas []ReqCuenta

	for i := 1; i <= *numMonedas; i++ {
		// Crear Moneda
		reqMoneda := ReqMoneda{IdMoneda: i}
		enviarPOST(client, *apiURL+"/monedas", reqMoneda)

		// Crear Cuentas
		for j := 1; j <= *cuentasPorMoneda; j++ {
			reqCuenta := ReqCuenta{
				IdUsuarioFinal: uint64(j),
				IdMoneda:       uint32(i),
				Fecha:          "2026-03-11",
			}
			status := enviarPOST(client, *apiURL+"/cuentas", reqCuenta)
			// Si devuelve 201 (Creado) o 200 (Ya existía), la consideramos válida para la prueba
			if status == http.StatusCreated || status == http.StatusOK {
				cuentasValidas = append(cuentasValidas, reqCuenta)
			}
		}
	}
	log.Printf("Llenado finalizado. %d cuentas listas para la prueba.", len(cuentasValidas))
	return cuentasValidas
}

func enviarPOST(client *http.Client, url string, payload interface{}) int {
	jsonData, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", *apiKey)

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error conectando a %s: %v", url, err)
		return 0
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("Advertencia en POST %s [%d]: %s", url, resp.StatusCode, string(body))
	}
	return resp.StatusCode
}

// FUNC AUX P DOCKER

func manejarDocker(accion string) {
	if !*autoDocker {
		fmt.Printf("\n>>> Ejecutar manualmente: `docker compose %s %s`\n", accion, *dockerService)
		fmt.Printf(">>> Tocar ENTER cuando lo esté...")
		fmt.Scanln()
		return
	}

	log.Printf("Ejecutando docker compose %s %s...", accion, *dockerService)
	cmd := exec.Command("docker", "compose", accion, *dockerService)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		log.Fatalf("Error ejecutando docker compose: %v", err)
	}
}

// FUNCIONES P LA SEGUNDA PARTE (KAFKA)
func llenarKafka(cuentas []ReqCuenta) {
	log.Println("Verificando conexión con Kafka en", *kafkaBroker, "...")
	conn, err := kafka.Dial("tcp", *kafkaBroker)
	if err != nil {
		log.Fatalf("FALLO FATAL: Kafka no responde: %v", err)
	}
	conn.Close()
	log.Println("Kafka responde correctamente. Preparando test...")

	writer := &kafka.Writer{
		Addr:                   kafka.TCP(*kafkaBroker),
		Topic:                  *kafkaTopic,
		AllowAutoTopicCreation: true,
	}
	defer writer.Close()

	var mensajes []kafka.Message

	for i := 1; i <= *totalTransfers; i++ {
		cuenta := cuentas[rand.Intn(len(cuentas))]

		msg := KafkaMsg{
			IdTransferencia: strconv.FormatInt(time.Now().UnixNano(), 10) + strconv.Itoa(i),
			IdUsuarioFinal:  cuenta.IdUsuarioFinal,
			Monto:           100.50,
			IdMoneda:        cuenta.IdMoneda,
			Tipo:            "I",
			IdCategoria:     1,
			Fecha:           "2026-03-11",
		}

		b, _ := json.Marshal(msg)
		mensajes = append(mensajes, kafka.Message{Value: b})

		// Publicar en lotes de 500
		if i%500 == 0 {
			var errWrite error
			//Reintenta hasta 5 veces si Kafka corta la conexión
			for intentos := 1; intentos <= 5; intentos++ {
				errWrite = writer.WriteMessages(context.Background(), mensajes...)
				if errWrite == nil {
					break
				}
				log.Printf("Aviso: Corte en Kafka (EOF) - Intento %d/5. Reintentando en 1s...", intentos)
				time.Sleep(1 * time.Second)
			}

			if errWrite != nil {
				log.Fatalf("\nError fatal escribiendo en Kafka tras 5 intentos: %v", errWrite)
			}

			mensajes = nil // Limpiar buffer
			fmt.Printf("\rEnviados: %d / %d", i, *totalTransfers)
		}
	}

	// Enviar el remanente si la cantidad no es múltiplo de 500
	if len(mensajes) > 0 {
		_ = writer.WriteMessages(context.Background(), mensajes...)
	}
	fmt.Printf("\rEnviados: %d / %d\n", *totalTransfers, *totalTransfers)
}

//FUNCIONES DE 3ERA PARTE (WEBHOOK )

func iniciarServidorWebhook() {
	http.HandleFunc("/webhook", func(w http.ResponseWriter, r *http.Request) {
		var lote LoteWebhook
		err := json.NewDecoder(r.Body).Decode(&lote)
		if err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		// Tomar T0 exacto cuando entra el primer request
		primerLote.Do(func() {
			inicioTimer = time.Now()
			log.Println("Recibido 1er batch. Inicia timer...")
		})

		actual := procesadas.Add(int64(lote.CantidadProcesada))

		w.WriteHeader(http.StatusOK)

		if actual >= int64(*totalTransfers) {
			finTimer = time.Now()
			select {
			case doneChan <- true:
			default:
			}
		}
	})

	log.Printf("Receptor Webhook escuchando en puerto %s...", *webhookPort)
	log.Fatal(http.ListenAndServe(":"+*webhookPort, nil))
}

func calcularResultados() {
	total := procesadas.Load()
	duracion := finTimer.Sub(inicioTimer)
	segundos := duracion.Seconds()
	tps := float64(total) / segundos

	fmt.Println("\n==================================================")
	fmt.Println("             RESULTADOS DEL TEST              ")
	fmt.Println("==================================================")
	fmt.Printf("Total esperado:       %d\n", *totalTransfers)
	fmt.Printf("Total procesado:      %d\n", total)
	fmt.Printf("Tiempo total:         %.3f segundos\n", segundos)
	fmt.Printf("THROUGHPUT OBTENIDO:  %.2f TPS (Transacciones/seg)\n", tps)
	fmt.Println("==================================================")
	os.Exit(0)
}
