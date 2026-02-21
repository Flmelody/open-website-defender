<script setup lang="ts">
import { computed } from "vue";
import { RouterView } from "vue-router";
import { ElConfigProvider } from "element-plus";
// @ts-ignore
import zhCn from "element-plus/dist/locale/zh-cn.mjs";
// @ts-ignore
import en from "element-plus/dist/locale/en.mjs";
// @ts-ignore
import de from "element-plus/dist/locale/de.mjs";
// @ts-ignore
import fr from "element-plus/dist/locale/fr.mjs";
// @ts-ignore
import ja from "element-plus/dist/locale/ja.mjs";
// @ts-ignore
import ru from "element-plus/dist/locale/ru.mjs";
import { useI18n } from "vue-i18n";

const { locale } = useI18n();

const localeMap: Record<string, any> = {
  zh: zhCn,
  en: en,
  de: de,
  fr: fr,
  ja: ja,
  ru: ru,
};

const elLocale = computed(() => {
  return localeMap[locale.value] || en;
});
</script>

<template>
  <el-config-provider :locale="elLocale">
    <RouterView />
  </el-config-provider>
</template>
