<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import Masthead from './components/Masthead.vue'
import MetadataColumn from './components/MetadataColumn.vue'
import DependencyColumn from './components/DependencyColumn.vue'
import SelectedTray from './components/SelectedTray.vue'
import FooterActions from './components/FooterActions.vue'
import ExploreModal from './components/ExploreModal.vue'
import Toast from './components/Toast.vue'
import { useCatalog } from './composables/useCatalog'
import { useSelection } from './composables/useSelection'
import { buildProjectTree } from './composables/useProjectTree'

const { t } = useI18n()
const { routers, dependencies, loading, error, load } = useCatalog()
const { selectedIds, toggle, remove } = useSelection()

// Project metadata (left column).
const modulePath = ref('github.com/me/my-app')
const projectName = ref('my-app')
const goVersion = ref('1.24')
const router = ref('stdlib')

const selectedDeps = computed(() =>
  dependencies.value.filter((d) => selectedIds.value.includes(d.id)),
)

const exploreOpen = ref(false)
const toast = ref({ show: false, message: '' })
let toastTimer = null

function showToast(message) {
  toast.value = { show: true, message }
  clearTimeout(toastTimer)
  toastTimer = setTimeout(() => {
    toast.value = { ...toast.value, show: false }
  }, 3200)
}

onMounted(load)

// Default the router to whatever the catalog marks as default, once loaded.
watch(routers, (list) => {
  const def = list.find((r) => r.default)
  if (def) router.value = def.id
})

async function generate() {
  showToast(t('generating'))
  const payload = {
    modulePath: modulePath.value,
    projectName: projectName.value,
    goVersion: goVersion.value,
    router: router.value,
    deps: selectedDeps.value,
  }
  try {
    const res = await fetch('/api/generate', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    })
    if (!res.ok) throw new Error(`HTTP ${res.status}`)
    const blob = await res.blob()
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `${projectName.value || 'project'}.zip`
    a.click()
    URL.revokeObjectURL(url)
    const { fileCount } = buildProjectTree(selectedIds.value, projectName.value)
    showToast(t('filesGenerated', { count: fileCount }))
  } catch (e) {
    // Backend is still a skeleton (POST /api/generate returns 501) — surface a
    // friendly message rather than failing silently.
    showToast(t('genFailed'))
  }
}
</script>

<template>
  <div class="shell">
    <Masthead />

    <div v-if="error" class="catalog-error" role="alert">
      <span>{{ t('catalogFailed') }}</span>
      <button class="retry" type="button" :disabled="loading" @click="load">
        {{ t('retry') }}
      </button>
    </div>

    <main class="layout">
      <MetadataColumn
        v-model:module-path="modulePath"
        v-model:project-name="projectName"
        v-model:go-version="goVersion"
        v-model:router="router"
        :routers="routers"
      />
      <DependencyColumn
        :dependencies="dependencies"
        :selected-ids="selectedIds"
        @toggle="toggle"
      />
    </main>

    <SelectedTray :selected-deps="selectedDeps" @remove="remove" />

    <FooterActions @generate="generate" @explore="exploreOpen = true" />

    <ExploreModal
      :open="exploreOpen"
      :selected-ids="selectedIds"
      :project-name="projectName"
      @close="exploreOpen = false"
    />

    <Toast :show="toast.show" :message="toast.message" />
  </div>
</template>

<style scoped>
.shell {
  max-width: var(--maxw);
  margin: 0 auto;
  padding: 0 28px;
}

.catalog-error {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  margin: 16px 0 0;
  padding: 12px 16px;
  border: 1px solid var(--danger, #e5484d);
  border-radius: var(--radius-sm);
  background: color-mix(in srgb, var(--danger, #e5484d) 10%, transparent);
  color: var(--ink);
  font-size: 13px;
}

.catalog-error .retry {
  flex-shrink: 0;
  padding: 6px 14px;
  border: 1px solid var(--danger, #e5484d);
  border-radius: var(--radius-sm);
  background: transparent;
  color: var(--danger, #e5484d);
  font-weight: 600;
  cursor: pointer;
}

.catalog-error .retry:disabled {
  opacity: 0.5;
  cursor: default;
}

.layout {
  display: grid;
  grid-template-columns: minmax(0, 38fr) minmax(0, 62fr);
  gap: 56px;
  padding: 36px 0;
}

@media (max-width: 860px) {
  .layout {
    grid-template-columns: 1fr;
    gap: 40px;
  }
}
</style>
