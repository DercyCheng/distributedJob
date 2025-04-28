<template>
  <div class="dashboard-container">
    <el-row :gutter="20">
      <!-- Overview Stats Cards -->
      <el-col :span="6" v-for="(item, index) in statsCards" :key="index">
        <el-card class="stat-card" shadow="hover">
          <div class="stat-card-content">
            <div class="stat-icon" :style="{ backgroundColor: item.color }">
              <el-icon><component :is="item.icon" /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ item.value }}</div>
              <div class="stat-title">{{ item.title }}</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20" class="mt-20">
      <!-- Task Execution Trend -->
      <el-col :span="16">
        <el-card shadow="hover">
          <template #header>
            <div class="card-header">
              <span>任务执行趋势</span>
              <el-radio-group v-model="timeRange" size="small">
                <el-radio-button label="week">近一周</el-radio-button>
                <el-radio-button label="month">近一月</el-radio-button>
              </el-radio-group>
            </div>
          </template>
          <div class="chart-container">
            <!-- Chart placeholder -->
            <div class="chart-placeholder">
              <el-empty description="加载数据中..." />
            </div>
          </div>
        </el-card>
      </el-col>

      <!-- Department Task Distribution -->
      <el-col :span="8">
        <el-card shadow="hover">
          <template #header>
            <div class="card-header">
              <span>部门任务分布</span>
            </div>
          </template>
          <div class="chart-container">
            <!-- Chart placeholder -->
            <div class="chart-placeholder">
              <el-empty description="加载数据中..." />
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20" class="mt-20">
      <!-- Recent Execution Records -->
      <el-col :span="24">
        <el-card shadow="hover">
          <template #header>
            <div class="card-header">
              <span>最近执行记录</span>
              <el-button type="primary" link @click="navigateToRecords"
                >查看更多</el-button
              >
            </div>
          </template>
          <el-table
            :data="recentRecords"
            style="width: 100%"
            v-loading="loading"
          >
            <el-table-column prop="taskName" label="任务名称" min-width="150" />
            <el-table-column
              prop="departmentName"
              label="所属部门"
              width="120"
            />
            <el-table-column prop="startTime" label="执行时间" width="180" />
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
            <el-table-column label="操作" width="120">
              <template #default="scope">
                <el-button
                  type="primary"
                  link
                  @click="viewRecordDetail(scope.row.id)"
                >
                  查看详情
                </el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from "vue";
import { useRouter } from "vue-router";
import { ElMessage } from "element-plus";
import { Clock, List, Check, WarningFilled } from "@element-plus/icons-vue";
import { getRecordList } from "@/api/record";
import type { Record } from "@/api/record";

const router = useRouter();
const loading = ref(false);
const timeRange = ref("week");
const recentRecords = ref<Record[]>([]);

// Stats cards data
const statsCards = ref([
  {
    title: "任务总数",
    value: "--",
    icon: "List",
    color: "#409EFF",
  },
  {
    title: "今日执行次数",
    value: "--",
    icon: "Clock",
    color: "#67C23A",
  },
  {
    title: "执行成功率",
    value: "--",
    icon: "Check",
    color: "#E6A23C",
  },
  {
    title: "异常任务数",
    value: "--",
    icon: "WarningFilled",
    color: "#F56C6C",
  },
]);

// Fetch recent records
const fetchRecentRecords = async () => {
  loading.value = true;
  try {
    const res = await getRecordList({
      page: 1,
      pageSize: 10,
    });
    recentRecords.value = res.list;
  } catch (error) {
    console.error("Failed to fetch recent records:", error);
    ElMessage.error("获取最近执行记录失败");
  } finally {
    loading.value = false;
  }
};

// View record detail
const viewRecordDetail = (id: number) => {
  router.push(`/record/detail/${id}`);
};

// Navigate to records page
const navigateToRecords = () => {
  router.push("/record/list");
};

onMounted(() => {
  fetchRecentRecords();
});
</script>

<style lang="scss" scoped>
.dashboard-container {
  .stat-card {
    margin-bottom: 20px;

    .stat-card-content {
      display: flex;
      align-items: center;

      .stat-icon {
        width: 60px;
        height: 60px;
        border-radius: 50%;
        display: flex;
        justify-content: center;
        align-items: center;
        margin-right: 15px;

        .el-icon {
          font-size: 24px;
          color: white;
        }
      }

      .stat-info {
        .stat-value {
          font-size: 24px;
          font-weight: bold;
          margin-bottom: 5px;
        }

        .stat-title {
          font-size: 14px;
          color: #909399;
        }
      }
    }
  }

  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .chart-container {
    height: 300px;

    .chart-placeholder {
      height: 100%;
      display: flex;
      justify-content: center;
      align-items: center;
    }
  }
}
</style>
