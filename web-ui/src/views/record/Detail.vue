<template>
  <div class="record-detail-container">
    <el-page-header
      @back="goBack"
      title="返回"
      :content="`执行记录详情 #${recordId}`"
    />

    <div class="main-content" v-loading="loading">
      <el-card v-if="record" shadow="never">
        <el-descriptions title="基本信息" :column="2" border>
          <el-descriptions-item label="任务名称">{{
            record.taskName
          }}</el-descriptions-item>
          <el-descriptions-item label="所属部门">{{
            record.departmentName
          }}</el-descriptions-item>
          <el-descriptions-item label="执行状态">
            <el-tag :type="record.status === 'success' ? 'success' : 'danger'">
              {{ record.status === "success" ? "成功" : "失败" }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="执行耗时"
            >{{ record.duration }} ms</el-descriptions-item
          >
          <el-descriptions-item label="开始时间">{{
            record.startTime
          }}</el-descriptions-item>
          <el-descriptions-item label="结束时间">{{
            record.endTime
          }}</el-descriptions-item>
        </el-descriptions>
      </el-card>

      <el-card v-if="record" shadow="never" class="mt-20">
        <template #header>
          <div class="card-header">
            <span>请求信息</span>
          </div>
        </template>
        <div class="code-container">
          <pre class="code-block">{{ formatJson(record.request) }}</pre>
        </div>
      </el-card>

      <el-card v-if="record" shadow="never" class="mt-20">
        <template #header>
          <div class="card-header">
            <span>{{
              record.status === "success" ? "响应信息" : "错误信息"
            }}</span>
          </div>
        </template>
        <div class="code-container">
          <pre
            :class="[
              'code-block',
              record.status === 'success' ? 'success' : 'error',
            ]"
            >{{
              record.status === "success"
                ? formatJson(record.response)
                : record.error
            }}</pre
          >
        </div>
      </el-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from "vue";
import { useRouter, useRoute } from "vue-router";
import { ElMessage } from "element-plus";
import { getRecordById } from "@/api/record";
import type { Record } from "@/api/record";

const router = useRouter();
const route = useRoute();
const loading = ref(false);
const record = ref<Record | null>(null);
const recordId = ref(parseInt(route.params.id as string));

// Go back to record list
const goBack = () => {
  router.back();
};

// Format JSON for display
const formatJson = (jsonStr: string) => {
  try {
    const obj = JSON.parse(jsonStr);
    return JSON.stringify(obj, null, 2);
  } catch (error) {
    return jsonStr;
  }
};

// Fetch record details
const fetchRecordDetail = async () => {
  if (!recordId.value) {
    ElMessage.error("记录ID无效");
    return;
  }

  loading.value = true;
  try {
    const data = await getRecordById(recordId.value);
    record.value = data;
  } catch (error) {
    console.error("Failed to fetch record details:", error);
    ElMessage.error("获取执行记录详情失败");
  } finally {
    loading.value = false;
  }
};

onMounted(() => {
  fetchRecordDetail();
});
</script>

<style lang="scss" scoped>
.record-detail-container {
  .main-content {
    margin-top: 20px;
  }

  .code-container {
    overflow-x: auto;

    .code-block {
      padding: 15px;
      border-radius: 4px;
      background-color: #f5f5f5;
      font-family: "Courier New", Courier, monospace;
      white-space: pre-wrap;
      word-break: break-all;

      &.success {
        background-color: #f0f9eb;
        border: 1px solid #e1f3d8;
      }

      &.error {
        background-color: #fef0f0;
        border: 1px solid #fbc4c4;
      }
    }
  }
}
</style>
