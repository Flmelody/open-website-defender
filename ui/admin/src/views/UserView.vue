<template>
  <div class="user-view">
    <div class="terminal-card glass-panel">
      <div class="card-header no-select">
        <div class="header-left">
          <span class="prefix">root@system:~/users$</span>
          <span class="command blink-cursor">./list_users.sh</span>
        </div>
        <div class="header-right">
          <el-button type="primary" size="small" @click="handleAdd">{{ t('user.new_user') }}</el-button>
          <el-button size="small" @click="fetchData">{{ t('common.refresh') }}</el-button>
        </div>
      </div>

      <div class="data-grid">
        <el-table 
          :data="tableData" 
          v-loading="loading" 
          style="width: 100%" 
          class="hacker-table"
        >
          <el-table-column prop="id" label="ID" width="80">
            <template #default="scope">
              <span class="dim-text">#{{ scope.row.id }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="username" :label="t('user.username')">
            <template #default="scope">
              <span class="bright-text">{{ scope.row.username }}</span>
              <el-tag v-if="scope.row.is_admin" size="small" type="success" class="admin-tag">ADMIN</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="git_token" :label="t('user.git_token')" show-overflow-tooltip>
            <template #default="scope">
              <span v-if="scope.row.git_token" class="token-mask">****************</span>
              <span v-else class="null-value">{{ t('user.undefined') }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="scopes" :label="t('user.authorized_domains')" show-overflow-tooltip>
            <template #default="scope">
              <span v-if="scope.row.scopes" class="bright-text">{{ scope.row.scopes }}</span>
              <span v-else class="null-value">{{ t('user.unrestricted') }}</span>
            </template>
          </el-table-column>
          <el-table-column :label="t('common.actions')" width="200" align="right">
            <template #default="scope">
              <div class="ops-cell">
                <el-button 
                  type="primary" 
                  link 
                  size="small" 
                  @click="handleEdit(scope.row)"
                  class="action-link"
                >
                  {{ t('common.edit') }}
                </el-button>
                <el-button 
                  type="danger" 
                  link 
                  size="small" 
                  v-if="!scope.row.is_admin"
                  @click="handleDelete(scope.row)"
                  class="action-link delete"
                >
                  {{ t('common.delete') }}
                </el-button>
              </div>
            </template>
          </el-table-column>
        </el-table>
      </div>

      <div class="card-footer no-select">
        <span class="status-text">{{ t('common.total_records', {total: total}) }}</span>
        <el-pagination
          v-model:current-page="queryParams.page"
          v-model:page-size="queryParams.size"
          :page-sizes="[10, 20, 50]"
          layout="sizes, prev, pager, next"
          :total="total"
          @size-change="handleSizeChange"
          @current-change="handleCurrentChange"
          small
        />
      </div>
    </div>

    <el-dialog
      v-model="dialogVisible"
      :title="dialogTitle.toUpperCase()"
      width="500px"
      destroy-on-close
    >
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-position="top"
        class="hacker-form"
      >
        <el-form-item :label="'> ' + t('user.username')" prop="username">
          <el-input v-model="form.username" placeholder="_" />
        </el-form-item>
        <el-form-item :label="'> ' + t('user.password')" prop="password">
          <el-input
            v-model="form.password"
            type="password"
            show-password
            :placeholder="form.id ? t('user.unchanged') : '_'"
          />
        </el-form-item>
        <el-form-item :label="'> ' + t('user.git_token')" prop="git_token">
          <div class="git-token-row">
            <el-input v-model="form.git_token" type="password" show-password placeholder="_" />
            <el-button type="primary" size="default" @click="generateGitToken" class="generate-btn">
              {{ t('user.generate') }}
            </el-button>
          </div>
        </el-form-item>
        <el-form-item prop="is_admin">
          <el-checkbox v-model="form.is_admin" :label="t('user.is_admin')" />
        </el-form-item>
        <el-form-item :label="'> ' + t('user.authorized_domains')" prop="scopes">
          <el-select
            v-model="scopesArray"
            multiple
            filterable
            allow-create
            default-first-option
            :placeholder="t('user.authorized_domains_placeholder')"
            :disabled="form.is_admin"
            style="width: 100%"
          >
            <el-option v-for="d in domainOptions" :key="d" :label="d" :value="d" />
          </el-select>
          <div class="scope-hint">{{ t('user.authorized_domains_hint') }}</div>
        </el-form-item>
      </el-form>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="dialogVisible = false">{{ t('common.cancel') }}</el-button>
          <el-button type="primary" :loading="formLoading" @click="handleSubmit">
            {{ t('common.confirm') }}
          </el-button>
        </div>
      </template>
    </el-dialog>

    <!-- Git Token Result Dialog -->
    <el-dialog
      v-model="tokenDialogVisible"
      :title="t('user.token_generated').toUpperCase()"
      width="600px"
      :close-on-click-modal="false"
      :close-on-press-escape="false"
    >
      <div class="token-result">
        <div class="token-warning">
          <el-icon class="warning-icon"><WarningFilled /></el-icon>
          <span>{{ t('user.token_warning') }}</span>
        </div>
        <div class="token-display">
          <code class="token-value">{{ generatedToken }}</code>
          <el-button type="primary" size="small" class="copy-btn" @click="copyToken">
            <el-icon><CopyDocument /></el-icon>
            {{ t('user.copy') }}
          </el-button>
        </div>
        <div v-if="copiedVisible" class="copied-hint">{{ t('user.copied') }}</div>
      </div>
      <template #footer>
        <div class="dialog-footer">
          <el-button type="primary" @click="closeTokenDialog">{{ t('user.understood') }}</el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, reactive, computed, watch } from 'vue'
