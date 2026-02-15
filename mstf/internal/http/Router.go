package http

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"MSTransaccionesFinancieras/internal/controllers"
	"MSTransaccionesFinancieras/internal/gestores"
	httpMiddleware "MSTransaccionesFinancieras/internal/http/middlewares"
	"MSTransaccionesFinancieras/internal/infra/kafkamstf"
	"MSTransaccionesFinancieras/internal/infra/persistence"
	"MSTransaccionesFinancieras/internal/infra/webhook"
)

func InitRouter(notificador *webhook.Notificador, productor *kafkamstf.ProductorKafka) *echo.Echo {
	e := echo.New()
	e.HideBanner = true

	// Middlewares
	e.Use(
		middleware.Recover(),
		middleware.Logger(),
		middleware.CORS(),
		httpMiddleware.TokenAuth(),
	)

	initRoutes(e, notificador, productor)

	return e
}

func initRoutes(router *echo.Echo, notificador *webhook.Notificador, productor *kafkamstf.ProductorKafka) {
	// Inicializac de controladores
	mainControlador := controllers.NewMainControlador()
	gestorCuentas := gestores.NewGestorCuentas(persistence.ClienteTB)
	gestorTransferencias := gestores.NewGestorTransferencias(persistence.ClienteTB, notificador)
	cuentasControlador := controllers.NewCuentasControlador(gestorCuentas)
	transferenciasControlador := controllers.NewTransferenciasControlador(gestorTransferencias, productor)
	gestorUsuarios := gestores.NewGestorUsuarios(persistence.ClienteMySQL)
	usuariosControlador := controllers.NewUsuariosControlador(gestorUsuarios)
	paramControlador := controllers.NewParametrosControlador()
	gestorMonedas := gestores.NewGestorMonedas(persistence.ClienteMySQL)
	monedasControlador := controllers.NewMonedasControlador(gestorMonedas)

	// Endpoint de prueba
	router.GET("/ping", mainControlador.Ping)

	// Cuentas
	router.GET("/cuentas/:id_cuenta/historial", cuentasControlador.DameHistorial)
	router.GET("/cuentas/:id_cuenta", cuentasControlador.Dame)
	router.POST("/cuentas", cuentasControlador.Crear)
	router.GET("/cuentas", cuentasControlador.Buscar)

	//Transferencias
	router.GET("/transferencias/:id_transferencia", transferenciasControlador.Dame)
	router.POST("/transferencias", transferenciasControlador.Crear)

	// Usuarios
	router.GET("/usuarios/:id_usuario", usuariosControlador.Dame)
	router.GET("/usuarios", usuariosControlador.Buscar)
	router.POST("/usuarios", usuariosControlador.Crear)
	router.POST("/usuarios/login", usuariosControlador.Login)
	router.PUT("/usuarios/activar/:id_usuario", usuariosControlador.Activar)
	router.PUT("/usuarios/desactivar/:id_usuario", usuariosControlador.Desactivar)
	router.PUT("/usuarios/confirmar-cuenta/:id_usuario", usuariosControlador.ConfirmarUsuario)
	router.PUT("/usuarios/password/modificar", usuariosControlador.ModificarPassword)
	router.PUT("/usuarios/password/reestablecer", usuariosControlador.ReestablecerPassword)
	router.DELETE("/usuarios/:id_usuario", usuariosControlador.Borrar)

	// Parámetros
	router.GET("/parametros/:parametro", paramControlador.Dame)
	router.GET("/parametros", paramControlador.Buscar)
	router.PUT("/parametros/:parametro", paramControlador.Modificar)

	// Monedas
	router.GET("/monedas/:id_moneda", monedasControlador.Dame)
	router.POST("/monedas", monedasControlador.Crear)
	router.DELETE("/monedas/:id_moneda", monedasControlador.Borrar)
	router.GET("/monedas", monedasControlador.Listar)
	router.PUT("/monedas/:id_moneda/desactivar", monedasControlador.Desctivar)

}
