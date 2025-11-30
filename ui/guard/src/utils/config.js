// 编译时注入的配置（由 Vite define 替换）
const BUILD_CONFIG = {
  baseURL: typeof __BUILD_BACKEND_HOST__ !== 'undefined' ? __BUILD_BACKEND_HOST__ : 'http://localhost:9999/wall',
  rootPath: typeof __BUILD_ROOT_PATH__ !== 'undefined' ? __BUILD_ROOT_PATH__ : '/wall',
  guardPath: typeof __BUILD_GUARD_PATH__ !== 'undefined' ? __BUILD_GUARD_PATH__ : '/guard',
  guardDomain: typeof __BUILD_GUARD_DOMAIN__ !== 'undefined' ? __BUILD_GUARD_DOMAIN__ : '/guard',
  adminPath: typeof __BUILD_ADMIN_PATH__ !== 'undefined' ? __BUILD_ADMIN_PATH__ : '/admin',
}

/**
 * 获取应用配置
 * @returns {{baseURL: string, rootPath: string, guardPath: string,guardDomain: string, adminPath: string}}
 */
export function getAppConfig() {
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
