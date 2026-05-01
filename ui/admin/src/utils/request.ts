import axios from "axios";
import type { InternalAxiosRequestConfig, AxiosResponse } from "axios";
import { getAppConfig, getPagePath } from "./config";
import { ElMessage } from "element-plus";

declare module "axios" {
  export interface AxiosRequestConfig {
    skipAuthRedirect?: boolean;
    skipErrorMessage?: boolean;
  }
}

const request = axios.create({
  baseURL: getAppConfig().baseURL,
  timeout: 5000,
  withCredentials: true,
});

request.interceptors.request.use((config: InternalAxiosRequestConfig) => {
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
    const requestConfig = error.config || {};
    if (
      error.response?.status === 401 &&
      !requestConfig.skipAuthRedirect &&
      !requestConfig.url?.includes("/login")
    ) {
      window.location.href = `${getPagePath("admin")}/login`;
    }

    const errorMessage =
      error.response?.data?.error ||
      error.response?.data?.message ||
      error.message ||
      "Network error";

    if (!requestConfig.skipErrorMessage) {
      ElMessage.error(errorMessage);
    }
    return Promise.reject(new Error(errorMessage));
  },
);

export default request;
