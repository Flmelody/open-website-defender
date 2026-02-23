// 编译时注入的配置（路径由 Vite define 替换，baseURL 由运行时注入）
const BUILD_CONFIG = {
  baseURL: "/wall", // fallback for dev; overridden by runtime __APP_CONFIG__
  rootPath:
    typeof __BUILD_ROOT_PATH__ !== "undefined" ? __BUILD_ROOT_PATH__ : "/wall",
  guardPath:
    typeof __BUILD_GUARD_PATH__ !== "undefined"
      ? __BUILD_GUARD_PATH__
      : "/guard",
  adminPath:
    typeof __BUILD_ADMIN_PATH__ !== "undefined"
      ? __BUILD_ADMIN_PATH__
      : "/admin",
};

export interface AppConfig {
  baseURL: string;
  rootPath: string;
  guardPath: string;
  adminPath: string;
}

export function getAppConfig(): AppConfig {
  const runtime = (window as any).__APP_CONFIG__ as Partial<AppConfig> | undefined;
  if (runtime) {
    return { ...BUILD_CONFIG, ...runtime };
  }
  return BUILD_CONFIG;
}

export function getPagePath(page: "guard" | "admin"): string {
  const config = getAppConfig();
  const rootPath = config.rootPath || "/wall";
  const pagePath =
    page === "guard"
      ? config.guardPath || "/guard"
      : config.adminPath || "/admin";
  return `${rootPath}${pagePath}`.replace(/\/+/g, "/");
}
