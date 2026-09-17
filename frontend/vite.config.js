import { svelte } from '@sveltejs/vite-plugin-svelte'
import { defineConfig } from 'vite'
import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))

// Garantiza compatibilidad automática con Cloudflare Pages replicando dist en out
function syncOutputToOutPlugin() {
  return {
    name: 'sync-output-to-out',
    closeBundle() {
      const distDir = path.resolve(__dirname, 'dist')
      const outDir = path.resolve(__dirname, 'out')
      if (fs.existsSync(distDir)) {
        fs.cpSync(distDir, outDir, { recursive: true })
      }
    }
  }
}

// https://vite.dev/config/
export default defineConfig({
  plugins: [svelte(), syncOutputToOutPlugin()],
  server: {
    host: true, // Expone el servidor en la red local (0.0.0.0)
    port: 5173,
  },
})

