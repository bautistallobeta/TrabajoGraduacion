package controllers

import (
	"MSTransaccionesFinancieras/internal/gestores"
	"MSTransaccionesFinancieras/internal/models"
	"MSTransaccionesFinancieras/internal/utils"
	"log"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

type MonedasControlador struct {
	Gestor        *gestores.GestorMonedas
	GestorCuentas *gestores.GestorCuentas
}

func NewMonedasControlador(gestor *gestores.GestorMonedas, gestorCuentas *gestores.GestorCuentas) *MonedasControlador {
	return &MonedasControlador{Gestor: gestor, GestorCuentas: gestorCuentas}
}

func (mc *MonedasControlador) Dame(c echo.Context) error {
	type Request struct {
		IdMoneda int `param:"IdMoneda"`
	}
	req := &Request{}
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("Parámetros inválidos: "+utils.SanitizarError(err)))
	}
	if req.IdMoneda <= 0 {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("IdMoneda es campo obligatorio"))
	}
	param := &models.Monedas{IdMoneda: req.IdMoneda}
	mensaje, err := param.Dame()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.NewErrorRespuesta("Error al obtener moneda: "+utils.SanitizarError(err)))
	}
	if mensaje != "OK" {
		return c.JSON(http.StatusNotFound, models.NewErrorRespuesta(mensaje))
	}
	return c.JSON(http.StatusOK, param)
}

func (mc *MonedasControlador) Listar(c echo.Context) error {
	type Request struct {
		IncluyeInactivos string `query:"IncluyeInactivos"`
	}
	req := &Request{}
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("Parámetros inválidos: "+utils.SanitizarError(err)))
	}
	if req.IncluyeInactivos == "" {
		req.IncluyeInactivos = "N"
	} else if req.IncluyeInactivos != "N" && req.IncluyeInactivos != "S" {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("IncluyeInactivos debe ser 'S' o 'N'"))
	}
	monedas, err := mc.Gestor.Listar(req.IncluyeInactivos)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.NewErrorRespuesta("Error al buscar monedas: "+utils.SanitizarError(err)))
	}
	return c.JSON(http.StatusOK, monedas)
}

func (mc *MonedasControlador) Crear(c echo.Context) error {
	type Request struct {
		IdMoneda int `json:"IdMoneda"`
	}
	req := &Request{}
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("Parámetros inválidos: "+utils.SanitizarError(err)))
	}
	if req.IdMoneda <= 0 {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("IdMoneda es campo obligatorio"))
	}
	// crea la moneda (estado P), crea la cuenta empresa en TB y activa la moneda (estado A)
	ctx := c.Request().Context()
	mensaje, err := mc.Gestor.Crear(ctx, models.Monedas{IdMoneda: req.IdMoneda, IdCuentaEmpresa: utils.ConcatenarIDString(uint64(req.IdMoneda), uint64(0))})
	log.Printf("\n\nMonedasControlador.Crear: Resultado de creación en GestorMonedas: mensaje='%s', error='%v'", mensaje, err)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.NewErrorRespuesta("Error al crear moneda: "+utils.SanitizarError(err)))
	}
	if mensaje != "OK" {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta(mensaje))
	}

	// intenta crear la cuenta empresa en TB, si falla, borra la moneda creada
	mensaje, _, err = mc.GestorCuentas.Crear(models.Cuentas{IdMoneda: uint32(req.IdMoneda), IdUsuarioFinal: 0, Fecha: time.Now().Format("2006-01-02")})
	if err != nil {
		msjBorrar, errBorrar := mc.Gestor.Borrar(ctx, models.Monedas{IdMoneda: req.IdMoneda})
		if errBorrar != nil {
			log.Printf("Error en rollback de creación moneda después de fallo en creación de cuenta: %v", errBorrar)
		}
		if msjBorrar != "OK" {
			log.Printf("Error en rollback de creación moneda después de fallo en creación de cuenta: %s", msjBorrar)
		}
		return c.JSON(http.StatusInternalServerError, models.NewErrorRespuesta("Error al crear cuenta empresa: "+utils.SanitizarError(err)))
	}

	// si se creó la cuenta en TB, intenta activar la moneda, si falla, borra la moneda pero no la cuenta empresa (TB no permite borrado)
	moneda := &models.Monedas{IdMoneda: req.IdMoneda}
	mensaje, err = moneda.Activar(ctx)
	if err != nil || mensaje != "OK" {
		msjBorrar, errBorrar := mc.Gestor.Borrar(ctx, models.Monedas{IdMoneda: req.IdMoneda})
		if errBorrar != nil {
			log.Printf("Error en rollback de creación moneda después de fallo en activación de moneda: %v", errBorrar)
		}
		if msjBorrar != "OK" {
			log.Printf("Error en rollback de creación moneda después de fallo en activación de moneda: %s", msjBorrar)
		}
		if err != nil {
			return c.JSON(http.StatusInternalServerError, models.NewErrorRespuesta("Error al activar moneda: "+utils.SanitizarError(err)))
		} else {
			return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta(mensaje))
		}
	}
	return c.JSON(http.StatusCreated, map[string]string{"Mensaje": mensaje})
}

func (mc *MonedasControlador) Borrar(c echo.Context) error {
	type Request struct {
		IdMoneda int `param:"IdMoneda"`
	}
	req := &Request{}
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("Parámetros inválidos: "+utils.SanitizarError(err)))
	}
	mensaje, err := mc.Gestor.Borrar(c.Request().Context(), models.Monedas{IdMoneda: req.IdMoneda})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.NewErrorRespuesta("Error al borrar moneda: "+utils.SanitizarError(err)))
	}
	if mensaje != "OK" {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta(mensaje))
	}
	return c.JSON(http.StatusOK, map[string]string{"Mensaje": mensaje})
}

func (mc *MonedasControlador) Activar(c echo.Context) error {
	type Request struct {
		IdMoneda int `param:"IdMoneda"`
	}
	req := &Request{}
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("Parámetros inválidos: "+utils.SanitizarError(err)))
	}
	if req.IdMoneda <= 0 {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("IdMoneda es campo obligatorio"))
	}
	moneda := &models.Monedas{IdMoneda: req.IdMoneda}
	mensaje, err := moneda.Activar(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.NewErrorRespuesta("Error al activar moneda: "+utils.SanitizarError(err)))
	}
	if mensaje != "OK" {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta(mensaje))
	}
	return c.JSON(http.StatusOK, map[string]string{"Mensaje": mensaje})
}

func (mc *MonedasControlador) Desactivar(c echo.Context) error {
	type Request struct {
		IdMoneda int `param:"IdMoneda"`
	}
	req := &Request{}
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("Parámetros inválidos: "+utils.SanitizarError(err)))
	}
	if req.IdMoneda <= 0 {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("IdMoneda es campo obligatorio"))
	}
	moneda := &models.Monedas{IdMoneda: req.IdMoneda}
	mensaje, err := moneda.Desactivar(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.NewErrorRespuesta("Error al desactivar moneda: "+utils.SanitizarError(err)))
	}
	if mensaje != "OK" {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta(mensaje))
	}
	return c.JSON(http.StatusOK, map[string]string{"Mensaje": mensaje})
}
