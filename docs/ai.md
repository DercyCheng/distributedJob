# DistributedJob - AI 功能文档

## AI 能力概述

DistributedJob 现已集成强大的 AI 能力，包括 Agent（智能代理）、MCP（模型上下文协议）和 RAG（检索增强生成）技术，显著增强了系统的智能化水平和自动化程度。这些 AI 功能采用模块化设计和预构建组件，支持在 10 天快速开发周期内完成集成，大大加速了项目交付速度。AI 能力集成仅需 1 天时间，可与现有系统无缝对接，立即提升产品价值。

## Agent 智能代理

### 概述

Agent 智能代理是一种可自主执行任务、做出决策并与环境交互的 AI 系统。在 DistributedJob 中，我们实现了可配置、可扩展的智能代理框架，能够处理复杂的工作流并优化调度决策。

### 核心功能

1. **自主决策** - 基于当前系统状态和历史数据做出智能调度决策
2. **任务分解** - 将复杂任务分解为可管理的子任务
3. **异常处理** - 智能检测和响应系统异常情况
4. **资源优化** - 动态优化资源分配，提高系统整体效率
5. **智能任务调度** - 利用 Agent 智能代理根据系统负载、任务优先级和历史数据进行优化调度

### 架构设计

```
internal/
├── agent/                # Agent 智能代理模块
│   ├── core/             # Agent 核心实现
│   │   ├── agent.go      # 基础 Agent 定义
│   │   ├── executor.go   # 执行器实现
│   │   ├── planner.go    # 计划生成器
│   │   └── memory.go     # Agent 记忆管理
│   ├── tools/            # Agent 可用工具集
│   │   ├── scheduler_tool.go # 调度工具
│   │   ├── data_tool.go      # 数据操作工具
│   │   └── system_tool.go    # 系统工具
│   ├── config/           # Agent 配置
│   │   └── agent_config.go # Agent 配置定义
│   └── types/            # 类型定义
│       ├── tool.go       # 工具接口定义
│       └── memory.go     # 记忆类型定义
```

### API 接口

- **POST /api/v1/agent/create** - 创建新的智能代理实例
- **GET /api/v1/agent/{id}** - 获取智能代理详情
- **POST /api/v1/agent/{id}/execute** - 指派智能代理执行任务
- **GET /api/v1/agent/{id}/status** - 查询智能代理执行状态

### 配置示例

```yaml
agents:
  - name: "scheduler-agent"
    description: "负责优化任务调度的智能代理"
    model: "deepseekv3-7b" # 采用 deepseekv3 低尺寸模型
    tools:
      - scheduler-tool
      - system-tool
    memory:
      type: "vector"
      capacity: 1000

  - name: "monitoring-agent"
    description: "系统监控智能代理"
    model: "qwen3-7b" # 采用 qwen3 模型
    tools:
      - data-tool
      - alert-tool
    memory:
      type: "buffer"
      capacity: 500
```

### 模型说明

本项目采用以下 AI 模型：

- **Qwen3**: 阿里通义千问第三代模型，支持多种尺寸，用于通用任务处理
- **DeepseekV3**: 深度求索第三代模型，使用低尺寸版本，优化部署资源需求

所有本地模型通过 docker-compose 进行部署，使用阿里云镜像源加速拉取。

```

## MCP (Model Context Protocol)

### 概述

MCP（模型上下文协议）是一种标准化的协议，用于与各种 AI 模型交互，提供统一的接口以简化多模型集成和交互流程。

### 核心组件

1. **模型连接器** - 支持连接到各种 AI 模型服务（如 OpenAI、Anthropic、本地模型等）
2. **上下文管理** - 高效管理和传递上下文信息，优化模型响应质量
3. **消息格式化** - 统一的消息格式，简化前后端通信
4. **流式处理** - 支持 AI 模型的流式输出处理

### 架构设计

```

internal/
├── mcp/ # MCP 模块
│ ├── client/ # MCP 客户端实现
│ │ ├── openai.go # OpenAI 连接器
│ │ ├── anthropic.go # Anthropic 连接器
│ │ └── local.go # 本地模型连接器
│ ├── context/ # 上下文管理
│ │ ├── manager.go # 上下文管理器
│ │ └── window.go # 滑动窗口实现
│ ├── protocol/ # 协议定义
│ │ ├── message.go # 消息结构定义
│ │ └── session.go # 会话管理
│ └── stream/ # 流处理
│ └── handler.go # 流处理器

````

### API 接口

- **POST /api/v1/mcp/chat** - 发送对话请求到指定模型
- **GET /api/v1/mcp/models** - 获取可用 AI 模型列表
- **POST /api/v1/mcp/stream-chat** - 流式对话接口
- **POST /api/v1/mcp/complete** - 文本补全接口

### 使用示例

