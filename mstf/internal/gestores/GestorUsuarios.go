package gestores

import (
	"MSTransaccionesFinancieras/internal/infra/persistence"
	"MSTransaccionesFinancieras/internal/models"
	"database/sql"
)

type GestorUsuarios struct {
}

func NewGestorUsuarios() *GestorUsuarios {
	return &GestorUsuarios{}
}

// Permite dar de alta un usuario administrativo en estado P: Pendiente.
// Genera una contraseña aleatoria que se devuelve para informar al usuario.
// Al iniciar sesión por primera vez, deberá cambiar su contraseña y se activará.
// Devuelve OK + Id + PasswordTemporal o el mensaje de error.
// tsp_crear_usuario
func (gu *GestorUsuarios) Crear(usuario string) (string, int, string, error) {
	var mensaje string
	var id sql.NullInt64
	var passwordTemporal sql.NullString
	err := persistence.ClienteMySQL.QueryRow("CALL tsp_crear_usuario(?)", usuario).Scan(&mensaje, &id, &passwordTemporal)

	if err != nil {
		return "", 0, "", err
	}

	if !id.Valid {
		return mensaje, 0, "", nil
	}

	return mensaje, int(id.Int64), passwordTemporal.String, nil
}

// Permite listar todos los usuarios que cumplan con la condición de búsqueda.
// tsp_buscar_usuarios
// - cadena: cadena de búsqueda para filtrar por nombre de usuario
// - incluyeBajas: S para incluir usuarios dados de baja, N para excluirlos
func (gu *GestorUsuarios) Buscar(cadena string, incluyeBajas string) ([]*models.Usuarios, error) {
	rows, err := persistence.ClienteMySQL.Query("CALL tsp_buscar_usuarios(?, ?)", cadena, incluyeBajas)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var usuarios []*models.Usuarios
	for rows.Next() {
		var m models.Usuarios
		err = rows.Scan(&m.IdUsuario, &m.Usuario, &m.FechaAlta, &m.Estado)
		if err != nil {
			return nil, err
		}
		usuarios = append(usuarios, &m)
	}

	return usuarios, nil
}

// Permite eliminar un usuario siempre y cuando no tenga registros en aud_Operaciones.
// tsp_borrar_usuario
func (gu *GestorUsuarios) Borrar(idUsuario int) (string, error) {
	var mensaje string
	err := persistence.ClienteMySQL.QueryRow("CALL tsp_borrar_usuario(?)", idUsuario).Scan(&mensaje)
	if err != nil {
		return "", err
	}
	return mensaje, nil
}

// Permite al usuario modificar su contraseña.
// tsp_modificar_password_usuario
// - credencial: token de sesión del usuario (para identificarlo en el SP)
// - passwordAnterior: contraseña actual hasheada con md5
// - passwordNuevo: nueva contraseña hasheada con md5
// - confirmarPassword: confirmación de la nueva contraseña hasheada con md5
func (gu *GestorUsuarios) ModificarPassword(credencial string, passwordAnterior string, passwordNuevo string, confirmarPassword string) (string, error) {
	var mensaje string
	err := persistence.ClienteMySQL.QueryRow("CALL tsp_modificar_password_usuario(?, ?, ?, ?)", credencial, passwordAnterior, passwordNuevo, confirmarPassword).Scan(&mensaje)
	if err != nil {
		return "", err
	}
	return mensaje, nil
}

// Permite a un administrador logueado restablecer la contraseña de otro usuario.
// tsp_restablecer_password_usuario
func (gu *GestorUsuarios) RestablecerPassword(idUsuario int) (string, string, error) {
	var mensaje string
	var passwordTemporal sql.NullString
	err := persistence.ClienteMySQL.QueryRow("CALL tsp_restablecer_password_usuario(?)", idUsuario).Scan(&mensaje, &passwordTemporal)
	if err != nil {
		return "", "", err
	}

	if !passwordTemporal.Valid {
		return mensaje, "", nil
	}

	return mensaje, passwordTemporal.String, nil
}
