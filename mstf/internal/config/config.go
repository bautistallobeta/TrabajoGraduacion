package config

import (
	"log"
	"os"
	"strconv"
	"strings"

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
	// MySQL
	MySQLHost     string
	MySQLPort     int
	MySQLUser     string
	MySQLPassword string
	MySQLDatabase string
}

func Load() Config {
	// Carga de .env. Si no existe (ej: producción con vars de entorno del sistema), se continúa sin error.
	if err := godotenv.Load(); err != nil {
		log.Printf("Archivo .env no encontrado, se usarán variables de entorno del sistema: %v", err)
	}

	cfg := Config{}

	cfg.Puerto = getEnvInt("PORT", 8080)

	// TB
	cfg.DireccionesTigerBeetle = []string{requireEnv("TB_ADDRESSES")}

	// Webhook
	cfg.URLWebhook = getEnv("WEBHOOK_URL", "")

	// Kafka
	cfg.BrokersKafka = strings.Split(requireEnv("KAFKA_BROKERS"), ",")
	cfg.TopicKafka = requireEnv("KAFKA_TOPIC_TRANSFERS")
	cfg.GroupIDKafka = requireEnv("KAFKA_GROUP_ID")
	// MySQL
	cfg.MySQLHost = requireEnv("MYSQL_HOST")
	cfg.MySQLPort = getEnvInt("MYSQL_PORT", 3306)
	cfg.MySQLUser = requireEnv("MYSQL_USER")
	cfg.MySQLPassword = requireEnv("MYSQL_PASSWORD")
	cfg.MySQLDatabase = requireEnv("MYSQL_DATABASE")

	return cfg
}

// requireEnv lee una variable de entorno requerida. Si no está definida, detiene el proceso.
func requireEnv(key string) string {
	value, exists := os.LookupEnv(key)
	if !exists || value == "" {
		log.Fatalf("FATAL: variable de entorno requerida no configurada: %s", key)
	}
	return value
}

func getEnv(key string, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return defaultValue
}
