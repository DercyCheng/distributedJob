<template>
  <div class="login-container">
    <div class="login-box">
      <div class="login-title">
        <h1>DistributedJob</h1>
        <p>分布式任务调度系统</p>
      </div>
      <el-form
        ref="loginFormRef"
        :model="loginForm"
        :rules="loginRules"
        class="login-form"
      >
        <el-form-item prop="username">
          <el-input
            v-model="loginForm.username"
            placeholder="用户名"
            prefix-icon="User"
          />
        </el-form-item>
        <el-form-item prop="password">
          <el-input
            v-model="loginForm.password"
            placeholder="密码"
            prefix-icon="Lock"
            type="password"
            show-password
            @keyup.enter="handleLogin"
          />
        </el-form-item>
        <el-form-item>
          <el-button
            :loading="loading"
            type="primary"
            class="login-button"
            @click="handleLogin"
          >
            登录
          </el-button>
        </el-form-item>
      </el-form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from "vue";
import { useRouter } from "vue-router";
import { ElMessage, type FormInstance } from "element-plus";
import { User, Lock } from "@element-plus/icons-vue";
import { useUserStore } from "@/store/modules/user";
import { validateUsername, validatePassword } from "@/utils/validate";

const router = useRouter();
const userStore = useUserStore();
const loginForm = reactive({
  username: "admin", // 默认用户名
  password: "admin123", // 默认密码与数据库一致
});

const loginRules = {
  username: [
    { required: true, message: "请输入用户名", trigger: "blur" },
    {
      validator: (rule: any, value: string, callback: any) => {
        if (!value) {
          callback(new Error("请输入用户名"));
        } else if (!validateUsername(value)) {
          callback(new Error("用户名格式不正确"));
        } else {
          callback();
        }
      },
      trigger: "blur",
    },
  ],
  password: [
    { required: true, message: "请输入密码", trigger: "blur" },
    { min: 8, message: "密码长度不能少于8个字符", trigger: "blur" },
  ],
};

const loginFormRef = ref<FormInstance>();
const loading = ref(false);

const handleLogin = async () => {
  if (!loginFormRef.value) return;

  await loginFormRef.value.validate(async (valid) => {
    if (valid) {
      loading.value = true;
      try {
        const result = await userStore.login(loginForm.username, loginForm.password);
        console.log("Login result:", result); // 调试信息
        ElMessage.success("登录成功");
        router.push({ path: "/" });
      } catch (error: any) {
        console.error("Login error:", error); // 添加详细错误日志
        
        if (error.response && error.response.status === 401) {
          // 具体处理401未授权错误
          const data = error.response.data;
          if (data && data.message) {
            ElMessage.error(data.message);
          } else {
            ElMessage.error("用户名或密码错误");
          }
        } else if (error.message) {
          ElMessage.error(error.message);
        } else {
          ElMessage.error("登录失败，请检查用户名和密码");
        }
      } finally {
        loading.value = false;
      }
    }
  });
};
</script>

<style lang="scss" scoped>
.login-container {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100vh;
  background-color: #f5f7fa;

  .login-box {
    width: 400px;
    padding: 40px;
    margin: 0 auto;
    background: #fff;
    border-radius: 4px;
    box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);

    .login-title {
      text-align: center;
      margin-bottom: 30px;

      h1 {
        font-size: 28px;
        color: #409eff;
        margin-bottom: 10px;
      }

      p {
        font-size: 16px;
        color: #606266;
      }
    }

    .login-form {
      .login-button {
        width: 100%;
      }
    }
  }
}
</style>