```go
// 创建 MCP 客户端 (使用本地 DeepseekV3 模型)
mcpClient := mcp.NewClient(mcp.Config{
    Provider: "local",
    Model:    "deepseekv3-7b",
    LocalConfig: mcp.LocalConfig{
        APIServerURL: "http://llm-server:8080",
        Models: []mcp.LocalModelConfig{
            {Name: "deepseekv3-7b", Path: "deepseekv3-7b-q5_k_m.gguf"},
        },
    },
})

// 创建聊天请求
chatRequest := &mcp.ChatRequest{
    Messages: []mcp.Message{
        {Role: "system", Content: "你是一个任务调度专家。"},
        {Role: "user", Content: "如何优化并发任务的执行顺序？"},
    },
    MaxTokens: 500,
}

// 获取响应
response, err := mcpClient.Chat(context.Background(), chatRequest)
if err != nil {
    log.Fatalf("Chat error: %v", err)
}

fmt.Println("AI Response:", response.Content)
````

## RAG (检索增强生成)

### 概述

RAG（检索增强生成）技术通过在生成响应前检索相关信息来增强 AI 模型的输出质量，使模型能够基于系统内部数据和知识提供更准确、相关的响应。

### 核心组件

1. **向量存储** - 将文档和数据转换为向量并高效存储
2. **文档处理器** - 处理和分块各种格式的文档
3. **检索引擎** - 实现高效相似度搜索
4. **查询处理器** - 优化用户查询以提高检索效果
5. **增强生成器** - 结合检索结果生成高质量回答

### 架构设计

```
internal/
├── rag/                  # RAG 模块
│   ├── vectorstore/      # 向量存储实现
│   │   ├── manager.go    # 存储管理器
│   │   ├── postgres.go   # PostgreSQL 向量存储
│   │   └── memory.go     # 内存向量存储
│   ├── document/         # 文档处理
│   │   ├── processor.go  # 文档处理器
│   │   └── chunker.go    # 文档分块器
│   ├── retriever/        # 检索实现
│   │   ├── engine.go     # 检索引擎
│   │   └── hybrid.go     # 混合检索策略
│   ├── generator/        # 增强生成器
│   │   └── generator.go  # 生成器实现
│   └── embedding/        # 嵌入模型
│       └── provider.go   # 嵌入提供者
```

### API 接口

- **POST /api/v1/rag/documents** - 上传和索引文档
- **GET /api/v1/rag/documents** - 获取已索引文档列表
- **DELETE /api/v1/rag/documents/{id}** - 删除索引文档
- **POST /api/v1/rag/query** - 基于检索增强生成回答用户查询

### 使用示例

```go
// 初始化 RAG 系统 (使用本地向量数据库和模型)
ragSystem := rag.NewSystem(rag.Config{
    VectorStore: "qdrant",            // docker-compose 部署的向量数据库
    VectorStoreURL: "http://qdrant:6333",
    EmbeddingProvider: "local",       // 使用本地嵌入模型
    EmbeddingModel: "bge-base",
    ChunkSize: 1000,
    ChunkOverlap: 200,
})

// 索引文档
err := ragSystem.IndexDocument(context.Background(), rag.Document{
    Content: documentContent,
    Metadata: map[string]interface{}{
        "title": "系统架构文档",
        "author": "技术团队",
        "date": "2025-05-01",
    },
})
if err != nil {
    log.Fatalf("Index error: %v", err)
}

// 查询问题
response, err := ragSystem.Query(context.Background(), "任务调度器的工作原理是什么？")
if err != nil {
    log.Fatalf("Query error: %v", err)
}

fmt.Println("RAG Response:", response.Answer)
fmt.Println("Sources:", response.Sources)
```

## 本地 AI 模型支持

### 概述

DistributedJob 支持本地 AI 模型部署，使您能够在自己的基础设施上运行 AI 功能，无需依赖外部 API，提高数据安全性并降低运营成本。本地模型支持与云端模型保持相同的接口和使用体验，实现无缝切换。所有基础设施（包括本地模型服务、MySQL、Redis 等）均使用 docker-compose 进行部署，并默认配置使用阿里云镜像源以提高国内网络环境下的部署速度。

### 支持的本地模型

1. **Qwen3 系列 (低尺寸)**

   - Qwen3-1.8B
   - Qwen3-4B
   - Qwen3-7B

2. **DeepseekV3 系列**

   - DeepseekV3-7B
   - DeepseekV3-Coder-7B
   - DeepseekV3-Lite-4B

3. **其他可选模型**

   - LLaMa 系列
   - Mistral 系列
   - Baichuan 系列
   - Yi 系列

### 部署方式

所有本地模型均通过 Docker 容器化部署，详细部署方式请参考[部署文档](/docs/deployment.md)。

### 量化支持

支持模型量化以减少内存占用和提高推理速度：

- **GGUF 格式** - 通过 llama.cpp 支持多级量化 (Q4_K_M, Q5_K_M, Q8_0)
- **GPTQ 量化** - 支持 4 位和 8 位量化
- **AWQ 量化** - 支持高性能权重量化
- **ONNX 导出** - 支持 ONNX 运行时优化

### 使用示例

```go
// 初始化本地 Qwen3 模型客户端
localModel := mcp.NewClient(mcp.Config{
    Provider: "local",
    Model:    "qwen3-4b-q5",
    LocalConfig: mcp.LocalConfig{
        APIServerURL: "http://llm-server:8080",  // docker-compose 服务名称
        Models: []mcp.LocalModelConfig{
            {
                Name: "qwen3-4b",
                Path: "qwen3-4b-q5_k_m.gguf",
            },
        },
    },
})

