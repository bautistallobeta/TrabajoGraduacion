package controllers

import (
	"MSTransaccionesFinancieras/internal/models"
	"net/http"

	"github.com/labstack/echo/v4"
)

type ParametrosControlador struct {
}

func NewParametrosControlador() *ParametrosControlador {
	return &ParametrosControlador{}
}

func (pc *ParametrosControlador) Dame(c echo.Context) error {
	type Request struct {
		Parametro string `param:"Parametro"`
	}
	tokenSesion, _ := c.Get("adminToken").(string)
	req := &Request{}
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("Parámetros inválidos: "+err.Error()))
	}
	if req.Parametro == "" {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("Parámetro es campo obligatorio"))
	}
	param := &models.Parametros{Parametro: req.Parametro}
	mensaje, err := param.Dame(tokenSesion)
	if mensaje != "OK" {
		return c.JSON(http.StatusNotFound, models.NewErrorRespuesta(mensaje))
	}
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.NewErrorRespuesta("Error al obtener parámetro: "+err.Error()))
	}

	return c.JSON(http.StatusOK, param)
}

func (pc *ParametrosControlador) Modificar(c echo.Context) error {
	type Request struct {
		Parametro string `param:"Parametro"`
		Valor     string `json:"Valor"`
	}
	tokenSesion, _ := c.Get("adminToken").(string)
	req := &Request{}
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("Parámetros inválidos: "+err.Error()))
	}
	if req.Parametro == "" {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("Parámetro es campo obligatorio"))
	}
	if req.Valor == "" {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("Valor es campo obligatorio"))
	}
	param := &models.Parametros{Parametro: req.Parametro}
	mensaje, err := param.ModificarParametro(tokenSesion, req.Valor)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.NewErrorRespuesta("Error al modificar parámetro: "+err.Error()))
	}
	if mensaje != "OK" {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta(mensaje))
	}
	return c.JSON(http.StatusOK, map[string]string{"Mensaje": mensaje})
}

func (pc *ParametrosControlador) Buscar(c echo.Context) error {
	type Request struct {
		Cadena string `query:"Cadena"`
	}
	tokenSesion, _ := c.Get("adminToken").(string)
	req := &Request{}
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("Parámetros inválidos: "+err.Error()))
	}
	param := &models.Parametros{}
	parametros, err := param.BuscarParametros(tokenSesion, req.Cadena)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.NewErrorRespuesta("Error al buscar parámetros: "+err.Error()))
	}
	return c.JSON(http.StatusOK, parametros)
}
