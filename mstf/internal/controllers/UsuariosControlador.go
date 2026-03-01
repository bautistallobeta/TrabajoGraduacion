package controllers

import (
	"MSTransaccionesFinancieras/internal/auth"
	"MSTransaccionesFinancieras/internal/gestores"
	"MSTransaccionesFinancieras/internal/models"
	"MSTransaccionesFinancieras/internal/utils"
	"context"
	"log"
	"net/http"
	"strings"

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
		Usuario string `json:"Usuario"`
	}
	req := &Request{}

	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("Parámetros inválidos: "+utils.SanitizarError(err)))
	}
	if req.Usuario == "" {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("Usuario es campo obligatorio"))
	}
	mensaje, id, passTemporal, err := uc.Gestor.Crear(c.Request().Context(), models.Usuarios{Usuario: req.Usuario})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.NewErrorRespuesta("Error al crear usuario: "+utils.SanitizarError(err)))
	}
	if id == 0 {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta(mensaje))
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"Id": id, "PasswordTemporal": passTemporal})
}

func (uc *UsuariosControlador) Buscar(c echo.Context) error {
	type Request struct {
		Cadena           string `query:"cadena"`
		IncluyeInactivos string `query:"incluyeInactivos"`
	}
	req := &Request{}
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("Parámetros inválidos: "+utils.SanitizarError(err)))
	}
	if req.IncluyeInactivos == "" {
		req.IncluyeInactivos = "N"
	} else if req.IncluyeInactivos != "S" && req.IncluyeInactivos != "N" {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("IncluyeInactivos debe ser 'S' o 'N'"))
	}
	usuarios, err := uc.Gestor.Buscar(req.Cadena, req.IncluyeInactivos)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.NewErrorRespuesta("Error al buscar usuarios: "+utils.SanitizarError(err)))
	}
	return c.JSON(http.StatusOK, usuarios)
}

func (uc *UsuariosControlador) Borrar(c echo.Context) error {
	type Request struct {
		IdUsuario int `param:"IdUsuario"`
	}
	req := &Request{}
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("Parámetros inválidos: "+utils.SanitizarError(err)))
	}
	mensaje, err := uc.Gestor.Borrar(c.Request().Context(), models.Usuarios{IdUsuario: req.IdUsuario})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.NewErrorRespuesta("Error al borrar usuario: "+utils.SanitizarError(err)))
	}
	if mensaje != "OK" {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta(mensaje))
	}
	return c.JSON(http.StatusOK, map[string]string{"Mensaje": mensaje})
}

func (uc *UsuariosControlador) ModificarPassword(c echo.Context) error {
	type Request struct {
		PasswordAnterior  string `json:"PasswordAnterior"`
		PasswordNuevo     string `json:"PasswordNuevo"`
		ConfirmarPassword string `json:"ConfirmarPassword"`
	}
	req := &Request{}
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("Parámetros inválidos: "+utils.SanitizarError(err)))
	}
	log.Printf("\n\n\n  Request: %+v\n\n\n", req)
	if strings.TrimSpace(req.PasswordAnterior) == "" || strings.TrimSpace(req.PasswordNuevo) == "" || strings.TrimSpace(req.ConfirmarPassword) == "" {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("PasswordAnterior, PasswordNuevo y ConfirmarPassword son campos obligatorios"))
	}
	if err := utils.ValidarFormatoPassword(req.PasswordNuevo); err != nil {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("Formato de contraseña inválido: "+utils.SanitizarError(err)))
	}
	if req.PasswordNuevo != req.ConfirmarPassword {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("La confirmación de la nueva contraseña no coincide"))
	}
	mensaje, err := uc.Gestor.ModificarPassword(c.Request().Context(), utils.MD5Hash(req.PasswordAnterior), utils.MD5Hash(req.PasswordNuevo), utils.MD5Hash(req.ConfirmarPassword))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.NewErrorRespuesta("Error al modificar contraseña: "+utils.SanitizarError(err)))
	}
	if mensaje != "OK" {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta(mensaje))
	}
	return c.JSON(http.StatusOK, map[string]string{"Mensaje": mensaje})
}

func (uc *UsuariosControlador) ReestablecerPassword(c echo.Context) error {
	type Request struct {
		IdUsuario int `json:"IdUsuario"`
	}
	req := &Request{}
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("Parámetros inválidos: "+utils.SanitizarError(err)))
	}
	if req.IdUsuario <= 0 {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("IdUsuario es campo obligatorio"))
	}
	mensaje, passTemporal, err := uc.Gestor.RestablecerPassword(models.Usuarios{IdUsuario: req.IdUsuario})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.NewErrorRespuesta("Error al restablecer contraseña: "+utils.SanitizarError(err)))
	}
	if mensaje != "OK" {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta(mensaje))
	}
	return c.JSON(http.StatusOK, map[string]string{"PasswordTemporal": passTemporal})
}

