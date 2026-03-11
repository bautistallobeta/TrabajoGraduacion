import axios from 'axios'
import { useAuth } from '../stores/auth'

const cliente = axios.create({
  baseURL: '/api',
  timeout: 10000,
  headers: { 'Content-Type': 'application/json' }
})

// Agrega el token Bearer en cada request autenticado
cliente.interceptors.request.use((config) => {
  const { token } = useAuth()
  if (token.value) {
    config.headers.Authorization = `Bearer ${token.value}`
  }
  return config
})

// Si el servidor responde 401, cierra la sesión y redirige al login
cliente.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      const { cerrarSesion } = useAuth()
      cerrarSesion()
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

export default cliente
