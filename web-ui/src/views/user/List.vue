<template>
  <div class="user-list-container">
    <!-- Search and Filter Form -->
    <el-card shadow="never" class="filter-container">
      <el-form :model="searchForm" inline>
        <el-form-item label="用户名">
          <el-input
            v-model="searchForm.username"
            placeholder="请输入用户名"
            clearable
            @keyup.enter="handleSearch"
          />
        </el-form-item>
        <el-form-item label="姓名">
          <el-input
            v-model="searchForm.name"
            placeholder="请输入姓名"
            clearable
            @keyup.enter="handleSearch"
          />
        </el-form-item>
        <el-form-item label="部门">
          <el-select
            v-model="searchForm.departmentId"
            placeholder="请选择部门"
            clearable
          >
            <el-option
              v-for="item in departments"
              :key="item.id"
              :label="item.name"
              :value="item.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="角色">
          <el-select
            v-model="searchForm.roleId"
            placeholder="请选择角色"
            clearable
          >
            <el-option
              v-for="item in roles"
              :key="item.id"
              :label="item.name"
              :value="item.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select
            v-model="searchForm.status"
            placeholder="请选择状态"
            clearable
          >
            <el-option label="启用" value="active" />
            <el-option label="禁用" value="disabled" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">查询</el-button>
          <el-button @click="resetSearch">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- User Table -->
    <el-card shadow="never" class="table-container">
      <template #header>
        <div class="card-header">
          <span>用户列表</span>
          <el-button type="primary" @click="handleCreate">创建用户</el-button>
        </div>
      </template>

      <el-table v-loading="loading" :data="userList" border style="width: 100%">
        <el-table-column
          prop="username"
          label="用户名"
          min-width="120"
          show-overflow-tooltip
        />
        <el-table-column
          prop="name"
          label="姓名"
          min-width="120"
          show-overflow-tooltip
        />
        <el-table-column
          prop="email"
          label="邮箱"
          min-width="180"
          show-overflow-tooltip
        />
        <el-table-column
          prop="departmentName"
          label="部门"
          min-width="120"
          show-overflow-tooltip
        />
        <el-table-column
          prop="roleName"
          label="角色"
          min-width="120"
          show-overflow-tooltip
        />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="scope">
            <el-switch
              v-model="scope.row.status"
              :active-value="'active'"
              :inactive-value="'disabled'"
              @change="() => handleStatusChange(scope.row)"
            />
          </template>
        </el-table-column>
        <el-table-column
          prop="createdAt"
          label="创建时间"
          width="180"
          show-overflow-tooltip
        />
        <el-table-column label="操作" width="220" fixed="right">
          <template #default="scope">
            <el-button type="primary" link @click="handleEdit(scope.row)">
              编辑
            </el-button>
            <el-button type="warning" link @click="handleResetPassword(scope.row)">
              重置密码
            </el-button>
            <el-button type="danger" link @click="handleDelete(scope.row)">
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- Pagination -->
      <div class="pagination-container">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.pageSize"
          :total="pagination.total"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleSizeChange"
          @current-change="handleCurrentChange"
        />
      </div>
    </el-card>

    <!-- User Form Dialog -->
    <el-dialog
      v-model="dialogVisible"
      :title="isEdit ? '编辑用户' : '新建用户'"
      width="500px"
      append-to-body
    >
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="80px"
        label-position="right"
      >
        <el-form-item label="用户名" prop="username" v-if="!isEdit">
          <el-input v-model="form.username" placeholder="请输入用户名" />
        </el-form-item>
        <el-form-item label="姓名" prop="name">
          <el-input v-model="form.name" placeholder="请输入姓名" />
        </el-form-item>
        <el-form-item label="邮箱" prop="email">
          <el-input v-model="form.email" placeholder="请输入邮箱" />
        </el-form-item>
        <el-form-item label="密码" prop="password" v-if="!isEdit">
          <el-input
            v-model="form.password"
            type="password"
            placeholder="请输入密码"
            show-password
          />
        </el-form-item>
        <el-form-item label="部门" prop="departmentId">
          <el-select
            v-model="form.departmentId"
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
        <el-form-item label="角色" prop="roleId">
          <el-select
            v-model="form.roleId"
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
      </el-form>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="dialogVisible = false">取消</el-button>
          <el-button type="primary" @click="submitForm">确定</el-button>
        </div>
      </template>
    </el-dialog>

    <!-- Reset Password Dialog -->
    <el-dialog
      v-model="resetPasswordVisible"
      title="重置密码"
      width="500px"
      append-to-body
    >
      <el-form
        ref="resetFormRef"
        :model="resetForm"
        :rules="resetRules"
        label-width="80px"
      >
        <el-form-item label="新密码" prop="password">
          <el-input
            v-model="resetForm.password"
            type="password"
            placeholder="请输入新密码"
            show-password
          />
        </el-form-item>
        <el-form-item label="确认密码" prop="confirmPassword">
          <el-input
            v-model="resetForm.confirmPassword"
            type="password"
            placeholder="请再次输入新密码"
            show-password
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="resetPasswordVisible = false">取消</el-button>
          <el-button type="primary" @click="submitResetPassword">确定</el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import type { FormInstance, FormRules } from 'element-plus';
