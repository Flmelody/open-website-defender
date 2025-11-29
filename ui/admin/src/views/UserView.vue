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
          layout="prev, pager, next"
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
          <el-input v-model="form.git_token" type="password" show-password placeholder="_" />
        </el-form-item>
        <el-form-item prop="is_admin">
          <el-checkbox v-model="form.is_admin" :label="t('user.is_admin')" />
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
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, reactive, computed } from 'vue'
import request from '@/utils/request'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useI18n } from 'vue-i18n'

interface User {
  id: number
  username: string
  git_token?: string
  is_admin?: boolean
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

const form = reactive({
  id: 0,
  username: '',
  password: '',
  git_token: '',
  is_admin: false
})

const rules = computed(() => ({
  username: [{ required: true, message: t('login.required'), trigger: 'blur' }],
  password: [{ required: !isEditMode.value, message: t('login.required'), trigger: 'blur' }]
}))

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
  isEditMode.value = false
  dialogVisible.value = true
}

const handleEdit = (row: User) => {
  dialogTitle.value = t('user.title_edit')
  form.id = row.id
  form.username = row.username
  form.password = ''
  form.git_token = row.git_token || ''
  form.is_admin = row.is_admin || false
  isEditMode.value = true
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
      } catch (error) {
        // handled
      } finally {
        formLoading.value = false
      }
    }
  })
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
</style>
