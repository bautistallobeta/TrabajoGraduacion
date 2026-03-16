<script setup>
import { ref, onMounted, onBeforeUnmount } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { Modal, Dropdown } from 'bootstrap'
import { useAuth } from '../stores/auth'
import cliente from '../api/cliente'
import * as apiUsuarios from '../api/usuarios'

const router = useRouter()
const route  = useRoute()
const { nombreUsuario, cerrarSesion } = useAuth()

const menuAbierto = ref(false)

function cerrarMenu() { menuAbierto.value = false }

const navLinks = [
  { name: 'usuarios',       label: 'Usuarios'       },
  { name: 'monedas',        label: 'Monedas'         },
  { name: 'parametros',     label: 'Parámetros'      },
  { name: 'cuentas',        label: 'Cuentas'         },
  { name: 'transferencias', label: 'Transferencias'  },
]

async function logout() {
  try {
    await cliente.post('/usuarios/logout')
  } catch {
    // Si falla igual limpiamos localmente
  } finally {
    cerrarSesion()
    router.push('/login')
  }
}

// Dropdown usuario
const userDropdownEl = ref(null)
let bsDropdown       = null

// Modal cambiar contraseña
const pwdModalEl  = ref(null)
const formPwd     = ref({ PasswordAnterior: '', PasswordNuevo: '', ConfirmarPassword: '' })
const cambiando   = ref(false)
const alertaPwd   = ref(null)
let bsPwdModal    = null

onMounted(() => {
  bsDropdown = new Dropdown(userDropdownEl.value)
  bsPwdModal = new Modal(pwdModalEl.value)
  pwdModalEl.value.addEventListener('hidden.bs.modal', () => {
    formPwd.value  = { PasswordAnterior: '', PasswordNuevo: '', ConfirmarPassword: '' }
    cambiando.value = false
    alertaPwd.value = null
  })
})

onBeforeUnmount(() => {
  bsDropdown?.dispose()
  bsPwdModal?.dispose()
})

function toggleDropdown() {
  bsDropdown?.toggle()
}

function abrirCambiarPassword() {
  bsDropdown?.hide()
  cerrarMenu()
  bsPwdModal?.show()
}

async function cambiarPassword() {
  cambiando.value = true
  alertaPwd.value = null
  try {
    await apiUsuarios.modificarPassword(formPwd.value)
    bsPwdModal?.hide()
    // breve feedback via alert (se mostrará en la vista activa, no hay forma facil de notificar globalmente)
    // Simplemente cerramos el modal; el usuario verá que se cerró exitosamente
  } catch (e) {
    alertaPwd.value = e.response?.data?.error ?? 'Error al cambiar contraseña'
  } finally {
    cambiando.value = false
  }
}
</script>

