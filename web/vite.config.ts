import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'
import { defineConfig } from 'vite'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react(), tailwindcss()],
  server: {
    proxy: {
      '/price': 'http://localhost:8080',
      '/health': 'http://localhost:8080',
      '/config': 'http://localhost:8080',
      '/metrics': 'http://localhost:8080',
    },
  },
})
