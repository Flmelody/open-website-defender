<template>
  <div class="system-settings-view">
    <div class="terminal-card glass-panel">
      <div class="card-header no-select">
        <div class="header-left">
          <span class="prefix">root@system:~/config$</span>
          <span class="command blink-cursor">./system_settings.sh</span>
        </div>
        <div class="header-right">
          <el-button size="small" @click="resetForm">{{
            t("system.reset")
          }}</el-button>
          <el-button
            size="small"
            type="primary"
            :loading="saving"
            @click="handleSave"
            >{{ t("system.save") }}</el-button
          >
          <el-button size="small" @click="fetchData">{{
            t("common.refresh")
          }}</el-button>
        </div>
      </div>

      <div class="settings-content" v-loading="loading">
        <el-form
          ref="formRef"
          :model="form"
          :rules="rules"
          label-position="top"
          class="hacker-form"
        >
          <div class="settings-section">
            <div class="section-title no-select">
              <span class="prefix">&gt;</span> {{ t("system.section_headers") }}
              <el-tooltip
                :content="t('system.headers_desc')"
                placement="right"
                effect="dark"
              >
                <el-icon class="info-icon"><InfoFilled /></el-icon>
              </el-tooltip>
            </div>

            <el-form-item
              :label="'> ' + t('system.git_token_header')"
              prop="git_token_header"
            >
              <el-input
                v-model="form.git_token_header"
                placeholder="Defender-Git-Token"
              />
            </el-form-item>

            <el-form-item
              :label="'> ' + t('system.license_header')"
              prop="license_header"
            >
              <el-input
                v-model="form.license_header"
                placeholder="Defender-License"
              />
            </el-form-item>
          </div>

          <div class="settings-section separator">
            <div class="section-title no-select">
              <span class="prefix">&gt;</span>
              {{ t("system.section_bot_protection") }}
              <el-tooltip
                :content="t('system.bot_protection_desc')"
                placement="right"
                effect="dark"
                :show-after="200"
              >
                <el-icon class="info-icon"><InfoFilled /></el-icon>
              </el-tooltip>
            </div>

            <!-- JS Challenge sub-section -->
            <div class="sub-section-title no-select">
              <span class="prefix">$</span> {{ t("js_challenge.title") }}
              <el-tooltip
                :content="t('js_challenge.desc')"
                placement="right"
                effect="dark"
                :show-after="200"
              >
                <el-icon class="info-icon"><InfoFilled /></el-icon>
              </el-tooltip>
            </div>

            <div class="inline-field">
              <span class="field-label dim-text"
                >{{ t("js_challenge.enabled") }}:</span
              >
              <el-switch v-model="form.js_challenge_enabled" />
            </div>

            <el-form-item v-if="form.js_challenge_enabled">
              <template #label>
                <span class="label-with-tip"
                  >> {{ t("js_challenge.mode") }}
                  <el-tooltip
                    :content="t('js_challenge.mode_desc')"
                    placement="right"
                    effect="dark"
                  >
                    <el-icon class="info-icon"><InfoFilled /></el-icon>
                  </el-tooltip>
                </span>
              </template>
              <el-select v-model="form.js_challenge_mode" style="width: 200px">
                <el-option :label="t('js_challenge.mode_off')" value="off" />
                <el-option
                  :label="t('js_challenge.mode_suspicious')"
                  value="suspicious"
                />
                <el-option :label="t('js_challenge.mode_all')" value="all" />
              </el-select>
            </el-form-item>

            <el-form-item v-if="form.js_challenge_enabled">
              <template #label>
                <span class="label-with-tip"
                  >> {{ t("js_challenge.difficulty") }}
                  <el-tooltip
                    :content="t('js_challenge.difficulty_desc')"
                    placement="right"
                    effect="dark"
                  >
                    <el-icon class="info-icon"><InfoFilled /></el-icon>
                  </el-tooltip>
                </span>
              </template>
              <el-slider
                v-model="form.js_challenge_difficulty"
                :min="1"
                :max="6"
                :step="1"
                show-stops
                style="max-width: 280px"
              />
            </el-form-item>

            <!-- Bot Detection sub-section -->
            <div class="sub-section-title no-select" style="margin-top: 20px">
              <span class="prefix">$</span>
              {{ t("system.section_bot_detection") }}
              <el-tooltip
                :content="t('system.bot_detection_desc')"
                placement="right"
                effect="dark"
                :show-after="200"
              >
                <el-icon class="info-icon"><InfoFilled /></el-icon>
              </el-tooltip>
            </div>

            <div class="inline-field">
              <span class="field-label dim-text"
                >{{ t("system.bot_management_enabled") }}:</span
              >
              <el-switch v-model="form.bot_management_enabled" />
            </div>

            <template v-if="form.bot_management_enabled">
              <div class="inline-field">
                <span class="field-label dim-text"
                  >{{ t("system.challenge_escalation") }}:</span
                >
                <el-switch v-model="form.challenge_escalation" />
                <el-tooltip
                  :content="t('system.challenge_escalation_desc')"
                  placement="right"
                  effect="dark"
                >
                  <el-icon class="info-icon"><InfoFilled /></el-icon>
                </el-tooltip>
              </div>

              <el-form-item :label="'> ' + t('system.captcha_provider')">
                <el-select v-model="form.captcha_provider" style="width: 200px">
                  <el-option
                    :label="t('system.captcha_provider_builtin')"
                    value="builtin"
                  />
                  <el-option
                    :label="t('system.captcha_provider_turnstile')"
                    value="turnstile"
                  />
                </el-select>
              </el-form-item>

              <template v-if="form.captcha_provider !== 'builtin'">
                <el-form-item
                  :label="'> ' + t('system.captcha_site_key')"
                  prop="captcha_site_key"
                  :rules="[
                    {
                      required: true,
                      message: t('system.captcha_site_key_required'),
                      trigger: 'blur',
                    },
                  ]"
                >
                  <el-input v-model="form.captcha_site_key" />
                </el-form-item>

                <el-form-item
                  :label="'> ' + t('system.captcha_secret_key')"
                  prop="captcha_secret_key"
                  :rules="[
                    {
                      required: true,
                      message: t('system.captcha_secret_key_required'),
                      trigger: 'blur',
                    },
                  ]"
                >
                  <el-input
                    v-model="form.captcha_secret_key"
                    type="password"
                    show-password
                  />
                </el-form-item>
              </template>

              <el-form-item :label="'> ' + t('system.captcha_cookie_ttl')">
                <el-input-number
                  v-model="form.captcha_cookie_ttl"
                  :min="60"
                  :max="604800"
                  :step="3600"
                />
              </el-form-item>
            </template>
          </div>

          <div class="settings-section separator">
            <div class="section-title no-select">
              <span class="prefix">&gt;</span> {{ t("webhook.title") }}
              <el-tooltip
                :content="t('webhook.desc')"
                placement="right"
                effect="dark"
                :show-after="200"
              >
                <el-icon class="info-icon"><InfoFilled /></el-icon>
              </el-tooltip>
            </div>

            <el-form-item :label="'> ' + t('webhook.url')">
              <el-input
                v-model="form.webhook_url"
                :placeholder="t('webhook.url_placeholder')"
              />
            </el-form-item>
          </div>
        </el-form>

        <!-- Semantic Analysis -->
        <div class="settings-section separator">
          <div class="section-title no-select">
            <span class="prefix">&gt;</span> {{ t("system.section_semantic_analysis") }}
            <el-tooltip
              :content="t('system.semantic_analysis_desc')"
              placement="right"
              effect="dark"
              :show-after="200"
            >
              <el-icon class="info-icon"><InfoFilled /></el-icon>
            </el-tooltip>
          </div>

          <el-form label-position="top" class="hacker-form">
            <el-form-item :label="t('system.semantic_analysis_enabled')">
              <el-switch v-model="form.semantic_analysis_enabled" />
            </el-form-item>
          </el-form>
        </div>

        <!-- Cache Management -->
        <div class="settings-section separator">
          <div class="section-title no-select">
            <span class="prefix">&gt;</span> {{ t("system.section_cache") }}
            <el-tooltip
              :content="t('system.cache_desc')"
              placement="right"
              effect="dark"
              :show-after="200"
            >
              <el-icon class="info-icon"><InfoFilled /></el-icon>
            </el-tooltip>
          </div>

          <el-form label-position="top" class="hacker-form">
            <el-form-item>
              <template #label>
                <span class="label-with-tip"
                  >> {{ t("system.cache_sync_interval") }}
                  <el-tooltip
                    :content="t('system.cache_sync_desc')"
                    placement="right"
                    effect="dark"
                  >
                    <el-icon class="info-icon"><InfoFilled /></el-icon>
                  </el-tooltip>
                </span>
              </template>
              <el-input-number
                v-model="form.cache_sync_interval"
                :min="0"
                :max="300"
              />
            </el-form-item>
          </el-form>

          <el-button
            type="danger"
            :loading="clearing"
            @click="handleClearCache"
            >{{ t("system.clear_cache") }}</el-button
          >
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, reactive, computed } from "vue";
import request from "@/utils/request";
import { ElMessage, ElMessageBox } from "element-plus";
import { InfoFilled } from "@element-plus/icons-vue";
import { useI18n } from "vue-i18n";

