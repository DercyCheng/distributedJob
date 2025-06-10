<template>
  <div id="app">
    <el-container>
      <!-- 侧边栏 -->
      <el-aside width="250px" class="sidebar">
        <div class="logo">
          <el-icon size="32" color="#fff"><Timer /></el-icon>
          <div class="logo-text">
            <h2>Go Job</h2>
            <p>分布式任务调度系统</p>
          </div>
        </div>
        <el-menu
          :default-active="$route.path"
          router
          class="sidebar-menu"
          background-color="#2c3e50"
          text-color="#bdc3c7"
          active-text-color="#3498db"
          unique-opened
        >
          <el-menu-item index="/dashboard">
            <el-icon><HomeFilled /></el-icon>
            <span>仪表板</span>
          </el-menu-item>
          <el-menu-item index="/jobs">
            <el-icon><Timer /></el-icon>
            <span>任务管理</span>
          </el-menu-item>
          <el-menu-item index="/executions">
            <el-icon><List /></el-icon>
            <span>执行记录</span>
          </el-menu-item>
          <el-menu-item index="/workers">
            <el-icon><Monitor /></el-icon>
            <span>工作节点</span>
          </el-menu-item>
          <el-menu-item index="/logs">
            <el-icon><Document /></el-icon>
            <span>日志查看</span>
          </el-menu-item>
        </el-menu>
      </el-aside>

      <!-- 主内容区 -->
      <el-container>
        <!-- 头部 -->
        <el-header class="header">
          <div class="header-left">
            <h2>{{ currentPageTitle }}</h2>
          </div>
          <div class="header-right">
            <div class="header-actions">
              <el-button type="primary" size="default" @click="handleNotifications">
                <el-icon><Bell /></el-icon>
                通知
              </el-button>
              
              <el-dropdown @command="handleUserAction">
                <span class="user-info">
                  <el-avatar :src="userAvatar" size="small">
                    <el-icon><User /></el-icon>
                  </el-avatar>
                  <span class="username">管理员</span>
                  <el-icon class="dropdown-icon"><ArrowDown /></el-icon>
                </span>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item command="profile">
                      <el-icon><User /></el-icon>
                      个人设置
                    </el-dropdown-item>
                    <el-dropdown-item command="settings">
                      <el-icon><Setting /></el-icon>
                      系统设置
                    </el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
            </div>
          </div>
        </el-header>

        <!-- 主体内容 -->
        <el-main class="main-content">
          <router-view />
        </el-main>
      </el-container>
    </el-container>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { 
  Timer, HomeFilled, List, Monitor, Document,
  Bell, User, ArrowDown, Setting
} from '@element-plus/icons-vue'

const route = useRoute()
const router = useRouter()

const userAvatar = ''

const currentPageTitle = computed(() => {
  const titleMap = {
    '/dashboard': '仪表板',
    '/jobs': '任务管理',
    '/executions': '执行记录',
    '/workers': '工作节点',
    '/logs': '日志查看',
    '/profile': '个人设置'
  }
  return titleMap[route.path] || '未知页面'
})

const handleNotifications = () => {
  ElMessage.info('通知功能待开发')
}

const handleUserAction = (command) => {
  if (command === 'profile') {
    router.push('/profile')
    ElMessage.success('跳转到个人设置页面')
  } else if (command === 'settings') {
    ElMessage.info('系统设置功能待开发')
  }
}
</script>

<style scoped>
#app {
  height: 100vh;
  font-family: 'Helvetica Neue', Helvetica, 'PingFang SC', 'Hiragino Sans GB', 'Microsoft YaHei', SimSun, sans-serif;
}

.sidebar {
  background-color: #2c3e50;
  overflow: hidden;
}

.logo {
  padding: 20px;
  display: flex;
  align-items: center;
  gap: 12px;
  color: #ecf0f1;
  border-bottom: 1px solid #34495e;
}

.logo-text h2 {
  margin: 0 0 5px 0;
  font-size: 24px;
  font-weight: bold;
  color: #fff;
}

.logo-text p {
  margin: 0;
  font-size: 12px;
  opacity: 0.8;
  color: #bdc3c7;
}

.sidebar-menu {
  border: none;
}

.sidebar-menu .el-menu-item {
  border-radius: 8px;
  margin: 4px 8px;
  transition: all 0.3s ease;
}

.sidebar-menu .el-menu-item:hover {
  background-color: #34495e !important;
  transform: translateX(4px);
}

.sidebar-menu .el-menu-item.is-active {
  background-color: #3498db !important;
  color: #fff !important;
}

.header {
  background-color: #fff;
  border-bottom: 1px solid #e4e7ed;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 20px;
}

.header-left h2 {
  margin: 0;
  color: #2c3e50;
  font-size: 20px;
  font-weight: 600;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 20px;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 12px;
  cursor: pointer;
  padding: 8px 12px;
  border-radius: 8px;
  transition: all 0.3s ease;
}

.user-info:hover {
  background-color: #f5f7fa;
}

.username {
  font-weight: 500;
  color: #2c3e50;
}

.dropdown-icon {
  font-size: 12px;
  transition: transform 0.3s ease;
}

.main-content {
  background-color: #f5f7fa;
  padding: 20px;
}
</style>
