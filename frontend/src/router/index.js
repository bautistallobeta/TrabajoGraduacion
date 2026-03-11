import { createRouter, createWebHistory } from 'vue-router'
import { useAuth } from '../stores/auth'
import LoginView from '../views/LoginView.vue'
import AppLayout from '../components/AppLayout.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/login',
      name: 'login',
      component: LoginView,
      meta: { publica: true }
    },
    {
      path: '/',
      component: AppLayout,
      redirect: '/usuarios',
      children: [
        {
          path: 'usuarios',
          name: 'usuarios',
          component: () => import('../views/UsuariosView.vue'),
          meta: { titulo: 'Usuarios' }
        },
        {
          path: 'monedas',
          name: 'monedas',
          component: () => import('../views/MonedasView.vue'),
          meta: { titulo: 'Monedas' }
        },
        {
          path: 'parametros',
          name: 'parametros',
          component: () => import('../views/ParametrosView.vue'),
          meta: { titulo: 'Parámetros' }
        },
        {
          path: 'cuentas',
          name: 'cuentas',
          component: () => import('../views/CuentasView.vue'),
          meta: { titulo: 'Cuentas' }
        },
        {
          path: 'transferencias',
          name: 'transferencias',
          component: () => import('../views/TransferenciasView.vue'),
          meta: { titulo: 'Transferencias' }
        }
      ]
    },
    // Cualquier ruta desconocida redirige
    { path: '/:pathMatch(.*)*', redirect: '/usuarios' }
  ]
})

router.beforeEach((to) => {
  const { verificarExpiracion } = useAuth()
  const autenticado = verificarExpiracion()

  if (!to.meta.publica && !autenticado) {
    return { name: 'login' }
  }
  if (to.name === 'login' && autenticado) {
    return { name: 'usuarios' }
  }
})

export default router
