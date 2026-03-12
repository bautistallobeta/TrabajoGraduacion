<script setup>
import { ref, onMounted, onBeforeUnmount } from 'vue'
import { Modal } from 'bootstrap'
import * as api from '../api/parametros'

const parametros = ref([])
const cargando   = ref(false)
const filtros    = ref({ cadena: '' })

const alerta = ref(null)
let alertaTimer = null

function mostrarAlerta(tipo, mensaje) {
  clearTimeout(alertaTimer)
  alerta.value = { tipo, mensaje }
  alertaTimer = setTimeout(() => { alerta.value = null }, 4000)
}

async function buscar() {
  cargando.value = true
  try {
    parametros.value = await api.buscar(filtros.value.cadena)
  } catch (e) {
    mostrarAlerta('danger', e.response?.data?.error ?? 'Error al cargar parámetros')
  } finally {
    cargando.value = false
  }
}

onMounted(buscar)

// Modal p editar
const editModalEl  = ref(null)
const editando     = ref(null)
const nuevoValor   = ref('')
const guardando    = ref(false)
let bsEditModal    = null

onMounted(() => {
  bsEditModal = new Modal(editModalEl.value)
  editModalEl.value.addEventListener('hidden.bs.modal', () => {
    editando.value = null
    nuevoValor.value = ''
    guardando.value = false
  })
})

onBeforeUnmount(() => {
  bsEditModal?.dispose()
})

function abrirEditar(p) {
  editando.value = p
  nuevoValor.value = p.Valor
  bsEditModal?.show()
}

async function guardar() {
  if (!editando.value) return
  guardando.value = true
  try {
    await api.modificar(editando.value.Parametro, nuevoValor.value)
    bsEditModal?.hide()
    mostrarAlerta('success', `Parámetro "${editando.value.Parametro}" actualizado`)
    buscar()
  } catch (e) {
    mostrarAlerta('danger', e.response?.data?.error ?? 'Error al modificar parámetro')
    bsEditModal?.hide()
  } finally {
    guardando.value = false
  }
}
</script>

<template>
  <div>

    <div class="section-header mb-4">
      <h1 class="page-title">Parámetros</h1>
    </div>

    <div v-if="alerta" :class="`alert alert-${alerta.tipo} alert-dismissible mb-4`" role="alert">
      {{ alerta.mensaje }}
      <button type="button" class="btn-close" @click="alerta = null"></button>
    </div>

    <div class="card mb-3">
      <div class="card-body py-3">
        <form @submit.prevent="buscar" class="d-flex align-items-end gap-3">
          <div>
            <label class="form-label">Buscar</label>
            <input
              v-model="filtros.cadena"
              type="text"
              class="form-control"
              style="width: 260px"
              placeholder="Nombre del parámetro..."
            />
          </div>
          <div style="padding-bottom: 2px">
            <button type="submit" class="btn btn-outline-primary btn-sm" :disabled="cargando">
              Buscar
            </button>
          </div>
        </form>
      </div>
    </div>

    <div class="card">
      <div class="card-body p-0">

        <div v-if="cargando" class="loading-state">
          <div class="spinner-border spinner-border-sm" role="status" aria-hidden="true"></div>
          <span>Cargando...</span>
        </div>

        <div v-else-if="parametros.length === 0" class="empty-state">
          <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <circle cx="12" cy="12" r="3"/><path d="M19.07 4.93a10 10 0 0 1 0 14.14M4.93 4.93a10 10 0 0 0 0 14.14"/>
          </svg>
          <p>No se encontraron parámetros</p>
        </div>

        <div v-else class="table-responsive">
          <table class="table mb-0">
            <thead>
              <tr>
                <th>Parámetro</th>
                <th>Valor</th>
                <th>Descripción</th>
                <th>Modificable</th>
                <th style="width: 1%; white-space: nowrap">Acciones</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="p in parametros" :key="p.Parametro">
                <td class="cell-mono">{{ p.Parametro }}</td>
                <td class="cell-mono" style="color: var(--accent)">{{ p.Valor }}</td>
                <td style="color: var(--text-secondary); font-size: 0.8125rem">{{ p.Descripcion }}</td>
                <td>
                  <span v-if="p.EsModificable === 'S'" class="badge badge-activo">Sí</span>
                  <span v-else class="badge badge-inactivo">No</span>
                </td>
                <td>
                  <button
                    v-if="p.EsModificable === 'S'"
                    class="btn btn-outline-secondary btn-icon"
                    title="Editar"
                    @click="abrirEditar(p)"
                  >
                    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                      <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/>
                      <path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/>
                    </svg>
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

      </div>
    </div>

    <div ref="editModalEl" class="modal fade" tabindex="-1" aria-hidden="true">
      <div class="modal-dialog modal-dialog-centered" style="max-width: 400px">
        <div class="modal-content">
          <div class="modal-header">
            <h5 class="modal-title">{{ editando?.Parametro }}</h5>
            <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Cerrar"></button>
          </div>
          <form @submit.prevent="guardar">
            <div class="modal-body">
              <p v-if="editando?.Descripcion" class="text-secondary mb-3" style="font-size: 0.875rem">
                {{ editando.Descripcion }}
              </p>
              <div class="mb-1">
                <label class="form-label">Valor</label>
                <input
                  v-model="nuevoValor"
                  type="text"
                  class="form-control"
                  :disabled="guardando"
                  required
                  autofocus
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
                :disabled="guardando || !nuevoValor"
              >
                <span v-if="guardando" class="spinner-border spinner-border-sm me-1" role="status" aria-hidden="true"></span>
                Guardar
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>

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
</style>
