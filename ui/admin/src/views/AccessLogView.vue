<template>
  <div class="access-log-view">
    <div class="terminal-card glass-panel">
      <div class="card-header no-select">
        <div class="header-left">
          <span class="prefix">root@system:~/firewall$</span>
          <span class="command blink-cursor">./access_log.sh</span>
        </div>
        <div class="header-right">
          <el-button type="danger" size="small" @click="handleClear">{{
            t("access_log.clear_all")
          }}</el-button>
          <el-button size="small" @click="fetchData">{{
            t("common.refresh")
          }}</el-button>
        </div>
      </div>

      <div class="filter-bar">
        <el-input
          v-model="queryParams.client_ip"
          :placeholder="t('access_log.filter_ip')"
          size="small"
          clearable
          style="width: 160px"
          @clear="fetchData"
          @keyup.enter="fetchData"
        />
        <el-select
          v-model="queryParams.action"
          :placeholder="t('access_log.filter_action')"
          size="small"
          clearable
          style="width: 160px"
          @change="fetchData"
        >
          <el-option label="Allowed" value="allowed" />
          <el-option label="Blocked" value="blocked" />
          <el-option label="Blocked (WAF)" value="blocked_waf" />
          <el-option label="Blocked (Rate)" value="blocked_ratelimit" />
          <el-option label="Blocked (Geo)" value="blocked_geo" />
        </el-select>
        <el-date-picker
          v-model="dateRange"
          type="datetimerange"
          size="small"
          :start-placeholder="t('access_log.filter_time_start')"
          :end-placeholder="t('access_log.filter_time_end')"
          value-format="YYYY-MM-DDTHH:mm:ssZ"
          @change="fetchData"
        />
        <el-button type="primary" size="small" @click="fetchData">{{
          t("access_log.search")
        }}</el-button>
      </div>

      <div class="data-grid">
        <el-table
          ref="tableRef"
          :data="tableData"
          v-loading="loading"
          style="width: 100%"
          class="hacker-table"
          row-key="id"
          @row-click="handleRowClick"
        >
          <el-table-column type="expand">
            <template #default="{ row }">
              <div class="expand-detail">
                <!-- Request Line -->
                <div class="request-line">
                  <span class="hl-method">{{ row.method }}</span>
                  <span class="hl-url"
                    >{{ row.scheme }}://{{ row.host }}{{ row.path
                    }}{{ row.query_string ? "?" + row.query_string : "" }}</span
                  >
                  <span
                    class="hl-status"
                    :style="{
                      color: row.status_code >= 400 ? '#f56c6c' : '#67c23a',
                    }"
                    >{{ row.status_code }}</span
                  >
                </div>

                <!-- Info Columns -->
                <div class="expand-columns">
                  <!-- Left: Overview -->
                  <div class="expand-col">
                    <div class="section-label">
                      {{ t("access_log.detail_overview") }}
                    </div>
                    <table class="kv-table">
                      <tr>
                        <td class="kv-key">{{ t("access_log.client_ip") }}</td>
                        <td class="kv-val bright-text">{{ row.client_ip }}</td>
                      </tr>
                      <tr>
                        <td class="kv-key">{{ t("access_log.action") }}</td>
                        <td class="kv-val">
                          <el-tag
                            :type="actionTag(row.action)"
                            size="small"
                            effect="dark"
                            >{{ row.action }}</el-tag
                          >
                        </td>
                      </tr>
                      <tr v-if="row.rule_name">
                        <td class="kv-key">{{ t("access_log.rule") }}</td>
                        <td class="kv-val">{{ row.rule_name }}</td>
                      </tr>
                      <tr>
                        <td class="kv-key">{{ t("access_log.latency") }}</td>
                        <td class="kv-val">
                          {{ (row.latency_us / 1000).toFixed(1) }} ms
                        </td>
                      </tr>
                      <tr>
                        <td class="kv-key">
                          {{ t("access_log.response_size") }}
                        </td>
                        <td class="kv-val">
                          {{ formatBytes(row.response_size) }}
                        </td>
                      </tr>
                      <tr>
                        <td class="kv-key">
                          {{ t("access_log.content_length") }}
                        </td>
                        <td class="kv-val">
                          {{ formatBytes(row.content_length) }}
                        </td>
                      </tr>
                      <tr v-if="row.content_type">
                        <td class="kv-key">
                          {{ t("access_log.content_type") }}
                        </td>
                        <td class="kv-val">{{ row.content_type }}</td>
                      </tr>
                      <tr v-if="row.referer">
                        <td class="kv-key">{{ t("access_log.referer") }}</td>
                        <td class="kv-val break-all">{{ row.referer }}</td>
                      </tr>
                      <tr>
                        <td class="kv-key">{{ t("access_log.user_agent") }}</td>
                        <td class="kv-val break-all">
                          {{ row.user_agent || "-" }}
                        </td>
                      </tr>
                      <tr>
                        <td class="kv-key">{{ t("common.created_at") }}</td>
                        <td class="kv-val">
                          {{ new Date(row.created_at).toLocaleString() }}
                        </td>
                      </tr>
                    </table>

                    <!-- Query Parameters -->
                    <template v-if="row.query_string">
                      <div class="section-label" style="margin-top: 16px">
                        {{ t("access_log.query_params") }}
                      </div>
                      <table class="kv-table">
                        <tr
                          v-for="(val, key) in parseQuery(row.query_string)"
                          :key="key"
                        >
                          <td class="kv-key">{{ key }}</td>
                          <td class="kv-val break-all">{{ val }}</td>
                        </tr>
                      </table>
                    </template>
                  </div>

                  <!-- Right: Headers + Body -->
                  <div class="expand-col">
                    <div class="section-label">
                      {{ t("access_log.request_headers") }}
                    </div>
                    <table
                      class="kv-table"
                      v-if="parseHeaders(row.request_headers)"
                    >
                      <tr
                        v-for="(vals, key) in parseHeaders(row.request_headers)"
                        :key="key"
                      >
                        <td class="kv-key">{{ key }}</td>
                        <td class="kv-val break-all">
                          {{ Array.isArray(vals) ? vals.join(", ") : vals }}
                        </td>
                      </tr>
                    </table>
                    <span v-else class="dim-text">-</span>

                    <template v-if="row.request_body">
                      <div class="section-label" style="margin-top: 16px">
                        {{ t("access_log.request_body") }}
                      </div>
                      <pre class="body-block">{{ row.request_body }}</pre>
                    </template>
                  </div>
                </div>
              </div>
            </template>
          </el-table-column>
          <el-table-column prop="id" label="ID" width="70">
            <template #default="scope">
              <span class="dim-text">#{{ scope.row.id }}</span>
            </template>
          </el-table-column>
          <el-table-column
            prop="client_ip"
            :label="t('access_log.client_ip')"
            width="140"
          >
            <template #default="scope">
              <span class="bright-text">{{ scope.row.client_ip }}</span>
            </template>
          </el-table-column>
          <el-table-column
            prop="method"
            :label="t('access_log.method')"
            width="70"
          >
            <template #default="scope">
              <span class="dim-text">{{ scope.row.method }}</span>
            </template>
          </el-table-column>
          <el-table-column
            prop="host"
            :label="t('access_log.host')"
            width="160"
          >
            <template #default="scope">
              <span class="dim-text">{{ scope.row.host || "-" }}</span>
            </template>
          </el-table-column>
          <el-table-column
            prop="path"
            :label="t('access_log.path')"
            min-width="180"
          >
            <template #default="scope">
              <span class="dim-text" style="word-break: break-all"
                >{{ scope.row.path
                }}{{
                  scope.row.query_string ? "?" + scope.row.query_string : ""
                }}</span
              >
            </template>
          </el-table-column>
          <el-table-column
            prop="status_code"
            :label="t('access_log.status')"
            width="70"
          >
            <template #default="scope">
              <span
                :style="{
                  color: scope.row.status_code >= 400 ? '#f56c6c' : '#67c23a',
                }"
              >
                {{ scope.row.status_code }}
              </span>
            </template>
          </el-table-column>
          <el-table-column
            prop="action"
            :label="t('access_log.action')"
            width="130"
          >
            <template #default="scope">
              <el-tag
                :type="actionTag(scope.row.action)"
                size="small"
                effect="dark"
              >
                {{ scope.row.action }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column
            prop="created_at"
            :label="t('common.created_at')"
            width="170"
          >
            <template #default="scope">
              <span class="dim-text">{{
                new Date(scope.row.created_at).toLocaleString()
              }}</span>
            </template>
          </el-table-column>
          <el-table-column
            :label="t('common.actions')"
            width="90"
            fixed="right"
          >
            <template #default="scope">
              <el-button
                type="danger"
                link
                size="small"
                class="action-link"
                @click.stop="handleBlockIp(scope.row)"
              >
                {{ t("access_log.block_ip") }}
              </el-button>
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
import { ref, onMounted, reactive } from "vue";
import { useRoute } from "vue-router";
import request from "@/utils/request";
import { ElMessage, ElMessageBox } from "element-plus";
import { useI18n } from "vue-i18n";

const { t } = useI18n();
const route = useRoute();
const tableRef = ref();
const tableData = ref<any[]>([]);
const total = ref(0);
const loading = ref(false);
const queryParams = reactive({
  page: 1,
  size: 20,
  client_ip: "",
  action: "",
});
const dateRange = ref<[string, string] | null>(null);

const actionTag = (action: string) => {
  if (action === "allowed") return "success";
  if (action.startsWith("blocked")) return "danger";
  return "info";
};

const formatBytes = (bytes: number) => {
  if (!bytes || bytes <= 0) return "0 B";
  if (bytes < 1024) return bytes + " B";
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + " KB";
  return (bytes / (1024 * 1024)).toFixed(1) + " MB";
};

const parseQuery = (qs: string): Record<string, string> => {
  const params: Record<string, string> = {};
  if (!qs) return params;
  const sp = new URLSearchParams(qs);
  sp.forEach((v, k) => {
    params[k] = v;
  });
  return params;
};

const parseHeaders = (raw: string): Record<string, string[]> | null => {
  if (!raw) return null;
  try {
    return JSON.parse(raw);
  } catch {
    return null;
  }
};

const fetchData = async () => {
  loading.value = true;
  try {
    const params: any = { page: queryParams.page, size: queryParams.size };
    if (queryParams.client_ip) params.client_ip = queryParams.client_ip;
    if (queryParams.action) params.action = queryParams.action;
    if (dateRange.value) {
      params.start_time = dateRange.value[0];
      params.end_time = dateRange.value[1];
    }
    const res: any = await request.get("/access-logs", { params });
    tableData.value = res.list || [];
    total.value = res.total || 0;
  } finally {
    loading.value = false;
  }
};

const handleClear = () => {
  ElMessageBox.confirm(t("access_log.clear_confirm"), t("common.warning"), {
    confirmButtonText: t("common.confirm"),
    cancelButtonText: t("common.cancel"),
    type: "warning",
  }).then(async () => {
    try {
      const res: any = await request.delete("/access-logs");
      ElMessage.success(t("access_log.cleared", { count: res?.deleted || 0 }));
      fetchData();
    } catch {
      // handled
    }
  });
};

const handleBlockIp = (row: any) => {
  ElMessageBox.confirm(
    t("access_log.block_ip_confirm", { ip: row.client_ip }),
    t("common.warning"),
    {
      confirmButtonText: t("common.confirm"),
      cancelButtonText: t("common.cancel"),
      type: "warning",
    },
  ).then(async () => {
    try {
      await request.post("/ip-black-list", { ip: row.client_ip });
      ElMessage.success(
        t("access_log.block_ip_success", { ip: row.client_ip }),
      );
      fetchData();
    } catch {
      // handled
    }
  });
};

const handleRowClick = (row: any, column: any, event: Event) => {
  const target = event.target as HTMLElement;
  if (target.closest('.action-link') || target.closest('.el-button')) return;
  tableRef.value?.toggleRowExpansion(row);
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
  if (route.query.client_ip) {
    queryParams.client_ip = route.query.client_ip as string;
  }
  fetchData();
});
</script>

