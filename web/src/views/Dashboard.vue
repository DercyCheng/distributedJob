<template>
  <div class="dashboard">
    <div class="page-header">
      <h1>
        <el-icon size="24"><TrendCharts /></el-icon>
        仪表板
      </h1>
      <p>实时监控系统运行状态和性能指标</p>
    </div>
    
    <!-- 统计卡片 -->
    <el-row :gutter="20" class="stats-cards">
      <el-col :span="6">
        <el-card class="stat-card total-jobs" shadow="hover">
          <div class="stat-content">
            <div class="stat-info">
              <div class="stat-number">{{ dashboardData.totalJobs }}</div>
              <div class="stat-label">总任务数</div>
            </div>
            <div class="stat-icon">
              <el-icon size="48" color="#409EFF"><Timer /></el-icon>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card active-jobs" shadow="hover">
          <div class="stat-content">
            <div class="stat-info">
              <div class="stat-number">{{ dashboardData.activeJobs }}</div>
              <div class="stat-label">活跃任务</div>
            </div>
            <div class="stat-icon">
              <el-icon size="48" color="#67C23A"><CircleCheck /></el-icon>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card online-workers" shadow="hover">
          <div class="stat-content">
            <div class="stat-info">
              <div class="stat-number">{{ dashboardData.onlineWorkers }}/{{ dashboardData.totalWorkers }}</div>
              <div class="stat-label">在线节点</div>
            </div>
            <div class="stat-icon">
              <el-icon size="48" color="#E6A23C"><Monitor /></el-icon>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card success-rate" shadow="hover">
          <div class="stat-content">
            <div class="stat-info">
              <div class="stat-number">{{ dashboardData.successRate.toFixed(1) }}%</div>
              <div class="stat-label">成功率</div>
            </div>
            <div class="stat-icon">
              <el-icon size="48" color="#F56C6C"><TrendCharts /></el-icon>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>    <!-- 图表区域 -->
    <el-row :gutter="20" class="charts-row">
      <el-col :span="12">
        <el-card shadow="hover" class="chart-card">
          <template #header>
            <div class="card-header">
              <el-icon><TrendCharts /></el-icon>
              <span>执行趋势</span>
            </div>
          </template>
          <div id="execution-chart" style="height: 300px;"></div>
        </el-card>
      </el-col>
      <el-col :span="12">
        <el-card shadow="hover" class="chart-card">
          <template #header>
            <div class="card-header">
              <el-icon><PieChart /></el-icon>
              <span>任务状态分布</span>
            </div>
          </template>
          <div id="status-chart" style="height: 300px;"></div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 最近任务和执行记录 -->
    <el-row :gutter="20" class="tables-row">
      <el-col :span="12">
        <el-card shadow="hover" class="table-card">
          <template #header>
            <div class="card-header">
              <el-icon><Timer /></el-icon>
              <span>最近任务</span>
            </div>
          </template>
          <el-table :data="dashboardData.recentJobs" stripe class="dashboard-table">
            <el-table-column prop="name" label="任务名称" show-overflow-tooltip />
            <el-table-column prop="cron" label="调度规则" width="120" />
            <el-table-column prop="enabled" label="状态" width="80" align="center">
              <template #default="{ row }">
                <el-tag :type="row.enabled ? 'success' : 'danger'" size="small">
                  {{ row.enabled ? '启用' : '禁用' }}
                </el-tag>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
      <el-col :span="12">
        <el-card shadow="hover" class="table-card">
          <template #header>
            <div class="card-header">
              <el-icon><List /></el-icon>
              <span>最近执行</span>
            </div>
          </template>
          <el-table :data="dashboardData.recentExecutions" stripe class="dashboard-table">
            <el-table-column prop="job.name" label="任务名称" show-overflow-tooltip />
            <el-table-column prop="status" label="状态" width="80" align="center">
              <template #default="{ row }">
                <el-tag :type="getStatusType(row.status)" size="small">
                  {{ getStatusText(row.status) }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="created_at" label="执行时间" width="120">
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
  padding: 0;
}

.page-header {
  margin-bottom: 24px;
  padding: 24px 0;
  border-bottom: 1px solid #e8e8e8;
}

.page-header h1 {
  margin: 0 0 8px 0;
  font-size: 28px;
  font-weight: 600;
  color: #2c3e50;
  display: flex;
  align-items: center;
  gap: 12px;
}

.page-header p {
  margin: 0;
  color: #666;
  font-size: 16px;
}

.stats-cards {
  margin-bottom: 24px;
}

.stat-card {
  border-radius: 12px;
  transition: all 0.3s ease;
  border: none;
  overflow: hidden;
}

.stat-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.12);
}

.total-jobs {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
}

.active-jobs {
  background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%);
  color: white;
}

.online-workers {
  background: linear-gradient(135deg, #43e97b 0%, #38f9d7 100%);
  color: white;
}

.success-rate {
  background: linear-gradient(135deg, #fa709a 0%, #fee140 100%);
  color: white;
}

.stat-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px;
}

.stat-info {
  flex: 1;
}

.stat-number {
  font-size: 32px;
  font-weight: bold;
  margin-bottom: 8px;
  text-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.stat-label {
  font-size: 14px;
  opacity: 0.9;
  font-weight: 500;
}

.stat-icon {
  opacity: 0.3;
}

.charts-row {
  margin-bottom: 24px;
}

.tables-row {
  margin-bottom: 24px;
}

.chart-card,
.table-card {
  border-radius: 12px;
  border: none;
}

.card-header {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 600;
  color: #2c3e50;
}

.dashboard-table {
  border-radius: 8px;
}

.dashboard-table .el-table__header {
  background-color: #f8f9fa;
}

/* 响应式设计 */
@media (max-width: 1200px) {
  .stat-card .stat-number {
    font-size: 28px;
  }
}

@media (max-width: 768px) {
  .page-header h1 {
    font-size: 24px;
  }
  
  .stat-card .stat-number {
    font-size: 24px;
  }
  
  .stat-card .stat-label {
    font-size: 12px;
  }
}</style>
