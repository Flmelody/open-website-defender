<template>
  <div class="bot-management-view">
    <div class="terminal-card glass-panel">
      <div class="card-header no-select">
        <div class="header-left">
          <span class="prefix">root@system:~/bot$</span>
          <span class="command blink-cursor">./bot_management.sh</span>
        </div>
        <div class="header-right">
          <el-button type="primary" size="small" @click="handleAdd">{{
            t("bot.new_signature")
          }}</el-button>
          <el-button size="small" @click="fetchData">{{
            t("common.refresh")
          }}</el-button>
        </div>
      </div>

      <div class="desc-bar no-select">
        <span class="dim-text">{{ t("bot.desc") }}</span>
      </div>

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
          <el-table-column prop="name" :label="t('bot.name')" min-width="150">
            <template #default="scope">
              <span class="bright-text">{{ scope.row.name }}</span>
            </template>
          </el-table-column>
          <el-table-column
            prop="category"
            :label="t('bot.category')"
            width="130"
          >
            <template #default="scope">
              <el-tag
                :type="categoryTag(scope.row.category)"
                size="small"
                effect="dark"
              >
                {{ categoryLabel(scope.row.category) }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column
            prop="match_target"
            :label="t('bot.match_target')"
            width="120"
          >
            <template #default="scope">
              <span class="dim-text">{{
                targetLabel(scope.row.match_target)
              }}</span>
            </template>
          </el-table-column>
          <el-table-column
            prop="pattern"
            :label="t('bot.pattern')"
            min-width="200"
          >
            <template #default="scope">
              <span
                class="dim-text"
                style="font-size: 12px; word-break: break-all"
                >{{ scope.row.pattern }}</span
              >
            </template>
          </el-table-column>
          <el-table-column prop="action" :label="t('bot.action')" width="100">
            <template #default="scope">
              <span :style="{ color: actionColor(scope.row.action) }">
                {{ actionLabel(scope.row.action) }}
              </span>
            </template>
          </el-table-column>
          <el-table-column prop="enabled" :label="t('bot.enabled')" width="80">
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

      <div class="card-footer no-select">
        <span class="status-text">{{
          t("common.total_records", { total })
        }}</span>
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
      width="600px"
      destroy-on-close
    >
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-position="top"
        class="hacker-form"
      >
        <el-form-item :label="'> ' + t('bot.name')" prop="name">
          <el-input v-model="form.name" placeholder="Googlebot" />
        </el-form-item>
        <el-form-item :label="'> ' + t('bot.pattern')" prop="pattern">
          <el-input
            v-model="form.pattern"
            placeholder="(?i)googlebot"
            type="textarea"
            :rows="2"
          />
        </el-form-item>
        <el-row :gutter="16">
          <el-col :span="8">
            <el-form-item
              :label="'> ' + t('bot.match_target')"
              prop="match_target"
            >
              <el-select v-model="form.match_target" style="width: 100%">
                <el-option :label="t('bot.target_ua')" value="ua" />
                <el-option :label="t('bot.target_header')" value="header" />
                <el-option :label="t('bot.target_behavior')" value="behavior" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item :label="'> ' + t('bot.category')" prop="category">
              <el-select v-model="form.category" style="width: 100%">
                <el-option
                  :label="t('bot.category_search_engine')"
                  value="search_engine"
                />
                <el-option
                  :label="t('bot.category_good_bot')"
                  value="good_bot"
                />
                <el-option
                  :label="t('bot.category_malicious')"
                  value="malicious"
                />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item :label="'> ' + t('bot.action')" prop="action">
              <el-select v-model="form.action" style="width: 100%">
                <el-option :label="t('bot.action_allow')" value="allow" />
                <el-option :label="t('bot.action_block')" value="block" />
                <el-option
                  :label="t('bot.action_challenge')"
                  value="challenge"
                />
                <el-option :label="t('bot.action_monitor')" value="monitor" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
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
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, reactive, computed } from "vue";
import request from "@/utils/request";
import { ElMessage, ElMessageBox } from "element-plus";
import { useI18n } from "vue-i18n";

