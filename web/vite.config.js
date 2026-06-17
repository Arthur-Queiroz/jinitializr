import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// The build output (web/dist) is embedded into the Go binary via
// `//go:embed all:web/dist`. Relative base keeps asset URLs working whatever
// path the Go server mounts the SPA on.
export default defineConfig({
  plugins: [vue()],
  base: './',
  build: {
    outDir: 'dist',
    emptyOutDir: true,
  },
  // Silence vue-i18n bundler feature-flag warnings without pulling in the
  // @intlify build plugin (kept out to honour the minimal dependency set).
  define: {
    __VUE_I18N_FULL_INSTALL__: true,
    __VUE_I18N_LEGACY_API__: false,
    __INTLIFY_PROD_DEVTOOLS__: false,
  },
  server: {
    // Dev server proxies the API to the Go backend on :8080.
    proxy: {
      '/api': 'http://localhost:8080',
    },
  },
})
