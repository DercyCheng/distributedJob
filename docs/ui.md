# DistributedJob - 前端开发文档

## 前端快速开发

DistributedJob 前端开发遵循 10 天快速开发周期，通过组件化设计、自动化工具和预构建模板，实现了高效的开发流程。前端团队使用 Vue 3、Vite 和 Element Plus 等现代技术栈，结合定制的代码生成工具和组件库，将常规前端开发时间缩短 60%。完整的前端开发可在 5 天内完成，包括基础框架搭建、页面实现和后端集成。详细方法请参考 [快速开发指南](rapid_development.md)。

## 前端开发

### 技术栈

DistributedJob 前端采用现代化的技术栈：

- **Vue 3** - 核心框架，采用 Composition API
- **TypeScript** - 提供类型安全
- **Vite** - 构建工具，提供快速的开发体验
- **Pinia** - 状态管理库
- **Vue Router** - 路由管理
- **Axios** - HTTP 客户端
- **Element Plus** - UI 组件库
- **SCSS** - CSS 预处理器
- **ESLint** - 代码质量工具
- **Vitest** - 单元测试框架

### 前端项目结构

```
web-ui/
├── public/                # 静态资源
├── src/
│   ├── api/              # API 客户端
│   │   ├── agent.ts      # 智能代理相关 API
│   │   ├── auth.ts       # 认证相关 API
│   │   ├── department.ts # 部门相关 API
│   │   ├── http.ts       # HTTP 客户端配置
│   │   ├── mcp.ts        # MCP 相关 API
│   │   ├── rag.ts        # RAG 相关 API
│   │   ├── record.ts     # 记录相关 API
│   │   ├── role.ts       # 角色相关 API
│   │   ├── task.ts       # 任务相关 API
│   │   └── user.ts       # 用户相关 API
│   ├── assets/           # 静态资源
│   │   └── styles/       # 样式文件
│   │       └── main.scss # 主样式文件
│   ├── components/       # Vue 组件
│   │   ├── agent/        # 智能代理相关组件
│   │   │   ├── AgentCard.vue        # 代理卡片组件
│   │   │   ├── AgentConfigForm.vue  # 代理配置表单
│   │   │   ├── AgentExecutor.vue    # 代理执行组件
│   │   │   └── AgentStatus.vue      # 代理状态组件
│   │   ├── chat/         # 对话相关组件
│   │   │   ├── ChatBox.vue          # 对话框组件
│   │   │   ├── MessageBubble.vue    # 消息气泡组件
│   │   │   └── StreamingText.vue    # 流式文本组件
│   │   ├── layout/       # 布局组件
│   │   │   ├── AppLink.vue      # 应用链接组件
│   │   │   ├── AppMain.vue      # 主内容组件
│   │   │   ├── Breadcrumb.vue   # 面包屑组件
│   │   │   └── Layout.vue       # 布局容器组件
│   │   ├── rag/          # RAG 相关组件
│   │   │   ├── DocumentUploader.vue # 文档上传组件
│   │   │   ├── QueryForm.vue        # 查询表单组件
│   │   │   ├── ResultDisplay.vue    # 结果展示组件
│   │   │   └── SourceViewer.vue     # 来源查看组件
│   │   └── workflow/     # 工作流相关组件
│   ├── router/           # 路由配置
│   │   └── index.ts      # 路由定义
│   ├── store/            # 状态管理
│   │   ├── index.ts      # 主 store 配置
│   │   └── modules/      # 模块化 store
│   │       ├── agent.ts  # 智能代理状态管理
│   │       ├── auth.ts   # 认证状态管理
│   │       ├── mcp.ts    # MCP 状态管理
│   │       ├── rag.ts    # RAG 状态管理
│   │       └── task.ts   # 任务状态管理
│   ├── types/            # TypeScript 类型定义
│   │   ├── agent.ts      # 智能代理类型定义
│   │   ├── mcp.ts        # MCP 类型定义
│   │   └── rag.ts        # RAG 类型定义
│   ├── utils/            # 工具函数
│   │   ├── token.ts      # 令牌相关工具
│   │   └── validate.ts   # 表单验证工具
│   ├── views/            # 页面组件
│   │   ├── agent/        # 智能代理页面
│   │   │   ├── AgentCreate.vue     # 创建代理页面
│   │   │   ├── AgentDetail.vue     # 代理详情页面
│   │   │   ├── AgentList.vue       # 代理列表页面
│   │   │   └── AgentWorkspace.vue  # 代理工作空间页面
│   │   ├── ai-chat/      # AI 对话页面
│   │   │   ├── ChatDashboard.vue   # 对话仪表盘
│   │   │   └── ChatInterface.vue   # 对话界面
│   │   ├── auth/         # 认证相关页面
│   │   ├── dashboard/    # 仪表盘页面
│   │   ├── department/   # 部门管理页面
│   │   ├── rag/          # RAG 相关页面
│   │   │   ├── DocumentLibrary.vue # 文档库页面
│   │   │   ├── QueryInterface.vue  # 查询界面页面
│   │   │   └── SourceManagement.vue # 来源管理页面
│   │   ├── record/       # 记录查询页面
│   │   ├── role/         # 角色管理页面
│   │   ├── task/         # 任务管理页面
│   │   ├── user/         # 用户管理页面
│   │   └── workflow/     # 工作流管理页面
│   ├── App.vue           # 根组件
│   ├── env.d.ts          # 环境变量类型定义
│   └── main.ts           # 应用入口
├── index.html            # HTML 入口文件
├── package.json          # 依赖和脚本配置
├── tsconfig.json         # TypeScript 配置
├── tsconfig.node.json    # 节点特定 TypeScript 配置
└── vite.config.ts        # Vite 配置
```

