<template>
  <div class="logs">
    <div class="page-header">
      <h1>日志查看</h1>
      <el-button type="primary" @click="refreshLogs">
        <el-icon><Refresh /></el-icon>
        刷新
      </el-button>
    </div>

    <!-- 日志筛选 -->
    <el-card class="filter-card">
      <el-form :model="filterForm" inline>
        <el-form-item label="日志级别">
          <el-select v-model="filterForm.level" placeholder="请选择级别">
            <el-option label="全部" value="" />
            <el-option label="DEBUG" value="debug" />
            <el-option label="INFO" value="info" />
            <el-option label="WARN" value="warn" />
            <el-option label="ERROR" value="error" />
          </el-select>
        </el-form-item>
        <el-form-item label="服务类型">
          <el-select v-model="filterForm.service" placeholder="请选择服务">
            <el-option label="全部" value="" />
            <el-option label="调度器" value="scheduler" />
            <el-option label="工作节点" value="worker" />
            <el-option label="Web服务" value="web" />
          </el-select>
        </el-form-item>
        <el-form-item label="关键字">
          <el-input
            v-model="filterForm.keyword"
            placeholder="请输入关键字"
            clearable
            style="width: 200px"
          />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="applyFilter">筛选</el-button>
          <el-button @click="resetFilter">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 日志内容 -->
    <el-card>
      <div class="log-container">
        <div class="log-header">
          <span>实时日志</span>
          <div class="log-controls">
            <el-switch
              v-model="autoScroll"
              active-text="自动滚动"
              inactive-text="停止滚动"
            />
            <el-button size="small" @click="clearLogs">清空日志</el-button>
          </div>
        </div>
        <div
          ref="logContent"
          class="log-content"
          :class="{ 'auto-scroll': autoScroll }"
        >
          <div
            v-for="(log, index) in filteredLogs"
            :key="index"
            class="log-line"
            :class="getLevelClass(log.level)"
          >
            <span class="log-time">{{ formatTime(log.time) }}</span>
            <span class="log-level" :class="getLevelClass(log.level)">
              {{ log.level.toUpperCase() }}
            </span>
            <span class="log-service">{{ log.service }}</span>
            <span class="log-message">{{ log.message }}</span>
          </div>
        </div>
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive, computed, nextTick, onMounted, onUnmounted } from 'vue'
import { Refresh } from '@element-plus/icons-vue'
import wsClient from '@/utils/websocket'

const logs = ref([])
const autoScroll = ref(true)
const logContent = ref()

const filterForm = reactive({
  level: '',
  service: '',
  keyword: ''
})

const filteredLogs = computed(() => {
  return logs.value.filter(log => {
    if (filterForm.level && log.level !== filterForm.level) {
      return false
    }
    if (filterForm.service && log.service !== filterForm.service) {
      return false
    }
    if (filterForm.keyword && !log.message.toLowerCase().includes(filterForm.keyword.toLowerCase())) {
      return false
    }
    return true
  })
})

onMounted(() => {
  startRealTimeLogStream()
})

onUnmounted(() => {
  stopRealTimeLogStream()
})

const startRealTimeLogStream = () => {
  // 连接WebSocket获取实时日志
  wsClient.connect(`ws://${window.location.host}/ws`)
  
  wsClient.on('log', (logData) => {
    addLogEntry(logData)
  })

  wsClient.on('connected', () => {
    // 订阅日志流
    wsClient.send({
      type: 'subscribe',
      channel: 'logs'
    })
  })
}

const stopRealTimeLogStream = () => {
  wsClient.disconnect()
}

const addLogEntry = (logData) => {
  logs.value.push({
    time: new Date(logData.time),
    level: logData.level,
    service: logData.service,
    message: logData.message
  })
  
  // 限制日志数量
  if (logs.value.length > 1000) {
    logs.value = logs.value.slice(-800)
  }
  
  if (autoScroll.value) {
    nextTick(() => {
      scrollToBottom()
    })
  }
}

const scrollToBottom = () => {
  if (logContent.value) {
    logContent.value.scrollTop = logContent.value.scrollHeight
  }
}

const refreshLogs = () => {
  // 重新连接WebSocket获取最新日志
  stopRealTimeLogStream()
  logs.value = []
  startRealTimeLogStream()
}

const clearLogs = () => {
  logs.value = []
}

const applyFilter = () => {
  // 筛选逻辑在 computed 中处理
}

const resetFilter = () => {
  filterForm.level = ''
  filterForm.service = ''
  filterForm.keyword = ''
}

const getLevelClass = (level) => {
  return `level-${level}`
}

const formatTime = (time) => {
  return time.toLocaleTimeString()
}
</script>

<style scoped>
.logs {
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

.filter-card {
  margin-bottom: 24px;
  background: rgba(255, 255, 255, 0.95);
  border-radius: 16px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.1);
  backdrop-filter: blur(10px);
  border: 1px solid rgba(255, 255, 255, 0.2);
  overflow: hidden;
}

