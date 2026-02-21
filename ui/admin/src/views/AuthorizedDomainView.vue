<template>
  <div class="authorized-domain-view">
    <div class="terminal-card glass-panel">
      <div class="card-header no-select">
        <div class="header-left">
          <span class="prefix">root@system:~/firewall$</span>
          <span class="command blink-cursor">./authorized_domains.sh</span>
        </div>
        <div class="header-right">
          <el-button type="primary" size="small" @click="handleAdd">{{
            t("authorized_domain.new_domain")
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
          <el-table-column prop="name" :label="t('authorized_domain.name')">
            <template #default="scope">
              <span class="bright-text">{{ scope.row.name }}</span>
            </template>
          </el-table-column>
          <el-table-column
            prop="created_at"
            :label="t('common.created_at')"
            width="200"
          >
            <template #default="scope">
              <span class="dim-text">{{
                new Date(scope.row.created_at).toLocaleString()
              }}</span>
            </template>
          </el-table-column>
          <el-table-column
            :label="t('common.actions')"
            width="120"
            align="right"
          >
            <template #default="scope">
              <div class="ops-cell">
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
        <el-form-item :label="'> ' + t('authorized_domain.name')" prop="name">
          <el-input
            v-model="form.name"
            :placeholder="t('authorized_domain.name_placeholder')"
          />
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
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, reactive, computed } from "vue";
import request from "@/utils/request";
import { ElMessage, ElMessageBox } from "element-plus";
import { useI18n } from "vue-i18n";

interface AuthorizedDomain {
  id: number;
  name: string;
  created_at: string;
}

const { t } = useI18n();
const tableData = ref<AuthorizedDomain[]>([]);
const total = ref(0);
const loading = ref(false);
const queryParams = reactive({ page: 1, size: 10 });

const dialogVisible = ref(false);
const dialogTitle = ref("");
const formRef = ref();
const formLoading = ref(false);
const form = reactive({ name: "" });

const domainValidator = (
  _rule: any,
  value: string,
  callback: (err?: Error) => void,
) => {
  if (!value) return callback();
  const pattern =
    /^(\*\.)?([a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$/;
  if (!pattern.test(value)) {
    callback(new Error(t("authorized_domain.name_invalid")));
  } else {
    callback();
  }
};

const rules = computed(() => ({
  name: [
    { required: true, message: t("login.required"), trigger: "blur" },
    { validator: domainValidator, trigger: ["blur", "change"] },
  ],
}));

const fetchData = async () => {
  loading.value = true;
  try {
    const res: any = await request.get("/authorized-domains", {
      params: queryParams,
    });
    tableData.value = res.list || [];
    total.value = res.total || 0;
  } finally {
    loading.value = false;
  }
};

const handleAdd = () => {
  dialogTitle.value = t("authorized_domain.title_create");
  form.name = "";
  dialogVisible.value = true;
};

const handleDelete = (row: AuthorizedDomain) => {
  ElMessageBox.confirm(
    t("authorized_domain.delete_confirm", { name: row.name }),
    t("common.warning"),
    {
      confirmButtonText: t("common.remove"),
      cancelButtonText: t("common.cancel"),
      type: "warning",
    },
  ).then(async () => {
    try {
      await request.delete(`/authorized-domains/${row.id}`);
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
        await request.post("/authorized-domains", form);
        ElMessage.success(t("common.added"));
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
.authorized-domain-view {
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
</style>
