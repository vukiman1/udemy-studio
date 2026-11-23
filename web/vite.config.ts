import react from '@vitejs/plugin-react'
import { defineConfig } from 'vitest/config'

const GO_DEV_SERVER = 'http://127.0.0.1:8765'

export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      '/api': GO_DEV_SERVER,
      '/healthz': GO_DEV_SERVER,
      '/media': GO_DEV_SERVER,
      '/articles': GO_DEV_SERVER,
      '/attachments': GO_DEV_SERVER,
    },
  },
  test: {
    environment: 'jsdom',
    setupFiles: ['./src/test-setup.ts'],
  },
})
