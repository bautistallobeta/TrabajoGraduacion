import { ref, computed } from 'vue'

function paginasBotones(paginaActual, totalPaginas) {
  const tot = totalPaginas
  const act = paginaActual
  if (tot <= 7) return Array.from({ length: tot }, (_, i) => i + 1)
  const pags = [1]
  if (act > 3) pags.push('...')
  for (let p = Math.max(2, act - 1); p <= Math.min(tot - 1, act + 1); p++) pags.push(p)
  if (act < tot - 2) pags.push('...')
  pags.push(tot)
  return pags
}

export function usePagination(items, porPagina = 50) {
  const paginaActual  = ref(1)
  const totalPaginas  = computed(() => Math.max(1, Math.ceil(items.value.length / porPagina)))
  const itemsEnPagina = computed(() => items.value.slice((paginaActual.value - 1) * porPagina, paginaActual.value * porPagina))
  const botones       = computed(() => paginasBotones(paginaActual.value, totalPaginas.value))

  return { paginaActual, totalPaginas, itemsEnPagina, botones }
}
