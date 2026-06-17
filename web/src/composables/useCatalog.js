import { ref } from 'vue'

// Loads the dependency/router catalog from the Go backend. The backend owns the
// catalog (names, modules, versions); the frontend only filters it client-side
// for search. On failure we expose `error` so the UI can tell the user instead
// of silently rendering an empty form.
export function useCatalog() {
  const routers = ref([])
  const dependencies = ref([])
  const loading = ref(false)
  const error = ref(null)

  async function load() {
    loading.value = true
    error.value = null
    try {
      const res = await fetch('/api/catalog')
      if (!res.ok) throw new Error(`HTTP ${res.status}`)
      const data = await res.json()
      routers.value = data.routers ?? []
      dependencies.value = data.dependencies ?? []
    } catch (e) {
      error.value = e
      routers.value = []
      dependencies.value = []
    } finally {
      loading.value = false
    }
  }

  return { routers, dependencies, loading, error, load }
}
