// vite.config.ts
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  base: '/',              // <- absolute base for domain root
  server: {
    port: 5174,
    strictPort: true,     // fail if port is already in use instead of auto-incrementing
  },
  build: {
    outDir: 'dist'
  }
})
