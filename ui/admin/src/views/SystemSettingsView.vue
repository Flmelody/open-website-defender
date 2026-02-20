<template>
  <div class="system-settings-view">
    <div class="terminal-card glass-panel">
      <div class="card-header no-select">
        <div class="header-left">
          <span class="prefix">root@system:~/config$</span>
          <span class="command blink-cursor">./system_settings.sh</span>
        </div>
        <div class="header-right">
          <el-button size="small" @click="resetForm">{{ t('system.reset') }}</el-button>
          <el-button size="small" type="primary" :loading="saving" @click="handleSave">{{ t('system.save') }}</el-button>
          <el-button size="small" @click="fetchData">{{ t('common.refresh') }}</el-button>
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
              <span class="prefix">&gt;</span> {{ t('system.section_headers') }}
              <el-tooltip :content="t('system.headers_desc')" placement="right" effect="dark">
                <el-icon class="info-icon"><InfoFilled /></el-icon>
              </el-tooltip>
            </div>

            <el-form-item :label="'> ' + t('system.git_token_header')" prop="git_token_header">
              <el-input v-model="form.git_token_header" placeholder="Defender-Git-Token" />
            </el-form-item>

            <el-form-item :label="'> ' + t('system.license_header')" prop="license_header">
              <el-input v-model="form.license_header" placeholder="Defender-License" />
            </el-form-item>
          </div>

          <div class="settings-section separator">
            <div class="section-title no-select">
              <span class="prefix">&gt;</span> {{ t('js_challenge.title') }}
              <el-tooltip :content="t('js_challenge.desc')" placement="right" effect="dark" :show-after="200">
                <el-icon class="info-icon"><InfoFilled /></el-icon>
              </el-tooltip>
            </div>

            <div class="inline-field">
              <span class="field-label dim-text">{{ t('js_challenge.enabled') }}:</span>
              <el-switch v-model="form.js_challenge_enabled" />
            </div>

            <el-form-item v-if="form.js_challenge_enabled">
              <template #label>
                <span class="label-with-tip">> {{ t('js_challenge.mode') }}
                  <el-tooltip :content="t('js_challenge.mode_desc')" placement="right" effect="dark">
                    <el-icon class="info-icon"><InfoFilled /></el-icon>
                  </el-tooltip>
                </span>
              </template>
              <el-select v-model="form.js_challenge_mode" style="width: 200px">
                <el-option :label="t('js_challenge.mode_off')" value="off" />
                <el-option :label="t('js_challenge.mode_suspicious')" value="suspicious" />
                <el-option :label="t('js_challenge.mode_all')" value="all" />
              </el-select>
            </el-form-item>

            <el-form-item v-if="form.js_challenge_enabled">
              <template #label>
                <span class="label-with-tip">> {{ t('js_challenge.difficulty') }}
                  <el-tooltip :content="t('js_challenge.difficulty_desc')" placement="right" effect="dark">
                    <el-icon class="info-icon"><InfoFilled /></el-icon>
                  </el-tooltip>
                </span>
              </template>
              <el-slider v-model="form.js_challenge_difficulty" :min="1" :max="6" :step="1" show-stops style="max-width: 280px" />
            </el-form-item>
          </div>

          <div class="settings-section separator">
            <div class="section-title no-select">
              <span class="prefix">&gt;</span> {{ t('webhook.title') }}
              <el-tooltip :content="t('webhook.desc')" placement="right" effect="dark" :show-after="200">
                <el-icon class="info-icon"><InfoFilled /></el-icon>
              </el-tooltip>
            </div>

            <el-form-item :label="'> ' + t('webhook.url')">
              <el-input v-model="form.webhook_url" :placeholder="t('webhook.url_placeholder')" />
            </el-form-item>
          </div>
        </el-form>

        <div class="settings-section separator">
          <div class="section-title no-select">
            <span class="prefix">&gt;</span> {{ t('system.section_cache') }}
            <el-tooltip :content="t('system.cache_desc')" placement="right" effect="dark" :show-after="200">
              <el-icon class="info-icon"><InfoFilled /></el-icon>
            </el-tooltip>
          </div>
          <el-button type="danger" :loading="clearing" @click="handleClearCache">{{ t('system.clear_cache') }}</el-button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, reactive, computed } from 'vue'
