import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react({
    babel: {
      plugins: [
        ['@babel/plugin-transform-runtime', {
          corejs: 3,
          helpers: true,
          regenerator: true,
          useESModules: true
        }]
      ]
    }
  })],
  server: {
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true
      },
      '/wx-images': {
        target: 'http://localhost:8080',
        changeOrigin: true
      },
      '/wx-qim': {
        target: 'http://localhost:8080',
        changeOrigin: true
      },
      '/wx-mp': {
        target: 'http://localhost:8080',
        changeOrigin: true
      }
    }
  }
})