<template>
  <nav class="app-navbar">
    <div class="navbar-inner">

      <router-link to="/usuarios" class="navbar-brand">
        <span class="brand-name">MSTF</span>
        <span class="brand-sub">Panel Administrativo</span>
      </router-link>

      <div class="navbar-links">
        <router-link
          v-for="link in navLinks"
          :key="link.name"
          :to="{ name: link.name }"
          class="nav-link"
          :class="{ active: route.name === link.name }"
        >
          {{ link.label }}
        </router-link>
      </div>

      <div class="navbar-actions">
        <div v-if="nombreUsuario" class="dropdown">
          <button
            ref="userDropdownEl"
            class="navbar-user-btn dropdown-toggle"
            type="button"
            @click="toggleDropdown"
          >
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" style="flex-shrink:0">
              <circle cx="12" cy="8" r="4"/><path d="M20 21a8 8 0 1 0-16 0"/>
            </svg>
            {{ nombreUsuario }}
          </button>
          <ul class="dropdown-menu dropdown-menu-end">
            <li>
              <button class="dropdown-item" @click="abrirCambiarPassword">
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="me-2">
                  <rect x="3" y="11" width="18" height="11" rx="2"/><path d="M7 11V7a5 5 0 0 1 10 0v4"/>
                </svg>
                Cambiar contraseña
              </button>
            </li>
            <li><hr class="dropdown-divider"></li>
            <li>
              <button class="dropdown-item text-danger" @click="logout">
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="me-2">
                  <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"/><polyline points="16 17 21 12 16 7"/><line x1="21" y1="12" x2="9" y2="12"/>
                </svg>
                Cerrar sesión
              </button>
            </li>
          </ul>
        </div>
      </div>

      <!-- Hamburger (solo mobile) -->
      <button class="navbar-hamburger" @click="menuAbierto = !menuAbierto" aria-label="Menú">
        <svg v-if="!menuAbierto" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <line x1="3" y1="6" x2="21" y2="6"/><line x1="3" y1="12" x2="21" y2="12"/><line x1="3" y1="18" x2="21" y2="18"/>
        </svg>
        <svg v-else width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/>
        </svg>
      </button>

    </div>

    <!-- Menú mobile desplegable -->
    <div v-show="menuAbierto" class="navbar-mobile-menu">
      <router-link
        v-for="link in navLinks"
        :key="link.name"
        :to="{ name: link.name }"
        class="mobile-nav-link"
        :class="{ active: route.name === link.name }"
        @click="cerrarMenu"
      >
        {{ link.label }}
      </router-link>
      <div class="mobile-nav-divider"></div>
      <button class="mobile-nav-link" @click="abrirCambiarPassword">Cambiar contraseña</button>
      <button class="mobile-nav-link mobile-nav-danger" @click="() => { cerrarMenu(); logout() }">Cerrar sesión</button>
    </div>

  </nav>

  <!-- Modal cambiar contraseña -->
  <div ref="pwdModalEl" class="modal fade" tabindex="-1" aria-hidden="true">
    <div class="modal-dialog modal-dialog-centered" style="max-width: 400px">
      <div class="modal-content">
        <div class="modal-header">
          <h5 class="modal-title">Cambiar contraseña</h5>
          <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Cerrar"></button>
        </div>
        <form @submit.prevent="cambiarPassword">
          <div class="modal-body">
            <div v-if="alertaPwd" class="alert alert-danger mb-3" role="alert">{{ alertaPwd }}</div>
            <div class="mb-3">
              <label class="form-label">Contraseña actual</label>
              <input
                v-model="formPwd.PasswordAnterior"
                type="password"
                class="form-control"
                autocomplete="current-password"
                :disabled="cambiando"
                required
                autofocus
              />
            </div>
            <div class="mb-3">
              <label class="form-label">Nueva contraseña</label>
              <input
                v-model="formPwd.PasswordNuevo"
                type="password"
                class="form-control"
                autocomplete="new-password"
                :disabled="cambiando"
                required
              />
            </div>
            <div class="mb-1">
              <label class="form-label">Confirmar nueva contraseña</label>
              <input
                v-model="formPwd.ConfirmarPassword"
                type="password"
                class="form-control"
                autocomplete="new-password"
                :disabled="cambiando"
                required
              />
            </div>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-outline-secondary btn-sm" data-bs-dismiss="modal">Cancelar</button>
            <button
              type="submit"
              class="btn btn-primary btn-sm"
              :disabled="cambiando || !formPwd.PasswordAnterior || !formPwd.PasswordNuevo || !formPwd.ConfirmarPassword"
            >
              <span v-if="cambiando" class="spinner-border spinner-border-sm me-1" role="status" aria-hidden="true"></span>
              Guardar
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<style scoped>
.app-navbar {
  position: sticky;
  top: 0;
  z-index: 1000;
  background: var(--surface);
  border-bottom: 1px solid var(--border);
  box-shadow: 0 1px 0 0 var(--border);
}

.navbar-inner {
  max-width: 1400px;
  margin: 0 auto;
  padding: 0 1.5rem;
  height: 64px;
  display: flex;
  align-items: center;
  gap: 2rem;
}

.navbar-brand {
  text-decoration: none;
  display: flex;
  flex-direction: column;
  line-height: 1.15;
  flex-shrink: 0;
}

