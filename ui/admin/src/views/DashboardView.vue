<template>
  <div class="dashboard-view">
    <div class="terminal-card glass-panel">
      <div class="card-header no-select">
        <div class="header-left">
          <span class="prefix">root@system:~/firewall$</span>
          <span class="command blink-cursor">./dashboard.sh</span>
        </div>
        <div class="header-right">
          <el-button size="small" @click="fetchData">{{
            t("common.refresh")
          }}</el-button>
        </div>
      </div>

      <div class="dashboard-content" v-loading="loading">
        <!-- Stats Cards Row -->
        <div class="stats-row">
          <div class="stat-card">
            <div class="stat-icon total-icon">
              <el-icon :size="22"><DataBoard /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value bright-text">
                {{ stats.total_requests || 0 }}
              </div>
              <div class="stat-label dim-text">
                {{ t("dashboard.total_requests") }}
              </div>
            </div>
          </div>
          <div class="stat-card">
            <div class="stat-icon blocked-icon">
              <el-icon :size="22"><WarningFilled /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value" style="color: #f56c6c">
                {{ stats.blocked_requests || 0 }}
              </div>
              <div class="stat-label dim-text">
                {{ t("dashboard.blocked_requests") }}
              </div>
            </div>
          </div>
          <div class="stat-card">
            <div class="stat-icon rule-icon">
              <el-icon :size="22"><Lock /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value bright-text">
                {{ stats.waf_rule_count || 0 }}
              </div>
              <div class="stat-label dim-text">
                {{ t("dashboard.waf_rules") }}
              </div>
            </div>
          </div>
          <div class="stat-card">
            <div class="stat-icon uptime-icon">
              <el-icon :size="22"><Timer /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value bright-text">
                {{ formatUptime(stats.uptime_seconds || 0) }}
              </div>
              <div class="stat-label dim-text">{{ t("dashboard.uptime") }}</div>
            </div>
          </div>
        </div>

        <!-- Charts Row -->
        <div class="charts-row">
          <div class="chart-panel">
            <div class="chart-title no-select">
              <span class="prefix">&gt;</span>
              <span>{{ t("dashboard.request_trend") }}</span>
              <div class="trend-range-btns">
                <button
                  v-for="opt in trendRangeOptions"
                  :key="opt.hours"
                  :class="['range-btn', { active: trendHours === opt.hours }]"
                  @click="switchTrendRange(opt.hours)"
                >
                  {{ opt.label }}
                </button>
              </div>
            </div>
            <v-chart class="chart" :option="trendOption" autoresize />
          </div>
          <div class="chart-panel small">
            <div class="chart-title no-select">
              <span class="prefix">&gt;</span>
              <span>{{ t("dashboard.block_reasons") }}</span>
            </div>
            <v-chart
              v-if="blockReasons.length > 0"
              class="chart"
              :option="pieOption"
              autoresize
            />
            <div v-else class="empty-state dim-text">
              {{ t("dashboard.no_data") }}
            </div>
          </div>
        </div>

        <!-- Bottom Row: Top Blocked IPs + Resource Counters -->
        <div class="bottom-row">
          <div class="chart-panel">
            <div class="chart-title no-select">
              <span class="prefix">&gt;</span>
              <span>{{ t("dashboard.top_blocked_ips") }}</span>
            </div>
            <v-chart
              v-if="topBlocked.length > 0"
              class="chart"
              :option="barOption"
              autoresize
            />
            <div v-else class="empty-state dim-text">
              {{ t("dashboard.no_data") }}
            </div>
          </div>
          <div class="counters-panel">
            <div class="counter-item">
              <div class="counter-label dim-text">
                {{ t("dashboard.blacklist_count") }}
              </div>
              <div class="counter-value bright-text">
                {{ stats.blacklist_count || 0 }}
              </div>
            </div>
            <div class="counter-item">
              <div class="counter-label dim-text">
                {{ t("dashboard.whitelist_count") }}
              </div>
              <div class="counter-value bright-text">
                {{ stats.whitelist_count || 0 }}
              </div>
            </div>
            <div class="counter-item">
              <div class="counter-label dim-text">
                {{ t("dashboard.user_count") }}
              </div>
              <div class="counter-value bright-text">
                {{ stats.user_count || 0 }}
              </div>
            </div>
            <div class="counter-item">
              <div class="counter-label dim-text">
                {{ t("dashboard.block_rate") }}
              </div>
              <div class="counter-value" style="color: #f56c6c">
                {{ blockRate }}%
              </div>
            </div>
            <div class="counter-item">
              <div class="counter-label dim-text">
                {{ t("dashboard.auto_bans_24h") }}
              </div>
              <div class="counter-value" style="color: #e6a23c">
                {{ securityStats.auto_bans_24h || 0 }}
              </div>
            </div>
            <div class="counter-item">
              <div class="counter-label dim-text">
                {{ t("dashboard.active_threats") }}
              </div>
              <div class="counter-value" style="color: #f56c6c">
                {{ securityStats.top_ips?.length || 0 }}
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from "vue";
import request from "@/utils/request";
import { useI18n } from "vue-i18n";
import VChart from "vue-echarts";
import { use } from "echarts/core";
import { CanvasRenderer } from "echarts/renderers";
import { LineChart, PieChart, BarChart } from "echarts/charts";
import {
  TitleComponent,
  TooltipComponent,
  GridComponent,
  LegendComponent,
} from "echarts/components";
import { DataBoard, WarningFilled, Lock, Timer } from "@element-plus/icons-vue";

