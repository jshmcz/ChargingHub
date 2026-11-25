import { defineConfig, loadEnv, type ProxyOptions } from 'vite'
import vue from '@vitejs/plugin-vue'
import { fileURLToPath, URL } from 'node:url'

const DEFAULT_API_BASE = '/api'
const API_TARGET = 'http://localhost:8080'

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')
  const apiBase = (env.VITE_API_BASE || DEFAULT_API_BASE).replace(/\/$/, '')

  const proxy: Record<string, ProxyOptions> = {
    '/api': {
      target: API_TARGET,
      changeOrigin: true,
    },
  }

  // 自定义 VITE_API_BASE（相对路径）时同样代理到本地后端
  if (apiBase.startsWith('/') && apiBase !== DEFAULT_API_BASE) {
    proxy[apiBase] = {
      target: API_TARGET,
      changeOrigin: true,
    }
  }

  return {
    base: '/m/',
    plugins: [vue()],
    resolve: {
      alias: {
        '@': fileURLToPath(new URL('./src', import.meta.url)),
      },
    },
    server: {
      host: '0.0.0.0',
      port: 5173,
      proxy,
    },
    preview: {
      port: 4173,
      proxy,
    },
  }
})
