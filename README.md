# Go-Job

一个现代化的分布式任务调度和管理系统，基于 Go 语言开发，提供 Web 界面和 API 接口。

![Go Version](https://img.shields.io/badge/Go-1.23+-blue.svg)
![Vue Version](https://img.shields.io/badge/Vue-3.3+-green.svg)
![License](https://img.shields.io/badge/license-MIT-blue.svg)

## ✨ 特性

- 🚀 **高性能**: 基于 Go 语言开发，支持高并发任务调度
- 🎯 **多协议支持**: 提供 HTTP REST API 和 gRPC 接口
- 🖥️ **现代化 UI**: 基于 Vue 3 + Element Plus 的响应式 Web 界面
- 📊 **实时监控**: 集成 Prometheus + Grafana 监控体系
- 🔐 **权限管理**: 完整的用户认证和权限控制系统
- 🔄 **任务调度**: 灵活的 Cron 表达式支持
- 💾 **数据持久化**: MySQL 数据库 + Redis 缓存
- 🐳 **容器化部署**: 完整的 Docker 和 Docker Compose 支持
- 🤖 **AI 集成**: 支持 MCP (Model Context Protocol) 智能调度

## 🏗️ 系统架构

```
go-job/
├── api/                    # API 层
│   ├── grpc/              # gRPC 接口定义
│   └── http/              # HTTP REST API
├── internal/              # 内部业务逻辑
│   ├── auth/              # 认证服务
│   ├── job/               # 任务管理
│   ├── scheduler/         # 调度器
│   ├── mcp/               # MCP AI 调度
│   └── ...
├── pkg/                   # 公共包
│   ├── database/          # 数据库连接
│   ├── redis/             # Redis 客户端
│   ├── auth/              # JWT 认证
│   └── ...
├── web/                   # Vue.js 前端
├── configs/               # 配置文件
└── scripts/               # 部署脚本
```

## 🚀 快速开始

### 环境要求

- Go 1.23+
- Node.js 18+
- Docker & Docker Compose (可选)
- MySQL 8.0+
- Redis 7+

### 方式一：Docker 部署 (推荐)

1. **克隆项目**
```bash
git clone <repository-url>
cd go-job
```

2. **启动所有服务**
```bash
make docker
# 或者
docker-compose up -d
```

3. **访问应用**
- Web 界面: http://localhost:8080
- API 文档: http://localhost:8080/api/docs
- Grafana 监控: http://localhost:3000 (admin/admin)
- Prometheus: http://localhost:9090

### 方式二：本地开发

1. **安装依赖**
```bash
# 后端依赖
make deps

# 前端依赖
cd web
npm install
```

2. **配置数据库**
```bash
# 创建数据库
mysql -u root -p < scripts/init.sql
```

3. **启动 Redis**
```bash
redis-server
```

4. **启动后端服务**
```bash
make dev
```

5. **启动前端服务**
```bash
cd web
npm run dev
```

## 📋 可用命令

```bash
# 构建和运行
make build          # 编译应用
make run            # 运行应用
make dev            # 开发模式运行

# Docker 相关
make docker         # 构建并启动 Docker 容器
make docker-build   # 构建 Docker 镜像
make docker-up      # 启动 Docker 容器
make docker-down    # 停止 Docker 容器

# 开发工具
make test           # 运行测试
make lint           # 代码检查
make fmt            # 格式化代码
make proto          # 生成 protobuf 文件
make clean          # 清理构建文件
```

## 🔧 配置说明

主要配置文件位于 `configs/config.yaml`:

```yaml
server:
  port: 8080
  grpc_port: 9090

database:
  host: localhost
  port: 3306
  username: go_job
  password: password
  database: go_job

redis:
  addr: localhost:6379
  password: ""
  db: 0

auth:
  jwt_secret: "your-secret-key"
  expire_hours: 24
```

## 📊 API 接口

### HTTP REST API

- `GET /api/jobs` - 获取任务列表
- `POST /api/jobs` - 创建新任务
- `PUT /api/jobs/:id` - 更新任务
- `DELETE /api/jobs/:id` - 删除任务
- `POST /api/jobs/:id/execute` - 手动执行任务

### gRPC API

gRPC 服务定义在 `api/grpc/job.proto`，支持：
- 任务 CRUD 操作
- 实时任务状态订阅
- 批量操作接口

## 🖥️ Web 界面功能

- **仪表板**: 系统概览和实时统计
- **任务管理**: 任务的创建、编辑、删除和执行
- **执行历史**: 任务执行记录和日志查看
- **工作节点**: 分布式工作节点管理
- **用户管理**: 用户和权限管理
- **系统监控**: 性能指标和告警

## 🔐 认证和权限

系统支持基于 JWT 的认证机制：

1. **用户登录**获取 JWT Token
2. **权限控制**基于角色和资源
3. **会话管理**支持 Token 刷新

默认管理员账号：
- 用户名: `admin`
- 密码: `admin123`

## 📈 监控和告警

### Prometheus 指标

系统暴露以下关键指标：
- `job_execution_total` - 任务执行总数
- `job_execution_duration` - 任务执行时长
- `job_queue_size` - 任务队列长度
- `worker_active_count` - 活跃工作节点数

### Grafana 仪表板

预配置的 Grafana 仪表板包含：
- 系统整体性能概览
- 任务执行统计图表
- 错误率和延迟监控
- 资源使用情况

## 🤖 AI 智能调度

集成 MCP (Model Context Protocol) 支持：

- **智能任务分析**: AI 分析任务特征和历史数据
- **动态调度优化**: 基于系统负载智能调整调度策略
- **异常检测**: 自动识别和处理异常任务
- **性能预测**: 预测任务执行时间和资源需求

启用 MCP 功能需要配置 AI 服务端点。

## 🛠️ 开发指南

### 添加新的任务类型

1. 在 `internal/job/` 中定义任务逻辑
2. 更新 `api/grpc/job.proto` 接口定义
3. 实现对应的 HTTP API 处理器
4. 添加前端界面支持

### 扩展监控指标

1. 在相应服务中定义 Prometheus 指标
2. 更新 Grafana 仪表板配置
3. 添加告警规则 (如需要)

### 自定义权限

1. 在 `internal/permission/` 中定义权限逻辑
2. 更新 JWT 中间件
3. 前端添加权限检查

## 🐛 故障排除

### 常见问题

1. **数据库连接失败**
   - 检查 MySQL 服务状态
   - 验证配置文件中的数据库连接信息

2. **Redis 连接失败**
   - 确认 Redis 服务运行
   - 检查网络连接和防火墙设置

3. **前端访问 404**
   - 确认后端服务正常启动
   - 检查 Nginx 代理配置

4. **gRPC 连接超时**
   - 验证防火墙规则
   - 检查 gRPC 端口配置

### 日志查看

```bash
# 查看应用日志
docker-compose logs go-job

# 查看数据库日志
docker-compose logs mysql

# 查看 Redis 日志
docker-compose logs redis
```

## 🤝 贡献指南

1. Fork 本项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 创建 Pull Request

## 📄 许可证

本项目采用 MIT 许可证。查看 [LICENSE](LICENSE) 文件了解更多信息。

## 📞 支持

如果您遇到问题或有建议，请通过以下方式联系：

- 创建 [Issue](../../issues)
- 发送邮件至: [your-email@example.com]
- 加入讨论群: [群号或链接]

---

**感谢使用 Go-Job！** 🎉
