<script setup>
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useAuth } from '../stores/auth'
import { confirmarCuenta } from '../api/usuarios'

const router = useRouter()
const { cerrarSesion } = useAuth()

const form = ref({ password: '', confirmarPassword: '' })
const cargando = ref(false)
const error = ref(null)

const errorCoincidencia = computed(() => {
  if (form.value.confirmarPassword && form.value.password !== form.value.confirmarPassword)
    return 'Las contraseñas no coinciden.'
  return null
})

async function confirmar() {
  if (!form.value.password || !form.value.confirmarPassword) return
  if (errorCoincidencia.value) return
  cargando.value = true
  error.value = null
  try {
    await confirmarCuenta(
      form.value.password,
      form.value.confirmarPassword
    )
    // si cuenta confirmada entonces limpiar sesión temporal y volver al login
    cerrarSesion()
    router.push({ name: 'login', query: { confirmado: '1' } })
  } catch (e) {
    error.value = e.response?.data?.error ?? 'Error al confirmar la cuenta'
  } finally {
    cargando.value = false
  }
}
</script>

<template>
  <div class="confirm-page">
    <div class="confirm-card">

      <div class="confirm-header">
        <div class="confirm-logo">MSTF</div>
        <div class="confirm-title">Activar cuenta</div>
        <p class="confirm-desc">
          Tu cuenta está pendiente de activación. Establecé tu contraseña definitiva.
        </p>
      </div>

      <form @submit.prevent="confirmar">
        <div v-if="error" class="alert alert-danger mb-4" role="alert">
          {{ error }}
        </div>

        <div class="mb-3">
          <label for="password" class="form-label">Nueva contraseña</label>
          <input
            id="password"
            v-model="form.password"
            type="password"
            class="form-control"
            autocomplete="new-password"
            :disabled="cargando"
            required
          />
        </div>

        <div class="mb-4">
          <label for="confirmarPassword" class="form-label">Confirmar contraseña</label>
          <input
            id="confirmarPassword"
            v-model="form.confirmarPassword"
            type="password"
            class="form-control"
            :class="{ 'is-invalid': errorCoincidencia }"
            autocomplete="new-password"
            :disabled="cargando"
            required
          />
          <div v-if="errorCoincidencia" class="invalid-feedback">{{ errorCoincidencia }}</div>
        </div>

        <button type="submit" class="btn btn-primary w-100 confirm-btn" :disabled="cargando || !!errorCoincidencia">
          <span v-if="cargando" class="spinner-border spinner-border-sm me-2" role="status" aria-hidden="true"></span>
          {{ cargando ? 'Activando cuenta...' : 'Activar cuenta' }}
        </button>

        <div class="mt-3 text-center">
          <button type="button" class="btn-link" @click="cerrarSesion(); router.push('/login')">
            Volver al inicio de sesión
          </button>
        </div>
      </form>

    </div>
  </div>
</template>

<style scoped>
.confirm-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--bg);
  padding: 1.5rem;
}

.confirm-card {
  width: 100%;
  max-width: 420px;
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: var(--radius-xl);
  box-shadow: var(--shadow-sm);
  padding: 2.5rem;
}

.confirm-header {
  text-align: center;
  margin-bottom: 2rem;
}

.confirm-logo {
  font-family: var(--font-mono);
  font-size: 2.25rem;
  font-weight: 700;
  color: var(--accent);
  letter-spacing: -0.04em;
  line-height: 1;
  margin-bottom: 0.5rem;
}

.confirm-title {
  font-family: var(--font-mono);
  font-size: 0.9375rem;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 0.625rem;
}

.confirm-desc {
  font-family: var(--font-sans);
  font-size: 0.875rem;
  color: var(--text-secondary);
  margin-bottom: 0;
  line-height: 1.5;
}

.confirm-btn {
  height: 2.5rem;
  font-size: 0.875rem;
}

.btn-link {
  background: none;
  border: none;
  padding: 0;
  font-family: var(--font-mono);
  font-size: 0.75rem;
  color: var(--text-tertiary);
  cursor: pointer;
  text-decoration: underline;
  transition: color var(--t-base);
}

.btn-link:hover {
  color: var(--text-secondary);
}
</style>
