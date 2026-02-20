<template>
  <div class="ip-list-view">
    <div class="terminal-card glass-panel">
      <div class="card-header no-select">
        <div class="header-left">
          <span class="prefix">root@system:~/firewall$</span>
          <span class="command blink-cursor">./list_blacklist.sh</span>
        </div>
        <div class="header-right">
          <el-button type="primary" size="small" @click="handleAdd">{{ t('ip_list.new_ip') }}</el-button>
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
          <el-table-column prop="ip" :label="t('ip_list.ip_address')">
            <template #default="scope">
              <span class="bright-text">{{ scope.row.ip }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="remark" :label="t('ip_list.remark')" min-width="120">
            <template #default="scope">
              <span v-if="scope.row.remark" class="remark-text">{{ scope.row.remark }}</span>
              <span v-else class="dim-text">-</span>
            </template>
          </el-table-column>
          <el-table-column prop="expires_at" :label="t('ip_list.expires_at')" width="180">
            <template #default="scope">
              <span v-if="scope.row.expires_at" class="expires-text">{{ formatExpiry(scope.row.expires_at) }}</span>
              <span v-else class="dim-text">{{ t('ip_list.permanent') }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="created_at" :label="t('common.created_at')" width="180">
            <template #default="scope">
              <span class="dim-text">{{ new Date(scope.row.created_at).toLocaleString() }}</span>
            </template>
          </el-table-column>
          <el-table-column :label="t('common.actions')" width="120" align="right">
            <template #default="scope">
              <div class="ops-cell">
                <el-button
                  type="danger"
                  link
                  size="small"
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
        <el-form-item :label="'> ' + t('ip_list.ip_address')" prop="ip">
          <el-input v-model="form.ip" placeholder="192.168.1.1" />
          <div class="field-hint">{{ t('ip_list.ip_hint') }}</div>
        </el-form-item>
        <el-form-item :label="'> ' + t('ip_list.remark')" prop="remark">
          <el-input v-model="form.remark" :placeholder="t('ip_list.remark_placeholder')" />
        </el-form-item>
        <el-form-item :label="'> ' + t('ip_list.duration')" prop="duration">
          <el-select v-model="form.duration" style="width: 100%">
            <el-option :label="t('ip_list.permanent')" value="permanent" />
            <el-option :label="t('ip_list.duration_1h')" value="1h" />
            <el-option :label="t('ip_list.duration_6h')" value="6h" />
            <el-option :label="t('ip_list.duration_24h')" value="24h" />
            <el-option :label="t('ip_list.duration_7d')" value="7d" />
            <el-option :label="t('ip_list.duration_30d')" value="30d" />
          </el-select>
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

interface IpItem {
  id: number
  ip: string
  remark: string
  expires_at: string | null
  created_at: string
}

const { t } = useI18n()
const tableData = ref<IpItem[]>([])
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

const form = reactive({
  ip: '',
  remark: '',
  duration: 'permanent'
})

const durationToMs: Record<string, number> = {
  '1h': 3600000,
  '6h': 21600000,
  '24h': 86400000,
  '7d': 604800000,
  '30d': 2592000000
}

const formatExpiry = (expiresAt: string) => {
  const expiry = new Date(expiresAt)
  const now = new Date()
  const diff = expiry.getTime() - now.getTime()
  if (diff <= 0) return t('ip_list.expired')
  const hours = Math.floor(diff / 3600000)
  const minutes = Math.floor((diff % 3600000) / 60000)
  if (hours >= 24) {
    const days = Math.floor(hours / 24)
    return `${days}d ${hours % 24}h`
  }
  if (hours > 0) return `${hours}h ${minutes}m`
  return `${minutes}m`
}

const ipValidator = (_rule: any, value: string, callback: (err?: Error) => void) => {
  if (!value) return callback()
  const ipv4Seg = '(25[0-5]|2[0-4]\\d|[01]?\\d\\d?)'
  const wildSeg = `(${ipv4Seg}|\\*)`
  const ipv4 = `^${ipv4Seg}\\.${ipv4Seg}\\.${ipv4Seg}\\.${ipv4Seg}$`
  const cidr = `^${ipv4Seg}\\.${ipv4Seg}\\.${ipv4Seg}\\.${ipv4Seg}/(3[0-2]|[12]?\\d)$`
  const wildcard = `^${wildSeg}\\.${wildSeg}\\.${wildSeg}\\.${wildSeg}$`
  const pattern = new RegExp(`${ipv4}|${cidr}|${wildcard}`)
  if (!pattern.test(value)) {
    callback(new Error(t('ip_list.ip_invalid')))
  } else {
    callback()
  }
}

const rules = computed(() => ({
  ip: [
    { required: true, message: t('login.required'), trigger: 'blur' },
    { validator: ipValidator, trigger: ['blur', 'change'] }
  ]
}))

const fetchData = async () => {
  loading.value = true
  try {
    const res: any = await request.get('/ip-black-list', { params: queryParams })
    tableData.value = res.list || []
    total.value = res.total || 0
  } catch (error) {
    // handled
  } finally {
    loading.value = false
  }
}

const handleAdd = () => {
  dialogTitle.value = t('ip_list.title_create')
  form.ip = ''
  form.remark = ''
  form.duration = 'permanent'
  dialogVisible.value = true
}

const handleDelete = (row: IpItem) => {
  ElMessageBox.confirm(
    t('ip_list.delete_confirm', { ip: row.ip }),
    t('common.warning'),
    {
      confirmButtonText: t('common.remove'),
      cancelButtonText: t('common.cancel'),
      type: 'warning',
    }
  ).then(async () => {
    try {
      await request.delete(`/ip-black-list/${row.id}`)
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
        const payload: any = { ip: form.ip, remark: form.remark }
        if (form.duration !== 'permanent' && durationToMs[form.duration]) {
          const expiresAt = new Date(Date.now() + durationToMs[form.duration])
          payload.expires_at = expiresAt.toISOString()
        }
        await request.post('/ip-black-list', payload)
        ElMessage.success(t('common.added'))
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
.ip-list-view {
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

.remark-text {
  color: #ff0;
  font-size: 13px;
}

.expires-text {
  color: #f80;
  font-family: 'Courier New', monospace;
  font-size: 13px;
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

.dialog-footer {
  text-align: right;
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

.field-hint {
  color: #8a8;
  font-size: 12px;
  margin-top: 4px;
  font-family: 'Courier New', monospace;
}
</style>
