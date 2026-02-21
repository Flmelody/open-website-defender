<template>
  <div class="security-events-view">
    <div class="stats-row">
      <div
        class="stat-card glass-panel"
        v-for="stat in statsCards"
        :key="stat.label"
      >
        <div class="stat-value">{{ stat.value }}</div>
        <div class="stat-label">{{ stat.label }}</div>
      </div>
    </div>

    <!-- Top Threat IPs -->
    <div
      class="terminal-card glass-panel"
      v-if="stats?.top_ips?.length"
      style="margin-bottom: 20px"
    >
      <div class="card-header no-select">
        <div class="header-left">
          <span class="prefix">root@system:~/security$</span>
          <span class="command">./top_threats.sh</span>
        </div>
      </div>
      <div class="data-grid">
        <el-table
          :data="stats.top_ips"
          style="width: 100%"
          class="hacker-table"
        >
          <el-table-column
            prop="client_ip"
            :label="t('security_events.client_ip')"
            width="200"
          >
            <template #default="scope">
              <span class="bright-text">{{ scope.row.client_ip }}</span>
            </template>
          </el-table-column>
          <el-table-column
            prop="count"
            :label="t('security_events.event_count')"
            width="140"
          >
            <template #default="scope">
              <span class="dim-text">{{ scope.row.count }}</span>
            </template>
          </el-table-column>
          <el-table-column
            prop="threat_score"
            :label="t('security_events.threat_score')"
          >
            <template #default="scope">
              <div class="score-bar-wrapper">
                <div
                  class="score-bar"
                  :style="{
                    width: Math.min(scope.row.threat_score, 100) + '%',
                    background: scoreColor(scope.row.threat_score),
                  }"
                ></div>
                <span
                  class="score-text"
                  :style="{ color: scoreColor(scope.row.threat_score) }"
                  >{{ scope.row.threat_score }}</span
                >
              </div>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </div>

    <!-- Event Log -->
    <div class="terminal-card glass-panel">
      <div class="card-header no-select">
        <div class="header-left">
          <span class="prefix">root@system:~/security$</span>
          <span class="command blink-cursor">./events.sh</span>
        </div>
        <div class="header-right">
          <el-select
            v-model="filters.event_type"
            :placeholder="t('security_events.filter_type')"
            clearable
            size="small"
            style="width: 160px; margin-right: 8px"
          >
            <el-option
              :label="t('security_events.type_auto_ban')"
              value="auto_ban"
            />
            <el-option
              :label="t('security_events.type_brute_force')"
              value="brute_force"
            />
            <el-option
              :label="t('security_events.type_scan_detected')"
              value="scan_detected"
            />
            <el-option
              :label="t('security_events.type_js_challenge_fail')"
              value="js_challenge_fail"
            />
          </el-select>
          <el-input
            v-model="filters.client_ip"
            :placeholder="t('security_events.filter_ip')"
            clearable
            size="small"
            style="width: 150px; margin-right: 8px"
            @keyup.enter="fetchData"
          />
          <el-button size="small" @click="fetchData">{{
            t("common.refresh")
          }}</el-button>
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
          <el-table-column
            prop="event_type"
            :label="t('security_events.event_type')"
            width="160"
          >
            <template #default="scope">
              <el-tag
                :type="eventTypeColor(scope.row.event_type)"
                size="small"
                effect="dark"
              >
                {{ eventTypeLabel(scope.row.event_type) }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column
            prop="client_ip"
            :label="t('security_events.client_ip')"
            width="160"
          >
            <template #default="scope">
              <span class="bright-text">{{ scope.row.client_ip }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="detail" :label="t('security_events.detail')">
            <template #default="scope">
              <span class="dim-text">{{ scope.row.detail }}</span>
            </template>
          </el-table-column>
          <el-table-column
            prop="created_at"
            :label="t('security_events.timestamp')"
            width="180"
          >
            <template #default="scope">
              <span class="dim-text">{{
                new Date(scope.row.created_at).toLocaleString()
              }}</span>
            </template>
          </el-table-column>
        </el-table>
      </div>

      <div class="card-footer no-select">
        <span class="status-text">{{
          t("common.total_records", { total })
        }}</span>
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
import { ref, reactive, onMounted, computed } from "vue";
import request from "@/utils/request";
import { useI18n } from "vue-i18n";

interface SecurityEvent {
  id: number;
  event_type: string;
  client_ip: string;
  detail: string;
  created_at: string;
}

interface Stats {
  total_events: number;
  auto_bans_24h: number;
  top_ips: { client_ip: string; count: number; threat_score: number }[];
  type_counts: { event_type: string; count: number }[];
}

const { t } = useI18n();
const tableData = ref<SecurityEvent[]>([]);
const total = ref(0);
const loading = ref(false);
const stats = ref<Stats | null>(null);
const queryParams = reactive({ page: 1, size: 20 });
const filters = reactive({ event_type: "", client_ip: "" });

const statsCards = computed(() => [
  {
    label: t("security_events.stats_total"),
    value: stats.value?.total_events ?? 0,
  },
  {
    label: t("security_events.stats_auto_bans"),
    value: stats.value?.auto_bans_24h ?? 0,
  },
  {
    label: t("security_events.stats_threats"),
    value: stats.value?.top_ips?.length ?? 0,
  },
]);

const scoreColor = (score: number) => {
  if (score >= 20) return "#f56c6c";
  if (score >= 10) return "#e6a23c";
  return "#67c23a";
};

const eventTypeColor = (type: string) => {
  switch (type) {
    case "auto_ban":
      return "danger";
    case "brute_force":
      return "warning";
    case "scan_detected":
      return "info";
    case "js_challenge_fail":
      return "";
    default:
      return "info";
  }
};

const eventTypeLabel = (type: string) => {
  switch (type) {
    case "auto_ban":
      return t("security_events.type_auto_ban");
    case "brute_force":
      return t("security_events.type_brute_force");
    case "scan_detected":
      return t("security_events.type_scan_detected");
    case "js_challenge_fail":
      return t("security_events.type_js_challenge_fail");
    default:
      return type;
  }
};

const fetchData = async () => {
  loading.value = true;
  try {
    const params: any = { ...queryParams };
    if (filters.event_type) params.event_type = filters.event_type;
    if (filters.client_ip) params.client_ip = filters.client_ip;
    const res: any = await request.get("/security-events", { params });
    tableData.value = res.list || [];
    total.value = res.total || 0;
  } finally {
    loading.value = false;
  }
};

const fetchStats = async () => {
  try {
    const res: any = await request.get("/security-events/stats");
    stats.value = res;
  } catch {
    // handled
  }
};

const handleSizeChange = (val: number) => {
  queryParams.size = val;
  fetchData();
};
const handleCurrentChange = (val: number) => {
  queryParams.page = val;
  fetchData();
};

onMounted(() => {
  fetchData();
  fetchStats();
});
</script>

<style scoped>
.security-events-view {
  width: 100%;
}
.stats-row {
  display: flex;
  gap: 16px;
  margin-bottom: 20px;
}
.stat-card {
  flex: 1;
  padding: 20px;
  text-align: center;
}
.stat-value {
  font-size: 28px;
  font-weight: bold;
  color: #0f0;
  font-family: "Courier New", monospace;
  text-shadow: 0 0 10px rgba(0, 255, 0, 0.3);
}
.stat-label {
  font-size: 12px;
  color: #8a8;
  margin-top: 6px;
  font-family: "Courier New", monospace;
  text-transform: uppercase;
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
  flex-wrap: wrap;
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
.hacker-table {
  font-family: "Courier New", monospace;
}
.dim-text {
  color: #8a8;
}
.bright-text {
  color: #fff;
  font-weight: bold;
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
  color: #0f0;
  font-size: 13px;
  font-family: "Courier New", monospace;
}

.score-bar-wrapper {
  display: flex;
  align-items: center;
  gap: 10px;
}

.score-bar {
  height: 6px;
  border-radius: 3px;
  min-width: 2px;
  transition: width 0.3s;
}

.score-text {
  font-family: "Courier New", monospace;
  font-size: 14px;
  font-weight: bold;
  min-width: 30px;
}
</style>
