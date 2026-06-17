<script setup>
import { useI18n } from 'vue-i18n'
import RouterPicker from './RouterPicker.vue'

const { t } = useI18n()

const modulePath = defineModel('modulePath', { type: String })
const projectName = defineModel('projectName', { type: String })
const goVersion = defineModel('goVersion', { type: String })
const router = defineModel('router', { type: String })

defineProps({
  routers: { type: Array, default: () => [] },
})

const GO_VERSIONS = ['1.24', '1.23', '1.22']
</script>

<template>
  <section class="metadata">
    <h2 class="label section-label">{{ t('metadata') }}</h2>

    <div class="field">
      <label class="label" for="module-path">{{ t('modulePathLabel') }}</label>
      <input id="module-path" v-model="modulePath" class="input mono" type="text" spellcheck="false" />
      <p class="hint mono">{{ t('moduleHint') }}</p>
    </div>

    <div class="field">
      <label class="label" for="project-name">{{ t('projectNameLabel') }}</label>
      <input id="project-name" v-model="projectName" class="input mono" type="text" spellcheck="false" />
    </div>

    <div class="field">
      <label class="label" for="go-version">{{ t('goVersionLabel') }}</label>
      <select id="go-version" v-model="goVersion" class="input mono">
        <option v-for="v in GO_VERSIONS" :key="v" :value="v">{{ v }}</option>
      </select>
    </div>

    <div class="field">
      <span class="label">{{ t('routerLabel') }}</span>
      <RouterPicker v-model="router" :routers="routers" />
    </div>
  </section>
</template>

<style scoped>
.metadata {
  display: flex;
  flex-direction: column;
  gap: 22px;
}

.section-label {
  margin: 0;
}

.field {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.input {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid var(--line);
  border-radius: var(--radius-sm);
  background: var(--panel);
  color: var(--ink);
  font-size: 13px;
  transition: border-color 0.15s ease;
}

.input:focus {
  outline: none;
  border-color: var(--accent);
}

select.input {
  cursor: pointer;
}

.hint {
  margin: 0;
  font-size: 11px;
  color: var(--ink-faint);
}
</style>
