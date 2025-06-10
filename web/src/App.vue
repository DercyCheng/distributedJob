<template>
  <div id="app">
    <el-container>
      <!-- 侧边栏 -->
      <el-aside width="250px" class="sidebar">
        <div class="logo">
          <h2>Go Job</h2>
          <p>分布式任务调度系统</p>
        </div>
        <el-menu
          :default-active="$route.path"
          router
          class="sidebar-menu"
          background-color="#2c3e50"
          text-color="#ecf0f1"
          active-text-color="#3498db"
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
          <div class="header-content">
            <div class="breadcrumb">
              <el-breadcrumb separator="/">
                <el-breadcrumb-item :to="{ path: '/dashboard' }">首页</el-breadcrumb-item>
                <el-breadcrumb-item>{{ currentPageTitle }}</el-breadcrumb-item>
              </el-breadcrumb>
            </div>
            <div class="header-actions">
              <el-badge :value="12" class="notification">
                <el-icon size="20"><Bell /></el-icon>
              </el-badge>
              <el-dropdown>
                <span class="user-info">
                  <el-icon><User /></el-icon>
                  <span>管理员</span>
                </span>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item>个人设置</el-dropdown-item>
                    <el-dropdown-item divided>退出登录</el-dropdown-item>
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
import { useRoute } from 'vue-router'

const route = useRoute()

const currentPageTitle = computed(() => {
  const titleMap = {
    '/dashboard': '仪表板',
    '/jobs': '任务管理',
    '/executions': '执行记录',
    '/workers': '工作节点',
    '/logs': '日志查看'
  }
  return titleMap[route.path] || '未知页面'
})
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
  text-align: center;
  color: #ecf0f1;
  border-bottom: 1px solid #34495e;
}

.logo h2 {
  margin: 0 0 5px 0;
  font-size: 24px;
  font-weight: bold;
}

.logo p {
  margin: 0;
  font-size: 12px;
  opacity: 0.8;
}

.sidebar-menu {
  border: none;
}

.header {
  background-color: #fff;
  border-bottom: 1px solid #e4e7ed;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.header-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
  height: 100%;
}

.breadcrumb {
  flex: 1;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 20px;
}

.notification {
  cursor: pointer;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  padding: 8px;
  border-radius: 4px;
  transition: background-color 0.3s;
}

.user-info:hover {
  background-color: #f5f7fa;
}

.main-content {
  background-color: #f5f7fa;
  padding: 20px;
}
</style>
