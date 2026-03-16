<script setup>
import { ref, onMounted } from 'vue'
import { useAlert } from '../composables/useAlert'
import { useModal } from '../composables/useModal'
import { usePagination } from '../composables/usePagination'
import Paginacion from '../components/Paginacion.vue'
import {
  formatFecha, formatTimestamp, formatMonto, hoy,
  TIPO_LABEL, TIPO_CLASS,
  ESTADO_TRANS_LABEL as ESTADO_LABEL, ESTADO_TRANS_CLASS as ESTADO_CLASS
} from '../utils/formatters'
import * as api from '../api/transferencias'
import { dame as dameParametro } from '../api/parametros'

const esDemo = import.meta.env.VITE_DEMO_MODE === 'true'

const transferencias = ref([])
const cargando       = ref(false)
const total          = ref(0)
const limite         = ref(100)

const { paginaActual, totalPaginas, itemsEnPagina: transferenciasEnPagina, botones: paginasBotones } = usePagination(transferencias)

const filtros = ref({
  idTransferencia:   '',
  idUsuarioFinal:    '',
  idMoneda:          '',
  idCategoria:       '',
  montoMin:          '',
  montoMax:          '',
  fechaDesde:        '',
  fechaHasta:        '',
  incluyeRevertidas: false
})

const { alerta, mostrarAlerta } = useAlert()

async function buscar() {
  cargando.value = true
  paginaActual.value = 1
  try {
    const params = {}
    if (filtros.value.idTransferencia)   params.IdsTransferencia  = filtros.value.idTransferencia
    if (filtros.value.idUsuarioFinal)    params.IdUsuarioFinal    = filtros.value.idUsuarioFinal
    if (filtros.value.idMoneda)          params.IdMoneda          = filtros.value.idMoneda
    if (filtros.value.idCategoria)       params.IdCategoria       = filtros.value.idCategoria
    if (filtros.value.montoMin)          params.MontoMin          = filtros.value.montoMin
    if (filtros.value.montoMax)          params.MontoMax          = filtros.value.montoMax
    if (filtros.value.fechaDesde)        params.FechaDesde        = filtros.value.fechaDesde
    if (filtros.value.fechaHasta)        params.FechaHasta        = filtros.value.fechaHasta
    if (filtros.value.incluyeRevertidas) params.IncluyeRevertidas = true

    const res = await api.buscar(params)
    transferencias.value = res.Transferencias ?? []
    total.value          = res.Total ?? 0
  } catch (e) {
    mostrarAlerta('danger', e.response?.data?.error ?? 'Error al cargar transferencias')
  } finally {
    cargando.value = false
  }
}

onMounted(async () => {
  try {
    const p = await dameParametro('LIMITEBUSCARTRANSFERENCIAS')
    const val = parseInt(p.Valor)
    if (!isNaN(val) && val > 0) limite.value = val
  } catch { /* usa el default */ }
  buscar()
})

// Modal p ver detalle
const detalleModalEl  = ref(null)
const detalleActual   = ref(null)
const cargandoDetalle = ref(false)

const detalleModal = useModal(detalleModalEl, () => {
  detalleActual.value   = null
  cargandoDetalle.value = false
})

async function verDetalle(t) {
  detalleActual.value = t
  detalleModal.show()
  cargandoDetalle.value = true
  try {
    detalleActual.value = await api.dame(t.IdTransferencia)
  } catch {
    // Mantener datos del listado si falla
  } finally {
    cargandoDetalle.value = false
  }
}

// Modal p crear (solo demo)
const crearModalEl = ref(null)
const nuevaTransf  = ref({ id: '', idUsuarioFinal: '', idMoneda: '', tipo: 'I', monto: '', idCategoria: '', fecha: hoy() })
const creando      = ref(false)

const crearModal = useModal(crearModalEl, () => {
  nuevaTransf.value = { id: '', idUsuarioFinal: '', idMoneda: '', tipo: 'I', monto: '', idCategoria: '', fecha: hoy() }
  creando.value = false
})

function abrirModalCrear() {
  crearModal.show()
}

