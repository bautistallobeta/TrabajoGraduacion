<script setup>
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { Modal } from 'bootstrap'
import ConfirmModal from '../components/ConfirmModal.vue'
import * as api from '../api/cuentas'
import { dame as dameParametro } from '../api/parametros'

const cuentas  = ref([])
const cargando = ref(false)
const total    = ref(0)
const filtros  = ref({ idUsuarioFinal: '', idMoneda: '', estado: '' })

const limiteCuentas       = ref(100)
const paginaCuentas       = ref(1)
const limiteDetalle       = ref(100)
const paginaTransfs       = ref(1)
const paginaHistorial     = ref(1)
const POR_PAGINA          = 50

function paginasBotones(paginaActual, totalPaginas) {
  const tot = totalPaginas
  const act = paginaActual
  if (tot <= 7) return Array.from({ length: tot }, (_, i) => i + 1)
  const pags = [1]
  if (act > 3) pags.push('...')
  for (let p = Math.max(2, act - 1); p <= Math.min(tot - 1, act + 1); p++) pags.push(p)
  if (act < tot - 2) pags.push('...')
  pags.push(tot)
  return pags
}

const totalPaginasCuentas = computed(() => Math.max(1, Math.ceil(cuentas.value.length / POR_PAGINA)))
const cuentasEnPagina     = computed(() => cuentas.value.slice((paginaCuentas.value - 1) * POR_PAGINA, paginaCuentas.value * POR_PAGINA))
const botonesCuentas      = computed(() => paginasBotones(paginaCuentas.value, totalPaginasCuentas.value))

const alerta = ref(null)
let alertaTimer = null

function mostrarAlerta(tipo, mensaje) {
  clearTimeout(alertaTimer)
  alerta.value = { tipo, mensaje }
  alertaTimer = setTimeout(() => { alerta.value = null }, 4000)
}

async function buscar() {
  cargando.value = true
  paginaCuentas.value = 1
  try {
    const res = await api.buscar({
      idUsuarioFinal: filtros.value.idUsuarioFinal || undefined,
      idMoneda:       filtros.value.idMoneda || undefined,
      estado:         filtros.value.estado || undefined
    })
    cuentas.value = res.Cuentas ?? []
    total.value   = res.Total ?? 0
  } catch (e) {
    mostrarAlerta('danger', e.response?.data?.error ?? 'Error al cargar cuentas')
  } finally {
    cargando.value = false
  }
}

onMounted(async () => {
  try {
    const [pC, pD] = await Promise.all([
      dameParametro('LIMITEBUSCARCUENTAS'),
      dameParametro('LIMITEHISTORIALBALANCE')
    ])
    const vC = parseInt(pC.Valor)
    const vD = parseInt(pD.Valor)
    if (!isNaN(vC) && vC > 0) limiteCuentas.value = vC
    if (!isNaN(vD) && vD > 0) limiteDetalle.value = vD
  } catch { /* usa los defaults */ }
  buscar()
})

// Modal p crear
const crearModalEl = ref(null)
const nuevaCuenta  = ref({ idUsuarioFinal: '', idMoneda: '', fecha: hoy() })
const creando      = ref(false)
let bsCrearModal   = null

onMounted(() => {
  bsCrearModal = new Modal(crearModalEl.value)
  crearModalEl.value.addEventListener('hidden.bs.modal', () => {
    nuevaCuenta.value = { idUsuarioFinal: '', idMoneda: '', fecha: hoy() }
    creando.value = false
  })
})

onBeforeUnmount(() => {
  bsCrearModal?.dispose()
})

function abrirModalCrear() {
  bsCrearModal?.show()
}

async function crearCuenta() {
  const { idUsuarioFinal, idMoneda, fecha } = nuevaCuenta.value
  if (!idUsuarioFinal || !idMoneda || !fecha) return
  creando.value = true
  try {
    const res = await api.crear({
      IdUsuarioFinal: parseInt(idUsuarioFinal),
      IdMoneda:       parseInt(idMoneda),
      Fecha:          fecha
    })
    bsCrearModal?.hide()
    mostrarAlerta('success', res.Mensaje ?? 'Cuenta creada')
    buscar()
  } catch (e) {
    mostrarAlerta('danger', e.response?.data?.error ?? 'Error al crear cuenta')
    bsCrearModal?.hide()
  } finally {
    creando.value = false
  }
}

