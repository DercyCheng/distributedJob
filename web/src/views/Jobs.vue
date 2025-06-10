<template>
  <div class="jobs">
    <div class="page-header">
      <h1>任务管理</h1>
      <el-button type="primary" @click="showCreateDialog = true">
        <el-icon><Plus /></el-icon>
        创建任务
      </el-button>
    </div>

    <!-- 搜索筛选 -->
    <el-card class="search-card">
      <el-form :model="searchForm" inline>
        <el-form-item label="任务名称">
          <el-input
            v-model="searchForm.keyword"
            placeholder="请输入任务名称"
            clearable
            @clear="loadJobs"
          />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="searchForm.enabled" placeholder="请选择状态" clearable>
            <el-option label="全部" :value="null" />
            <el-option label="启用" :value="true" />
            <el-option label="禁用" :value="false" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="loadJobs">搜索</el-button>
          <el-button @click="resetSearch">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 任务列表 -->
    <el-card>
      <el-table :data="jobs" stripe v-loading="loading">
        <el-table-column prop="name" label="任务名称" />
        <el-table-column prop="description" label="描述" />
        <el-table-column prop="cron" label="调度规则" />
        <el-table-column prop="command" label="执行命令" />
        <el-table-column prop="enabled" label="状态">
          <template #default="{ row }">
            <el-switch
              v-model="row.enabled"
              @change="toggleJobStatus(row)"
            />
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间">
          <template #default="{ row }">
            {{ formatTime(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200">
          <template #default="{ row }">
            <el-button size="small" @click="triggerJob(row)">
              <el-icon><VideoPlay /></el-icon>
              执行
            </el-button>
            <el-button size="small" type="primary" @click="editJob(row)">
              <el-icon><Edit /></el-icon>
              编辑
            </el-button>
            <el-button size="small" type="danger" @click="deleteJob(row)">
              <el-icon><Delete /></el-icon>
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div class="pagination">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.size"
          :page-sizes="[10, 20, 50, 100]"
          :total="pagination.total"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="loadJobs"
          @current-change="loadJobs"
        />
      </div>
    </el-card>

    <!-- 创建/编辑任务对话框 -->
    <el-dialog
      v-model="showCreateDialog"
      :title="editingJob ? '编辑任务' : '创建任务'"
      width="600px"
    >
      <el-form :model="jobForm" :rules="jobRules" ref="jobFormRef" label-width="100px">
        <el-form-item label="任务名称" prop="name">
          <el-input v-model="jobForm.name" placeholder="请输入任务名称" />
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input
            v-model="jobForm.description"
            type="textarea"
            placeholder="请输入任务描述"
            :rows="3"
          />
        </el-form-item>
        <el-form-item label="调度规则" prop="cron">
          <el-input v-model="jobForm.cron" placeholder="如: 0 */5 * * * *" />
          <div class="form-help">
            Cron表达式格式: 秒 分 时 日 月 周
          </div>
        </el-form-item>
        <el-form-item label="执行命令" prop="command">
          <el-input v-model="jobForm.command" placeholder="如: echo 'Hello World'" />
        </el-form-item>
        <el-form-item label="超时时间" prop="timeout">
          <el-input-number
            v-model="jobForm.timeout"
            :min="1"
            :max="3600"
            placeholder="秒"
          />
          <span class="input-suffix">秒</span>
        </el-form-item>
        <el-form-item label="重试次数" prop="retry_attempts">
          <el-input-number
            v-model="jobForm.retry_attempts"
            :min="0"
            :max="10"
          />
        </el-form-item>
        <el-form-item label="参数设置">
          <div class="params-editor">
            <div
              v-for="(param, index) in jobForm.params"
              :key="index"
              class="param-item"
            >
              <el-input
                v-model="param.key"
                placeholder="参数名"
                style="width: 200px"
              />
              <el-input
                v-model="param.value"
                placeholder="参数值"
                style="width: 200px; margin-left: 10px"
              />
              <el-button
                type="danger"
                size="small"
                @click="removeParam(index)"
                style="margin-left: 10px"
              >
                删除
              </el-button>
            </div>
            <el-button type="primary" size="small" @click="addParam">
              添加参数
            </el-button>
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreateDialog = false">取消</el-button>
        <el-button type="primary" @click="saveJob">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, VideoPlay, Edit, Delete } from '@element-plus/icons-vue'
import { getJobs, createJob, updateJob, deleteJob as deleteJobApi, triggerJob as triggerJobApi } from '@/api/jobs'

const jobs = ref([])
const loading = ref(false)
const showCreateDialog = ref(false)
const editingJob = ref(null)
const jobFormRef = ref()

const searchForm = reactive({
  keyword: '',
  enabled: null
})

const pagination = reactive({
  page: 1,
  size: 10,
  total: 0
})

const jobForm = reactive({
  name: '',
  description: '',
  cron: '',
  command: '',
  timeout: 300,
  retry_attempts: 3,
  params: []
})

const jobRules = {
  name: [{ required: true, message: '请输入任务名称', trigger: 'blur' }],
  cron: [{ required: true, message: '请输入调度规则', trigger: 'blur' }],
  command: [{ required: true, message: '请输入执行命令', trigger: 'blur' }]
}

onMounted(() => {
  loadJobs()
})

const loadJobs = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.page,
      size: pagination.size,
      keyword: searchForm.keyword,
      enabled: searchForm.enabled
    }
    const response = await getJobs(params)
    jobs.value = response.data.jobs || []
    pagination.total = response.data.total || 0
  } catch (error) {
    ElMessage.error('加载任务列表失败')
  } finally {
    loading.value = false
  }
}

