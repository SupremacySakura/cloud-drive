import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import tailwindcss from '@tailwindcss/vite'
import vueJsx from '@vitejs/plugin-vue-jsx'
import { resolve } from 'path'

// https://vite.dev/config/
export default defineConfig({
  plugins: [vue(), tailwindcss(), vueJsx()],
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src'),
    },
  },
  build: {
    target: 'es2020',
    sourcemap: true,
    rollupOptions: {
      output: {
        manualChunks(id) {
          if (id.includes('node_modules')) {
            if (id.includes('vue')) return 'vue'
            if (id.includes('axios')) return 'axios'
            if (id.includes('pinia')) return 'pinia'
          }
        },
      },
    },
  },
  server: {
    host: '127.0.0.1',
    port: Number(process.env.DEV_FRONTEND_PORT || 5173),
    strictPort: true,
    proxy: {
      '/api': {
        target: process.env.DEV_BACKEND_URL || 'http://127.0.0.1:9000',
        changeOrigin: true,
        secure: false,
        rewrite: path => path.replace(/^\/api/, ''),
      },
    },
  },
})
