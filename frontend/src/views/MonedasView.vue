<script setup>
import { ref, onMounted, onBeforeUnmount } from 'vue'
import { Modal } from 'bootstrap'
import ConfirmModal from '../components/ConfirmModal.vue'
import * as api from '../api/monedas'

const monedas = ref([])
const cargando = ref(false)
const filtros  = ref({ incluyeInactivos: false })

const alerta = ref(null)
let alertaTimer = null

function mostrarAlerta(tipo, mensaje) {
  clearTimeout(alertaTimer)
  alerta.value = { tipo, mensaje }
  alertaTimer = setTimeout(() => { alerta.value = null }, 4000)
}

async function listar() {
  cargando.value = true
  try {
    monedas.value = await api.listar(filtros.value.incluyeInactivos ? 'S' : 'N')
  } catch (e) {
    mostrarAlerta('danger', e.response?.data?.error ?? 'Error al cargar monedas')
  } finally {
    cargando.value = false
  }
}

onMounted(listar)

// Modal p crear
const crearModalEl  = ref(null)
const nuevoIdMoneda = ref('')
const creando       = ref(false)
let bsCrearModal    = null

onMounted(() => {
  bsCrearModal = new Modal(crearModalEl.value)
  crearModalEl.value.addEventListener('hidden.bs.modal', () => {
    nuevoIdMoneda.value = ''
    creando.value = false
  })
})

onBeforeUnmount(() => {
  bsCrearModal?.dispose()
})

function abrirModalCrear() {
  bsCrearModal?.show()
}

async function crearMoneda() {
  const id = parseInt(nuevoIdMoneda.value)
  if (!id || id <= 0) return
  creando.value = true
  try {
    await api.crear(id)
    bsCrearModal?.hide()
    mostrarAlerta('success', `Moneda ${id} creada correctamente`)
    listar()
  } catch (e) {
    mostrarAlerta('danger', e.response?.data?.error ?? 'Error al crear moneda')
    bsCrearModal?.hide()
  } finally {
    creando.value = false
  }
}

//Modal p confirmar acción
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

function activar(m) {
  pedirConfirmacion(
    {
      title: 'Activar moneda',
      message: `¿Activar la moneda <strong>${m.IdMoneda}</strong>?`,
      confirmLabel: 'Activar',
      confirmVariant: 'btn-primary'
    },
    async () => {
      try {
        await api.activar(m.IdMoneda)
        mostrarAlerta('success', `Moneda ${m.IdMoneda} activada`)
        listar()
      } catch (e) {
        mostrarAlerta('danger', e.response?.data?.error ?? 'Error al activar moneda')
      }
    }
  )
}

function desactivar(m) {
  pedirConfirmacion(
    {
      title: 'Desactivar moneda',
      message: `¿Desactivar la moneda <strong>${m.IdMoneda}</strong>?`,
      confirmLabel: 'Desactivar',
      confirmVariant: 'btn-outline-danger'
    },
    async () => {
      try {
        await api.desactivar(m.IdMoneda)
        mostrarAlerta('success', `Moneda ${m.IdMoneda} desactivada`)
        listar()
      } catch (e) {
        mostrarAlerta('danger', e.response?.data?.error ?? 'Error al desactivar moneda')
      }
    }
  )
}

function borrar(m) {
  pedirConfirmacion(
    {
      title: 'Borrar moneda',
      message: `¿Borrar permanentemente la moneda <strong>${m.IdMoneda}</strong>? Esta acción no se puede deshacer.`,
      confirmLabel: 'Borrar',
      confirmVariant: 'btn-outline-danger'
    },
    async () => {
      try {
        await api.borrar(m.IdMoneda)
        mostrarAlerta('success', `Moneda ${m.IdMoneda} borrada`)
        listar()
      } catch (e) {
        mostrarAlerta('danger', e.response?.data?.error ?? 'Error al borrar moneda')
      }
    }
  )
}

// Helpers
function formatFecha(f) {
  return f ? f.slice(0, 10) : '—'
}

const ESTADO_LABEL = { A: 'Activa', I: 'Inactiva' }
const ESTADO_CLASS = { A: 'badge-activo', I: 'badge-inactivo' }
</script>

