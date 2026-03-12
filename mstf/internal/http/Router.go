package http

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"MSTransaccionesFinancieras/internal/controllers"
	"MSTransaccionesFinancieras/internal/gestores"
	httpMiddleware "MSTransaccionesFinancieras/internal/http/middlewares"
	"MSTransaccionesFinancieras/internal/infra/kafkamstf"
)

func InitRouter(productor *kafkamstf.ProductorKafka) *echo.Echo {
	e := echo.New()
	e.HideBanner = true

	// Middlewares
	e.Use(
		middleware.Recover(),
		middleware.Logger(),
		middleware.CORS(),
		httpMiddleware.AutenticacionDual(func(c echo.Context) bool {
			path := c.Request().URL.Path
			// confirmar-cuenta SÍ usa token de sesión Estado=P; el SP valida internamente.
			return path == "/ping" || path == "/usuarios/login" ||
				path == "/usuarios/confirmar-cuenta"
		}),
	)

	initRoutes(e, productor)

	return e
}

func initRoutes(router *echo.Echo, productor *kafkamstf.ProductorKafka) {
	// Inicializac de controladores
	mainControlador := controllers.NewMainControlador()
	gestorCuentas := gestores.NewGestorCuentas()
	gestorTransferencias := gestores.NewGestorTransferencias()
	cuentasControlador := controllers.NewCuentasControlador(gestorCuentas, gestorTransferencias)
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
	router.GET("/cuentas/:idusuariofinal/:idmoneda/transferencias", cuentasControlador.DameTransferencias)
	router.GET("/cuentas/:idusuariofinal/:idmoneda", cuentasControlador.Dame)
	router.POST("/cuentas", cuentasControlador.Crear)
	router.GET("/cuentas", cuentasControlador.Buscar)
	router.PUT("/cuentas/:idusuariofinal/:idmoneda/desactivar", cuentasControlador.Desactivar)
	router.PUT("/cuentas/:idusuariofinal/:idmoneda/activar", cuentasControlador.Activar)

	//Transferencias
	router.GET("/transferencias/:idtransferencia", transferenciasControlador.Dame)
	router.GET("/transferencias", transferenciasControlador.Buscar)
	router.POST("/transferencias", transferenciasControlador.Crear)

	// Usuarios
	router.GET("/usuarios/:idusuario", usuariosControlador.Dame)
	router.GET("/usuarios", usuariosControlador.Buscar)
	router.POST("/usuarios", usuariosControlador.Crear)
	router.POST("/usuarios/login", usuariosControlador.Login)
	router.POST("/usuarios/logout", usuariosControlador.Logout)
	router.PUT("/usuarios/activar/:idusuario", usuariosControlador.Activar)
	router.PUT("/usuarios/desactivar/:idusuario", usuariosControlador.Desactivar)
	router.PUT("/usuarios/confirmar-cuenta", usuariosControlador.ConfirmarUsuario)
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
