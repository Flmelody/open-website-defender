// 编译时注入的配置（由 Vite define 替换）
const BUILD_CONFIG = {
    baseURL: typeof __BUILD_BACKEND_HOST__ !== 'undefined' ? __BUILD_BACKEND_HOST__ : 'http://localhost:9999/wall',
    rootPath: typeof __BUILD_ROOT_PATH__ !== 'undefined' ? __BUILD_ROOT_PATH__ : '/wall',
    guardPath: typeof __BUILD_GUARD_PATH__ !== 'undefined' ? __BUILD_GUARD_PATH__ : '/guard',
    adminPath: typeof __BUILD_ADMIN_PATH__ !== 'undefined' ? __BUILD_ADMIN_PATH__ : '/admin',
}

export interface AppConfig {
    baseURL: string;
    rootPath: string;
    guardPath: string;
    adminPath: string;
}

export function getAppConfig(): AppConfig {
    return BUILD_CONFIG
}

export function getPagePath(page: 'guard' | 'admin'): string {
    const config = getAppConfig()
    const rootPath = config.rootPath || '/wall'
    const pagePath = page === 'guard'
        ? (config.guardPath || '/guard')
        : (config.adminPath || '/admin')
    return `${rootPath}${pagePath}`.replace(/\/+/g, '/')
}

