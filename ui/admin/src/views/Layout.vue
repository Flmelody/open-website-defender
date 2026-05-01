<template>
  <el-container class="layout-container">
    <el-aside width="260px" class="aside">
      <div class="brand no-select">
        <span class="terminal-prompt">></span>
        <span class="brand-text">{{ t("common.brand") }}</span>
        <span class="cursor-blink">_</span>
      </div>

      <el-menu
        :default-active="route.path"
        class="sidebar-menu no-select"
        router
        text-color="var(--theme-text-dim)"
        active-text-color="var(--theme-accent)"
        background-color="transparent"
      >
        <div class="menu-label">{{ t("menu.system_modules") }}</div>
        <template v-for="item in filteredMenuItems" :key="item.index">
          <el-sub-menu v-if="item.children" :index="item.index">
            <template #title>
              <el-icon><component :is="item.icon" /></el-icon>
              <span>{{ t(item.label) }}</span>
            </template>
            <el-menu-item
              v-for="child in item.children"
              :key="child.index"
              :index="child.index"
            >
              <el-icon><component :is="child.icon" /></el-icon>
              <span>{{ t(child.label) }}</span>
            </el-menu-item>
          </el-sub-menu>
          <el-menu-item v-else :index="item.index">
            <el-icon><component :is="item.icon" /></el-icon>
            <span>{{ t(item.label) }}</span>
          </el-menu-item>
        </template>
      </el-menu>

      <div class="sidebar-footer no-select">
        <div class="sys-status">
          <div class="status-line">
            <span class="status-dot"></span>
            <span class="status-text">{{ t("layout.net_status") }}</span>
          </div>
          <div class="user-line">
            <span class="user-label">{{ t("layout.operator") }}</span>
            <span class="user-name">{{
              authStore.user?.username?.toUpperCase() || t("common.unknown")
            }}</span>
          </div>
          <div class="lang-line">
            <span class="lang-label">LANG:</span>
            <el-dropdown
              trigger="click"
              @command="handleLangCommand"
              class="lang-dropdown"
              popper-class="terminal-popper"
            >
              <span class="lang-switch">
                {{ currentLangLabel }}
                <el-icon class="el-icon--right"><arrow-down /></el-icon>
              </span>
              <template #dropdown>
                <el-dropdown-menu class="terminal-dropdown">
                  <el-dropdown-item
                    command="en"
                    :class="{ active: locale === 'en' }"
                    >EN - English</el-dropdown-item
                  >
                  <el-dropdown-item
                    command="zh"
                    :class="{ active: locale === 'zh' }"
                    >CN - 简体中文</el-dropdown-item
                  >
                  <el-dropdown-item
                    command="de"
                    :class="{ active: locale === 'de' }"
                    >DE - Deutsch</el-dropdown-item
                  >
                  <el-dropdown-item
                    command="fr"
                    :class="{ active: locale === 'fr' }"
                    >FR - Français</el-dropdown-item
                  >
                  <el-dropdown-item
                    command="ja"
                    :class="{ active: locale === 'ja' }"
                    >JP - 日本語</el-dropdown-item
                  >
                  <el-dropdown-item
                    command="ru"
                    :class="{ active: locale === 'ru' }"
                    >RU - Русский</el-dropdown-item
                  >
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </div>
          <div class="theme-line">
            <span class="theme-label">{{ t("layout.theme") }}:</span>
            <el-dropdown
              trigger="click"
              @command="handleThemeCommand"
              class="theme-dropdown"
              popper-class="terminal-popper"
            >
              <span class="theme-switch">
                <span
                  class="theme-swatch"
                  :style="{ backgroundColor: currentThemeOption.accent }"
                ></span>
                {{ currentThemeOption.label }}
                <el-icon class="el-icon--right"><arrow-down /></el-icon>
              </span>
              <template #dropdown>
                <el-dropdown-menu class="terminal-dropdown">
                  <el-dropdown-item
                    v-for="theme in themeOptions"
                    :key="theme.id"
                    :command="theme.id"
                    :class="{ active: currentTheme === theme.id }"
                  >
                    <span
                      class="theme-swatch menu-swatch"
                      :style="{ backgroundColor: theme.accent }"
                    ></span>
                    {{ theme.label }}
                  </el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </div>
          <el-button link class="logout-btn" @click="handleLogout">
            {{ t("layout.terminate_session") }}
          </el-button>
        </div>
      </div>
    </el-aside>

    <el-container class="main-container">
      <el-main class="main-content">
        <RouterView />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup lang="ts">
