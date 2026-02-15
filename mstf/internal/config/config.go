package config

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Puerto                 int
	DireccionesTigerBeetle []string
	URLWebhook             string
	// Kafka
	BrokersKafka []string
	TopicKafka   string
	GroupIDKafka string
	// Batching
	TamanoLoteKafka  int
	TimeoutLoteKafka time.Duration
	// MySQL
	MySQLHost     string
	MySQLPort     int
	MySQLUser     string
	MySQLPassword string
	MySQLDatabase string
}

func Load() Config {
	//TODO: corregir los valores por defecto, que solo tome variables de .env
	// Carga de .env (TODO: verificar el manejo de este error - diciembre 2025: no cargó el .env pero no falló
	err := godotenv.Load()
	if err != nil {
		log.Printf("No se pudo cargar el archivo .env: %v", err)
	}

	cfg := Config{}

	cfg.Puerto = mustGetEnvInt("PORT", 8080)

	// TB
	tbAddr := mustGetEnv("TB_ADDRESSES", "3000")
	cfg.DireccionesTigerBeetle = []string{tbAddr}

	// Webhook
	cfg.URLWebhook = mustGetEnv("WEBHOOK_URL", "")

	//  Kafka
	kafkaBrokerStr := mustGetEnv("KAFKA_BROKERS", "")
	cfg.BrokersKafka = strings.Split(kafkaBrokerStr, ",")
	cfg.TopicKafka = mustGetEnv("KAFKA_TOPIC", "")
	cfg.GroupIDKafka = mustGetEnv("KAFKA_GROUP_ID", "")
	cfg.TamanoLoteKafka = mustGetEnvInt("KAFKA_BATCH_SIZE", 1000)
	cfg.TimeoutLoteKafka = time.Duration(mustGetEnvInt("KAFKA_BATCH_TIMEOUT_MS", 20000)) * time.Millisecond

	// MySQL
	cfg.MySQLHost = mustGetEnv("MYSQL_HOST", "")
	cfg.MySQLPort = mustGetEnvInt("MYSQL_PORT", 3306)
	cfg.MySQLUser = mustGetEnv("MYSQL_USER", "")
	cfg.MySQLPassword = mustGetEnv("MYSQL_PASSWORD", "")
	cfg.MySQLDatabase = mustGetEnv("MYSQL_DATABASE", "")

	return cfg
}

func mustGetEnv(key string, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func mustGetEnvInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return defaultValue
}
