/// <reference types="vite/client" />
declare const __BUILD_ROOT_PATH__: string
declare const __BUILD_GUARD_PATH__: string
declare const __BUILD_ADMIN_PATH__: string

interface Window {
  __APP_CONFIG__?: import('./src/utils/config').AppConfig
}
