import axios from 'axios'
import { useAuth } from '../stores/auth'

const cliente = axios.create({
  baseURL: '/api',
  timeout: 10000,
  headers: { 'Content-Type': 'application/json' }
})

cliente.interceptors.request.use((config) => {
  const { token } = useAuth()
  if (token.value) {
    config.headers.Authorization = `Bearer ${token.value}`
  }
  return config
})

// si 401 cierra sesión y redirige, salvo que la request tenga _noRedirect: true
cliente.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401 && !error.config._noRedirect) {
      const { cerrarSesion } = useAuth()
      cerrarSesion()
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

export default cliente
