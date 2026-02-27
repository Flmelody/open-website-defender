<template>
  <div class="user-view">
    <div class="terminal-card glass-panel">
      <div class="card-header no-select">
        <div class="header-left">
          <span class="prefix">root@system:~/users$</span>
          <span class="command blink-cursor">./list_users.sh</span>
        </div>
        <div class="header-right">
          <el-button type="primary" size="small" @click="handleAdd">{{
            t("user.new_user")
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
          <el-table-column prop="id" label="ID" width="80">
            <template #default="scope">
              <span class="dim-text">#{{ scope.row.id }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="username" :label="t('user.username')">
            <template #default="scope">
              <span class="bright-text">{{ scope.row.username }}</span>
              <el-tag
                v-if="scope.row.is_admin"
                size="small"
                type="success"
                class="admin-tag"
                >ADMIN</el-tag
              >
              <el-tag
                v-if="scope.row.totp_enabled"
                size="small"
                type="warning"
                class="admin-tag"
                >2FA</el-tag
              >
            </template>
          </el-table-column>
          <el-table-column
            prop="email"
            :label="t('user.email')"
            show-overflow-tooltip
          >
            <template #default="scope">
              <span v-if="scope.row.email" class="dim-text">{{
                scope.row.email
              }}</span>
              <span v-else class="null-value">{{ t("user.undefined") }}</span>
            </template>
          </el-table-column>
          <el-table-column
            prop="remark"
            :label="t('user.remark')"
            show-overflow-tooltip
          >
            <template #default="scope">
              <span v-if="scope.row.remark" class="dim-text">{{
                scope.row.remark
              }}</span>
              <span v-else class="null-value">{{ t("user.undefined") }}</span>
            </template>
          </el-table-column>
          <el-table-column
            prop="git_token"
            :label="t('user.git_token')"
            show-overflow-tooltip
          >
            <template #default="scope">
              <span v-if="scope.row.git_token" class="token-mask"
                >****************</span
              >
              <span v-else class="null-value">{{ t("user.undefined") }}</span>
            </template>
          </el-table-column>
          <el-table-column
            prop="scopes"
            :label="t('user.authorized_domains')"
            show-overflow-tooltip
          >
            <template #default="scope">
              <span v-if="scope.row.scopes" class="bright-text">{{
                scope.row.scopes
              }}</span>
              <span v-else class="null-value">{{
                t("user.unrestricted")
              }}</span>
            </template>
          </el-table-column>
          <el-table-column
            :label="t('user.enabled')"
            width="100"
            align="center"
          >
            <template #default="scope">
              <el-switch
                v-model="scope.row.enabled"
                size="small"
                :disabled="scope.row.is_admin"
                @change="(val: boolean) => handleToggleEnabled(scope.row, val)"
              />
            </template>
          </el-table-column>
          <el-table-column
            :label="t('common.actions')"
            width="340"
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
                  type="warning"
                  link
                  size="small"
                  @click="handleViewAuthorizations(scope.row)"
                  class="action-link oauth-btn"
                >
                  {{ t("user.oauth_authorizations") }}
                </el-button>
                <el-button
                  v-if="scope.row.totp_enabled"
                  type="danger"
                  link
                  size="small"
                  @click="handleReset2FA(scope.row)"
                  class="action-link delete"
                >
                  {{ t("user.reset_2fa") }}
                </el-button>
                <el-button
                  v-else
                  type="success"
                  link
                  size="small"
                  @click="handleSetup2FA(scope.row)"
                  class="action-link totp-btn"
                >
                  {{ t("user.setup_2fa") }}
                </el-button>
                <el-button
                  type="danger"
                  link
                  size="small"
                  v-if="!scope.row.is_admin"
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
          t("common.total_records", { total: total })
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
        <el-form-item :label="'> ' + t('user.username')" prop="username">
          <el-input v-model="form.username" placeholder="_" />
        </el-form-item>
        <el-form-item :label="'> ' + t('user.password')" prop="password">
          <div class="password-row">
            <el-input
              v-model="form.password"
              type="password"
              show-password
              :placeholder="form.id ? t('user.unchanged') : '_'"
            />
            <el-button
              type="primary"
              size="default"
              @click="generatePassword"
              class="generate-btn"
            >
              {{ t("user.generate") }}
            </el-button>
            <el-button
              size="default"
              @click="copyPassword"
              :disabled="!form.password"
              class="copy-pwd-btn"
            >
              <el-icon><CopyDocument /></el-icon>
            </el-button>
          </div>
        </el-form-item>
        <el-form-item :label="'> ' + t('user.email')" prop="email">
          <el-input
            v-model="form.email"
            :placeholder="t('user.email_placeholder')"
          />
        </el-form-item>
        <el-form-item :label="'> ' + t('user.remark')" prop="remark">
          <el-input
            v-model="form.remark"
            :placeholder="t('user.remark_placeholder')"
          />
        </el-form-item>
        <el-form-item :label="'> ' + t('user.git_token')" prop="git_token">
          <div class="git-token-row">
            <el-input
              v-model="form.git_token"
              type="password"
              show-password
              placeholder="_"
            />
            <el-button
              type="primary"
              size="default"
              @click="generateGitToken"
              class="generate-btn"
            >
              {{ t("user.generate") }}
            </el-button>
          </div>
        </el-form-item>
        <el-form-item prop="is_admin">
          <el-checkbox v-model="form.is_admin" :label="t('user.is_admin')" />
        </el-form-item>
        <el-form-item
          v-show="!form.is_admin"
          :label="'> ' + t('user.authorized_domains')"
          prop="scopes"
        >
          <el-select
            v-model="scopesArray"
            multiple
            filterable
            allow-create
            default-first-option
            :placeholder="t('user.authorized_domains_placeholder')"
            style="width: 100%"
          >
            <el-option
              v-for="d in domainOptions"
              :key="d"
              :label="d"
              :value="d"
            />
          </el-select>
          <div class="scope-hint">{{ t("user.authorized_domains_hint") }}</div>
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
          >
            {{ t("common.confirm") }}
          </el-button>
        </div>
      </template>
    </el-dialog>

    <!-- Git Token Result Dialog -->
    <el-dialog
      v-model="tokenDialogVisible"
      :title="t('user.token_generated').toUpperCase()"
      width="600px"
      :close-on-click-modal="false"
      :close-on-press-escape="false"
    >
      <div class="token-result">
        <div class="token-warning">
          <el-icon class="warning-icon"><WarningFilled /></el-icon>
          <span>{{ t("user.token_warning") }}</span>
        </div>
        <div class="token-display">
          <code class="token-value">{{ generatedToken }}</code>
          <el-button
            type="primary"
            size="small"
            class="copy-btn"
            @click="copyToken"
          >
            <el-icon><CopyDocument /></el-icon>
            {{ t("user.copy") }}
          </el-button>
        </div>
        <div v-if="copiedVisible" class="copied-hint">
          {{ t("user.copied") }}
        </div>
      </div>
      <template #footer>
        <div class="dialog-footer">
          <el-button type="primary" @click="closeTokenDialog">{{
            t("user.understood")
          }}</el-button>
        </div>
      </template>
    </el-dialog>
    <!-- OAuth Authorizations Dialog -->
    <el-dialog
      v-model="authDialogVisible"
      :title="authDialogTitle"
      width="650px"
      destroy-on-close
    >
      <div class="auth-list">
        <div v-if="authLoading" class="auth-loading">
          <el-icon class="is-loading"><Loading /></el-icon>
        </div>
        <div v-else-if="authorizations.length === 0" class="auth-empty">
          <span class="dim-text">{{ t("user.no_authorizations") }}</span>
        </div>
        <div v-else>
          <div
            v-for="auth in authorizations"
            :key="auth.client_id"
            class="auth-item"
          >
            <div class="auth-info">
              <div class="auth-client-name">{{ auth.client_name }}</div>
              <div class="auth-meta">
                <span class="dim-text"
                  >{{ t("user.auth_scope") }}: {{ auth.scope }}</span
                >
                <span class="dim-text">
                  | {{ t("user.auth_since") }}:
                  {{ formatTime(auth.authorized_at) }}</span
                >
              </div>
            </div>
            <el-button
              type="danger"
              size="small"
              @click="handleRevokeAuthorization(auth)"
            >
              {{ t("user.auth_revoke") }}
            </el-button>
          </div>
        </div>
      </div>
    </el-dialog>

    <!-- 2FA Setup Dialog -->
    <el-dialog
      v-model="totpDialogVisible"
      :title="t('user.setup_2fa').toUpperCase()"
      width="500px"
      destroy-on-close
    >
      <div class="totp-setup">
        <div v-if="totpSetupData">
          <p class="totp-instruction">{{ t("user.scan_qr") }}</p>
          <div class="totp-qr-wrapper">
            <img :src="totpSetupData.qr_code" alt="QR Code" class="totp-qr" />
          </div>
          <div class="totp-manual-key">
            <span class="totp-key-label">{{ t("user.manual_key") }}</span>
            <div class="totp-key-row">
              <code class="totp-key-value">{{ totpSetupData.secret }}</code>
              <el-button
                type="primary"
                size="small"
                class="totp-copy-btn"
                @click="copyManualKey"
              >
                <el-icon><CopyDocument /></el-icon>
              </el-button>
            </div>
          </div>
          <div class="totp-verify-section">
            <div class="totp-verify-row">
              <span class="totp-verify-label"
                >{{ t("user.verify_code") }}:</span
              >
              <el-input
                v-model="totpVerifyCode"
                maxlength="6"
                :placeholder="'000000'"
                class="totp-verify-input"
              />
            </div>
          </div>
        </div>
        <div v-else class="totp-loading">
          <el-icon class="is-loading"><Loading /></el-icon>
        </div>
      </div>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="totpDialogVisible = false">{{
            t("common.cancel")
          }}</el-button>
          <el-button
            type="primary"
            :loading="totpConfirmLoading"
            @click="handleConfirm2FA"
            :disabled="totpVerifyCode.length !== 6"
          >
            {{ t("user.enable_2fa") }}
          </el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, reactive, computed, watch } from "vue";
import request from "@/utils/request";
import { ElMessage, ElMessageBox } from "element-plus";
import { CopyDocument, WarningFilled, Loading } from "@element-plus/icons-vue";
import { useI18n } from "vue-i18n";

interface User {
  id: number;
  username: string;
  email?: string;
  remark?: string;
  git_token?: string;
  is_admin?: boolean;
  enabled?: boolean;
  scopes?: string;
  totp_enabled?: boolean;
}

const { t } = useI18n();
const tableData = ref<User[]>([]);
const total = ref(0);
const loading = ref(false);
const queryParams = reactive({
  page: 1,
  size: 10,
});

const dialogVisible = ref(false);
const dialogTitle = ref("");
const formRef = ref();
const formLoading = ref(false);
const isEditMode = ref(false);

const domainOptions = ref<string[]>([]);
const scopesArray = ref<string[]>([]);
const tokenDialogVisible = ref(false);
const generatedToken = ref("");
const copiedVisible = ref(false);
const pendingToken = ref("");

const form = reactive({
  id: 0,
  username: "",
  password: "",
  email: "",
  remark: "",
  git_token: "",
  is_admin: false,
  enabled: true,
  scopes: "",
});

watch(
  () => form.is_admin,
  (val) => {
    if (val) {
      form.scopes = "";
      scopesArray.value = [];
    }
  },
);

watch(scopesArray, (val) => {
  form.scopes = val.join(", ");
});

const domainPattern =
  /^(\*\.)?([a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$/;

const scopesValidator = (
  _rule: any,
  _value: string,
  callback: (err?: Error) => void,
) => {
  for (const item of scopesArray.value) {
    if (item && !domainPattern.test(item)) {
      callback(new Error(t("user.authorized_domains_invalid")));
      return;
    }
  }
  callback();
};

const rules = computed(() => ({
  username: [{ required: true, message: t("login.required"), trigger: "blur" }],
  password: [
    {
      required: !isEditMode.value,
      message: t("login.required"),
      trigger: "blur",
    },
  ],
  scopes: [{ validator: scopesValidator, trigger: ["blur", "change"] }],
}));

const fetchDomainOptions = async () => {
  try {
    const res: any = await request.get("/authorized-domains", {
      params: { all: "true" },
    });
    domainOptions.value = (res || []).map((d: any) => d.name);
  } catch {
    // handled
  }
};

const fetchData = async () => {
  loading.value = true;
  try {
    const res: any = await request.get("/users", { params: queryParams });
    tableData.value = res.list || [];
    total.value = res.total || 0;
  } catch (error) {
    // handled
  } finally {
    loading.value = false;
  }
};

const handleAdd = () => {
  dialogTitle.value = t("user.title_create");
  form.id = 0;
  form.username = "";
  form.password = "";
  form.email = "";
  form.remark = "";
  form.git_token = "";
  form.is_admin = false;
  form.enabled = true;
  form.scopes = "";
  scopesArray.value = [];
  isEditMode.value = false;
  pendingToken.value = "";
  dialogVisible.value = true;
};

const handleEdit = (row: User) => {
  dialogTitle.value = t("user.title_edit");
  form.id = row.id;
  form.username = row.username;
  form.password = "";
  form.email = row.email || "";
  form.remark = row.remark || "";
  form.git_token = "";
  form.is_admin = row.is_admin || false;
  form.enabled = row.enabled !== false;
  form.scopes = row.scopes || "";
  scopesArray.value = row.scopes
    ? row.scopes
        .split(",")
        .map((s: string) => s.trim())
        .filter(Boolean)
    : [];
  isEditMode.value = true;
  pendingToken.value = "";
  dialogVisible.value = true;
};

const handleToggleEnabled = async (row: User, val: boolean) => {
  try {
    await request.put(`/users/${row.id}`, { enabled: val });
    ElMessage.success(t("common.updated"));
  } catch {
    row.enabled = !val;
  }
};

const handleDelete = (row: User) => {
  ElMessageBox.confirm(
    t("user.delete_confirm", { name: row.username }),
    t("common.warning"),
    {
      confirmButtonText: t("common.delete"),
      cancelButtonText: t("common.cancel"),
      type: "warning",
    },
  ).then(async () => {
    try {
      await request.delete(`/users/${row.id}`);
      ElMessage.success(t("common.deleted"));
      fetchData();
    } catch (error) {
      // handled
    }
  });
};

const handleSubmit = async () => {
  if (!formRef.value) return;
  await formRef.value.validate(async (valid: boolean) => {
    if (valid) {
      formLoading.value = true;
      try {
        if (form.id === 0) {
          await request.post("/users", form);
          ElMessage.success(t("common.created"));
        } else {
          await request.put(`/users/${form.id}`, form);
          ElMessage.success(t("common.updated"));
        }
        dialogVisible.value = false;
        fetchData();
        if (pendingToken.value) {
          generatedToken.value = pendingToken.value;
          copiedVisible.value = false;
          tokenDialogVisible.value = true;
          pendingToken.value = "";
        }
      } catch (error) {
        // handled
      } finally {
        formLoading.value = false;
      }
    }
  });
};

const generateRandomHex = (bytes: number): string => {
  const arr = new Uint8Array(bytes);
  crypto.getRandomValues(arr);
  return Array.from(arr, (b) => b.toString(16).padStart(2, "0")).join("");
};

const generatePassword = () => {
  const upper = "ABCDEFGHJKLMNPQRSTUVWXYZ";
  const lower = "abcdefghjkmnpqrstuvwxyz";
  const digits = "23456789";
  const special = "!@#$%^&*_+-=";
  const all = upper + lower + digits + special;

  const pick = (charset: string) =>
    charset[Math.floor(Math.random() * charset.length)];

  // Ensure at least 2 of each category
  const required = [
    pick(upper),
    pick(upper),
    pick(lower),
    pick(lower),
    pick(digits),
    pick(digits),
    pick(special),
    pick(special),
  ];

  // Fill remaining with random from all
  const remaining = Array.from({ length: 8 }, () => pick(all));
  const chars = [...required, ...remaining];

  // Shuffle
  for (let i = chars.length - 1; i > 0; i--) {
    const j = Math.floor(Math.random() * (i + 1));
    [chars[i], chars[j]] = [chars[j], chars[i]];
  }

  form.password = chars.join("");
};

const copyPassword = async () => {
  if (!form.password) return;
  try {
    await navigator.clipboard.writeText(form.password);
    ElMessage.success(t("user.copied"));
  } catch {
    const textarea = document.createElement("textarea");
    textarea.value = form.password;
    textarea.style.position = "fixed";
    textarea.style.opacity = "0";
    document.body.appendChild(textarea);
    textarea.select();
    document.execCommand("copy");
    document.body.removeChild(textarea);
    ElMessage.success(t("user.copied"));
  }
};

const generateGitToken = () => {
  const token = generateRandomHex(32);
  form.git_token = token;
  pendingToken.value = token;
};

const copyToken = async () => {
  try {
    await navigator.clipboard.writeText(generatedToken.value);
    copiedVisible.value = true;
    ElMessage.success(t("user.copied"));
  } catch {
    const textarea = document.createElement("textarea");
    textarea.value = generatedToken.value;
    textarea.style.position = "fixed";
    textarea.style.opacity = "0";
    document.body.appendChild(textarea);
    textarea.select();
    document.execCommand("copy");
    document.body.removeChild(textarea);
    copiedVisible.value = true;
    ElMessage.success(t("user.copied"));
  }
};

const closeTokenDialog = () => {
  tokenDialogVisible.value = false;
  generatedToken.value = "";
};

// OAuth Authorizations
interface OAuthAuthorization {
  client_id: string;
  client_name: string;
  scope: string;
  authorized_at: string;
}

const authDialogVisible = ref(false);
const authDialogTitle = ref("");
const authLoading = ref(false);
const authorizations = ref<OAuthAuthorization[]>([]);
const authUserId = ref(0);

const formatTime = (isoStr: string) => {
  return new Date(isoStr).toLocaleString();
};

const handleViewAuthorizations = async (row: User) => {
  authUserId.value = row.id;
  authDialogTitle.value = `OAUTH > ${row.username}`;
  authDialogVisible.value = true;
  authLoading.value = true;
  try {
    const res: any = await request.get(`/users/${row.id}/oauth-authorizations`);
    authorizations.value = res || [];
  } catch {
    authorizations.value = [];
  } finally {
    authLoading.value = false;
  }
};

const handleRevokeAuthorization = (auth: OAuthAuthorization) => {
  ElMessageBox.confirm(
    t("user.auth_revoke_confirm", { name: auth.client_name }),
    t("common.warning"),
    {
      confirmButtonText: t("common.confirm"),
      cancelButtonText: t("common.cancel"),
      type: "warning",
    },
  ).then(async () => {
    try {
      await request.delete(
        `/users/${authUserId.value}/oauth-authorizations/${auth.client_id}`,
      );
      ElMessage.success(t("user.auth_revoked"));
      // Refresh the list
      const res: any = await request.get(
        `/users/${authUserId.value}/oauth-authorizations`,
      );
      authorizations.value = res || [];
    } catch {
      // handled
    }
  });
};

// 2FA Management
interface TotpSetupData {
  secret: string;
  qr_code: string;
}

const totpDialogVisible = ref(false);
const totpSetupData = ref<TotpSetupData | null>(null);
const totpVerifyCode = ref("");
const totpConfirmLoading = ref(false);
const totpTargetUserId = ref(0);

const handleSetup2FA = async (row: User) => {
  totpTargetUserId.value = row.id;
  totpSetupData.value = null;
  totpVerifyCode.value = "";
  totpDialogVisible.value = true;
  try {
    const res: any = await request.post(`/users/${row.id}/totp/setup`);
    totpSetupData.value = res;
  } catch {
    totpDialogVisible.value = false;
  }
};

const copyManualKey = async () => {
  if (!totpSetupData.value) return;
  try {
    await navigator.clipboard.writeText(totpSetupData.value.secret);
    ElMessage.success(t("user.copied"));
  } catch {
    const textarea = document.createElement("textarea");
    textarea.value = totpSetupData.value.secret;
    textarea.style.position = "fixed";
    textarea.style.opacity = "0";
    document.body.appendChild(textarea);
    textarea.select();
    document.execCommand("copy");
    document.body.removeChild(textarea);
    ElMessage.success(t("user.copied"));
  }
};

const handleConfirm2FA = async () => {
  if (totpVerifyCode.value.length !== 6) return;
  totpConfirmLoading.value = true;
  try {
    await request.post(`/users/${totpTargetUserId.value}/totp/confirm`, {
      code: totpVerifyCode.value,
    });
    ElMessage.success(t("user.two_factor_enabled"));
    totpDialogVisible.value = false;
    fetchData();
  } catch {
    totpVerifyCode.value = "";
  } finally {
    totpConfirmLoading.value = false;
  }
};

const handleReset2FA = (row: User) => {
  ElMessageBox.confirm(
    t("user.reset_2fa_confirm", { name: row.username }),
    t("common.warning"),
    {
      confirmButtonText: t("common.confirm"),
      cancelButtonText: t("common.cancel"),
      type: "warning",
    },
  ).then(async () => {
    try {
      await request.delete(`/users/${row.id}/totp`);
      ElMessage.success(t("user.two_factor_disabled"));
      fetchData();
    } catch {
      // handled
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
  fetchDomainOptions();
});
</script>

<style scoped>
.user-view {
  width: 100%;
}

.glass-panel {
  background: rgba(10, 30, 10, 0.75); /* More visible background */
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

.token-mask {
  color: #00a000;
  letter-spacing: 2px;
}

.null-value {
  color: #006000;
  font-style: italic;
}

.ops-cell {
  display: flex;
  justify-content: flex-end;
  gap: 6px;
}

.action-link {
  font-weight: bold;
  text-decoration: underline;
}

.oauth-btn {
  text-decoration: none !important;
  border: 1px solid rgba(230, 162, 60, 0.5) !important;
  border-radius: 3px !important;
  padding: 2px 8px !important;
}

.totp-btn {
  text-decoration: none !important;
  border: 1px solid rgba(64, 158, 255, 0.5) !important;
  border-radius: 3px !important;
  padding: 2px 8px !important;
  color: rgba(64, 158, 255, 0.9) !important;
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

.admin-tag {
  margin-left: 10px;
  font-family: "Courier New", monospace;
  font-weight: bold;
  letter-spacing: 1px;
}

.dialog-footer {
  text-align: right;
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

.password-row,
.git-token-row {
  display: flex;
  gap: 10px;
  width: 100%;
}

.password-row .el-input,
.git-token-row .el-input {
  flex: 1;
}

.generate-btn,
.copy-pwd-btn {
  flex-shrink: 0;
}

.copy-pwd-btn {
  background: transparent !important;
  border: 1px solid #005000 !important;
  color: #0f0 !important;
}

.copy-pwd-btn:hover:not(:disabled) {
  background: rgba(0, 60, 0, 0.4) !important;
  border-color: #0f0 !important;
}

.copy-pwd-btn:disabled {
  opacity: 0.4;
}

/* Token result dialog styles */
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

.token-display {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 14px 16px;
  background: rgba(0, 40, 0, 0.6);
  border: 1px solid #005000;
  border-radius: 4px;
}

.token-value {
  flex: 1;
  color: #0f0;
  font-size: 13px;
  word-break: break-all;
  line-height: 1.5;
  text-shadow: 0 0 5px rgba(0, 255, 0, 0.2);
}

.copy-btn {
  flex-shrink: 0;
}

.copied-hint {
  margin-top: 10px;
  color: #0f0;
  font-size: 12px;
  text-align: right;
}

.scope-hint {
  color: #8a8;
  font-size: 12px;
  margin-top: 4px;
  font-family: "Courier New", monospace;
}

/* OAuth Authorization dialog */
.auth-list {
  font-family: "Courier New", monospace;
}

.auth-loading {
  text-align: center;
  padding: 30px;
  color: #0f0;
}

.auth-empty {
  text-align: center;
  padding: 30px;
}

.auth-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 14px 16px;
  border: 1px solid #005000;
  border-radius: 4px;
  background: rgba(0, 40, 0, 0.4);
  margin-bottom: 10px;
}

.auth-client-name {
  color: #0f0;
  font-weight: bold;
  font-size: 15px;
  margin-bottom: 4px;
  text-shadow: 0 0 5px rgba(0, 255, 0, 0.2);
}

.auth-meta {
  font-size: 12px;
}

/* 2FA Setup dialog */
.totp-setup {
  font-family: "Courier New", monospace;
}

.totp-instruction {
  color: #8a8;
  font-size: 13px;
  margin-bottom: 16px;
}

.totp-qr-wrapper {
  text-align: center;
  margin: 0 auto 16px;
  display: flex;
  justify-content: center;
}

.totp-qr {
  width: 200px;
  height: 200px;
  border-radius: 4px;
}

.totp-manual-key {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  background: rgba(0, 40, 0, 0.4);
  border: 1px solid #005000;
  border-radius: 4px;
  margin-bottom: 16px;
}

.totp-key-label {
  color: #8a8;
  font-size: 12px;
  flex-shrink: 0;
}

.totp-key-row {
  display: flex;
  align-items: center;
  gap: 8px;
  flex: 1;
  min-width: 0;
}

.totp-key-value {
  color: #0f0;
  font-size: 12px;
  word-break: break-all;
  letter-spacing: 1px;
  text-shadow: 0 0 5px rgba(0, 255, 0, 0.2);
  flex: 1;
}

.totp-copy-btn {
  flex-shrink: 0;
  background: transparent !important;
  border: 1px solid #005000 !important;
  color: #0f0 !important;
  padding: 4px 8px !important;
}

.totp-copy-btn:hover {
  background: rgba(0, 60, 0, 0.4) !important;
  border-color: #0f0 !important;
}

.totp-verify-section {
  margin-top: 16px;
}

.totp-verify-label {
  color: #0f0;
  font-size: 13px;
  font-weight: bold;
  flex-shrink: 0;
}

.totp-verify-row {
  display: flex;
  align-items: center;
  gap: 12px;
}

.totp-verify-input {
  flex: 1;
}

:deep(.totp-verify-input .el-input__inner) {
  font-family: "Courier New", monospace;
  font-size: 18px;
  letter-spacing: 6px;
  text-align: center;
}

.totp-loading {
  text-align: center;
  padding: 40px;
  color: #0f0;
}
</style>
