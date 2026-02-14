package models

import (
	"MSTransaccionesFinancieras/internal/infra/persistence"
	"database/sql"
	"errors"
)

type Usuarios struct {
	IdUsuario              int    `json:"id_usuario"`
	Usuario                string `json:"usuario"`
	TokenSesion            string `json:"token_sesion"`
	FechaAlta              string `json:"fecha_alta"`
	Estado                 string `json:"estado"`
	RequiereCambioPassword string `json:"requiere_cambio_password"`
}

// Instancia un usuario específico por su ID.
// tsp_dame_usuario
// - tokenSesion: token de sesión del usuario que realiza la operación
// - idUsuario: Id del usuario a instanciar
func (u *Usuarios) Dame(tokenSesion string) (string, error) {
	rows, err := persistence.ClienteMySQL.Query("CALL tsp_dame_usuario(?, ?)", tokenSesion, u.IdUsuario)
	if err != nil {
		return "", err
	}
	defer rows.Close()
	var mensaje string
	var usuario sql.NullString
	var fechaAlta sql.NullString
	var estado sql.NullString
	if rows.Next() {
		err = rows.Scan(&mensaje, &u.IdUsuario, &usuario, &fechaAlta, &estado)
		if err != nil {
			return mensaje, err
		}
		if usuario.Valid {
			u.Usuario = usuario.String
		} else {
			u.Usuario = ""
		}
		if fechaAlta.Valid {
			u.FechaAlta = fechaAlta.String
		} else {
			u.FechaAlta = ""
		}
		if estado.Valid {
			u.Estado = estado.String
		} else {
			u.Estado = ""
		}
	}
	if mensaje != "OK" {
		return mensaje, errors.New(mensaje)
	}

	return mensaje, nil
}

// Permite a un usuario iniciar sesión en el sistema administrativo de MSTF.
// Valida credenciales, regenera el token de sesión y devuelve los datos del usuario.
// Si el usuario está Pendiente, permite login pero indica que debe cambiar contraseña.
// pPassword debe venir ya hasheado con md5 desde el cliente.
// Devuelve OK + token o el mensaje de error
// tsp_login_usuario
// - usuario: nombre de usuario que intenta iniciar sesión
// - password: contraseña hasheada con md5 del usuario que intenta iniciar sesión
func (u *Usuarios) Login(usuario string, password string) (string, string, error) {
	rows, err := persistence.ClienteMySQL.Query("CALL tsp_login_usuario(?, ?)", usuario, password)
	if err != nil {
		return "", "", err
	}
	defer rows.Close()
	var mensaje string
	var usr sql.NullString
	var fechaAlta sql.NullString
	var tokenSesion sql.NullString
	var requiereCambioPassword sql.NullString
	var estado sql.NullString
	if rows.Next() {
		err = rows.Scan(&mensaje, &u.IdUsuario, &usr, &tokenSesion, &requiereCambioPassword, &fechaAlta, &estado)
		if err != nil {
			return "", "", err
		}
		if usr.Valid {
			u.Usuario = usr.String
		} else {
			u.Usuario = ""
		}
		if fechaAlta.Valid {
			u.FechaAlta = fechaAlta.String
		} else {
			u.FechaAlta = ""
		}
		if tokenSesion.Valid {
			u.TokenSesion = tokenSesion.String
		} else {
			u.TokenSesion = ""
		}
		if requiereCambioPassword.Valid {
			if requiereCambioPassword.String == "S" {
				mensaje += " - Se requiere cambio de contraseña temporal"
			}
			u.RequiereCambioPassword = requiereCambioPassword.String
		} else {
			u.RequiereCambioPassword = ""
		}
		if estado.Valid {
			u.Estado = estado.String
		} else {
			u.Estado = ""
		}

		return mensaje, u.TokenSesion, nil
	}

	return "", "", errors.New("Error: intente nuevamente o contacte al administrador")
}

// Permite cambiar el estado de un usuario a A: Activo siempre y cuando esté dado de baja.
// Devuelve OK o el mensaje de error
// tsp_activar_usuario
// - tokenSesion: token de sesión del usuario que realiza la operación
// - idUsuario: Id del usuario a activar
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

// Permite al usuario Pendiente cambiar su contraseña temporal y activarse.
// Requiere haber iniciado sesión (tener token válido de tsp_login_usuario).
// Devuelve OK o el mensaje de error
// tsp_confirmar_cuenta_usuario
// - tokenSesion: token de sesión del usuario a confirmar
// - password: contraseña hasheada con md5 que el usuario ingresa para confirmar su cuenta
// - confirmarPassword: confirmación de la contraseña hasheada con md5 que el usuario ingresa para confirmar su cuenta
func (u *Usuarios) ConfirmarCuenta(tokenSesion string, password string, confirmarPassword string) (string, error) {
	var mensaje string
	err := persistence.ClienteMySQL.QueryRow("CALL tsp_confirmar_cuenta_usuario(?, ?, ?)", tokenSesion, password, confirmarPassword).Scan(&mensaje)
	return mensaje, err
}
