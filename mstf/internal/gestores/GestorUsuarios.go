package gestores

import (
	"MSTransaccionesFinancieras/internal/auth"
	"MSTransaccionesFinancieras/internal/infra/persistence"
	"MSTransaccionesFinancieras/internal/models"
	"context"
	"database/sql"
)

type GestorUsuarios struct {
}

func NewGestorUsuarios() *GestorUsuarios {
	return &GestorUsuarios{}
}

// Permite crear un usuario administrativo en estado P: Pendiente.
// Genera una contraseña aleatoria que se devuelve para informar al usuario.
// Al iniciar sesión por primera vez, deberá cambiar su contraseña y se activará.
// tsp_crear_usuario
// - Usuario.Usuario: nombre de usuario a crear
func (gu *GestorUsuarios) Crear(ctx context.Context, Usuario models.Usuarios) (string, int, string, error) {
	credencial, actor := auth.CredencialDesdeCtx(ctx)
	var mensaje string
	var id sql.NullInt64
	var passwordTemporal sql.NullString
	err := persistence.ClienteMySQL.QueryRow("CALL tsp_crear_usuario(?, ?, ?)", Usuario.Usuario, credencial, actor).Scan(&mensaje, &id, &passwordTemporal)

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
// - Cadena: cadena de búsqueda para filtrar por nombre de usuario
// - IncluyeInactivos: S para incluir usuarios inactivos, N para excluirlos
func (gu *GestorUsuarios) Buscar(Cadena string, IncluyeInactivos string) ([]*models.Usuarios, error) {
	rows, err := persistence.ClienteMySQL.Query("CALL tsp_buscar_usuarios(?, ?)", Cadena, IncluyeInactivos)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	usuarios := make([]*models.Usuarios, 0)
	for rows.Next() {
		var m models.Usuarios
		err = rows.Scan(&m.IdUsuario, &m.Usuario, &m.FechaAlta, &m.Estado, &m.Rol)
		if err != nil {
			return nil, err
		}
		usuarios = append(usuarios, &m)
	}

	return usuarios, nil
}

// Permite eliminar un usuario siempre y cuando no tenga registros en Operaciones y se encuentre Inactivo.
// tsp_borrar_usuario
// - Usuario.IdUsuario: ID del usuario a eliminar
func (gu *GestorUsuarios) Borrar(ctx context.Context, Usuario models.Usuarios) (string, error) {
	credencial, actor := auth.CredencialDesdeCtx(ctx)
	var mensaje string
	err := persistence.ClienteMySQL.QueryRow("CALL tsp_borrar_usuario(?, ?, ?)", Usuario.IdUsuario, credencial, actor).Scan(&mensaje)
	if err != nil {
		return "", err
	}
	return mensaje, nil
}
