package controllers

import (
	"MSTransaccionesFinancieras/internal/gestores"
	"MSTransaccionesFinancieras/internal/models"
	"net/http"

	"github.com/labstack/echo/v4"
)

type UsuariosControlador struct {
	Gestor *gestores.GestorUsuarios
}

func NewUsuariosControlador(gu *gestores.GestorUsuarios) *UsuariosControlador {
	return &UsuariosControlador{Gestor: gu}
}

func (uc *UsuariosControlador) Crear(c echo.Context) error {
	type Request struct {
		Usuario string `json:"usuario"`
	}
	// el token se obtiene del header authorization y se pasa al gestor
	// el token viene dado del middleware de autenticación, no se valida en el controlador
	tokenSesion, _ := c.Get("adminToken").(string)
	req := &Request{}

	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("Parámetros inválidos: "+err.Error()))
	}
	if req.Usuario == "" {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("Usuario es campo obligatorio"))
	}
	mensaje, id, passTemporal, err := uc.Gestor.Crear(tokenSesion, req.Usuario)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.NewErrorRespuesta("Error al crear usuario: "+err.Error()))
	}
	if id == 0 {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta(mensaje))
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"id": id, "passwordTemporal": passTemporal})
}
