package controllers

import (
	"MSTransaccionesFinancieras/internal/gestores"
	"MSTransaccionesFinancieras/internal/models"
	"net/http"

	"github.com/labstack/echo/v4"
)

type MonedasControlador struct {
	Gestor *gestores.GestorMonedas
}

func NewMonedasControlador(gestor *gestores.GestorMonedas) *MonedasControlador {
	return &MonedasControlador{Gestor: gestor}
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
	if req.IdMoneda == 0 {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("IdMoneda es campo obligatorio"))
	}
	param := &models.Monedas{IdMoneda: req.IdMoneda}
	mensaje, err := param.Dame(tokenSesion)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.NewErrorRespuesta("Error al obtener moneda: "+err.Error()))
	}
	if mensaje != "OK" {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta(mensaje))
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
	if req.IdMoneda == 0 {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("IdMoneda es campo obligatorio"))
	}
	tokenSesion, _ := c.Get("adminToken").(string)
	mensaje, err := mc.Gestor.Crear(tokenSesion, req.IdMoneda)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.NewErrorRespuesta("Error al crear moneda: "+err.Error()))
	}
	return c.JSON(http.StatusOK, map[string]string{"Mensaje": mensaje})
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

func (mc *MonedasControlador) Desctivar(c echo.Context) error {
	type Request struct {
		IdMoneda int `param:"IdMoneda"`
	}
	tokenSesion, _ := c.Get("adminToken").(string)
	req := &Request{}
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("Parámetros inválidos: "+err.Error()))
	}
	if req.IdMoneda == 0 {
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