//Modal p confirmar
const confirmModalRef = ref(null)
const confirmConfig   = ref({ title: '', message: '', confirmLabel: '', confirmVariant: 'btn-outline-danger' })
let accionPendiente   = null

function pedirConfirmacion(config, accion) {
  confirmConfig.value = config
  accionPendiente = accion
  confirmModalRef.value.open()
}

async function onConfirmar() {
  confirmModalRef.value.close()
  if (accionPendiente) {
    await accionPendiente()
    accionPendiente = null
  }
}

function onCancelar() {
  accionPendiente = null
}

function activar(c) {
  pedirConfirmacion(
    {
      title: 'Activar cuenta',
      message: `¿Activar la cuenta del usuario <strong>${c.IdUsuarioFinal}</strong> en moneda <strong>${c.IdMoneda}</strong>?`,
      confirmLabel: 'Activar',
      confirmVariant: 'btn-primary'
    },
    async () => {
      try {
        await api.activar(c.IdUsuarioFinal, c.IdMoneda)
        mostrarAlerta('success', `Cuenta activada`)
        buscar()
      } catch (e) {
        mostrarAlerta('danger', e.response?.data?.error ?? 'Error al activar cuenta')
      }
    }
  )
}

function desactivar(c) {
  pedirConfirmacion(
    {
      title: 'Desactivar cuenta',
      message: `¿Desactivar la cuenta del usuario <strong>${c.IdUsuarioFinal}</strong> en moneda <strong>${c.IdMoneda}</strong>? Se requiere saldo cero.`,
      confirmLabel: 'Desactivar',
      confirmVariant: 'btn-outline-danger'
    },
    async () => {
      try {
        await api.desactivar(c.IdUsuarioFinal, c.IdMoneda)
        mostrarAlerta('success', `Cuenta desactivada`)
        buscar()
      } catch (e) {
        mostrarAlerta('danger', e.response?.data?.error ?? 'Error al desactivar cuenta')
      }
    }
  )
}

