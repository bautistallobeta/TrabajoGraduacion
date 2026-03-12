import cliente from './cliente'

// GET /transferencias
export async function buscar(params = {}) {
  const res = await cliente.get('/transferencias', { params })
  return res.data
}

// GET /transferencias/:id
export async function dame(id) {
  const res = await cliente.get(`/transferencias/${id}`)
  return res.data
}

// POST /transferencias
export async function crear(data) {
  const res = await cliente.post('/transferencias', data)
  return res.data
}
