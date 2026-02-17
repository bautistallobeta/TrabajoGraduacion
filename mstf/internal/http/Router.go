package http

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"MSTransaccionesFinancieras/internal/controllers"
	"MSTransaccionesFinancieras/internal/gestores"
	httpMiddleware "MSTransaccionesFinancieras/internal/http/middlewares"
	"MSTransaccionesFinancieras/internal/infra/kafkamstf"
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
	gestorCuentas := gestores.NewGestorCuentas()
	gestorTransferencias := gestores.NewGestorTransferencias(notificador)
	cuentasControlador := controllers.NewCuentasControlador(gestorCuentas)
	transferenciasControlador := controllers.NewTransferenciasControlador(gestorTransferencias, productor)
	gestorUsuarios := gestores.NewGestorUsuarios()
	usuariosControlador := controllers.NewUsuariosControlador(gestorUsuarios)
	paramControlador := controllers.NewParametrosControlador()
	gestorMonedas := gestores.NewGestorMonedas()
	monedasControlador := controllers.NewMonedasControlador(gestorMonedas, gestorCuentas)

	// Endpoint de prueba
	router.GET("/ping", mainControlador.Ping)

	// Cuentas
	router.GET("/cuentas/:idusuariofinal/:idmoneda/historial", cuentasControlador.DameHistorial)
	router.GET("/cuentas/:idusuariofinal/:idmoneda", cuentasControlador.Dame)
	router.POST("/cuentas", cuentasControlador.Crear)
	router.GET("/cuentas", cuentasControlador.Buscar)

	//Transferencias
	router.GET("/transferencias/:idtransferencia", transferenciasControlador.Dame)
	router.POST("/transferencias", transferenciasControlador.Crear)

	// Usuarios
	router.GET("/usuarios/:idusuario", usuariosControlador.Dame)
	router.GET("/usuarios", usuariosControlador.Buscar)
	router.POST("/usuarios", usuariosControlador.Crear)
	router.POST("/usuarios/login", usuariosControlador.Login)
	router.PUT("/usuarios/activar/:idusuario", usuariosControlador.Activar)
	router.PUT("/usuarios/desactivar/:idusuario", usuariosControlador.Desactivar)
	router.PUT("/usuarios/confirmar-cuenta/:idusuario", usuariosControlador.ConfirmarUsuario)
	router.PUT("/usuarios/password/modificar", usuariosControlador.ModificarPassword)
	router.PUT("/usuarios/password/reestablecer", usuariosControlador.ReestablecerPassword)
	router.DELETE("/usuarios/:idusuario", usuariosControlador.Borrar)

	// Parámetros
	router.GET("/parametros/:parametro", paramControlador.Dame)
	router.GET("/parametros", paramControlador.Buscar)
	router.PUT("/parametros/:parametro", paramControlador.Modificar)

	// Monedas
	router.GET("/monedas/:idmoneda", monedasControlador.Dame)
	router.POST("/monedas", monedasControlador.Crear)
	router.DELETE("/monedas/:idmoneda", monedasControlador.Borrar)
	router.GET("/monedas", monedasControlador.Listar)
	router.PUT("/monedas/:idmoneda/desactivar", monedasControlador.Desactivar)
	router.PUT("/monedas/:idmoneda/activar", monedasControlador.Activar)

}