func (uc *UsuariosControlador) Dame(c echo.Context) error {
	type Request struct {
		IdUsuario int `param:"IdUsuario"`
	}
	req := &Request{}
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("Parámetros inválidos: "+utils.SanitizarError(err)))
	}
	if req.IdUsuario <= 0 {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("IdUsuario es campo obligatorio"))
	}
	usuario := &models.Usuarios{IdUsuario: req.IdUsuario}
	mensaje, err := usuario.Dame()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.NewErrorRespuesta("Error al obtener usuario: "+utils.SanitizarError(err)))
	}
	if mensaje != "OK" {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta(mensaje))
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"Mensaje": mensaje, "Usuario": usuario})
}

func (uc *UsuariosControlador) Login(c echo.Context) error {
	type Request struct {
		Usuario  string `json:"Usuario"`
		Password string `json:"Password"`
	}
	req := &Request{}
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("Parámetros inválidos: "+utils.SanitizarError(err)))
	}
	if req.Usuario == "" || req.Password == "" {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("Usuario y Password son campos obligatorios"))
	}
	usuario := &models.Usuarios{}
	mensaje, tokenSesion, err := usuario.Login(req.Usuario, utils.MD5Hash(req.Password))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.NewErrorRespuesta("Error al iniciar sesión: "+utils.SanitizarError(err)))
	}
	if mensaje[:2] != "OK" {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta(mensaje))
	}
	return c.JSON(http.StatusOK, map[string]string{"Mensaje": mensaje, "TokenSesion": tokenSesion})
}

func (uc *UsuariosControlador) Activar(c echo.Context) error {
	type Request struct {
		IdUsuario int `param:"IdUsuario"`
	}
	req := &Request{}
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("Parámetros inválidos: "+utils.SanitizarError(err)))
	}
	if req.IdUsuario <= 0 {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("IdUsuario es campo obligatorio"))
	}
	usuario := &models.Usuarios{IdUsuario: req.IdUsuario}
	mensaje, err := usuario.Activar(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.NewErrorRespuesta("Error al activar usuario: "+utils.SanitizarError(err)))
	}
	if mensaje != "OK" {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta(mensaje))
	}
	return c.JSON(http.StatusOK, map[string]string{"Mensaje": mensaje})
}

func (uc *UsuariosControlador) Desactivar(c echo.Context) error {
	type Request struct {
		IdUsuario int `param:"IdUsuario"`
	}
	req := &Request{}
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("Parámetros inválidos: "+utils.SanitizarError(err)))
	}
	if req.IdUsuario <= 0 {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("IdUsuario es campo obligatorio"))
	}
	usuario := &models.Usuarios{IdUsuario: req.IdUsuario}
	mensaje, err := usuario.Desactivar(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.NewErrorRespuesta("Error al desactivar usuario: "+utils.SanitizarError(err)))
	}
	if mensaje != "OK" {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta(mensaje))
	}
	return c.JSON(http.StatusOK, map[string]string{"Mensaje": mensaje})
}

func (uc *UsuariosControlador) ConfirmarUsuario(c echo.Context) error {
	type Request struct {
		IdUsuario         int    `param:"IdUsuario"`
		Password          string `json:"Password"`
		ConfirmarPassword string `json:"ConfirmarPassword"`
	}
	req := &Request{}
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("Parámetros inválidos: "+utils.SanitizarError(err)))
	}
	if req.IdUsuario <= 0 {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("IdUsuario es campo obligatorio"))
	}
	if req.Password == "" || req.ConfirmarPassword == "" {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("Password y ConfirmarPassword son campos obligatorios"))
	}
	if err := utils.ValidarFormatoPassword(req.Password); err != nil {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("Formato de contraseña inválido: "+utils.SanitizarError(err)))
	}
	if req.Password != req.ConfirmarPassword {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("La confirmación de la contraseña no coincide"))
	}
	// Esta ruta se omite del middleware de authpero el SP valida el token de sesión internamente
	authHeader := c.Request().Header.Get("Authorization")
	partes := strings.SplitN(authHeader, " ", 2)
	if len(partes) != 2 || strings.ToLower(partes[0]) != "bearer" || partes[1] == "" {
		return c.JSON(http.StatusUnauthorized, models.NewErrorRespuesta("Se requiere token de sesión Bearer"))
	}
	ctx := context.WithValue(c.Request().Context(), auth.ClaveCredencial, partes[1])
	usuario := &models.Usuarios{IdUsuario: req.IdUsuario}
	mensaje, err := usuario.ConfirmarCuenta(ctx, utils.MD5Hash(req.Password), utils.MD5Hash(req.ConfirmarPassword))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.NewErrorRespuesta("Error al confirmar cuenta del usuario: "+utils.SanitizarError(err)))
	}
	if mensaje != "OK" {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta(mensaje))
	}
	return c.JSON(http.StatusOK, map[string]string{"Mensaje": mensaje})
}
