<script setup>
const model = defineModel({ type: String })

defineProps({
  routers: { type: Array, default: () => [] },
})
</script>

<template>
  <div class="router-picker" role="radiogroup">
    <label
      v-for="r in routers"
      :key="r.id"
      class="router-opt"
      :class="{ active: model === r.id }"
    >
      <input
        type="radio"
        name="router"
        :value="r.id"
        :checked="model === r.id"
        @change="model = r.id"
      />
      <span class="dot" aria-hidden="true"></span>
      <span class="r-name mono">{{ r.name }}</span>
      <span v-if="r.module" class="r-mod mono">{{ r.module }}</span>
    </label>
  </div>
</template>

<style scoped>
.router-picker {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.router-opt {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 11px 13px;
  border: 1px solid var(--line);
  border-radius: var(--radius-sm);
  cursor: pointer;
  transition: border-color 0.15s ease, background-color 0.15s ease;
}

.router-opt:hover {
  border-color: var(--accent);
}

.router-opt.active {
  border-color: var(--accent);
  background: var(--accent-soft);
}

.router-opt input {
  position: absolute;
  opacity: 0;
  pointer-events: none;
}

.dot {
  width: 14px;
  height: 14px;
  border-radius: 50%;
  border: 1.5px solid var(--ink-faint);
  flex-shrink: 0;
  transition: border-color 0.15s ease;
}

.router-opt.active .dot {
  border-color: var(--accent);
  background:
    radial-gradient(circle at center, var(--accent) 0 4px, transparent 4px);
}

.r-name {
  font-size: 13px;
  font-weight: 600;
}

.r-mod {
  margin-left: auto;
  font-size: 11px;
  color: var(--ink-faint);
}
</style>