import { RouterView, useRoute, useRouter } from "vue-router";
import { useAuthStore } from "@/stores/auth";
import { useSystemStore } from "@/stores/system";
import { menuItems } from "@/config/menu";
import type { MenuItem } from "@/config/menu";
import { useI18n } from "vue-i18n";
import { ArrowDown } from "@element-plus/icons-vue";
import { computed, onMounted } from "vue";
import { useTheme } from "@/composables/useTheme";

const route = useRoute();
const router = useRouter();
const authStore = useAuthStore();
const systemStore = useSystemStore();
const { t, locale } = useI18n();
const { themeOptions, currentTheme, setTheme } = useTheme();

onMounted(() => {
  systemStore.fetchSettings();
});

const filteredMenuItems = computed(() => {
  const mode = systemStore.mode;
  return menuItems.reduce<MenuItem[]>((acc, item) => {
    if (item.requireMode && item.requireMode !== mode) return acc;
    if (item.children) {
      const children = item.children.filter(
        (c) => !c.requireMode || c.requireMode === mode,
      );
      if (children.length === 0) return acc;
      acc.push({ ...item, children });
    } else {
      acc.push(item);
    }
    return acc;
  }, []);
});

const langMap: Record<string, string> = {
  en: "EN",
  zh: "CN",
  de: "DE",
  fr: "FR",
  ja: "JP",
  ru: "RU",
};

const currentLangLabel = computed(() => langMap[locale.value] || "EN");
const currentThemeOption = computed(
  () =>
    themeOptions.find((theme) => theme.id === currentTheme.value) ||
    themeOptions[0],
);

const handleLogout = async () => {
  await authStore.logout();
  router.push("/login");
};

const handleLangCommand = (command: string) => {
  locale.value = command;
  localStorage.setItem("locale", command);
};

const handleThemeCommand = (command: string | number | object) => {
  setTheme(String(command));
};
</script>

<style scoped>
.layout-container {
  height: 100vh;
  background-color: var(--theme-bg);
}

.aside {
  background-color: rgba(var(--theme-panel-rgb), 0.95);
  display: flex;
  flex-direction: column;
  border-right: 1px solid var(--theme-border-soft);
  position: relative;
  z-index: 100;
  box-shadow: 5px 0 20px rgba(0, 0, 0, 0.3);
}

.brand {
  height: 70px;
  display: flex;
  align-items: center;
  padding: 0 24px;
  color: var(--theme-accent);
  font-family: "Courier New", monospace;
  font-weight: 900;
  font-size: 18px;
  border-bottom: 1px solid var(--theme-border-soft);
  letter-spacing: 1px;
  text-shadow: 0 0 10px rgba(var(--theme-accent-rgb), 0.4);
  background: rgba(var(--theme-panel-rgb), 0.1);
}

.terminal-prompt {
  margin-right: 10px;
  color: #fff;
}

.cursor-blink {
  animation: blink 1s step-end infinite;
  margin-left: 5px;
  color: var(--theme-accent);
}

@keyframes blink {
  50% {
    opacity: 0;
  }
}

.sidebar-menu {
  flex: 1;
  border-right: none;
  padding-top: 30px;
}

.menu-label {
  padding: 0 24px 10px;
  color: var(--theme-accent-muted);
  font-size: 12px;
  font-weight: bold;
  letter-spacing: 1px;
}

:deep(.el-menu-item) {
  font-family: "Courier New", monospace;
  height: 48px;
  line-height: 48px;
  margin: 4px 0;
  border-left: 3px solid transparent;
  font-weight: bold;
  font-size: 14px;
}

:deep(.el-menu-item:hover) {
  background-color: rgba(var(--theme-accent-rgb), 0.1) !important;
  color: #fff !important;
}

:deep(.el-menu-item.is-active) {
  background-color: rgba(var(--theme-accent-rgb), 0.15) !important;
  border-left: 3px solid var(--theme-accent);
  color: var(--theme-accent) !important;
  text-shadow: 0 0 5px rgba(var(--theme-accent-rgb), 0.3);
}

