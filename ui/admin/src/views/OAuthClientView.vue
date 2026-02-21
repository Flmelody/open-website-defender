<template>
  <div class="oauth-client-view">
    <div class="terminal-card glass-panel">
      <div class="card-header no-select">
        <div class="header-left">
          <span class="prefix">root@system:~/oauth$</span>
          <span class="command blink-cursor">./list_clients.sh</span>
        </div>
        <div class="header-right">
          <el-button type="primary" size="small" @click="handleAdd">{{
            t("oauth.new_client")
          }}</el-button>
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
          <el-table-column prop="id" label="ID" width="60">
            <template #default="scope">
              <span class="dim-text">#{{ scope.row.id }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="name" :label="t('oauth.name')">
            <template #default="scope">
              <span class="bright-text">{{ scope.row.name }}</span>
            </template>
          </el-table-column>
          <el-table-column
            prop="client_id"
            :label="t('oauth.client_id')"
            width="200"
          >
            <template #default="scope">
              <span class="dim-text mono-text"
                >{{ scope.row.client_id.substring(0, 16) }}...</span
              >
            </template>
          </el-table-column>
          <el-table-column
            prop="trusted"
            :label="t('oauth.trusted')"
            width="100"
          >
            <template #default="scope">
              <el-tag
                :type="scope.row.trusted ? 'warning' : 'info'"
                size="small"
                effect="dark"
              >
                {{ scope.row.trusted ? t("oauth.yes") : t("oauth.no") }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="active" :label="t('oauth.status')" width="100">
            <template #default="scope">
              <el-tag
                :type="scope.row.active ? 'success' : 'danger'"
                size="small"
                effect="dark"
              >
                {{ scope.row.active ? t("oauth.active") : t("oauth.inactive") }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column
            prop="created_at"
            :label="t('common.created_at')"
            width="180"
          >
            <template #default="scope">
              <span class="dim-text">{{
                new Date(scope.row.created_at).toLocaleString()
              }}</span>
            </template>
          </el-table-column>
          <el-table-column
            :label="t('common.actions')"
            width="150"
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

    <!-- Create/Edit Client Dialog -->
    <el-dialog
      v-model="dialogVisible"
      :title="
        (isEdit ? t('oauth.title_edit') : t('oauth.title_create')).toUpperCase()
      "
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
        <el-form-item :label="'> ' + t('oauth.name')" prop="name">
          <el-input
            v-model="form.name"
            :placeholder="t('oauth.name_placeholder')"
          />
        </el-form-item>
        <el-form-item prop="redirect_uris_text">
          <template #label>
            <span class="label-with-tip"
              >> {{ t("oauth.redirect_uris") }}
              <el-tooltip
                :content="t('oauth.redirect_uris_hint')"
                placement="top"
                effect="dark"
              >
                <el-icon class="info-icon"><InfoFilled /></el-icon>
              </el-tooltip>
            </span>
          </template>
          <el-input
            v-model="form.redirect_uris_text"
            type="textarea"
            :rows="3"
            :placeholder="t('oauth.redirect_uris_placeholder')"
          />
        </el-form-item>
        <el-form-item :label="'> ' + t('oauth.scopes')">
          <el-input v-model="form.scopes" placeholder="openid profile email" />
        </el-form-item>
        <el-form-item>
          <el-checkbox v-model="form.trusted">{{
            t("oauth.trusted_label")
          }}</el-checkbox>
        </el-form-item>
        <el-form-item v-if="isEdit">
          <el-checkbox v-model="form.active">{{
            t("oauth.active_label")
          }}</el-checkbox>
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

    <!-- Secret Result Dialog (shown only on create) -->
    <el-dialog
      v-model="secretDialogVisible"
      :title="t('oauth.secret_generated').toUpperCase()"
      width="700px"
      :close-on-click-modal="false"
      :close-on-press-escape="false"
    >
      <div class="token-result">
        <div class="token-warning">
          <el-icon class="warning-icon">
            <WarningFilled />
          </el-icon>
          <span>{{ t("oauth.secret_warning") }}</span>
        </div>
        <div class="info-row">
          <span class="info-label">Client ID:</span>
          <code class="info-value">{{ createdClient.client_id }}</code>
          <el-button
            type="primary"
            size="small"
            class="copy-btn"
            @click="copyText(createdClient.client_id)"
          >
            <el-icon>
              <CopyDocument />
            </el-icon>
          </el-button>
        </div>
        <div class="info-row">
          <span class="info-label">Client Secret:</span>
          <code class="info-value highlight">{{
            createdClient.client_secret
          }}</code>
          <el-button
            type="primary"
            size="small"
            class="copy-btn"
            @click="copyText(createdClient.client_secret)"
          >
            <el-icon>
              <CopyDocument />
            </el-icon>
          </el-button>
        </div>
      </div>
      <template #footer>
        <div class="dialog-footer">
          <el-button type="primary" @click="closeSecretDialog">{{
            t("oauth.understood")
          }}</el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from "vue";
import request from "@/utils/request";
import { ElMessage, ElMessageBox } from "element-plus";
import {
  CopyDocument,
  InfoFilled,
  WarningFilled,
} from "@element-plus/icons-vue";
import { useI18n } from "vue-i18n";

interface OAuthClientItem {
  id: number;
  client_id: string;
  name: string;
  redirect_uris: string[];
  scopes: string;
  trusted: boolean;
  active: boolean;
  created_at: string;
}

interface CreatedClient {
  client_id: string;
  client_secret: string;
}

const { t } = useI18n();
const tableData = ref<OAuthClientItem[]>([]);
const total = ref(0);
const loading = ref(false);
const queryParams = reactive({ page: 1, size: 10 });

const dialogVisible = ref(false);
const isEdit = ref(false);
const editId = ref<number>(0);
const formRef = ref();
const formLoading = ref(false);
const form = reactive({
  name: "",
  redirect_uris_text: "",
  scopes: "openid profile email",
  trusted: false,
  active: true,
});

const secretDialogVisible = ref(false);
const createdClient = reactive<CreatedClient>({
  client_id: "",
  client_secret: "",
});

const rules = computed(() => ({
  name: [{ required: true, message: t("login.required"), trigger: "blur" }],
  redirect_uris_text: [
    {
      required: true,
      message: t("oauth.redirect_uris_required"),
      trigger: "blur",
    },
  ],
}));

const fetchData = async () => {
  loading.value = true;
  try {
    const res: any = await request.get("/oauth-clients", {
      params: queryParams,
    });
    tableData.value = res.list || [];
    total.value = res.total || 0;
  } finally {
    loading.value = false;
  }
};

const handleAdd = () => {
  isEdit.value = false;
  form.name = "";
  form.redirect_uris_text = "";
  form.scopes = "openid profile email";
  form.trusted = false;
  form.active = true;
  dialogVisible.value = true;
};

const handleEdit = (row: OAuthClientItem) => {
  isEdit.value = true;
  editId.value = row.id;
  form.name = row.name;
  form.redirect_uris_text = (row.redirect_uris || []).join("\n");
  form.scopes = row.scopes;
  form.trusted = row.trusted;
  form.active = row.active;
  dialogVisible.value = true;
};

const handleDelete = (row: OAuthClientItem) => {
  ElMessageBox.confirm(
    t("oauth.delete_confirm", { name: row.name }),
    t("common.warning"),
    {
      confirmButtonText: t("common.remove"),
      cancelButtonText: t("common.cancel"),
      type: "warning",
    },
  ).then(async () => {
    try {
      await request.delete(`/oauth-clients/${row.id}`);
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
        const redirectURIs = form.redirect_uris_text
          .split("\n")
          .map((s: string) => s.trim())
          .filter((s: string) => s.length > 0);

        if (redirectURIs.length === 0) {
          ElMessage.error(t("oauth.redirect_uris_required"));
          return;
        }

        const payload = {
          name: form.name,
          redirect_uris: redirectURIs,
          scopes: form.scopes,
          trusted: form.trusted,
          active: form.active,
        };

        if (isEdit.value) {
          await request.put(`/oauth-clients/${editId.value}`, payload);
          dialogVisible.value = false;
          ElMessage.success(t("common.updated"));
        } else {
          const res: any = await request.post("/oauth-clients", payload);
          dialogVisible.value = false;
          createdClient.client_id = res.client_id;
          createdClient.client_secret = res.client_secret;
          secretDialogVisible.value = true;
        }
        fetchData();
      } finally {
        formLoading.value = false;
      }
    }
  });
};

const copyText = async (text: string) => {
  try {
    await navigator.clipboard.writeText(text);
    ElMessage.success(t("oauth.copied"));
  } catch {
    const textarea = document.createElement("textarea");
    textarea.value = text;
    textarea.style.position = "fixed";
    textarea.style.opacity = "0";
    document.body.appendChild(textarea);
    textarea.select();
    document.execCommand("copy");
    document.body.removeChild(textarea);
    ElMessage.success(t("oauth.copied"));
  }
};

const closeSecretDialog = () => {
  secretDialogVisible.value = false;
  createdClient.client_id = "";
  createdClient.client_secret = "";
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
.oauth-client-view {
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

.mono-text {
  font-family: "Courier New", monospace;
  font-size: 12px;
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

/* Secret result dialog styles */
.token-result {
  font-family: "Courier New", monospace;
}

.token-warning {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 16px;
  background: rgba(255, 165, 0, 0.15);
  border: 1px solid rgba(255, 165, 0, 0.4);
  border-radius: 4px;
  color: #ffa500;
  font-size: 13px;
  margin-bottom: 20px;
}

.warning-icon {
  font-size: 18px;
  flex-shrink: 0;
}

.info-row {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 16px;
  background: rgba(0, 40, 0, 0.6);
  border: 1px solid #005000;
  border-radius: 4px;
  margin-bottom: 8px;
}

.info-label {
  color: #8a8;
  white-space: nowrap;
  min-width: 100px;
}

.info-value {
  flex: 1;
  color: #0f0;
  font-size: 13px;
  word-break: break-all;
  line-height: 1.5;
  text-shadow: 0 0 5px rgba(0, 255, 0, 0.2);
}

.info-value.highlight {
  color: #ff0;
  text-shadow: 0 0 5px rgba(255, 255, 0, 0.3);
}

.copy-btn {
  flex-shrink: 0;
}
</style>