<template>
  <div>

    <div class="section-header mb-4">
      <h1 class="page-title">Monedas</h1>
    </div>

    <div v-if="alerta" :class="`alert alert-${alerta.tipo} alert-dismissible mb-4`" role="alert">
      {{ alerta.mensaje }}
      <button type="button" class="btn-close" @click="alerta = null"></button>
    </div>

    <div class="card mb-3">
      <div class="card-body py-3">
        <div class="d-flex align-items-end justify-content-between gap-3">
          <form @submit.prevent="listar" class="d-flex align-items-center gap-3">
            <div class="d-flex align-items-center gap-2">
              <input
                id="incluyeInactivos"
                v-model="filtros.incluyeInactivos"
                type="checkbox"
                class="form-check-input mt-0"
              />
              <label for="incluyeInactivos" class="form-label mb-0">Incluir inactivas</label>
            </div>
            <div>
              <button type="submit" class="btn btn-outline-primary btn-sm" :disabled="cargando">
                Actualizar
              </button>
            </div>
          </form>
          <div>
            <button type="button" class="btn btn-primary btn-sm" @click="abrirModalCrear">
              + Crear moneda
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

        <div v-else-if="monedas.length === 0" class="empty-state">
          <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/>
          </svg>
          <p>No se encontraron monedas</p>
        </div>

        <div v-else class="table-responsive">
          <table class="table mb-0">
            <thead>
              <tr>
                <th>ID</th>
                <th>Cuenta empresa</th>
                <th>Estado</th>
                <th>Fecha Alta</th>
                <th style="width: 1%; white-space: nowrap">Acciones</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="m in monedas" :key="m.IdMoneda">
                <td class="cell-id">{{ m.IdMoneda }}</td>
                <td class="cell-mono" style="font-size: 0.75rem; color: var(--text-secondary)">{{ m.IdCuentaEmpresa }}</td>
                <td>
                  <span :class="`badge ${ESTADO_CLASS[m.Estado] ?? ''}`">
                    {{ ESTADO_LABEL[m.Estado] ?? m.Estado }}
                  </span>
                </td>
                <td>{{ formatFecha(m.FechaAlta) }}</td>
                <td style="white-space: nowrap">
                  <div class="d-flex gap-1">
                    <button
                      v-if="m.Estado === 'I'"
                      class="btn btn-outline-primary btn-icon"
                      title="Activar"
                      @click="activar(m)"
                    >
                      <svg width="17" height="17" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                        <path d="M9 12l2 2 4-4"/><circle cx="12" cy="12" r="10"/>
                      </svg>
                    </button>
                    <button
                      v-if="m.Estado === 'A'"
                      class="btn btn-outline-secondary btn-icon"
                      title="Desactivar"
                      @click="desactivar(m)"
                    >
                      <svg width="17" height="17" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                        <circle cx="12" cy="12" r="10"/><line x1="8" y1="12" x2="16" y2="12"/>
                      </svg>
                    </button>
                    <button
                      class="btn btn-outline-danger btn-icon"
                      title="Borrar"
                      @click="borrar(m)"
                    >
                      <svg width="17" height="17" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                        <polyline points="3 6 5 6 21 6"/><path d="M19 6l-1 14H6L5 6"/><path d="M10 11v6"/><path d="M14 11v6"/><path d="M9 6V4h6v2"/>
                      </svg>
                    </button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

      </div>
    </div>

    <div ref="crearModalEl" class="modal fade" tabindex="-1" aria-hidden="true">
      <div class="modal-dialog modal-dialog-centered" style="max-width: 360px">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">Crear moneda</h5>
            <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Cerrar"></button>
          </div>
          <form @submit.prevent="crearMoneda">
            <div class="modal-body">
              <div class="mb-1">
                <label class="form-label">ID de moneda</label>
                <input
                  v-model="nuevoIdMoneda"
                  type="number"
                  min="1"
                  class="form-control"
                  placeholder="ej: 2"
                  :disabled="creando"
                  required
                  autofocus
                />
                <div class="form-text mt-2">
                  Identificador numérico único para el ledger de TigerBeetle.
                </div>
              </div>
            </div>
            <div class="modal-footer">
              <button type="button" class="btn btn-outline-secondary btn-sm" data-bs-dismiss="modal">
                Cancelar
              </button>
              <button
                type="submit"
                class="btn btn-primary btn-sm"
                :disabled="creando || !nuevoIdMoneda"
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

  </div>
</template>

<style scoped>
.btn-icon {
  padding: 0.375rem 0.5rem;
  line-height: 1;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

:deep(tbody td) {
  padding-top: 0.875rem;
  padding-bottom: 0.875rem;
}
</style>
