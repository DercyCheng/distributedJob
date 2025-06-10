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

const logs = ref([])
const autoScroll = ref(true)
const logContent = ref()

const filterForm = reactive({
  level: '',
  service: '',
  keyword: ''
})

// 模拟实时日志数据
const mockLogs = [
  {
    time: new Date(),
    level: 'info',
    service: 'scheduler',
    message: '调度器服务启动成功'
  },
  {
    time: new Date(Date.now() - 1000),
    level: 'info',
    service: 'worker',
    message: '工作节点注册成功: worker-001'
  },
  {
    time: new Date(Date.now() - 2000),
    level: 'debug',
    service: 'scheduler',
    message: '加载任务配置: 5个任务已加载'
  },
  {
    time: new Date(Date.now() - 3000),
    level: 'warn',
    service: 'worker',
    message: '工作节点负载较高: 90%'
  },
  {
    time: new Date(Date.now() - 4000),
    level: 'error',
    service: 'scheduler',
    message: '任务执行失败: job-123'
  }
]

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

let logInterval = null

onMounted(() => {
  logs.value = [...mockLogs]
  startLogStream()
})

onUnmounted(() => {
  stopLogStream()
})

const startLogStream = () => {
  // 模拟实时日志流
  logInterval = setInterval(() => {
    const newLog = {
      time: new Date(),
      level: ['debug', 'info', 'warn', 'error'][Math.floor(Math.random() * 4)],
      service: ['scheduler', 'worker', 'web'][Math.floor(Math.random() * 3)],
      message: `模拟日志消息 ${Date.now()}`
    }
    logs.value.push(newLog)
    
    // 限制日志数量
    if (logs.value.length > 1000) {
      logs.value = logs.value.slice(-800)
    }
    
    if (autoScroll.value) {
      nextTick(() => {
        scrollToBottom()
      })
    }
  }, 2000)
}

const stopLogStream = () => {
  if (logInterval) {
    clearInterval(logInterval)
    logInterval = null
  }
}

const scrollToBottom = () => {
  if (logContent.value) {
    logContent.value.scrollTop = logContent.value.scrollHeight
  }
}

const refreshLogs = () => {
  logs.value = [...mockLogs]
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
  padding: 20px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.filter-card {
  margin-bottom: 20px;
}

.log-container {
  height: 600px;
  display: flex;
  flex-direction: column;
}

.log-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-bottom: 10px;
  border-bottom: 1px solid #e4e7ed;
  margin-bottom: 10px;
}

.log-controls {
  display: flex;
  align-items: center;
  gap: 15px;
}

.log-content {
  flex: 1;
  overflow-y: auto;
  background-color: #1e1e1e;
  color: #d4d4d4;
  font-family: 'Courier New', monospace;
  font-size: 13px;
  line-height: 1.4;
  padding: 10px;
  border-radius: 4px;
}

.log-line {
  display: flex;
  margin-bottom: 2px;
  white-space: nowrap;
}

.log-time {
  color: #569cd6;
  margin-right: 10px;
  min-width: 80px;
}

.log-level {
  margin-right: 10px;
  min-width: 60px;
  font-weight: bold;
}

.log-service {
  color: #4ec9b0;
  margin-right: 10px;
  min-width: 80px;
}

.log-message {
  flex: 1;
  word-break: break-all;
  white-space: pre-wrap;
}

.level-debug {
  color: #808080;
}

.level-info {
  color: #4fc3f7;
}

.level-warn {
  color: #ffb74d;
}

.level-error {
  color: #f44336;
}

.auto-scroll {
  scroll-behavior: smooth;
}

::-webkit-scrollbar {
  width: 8px;
}

::-webkit-scrollbar-track {
  background: #2d2d2d;
}

::-webkit-scrollbar-thumb {
  background: #555;
  border-radius: 4px;
}

::-webkit-scrollbar-thumb:hover {
  background: #777;
}
</style>
