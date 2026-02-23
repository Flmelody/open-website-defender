// 编译时注入的配置（路径由 Vite define 替换，baseURL/guardDomain 由运行时注入）
const BUILD_CONFIG = {
  baseURL: '/wall', // fallback for dev; overridden by runtime __APP_CONFIG__
  rootPath: typeof __BUILD_ROOT_PATH__ !== 'undefined' ? __BUILD_ROOT_PATH__ : '/wall',
  guardPath: typeof __BUILD_GUARD_PATH__ !== 'undefined' ? __BUILD_GUARD_PATH__ : '/guard',
  guardDomain: '', // runtime-only via __APP_CONFIG__
  adminPath: typeof __BUILD_ADMIN_PATH__ !== 'undefined' ? __BUILD_ADMIN_PATH__ : '/admin',
}

/**
 * 获取应用配置
 * @returns {{baseURL: string, rootPath: string, guardPath: string,guardDomain: string, adminPath: string}}
 */
export function getAppConfig() {
  const runtime = window.__APP_CONFIG__
  if (runtime) {
    return { ...BUILD_CONFIG, ...runtime }
  }
  return BUILD_CONFIG
}

/**
 * 获取页面路径
 * @param {'guard' | 'admin'} page - 页面类型
 * @returns {string} 完整路径，如 '/wall/guard'
 */
export function getPagePath(page) {
  const config = getAppConfig()
  const rootPath = config.rootPath || '/wall'
  const pagePath = page === 'guard'
    ? (config.guardPath || '/guard')
    : (config.adminPath || '/admin')
  return `${rootPath}${pagePath}`.replace(/\/+/g, '/')
}
