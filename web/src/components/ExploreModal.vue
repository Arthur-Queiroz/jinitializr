<script setup>
import { computed, onMounted, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { buildProjectTree } from '../composables/useProjectTree'

const { t, te } = useI18n()

const props = defineProps({
  open: { type: Boolean, default: false },
  selectedIds: { type: Array, default: () => [] },
  projectName: { type: String, default: 'my-app' },
})

const emit = defineEmits(['close'])

const tree = computed(() => buildProjectTree(props.selectedIds, props.projectName))

const comment = (key) => (key && te(key) ? t(key) : '')

function onKey(e) {
  if (e.key === 'Escape' && props.open) emit('close')
}

onMounted(() => window.addEventListener('keydown', onKey))
onUnmounted(() => window.removeEventListener('keydown', onKey))
</script>

<template>
  <Transition name="modal">
    <div v-if="open" class="overlay" @click.self="emit('close')">
      <div class="modal" role="dialog" aria-modal="true">
        <header class="modal-head">
          <h2 class="modal-title">{{ t('projectStructure') }}</h2>
          <button class="close" type="button" aria-label="close" @click="emit('close')">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round">
              <path d="M6 6l12 12M18 6L6 18" />
            </svg>
          </button>
        </header>

        <div class="tree mono">
          <div v-for="(line, i) in tree.lines" :key="i" class="tree-line">
            <span class="connector">{{ line.prefix }}</span><span
              class="node"
              :class="{ dir: line.dir }"
            >{{ line.name }}</span><span v-if="comment(line.comment)" class="comment">
              # {{ comment(line.comment) }}</span>
          </div>
        </div>

        <footer class="modal-foot mono">
          {{ t('filesGenerated', { count: tree.fileCount }) }}
        </footer>
      </div>
    </div>
  </Transition>
</template>

<style scoped>
.overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.4);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 24px;
  z-index: 40;
}

.modal {
  width: 100%;
  max-width: 640px;
  max-height: 84vh;
  display: flex;
  flex-direction: column;
  background: var(--panel);
  border: 1px solid var(--line);
  border-radius: var(--radius);
  overflow: hidden;
}

.modal-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 18px 22px;
  border-bottom: 1px solid var(--line);
}

.modal-title {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
}

.close {
  display: inline-flex;
  border: none;
  background: none;
  color: var(--ink-soft);
}

.close:hover {
  color: var(--accent);
}

.tree {
  padding: 20px 24px;
  overflow: auto;
  font-size: 13px;
  line-height: 1.85;
}

.tree-line {
  white-space: pre;
}

.connector {
  color: var(--ink-faint);
}

.node {
  color: var(--ink);
}

.node.dir {
  color: var(--accent);
  font-weight: 600;
}

.comment {
  color: var(--ink-faint);
}

.modal-foot {
  padding: 14px 22px;
  border-top: 1px solid var(--line);
  font-size: 11px;
  color: var(--ink-soft);
}

.modal-enter-active,
.modal-leave-active {
  transition: opacity 0.2s ease;
}

.modal-enter-from,
.modal-leave-to {
  opacity: 0;
}
</style>