// Import each function individually rather than using grouped imports
import { getUserList } from '@/api/user';
import { createUser } from '@/api/user';
import { updateUser } from '@/api/user';
import { deleteUser } from '@/api/user';
import { resetPassword } from '@/api/user';
import { updateUserStatus } from '@/api/user';
import type { User, UserQueryParams } from '@/api/user';
import { getDepartmentList } from '@/api/department';
import { getRoleList } from '@/api/role';
import type { Department } from '@/api/department';
import type { Role } from '@/api/role';

// Data
const loading = ref(false);
const userList = ref<User[]>([]);
const departments = ref<Department[]>([]);
const roles = ref<Role[]>([]);

// Search form
const searchForm = reactive<UserQueryParams>({
  page: 1,
  pageSize: 10,
  username: '',
  name: '',
  departmentId: undefined,
  roleId: undefined,
  status: ''
});

// Pagination
const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0
});

// Dialog control
const dialogVisible = ref(false);
const isEdit = ref(false);
const formRef = ref<FormInstance>();

// Form data
const form = reactive({
  id: 0,
  username: '',
  name: '',
  email: '',
  password: '',
  departmentId: undefined as number | undefined,
  roleId: undefined as number | undefined,
  status: 'active' as 'active' | 'disabled'
});

// Form validation rules
const rules = reactive<FormRules>({
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 3, max: 30, message: '长度在 3 到 30 个字符', trigger: 'blur' }
  ],
  name: [
    { required: true, message: '请输入姓名', trigger: 'blur' }
  ],
  email: [
    { required: true, message: '请输入邮箱', trigger: 'blur' },
    { type: 'email', message: '请输入正确的邮箱格式', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, message: '密码长度不能少于6个字符', trigger: 'blur' }
  ],
  departmentId: [
    { required: true, message: '请选择部门', trigger: 'change' }
  ],
  roleId: [
    { required: true, message: '请选择角色', trigger: 'change' }
  ]
});

// Reset password
const resetPasswordVisible = ref(false);
const resetFormRef = ref<FormInstance>();
const currentUserId = ref<number>(0);
const resetForm = reactive({
  password: '',
  confirmPassword: ''
});

// Reset password validation rules
const resetRules = reactive<FormRules>({
  password: [
    { required: true, message: '请输入新密码', trigger: 'blur' },
    { min: 6, message: '密码长度不能少于6个字符', trigger: 'blur' }
  ],
  confirmPassword: [
    { 
      required: true, 
      message: '请再次输入新密码', 
      trigger: 'blur' 
    },
    { 
      validator: (rule, value, callback) => {
        if (value !== resetForm.password) {
          callback(new Error('两次输入的密码不一致'));
        } else {
          callback();
        }
      }, 
      trigger: 'blur' 
    }
  ]
});

// Methods
// Fetch user list
const fetchUserList = async () => {
  loading.value = true;
  try {
    const params: UserQueryParams = {
      page: pagination.page,
      pageSize: pagination.pageSize
    };
    
    if (searchForm.username) params.username = searchForm.username;
    if (searchForm.name) params.name = searchForm.name;
    if (searchForm.departmentId) params.departmentId = searchForm.departmentId;
    if (searchForm.roleId) params.roleId = searchForm.roleId;
    if (searchForm.status) params.status = searchForm.status;
    
    const res = await getUserList(params);
    userList.value = res.list;
    pagination.total = res.total;
  } catch (error) {
    console.error('Failed to fetch user list:', error);
    ElMessage.error('获取用户列表失败');
  } finally {
    loading.value = false;
  }
};

// Fetch departments
const fetchDepartments = async () => {
  try {
    const res = await getDepartmentList();
    departments.value = res;
  } catch (error) {
    console.error('Failed to fetch departments:', error);
    ElMessage.error('获取部门列表失败');
  }
};

