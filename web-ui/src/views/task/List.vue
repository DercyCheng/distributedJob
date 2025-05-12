<template>
  <div class="task-list-container">
    <!-- Search and Filter Form -->
    <el-card shadow="never" class="filter-container">
      <el-form :model="searchForm" inline>
        <el-form-item label="任务名称">
          <el-input
            v-model="searchForm.name"
            placeholder="请输入任务名称"
            clearable
            @keyup.enter="handleSearch"
          />
        </el-form-item>
        <el-form-item label="任务类型">
          <el-select
            v-model="searchForm.type"
            placeholder="请选择任务类型"
            clearable
          >
            <el-option label="HTTP" value="http" />
            <el-option label="gRPC" value="grpc" />
          </el-select>
        </el-form-item>
        <el-form-item label="任务状态">
          <el-select
            v-model="searchForm.status"
            placeholder="请选择任务状态"
            clearable
          >
            <el-option label="运行中" value="active" />
            <el-option label="已暂停" value="paused" />
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
        <el-form-item>
          <el-button type="primary" @click="handleSearch">查询</el-button>
          <el-button @click="resetSearch">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- Task Table -->
    <el-card shadow="never" class="table-container">
      <template #header>
        <div class="card-header">
          <span>任务列表</span>
          <div>
            <el-button type="primary" @click="handleCreate">创建任务</el-button>
          </div>
        </div>
      </template>

      <el-table v-loading="loading" :data="taskList" border style="width: 100%">
        <el-table-column
          prop="name"
          label="任务名称"
          min-width="150"
          show-overflow-tooltip
        />
        <el-table-column prop="type" label="任务类型" width="100">
          <template #default="scope">
            <el-tag :type="scope.row.type === 'http' ? 'primary' : 'success'">
              {{ scope.row.type === "http" ? "HTTP" : "gRPC" }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column
          prop="cronExpression"
          label="执行计划"
          width="150"
          show-overflow-tooltip
        />
        <el-table-column
          prop="departmentName"
          label="所属部门"
          width="120"
          show-overflow-tooltip
        />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="scope">
            <el-switch
              v-model="scope.row.status"
              :active-value="'active'"
              :inactive-value="'paused'"
              @change="handleStatusChange(scope.row)"
            />
          </template>
        </el-table-column>
        <el-table-column
          prop="creatorName"
          label="创建人"
          width="120"
          show-overflow-tooltip
        />
        <el-table-column
          prop="createdAt"
          label="创建时间"
          width="180"
          show-overflow-tooltip
        />
        <el-table-column label="操作" width="220" fixed="right">
          <template #default="scope">
            <el-button type="primary" link @click="handleEdit(scope.row)"
              >编辑</el-button
            >
            <el-button type="primary" link @click="handleViewRecords(scope.row)"
              >执行记录</el-button
            >
            <el-button type="success" link @click="handleExecute(scope.row)"
              >立即执行</el-button
            >
            <el-button type="danger" link @click="handleDelete(scope.row)"
              >删除</el-button
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
import { ref, reactive, onMounted } from "vue";
import { useRouter } from "vue-router";
import { ElMessage, ElMessageBox } from "element-plus";
import {
  getTaskList,
  deleteTask,
  pauseTask,
  resumeTask,
  executeTask,
} from "@/api/task";
import { getDepartmentList } from "@/api/department";
import type { Task } from "@/api/task";
import type { Department } from "@/api/department";

const router = useRouter();
const loading = ref(false);
const taskList = ref<Task[]>([]);
const departments = ref<Department[]>([]);

// Search form
const searchForm = reactive({
  name: "",
  type: "",
  status: "",
  departmentId: undefined,
});

// Pagination
const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0,
});

// Fetch task list
const fetchTaskList = async () => {
  loading.value = true;
  try {
    const res = await getTaskList({
      page: pagination.page,
      pageSize: pagination.pageSize,
      name: searchForm.name || undefined,
      type: searchForm.type || undefined,
      status: searchForm.status || undefined,
      departmentId: searchForm.departmentId,
    });
    taskList.value = res.list;
    pagination.total = res.total;
  } catch (error) {
    console.error("Failed to fetch task list:", error);
    ElMessage.error("获取任务列表失败");
  } finally {
    loading.value = false;
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
  fetchTaskList();
};

// Reset search form
const resetSearch = () => {
  searchForm.name = "";
  searchForm.type = "";
  searchForm.status = "";
  searchForm.departmentId = undefined;
  handleSearch();
};

// Handle pagination size change
const handleSizeChange = (val: number) => {
  pagination.pageSize = val;
  fetchTaskList();
};

// Handle pagination page change
const handleCurrentChange = (val: number) => {
  pagination.page = val;
  fetchTaskList();
};

// Handle task status change
const handleStatusChange = async (row: Task) => {
  // Store the original status to restore in case of failure
  const originalStatus = row.status;
  
  // Set the new status optimistically (for better UX)
  const newStatus = originalStatus === "active" ? "paused" : "active";
  row.status = newStatus;
  
  try {
    if (newStatus === "active") {
      await resumeTask(row.id);
      ElMessage.success("任务已启用");
    } else {
      await pauseTask(row.id);
      ElMessage.success("任务已暂停");
    }
  } catch (error) {
    console.error("Failed to change task status:", error);
    ElMessage.error("更改任务状态失败");
    // Revert UI status on error
    row.status = originalStatus;
    
    // Refresh the task list to ensure UI is in sync with backend
    fetchTaskList();
  }
};

// Handle create task
const handleCreate = () => {
  router.push("/task/create");
};

// Handle edit task
const handleEdit = (row: Task) => {
  router.push(`/task/edit/${row.id}`);
};

// Handle view task records
const handleViewRecords = (row: Task) => {
  router.push({
    path: "/record/list",
    query: { taskId: row.id.toString() },
  });
};

// Handle delete task
const handleDelete = (row: Task) => {
  ElMessageBox.confirm(
    `确定要删除任务"${row.name}"吗？删除后无法恢复。`,
    "删除确认",
    {
      confirmButtonText: "确定",
      cancelButtonText: "取消",
      type: "warning",
    }
  ).then(async () => {
    try {
      await deleteTask(row.id);
      ElMessage.success("删除成功");
      fetchTaskList();
    } catch (error) {
      console.error("Failed to delete task:", error);
      ElMessage.error("删除任务失败");
    }
  });
};

// Handle execute task immediately
const handleExecute = async (row: Task) => {
  try {
    await executeTask(row.id);
    ElMessage.success("已触发任务执行");
  } catch (error) {
    console.error("Failed to execute task:", error);
    ElMessage.error("触发任务执行失败");
  }
};

onMounted(() => {
  fetchTaskList();
  fetchDepartments();
});
</script>

<style lang="scss" scoped>
.task-list-container {
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
