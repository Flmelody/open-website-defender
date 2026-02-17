<template>
  <div class="access-log-view">
    <div class="terminal-card glass-panel">
      <div class="card-header no-select">
        <div class="header-left">
          <span class="prefix">root@system:~/firewall$</span>
          <span class="command blink-cursor">./access_log.sh</span>
        </div>
        <div class="header-right">
          <el-button type="danger" size="small" @click="handleClear">{{ t('access_log.clear_all') }}</el-button>
          <el-button size="small" @click="fetchData">{{ t('common.refresh') }}</el-button>
        </div>
      </div>

      <div class="filter-bar">
        <el-input v-model="queryParams.client_ip" :placeholder="t('access_log.filter_ip')" size="small" clearable style="width: 160px" @clear="fetchData" @keyup.enter="fetchData" />
        <el-select v-model="queryParams.action" :placeholder="t('access_log.filter_action')" size="small" clearable style="width: 160px" @change="fetchData">
          <el-option label="Allowed" value="allowed" />
          <el-option label="Blocked" value="blocked" />
          <el-option label="Blocked (WAF)" value="blocked_waf" />
          <el-option label="Blocked (Rate)" value="blocked_ratelimit" />
          <el-option label="Blocked (Geo)" value="blocked_geo" />
        </el-select>
        <el-button type="primary" size="small" @click="fetchData">{{ t('access_log.search') }}</el-button>
      </div>

      <div class="data-grid">
        <el-table :data="tableData" v-loading="loading" style="width: 100%" class="hacker-table">
          <el-table-column prop="id" label="ID" width="70">
            <template #default="scope">
              <span class="dim-text">#{{ scope.row.id }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="client_ip" :label="t('access_log.client_ip')" width="140">
            <template #default="scope">
              <span class="bright-text">{{ scope.row.client_ip }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="method" :label="t('access_log.method')" width="70">
            <template #default="scope">
              <span class="dim-text">{{ scope.row.method }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="path" :label="t('access_log.path')" min-width="180">
            <template #default="scope">
              <span class="dim-text" style="word-break: break-all;">{{ scope.row.path }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="status_code" :label="t('access_log.status')" width="70">
            <template #default="scope">
              <span :style="{ color: scope.row.status_code >= 400 ? '#f56c6c' : '#67c23a' }">
                {{ scope.row.status_code }}
              </span>
            </template>
          </el-table-column>
          <el-table-column prop="action" :label="t('access_log.action')" width="130">
            <template #default="scope">
              <el-tag :type="actionTag(scope.row.action)" size="small" effect="dark">
                {{ scope.row.action }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="rule_name" :label="t('access_log.rule')" width="140">
            <template #default="scope">
              <span class="dim-text">{{ scope.row.rule_name || '-' }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="created_at" :label="t('common.created_at')" width="170">
            <template #default="scope">
              <span class="dim-text">{{ new Date(scope.row.created_at).toLocaleString() }}</span>
            </template>
          </el-table-column>
        </el-table>
      </div>

      <div class="card-footer no-select">
        <span class="status-text">{{ t('common.total_records', { total }) }}</span>
        <el-pagination
          v-model:current-page="queryParams.page"
          v-model:page-size="queryParams.size"
          :page-sizes="[20, 50, 100]"
          layout="sizes, prev, pager, next"
          :total="total"
          @size-change="handleSizeChange"
          @current-change="handleCurrentChange"
          small
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, reactive } from 'vue'
import request from '@/utils/request'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
const tableData = ref<any[]>([])
const total = ref(0)
const loading = ref(false)
const queryParams = reactive({
  page: 1,
  size: 20,
  client_ip: '',
  action: ''
})

const actionTag = (action: string) => {
  if (action === 'allowed') return 'success'
  if (action.startsWith('blocked')) return 'danger'
  return 'info'
}

const fetchData = async () => {
  loading.value = true
  try {
    const params: any = { page: queryParams.page, size: queryParams.size }
    if (queryParams.client_ip) params.client_ip = queryParams.client_ip
    if (queryParams.action) params.action = queryParams.action
    const res: any = await request.get('/access-logs', { params })
    tableData.value = res.list || []
    total.value = res.total || 0
  } finally {
    loading.value = false
  }
}

const handleClear = () => {
  ElMessageBox.confirm(
    t('access_log.clear_confirm'),
    t('common.warning'),
    {
      confirmButtonText: t('common.confirm'),
      cancelButtonText: t('common.cancel'),
      type: 'warning',
    }
  ).then(async () => {
    try {
      const res: any = await request.delete('/access-logs')
      ElMessage.success(t('access_log.cleared', { count: res?.deleted || 0 }))
      fetchData()
    } catch {
      // handled
    }
  })
}

const handleSizeChange = (val: number) => { queryParams.size = val; fetchData() }
const handleCurrentChange = (val: number) => { queryParams.page = val; fetchData() }

onMounted(() => { fetchData() })
</script>

<style scoped>
.access-log-view { width: 100%; }
.glass-panel { background: rgba(10, 30, 10, 0.75); backdrop-filter: blur(10px); border: 1px solid #005000; box-shadow: 0 5px 25px rgba(0, 0, 0, 0.5); border-radius: 4px; }
.card-header { padding: 18px 25px; border-bottom: 1px solid #005000; display: flex; justify-content: space-between; align-items: center; background: rgba(0, 60, 0, 0.25); border-radius: 4px 4px 0 0; }
.header-left { font-family: 'Courier New', monospace; font-size: 15px; display: flex; gap: 10px; }
.prefix { color: #0F0; font-weight: bold; text-shadow: 0 0 5px rgba(0, 255, 0, 0.3); }
.command { color: #fff; }
.blink-cursor::after { content: '_'; animation: blink 1s step-end infinite; }
@keyframes blink { 50% { opacity: 0; } }
.filter-bar { padding: 12px 25px; display: flex; gap: 10px; align-items: center; border-bottom: 1px solid #003000; background: rgba(0, 40, 0, 0.2); }
.hacker-table { font-family: 'Courier New', monospace; }
.dim-text { color: #8a8; }
.bright-text { color: #fff; font-weight: bold; }
.card-footer { padding: 12px 25px; border-top: 1px solid #005000; display: flex; justify-content: space-between; align-items: center; background: rgba(0, 60, 0, 0.2); border-radius: 0 0 4px 4px; }
.status-text { color: #0F0; font-size: 13px; font-family: 'Courier New', monospace; }
</style>
