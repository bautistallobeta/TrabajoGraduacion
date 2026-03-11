<script setup>
import { useRouter, useRoute } from 'vue-router'
import { useAuth } from '../stores/auth'
import cliente from '../api/cliente'

const router = useRouter()
const route = useRoute()
const { nombreUsuario, cerrarSesion } = useAuth()

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
    // Si falla el logout en el servidor igual limpiamos localmente
  } finally {
    cerrarSesion()
    router.push('/login')
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
        <span v-if="nombreUsuario" class="navbar-user">{{ nombreUsuario }}</span>
        <button class="btn btn-outline-primary btn-sm" @click="logout">
          Cerrar sesión
        </button>
      </div>

    </div>
  </nav>
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

.navbar-user {
  font-family: var(--font-mono);
  font-size: 0.75rem;
  color: var(--text-secondary);
}
</style>