// Fetch roles
const fetchRoles = async () => {
  try {
    const res = await getRoleList();
    roles.value = res;
  } catch (error) {
    console.error('Failed to fetch roles:', error);
    ElMessage.error('获取角色列表失败');
  }
};

// Handle search
const handleSearch = () => {
  pagination.page = 1;
  fetchUserList();
};

// Reset search form
const resetSearch = () => {
  searchForm.username = '';
  searchForm.name = '';
  searchForm.departmentId = undefined;
  searchForm.roleId = undefined;
  searchForm.status = '';
  handleSearch();
};

// Handle pagination size change
const handleSizeChange = (val: number) => {
  pagination.pageSize = val;
  fetchUserList();
};

// Handle pagination page change
const handleCurrentChange = (val: number) => {
  pagination.page = val;
  fetchUserList();
};

// Handle create user
const handleCreate = () => {
  isEdit.value = false;
  form.id = 0;
  form.username = '';
  form.name = '';
  form.email = '';
  form.password = '';
  form.departmentId = undefined;
  form.roleId = undefined;
  form.status = 'active';
  dialogVisible.value = true;
};

// Handle edit user
const handleEdit = (row: User) => {
  isEdit.value = true;
  form.id = row.id;
  form.username = row.username;
  form.name = row.name;
  form.email = row.email;
  form.password = ''; // Clear password
  form.departmentId = row.departmentId;
  form.roleId = row.roleId;
  form.status = row.status;
  dialogVisible.value = true;
};

// Handle reset password
const handleResetPassword = (row: User) => {
  resetForm.password = '';
  resetForm.confirmPassword = '';
  currentUserId.value = row.id;
  resetPasswordVisible.value = true;
};

// Submit reset password
const submitResetPassword = async () => {
  if (!resetFormRef.value) return;
  
  await resetFormRef.value.validate(async (valid) => {
    if (valid) {
      try {
        await resetPassword(currentUserId.value, resetForm.password);
        ElMessage.success('密码重置成功');
        resetPasswordVisible.value = false;
      } catch (error) {
        console.error('Failed to reset password:', error);
        ElMessage.error('密码重置失败');
      }
    }
  });
};

// Handle user status change
const handleStatusChange = async (row: User) => {
  try {
    await updateUserStatus(row.id, row.status);
    ElMessage.success(`用户已${row.status === 'active' ? '启用' : '禁用'}`);
  } catch (error) {
    console.error('Failed to change user status:', error);
    ElMessage.error('更改用户状态失败');
    // Revert UI status on error
    row.status = row.status === 'active' ? 'disabled' : 'active';
  }
};

// Handle delete user
const handleDelete = (row: User) => {
  ElMessageBox.confirm(
    `确定要删除用户"${row.name}"吗？删除后无法恢复。`,
    '删除确认',
    {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning',
    }
  ).then(async () => {
    try {
      await deleteUser(row.id);
      ElMessage.success('删除成功');
      fetchUserList();
    } catch (error) {
      console.error('Failed to delete user:', error);
      ElMessage.error('删除用户失败');
    }
  });
};

// Submit form
const submitForm = async () => {
  if (!formRef.value) return;
  
  await formRef.value.validate(async (valid) => {
    if (valid) {
      try {
        if (isEdit.value) {
          const updateData = {
            name: form.name,
            email: form.email,
            departmentId: form.departmentId,
            roleId: form.roleId,
            status: form.status
          };
          await updateUser(form.id, updateData);
          ElMessage.success('用户更新成功');
        } else {
          await createUser({
            username: form.username,
            name: form.name,
            email: form.email,
            password: form.password,
            departmentId: form.departmentId!,
            roleId: form.roleId!,
            status: form.status
          });
          ElMessage.success('用户创建成功');
        }
        dialogVisible.value = false;
        fetchUserList();
      } catch (error) {
        console.error('Failed to save user:', error);
        ElMessage.error('保存用户失败');
      }
    }
  });
};

// Lifecycle hooks
onMounted(() => {
  fetchUserList();
  fetchDepartments();
  fetchRoles();
});
</script>

<style lang="scss" scoped>
.user-list-container {
  .filter-container {
    margin-bottom: 20px;
  }

  .table-container {
    .card-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
    }
  }

  .pagination-container {
    margin-top: 20px;
    display: flex;
    justify-content: flex-end;
  }
}
</style>