.brand-name {
  font-family: var(--font-mono);
  font-size: 1.25rem;
  font-weight: 700;
  color: var(--accent);
  letter-spacing: -0.03em;
}

.brand-sub {
  font-family: var(--font-mono);
  font-size: 0.625rem;
  font-weight: 500;
  color: var(--text-tertiary);
  text-transform: uppercase;
  letter-spacing: 0.08em;
}

.navbar-links {
  display: flex;
  align-items: center;
  gap: 0.125rem;
  flex: 1;
}

.nav-link {
  font-family: var(--font-mono);
  font-size: 0.8125rem;
  font-weight: 500;
  color: var(--text-secondary);
  text-decoration: none;
  padding: 0.375rem 0.75rem;
  border-radius: var(--radius-md);
  transition: color var(--t-base), background-color var(--t-base);
  white-space: nowrap;
}

.nav-link:hover {
  color: var(--text-primary);
  background: var(--gray-50);
}

.nav-link.active {
  color: var(--accent);
  background: var(--accent-bg);
  font-weight: 600;
}

.navbar-actions {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  flex-shrink: 0;
}

.navbar-user-btn {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-family: var(--font-mono);
  font-size: 0.8125rem;
  font-weight: 500;
  color: var(--text-secondary);
  background: transparent;
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
  padding: 0.375rem 0.75rem;
  cursor: pointer;
  transition: color var(--t-base), background-color var(--t-base), border-color var(--t-base);
  white-space: nowrap;
}

.navbar-user-btn:hover {
  color: var(--text-primary);
  background: var(--gray-50);
  border-color: var(--text-tertiary);
}

.navbar-user-btn::after {
  margin-left: 0.25rem;
}

/* Dropdown menu overrides */
:deep(.dropdown-menu) {
  font-family: var(--font-mono);
  font-size: 0.8125rem;
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
  box-shadow: var(--shadow-sm);
  min-width: 200px;
  padding: 0.375rem;
}

:deep(.dropdown-item) {
  border-radius: var(--radius-sm);
  padding: 0.5rem 0.75rem;
  display: flex;
  align-items: center;
  color: var(--text-primary);
  transition: background-color var(--t-base);
}

:deep(.dropdown-item:hover) {
  background: var(--gray-50);
}

:deep(.dropdown-item.text-danger) {
  color: var(--error) !important;
}

:deep(.dropdown-divider) {
  margin: 0.375rem 0;
  border-color: var(--border);
}

/* Hamburger — oculto en desktop */
.navbar-hamburger {
  display: none;
  align-items: center;
  justify-content: center;
  background: transparent;
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
  padding: 0.375rem 0.5rem;
  cursor: pointer;
  color: var(--text-secondary);
  flex-shrink: 0;
}

.navbar-mobile-menu {
  display: flex;
  flex-direction: column;
  padding: 0.5rem 1rem 1rem;
  border-top: 1px solid var(--border);
  background: var(--surface);
}

.mobile-nav-link {
  font-family: var(--font-mono);
  font-size: 0.9375rem;
  font-weight: 500;
  color: var(--text-secondary);
  text-decoration: none;
  padding: 0.75rem 0.75rem;
  border-radius: var(--radius-md);
  border: none;
  background: transparent;
  text-align: left;
  cursor: pointer;
  display: block;
  width: 100%;
  transition: background-color var(--t-base), color var(--t-base);
}

.mobile-nav-link:hover { background: var(--gray-50); color: var(--text-primary); }
.mobile-nav-link.active { color: var(--accent); background: var(--accent-bg); font-weight: 600; }
.mobile-nav-danger { color: var(--error); }
.mobile-nav-danger:hover { background: #FEE2E2; color: var(--error); }

.mobile-nav-divider {
  height: 1px;
  background: var(--border);
  margin: 0.375rem 0;
}

@media (max-width: 767px) {
  .navbar-hamburger { display: flex; }
  .navbar-links     { display: none !important; }
  .navbar-actions   { display: none !important; }
  .navbar-inner     { gap: 0; }
  .navbar-brand     { margin-right: auto; }
}
</style>
