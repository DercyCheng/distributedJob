<template>
  <div class="workers">
    <div class="page-header">
      <h1>工作节点</h1>
    </div>

    <!-- 搜索筛选 -->
    <el-card class="search-card">
      <el-form :model="searchForm" inline>
        <el-form-item label="状态">
          <el-select v-model="searchForm.status" placeholder="请选择状态" clearable>
            <el-option label="全部" value="" />
            <el-option label="在线" value="online" />
            <el-option label="离线" value="offline" />
            <el-option label="忙碌" value="busy" />
            <el-option label="维护中" value="maintenance" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="loadWorkers">搜索</el-button>
          <el-button @click="resetSearch">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 工作节点列表 -->
    <el-card>
      <el-table :data="workers" stripe v-loading="loading">
        <el-table-column prop="name" label="节点名称" />
        <el-table-column prop="ip" label="IP地址" />
        <el-table-column prop="port" label="端口" />
        <el-table-column prop="status" label="状态" width="120">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.status)">
              {{ getStatusText(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="负载" width="200">
          <template #default="{ row }">
            <div class="load-info">
              <el-progress
                :percentage="getLoadPercentage(row)"
                :color="getLoadColor(row)"
                :stroke-width="12"
              />
              <span class="load-text">{{ row.current_load }}/{{ row.capacity }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="last_heartbeat" label="最后心跳" width="180">
          <template #default="{ row }">
            {{ row.last_heartbeat ? formatTime(row.last_heartbeat) : '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="注册时间" width="180">
          <template #default="{ row }">
            {{ formatTime(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="120">
          <template #default="{ row }">
            <el-button size="small" type="primary" @click="viewWorker(row)">
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
          @size-change="loadWorkers"
          @current-change="loadWorkers"
        />
      </div>
    </el-card>

    <!-- 工作节点详情对话框 -->
    <el-dialog
      v-model="showDetailDialog"
      title="工作节点详情"
      width="600px"
    >
      <div v-if="currentWorker">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="节点ID">
            {{ currentWorker.id }}
          </el-descriptions-item>
          <el-descriptions-item label="节点名称">
            {{ currentWorker.name }}
          </el-descriptions-item>
          <el-descriptions-item label="IP地址">
            {{ currentWorker.ip }}
          </el-descriptions-item>
          <el-descriptions-item label="端口">
            {{ currentWorker.port }}
          </el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="getStatusType(currentWorker.status)">
              {{ getStatusText(currentWorker.status) }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="容量">
            {{ currentWorker.capacity }}
          </el-descriptions-item>
          <el-descriptions-item label="当前负载">
            {{ currentWorker.current_load }}
          </el-descriptions-item>
          <el-descriptions-item label="负载率">
            {{ getLoadPercentage(currentWorker) }}%
          </el-descriptions-item>
          <el-descriptions-item label="最后心跳">
            {{ currentWorker.last_heartbeat ? formatTime(currentWorker.last_heartbeat) : '-' }}
          </el-descriptions-item>
          <el-descriptions-item label="注册时间">
            {{ formatTime(currentWorker.created_at) }}
          </el-descriptions-item>
        </el-descriptions>

        <div v-if="currentWorker.metadata" class="metadata-section">
          <h3>节点元数据</h3>
          <el-table :data="getMetadataList(currentWorker.metadata)" size="small">
            <el-table-column prop="key" label="键" />
            <el-table-column prop="value" label="值" />
          </el-table>
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { getWorkers, getWorker } from '@/api/workers'

const workers = ref([])
const loading = ref(false)
const showDetailDialog = ref(false)
const currentWorker = ref(null)

const searchForm = reactive({
  status: ''
})

const pagination = reactive({
  page: 1,
  size: 10,
  total: 0
})

onMounted(() => {
  loadWorkers()
})

const loadWorkers = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.page,
      size: pagination.size,
      status: searchForm.status
    }
    const response = await getWorkers(params)
    workers.value = response.data.workers || []
    pagination.total = response.data.total || 0
  } catch (error) {
    ElMessage.error('加载工作节点失败')
  } finally {
    loading.value = false
  }
}

const resetSearch = () => {
  searchForm.status = ''
  pagination.page = 1
  loadWorkers()
}

const viewWorker = async (worker) => {
  try {
    const response = await getWorker(worker.id)
    currentWorker.value = response.data
    showDetailDialog.value = true
  } catch (error) {
    ElMessage.error('加载工作节点详情失败')
  }
}

const getStatusType = (status) => {
  const typeMap = {
    'online': 'success',
    'offline': 'danger',
    'busy': 'warning',
    'maintenance': 'info'
  }
  return typeMap[status] || 'info'
}

const getStatusText = (status) => {
  const textMap = {
    'online': '在线',
    'offline': '离线',
    'busy': '忙碌',
    'maintenance': '维护中'
  }
  return textMap[status] || status
}

const getLoadPercentage = (worker) => {
  if (!worker.capacity) return 0
  return Math.round((worker.current_load / worker.capacity) * 100)
}

const getLoadColor = (worker) => {
  const percentage = getLoadPercentage(worker)
  if (percentage >= 90) return '#F56C6C'
  if (percentage >= 70) return '#E6A23C'
  return '#67C23A'
}

const formatTime = (time) => {
  return new Date(time).toLocaleString()
}

const getMetadataList = (metadata) => {
  if (!metadata) return []
  try {
    const parsed = typeof metadata === 'string' ? JSON.parse(metadata) : metadata
    return Object.entries(parsed).map(([key, value]) => ({ key, value }))
  } catch {
    return []
  }
}
</script>

<style scoped>
.workers {
  padding: 24px;
  min-height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
  background: rgba(255, 255, 255, 0.95);
  padding: 20px 24px;
  border-radius: 16px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.1);
  backdrop-filter: blur(10px);
  border: 1px solid rgba(255, 255, 255, 0.2);
}

.page-header h1 {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  margin: 0;
  font-size: 28px;
  font-weight: 600;
}

.search-card {
  margin-bottom: 24px;
  background: rgba(255, 255, 255, 0.95);
  border-radius: 16px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.1);
  backdrop-filter: blur(10px);
  border: 1px solid rgba(255, 255, 255, 0.2);
  overflow: hidden;
}

.search-card :deep(.el-card__body) {
  padding: 24px;
}

.workers :deep(.el-card) {
  background: rgba(255, 255, 255, 0.95);
  border-radius: 16px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.1);
  backdrop-filter: blur(10px);
  border: 1px solid rgba(255, 255, 255, 0.2);
  overflow: hidden;
}

.workers :deep(.el-card__body) {
  padding: 24px;
}

.workers :deep(.el-table) {
  background: transparent;
  border-radius: 12px;
  overflow: hidden;
}

.workers :deep(.el-table__header) {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.workers :deep(.el-table__header th) {
  background: transparent;
  color: white;
  font-weight: 600;
  border: none;
  padding: 16px 12px;
}

.workers :deep(.el-table__body tr) {
  transition: all 0.3s ease;
}

.workers :deep(.el-table__body tr:hover) {
  background: rgba(102, 126, 234, 0.1);
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.workers :deep(.el-table__body td) {
  border: none;
  padding: 16px 12px;
}

.workers :deep(.el-button) {
  border-radius: 8px;
  transition: all 0.3s ease;
  font-weight: 500;
}

.workers :deep(.el-button:hover) {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.workers :deep(.el-button--primary) {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border: none;
}

.workers :deep(.el-tag) {
  border-radius: 20px;
  padding: 4px 12px;
  font-weight: 500;
  border: none;
}

.workers :deep(.el-tag--success) {
  background: linear-gradient(135deg, #67c23a 0%, #85ce61 100%);
  color: white;
}

.workers :deep(.el-tag--danger) {
  background: linear-gradient(135deg, #f56c6c 0%, #f78989 100%);
  color: white;
}

.workers :deep(.el-tag--warning) {
  background: linear-gradient(135deg, #e6a23c 0%, #ebb563 100%);
  color: white;
}

.workers :deep(.el-tag--info) {
  background: linear-gradient(135deg, #909399 0%, #a6a9ad 100%);
  color: white;
}

.load-info {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 8px 12px;
  background: linear-gradient(135deg, #f8f9fa 0%, #e9ecef 100%);
  border-radius: 8px;
  border: 1px solid rgba(102, 126, 234, 0.1);
}

.load-text {
  font-size: 12px;
  color: #667eea;
  min-width: 50px;
  font-weight: 600;
}

.workers :deep(.el-progress) {
  width: 120px;
}

.workers :deep(.el-progress__text) {
  display: none;
}

.pagination {
  margin-top: 24px;
  text-align: right;
  padding: 20px;
  background: rgba(255, 255, 255, 0.95);
  border-radius: 16px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.1);
  backdrop-filter: blur(10px);
  border: 1px solid rgba(255, 255, 255, 0.2);
}

.workers :deep(.el-pagination) {
  justify-content: flex-end;
}

.workers :deep(.el-pagination .el-pager li.is-active) {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  border-radius: 6px;
}

.workers :deep(.el-dialog) {
  border-radius: 16px;
  overflow: hidden;
  box-shadow: 0 20px 40px rgba(0, 0, 0, 0.15);
}

.workers :deep(.el-dialog__header) {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  padding: 24px;
}

.workers :deep(.el-dialog__title) {
  color: white;
  font-weight: 600;
  font-size: 18px;
}

.workers :deep(.el-dialog__body) {
  padding: 24px;
}

.workers :deep(.el-descriptions) {
  margin-bottom: 24px;
}

.workers :deep(.el-descriptions__header) {
  background: linear-gradient(135deg, #f8f9fa 0%, #e9ecef 100%);
  font-weight: 600;
  color: #2c3e50;
}

.workers :deep(.el-descriptions__body .el-descriptions__table) {
  border-radius: 8px;
  overflow: hidden;
}

.workers :deep(.el-descriptions__label) {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  font-weight: 500;
}

.metadata-section {
  margin-top: 24px;
  padding: 20px;
  background: linear-gradient(135deg, #f8f9fa 0%, #e9ecef 100%);
  border-radius: 12px;
  border: 1px solid rgba(102, 126, 234, 0.1);
}

.metadata-section h3 {
  margin: 0 0 16px 0;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  font-size: 16px;
  font-weight: 600;
}

.metadata-section :deep(.el-table) {
  border-radius: 8px;
  overflow: hidden;
  background: white;
}

.metadata-section :deep(.el-table th) {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  font-weight: 500;
}

.workers :deep(.el-input__wrapper) {
  border-radius: 8px;
  transition: all 0.3s ease;
}

.workers :deep(.el-input__wrapper:hover) {
  box-shadow: 0 2px 8px rgba(102, 126, 234, 0.2);
}

.workers :deep(.el-input__wrapper.is-focus) {
  box-shadow: 0 0 0 2px rgba(102, 126, 234, 0.2);
  border-color: #667eea;
}

.workers :deep(.el-select .el-input__wrapper) {
  border-radius: 8px;
}

@media (max-width: 768px) {
  .workers {
    padding: 16px;
  }
  
  .page-header {
    flex-direction: column;
    gap: 16px;
    text-align: center;
  }
  
  .page-header h1 {
    font-size: 24px;
  }
  
  .workers :deep(.el-table) {
    font-size: 14px;
  }
  
  .workers :deep(.el-button) {
    padding: 8px 12px;
    font-size: 12px;
  }
  
  .load-info {
    flex-direction: column;
    gap: 8px;
  }
  
  .workers :deep(.el-progress) {
    width: 100px;
  }
  
  .workers :deep(.el-dialog) {
    width: 95%;
    margin: 0 auto;
  }
}
</style>
