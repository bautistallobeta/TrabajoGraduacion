package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"MSTransaccionesFinancieras/internal/config"
	"MSTransaccionesFinancieras/internal/gestores"
	httpRouter "MSTransaccionesFinancieras/internal/http"
	"MSTransaccionesFinancieras/internal/infra/kafkamstf"
	"MSTransaccionesFinancieras/internal/infra/persistence"
	"MSTransaccionesFinancieras/internal/infra/webhook"

	"github.com/tigerbeetle/tigerbeetle-go/pkg/types"
)

// función (de momnto hardcodeada) para inicializar cuenta "empresa"
// TODO: inicializar una cuenta empresa por c/ledger en db relacional
func inicializarCuentaEmpresa() {
	log.Println("Inicializando cuenta empresa (ID=1)...")

	// ID fijo: 1
	idCuentaEmpresa := types.ToUint128(1)

	cuentas, err := persistence.ClienteTB.LookupAccounts([]types.Uint128{idCuentaEmpresa})
	if err != nil {
		log.Fatalf("FATAL: Error al verificar cuenta empresa: %v", err)
	}
	if len(cuentas) > 0 {
		return
	}

	cuentaEmpresa := types.Account{
		ID:         idCuentaEmpresa,
		Ledger:     1,   // Ledger 1 de momento)
		Code:       999, // Código hardcodeado para diferenciar cuenta empresa
		UserData64: 0,
		UserData32: 0,
		Flags: types.AccountFlags{
			DebitsMustNotExceedCredits: false, // permitir saldo negativo
			History:                    true,  // Habilita historial para auditoría (a definir - TODO)
		}.ToUint16(),
	}

	_, err = persistence.ClienteTB.CreateAccounts([]types.Account{cuentaEmpresa})
	if err != nil {
		log.Fatalf("FATAL: Error al crear cuenta empresa: %v", err)
	}

	log.Println("Cuenta empresa creada exitosamente (ID=1)")
}

func main() {
	cfg := config.Load()

	// Client de TigerBeetle
	if err := persistence.InitTBClient(cfg); err != nil {
		log.Fatalf("FATAL: No se pudo conectar a TigerBeetle: %v", err)
	}

	// Conexión a db MySQL
	if err := persistence.InitMySQLClient(cfg); err != nil {
		log.Fatalf("FATAL: No se pudo conectar a MySQL: %v", err)
	}

	// TODO: corregir creacipon de cuentas empresa - de momento hardcodeado
	inicializarCuentaEmpresa()

	// Notificador Webhook
	notificador := webhook.NewNotificador(cfg)

	// Gestor de Transferencias
	gestorTransferencias := gestores.NewGestorTransferencias(persistence.ClienteTB, notificador)

	// Consumidor Kafka
	consumidor := kafkamstf.NewConsumidor(cfg, gestorTransferencias)
	consumidor.Start()

	// Productor Kafka (unicamente para probar inserción de mensajes en la cola)
	// TODO: aclarar en docmentación que esto es solo para pruebas y para la demostración de la creación de transferancias
	productor, err := kafkamstf.InitProductor(cfg)
	if err != nil {
		log.Fatalf("FATAL: No se pudo conectar a Kafka (Productor): %v", err)
	}

	// Inicializar router HTTP
	e := httpRouter.InitRouter(notificador, productor)

	// Arranque del server
	go func() {
		addr := fmt.Sprintf(":%d", cfg.Puerto)
		log.Printf("HTTP escuchando en  %s", addr)
		if err := e.Start(addr); err != nil {
			log.Printf("Servidor parado: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Println("Apagando servidor...")

	// apagar consumer y producer kafka y cerrar conexiones a TB y MySQL
	consumidor.Close()
	productor.Close()
	persistence.CloseTBClient()
	persistence.CloseMySQLClient()

	// apagar server (con timeout)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		log.Printf("Shutdown forzado: %v", err)
	}
	log.Println("El servidor dejó de funcionar.")
}
