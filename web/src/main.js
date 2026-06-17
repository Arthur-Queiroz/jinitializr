import { createApp } from 'vue'
import { createI18n } from 'vue-i18n'
import App from './App.vue'
import en from './locales/en.json'
import pt from './locales/pt.json'
import './assets/tokens.css'

const LANG_KEY = 'ji-lang'
const saved = localStorage.getItem(LANG_KEY)
const locale = saved === 'pt' || saved === 'en' ? saved : 'en'

const i18n = createI18n({
  legacy: false,
  globalInjection: true,
  locale,
  fallbackLocale: 'en',
  messages: { en, pt },
})

createApp(App).use(i18n).mount('#app')
