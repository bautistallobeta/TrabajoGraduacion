package persistence

import (
	"MSTransaccionesFinancieras/internal/config"
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

var ClienteMySQL *sql.DB

func InitMySQLClient(cfg config.Config) error {
	var err error

	// DSN: user:password@tcp(host:port)/database?parseTime=true
	sqlConfig := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
		cfg.MySQLUser,
		cfg.MySQLPassword,
		cfg.MySQLHost,
		cfg.MySQLPort,
		cfg.MySQLDatabase,
	)

	ClienteMySQL, err = sql.Open("mysql", sqlConfig)
	if err != nil {
		log.Printf("Error al crear cliente de MySQL: %v", err)
		return err
	}

	// Verificar conexión
	if err = ClienteMySQL.Ping(); err != nil {
		log.Printf("Error al conectar a MySQL: %v", err)
		return err
	}

	log.Printf("Conexión a MySQL establecida (%s:%d/%s)", cfg.MySQLHost, cfg.MySQLPort, cfg.MySQLDatabase)
	return nil
}

func CloseMySQLClient() {
	if ClienteMySQL != nil {
		ClienteMySQL.Close()
		log.Println("Conexión a MySQL cerrada")
	}
}