async function crearTransferencia() {
  const { id, idUsuarioFinal, idMoneda, tipo, monto, idCategoria, fecha } = nuevaTransf.value
  if (!id || !idUsuarioFinal || !idMoneda || !tipo || !idCategoria || !fecha) return
  if (!monto) return
  creando.value = true
  try {
    const body = {
      IdTransferencia: id,
      IdUsuarioFinal:  parseInt(idUsuarioFinal),
      IdMoneda:        parseInt(idMoneda),
      Tipo:            tipo,
      IdCategoria:     parseInt(idCategoria),
      Fecha:           fecha
    }
    if (monto) body.Monto = parseFloat(monto)
    const res = await api.crear(body)
    crearModal.hide()
    mostrarAlerta('success', `Transferencia encolada — ID: ${res.Id}`)
  } catch (e) {
    mostrarAlerta('danger', e.response?.data?.error ?? 'Error al crear transferencia')
    crearModal.hide()
  } finally {
    creando.value = false
  }
}

// Modal p revertir
const revertirModalEl        = ref(null)
const transferenciaARevertir = ref(null)
const fechaReversion         = ref('')
const revirtiendo            = ref(false)

const revertirModal = useModal(revertirModalEl, () => {
  transferenciaARevertir.value = null
  fechaReversion.value = ''
  revirtiendo.value = false
})

function abrirReversion(t) {
  transferenciaARevertir.value = t
  fechaReversion.value = new Date().toISOString().slice(0, 10)
  detalleModal.hide()
  // esperar a que cierre el detalle antes de abrir
  detalleModalEl.value.addEventListener('hidden.bs.modal', () => {
    revertirModal.show()
  }, { once: true })
}

async function realizarReversion() {
  const t = transferenciaARevertir.value
  if (!t || !fechaReversion.value) return
  revirtiendo.value = true
  try {
    const res = await api.crear({
      IdTransferencia: t.IdTransferencia,
      IdUsuarioFinal:  t.IdUsuarioFinal,
      IdMoneda:        t.IdMoneda,
      Tipo:            'R',
      IdCategoria:     t.Categoria,
      Fecha:           fechaReversion.value,
      Monto:           0
    })
    revertirModal.hide()
    mostrarAlerta('success', `Reversión encolada — ID: ${res.Id}`)
    buscar()
  } catch (e) {
    mostrarAlerta('danger', e.response?.data?.error ?? 'Error al revertir transferencia')
    revertirModal.hide()
  } finally {
    revirtiendo.value = false
  }
}
</script>