const { t } = useI18n();
const loading = ref(false);
const saving = ref(false);
const clearing = ref(false);
const formRef = ref();

const form = reactive({
  git_token_header: "",
  license_header: "",
  js_challenge_enabled: false,
  js_challenge_mode: "suspicious",
  js_challenge_difficulty: 4,
  webhook_url: "",
  bot_management_enabled: false,
  challenge_escalation: false,
  captcha_provider: "builtin",
  captcha_site_key: "",
  captcha_secret_key: "",
  captcha_cookie_ttl: 86400,
  cache_sync_interval: 0,
  semantic_analysis_enabled: false,
});

const originalValues = reactive({
  git_token_header: "",
  license_header: "",
  js_challenge_enabled: false,
  js_challenge_mode: "suspicious",
  js_challenge_difficulty: 4,
  webhook_url: "",
  bot_management_enabled: false,
  challenge_escalation: false,
  captcha_provider: "builtin",
  captcha_site_key: "",
  captcha_secret_key: "",
  captcha_cookie_ttl: 86400,
  cache_sync_interval: 0,
  semantic_analysis_enabled: false,
});

const rules = computed(() => ({
  git_token_header: [
    { required: true, message: t("login.required"), trigger: "blur" },
  ],
  license_header: [
    { required: true, message: t("login.required"), trigger: "blur" },
  ],
}));

