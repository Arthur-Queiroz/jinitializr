import { ref } from 'vue'

// Selected dependency ids. Shared singleton so every component (the list, the
// tray, the explore modal, the footer) reads and mutates the same state.
// Nothing is selected by default.
const selectedIds = ref([])

export function useSelection() {
  function isSelected(id) {
    return selectedIds.value.includes(id)
  }

  function toggle(id) {
    if (isSelected(id)) {
      remove(id)
    } else {
      selectedIds.value = [...selectedIds.value, id]
    }
  }

  function remove(id) {
    selectedIds.value = selectedIds.value.filter((x) => x !== id)
  }

  function clear() {
    selectedIds.value = []
  }

  return { selectedIds, isSelected, toggle, remove, clear }
}
