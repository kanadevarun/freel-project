import { defineConfig } from 'vitest/config'
import react from '@vitejs/plugin-react'

/**
 * Dedicated Vitest configuration.
 * Separated from vite.config.js to ensure Vitest picks up
 * this config with full jsdom support on Node 22+/26.
 */
export default defineConfig({
  plugins: [react()],
  test: {
    environment: 'jsdom',
    globals: true,
    setupFiles: ['./src/__tests__/setup.js'],
    css: false,
    // Disable Node 22+/26's experimental Web Storage API which shadows
    // jsdom's localStorage implementation causing tests to fail.
    execArgv: ['--no-webstorage'],
    environmentOptions: {
      jsdom: {
        url: 'http://localhost:3000',
      },
    },
    coverage: {
      provider: 'v8',
      reporter: ['text', 'html'],
      include: ['src/**/*.{js,jsx}'],
      exclude: ['src/__tests__/**', 'src/main.jsx'],
    },
  },
})
