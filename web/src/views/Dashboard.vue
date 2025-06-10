<template>
  <div class="dashboard">
    <h1>仪表板</h1>
    
    <!-- 统计卡片 -->
    <el-row :gutter="20" class="stats-cards">
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-number">{{ dashboardData.totalJobs }}</div>
            <div class="stat-label">总任务数</div>
          </div>
          <el-icon class="stat-icon" color="#409EFF"><Timer /></el-icon>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-number">{{ dashboardData.activeJobs }}</div>
            <div class="stat-label">活跃任务</div>
          </div>
          <el-icon class="stat-icon" color="#67C23A"><CircleCheck /></el-icon>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-number">{{ dashboardData.onlineWorkers }}/{{ dashboardData.totalWorkers }}</div>
            <div class="stat-label">在线节点</div>
          </div>
          <el-icon class="stat-icon" color="#E6A23C"><Monitor /></el-icon>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-number">{{ dashboardData.successRate.toFixed(1) }}%</div>
            <div class="stat-label">成功率</div>
          </div>
          <el-icon class="stat-icon" color="#F56C6C"><TrendCharts /></el-icon>
        </el-card>
      </el-col>
    </el-row>

    <!-- 图表区域 -->
    <el-row :gutter="20" class="charts-row">
      <el-col :span="12">
        <el-card>
          <template #header>
            <span>执行趋势</span>
          </template>
          <div id="execution-chart" style="height: 300px;"></div>
        </el-card>
      </el-col>
      <el-col :span="12">
        <el-card>
          <template #header>
            <span>任务状态分布</span>
          </template>
          <div id="status-chart" style="height: 300px;"></div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 最近任务和执行记录 -->
    <el-row :gutter="20">
      <el-col :span="12">
        <el-card>
          <template #header>
            <span>最近任务</span>
          </template>
          <el-table :data="dashboardData.recentJobs" stripe>
            <el-table-column prop="name" label="任务名称" />
            <el-table-column prop="cron" label="调度规则" />
            <el-table-column prop="enabled" label="状态">
              <template #default="{ row }">
                <el-tag :type="row.enabled ? 'success' : 'danger'">
                  {{ row.enabled ? '启用' : '禁用' }}
                </el-tag>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
      <el-col :span="12">
        <el-card>
          <template #header>
            <span>最近执行</span>
          </template>
          <el-table :data="dashboardData.recentExecutions" stripe>
            <el-table-column prop="job.name" label="任务名称" />
            <el-table-column prop="status" label="状态">
              <template #default="{ row }">
                <el-tag :type="getStatusType(row.status)">
                  {{ getStatusText(row.status) }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="created_at" label="执行时间">
              <template #default="{ row }">
                {{ formatTime(row.created_at) }}
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, onMounted, nextTick } from 'vue'
import * as echarts from 'echarts'
import { getDashboardData } from '@/api/stats'

const dashboardData = ref({
  totalJobs: 0,
  activeJobs: 0,
  totalWorkers: 0,
  onlineWorkers: 0,
  successRate: 0,
  recentJobs: [],
  recentExecutions: [],
  executionStats: []
})

onMounted(async () => {
  await loadDashboardData()
  await nextTick()
  initCharts()
})

const loadDashboardData = async () => {
  try {
    const response = await getDashboardData()
    dashboardData.value = response.data
  } catch (error) {
    console.error('加载仪表板数据失败:', error)
  }
}

const initCharts = () => {
  // 执行趋势图
  const executionChart = echarts.init(document.getElementById('execution-chart'))
  const executionOption = {
    tooltip: {
      trigger: 'axis'
    },
    legend: {
      data: ['成功', '失败', '总计']
    },
    xAxis: {
      type: 'category',
      data: dashboardData.value.executionStats.map(item => item.date)
    },
    yAxis: {
      type: 'value'
    },
    series: [
      {
        name: '成功',
        type: 'line',
        data: dashboardData.value.executionStats.map(item => item.success),
        itemStyle: { color: '#67C23A' }
      },
      {
        name: '失败',
        type: 'line',
        data: dashboardData.value.executionStats.map(item => item.failed),
        itemStyle: { color: '#F56C6C' }
      },
      {
        name: '总计',
        type: 'line',
        data: dashboardData.value.executionStats.map(item => item.total),
        itemStyle: { color: '#409EFF' }
      }
    ]
  }
  executionChart.setOption(executionOption)

  // 状态分布饼图
  const statusChart = echarts.init(document.getElementById('status-chart'))
  const totalSuccess = dashboardData.value.executionStats.reduce((sum, item) => sum + item.success, 0)
  const totalFailed = dashboardData.value.executionStats.reduce((sum, item) => sum + item.failed, 0)
  
  const statusOption = {
    tooltip: {
      trigger: 'item'
    },
    legend: {
      orient: 'vertical',
      left: 'left'
    },
    series: [
      {
        name: '执行状态',
        type: 'pie',
        radius: '50%',
        data: [
          { value: totalSuccess, name: '成功', itemStyle: { color: '#67C23A' } },
          { value: totalFailed, name: '失败', itemStyle: { color: '#F56C6C' } }
        ],
        emphasis: {
          itemStyle: {
            shadowBlur: 10,
            shadowOffsetX: 0,
            shadowColor: 'rgba(0, 0, 0, 0.5)'
          }
        }
      }
    ]
  }
  statusChart.setOption(statusOption)
}

const getStatusType = (status) => {
  const typeMap = {
    'success': 'success',
    'failed': 'danger',
    'running': 'warning',
    'pending': 'info'
  }
  return typeMap[status] || 'info'
}

const getStatusText = (status) => {
  const textMap = {
    'success': '成功',
    'failed': '失败',
    'running': '运行中',
    'pending': '等待中'
  }
  return textMap[status] || status
}

const formatTime = (time) => {
  return new Date(time).toLocaleString()
}
</script>

<style scoped>
.dashboard {
  padding: 20px;
}

.stats-cards {
  margin-bottom: 20px;
}

.stat-card {
  .stat-content {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
  }
  
  .stat-number {
    font-size: 28px;
    font-weight: bold;
    color: #333;
    margin-bottom: 5px;
  }
  
  .stat-label {
    font-size: 14px;
    color: #666;
  }
  
  .stat-icon {
    position: absolute;
    right: 20px;
    top: 50%;
    transform: translateY(-50%);
    font-size: 32px;
  }
}

.el-card {
  position: relative;
  margin-bottom: 20px;
}

.charts-row {
  margin-bottom: 20px;
}
</style>
