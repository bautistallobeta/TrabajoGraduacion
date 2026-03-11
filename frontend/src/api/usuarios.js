import cliente from './cliente'

// GET /usuarios
export async function buscar(cadena = '', incluyeInactivos = 'N') {
  const params = { incluyeInactivos }
  if (cadena) params.cadena = cadena
  const res = await cliente.get('/usuarios', { params })
  return res.data
}

// POST /usuarios
export async function crear(usuario) {
  const res = await cliente.post('/usuarios', { Usuario: usuario })
  return res.data
}

// PUT /usuarios/activar/:id
export async function activar(id) {
  const res = await cliente.put(`/usuarios/activar/${id}`)
  return res.data
}

// PUT /usuarios/desactivar/:id
export async function desactivar(id) {
  const res = await cliente.put(`/usuarios/desactivar/${id}`)
  return res.data
}

// DELETE /usuarios/:id
export async function borrar(id) {
  const res = await cliente.delete(`/usuarios/${id}`)
  return res.data
}

// PUT /usuarios/password/reestablecer
export async function restablecerPassword(idUsuario) {
  const res = await cliente.put('/usuarios/password/reestablecer', { IdUsuario: idUsuario })
  return res.data
}

// PUT /usuarios/password/modificar
export async function modificarPassword(data) {
  const res = await cliente.put('/usuarios/password/modificar', data)
  return res.data
}

// PUT /usuarios/confirmar-cuenta
export async function confirmarCuenta(password, confirmarPassword) {
  const res = await cliente.put('/usuarios/confirmar-cuenta', {
    Password: password,
    ConfirmarPassword: confirmarPassword
  })
  return res.data
}
