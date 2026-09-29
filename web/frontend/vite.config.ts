import { defineConfig, loadEnv } from 'vite'
import react from '@vitejs/plugin-react'

// https://vite.dev/config/
export default defineConfig(({ mode }) => {
  // '.' is web/frontend: yarn and make run vite from there
  const env = loadEnv(mode, '.', '')

  return {
    plugins: [react()],
    server: {
      // the dev server forwards API calls to a locally running backend
      proxy: {
        '/api': env.VITE_API_PROXY || 'http://localhost:8888',
      },
    },
  }
})