use([
  CanvasRenderer,
  LineChart,
  PieChart,
  BarChart,
  TitleComponent,
  TooltipComponent,
  GridComponent,
  LegendComponent,
]);

const { t } = useI18n();
const loading = ref(false);
const stats = ref<any>({});
const topBlocked = ref<any[]>([]);
const requestTrend = ref<any[]>([]);
const blockReasons = ref<any[]>([]);
const securityStats = ref<any>({});
const trendHours = ref(24);

const trendRangeOptions = computed(() => [
  { hours: 24, label: t("dashboard.trend_1d") },
  { hours: 168, label: t("dashboard.trend_7d") },
  { hours: 720, label: t("dashboard.trend_30d") },
]);

const switchTrendRange = async (hours: number) => {
  trendHours.value = hours;
  try {
    const res: any = await request.get("/dashboard/stats", {
      params: { hours },
    });
    requestTrend.value = res.request_trend || [];
  } catch {
    // handled
  }
};

const blockRate = computed(() => {
  const total = stats.value.total_requests || 0;
  const blocked = stats.value.blocked_requests || 0;
  if (total === 0) return "0.0";
  return ((blocked / total) * 100).toFixed(1);
});

const formatUptime = (seconds: number) => {
  const d = Math.floor(seconds / 86400);
  const h = Math.floor((seconds % 86400) / 3600);
  const m = Math.floor((seconds % 3600) / 60);
  if (d > 0) return `${d}d ${h}h ${m}m`;
  if (h > 0) return `${h}h ${m}m`;
  return `${m}m`;
};

const formatHour = (hourStr: string) => {
  // hourStr is like "2025-01-15 14:00"
  const parts = hourStr.split(" ");
  if (parts.length < 2) return hourStr;
  if (trendHours.value <= 24) return parts[1];
  return `${parts[0].slice(5)} ${parts[1]}`;
};

