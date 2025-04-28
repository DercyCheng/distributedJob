<template>
  <div class="role-list-container">
    <!-- Search and Filter Form -->
    <el-card shadow="never" class="filter-container">
      <el-form :model="searchForm" inline>
        <el-form-item label="角色名称">
          <el-input
            v-model="searchForm.name"
            placeholder="请输入角色名称"
            clearable
            @keyup.enter="handleSearch"
          />
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

    <!-- Role Table -->
    <el-card shadow="never" class="table-container">
      <template #header>
        <div class="card-header">
          <span>角色列表</span>
          <el-button type="primary" @click="handleCreate">创建角色</el-button>
        </div>
      </template>

      <el-table v-loading="loading" :data="roleList" border style="width: 100%">
        <el-table-column
          prop="name"
          label="角色名称"
          min-width="150"
          show-overflow-tooltip
        />
        <el-table-column
          prop="description"
          label="角色描述"
          min-width="200"
          show-overflow-tooltip
        />
        <el-table-column
          prop="status"
          label="状态"
          width="100"
        >
          <template #default="scope">
            <el-tag :type="scope.row.status === 'active' ? 'success' : 'info'">
              {{ scope.row.status === 'active' ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column
          prop="createdAt"
          label="创建时间"
          width="180"
          show-overflow-tooltip
        />
        <el-table-column
          label="操作"
          width="180"
          fixed="right"
        >
          <template #default="scope">
            <el-button type="primary" link @click="handleEdit(scope.row)">
              编辑
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

    <!-- Role Form Dialog -->
    <el-dialog
      v-model="dialogVisible"
      :title="isEdit ? '编辑角色' : '新建角色'"
      width="700px"
      append-to-body
    >
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="100px"
        label-position="right"
      >
        <el-form-item label="角色名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入角色名称" />
        </el-form-item>
        <el-form-item label="角色描述" prop="description">
          <el-input
            v-model="form.description"
            type="textarea"
            :rows="3"
            placeholder="请输入角色描述"
          />
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-radio-group v-model="form.status">
            <el-radio label="active">启用</el-radio>
            <el-radio label="disabled">禁用</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="权限分配" prop="permissions">
          <el-tree
            ref="permissionTreeRef"
            :data="permissionTree"
            :props="{
              label: 'name',
              children: 'children'
            }"
            node-key="id"
            show-checkbox
            default-expand-all
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="dialogVisible = false">取消</el-button>
          <el-button type="primary" @click="submitForm">确定</el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import type { FormInstance } from 'element-plus';
import type { ElTree } from 'element-plus';
// Import each function individually rather than using grouped imports
import { getRoleList } from '@/api/role';
import { getRoleById } from '@/api/role';
import { createRole } from '@/api/role';
import { updateRole } from '@/api/role';
import { deleteRole } from '@/api/role';
import { getPermissionTree } from '@/api/role';
import type { Role, Permission, RoleQueryParams } from '@/api/role';

// Data
const loading = ref(false);
const roleList = ref<Role[]>([]);
const permissionTree = ref<Permission[]>([]);

// Form refs
const formRef = ref<FormInstance>();
const permissionTreeRef = ref<InstanceType<typeof ElTree>>();

// Search form
const searchForm = reactive<RoleQueryParams>({
  page: 1,
  pageSize: 10,
  name: '',
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

// Form data
const form = reactive({
  id: 0,
  name: '',
  description: '',
  status: 'active' as 'active' | 'disabled',
  permissions: [] as number[]
});

// Form validation rules
const rules = {
  name: [
    { required: true, message: '请输入角色名称', trigger: 'blur' },
    { min: 2, max: 30, message: '长度在 2 到 30 个字符', trigger: 'blur' }
  ],
  description: [
    { required: true, message: '请输入角色描述', trigger: 'blur' },
    { max: 200, message: '长度不能超过 200 个字符', trigger: 'blur' }
  ],
  status: [
    { required: true, message: '请选择角色状态', trigger: 'change' }
  ]
};

// Methods
// Fetch role list
const fetchRoleList = async () => {
  loading.value = true;
  try {
    const params: RoleQueryParams = {
      page: pagination.page,
      pageSize: pagination.pageSize
    };
    
    if (searchForm.name) params.name = searchForm.name;
    if (searchForm.status) params.status = searchForm.status;
    
    const res = await getRoleList(params);
    roleList.value = res.list;
    pagination.total = res.total;
  } catch (error) {
    console.error('Failed to fetch role list:', error);
    ElMessage.error('获取角色列表失败');
  } finally {
    loading.value = false;
  }
};

// Fetch permission tree
const fetchPermissionTree = async () => {
  try {
    permissionTree.value = await getPermissionTree();
  } catch (error) {
    console.error('Failed to fetch permission tree:', error);
    ElMessage.error('获取权限列表失败');
  }
};

// Handle search
const handleSearch = () => {
  pagination.page = 1;
  fetchRoleList();
};

// Reset search form
const resetSearch = () => {
  searchForm.name = '';
  searchForm.status = '';
  handleSearch();
};

// Handle pagination size change
const handleSizeChange = (val: number) => {
  pagination.pageSize = val;
  fetchRoleList();
};

// Handle pagination page change
const handleCurrentChange = (val: number) => {
  pagination.page = val;
  fetchRoleList();
};

// Handle create role
const handleCreate = () => {
  isEdit.value = false;
  form.id = 0;
  form.name = '';
  form.description = '';
  form.status = 'active';
  form.permissions = [];
  
  if (permissionTreeRef.value) {
    permissionTreeRef.value.setCheckedKeys([]);
  }
  
  dialogVisible.value = true;
};

// Handle edit role
const handleEdit = async (row: Role) => {
  isEdit.value = true;
  form.id = row.id;
  
  // Fetch complete role details with permissions
  try {
    const roleDetail = await getRoleById(row.id);
    form.name = roleDetail.name;
    form.description = roleDetail.description;
    form.status = roleDetail.status;
    form.permissions = roleDetail.permissions.map(p => p.id);
    
    // Set tree checked state
    if (permissionTreeRef.value) {
      permissionTreeRef.value.setCheckedKeys(form.permissions);
    }
    
    dialogVisible.value = true;
  } catch (error) {
    console.error('Failed to fetch role details:', error);
    ElMessage.error('获取角色详情失败');
  }
};

// Handle delete role
const handleDelete = (row: Role) => {
  ElMessageBox.confirm(
    `确定要删除角色"${row.name}"吗？删除后无法恢复。`,
    '删除确认',
    {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning',
    }
  ).then(async () => {
    try {
      await deleteRole(row.id);
      ElMessage.success('删除成功');
      fetchRoleList();
    } catch (error) {
      console.error('Failed to delete role:', error);
      ElMessage.error('删除角色失败');
    }
  });
};

// Submit form
const submitForm = async () => {
  if (!formRef.value || !permissionTreeRef.value) return;
  
  await formRef.value.validate(async (valid) => {
    if (valid) {
      try {
        // Get selected permissions
        const checkedKeys = permissionTreeRef.value!.getCheckedKeys(false) as number[];
        const halfCheckedKeys = permissionTreeRef.value!.getHalfCheckedKeys() as number[];
        const allCheckedKeys = [...checkedKeys, ...halfCheckedKeys];
        
        // Prepare data
        const roleData = {
          name: form.name,
          description: form.description,
          status: form.status,
          permissions: allCheckedKeys.map(id => ({ id }))
        };
        
        if (isEdit.value) {
          await updateRole(form.id, roleData);
          ElMessage.success('角色更新成功');
        } else {
          await createRole(roleData);
          ElMessage.success('角色创建成功');
        }
        dialogVisible.value = false;
        fetchRoleList();
      } catch (error) {
        console.error('Failed to save role:', error);
        ElMessage.error('保存角色失败');
      }
    }
  });
};

// Lifecycle hooks
onMounted(() => {
  fetchRoleList();
  fetchPermissionTree();
});
</script>

<style lang="scss" scoped>
.role-list-container {
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