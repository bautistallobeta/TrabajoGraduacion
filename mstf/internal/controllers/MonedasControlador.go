package controllers

import (
	"MSTransaccionesFinancieras/internal/gestores"
	"MSTransaccionesFinancieras/internal/models"
	"MSTransaccionesFinancieras/internal/utils"
	"log"
	"net/http"
	"strings"
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
	tokenSesion, _ := c.Get("adminToken").(string)
	req := &Request{}
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("Parámetros inválidos: "+err.Error()))
	}
	if req.IdMoneda <= 0 {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("IdMoneda es campo obligatorio"))
	}
	param := &models.Monedas{IdMoneda: req.IdMoneda}
	mensaje, err := param.Dame(tokenSesion)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.NewErrorRespuesta("Error al obtener moneda: "+err.Error()))
	}
	if mensaje != "OK" {
		return c.JSON(http.StatusNotFound, models.NewErrorRespuesta(mensaje))
	}
	return c.JSON(http.StatusOK, param)
}

func (mc *MonedasControlador) Listar(c echo.Context) error {
	type Request struct {
		IncluyeBajas string `query:"IncluyeBajas"`
	}
	req := &Request{}
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("Parámetros inválidos: "+err.Error()))
	}
	tokenSesion, _ := c.Get("adminToken").(string)
	monedas, err := mc.Gestor.Listar(tokenSesion, req.IncluyeBajas)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.NewErrorRespuesta("Error al buscar monedas: "+err.Error()))
	}
	return c.JSON(http.StatusOK, monedas)
}

func (mc *MonedasControlador) Crear(c echo.Context) error {
	type Request struct {
		IdMoneda int `json:"IdMoneda"`
	}
	req := &Request{}
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("Parámetros inválidos: "+err.Error()))
	}
	if req.IdMoneda <= 0 {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("IdMoneda es campo obligatorio"))
	}
	// crea la moneda (estado P)
	// si no falla, crea la cuenta empresa en TB
	// si eso no falla, activa la moneda (estado A)
	// si la primera creacion falla, retorna el error;
	// si tigerbeetle falla, borra la moneda creada;
	// si falla la activacion, borra la moneda pero no la cuenta empresa (TB no permite borrar cuentas)
	// si se reintenta la creacion de la misma moneda, al llegar al punto de TB, si TB retorna AccountExists y se busca que la account no tenga transfers, entonces no se lo toma por error pues indicaría que se está reintentando.
	tokenSesion, _ := c.Get("adminToken").(string)
	mensaje, err := mc.Gestor.Crear(tokenSesion, req.IdMoneda, utils.ConcatenarIDString(uint64(req.IdMoneda), uint64(0)))
	log.Printf("\n\nMonedasControlador.Crear: Resultado de creación en GestorMonedas: mensaje='%s', error='%v'", mensaje, err)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.NewErrorRespuesta("Error al crear moneda: "+err.Error()))
	}
	if mensaje != "OK" {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta(mensaje))
	}

	// intenta crear la cuenta empresa en TB, si falla, borra la moneda creada
	mensaje, err = mc.GestorCuentas.Crear(uint32(req.IdMoneda), uint64(0), time.Now().Format("2006-01-02"), false)
	if err != nil && !strings.Contains(err.Error(), "AccountExists") {
		msjBorrar, errBorrar := mc.Gestor.Borrar(tokenSesion, req.IdMoneda)
		if errBorrar != nil {
			log.Printf("Error en rollback de creación moneda después de fallo en creación de cuenta: %v", errBorrar)
		}
		if msjBorrar != "OK" {
			log.Printf("Error en rollback de creación moneda después de fallo en creación de cuenta: %s", msjBorrar)
		}
		return c.JSON(http.StatusInternalServerError, models.NewErrorRespuesta("Error al crear cuenta empresa: "+err.Error()))
	}

	// si se creó la cuenta en TB, intenta activar la moneda, si falla, borra la moneda pero no la cuenta empresa (TB no permite borrar cuentas)
	moneda := &models.Monedas{IdMoneda: req.IdMoneda}
	mensaje, err = moneda.Activar(tokenSesion)
	if err != nil || mensaje != "OK" {
		msjBorrar, errBorrar := mc.Gestor.Borrar(tokenSesion, req.IdMoneda)
		if errBorrar != nil {
			log.Printf("Error en rollback de creación moneda después de fallo en activación de moneda: %v", errBorrar)
		}
		if msjBorrar != "OK" {
			log.Printf("Error en rollback de creación moneda después de fallo en activación de moneda: %s", msjBorrar)
		}
		if err != nil {
			return c.JSON(http.StatusInternalServerError, models.NewErrorRespuesta("Error al activar moneda: "+err.Error()))
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
	tokenSesion, _ := c.Get("adminToken").(string)
	req := &Request{}
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("Parámetros inválidos: "+err.Error()))
	}
	mensaje, err := mc.Gestor.Borrar(tokenSesion, req.IdMoneda)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.NewErrorRespuesta("Error al borrar moneda: "+err.Error()))
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
	tokenSesion, _ := c.Get("adminToken").(string)
	req := &Request{}
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("Parámetros inválidos: "+err.Error()))
	}
	if req.IdMoneda <= 0 {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("IdMoneda es campo obligatorio"))
	}
	moneda := &models.Monedas{IdMoneda: req.IdMoneda}
	mensaje, err := moneda.Activar(tokenSesion)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.NewErrorRespuesta("Error al activar moneda: "+err.Error()))
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
	tokenSesion, _ := c.Get("adminToken").(string)
	req := &Request{}
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("Parámetros inválidos: "+err.Error()))
	}
	if req.IdMoneda <= 0 {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("IdMoneda es campo obligatorio"))
	}
	moneda := &models.Monedas{IdMoneda: req.IdMoneda}
	mensaje, err := moneda.Desactivar(tokenSesion)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.NewErrorRespuesta("Error al desactivar moneda: "+err.Error()))
	}
	if mensaje != "OK" {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta(mensaje))
	}
	return c.JSON(http.StatusOK, map[string]string{"Mensaje": mensaje})
}
