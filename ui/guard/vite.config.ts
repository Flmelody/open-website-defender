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

  // 构建时读取环境变量（会被打包到代码中）
  // 开发时从 process.env 读取，构建时固化到代码
  const backendHost = process.env.VITE_BACKEND_HOST || env.BACKEND_HOST || 'http://localhost:9999/wall'
  const rootPath = process.env.VITE_ROOT_PATH || env.ROOT_PATH || '/wall'
  const guardPath = process.env.VITE_GUARD_PATH || env.GUARD_PATH || '/guard'
  const adminPath = process.env.VITE_ADMIN_PATH || env.ADMIN_PATH || '/admin'

  console.log('Build config:', { backendHost, rootPath, guardPath, adminPath })

  // 提取后端地址（不含路径）用于 proxy
  const backendOrigin = new URL(backendHost).origin

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
          target: backendOrigin,
          changeOrigin: true,
          rewrite: (path) => path.replace(/^\/api/, '')
        }
      }
    },
    base: `${rootPath}${guardPath}`,
    // 定义全局常量，构建时会被替换（但生产环境由 Go 注入覆盖）
    define: {
      '__BUILD_BACKEND_HOST__': JSON.stringify(backendHost),
      '__BUILD_ROOT_PATH__': JSON.stringify(rootPath),
      '__BUILD_GUARD_PATH__': JSON.stringify(guardPath),
      '__BUILD_ADMIN_PATH__': JSON.stringify(adminPath),
    }
  }
})
