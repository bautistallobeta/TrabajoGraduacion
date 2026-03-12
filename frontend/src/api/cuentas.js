import cliente from './cliente'

// GET /cuentas
export async function buscar({ idUsuarioFinal, idMoneda, estado } = {}) {
  const params = {}
  if (idUsuarioFinal) params.IdUsuarioFinal = idUsuarioFinal
  if (idMoneda)       params.IdMoneda = idMoneda
  if (estado)         params.Estado = estado
  const res = await cliente.get('/cuentas', { params })
  return res.data
}

// POST /cuentas
export async function crear(data) {
  const res = await cliente.post('/cuentas', data)
  return res.data
}

// PUT /cuentas/:idusuariofinal/:idmoneda/activar
export async function activar(idUsuarioFinal, idMoneda) {
  const res = await cliente.put(`/cuentas/${idUsuarioFinal}/${idMoneda}/activar`)
  return res.data
}

// PUT /cuentas/:idusuariofinal/:idmoneda/desactivar
export async function desactivar(idUsuarioFinal, idMoneda) {
  const res = await cliente.put(`/cuentas/${idUsuarioFinal}/${idMoneda}/desactivar`)
  return res.data
}

// GET /cuentas/:idusuariofinal/:idmoneda
export async function dame(idUsuarioFinal, idMoneda) {
  const res = await cliente.get(`/cuentas/${idUsuarioFinal}/${idMoneda}`)
  return res.data
}

// Convierte un string datetime-local ("2025-01-15T10:30") a nanosegundos epoch
function datetimeLocalANs(str) {
  if (!str) return null
  const ms = new Date(str).getTime()
  return isNaN(ms) ? null : ms * 1_000_000
}

// GET /cuentas/:idusuariofinal/:idmoneda/transferencias
export async function dameTransferencias(idUsuarioFinal, idMoneda, { timestampMin, timestampMax, incluyeRevertidas } = {}) {
  const params = {}
  const tsMin = datetimeLocalANs(timestampMin)
  const tsMax = datetimeLocalANs(timestampMax)
  if (tsMin)             params.TimestampMin      = tsMin
  if (tsMax)             params.TimestampMax      = tsMax
  if (incluyeRevertidas) params.IncluyeRevertidas = true
  const res = await cliente.get(`/cuentas/${idUsuarioFinal}/${idMoneda}/transferencias`, { params })
  return res.data
}

// GET /cuentas/:idusuariofinal/:idmoneda/historial
export async function dameHistorial(idUsuarioFinal, idMoneda, { timestampMin, timestampMax } = {}) {
  const params = {}
  const tsMin = datetimeLocalANs(timestampMin)
  const tsMax = datetimeLocalANs(timestampMax)
  if (tsMin) params.TimeStampMin = tsMin
  if (tsMax) params.TimeStampMax = tsMax
  const res = await cliente.get(`/cuentas/${idUsuarioFinal}/${idMoneda}/historial`, { params })
  return res.data
}
