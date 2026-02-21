import { defineStore } from "pinia";
import { ref, computed } from "vue";
import request from "@/utils/request";

export const useSystemStore = defineStore("system", () => {
  const mode = ref("auth_request");
  const fetched = ref(false);

  const isReverseProxy = computed(() => mode.value === "reverse_proxy");

  async function fetchSettings() {
    if (fetched.value) return;
    try {
      const res: any = await request.get("/system/settings");
      mode.value = res.mode || "auth_request";
      fetched.value = true;
    } catch {
      // fallback to auth_request on error
    }
  }

  return { mode, isReverseProxy, fetchSettings };
});
