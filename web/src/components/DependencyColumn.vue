<script setup>
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import DependencyRow from './DependencyRow.vue'

const { t, te } = useI18n()

const props = defineProps({
  dependencies: { type: Array, default: () => [] },
  selectedIds: { type: Array, default: () => [] },
})

defineEmits(['toggle'])

const query = ref('')

// Client-side search: matches name, module, id, category and the translated
// description + category label. No round-trip to the server.
const filtered = computed(() => {
  const q = query.value.trim().toLowerCase()
  if (!q) return props.dependencies
  return props.dependencies.filter((d) => {
    const desc = te(`deps.${d.id}`) ? t(`deps.${d.id}`) : ''
    const cat = te(`cat.${d.category}`) ? t(`cat.${d.category}`) : ''
    return [d.name, d.module, d.id, d.category, desc, cat]
      .filter(Boolean)
      .join(' ')
      .toLowerCase()
      .includes(q)
  })
})

const isSelected = (id) => props.selectedIds.includes(id)
</script>

<template>
  <section class="deps">
    <div class="deps-head">
      <h2 class="label section-label">{{ t('dependencies') }}</h2>
      <span class="add-only">{{ t('addOnly') }}</span>
    </div>

    <div class="search">
      <svg class="search-icon" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
        <circle cx="11" cy="11" r="7" />
        <path d="m21 21-4.3-4.3" />
      </svg>
      <input v-model="query" class="search-input" type="search" :placeholder="t('searchPlaceholder')" />
    </div>

    <p class="counter mono">
      {{ t('showing') }} {{ filtered.length }} {{ t('of') }} {{ dependencies.length }}
    </p>

    <div v-if="filtered.length" class="dep-list">
      <DependencyRow
        v-for="d in filtered"
        :key="d.id"
        :dep="d"
        :selected="isSelected(d.id)"
        @toggle="$emit('toggle', $event)"
      />
    </div>
    <p v-else class="no-match">{{ t('noMatch') }}</p>
  </section>
</template>

<style scoped>
.deps {
  display: flex;
  flex-direction: column;
}

.deps-head {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 16px;
}

.section-label {
  margin: 0;
}

.add-only {
  font-size: 12px;
  color: var(--ink-faint);
}

.search {
  position: relative;
  display: flex;
  align-items: center;
}

.search-icon {
  position: absolute;
  left: 12px;
  color: var(--ink-faint);
  pointer-events: none;
}

.search-input {
  width: 100%;
  padding: 11px 12px 11px 36px;
  border: 1px solid var(--line);
  border-radius: var(--radius-sm);
  background: var(--panel);
  color: var(--ink);
  font-size: 14px;
  transition: border-color 0.15s ease;
}

.search-input:focus {
  outline: none;
  border-color: var(--accent);
}

.counter {
  margin: 12px 2px 0;
  font-size: 11px;
  color: var(--ink-faint);
}

.dep-list {
  margin-top: 4px;
  border-top: 1px solid var(--line);
}

.dep-list :deep(.dep-row:last-child) {
  border-bottom: none;
}

.no-match {
  margin: 28px 0;
  text-align: center;
  color: var(--ink-soft);
  font-size: 14px;
}
</style>
