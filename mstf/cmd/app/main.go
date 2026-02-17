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
	"MSTransaccionesFinancieras/internal/utils"

	"github.com/tigerbeetle/tigerbeetle-go/pkg/types"
)

// Inicializa las cuentas empresa en TigerBeetle para cada moneda activa.
// Lee las monedas activas de MySQL, verifica cuáles ya existen en TB (un solo lookup),
// y crea las faltantes en un solo llamado a TB.
func inicializarCuentasEmpresa() error {
	log.Println("Inicializando cuentas empresa...")

	// Listar monedas activas desde MySQL
	// TODO: eliminar hardcodeo de token
	gm := gestores.NewGestorMonedas()
	monedas, err := gm.Listar("cf904666e02a79cfd50b074ab3c360c0", "N")
	if err != nil {
		log.Printf("ERROR [inicializarCuentasEmpresa]: No se pudieron listar monedas activas: %v", err)
		return err
	}
	if len(monedas) == 0 {
		log.Println("No hay monedas activas. No se inicializan cuentas empresa.")
		return nil
	}

	// Armar array de IDs de cuentas empresa para consultar a TB en un solo llamado
	type monedaPendiente struct {
		idMoneda  int
		fechaAlta string
	}
	var ids []types.Uint128
	pendientes := make(map[types.Uint128]monedaPendiente)

	for _, m := range monedas {
		if m.IdCuentaEmpresa == "" {
			log.Printf("ADVERTENCIA: Moneda %d activa sin IdCuentaEmpresa, omitiendo", m.IdMoneda)
			continue
		}
		tbId, err := utils.ParsearUint128(m.IdCuentaEmpresa)
		if err != nil {
			log.Printf("ADVERTENCIA: IdCuentaEmpresa '%s' inválido para moneda %d, omitiendo", m.IdCuentaEmpresa, m.IdMoneda)
			continue
		}
		ids = append(ids, tbId)
		pendientes[tbId] = monedaPendiente{
			idMoneda:  m.IdMoneda,
			fechaAlta: m.FechaAlta.Format("2006-01-02"),
		}
	}

	if len(ids) == 0 {
		log.Println("No hay cuentas empresa para verificar.")
		return nil
	}

	// Consultar TB en un solo llamado
	cuentasExistentes, err := persistence.ClienteTB.LookupAccounts(ids)
	log.Printf("Cuentas empresa encontradas en TB: %d de %d", len(cuentasExistentes), len(ids))
	log.Printf("IDs de cuentas empresa pendientes de creación: %v", ids)

	if err != nil {
		log.Printf("ERROR [inicializarCuentasEmpresa]: No se pudieron consultar cuentas empresa en TigerBeetle: %v", err)
		return err
	}

	// Marcar las que ya existen
	existe := make(map[types.Uint128]bool)
	for _, c := range cuentasExistentes {
		existe[c.ID] = true
	}

	// Armar lote de cuentas faltantes
	var faltantes []gestores.CuentaNueva
	for _, tbId := range ids {
		if existe[tbId] {
			continue
		}
		mp := pendientes[tbId]
		faltantes = append(faltantes, gestores.CuentaNueva{
			IdMoneda:                      uint32(mp.idMoneda),
			IdUsuarioFinal:                0,
			FechaAlta:                     mp.fechaAlta,
			DebitosNoDebenExcederCreditos: false,
		})
	}

	if len(faltantes) == 0 {
		log.Println("Todas las cuentas empresa ya existen en TigerBeetle.")
		return nil
	}

	// Crear las faltantes en un solo llamado
	log.Printf("Creando %d cuentas empresa faltantes...", len(faltantes))
	gc := gestores.NewGestorCuentas()
	idsCreados, err := gc.CrearLote(faltantes)
	if err != nil {
		log.Printf("ERROR [inicializarCuentasEmpresa]: No se pudieron crear cuentas empresa: %v", err)
		return err
	}
	log.Printf("Cuentas empresa creadas exitosamente: %v", idsCreados)
	return nil
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

	// Inicializar cuentas empresa para cada moneda activa
	err := inicializarCuentasEmpresa()
	if err != nil {
		log.Fatalf("FATAL: No se pudo inicializar cuentas empresa: %v", err)
	}

	// Notificador Webhook
	notificador := webhook.NewNotificador(cfg)

	// Gestor de Transferencias
	gestorTransferencias := gestores.NewGestorTransferencias(notificador)

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