:deep(.el-sub-menu__title) {
  font-family: "Courier New", monospace;
  height: 48px;
  line-height: 48px;
  border-left: 3px solid transparent;
  font-weight: bold;
  font-size: 14px;
  color: var(--theme-text-dim) !important;
}

:deep(.el-sub-menu__title:hover) {
  background-color: rgba(var(--theme-accent-rgb), 0.1) !important;
  color: #fff !important;
}

:deep(.el-sub-menu.is-active > .el-sub-menu__title) {
  color: var(--theme-accent) !important;
}

:deep(.el-sub-menu .el-menu-item) {
  padding-left: 56px !important;
  font-size: 13px;
  height: 42px;
  line-height: 42px;
}

:deep(.el-sub-menu__title .el-sub-menu__icon-arrow) {
  color: var(--theme-text-dim);
}

.sidebar-footer {
  padding: 24px;
  border-top: 1px solid var(--theme-border-soft);
  font-family: "Courier New", monospace;
  font-size: 12px;
  background: rgba(var(--theme-panel-rgb), 0.3);
}

.sys-status {
  color: var(--theme-text-dim);
}

.status-line {
  display: flex;
  align-items: center;
  margin-bottom: 12px;
  color: var(--theme-accent);
  font-weight: bold;
}

.status-dot {
  width: 8px;
  height: 8px;
  background-color: var(--theme-accent);
  border-radius: 50%;
  margin-right: 10px;
  box-shadow: 0 0 8px var(--theme-accent);
}

.user-line {
  margin-bottom: 12px;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.user-label {
  color: var(--theme-accent-muted);
  font-size: 10px;
}

.user-name {
  color: #fff;
  font-weight: bold;
  font-size: 14px;
}

.lang-line,
.theme-line {
  margin-bottom: 12px;
  display: flex;
  align-items: center;
  gap: 8px;
  color: var(--theme-accent-muted);
  font-size: 10px;
}

.theme-line {
  margin-bottom: 20px;
}

.lang-label,
.theme-label {
  color: var(--theme-accent-muted);
}

.lang-switch,
.theme-switch {
  color: var(--theme-text-dim);
  cursor: pointer;
  font-weight: bold;
  font-size: 12px;
  display: flex;
  align-items: center;
  gap: 4px;
}

.lang-switch:hover,
.theme-switch:hover {
  color: var(--theme-accent);
  text-shadow: 0 0 5px rgba(var(--theme-accent-rgb), 0.5);
}

.theme-swatch {
  width: 10px;
  height: 10px;
  border: 1px solid rgba(255, 255, 255, 0.4);
  box-shadow: 0 0 8px currentColor;
  display: inline-block;
  flex-shrink: 0;
}

.menu-swatch {
  margin-right: 6px;
}

.logout-btn {
  color: #ff5555;
  padding: 0;
  font-size: 12px;
  font-family: inherit;
  width: 100%;
  text-align: left;
  justify-content: flex-start;
}

.logout-btn:hover {
  color: #ff7875;
  text-shadow: 0 0 8px rgba(255, 85, 85, 0.6);
  text-decoration: none;
}

.main-container {
  background-color: var(--theme-bg);
  /* Grid pattern defined in body also applies here visually */
}

.main-content {
  padding: 40px;
  background-color: transparent;
}
</style>

<style>
.terminal-popper.el-popper {
  background: rgba(var(--theme-panel-rgb), 0.95) !important;
  border: 1px solid var(--theme-border-soft) !important;
}

.terminal-popper .el-dropdown-menu {
  background: transparent !important;
  border: none !important;
  padding: 5px 0 !important;
}

.terminal-popper .el-dropdown-menu__item {
  color: var(--theme-text-dim) !important;
  font-family: "Courier New", monospace !important;
  font-size: 12px !important;
}

.terminal-popper .el-dropdown-menu__item:hover,
.terminal-popper .el-dropdown-menu__item:focus {
  background-color: rgba(var(--theme-accent-rgb), 0.1) !important;
  color: #fff !important;
}

.terminal-popper .el-dropdown-menu__item.active {
  color: var(--theme-accent) !important;
  font-weight: bold !important;
}

.terminal-popper .el-popper__arrow::before {
  background: rgba(var(--theme-panel-rgb), 0.95) !important;
  border: 1px solid var(--theme-border-soft) !important;
}
</style>
