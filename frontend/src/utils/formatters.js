const _fmt = new Intl.NumberFormat('es-AR', { minimumFractionDigits: 2, maximumFractionDigits: 2 })

export function formatFecha(f)      { return f ? f.slice(0, 10) : '—' }
export function formatTimestamp(ts) { return ts ? String(ts).slice(0, 19).replace('T', ' ') : '—' }
export function formatMonto(v)      { return _fmt.format(parseFloat(v) || 0) }
export function hoy()               { return new Date().toISOString().slice(0, 10) }

export const TIPO_LABEL = { I: 'Ingreso', E: 'Egreso', R: 'Reversión' }
export const TIPO_CLASS  = { I: 'tipo-ingreso', E: 'tipo-egreso', R: 'tipo-reversion' }

export const ESTADO_CUENTA_LABEL   = { A: 'Activa',   I: 'Inactiva' }
export const ESTADO_CUENTA_CLASS   = { A: 'badge-activo', I: 'badge-inactivo' }

export const ESTADO_USUARIO_LABEL  = { A: 'Activo',   I: 'Inactivo',  P: 'Pendiente' }
export const ESTADO_USUARIO_CLASS  = { A: 'badge-activo', I: 'badge-inactivo', P: 'badge-pendiente' }

export const ESTADO_MONEDA_LABEL   = { A: 'Activa',   I: 'Inactiva' }
export const ESTADO_MONEDA_CLASS   = { A: 'badge-activo', I: 'badge-inactivo' }

export const ESTADO_TRANS_LABEL    = { F: 'Finalizada', R: 'Revertida' }
export const ESTADO_TRANS_CLASS    = { F: 'badge-activo', R: 'badge-pendiente' }
