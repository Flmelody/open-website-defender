<template>
  <div class="system-settings-view">
    <div class="terminal-card glass-panel">
      <div class="card-header no-select">
        <div class="header-left">
          <span class="prefix">root@system:~/config$</span>
          <span class="command blink-cursor">./system_settings.sh</span>
        </div>
        <div class="header-right">
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
            </div>
            <div class="section-desc dim-text">{{ t('system.headers_desc') }}</div>

            <el-form-item :label="'> ' + t('system.git_token_header')" prop="git_token_header">
              <el-input v-model="form.git_token_header" placeholder="Defender-Git-Token" />
            </el-form-item>

            <el-form-item :label="'> ' + t('system.license_header')" prop="license_header">
              <el-input v-model="form.license_header" placeholder="Defender-License" />
            </el-form-item>
          </div>

          <div class="form-actions">
            <el-button @click="resetForm">{{ t('system.reset') }}</el-button>
            <el-button type="primary" :loading="saving" @click="handleSave">{{ t('system.save') }}</el-button>
          </div>
        </el-form>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, reactive, computed } from 'vue'
import request from '@/utils/request'
import { ElMessage } from 'element-plus'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
const loading = ref(false)
const saving = ref(false)
const formRef = ref()

const form = reactive({
  git_token_header: '',
  license_header: ''
})

const originalValues = reactive({
  git_token_header: '',
  license_header: ''
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
    originalValues.git_token_header = form.git_token_header
    originalValues.license_header = form.license_header
  } finally {
    loading.value = false
  }
}

const resetForm = () => {
  form.git_token_header = originalValues.git_token_header
  form.license_header = originalValues.license_header
}

const handleSave = async () => {
  if (!formRef.value) return
  await formRef.value.validate(async (valid: boolean) => {
    if (valid) {
      saving.value = true
      try {
        await request.put('/system/settings', form)
        originalValues.git_token_header = form.git_token_header
        originalValues.license_header = form.license_header
        ElMessage.success(t('common.updated'))
      } finally {
        saving.value = false
      }
    }
  })
}

onMounted(() => { fetchData() })
</script>

<style scoped>
.system-settings-view { width: 100%; }
.glass-panel { background: rgba(10, 30, 10, 0.75); backdrop-filter: blur(10px); border: 1px solid #005000; box-shadow: 0 5px 25px rgba(0, 0, 0, 0.5); border-radius: 4px; }
.card-header { padding: 18px 25px; border-bottom: 1px solid #005000; display: flex; justify-content: space-between; align-items: center; background: rgba(0, 60, 0, 0.25); border-radius: 4px 4px 0 0; }
.header-left { font-family: 'Courier New', monospace; font-size: 15px; display: flex; gap: 10px; }
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

.section-title {
  font-family: 'Courier New', monospace;
  font-size: 16px;
  color: #0F0;
  font-weight: bold;
  margin-bottom: 6px;
  text-shadow: 0 0 5px rgba(0, 255, 0, 0.2);
}

.section-desc {
  font-family: 'Courier New', monospace;
  font-size: 13px;
  margin-bottom: 20px;
}

.dim-text { color: #8a8; }

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding-top: 15px;
  border-top: 1px solid #003000;
}
</style>
