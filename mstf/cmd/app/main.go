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
	"MSTransaccionesFinancieras/internal/models"
	"MSTransaccionesFinancieras/internal/utils"

	"github.com/tigerbeetle/tigerbeetle-go/pkg/types"
)

// Inicializa las cuentas empresa en TigerBeetle para cada moneda activa,
// y recupera monedas que quedaron en estado P por una caída del servicio.
func inicializarCuentasEmpresa() error {
	log.Println("Inicializando cuentas empresa...")

	// TODO: eliminar hardcodeo de token
	const tokenAdmin = "cf904666e02a79cfd50b074ab3c360c0"

	gm := gestores.NewGestorMonedas()
	monedas, err := gm.Listar("N")
	if err != nil {
		log.Printf("ERROR [inicializarCuentasEmpresa]: No se pudieron listar monedas: %v", err)
		return err
	}
	if len(monedas) == 0 {
		log.Println("No hay monedas activas o pendientes. No se inicializan cuentas empresa.")
		return nil
	}

	type monedaInfo struct {
		idMoneda  int
		fechaAlta string
		estado    string
	}
	// array de IDs de cuentas empresa (única consulta a TB)
	var ids []types.Uint128
	infoMap := make(map[types.Uint128]monedaInfo)

	for _, m := range monedas {
		if m.IdCuentaEmpresa == "" {
			log.Printf("ADVERTENCIA: Moneda %d sin IdCuentaEmpresa, omitiendo", m.IdMoneda)
			continue
		}
		tbId, err := utils.ParsearUint128(m.IdCuentaEmpresa)
		if err != nil {
			log.Printf("ADVERTENCIA: IdCuentaEmpresa '%s' inválido para moneda %d, omitiendo", m.IdCuentaEmpresa, m.IdMoneda)
			continue
		}
		ids = append(ids, tbId)
		infoMap[tbId] = monedaInfo{
			idMoneda:  m.IdMoneda,
			fechaAlta: m.FechaAlta.Format("2006-01-02"),
			estado:    m.Estado,
		}
	}

	if len(ids) == 0 {
		log.Println("No hay cuentas empresa para verificar.")
		return nil
	}

	// Única consulta a TB para verificar qué cuentas ya existen
	cuentasExistentes, err := persistence.ClienteTB.LookupAccounts(ids)
	log.Printf("Cuentas empresa encontradas en TB: %d de %d", len(cuentasExistentes), len(ids))
	if err != nil {
		log.Printf("ERROR [inicializarCuentasEmpresa]: No se pudieron consultar cuentas empresa en TigerBeetle: %v", err)
		return err
	}

	existe := make(map[types.Uint128]bool)
	for _, c := range cuentasExistentes {
		existe[c.ID] = true
	}

	gc := gestores.NewGestorCuentas()

	// --- Monedas activas (A): crear las cuentas empresa que faltan en TB ---
	var faltantes []gestores.CuentaNueva
	for _, tbId := range ids {
		mi := infoMap[tbId]
		if mi.estado != "A" || existe[tbId] {
			continue
		}
		faltantes = append(faltantes, gestores.CuentaNueva{
			IdMoneda:                      uint32(mi.idMoneda),
			IdUsuarioFinal:                0,
			Fecha:                         mi.fechaAlta,
			DebitosNoDebenExcederCreditos: false,
		})
	}
	if len(faltantes) > 0 {
		log.Printf("Creando %d cuentas empresa faltantes para monedas activas...", len(faltantes))
		idsCreados, err := gc.CrearLote(faltantes)
		if err != nil {
			log.Printf("ERROR [inicializarCuentasEmpresa]: No se pudieron crear cuentas empresa: %v", err)
			return err
		}
		log.Printf("Cuentas empresa creadas exitosamente: %v", idsCreados)
	} else {
		log.Println("Todas las cuentas empresa de monedas activas ya existen en TigerBeetle.")
	}

	// --- Monedas pendientes (P): recuperar creación interrumpida por caída del servicio ---
	for _, tbId := range ids {
		mi := infoMap[tbId]
		if mi.estado != "P" {
			continue
		}
		log.Printf("Recuperando moneda pendiente %d...", mi.idMoneda)

		// Si la cuenta empresa no llegó a crearse en TB, crearla ahora
		if !existe[tbId] {
			log.Printf("Creando cuenta empresa faltante para moneda pendiente %d...", mi.idMoneda)
			_, _, err := gc.Crear(uint32(mi.idMoneda), 0, mi.fechaAlta, false)
			if err != nil {
				log.Printf("ERROR [inicializarCuentasEmpresa]: No se pudo crear cuenta empresa para moneda pendiente %d: %v", mi.idMoneda, err)
				return err
			}
			log.Printf("Cuenta empresa creada para moneda pendiente %d.", mi.idMoneda)
		}

		// Activar la moneda (P -> A)
		moneda := &models.Monedas{IdMoneda: mi.idMoneda}
		mensaje, err := moneda.Activar(tokenAdmin, "SISTEMA")
		if err != nil {
			log.Printf("ERROR [inicializarCuentasEmpresa]: No se pudo activar moneda pendiente %d: %v", mi.idMoneda, err)
			return err
		}
		if mensaje != "OK" {
			log.Printf("ERROR [inicializarCuentasEmpresa]: No se pudo activar moneda pendiente %d: %s", mi.idMoneda, mensaje)
			return fmt.Errorf("no se pudo activar moneda %d: %s", mi.idMoneda, mensaje)
		}
		log.Printf("Moneda pendiente %d activada exitosamente.", mi.idMoneda)
	}

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

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		log.Printf("Shutdown forzado: %v", err)
	}
	log.Println("El servidor dejó de funcionar.")
}