.filter-card :deep(.el-card__body) {
  padding: 24px;
}

.logs :deep(.el-card) {
  background: rgba(255, 255, 255, 0.95);
  border-radius: 16px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.1);
  backdrop-filter: blur(10px);
  border: 1px solid rgba(255, 255, 255, 0.2);
  overflow: hidden;
}

.logs :deep(.el-card__body) {
  padding: 24px;
}

.logs :deep(.el-button) {
  border-radius: 8px;
  transition: all 0.3s ease;
  font-weight: 500;
}

.logs :deep(.el-button:hover) {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.logs :deep(.el-button--primary) {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border: none;
}

.logs :deep(.el-input__wrapper) {
  border-radius: 8px;
  transition: all 0.3s ease;
}

.logs :deep(.el-input__wrapper:hover) {
  box-shadow: 0 2px 8px rgba(102, 126, 234, 0.2);
}

.logs :deep(.el-input__wrapper.is-focus) {
  box-shadow: 0 0 0 2px rgba(102, 126, 234, 0.2);
  border-color: #667eea;
}

.logs :deep(.el-select .el-input__wrapper) {
  border-radius: 8px;
}

.logs :deep(.el-switch.is-checked .el-switch__core) {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-color: #667eea;
}

.log-container {
  height: 600px;
  display: flex;
  flex-direction: column;
  background: linear-gradient(135deg, #2d3748 0%, #1a202c 100%);
  border-radius: 12px;
  overflow: hidden;
  border: 2px solid rgba(102, 126, 234, 0.2);
}

.log-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  font-weight: 600;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}

.log-controls {
  display: flex;
  align-items: center;
  gap: 16px;
}

.log-controls :deep(.el-switch__label) {
  color: white;
  font-weight: 500;
}

.log-controls :deep(.el-switch__label.is-active) {
  color: white;
}

.log-controls .el-button {
  background: rgba(255, 255, 255, 0.2);
  border: 1px solid rgba(255, 255, 255, 0.3);
  color: white;
  border-radius: 6px;
  font-size: 12px;
  padding: 6px 12px;
}

.log-controls .el-button:hover {
  background: rgba(255, 255, 255, 0.3);
  transform: none;
  box-shadow: none;
}

.log-content {
  flex: 1;
  overflow-y: auto;
  background: linear-gradient(135deg, #1a202c 0%, #2d3748 100%);
  color: #e2e8f0;
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
  font-size: 13px;
  line-height: 1.5;
  padding: 16px 20px;
}

.log-line {
  display: flex;
  margin-bottom: 4px;
  white-space: nowrap;
  padding: 4px 8px;
  border-radius: 4px;
  transition: all 0.2s ease;
}

.log-line:hover {
  background: rgba(102, 126, 234, 0.1);
}

.log-time {
  color: #81e6d9;
  margin-right: 12px;
  min-width: 85px;
  font-weight: 500;
}

.log-level {
  margin-right: 12px;
  min-width: 65px;
  font-weight: bold;
  text-align: center;
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 11px;
}

.log-service {
  color: #68d391;
  margin-right: 12px;
  min-width: 85px;
  font-weight: 500;
}

.log-message {
  flex: 1;
  word-break: break-all;
  white-space: pre-wrap;
  color: #e2e8f0;
}

.level-debug {
  background: rgba(113, 128, 150, 0.3);
  color: #a0aec0;
}

.level-info {
  background: rgba(66, 153, 225, 0.3);
  color: #63b3ed;
}

.level-warn {
  background: rgba(237, 137, 54, 0.3);
  color: #f6ad55;
}

.level-error {
  background: rgba(245, 101, 101, 0.3);
  color: #fc8181;
}

.auto-scroll {
  scroll-behavior: smooth;
}

/* 自定义滚动条 */
.log-content::-webkit-scrollbar {
  width: 8px;
}

.log-content::-webkit-scrollbar-track {
  background: rgba(45, 55, 72, 0.5);
  border-radius: 4px;
}

.log-content::-webkit-scrollbar-thumb {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: 4px;
  transition: all 0.3s ease;
}

.log-content::-webkit-scrollbar-thumb:hover {
  background: linear-gradient(135deg, #5a67d8 0%, #6b46c1 100%);
}

@media (max-width: 768px) {
  .logs {
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
  
  .log-container {
    height: 500px;
  }
  
  .log-header {
    flex-direction: column;
    gap: 12px;
  }
  
  .log-controls {
    gap: 12px;
  }
  
  .log-line {
    flex-direction: column;
    gap: 4px;
    padding: 8px;
  }
  
  .log-time,
  .log-level,
  .log-service {
    min-width: auto;
    margin-right: 0;
  }
  
  .log-content {
    font-size: 12px;
    padding: 12px 16px;
  }
}
</style>
