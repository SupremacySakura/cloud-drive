import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import tailwindcss from '@tailwindcss/vite'
import vueJsx from '@vitejs/plugin-vue-jsx'
// https://vite.dev/config/
export default defineConfig({
  plugins: [vue(), tailwindcss(), vueJsx()],
  server:{
    proxy:{
      '/api':{
        target:'http://localhost:9000',
        changeOrigin:true,
        secure:false,
        rewrite:(path) => path.replace(/^\/api/,'')
      }
    }
  }
})
