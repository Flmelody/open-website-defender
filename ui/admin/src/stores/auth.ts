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
  user: UserInfo;
}

interface LoginResponse {
  user: UserInfo;
}

export const useAuthStore = defineStore("auth", () => {
  const user = ref<UserInfo | null>(null);
  const pendingChallengeToken = ref("");
  const requires2FA = ref(false);
  const sessionChecked = ref(false);

  localStorage.removeItem("token");
  localStorage.removeItem("user");

  const isLoggedIn = computed(() => !!user.value);

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

      if (res.user) {
        user.value = res.user;
        sessionChecked.value = true;
      }
      return { requires2FA: false };
    } catch (error) {
      throw error;
    }
  }

  async function verify2FA(code: string, trustDevice: boolean = false): Promise<void> {
    try {
      const res = await request.post<any, LoginResponse>("/admin-login/2fa", {
        challenge_token: pendingChallengeToken.value,
        code,
        trust_device: trustDevice,
      });

      if (res.user) {
        user.value = res.user;
        sessionChecked.value = true;
      }
      pendingChallengeToken.value = "";
      requires2FA.value = false;
    } catch (error) {
      throw error;
    }
  }

  async function restoreSession(force = false): Promise<boolean> {
    if (user.value) return true;
    if (sessionChecked.value && !force) return false;

    try {
      const res = await request.get<any, { user: UserInfo }>("/admin-session", {
        skipAuthRedirect: true,
        skipErrorMessage: true,
      });
      user.value = res.user;
      return true;
    } catch {
      user.value = null;
      return false;
    } finally {
      sessionChecked.value = true;
    }
  }

  function cancelChallenge() {
    pendingChallengeToken.value = "";
    requires2FA.value = false;
  }

  async function logout() {
    try {
      await request.post("/logout", null, {
        skipAuthRedirect: true,
        skipErrorMessage: true,
      });
    } catch {
      // Local logout still proceeds if the backend call fails.
    }
    user.value = null;
    pendingChallengeToken.value = "";
    requires2FA.value = false;
    sessionChecked.value = true;
  }

  return {
    user,
    isLoggedIn,
    pendingChallengeToken,
    requires2FA,
    restoreSession,
    login,
    verify2FA,
    cancelChallenge,
    logout,
  };
});
