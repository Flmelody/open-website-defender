import { defineStore } from "pinia";
import { ref, computed } from "vue";
import request from "@/utils/request";

interface UserInfo {
  id: number;
  username: string;
}

interface AdminLoginResponse {
  requires_two_factor: boolean;
  challenge_token?: string;
  token?: string;
  user: UserInfo;
}

interface LoginResponse {
  token: string;
  user: UserInfo;
}

export const useAuthStore = defineStore("auth", () => {
  const token = ref(localStorage.getItem("token") || "");
  const user = ref<UserInfo | null>(null);
  const pendingChallengeToken = ref("");
  const requires2FA = ref(false);

  // Initialize user from localStorage if available
  const storedUser = localStorage.getItem("user");
  if (storedUser) {
    try {
      user.value = JSON.parse(storedUser);
    } catch (e) {
      console.error("Failed to parse stored user data", e);
      localStorage.removeItem("user");
    }
  }

  const isLoggedIn = computed(() => !!token.value);

  async function login(
    username: string,
    password: string,
  ): Promise<{ requires2FA: boolean }> {
    try {
      const res = await request.post<any, AdminLoginResponse>("/admin-login", {
        username,
        password,
      });

      if (res.requires_two_factor) {
        pendingChallengeToken.value = res.challenge_token || "";
        requires2FA.value = true;
        return { requires2FA: true };
      }

      if (res.token) {
        token.value = res.token;
        user.value = res.user;
        localStorage.setItem("token", res.token);
        localStorage.setItem("user", JSON.stringify(res.user));
      }
      return { requires2FA: false };
    } catch (error) {
      throw error;
    }
  }

  async function verify2FA(code: string): Promise<void> {
    try {
      const res = await request.post<any, LoginResponse>("/admin-login/2fa", {
        challenge_token: pendingChallengeToken.value,
        code,
      });

      if (res.token) {
        token.value = res.token;
        user.value = res.user;
        localStorage.setItem("token", res.token);
        localStorage.setItem("user", JSON.stringify(res.user));
      }
      pendingChallengeToken.value = "";
      requires2FA.value = false;
    } catch (error) {
      throw error;
    }
  }

  function cancelChallenge() {
    pendingChallengeToken.value = "";
    requires2FA.value = false;
  }

  function logout() {
    token.value = "";
    user.value = null;
    pendingChallengeToken.value = "";
    requires2FA.value = false;
    localStorage.removeItem("token");
    localStorage.removeItem("user");
  }

  return {
    token,
    user,
    isLoggedIn,
    pendingChallengeToken,
    requires2FA,
    login,
    verify2FA,
    cancelChallenge,
    logout,
  };
});