<template>
  <div>

    <div class="section-header mb-4">
      <h1 class="page-title">Transferencias</h1>
      <button v-if="esDemo" type="button" class="btn btn-success" @click="abrirModalCrear">
        + Nueva transferencia (demo)
      </button>
    </div>

    <div v-if="alerta" :class="`alert alert-${alerta.tipo} alert-dismissible mb-4`" role="alert">
      {{ alerta.mensaje }}
      <button type="button" class="btn-close" @click="alerta = null"></button>
    </div>

    <!-- Filtros -->
    <div class="card mb-3">
      <div class="card-body py-3">
        <form @submit.prevent="buscar">
          <div class="filtros-grid">
            <div>
              <label class="form-label">ID Transferencia</label>
              <input v-model="filtros.idTransferencia" type="text" class="form-control" placeholder="ID exacto..." />
            </div>
            <div>
              <label class="form-label">Usuario</label>
              <input v-model="filtros.idUsuarioFinal" type="number" min="1" class="form-control" placeholder="ID..." />
            </div>
            <div>
              <label class="form-label">Moneda</label>
              <input v-model="filtros.idMoneda" type="number" min="1" class="form-control" placeholder="ID..." />
            </div>
            <div>
              <label class="form-label">Categoría</label>
              <input v-model="filtros.idCategoria" type="number" min="1" class="form-control" placeholder="ID..." />
            </div>
            <div>
              <label class="form-label">Monto mín</label>
              <input v-model="filtros.montoMin" type="number" min="0" step="0.01" class="form-control" placeholder="0.00" />
            </div>
            <div>
              <label class="form-label">Monto máx</label>
              <input v-model="filtros.montoMax" type="number" min="0" step="0.01" class="form-control" placeholder="0.00" />
            </div>
            <div>
              <label class="form-label">Desde</label>
              <input v-model="filtros.fechaDesde" type="date" class="form-control" />
            </div>
            <div>
              <label class="form-label">Hasta</label>
              <input v-model="filtros.fechaHasta" type="date" class="form-control" />
            </div>
            <div style="padding-bottom: 6px" class="d-flex align-items-end">
              <div class="d-flex align-items-center gap-2">
                <input
                  id="incluyeRevertidas"
                  v-model="filtros.incluyeRevertidas"
                  type="checkbox"
                  class="form-check-input mt-0"
                />
                <label for="incluyeRevertidas" class="form-label mb-0" style="white-space: nowrap">Incl. revertidas</label>
              </div>
            </div>
            <div class="d-flex align-items-end">
              <button type="submit" class="btn btn-outline-primary w-100" :disabled="cargando">
                Buscar
              </button>
            </div>
          </div>
        </form>
      </div>
    </div>

    <!-- Tabla -->
    <div class="card">
      <div class="card-body p-0">

        <div v-if="cargando" class="loading-state">
          <div class="spinner-border spinner-border-sm" role="status" aria-hidden="true"></div>
          <span>Cargando...</span>
        </div>

        <div v-else-if="transferencias.length === 0" class="empty-state">
          <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <polyline points="17 1 21 5 17 9"/><path d="M3 11V9a4 4 0 0 1 4-4h14"/><polyline points="7 23 3 19 7 15"/><path d="M21 13v2a4 4 0 0 1-4 4H3"/>
          </svg>
          <p>No se encontraron transferencias</p>
        </div>

        <div v-else class="table-responsive">
          <div class="total-row d-flex align-items-center justify-content-between">
            <span>{{ total }} resultado{{ total !== 1 ? 's' : '' }}</span>
            <span v-if="total === limite" class="limite-aviso">
              Se alcanzó el límite de búsqueda. Acotar los filtros para resultados más precisos.
            </span>
          </div>
          <table class="table mb-0">
            <thead>
              <tr>
                <th>ID</th>
                <th>Usuario</th>
                <th>Moneda</th>
                <th>Tipo</th>
                <th>Categoría</th>
                <th class="text-end">Monto</th>
                <th>Fecha Alta</th>
                <th>Procesada</th>
                <th>Estado</th>
                <th style="width: 1%; white-space: nowrap"></th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="t in transferenciasEnPagina" :key="t.IdTransferencia" class="clickable" @click="verDetalle(t)">
                <td class="cell-id" style="max-width: 140px; overflow: hidden; text-overflow: ellipsis">
                  {{ t.IdTransferencia }}
                </td>
                <td class="cell-id">{{ t.IdUsuarioFinal }}</td>
                <td class="cell-id">{{ t.IdMoneda }}</td>
                <td>
                  <span v-if="t.Tipo" :class="`tipo-badge ${TIPO_CLASS[t.Tipo] ?? ''}`">
                    {{ TIPO_LABEL[t.Tipo] ?? t.Tipo }}
                  </span>
                  <span v-else class="text-muted">—</span>
                </td>
                <td class="cell-id">{{ t.Categoria }}</td>
                <td class="cell-monto">{{ formatMonto(t.Monto) }}</td>
                <td>{{ formatFecha(t.Fecha) }}</td>
                <td class="cell-id">{{ formatTimestamp(t.FechaProceso) }}</td>
                <td>
                  <span :class="`badge ${ESTADO_CLASS[t.Estado] ?? ''}`">
                    {{ ESTADO_LABEL[t.Estado] ?? t.Estado }}
                  </span>
                </td>
                <td>
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" style="color: var(--text-tertiary)">
                    <polyline points="9 18 15 12 9 6"/>
                  </svg>
                </td>
              </tr>
            </tbody>
          </table>
          <Paginacion v-model="paginaActual" :total-paginas="totalPaginas" :botones="paginasBotones" />
        </div>

      </div>
    </div>

    <!-- Modal crear (demo) -->
    <div ref="crearModalEl" class="modal fade" tabindex="-1" aria-hidden="true">
      <div class="modal-dialog modal-dialog-centered" style="max-width: 440px">
        <div class="modal-content">
          <div class="modal-header">
            <div>
              <h5 class="modal-title">Nueva transferencia</h5>
              <div style="font-size: 0.75rem; color: var(--text-tertiary); margin-top: 2px">Solo para demo — publica en Kafka y retorna 202</div>
            </div>
            <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Cerrar"></button>
          </div>
          <form @submit.prevent="crearTransferencia">
            <div class="modal-body">
              <div class="crear-grid">
                <div class="crear-full">
                  <label class="form-label">ID de transferencia</label>
                  <input v-model="nuevaTransf.id" type="text" class="form-control" placeholder="ej: 98765432100000000001" :disabled="creando" required autofocus />
                </div>
                <div>
                  <label class="form-label">Tipo</label>
                  <select v-model="nuevaTransf.tipo" class="form-select" :disabled="creando" required>
                    <option value="I">Ingreso</option>
                    <option value="E">Egreso</option>
                  </select>
                </div>
                <div>
                  <label class="form-label">Monto</label>
                  <input
                    v-model="nuevaTransf.monto"
                    type="number"
                    min="0"
                    step="0.01"
                    class="form-control"
                    placeholder="0.00"
                    :disabled="creando"
                    required
                  />
                </div>
                <div>
                  <label class="form-label">Usuario</label>
                  <input v-model="nuevaTransf.idUsuarioFinal" type="number" min="1" class="form-control" placeholder="ID..." :disabled="creando" required />
                </div>
                <div>
                  <label class="form-label">Moneda</label>
                  <input v-model="nuevaTransf.idMoneda" type="number" min="1" class="form-control" placeholder="ID..." :disabled="creando" required />
                </div>
                <div>
                  <label class="form-label">Categoría</label>
                  <input v-model="nuevaTransf.idCategoria" type="number" min="0" class="form-control" placeholder="ID..." :disabled="creando" required />
                </div>
                <div>
                  <label class="form-label">Fecha</label>
                  <input v-model="nuevaTransf.fecha" type="date" class="form-control" :disabled="creando" required />
                </div>
              </div>
            </div>
            <div class="modal-footer">
              <button type="button" class="btn btn-outline-secondary btn-sm" data-bs-dismiss="modal">Cancelar</button>
              <button type="submit" class="btn btn-primary btn-sm" :disabled="creando">
                <span v-if="creando" class="spinner-border spinner-border-sm me-1" role="status" aria-hidden="true"></span>
                Enviar
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>

    <!-- Modal detalle -->
    <div ref="detalleModalEl" class="modal fade" tabindex="-1" aria-hidden="true">
      <div class="modal-dialog modal-dialog-centered" style="max-width: 480px">
        <div class="modal-content" v-if="detalleActual">
          <div class="modal-header">
            <div>
              <h5 class="modal-title d-flex align-items-center gap-2">
                Transferencia
                <span v-if="cargandoDetalle" class="spinner-border spinner-border-sm text-secondary" role="status" aria-hidden="true"></span>
              </h5>
              <div class="modal-id">{{ detalleActual.IdTransferencia }}</div>
            </div>
            <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Cerrar"></button>
          </div>
          <div class="modal-body">
            <div class="detalle-grid">
              <div class="detalle-item">
                <span class="detalle-label">Tipo</span>
                <span class="detalle-tipo" :class="TIPO_CLASS[detalleActual.Tipo]">
                  {{ TIPO_LABEL[detalleActual.Tipo] ?? (detalleActual.Tipo || '—') }}
                </span>
              </div>
              <div class="detalle-item">
                <span class="detalle-label">Estado</span>
                <span :class="`badge ${ESTADO_CLASS[detalleActual.Estado] ?? ''}`">
                  {{ ESTADO_LABEL[detalleActual.Estado] ?? detalleActual.Estado }}
                </span>
              </div>
              <div class="detalle-item">
                <span class="detalle-label">Monto</span>
                <span class="detalle-monto">{{ formatMonto(detalleActual.Monto) }}</span>
              </div>
              <div class="detalle-item">
                <span class="detalle-label">Usuario</span>
                <span class="detalle-value">{{ detalleActual.IdUsuarioFinal }}</span>
              </div>
              <div class="detalle-item">
                <span class="detalle-label">Moneda</span>
                <span class="detalle-value">{{ detalleActual.IdMoneda }}</span>
              </div>
              <div class="detalle-item">
                <span class="detalle-label">Categoría</span>
                <span class="detalle-value">{{ detalleActual.Categoria }}</span>
              </div>
              <div class="detalle-item">
                <span class="detalle-label">Fecha negocio</span>
                <span class="detalle-value">{{ formatFecha(detalleActual.Fecha) }}</span>
              </div>
              <div class="detalle-item">
                <span class="detalle-label">Procesada</span>
                <span class="detalle-value" style="font-size: 0.8rem">{{ formatTimestamp(detalleActual.FechaProceso) }}</span>
              </div>
              <div class="detalle-item detalle-full">
                <span class="detalle-label">Cuenta débito</span>
                <span class="detalle-value mono">{{ detalleActual.IdCuentaDebito }}</span>
              </div>
              <div class="detalle-item detalle-full">
                <span class="detalle-label">Cuenta crédito</span>
                <span class="detalle-value mono">{{ detalleActual.IdCuentaCredito }}</span>
              </div>
              <div v-if="detalleActual.IdTransferenciaOriginal" class="detalle-item detalle-full">
                <span class="detalle-label">Transfer original</span>
                <span class="detalle-value mono">{{ detalleActual.IdTransferenciaOriginal }}</span>
              </div>
            </div>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-outline-secondary btn-sm" data-bs-dismiss="modal">Cerrar</button>
            <button
              v-if="esDemo && detalleActual.Tipo !== 'R' && detalleActual.Estado !== 'R'"
              type="button"
              class="btn btn-outline-danger btn-sm"
              @click="abrirReversion(detalleActual)"
            >
              Revertir
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Modal reversión -->
    <div ref="revertirModalEl" class="modal fade" tabindex="-1" aria-hidden="true">
      <div class="modal-dialog modal-dialog-centered" style="max-width: 400px">
        <div class="modal-content" v-if="transferenciaARevertir">
          <div class="modal-header">
            <div>
              <h5 class="modal-title">Revertir transferencia</h5>
              <div style="font-size: 0.75rem; color: var(--text-tertiary); margin-top: 2px">Solo para demo — publica en Kafka y retorna 202</div>
            </div>
            <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Cerrar"></button>
          </div>
          <form @submit.prevent="realizarReversion">
            <div class="modal-body">
              <div class="detalle-grid mb-3">
                <div class="detalle-item detalle-full">
                  <span class="detalle-label">ID a revertir</span>
                  <span class="detalle-value mono">{{ transferenciaARevertir.IdTransferencia }}</span>
                </div>
                <div class="detalle-item">
                  <span class="detalle-label">Usuario</span>
                  <span class="detalle-value">{{ transferenciaARevertir.IdUsuarioFinal }}</span>
                </div>
                <div class="detalle-item">
                  <span class="detalle-label">Moneda</span>
                  <span class="detalle-value">{{ transferenciaARevertir.IdMoneda }}</span>
                </div>
                <div class="detalle-item">
                  <span class="detalle-label">Monto original</span>
                  <span class="detalle-monto" style="font-size: 0.9375rem">{{ formatMonto(transferenciaARevertir.Monto) }}</span>
                </div>
                <div class="detalle-item">
                  <span class="detalle-label">Categoría</span>
                  <span class="detalle-value">{{ transferenciaARevertir.Categoria }}</span>
                </div>
              </div>
              <div>
                <label class="form-label">Fecha de reversión</label>
                <input
                  v-model="fechaReversion"
                  type="date"
                  class="form-control"
                  :disabled="revirtiendo"
                  required
                />
              </div>
            </div>
            <div class="modal-footer">
              <button type="button" class="btn btn-outline-secondary btn-sm" data-bs-dismiss="modal">Cancelar</button>
              <button type="submit" class="btn btn-outline-danger btn-sm" :disabled="revirtiendo || !fechaReversion">
                <span v-if="revirtiendo" class="spinner-border spinner-border-sm me-1" role="status" aria-hidden="true"></span>
                Confirmar reversión
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>

  </div>