import request from '@/utils/request'
import { ElMessage, ElMessageBox } from 'element-plus'
import { InfoFilled } from '@element-plus/icons-vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
const loading = ref(false)
const saving = ref(false)
const clearing = ref(false)
const formRef = ref()

const form = reactive({
  git_token_header: '',
  license_header: '',
  js_challenge_enabled: false,
  js_challenge_mode: 'suspicious',
  js_challenge_difficulty: 4,
  webhook_url: ''
})

const originalValues = reactive({
  git_token_header: '',
  license_header: '',
  js_challenge_enabled: false,
  js_challenge_mode: 'suspicious',
  js_challenge_difficulty: 4,
  webhook_url: ''
})

const rules = computed(() => ({
  git_token_header: [{ required: true, message: t('login.required'), trigger: 'blur' }],
  license_header: [{ required: true, message: t('login.required'), trigger: 'blur' }]
}))

const fetchData = async () => {
  loading.value = true
  try {
    const res: any = await request.get('/system/settings')
    form.git_token_header = res.git_token_header || ''
    form.license_header = res.license_header || ''
    form.js_challenge_enabled = res.js_challenge_enabled || false
    form.js_challenge_mode = res.js_challenge_mode || 'suspicious'
    form.js_challenge_difficulty = res.js_challenge_difficulty || 4
    form.webhook_url = res.webhook_url || ''
    Object.assign(originalValues, { ...form })
  } finally {
    loading.value = false
  }
}

const resetForm = () => {
  Object.assign(form, { ...originalValues })
}

const handleSave = async () => {
  if (!formRef.value) return
  await formRef.value.validate(async (valid: boolean) => {
    if (valid) {
      saving.value = true
      try {
        await request.put('/system/settings', form)
        Object.assign(originalValues, { ...form })
        ElMessage.success(t('common.updated'))
      } finally {
        saving.value = false
      }
    }
  })
}

const handleClearCache = () => {
  ElMessageBox.confirm(
    t('system.clear_cache_confirm'),
    t('common.warning'),
    {
      confirmButtonText: t('common.confirm'),
      cancelButtonText: t('common.cancel'),
      type: 'warning',
    }
  ).then(async () => {
    clearing.value = true
    try {
      await request.post('/system/reload')
      ElMessage.success(t('system.cache_cleared'))
    } finally {
      clearing.value = false
    }
  })
}

onMounted(() => { fetchData() })
</script>

<style scoped>
.system-settings-view { width: 100%; }
.glass-panel { background: rgba(10, 30, 10, 0.75); backdrop-filter: blur(10px); border: 1px solid #005000; box-shadow: 0 5px 25px rgba(0, 0, 0, 0.5); border-radius: 4px; }
.card-header { padding: 18px 25px; border-bottom: 1px solid #005000; display: flex; justify-content: space-between; align-items: center; background: rgba(0, 60, 0, 0.25); border-radius: 4px 4px 0 0; flex-wrap: wrap; gap: 10px; }
.header-left { font-family: 'Courier New', monospace; font-size: 15px; display: flex; gap: 10px; }
.header-right { display: flex; align-items: center; gap: 8px; }
.prefix { color: #0F0; font-weight: bold; text-shadow: 0 0 5px rgba(0, 255, 0, 0.3); }
.command { color: #fff; }
.blink-cursor::after { content: '_'; animation: blink 1s step-end infinite; }
@keyframes blink { 50% { opacity: 0; } }
.hacker-form :deep(.el-form-item__label) { color: #0F0 !important; font-weight: bold; font-size: 14px; }

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
  font-family: 'Courier New', monospace;
  font-size: 16px;
  color: #0F0;
  font-weight: bold;
  margin-bottom: 16px;
  text-shadow: 0 0 5px rgba(0, 255, 0, 0.2);
  display: flex;
  align-items: center;
  gap: 8px;
}

.info-icon {
  color: #8a8;
  font-size: 16px;
  cursor: pointer;
  transition: color 0.2s;
  flex-shrink: 0;
}

.info-icon:hover {
  color: #0f0;
}

.dim-text { color: #8a8; }

.inline-field {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
}

.field-label {
  font-family: 'Courier New', monospace;
  font-size: 14px;
}

.label-with-tip {
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

@media (max-width: 768px) {
  .header-right { flex-wrap: wrap; }
}
</style>
