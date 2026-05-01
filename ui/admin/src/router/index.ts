import { createRouter, createWebHistory } from "vue-router";
import { getAppConfig } from "@/utils/config";
import { useAuthStore } from "@/stores/auth";
import Layout from "@/views/Layout.vue";
import LoginView from "@/views/LoginView.vue";
import DashboardView from "@/views/DashboardView.vue";
import UserView from "@/views/UserView.vue";
import IpWhiteListView from "@/views/IpWhiteListView.vue";
import IpBlackListView from "@/views/IpBlackListView.vue";
import WafRulesView from "@/views/WafRulesView.vue";
import AccessLogView from "@/views/AccessLogView.vue";
import GeoBlockView from "@/views/GeoBlockView.vue";
import AuthorizedDomainView from "@/views/AuthorizedDomainView.vue";
import LicenseView from "@/views/LicenseView.vue";
import SystemSettingsView from "@/views/SystemSettingsView.vue";
import OAuthClientView from "@/views/OAuthClientView.vue";
import SecurityEventsView from "@/views/SecurityEventsView.vue";
import BotManagementView from "@/views/BotManagementView.vue";

const config = getAppConfig();
// Clean up double slashes if any
const base = `${config.rootPath}${config.adminPath}`.replace(/\/+/g, "/");

const router = createRouter({
  history: createWebHistory(base),
  routes: [
    {
      path: "/login",
      name: "login",
      component: LoginView,
      meta: { guest: true },
    },
    {
      path: "/",
      component: Layout,
      meta: { requiresAuth: true },
      children: [
        {
          path: "",
          redirect: "dashboard",
        },
        {
          path: "dashboard",
          name: "dashboard",
          component: DashboardView,
        },
        {
          path: "users",
          name: "users",
          component: UserView,
        },
        {
          path: "ip-white-list",
          name: "ip-white-list",
          component: IpWhiteListView,
        },
        {
          path: "ip-black-list",
          name: "ip-black-list",
          component: IpBlackListView,
        },
        {
          path: "waf-rules",
          name: "waf-rules",
          component: WafRulesView,
        },
        {
          path: "access-logs",
          name: "access-logs",
          component: AccessLogView,
        },
        {
          path: "geo-block",
          name: "geo-block",
          component: GeoBlockView,
        },
        {
          path: "authorized-domains",
          name: "authorized-domains",
          component: AuthorizedDomainView,
        },
        {
          path: "licenses",
          name: "licenses",
          component: LicenseView,
        },
        {
          path: "oauth-clients",
          name: "oauth-clients",
          component: OAuthClientView,
        },
        {
          path: "security-events",
          name: "security-events",
          component: SecurityEventsView,
        },
        {
          path: "bot-management",
          name: "bot-management",
          component: BotManagementView,
        },
        {
          path: "system-settings",
          name: "system-settings",
          component: SystemSettingsView,
        },
      ],
    },
  ],
});

router.beforeEach(async (to) => {
  const authStore = useAuthStore();

  if (to.meta.requiresAuth) {
    const authenticated =
      authStore.isLoggedIn || (await authStore.restoreSession());
    if (!authenticated) {
      return { name: "login" };
    }
    return true;
  }

  if (to.meta.guest) {
    const authenticated =
      authStore.isLoggedIn || (await authStore.restoreSession());
    if (authenticated) {
      return { name: "dashboard" };
    }
  }

  return true;
});

export default router;