interface BotSignature {
  id: number;
  name: string;
  pattern: string;
  match_target: string;
  category: string;
  action: string;
  enabled: boolean;
  created_at: string;
}

const { t } = useI18n();
const tableData = ref<BotSignature[]>([]);
const total = ref(0);
const loading = ref(false);
const queryParams = reactive({ page: 1, size: 10 });

const dialogVisible = ref(false);
const dialogTitle = ref("");
const formRef = ref();
const formLoading = ref(false);
const editingId = ref<number | null>(null);

const form = reactive({
  name: "",
  pattern: "",
  match_target: "ua",
  category: "malicious",
  action: "block",
});

const rules = computed(() => ({
  name: [{ required: true, message: t("login.required"), trigger: "blur" }],
  pattern: [{ required: true, message: t("login.required"), trigger: "blur" }],
  match_target: [
    { required: true, message: t("login.required"), trigger: "change" },
  ],
  category: [
    { required: true, message: t("login.required"), trigger: "change" },
  ],
}));

const categoryTag = (cat: string) => {
  const map: Record<string, string> = {
    malicious: "danger",
    search_engine: "success",
    good_bot: "",
  };
  return map[cat] || "info";
};

const categoryLabel = (cat: string) => {
  const map: Record<string, string> = {
    malicious: t("bot.category_malicious"),
    search_engine: t("bot.category_search_engine"),
    good_bot: t("bot.category_good_bot"),
  };
  return map[cat] || cat.toUpperCase();
};

const targetLabel = (target: string) => {
  const map: Record<string, string> = {
    ua: t("bot.target_ua"),
    header: t("bot.target_header"),
    behavior: t("bot.target_behavior"),
  };
  return map[target] || target;
};

const actionColor = (action: string) => {
  const map: Record<string, string> = {
    allow: "#67c23a",
    block: "#f56c6c",
    challenge: "#e6a23c",
    monitor: "#909399",
  };
  return map[action] || "#e6a23c";
};

const actionLabel = (action: string) => {
  const map: Record<string, string> = {
    allow: t("bot.action_allow"),
    block: t("bot.action_block"),
    challenge: t("bot.action_challenge"),
    monitor: t("bot.action_monitor"),
  };
  return map[action] || action.toUpperCase();
};

const fetchData = async () => {
  loading.value = true;
  try {
    const res: any = await request.get("/bot-signatures", {
      params: queryParams,
    });
    tableData.value = res.list || [];
    total.value = res.total || 0;
  } finally {
    loading.value = false;
  }
};

const handleAdd = () => {
  editingId.value = null;
  dialogTitle.value = t("bot.title_create");
  form.name = "";
  form.pattern = "";
  form.match_target = "ua";
  form.category = "malicious";
  form.action = "block";
  dialogVisible.value = true;
};

const handleEdit = (row: BotSignature) => {
  editingId.value = row.id;
  dialogTitle.value = t("bot.title_edit");
  form.name = row.name;
  form.pattern = row.pattern;
  form.match_target = row.match_target;
  form.category = row.category;
  form.action = row.action;
  dialogVisible.value = true;
};

const handleToggle = async (row: BotSignature) => {
  try {
    await request.put(`/bot-signatures/${row.id}`, { enabled: row.enabled });
    ElMessage.success(t("common.updated"));
  } catch {
    row.enabled = !row.enabled;
  }
};

const handleDelete = (row: BotSignature) => {
  ElMessageBox.confirm(
    t("bot.delete_confirm", { name: row.name }),
    t("common.warning"),
    {
      confirmButtonText: t("common.remove"),
      cancelButtonText: t("common.cancel"),
      type: "warning",
    },
  ).then(async () => {
    try {
      await request.delete(`/bot-signatures/${row.id}`);
      ElMessage.success(t("common.deleted"));
      fetchData();
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
          await request.put(`/bot-signatures/${editingId.value}`, form);
          ElMessage.success(t("common.updated"));
        } else {
          await request.post("/bot-signatures", form);
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
.bot-management-view {
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
.desc-bar {
  padding: 12px 25px;
  border-bottom: 1px solid #003000;
  font-size: 13px;
  font-family: "Courier New", monospace;
}
</style>
