/// <reference types="vite/client" />

// 前端配置接口
interface AppConfig {
  baseURL: string;
}

// 扩展 Window 接口
interface Window {
  __APP_CONFIG__: AppConfig;
}
