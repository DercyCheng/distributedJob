<template>
  <div class="login-container">
    <div class="login-wrapper">
      <!-- 背景装饰 -->
      <div class="bg-decoration">
        <div class="circle circle1"></div>
        <div class="circle circle2"></div>
        <div class="circle circle3"></div>
      </div>
      
      <!-- 登录表单 -->
      <el-card class="login-card" shadow="always">
        <div class="login-header">
          <div class="logo">
            <el-icon size="40" color="#409EFF"><Timer /></el-icon>
          </div>
          <h1>Go Job</h1>
          <p>分布式任务调度系统</p>
        </div>
          <!-- 登录功能已移除 -->
        <div class="welcome-section">
          <div class="welcome-message">
            <h2>欢迎使用 Go Job</h2>
            <p>分布式任务调度系统已准备就绪</p>
            <el-button 
              type="primary" 
              size="large" 
              class="enter-btn"
              @click="enterSystem"
            >
              进入系统
            </el-button>
          </div>
        </div>        
        <!-- 系统信息 -->
        <div class="system-info">
          <el-divider>系统信息</el-divider>
          <div class="info-content">
            <p><strong>系统名称:</strong> Go Job 分布式任务调度系统</p>
            <p><strong>版本:</strong> v1.0.0</p>
            <p><strong>状态:</strong> <el-tag type="success" size="small">运行中</el-tag></p>
          </div>
        </div>        
        <!-- 调试信息 - 移除登录相关调试 -->
        <div v-if="showDebug" class="debug-info">
          <el-divider>调试信息</el-divider>
          <div class="debug-content">
            <p><strong>系统状态:</strong> 运行中</p>
            <p><strong>版本:</strong> v1.0.0</p>
            <p><strong>环境:</strong> {{ process.env.NODE_ENV || 'development' }}</p>
          </div>
          <el-button size="small" @click="showDebug = false">隐藏调试</el-button>
        </div>
        
        <!-- 调试按钮 -->
        <div class="debug-toggle">
          <el-button 
            size="small" 
            type="info" 
            plain
            @click="showDebug = !showDebug"
          >
            {{ showDebug ? '隐藏' : '显示' }}调试信息
          </el-button>
        </div>
      </el-card>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { Timer } from '@element-plus/icons-vue'

const router = useRouter()
const showDebug = ref(false)

// 直接进入系统
const enterSystem = () => {
  router.push('/dashboard')
}
</script>
</script>

<style scoped>
.login-container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  position: relative;
  overflow: hidden;
}

.login-wrapper {
  position: relative;
  z-index: 10;
}

.bg-decoration {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  overflow: hidden;
  z-index: 1;
}

.circle {
  position: absolute;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.1);
  animation: float 6s ease-in-out infinite;
}

.circle1 {
  width: 80px;
  height: 80px;
  top: 10%;
  left: 10%;
  animation-delay: 0s;
}

.circle2 {
  width: 120px;
  height: 120px;
  top: 70%;
  right: 10%;
  animation-delay: 2s;
}

.circle3 {
  width: 60px;
  height: 60px;
  top: 30%;
  right: 30%;
  animation-delay: 4s;
}

@keyframes float {
  0%, 100% {
    transform: translateY(0px) rotate(0deg);
  }
  50% {
    transform: translateY(-20px) rotate(180deg);
  }
}

.login-card {
  width: 400px;
  padding: 40px;
  border-radius: 20px;
  backdrop-filter: blur(10px);
  background: rgba(255, 255, 255, 0.95);
  box-shadow: 0 25px 45px rgba(0, 0, 0, 0.1);
  border: 1px solid rgba(255, 255, 255, 0.2);
}

.login-header {
  text-align: center;
  margin-bottom: 40px;
}

.logo {
  margin-bottom: 20px;
}

.login-header h1 {
  margin: 0 0 10px 0;
  font-size: 32px;
  font-weight: bold;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.login-header p {
  margin: 0;
  color: #666;
  font-size: 16px;
}

.login-form {
  margin-bottom: 30px;
}

.welcome-section {
  text-align: center;
  margin-bottom: 30px;
}

.welcome-message h2 {
  margin: 0 0 15px 0;
  color: #333;
  font-size: 24px;
}

.welcome-message p {
  margin: 0 0 25px 0;
  color: #666;
  font-size: 16px;
}

.enter-btn {
  width: 100%;
  height: 50px;
  font-size: 16px;
  font-weight: bold;
  border-radius: 25px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border: none;
  transition: all 0.3s ease;
}

.enter-btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 10px 20px rgba(102, 126, 234, 0.4);
}

.system-info {
  margin-bottom: 20px;
}

.info-content {
  font-size: 14px;
  color: #495057;
  line-height: 1.6;
}

.info-content p {
  margin: 8px 0;
}

.debug-toggle {
  text-align: center;
  margin-top: 15px;
}

.login-form .el-input {
  height: 48px;
}

.login-form .el-input__inner {
  height: 48px;
  font-size: 16px;
  border-radius: 8px;
}

.el-divider {
  margin: 20px 0 10px 0;
}

.el-divider__text {
  color: #6c757d;
  font-size: 14px;
}

/* 响应式设计 */
@media (max-width: 480px) {
  .login-card {
    width: 90%;
    padding: 30px 20px;
  }
  
  .login-header h1 {
    font-size: 28px;
  }
}
</style>
