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

// Permite dar de alta un usuario administrativo en estado P: Pendiente controlando que
// el nombre de usuario y el correo electrónico no existan ya, y sean obligatorios ambos campos.
// Genera un token de sesión y un password aleatorios. Devuelve OK + Id o el mensaje de error
// tsp_crear_usuario
// - tokenSesion: token de sesión del usuario que realiza la operación
// - usuario: nombre de usuario a crear
// - email: correo electrónico del usuario a crear
func (gu *GestorUsuarios) Crear(tokenSesion string, usuario string, email string) (string, int, error) {
	var mensaje string
	var id *int

	err := gu.Db.QueryRow("CALL tsp_crear_usuario(?, ?, ?)", tokenSesion, usuario, email).Scan(&mensaje, &id)
	if err != nil {
		return "", 0, err
	}

	if id == nil {
		return mensaje, 0, nil
	}

	return mensaje, *id, nil
}

// Permite listar todos los usuarios que cumplan con la condición de búsqueda: la cadena debe estar
// contenida en el nombre de usuario o en el email. Puede o no incluir los usuarios dados de baja
// según pIncluyeBajas (S: Si - N: No). Ordena por nombre de usuario.
// tsp_buscar_usuarios
// - tokenSesion: token de sesión del usuario que realiza la operación
// - cadena: cadena de búsqueda para filtrar por nombre de usuario o email
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
		err = rows.Scan(&m.IdUsuario, &m.NombreUsuario, &m.Email, &m.Estado)
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

// Permite modificar el email del usuario logueado. El email es obligatorio y no debe existir ya.
// Devuelve OK o el mensaje de error
// tsp_modificar_usuario
// - tokenSesion: token de sesión del usuario que realiza la operación
// - email: nuevo correo electrónico del usuario
func (gu *GestorUsuarios) Modificar(tokenSesion string, email string) (string, error) {
	var mensaje string
	err := gu.Db.QueryRow("CALL tsp_modificar_usuario(?, ?)", tokenSesion, email).Scan(&mensaje)
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
// Deja al usuario objetivo en estado pendiente y cambia la contraseña por una generada aleatoriamente.
// Devuelve OK o el mensaje de error
// tsp_restablecer_password_usuario
// - tokenSesion: token de sesión del usuario que realiza la operación
// - idUsuarioObjetivo: Id del usuario al que se le restablecerá la contraseña
func (gu *GestorUsuarios) RestablecerPassword(tokenSesion string, idUsuarioObjetivo int) (string, error) {
	var mensaje string
	err := gu.Db.QueryRow("CALL tsp_restablecer_password_usuario(?, ?)", tokenSesion, idUsuarioObjetivo).Scan(&mensaje)
	if err != nil {
		return "", err
	}
	return mensaje, nil
}
