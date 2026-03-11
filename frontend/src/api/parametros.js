import cliente from './cliente'

// GET /parametros
export async function buscar(cadena = '') {
  const params = {}
  if (cadena) params.Cadena = cadena
  const res = await cliente.get('/parametros', { params })
  return res.data
}

// PUT /parametros/:parametro
export async function modificar(parametro, valor) {
  const res = await cliente.put(`/parametros/${parametro}`, { Valor: valor })
  return res.data
}
