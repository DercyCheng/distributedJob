<template>
  <div class="register-container">
    <div class="register-box">
      <div class="register-title">
        <h1>DistributedJob</h1>
        <p>分布式任务调度系统</p>
      </div>
      <el-form
        ref="registerFormRef"
        :model="registerForm"
        :rules="registerRules"
        class="register-form"
      >
        <el-form-item prop="username">
          <el-input
            v-model="registerForm.username"
            placeholder="用户名"
            prefix-icon="User"
          />
        </el-form-item>
        <el-form-item prop="name">
          <el-input
            v-model="registerForm.name"
            placeholder="姓名"
            prefix-icon="UserFilled"
          />
        </el-form-item>
        <el-form-item prop="email">
          <el-input
            v-model="registerForm.email"
            placeholder="邮箱"
            prefix-icon="Message"
          />
        </el-form-item>
        <el-form-item prop="password">
          <el-input
            v-model="registerForm.password"
            placeholder="密码"
            prefix-icon="Lock"
            type="password"
            show-password
          />
        </el-form-item>
        <el-form-item prop="confirmPassword">
          <el-input
            v-model="registerForm.confirmPassword"
            placeholder="确认密码"
            prefix-icon="Lock"
            type="password"
            show-password
          />
        </el-form-item>
        <el-form-item prop="departmentId" v-if="departments.length > 0">
          <el-select
            v-model="registerForm.departmentId"
            placeholder="请选择部门"
            style="width: 100%"
          >
            <el-option
              v-for="item in departments"
              :key="item.id"
              :label="item.name"
              :value="item.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item prop="roleId" v-if="roles.length > 0">
          <el-select
            v-model="registerForm.roleId"
            placeholder="请选择角色"
            style="width: 100%"
          >
            <el-option
              v-for="item in roles"
              :key="item.id"
              :label="item.name"
              :value="item.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button
            :loading="loading"
            type="primary"
            class="register-button"
            @click="handleRegister"
          >
            注册
          </el-button>
        </el-form-item>
        <div class="login-link">
          已有账号？<el-link type="primary" @click="goToLogin">立即登录</el-link>
        </div>
      </el-form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from "vue";
import { useRouter } from "vue-router";
import { ElMessage, type FormInstance } from "element-plus";
import { User, UserFilled, Message, Lock } from "@element-plus/icons-vue";
import { useUserStore } from "@/store/modules/user";
// import { register } from "@/api/auth";
import { validateUsername, validatePassword } from "@/utils/validate";
import { getDepartmentList } from "@/api/department";
import { getRoleList } from "@/api/role";
import type { Department } from "@/api/department";
import type { Role } from "@/api/role";

const router = useRouter();
const userStore = useUserStore();
const registerForm = reactive({
  username: "",
  name: "",
  email: "",
  password: "",
  confirmPassword: "",
  departmentId: undefined as number | undefined,
  roleId: undefined as number | undefined
});

const departments = ref<Department[]>([]);
const roles = ref<Role[]>([]);

const registerRules = {
  username: [
    { required: true, message: "请输入用户名", trigger: "blur" },
    {
      validator: (rule: any, value: string, callback: any) => {
        if (!value) {
          callback(new Error("请输入用户名"));
        } else if (!validateUsername(value)) {
          callback(new Error("用户名格式不正确（4-16位字母、数字、下划线或短横线）"));
        } else {
          callback();
        }
      },
      trigger: "blur",
    },
  ],
  name: [
    { required: true, message: "请输入姓名", trigger: "blur" },
    { min: 2, message: "姓名长度不能少于2个字符", trigger: "blur" },
  ],
  email: [
    { required: true, message: "请输入邮箱", trigger: "blur" },
    { type: "email", message: "请输入正确的邮箱格式", trigger: "blur" },
  ],
  password: [
    { required: true, message: "请输入密码", trigger: "blur" },
    { min: 8, message: "密码长度不能少于8个字符", trigger: "blur" },
    {
      validator: (rule: any, value: string, callback: any) => {
        if (!value) {
          callback(new Error("请输入密码"));
        } else if (!validatePassword(value)) {
          callback(new Error("密码格式不正确（至少8位，必须包含字母和数字）"));
        } else {
          callback();
        }
      },
      trigger: "blur",
    }
  ],
  confirmPassword: [
    { required: true, message: "请再次输入密码", trigger: "blur" },
    {
      validator: (rule: any, value: string, callback: any) => {
        if (value !== registerForm.password) {
          callback(new Error("两次输入的密码不一致"));
        } else {
          callback();
        }
      },
      trigger: "blur",
    }
  ],
  departmentId: [
    { required: true, message: "请选择部门", trigger: "change" }
  ],
  roleId: [
    { required: true, message: "请选择角色", trigger: "change" }
  ]
};

const registerFormRef = ref<FormInstance>();
const loading = ref(false);

// 获取部门列表
const fetchDepartments = async () => {
  try {
    const res = await getDepartmentList();
    departments.value = res;
  } catch (error) {
    console.error("Failed to fetch departments:", error);
    ElMessage.error("获取部门列表失败");
  }
};

// 获取角色列表
const fetchRoles = async () => {
  try {
    const res = await getRoleList();
    roles.value = res;
  } catch (error) {
    console.error("Failed to fetch roles:", error);
    ElMessage.error("获取角色列表失败");
  }
};

const handleRegister = async () => {
  if (!registerFormRef.value) return;

  await registerFormRef.value.validate(async (valid) => {
    if (valid) {
      loading.value = true;      try {
        // 调用注册 API
        const result = await userStore.register({
          username: registerForm.username,
          name: registerForm.name,
          email: registerForm.email,
          password: registerForm.password,
          departmentId: registerForm.departmentId as number,
          roleId: registerForm.roleId as number
        });
        
        // 注册成功后直接登录
        if (result && result.accessToken) {
          ElMessage.success("注册成功，即将进入系统");
          router.push({ path: "/" });
        } else {
          ElMessage.success("注册成功，请登录");
          router.push({ path: "/login" });
        }
      } catch (error: any) {
        console.error("Register error:", error);
        
        if (error.response && error.response.data) {
          const data = error.response.data;
          if (data.message) {
            ElMessage.error(data.message);
          } else {
            ElMessage.error("注册失败，请稍后重试");
          }
        } else if (error.message) {
          ElMessage.error(error.message);
        } else {
          ElMessage.error("注册失败，请稍后重试");
        }
      } finally {
        loading.value = false;
      }
    }
  });
};

const goToLogin = () => {
  router.push({ path: "/login" });
};

onMounted(() => {
  fetchDepartments();
  fetchRoles();
});
</script>

<style lang="scss" scoped>
.register-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  background-color: #f5f7fa;
  padding: 20px 0;

  .register-box {
    width: 450px;
    padding: 40px;
    margin: 0 auto;
    background: #fff;
    border-radius: 4px;
    box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);

    .register-title {
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

    .register-form {
      .register-button {
        width: 100%;
      }
      
      .login-link {
        text-align: center;
        margin-top: 15px;
        font-size: 14px;
        color: #606266;
      }
    }
  }
}
</style>