const trendOption = computed(() => ({
  tooltip: {
    trigger: "axis",
    backgroundColor: "rgba(10, 30, 10, 0.9)",
    borderColor: "#005000",
    textStyle: {
      color: "#0f0",
      fontFamily: "Courier New, monospace",
      fontSize: 12,
    },
  },
  legend: {
    data: [t("dashboard.total_requests"), t("dashboard.blocked_requests")],
    textStyle: {
      color: "#8a8",
      fontFamily: "Courier New, monospace",
      fontSize: 11,
    },
    top: 4,
    right: 10,
  },
  grid: { top: 35, right: 15, bottom: 25, left: 50 },
  xAxis: {
    type: "category",
    data: requestTrend.value.map((i) => formatHour(i.hour)),
    axisLine: { lineStyle: { color: "#005000" } },
    axisLabel: {
      color: "#8a8",
      fontFamily: "Courier New, monospace",
      fontSize: 10,
    },
    axisTick: { show: false },
  },
  yAxis: {
    type: "value",
    splitLine: { lineStyle: { color: "#002800" } },
    axisLine: { lineStyle: { color: "#005000" } },
    axisLabel: {
      color: "#8a8",
      fontFamily: "Courier New, monospace",
      fontSize: 10,
    },
  },
  series: [
    {
      name: t("dashboard.total_requests"),
      type: "line",
      smooth: true,
      symbol: "none",
      data: requestTrend.value.map((i) => i.total),
      areaStyle: {
        color: {
          type: "linear",
          x: 0,
          y: 0,
          x2: 0,
          y2: 1,
          colorStops: [
            { offset: 0, color: "rgba(0, 200, 0, 0.3)" },
            { offset: 1, color: "rgba(0, 200, 0, 0.02)" },
          ],
        },
      },
      lineStyle: { color: "#0c0", width: 2 },
      itemStyle: { color: "#0c0" },
    },
    {
      name: t("dashboard.blocked_requests"),
      type: "line",
      smooth: true,
      symbol: "none",
      data: requestTrend.value.map((i) => i.blocked),
      areaStyle: {
        color: {
          type: "linear",
          x: 0,
          y: 0,
          x2: 0,
          y2: 1,
          colorStops: [
            { offset: 0, color: "rgba(245, 108, 108, 0.3)" },
            { offset: 1, color: "rgba(245, 108, 108, 0.02)" },
          ],
        },
      },
      lineStyle: { color: "#f56c6c", width: 2 },
      itemStyle: { color: "#f56c6c" },
    },
  ],
}));

const actionLabelMap: Record<string, string> = {
  blocked_blacklist: "Blacklist",
  blocked_waf: "WAF",
  blocked_ratelimit: "Rate Limit",
  blocked_geo: "Geo Block",
};

const pieColors = ["#f56c6c", "#e6a23c", "#409eff", "#67c23a", "#909399"];

const pieOption = computed(() => ({
  tooltip: {
    backgroundColor: "rgba(10, 30, 10, 0.9)",
    borderColor: "#005000",
    textStyle: {
      color: "#0f0",
      fontFamily: "Courier New, monospace",
      fontSize: 12,
    },
  },
  series: [
    {
      type: "pie",
      radius: ["40%", "70%"],
      center: ["50%", "55%"],
      avoidLabelOverlap: true,
      label: {
        color: "#8a8",
        fontFamily: "Courier New, monospace",
        fontSize: 11,
        formatter: "{b}: {c}",
      },
      labelLine: { lineStyle: { color: "#005000" } },
      data: blockReasons.value.map((item, idx) => ({
        name:
          actionLabelMap[item.action] || item.action.replace("blocked_", ""),
        value: item.count,
        itemStyle: { color: pieColors[idx % pieColors.length] },
      })),
      emphasis: {
        itemStyle: {
          shadowBlur: 10,
          shadowOffsetX: 0,
          shadowColor: "rgba(0, 0, 0, 0.5)",
        },
      },
    },
  ],
}));

const barOption = computed(() => {
  const ips = topBlocked.value;
  return {
    tooltip: {
      trigger: "axis",
      axisPointer: { type: "shadow" },
      backgroundColor: "rgba(10, 30, 10, 0.9)",
      borderColor: "#005000",
      textStyle: {
        color: "#0f0",
        fontFamily: "Courier New, monospace",
        fontSize: 12,
      },
    },
    grid: { top: 10, right: 15, bottom: 55, left: 45 },
    xAxis: {
      type: "category",
      data: ips.map((i) => i.client_ip),
      axisLine: { lineStyle: { color: "#005000" } },
      axisLabel: {
        color: "#8a8",
        fontFamily: "Courier New, monospace",
        fontSize: 10,
        rotate: 35,
      },
      axisTick: { show: false },
    },
    yAxis: {
      type: "value",
      splitLine: { lineStyle: { color: "#002800" } },
      axisLine: { lineStyle: { color: "#005000" } },
      axisLabel: {
        color: "#8a8",
        fontFamily: "Courier New, monospace",
        fontSize: 10,
      },
    },
    series: [
      {
        type: "bar",
        data: ips.map((i) => i.count),
        barWidth: "50%",
        itemStyle: {
          color: {
            type: "linear",
            x: 0,
            y: 0,
            x2: 0,
            y2: 1,
            colorStops: [
              { offset: 0, color: "rgba(245, 108, 108, 0.8)" },
              { offset: 1, color: "rgba(245, 108, 108, 0.3)" },
            ],
          },
          borderRadius: [3, 3, 0, 0],
        },
      },
    ],
  };
});

