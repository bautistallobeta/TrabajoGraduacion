<script setup>
import { ref, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuth } from '../stores/auth'
import cliente from '../api/cliente'

const router = useRouter()
const route  = useRoute()
const { iniciarSesion } = useAuth()

const form    = ref({ usuario: '', password: '' })
const cargando = ref(false)
const error   = ref(null)
const exito   = ref(null)

onMounted(() => {
  if (route.query.confirmado === '1') {
    exito.value = 'Cuenta activada correctamente. Ya podés iniciar sesión con tu nueva contraseña.'
  }
})

async function login() {
  if (!form.value.usuario || !form.value.password) return
  cargando.value = true
  error.value = null
  try {
    const res = await cliente.post('/usuarios/login', {
      Usuario: form.value.usuario,
      Password: form.value.password
    })
    iniciarSesion(res.data.TokenSesion, form.value.usuario)

    try {
      await cliente.get('/parametros', { _noRedirect: true })
      router.push('/usuarios')
    } catch (testError) {
      if (testError.response?.status === 401) {
        // Mantener el token xq lo necesita el endpoint de confirmar-cuenta
        router.push('/confirmar-cuenta')
      } else {
        router.push('/usuarios')
      }
    }
  } catch (e) {
    error.value = e.response?.data?.error ?? 'Credenciales incorrectas'
  } finally {
    cargando.value = false
  }
}
</script>

<template>
  <div class="login-page">
    <div class="login-card">

      <div class="login-header">
        <div class="login-logo">MSTF</div>
        <div class="login-subtitle">Panel Administrativo</div>
      </div>

      <form @submit.prevent="login">
        <div v-if="exito" class="alert alert-success mb-4" role="alert">
          {{ exito }}
        </div>
        <div v-if="error" class="alert alert-danger mb-4" role="alert">
          {{ error }}
        </div>

        <div class="mb-3">
          <label for="usuario" class="form-label">Usuario</label>
          <input
            id="usuario"
            v-model="form.usuario"
            type="text"
            class="form-control"
            placeholder="Nombre de usuario"
            autocomplete="username"
            :disabled="cargando"
            required
          />
        </div>

        <div class="mb-4">
          <label for="password" class="form-label">Contraseña</label>
          <input
            id="password"
            v-model="form.password"
            type="password"
            class="form-control"
            placeholder="••••••••"
            autocomplete="current-password"
            :disabled="cargando"
            required
          />
        </div>

        <button type="submit" class="btn btn-primary w-100 login-btn" :disabled="cargando">
          <span
            v-if="cargando"
            class="spinner-border spinner-border-sm me-2"
            role="status"
            aria-hidden="true"
          ></span>
          {{ cargando ? 'Iniciando sesión...' : 'Iniciar sesión' }}
        </button>
      </form>

    </div>
  </div>
</template>

<style scoped>
.login-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--bg);
  padding: 1.5rem;
}

.login-card {
  width: 100%;
  max-width: 400px;
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: var(--radius-xl);
  box-shadow: var(--shadow-sm);
  padding: 2.5rem;
}

.login-header {
  text-align: center;
  margin-bottom: 2rem;
}

.login-logo {
  font-family: var(--font-mono);
  font-size: 2.25rem;
  font-weight: 700;
  color: var(--accent);
  letter-spacing: -0.04em;
  line-height: 1;
  margin-bottom: 0.5rem;
}

.login-subtitle {
  font-family: var(--font-mono);
  font-size: 0.6875rem;
  font-weight: 500;
  text-transform: uppercase;
  letter-spacing: 0.1em;
  color: var(--text-tertiary);
}

.login-btn {
  height: 2.5rem;
  font-size: 0.875rem;
}
</style>
