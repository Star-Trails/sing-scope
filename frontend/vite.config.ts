import vue from '@vitejs/plugin-vue'
import vueJsx from '@vitejs/plugin-vue-jsx'
import path from 'path'
import { defineConfig } from 'vite'

export default defineConfig({
  define: {
    __APP_VERSION__: JSON.stringify('0.1.0'),
    __COMMIT_ID__: JSON.stringify('c82b9b8'),
    __FONT__: JSON.stringify('none'),
  },
  base: './',
  worker: {
    format: 'es',
  },
  plugins: [
    vue(),
    vueJsx(),
  ],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
      net: path.resolve(__dirname, './src/helper/netShim.ts'),
    },
  },
  server: {
    port: 5173,
    strictPort: true,
  },
  build: {
    outDir: 'dist',
    emptyOutDir: true,
    target: 'esnext',
  },
})
