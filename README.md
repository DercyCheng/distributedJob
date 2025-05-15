<div align="center"><h1>DistributedJob</h1></div>

<div align="center">
  <img src="https://img.shields.io/badge/go-%2300ADD8.svg?style=for-the-badge&logo=go&logoColor=white" alt="Go">
  <img src="https://img.shields.io/badge/docker-%230db7ed.svg?style=for-the-badge&logo=docker&logoColor=white" alt="Docker">
  <img src="https://img.shields.io/badge/AI%20Powered-8A2BE2?style=for-the-badge" alt="AI Powered">
  <img src="https://img.shields.io/badge/License-MIT-yellow.svg?style=for-the-badge" alt="License: MIT">
</div>

一个功能强大的分布式任务调度系统，集成了 AI 智能代理、RAG 检索增强生成以及 MCP 模型上下文协议的功能。该系统可以可靠地处理各种类型的计划任务，从简单的 HTTP 请求到复杂的 GRPC 服务调用，同时拥有 AI 增强的能力，支持智能分析和决策。

## ✨ 特性

- **分布式任务调度**：强大的任务调度引擎，支持 Cron 表达式和精确调度
- **多种任务类型**：支持 HTTP 和 GRPC 任务类型，灵活处理不同场景
- **重试机制**：任务失败自动重试，可配置重试次数和间隔
- **故障转移**：内置故障转移机制，确保任务可靠执行
- **AI 能力集成**：
  - **智能代理**：基于 LLM 的智能代理，可执行多种工具和任务
  - **RAG 系统**：检索增强生成系统，支持文档处理和智能信息检索
  - **MCP 支持**：实现 Model Context Protocol，支持 Anthropic 和 OpenAI 等模型
- **分布式协调**：通过 etcd 实现节点协调和领导者选举
- **消息队列**：使用 Kafka 实现任务分发和负载均衡
- **可观测性**：集成 Prometheus 监控和 OpenTelemetry 分布式追踪
- **权限管理**：基于 RBAC 的完整权限管理系统
- **现代化 Web UI**：Vue3 构建的友好管理界面

## 📋 先决条件

- Go 1.18+
- Docker 和 Docker Compose
- MySQL 8.0+
- Redis 7.0+
- Kafka (可选，用于分布式任务队列)
- etcd (可选，用于分布式协调)
- Vector 数据库 (可选，用于 RAG 嵌入存储)

## 🚀 快速开始

### 使用 Docker Compose

最简单的方式是使用 Docker Compose 来启动整个系统：

```bash
# 克隆仓库
git clone https://github.com/yourusername/distributedJob.git
cd distributedJob

# 启动所有服务
docker-compose up -d
```

### 手动启动

1. **准备数据库**

```bash
# MySQL初始化
mysql -u root -p < scripts/init-db/init.sql

# 初始化向量数据库(如果使用RAG功能)
mysql -u root -p < scripts/init-vector-db.sql
```

2. **构建并运行**

```bash
# 构建项目
go build -o distributed-job ./cmd

# 运行服务
./distributed-job --config configs/config.yaml
```

3. **访问 Web 界面**

打开浏览器访问： `http://localhost:8080`

## ⚙️ 配置

主要配置文件位于 `configs/config.yaml`。您可以根据需要调整以下关键配置：

```yaml
server:
  host: 0.0.0.0
  port: 8080
  context_path: /v1

database:
  url: localhost:3306
  username: root
  password: root
  schema: distributed_job
# 其他配置项...
```

## 📚 架构

DistributedJob 采用模块化设计，围绕几个核心组件构建：

- **任务调度器**：负责任务的调度和执行
- **HTTP/GRPC 工作器**：处理不同类型的任务执行
- **AI 控制器**：整合 Agent、MCP 和 RAG 功能
- **Web API**：提供 RESTful 接口管理任务和系统
- **存储层**：包含 MySQL、Redis、Vector 数据库等多种存储

更多架构细节，请参考 [架构文档](docs/structure.md)。

## 🧠 AI 功能

系统集成了三个主要的 AI 功能模块：

1. **Agent 智能代理**：

   - 基于 LLM 的智能代理可执行复杂任务
   - 支持多种工具集成，包括数据工具、调度工具和系统工具

2. **RAG 检索增强生成**：

   - 文档处理和分块能力
   - 多种 Embedding 提供者支持
   - 智能信息检索和生成

3. **MCP 模型上下文协议**：

   - 支持 Anthropic、OpenAI 等模型
   - 上下文管理和窗口控制
   - 流式处理能力

更多 AI 功能的细节，请参考 [AI 功能文档](docs/ai.md)。

## 🔧 API 参考

API 文档可通过启动服务后访问 Swagger UI 查看：
`http://localhost:8080/swagger/index.html`

主要 API 端点包括：

- `/v1/task`：任务管理 API
- `/v1/auth`：认证和权限管理
- `/v1/ai`：AI 功能 API
- `/v1/dashboard`：仪表盘数据 API

## 🧪 测试

运行单元测试：

```bash
go test ./...
```

运行集成测试：

```bash
go test -tags=integration ./...
```

## 🤝 贡献

欢迎贡献！请查看[贡献指南](CONTRIBUTING.md)了解如何开始。

## 📄 许可证

本项目采用 MIT 许可证 - 详情请参见[LICENSE](LICENSE)文件。

## 📞 联系方式

如有任何问题或建议，请开启一个 issue 或联系项目维护者。
