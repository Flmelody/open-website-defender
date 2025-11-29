/**
 * 配置工具 - 编译时固化配置
 * 
 * 配置在编译时从环境变量读取并固化到代码中
 * 前后端使用相同的环境变量：BACKEND_HOST, ROOT_PATH, ADMIN_PATH, GUARD_PATH
 */

// 编译时注入的配置（由 Vite define 替换）
const BUILD_CONFIG = {
  baseURL: typeof __BUILD_BACKEND_HOST__ !== 'undefined' ? __BUILD_BACKEND_HOST__ : 'http://localhost:9999/wall',
  rootPath: typeof __BUILD_ROOT_PATH__ !== 'undefined' ? __BUILD_ROOT_PATH__ : '/wall',
  guardPath: typeof __BUILD_GUARD_PATH__ !== 'undefined' ? __BUILD_GUARD_PATH__ : '/guard',
  adminPath: typeof __BUILD_ADMIN_PATH__ !== 'undefined' ? __BUILD_ADMIN_PATH__ : '/admin',
}

/**
 * 获取应用配置（编译时已固化）
 * @returns {{baseURL: string, rootPath: string, guardPath: string, adminPath: string}}
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