const resetSearch = () => {
  searchForm.keyword = ''
  searchForm.enabled = null
  pagination.page = 1
  loadJobs()
}

const editJob = (job) => {
  editingJob.value = job
  Object.assign(jobForm, {
    name: job.name,
    description: job.description,
    cron: job.cron,
    command: job.command,
    timeout: job.timeout,
    retry_attempts: job.retry_attempts,
    params: Object.entries(job.params || {}).map(([key, value]) => ({ key, value }))
  })
  showCreateDialog.value = true
}

const saveJob = async () => {
  if (!jobFormRef.value) return
  
  try {
    await jobFormRef.value.validate()
    
    const params = {}
    jobForm.params.forEach(param => {
      if (param.key && param.value) {
        params[param.key] = param.value
      }
    })

    const jobData = {
      name: jobForm.name,
      description: jobForm.description,
      cron: jobForm.cron,
      command: jobForm.command,
      timeout: jobForm.timeout,
      retry_attempts: jobForm.retry_attempts,
      params
    }

    if (editingJob.value) {
      await updateJob(editingJob.value.id, jobData)
      ElMessage.success('任务更新成功')
    } else {
      await createJob(jobData)
      ElMessage.success('任务创建成功')
    }
    
    showCreateDialog.value = false
    resetJobForm()
    loadJobs()
  } catch (error) {
    ElMessage.error(editingJob.value ? '任务更新失败' : '任务创建失败')
  }
}

const resetJobForm = () => {
  Object.assign(jobForm, {
    name: '',
    description: '',
    cron: '',
    command: '',
    timeout: 300,
    retry_attempts: 3,
    params: []
  })
  editingJob.value = null
}

const addParam = () => {
  jobForm.params.push({ key: '', value: '' })
}

const removeParam = (index) => {
  jobForm.params.splice(index, 1)
}

const toggleJobStatus = async (job) => {
  try {
    await updateJob(job.id, { ...job, enabled: job.enabled })
    ElMessage.success(`任务已${job.enabled ? '启用' : '禁用'}`)
  } catch (error) {
    job.enabled = !job.enabled // 回滚状态
    ElMessage.error('状态更新失败')
  }
}

const triggerJob = async (job) => {
  try {
    await triggerJobApi(job.id)
    ElMessage.success('任务已手动触发')
  } catch (error) {
    ElMessage.error('任务触发失败')
  }
}

