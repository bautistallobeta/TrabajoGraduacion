package controllers

import (
	"MSTransaccionesFinancieras/internal/gestores"
	"MSTransaccionesFinancieras/internal/models"
	"MSTransaccionesFinancieras/internal/utils"
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

func (uc *UsuariosControlador) Buscar(c echo.Context) error {
	type Request struct {
		Cadena       string `query:"cadena"`
		IncluyeBajas string `query:"incluyeBajas"`
	}
	tokenSesion, _ := c.Get("adminToken").(string)
	req := &Request{}
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("Parámetros inválidos: "+err.Error()))
	}
	if req.IncluyeBajas != "S" && req.IncluyeBajas != "N" {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("IncluyeBajas debe ser 'S' o 'N'"))
	}
	usuarios, err := uc.Gestor.Buscar(tokenSesion, req.Cadena, req.IncluyeBajas)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.NewErrorRespuesta("Error al buscar usuarios: "+err.Error()))
	}
	return c.JSON(http.StatusOK, usuarios)
}

func (uc *UsuariosControlador) Borrar(c echo.Context) error {
	type Request struct {
		IdUsuario int `param:"id_usuario"`
	}
	tokenSesion, _ := c.Get("adminToken").(string)
	req := &Request{}
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("Parámetros inválidos: "+err.Error()))
	}
	mensaje, err := uc.Gestor.Borrar(tokenSesion, req.IdUsuario)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.NewErrorRespuesta("Error al borrar usuario: "+err.Error()))
	}
	if mensaje != "OK" {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta(mensaje))
	}
	return c.JSON(http.StatusOK, map[string]string{"mensaje": mensaje})
}

func (uc *UsuariosControlador) ModificarPassword(c echo.Context) error {
	type Request struct {
		PasswordAnterior  string `json:"password_anterior"`
		PasswordNuevo     string `json:"password_nuevo"`
		ConfirmarPassword string `json:"confirmar_password"`
	}
	tokenSesion, _ := c.Get("adminToken").(string)
	req := &Request{}
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("Parámetros inválidos: "+err.Error()))
	}
	if req.PasswordAnterior == "" || req.PasswordNuevo == "" || req.ConfirmarPassword == "" {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("PasswordAnterior, PasswordNuevo y ConfirmarPassword son campos obligatorios"))
	}
	if req.PasswordNuevo != req.ConfirmarPassword {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("La confirmación de la nueva contraseña no coincide"))
	}
	mensaje, err := uc.Gestor.ModificarPassword(tokenSesion, utils.MD5Hash(req.PasswordAnterior), utils.MD5Hash(req.PasswordNuevo), utils.MD5Hash(req.ConfirmarPassword))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.NewErrorRespuesta("Error al modificar contraseña: "+err.Error()))
	}
	if mensaje != "OK" {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta(mensaje))
	}
	return c.JSON(http.StatusOK, map[string]string{"mensaje": mensaje})
}

func (uc *UsuariosControlador) ReestablecerPassword(c echo.Context) error {
	type Request struct {
		IdUsuario int `json:"id_usuario"`
	}
	tokenSesion, _ := c.Get("adminToken").(string)
	req := &Request{}
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("Parámetros inválidos: "+err.Error()))
	}
	if req.IdUsuario == 0 {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("IdUsuario es campo obligatorio"))
	}
	mensaje, passTemporal, err := uc.Gestor.RestablecerPassword(tokenSesion, req.IdUsuario)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.NewErrorRespuesta("Error al restablecer contraseña: "+err.Error()))
	}
	if mensaje != "OK" {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta(mensaje))
	}
	return c.JSON(http.StatusOK, map[string]string{"mensaje": mensaje, "passwordTemporal": passTemporal})
}

