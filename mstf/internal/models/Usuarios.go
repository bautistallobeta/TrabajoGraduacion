package models

import (
	"MSTransaccionesFinancieras/internal/infra/persistence"
	"errors"
)

type Usuarios struct {
	IdUsuario     int    `json:"id_usuario"`
	NombreUsuario string `json:"nombre_usuario"`
	Email         string `json:"email"`
	TokenSesion   string `json:"token_sesion"`
	Estado        string `json:"estado"`
}

// Instancia un usuario específico por su ID.
// tsp_dame_usuario
// - tokenSesion: token de sesión del usuario que realiza la operación
func (u *Usuarios) Dame(tokenSesion string) error {
	rows, err := persistence.ClienteMySQL.Query("CALL tsp_dame_usuario(?)", tokenSesion)
	if err != nil {
		return err
	}
	defer rows.Close()

	if rows.Next() {
		return rows.Scan(&u.IdUsuario, &u.NombreUsuario, &u.Email, &u.Estado)
	}

	return nil
}

// Permite a un usuario iniciar sesión en el sistema administrativo de MSTF.
// Valida credenciales, regenera el token de sesión y devuelve los datos del usuario.
// pPassword debe venir ya hasheado con md5 desde el cliente.
// Devuelve OK + datos del usuario o el mensaje
// tsp_iniciar_sesion_usuario
// - usuario: nombre de usuario que intenta iniciar sesión
// - password: contraseña hasheada con md5 del usuario que intenta iniciar sesión
func (u *Usuarios) IniciarSesion(usuario string, password string) (string, error) {
	rows, err := persistence.ClienteMySQL.Query("CALL tsp_iniciar_sesion_usuario(?, ?)", usuario, password)
	if err != nil {
		return "", err
	}
	defer rows.Close()
	var mensaje string
	if rows.Next() {
		err = rows.Scan(&mensaje, &u.IdUsuario, &u.NombreUsuario, &u.Email, &u.TokenSesion)
		if err != nil {
			return "", err
		}
		return mensaje, nil
	}
	return "", errors.New("Error al iniciar sesión: intente nuevamente o contacte al administrador")
}

// Permite cambiar el estado de un usuario a A: Activo siempre y cuando esté dado de baja.
// Devuelve OK o el mensaje de error
// tsp_activar_usuario
// - tokenSesion: token de sesión del usuario que realiza la operación
func (u *Usuarios) Activar(tokenSesion string) (string, error) {
	var mensaje string
	err := persistence.ClienteMySQL.QueryRow("CALL tsp_activar_usuario(?, ?)", tokenSesion, u.IdUsuario).Scan(&mensaje)
	return mensaje, err
}

// Permite cambiar el estado de un usuario a I: Inactivo siempre y cuando no esté desactivado.
// No puede desactivarse a sí mismo.
// Devuelve OK o el mensaje de error
// tsp_desactivar_usuario
// - tokenSesion: token de sesión del usuario que realiza la operación
// - idUsuario: Id del usuario a desactivar
func (u *Usuarios) Desactivar(tokenSesion string) (string, error) {
	var mensaje string
	err := persistence.ClienteMySQL.QueryRow("CALL tsp_desactivar_usuario(?, ?)", tokenSesion, u.IdUsuario).Scan(&mensaje)
	return mensaje, err
}

// Permite al usuario pendiente confirmar la cuenta creada ingresando una contraseña nueva y pasando al
// estado activo. La política de contraseñas establece que ésta debe tener una longitud mínima de 6
// caracteres y debe incluir por lo menos una letra y un número. La nueva contraseña debe coincidir
// con la confirmación. Crea la contraseña y activa al usuario.
// Devuelve OK o el mensaje de error
// tsp_confirmar_cuenta_usuario
// - idUsuario: Id del usuario a confirmar
// - password: contraseña hasheada con md5 que el usuario ingresa para confirmar su cuenta
// - confirmarPassword: confirmación de la contraseña hasheada con md5 que el usuario ingresa para confirmar su cuenta
func (u *Usuarios) ConfirmarCuenta(password string, confirmarPassword string) (string, error) {
	var mensaje string
	err := persistence.ClienteMySQL.QueryRow("CALL tsp_confirmar_cuenta_usuario(?, ?, ?)", u.IdUsuario, password, confirmarPassword).Scan(&mensaje)
	return mensaje, err
}
