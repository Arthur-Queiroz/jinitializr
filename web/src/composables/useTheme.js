import { ref } from 'vue'

// Light/dark theme, applied via [data-theme] on <html>, persisted in localStorage.
const THEME_KEY = 'ji-theme'

const saved = localStorage.getItem(THEME_KEY)
const theme = ref(saved === 'dark' || saved === 'light' ? saved : 'light')

function apply(value) {
  document.documentElement.setAttribute('data-theme', value)
}

// Apply once at module load so the initial paint matches the stored preference.
apply(theme.value)

export function useTheme() {
  function toggle() {
    theme.value = theme.value === 'light' ? 'dark' : 'light'
    localStorage.setItem(THEME_KEY, theme.value)
    apply(theme.value)
  }

  return { theme, toggle }
}
