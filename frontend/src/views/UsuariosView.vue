<script setup>
import { ref, onMounted } from 'vue'
import ConfirmModal from '../components/ConfirmModal.vue'
import { useAlert } from '../composables/useAlert'
import { useModal } from '../composables/useModal'
import { useConfirmModal } from '../composables/useConfirmModal'
import { formatFecha, ESTADO_USUARIO_LABEL as ESTADO_LABEL, ESTADO_USUARIO_CLASS as ESTADO_CLASS, ROL_USUARIO_LABEL } from '../utils/formatters'
import * as api from '../api/usuarios'

const { alerta, mostrarAlerta } = useAlert()

const alertaPwd = ref(null)

const usuarios = ref([])
const cargando = ref(false)
const filtros  = ref({ cadena: '', incluyeInactivos: false })

async function buscar() {
  cargando.value = true
  try {
    usuarios.value = await api.buscar(
      filtros.value.cadena,
      filtros.value.incluyeInactivos ? 'S' : 'N'
    )
  } catch (e) {
    mostrarAlerta('danger', e.response?.data?.error ?? 'Error al cargar usuarios')
  } finally {
    cargando.value = false
  }
}

onMounted(buscar)

// Modal p crear usuario
const crearModalEl    = ref(null)
const nuevoUsuario    = ref('')
const creandoUsuario  = ref(false)
const resultadoCreacion = ref(null)

const crearModal = useModal(crearModalEl, () => {
  nuevoUsuario.value   = ''
  creandoUsuario.value = false
})

function abrirModalCrear() {
  resultadoCreacion.value = null
  crearModal.show()
}

async function crearUsuario() {
  if (!nuevoUsuario.value.trim()) return
  creandoUsuario.value = true
  try {
    const res = await api.crear(nuevoUsuario.value.trim())
    resultadoCreacion.value = { ...res, usuario: nuevoUsuario.value.trim() }
    buscar()
    crearModal.hide()
    mostrarAlerta('success', `Usuario "${res.Id} — ${nuevoUsuario.value.trim()}" creado. Contraseña temporal: ${res.PasswordTemporal}`)
  } catch (e) {
    mostrarAlerta('danger', e.response?.data?.error ?? 'Error al crear usuario')
    crearModal.hide()
  } finally {
    creandoUsuario.value = false
  }
}

// Modal p confirmar
const { confirmModalRef, confirmConfig, pedirConfirmacion, onConfirmar, onCancelar } = useConfirmModal()

function activar(u) {
  pedirConfirmacion(
    {
      title: 'Activar usuario',
      message: `¿Confirma que desea activar al usuario <strong>${u.Usuario}</strong>?`,
      confirmLabel: 'Activar',
      confirmVariant: 'btn-primary'
    },
    async () => {
      try {
        await api.activar(u.IdUsuario)
        mostrarAlerta('success', `Usuario "${u.Usuario}" activado correctamente`)
        buscar()
      } catch (e) {
        mostrarAlerta('danger', e.response?.data?.error ?? 'Error al activar usuario')
      }
    }
  )
}

function desactivar(u) {
  pedirConfirmacion(
    {
      title: 'Desactivar usuario',
      message: `¿Confirma que desea desactivar al usuario <strong>${u.Usuario}</strong>?`,
      confirmLabel: 'Desactivar',
      confirmVariant: 'btn-outline-danger'
    },
    async () => {
      try {
        await api.desactivar(u.IdUsuario)
        mostrarAlerta('success', `Usuario "${u.Usuario}" desactivado`)
        buscar()
      } catch (e) {
        mostrarAlerta('danger', e.response?.data?.error ?? 'Error al desactivar usuario')
      }
    }
  )
}

function borrar(u) {
  pedirConfirmacion(
    {
      title: 'Borrar usuario',
      message: `¿Borrar permanentemente al usuario <strong>${u.Usuario}</strong>? Esta acción no se puede deshacer.`,
      confirmLabel: 'Borrar',
      confirmVariant: 'btn-outline-danger'
    },
    async () => {
      try {
        await api.borrar(u.IdUsuario)
        mostrarAlerta('success', `Usuario "${u.Usuario}" borrado`)
        buscar()
      } catch (e) {
        mostrarAlerta('danger', e.response?.data?.error ?? 'Error al borrar usuario')
      }
    }
  )
}

async function restablecerPassword(u) {
  alertaPwd.value = null
  try {
    const res = await api.restablecerPassword(u.IdUsuario)
    alertaPwd.value = { usuario: u.Usuario, passwordTemporal: res.PasswordTemporal }
  } catch (e) {
    mostrarAlerta('danger', e.response?.data?.error ?? 'Error al restablecer contraseña')
  }
}
</script>