const _fmt = new Intl.NumberFormat('es-AR', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
function formatMonto(v) { return _fmt.format(parseFloat(v) || 0) }

function balance(c) {
  const cr = parseFloat(c.Creditos) || 0
  const db = parseFloat(c.Debitos)  || 0
  const b  = cr - db
  return (b > 0 ? '+' : '') + _fmt.format(b)
}

function balanceClass(c) {
  const b = (parseFloat(c.Creditos) || 0) - (parseFloat(c.Debitos) || 0)
  return b > 0 ? 'balance-positivo' : b < 0 ? 'balance-negativo' : ''
}

function hoy() {
  return new Date().toISOString().slice(0, 10)
}

function formatFecha(f) {
  return f ? f.slice(0, 10) : '—'
}

const ESTADO_LABEL = { A: 'Activa', I: 'Inactiva' }
const ESTADO_CLASS = { A: 'badge-activo', I: 'badge-inactivo' }

//  Detalle de cuenta 
const cuentaSeleccionada = ref(null)
const tabActiva          = ref('transferencias')

function verDetalle(c) {
  cuentaSeleccionada.value = c
  tabActiva.value = 'transferencias'
  cargarTransfs()
}

function volverALista() {
  cuentaSeleccionada.value  = null
  detalleTransfs.value      = []
  detalleTransfsTotal.value = 0
  historial.value           = []
  historialTotal.value      = 0
  paginaTransfs.value       = 1
  paginaHistorial.value     = 1
  filtrosTransfs.value      = { timestampMin: '', timestampMax: '', incluyeRevertidas: false }
  filtrosHistorial.value    = { timestampMin: '', timestampMax: '' }
}

// Sub-tab: Transferencias de la cuenta
const detalleTransfs      = ref([])
const detalleTransfsTotal = ref(0)
const cargandoTransfs     = ref(false)
const filtrosTransfs      = ref({ timestampMin: '', timestampMax: '', incluyeRevertidas: false })

const totalPaginasTransfs = computed(() => Math.max(1, Math.ceil(detalleTransfs.value.length / POR_PAGINA)))
const transfsEnPagina     = computed(() => detalleTransfs.value.slice((paginaTransfs.value - 1) * POR_PAGINA, paginaTransfs.value * POR_PAGINA))
const botonesTransfs      = computed(() => paginasBotones(paginaTransfs.value, totalPaginasTransfs.value))

async function cargarTransfs() {
  if (!cuentaSeleccionada.value) return
  cargandoTransfs.value = true
  paginaTransfs.value = 1
  try {
    const res = await api.dameTransferencias(
      cuentaSeleccionada.value.IdUsuarioFinal,
      cuentaSeleccionada.value.IdMoneda,
      filtrosTransfs.value
    )
    detalleTransfs.value      = res.Transferencias ?? []
    detalleTransfsTotal.value = res.Total ?? 0
  } catch (e) {
    mostrarAlerta('danger', e.response?.data?.error ?? 'Error al cargar transferencias')
  } finally {
    cargandoTransfs.value = false
  }
}

// Sub-tab: Historial de balances
const historial         = ref([])
const historialTotal    = ref(0)
const cargandoHistorial = ref(false)
const filtrosHistorial  = ref({ timestampMin: '', timestampMax: '' })

const totalPaginasHistorial = computed(() => Math.max(1, Math.ceil(historial.value.length / POR_PAGINA)))
const historialEnPagina     = computed(() => historial.value.slice((paginaHistorial.value - 1) * POR_PAGINA, paginaHistorial.value * POR_PAGINA))
const botonesHistorial      = computed(() => paginasBotones(paginaHistorial.value, totalPaginasHistorial.value))

async function cargarHistorial() {
  if (!cuentaSeleccionada.value) return
  cargandoHistorial.value = true
  paginaHistorial.value = 1
  try {
    const res = await api.dameHistorial(
      cuentaSeleccionada.value.IdUsuarioFinal,
      cuentaSeleccionada.value.IdMoneda,
      filtrosHistorial.value
    )
    historial.value      = res.Historial ?? []
    historialTotal.value = res.Total ?? 0
  } catch (e) {
    mostrarAlerta('danger', e.response?.data?.error ?? 'Error al cargar historial')
  } finally {
    cargandoHistorial.value = false
  }
}

function cambiarTab(tab) {
  tabActiva.value = tab
  if (tab === 'transferencias' && detalleTransfs.value.length === 0) cargarTransfs()
  if (tab === 'historial'      && historial.value.length === 0)      cargarHistorial()
}

// Helpers detalle
function formatTimestamp(ts) {
  // Para strings ISO (FechaProceso)
  return ts ? String(ts).slice(0, 19).replace('T', ' ') : '—'
}


const TIPO_LABEL = { I: 'Ingreso', E: 'Egreso', R: 'Reversión' }
const TIPO_CLASS  = { I: 'tipo-ingreso', E: 'tipo-egreso', R: 'tipo-reversion' }
const TTRANS_ESTADO_LABEL = { F: 'Finalizada', R: 'Revertida' }
const TTRANS_ESTADO_CLASS = { F: 'badge-activo', R: 'badge-pendiente' }
</script>

<template>
  <div>

    <div v-if="alerta" :class="`alert alert-${alerta.tipo} alert-dismissible mb-4`" role="alert">
      {{ alerta.mensaje }}
      <button type="button" class="btn-close" @click="alerta = null"></button>
    </div>

    <template v-if="!cuentaSeleccionada">

    <div class="section-header mb-4">
      <h1 class="page-title">Cuentas</h1>
    </div>

    <div class="card mb-3">
      <div class="card-body py-3">
        <div class="d-flex align-items-end justify-content-between gap-3 flex-wrap">
          <form @submit.prevent="buscar" class="d-flex align-items-end gap-3 flex-wrap">
            <div>
              <label class="form-label">Usuario</label>
              <input
                v-model="filtros.idUsuarioFinal"
                type="number"
                min="1"
                class="form-control"
                style="width: 120px"
                placeholder="ID..."
              />
            </div>
            <div>
              <label class="form-label">Moneda</label>
              <input
                v-model="filtros.idMoneda"
                type="number"
                min="1"
                class="form-control"
                style="width: 100px"
                placeholder="ID..."
              />
            </div>
            <div>
              <label class="form-label">Estado</label>
              <select v-model="filtros.estado" class="form-select" style="width: 130px">
                <option value="">Todos</option>
                <option value="A">Activa</option>
                <option value="I">Inactiva</option>
              </select>
            </div>
            <div>
              <button type="submit" class="btn btn-outline-primary" :disabled="cargando">
                Buscar
              </button>
            </div>
          </form>
          <div>
            <button type="button" class="btn btn-primary" @click="abrirModalCrear">
              + Crear cuenta
            </button>
          </div>
        </div>
      </div>
    </div>

    <div class="card">
      <div class="card-body p-0">

        <div v-if="cargando" class="loading-state">
          <div class="spinner-border spinner-border-sm" role="status" aria-hidden="true"></div>
          <span>Cargando...</span>
        </div>

        <div v-else-if="cuentas.length === 0" class="empty-state">
          <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <rect x="2" y="5" width="20" height="14" rx="2"/><line x1="2" y1="10" x2="22" y2="10"/>
          </svg>
          <p>No se encontraron cuentas</p>
        </div>

        <div v-else class="table-responsive">
          <div class="total-row d-flex align-items-center justify-content-between">
            <span>{{ total }} resultado{{ total !== 1 ? 's' : '' }}</span>
            <span v-if="total === limiteCuentas" class="limite-aviso">
              Se alcanzó el límite de búsqueda. Acotar los filtros para resultados más precisos.
            </span>
          </div>
          <table class="table mb-0">
            <thead>
              <tr>
                <th>IdUsuario</th>
                <th>Moneda</th>
                <th class="text-end">Créditos</th>
                <th class="text-end">Débitos</th>
                <th class="text-end">Balance</th>
                <th>Fecha Alta</th>
                <th>Fecha Proceso</th>
                <th>Estado</th>
                <th style="width: 1%; white-space: nowrap">Acciones</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="c in cuentasEnPagina" :key="`${c.IdUsuarioFinal}-${c.IdMoneda}`">
                <td class="cell-id">{{ c.IdUsuarioFinal }}</td>
                <td class="cell-id">{{ c.IdMoneda }}</td>
                <td class="cell-monto">{{ formatMonto(c.Creditos) }}</td>
                <td class="cell-monto">{{ formatMonto(c.Debitos) }}</td>
                <td :class="`cell-monto ${balanceClass(c)}`">{{ balance(c) }}</td>
                <td>{{ formatFecha(c.Fecha) }}</td>
                <td class="cell-id">{{ formatTimestamp(c.FechaProceso) }}</td>
                <td>
                  <span v-if="c.IdUsuarioFinal === 0" class="badge badge-empresa">Empresa</span>
                  <span v-else :class="`badge ${ESTADO_CLASS[c.Estado] ?? ''}`">
                    {{ ESTADO_LABEL[c.Estado] ?? c.Estado }}
                  </span>
                </td>
                <td style="white-space: nowrap">
                  <div class="d-flex gap-1">
                    <button
                      class="btn btn-outline-primary btn-icon"
                      title="Ver detalle"
                      @click="verDetalle(c)"
                    >
                      <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                        <circle cx="12" cy="12" r="10"/><line x1="12" y1="12" x2="12" y2="16"/><line x1="12" y1="8" x2="12.01" y2="8"/>
                      </svg>
                    </button>
                    <button
                      v-if="c.IdUsuarioFinal !== 0 && c.Estado === 'I'"
                      class="btn btn-outline-secondary btn-icon"
                      title="Activar"
                      @click="activar(c)"
                    >
                      <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                        <path d="M9 12l2 2 4-4"/><circle cx="12" cy="12" r="10"/>
                      </svg>
                    </button>
                    <button
                      v-if="c.IdUsuarioFinal !== 0 && c.Estado === 'A'"
                      class="btn btn-outline-secondary btn-icon"
                      title="Desactivar"
                      @click="desactivar(c)"
                    >
                      <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                        <circle cx="12" cy="12" r="10"/><line x1="8" y1="12" x2="16" y2="12"/>
                      </svg>
                    </button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
          <nav v-if="totalPaginasCuentas > 1" class="d-flex justify-content-center py-3">
            <ul class="pagination pagination-sm mb-0">
              <li class="page-item" :class="{ disabled: paginaCuentas === 1 }">
                <button class="page-link" @click="paginaCuentas--">&lsaquo;</button>
              </li>
              <li v-for="p in botonesCuentas" :key="String(p)" class="page-item" :class="{ active: p === paginaCuentas, disabled: p === '...' }">
                <button class="page-link" @click="typeof p === 'number' && (paginaCuentas = p)">{{ p }}</button>
              </li>
              <li class="page-item" :class="{ disabled: paginaCuentas === totalPaginasCuentas }">
                <button class="page-link" @click="paginaCuentas++">&rsaquo;</button>
              </li>
            </ul>
          </nav>
        </div>

      </div>
    </div>

    <div ref="crearModalEl" class="modal fade" tabindex="-1" aria-hidden="true">
      <div class="modal-dialog modal-dialog-centered" style="max-width: 380px">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">Crear cuenta</h5>
            <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Cerrar"></button>
          </div>
          <form @submit.prevent="crearCuenta">
            <div class="modal-body">
              <div class="mb-3">
                <label class="form-label">ID de usuario</label>
                <input
                  v-model="nuevaCuenta.idUsuarioFinal"
                  type="number"
                  min="1"
                  class="form-control"
                  placeholder="ej: 12345"
                  :disabled="creando"
                  required
                  autofocus
                />
              </div>
              <div class="mb-3">
                <label class="form-label">ID de moneda</label>
                <input
                  v-model="nuevaCuenta.idMoneda"
                  type="number"
                  min="1"
                  class="form-control"
                  placeholder="ej: 1"
                  :disabled="creando"
                  required
                />
              </div>
              <div class="mb-1">
                <label class="form-label">Fecha de alta</label>
                <input
                  v-model="nuevaCuenta.fecha"
                  type="date"
                  class="form-control"
                  :disabled="creando"
                  required
                />
              </div>
            </div>
            <div class="modal-footer">
              <button type="button" class="btn btn-outline-secondary btn-sm" data-bs-dismiss="modal">
                Cancelar
              </button>
              <button
                type="submit"
                class="btn btn-primary btn-sm"
                :disabled="creando || !nuevaCuenta.idUsuarioFinal || !nuevaCuenta.idMoneda || !nuevaCuenta.fecha"
              >
                <span v-if="creando" class="spinner-border spinner-border-sm me-1" role="status" aria-hidden="true"></span>
                Crear
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>

    <ConfirmModal
      ref="confirmModalRef"
      :title="confirmConfig.title"
      :message="confirmConfig.message"
      :confirm-label="confirmConfig.confirmLabel"
      :confirm-variant="confirmConfig.confirmVariant"
      @confirm="onConfirmar"
      @cancel="onCancelar"
    />

    </template>

    <template v-if="cuentaSeleccionada">

      <div class="section-header mb-4">
        <div class="d-flex align-items-center gap-3">
          <button class="btn btn-outline-secondary btn-sm d-flex align-items-center gap-1" @click="volverALista">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <polyline points="15 18 9 12 15 6"/>
            </svg>
            Volver
          </button>
          <h1 class="page-title mb-0">Cuenta</h1>
        </div>
      </div>

      <div class="card mb-4">
        <div class="card-body">
          <div class="cuenta-info-grid">
            <div class="info-item">
              <span class="info-label">Usuario</span>
              <span class="info-value mono">{{ cuentaSeleccionada.IdUsuarioFinal }}</span>
            </div>
            <div class="info-item">
              <span class="info-label">Moneda</span>
              <span class="info-value mono">{{ cuentaSeleccionada.IdMoneda }}</span>
            </div>
            <div class="info-item">
              <span class="info-label">Estado</span>
              <span :class="`badge ${ESTADO_CLASS[cuentaSeleccionada.Estado] ?? ''}`">
                {{ ESTADO_LABEL[cuentaSeleccionada.Estado] ?? cuentaSeleccionada.Estado }}
              </span>
            </div>
            <div class="info-item">
              <span class="info-label">Fecha alta</span>
              <span class="info-value">{{ formatFecha(cuentaSeleccionada.Fecha) }}</span>
            </div>
            <div class="info-item">
              <span class="info-label">Créditos</span>
              <span class="info-value mono" style="color: var(--success)">{{ formatMonto(cuentaSeleccionada.Creditos) }}</span>
            </div>
            <div class="info-item">
              <span class="info-label">Débitos</span>
              <span class="info-value mono" style="color: var(--error)">{{ formatMonto(cuentaSeleccionada.Debitos) }}</span>
            </div>
            <div class="info-item">
              <span class="info-label">Balance</span>
              <span :class="`info-value mono ${balanceClass(cuentaSeleccionada)}`" style="font-size: 1.125rem; font-weight: 700">
                {{ balance(cuentaSeleccionada) }}
              </span>
            </div>
          </div>
        </div>
      </div>

      <ul class="nav nav-tabs mb-0">
        <li class="nav-item">
          <a
            class="nav-link"
            :class="{ active: tabActiva === 'transferencias' }"
            href="#"
            @click.prevent="cambiarTab('transferencias')"
          >Transferencias</a>
        </li>
        <li class="nav-item">
          <a
            class="nav-link"
            :class="{ active: tabActiva === 'historial' }"
            href="#"
            @click.prevent="cambiarTab('historial')"
          >Historial de balances</a>
        </li>
      </ul>

      <div v-if="tabActiva === 'transferencias'" class="card" style="border-top-left-radius: 0">
        <div class="card-body py-3 border-bottom">
          <form @submit.prevent="cargarTransfs" class="d-flex align-items-end gap-3 flex-wrap">
            <div>
              <label class="form-label">Desde</label>
              <input v-model="filtrosTransfs.timestampMin" type="datetime-local" class="form-control" style="width: 195px" />
            </div>
            <div>
              <label class="form-label">Hasta</label>
              <input v-model="filtrosTransfs.timestampMax" type="datetime-local" class="form-control" style="width: 195px" />
            </div>
            <div style="padding-bottom: 6px" class="d-flex align-items-end">
              <div class="d-flex align-items-center gap-2">
                <input
                  id="incluyeRevertidas"
                  v-model="filtrosTransfs.incluyeRevertidas"
                  type="checkbox"
                  class="form-check-input mt-0"
                />
                <label for="incluyeRevertidas" class="form-label mb-0" style="white-space: nowrap">Incl. revertidas</label>
              </div>
            </div>
            <div>
              <button type="submit" class="btn btn-outline-primary" :disabled="cargandoTransfs">Buscar</button>
            </div>
          </form>
        </div>

        <div class="card-body p-0">
          <div v-if="cargandoTransfs" class="loading-state">
            <div class="spinner-border spinner-border-sm" role="status" aria-hidden="true"></div>
            <span>Cargando...</span>
          </div>
          <div v-else-if="detalleTransfs.length === 0" class="empty-state">
            <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <polyline points="17 1 21 5 17 9"/><path d="M3 11V9a4 4 0 0 1 4-4h14"/><polyline points="7 23 3 19 7 15"/><path d="M21 13v2a4 4 0 0 1-4 4H3"/>
            </svg>
            <p>No se encontraron transferencias</p>
          </div>
          <div v-else class="table-responsive">
            <div class="total-row d-flex align-items-center justify-content-between">
              <span>{{ detalleTransfsTotal }} resultado{{ detalleTransfsTotal !== 1 ? 's' : '' }}</span>
              <span v-if="detalleTransfsTotal === limiteDetalle" class="limite-aviso">
                Se alcanzó el límite de búsqueda. Acotar los filtros para resultados más precisos.
              </span>
            </div>
            <table class="table mb-0">
              <thead>
                <tr>
                  <th>ID</th>
                  <th>Tipo</th>
                  <th>Categoría</th>
                  <th class="text-end">Monto</th>
                  <th>Estado</th>
                  <th>Fecha</th>
                  <th>Procesada</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="t in transfsEnPagina" :key="t.IdTransferencia">
                  <td class="cell-mono" style="font-size: 0.75rem; max-width: 180px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap">
                    {{ t.IdTransferencia }}
                  </td>
                  <td>
                    <span v-if="t.Tipo" :class="`tipo-badge ${TIPO_CLASS[t.Tipo] ?? ''}`">
                      {{ TIPO_LABEL[t.Tipo] ?? t.Tipo }}
                    </span>
                    <span v-else class="text-muted">—</span>
                  </td>
                  <td class="cell-id">{{ t.Categoria }}</td>
                  <td class="cell-monto">{{ formatMonto(t.Monto) }}</td>
                  <td>
                    <span :class="`badge ${TTRANS_ESTADO_CLASS[t.Estado] ?? ''}`">
                      {{ TTRANS_ESTADO_LABEL[t.Estado] ?? t.Estado }}
                    </span>
                  </td>
                  <td>{{ formatFecha(t.Fecha) }}</td>
                  <td style="font-size: 0.8rem; color: var(--text-secondary)">{{ formatTimestamp(t.FechaProceso) }}</td>
                </tr>
              </tbody>
            </table>
            <nav v-if="totalPaginasTransfs > 1" class="d-flex justify-content-center py-3">
              <ul class="pagination pagination-sm mb-0">
                <li class="page-item" :class="{ disabled: paginaTransfs === 1 }">
                  <button class="page-link" @click="paginaTransfs--">&lsaquo;</button>
                </li>
                <li v-for="p in botonesTransfs" :key="String(p)" class="page-item" :class="{ active: p === paginaTransfs, disabled: p === '...' }">
                  <button class="page-link" @click="typeof p === 'number' && (paginaTransfs = p)">{{ p }}</button>
                </li>
                <li class="page-item" :class="{ disabled: paginaTransfs === totalPaginasTransfs }">
                  <button class="page-link" @click="paginaTransfs++">&rsaquo;</button>
                </li>
              </ul>
            </nav>
          </div>
        </div>
      </div>

      <div v-if="tabActiva === 'historial'" class="card" style="border-top-left-radius: 0">
        <div class="card-body py-3 border-bottom">
          <form @submit.prevent="cargarHistorial" class="d-flex align-items-end gap-3 flex-wrap">
            <div>
              <label class="form-label">Desde</label>
              <input v-model="filtrosHistorial.timestampMin" type="datetime-local" class="form-control" style="width: 195px" />
            </div>
            <div>
              <label class="form-label">Hasta</label>
              <input v-model="filtrosHistorial.timestampMax" type="datetime-local" class="form-control" style="width: 195px" />
            </div>
            <div>
              <button type="submit" class="btn btn-outline-primary" :disabled="cargandoHistorial">Buscar</button>
            </div>
          </form>
        </div>

        <div class="card-body p-0">
          <div v-if="cargandoHistorial" class="loading-state">
            <div class="spinner-border spinner-border-sm" role="status" aria-hidden="true"></div>
            <span>Cargando...</span>
          </div>
          <div v-else-if="historial.length === 0" class="empty-state">
            <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <polyline points="22 12 18 12 15 21 9 3 6 12 2 12"/>
            </svg>
            <p>No se encontró historial de balances</p>
          </div>
          <div v-else class="table-responsive">
            <div class="total-row d-flex align-items-center justify-content-between">
              <span>{{ historialTotal }} registro{{ historialTotal !== 1 ? 's' : '' }}</span>
              <span v-if="historialTotal === limiteDetalle" class="limite-aviso">
                Se alcanzó el límite de búsqueda. Acotar los filtros para resultados más precisos.
              </span>
            </div>
            <table class="table mb-0">
              <thead>
                <tr>
                  <th>Fecha proceso</th>
                  <th class="text-end">Créditos</th>
                  <th class="text-end">Débitos</th>
                  <th class="text-end">Balance</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="(h, i) in historialEnPagina" :key="i">
                  <td style="font-family: var(--font-mono); font-size: 0.8rem; color: var(--text-secondary)">
                    {{ formatTimestamp(h.Fecha) }}
                  </td>
                  <td class="cell-monto" style="color: var(--success)">{{ formatMonto(h.Creditos) }}</td>
                  <td class="cell-monto" style="color: var(--error)">{{ formatMonto(h.Debitos) }}</td>
                  <td class="cell-monto" style="font-weight: 600">{{ formatMonto(h.Balance) }}</td>
                </tr>
              </tbody>
            </table>
            <nav v-if="totalPaginasHistorial > 1" class="d-flex justify-content-center py-3">
              <ul class="pagination pagination-sm mb-0">
                <li class="page-item" :class="{ disabled: paginaHistorial === 1 }">
                  <button class="page-link" @click="paginaHistorial--">&lsaquo;</button>
                </li>
                <li v-for="p in botonesHistorial" :key="String(p)" class="page-item" :class="{ active: p === paginaHistorial, disabled: p === '...' }">
                  <button class="page-link" @click="typeof p === 'number' && (paginaHistorial = p)">{{ p }}</button>
                </li>
                <li class="page-item" :class="{ disabled: paginaHistorial === totalPaginasHistorial }">
                  <button class="page-link" @click="paginaHistorial++">&rsaquo;</button>
                </li>
              </ul>
            </nav>
          </div>
        </div>
      </div>

    </template>

  </div>
</template>

<style scoped>
.btn-icon {
  padding: 0.5rem 0.625rem;
  line-height: 1;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

:deep(tbody td) {
  padding-top: 0.875rem;
  padding-bottom: 0.875rem;
}

.total-row {
  font-family: var(--font-mono);
  font-size: 0.6875rem;
  color: var(--text-tertiary);
  padding: 0.5rem 0.75rem;
  border-bottom: 1px solid var(--border);
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.limite-aviso {
  font-family: var(--font-mono);
  font-size: 0.6875rem;
  color: #92400E;
  background: #FEF3C7;
  padding: 0.2em 0.6em;
  border-radius: 4px;
  text-transform: none;
  letter-spacing: 0;
}

.badge-empresa {
  background: #EDE9FE;
  color: #5B21B6;
  font-weight: 600;
}

.balance-positivo {
  color: var(--success);
  font-weight: 600;
}

.balance-negativo {
  color: var(--error);
  font-weight: 600;
}

.cuenta-info-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(140px, 1fr));
  gap: 1.25rem 2rem;
}

.info-item {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.info-label {
  font-family: var(--font-mono);
  font-size: 0.625rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: var(--text-tertiary);
}

.info-value {
  font-size: 0.9375rem;
  color: var(--text-primary);
}

.info-value.mono {
  font-family: var(--font-mono);
}

.tipo-badge {
  font-family: var(--font-mono);
  font-size: 0.6875rem;
  font-weight: 600;
  padding: 0.2em 0.6em;
  border-radius: 4px;
  letter-spacing: 0.02em;
  white-space: nowrap;
}

.tipo-ingreso   { background: #D1FAE5; color: #065F46; }
.tipo-egreso    { background: #FEE2E2; color: #991B1B; }
.tipo-reversion { background: #E0F2FE; color: #0C4A6E; }
</style>
