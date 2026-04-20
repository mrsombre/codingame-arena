import {readdirSync, statSync} from 'node:fs'
import {resolve} from 'node:path'
import tailwindcss from '@tailwindcss/vite'
import react from '@vitejs/plugin-react'
import {defineConfig} from 'vite'

const packagesDir = resolve(__dirname, 'packages')
const input = {main: resolve(packagesDir, 'index.html')}

for (const name of readdirSync(packagesDir)) {
  const html = resolve(packagesDir, name, 'index.html')
  try {
    if (statSync(html).isFile()) {
      input[name] = html
    }
  } catch {}
}

const apiTarget = process.env.ARENA_API ?? 'http://localhost:5757'

export default defineConfig({
  plugins: [react(), tailwindcss()],
  root: 'packages',
  resolve: {
    alias: {
      '@shared': resolve(__dirname, 'packages/shared')
    }
  },
  build: {
    outDir: '../dist',
    emptyOutDir: true,
    rollupOptions: {input}
  },
  server: {
    port: 3000,
    proxy: {
      '/api': {target: apiTarget, changeOrigin: true},
      '/healthz': {target: apiTarget, changeOrigin: true}
    }
  }
})
