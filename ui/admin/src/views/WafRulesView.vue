<template>
  <div class="waf-rules-view">
    <div class="terminal-card glass-panel">
      <div class="card-header no-select">
        <div class="header-left">
          <span class="prefix">root@system:~/firewall$</span>
          <span class="command blink-cursor">./list_rules.sh</span>
        </div>
        <div class="header-right">
          <el-button type="primary" size="small" @click="handleAdd">{{ t('waf.new_rule') }}</el-button>
          <el-button size="small" @click="fetchData">{{ t('common.refresh') }}</el-button>
        </div>
      </div>

      <div class="data-grid">
        <el-table :data="tableData" v-loading="loading" style="width: 100%" class="hacker-table">
          <el-table-column prop="id" label="ID" width="60">
            <template #default="scope">
              <span class="dim-text">#{{ scope.row.id }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="name" :label="t('waf.name')" min-width="160">
            <template #default="scope">
              <span class="bright-text">{{ scope.row.name }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="category" :label="t('waf.category')" width="100">
            <template #default="scope">
              <el-tag :type="categoryTag(scope.row.category)" size="small" effect="dark">
                {{ scope.row.category.toUpperCase() }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="pattern" :label="t('waf.pattern')" min-width="200">
            <template #default="scope">
              <span class="dim-text" style="font-size: 12px; word-break: break-all;">{{ scope.row.pattern }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="action" :label="t('waf.action')" width="80">
            <template #default="scope">
              <span :style="{ color: scope.row.action === 'block' ? '#f56c6c' : '#e6a23c' }">
                {{ scope.row.action.toUpperCase() }}
              </span>
            </template>
          </el-table-column>
          <el-table-column prop="enabled" :label="t('waf.enabled')" width="80">
            <template #default="scope">
              <el-switch
                v-model="scope.row.enabled"
                size="small"
                @change="handleToggle(scope.row)"
              />
            </template>
          </el-table-column>
          <el-table-column :label="t('common.actions')" width="140" align="right">
            <template #default="scope">
              <div class="ops-cell">
                <el-button type="primary" link size="small" @click="handleEdit(scope.row)" class="action-link">
                  {{ t('common.edit') }}
                </el-button>
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

    <el-dialog v-model="dialogVisible" :title="dialogTitle.toUpperCase()" width="600px" destroy-on-close>
      <el-form ref="formRef" :model="form" :rules="rules" label-position="top" class="hacker-form">
        <el-form-item :label="'> ' + t('waf.name')" prop="name">
          <el-input v-model="form.name" placeholder="SQLi - Union Select" />
        </el-form-item>
        <el-form-item :label="'> ' + t('waf.pattern')" prop="pattern">
          <el-input v-model="form.pattern" placeholder="(?i)(union\s+select)" type="textarea" :rows="2" />
        </el-form-item>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item :label="'> ' + t('waf.category')" prop="category">
              <el-select v-model="form.category" style="width: 100%">
                <el-option label="SQLi" value="sqli" />
                <el-option label="XSS" value="xss" />
                <el-option label="Traversal" value="traversal" />
                <el-option label="Custom" value="custom" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item :label="'> ' + t('waf.action')" prop="action">
              <el-select v-model="form.action" style="width: 100%">
                <el-option label="BLOCK" value="block" />
                <el-option label="LOG" value="log" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="dialogVisible = false">{{ t('common.cancel') }}</el-button>
          <el-button type="primary" :loading="formLoading" @click="handleSubmit">{{ t('common.confirm') }}</el-button>
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

interface WafRule {
  id: number
  name: string
  pattern: string
  category: string
  action: string
  enabled: boolean
  created_at: string
}

const { t } = useI18n()
const tableData = ref<WafRule[]>([])
const total = ref(0)
const loading = ref(false)
const queryParams = reactive({ page: 1, size: 10 })

const dialogVisible = ref(false)
const dialogTitle = ref('')
const formRef = ref()
const formLoading = ref(false)
const editingId = ref<number | null>(null)

const form = reactive({
  name: '',
  pattern: '',
  category: 'custom',
  action: 'block'
})

const rules = computed(() => ({
  name: [{ required: true, message: t('login.required'), trigger: 'blur' }],
  pattern: [{ required: true, message: t('login.required'), trigger: 'blur' }],
  category: [{ required: true, message: t('login.required'), trigger: 'change' }]
}))

const categoryTag = (cat: string) => {
  const map: Record<string, string> = { sqli: 'danger', xss: 'warning', traversal: '', custom: 'info' }
  return map[cat] || 'info'
}

const fetchData = async () => {
  loading.value = true
  try {
    const res: any = await request.get('/waf-rules', { params: queryParams })
    tableData.value = res.list || []
    total.value = res.total || 0
  } finally {
    loading.value = false
  }
}

const handleAdd = () => {
  editingId.value = null
  dialogTitle.value = t('waf.title_create')
  form.name = ''
  form.pattern = ''
  form.category = 'custom'
  form.action = 'block'
  dialogVisible.value = true
}

const handleEdit = (row: WafRule) => {
  editingId.value = row.id
  dialogTitle.value = t('waf.title_edit')
  form.name = row.name
  form.pattern = row.pattern
  form.category = row.category
  form.action = row.action
  dialogVisible.value = true
}

const handleToggle = async (row: WafRule) => {
  try {
    await request.put(`/waf-rules/${row.id}`, { enabled: row.enabled })
    ElMessage.success(t('common.updated'))
  } catch {
    row.enabled = !row.enabled
  }
}

const handleDelete = (row: WafRule) => {
  ElMessageBox.confirm(
    t('waf.delete_confirm', { name: row.name }),
    t('common.warning'),
    { confirmButtonText: t('common.remove'), cancelButtonText: t('common.cancel'), type: 'warning' }
  ).then(async () => {
    try {
      await request.delete(`/waf-rules/${row.id}`)
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
        if (editingId.value) {
          await request.put(`/waf-rules/${editingId.value}`, form)
          ElMessage.success(t('common.updated'))
        } else {
          await request.post('/waf-rules', form)
          ElMessage.success(t('common.added'))
        }
        dialogVisible.value = false
        fetchData()
      } finally {
        formLoading.value = false
      }
    }
  })
}

const handleSizeChange = (val: number) => { queryParams.size = val; fetchData() }
const handleCurrentChange = (val: number) => { queryParams.page = val; fetchData() }

onMounted(() => { fetchData() })
</script>

<style scoped>
.waf-rules-view { width: 100%; }
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
.ops-cell { display: flex; gap: 8px; justify-content: flex-end; }
</style>