### 开发指南

#### 1. 环境准备

确保已安装以下工具：

- Node.js 18+
- npm 8+ 或 Yarn 1.22+

#### 2. 安装依赖

```bash
cd web-ui
npm install
```

#### 3. 开发模式

```bash
npm run dev
```

这将启动开发服务器，默认访问地址为 http://localhost:3000。

#### 4. 构建生产版本

```bash
npm run build
```

### AI 功能前端开发指南

#### Agent 智能代理

Agent 智能代理前端提供了直观的界面来创建、配置和监控智能代理：

1. **创建代理**：通过 `AgentCreate.vue` 页面，用户可以配置代理名称、描述、使用的模型以及可用工具集。

2. **代理列表**：`AgentList.vue` 提供所有代理的概览，展示其状态和最近活动。

3. **代理详情**：`AgentDetail.vue` 展示代理的详细信息、历史执行记录和性能指标。

4. **工作空间**：`AgentWorkspace.vue` 是与代理交互的主要界面，支持任务指派和结果查看。

代理组件示例：

```vue
<!-- AgentCard.vue -->
<template>
  <el-card class="agent-card" :body-style="{ padding: '0px' }">
    <div class="agent-header">
      <div class="agent-avatar">
        <el-avatar :icon="Robot" />
      </div>
      <div class="agent-info">
        <h3>{{ agent.name }}</h3>
        <span class="agent-status" :class="statusClass">
          {{ statusText }}
        </span>
      </div>
    </div>
    <div class="agent-body">
      <p>{{ agent.description }}</p>
      <div class="agent-tools">
        <el-tag v-for="tool in agent.tools" :key="tool" size="small">
          {{ tool }}
        </el-tag>
      </div>
    </div>
    <div class="agent-footer">
      <el-button type="primary" size="small" @click="$emit('execute')">
        执行任务
      </el-button>
      <el-button type="info" size="small" @click="$emit('detail')">
        查看详情
      </el-button>
    </div>
  </el-card>
</template>
```

#### MCP 模型交互

MCP 前端组件提供流畅的 AI 模型交互体验：

1. **对话界面**：`ChatInterface.vue` 提供类似聊天的交互界面，支持文本、代码和其他格式消息。

2. **流式响应**：`StreamingText.vue` 组件实现流式文本展示，提供打字机效果的实时响应。

3. **模型选择器**：允许用户选择不同的 AI 模型以满足不同任务需求。

聊天组件示例：

