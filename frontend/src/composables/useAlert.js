import { ref, onBeforeUnmount } from 'vue'

export function useAlert() {
  const alerta = ref(null)
  let alertaTimer = null

  function mostrarAlerta(tipo, mensaje) {
    clearTimeout(alertaTimer)
    alerta.value = { tipo, mensaje }
    alertaTimer = setTimeout(() => { alerta.value = null }, 4000)
  }

  onBeforeUnmount(() => { clearTimeout(alertaTimer) })

  return { alerta, mostrarAlerta }
}
