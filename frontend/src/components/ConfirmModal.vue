<script setup>
import { ref, onMounted } from 'vue'
import { Modal } from 'bootstrap'

const props = defineProps({
  title:          { type: String, required: true },
  message:        { type: String, default: '' },
  confirmLabel:   { type: String, default: 'Confirmar' },
  confirmVariant: { type: String, default: 'btn-outline-primary' }
})

const emit = defineEmits(['confirm', 'cancel'])

const modalEl = ref(null)
let bsModal = null

onMounted(() => {
  bsModal = new Modal(modalEl.value, { backdrop: 'static' })
})

function open()  { bsModal?.show() }
function close() { bsModal?.hide() }

defineExpose({ open, close })
</script>

<template>
  <div ref="modalEl" class="modal fade" tabindex="-1" aria-hidden="true">
    <div class="modal-dialog modal-dialog-centered">
      <div class="modal-content">

        <div class="modal-header">
          <h5 class="modal-title">{{ title }}</h5>
          <button
            type="button"
            class="btn-close"
            aria-label="Cerrar"
            @click="close(); emit('cancel')"
          ></button>
        </div>

        <div class="modal-body">
          <p class="mb-0" style="font-size: 0.875rem; color: var(--text-primary);" v-html="message"></p>
        </div>

        <div class="modal-footer">
          <button
            type="button"
            class="btn btn-outline-secondary btn-sm"
            @click="close(); emit('cancel')"
          >Cancelar</button>
          <button
            type="button"
            :class="`btn ${confirmVariant} btn-sm`"
            @click="emit('confirm')"
          >{{ confirmLabel }}</button>
        </div>

      </div>
    </div>
  </div>
</template>
