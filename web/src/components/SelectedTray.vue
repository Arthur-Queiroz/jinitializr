<script setup>
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

defineProps({
  selectedDeps: { type: Array, default: () => [] },
})

defineEmits(['remove'])
</script>

<template>
  <section class="tray">
    <div class="tray-head">
      <h2 class="label section-label">{{ t('selectedDependencies') }}</h2>
      <span v-if="selectedDeps.length" class="count mono">
        {{ selectedDeps.length }} {{ t('selected') }}
      </span>
    </div>

    <div v-if="selectedDeps.length" class="chips">
      <span v-for="d in selectedDeps" :key="d.id" class="chip">
        <span class="chip-dot" :style="{ background: `var(--cat-${d.category}, var(--accent))` }"></span>
        <span class="chip-name">{{ d.name }}</span>
        <button class="chip-x" type="button" :aria-label="`remove ${d.name}`" @click="$emit('remove', d.id)">
          <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round">
            <path d="M6 6l12 12M18 6L6 18" />
          </svg>
        </button>
      </span>
    </div>

    <div v-else class="empty">
      <div class="empty-mark" aria-hidden="true">{ }</div>
      <p class="empty-title">{{ t('emptyTitle') }}</p>
      <p class="empty-sub">{{ t('emptySub') }}</p>
    </div>
  </section>
</template>

<style scoped>
.tray {
  border-top: 1px solid var(--line);
  padding-top: 24px;
}

.tray-head {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  margin-bottom: 14px;
}

.section-label {
  margin: 0;
}

.count {
  font-size: 11px;
  color: var(--accent);
}

.chips {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.chip {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 7px 8px 7px 12px;
  border: 1px solid var(--line);
  border-radius: 999px;
  background: var(--panel);
  font-size: 13px;
  font-weight: 500;
}

.chip-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
}

.chip-x {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 20px;
  height: 20px;
  border: none;
  border-radius: 50%;
  background: var(--accent-soft);
  color: var(--ink-soft);
  transition: background-color 0.15s ease, color 0.15s ease;
}

.chip-x:hover {
  background: var(--accent);
  color: #fff;
}

.empty {
  padding: 36px 20px;
  text-align: center;
  border: 1px dashed var(--line);
  border-radius: var(--radius);
}

.empty-mark {
  font-family: var(--font-mono);
  font-size: 30px;
  color: var(--ink-faint);
  letter-spacing: 0.1em;
}

.empty-title {
  margin: 10px 0 4px;
  font-size: 15px;
  font-weight: 600;
}

.empty-sub {
  margin: 0;
  font-size: 13px;
  color: var(--ink-soft);
  max-width: 42ch;
  margin-inline: auto;
}
</style>
