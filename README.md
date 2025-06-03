<div align="center">
  <h1>DistributedJob</h1>
  <h3>高性能分布式任务调度系统</h3>
</div>

![Go Version](https://img.shields.io/badge/go-1.24+-blue.svg)
![License](https://img.shields.io/badge/license-MIT-green.svg)

## 项目简介

DistributedJob 是一个功能强大、高可用的分布式任务调度系统，基于 Go 语言开发，提供任务的可靠调度、执行和监控。系统支持多种类型的任务执行方式，包括 HTTP 回调和 gRPC 调用，同时具备完整的用户权限管理和系统监控能力。

## 主要特性

- **分布式调度**: 基于 etcd 实现分布式锁，确保任务不重复执行
- **多种执行方式**: 支持 HTTP 回调和 gRPC 协议执行任务
- **可靠性保障**:
  - Kafka 消息队列支持，确保任务不丢失
  - 任务重试机制，处理临时故障
  - 事务支持，确保数据一致性
- **权限管理**: 完整的用户、角色、权限和部门管理
- **可观测性**:
  - Prometheus 指标收集
  - Jaeger 分布式跟踪
  - 结构化日志支持
- **高性能**:
  - 协程池管理
  - 连接池优化
  - 并发控制
- **Web界面**: 现代化 Vue3 前端，支持任务管理和监控

## 系统架构

DistributedJob 采用现代化的微服务架构，主要由以下组件构成:

- **核心调度服务**: 负责任务调度和分发
- **API服务**: RESTful API，提供管理接口
- **RPC服务**: gRPC 接口，用于任务执行和服务间通信
- **Web界面**: Vue3 构建的管理界面
- **存储层**: MySQL 用于持久化数据，Redis 用于缓存
- **消息队列**: Kafka 用于任务分发
- **服务协调**: etcd 用于服务发现和分布式锁
- **可观测性组件**: Prometheus, Grafana, Jaeger, Elasticsearch

## 快速开始

### 系统要求

- Go 1.24+
- Docker & Docker Compose
- Node.js 16+

### 使用 Docker Compose 启动

1. 克隆代码库

   ```bash
   git clone https://github.com/yourusername/distributedJob.git
   cd distributedJob
   ```
2. 使用 Docker Compose 启动所有组件

   ```bash
   docker-compose up -d
   ```

   这将启动所有必要的组件，包括:

   - MySQL 数据库
   - Redis 缓存
   - Kafka & Zookeeper
   - etcd 服务
   - Prometheus & Grafana
   - Elasticsearch & Kibana
   - Jaeger 链路追踪
3. 访问服务

   - API服务: http://localhost:8080/v1
   - Web界面: http://localhost:5173
   - Grafana: http://localhost:3000
   - Jaeger UI: http://localhost:16686
   - Kibana: http://localhost:5601

### 手动部署

1. 准备依赖组件 (可参考 docker-compose.yml)
2. 编译服务

   ```bash
   go build -o distributedJob ./cmd/main.go
   ```
3. 配置服务

   ```bash
   # 编辑配置文件以适应你的环境
   vim configs/config.yaml
   ```
4. 初始化数据库

   ```bash
   mysql -u root -p < scripts/init-db/init.sql
   ```
5. 启动服务

   ```bash
   ./distributedJob --config configs/config.yaml
   ```
6. 构建并部署前端

   ```bash
   cd web-ui
   npm install
   npm run build
   # 将生成的 dist 目录部署到 web 服务器
   ```
7. 更多使用说明请参见 `docs/build.md`

## 配置说明

主要配置文件位于 `configs/config.yaml`，包含以下配置项:

```yaml
server:
  host: 0.0.0.0
  port: 8080
  context_path: /v1
  shutdown_timeout: 30

database:
  url: localhost:3306
  username: root
  password: root
  schema: distributed_job

redis:
  url: localhost:6379
  password: ""
  db: 0

# 更多配置项...
```

## API

系统提供完整的 RESTful API:

- **认证接口**: 用户登录、注销、刷新令牌
- **用户管理**: 创建、查询、更新用户信息
- **角色权限**: 角色分配、权限管理
- **部门管理**: 部门层级结构管理
- **任务管理**: 任务的创建、执行、查询和控制
- **系统监控**: 健康检查、性能指标收集

## 贡献指南

欢迎提交 Issue 和 Pull Request 贡献代码。在提交 PR 前，请确保:

1. 代码风格符合项目规范
2. 提供完整的单元测试和集成测试
3. 必要时更新文档

## 开发与测试

### 后端开发
```bash
# 启动后端服务
go run ./cmd/main.go

# 运行单元测试
go test ./...
```

### 前端开发
```bash
cd web-ui
npm install
npm run dev
# 前端开发服务器将在 http://localhost:5173 启动
# API请求会自动代理到后端服务 http://localhost:8080
```

### 端口说明
- 后端API服务: http://localhost:8080/v1
- 前端开发服务器: http://localhost:5173
- 后端RPC服务: localhost:8081

## 许可证

本项目采用 MIT 许可证。
