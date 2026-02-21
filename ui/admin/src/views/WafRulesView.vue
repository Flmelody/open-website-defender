<template>
  <div class="waf-rules-view">
    <div class="terminal-card glass-panel">
      <div class="card-header no-select">
        <div class="header-left">
          <span class="prefix">root@system:~/firewall$</span>
          <span class="command blink-cursor">./list_rules.sh</span>
        </div>
        <div class="header-right">
          <el-select
            v-model="filterGroup"
            size="small"
            style="width: 160px"
            clearable
            :placeholder="t('waf.filter_group')"
            @change="fetchData"
          >
            <el-option :label="t('waf.all_groups')" value="" />
            <el-option v-for="g in groups" :key="g" :label="g" :value="g" />
          </el-select>
          <el-button type="primary" size="small" @click="handleAdd">{{
            activeTab === "rules" ? t("waf.new_rule") : t("waf.new_exclusion")
          }}</el-button>
          <el-button size="small" @click="fetchData">{{
            t("common.refresh")
          }}</el-button>
        </div>
      </div>

      <el-tabs
        v-model="activeTab"
        class="waf-tabs"
        @tab-change="handleTabChange"
      >
        <el-tab-pane :label="t('waf.rules_tab')" name="rules">
          <div class="data-grid">
            <el-table
              :data="tableData"
              v-loading="loading"
              style="width: 100%"
              class="hacker-table"
            >
              <el-table-column prop="id" label="ID" width="60">
                <template #default="scope">
                  <span class="dim-text">#{{ scope.row.id }}</span>
                </template>
              </el-table-column>
              <el-table-column
                prop="priority"
                :label="t('waf.priority')"
                width="80"
                sortable
              >
                <template #default="scope">
                  <span class="dim-text">{{ scope.row.priority }}</span>
                </template>
              </el-table-column>
              <el-table-column
                prop="name"
                :label="t('waf.name')"
                min-width="150"
              >
                <template #default="scope">
                  <span class="bright-text">{{ scope.row.name }}</span>
                </template>
              </el-table-column>
              <el-table-column
                prop="group_name"
                :label="t('waf.group_name')"
                width="110"
              >
                <template #default="scope">
                  <el-tag
                    v-if="scope.row.group_name"
                    size="small"
                    effect="plain"
                    >{{ scope.row.group_name }}</el-tag
                  >
                  <span v-else class="dim-text">—</span>
                </template>
              </el-table-column>
              <el-table-column
                prop="category"
                :label="t('waf.category')"
                width="100"
              >
                <template #default="scope">
                  <el-tag
                    :type="categoryTag(scope.row.category)"
                    size="small"
                    effect="dark"
                  >
                    {{ scope.row.category.toUpperCase() }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column
                prop="target"
                :label="t('waf.target')"
                width="90"
              >
                <template #default="scope">
                  <span class="dim-text">{{ scope.row.target || "all" }}</span>
                </template>
              </el-table-column>
              <el-table-column
                prop="operator"
                :label="t('waf.operator')"
                width="90"
              >
                <template #default="scope">
                  <span class="dim-text">{{
                    scope.row.operator || "regex"
                  }}</span>
                </template>
              </el-table-column>
              <el-table-column
                prop="pattern"
                :label="t('waf.pattern')"
                min-width="180"
              >
                <template #default="scope">
                  <span
                    class="dim-text"
                    style="font-size: 12px; word-break: break-all"
                    >{{ scope.row.pattern }}</span
                  >
                </template>
              </el-table-column>
              <el-table-column
                prop="action"
                :label="t('waf.action')"
                width="90"
              >
                <template #default="scope">
                  <span :style="{ color: actionColor(scope.row.action) }">
                    {{ scope.row.action.toUpperCase() }}
                  </span>
                </template>
              </el-table-column>
              <el-table-column
                prop="enabled"
                :label="t('waf.enabled')"
                width="80"
              >
                <template #default="scope">
                  <el-switch
                    v-model="scope.row.enabled"
                    size="small"
                    @change="handleToggle(scope.row)"
                  />
                </template>
              </el-table-column>
              <el-table-column
                :label="t('common.actions')"
                width="140"
                align="right"
              >
                <template #default="scope">
                  <div class="ops-cell">
                    <el-button
                      type="primary"
                      link
                      size="small"
                      @click="handleEdit(scope.row)"
                      class="action-link"
                    >
                      {{ t("common.edit") }}
                    </el-button>
                    <el-button
                      type="danger"
                      link
                      size="small"
                      @click="handleDelete(scope.row)"
                      class="action-link delete"
                    >
                      {{ t("common.delete") }}
                    </el-button>
                  </div>
                </template>
              </el-table-column>
            </el-table>
          </div>
        </el-tab-pane>

        <el-tab-pane :label="t('waf.exclusions_tab')" name="exclusions">
          <div class="data-grid">
            <el-table
              :data="exclusionData"
              v-loading="exLoading"
              style="width: 100%"
              class="hacker-table"
            >
              <el-table-column prop="id" label="ID" width="60">
                <template #default="scope">
                  <span class="dim-text">#{{ scope.row.id }}</span>
                </template>
              </el-table-column>
              <el-table-column
                prop="path"
                :label="t('waf.exclusion_path')"
                min-width="200"
              >
                <template #default="scope">
                  <span class="bright-text">{{ scope.row.path }}</span>
                </template>
              </el-table-column>
              <el-table-column
                prop="operator"
                :label="t('waf.exclusion_operator')"
                width="120"
              >
                <template #default="scope">
                  <span class="dim-text">{{ scope.row.operator }}</span>
                </template>
              </el-table-column>
              <el-table-column
                prop="rule_id"
                :label="t('waf.exclusion_rule_id')"
                width="120"
              >
                <template #default="scope">
                  <span class="dim-text">{{
                    scope.row.rule_id === 0
                      ? t("waf.exclusion_all_rules")
                      : "#" + scope.row.rule_id
                  }}</span>
                </template>
              </el-table-column>
              <el-table-column
                prop="enabled"
                :label="t('waf.enabled')"
                width="80"
              >
                <template #default="scope">
                  <el-tag
                    :type="scope.row.enabled ? 'success' : 'info'"
                    size="small"
                    effect="dark"
                  >
                    {{ scope.row.enabled ? "ON" : "OFF" }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column
                :label="t('common.actions')"
                width="100"
                align="right"
              >
                <template #default="scope">
                  <el-button
                    type="danger"
                    link
                    size="small"
                    @click="handleDeleteExclusion(scope.row)"
                    class="action-link delete"
                  >
                    {{ t("common.delete") }}
                  </el-button>
                </template>
              </el-table-column>
            </el-table>
          </div>
        </el-tab-pane>
      </el-tabs>

      <div class="card-footer no-select">
        <span class="status-text">{{
          t("common.total_records", {
            total: activeTab === "rules" ? total : exTotal,
          })
        }}</span>
        <el-pagination
          v-model:current-page="queryParams.page"
          v-model:page-size="queryParams.size"
          :page-sizes="[10, 20, 50]"
          layout="sizes, prev, pager, next"
          :total="activeTab === 'rules' ? total : exTotal"
          @size-change="handleSizeChange"
          @current-change="handleCurrentChange"
          small
        />
      </div>
    </div>

    <!-- Rule dialog -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogTitle.toUpperCase()"
      width="700px"
      destroy-on-close
    >
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-position="top"
        class="hacker-form"
      >
        <el-form-item :label="'> ' + t('waf.name')" prop="name">
          <el-input v-model="form.name" placeholder="SQLi - Union Select" />
        </el-form-item>
        <el-row :gutter="16">
          <el-col :span="8">
            <el-form-item :label="'> ' + t('waf.operator')" prop="operator">
              <el-select v-model="form.operator" style="width: 100%">
                <el-option :label="t('waf.operator_regex')" value="regex" />
                <el-option
                  :label="t('waf.operator_contains')"
                  value="contains"
                />
                <el-option :label="t('waf.operator_prefix')" value="prefix" />
                <el-option :label="t('waf.operator_suffix')" value="suffix" />
                <el-option :label="t('waf.operator_equals')" value="equals" />
                <el-option :label="t('waf.operator_gt')" value="gt" />
                <el-option :label="t('waf.operator_lt')" value="lt" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item :label="'> ' + t('waf.target')" prop="target">
              <el-select v-model="form.target" style="width: 100%">
                <el-option :label="t('waf.target_all')" value="all" />
                <el-option :label="t('waf.target_url')" value="url" />
                <el-option :label="t('waf.target_headers')" value="headers" />
                <el-option :label="t('waf.target_body')" value="body" />
                <el-option :label="t('waf.target_cookies')" value="cookies" />
                <el-option :label="t('waf.target_query')" value="query" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item prop="priority">
              <template #label>
                <span class="label-with-tip"
                  >> {{ t("waf.priority") }}
                  <el-tooltip
                    :content="t('waf.priority_hint')"
                    placement="top"
                    effect="dark"
                  >
                    <el-icon class="info-icon"><InfoFilled /></el-icon>
                  </el-tooltip>
                </span>
              </template>
              <el-input-number
                v-model="form.priority"
                :min="0"
                :max="9999"
                style="width: 100%"
              />
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item :label="'> ' + t('waf.pattern')" prop="pattern">
          <el-input
            v-model="form.pattern"
            placeholder="(?i)(union\s+select)"
            type="textarea"
            :rows="2"
          />
        </el-form-item>
        <el-row :gutter="16">
          <el-col :span="8">
            <el-form-item :label="'> ' + t('waf.category')" prop="category">
              <el-select v-model="form.category" style="width: 100%">
                <el-option label="SQLi" value="sqli" />
                <el-option label="XSS" value="xss" />
                <el-option label="Traversal" value="traversal" />
                <el-option label="UA" value="ua" />
                <el-option label="Custom" value="custom" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item :label="'> ' + t('waf.action')" prop="action">
              <el-select v-model="form.action" style="width: 100%">
                <el-option :label="t('waf.action_block')" value="block" />
                <el-option :label="t('waf.action_log')" value="log" />
                <el-option :label="t('waf.action_redirect')" value="redirect" />
                <el-option
                  :label="t('waf.action_challenge')"
                  value="challenge"
                />
                <el-option
                  :label="t('waf.action_rate_limit')"
                  value="rate-limit"
                />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item :label="'> ' + t('waf.group_name')">
              <el-input
                v-model="form.group_name"
                placeholder="e.g. sqli-core"
              />
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item
          v-if="form.action === 'redirect'"
          :label="'> ' + t('waf.redirect_url')"
        >
          <el-input
            v-model="form.redirect_url"
            placeholder="https://example.com/blocked"
          />
        </el-form-item>
        <el-form-item v-if="form.action === 'rate-limit'">
          <template #label>
            <span class="label-with-tip"
              >> {{ t("waf.rate_limit") }}
              <el-tooltip
                :content="t('waf.rate_limit_hint')"
                placement="top"
                effect="dark"
              >
                <el-icon class="info-icon"><InfoFilled /></el-icon>
              </el-tooltip>
            </span>
          </template>
          <el-input-number v-model="form.rate_limit" :min="1" :max="10000" />
        </el-form-item>
        <el-form-item :label="'> ' + t('waf.description')">
          <el-input v-model="form.description" type="textarea" :rows="2" />
        </el-form-item>
      </el-form>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="dialogVisible = false">{{
            t("common.cancel")
          }}</el-button>
          <el-button
            type="primary"
            :loading="formLoading"
            @click="handleSubmit"
            >{{ t("common.confirm") }}</el-button
          >
        </div>
      </template>
    </el-dialog>

    <!-- Exclusion dialog -->
    <el-dialog
      v-model="exDialogVisible"
      :title="t('waf.title_create_exclusion').toUpperCase()"
      width="500px"
      destroy-on-close
    >
      <el-form
        ref="exFormRef"
        :model="exForm"
        :rules="exRules"
        label-position="top"
        class="hacker-form"
      >
        <el-form-item :label="'> ' + t('waf.exclusion_path')" prop="path">
          <el-input v-model="exForm.path" placeholder="/api/webhook" />
        </el-form-item>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item
              :label="'> ' + t('waf.exclusion_operator')"
              prop="operator"
            >
              <el-select v-model="exForm.operator" style="width: 100%">
                <el-option label="Prefix" value="prefix" />
                <el-option label="Exact" value="exact" />
                <el-option label="Regex" value="regex" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item>
              <template #label>
                <span class="label-with-tip"
                  >> {{ t("waf.exclusion_rule_id") }}
                  <el-tooltip
                    :content="'0 = ' + t('waf.exclusion_all_rules')"
                    placement="top"
                    effect="dark"
                  >
                    <el-icon class="info-icon"><InfoFilled /></el-icon>
                  </el-tooltip>
                </span>
              </template>
              <el-input-number
                v-model="exForm.rule_id"
                :min="0"
                style="width: 100%"
              />
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="exDialogVisible = false">{{
            t("common.cancel")
          }}</el-button>
          <el-button
            type="primary"
            :loading="exFormLoading"
            @click="handleExSubmit"
            >{{ t("common.confirm") }}</el-button
          >
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, reactive, computed } from "vue";
import request from "@/utils/request";
import { ElMessage, ElMessageBox } from "element-plus";
import { InfoFilled } from "@element-plus/icons-vue";
import { useI18n } from "vue-i18n";

interface WafRule {
  id: number;
  name: string;
  pattern: string;
  category: string;
  action: string;
  operator: string;
  target: string;
  priority: number;
  group_name: string;
  redirect_url: string;
  rate_limit: number;
  description: string;
  enabled: boolean;
  created_at: string;
}

interface WafExclusion {
  id: number;
  rule_id: number;
  path: string;
  operator: string;
  enabled: boolean;
  created_at: string;
}

const { t } = useI18n();
const activeTab = ref("rules");
const tableData = ref<WafRule[]>([]);
const total = ref(0);
const loading = ref(false);
const groups = ref<string[]>([]);
const filterGroup = ref("");
const queryParams = reactive({ page: 1, size: 10 });

// Exclusion state
const exclusionData = ref<WafExclusion[]>([]);
const exTotal = ref(0);
const exLoading = ref(false);

const dialogVisible = ref(false);
const dialogTitle = ref("");
const formRef = ref();
const formLoading = ref(false);
const editingId = ref<number | null>(null);

const form = reactive({
  name: "",
  pattern: "",
  category: "custom",
  action: "block",
  operator: "regex",
  target: "all",
  priority: 100,
  group_name: "",
  redirect_url: "",
  rate_limit: 0,
  description: "",
});

// Exclusion dialog
const exDialogVisible = ref(false);
const exFormRef = ref();
const exFormLoading = ref(false);
const exForm = reactive({
  path: "",
  operator: "prefix",
  rule_id: 0,
});

const rules = computed(() => ({
  name: [{ required: true, message: t("login.required"), trigger: "blur" }],
  pattern: [{ required: true, message: t("login.required"), trigger: "blur" }],
  category: [
    { required: true, message: t("login.required"), trigger: "change" },
  ],
}));

const exRules = computed(() => ({
  path: [{ required: true, message: t("login.required"), trigger: "blur" }],
  operator: [
    { required: true, message: t("login.required"), trigger: "change" },
  ],
}));

const categoryTag = (cat: string) => {
  const map: Record<string, string> = {
    sqli: "danger",
    xss: "warning",
    traversal: "",
    ua: "info",
    custom: "info",
  };
  return map[cat] || "info";
};

const actionColor = (action: string) => {
  const map: Record<string, string> = {
    block: "#f56c6c",
    log: "#e6a23c",
    redirect: "#409eff",
    challenge: "#e6a23c",
    "rate-limit": "#909399",
  };
  return map[action] || "#e6a23c";
};

const fetchData = async () => {
  if (activeTab.value === "exclusions") {
    await fetchExclusions();
    return;
  }
  loading.value = true;
  try {
    const params: any = { ...queryParams };
    if (filterGroup.value) params.group = filterGroup.value;
    const res: any = await request.get("/waf-rules", { params });
    tableData.value = res.list || [];
    total.value = res.total || 0;
    // Extract unique groups
    const allGroups = new Set<string>();
    tableData.value.forEach((r: WafRule) => {
      if (r.group_name) allGroups.add(r.group_name);
    });
    groups.value = [...allGroups].sort();
  } finally {
    loading.value = false;
  }
};

const fetchExclusions = async () => {
  exLoading.value = true;
  try {
    const res: any = await request.get("/waf-exclusions", {
      params: queryParams,
    });
    exclusionData.value = res.list || [];
    exTotal.value = res.total || 0;
  } finally {
    exLoading.value = false;
  }
};

const handleTabChange = () => {
  queryParams.page = 1;
  fetchData();
};

const handleAdd = () => {
  if (activeTab.value === "exclusions") {
    exForm.path = "";
    exForm.operator = "prefix";
    exForm.rule_id = 0;
    exDialogVisible.value = true;
    return;
  }
  editingId.value = null;
  dialogTitle.value = t("waf.title_create");
  form.name = "";
  form.pattern = "";
  form.category = "custom";
  form.action = "block";
  form.operator = "regex";
  form.target = "all";
  form.priority = 100;
  form.group_name = "";
  form.redirect_url = "";
  form.rate_limit = 0;
  form.description = "";
  dialogVisible.value = true;
};

const handleEdit = (row: WafRule) => {
  editingId.value = row.id;
  dialogTitle.value = t("waf.title_edit");
  form.name = row.name;
  form.pattern = row.pattern;
  form.category = row.category;
  form.action = row.action;
  form.operator = row.operator || "regex";
  form.target = row.target || "all";
  form.priority = row.priority ?? 100;
  form.group_name = row.group_name || "";
  form.redirect_url = row.redirect_url || "";
  form.rate_limit = row.rate_limit || 0;
  form.description = row.description || "";
  dialogVisible.value = true;
};

const handleToggle = async (row: WafRule) => {
  try {
    await request.put(`/waf-rules/${row.id}`, { enabled: row.enabled });
    ElMessage.success(t("common.updated"));
  } catch {
    row.enabled = !row.enabled;
  }
};

const handleDelete = (row: WafRule) => {
  ElMessageBox.confirm(
    t("waf.delete_confirm", { name: row.name }),
    t("common.warning"),
    {
      confirmButtonText: t("common.remove"),
      cancelButtonText: t("common.cancel"),
      type: "warning",
    },
  ).then(async () => {
    try {
      await request.delete(`/waf-rules/${row.id}`);
      ElMessage.success(t("common.deleted"));
      fetchData();
    } catch {
      /* handled */
    }
  });
};

const handleDeleteExclusion = (row: WafExclusion) => {
  ElMessageBox.confirm(
    t("waf.delete_exclusion_confirm", { path: row.path }),
    t("common.warning"),
    {
      confirmButtonText: t("common.remove"),
      cancelButtonText: t("common.cancel"),
      type: "warning",
    },
  ).then(async () => {
    try {
      await request.delete(`/waf-exclusions/${row.id}`);
      ElMessage.success(t("common.deleted"));
      fetchExclusions();
    } catch {
      /* handled */
    }
  });
};

const handleSubmit = async () => {
  if (!formRef.value) return;
  await formRef.value.validate(async (valid: boolean) => {
    if (valid) {
      formLoading.value = true;
      try {
        if (editingId.value) {
          await request.put(`/waf-rules/${editingId.value}`, form);
          ElMessage.success(t("common.updated"));
        } else {
          await request.post("/waf-rules", form);
          ElMessage.success(t("common.added"));
        }
        dialogVisible.value = false;
        fetchData();
      } finally {
        formLoading.value = false;
      }
    }
  });
};

const handleExSubmit = async () => {
  if (!exFormRef.value) return;
  await exFormRef.value.validate(async (valid: boolean) => {
    if (valid) {
      exFormLoading.value = true;
      try {
        await request.post("/waf-exclusions", exForm);
        ElMessage.success(t("common.added"));
        exDialogVisible.value = false;
        fetchExclusions();
      } finally {
        exFormLoading.value = false;
      }
    }
  });
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
});
</script>

<style scoped>
.waf-rules-view {
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
  gap: 8px;
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
  font-size: 15px;
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
  color: #0f0;
  font-size: 13px;
  font-family: "Courier New", monospace;
}
.hacker-form :deep(.el-form-item__label) {
  color: #0f0 !important;
  font-weight: bold;
  font-size: 14px;
}
.dialog-footer {
  text-align: right;
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}
.ops-cell {
  display: flex;
  gap: 8px;
  justify-content: flex-end;
}
.waf-tabs {
  padding: 0 25px;
}
.waf-tabs :deep(.el-tabs__item) {
  color: #8a8;
}
.waf-tabs :deep(.el-tabs__item.is-active) {
  color: #0f0;
}
.waf-tabs :deep(.el-tabs__active-bar) {
  background-color: #0f0;
}
.waf-tabs :deep(.el-tabs__nav-wrap::after) {
  display: none !important;
}
</style>