<template>
  <div>

    <div class="section-header mb-4">
      <h1 class="page-title">Usuarios</h1>
    </div>

    <div v-if="alerta" :class="`alert alert-${alerta.tipo} alert-dismissible mb-4`" role="alert">
      {{ alerta.mensaje }}
      <button type="button" class="btn-close" @click="alerta = null"></button>
    </div>

    <div>

      <div v-if="alertaPwd" class="alert alert-warning alert-dismissible mb-3" role="alert">
        <div class="mb-2">
          Contraseña restablecida para <strong>{{ alertaPwd.usuario }}</strong>.
          Comparta esta contraseña con el usuario — no se mostrará nuevamente.
        </div>
        <div class="d-flex align-items-center gap-2 flex-wrap">
          <span class="label-inline">Contraseña temporal:</span>
          <code class="pwd-chip">{{ alertaPwd.passwordTemporal }}</code>
        </div>
        <button type="button" class="btn-close" @click="alertaPwd = null"></button>
      </div>

      <div class="card mb-3">
        <div class="card-body py-3">
          <div class="d-flex align-items-end justify-content-between gap-3 flex-wrap">
            <form @submit.prevent="buscar" class="d-flex align-items-end gap-3 flex-wrap">
              <div>
                <label class="form-label">Buscar</label>
                <input
                  v-model="filtros.cadena"
                  type="text"
                  class="form-control"
                  style="width: 220px; max-width: 100%"
                  placeholder="Nombre de usuario..."
                />
              </div>
              <div class="d-flex align-items-center gap-2" style="min-height: 2.375rem">
                <input
                  id="incluyeInactivos"
                  v-model="filtros.incluyeInactivos"
                  type="checkbox"
                  class="form-check-input mt-0"
                />
                <label for="incluyeInactivos" class="form-label mb-0" style="font-size: 0.6875rem">Incluir inactivos</label>
              </div>
              <div>
                <button type="submit" class="btn btn-outline-primary" :disabled="cargando">
                  Buscar
                </button>
              </div>
            </form>
            <div style="padding-bottom: 2px">
              <button type="button" class="btn btn-primary" @click="abrirModalCrear">
                + Crear usuario
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

          <div v-else-if="usuarios.length === 0" class="empty-state">
            <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <circle cx="12" cy="8" r="4"/>
              <path d="M20 21a8 8 0 1 0-16 0"/>
            </svg>
            <p>No se encontraron usuarios</p>
          </div>

          <div v-else class="table-responsive">
            <table class="table mb-0">
              <thead>
                <tr>
                  <th>ID</th>
                  <th>Usuario</th>
                  <th>Rol</th>
                  <th>Estado</th>
                  <th>Fecha Alta</th>
                  <th style="width: 1%; white-space: nowrap">Acciones</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="u in usuarios" :key="u.IdUsuario">
                  <td class="cell-id">{{ u.IdUsuario }}</td>
                  <td class="cell-mono">{{ u.Usuario }}</td>
                  <td>{{ ROL_USUARIO_LABEL[u.Rol] ?? '—' }}</td>
                  <td>
                    <span :class="`badge ${ESTADO_CLASS[u.Estado] ?? ''}`">
                      {{ ESTADO_LABEL[u.Estado] ?? u.Estado }}
                    </span>
                  </td>
                  <td>{{ formatFecha(u.FechaAlta) }}</td>
                  <td style="white-space: nowrap">
                    <div class="d-flex gap-1 justify-content-end">
                      <button
                        v-if="u.Estado === 'I'"
                        class="btn btn-outline-primary btn-icon"
                        title="Activar"
                        @click="activar(u)"
                      >
                        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                          <path d="M9 12l2 2 4-4"/><circle cx="12" cy="12" r="10"/>
                        </svg>
                      </button>
                      <button
                        v-if="u.Estado === 'A'"
                        class="btn btn-outline-secondary btn-icon"
                        title="Desactivar"
                        @click="desactivar(u)"
                      >
                        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                          <circle cx="12" cy="12" r="10"/><line x1="8" y1="12" x2="16" y2="12"/>
                        </svg>
                      </button>
                      <button
                        v-if="u.Estado !== 'P'"
                        class="btn btn-outline-secondary btn-icon"
                        title="Restablecer contraseña"
                        @click="restablecerPassword(u)"
                      >
                        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                          <path d="M21 2l-2 2m-7.61 7.61a5.5 5.5 0 1 1-7.778 7.778 5.5 5.5 0 0 1 7.777-7.777zm0 0L15.5 7.5m0 0l3 3L22 7l-3-3m-3.5 3.5L19 4"/>
                        </svg>
                      </button>
                      <button
                        class="btn btn-outline-danger btn-icon"
                        title="Borrar"
                        @click="borrar(u)"
                      >
                        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
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

    </div>

    <div ref="crearModalEl" class="modal fade" tabindex="-1" aria-hidden="true">
      <div class="modal-dialog modal-dialog-centered">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">Crear usuario</h5>
            <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Cerrar"></button>
          </div>
          <form @submit.prevent="crearUsuario">
            <div class="modal-body">
              <div class="mb-1">
                <label class="form-label">Nombre de usuario</label>
                <input
                  v-model="nuevoUsuario"
                  type="text"
                  class="form-control"
                  :class="{ 'is-invalid': nuevoUsuario.trim() && nuevoUsuario.trim().length < 3 }"
                  placeholder="ej: juan.perez"
                  :disabled="creandoUsuario"
                  required
                  autofocus
                />
                <div v-if="nuevoUsuario.trim() && nuevoUsuario.trim().length < 3" class="invalid-feedback">
                  El nombre de usuario debe tener al menos 3 caracteres.
                </div>
                <div class="form-text mt-2">
                  El usuario se creará en estado Pendiente con una contraseña temporal generada automáticamente.
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
                :disabled="creandoUsuario || !nuevoUsuario.trim() || nuevoUsuario.trim().length < 3"
              >
                <span v-if="creandoUsuario" class="spinner-border spinner-border-sm me-1" role="status" aria-hidden="true"></span>
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
:deep(tbody td) {
  padding-top: 0.875rem;
  padding-bottom: 0.875rem;
}

.pwd-chip {
  font-family: var(--font-mono);
  font-size: 0.875rem;
  font-weight: 600;
  background: rgba(0, 0, 0, 0.07);
  padding: 0.2em 0.55em;
  border-radius: var(--radius-sm);
  color: inherit;
  letter-spacing: 0.03em;
}

.label-inline {
  font-family: var(--font-mono);
  font-size: 0.6875rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: inherit;
  opacity: 0.75;
}
</style>
