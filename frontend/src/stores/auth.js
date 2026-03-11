import { ref, computed } from 'vue'

const SESSION_KEY = 'mstf_session'
const SESSION_DURATION = 15 * 60 * 1000 // 15 minutos en ms

function cargarSesion() {
  try {
    const raw = sessionStorage.getItem(SESSION_KEY)
    if (!raw) return null
    const data = JSON.parse(raw)
    if (Date.now() > data.expira) {
      sessionStorage.removeItem(SESSION_KEY)
      return null
    }
    return data
  } catch {
    return null
  }
}

// Estado global (singleton)
const sesion = ref(cargarSesion())

export function useAuth() {
  const estaAutenticado = computed(() => sesion.value !== null)
  const token = computed(() => sesion.value?.token ?? null)
  const nombreUsuario = computed(() => sesion.value?.nombreUsuario ?? null)

  function iniciarSesion(tokenRecibido, nombreUsuarioRecibido) {
    const data = {
      token: tokenRecibido,
      nombreUsuario: nombreUsuarioRecibido,
      expira: Date.now() + SESSION_DURATION
    }
    sesion.value = data
    sessionStorage.setItem(SESSION_KEY, JSON.stringify(data))
  }

  function cerrarSesion() {
    sesion.value = null
    sessionStorage.removeItem(SESSION_KEY)
  }

  // Verifica expiración y limpia si corresponde
  // Retorna true si la sesión es válida
  function verificarExpiracion() {
    if (sesion.value && Date.now() > sesion.value.expira) {
      cerrarSesion()
      return false
    }
    return sesion.value !== null
  }

  return { estaAutenticado, token, nombreUsuario, iniciarSesion, cerrarSesion, verificarExpiracion }
}