const deleteJob = async (job) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除任务 "${job.name}" 吗？`,
      '确认删除',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    await deleteJobApi(job.id)
    ElMessage.success('任务删除成功')
    loadJobs()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('任务删除失败')
    }
  }
}

const formatTime = (time) => {
  return new Date(time).toLocaleString()
}
</script>

<style scoped>
.jobs {
  padding: 24px;
  min-height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
  background: rgba(255, 255, 255, 0.95);
  padding: 20px 24px;
  border-radius: 16px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.1);
  backdrop-filter: blur(10px);
  border: 1px solid rgba(255, 255, 255, 0.2);
}

.page-header h1 {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  margin: 0;
  font-size: 28px;
  font-weight: 600;
}

.search-card {
  margin-bottom: 24px;
  background: rgba(255, 255, 255, 0.95);
  border-radius: 16px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.1);
  backdrop-filter: blur(10px);
  border: 1px solid rgba(255, 255, 255, 0.2);
  overflow: hidden;
}

.search-card :deep(.el-card__body) {
  padding: 24px;
}

.jobs :deep(.el-card) {
  background: rgba(255, 255, 255, 0.95);
  border-radius: 16px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.1);
  backdrop-filter: blur(10px);
  border: 1px solid rgba(255, 255, 255, 0.2);
  overflow: hidden;
}

.jobs :deep(.el-card__body) {
  padding: 24px;
}

.jobs :deep(.el-table) {
  background: transparent;
  border-radius: 12px;
  overflow: hidden;
}

.jobs :deep(.el-table__header) {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.jobs :deep(.el-table__header th) {
  background: transparent;
  color: white;
  font-weight: 600;
  border: none;
  padding: 16px 12px;
}

.jobs :deep(.el-table__body tr) {
  transition: all 0.3s ease;
}

.jobs :deep(.el-table__body tr:hover) {
  background: rgba(102, 126, 234, 0.1);
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.jobs :deep(.el-table__body td) {
  border: none;
  padding: 16px 12px;
}

.jobs :deep(.el-button) {
  border-radius: 8px;
  transition: all 0.3s ease;
  font-weight: 500;
}

.jobs :deep(.el-button:hover) {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.jobs :deep(.el-button--primary) {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border: none;
}

.jobs :deep(.el-button--danger) {
  background: linear-gradient(135deg, #ff6b6b 0%, #ee5a52 100%);
  border: none;
}

.jobs :deep(.el-switch.is-checked .el-switch__core) {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-color: #667eea;
}

.pagination {
  margin-top: 24px;
  text-align: right;
  padding: 20px;
  background: rgba(255, 255, 255, 0.95);
  border-radius: 16px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.1);
  backdrop-filter: blur(10px);
  border: 1px solid rgba(255, 255, 255, 0.2);
}

.jobs :deep(.el-pagination) {
  justify-content: flex-end;
}

.jobs :deep(.el-pagination .el-pager li.is-active) {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  border-radius: 6px;
}

.jobs :deep(.el-dialog) {
  border-radius: 16px;
  overflow: hidden;
  box-shadow: 0 20px 40px rgba(0, 0, 0, 0.15);
}

.jobs :deep(.el-dialog__header) {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  padding: 24px;
}

.jobs :deep(.el-dialog__title) {
  color: white;
  font-weight: 600;
  font-size: 18px;
}

.jobs :deep(.el-dialog__body) {
  padding: 24px;
}

.jobs :deep(.el-form-item__label) {
  font-weight: 500;
  color: #2c3e50;
}

.form-help {
  font-size: 12px;
  color: #67c23a;
  margin-top: 5px;
  font-style: italic;
}

.params-editor {
  border: 2px solid #e1f3d8;
  border-radius: 12px;
  padding: 16px;
  background: linear-gradient(135deg, #f0f9ff 0%, #e0f2fe 100%);
}

.param-item {
  display: flex;
  align-items: center;
  margin-bottom: 12px;
  padding: 8px;
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
}

.param-item:last-child {
  margin-bottom: 0;
}

.input-suffix {
  margin-left: 10px;
  color: #67c23a;
  font-weight: 500;
}

.jobs :deep(.el-input__wrapper) {
  border-radius: 8px;
  transition: all 0.3s ease;
}

.jobs :deep(.el-input__wrapper:hover) {
  box-shadow: 0 2px 8px rgba(102, 126, 234, 0.2);
}

.jobs :deep(.el-input__wrapper.is-focus) {
  box-shadow: 0 0 0 2px rgba(102, 126, 234, 0.2);
  border-color: #667eea;
}

.jobs :deep(.el-select .el-input__wrapper) {
  border-radius: 8px;
}

.jobs :deep(.el-textarea__inner) {
  border-radius: 8px;
  transition: all 0.3s ease;
}

.jobs :deep(.el-textarea__inner:hover) {
  box-shadow: 0 2px 8px rgba(102, 126, 234, 0.2);
}

.jobs :deep(.el-textarea__inner:focus) {
  box-shadow: 0 0 0 2px rgba(102, 126, 234, 0.2);
  border-color: #667eea;
}

@media (max-width: 768px) {
  .jobs {
    padding: 16px;
  }
  
  .page-header {
    flex-direction: column;
    gap: 16px;
    text-align: center;
  }
  
  .page-header h1 {
    font-size: 24px;
  }
  
  .jobs :deep(.el-table) {
    font-size: 14px;
  }
  
  .jobs :deep(.el-button) {
    padding: 8px 12px;
    font-size: 12px;
  }
}
</style>
