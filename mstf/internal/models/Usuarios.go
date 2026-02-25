package models

import (
	"MSTransaccionesFinancieras/internal/infra/persistence"
	"database/sql"
	"errors"
)

type Usuarios struct {
	IdUsuario              int    `json:"IdUsuario"`
	Usuario                string `json:"Usuario"`
	TokenSesion            string `json:"TokenSesion"`
	FechaAlta              string `json:"FechaAlta"`
	Estado                 string `json:"Estado"`
	RequiereCambioPassword string `json:"RequiereCambioPassword"`
}

// Instancia un usuario específico por su ID.
// tsp_dame_usuario
func (u *Usuarios) Dame() (string, error) {
	rows, err := persistence.ClienteMySQL.Query("CALL tsp_dame_usuario(?)", u.IdUsuario)
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
	return mensaje, nil
}

// Permite a un usuario iniciar sesión en el sistema administrativo de MSTF.
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
// tsp_activar_usuario
func (u *Usuarios) Activar() (string, error) {
	var mensaje string
	err := persistence.ClienteMySQL.QueryRow("CALL tsp_activar_usuario(?)", u.IdUsuario).Scan(&mensaje)
	return mensaje, err
}

// Permite cambiar el estado de un usuario a I: Inactivo siempre y cuando no esté desactivado.
// tsp_desactivar_usuario
func (u *Usuarios) Desactivar() (string, error) {
	var mensaje string
	err := persistence.ClienteMySQL.QueryRow("CALL tsp_desactivar_usuario(?)", u.IdUsuario).Scan(&mensaje)
	return mensaje, err
}

// Permite al usuario Pendiente cambiar su contraseña temporal y activarse.
// tsp_confirmar_cuenta_usuario
// - credencial: token de sesión del usuario (para identificarlo en el SP)
// - password: contraseña hasheada con md5
// - confirmarPassword: confirmación de la contraseña hasheada con md5
func (u *Usuarios) ConfirmarCuenta(credencial string, password string, confirmarPassword string) (string, error) {
	var mensaje string
	err := persistence.ClienteMySQL.QueryRow("CALL tsp_confirmar_cuenta_usuario(?, ?, ?)", credencial, password, confirmarPassword).Scan(&mensaje)
	return mensaje, err
}