</template>

<style scoped>
:deep(tbody td) {
  padding-top: 0.75rem;
  padding-bottom: 0.75rem;
}

.filtros-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(130px, 1fr));
  gap: 0.75rem;
  align-items: end;
}

.modal-id {
  font-family: var(--font-mono);
  font-size: 0.6875rem;
  color: var(--text-tertiary);
  margin-top: 0.125rem;
  word-break: break-all;
}

.detalle-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1rem 1.5rem;
}

.detalle-full {
  grid-column: 1 / -1;
}

.detalle-item {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.detalle-label {
  font-family: var(--font-mono);
  font-size: 0.625rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: var(--text-tertiary);
}

.detalle-value {
  font-size: 0.875rem;
  color: var(--text-primary);
}

.detalle-value.mono {
  font-family: var(--font-mono);
  font-size: 0.75rem;
  word-break: break-all;
}

.crear-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.875rem;
}

.crear-full {
  grid-column: 1 / -1;
}

.detalle-monto {
  font-family: var(--font-mono);
  font-size: 1.125rem;
  font-weight: 600;
  color: var(--accent);
}

.detalle-tipo {
  font-size: 0.9375rem;
  font-weight: 600;
  color: var(--text-primary);
}

.detalle-tipo.tipo-ingreso  { color: #065F46; }
.detalle-tipo.tipo-egreso   { color: #991B1B; }
.detalle-tipo.tipo-reversion { color: #0C4A6E; }
</style>
