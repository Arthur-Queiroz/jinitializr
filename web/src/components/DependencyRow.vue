<script setup>
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

const { t, te } = useI18n()

const props = defineProps({
  dep: { type: Object, required: true },
  selected: { type: Boolean, default: false },
})

defineEmits(['toggle'])

const catColor = computed(() => `var(--cat-${props.dep.category}, var(--accent))`)
const catLabel = computed(() =>
  te(`cat.${props.dep.category}`) ? t(`cat.${props.dep.category}`) : props.dep.category,
)
const description = computed(() => (te(`deps.${props.dep.id}`) ? t(`deps.${props.dep.id}`) : ''))
</script>

<template>
  <div class="dep-row" :class="{ selected }">
    <div class="dep-main">
      <div class="dep-head">
        <span class="dep-name">{{ dep.name }}</span>
        <span class="cat-tag mono" :style="{ color: catColor }">
          <span class="cat-dot" :style="{ background: catColor }"></span>{{ catLabel }}
        </span>
      </div>
      <p v-if="description" class="dep-desc">{{ description }}</p>
      <div class="dep-meta mono">
        <span v-if="dep.module" class="dep-mod">{{ dep.module }}</span>
        <span v-if="dep.version" class="dep-ver">{{ dep.version }}</span>
      </div>
    </div>

    <button
      class="toggle"
      :class="{ on: selected }"
      type="button"
      :aria-pressed="selected"
      @click="$emit('toggle', dep.id)"
    >
      <svg v-if="selected" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <path d="M5 12h14" />
      </svg>
      <svg v-else width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <path d="M12 5v14M5 12h14" />
      </svg>
    </button>
  </div>
</template>

<style scoped>
.dep-row {
  display: flex;
  align-items: flex-start;
  gap: 16px;
  padding: 16px 0;
  border-bottom: 1px solid var(--line);
}

.dep-main {
  min-width: 0;
  flex: 1;
}

.dep-head {
  display: flex;
  align-items: baseline;
  gap: 10px;
  flex-wrap: wrap;
}

.dep-name {
  font-size: 15px;
  font-weight: 600;
}

.cat-tag {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  font-size: 10px;
  font-weight: 600;
  letter-spacing: 0.06em;
  text-transform: uppercase;
}

.cat-dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
}

.dep-desc {
  margin: 6px 0 0;
  font-size: 13px;
  color: var(--ink-soft);
}

.dep-meta {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-top: 8px;
  font-size: 11px;
  color: var(--ink-faint);
}

.toggle {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 34px;
  height: 34px;
  flex-shrink: 0;
  border: 1px solid var(--line);
  border-radius: var(--radius-sm);
  background: var(--panel);
  color: var(--accent);
  transition: all 0.15s ease;
}

.toggle:hover {
  border-color: var(--accent);
}

.toggle.on {
  background: var(--accent);
  border-color: var(--accent);
  color: #fff;
}
</style>
