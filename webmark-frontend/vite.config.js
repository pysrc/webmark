import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  server: {
    port: 3000,
    open: true,
    proxy: {
      '/wmapi': {
        target: 'http://127.0.0.1:11990',
        changeOrigin: true,
      }
    }
  },
  build: {
    outDir: 'build',  // 输出到 build 目录，与之前 react-scripts 保持一致
  }
})