// 使用方式与云端模型相同
response, err := localModel.Chat(ctx, &mcp.ChatRequest{
    Messages: []mcp.Message{
        {Role: "user", Content: "优化任务调度的最佳实践是什么？"},
    },
})
```

### 混合部署策略

支持云端模型和本地模型混合部署，根据不同需求选择最合适的模型：

```go
// 创建模型路由器
modelRouter := mcp.NewModelRouter(mcp.RouterConfig{
    DefaultProvider: "local",
    RouteRules: []mcp.RouteRule{
        {
            Pattern:  "sensitive-*", // 敏感数据查询
            Provider: "local",       // 使用本地模型
            Model:    "deepseekv3-7b",
        },
        {
            Pattern:  "code-*",      // 代码相关查询
            Provider: "local",       // 使用本地模型
            Model:    "deepseekv3-coder-7b",
        },
        {
            Pattern:  "complex-*",   // 复杂问题
            Provider: "openai",      // 使用外部 API
            Model:    "gpt-4",
        },
    },
})
```

## 集成场景

### Agent + MCP 集成

Agent 系统通过 MCP 协议与 AI 模型通信，实现智能决策：

```go
// Agent 使用 MCP 客户端进行决策
func (a *SchedulerAgent) MakeDecision(ctx context.Context, input string) (string, error) {
    request := &mcp.ChatRequest{
        Messages: []mcp.Message{
            {Role: "system", Content: a.systemPrompt},
            {Role: "user", Content: input},
        },
    }

    response, err := a.mcpClient.Chat(ctx, request)
    if err != nil {
        return "", fmt.Errorf("AI decision error: %v", err)
    }

    return response.Content, nil
}
```

### RAG + MCP 集成

RAG 系统利用 MCP 协议实现知识增强的生成：

```go
// RAG 生成器使用 MCP 客户端
func (g *RagGenerator) Generate(ctx context.Context, query string, documents []Document) (string, error) {
    prompt := g.buildPromptWithDocuments(query, documents)

    request := &mcp.ChatRequest{
        Messages: []mcp.Message{
            {Role: "system", Content: "根据提供的文档信息回答问题。"},
            {Role: "user", Content: prompt},
        },
    }

    response, err := g.mcpClient.Chat(ctx, request)
    if err != nil {
        return "", fmt.Errorf("RAG generation error: %v", err)
    }

    return response.Content, nil
}
```

### 完整集成示例

实现全功能 AI 助手，结合 Agent、MCP 和 RAG：

```go
// AI 助手集成所有功能
func NewAIAssistant(config Config) *AIAssistant {
    // 初始化 MCP 客户端
    mcpClient := mcp.NewClient(config.MCPConfig)

    // 初始化 RAG 系统
    ragSystem := rag.NewSystem(config.RAGConfig)

    // 初始化 Agent 工具集
    tools := []agent.Tool{
        agent.NewSchedulerTool(),
        agent.NewDataTool(),
    }

    // 创建 Agent
    agentSystem := agent.NewAgent(config.AgentConfig, mcpClient, tools)

    return &AIAssistant{
        mcpClient:   mcpClient,
        ragSystem:   ragSystem,
        agentSystem: agentSystem,
    }
}

// 处理用户查询
func (a *AIAssistant) HandleQuery(ctx context.Context, query string) (*Response, error) {
    // 1. 使用 RAG 检索相关信息
    ragResult, err := a.ragSystem.Query(ctx, query)
    if err != nil {
        return nil, fmt.Errorf("RAG query error: %v", err)
    }

    // 2. 构建增强上下文
    enhancedContext := fmt.Sprintf("用户查询: %s\n\n相关信息: %s",
                                  query, ragResult.Answer)

    // 3. 使用 Agent 系统处理增强后的查询
    agentResponse, err := a.agentSystem.Process(ctx, enhancedContext)
    if err != nil {
        return nil, fmt.Errorf("Agent processing error: %v", err)
    }

    return &Response{
        Answer:  agentResponse,
        Sources: ragResult.Sources,
    }, nil
}
```

## 性能指标和监控

AI 功能集成了 DistributedJob 的性能监控系统，提供以下指标：

1. **Agent 性能指标**

   - 决策响应时间
   - 任务成功率
   - 资源利用效率

2. **MCP 性能指标**

   - API 调用延迟
   - 令牌使用量
   - 请求成功率

3. **RAG 性能指标**

   - 检索延迟
   - 检索准确率
   - 向量存储利用率

所有指标通过 Prometheus 收集，并在 Grafana 面板中可视化。

## 安全考虑

1. **数据安全**

   - 所有敏感数据在向量化前进行屏蔽
   - 支持本地部署模型以避免数据离开私有环境

2. **访问控制**

   - 基于现有权限系统控制 AI 功能访问
   - 操作日志记录，支持审计

3. **安全审查**

   - AI 生成内容的审核机制
   - 防止提示注入和其他 AI 相关安全风险的措施
