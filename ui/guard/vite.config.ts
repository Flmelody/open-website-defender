import {fileURLToPath, URL} from 'node:url'
import path from 'node:path'

import {defineConfig, loadEnv} from 'vite'
import vue from '@vitejs/plugin-vue'
import vueDevTools from 'vite-plugin-vue-devtools'

// https://vite.dev/config/
export default defineConfig(({mode}) => {
  // 从根目录读取 .env 文件（与 Go 后端共享）
  const rootDir = path.resolve(__dirname, '../..')
  const env = loadEnv(mode, rootDir, '')

  // 构建时读取环境变量（路径会被打包到代码中，BACKEND_HOST/GUARD_DOMAIN 由运行时注入）
  const rootPath = process.env.VITE_ROOT_PATH || env.ROOT_PATH || '/wall'
  const guardPath = process.env.VITE_GUARD_PATH || env.GUARD_PATH || '/guard'
  const adminPath = process.env.VITE_ADMIN_PATH || env.ADMIN_PATH || '/admin'

  console.log('Build config:', {rootPath, guardPath, adminPath})

  // Dev proxy target (only used in dev server, not baked into build)
  const devBackend = process.env.VITE_BACKEND_HOST || env.BACKEND_HOST || 'http://localhost:9999'

  return {
    plugins: [
      vue(),
      vueDevTools(),
    ],
    resolve: {
      alias: {
        '@': fileURLToPath(new URL('./src', import.meta.url))
      },
    },
    server: {
      proxy: {
        '/api': {
          target: devBackend,
          changeOrigin: true,
          rewrite: (path) => path.replace(/^\/api/, '')
        }
      }
    },
    base: `${rootPath}${guardPath}`,
    define: {
      '__BUILD_ROOT_PATH__': JSON.stringify(rootPath),
      '__BUILD_GUARD_PATH__': JSON.stringify(guardPath),
      '__BUILD_ADMIN_PATH__': JSON.stringify(adminPath),
    }
  }
})
