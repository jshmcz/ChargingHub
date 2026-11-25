import { defineConfig, loadEnv } from 'vite'
import vue from '@vitejs/plugin-vue'
import { fileURLToPath, URL } from 'node:url'

// VITE_API_BASE 默认为相对路径 /api，开发环境由 Vite 代理到后端。
// 若配置为完整 URL（例如生产环境的 https://api.example.com/api），则跳过代理。
export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')
  const apiBase = env.VITE_API_BASE || '/api'

  const proxy: Record<string, { target: string; changeOrigin: boolean }> = {}
  if (apiBase.startsWith('/')) {
    proxy[apiBase] = {
      target: 'http://localhost:8080',
      changeOrigin: true,
    }
  }

  return {
    plugins: [vue()],
    resolve: {
      alias: {
        '@': fileURLToPath(new URL('./src', import.meta.url)),
      },
    },
    server: {
      proxy,
    },
    preview: {
      proxy,
    },
  }
})
