import axios from "axios";
import type { InternalAxiosRequestConfig, AxiosResponse } from "axios";
import { getAppConfig, getPagePath } from "./config";
import { ElMessage } from "element-plus";

export function isTokenExpired(token: string): boolean {
  try {
    const payload = JSON.parse(atob(token.split(".")[1]));
    if (!payload.exp) return false;
    // Consider expired 5 seconds early to avoid edge cases
    return Date.now() >= (payload.exp - 5) * 1000;
  } catch {
    return true;
  }
}

function handleExpiredToken() {
  localStorage.removeItem("token");
  window.location.href = `${getPagePath("admin")}/login`;
}

const request = axios.create({
  baseURL: getAppConfig().baseURL,
  timeout: 5000,
});

request.interceptors.request.use((config: InternalAxiosRequestConfig) => {
  const token = localStorage.getItem("token");
  if (token) {
    if (isTokenExpired(token) && !config.url?.includes("/login")) {
      handleExpiredToken();
      return Promise.reject(new axios.Cancel("Token expired"));
    }
    config.headers["Defender-Authorization"] = `Bearer ${token}`;
  }
  return config;
});

request.interceptors.response.use(
  (response: AxiosResponse) => {
    const { data } = response;
    if (data && typeof data === "object" && "code" in data) {
      if (data.code === 0) {
        return data.data;
      } else {
        const msg = data.error || data.message || "Request failed";
        ElMessage.error(msg);
        return Promise.reject(new Error(msg));
      }
    }
    return data;
  },
  (error) => {
    if (
      error.response?.status === 401 &&
      !error.config?.url?.includes("/login")
    ) {
      localStorage.removeItem("token");
      window.location.href = `${getPagePath("admin")}/login`;
    }

    const errorMessage =
      error.response?.data?.error ||
      error.response?.data?.message ||
      error.message ||
      "Network error";

    ElMessage.error(errorMessage);
    return Promise.reject(new Error(errorMessage));
  },
);

export default request;