const fetchData = async () => {
  loading.value = true;
  try {
    const res: any = await request.get("/dashboard/stats", {
      params: { hours: trendHours.value },
    });
    stats.value = res || {};
    topBlocked.value = res.top_blocked_ips || [];
    requestTrend.value = res.request_trend || [];
    blockReasons.value = res.block_reasons || [];
  } catch (error) {
    // handled
  } finally {
    loading.value = false;
  }
};

const fetchSecurityStats = async () => {
  try {
    const res: any = await request.get("/security-events/stats");
    securityStats.value = res || {};
  } catch {
    // handled
  }
};

onMounted(() => {
  fetchData();
  fetchSecurityStats();
});
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
  font-family: "Courier New", monospace;
  font-size: 15px;
  display: flex;
  gap: 10px;
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

.dashboard-content {
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

/* Stats Row */
.stats-row {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
}

.stat-card {
  background: rgba(0, 40, 0, 0.5);
  border: 1px solid #003000;
  border-radius: 4px;
  padding: 16px 18px;
  display: flex;
  align-items: center;
  gap: 14px;
}

.stat-icon {
  width: 44px;
  height: 44px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.total-icon {
  background: rgba(0, 200, 0, 0.15);
  color: #0c0;
}
.blocked-icon {
  background: rgba(245, 108, 108, 0.15);
  color: #f56c6c;
}
.rule-icon {
  background: rgba(64, 158, 255, 0.15);
  color: #409eff;
}
.uptime-icon {
  background: rgba(230, 162, 60, 0.15);
  color: #e6a23c;
}

.stat-info {
  min-width: 0;
}

.stat-value {
  font-family: "Courier New", monospace;
  font-size: 24px;
  font-weight: bold;
  line-height: 1.2;
}

.stat-label {
  font-family: "Courier New", monospace;
  font-size: 11px;
  text-transform: uppercase;
  margin-top: 2px;
}

/* Charts Row */
.charts-row {
  display: grid;
  grid-template-columns: 1fr 340px;
  gap: 16px;
}

.chart-panel {
  background: rgba(0, 40, 0, 0.35);
  border: 1px solid #003000;
  border-radius: 4px;
  overflow: hidden;
}

.chart-title {
  padding: 10px 16px;
  font-family: "Courier New", monospace;
  font-size: 13px;
  color: #8a8;
  border-bottom: 1px solid #002800;
  display: flex;
  align-items: center;
  gap: 8px;
}

.trend-range-btns {
  margin-left: auto;
  display: flex;
  gap: 4px;
}

.range-btn {
  background: transparent;
  border: 1px solid #005000;
  color: #8a8;
  font-family: "Courier New", monospace;
  font-size: 11px;
  padding: 2px 10px;
  cursor: pointer;
  transition: all 0.2s;
}

.range-btn:hover {
  color: #0f0;
  border-color: #0f0;
}

.range-btn.active {
  background: rgba(0, 255, 0, 0.15);
  color: #0f0;
  border-color: #0f0;
}

.chart {
  width: 100%;
  height: 260px;
}

/* Bottom Row */
.bottom-row {
  display: grid;
  grid-template-columns: 1fr 280px;
  gap: 16px;
}

.counters-panel {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.counter-item {
  background: rgba(0, 40, 0, 0.5);
  border: 1px solid #003000;
  border-radius: 4px;
  padding: 14px 18px;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.counter-label {
  font-family: "Courier New", monospace;
  font-size: 12px;
  text-transform: uppercase;
}

.counter-value {
  font-family: "Courier New", monospace;
  font-size: 22px;
  font-weight: bold;
}

.empty-state {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 260px;
  font-family: "Courier New", monospace;
  font-size: 13px;
}

.dim-text {
  color: #8a8;
}
.bright-text {
  color: #fff;
  font-weight: bold;
}

@media (max-width: 1100px) {
  .stats-row {
    grid-template-columns: repeat(2, 1fr);
  }
  .charts-row {
    grid-template-columns: 1fr;
  }
  .charts-row .chart-panel.small {
    order: -1;
  }
  .bottom-row {
    grid-template-columns: 1fr;
  }
}
</style>
