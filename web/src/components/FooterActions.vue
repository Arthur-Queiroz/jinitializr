<script setup>
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

defineEmits(['generate', 'explore'])
</script>

<template>
  <footer class="footer">
    <button class="btn ghost" type="button" @click="$emit('explore')">
      <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
        <path d="M3 7a2 2 0 0 1 2-2h4l2 2h8a2 2 0 0 1 2 2v8a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z" />
      </svg>
      {{ t('explore') }}
    </button>

    <button class="btn primary" type="button" @click="$emit('generate')">
      <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
        <path d="M12 3v12M7 10l5 5 5-5M5 21h14" />
      </svg>
      {{ t('generate') }}
    </button>
  </footer>
</template>

<style scoped>
.footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding: 24px 0 40px;
}

.btn {
  display: inline-flex;
  align-items: center;
  gap: 9px;
  padding: 12px 22px;
  border-radius: var(--radius-sm);
  font-size: 14px;
  font-weight: 600;
  border: 1px solid transparent;
  transition: all 0.15s ease;
}

.ghost {
  background: var(--panel);
  border-color: var(--line);
  color: var(--ink);
}

.ghost:hover {
  border-color: var(--accent);
  color: var(--accent);
}

.primary {
  background: var(--accent);
  color: #fff;
}

.primary:hover {
  filter: brightness(1.07);
}
</style>
