<template>
  <div class="profile-container">
    <el-card class="profile-card">
      <template #header>
        <div class="card-header">
          <span>个人设置</span>
        </div>
      </template>
      
      <el-tabs v-model="activeTab" type="border-card">
        <!-- 基本信息 -->
        <el-tab-pane label="基本信息" name="profile">          <el-form
            ref="profileFormRef"
            :model="profileForm"
            :rules="profileRules"
            label-width="100px"
            class="profile-form"
          >
            <el-form-item label="用户名">
              <el-input :value="userStore.user.username" disabled />
            </el-form-item>
            <el-form-item label="邮箱" prop="email">
              <el-input v-model="profileForm.email" placeholder="请输入邮箱" />
            </el-form-item>
            <el-form-item label="手机号" prop="phone">
              <el-input v-model="profileForm.phone" placeholder="请输入手机号" />
            </el-form-item>
            <el-form-item label="显示名称" prop="display_name">
              <el-input v-model="profileForm.display_name" placeholder="请输入显示名称" />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="updateProfile" :loading="profileLoading">
                更新信息
              </el-button>
              <el-button @click="resetProfileForm">重置</el-button>
            </el-form-item>
          </el-form>
        </el-tab-pane>

        <!-- 修改密码 -->
        <el-tab-pane label="修改密码" name="password">          <el-form
            ref="passwordFormRef"
            :model="passwordForm"
            :rules="passwordRules"
            label-width="100px"
            class="password-form"
          >
            <el-form-item label="旧密码" prop="old_password">
              <el-input
                v-model="passwordForm.old_password"
                type="password"
                placeholder="请输入旧密码"
                show-password
              />
            </el-form-item>
            <el-form-item label="新密码" prop="new_password">
              <el-input
                v-model="passwordForm.new_password"
                type="password"
                placeholder="请输入新密码"
                show-password
              />
            </el-form-item>
            <el-form-item label="确认密码" prop="confirm_password">
              <el-input
                v-model="passwordForm.confirm_password"
                type="password"
                placeholder="请再次输入新密码"
                show-password
              />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="changePassword" :loading="passwordLoading">
                修改密码
              </el-button>
              <el-button @click="resetPasswordForm">重置</el-button>
            </el-form-item>
          </el-form>
        </el-tab-pane>
      </el-tabs>
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { useAuthStore } from '@/stores/auth'
import { updateProfile as updateUserProfile, changePassword as changeUserPassword } from '@/api/auth'

const userStore = useAuthStore()
const activeTab = ref('profile')
const profileLoading = ref(false)
const passwordLoading = ref(false)

// 个人信息表单
const profileForm = reactive({
  email: '',
  phone: '',
  display_name: ''
})

const profileRules = {
  email: [
    { type: 'email', message: '请输入正确的邮箱地址', trigger: 'blur' }
  ],
  phone: [
    { pattern: /^1[3-9]\d{9}$/, message: '请输入正确的手机号码', trigger: 'blur' }
  ]
}

// 修改密码表单
const passwordForm = reactive({
  old_password: '',
  new_password: '',
  confirm_password: ''
})

const passwordRules = {
  old_password: [
    { required: true, message: '请输入旧密码', trigger: 'blur' }
  ],
  new_password: [
    { required: true, message: '请输入新密码', trigger: 'blur' },
    { min: 6, message: '密码长度不能少于6位', trigger: 'blur' }
  ],
  confirm_password: [
    { required: true, message: '请再次输入新密码', trigger: 'blur' },
    {
      validator: (rule, value, callback) => {
        if (value !== passwordForm.new_password) {
          callback(new Error('两次输入的密码不一致'))
        } else {
          callback()
        }
      },
      trigger: 'blur'
    }
  ]
}

const profileFormRef = ref()
const passwordFormRef = ref()

// 更新个人信息
const updateProfile = async () => {
  if (!profileFormRef.value) return
  
  try {
    await profileFormRef.value.validate()
    profileLoading.value = true
    
    await updateUserProfile({
      email: profileForm.email,
      phone: profileForm.phone,
      display_name: profileForm.display_name
    })
    
    ElMessage.success('个人信息更新成功')
  } catch (error) {
    if (error?.message) {
      ElMessage.error(error.message)
    }
  } finally {
    profileLoading.value = false
  }
}

// 修改密码
const changePassword = async () => {
  if (!passwordFormRef.value) return
  
  try {
    await passwordFormRef.value.validate()
    passwordLoading.value = true
    
    await changeUserPassword({
      old_password: passwordForm.old_password,
      new_password: passwordForm.new_password
    })
    
    ElMessage.success('密码修改成功')
    resetPasswordForm()
  } catch (error) {
    if (error?.message) {
      ElMessage.error(error.message)
    }
  } finally {
    passwordLoading.value = false
  }
}

// 重置表单
const resetProfileForm = () => {
  profileFormRef.value?.resetFields()
}

const resetPasswordForm = () => {
  passwordFormRef.value?.resetFields()
}

// 初始化数据
onMounted(() => {
  // 这里可以获取用户当前的个人信息
  // 由于当前实现中用户信息存储在 store 中，可以从那里获取
})
</script>

<style scoped>
.profile-container {
  padding: 20px;
}

.profile-card {
  max-width: 800px;
  margin: 0 auto;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.profile-form,
.password-form {
  max-width: 500px;
  margin: 20px auto;
}

.el-form-item {
  margin-bottom: 22px;
}
</style>