```vue
<!-- ChatBox.vue -->
<template>
  <div class="chat-container">
    <div class="messages-container" ref="messagesContainer">
      <message-bubble
        v-for="(msg, index) in messages"
        :key="index"
        :message="msg"
        :is-user="msg.role === 'user'"
      />
      <div v-if="isTyping" class="typing-indicator">
        <span></span>
        <span></span>
        <span></span>
      </div>
    </div>
    <div class="input-container">
      <el-input
        v-model="userInput"
        type="textarea"
        :rows="3"
        placeholder="输入消息..."
        @keydown.enter.exact.prevent="sendMessage"
      />
      <el-button
        type="primary"
        @click="sendMessage"
        :disabled="!userInput.trim() || isSending"
      >
        发送
      </el-button>
    </div>
  </div>
</template>
```

#### RAG 检索增强生成

RAG 前端组件让用户能够轻松管理知识库和执行查询：

1. **文档上传**：`DocumentUploader.vue` 提供文档上传和索引功能，支持多种格式。

2. **文档库管理**：`DocumentLibrary.vue` 管理所有已索引文档，支持查看、更新和删除操作。

3. **查询界面**：`QueryInterface.vue` 提供问答交互界面，展示查询结果及其来源。

4. **来源查看器**：`SourceViewer.vue` 让用户可以深入了解结果的信息来源。

文档上传组件示例：

```vue
<!-- DocumentUploader.vue -->
<template>
  <div class="document-uploader">
    <el-upload
      class="upload-container"
      drag
      multiple
      :http-request="customUpload"
      :file-list="fileList"
      :before-upload="beforeUpload"
    >
      <el-icon class="el-icon--upload"><upload-filled /></el-icon>
      <div class="el-upload__text">拖拽文件至此处或 <em>点击上传</em></div>
      <template #tip>
        <div class="el-upload__tip">
          支持 PDF、Word、文本、Markdown 等格式文件
        </div>
      </template>
    </el-upload>

    <div class="processing-files" v-if="processingFiles.length > 0">
      <h3>处理中的文件：</h3>
      <el-progress
        v-for="file in processingFiles"
        :key="file.name"
        :percentage="file.progress"
        :status="file.status"
      >
        <span>{{ file.name }}</span>
      </el-progress>
    </div>
  </div>
</template>
```

### 组件交互示例

以下是 Agent、MCP 和 RAG 组件如何协同工作的示例：

```typescript
// agent 工作空间的业务逻辑
import { ref, reactive } from "vue";
import { useAgentStore } from "@/store/modules/agent";
import { useRagStore } from "@/store/modules/rag";
import { AgentService } from "@/api/agent";
import { RagService } from "@/api/rag";

export function useAgentWorkspace(agentId: string) {
  const agentStore = useAgentStore();
  const ragStore = useRagStore();

  const agent = ref(null);
  const taskInput = ref("");
  const isProcessing = ref(false);
  const result = reactive({
    content: "",
    sources: [],
  });

  // 加载代理详情
  const loadAgent = async () => {
    agent.value = await agentStore.getAgentById(agentId);
  };

  // 执行代理任务
  const executeTask = async () => {
    if (!taskInput.value.trim() || isProcessing.value) return;

    isProcessing.value = true;

    try {
      // 1. 首先使用 RAG 增强查询
      const ragResult = await ragStore.query(taskInput.value);

      // 2. 将增强后的查询发送给 Agent
      const response = await AgentService.executeTask(agentId, {
        task: taskInput.value,
        context: ragResult.sources,
      });

      result.content = response.result;
      result.sources = response.sources;

      // 3. 添加到任务历史
      agentStore.addTaskHistory(agentId, {
        input: taskInput.value,
        output: response.result,
        timestamp: new Date(),
      });
    } finally {
      isProcessing.value = false;
    }
  };

  return {
    agent,
    taskInput,
    isProcessing,
    result,
    loadAgent,
    executeTask,
  };
}
```

通过这种方式，我们的前端界面能够无缝集成 Agent、MCP 和 RAG 功能，提供直观且功能丰富的用户体验。
