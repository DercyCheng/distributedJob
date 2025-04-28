<template>
  <div class="record-list-container">
    <!-- Search and Filter Form -->
    <el-card shadow="never" class="filter-container">
      <el-form :model="searchForm" inline>
        <el-form-item label="任务名称">
          <el-select
            v-model="searchForm.taskId"
            placeholder="请选择任务"
            clearable
            filterable
            :loading="tasksLoading"
          >
            <el-option
              v-for="item in taskOptions"
              :key="item.id"
              :label="item.name"
              :value="item.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="执行状态">
          <el-select
            v-model="searchForm.status"
            placeholder="请选择状态"
            clearable
          >
            <el-option label="成功" value="success" />
            <el-option label="失败" value="fail" />
          </el-select>
        </el-form-item>
        <el-form-item label="所属部门">
          <el-select
            v-model="searchForm.departmentId"
            placeholder="请选择部门"
            clearable
          >
            <el-option
              v-for="item in departments"
              :key="item.id"
              :label="item.name"
              :value="item.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="执行时间">
          <el-date-picker
            v-model="timeRange"
            type="datetimerange"
            range-separator="至"
            start-placeholder="开始时间"
            end-placeholder="结束时间"
            format="YYYY-MM-DD HH:mm:ss"
            value-format="YYYY-MM-DD HH:mm:ss"
          />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">查询</el-button>
          <el-button @click="resetSearch">重置</el-button>
          <el-button type="success" @click="handleExport">导出</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- Records Table -->
    <el-card shadow="never" class="table-container">
      <template #header>
        <div class="card-header">
          <span>执行记录列表</span>
        </div>
      </template>

      <el-table
        v-loading="loading"
        :data="recordList"
        border
        style="width: 100%"
      >
        <el-table-column
          prop="taskName"
          label="任务名称"
          min-width="150"
          show-overflow-tooltip
        />
        <el-table-column
          prop="departmentName"
          label="所属部门"
          width="120"
          show-overflow-tooltip
        />
        <el-table-column
          prop="startTime"
          label="开始时间"
          width="180"
          show-overflow-tooltip
        />
        <el-table-column
          prop="endTime"
          label="结束时间"
          width="180"
          show-overflow-tooltip
        />
        <el-table-column prop="duration" label="耗时(ms)" width="100" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="scope">
            <el-tag
              :type="scope.row.status === 'success' ? 'success' : 'danger'"
            >
              {{ scope.row.status === "success" ? "成功" : "失败" }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="scope">
            <el-button type="primary" link @click="handleViewDetail(scope.row)"
              >查看详情</el-button
            >
          </template>
        </el-table-column>
      </el-table>

      <!-- Pagination -->
      <div class="pagination-container">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.pageSize"
          :total="pagination.total"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleSizeChange"
          @current-change="handleCurrentChange"
        />
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, watch } from "vue";
import { useRouter, useRoute } from "vue-router";
import { ElMessage } from "element-plus";
import { getRecordList, exportRecords } from "@/api/record";
import { getTaskList } from "@/api/task";
import { getDepartmentList } from "@/api/department";
import type { Record } from "@/api/record";
import type { Task } from "@/api/task";
import type { Department } from "@/api/department";

const router = useRouter();
const route = useRoute();
const loading = ref(false);
const tasksLoading = ref(false);
const recordList = ref<Record[]>([]);
const taskOptions = ref<Task[]>([]);
const departments = ref<Department[]>([]);
const timeRange = ref<[string, string] | null>(null);

// Search form
const searchForm = reactive({
  taskId: undefined as number | undefined,
  status: "",
  departmentId: undefined as number | undefined,
});

// Pagination
const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0,
});

// Fetch record list
const fetchRecordList = async () => {
  loading.value = true;
  try {
    const startTime = timeRange.value ? timeRange.value[0] : undefined;
    const endTime = timeRange.value ? timeRange.value[1] : undefined;

    const res = await getRecordList({
      page: pagination.page,
      pageSize: pagination.pageSize,
      taskId: searchForm.taskId,
      status: searchForm.status || undefined,
      departmentId: searchForm.departmentId,
      startTime,
      endTime,
    });
    recordList.value = res.list;
    pagination.total = res.total;
  } catch (error) {
    console.error("Failed to fetch record list:", error);
    ElMessage.error("获取执行记录列表失败");
  } finally {
    loading.value = false;
  }
};

// Fetch task options for filter
const fetchTaskOptions = async () => {
  tasksLoading.value = true;
  try {
    const res = await getTaskList({
      page: 1,
      pageSize: 1000, // Get all tasks for dropdown
    });
    taskOptions.value = res.list;
  } catch (error) {
    console.error("Failed to fetch tasks:", error);
  } finally {
    tasksLoading.value = false;
  }
};

// Fetch departments for filter
const fetchDepartments = async () => {
  try {
    const res = await getDepartmentList();
    departments.value = res;
  } catch (error) {
    console.error("Failed to fetch departments:", error);
  }
};

// Handle search
const handleSearch = () => {
  pagination.page = 1;
  fetchRecordList();
};

// Reset search form
const resetSearch = () => {
  searchForm.taskId = undefined;
  searchForm.status = "";
  searchForm.departmentId = undefined;
  timeRange.value = null;
  handleSearch();
};

// Handle pagination size change
const handleSizeChange = (val: number) => {
  pagination.pageSize = val;
  fetchRecordList();
};

// Handle pagination page change
const handleCurrentChange = (val: number) => {
  pagination.page = val;
  fetchRecordList();
};

// Handle view record detail
const handleViewDetail = (row: Record) => {
  router.push(`/record/detail/${row.id}`);
};

// Handle export records
const handleExport = async () => {
  try {
    const startTime = timeRange.value ? timeRange.value[0] : undefined;
    const endTime = timeRange.value ? timeRange.value[1] : undefined;

    const blob = await exportRecords({
      page: 1,
      pageSize: 10000, // Export more records
      taskId: searchForm.taskId,
      status: searchForm.status || undefined,
      departmentId: searchForm.departmentId,
      startTime,
      endTime,
    });

    // Create download link
    const url = window.URL.createObjectURL(blob);
    const link = document.createElement("a");
    link.href = url;
    link.download = `执行记录_${new Date().toISOString().slice(0, 10)}.xlsx`;
    link.click();
    window.URL.revokeObjectURL(url);

    ElMessage.success("导出成功");
  } catch (error) {
    console.error("Failed to export records:", error);
    ElMessage.error("导出执行记录失败");
  }
};

// Check if there are query parameters and apply them
const initFromQuery = () => {
  if (route.query.taskId) {
    searchForm.taskId = Number(route.query.taskId);
  }
};

onMounted(() => {
  initFromQuery();
  fetchTaskOptions();
  fetchDepartments();
  fetchRecordList();
});

// Watch for route query changes
watch(() => route.query, initFromQuery, { immediate: true });
</script>

<style lang="scss" scoped>
.record-list-container {
  .filter-container {
    margin-bottom: 20px;
  }

  .table-container {
    .card-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
    }
  }

  .pagination-container {
    margin-top: 20px;
    display: flex;
    justify-content: flex-end;
  }
}
</style>
