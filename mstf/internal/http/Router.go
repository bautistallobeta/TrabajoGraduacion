package http

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"MSTransaccionesFinancieras/internal/config"
	"MSTransaccionesFinancieras/internal/controllers"
	"MSTransaccionesFinancieras/internal/gestores"
	httpMiddleware "MSTransaccionesFinancieras/internal/http/middlewares"
	"MSTransaccionesFinancieras/internal/infra/kafkamstf"
	"MSTransaccionesFinancieras/internal/infra/persistence"
	"MSTransaccionesFinancieras/internal/infra/webhook"
)

func InitRouter(cfg config.Config, notificador *webhook.Notificador) *echo.Echo {
	e := echo.New()
	e.HideBanner = true

	// Middlewares
	e.Use(
		middleware.Recover(),
		middleware.Logger(),
		middleware.CORS(),
		httpMiddleware.TokenAuth(),
	)

	initRoutes(e, notificador)

	return e
}

func initRoutes(router *echo.Echo, notificador *webhook.Notificador) {
	// Inicializac de controladores
	mainControlador := controllers.NewMainControlador()
	gestorCuentas := gestores.NewGestorCuentas(persistence.ClienteTB)
	gestorTransferencias := gestores.NewGestorTransferencias(persistence.ClienteTB, notificador)
	productorKafka, _ := kafkamstf.InitProductor(config.Load())
	cuentasControlador := controllers.NewCuentasControlador(gestorCuentas)
	transferenciasControlador := controllers.NewTransferenciasControlador(gestorTransferencias, productorKafka)
	gestorUsuarios := gestores.NewGestorUsuarios(persistence.ClienteMySQL)
	usuariosControlador := controllers.NewUsuariosControlador(gestorUsuarios)

	// Endpoint de prueba
	router.GET("/hola", mainControlador.Hola)

	// Cuentas
	router.GET("/cuentas/:id_cuenta/historial", cuentasControlador.DameHistorialCuenta)
	router.GET("/cuentas/:id_cuenta", cuentasControlador.DameCuenta)
	router.POST("/cuentas", cuentasControlador.CrearCuenta)
	router.GET("/cuentas", cuentasControlador.BuscarCuentas)

	//Transferencias
	router.GET("/transferencias/:id_transferencia", transferenciasControlador.DameTransferencia)
	router.POST("/transferencias", transferenciasControlador.CrearTransferencia)

	// Usuarios
	router.POST("/usuarios", usuariosControlador.Crear)

}
