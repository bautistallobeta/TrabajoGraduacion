package models

import (
	"MSTransaccionesFinancieras/internal/auth"
	"MSTransaccionesFinancieras/internal/infra/persistence"
	"context"
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
func (u *Usuarios) Login(Usuario string, Password string) (string, string, error) {
	rows, err := persistence.ClienteMySQL.Query("CALL tsp_login_usuario(?, ?)", Usuario, Password)
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

// Permite cambiar el estado de un usuario a A: Activo siempre y cuando esté inactivo.
// tsp_activar_usuario
func (u *Usuarios) Activar(ctx context.Context) (string, error) {
	credencial, actor := auth.CredencialDesdeCtx(ctx)
	var mensaje string
	err := persistence.ClienteMySQL.QueryRow("CALL tsp_activar_usuario(?, ?, ?)", u.IdUsuario, credencial, actor).Scan(&mensaje)
	return mensaje, err
}

// Permite cambiar el estado de un usuario a I: Inactivo siempre y cuando esté activo.
// tsp_desactivar_usuario
func (u *Usuarios) Desactivar(ctx context.Context) (string, error) {
	credencial, actor := auth.CredencialDesdeCtx(ctx)
	var mensaje string
	err := persistence.ClienteMySQL.QueryRow("CALL tsp_desactivar_usuario(?, ?, ?)", u.IdUsuario, credencial, actor).Scan(&mensaje)
	return mensaje, err
}

// Permite al usuario Pendiente cambiar su contraseña temporal y activarse.
// tsp_confirmar_cuenta_usuario
// - credencial: token de sesión del usuario (para identificarlo en el SP)
// - password: contraseña hasheada con md5
// - confirmarPassword: confirmación de la contraseña hasheada con md5
func (u *Usuarios) ConfirmarCuenta(ctx context.Context, Password string, ConfirmarPassword string) (string, error) {
	credencial, _ := auth.CredencialDesdeCtx(ctx)
	var mensaje string
	err := persistence.ClienteMySQL.QueryRow("CALL tsp_confirmar_cuenta_usuario(?, ?, ?)", credencial, Password, ConfirmarPassword).Scan(&mensaje)
	return mensaje, err
}

// Permite al usuario modificar su contraseña.
// tsp_modificar_password_usuario
// - PasswordAnterior: contraseña actual hasheada con md5
// - PasswordNuevo: nueva contraseña hasheada con md5
// - ConfirmarPassword: confirmación de la nueva contraseña hasheada con md5
func (u *Usuarios) ModificarPassword(ctx context.Context, PasswordAnterior string, PasswordNuevo string, ConfirmarPassword string) (string, error) {
	credencial, _ := auth.CredencialDesdeCtx(ctx)
	var mensaje string
	err := persistence.ClienteMySQL.QueryRow("CALL tsp_modificar_password_usuario(?, ?, ?, ?)", credencial, PasswordAnterior, PasswordNuevo, ConfirmarPassword).Scan(&mensaje)
	if err != nil {
		return "", err
	}
	return mensaje, nil
}

// Permite a un administrador logueado restablecer la contraseña de otro usuario.
// tsp_restablecer_password_usuario
// - Usuario.IdUsuario: ID del usuario al que se le restablecerá la contraseña
func (u *Usuarios) RestablecerPassword() (string, string, error) {
	var mensaje string
	var passwordTemporal sql.NullString
	err := persistence.ClienteMySQL.QueryRow("CALL tsp_restablecer_password_usuario(?)", u.IdUsuario).Scan(&mensaje, &passwordTemporal)
	if err != nil {
		return "", "", err
	}

	if !passwordTemporal.Valid {
		return mensaje, "", nil
	}

	return mensaje, passwordTemporal.String, nil
}
