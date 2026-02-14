<template>
  <div class="license-view">
    <div class="terminal-card glass-panel">
      <div class="card-header no-select">
        <div class="header-left">
          <span class="prefix">root@system:~/license$</span>
          <span class="command blink-cursor">./list_licenses.sh</span>
        </div>
        <div class="header-right">
          <el-button type="primary" size="small" @click="handleAdd">{{ t('license.new_license') }}</el-button>
          <el-button size="small" @click="fetchData">{{ t('common.refresh') }}</el-button>
        </div>
      </div>

      <div class="data-grid">
        <el-table :data="tableData" v-loading="loading" style="width: 100%" class="hacker-table">
          <el-table-column prop="id" label="ID" width="80">
            <template #default="scope">
              <span class="dim-text">#{{ scope.row.id }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="name" :label="t('license.name')">
            <template #default="scope">
              <span class="bright-text">{{ scope.row.name }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="active" :label="t('license.status')" width="120">
            <template #default="scope">
              <el-tag :type="scope.row.active ? 'success' : 'danger'" size="small" effect="dark">
                {{ scope.row.active ? t('license.active') : t('license.inactive') }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="created_at" :label="t('common.created_at')" width="200">
            <template #default="scope">
              <span class="dim-text">{{ new Date(scope.row.created_at).toLocaleString() }}</span>
            </template>
          </el-table-column>
          <el-table-column :label="t('common.actions')" width="120" align="right">
            <template #default="scope">
              <div class="ops-cell">
                <el-button type="danger" link size="small" @click="handleDelete(scope.row)" class="action-link delete">
                  {{ t('common.delete') }}
                </el-button>
              </div>
            </template>
          </el-table-column>
        </el-table>
      </div>

      <div class="card-footer no-select">
        <span class="status-text">{{ t('common.total_records', { total }) }}</span>
        <el-pagination
          v-model:current-page="queryParams.page"
          v-model:page-size="queryParams.size"
          :page-sizes="[10, 20, 50]"
          layout="prev, pager, next"
          :total="total"
          @size-change="handleSizeChange"
          @current-change="handleCurrentChange"
          small
        />
      </div>
    </div>

    <!-- Create License Dialog -->
    <el-dialog v-model="dialogVisible" :title="t('license.title_create').toUpperCase()" width="500px" destroy-on-close>
      <el-form ref="formRef" :model="form" :rules="rules" label-position="top" class="hacker-form">
        <el-form-item :label="'> ' + t('license.name')" prop="name">
          <el-input v-model="form.name" :placeholder="t('license.name_placeholder')" />
        </el-form-item>
      </el-form>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="dialogVisible = false">{{ t('common.cancel') }}</el-button>
          <el-button type="primary" :loading="formLoading" @click="handleSubmit">{{ t('common.confirm') }}</el-button>
        </div>
      </template>
    </el-dialog>

    <!-- Token Result Dialog -->
    <el-dialog
      v-model="tokenDialogVisible"
      :title="t('license.token_generated').toUpperCase()"
      width="600px"
      :close-on-click-modal="false"
      :close-on-press-escape="false"
    >
      <div class="token-result">
        <div class="token-warning">
          <el-icon class="warning-icon"><WarningFilled /></el-icon>
          <span>{{ t('license.token_warning') }}</span>
        </div>
        <div class="token-display">
          <code class="token-value">{{ generatedToken }}</code>
          <el-button type="primary" size="small" class="copy-btn" @click="copyToken">
            <el-icon><CopyDocument /></el-icon>
            {{ t('license.copy') }}
          </el-button>
        </div>
        <div v-if="copiedVisible" class="copied-hint">{{ t('license.copied') }}</div>
      </div>
      <template #footer>
        <div class="dialog-footer">
          <el-button type="primary" @click="closeTokenDialog">{{ t('license.understood') }}</el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, reactive, computed } from 'vue'
import request from '@/utils/request'
import { ElMessage, ElMessageBox } from 'element-plus'
import { CopyDocument, WarningFilled } from '@element-plus/icons-vue'
import { useI18n } from 'vue-i18n'

interface LicenseItem {
  id: number
  name: string
  active: boolean
  created_at: string
}

const { t } = useI18n()
const tableData = ref<LicenseItem[]>([])
const total = ref(0)
const loading = ref(false)
const queryParams = reactive({ page: 1, size: 10 })

const dialogVisible = ref(false)
const formRef = ref()
const formLoading = ref(false)
const form = reactive({ name: '' })

const tokenDialogVisible = ref(false)
const generatedToken = ref('')
const copiedVisible = ref(false)

const rules = computed(() => ({
  name: [{ required: true, message: t('login.required'), trigger: 'blur' }]
}))

const fetchData = async () => {
  loading.value = true
  try {
    const res: any = await request.get('/licenses', { params: queryParams })
    tableData.value = res.list || []
    total.value = res.total || 0
  } finally {
    loading.value = false
  }
}

const handleAdd = () => {
  form.name = ''
  dialogVisible.value = true
}

const handleDelete = (row: LicenseItem) => {
  ElMessageBox.confirm(
    t('license.delete_confirm', { name: row.name }),
    t('common.warning'),
    { confirmButtonText: t('common.remove'), cancelButtonText: t('common.cancel'), type: 'warning' }
  ).then(async () => {
    try {
      await request.delete(`/licenses/${row.id}`)
      ElMessage.success(t('common.deleted'))
      fetchData()
    } catch { /* handled */ }
  })
}

const handleSubmit = async () => {
  if (!formRef.value) return
  await formRef.value.validate(async (valid: boolean) => {
    if (valid) {
      formLoading.value = true
      try {
        const res: any = await request.post('/licenses', form)
        dialogVisible.value = false
        generatedToken.value = res.token
        copiedVisible.value = false
        tokenDialogVisible.value = true
        fetchData()
      } finally {
        formLoading.value = false
      }
    }
  })
}

const copyToken = async () => {
  try {
    await navigator.clipboard.writeText(generatedToken.value)
    copiedVisible.value = true
    ElMessage.success(t('license.copied'))
  } catch {
    // Fallback for older browsers
    const textarea = document.createElement('textarea')
    textarea.value = generatedToken.value
    textarea.style.position = 'fixed'
    textarea.style.opacity = '0'
    document.body.appendChild(textarea)
    textarea.select()
    document.execCommand('copy')
    document.body.removeChild(textarea)
    copiedVisible.value = true
    ElMessage.success(t('license.copied'))
  }
}

const closeTokenDialog = () => {
  tokenDialogVisible.value = false
  generatedToken.value = ''
}

const handleSizeChange = (val: number) => { queryParams.size = val; fetchData() }
const handleCurrentChange = (val: number) => { queryParams.page = val; fetchData() }

onMounted(() => { fetchData() })
</script>

<style scoped>
.license-view { width: 100%; }
.glass-panel { background: rgba(10, 30, 10, 0.75); backdrop-filter: blur(10px); border: 1px solid #005000; box-shadow: 0 5px 25px rgba(0, 0, 0, 0.5); border-radius: 4px; }
.card-header { padding: 18px 25px; border-bottom: 1px solid #005000; display: flex; justify-content: space-between; align-items: center; background: rgba(0, 60, 0, 0.25); border-radius: 4px 4px 0 0; }
.header-left { font-family: 'Courier New', monospace; font-size: 15px; display: flex; gap: 10px; }
.prefix { color: #0F0; font-weight: bold; text-shadow: 0 0 5px rgba(0, 255, 0, 0.3); }
.command { color: #fff; }
.blink-cursor::after { content: '_'; animation: blink 1s step-end infinite; }
@keyframes blink { 50% { opacity: 0; } }
.hacker-table { font-family: 'Courier New', monospace; }
.dim-text { color: #8a8; }
.bright-text { color: #fff; font-weight: bold; font-size: 15px; }
.action-link { font-weight: bold; text-decoration: underline; }
.card-footer { padding: 12px 25px; border-top: 1px solid #005000; display: flex; justify-content: space-between; align-items: center; background: rgba(0, 60, 0, 0.2); border-radius: 0 0 4px 4px; }
.status-text { color: #0F0; font-size: 13px; font-family: 'Courier New', monospace; }
.hacker-form :deep(.el-form-item__label) { color: #0F0 !important; font-weight: bold; font-size: 14px; }
.dialog-footer { text-align: right; display: flex; justify-content: flex-end; gap: 12px; }

/* Token result dialog styles */
.token-result {
  font-family: 'Courier New', monospace;
}

.token-warning {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 16px;
  background: rgba(255, 165, 0, 0.15);
  border: 1px solid rgba(255, 165, 0, 0.4);
  border-radius: 4px;
  color: #ffa500;
  font-size: 13px;
  margin-bottom: 20px;
}

.warning-icon {
  font-size: 18px;
  flex-shrink: 0;
}

.token-display {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 14px 16px;
  background: rgba(0, 40, 0, 0.6);
  border: 1px solid #005000;
  border-radius: 4px;
}

.token-value {
  flex: 1;
  color: #0F0;
  font-size: 13px;
  word-break: break-all;
  line-height: 1.5;
  text-shadow: 0 0 5px rgba(0, 255, 0, 0.2);
}

.copy-btn {
  flex-shrink: 0;
}

.copied-hint {
  margin-top: 10px;
  color: #0F0;
  font-size: 12px;
  text-align: right;
}
</style>
