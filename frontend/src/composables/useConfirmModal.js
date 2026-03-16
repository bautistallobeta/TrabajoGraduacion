import { ref } from 'vue'

export function useConfirmModal() {
  const confirmModalRef = ref(null)
  const confirmConfig   = ref({ title: '', message: '', confirmLabel: '', confirmVariant: 'btn-outline-danger' })
  let accionPendiente   = null

  function pedirConfirmacion(config, accion) {
    confirmConfig.value = config
    accionPendiente = accion
    confirmModalRef.value.open()
  }

  async function onConfirmar() {
    confirmModalRef.value.close()
    if (accionPendiente) {
      await accionPendiente()
      accionPendiente = null
    }
  }

  function onCancelar() {
    accionPendiente = null
  }

  return { confirmModalRef, confirmConfig, pedirConfirmacion, onConfirmar, onCancelar }
}
