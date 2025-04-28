<template>
  <div class="task-edit-container">
    <el-page-header @back="goBack" :title="isEdit ? '编辑任务' : '创建任务'" />

    <el-card class="mt-20">
      <el-form
        ref="formRef"
        :model="taskForm"
        :rules="rules"
        label-width="120px"
        status-icon
      >
        <!-- Basic Info -->
        <el-divider content-position="left">基本信息</el-divider>
        <el-form-item label="任务名称" prop="name">
          <el-input v-model="taskForm.name" placeholder="请输入任务名称" />
        </el-form-item>
        <el-form-item label="任务描述" prop="description">
          <el-input
            v-model="taskForm.description"
            type="textarea"
            :rows="3"
            placeholder="请输入任务描述"
          />
        </el-form-item>
        <el-form-item label="任务类型" prop="type">
          <el-radio-group v-model="taskForm.type">
            <el-radio label="http">HTTP</el-radio>
            <el-radio label="grpc">gRPC</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="执行计划" prop="cronExpression">
          <el-input
            v-model="taskForm.cronExpression"
            placeholder="Cron表达式，如：0 0 * * *"
          />
        </el-form-item>
        <el-form-item label="所属部门" prop="departmentId">
          <el-select v-model="taskForm.departmentId" placeholder="请选择部门">
            <el-option
              v-for="item in departments"
              :key="item.id"
              :label="item.name"
              :value="item.id"
            />
          </el-select>
        </el-form-item>

        <!-- HTTP Task Config -->
        <template v-if="taskForm.type === 'http'">
          <el-divider content-position="left">HTTP请求配置</el-divider>
          <el-form-item label="请求URL" prop="config.url">
            <el-input
              v-model="httpConfig.url"
              placeholder="请输入完整URL，包括http(s)://"
            />
          </el-form-item>
          <el-form-item label="请求方法" prop="config.method">
            <el-select v-model="httpConfig.method" placeholder="请选择HTTP方法">
              <el-option label="GET" value="GET" />
              <el-option label="POST" value="POST" />
              <el-option label="PUT" value="PUT" />
              <el-option label="DELETE" value="DELETE" />
            </el-select>
          </el-form-item>
          <el-form-item label="请求超时" prop="config.timeout">
            <el-input-number
              v-model="httpConfig.timeout"
              :min="1"
              :max="60"
              :step="1"
              controls-position="right"
            />
            <span class="ml-10">秒</span>
          </el-form-item>
          <el-form-item label="请求头" prop="config.headers">
            <el-button type="primary" link @click="addHeader"
              >添加请求头</el-button
            >
            <div
              v-for="(_, index) in httpHeaders"
              :key="index"
              class="header-item"
            >
              <el-input
                v-model="httpHeaders[index].key"
                placeholder="Key"
                class="header-key"
              />
              <el-input
                v-model="httpHeaders[index].value"
                placeholder="Value"
                class="header-value"
              />
              <el-button type="danger" link @click="removeHeader(index)"
                >删除</el-button
              >
            </div>
          </el-form-item>
          <el-form-item label="请求体" prop="config.body">
            <el-input
              v-model="httpConfig.body"
              type="textarea"
              :rows="5"
              placeholder="请求体内容，支持JSON格式"
            />
          </el-form-item>
          <el-form-item label="成功状态码" prop="config.successCodes">
            <el-select
              v-model="httpConfig.successCodes"
              multiple
              placeholder="请选择成功的HTTP状态码"
            >
              <el-option label="200 - OK" :value="200" />
              <el-option label="201 - Created" :value="201" />
              <el-option label="202 - Accepted" :value="202" />
              <el-option label="204 - No Content" :value="204" />
            </el-select>
          </el-form-item>
        </template>

        <!-- gRPC Task Config -->
        <template v-else-if="taskForm.type === 'grpc'">
          <el-divider content-position="left">gRPC请求配置</el-divider>
          <el-form-item label="Host" prop="config.host">
            <el-input
              v-model="grpcConfig.host"
              placeholder="gRPC服务的主机地址"
            />
          </el-form-item>
          <el-form-item label="Port" prop="config.port">
            <el-input-number
              v-model="grpcConfig.port"
              :min="1"
              :max="65535"
              :step="1"
              controls-position="right"
            />
          </el-form-item>
          <el-form-item label="Service" prop="config.service">
            <el-input v-model="grpcConfig.service" placeholder="gRPC服务名称" />
          </el-form-item>
          <el-form-item label="Method" prop="config.method">
            <el-input v-model="grpcConfig.method" placeholder="gRPC方法名称" />
          </el-form-item>
          <el-form-item label="Request" prop="config.request">
            <el-input
              v-model="grpcConfig.request"
              type="textarea"
              :rows="5"
              placeholder="gRPC请求内容，支持JSON格式"
            />
          </el-form-item>
          <el-form-item label="请求超时" prop="config.timeout">
            <el-input-number
              v-model="grpcConfig.timeout"
              :min="1"
              :max="60"
              :step="1"
              controls-position="right"
            />
            <span class="ml-10">秒</span>
          </el-form-item>
        </template>

        <!-- Submit Buttons -->
        <el-form-item>
          <el-button type="primary" @click="handleSubmit" :loading="loading"
            >保存</el-button
          >
          <el-button @click="goBack">取消</el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from "vue";
