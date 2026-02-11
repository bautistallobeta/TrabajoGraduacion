package gestores

import (
	"MSTransaccionesFinancieras/internal/models"
	"database/sql"
)

type GestorUsuarios struct {
	Db *sql.DB
}

func NewGestorUsuarios(db *sql.DB) *GestorUsuarios {
	return &GestorUsuarios{Db: db}
}

// Permite dar de alta un usuario administrativo en estado P: Pendiente.
// Genera una contraseña aleatoria que se devuelve para informar al usuario.
// Al iniciar sesión por primera vez, deberá cambiar su contraseña y se activará.
// Devuelve OK + Id + PasswordTemporal o el mensaje de error.
// tsp_crear_usuario
// - tokenSesion: token de sesión del usuario que realiza la operación
// - usuario: nombre de usuario a crear
func (gu *GestorUsuarios) Crear(tokenSesion string, usuario string) (string, int, string, error) {
	var mensaje string
	var id sql.NullInt64
	var passwordTemporal sql.NullString
	err := gu.Db.QueryRow("CALL tsp_crear_usuario(?, ?)", tokenSesion, usuario).Scan(&mensaje, &id, &passwordTemporal)

	if err != nil {
		return "", 0, "", err
	}

	if !id.Valid {
		return mensaje, 0, "", nil
	}

	return mensaje, int(id.Int64), passwordTemporal.String, nil
}

// Permite listar todos los usuarios que cumplan con la condición de búsqueda: la cadena debe estar
// contenida en el nombre de usuario. Puede o no incluir los usuarios dados de baja
// según pIncluyeBajas (S: Si - N: No). Ordena por nombre de usuario.
// tsp_buscar_usuarios
// - tokenSesion: token de sesión del usuario que realiza la operación
// - cadena: cadena de búsqueda para filtrar por nombre de usuario
// - incluyeBajas: S para incluir usuarios dados de baja, N para excluirlos
func (gu *GestorUsuarios) Buscar(tokenSesion string, cadena string, incluyeBajas string) ([]*models.Usuarios, error) {
	rows, err := gu.Db.Query("CALL tsp_buscar_usuarios(?, ?, ?)", tokenSesion, cadena, incluyeBajas)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var usuarios []*models.Usuarios

	for rows.Next() {
		var m models.Usuarios
		err = rows.Scan(&m.IdUsuario, &m.NombreUsuario, &m.Estado)
		if err != nil {
			return nil, err
		}
		usuarios = append(usuarios, &m)
	}

	return usuarios, nil
}

// Permite eliminar un usuario siempre y cuando no tenga registros en aud_Operaciones.
// No puede eliminarse a sí mismo.
// Devuelve OK o el mensaje de error
// tsp_borrar_usuario
// - tokenSesion: token de sesión del usuario que realiza la operación
// - idUsuario: Id del usuario a eliminar
func (gu *GestorUsuarios) Borrar(tokenSesion string, idUsuario int) (string, error) {
	var mensaje string
	err := gu.Db.QueryRow("CALL tsp_borrar_usuario(?, ?)", tokenSesion, idUsuario).Scan(&mensaje)
	if err != nil {
		return "", err
	}
	return mensaje, nil
}

// Permite al usuario modificar su contraseña. Debe ingresar la contraseña anterior (hasheada con md5),
// la nueva y su confirmación. La política de contraseñas establece que ésta debe tener una longitud
// mínima de 6 caracteres y debe incluir por lo menos una letra y un número.
// Devuelve OK o el mensaje de error
// tsp_modificar_password_usuario
// - tokenSesion: token de sesión del usuario que realiza la operación
// - passwordAnterior: contraseña actual hasheada con md5
// - passwordNuevo: nueva contraseña hasheada con md5
// - confirmarPassword: confirmación de la nueva contraseña hasheada con md5
func (gu *GestorUsuarios) ModificarPassword(tokenSesion string, passwordAnterior string, passwordNuevo string, confirmarPassword string) (string, error) {
	var mensaje string
	err := gu.Db.QueryRow("CALL tsp_modificar_password_usuario(?, ?, ?, ?)", tokenSesion, passwordAnterior, passwordNuevo, confirmarPassword).Scan(&mensaje)
	if err != nil {
		return "", err
	}
	return mensaje, nil
}

// Permite a un administrador logueado restablecer la contraseña de otro usuario.
// Genera una contraseña temporal, deja al usuario en estado Pendiente y regenera su token.
// Devuelve OK + PasswordTemporal o el mensaje de error.
// tsp_restablecer_password_usuario
// - tokenSesion: token de sesión del usuario que realiza la operación
// - idUsuario: Id del usuario al que se le restablecerá la contraseña
func (gu *GestorUsuarios) RestablecerPassword(tokenSesion string, idUsuario int) (string, error) {
	var mensaje string
	err := gu.Db.QueryRow("CALL tsp_restablecer_password_usuario(?, ?)", tokenSesion, idUsuario).Scan(&mensaje)
	if err != nil {
		return "", err
	}
	return mensaje, nil
}
