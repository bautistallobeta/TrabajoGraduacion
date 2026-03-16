<script setup>
defineProps({
  modelValue:   { type: Number, required: true },
  totalPaginas: { type: Number, required: true },
  botones:      { type: Array,  required: true }
})
const emit = defineEmits(['update:modelValue'])
</script>

<template>
  <nav v-if="totalPaginas > 1" class="d-flex justify-content-center py-3">
    <ul class="pagination pagination-sm mb-0">
      <li class="page-item" :class="{ disabled: modelValue === 1 }">
        <button class="page-link" @click="emit('update:modelValue', modelValue - 1)">&lsaquo;</button>
      </li>
      <li
        v-for="p in botones"
        :key="String(p)"
        class="page-item"
        :class="{ active: p === modelValue, disabled: p === '...' }"
      >
        <button class="page-link" @click="typeof p === 'number' && emit('update:modelValue', p)">{{ p }}</button>
      </li>
      <li class="page-item" :class="{ disabled: modelValue === totalPaginas }">
        <button class="page-link" @click="emit('update:modelValue', modelValue + 1)">&rsaquo;</button>
      </li>
    </ul>
  </nav>
</template>
