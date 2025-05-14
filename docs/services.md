# DistributedJob - 服务文档

## 快速开发与部署

DistributedJob 服务架构设计优先考虑了开发速度和部署效率，支持 10 天快速开发周期。所有服务组件均采用标准化接口设计、代码生成工具和预构建模板，使得开发团队能够在极短时间内完成功能实现。通过容器化技术和自动化脚本，系统可在几分钟内完成部署，支持敏捷迭代和快速交付。详细方法请参考 [快速开发指南](rapid_development.md)。

## 核心服务组件

1. **API 层** - 用于任务/部门/用户管理的 RESTful API 端点
2. **RPC 服务层** - 提供高性能内部服务通信机制
3. **调度引擎** - 处理任务调度和分发
4. **存储层** - 持久化任务配置和执行记录
5. **认证/权限模块** - 管理身份验证和授权
6. **任务执行器** - 执行 HTTP 和 gRPC 任务，支持重试
7. **历史记录管理器** - 记录和分析任务执行历史
8. **智能代理系统** - 提供 Agent 自主决策和执行能力
9. **模型通信层** - 实现符合 MCP 标准的 AI 模型交互
10. **知识检索系统** - 基于 RAG 技术的信息检索与生成

## API 文档

### API 概述

系统提供全面的 API 用于管理和监控分布式任务，以及新增的 AI 功能交互。

### 用户认证 API

- **POST /api/v1/auth/login** - 用户登录，获取访问令牌和刷新令牌
- **POST /api/v1/auth/refresh** - 使用刷新令牌获取新的访问令牌
- **POST /api/v1/auth/logout** - 用户登出，撤销当前会话的令牌

### 部门管理 API

- **GET /api/v1/departments** - 获取部门列表
- **POST /api/v1/departments** - 创建新部门
- **GET /api/v1/departments/{id}** - 获取部门详情
- **PUT /api/v1/departments/{id}** - 更新部门信息
- **DELETE /api/v1/departments/{id}** - 删除部门

### 用户管理 API

- **GET /api/v1/users** - 获取用户列表
- **POST /api/v1/users** - 创建新用户
- **GET /api/v1/users/{id}** - 获取用户详情
- **PUT /api/v1/users/{id}** - 更新用户信息
- **DELETE /api/v1/users/{id}** - 删除用户

### 角色与权限管理 API

- **GET /api/v1/roles** - 获取角色列表
- **POST /api/v1/roles** - 创建新角色
- **GET /api/v1/roles/{id}** - 获取角色详情
- **PUT /api/v1/roles/{id}** - 更新角色信息
- **DELETE /api/v1/roles/{id}** - 删除角色
- **GET /api/v1/permissions** - 获取权限列表

### 任务管理 API

- **GET /api/v1/tasks** - 获取任务列表
- **POST /api/v1/tasks** - 创建新任务
- **GET /api/v1/tasks/{id}** - 获取任务详情
- **PUT /api/v1/tasks/{id}** - 更新任务信息
- **DELETE /api/v1/tasks/{id}** - 删除任务
- **POST /api/v1/tasks/{id}/execute** - 手动执行任务

### 执行记录查询 API

- **GET /api/v1/records** - 获取执行记录列表
- **GET /api/v1/records/{id}** - 获取执行记录详情

### 健康检查与服务管理 API

- **GET /health** - 系统健康检查
- **GET /metrics** - 获取系统指标

### 智能代理 API

- **GET /api/v1/agents** - 获取智能代理列表
- **POST /api/v1/agents** - 创建新的智能代理
- **GET /api/v1/agents/{id}** - 获取智能代理详情
- **PUT /api/v1/agents/{id}** - 更新智能代理配置
- **DELETE /api/v1/agents/{id}** - 删除智能代理
- **POST /api/v1/agents/{id}/execute** - 指派智能代理执行任务
- **GET /api/v1/agents/{id}/status** - 获取智能代理状态

### MCP 模型交互 API

- **GET /api/v1/mcp/models** - 获取可用模型列表
- **POST /api/v1/mcp/chat** - 发送对话请求
- **POST /api/v1/mcp/stream-chat** - 流式对话请求
- **POST /api/v1/mcp/complete** - 文本补全请求
- **GET /api/v1/mcp/usage** - 获取模型使用统计

### RAG 检索增强生成 API

- **POST /api/v1/rag/documents** - 上传并索引文档
- **GET /api/v1/rag/documents** - 获取已索引文档列表
- **GET /api/v1/rag/documents/{id}** - 获取索引文档详情
- **DELETE /api/v1/rag/documents/{id}** - 删除索引文档
- **POST /api/v1/rag/query** - 提交 RAG 查询请求
- **POST /api/v1/rag/batch-query** - 提交批量 RAG 查询请求

### RPC 服务 API

系统内部组件通过 gRPC 协议进行通信，包括：

1. **认证服务** - 处理用户认证和令牌管理

   - `AuthService.Login` - 用户登录
   - `AuthService.Verify` - 验证令牌
   - `AuthService.Refresh` - 刷新令牌

2. **数据服务** - 提供数据访问和管理

   - `DataService.GetUser` - 获取用户信息
   - `DataService.ListTasks` - 列出任务
   - `DataService.GetTaskDetails` - 获取任务详情

3. **调度服务** - 任务调度和执行

   - `SchedulerService.ScheduleTask` - 调度任务
   - `SchedulerService.CancelTask` - 取消任务
   - `SchedulerService.GetTaskStatus` - 获取任务状态

4. **智能代理服务** - 智能代理管理和交互

   - `AgentService.CreateAgent` - 创建智能代理
   - `AgentService.GetAgent` - 获取智能代理信息
   - `AgentService.ExecuteTask` - 指派智能代理执行任务
   - `AgentService.GetStatus` - 获取智能代理状态

5. **MCP 服务** - 模型交互服务

   - `MCPService.Chat` - 发送对话请求
   - `MCPService.StreamChat` - 流式对话
   - `MCPService.GetModels` - 获取可用模型列表

6. **RAG 服务** - 检索增强生成服务

   - `RAGService.IndexDocument` - 索引文档
   - `RAGService.Query` - 提交查询
   - `RAGService.ListDocuments` - 列出已索引文档

## 令牌安全机制

### 令牌概述

系统使用基于 JWT 的身份验证和授权机制，实现安全访问控制。

### 双令牌机制

系统采用双令牌机制提高安全性：

1. **访问令牌 (Access Token)**