import { useRouter, useRoute } from "vue-router";
import { ElMessage, type FormInstance, ElMessageBox } from "element-plus";
import { createTask, updateTask, getTaskById } from "@/api/task";
import { getDepartmentList } from "@/api/department";
import type { Task, HttpTaskConfig, GrpcTaskConfig } from "@/api/task";
import type { Department } from "@/api/department";

// Router and route
const router = useRouter();
const route = useRoute();
const isEdit = computed(() => route.name === "TaskEdit");
const taskId = computed(() =>
  isEdit.value ? parseInt(route.params.id as string) : 0
);

// Form refs and loading state
const formRef = ref<FormInstance>();
const loading = ref(false);
const departments = ref<Department[]>([]);

// HTTP headers array for UI
const httpHeaders = ref<{ key: string; value: string }[]>([]);

// Task form data
const taskForm = reactive<Partial<Task>>({
  name: "",
  description: "",
  type: "http",
  cronExpression: "0 0 * * *",
  departmentId: undefined,
  status: "active",
  config: {} as any,
});

// HTTP config
const httpConfig = reactive<HttpTaskConfig>({
  url: "",
  method: "GET",
  headers: {},
  body: "",
  timeout: 30,
  successCodes: [200],
});

// gRPC config
const grpcConfig = reactive<GrpcTaskConfig>({
  host: "",
  port: 50051,
  service: "",
  method: "",
  request: "",
  timeout: 30,
});

// Form validation rules
const rules = {
  name: [
    { required: true, message: "请输入任务名称", trigger: "blur" },
    { min: 2, max: 50, message: "长度在 2 到 50 个字符", trigger: "blur" },
  ],
  type: [{ required: true, message: "请选择任务类型", trigger: "change" }],
  cronExpression: [
    { required: true, message: "请输入Cron表达式", trigger: "blur" },
  ],
  departmentId: [
    { required: true, message: "请选择所属部门", trigger: "change" },
  ],
};

// Go back to task list
const goBack = () => {
  router.push("/task/list");
};

// Add HTTP header
const addHeader = () => {
  httpHeaders.value.push({ key: "", value: "" });
};

// Remove HTTP header
const removeHeader = (index: number) => {
  httpHeaders.value.splice(index, 1);
};

// Convert headers array to object
const convertHeadersToObject = () => {
  const headers: Record<string, string> = {};
  httpHeaders.value.forEach((item) => {
    if (item.key && item.value) {
      headers[item.key] = item.value;
    }
  });
  return headers;
};

// Convert headers object to array
const convertHeadersToArray = (headers: Record<string, string>) => {
  return Object.entries(headers).map(([key, value]) => ({
    key,
    value,
  }));
};

// Submit form
const handleSubmit = async () => {
  if (!formRef.value) return;

  await formRef.value.validate(async (valid) => {
    if (valid) {
      // Prepare config based on task type
      if (taskForm.type === "http") {
        httpConfig.headers = convertHeadersToObject();
        taskForm.config = httpConfig;
      } else if (taskForm.type === "grpc") {
        taskForm.config = grpcConfig;
      }

      loading.value = true;
      try {
        if (isEdit.value) {
          await updateTask(taskId.value, taskForm);
          ElMessage.success("任务更新成功");
        } else {
          await createTask(taskForm as any);
          ElMessage.success("任务创建成功");
        }
        goBack();
      } catch (error: any) {
        ElMessage.error(
          error.message || (isEdit.value ? "更新任务失败" : "创建任务失败")
        );
      } finally {
        loading.value = false;
      }
    }
  });
};

// Fetch task details if in edit mode
const fetchTaskDetails = async () => {
  if (!isEdit.value) return;

  loading.value = true;
  try {
    const task = await getTaskById(taskId.value);

    // Set basic task info
    taskForm.name = task.name;
    taskForm.description = task.description;
    taskForm.type = task.type;
    taskForm.cronExpression = task.cronExpression;
    taskForm.departmentId = task.departmentId;
    taskForm.status = task.status;

    // Set config based on task type
    if (task.type === "http") {
      const config = task.config as HttpTaskConfig;
      httpConfig.url = config.url;
      httpConfig.method = config.method;
      httpConfig.body = config.body;
      httpConfig.timeout = config.timeout;
      httpConfig.successCodes = config.successCodes;

      // Convert headers object to array for UI
      httpHeaders.value = convertHeadersToArray(config.headers);
    } else if (task.type === "grpc") {
      const config = task.config as GrpcTaskConfig;
      grpcConfig.host = config.host;
      grpcConfig.port = config.port;
      grpcConfig.service = config.service;
      grpcConfig.method = config.method;
      grpcConfig.request = config.request;
      grpcConfig.timeout = config.timeout;
    }
  } catch (error) {
    console.error("Failed to fetch task details:", error);
    ElMessage.error("获取任务详情失败");
    goBack();
  } finally {
    loading.value = false;
  }
};

// Fetch departments for selection
const fetchDepartments = async () => {
  try {
    departments.value = await getDepartmentList();
  } catch (error) {
    console.error("Failed to fetch departments:", error);
    ElMessage.error("获取部门列表失败");
  }
};

onMounted(() => {
  fetchDepartments();
  fetchTaskDetails();
});
</script>

<style lang="scss" scoped>
.task-edit-container {
  .header-item {
    display: flex;
    align-items: center;
    margin-bottom: 10px;

    .header-key {
      width: 200px;
      margin-right: 10px;
    }

    .header-value {
      flex: 1;
      margin-right: 10px;
    }
  }
}
</style>
