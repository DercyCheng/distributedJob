<template>
  <div class="department-list-container">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>部门列表</span>
          <el-button type="primary" size="small">新增部门</el-button>
        </div>
      </template>
      <el-table
        :data="departments"
        row-key="id"
        border
        :tree-props="{ children: 'children' }"
      >
        <el-table-column prop="name" label="部门名称" min-width="180" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="scope">
            <el-tag
              :type="scope.row.status === 'active' ? 'success' : 'danger'"
            >
              {{ scope.row.status === "active" ? "启用" : "禁用" }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="createdAt" label="创建时间" width="180" />
        <el-table-column prop="updatedAt" label="更新时间" width="180" />
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="scope">
            <el-button type="primary" link>编辑</el-button>
            <el-button type="primary" link>添加子部门</el-button>
            <el-button type="danger" link>删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref } from "vue";
import type { Department } from "@/api/department";

// This is a stub implementation to allow the build to succeed
const departments = ref<Department[]>([
  {
    id: 1,
    name: "总公司",
    parentId: null,
    parentName: "",
    status: "active",
    createdAt: "2025-01-01 00:00:00",
    updatedAt: "2025-01-01 00:00:00",
    children: [
      {
        id: 2,
        name: "技术部",
        parentId: 1,
        parentName: "总公司",
        status: "active",
        createdAt: "2025-01-01 00:00:00",
        updatedAt: "2025-01-01 00:00:00",
      },
    ],
  },
]);
</script>

<style lang="scss" scoped>
.department-list-container {
  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
}
</style>
