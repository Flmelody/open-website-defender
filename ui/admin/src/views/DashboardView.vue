<template>
  <div class="dashboard-view">
    <div class="terminal-card glass-panel">
      <div class="card-header no-select">
        <div class="header-left">
          <span class="prefix">root@system:~/firewall$</span>
          <span class="command blink-cursor">./dashboard.sh</span>
        </div>
        <div class="header-right">
          <el-button size="small" @click="fetchData">{{ t('common.refresh') }}</el-button>
        </div>
      </div>

      <div class="stats-grid" v-loading="loading">
        <div class="stat-card">
          <div class="stat-value bright-text">{{ stats.total_requests || 0 }}</div>
          <div class="stat-label dim-text">{{ t('dashboard.total_requests') }}</div>
        </div>
        <div class="stat-card">
          <div class="stat-value" style="color: #f56c6c;">{{ stats.blocked_requests || 0 }}</div>
          <div class="stat-label dim-text">{{ t('dashboard.blocked_requests') }}</div>
        </div>
        <div class="stat-card">
          <div class="stat-value bright-text">{{ stats.waf_rule_count || 0 }}</div>
          <div class="stat-label dim-text">{{ t('dashboard.waf_rules') }}</div>
        </div>
        <div class="stat-card">
          <div class="stat-value bright-text">{{ stats.blacklist_count || 0 }}</div>
          <div class="stat-label dim-text">{{ t('dashboard.blacklist_count') }}</div>
        </div>
        <div class="stat-card">
          <div class="stat-value bright-text">{{ stats.whitelist_count || 0 }}</div>
          <div class="stat-label dim-text">{{ t('dashboard.whitelist_count') }}</div>
        </div>
        <div class="stat-card">
          <div class="stat-value bright-text">{{ stats.user_count || 0 }}</div>
          <div class="stat-label dim-text">{{ t('dashboard.user_count') }}</div>
        </div>
        <div class="stat-card wide">
          <div class="stat-value bright-text">{{ formatUptime(stats.uptime_seconds || 0) }}</div>
          <div class="stat-label dim-text">{{ t('dashboard.uptime') }}</div>
        </div>
      </div>

      <div class="section-header no-select" v-if="topBlocked.length > 0">
        <span class="prefix">&gt;</span>
        <span class="command">{{ t('dashboard.top_blocked_ips') }}</span>
      </div>

      <div class="data-grid" v-if="topBlocked.length > 0">
        <el-table :data="topBlocked" class="hacker-table" style="width: 100%">
          <el-table-column prop="client_ip" label="IP" width="200">
            <template #default="scope">
              <span class="bright-text">{{ scope.row.client_ip }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="count" :label="t('dashboard.block_count')">
            <template #default="scope">
              <span style="color: #f56c6c; font-weight: bold;">{{ scope.row.count }}</span>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import request from '@/utils/request'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
const loading = ref(false)
const stats = ref<any>({})
const topBlocked = ref<any[]>([])

const formatUptime = (seconds: number) => {
  const d = Math.floor(seconds / 86400)
  const h = Math.floor((seconds % 86400) / 3600)
  const m = Math.floor((seconds % 3600) / 60)
  if (d > 0) return `${d}d ${h}h ${m}m`
  if (h > 0) return `${h}h ${m}m`
  return `${m}m`
}

const fetchData = async () => {
  loading.value = true
  try {
    const res: any = await request.get('/dashboard/stats')
    stats.value = res || {}
    topBlocked.value = res.top_blocked_ips || []
  } catch (error) {
    // handled
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchData()
})
</script>

<style scoped>
.dashboard-view {
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

.prefix { color: #0F0; font-weight: bold; text-shadow: 0 0 5px rgba(0, 255, 0, 0.3); }
.command { color: #fff; }
.blink-cursor::after { content: '_'; animation: blink 1s step-end infinite; }
@keyframes blink { 50% { opacity: 0; } }

.stats-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 16px;
  padding: 24px;
}

.stat-card {
  background: rgba(0, 40, 0, 0.5);
  border: 1px solid #003000;
  border-radius: 4px;
  padding: 20px;
  text-align: center;
}

.stat-card.wide {
  grid-column: span 3;
}

.stat-value {
  font-family: 'Courier New', monospace;
  font-size: 28px;
  font-weight: bold;
  margin-bottom: 8px;
}

.stat-label {
  font-family: 'Courier New', monospace;
  font-size: 12px;
  text-transform: uppercase;
}

.section-header {
  padding: 12px 25px;
  border-top: 1px solid #005000;
  font-family: 'Courier New', monospace;
  font-size: 14px;
  display: flex;
  gap: 8px;
  background: rgba(0, 60, 0, 0.15);
}

.dim-text { color: #8a8; }
.bright-text { color: #fff; font-weight: bold; }
.hacker-table { font-family: 'Courier New', monospace; }
</style>
