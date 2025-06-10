<template>
  <div class="executions">
    <div class="page-header">
      <h1>执行记录</h1>
    </div>

    <!-- 搜索筛选 -->
    <el-card class="search-card">
      <el-form :model="searchForm" inline>
        <el-form-item label="任务ID">
          <el-input
            v-model="searchForm.job_id"
            placeholder="请输入任务ID"
            clearable
          />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="searchForm.status" placeholder="请选择状态" clearable>
            <el-option label="全部" value="" />
            <el-option label="等待中" value="pending" />
            <el-option label="运行中" value="running" />
            <el-option label="成功" value="success" />
            <el-option label="失败" value="failed" />
            <el-option label="超时" value="timeout" />
            <el-option label="已取消" value="cancelled" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="loadExecutions">搜索</el-button>
          <el-button @click="resetSearch">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 执行记录列表 -->
    <el-card>
      <el-table :data="executions" stripe v-loading="loading">
        <el-table-column prop="id" label="执行ID" width="280" />
        <el-table-column prop="job.name" label="任务名称" />
        <el-table-column prop="worker.name" label="工作节点" />
        <el-table-column prop="status" label="状态" width="120">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.status)">
              {{ getStatusText(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="开始时间" width="180">
          <template #default="{ row }">
            {{ formatTime(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column prop="finished_at" label="结束时间" width="180">
          <template #default="{ row }">
            {{ row.finished_at ? formatTime(row.finished_at) : '-' }}
          </template>
        </el-table-column>
        <el-table-column label="耗时" width="100">
          <template #default="{ row }">
            {{ getDuration(row) }}
          </template>
        </el-table-column>
        <el-table-column prop="exit_code" label="退出码" width="80" />
        <el-table-column label="操作" width="120">
          <template #default="{ row }">
            <el-button size="small" type="primary" @click="viewExecution(row)">
              查看详情
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div class="pagination">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.size"
          :page-sizes="[10, 20, 50, 100]"
          :total="pagination.total"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="loadExecutions"
          @current-change="loadExecutions"
        />
      </div>
    </el-card>

    <!-- 执行详情对话框 -->
    <el-dialog
      v-model="showDetailDialog"
      title="执行详情"
      width="800px"
    >
      <div v-if="currentExecution">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="执行ID">
            {{ currentExecution.id }}
          </el-descriptions-item>
          <el-descriptions-item label="任务名称">
            {{ currentExecution.job?.name }}
          </el-descriptions-item>
          <el-descriptions-item label="工作节点">
            {{ currentExecution.worker?.name }}
          </el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="getStatusType(currentExecution.status)">
              {{ getStatusText(currentExecution.status) }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="开始时间">
            {{ formatTime(currentExecution.created_at) }}
          </el-descriptions-item>
          <el-descriptions-item label="结束时间">
            {{ currentExecution.finished_at ? formatTime(currentExecution.finished_at) : '-' }}
          </el-descriptions-item>
          <el-descriptions-item label="耗时">
            {{ getDuration(currentExecution) }}
          </el-descriptions-item>
          <el-descriptions-item label="退出码">
            {{ currentExecution.exit_code }}
          </el-descriptions-item>
        </el-descriptions>

        <div class="output-section">
          <h3>执行输出</h3>
          <el-input
            v-model="currentExecution.output"
            type="textarea"
            :rows="10"
            readonly
            placeholder="无输出"
          />
        </div>

        <div v-if="currentExecution.error" class="error-section">
          <h3>错误信息</h3>
          <el-input
            v-model="currentExecution.error"
            type="textarea"
            :rows="5"
            readonly
          />
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { getExecutions, getExecution } from '@/api/executions'

const executions = ref([])
const loading = ref(false)
const showDetailDialog = ref(false)
const currentExecution = ref(null)

const searchForm = reactive({
  job_id: '',
  status: ''
})

const pagination = reactive({
  page: 1,
  size: 10,
  total: 0
})

onMounted(() => {
  loadExecutions()
})

const loadExecutions = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.page,
      size: pagination.size,
      job_id: searchForm.job_id,
      status: searchForm.status
    }
    const response = await getExecutions(params)
    executions.value = response.data.executions || []
    pagination.total = response.data.total || 0
  } catch (error) {
    ElMessage.error('加载执行记录失败')
  } finally {
    loading.value = false
  }
}

const resetSearch = () => {
  searchForm.job_id = ''
  searchForm.status = ''
  pagination.page = 1
  loadExecutions()
}

const viewExecution = async (execution) => {
  try {
    const response = await getExecution(execution.id)
    currentExecution.value = response.data
    showDetailDialog.value = true
  } catch (error) {
    ElMessage.error('加载执行详情失败')
  }
}

const getStatusType = (status) => {
  const typeMap = {
    'success': 'success',
    'failed': 'danger',
    'running': 'warning',
    'pending': 'info',
    'timeout': 'danger',
    'cancelled': 'info'
  }
  return typeMap[status] || 'info'
}

const getStatusText = (status) => {
  const textMap = {
    'success': '成功',
    'failed': '失败',
    'running': '运行中',
    'pending': '等待中',
    'timeout': '超时',
    'cancelled': '已取消'
  }
  return textMap[status] || status
}

const formatTime = (time) => {
  return new Date(time).toLocaleString()
}

const getDuration = (execution) => {
  if (!execution.started_at || !execution.finished_at) {
    return '-'
  }
  const start = new Date(execution.started_at)
  const end = new Date(execution.finished_at)
  const duration = Math.floor((end - start) / 1000)
  return `${duration}秒`
}
</script>

<style scoped>
.executions {
  padding: 20px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.search-card {
  margin-bottom: 20px;
}

.pagination {
  margin-top: 20px;
  text-align: right;
}

.output-section,
.error-section {
  margin-top: 20px;
}

.output-section h3,
.error-section h3 {
  margin-bottom: 10px;
  color: #333;
}
</style>