import request from '@/utils/request'
import { ElMessage, ElMessageBox } from 'element-plus'
import { CopyDocument, WarningFilled } from '@element-plus/icons-vue'
import { useI18n } from 'vue-i18n'

interface User {
  id: number
  username: string
  git_token?: string
  is_admin?: boolean
  scopes?: string
}

const { t } = useI18n()
const tableData = ref<User[]>([])
const total = ref(0)
const loading = ref(false)
const queryParams = reactive({
  page: 1,
  size: 10
})

const dialogVisible = ref(false)
const dialogTitle = ref('')
const formRef = ref()
const formLoading = ref(false)
const isEditMode = ref(false)

const domainOptions = ref<string[]>([])
const scopesArray = ref<string[]>([])
const tokenDialogVisible = ref(false)
const generatedToken = ref('')
const copiedVisible = ref(false)
const pendingToken = ref('')

const form = reactive({
  id: 0,
  username: '',
  password: '',
  git_token: '',
  is_admin: false,
  scopes: ''
})

watch(() => form.is_admin, (val) => {
  if (val) {
    form.scopes = ''
    scopesArray.value = []
  }
})

watch(scopesArray, (val) => {
  form.scopes = val.join(', ')
})

const domainPattern = /^(\*\.)?([a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$/

const scopesValidator = (_rule: any, _value: string, callback: (err?: Error) => void) => {
  for (const item of scopesArray.value) {
    if (item && !domainPattern.test(item)) {
      callback(new Error(t('user.authorized_domains_invalid')))
      return
    }
  }
  callback()
}

const rules = computed(() => ({
  username: [{ required: true, message: t('login.required'), trigger: 'blur' }],
  password: [{ required: !isEditMode.value, message: t('login.required'), trigger: 'blur' }],
  scopes: [{ validator: scopesValidator, trigger: ['blur', 'change'] }]
}))

const fetchDomainOptions = async () => {
  try {
    const res: any = await request.get('/authorized-domains', { params: { all: 'true' } })
    domainOptions.value = (res || []).map((d: any) => d.name)
  } catch {
    // handled
  }
}

const fetchData = async () => {
  loading.value = true
  try {
    const res: any = await request.get('/users', { params: queryParams })
    tableData.value = res.list || []
    total.value = res.total || 0
  } catch (error) {
    // handled
  } finally {
    loading.value = false
  }
}

const handleAdd = () => {
  dialogTitle.value = t('user.title_create')
  form.id = 0
  form.username = ''
  form.password = ''
  form.git_token = ''
  form.is_admin = false
  form.scopes = ''
  scopesArray.value = []
  isEditMode.value = false
  pendingToken.value = ''
  dialogVisible.value = true
}

const handleEdit = (row: User) => {
  dialogTitle.value = t('user.title_edit')
  form.id = row.id
  form.username = row.username
  form.password = ''
  form.git_token = ''
  form.is_admin = row.is_admin || false
  form.scopes = row.scopes || ''
  scopesArray.value = row.scopes ? row.scopes.split(',').map((s: string) => s.trim()).filter(Boolean) : []
  isEditMode.value = true
  pendingToken.value = ''
  dialogVisible.value = true
}

const handleDelete = (row: User) => {
  ElMessageBox.confirm(
    t('user.delete_confirm', { name: row.username }),
    t('common.warning'),
    {
      confirmButtonText: t('common.delete'),
      cancelButtonText: t('common.cancel'),
      type: 'warning',
    }
  ).then(async () => {
    try {
      await request.delete(`/users/${row.id}`)
      ElMessage.success(t('common.deleted'))
      fetchData()
    } catch (error) {
      // handled
    }
  })
}

const handleSubmit = async () => {
  if (!formRef.value) return
  await formRef.value.validate(async (valid: boolean) => {
    if (valid) {
      formLoading.value = true
      try {
        if (form.id === 0) {
          await request.post('/users', form)
          ElMessage.success(t('common.created'))
        } else {
          await request.put(`/users/${form.id}`, form)
          ElMessage.success(t('common.updated'))
        }
        dialogVisible.value = false
        fetchData()
        if (pendingToken.value) {
          generatedToken.value = pendingToken.value
          copiedVisible.value = false
          tokenDialogVisible.value = true
          pendingToken.value = ''
        }
      } catch (error) {
        // handled
      } finally {
        formLoading.value = false
      }
    }
  })
}

const generateRandomHex = (bytes: number): string => {
  const arr = new Uint8Array(bytes)
  crypto.getRandomValues(arr)
  return Array.from(arr, b => b.toString(16).padStart(2, '0')).join('')
}

const generateGitToken = () => {
  const token = generateRandomHex(32)
  form.git_token = token
  pendingToken.value = token
}

const copyToken = async () => {
  try {
    await navigator.clipboard.writeText(generatedToken.value)
    copiedVisible.value = true
    ElMessage.success(t('user.copied'))
  } catch {
    const textarea = document.createElement('textarea')
    textarea.value = generatedToken.value
    textarea.style.position = 'fixed'
    textarea.style.opacity = '0'
    document.body.appendChild(textarea)
    textarea.select()
    document.execCommand('copy')
    document.body.removeChild(textarea)
    copiedVisible.value = true
    ElMessage.success(t('user.copied'))
  }
}

const closeTokenDialog = () => {
  tokenDialogVisible.value = false
  generatedToken.value = ''
}

const handleSizeChange = (val: number) => {
  queryParams.size = val
  fetchData()
}

const handleCurrentChange = (val: number) => {
  queryParams.page = val
  fetchData()
}

onMounted(() => {
  fetchData()
  fetchDomainOptions()
})
</script>

<style scoped>
.user-view {
  width: 100%;
}

.glass-panel {
  background: rgba(10, 30, 10, 0.75); /* More visible background */
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
}

.header-left {
  font-family: 'Courier New', monospace;
  font-size: 15px;
  display: flex;
  gap: 10px;
}

.prefix {
  color: #0F0;
  font-weight: bold;
  text-shadow: 0 0 5px rgba(0, 255, 0, 0.3);
}

.command {
  color: #fff;
}

.blink-cursor::after {
  content: '_';
  animation: blink 1s step-end infinite;
}

@keyframes blink {
  50% { opacity: 0; }
}

.hacker-table {
  font-family: 'Courier New', monospace;
}

.dim-text {
  color: #8a8;
}

.bright-text {
  color: #fff;
  font-weight: bold;
  font-size: 15px;
}

.token-mask {
  color: #00a000;
  letter-spacing: 2px;
}

.null-value {
  color: #006000;
  font-style: italic;
}

.action-link {
  font-weight: bold;
  text-decoration: underline;
}

.card-footer {
  padding: 12px 25px;
  border-top: 1px solid #005000;
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: rgba(0, 60, 0, 0.2);
  border-radius: 0 0 4px 4px;
}

.status-text {
  color: #0F0;
  font-size: 13px;
  font-family: 'Courier New', monospace;
}

.hacker-form :deep(.el-form-item__label) {
  color: #0F0 !important;
  font-weight: bold;
  font-size: 14px;
}

.admin-tag {
  margin-left: 10px;
  font-family: 'Courier New', monospace;
  font-weight: bold;
  letter-spacing: 1px;
}

.dialog-footer {
  text-align: right;
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

.git-token-row {
  display: flex;
  gap: 10px;
  width: 100%;
}

.git-token-row .el-input {
  flex: 1;
}

.generate-btn {
  flex-shrink: 0;
}

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

.scope-hint {
  color: #8a8;
  font-size: 12px;
  margin-top: 4px;
  font-family: 'Courier New', monospace;
}

</style>
