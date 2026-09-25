import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import legacy from '@vitejs/plugin-legacy'

export default defineConfig({
  plugins: [
    vue(),
    legacy({
      // Safari 13 needs the legacy bundle and its runtime polyfills.
      targets: ['defaults', 'Safari >= 11.1', 'iOS >= 11.3', 'not IE 11'],
    }),
  ],
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://127.0.0.1:7530',
        changeOrigin: true,
      },
    },
  },
  build: {
    // /assets is reserved by the Go server for user-uploaded files.
    assetsDir: 'static',
    outDir: 'dist',
    emptyOutDir: true,
  },
})
