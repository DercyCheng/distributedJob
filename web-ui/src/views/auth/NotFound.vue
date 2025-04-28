<template>
  <div class="not-found-container">
    <div class="not-found-content">
      <h1>404</h1>
      <h2>页面不存在</h2>
      <p>您访问的页面不存在或已被删除</p>
      <el-button type="primary" @click="goHome">返回首页</el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from "vue-router";

const router = useRouter();

const goHome = () => {
  router.push("/");
};
</script>

<style lang="scss" scoped>
.not-found-container {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100vh;
  background-color: #f5f7fa;

  .not-found-content {
    text-align: center;

    h1 {
      font-size: 80px;
      color: #409eff;
      margin: 0;
    }

    h2 {
      font-size: 30px;
      color: #303133;
      margin: 20px 0;
    }

    p {
      font-size: 16px;
      color: #606266;
      margin-bottom: 30px;
    }
  }
}
</style>
