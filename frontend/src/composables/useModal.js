import { onMounted, onBeforeUnmount } from 'vue'
import { Modal } from 'bootstrap'

export function useModal(elRef, onHide) {
  let bsModal = null

  onMounted(() => {
    bsModal = new Modal(elRef.value)
    if (onHide) elRef.value.addEventListener('hidden.bs.modal', onHide)
  })

  onBeforeUnmount(() => { bsModal?.dispose() })

  return {
    show: () => bsModal?.show(),
    hide: () => bsModal?.hide()
  }
}
