package controllers

import (
	"MSTransaccionesFinancieras/internal/models"
	"MSTransaccionesFinancieras/internal/utils"
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
	req := &Request{}
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("Parámetros inválidos: "+utils.SanitizarError(err)))
	}
	if req.Parametro == "" {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("Parámetro es campo obligatorio"))
	}
	param := &models.Parametros{Parametro: req.Parametro}
	mensaje, err := param.Dame()
	if mensaje != "OK" {
		return c.JSON(http.StatusNotFound, models.NewErrorRespuesta(mensaje))
	}
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.NewErrorRespuesta("Error al obtener parámetro: "+utils.SanitizarError(err)))
	}

	return c.JSON(http.StatusOK, param)
}

func (pc *ParametrosControlador) Modificar(c echo.Context) error {
	type Request struct {
		Parametro string `param:"Parametro"`
		Valor     string `json:"Valor"`
	}
	req := &Request{}
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("Parámetros inválidos: "+utils.SanitizarError(err)))
	}
	if req.Parametro == "" {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("Parámetro es campo obligatorio"))
	}
	if req.Valor == "" {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("Valor es campo obligatorio"))
	}
	param := &models.Parametros{Parametro: req.Parametro}
	mensaje, err := param.ModificarParametro(c.Request().Context(), req.Valor)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.NewErrorRespuesta("Error al modificar parámetro: "+utils.SanitizarError(err)))
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
	req := &Request{}
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("Parámetros inválidos: "+utils.SanitizarError(err)))
	}
	param := &models.Parametros{}
	parametros, err := param.BuscarParametros(req.Cadena)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.NewErrorRespuesta("Error al buscar parámetros: "+utils.SanitizarError(err)))
	}
	return c.JSON(http.StatusOK, parametros)
}
