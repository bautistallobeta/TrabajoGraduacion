import cliente from './cliente'

// GET /monedas
export async function listar(incluyeInactivos = 'N') {
  const res = await cliente.get('/monedas', { params: { IncluyeInactivos: incluyeInactivos } })
  return res.data
}

// POST /monedas
export async function crear(idMoneda) {
  const res = await cliente.post('/monedas', { IdMoneda: idMoneda })
  return res.data
}

// DELETE /monedas/:id
export async function borrar(idMoneda) {
  const res = await cliente.delete(`/monedas/${idMoneda}`)
  return res.data
}

// PUT /monedas/:id/activar
export async function activar(idMoneda) {
  const res = await cliente.put(`/monedas/${idMoneda}/activar`)
  return res.data
}

// PUT /monedas/:id/desactivar
export async function desactivar(idMoneda) {
  const res = await cliente.put(`/monedas/${idMoneda}/desactivar`)
  return res.data
}
