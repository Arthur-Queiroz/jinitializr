<script setup>
import { useI18n } from 'vue-i18n'
import { useTheme } from '../composables/useTheme'

const { t, locale } = useI18n()
const { theme, toggle } = useTheme()

const LANG_KEY = 'ji-lang'

function setLocale(lang) {
  locale.value = lang
  localStorage.setItem(LANG_KEY, lang)
}
</script>

<template>
  <header class="masthead">
    <div class="brand">
      <span class="wordmark mono">J<span class="accent">·</span>Initializr</span>
      <p class="tagline">{{ t('tagline') }}</p>
    </div>

    <div class="controls">
      <div class="lang" role="group" aria-label="Language">
        <button
          class="lang-btn"
          :class="{ active: locale === 'en' }"
          type="button"
          @click="setLocale('en')"
        >
          EN
        </button>
        <span class="lang-sep">/</span>
        <button
          class="lang-btn"
          :class="{ active: locale === 'pt' }"
          type="button"
          @click="setLocale('pt')"
        >
          PT
        </button>
      </div>

      <button class="theme-btn" type="button" :aria-label="theme" @click="toggle">
        <svg v-if="theme === 'light'" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round">
          <circle cx="12" cy="12" r="4" />
          <path d="M12 2v2M12 20v2M2 12h2M20 12h2M4.9 4.9l1.4 1.4M17.7 17.7l1.4 1.4M19.1 4.9l-1.4 1.4M6.3 17.7l-1.4 1.4" />
        </svg>
        <svg v-else width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round">
          <path d="M21 12.8A9 9 0 1 1 11.2 3a7 7 0 0 0 9.8 9.8z" />
        </svg>
      </button>
    </div>
  </header>
</template>

<style scoped>
.masthead {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 24px;
  padding: 28px 0 24px;
  border-bottom: 1px solid var(--line);
}

.wordmark {
  font-size: 22px;
  font-weight: 600;
  letter-spacing: -0.01em;
}

.wordmark .accent {
  color: var(--accent);
}

.tagline {
  margin: 6px 0 0;
  color: var(--ink-soft);
  font-size: 14px;
  max-width: 42ch;
}

.controls {
  display: flex;
  align-items: center;
  gap: 14px;
  flex-shrink: 0;
}

.lang {
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

.lang-btn {
  background: none;
  border: none;
  padding: 4px 6px;
  font-family: var(--font-mono);
  font-size: 12px;
  font-weight: 600;
  letter-spacing: 0.06em;
  color: var(--ink-faint);
  transition: color 0.15s ease;
}

.lang-btn.active {
  color: var(--accent);
}

.lang-sep {
  color: var(--ink-faint);
  font-size: 12px;
}

.theme-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  border: 1px solid var(--line);
  border-radius: var(--radius-sm);
  background: var(--panel);
  color: var(--ink);
  transition: border-color 0.15s ease, color 0.15s ease;
}

.theme-btn:hover {
  border-color: var(--accent);
  color: var(--accent);
}
</style>