func (uc *UsuariosControlador) Dame(c echo.Context) error {
	type Request struct {
		IdUsuario int `param:"id_usuario"`
	}
	tokenSesion, _ := c.Get("adminToken").(string)
	req := &Request{}
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("Parámetros inválidos: "+err.Error()))
	}
	if req.IdUsuario == 0 {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("IdUsuario es campo obligatorio"))
	}
	usuario := &models.Usuarios{IdUsuario: req.IdUsuario}
	err := usuario.Dame(tokenSesion)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.NewErrorRespuesta("Error al obtener usuario: "+err.Error()))
	}
	return c.JSON(http.StatusOK, usuario)
}

func (uc *UsuariosControlador) Login(c echo.Context) error {
	type Request struct {
		Usuario  string `json:"usuario"`
		Password string `json:"password"`
	}
	req := &Request{}
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("Parámetros inválidos: "+err.Error()))
	}
	if req.Usuario == "" || req.Password == "" {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("Usuario y Password son campos obligatorios"))
	}
	usuario := &models.Usuarios{}
	mensaje, err := usuario.Login(req.Usuario, utils.MD5Hash(req.Password))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.NewErrorRespuesta("Error al iniciar sesión: "+err.Error()))
	}
	if mensaje != "OK" {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta(mensaje))
	}
	return c.JSON(http.StatusOK, usuario)
}

func (uc *UsuariosControlador) Activar(c echo.Context) error {
	type Request struct {
		IdUsuario int `param:"id_usuario"`
	}
	tokenSesion, _ := c.Get("adminToken").(string)
	req := &Request{}
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("Parámetros inválidos: "+err.Error()))
	}
	if req.IdUsuario == 0 {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("IdUsuario es campo obligatorio"))
	}
	usuario := &models.Usuarios{IdUsuario: req.IdUsuario}
	mensaje, err := usuario.Activar(tokenSesion)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.NewErrorRespuesta("Error al activar usuario: "+err.Error()))
	}
	if mensaje != "OK" {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta(mensaje))
	}
	return c.JSON(http.StatusOK, map[string]string{"mensaje": mensaje})
}

func (uc *UsuariosControlador) Desactivar(c echo.Context) error {
	type Request struct {
		IdUsuario int `param:"id_usuario"`
	}
	tokenSesion, _ := c.Get("adminToken").(string)
	req := &Request{}
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("Parámetros inválidos: "+err.Error()))
	}
	if req.IdUsuario == 0 {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("IdUsuario es campo obligatorio"))
	}
	usuario := &models.Usuarios{IdUsuario: req.IdUsuario}
	mensaje, err := usuario.Desactivar(tokenSesion)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.NewErrorRespuesta("Error al desactivar usuario: "+err.Error()))
	}
	if mensaje != "OK" {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta(mensaje))
	}
	return c.JSON(http.StatusOK, map[string]string{"mensaje": mensaje})
}

func (uc *UsuariosControlador) ConfirmarCuenta(c echo.Context) error {
	type Request struct {
		IdUsuario         int    `param:"id_usuario"`
		Password          string `json:"password"`
		ConfirmarPassword string `json:"confirmar_password"`
	}
	tokenSesion, _ := c.Get("adminToken").(string)
	req := &Request{}
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("Parámetros inválidos: "+err.Error()))
	}
	if req.IdUsuario == 0 {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("IdUsuario es campo obligatorio"))
	}
	if req.Password == "" || req.ConfirmarPassword == "" {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta("Password y ConfirmarPassword son campos obligatorios"))
	}
	usuario := &models.Usuarios{IdUsuario: req.IdUsuario}
	mensaje, err := usuario.ConfirmarCuenta(tokenSesion, utils.MD5Hash(req.Password), utils.MD5Hash(req.ConfirmarPassword))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.NewErrorRespuesta("Error al confirmar cuenta del usuario: "+err.Error()))
	}
	if mensaje != "OK" {
		return c.JSON(http.StatusBadRequest, models.NewErrorRespuesta(mensaje))
	}
	return c.JSON(http.StatusOK, map[string]string{"mensaje": mensaje})
}