const fetchData = async () => {
  loading.value = true;
  try {
    const res: any = await request.get("/system/settings");
    form.git_token_header = res.git_token_header || "";
    form.license_header = res.license_header || "";
    form.js_challenge_enabled = res.js_challenge_enabled || false;
    form.js_challenge_mode = res.js_challenge_mode || "suspicious";
    form.js_challenge_difficulty = res.js_challenge_difficulty || 4;
    form.webhook_url = res.webhook_url || "";
    form.bot_management_enabled = res.bot_management_enabled || false;
    form.challenge_escalation = res.challenge_escalation || false;
    form.captcha_provider = res.captcha_provider || "builtin";
    form.captcha_site_key = res.captcha_site_key || "";
    form.captcha_secret_key = res.captcha_secret_key || "";
    form.captcha_cookie_ttl = res.captcha_cookie_ttl || 86400;
    form.cache_sync_interval = res.cache_sync_interval || 0;
    form.semantic_analysis_enabled = res.semantic_analysis_enabled || false;
    Object.assign(originalValues, { ...form });
  } finally {
    loading.value = false;
  }
};

const resetForm = () => {
  Object.assign(form, { ...originalValues });
};

const handleSave = async () => {
  if (!formRef.value) return;
  await formRef.value.validate(async (valid: boolean) => {
    if (valid) {
      saving.value = true;
      try {
        await request.put("/system/settings", form);
        Object.assign(originalValues, { ...form });
        ElMessage.success(t("common.updated"));
      } finally {
        saving.value = false;
      }
    }
  });
};

const handleClearCache = () => {
  ElMessageBox.confirm(t("system.clear_cache_confirm"), t("common.warning"), {
    confirmButtonText: t("common.confirm"),
    cancelButtonText: t("common.cancel"),
    type: "warning",
  }).then(async () => {
    clearing.value = true;
    try {
      await request.post("/system/reload");
      ElMessage.success(t("system.cache_cleared"));
    } finally {
      clearing.value = false;
    }
  });
};

onMounted(() => {
  fetchData();
});
</script>

<style scoped>
.system-settings-view {
  width: 100%;
}
.glass-panel {
  background: rgba(10, 30, 10, 0.75);
  backdrop-filter: blur(10px);
  border: 1px solid #005000;
  box-shadow: 0 5px 25px rgba(0, 0, 0, 0.5);
  border-radius: 4px;
}
.card-header {
  padding: 18px 25px;
  border-bottom: 1px solid #005000;
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: rgba(0, 60, 0, 0.25);
  border-radius: 4px 4px 0 0;
  flex-wrap: wrap;
  gap: 10px;
}
.header-left {
  font-family: "Courier New", monospace;
  font-size: 15px;
  display: flex;
  gap: 10px;
}
.header-right {
  display: flex;
  align-items: center;
  gap: 8px;
}
.prefix {
  color: #0f0;
  font-weight: bold;
  text-shadow: 0 0 5px rgba(0, 255, 0, 0.3);
}
.command {
  color: #fff;
}
.blink-cursor::after {
  content: "_";
  animation: blink 1s step-end infinite;
}
@keyframes blink {
  50% {
    opacity: 0;
  }
}
.hacker-form :deep(.el-form-item__label) {
  color: #0f0 !important;
  font-weight: bold;
  font-size: 14px;
}

.settings-content {
  padding: 25px;
}

.settings-section {
  margin-bottom: 10px;
}

.settings-section.separator {
  margin-top: 25px;
  padding-top: 20px;
  border-top: 1px solid #005000;
}

.section-title {
  font-family: "Courier New", monospace;
  font-size: 16px;
  color: #0f0;
  font-weight: bold;
  margin-bottom: 16px;
  text-shadow: 0 0 5px rgba(0, 255, 0, 0.2);
  display: flex;
  align-items: center;
  gap: 8px;
}

.sub-section-title {
  font-family: "Courier New", monospace;
  font-size: 14px;
  color: #8d8;
  font-weight: bold;
  margin-bottom: 12px;
  display: flex;
  align-items: center;
  gap: 8px;
}

.dim-text {
  color: #8a8;
}

.inline-field {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
}

.field-label {
  font-family: "Courier New", monospace;
  font-size: 14px;
}

@media (max-width: 768px) {
  .header-right {
    flex-wrap: wrap;
  }
}
</style>