<style scoped>
.access-log-view {
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
.filter-bar {
  padding: 12px 25px;
  display: flex;
  gap: 10px;
  align-items: center;
  border-bottom: 1px solid #003000;
  background: rgba(0, 40, 0, 0.2);
  flex-wrap: wrap;
}
.hacker-table {
  font-family: "Courier New", monospace;
}
.action-link {
  font-weight: bold;
  text-decoration: underline;
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
.break-all {
  word-break: break-all;
}

/* Expand row detail */
.expand-detail {
  font-family: "Courier New", monospace;
  font-size: 13px;
  color: #ccc;
  padding: 12px 8px;
}

.request-line {
  background: rgba(0, 40, 0, 0.5);
  padding: 10px 14px;
  border-radius: 4px;
  margin-bottom: 16px;
  word-break: break-all;
  line-height: 1.6;
  border-left: 3px solid #0a0;
}

.hl-method {
  color: #0f0;
  font-weight: bold;
  margin-right: 10px;
}

.hl-url {
  color: #ddd;
}

.hl-status {
  margin-left: 12px;
  font-weight: bold;
}

.expand-columns {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 24px;
}

.section-label {
  color: #0f0;
  font-size: 11px;
  font-weight: bold;
  letter-spacing: 1px;
  margin-bottom: 6px;
  padding-bottom: 4px;
  border-bottom: 1px solid #003000;
}

.kv-table {
  width: 100%;
  border-collapse: collapse;
}

.kv-table td {
  padding: 3px 0;
  vertical-align: top;
  line-height: 1.5;
}

.kv-key {
  color: #007000;
  width: 160px;
  font-size: 12px;
  padding-right: 12px !important;
  white-space: nowrap;
}

.kv-val {
  color: #ccc;
  font-size: 12px;
}

.body-block {
  background: rgba(0, 40, 0, 0.5);
  padding: 10px 12px;
  border-radius: 4px;
  color: #ccc;
  font-size: 12px;
  white-space: pre-wrap;
  word-break: break-all;
  max-height: 240px;
  overflow-y: auto;
  margin: 0;
  border-left: 3px solid #0a0;
}
</style>

<style>
/* Clickable rows */
.hacker-table .el-table__body tr {
  cursor: pointer;
}
/* Expand row background override */
.hacker-table .el-table__expanded-cell {
  background: rgba(0, 20, 0, 0.6) !important;
  border-bottom: 1px solid #003000;
}
</style>
