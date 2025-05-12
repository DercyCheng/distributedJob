<div align="center">
  <h1>DistributedJob - 系统文档</h1>
</div>

<div align="center">
  <h3>DistributedJob - 分布式调度系统</h3>
</div>

## 目录

1. [架构设计](#架构设计)
   - [概述](#概述)
   - [核心组件](#核心组件)
   - [项目结构](#项目结构)
   - [组件图](#组件图)
   - [工作流程](#工作流程)
   - [设计原则](#设计原则)
   - [RPC 通信](#rpc通信)
2. [数据库设计](#数据库设计)
   - [数据库概述](#数据库概述)
   - [表结构设计](#表结构设计)
   - [数据分表策略](#数据分表策略)
   - [ER 图](#er-图)
   - [数据库优化建议](#数据库优化建议)
3. [安装指南](#安装指南)
   - [系统要求](#系统要求)
   - [安装方法](#安装方法)
   - [配置](#配置)
   - [数据库设置](#数据库设置)
   - [运行服务](#运行服务)
   - [验证](#验证)
   - [部署选项](#部署选项)
4. [API 文档](#api-文档)
   - [API 概述](#api-概述)
   - [用户认证 API](#用户认证-api)
   - [部门管理 API](#部门管理-api)
   - [用户管理 API](#用户管理-api)
   - [角色与权限管理 API](#角色与权限管理-api)
   - [任务管理 API](#任务管理-api)
   - [执行记录查询 API](#执行记录查询-api)
   - [健康检查与服务管理 API](#健康检查与服务管理-api)
   - [RPC 服务 API](#rpc-服务-api)
5. [测试指南](#测试指南)
   - [测试架构](#测试架构)
   - [单元测试](#单元测试)
   - [集成测试](#集成测试)
   - [性能测试](#性能测试)
   - [测试自动化](#测试自动化)
   - [覆盖率分析](#覆盖率分析)
6. [前端开发](#前端开发)
   - [技术栈](#技术栈)
   - [项目结构](#前端项目结构)
   - [开发指南](#开发指南)
   - [构建与部署](#构建与部署)
7. [令牌安全机制](#令牌安全机制)
   - [概述](#令牌概述)
   - [双令牌机制](#双令牌机制)
   - [令牌撤销](#令牌撤销)
   - [令牌内容优化](#令牌内容优化)
   - [令牌传输安全](#令牌传输安全)
   - [令牌刷新流程](#令牌刷新流程)
   - [最佳实践](#令牌最佳实践)

---

## 架构设计

### 概述

DistributedJob 采用模块化设计，围绕几个核心组件构建，这些组件共同协作，提供可靠且可扩展的分布式调度系统。系统现已支持 RPC 通信，增强了组件间的通信效率与可靠性。

### 核心组件

1. **API 层** - 用于任务/部门/用户管理的 RESTful API 端点
2. **RPC 服务层** - 提供高性能内部服务通信机制
3. **调度引擎** - 处理任务调度和分发
4. **存储层** - 持久化任务配置和执行记录
5. **Web 控制台** - 基于 Vite 构建的现代化前端界面
6. **认证/权限模块** - 管理身份验证和授权
7. **任务执行器** - 执行 HTTP 和 gRPC 任务，支持重试
8. **历史记录管理器** - 记录和分析任务执行历史

### 项目结构

```
distributedJob/
├── cmd/                  # 命令行应用程序入口点
│   └── main.go           # 服务入口点
├── configs/              # 配置文件目录
│   ├── config.yaml       # 主配置文件
│   └── prometheus/       # Prometheus 相关配置
│       └── prometheus.yml # Prometheus 配置文件
├── docs/                 # 文档目录
│   └── build.md          # 构建和设计文档
├── internal/             # 私有应用程序和库代码
│   ├── api/              # API 相关代码
│   │   ├── server.go     # API 服务器
│   │   ├── handler/      # HTTP 处理器
│   │   │   ├── dashboard_handler.go # 仪表盘处理器
│   │   │   └── health_handler.go    # 健康检查处理器
│   │   └── middleware/   # HTTP 中间件
│   │       ├── cors.go          # 跨域请求中间件
│   │       ├── instrumentation.go # 监控中间件
│   │       ├── jwt_auth.go      # JWT 认证中间件
│   │       ├── logger.go        # 日志中间件
│   │       └── tracing.go       # 链路追踪中间件
│   ├── config/           # 配置管理
│   │   ├── config.go      # 配置结构和加载逻辑
│   │   └── extended_config.go # 扩展配置
│   ├── infrastructure/   # 基础设施
│   │   └── infrastructure.go # 基础设施初始化和管理
│   ├── job/              # 核心任务调度模块
│   │   ├── scheduler.go   # 任务调度器
│   │   ├── http_worker.go # HTTP 任务执行器
│   │   ├── grpc_worker.go # gRPC 任务执行器
│   │   ├── options.go     # 调度器选项
│   │   └── stats.go       # 任务统计
│   ├── model/            # 数据模型
│   │   └── entity/       # 业务实体对象
│   │       ├── department.go    # 部门实体
│   │       ├── permission.go    # 权限实体
│   │       ├── record.go        # 执行记录实体
│   │       ├── role_permission.go # 角色权限关系实体
│   │       ├── role.go          # 角色实体
│   │       ├── task.go          # 任务实体
│   │       └── user.go          # 用户实体
│   ├── rpc/              # RPC 服务相关代码
│   │   ├── client/       # RPC 客户端实现
│   │   ├── proto/        # Protocol Buffers 定义
│   │   │   ├── auth.proto        # 认证服务定义
│   │   │   ├── data.proto        # 数据服务定义
│   │   │   └── scheduler.proto   # 调度器服务定义
│   │   └── server/       # RPC 服务器实现
│   │       ├── auth_service_server.go   # 认证服务实现
│   │       ├── data_service_server.go   # 数据服务实现
│   │       ├── rpc_server.go            # RPC 服务器基础结构
│   │       └── task_scheduler_server.go # 任务调度服务实现
│   ├── service/          # 业务逻辑服务
│   │   ├── auth_service.go  # 认证服务实现
│   │   ├── init_service.go  # 初始化服务
│   │   └── task_service.go  # 任务服务实现
│   └── store/            # 存储层
│       ├── repository.go    # 存储接口定义
│       ├── token_revoker.go # 令牌撤销接口
│       ├── etcd/            # ETCD 存储实现
│       │   └── manager.go   # ETCD 管理器
│       ├── kafka/           # Kafka 存储实现
│       │   └── manager.go   # Kafka 管理器
│       ├── mysql/           # MySQL 实现
│       │   ├── manager.go   # MySQL 连接管理
│       │   └── repository/  # 数据访问对象
│       │       ├── department_repository.go # 部门仓库
│       │       ├── permission_repository.go # 权限仓库
│       │       ├── role_repository.go      # 角色仓库
│       │       ├── task_repository.go      # 任务仓库
│       │       └── user_repository.go      # 用户仓库
│       └── redis/           # Redis 实现
│           ├── manager.go      # Redis 连接管理
│           └── token_revoker.go # 基于 Redis 的令牌撤销
├── pkg/                  # 可被外部应用程序使用的库
│   ├── logger/           # 日志工具
│   │   ├── context.go    # 日志上下文
│   │   └── logger.go     # 日志实现
│   ├── memory/           # 内存相关工具
│   │   └── token_revoker.go # 内存令牌撤销实现
│   ├── metrics/          # 指标监控
│   │   ├── gauge_getter.go # 度量值获取
│   │   └── metrics.go      # 指标监控实现
│   ├── session/          # 会话管理
│   └── tracing/          # 分布式追踪
│       └── tracer.go     # 追踪器实现
├── scripts/              # 构建和部署脚本
│   └── init-db/          # 数据库初始化
│       └── init.sql      # 初始化 SQL 脚本
├── web-ui/               # 前端应用 (Vite 构建)
│   ├── src/              # 源代码
│   │   ├── api/          # API 客户端
│   │   ├── components/   # UI 组件
│   │   ├── router/       # 路由管理
│   │   ├── store/        # 状态管理
│   │   └── views/        # 页面视图
│   ├── index.html        # 入口 HTML
│   └── vite.config.ts    # Vite 配置
├── go.mod                # Go 模块依赖
├── go.sum                # Go 模块校验和
└── docker-compose.yml    # Docker Compose 配置
```

### 组件图

```
┌─────────────────────────────────────────────────────────────────┐
│                     Vite 构建的 Web 控制台                       │
└───────────────────────────────┬─────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                           API 层                                 │
├─────────────┬─────────────┬──────────────┬──────────────────────┤
│  任务 API    │  用户 API    │   部门 API    │    历史记录 API      │
│             │             │              │                      │
└─────────────┴─────────────┴──────────────┴──────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                          服务层                                  │
├─────────────┬─────────────┬──────────────┬──────────────────────┤
│ 任务服务     │ 用户服务     │ 部门服务      │    历史记录服务        │
│             │             │              │                      │
└─────────────┴─────────────┴──────────────┴──────────────────────┘
                                │
                 ┌──────────────┼──────────────┐
                 │              │              │
                 ▼              ▼              ▼
┌───────────────────┐ ┌──────────────────┐ ┌─────────────────────┐
│    调度引擎        │ │     认证模块      │ │      存储层          │
├───────────────────┤ └──────────────────┘ ├─────────────────────┤
│  HTTP 任务执行器   │         │            │      MySQL          │
├───────────────────┤         │            │                     │
│  gRPC 任务执行器   │         │            │                     │
└─────────┬─────────┘         │            └─────────────────────┘
          │                   │
          │                   │
          ▼                   ▼
┌────────────────────────────────────────────────────────────────┐
│                          RPC 服务层                             │
├────────────────────────────────────────────────────────────────┤
│  任务调度 RPC 服务  │  用户认证 RPC 服务  │  数据访问 RPC 服务     │
├────────────────────────────────────────────────────────────────┤
│                        gRPC / Protocol Buffers                 │
└────────────────────────────────────────────────────────────────┘
                              ▲
                              │
                              │
                              ▼
┌────────────────────────────────────────────────────────────────┐
│                        分布式客户端                              │
├────────────────────────────────────────────────────────────────┤
│  任务执行客户端  │  管理工具客户端  │  第三方系统集成客户端        │
└────────────────────────────────────────────────────────────────┘
```

### 工作流程

1. **系统初始化**

   - 从 config.yaml 加载配置
   - 初始化数据库连接
   - 设置日志记录
   - 启动 HTTP 服务器
   - 启动 RPC 服务器
   - 初始化调度器

2. **任务调度**

   - 调度器扫描数据库中的活动任务
   - 根据 Cron 表达式组织任务并分配执行上下文
   - 支持分布式部署模式，通过 ETCD 实现分布式锁
   - 可选启用 Kafka 支持，用于任务的可靠分发
   - 任务执行上下文通过 JobContext 传递，包含完整的任务信息
   - 调度器实现任务队列和并发控制，避免系统过载

3. **任务执行**

   - 支持 HTTP Worker 和 gRPC Worker 两种执行器类型
   - 执行器负责任务执行、结果收集和错误处理
   - HTTP Worker 支持多种 HTTP 方法、自定义请求头和请求体
   - gRPC Worker 支持服务发现和自动重连
   - 完善的重试机制，根据配置的重试次数和间隔进行重试
   - 支持主备地址切换策略，当主地址执行失败时自动切换到备用地址
   - 执行结果记录在执行历史中，支持按年月分表存储
   - 提供完善的指标收集，支持 Prometheus 监控和 OpenTelemetry 追踪

4. **用户身份与权限管理**

   - 完备的用户认证系统，包括登录、令牌验证和权限检查
   - 实现基于 JWT 的双令牌机制 (Access Token + Refresh Token)
   - 支持多种令牌撤销策略 (内存、Redis)，确保安全退出
   - 部门-角色-权限三层设计，实现细粒度权限控制
   - 角色与权限的多对多关系，支持灵活的权限分配
   - 用户资源按部门隔离，确保数据安全

5. **用户交互**

   - 用户通过基于 Vue 3 + Vite 构建的现代化 Web 控制台与系统交互
   - 完整的路由和状态管理，支持组件化开发
   - 集成响应式布局和主题切换，提供良好的用户体验
   - 支持仪表盘、任务管理、部门管理、用户管理、角色权限管理等功能
   - HTTP API 和 gRPC API 双渠道接入，满足不同场景需求

### 设计原则

- **模块化设计**：系统按功能划分为明确的模块，各模块间通过接口交互，降低耦合度
- **可扩展架构**：采用无状态设计，支持水平扩展，适应不同规模的部署需求
- **高可用保障**：完善的重试机制、主备切换和分布式锁，确保任务调度的可靠性
- **分布式友好**：支持多实例部署，通过 ETCD 协调，避免任务重复执行
- **安全性设计**：实现基于部门-角色-权限的三层访问控制模型，JWT 双令牌机制保障系统安全
- **可观测性**：集成日志、指标和分布式追踪，支持 Prometheus 监控和 OpenTelemetry 追踪
- **高性能通信**：使用 gRPC 实现服务间高效通信，二进制序列化减少网络开销
- **资源隔离**：基于部门的资源隔离设计，确保多租户场景下的数据安全
- **开发友好**：合理的项目结构和接口设计，降低开发和维护难度

### RPC 通信

DistributedJob 系统现在使用 gRPC 作为 RPC 框架，实现高效的内部服务通信。

#### 核心 RPC 服务

1. **任务调度 RPC 服务**

   - `ScheduleTask` - 调度一个任务
   - `PauseTask` - 暂停一个任务
   - `ResumeTask` - 恢复一个已暂停的任务
   - `GetTaskStatus` - 获取任务状态

2. **用户认证 RPC 服务**

   - `Authenticate` - 验证用户凭证
   - `ValidateToken` - 验证 JWT 令牌
   - `GetUserPermissions` - 获取用户权限

3. **数据访问 RPC 服务**

   - `GetTaskHistory` - 获取任务执行历史
   - `GetStatistics` - 获取系统统计数据

#### Protocol Buffers 定义

DistributedJob 使用 Protocol Buffers 来定义 RPC 服务接口。系统提供了三个主要的 RPC 服务：

1. **任务调度服务 (scheduler.proto)**

```protobuf
syntax = "proto3";
package scheduler;

option go_package = "distributedJob/internal/rpc/proto";

service TaskScheduler {
  rpc ScheduleTask(ScheduleTaskRequest) returns (ScheduleTaskResponse);
  rpc PauseTask(TaskRequest) returns (TaskResponse);
  rpc ResumeTask(TaskRequest) returns (TaskResponse);
  rpc GetTaskStatus(TaskRequest) returns (TaskStatusResponse);
  rpc ExecuteTaskImmediately(TaskRequest) returns (TaskResponse);
  rpc BatchScheduleTasks(BatchScheduleTasksRequest) returns (BatchScheduleTasksResponse);
  rpc DeleteTask(TaskRequest) returns (TaskResponse);
}

message ScheduleTaskRequest {
  string name = 1;
  string cron_expression = 2;
  string handler = 3;
  bytes params = 4;
  int32 max_retry = 5;
  int64 department_id = 6;
  int32 timeout = 7;
}

message ScheduleTaskResponse {
  int64 task_id = 1;
  bool success = 2;
  string message = 3;
}

message TaskRequest {
  int64 task_id = 1;
}

message TaskResponse {
  bool success = 1;
  string message = 2;
}

message TaskStatusResponse {
  int64 task_id = 1;
  int32 status = 2;
  string last_execute_time = 3;
  string next_execute_time = 4;
  int32 retry_count = 5;
  int32 success_count = 6;
  int32 fail_count = 7;
  float avg_execution_time = 8;
}

message BatchScheduleTasksRequest {
  repeated ScheduleTaskRequest tasks = 1;
}

message BatchScheduleTasksResponse {
  repeated ScheduleTaskResponse results = 1;
  bool overall_success = 2;
  string message = 3;
}
```

2. **认证服务 (auth.proto)**

```protobuf
syntax = "proto3";
package auth;

option go_package = "distributedJob/internal/rpc/proto";

service AuthService {
  rpc Authenticate(AuthenticateRequest) returns (AuthenticateResponse);
  rpc ValidateToken(ValidateTokenRequest) returns (ValidateTokenResponse);
  rpc RefreshToken(RefreshTokenRequest) returns (RefreshTokenResponse);
  rpc GetUserPermissions(UserPermissionsRequest) returns (UserPermissionsResponse);
}

message AuthenticateRequest {
  string username = 1;
  string password = 2;
}

message AuthenticateResponse {
  bool success = 1;
  string access_token = 2;
  string refresh_token = 3;
  UserInfo user_info = 4;
  string message = 5;
}

message UserInfo {
  int64 user_id = 1;
  string username = 2;
  string real_name = 3;
  int64 department_id = 4;
  string department_name = 5;
  int64 role_id = 6;
  string role_name = 7;
}

message ValidateTokenRequest {
  string token = 1;
}

message ValidateTokenResponse {
  bool valid = 1;
  int64 user_id = 2;
  string message = 3;
}

message RefreshTokenRequest {
  string refresh_token = 1;
}

message RefreshTokenResponse {
  bool success = 1;
  string access_token = 2;
  string refresh_token = 3;
  string message = 4;
}

message UserPermissionsRequest {
  int64 user_id = 1;
}

message UserPermissionsResponse {
  bool success = 1;
  repeated string permissions = 2;
  string message = 3;
}
```

3. **数据服务 (data.proto)**

```protobuf
syntax = "proto3";
package data;

option go_package = "distributedJob/internal/rpc/proto";

service DataService {
  rpc GetTaskHistory(TaskHistoryRequest) returns (TaskHistoryResponse);
  rpc GetUserList(UserListRequest) returns (UserListResponse);
  rpc GetDepartmentList(DepartmentListRequest) returns (DepartmentListResponse);
  rpc GetTaskStatistics(TaskStatisticsRequest) returns (TaskStatisticsResponse);
}

message TaskHistoryRequest {
  int64 task_id = 1;
  string start_time = 2;
  string end_time = 3;
  int32 limit = 4;
  int32 offset = 5;
  int32 year = 6;
  int32 month = 7;
}

message TaskHistoryRecord {
  int64 id = 1;
  int64 task_id = 2;
  string task_name = 3;
  bool success = 4;
  int32 status_code = 5;
  string response = 6;
  int32 cost_time = 7;
  string execute_time = 8;
  int32 retry_times = 9;
}

message TaskHistoryResponse {
  bool success = 1;
  repeated TaskHistoryRecord records = 2;
  int64 total = 3;
  string message = 4;
}

message UserListRequest {
  int64 department_id = 1;
  int32 page = 2;
  int32 size = 3;
}

message UserInfo {
  int64 id = 1;
  string username = 2;
  string real_name = 3;
  string email = 4;
  string phone = 5;
  int64 department_id = 6;
  string department_name = 7;
  int64 role_id = 8;
  string role_name = 9;
  int32 status = 10;
  string create_time = 11;
}

message UserListResponse {
  bool success = 1;
  repeated UserInfo users = 2;
  int64 total = 3;
  string message = 4;
}

message DepartmentListRequest {
  int32 page = 1;
  int32 size = 2;
}

message Department {
  int64 id = 1;
  string name = 2;
  string description = 3;
  string create_time = 4;
}

message DepartmentListResponse {
  bool success = 1;
  repeated Department departments = 2;
  int64 total = 3;
  string message = 4;
}

message TaskStatisticsRequest {
  int64 department_id = 1;
  string start_time = 2;
  string end_time = 3;
}

message TaskStatisticsResponse {
  bool success = 1;
  int32 task_count = 2;
  float success_rate = 3;
  float avg_execution_time = 4;
  map<string, float> execution_stats = 5;
  string message = 6;
}
```

}

````

#### RPC 客户端示例

```go
package main

import (
  "context"
  "log"
  "time"

  schedulerpb "github.com/username/distributedJob/internal/rpc/proto"
  "google.golang.org/grpc"
)

func main() {
  conn, err := grpc.Dial("localhost:9090", grpc.WithInsecure())
  if (err != nil) {
    log.Fatalf("Failed to connect: %v", err)
  }
  defer conn.Close()

  client := schedulerpb.NewTaskSchedulerClient(conn)

  ctx, cancel := context.WithTimeout(context.Background(), time.Second)
  defer cancel()

  resp, err := client.ScheduleTask(ctx, &schedulerpb.ScheduleTaskRequest{
    Name:           "ExampleTask",
    CronExpression: "*/5 * * * *",
    Handler:        "http",
    Params:         []byte(`{"url": "http://example.com/api"}`),
    MaxRetry:       3,
  })

  if (err != nil) {
    log.Fatalf("Could not schedule task: %v", err)
  }

  log.Printf("Task scheduled with ID: %d, Success: %v", resp.TaskId, resp.Success)
}
````

#### RPC 服务端实现

```go
package server

import (
  "context"

  schedulerpb "github.com/username/distributedJob/internal/rpc/proto"
  "github.com/username/distributedJob/internal/job"
)

type TaskSchedulerServer struct {
  schedulerpb.UnimplementedTaskSchedulerServer
  scheduler *job.Scheduler
}

func NewTaskSchedulerServer(scheduler *job.Scheduler) *TaskSchedulerServer {
  return &TaskSchedulerServer{scheduler: scheduler}
}

func (s *TaskSchedulerServer) ScheduleTask(ctx context.Context, req *schedulerpb.ScheduleTaskRequest) (*schedulerpb.ScheduleTaskResponse, error) {
  taskID, err := s.scheduler.ScheduleTask(req.Name, req.CronExpression, req.Handler, req.Params, int(req.MaxRetry))
  if (err != nil) {
    return &schedulerpb.ScheduleTaskResponse{
      Success: false,
      Message: err.Error(),
    }, nil
  }

  return &schedulerpb.ScheduleTaskResponse{
    TaskId:  taskID,
    Success: true,
    Message: "Task scheduled successfully",
  }, nil
}

// 其他 RPC 方法实现...
```

---

## 数据库设计

### 数据库概述

DistributedJob 使用 MySQL 数据库存储任务配置、用户权限和执行记录。数据库设计遵循以下原则：

- 简单实用：只设计必要的表结构，减少复杂度
- 良好性能：合理的索引设计，优化查询性能
- 权限分离：清晰的权限模型，支持多部门管理和权限控制

### 表结构设计

#### 部门表 (department)

部门表存储系统中的部门信息。

| 字段名      | 数据类型     | 是否为空 | 默认值                                        | 说明                       |
| ----------- | ------------ | -------- | --------------------------------------------- | -------------------------- |
| id          | bigint(20)   | 否       | 自增                                          | 主键                       |
| name        | varchar(255) | 否       | 无                                            | 部门名称                   |
| description | varchar(500) | 是       | NULL                                          | 部门描述                   |
| parent_id   | bigint(20)   | 是       | NULL                                          | 父部门 ID，顶级部门为 NULL |
| status      | tinyint(4)   | 否       | 1                                             | 状态：0-禁用，1-启用       |
| create_time | datetime     | 否       | CURRENT_TIMESTAMP                             | 创建时间                   |
| update_time | datetime     | 否       | CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP | 更新时间                   |

索引：

- PRIMARY KEY (`id`)
- KEY `idx_parent_id` (`parent_id`)
- KEY `idx_status` (`status`)

#### 用户表 (user)

用户表存储系统用户信息。

| 字段名        | 数据类型     | 是否为空 | 默认值                                        | 说明                 |
| ------------- | ------------ | -------- | --------------------------------------------- | -------------------- |
| id            | bigint(20)   | 否       | 自增                                          | 主键                 |
| username      | varchar(50)  | 否       | 无                                            | 用户名               |
| password      | varchar(100) | 否       | 无                                            | 密码（加密存储）     |
| real_name     | varchar(50)  | 否       | 无                                            | 真实姓名             |
| email         | varchar(100) | 是       | NULL                                          | 电子邮箱             |
| phone         | varchar(20)  | 是       | NULL                                          | 手机号码             |
| department_id | bigint(20)   | 否       | 无                                            | 所属部门 ID          |
| role_id       | bigint(20)   | 否       | 无                                            | 角色 ID              |
| status        | tinyint(4)   | 否       | 1                                             | 状态：0-禁用，1-启用 |
| create_time   | datetime     | 否       | CURRENT_TIMESTAMP                             | 创建时间             |
| update_time   | datetime     | 否       | CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP | 更新时间             |

索引：

- PRIMARY KEY (`id`)
- UNIQUE KEY `idx_username` (`username`)
- KEY `idx_department_id` (`department_id`)
- KEY `idx_role_id` (`role_id`)
- KEY `idx_status` (`status`)

#### 角色表 (role)

角色表存储系统角色信息。

| 字段名      | 数据类型     | 是否为空 | 默认值                                        | 说明                 |
| ----------- | ------------ | -------- | --------------------------------------------- | -------------------- |
| id          | bigint(20)   | 否       | 自增                                          | 主键                 |
| name        | varchar(50)  | 否       | 无                                            | 角色名称             |
| description | varchar(255) | 是       | NULL                                          | 角色描述             |
| status      | tinyint(4)   | 否       | 1                                             | 状态：0-禁用，1-启用 |
| create_time | datetime     | 否       | CURRENT_TIMESTAMP                             | 创建时间             |
| update_time | datetime     | 否       | CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP | 更新时间             |

索引：

- PRIMARY KEY (`id`)
- UNIQUE KEY `idx_name` (`name`)

#### 权限表 (permission)

权限表存储系统权限信息。

| 字段名      | 数据类型     | 是否为空 | 默认值                                        | 说明                 |
| ----------- | ------------ | -------- | --------------------------------------------- | -------------------- |
| id          | bigint(20)   | 否       | 自增                                          | 主键                 |
| name        | varchar(50)  | 否       | 无                                            | 权限名称             |
| code        | varchar(50)  | 否       | 无                                            | 权限编码             |
| description | varchar(255) | 是       | NULL                                          | 权限描述             |
| status      | tinyint(4)   | 否       | 1                                             | 状态：0-禁用，1-启用 |
| create_time | datetime     | 否       | CURRENT_TIMESTAMP                             | 创建时间             |
| update_time | datetime     | 否       | CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP | 更新时间             |

索引：

- PRIMARY KEY (`id`)
- UNIQUE KEY `idx_code` (`code`)

#### 角色权限关联表 (role_permission)

角色权限关联表存储角色与权限的多对多关系。

| 字段名        | 数据类型 | 是否为空 | 默认值                                        | 说明     |
| ------------- | -------- | -------- | --------------------------------------------- | -------- |
| id            | bigint   | 否       | 自增                                          | 主键     |
| role_id       | bigint   | 否       | 无                                            | 角色 ID  |
| permission_id | bigint   | 否       | 无                                            | 权限 ID  |
| create_time   | datetime | 否       | CURRENT_TIMESTAMP                             | 创建时间 |
| update_time   | datetime | 否       | CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP | 更新时间 |

索引：

- PRIMARY KEY (`id`)
- UNIQUE KEY `idx_role_perm` (`role_id`, `permission_id`)
- KEY `idx_permission_id` (`permission_id`)

#### 任务表 (task)

任务表存储所有定时任务的配置信息。

| 字段名            | 数据类型     | 是否为空 | 默认值                                        | 说明                         |
| ----------------- | ------------ | -------- | --------------------------------------------- | ---------------------------- |
| id                | bigint       | 否       | 自增                                          | 主键                         |
| name              | varchar(255) | 否       | 无                                            | 任务名称                     |
| description       | text         | 是       | NULL                                          | 任务描述                     |
| cron_expression   | varchar(50)  | 是       | NULL                                          | cron 表达式                  |
| handler           | varchar(255) | 否       | 无                                            | 任务处理器                   |
| params            | text         | 是       | NULL                                          | 任务参数(JSON 格式)          |
| status            | tinyint      | 否       | 0                                             | 状态：0-禁用，1-启用，2-临时 |
| max_retry         | int          | 否       | 0                                             | 最大重试次数                 |
| retry_count       | int          | 否       | 0                                             | 当前重试次数                 |
| last_execute_time | datetime     | 是       | NULL                                          | 上次执行时间                 |
| next_execute_time | datetime     | 是       | NULL                                          | 下次执行时间                 |
| creator_id        | bigint       | 否       | 无                                            | 创建人 ID                    |
| create_time       | datetime     | 否       | CURRENT_TIMESTAMP                             | 创建时间                     |
| update_time       | datetime     | 否       | CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP | 更新时间                     |

索引：

- PRIMARY KEY (`id`)
- KEY `idx_creator` (`creator_id`)
- KEY `idx_status` (`status`)
- KEY `idx_next_exec` (`next_execute_time`)

#### 执行记录表 (record)

执行记录表存储任务的每次执行记录。

| 字段名      | 数据类型     | 是否为空 | 默认值                                        | 说明                           |
| ----------- | ------------ | -------- | --------------------------------------------- | ------------------------------ |
| id          | bigint       | 否       | 自增                                          | 主键                           |
| task_id     | bigint       | 否       | 无                                            | 任务 ID                        |
| start_time  | datetime     | 否       | 无                                            | 开始执行时间                   |
| end_time    | datetime     | 是       | NULL                                          | 结束执行时间                   |
| status      | tinyint      | 否       | 0                                             | 状态：0-执行中，1-成功，2-失败 |
| result      | text         | 是       | NULL                                          | 执行结果                       |
| error       | text         | 是       | NULL                                          | 错误信息                       |
| executor    | varchar(100) | 是       | NULL                                          | 执行者标识                     |
| create_time | datetime     | 否       | CURRENT_TIMESTAMP                             | 创建时间                       |
| update_time | datetime     | 否       | CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP | 更新时间                       |

索引：

- PRIMARY KEY (`id`)
- KEY `idx_task_id` (`task_id`)
- KEY `idx_status` (`status`)

### 数据分表策略

DistributedJob 目前采用单表存储执行记录，但根据业务需求增长，可能在未来考虑以下优化方案：

1. **分区表**：可以考虑对 `record`表按时间范围进行分区，提高查询效率
2. **归档策略**：定期将历史记录归档到备份表中，保持主表数据量在可控范围
3. **冷热数据分离**：将常用的近期数据和不常用的历史数据分离存储

当数据量逐渐增长到百万级别时，建议实施上述优化措施。

### ER 图

下图展示了 DistributedJob 的实体关系图：

```
+-----------------+        +------------------+        +-----------------------+
|   department    |        |       task       |        |       record         |
+-----------------+        +------------------+        +-----------------------+
| id (PK)         |        | id (PK)          |        | id (PK)              |
| name            |        | name             |        | task_id (FK)         |
| description     |        | description      |------->| start_time           |
| parent_id       |        | cron_expression  |        | end_time             |
| status          |        | handler          |        | status               |
| create_time     |        | params           |        | result               |
| update_time     |        | status           |        | error                |
+-----------------+        | max_retry        |        | executor             |
       ^                   | retry_count      |        | create_time          |
       |                   | last_execute_time|        | update_time          |
       |                   | next_execute_time|        +-----------------------+
       |                   | creator_id (FK)  |
       |                   | create_time      |
       |                   | update_time      |
       |                   +------------------+
       |                          ^
       |                          |
       |                          |
+-----------------+              |          +--------------------+
|      user       |              |          |  role_permission   |
+-----------------+              |          +--------------------+
| id (PK)         |              |          | id (PK)            |
| username        |              |          | role_id (FK)       |
| password        |              |          | permission_id (FK) |
| real_name       |              |          | create_time        |
| email           |              +----------|--------------------+
| phone           |              |                   ^
| department_id(FK)--------------|                   |
| role_id (FK)    |              |                   |
| status          |--------------+                   |
| create_time     |                                  |
| update_time     |                                  |
+-----------------+                                  |
       |                                             |
       |                                             |
       v                                             |
+-----------------+                          +-----------------+
|      role       |                          |   permission    |
+-----------------+                          +-----------------+
| id (PK)         |------------------------->| id (PK)         |
| name            |                          | name            |
| description     |                          | code            |
| status          |                          | description     |
| create_time     |                          | status          |
| update_time     |                          | create_time     |
+-----------------+                          | update_time     |
                                            +-----------------+
```

### 数据库优化建议

#### 索引优化

- 任务表 (`task`) 已添加 `department_id`、`status`、`task_type` 字段的索引，用于优化常见查询场景
- 记录表已添加 `task_id`、`department_id`、`success`、`create_time`、`task_type` 字段的索引
- 如果经常按任务名称关键字查询，可考虑在 `task` 表的 `name` 字段上创建索引

#### 大数据量优化

- 记录表已按年月分表，但长期运行后仍可能有大量历史数据
- 建议实现自动归档策略，如保留最近 6 个月的记录，将更早的记录归档或清理
- 对于需要长期保存的记录，可导出到其他存储系统或归档数据库

#### 并发控制

- 任务调度采用乐观锁控制并发，确保同一任务不会被多个实例同时执行
- 在 MySQL 配置中适当调整 `max_connections` 参数，确保足够的连接数

#### 数据备份

- 定期备份数据库，保证数据安全
- 可使用 MySQL 自带的备份工具如 mysqldump 进行备份
- 示例备份命令：

  ```bash
  mysqldump -u username -p scheduler > scheduler_backup_$(date +%Y%m%d).sql
  ```

---

## 安装指南

### 系统要求

在安装 DistributedJob 之前，请确保您的系统满足以下要求：

| 组件             | 最低要求                           |
| ---------------- | ---------------------------------- |
| Go               | 1.16 或更高版本                    |
| MySQL            | 5.7 或更高版本                     |
| Node.js          | 16.0 或更高版本                    |
| 操作系统         | Windows、macOS 或 Linux            |
| 内存             | 2GB RAM（推荐）                    |
| 磁盘空间         | 应用程序 200MB，外加日志和数据空间 |
| gRPC             | 需要 gRPC 任务功能                 |
| Protocol Buffers | 用于 RPC 服务定义                  |

### 安装方法

#### 源码安装

从源代码构建允许您根据需要自定义和修改应用程序。

1. **克隆仓库**

   ```bash
   git clone https://github.com/username/distributedJob.git
   cd distributedJob
   ```

2. **构建应用程序**

   ```bash
   go build -o distributedJob ./cmd/server/main.go
   ```

3. **准备目录结构**

   确保以下目录结构：

   ```
   deployment-directory/
   ├── distributedJob      # 编译好的二进制文件
   ├── configs/
   │   └── config.yaml     # 配置文件
   └── web-ui/             # Web UI 文件
   ```

#### 二进制安装

对于快速部署，您可以下载预编译的二进制文件。

1. **下载发布版本**

   访问 [发布页面](https://github.com/username/distributedJob/releases) 并下载适合您操作系统的二进制文件。

2. **解压归档文件**

   ```bash
   # Linux/macOS
   tar -xzf distributedJob-v1.0.0-linux-amd64.tar.gz -C /opt/distributedJob

   # Windows
   # 使用您喜欢的解压工具解压到 C:\distributedJob
   ```

3. **验证结构**

   确保解压目录包含：

   - 可执行文件（`distributedJob` 或 `distributedJob.exe`）
   - 配置目录（`configs`）及 `config.yaml`
   - Web UI 目录（`web-ui`）

#### Docker 安装

使用 Docker 提供了跨不同平台的隔离一致环境。

1. **拉取 Docker 镜像**

   ```bash
   docker pull username/distributed-job:latest
   ```

   或者使用提供的 Dockerfile 构建自己的镜像：

   ```bash
   docker build -t distributed-job:latest .
   ```

2. **准备配置**

   为配置和数据库持久化创建本地目录：

   ```bash
   mkdir -p /data/distributed-job/configs
   # 将 config.yaml 复制到此目录
   cp config.yaml /data/distributed-job/configs/
   ```

### 配置

通过编辑 `config.yaml` 文件配置 DistributedJob：

#### 服务器配置

```yaml
server:
  port: 9088 # HTTP 服务端口
  contextPath: /v1 # API 基础路径
  timeout: 10 # HTTP 请求超时（秒）
```

#### RPC 服务器配置

```yaml
rpc:
  port: 9090 # gRPC 服务端口
  maxConcurrentStreams: 100 # 最大并发流
  keepAliveTime: 30 # keep-alive 时间（秒）
  keepAliveTimeout: 10 # keep-alive 超时（秒）
```

#### 数据库配置

```yaml
database:
  url: localhost:3306 # MySQL 服务器地址和端口
  username: root # 数据库用户名
  password: 123456 # 数据库密码
  schema: scheduler # 数据库名称
  maxConn: 10 # 最大连接数
  maxIdle: 5 # 最大空闲连接数
```

#### 日志配置

```yaml
log:
  path: ./log # 日志文件存储路径
  level: INFO # 日志级别（DEBUG、INFO、WARN、ERROR）
  maxSize: 100 # 单个日志文件的最大大小（MB）
  maxBackups: 10 # 日志文件备份的最大数量
  maxAge: 30 # 日志文件保留天数
```

#### 任务配置

```yaml
job:
  workers: 5 # 工作线程数
  queueSize: 100 # 任务队列大小
  httpWorkers: 3 # HTTP 任务工作线程数
  grpcWorkers: 2 # gRPC 任务工作线程数
```

#### 认证配置

```yaml
auth:
  jwtSecret: your-secret-key # JWT 密钥（请更改此项！）
  jwtExpireHours: 24 # JWT 过期时间（小时）
  adminUsername: admin # 默认管理员用户名
  adminPassword: admin123 # 默认管理员密码（请更改此项！）
```

### 数据库设置

DistributedJob 在首次启动时会自动创建必要的数据库表。但是，您也可以手动初始化数据库。

#### 手动数据库初始化

1. **创建数据库**

   ```sql
   CREATE DATABASE IF NOT EXISTS scheduler DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
   ```

2. **创建表**

   按照以下顺序执行 SQL 脚本：

   - [部门表](../scripts/init-db/init-department.sql)
   - [用户/角色/权限表](../scripts/init-db/init-user.sql)
   - [任务表](../scripts/init-db/init-task.sql)
   - [记录表模板](../scripts/init-db/init-record.sql)

3. **初始化默认数据**

   ```sql
   -- 插入默认权限
   INSERT INTO `permission` (`name`, `code`, `description`) VALUES
   ('任务查看', 'task:view', '查看任务的权限'),
   ('任务创建', 'task:create', '创建任务的权限'),
   ('任务更新', 'task:update', '编辑任务的权限'),
   ('任务删除', 'task:delete', '删除任务的权限'),
   ('记录查看', 'record:view', '查看执行记录的权限'),
   ('部门管理', 'department:manage', '管理部门的权限'),
   ('用户管理', 'user:manage', '管理用户的权限'),
   ('角色管理', 'role:manage', '管理角色的权限');

   -- 插入管理员角色
   INSERT INTO `role` (`name`, `description`) VALUES
   ('管理员', '拥有所有权限的系统管理员');

   -- 将管理员角色与所有权限关联
   INSERT INTO `role_permission` (`role_id`, `permission_id`)
   SELECT 1, id FROM `permission`;

   -- 插入默认部门
   INSERT INTO `department` (`name`, `description`, `parent_id`) VALUES
   ('总部', '总部', NULL);

   -- 插入管理员用户（生产环境中密码应加密）
   INSERT INTO `user` (`username`, `password`, `real_name`, `department_id`, `role_id`) VALUES
   ('admin', 'admin123', '系统管理员', 1, 1);
   ```

### 运行服务

#### Linux/macOS

```bash
# 导航到安装目录
cd /opt/distributedJob

# 运行服务
./distributedJob

# 作为后台服务运行
nohup ./distributedJob > /dev/null 2>&1 &
```

#### Windows

```cmd
# 导航到安装目录
cd C:\distributedJob

# 运行服务
distributedJob.exe
```

#### 使用 Systemd（Linux）

在 `/etc/systemd/system/distributed-job.service` 创建服务文件：

```ini
[Unit]
Description=DistributedJob 调度服务
After=network.target mysql.service

[Service]
Type=simple
User=distributed
WorkingDirectory=/opt/distributedJob
ExecStart=/opt/distributedJob/distributedJob
Restart=on-failure
RestartSec=5
LimitNOFILE=65536

[Install]
WantedBy=multi-user.target
```

启用并启动服务：

```bash
sudo systemctl daemon-reload
sudo systemctl enable distributed-job
sudo systemctl start distributed-job

# 检查状态
sudo systemctl status distributed-job
```

#### 使用 Docker

```bash
docker run -d \
  --name distributed-job \
  -p 9088:9088 \
  -v /data/distributed-job/configs:/app/configs \
  -v /data/distributed-job/log:/app/log \
  username/distributed-job:latest
```

### 验证

启动服务后，验证其是否正常运行：

1. **检查健康端点**

   ```bash
   curl http://localhost:9088/v1/health
   ```

   预期响应：

   ```json
   {
     "code": 0,
     "message": "success",
     "data": { "status": "up", "timestamp": "2023-01-01T12:00:00Z" }
   }
   ```

2. **访问 Web 控制台**

   打开浏览器并导航至：

   ```
   http://localhost:9088/v1/web/
   ```

   您应该会看到登录页面。使用默认凭据：

   - 用户名：`admin`
   - 密码：`admin123`

3. **检查日志**

   检查日志中是否有错误：

   ```bash
   # Linux/macOS
   tail -f /opt/distributedJob/log/app.log

   # Windows
   type C:\distributedJob\log\app.log
   ```

### 部署选项

#### 单实例部署

对于较小的环境或测试，单实例部署已足够：

1. **准备服务器**

   - 安装 MySQL 5.7+
   - 创建数据库和用户
   - 部署 DistributedJob 二进制文件和配置

2. **配置 MySQL 单实例**

   - 优化更高的连接限制：
     ```ini
     max_connections = 200
     innodb_buffer_pool_size = 1G
     ```

3. **运行服务**

   - 按照[运行服务](#运行服务)部分所述启动 DistributedJob

#### 多实例部署

为了高可用性和水平扩展，部署多个实例：

1. **共享数据库**

   - 配置所有实例使用相同的 MySQL 数据库
   - 确保数据库资源足够支持多个连接

2. **负载均衡器**

   - 在实例前设置负载均衡器（如 NGINX、HAProxy）
   - 配置健康检查以将流量路由到健康的实例

3. **配置一致性**

   - 在实例间使用相同的配置
   - 根据实例资源调整工作线程数量

4. **NGINX 配置示例**

   ```nginx
   upstream distributed_job {
       server instance1:9088;
       server instance2:9088;
       server instance3:9088;
   }

   server {
       listen 80;
       server_name job.example.com;

       location / {
           proxy_pass http://distributed_job;
           proxy_set_header Host $host;
           proxy_set_header X-Real-IP $remote_addr;
       }
   }
   ```

#### 容器化部署

使用 Docker 和 Docker Compose 部署提供了灵活性和隔离性：

1. **Docker Compose 配置**

   创建 `docker-compose.yml` 文件：

   ```yaml
   version: "3"

   services:
     mysql:
       image: mysql:8.0
       restart: always
       environment:
         MYSQL_ROOT_PASSWORD: root_password
         MYSQL_DATABASE: scheduler
         MYSQL_USER: distributed_job
         MYSQL_PASSWORD: distributed_job_password
       volumes:
         - mysql_data:/var/lib/mysql
         - ./scripts/init-db:/docker-entrypoint-initdb.d
       ports:
         - "3306:3306"

     distributed-job:
       image: username/distributed-job:latest
       restart: always
       depends_on:
         - mysql
       ports:
         - "9088:9088"
       volumes:
         - ./configs:/app/configs
         - ./log:/app/log
       environment:
         - TZ=UTC

   volumes:
     mysql_data:
   ```

2. **启动服务**

   ```bash
   docker-compose up -d
   ```

3. **扩展服务**

   ```bash
   docker-compose up -d --scale distributed-job=3
   ```

4. **生产环境建议**

   - 在生产环境中使用托管数据库服务而不是容器化 MySQL
   - 为数据持久性配置适当的卷挂载
   - 设置容器编排系统（如 Kubernetes）以实现高级扩展和管理

---

## API 文档

### API 概述

DistributedJob 提供了一套 RESTful API，用于管理定时任务、部门权限和查询执行记录。所有 API 都使用 JSON 格式进行数据交换，并返回统一的响应格式。

#### 基础 URL

所有 API 的基础路径为：`http://<host>:<port>/v1`

#### 统一响应格式

```json
{
  "code": 0, // 0 表示成功，非 0 表示错误
  "message": "", // 响应消息，成功时为 "success"，失败时为错误信息
  "data": null // 响应数据，可能是对象、数组或 null
}
```

#### 错误码说明

| 错误码 | 说明           |
| ------ | -------------- |
| 0      | 成功           |
| 4001   | 参数错误       |
| 4003   | 权限不足       |
| 4004   | 资源不存在     |
| 5000   | 服务器内部错误 |

#### 认证鉴权

除了健康检查接口外，所有 API 都需要通过认证鉴权。认证方式为基于 Token 的认证，Token 通过登录接口获取。

请求时需要在 HTTP Header 中添加 `Authorization` 字段：

```
Authorization: Bearer <token>
```

### 用户认证 API

#### 用户登录

```
POST /auth/login
Content-Type: application/json
```

请求参数：

| 参数名   | 类型   | 是否必填 | 说明   |
| -------- | ------ | -------- | ------ |
| username | string | 是       | 用户名 |
| password | string | 是       | 密码   |

请求示例：

```json
{
  "username": "admin",
  "password": "admin123"
}
```

响应示例：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
      "id": 1,
      "username": "admin",
      "realName": "系统管理员",
      "departmentId": 1,
      "departmentName": "技术部",
      "roleId": 1,
      "roleName": "管理员"
    }
  }
}
```

#### 刷新 Token

```
POST /auth/refresh
Authorization: Bearer <token>
```

响应示例：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }
}
```

#### 获取当前用户信息

```
GET /auth/userinfo
Authorization: Bearer <token>
```

响应示例：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "username": "admin",
    "realName": "系统管理员",
    "email": "admin@example.com",
    "phone": "13800138000",
    "departmentId": 1,
    "departmentName": "技术部",
    "roleId": 1,
    "roleName": "管理员",
    "permissions": [
      "task:create",
      "task:update",
      "task:delete",
      "task:view",
      "department:manage"
    ]
  }
}
```

### 部门管理 API

#### 获取部门列表

```
GET /departments?keyword={keyword}
Authorization: Bearer <token>
```

参数说明：

- `keyword` : 部门名称关键字（可选）

#### 获取部门详情

```
GET /departments/{id}
Authorization: Bearer <token>
```

参数说明：

- `id` : 部门 ID

#### 创建部门

```
POST /departments
Content-Type: application/json
Authorization: Bearer <token>
```

请求参数：

| 参数名      | 类型   | 是否必填 | 说明                    |
| ----------- | ------ | -------- | ----------------------- |
| name        | string | 是       | 部门名称                |
| description | string | 否       | 部门描述                |
| parentId    | number | 否       | 父部门 ID，顶级部门为空 |
| status      | number | 是       | 状态：0-禁用，1-启用    |

#### 更新部门

```
PUT /departments/{id}
Content-Type: application/json
Authorization: Bearer <token>
```

参数说明：

- `id` : 部门 ID

请求参数：同创建部门

#### 删除部门

```
DELETE /departments/{id}
Authorization: Bearer <token>
```

参数说明：

- `id` : 部门 ID

### 用户管理 API

#### 获取用户列表

```
GET /users?page={page}&size={size}&departmentId={departmentId}&keyword={keyword}
Authorization: Bearer <token>
```

参数说明：

- `page` : 页码，从 1 开始（可选，默认为 1）
- `size` : 每页记录数（可选，默认为 10）
- `departmentId` : 部门 ID（可选）
- `keyword` : 用户名或真实姓名关键字（可选）

#### 获取用户详情

```
GET /users/{id}
Authorization: Bearer <token>
```

参数说明：

- `id` : 用户 ID

#### 创建用户

```
POST /users
Content-Type: application/json
Authorization: Bearer <token>
```

请求参数：

| 参数名       | 类型   | 是否必填 | 说明                 |
| ------------ | ------ | -------- | -------------------- |
| username     | string | 是       | 用户名               |
| password     | string | 是       | 密码                 |
| realName     | string | 是       | 真实姓名             |
| email        | string | 否       | 电子邮箱             |
| phone        | string | 否       | 手机号码             |
| departmentId | number | 是       | 所属部门 ID          |
| roleId       | number | 是       | 角色 ID              |
| status       | number | 是       | 状态：0-禁用，1-启用 |

#### 更新用户

```
PUT /users/{id}
Content-Type: application/json
Authorization: Bearer <token>
```

参数说明：

- `id` : 用户 ID

请求参数：同创建用户，但 `password` 为可选

#### 删除用户

```
DELETE /users/{id}
Authorization: Bearer <token>
```

参数说明：

- `id` : 用户 ID

#### 修改用户密码

```
PATCH /users/{id}/password
Content-Type: application/json
Authorization: Bearer <token>
```

参数说明：

- `id` : 用户 ID

请求参数：

| 参数名      | 类型   | 是否必填 | 说明   |
| ----------- | ------ | -------- | ------ |
| newPassword | string | 是       | 新密码 |
| oldPassword | string | 是       | 原密码 |

### 角色与权限管理 API

#### 获取角色列表

```
GET /roles?page={page}&size={size}&keyword={keyword}
Authorization: Bearer <token>
```

参数说明：

- `page` : 页码，从 1 开始（可选，默认为 1）
- `size` : 每页记录数（可选，默认为 10）
- `keyword` : 角色名称关键字（可选）

#### 获取角色详情

```
GET /roles/{id}
Authorization: Bearer <token>
```

参数说明：

- `id` : 角色 ID

#### 创建角色

```
POST /roles
Content-Type: application/json
Authorization: Bearer <token>
```

请求参数：

| 参数名      | 类型     | 是否必填 | 说明                 |
| ----------- | -------- | -------- | -------------------- |
| name        | string   | 是       | 角色名称             |
| description | string   | 否       | 角色描述             |
| permissions | number[] | 是       | 权限 ID 数组         |
| status      | number   | 是       | 状态：0-禁用，1-启用 |

#### 更新角色

```
PUT /roles/{id}
Content-Type: application/json
Authorization: Bearer <token>
```

参数说明：

- `id` : 角色 ID

请求参数：同创建角色

#### 删除角色

```
DELETE /roles/{id}
Authorization: Bearer <token>
```

参数说明：

- `id` : 角色 ID

#### 获取所有权限列表

```
GET /permissions
Authorization: Bearer <token>
```

### 任务管理 API

#### 获取任务列表

```
GET /tasks?page={page}&size={size}&keyword={keyword}&departmentId={departmentId}&taskType={taskType}
Authorization: Bearer <token>
```

参数说明：

- `page` : 页码，从 1 开始（可选，默认为 1）
- `size` : 每页记录数（可选，默认为 10）
- `keyword` : 任务名称关键字（可选）
- `departmentId` : 部门 ID（可选）
- `taskType` : 任务类型，HTTP 或 GRPC（可选）

#### 获取任务详情

```
GET /tasks/{id}
Authorization: Bearer <token>
```

参数说明：

- `id` : 任务 ID

#### 创建 HTTP 任务

```
POST /tasks/http
Content-Type: application/json
Authorization: Bearer <token>
```

请求参数：

| 参数名        | 类型   | 是否必填 | 说明                                       |
| ------------- | ------ | -------- | ------------------------------------------ |
| name          | string | 是       | 任务名称                                   |
| departmentId  | number | 是       | 所属部门 ID                                |
| cron          | string | 是       | cron 表达式                                |
| url           | string | 是       | 调度 URL                                   |
| httpMethod    | string | 是       | HTTP 方法（GET、POST、PUT、PATCH、DELETE） |
| body          | string | 否       | 请求体                                     |
| headers       | string | 否       | 请求头（JSON 格式字符串）                  |
| retryCount    | number | 否       | 最大重试次数                               |
| retryInterval | number | 否       | 重试间隔（秒）                             |
| fallbackUrl   | string | 否       | 备用 URL                                   |
| status        | number | 是       | 状态：0-禁用，1-启用                       |

#### 创建 gRPC 任务

```
POST /tasks/grpc
Content-Type: application/json
Authorization: Bearer <token>
```

请求参数：

| 参数名              | 类型   | 是否必填 | 说明                   |
| ------------------- | ------ | -------- | ---------------------- |
| name                | string | 是       | 任务名称               |
| departmentId        | number | 是       | 所属部门 ID            |
| cron                | string | 是       | cron 表达式            |
| grpcService         | string | 是       | gRPC 服务名            |
| grpcMethod          | string | 是       | gRPC 方法名            |
| grpcParams          | string | 否       | gRPC 参数(JSON 字符串) |
| retryCount          | number | 否       | 最大重试次数           |
| retryInterval       | number | 否       | 重试间隔（秒）         |
| fallbackGrpcService | string | 否       | 备用 gRPC 服务名       |
| fallbackGrpcMethod  | string | 否       | 备用 gRPC 方法名       |
| status              | number | 是       | 状态：0-禁用，1-启用   |

#### 更新 HTTP 任务

```
PUT /tasks/http/{id}
Content-Type: application/json
Authorization: Bearer <token>
```

参数说明：

- `id` : HTTP 任务 ID

请求参数：同创建 HTTP 任务

#### 更新 gRPC 任务

```
PUT /tasks/grpc/{id}
Content-Type: application/json
Authorization: Bearer <token>
```

参数说明：

- `id` : gRPC 任务 ID

请求参数：同创建 gRPC 任务

#### 删除任务

```
DELETE /tasks/{id}
Authorization: Bearer <token>
```

参数说明：

- `id` : 任务 ID

#### 修改任务状态

```
PATCH /tasks/{id}/status
Content-Type: application/json
Authorization: Bearer <token>
```

参数说明：

- `id` : 任务 ID

请求参数：

| 参数名 | 类型   | 是否必填 | 说明                 |
| ------ | ------ | -------- | -------------------- |
| status | number | 是       | 状态：0-禁用，1-启用 |

### 执行记录查询 API

#### 获取任务执行记录

```
GET /records?taskId={taskId}&departmentId={departmentId}&year={year}&month={month}&page={page}&size={size}&success={success}&taskType={taskType}
Authorization: Bearer <token>
```

参数说明：

- `taskId` : 任务 ID（可选）
- `departmentId` : 部门 ID（可选）
- `year` : 年份，如 2025（必填）
- `month` : 月份，如 1-12（必填）
- `page` : 页码，从 1 开始（可选，默认为 1）
- `size` : 每页记录数（可选，默认为 10）
- `success` : 是否成功，1-成功，0-失败（可选）
- `taskType` : 任务类型，HTTP 或 GRPC（可选）

#### 获取记录详情

```
GET /records/{id}?year={year}&month={month}
Authorization: Bearer <token>
```

参数说明：

- `id` : 记录 ID
- `year` : 年份，如 2025（必填）
- `month` : 月份，如 1-12（必填）

#### 获取任务执行历史统计

```
GET /records/stats?taskId={taskId}&departmentId={departmentId}&year={year}&month={month}
Authorization: Bearer <token>
```

参数说明：

- `taskId` : 任务 ID（可选）
- `departmentId` : 部门 ID（可选）
- `year` : 年份，如 2025（必填）
- `month` : 月份，如 1-12（必填）

### 健康检查与服务管理 API

#### 获取服务健康状态

```
GET /health
```

正常情况：HTTP 状态码返回 `200`

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "status": "up",
    "timestamp": "2023-01-01T12:00:00Z"
  }
}
```

#### 平滑关闭服务实例

```
GET /shutdown?wait={wait}
```

参数说明：

- `wait` : 等待关闭时间（单位-秒），必须大于 0

**注意**：该接口仅限本机调用（只能使用 `localhost`、`127.0.0.1`、`0.0.0.0` 这三个 hostname 访问）

### RPC 服务 API

DistributedJob 现提供以下 RPC 服务 API:

#### 任务调度 RPC 服务

| 方法名        | 描述                 | 请求参数                                          | 响应参数                                                           |
| ------------- | -------------------- | ------------------------------------------------- | ------------------------------------------------------------------ |
| ScheduleTask  | 调度一个新任务       | name, cron_expression, handler, params, max_retry | task_id, success, message                                          |
| PauseTask     | 暂停一个运行中的任务 | task_id                                           | success, message                                                   |
| ResumeTask    | 恢复一个已暂停的任务 | task_id                                           | success, message                                                   |
| GetTaskStatus | 获取任务当前状态     | task_id                                           | task_id, status, last_execute_time, next_execute_time, retry_count |

#### 用户认证 RPC 服务

| 方法名             | 描述          | 请求参数           | 响应参数                         |
| ------------------ | ------------- | ------------------ | -------------------------------- |
| Authenticate       | 验证用户凭证  | username, password | token, user_id, success, message |
| ValidateToken      | 验证 JWT 令牌 | token              | valid, user_id, permissions      |
| GetUserPermissions | 获取用户权限  | user_id            | permissions, success             |

#### 数据访问 RPC 服务

| 方法名         | 描述             | 请求参数                                     | 响应参数                                     |
| -------------- | ---------------- | -------------------------------------------- | -------------------------------------------- |
| GetTaskHistory | 获取任务执行历史 | task_id, start_time, end_time, limit, offset | records, total, success                      |
| GetStatistics  | 获取系统统计数据 | department_id, period                        | task_count, success_rate, avg_execution_time |

---

## 测试指南

### 测试架构

DistributedJob 采用多层测试策略，确保系统的稳定性和可靠性。测试架构设计如下：

```
┌─────────────────────────────────────────────┐
│              端到端测试 (E2E)                │
│                                             │
│  ┌─────────────────────────────────────────┐│
│  │           集成测试 (Integration)         ││
│  │                                         ││
│  │  ┌─────────────────────────────────────┐││
│  │  │          单元测试 (Unit)            │││
│  │  └─────────────────────────────────────┘││
│  └─────────────────────────────────────────┘│
└─────────────────────────────────────────────┘
```

### 单元测试

单元测试是测试系统最基础的层级，主要聚焦于测试单个功能单元（如函数或方法）。DistributedJob 的单元测试采用 Go 标准库的 `testing` 包和 `github.com/stretchr/testify` 库进行辅助。

#### 核心业务逻辑测试

对系统的核心业务逻辑进行充分测试是确保系统稳定性的关键。以下是 DistributedJob 核心业务逻辑的测试覆盖范围：

1. **任务调度服务测试**

   ```go
   // internal/service/test/task_service_test.go
   package test

   import (
       "testing"
       "time"

       "distributedJob/internal/job"
       "distributedJob/internal/model/entity"
       "distributedJob/internal/service"
       "github.com/stretchr/testify/assert"
       "github.com/stretchr/testify/mock"
   )

   // 模拟任务仓库
   type MockTaskRepository struct {
       mock.Mock
   }

   // 实现任务仓库接口的方法
   func (m *MockTaskRepository) GetTaskByID(id int64) (*entity.Task, error) {
       args := m.Called(id)
       if args.Get(0) == nil {
           return nil, args.Error(1)
       }
       return args.Get(0).(*entity.Task), args.Error(1)
   }

   func (m *MockTaskRepository) CreateTask(task *entity.Task) (int64, error) {
       args := m.Called(task)
       return args.Get(0).(int64), args.Error(1)
   }

   // 其他必要方法实现...

   func TestCreateHTTPTask(t *testing.T) {
       // 创建模拟对象
       mockRepo := new(MockTaskRepository)
       mockScheduler := job.NewScheduler(nil)

       // 设置模拟行为
       mockTask := &entity.Task{
           Name:      "Test HTTP Task",
           Cron:      "*/5 * * * *",
           TaskType:  "HTTP",
           Status:    1,
           Params:    `{"url":"http://example.com","method":"GET"}`,
           CreatorID: 1,
       }
       mockRepo.On("CreateTask", mock.AnythingOfType("*entity.Task")).Return(int64(1), nil)

       // 创建服务实例
       taskService := service.NewTaskService(mockRepo, mockScheduler)

       // 执行测试
       taskID, err := taskService.CreateHTTPTask(mockTask)

       // 验证结果
       assert.NoError(t, err)
       assert.Equal(t, int64(1), taskID)
       mockRepo.AssertExpectations(t)
   }

   // 更多任务服务测试...
   ```

2. **认证服务测试**

   ```go
   // internal/service/test/auth_service_test.go
   package test

   import (
       "testing"
       "time"

       "distributedJob/internal/model/entity"
       "distributedJob/internal/service"
       "github.com/stretchr/testify/assert"
       "github.com/stretchr/testify/mock"
       "golang.org/x/crypto/bcrypt"
   )

   // 模拟用户仓库
   type MockUserRepository struct {
       mock.Mock
   }

   // 实现必要的接口方法
   func (m *MockUserRepository) GetUserByUsername(username string) (*entity.User, error) {
       args := m.Called(username)
       if args.Get(0) == nil {
           return nil, args.Error(1)
       }
       return args.Get(0).(*entity.User), args.Error(1)
   }

   // 其他必要方法实现...

   func TestLogin(t *testing.T) {
       // 创建模拟对象
       mockUserRepo := new(MockUserRepository)
       mockRoleRepo := new(MockRoleRepository)
       mockDeptRepo := new(MockDepartmentRepository)
       mockPermRepo := new(MockPermissionRepository)

       // 生成加密密码
       hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

       // 设置模拟行为
       mockUser := &entity.User{
           ID:           1,
           Username:     "testuser",
           Password:     string(hashedPassword),
           DepartmentID: 1,
           RoleID:       1,
       }
       mockUserRepo.On("GetUserByUsername", "testuser").Return(mockUser, nil)

       // 创建服务实例
       authService := service.NewAuthService(
           mockUserRepo,
           mockRoleRepo,
           mockDeptRepo,
           mockPermRepo,
           "test-secret",
           time.Hour*24,
       )

       // 执行测试
       token, err := authService.Login("testuser", "password123")

       // 验证结果
       assert.NoError(t, err)
       assert.NotEmpty(t, token)
       mockUserRepo.AssertExpectations(t)
   }

   // 更多认证服务测试...
   ```

3. **调度器测试**

   ```go
   // internal/job/test/scheduler_test.go
   package test

   import (
       "testing"
       "time"

       "distributedJob/internal/job"
       "distributedJob/internal/model/entity"
       "github.com/stretchr/testify/assert"
       "github.com/stretchr/testify/mock"
   )

   // 模拟TaskRepository
   type MockTaskRepository struct {
       mock.Mock
   }

   // 实现必要的接口方法
   func (m *MockTaskRepository) UpdateTaskStatus(id int64, status int8) error {
       args := m.Called(id, status)
       return args.Error(0)
   }

   func (m *MockTaskRepository) GetTaskByID(id int64) (*entity.Task, error) {
       args := m.Called(id)
       if args.Get(0) == nil {
           return nil, args.Error(1)
       }
       return args.Get(0).(*entity.Task), args.Error(1)
   }

   // 其他必要方法实现...

   func TestPauseTask(t *testing.T) {
       // 创建模拟对象
       mockRepo := new(MockTaskRepository)

       // 创建调度器
       scheduler := job.NewScheduler(nil)
       scheduler.SetTaskRepository(mockRepo)

       // 设置模拟任务
       taskID := int64(1)
       task := &entity.Task{
           ID:       taskID,
           Name:     "Test Task",
           Cron:     "*/5 * * * *",
           Status:   1,
           TaskType: "HTTP",
       }

       // 设置模拟行为
       mockRepo.On("GetTaskByID", taskID).Return(task, nil)
       mockRepo.On("UpdateTaskStatus", taskID, int8(0)).Return(nil)

       // 添加任务到调度器
       scheduler.AddTask(task)

       // 执行测试
       err := scheduler.PauseTask(taskID)

       // 验证结果
       assert.NoError(t, err)
       mockRepo.AssertExpectations(t)
   }

   // 更多调度器测试...
   ```

4. **存储库测试**

   ```go
   // internal/store/mysql/repository/test/task_repository_test.go
   package test

   import (
       "testing"
       "time"

       "github.com/DATA-DOG/go-sqlmock"
       "distributedJob/internal/model/entity"
       "distributedJob/internal/store/mysql/repository"
       "github.com/stretchr/testify/assert"
       "gorm.io/driver/mysql"
       "gorm.io/gorm"
   )

   func TestGetTaskByID(t *testing.T) {
       // 创建sqlmock
       db, mock, err := sqlmock.New()
       assert.NoError(t, err)
       defer db.Close()

       // 转换为gorm.DB
       gormDB, err := gorm.Open(mysql.New(mysql.Config{
           Conn:                      db,
           SkipInitializeWithVersion: true,
       }), &gorm.Config{})
       assert.NoError(t, err)

       // 创建仓库实例
       taskRepo := repository.NewTaskRepository(gormDB)

       // 设置模拟查询预期
       taskID := int64(1)
       rows := sqlmock.NewRows([]string{"id", "name", "cron", "task_type", "status", "params"}).
           AddRow(taskID, "Test Task", "*/5 * * * *", "HTTP", 1, `{"url":"http://example.com"}`)

       mock.ExpectQuery("SELECT (.+) FROM `tasks` WHERE").WithArgs(taskID).WillReturnRows(rows)

       // 执行测试
       task, err := taskRepo.GetTaskByID(taskID)

       // 验证结果
       assert.NoError(t, err)
       assert.NotNil(t, task)
       assert.Equal(t, "Test Task", task.Name)
       assert.NoError(t, mock.ExpectationsWereMet())
   }

   // 更多存储库测试...
   ```

5. **RPC 服务测试**

   ```go
   // internal/rpc/server/test/task_scheduler_server_test.go
   package test

   import (
       "context"
       "testing"

       "distributedJob/internal/job"
       "distributedJob/internal/model/entity"
       pb "distributedJob/internal/rpc/proto"
       "distributedJob/internal/rpc/server"
       "github.com/stretchr/testify/assert"
       "github.com/stretchr/testify/mock"
   )

   // 模拟Scheduler
   type MockScheduler struct {
       mock.Mock
   }

   func (m *MockScheduler) AddTaskAndStore(task *entity.Task) (int64, error) {
       args := m.Called(task)
       return args.Get(0).(int64), args.Error(1)
   }

   func (m *MockScheduler) PauseTask(taskID int64) error {
       args := m.Called(taskID)
       return args.Error(0)
   }

   // 其他必要方法实现...

   func TestScheduleTask(t *testing.T) {
       // 创建模拟调度器
       mockScheduler := new(MockScheduler)

       // 设置模拟行为
       mockScheduler.On("AddTaskAndStore", mock.AnythingOfType("*entity.Task")).Return(int64(1), nil)

       // 创建RPC服务器
       taskServer := server.NewTaskSchedulerServer(mockScheduler)

       // 创建请求
       req := &pb.ScheduleTaskRequest{
           Name:           "Test Task",
           CronExpression: "*/5 * * * *",
           Handler:        "http",
           Params:         []byte(`{"url":"http://example.com"}`),
           MaxRetry:       3,
       }

       // 执行测试
       resp, err := taskServer.ScheduleTask(context.Background(), req)

       // 验证结果
       assert.NoError(t, err)
       assert.Equal(t, int64(1), resp.TaskId)
       assert.True(t, resp.Success)
       mockScheduler.AssertExpectations(t)
   }

   // 更多RPC服务测试...
   ```

### 集成测试

集成测试验证系统的多个组件在一起工作时的正确性，通常会测试整个功能流程。

#### 核心业务流程测试

```go
// internal/test/integration/task_workflow_test.go
package integration

import (
    "testing"
    "time"

    "distributedJob/internal/config"
    "distributedJob/internal/job"
    "distributedJob/internal/model/entity"
    "distributedJob/internal/service"
    "distributedJob/internal/store/mysql"
    "github.com/stretchr/testify/assert"
)

func setupTestEnvironment(t *testing.T) (*service.TaskService, *job.Scheduler, func()) {
    // 加载测试配置
    cfg := &config.Config{
        Database: config.Database{
            URL:      "localhost:3306",
            Username: "test",
            Password: "test",
            Schema:   "test_scheduler",
        },
        // 其他必要配置...
    }

    // 初始化数据库连接
    db, err := mysql.InitDB(cfg)
    if err != nil {
        t.Fatalf("Failed to connect to database: %v", err)
    }

    // 初始化存储库管理器
    repoManager := mysql.NewRepositoryManager(db)

    // 初始化调度器
    scheduler, err := job.NewScheduler(cfg)
    if err != nil {
        t.Fatalf("Failed to create scheduler: %v", err)
    }
    scheduler.SetTaskRepository(repoManager.Task())

    // 创建任务服务
    taskService := service.NewTaskService(repoManager.Task(), scheduler)

    // 清理函数
    cleanup := func() {
        // 清理测试数据
        db.Exec("DELETE FROM tasks WHERE name LIKE 'Test%'")
        db.Exec("DELETE FROM records WHERE task_id IN (SELECT id FROM tasks WHERE name LIKE 'Test%')")
    }

    return taskService, scheduler, cleanup
}

func TestTaskCreationAndExecution(t *testing.T) {
    // 设置测试环境
    taskService, scheduler, cleanup := setupTestEnvironment(t)
    defer cleanup()

    // 启动调度器
    err := scheduler.Start()
    assert.NoError(t, err)
    defer scheduler.Stop()

    // 创建HTTP任务
    httpTask := &entity.Task{
        Name:      "Test Integration HTTP Task",
        Cron:      "* * * * *", // 每分钟执行一次
        TaskType:  "HTTP",
        Status:    1, // 启用
        Params:    `{"url":"http://example.com/test","method":"GET"}`,
        CreatorID: 1,
    }

    // 添加任务
    taskID, err := taskService.CreateHTTPTask(httpTask)
    assert.NoError(t, err)
    assert.Greater(t, taskID, int64(0))

    // 等待任务执行
    time.Sleep(65 * time.Second)

    // 验证任务执行记录
    task, err := taskService.GetTaskByID(taskID)
    assert.NoError(t, err)
    assert.NotNil(t, task.LastExecuteTime)

    // 暂停任务
    err = taskService.UpdateTaskStatus(taskID, 0) // 0表示暂停
    assert.NoError(t, err)

    // 验证任务已暂停
    task, err = taskService.GetTaskByID(taskID)
    assert.NoError(t, err)
    assert.Equal(t, int8(0), task.Status)
}
```

#### API 端点测试

```go
// internal/api/test/task_api_test.go
package test

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"

    "distributedJob/internal/api"
    "distributedJob/internal/config"
    "distributedJob/internal/job"
    "distributedJob/internal/store/mysql"
    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/assert"
)

func setupAPITestEnvironment(t *testing.T) (http.Handler, func()) {
    // 设置测试模式
    gin.SetMode(gin.TestMode)

    // 加载测试配置
    cfg := &config.Config{
        Database: config.Database{
            URL:      "localhost:3306",
            Username: "test",
            Password: "test",
            Schema:   "test_scheduler",
        },
        Auth: config.Auth{
            JwtSecret:     "test-secret",
            JwtExpireHours: 24,
        },
        // 其他必要配置...
    }

    // 初始化数据库连接
    db, err := mysql.InitDB(cfg)
    if err != nil {
        t.Fatalf("Failed to connect to database: %v", err)
    }

    // 初始化存储库管理器
    repoManager := mysql.NewRepositoryManager(db)

    // 初始化调度器
    scheduler, err := job.NewScheduler(cfg)
    if err != nil {
        t.Fatalf("Failed to create scheduler: %v", err)
    }

    // 创建API服务器
    server := api.NewServer(cfg, scheduler, repoManager)

    // 清理函数
    cleanup := func() {
        // 清理测试数据
        db.Exec("DELETE FROM tasks WHERE name LIKE 'Test%'")
        db.Exec("DELETE FROM records WHERE task_id IN (SELECT id FROM tasks WHERE name LIKE 'Test%')")
    }

    return server.Router(), cleanup
}

func TestCreateHttpTask(t *testing.T) {
    // 设置API测试环境
    router, cleanup := setupAPITestEnvironment(t)
    defer cleanup()

    // 登录以获取令牌
    loginPayload := map[string]string{
        "username": "admin",
        "password": "admin123",
    }
    loginBody, _ := json.Marshal(loginPayload)
    loginReq := httptest.NewRequest("POST", "/v1/auth/login", bytes.NewBuffer(loginBody))
    loginReq.Header.Set("Content-Type", "application/json")

    loginResp := httptest.NewRecorder()
    router.ServeHTTP(loginResp, loginReq)

    assert.Equal(t, http.StatusOK, loginResp.Code)

    var loginResult map[string]interface{}
    err := json.Unmarshal(loginResp.Body.Bytes(), &loginResult)
    assert.NoError(t, err)

    data := loginResult["data"].(map[string]interface{})
    token := data["token"].(string)

    // 创建HTTP任务
    taskPayload := map[string]interface{}{
        "name":         "Test API HTTP Task",
        "departmentId": 1,
        "cron":         "*/5 * * * *",
        "url":          "http://example.com/test",
        "httpMethod":   "GET",
        "retryCount":   3,
        "status":       1,
    }
    taskBody, _ := json.Marshal(taskPayload)

    req := httptest.NewRequest("POST", "/v1/tasks/http", bytes.NewBuffer(taskBody))
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer "+token)

    resp := httptest.NewRecorder()
    router.ServeHTTP(resp, req)

    // 验证结果
    assert.Equal(t, http.StatusOK, resp.Code)

    var result map[string]interface{}
    err = json.Unmarshal(resp.Body.Bytes(), &result)
    assert.NoError(t, err)

    assert.Equal(t, float64(0), result["code"])
    assert.Equal(t, "success", result["message"])
    assert.NotNil(t, result["data"])
}
```

### 性能测试

性能测试用于验证系统在不同负载下的表现，包括响应时间、吞吐量和资源使用情况。

```go
// internal/test/performance/scheduler_benchmark_test.go
package performance

import (
    "testing"
    "time"

    "distributedJob/internal/config"
    "distributedJob/internal/job"
    "distributedJob/internal/model/entity"
)

func BenchmarkTaskScheduling(b *testing.B) {
    // 创建调度器
    cfg := &config.Config{
        Job: config.Job{
            Workers:     5,
            QueueSize:   100,
            HTTPWorkers: 3,
            GRPCWorkers: 2,
        },
    }
    scheduler, _ := job.NewScheduler(cfg)
    scheduler.Start()
    defer scheduler.Stop()

    // 准备基准测试
    b.ResetTimer()

    // 运行基准测试
    for i := 0; i < b.N; i++ {
        task := &entity.Task{
            ID:       int64(i + 1),
            Name:     "Benchmark Task",
            Cron:     "* * * * *",
            TaskType: "HTTP",
            Status:   1,
            Params:   `{"url":"http://example.com/benchmark"}`,
        }
        scheduler.AddTask(task)
    }
}

func BenchmarkParallelTaskProcessing(b *testing.B) {
    // 创建调度器
    cfg := &config.Config{
        Job: config.Job{
            Workers:     10,  // 增加工作线程
            QueueSize:   1000, // 增加队列大小
            HTTPWorkers: 5,
            GRPCWorkers: 5,
        },
    }
    scheduler, _ := job.NewScheduler(cfg)
    scheduler.Start()
    defer scheduler.Stop()

    // 准备基准测试
    taskCount := 100
    for i := 0; i < taskCount; i++ {
        task := &entity.Task{
            ID:       int64(i + 1),
            Name:     "Parallel Task",
            Cron:     "* * * * *",
            TaskType: "HTTP",
            Status:   1,
            Params:   `{"url":"http://example.com/parallel"}`,
        }
        scheduler.AddTask(task)
    }

    b.ResetTimer()

    // 运行并行基准测试
    b.RunParallel(func(pb *testing.PB) {
        i := 0
        for pb.Next() {
            taskID := int64((i % taskCount) + 1)
            scheduler.GetTaskStatus(taskID)
            i++
        }
    })
}
```

### 测试自动化

为了简化测试流程，我们提供了自动化测试脚本：

```bash
#!/bin/bash
# scripts/run-tests.sh

# 设置测试环境变量
export TEST_ENV=true
export TEST_DB_URL=localhost:3306
export TEST_DB_USER=test
export TEST_DB_PASS=test
export TEST_DB_NAME=test_scheduler

# 运行单元测试
echo "Running unit tests..."
go test -v ./internal/service/test/...
go test -v ./internal/job/test/...
go test -v ./internal/store/mysql/repository/test/...
go test -v ./internal/rpc/server/test/...

# 运行集成测试
echo "Running integration tests..."
go test -v ./internal/test/integration/...

# 运行API测试
echo "Running API tests..."
go test -v ./internal/api/test/...

# 运行基准测试
echo "Running benchmark tests..."
go test -v -bench=. ./internal/test/performance/...

# 生成测试覆盖率报告
echo "Generating test coverage report..."
go test -coverprofile=coverage.out ./internal/...
go tool cover -html=coverage.out -o coverage.html

echo "Tests completed. See coverage.html for coverage report."
```

### 覆盖率分析

使用 Go 内置的覆盖率工具监控代码测试覆盖率：

```bash
# 生成覆盖率报告
go test -coverprofile=coverage.out ./internal/...

# 查看HTML格式的覆盖率报告
go tool cover -html=coverage.out

# 查看文本格式的覆盖率报告
go tool cover -func=coverage.out
```

#### 覆盖率目标

为了确保代码质量，我们为不同层次的代码设置了以下覆盖率目标：

1. **核心业务逻辑（服务层）**: 80%+
2. **数据存储层**: 70%+
3. **API 层**: 60%+
4. **RPC 服务层**: 70%+
5. **工具与辅助函数**: 50%+

#### 核心业务测试重点

针对 DistributedJob 的核心业务，测试应特别关注以下方面：

1. **调度逻辑**

   - 任务添加/删除/暂停/恢复的正确性
   - Cron 表达式解析和执行时间计算
   - 并发任务处理
   - 任务重试机制

2. **任务执行**

   - HTTP 任务的执行与结果处理
   - gRPC 任务的执行与结果处理
   - 失败重试策略
   - 备份机制激活

3. **认证授权**

   - 用户认证流程
   - JWT 令牌生成与验证
   - 权限检查
   - 安全相关功能（密码哈希等）

4. **数据一致性**

   - 任务状态同步
   - 执行记录与任务状态的一致性
   - 并发操作下的数据一致性

#### 测试数据管理

为了确保测试的可靠性和可重复性，我们使用以下策略管理测试数据：

1. 对于单元测试，使用模拟对象和存根
2. 对于存储库测试，使用 `go-sqlmock` 模拟数据库交互
3. 对于集成测试，使用专用的测试数据库
4. 每次测试前清理测试数据，测试后恢复初始状态
5. 使用测试固件（fixtures）提供一致的测试数据集

#### 测试与 CI/CD 集成

将测试流程集成到持续集成管道中，确保代码更改不会破坏现有功能：

```yaml
# .github/workflows/test.yml
name: Test

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main, develop]

jobs:
  test:
    runs-on: ubuntu-latest
    services:
      mysql:
        image: mysql:8.0
        env:
          MYSQL_ROOT_PASSWORD: root
          MYSQL_DATABASE: test_scheduler
          MYSQL_USER: test
          MYSQL_PASSWORD: test
        ports:
          - 3306:3306
        options: --health-cmd="mysqladmin ping" --health-interval=10s --health-timeout=5s --health-retries=3

    steps:
      - uses: actions/checkout@v2

      - name: Set up Go
        uses: actions/setup-go@v2
        with:
          go-version: 1.16

      - name: Install dependencies
        run: go mod download

      - name: Run unit tests
        run: go test -v ./internal/service/test/... ./internal/job/test/... ./internal/store/mysql/repository/test/... ./internal/rpc/server/test/...

      - name: Run integration tests
        run: go test -v ./internal/test/integration/...

      - name: Generate coverage report
        run: go test -coverprofile=coverage.out ./internal/...

      - name: Upload coverage to Codecov
        uses: codecov/codecov-action@v1
```

---

## 前端开发

### 技术栈

DistributedJob 前端应用采用现代化的前端技术栈：

- **构建工具**: Vite 4.x
- **前端框架**: Vue 3.x (使用 Composition API)
- **类型系统**: TypeScript 4.x
- **UI 组件库**: Element Plus 2.x
- **状态管理**: Pinia 2.x
- **路由管理**: Vue Router 4.x
- **HTTP 客户端**: Axios 1.x
- **CSS 预处理器**: SCSS
- **代码规范**: ESLint + Prettier
- **打包工具**: Rollup (由 Vite 内置)
- **图表可视化**: ECharts 5.x

### 前端项目结构

```
web-ui/
├── public/                  # 静态资源
│   ├── favicon.ico          # 网站图标
│   └── assets/              # 其他静态资源
├── src/                     # 源代码
│   ├── api/                 # API 请求模块
│   │   ├── auth.ts          # 认证相关 API
│   │   ├── department.ts    # 部门相关 API
│   │   ├── http.ts          # Axios 实例配置
│   │   ├── record.ts        # 执行记录相关 API
│   │   ├── role.ts          # 角色权限相关 API
│   │   ├── task.ts          # 任务相关 API
│   │   └── user.ts          # 用户相关 API
│   ├── assets/              # 资源文件
│   │   ├── images/          # 图片资源
│   │   └── styles/          # 样式文件
│   │       └── main.scss    # 主样式文件
│   ├── components/          # 通用组件
│   │   ├── common/          # 公共组件
│   │   │   ├── Pagination.vue  # 分页组件
│   │   │   ├── SearchForm.vue  # 搜索表单组件
│   │   │   └── StatusTag.vue   # 状态标签组件
│   │   └── layout/          # 布局组件
│   │       ├── AppLink.vue     # 应用链接组件
│   │       ├── AppMain.vue     # 主内容区组件
│   │       ├── Breadcrumb.vue  # 面包屑组件
│   │       ├── Layout.vue      # 整体布局组件
│   │       ├── SidebarItem.vue # 侧边栏项组件
│   │       └── TabsView.vue    # 标签页视图组件
│   ├── router/              # 路由配置
│   │   └── index.ts         # 路由定义
│   ├── store/               # 状态管理
│   │   ├── modules/         # 状态模块
│   │   │   ├── app.ts       # 应用状态模块
│   │   │   ├── permission.ts # 权限状态模块
│   │   │   ├── settings.ts  # 设置状态模块
│   │   │   ├── tagsView.ts  # 标签页状态模块
│   │   │   └── user.ts      # 用户状态模块
│   │   └── index.ts         # 状态入口
│   ├── utils/               # 工具函数
│   │   ├── auth.ts          # 认证相关工具
│   │   ├── date.ts          # 日期处理工具
│   │   ├── token.ts         # Token 相关工具
│   │   └── validate.ts      # 表单验证工具
│   ├── views/               # 页面视图组件
│   │   ├── auth/            # 认证相关页面
│   │   │   ├── Login.vue    # 登录页面
│   │   │   └── NotFound.vue # 404页面
│   │   ├── dashboard/       # 仪表板页面
│   │   │   └── Index.vue    # 首页仪表板
│   │   ├── department/      # 部门管理页面
│   │   │   └── List.vue     # 部门列表页
│   │   ├── record/          # 执行记录页面
│   │   │   ├── Detail.vue   # 记录详情页
│   │   │   └── List.vue     # 记录列表页
│   │   ├── role/            # 角色管理页面
│   │   │   └── List.vue     # 角色列表页
│   │   ├── task/            # 任务管理页面
│   │   │   ├── Edit.vue     # 任务编辑页
│   │   │   └── List.vue     # 任务列表页
│   │   └── user/            # 用户管理页面
│   │       └── List.vue     # 用户列表页
│   ├── App.vue              # 应用入口组件
│   ├── env.d.ts             # 环境声明文件
│   └── main.ts              # 应用入口TS文件
├── index.html               # HTML 入口文件
├── package.json             # 依赖配置
├── tsconfig.json            # TypeScript 配置
├── tsconfig.node.json       # Node.js TypeScript 配置
└── vite.config.ts           # Vite 配置文件
```

### 前端功能模块

DistributedJob 前端应用主要包含以下功能模块：

#### 1. 认证与授权

- 登录表单 - 用户名/密码认证
- Token 管理 - JWT token 存储与刷新
- 权限控制 - 基于角色的权限控制
- 路由守卫 - 拦截未授权访问

```typescript
// src/utils/token.ts
const TokenKey = "Admin-Token";

export function getToken(): string {
  return localStorage.getItem(TokenKey) || "";
}

export function setToken(token: string): void {
  return localStorage.setItem(TokenKey, token);
}

export function removeToken(): void {
  return localStorage.removeItem(TokenKey);
}

// src/api/auth.ts
import request from "./http";
import { LoginData, UserInfo } from "../types";

export function login(data: LoginData) {
  return request({
    url: "/auth/login",
    method: "post",
    data,
  });
}

export function getUserInfo() {
  return request({
    url: "/auth/userinfo",
    method: "get",
  });
}

export function logout() {
  return request({
    url: "/auth/logout",
    method: "post",
  });
}
```

#### 2. 布局系统

- 响应式布局 - 适配不同屏幕尺寸
- 侧边菜单 - 可折叠导航菜单
- 标签页视图 - 多标签切换功能
- 面包屑导航 - 显示当前页面位置

```vue
<!-- src/components/layout/Layout.vue -->
<template>
  <div class="app-wrapper">
    <div class="sidebar-container">
      <div class="logo">DistributedJob</div>
      <el-scrollbar>
        <el-menu
          :default-active="activeMenu"
          background-color="#304156"
          text-color="#bfcbd9"
          active-text-color="#409EFF"
        >
          <sidebar-item
            v-for="route in permission_routes"
            :key="route.path"
            :item="route"
            :base-path="route.path"
          />
        </el-menu>
      </el-scrollbar>
    </div>
    <div class="main-container">
      <div class="navbar">
        <breadcrumb class="breadcrumb-container" />
        <div class="right-menu">
          <el-dropdown trigger="click">
            <span class="el-dropdown-link">
              {{ userInfo.username }}
              <el-icon class="el-icon--right"><arrow-down /></el-icon>
            </span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item @click="handleLogout"
                  >退出登录</el-dropdown-item
                >
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </div>
      <tabs-view />
      <app-main />
    </div>
  </div>
</template>
```

#### 3. 任务管理

- 任务列表 - 分页展示所有任务
- 任务创建 - 创建 HTTP/gRPC 任务
- 任务编辑 - 修改任务配置
- 任务操作 - 启用/禁用/删除任务
- Cron 表达式验证 - 检查 cron 表达式合法性

```typescript
// src/api/task.ts
import request from "./http";
import { TaskQuery, TaskData } from "../types";

export function getTasks(params: TaskQuery) {
  return request({
    url: "/tasks",
    method: "get",
    params,
  });
}

export function getTaskById(id: number) {
  return request({
    url: `/tasks/${id}`,
    method: "get",
  });
}

export function createHttpTask(data: TaskData) {
  return request({
    url: "/tasks/http",
    method: "post",
    data,
  });
}

export function createGrpcTask(data: TaskData) {
  return request({
    url: "/tasks/grpc",
    method: "post",
    data,
  });
}

export function updateTaskStatus(id: number, status: number) {
  return request({
    url: `/tasks/${id}/status`,
    method: "patch",
    data: { status },
  });
}

export function deleteTask(id: number) {
  return request({
    url: `/tasks/${id}`,
    method: "delete",
  });
}
```

#### 4. 执行记录分析

- 记录列表 - 按任务、时间等筛选
- 记录详情 - 查看执行详细信息
- 执行统计 - 成功率、平均耗时等指标
- 图表可视化 - 使用 ECharts 展示执行趋势

```vue
<!-- src/views/record/List.vue (部分代码) -->
<template>
  <div class="app-container">
    <div class="filter-container">
      <el-form :inline="true" :model="queryParams" class="demo-form-inline">
        <el-form-item label="任务ID">
          <el-input
            v-model="queryParams.taskId"
            placeholder="任务ID"
            clearable
          />
        </el-form-item>
        <el-form-item label="部门">
          <el-select
            v-model="queryParams.departmentId"
            placeholder="所属部门"
            clearable
          >
            <el-option
              v-for="dept in departmentOptions"
              :key="dept.id"
              :label="dept.name"
              :value="dept.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="执行状态">
          <el-select
            v-model="queryParams.success"
            placeholder="执行状态"
            clearable
          >
            <el-option label="成功" :value="1" />
            <el-option label="失败" :value="0" />
          </el-select>
        </el-form-item>
        <el-form-item label="时间">
          <el-date-picker
            v-model="dateRange"
            type="daterange"
            range-separator="至"
            start-placeholder="开始日期"
            end-placeholder="结束日期"
            value-format="YYYY-MM-DD"
            @change="handleDateRangeChange"
          />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleQuery">查询</el-button>
          <el-button @click="resetQuery">重置</el-button>
        </el-form-item>
      </el-form>
    </div>

    <!-- 图表区域 -->
    <div class="chart-container">
      <div ref="executionChart" style="width: 100%; height: 300px"></div>
    </div>

    <!-- 表格区域 -->
    <el-table
      v-loading="loading"
      :data="recordList"
      stripe
      border
      style="width: 100%"
      @selection-change="handleSelectionChange"
    >
      <el-table-column type="selection" width="55" align="center" />
      <el-table-column label="记录ID" prop="id" width="80" align="center" />
      <el-table-column
        label="任务名称"
        prop="taskName"
        min-width="150"
        show-overflow-tooltip
      />
      <el-table-column
        label="所属部门"
        prop="departmentName"
        width="120"
        align="center"
      />
      <el-table-column
        label="开始时间"
        prop="startTime"
        width="180"
        align="center"
      />
      <el-table-column
        label="结束时间"
        prop="endTime"
        width="180"
        align="center"
      />
      <el-table-column label="状态" align="center" width="100">
        <template #default="{ row }">
          <el-tag :type="row.status === 1 ? 'success' : 'danger'">
            {{ row.status === 1 ? "成功" : "失败" }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" align="center" width="150">
        <template #default="{ row }">
          <el-button size="small" type="primary" @click="handleViewDetail(row)"
            >查看详情</el-button
          >
        </template>
      </el-table-column>
    </el-table>

    <!-- 分页器 -->
    <pagination
      v-show="total > 0"
      :total="total"
      v-model:page="queryParams.page"
      v-model:limit="queryParams.size"
      @pagination="getList"
    />
  </div>
</template>

<script lang="ts" setup>
import { ref, onMounted, reactive, toRefs } from "vue";
import { ElMessage } from "element-plus";
import * as echarts from "echarts/core";
import { getRecordList, getRecordStats } from "@/api/record";
import { getDepartments } from "@/api/department";
import Pagination from "@/components/common/Pagination.vue";

// 查询条件
const queryState = reactive({
  queryParams: {
    page: 1,
    size: 10,
    taskId: undefined,
    departmentId: undefined,
    success: undefined,
    year: new Date().getFullYear(),
    month: new Date().getMonth() + 1,
  },
  dateRange: [],
  departmentOptions: [],
  recordList: [],
  total: 0,
  loading: false,
  selectedRows: [],
});

const {
  queryParams,
  dateRange,
  departmentOptions,
  recordList,
  total,
  loading,
  selectedRows,
} = toRefs(queryState);

// 查询方法
const getList = async () => {
  queryState.loading = true;
  try {
    const { data } = await getRecordList(queryParams.value);
    queryState.recordList = data.list;
    queryState.total = data.total;
    initChart();
  } catch (error) {
    ElMessage.error("获取记录列表失败");
  } finally {
    queryState.loading = false;
  }
};

// 初始化图表
const executionChart = ref<HTMLDivElement | null>(null);
const chartInstance = ref<echarts.ECharts | null>(null);

const initChart = async () => {
  if (!executionChart.value) return;

  try {
    const { data } = await getRecordStats({
      taskId: queryParams.value.taskId,
      departmentId: queryParams.value.departmentId,
      year: queryParams.value.year,
      month: queryParams.value.month,
    });

    if (!chartInstance.value) {
      chartInstance.value = echarts.init(executionChart.value);
    }

    chartInstance.value.setOption({
      title: {
        text: "任务执行统计",
      },
      tooltip: {
        trigger: "axis",
      },
      legend: {
        data: ["成功次数", "失败次数", "成功率"],
      },
      xAxis: {
        type: "category",
        data: data.dates,
      },
      yAxis: [
        {
          type: "value",
          name: "次数",
          position: "left",
        },
        {
          type: "value",
          name: "成功率",
          min: 0,
          max: 100,
          position: "right",
          axisLabel: {
            formatter: "{value}%",
          },
        },
      ],
      series: [
        {
          name: "成功次数",
          type: "bar",
          stack: "总量",
          data: data.success,
        },
        {
          name: "失败次数",
          type: "bar",
          stack: "总量",
          data: data.fail,
        },
        {
          name: "成功率",
          type: "line",
          yAxisIndex: 1,
          data: data.successRate.map((rate: number) =>
            parseFloat(rate.toFixed(2))
          ),
          markLine: {
            data: [{ type: "average", name: "平均成功率" }],
          },
        },
      ],
    });
  } catch (error) {
    ElMessage.error("获取统计数据失败");
  }
};

// 处理查询
const handleQuery = () => {
  queryParams.value.page = 1;
  getList();
};

// 重置查询
const resetQuery = () => {
  queryParams.value = {
    page: 1,
    size: 10,
    taskId: undefined,
    departmentId: undefined,
    success: undefined,
    year: new Date().getFullYear(),
    month: new Date().getMonth() + 1,
  };
  dateRange.value = [];
  getList();
};

// 处理日期范围变化
const handleDateRangeChange = () => {
  if (dateRange.value && dateRange.value.length === 2) {
    const [start, end] = dateRange.value;
    const startDate = new Date(start);
    queryParams.value.year = startDate.getFullYear();
    queryParams.value.month = startDate.getMonth() + 1;
  }
};

// 查看详情
const handleViewDetail = (row: any) => {
  router.push(
    `/record/detail/${row.id}?year=${queryParams.value.year}&month=${queryParams.value.month}`
  );
};

// 加载部门列表
const loadDepartments = async () => {
  try {
    const { data } = await getDepartments();
    queryState.departmentOptions = data;
  } catch (error) {
    ElMessage.error("获取部门列表失败");
  }
};

// 选择行变化
const handleSelectionChange = (selection: any[]) => {
  queryState.selectedRows = selection;
};

// 初始化
onMounted(() => {
  loadDepartments();
  getList();
  window.addEventListener("resize", () => {
    chartInstance.value?.resize();
  });
});
</script>
```

#### 5. 部门与用户管理

- 部门树形结构 - 展示部门层级关系
- 部门 CRUD - 创建/修改/删除部门
- 用户列表 - 按部门筛选用户列表
- 用户 CRUD - 创建/修改/删除用户
- 角色分配 - 为用户分配角色

```typescript
// src/api/department.ts
import request from "./http";

export function getDepartments(params: any) {
  return request({
    url: "/departments",
    method: "get",
    params,
  });
}

export function getDepartmentById(id: number) {
  return request({
    url: `/departments/${id}`,
    method: "get",
  });
}

export function createDepartment(data: any) {
  return request({
    url: "/departments",
    method: "post",
    data,
  });
}

export function updateDepartment(id: number, data: any) {
  return request({
    url: `/departments/${id}`,
    method: "put",
    data,
  });
}

export function deleteDepartment(id: number) {
  return request({
    url: `/departments/${id}`,
    method: "delete",
  });
}
```

#### 6. 角色与权限管理

- 角色列表 - 展示系统中的角色
- 角色 CRUD - 创建/修改/删除角色
- 权限分配 - 为角色分配权限
- 权限树 - 树形结构展示权限

```vue
<!-- src/views/role/List.vue (部分代码) -->
<template>
  <div class="app-container">
    <div class="filter-container">
      <el-form :inline="true" :model="queryParams">
        <el-form-item label="角色名称">
          <el-input
            v-model="queryParams.keyword"
            placeholder="角色名称"
            clearable
          />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleQuery">查询</el-button>
          <el-button @click="resetQuery">重置</el-button>
          <el-button type="success" @click="handleAdd">新增角色</el-button>
        </el-form-item>
      </el-form>
    </div>

    <el-table
      v-loading="loading"
      :data="roleList"
      stripe
      border
      style="width: 100%"
    >
      <el-table-column label="角色ID" prop="id" width="80" align="center" />
      <el-table-column label="角色名称" prop="name" min-width="120" />
      <el-table-column
        label="描述"
        prop="description"
        min-width="180"
        show-overflow-tooltip
      />
      <el-table-column label="状态" width="100" align="center">
        <template #default="{ row }">
          <el-tag :type="row.status === 1 ? 'success' : 'info'">
            {{ row.status === 1 ? "启用" : "禁用" }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column
        label="创建时间"
        prop="createTime"
        width="180"
        align="center"
      />
      <el-table-column
        label="更新时间"
        prop="updateTime"
        width="180"
        align="center"
      />
      <el-table-column label="操作" align="center" width="250">
        <template #default="{ row }">
          <el-button size="small" type="primary" @click="handleEdit(row)"
            >编辑</el-button
          >
          <el-button size="small" type="success" @click="handlePermission(row)"
            >分配权限</el-button
          >
          <el-button size="small" type="danger" @click="handleDelete(row)"
            >删除</el-button
          >
        </template>
      </el-table-column>
    </el-table>

    <pagination
      v-show="total > 0"
      :total="total"
      v-model:page="queryParams.page"
      v-model:limit="queryParams.size"
      @pagination="getList"
    />

    <!-- 添加/编辑角色对话框 -->
    <el-dialog :title="dialogTitle" v-model="dialogVisible" width="500px">
      <el-form
        ref="roleFormRef"
        :model="roleForm"
        :rules="rules"
        label-width="100px"
      >
        <el-form-item label="角色名称" prop="name">
          <el-input v-model="roleForm.name" placeholder="请输入角色名称" />
        </el-form-item>
        <el-form-item label="角色描述" prop="description">
          <el-input
            v-model="roleForm.description"
            type="textarea"
            placeholder="请输入角色描述"
          />
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-radio-group v-model="roleForm.status">
            <el-radio :label="1">启用</el-radio>
            <el-radio :label="0">禁用</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="dialogVisible = false">取 消</el-button>
          <el-button type="primary" @click="submitForm">确 定</el-button>
        </span>
      </template>
    </el-dialog>

    <!-- 权限分配对话框 -->
    <el-dialog title="分配权限" v-model="permissionDialogVisible" width="600px">
      <el-tree
        ref="permissionTreeRef"
        :data="permissionTree"
        :props="{ label: 'name', children: 'children' }"
        show-checkbox
        node-key="id"
        :default-checked-keys="selectedPermissions"
      />
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="permissionDialogVisible = false">取 消</el-button>
          <el-button type="primary" @click="submitPermissions">确 定</el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>
```

#### 7. 仪表盘

- 系统概览 - 任务、部门、用户数量统计
- 执行情况 - 任务执行成功率饼图
- 执行趋势 - 近期执行量趋势图
- 资源使用 - 系统资源使用情况

```vue
<!-- src/views/dashboard/Index.vue (部分代码) -->
<template>
  <div class="dashboard-container">
    <el-row :gutter="20">
      <!-- 统计卡片 -->
      <el-col :xs="24" :sm="12" :md="6">
        <el-card class="stat-card">
          <div class="card-panel">
            <div class="card-panel-icon-wrapper">
              <el-icon class="card-panel-icon">
                <calendar />
              </el-icon>
            </div>
            <div class="card-panel-description">
              <div class="card-panel-text">任务总数</div>
              <div class="card-panel-num">{{ dashboardData.totalTasks }}</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="24" :sm="12" :md="6">
        <el-card class="stat-card">
          <div class="card-panel">
            <div
              class="card-panel-icon-wrapper"
              style="background: rgba(0,199,139,0.1)"
            >
              <el-icon class="card-panel-icon" style="color: #00C78B">
                <check-circle />
              </el-icon>
            </div>
            <div class="card-panel-description">
              <div class="card-panel-text">运行中任务</div>
              <div class="card-panel-num">{{ dashboardData.runningTasks }}</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="24" :sm="12" :md="6">
        <el-card class="stat-card">
          <div class="card-panel">
            <div
              class="card-panel-icon-wrapper"
              style="background: rgba(238,99,99,0.1)"
            >
              <el-icon class="card-panel-icon" style="color: #EE6363">
                <warning />
              </el-icon>
            </div>
            <div class="card-panel-description">
              <div class="card-panel-text">失败任务</div>
              <div class="card-panel-num">{{ dashboardData.failedTasks }}</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="24" :sm="12" :md="6">
        <el-card class="stat-card">
          <div class="card-panel">
            <div
              class="card-panel-icon-wrapper"
              style="background: rgba(84,112,198,0.1)"
            >
              <el-icon class="card-panel-icon" style="color: #5470C6">
                <user />
              </el-icon>
            </div>
            <div class="card-panel-description">
              <div class="card-panel-text">用户总数</div>
              <div class="card-panel-num">{{ dashboardData.totalUsers }}</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20" style="margin-top: 20px">
      <!-- 任务执行情况图表 -->
      <el-col :xs="24" :sm="24" :md="12">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>任务执行情况</span>
            </div>
          </template>
          <div ref="executionPieChart" style="width: 100%; height: 350px"></div>
        </el-card>
      </el-col>

      <!-- 执行趋势图表 -->
      <el-col :xs="24" :sm="24" :md="12">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>近7天执行趋势</span>
            </div>
          </template>
          <div ref="trendLineChart" style="width: 100%; height: 350px"></div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20" style="margin-top: 20px">
      <!-- 最近执行记录 -->
      <el-col :xs="24" :sm="24" :md="24">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>最近执行记录</span>
              <el-button
                class="more-button"
                text
                @click="router.push('/record/list')"
                >查看更多</el-button
              >
            </div>
          </template>
          <el-table
            :data="recentRecords"
            style="width: 100%"
            :row-class-name="tableRowClassName"
          >
            <el-table-column
              label="任务名称"
              prop="taskName"
              min-width="150"
              show-overflow-tooltip
            />
            <el-table-column
              label="任务类型"
              prop="taskType"
              width="100"
              align="center"
            >
              <template #default="{ row }">
                <el-tag :type="row.taskType === 'HTTP' ? 'success' : 'warning'">
                  {{ row.taskType }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column
              label="执行时间"
              prop="startTime"
              width="180"
              align="center"
            />
            <el-table-column label="执行耗时" width="120" align="center">
              <template #default="{ row }">
                {{ calculateDuration(row.startTime, row.endTime) }}
              </template>
            </el-table-column>
            <el-table-column label="状态" width="100" align="center">
              <template #default="{ row }">
                <el-tag :type="row.status === 1 ? 'success' : 'danger'">
                  {{ row.status === 1 ? "成功" : "失败" }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="120" align="center">
              <template #default="{ row }">
                <el-button
                  size="small"
                  type="primary"
                  @click="viewRecordDetail(row)"
                  >查看详情</el-button
                >
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script lang="ts" setup>
import { ref, onMounted, reactive, toRefs, onBeforeUnmount } from "vue";
import { useRouter } from "vue-router";
import * as echarts from "echarts/core";
import { getDashboardStats } from "@/api/dashboard";
import { getRecentRecords } from "@/api/record";
import { formatDate, diffTime } from "@/utils/date";

const router = useRouter();

// 仪表盘数据
const state = reactive({
  dashboardData: {
    totalTasks: 0,
    runningTasks: 0,
    failedTasks: 0,
    totalUsers: 0,
  },
  recentRecords: [],
});

const { dashboardData, recentRecords } = toRefs(state);

// 图表引用
const executionPieChart = ref<HTMLDivElement | null>(null);
const trendLineChart = ref<HTMLDivElement | null>(null);
const pieChartInstance = ref<echarts.ECharts | null>(null);
const lineChartInstance = ref<echarts.ECharts | null>(null);

// 加载仪表盘数据
const loadDashboardData = async () => {
  try {
    const { data } = await getDashboardStats();
    state.dashboardData = data.stats;

    // 加载最近记录
    const recordResponse = await getRecentRecords({ limit: 10 });
    state.recentRecords = recordResponse.data.records;

    // 初始化图表
    initPieChart(data.execution);
    initLineChart(data.trend);
  } catch (error) {
    console.error("加载仪表盘数据失败:", error);
  }
};

// 初始化饼图
const initPieChart = (data: any) => {
  if (!executionPieChart.value) return;

  pieChartInstance.value = echarts.init(executionPieChart.value);

  pieChartInstance.value.setOption({
    tooltip: {
      trigger: "item",
      formatter: "{a} <br/>{b}: {c} ({d}%)",
    },
    legend: {
      orient: "vertical",
      left: "left",
      data: ["成功", "失败", "超时", "其他异常"],
    },
    series: [
      {
        name: "执行情况",
        type: "pie",
        radius: ["50%", "70%"],
        avoidLabelOverlap: false,
        itemStyle: {
          borderRadius: 10,
          borderColor: "#fff",
          borderWidth: 2,
        },
        label: {
          show: false,
          position: "center",
        },
        emphasis: {
          label: {
            show: true,
            fontSize: "16",
            fontWeight: "bold",
          },
        },
        labelLine: {
          show: false,
        },
        data: [
          {
            value: data.success,
            name: "成功",
            itemStyle: { color: "#00C78B" },
          },
          { value: data.fail, name: "失败", itemStyle: { color: "#EE6363" } },
          {
            value: data.timeout,
            name: "超时",
            itemStyle: { color: "#FF9900" },
          },
          {
            value: data.error,
            name: "其他异常",
            itemStyle: { color: "#909399" },
          },
        ],
      },
    ],
  });
};

// 初始化趋势图
const initLineChart = (data: any) => {
  if (!trendLineChart.value) return;

  lineChartInstance.value = echarts.init(trendLineChart.value);

  lineChartInstance.value.setOption({
    tooltip: {
      trigger: "axis",
    },
    legend: {
      data: ["总执行次数", "成功次数", "失败次数"],
    },
    grid: {
      left: "3%",
      right: "4%",
      bottom: "3%",
      containLabel: true,
    },
    xAxis: {
      type: "category",
      boundaryGap: false,
      data: data.dates,
    },
    yAxis: {
      type: "value",
    },
    series: [
      {
        name: "总执行次数",
        type: "line",
        stack: "Total",
        data: data.total,
        areaStyle: {},
        emphasis: {
          focus: "series",
        },
      },
      {
        name: "成功次数",
        type: "line",
        stack: "Total",
        data: data.success,
        areaStyle: {},
        emphasis: {
          focus: "series",
        },
      },
      {
        name: "失败次数",
        type: "line",
        stack: "Total",
        data: data.fail,
        areaStyle: {},
        emphasis: {
          focus: "series",
        },
      },
    ],
  });
};

// 计算执行时长
const calculateDuration = (start: string, end: string): string => {
  if (!start || !end) return "-";
  return diffTime(new Date(start), new Date(end));
};

// 记录行样式
const tableRowClassName = ({ row }: { row: any }): string => {
  return row.status === 0 ? "error-row" : "";
};

// 查看记录详情
const viewRecordDetail = (row: any) => {
  const date = new Date(row.startTime);
  router.push(
    `/record/detail/${row.id}?year=${date.getFullYear()}&month=${
      date.getMonth() + 1
    }`
  );
};

// 生命周期钩子
onMounted(() => {
  loadDashboardData();

  window.addEventListener("resize", handleResize);
});

onBeforeUnmount(() => {
  window.removeEventListener("resize", handleResize);
  pieChartInstance.value?.dispose();
  lineChartInstance.value?.dispose();
});

// 处理窗口大小变化
const handleResize = () => {
  pieChartInstance.value?.resize();
  lineChartInstance.value?.resize();
};
</script>

<style lang="scss" scoped>
.dashboard-container {
  padding: 20px;

  .stat-card {
    margin-bottom: 20px;

    .card-panel {
      display: flex;
      justify-content: space-between;
      align-items: center;

      &-icon-wrapper {
        width: 60px;
        height: 60px;
        border-radius: 10px;
        background: rgba(84, 112, 198, 0.1);
        display: flex;
        justify-content: center;
        align-items: center;
      }

      &-icon {
        font-size: 28px;
        color: #5470c6;
      }

      &-description {
        text-align: right;

        .card-panel-text {
          color: #909399;
          font-size: 14px;
          margin-bottom: 5px;
        }

        .card-panel-num {
          font-size: 24px;
          font-weight: bold;
        }
      }
    }
  }

  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;

    .more-button {
      color: #409eff;
      font-size: 14px;
    }
  }

  :deep(.error-row) {
    background-color: #ffeeee;
  }
}
</style>
```

### 前端开发环境配置

#### 开发环境准备

1. **Node.js 环境**

   确保安装了 Node.js 16.x 或更高版本：

   ```bash
   # 检查 Node 版本
   node -v

   # 检查 npm 版本
   npm -v
   ```

2. **代码编辑器**

   推荐使用 Visual Studio Code，并安装以下扩展：

   - Volar (Vue 语言支持)
   - ESLint
   - Prettier
   - TypeScript Vue Plugin

3. **安装依赖**

   ```bash
   cd web-ui
   npm install
   ```

#### 启动开发服务器

```bash
npm run dev
```

这将在 `http://localhost:5173` 启动开发服务器。

#### 开发环境配置文件

1. **Vite 配置**

   ```typescript
   // vite.config.ts
   import { defineConfig } from "vite";
   import vue from "@vitejs/plugin-vue";
   import { resolve } from "path";

   export default defineConfig({
     plugins: [vue()],
     resolve: {
       alias: {
         "@": resolve(__dirname, "src"),
       },
     },
     server: {
       port: 5173,
       open: true,
       proxy: {
         "/v1": {
           target: "http://localhost:9088",
           changeOrigin: true,
           rewrite: (path) => path,
         },
       },
     },
     css: {
       preprocessorOptions: {
         scss: {
           additionalData: `@import "@/assets/styles/variables.scss";`,
         },
       },
     },
   });
   ```

2. **TypeScript 配置**

   ```json
   // tsconfig.json
   {
     "compilerOptions": {
       "target": "ESNext",
       "useDefineForClassFields": true,
       "module": "ESNext",
       "moduleResolution": "Node",
       "strict": true,
       "jsx": "preserve",
       "resolveJsonModule": true,
       "isolatedModules": true,
       "esModuleInterop": true,
       "lib": ["ESNext", "DOM"],
       "skipLibCheck": true,
       "noEmit": true,
       "baseUrl": ".",
       "paths": {
         "@/*": ["src/*"]
       }
     },
     "include": [
       "src/**/*.ts",
       "src/**/*.d.ts",
       "src/**/*.tsx",
       "src/**/*.vue"
     ],
     "references": [{ "path": "./tsconfig.node.json" }]
   }
   ```

3. **ESLint 配置**

   ```js
   // .eslintrc.js
   module.exports = {
     root: true,
     env: {
       browser: true,
       es2021: true,
       node: true,
     },
     extends: [
       "plugin:vue/vue3-recommended",
       "eslint:recommended",
       "@vue/typescript/recommended",
       "prettier",
     ],
     parserOptions: {
       ecmaVersion: 2021,
     },
     rules: {
       "no-console": process.env.NODE_ENV === "production" ? "warn" : "off",
       "no-debugger": process.env.NODE_ENV === "production" ? "warn" : "off",
       "vue/no-multiple-template-root": "off",
       "@typescript-eslint/no-explicit-any": "off",
       "@typescript-eslint/explicit-module-boundary-types": "off",
     },
   };
   ```

### 前端构建与部署

#### 生产环境构建

```bash
# 构建生产版本
npm run build
```

构建结果将输出到 `dist` 目录。

#### 预览构建结果

```bash
# 预览构建结果
npm run preview
```

#### 部署策略

1. **与后端集成部署**

   将前端构建产物复制到后端项目中，通过 Go 服务提供静态文件：

   ```go
   // internal/api/server.go
   func (s *Server) setupStaticFiles() {
       s.engine.Static("/v1/web", "./web-ui/dist")
       s.engine.NoRoute(func(c *gin.Context) {
           c.File("./web-ui/dist/index.html")
       })
   }
   ```

2. **独立部署**

   也可以将前端应用部署到 Nginx 或其他静态文件服务器：

   ```nginx
   # nginx.conf
   server {
       listen 80;
       server_name job.example.com;

       location / {
           root /var/www/distributed-job/web;
           try_files $uri $uri/ /index.html;
           index index.html;
       }

       location /v1/ {
           proxy_pass http://backend:9088;
           proxy_set_header Host $host;
           proxy_set_header X-Real-IP $remote_addr;
       }
   }
   ```

3. **Docker 部署**

   使用多阶段构建创建包含前端的镜像：

   ```dockerfile
   # 构建前端
   FROM node:16 AS web-builder
   WORKDIR /app/web-ui
   COPY web-ui/package*.json ./
   RUN npm install
   COPY web-ui .
   RUN npm run build

   # 最终镜像
   FROM nginx:1.21-alpine
   COPY --from=web-builder /app/web-ui/dist /usr/share/nginx/html
   COPY nginx.conf /etc/nginx/conf.d/default.conf
   EXPOSE 80
   CMD ["nginx", "-g", "daemon off;"]
   ```

### 前端性能优化

#### 代码分割

利用 Vite 的动态导入功能实现代码分割，减小主包大小：

```typescript
// src/router/index.ts
const routes = [
  {
    path: "/",
    component: Layout,
    redirect: "/dashboard",
    children: [
      {
        path: "dashboard",
        name: "Dashboard",
        component: () => import("@/views/dashboard/Index.vue"), // 动态导入
        meta: { title: "首页", icon: "el-icon-s-home" },
      },
    ],
  },
  {
    path: "/task",
    component: Layout,
    meta: { title: "任务管理", icon: "el-icon-s-order" },
    children: [
      {
        path: "list",
        name: "TaskList",
        component: () => import("@/views/task/List.vue"), // 动态导入
        meta: { title: "任务列表" },
      },
      {
        path: "edit/:id?",
        name: "TaskEdit",
        component: () => import("@/views/task/Edit.vue"), // 动态导入
        meta: { title: "编辑任务", activeMenu: "/task/list" },
        hidden: true,
      },
    ],
  },
];
```

#### 组件懒加载

对于不在首屏的组件，使用 Vue 的异步组件：

```typescript
// src/components/index.ts
import { defineAsyncComponent } from "vue";

// 异步加载组件
export const JsonEditor = defineAsyncComponent(
  () => import("./widgets/JsonEditor.vue")
);

// 异步加载带加载状态的组件
export const TaskChart = defineAsyncComponent({
  loader: () => import("./charts/TaskChart.vue"),
  delay: 200,
  loadingComponent: () => import("./common/LoadingComponent.vue"),
});
```

#### 静态资源优化

1. **图片优化**

   使用 `vite-plugin-imagemin` 插件压缩图片：

   ```typescript
   // vite.config.ts
   import viteImagemin from "vite-plugin-imagemin";

   export default defineConfig({
     plugins: [
       vue(),
       viteImagemin({
         gifsicle: {
           optimizationLevel: 7,
           interlaced: false,
         },
         optipng: {
           optimizationLevel: 7,
         },
         mozjpeg: {
           quality: 80,
         },
         pngquant: {
           quality: [0.8, 0.9],
           speed: 4,
         },
         svgo: {
           plugins: [
             {
               name: "removeViewBox",
             },
             {
               name: "removeEmptyAttrs",
               active: false,
             },
           ],
         },
       }),
     ],
   });
   ```

2. **CDN 加速**

   通过外部引入大型依赖库减小包体积：

   ```typescript
   // vite.config.ts
   export default defineConfig({
     build: {
       rollupOptions: {
         external: ["echarts"],
         output: {
           globals: {
             echarts: "echarts",
           },
         },
       },
     },
   });
   ```

   ```html
   <!-- index.html -->
   <head>
     <!-- CDN 引入 -->
     <script src="https://cdn.jsdelivr.net/npm/echarts@5.4.0/dist/echarts.min.js"></script>
   </head>
   ```

---

## 令牌安全机制

### 令牌概述

DistributedJob 系统采用基于 JWT (JSON Web Token) 的认证机制，通过实现双令牌系统来提高系统安全性，有效防范令牌滥用、XSS 和 CSRF 攻击。本系统采用的认证方式符合行业安全标准，保障用户和系统的安全。

### 双令牌机制

系统实现正规的长短令牌机制，具体包括：

1. **Access Token (短期令牌)**

   - 短期有效（默认设置为 30 分钟）
   - 用于日常 API 访问认证
   - 实现无状态验证
   - 携带最小权限信息

2. **Refresh Token (长期令牌)**

   - 长期有效（默认设置为 7 天）
   - 仅用于获取新的 Access Token
   - 通过安全的 HttpOnly Cookie 传输
   - 每次刷新会生成新的 Refresh Token

**实现细节**：

```go
// 生成访问令牌 (Access Token)
func (s *AuthService) generateAccessToken(user *entity.User) (string, error) {
    nowTime := time.Now()
    expireTime := nowTime.Add(time.Duration(s.jwtExpireMinutes) * time.Minute)

    claims := jwt.StandardClaims{
        ExpiresAt: expireTime.Unix(),
        IssuedAt:  nowTime.Unix(),
        Id:        fmt.Sprintf("%d", user.ID),
        Subject:   user.Username,
    }

    tokenClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    token, err := tokenClaims.SignedString([]byte(s.jwtSecret))
    return token, err
}

// 生成刷新令牌 (Refresh Token)
func (s *AuthService) generateRefreshToken(user *entity.User) (string, error) {
    nowTime := time.Now()
    expireTime := nowTime.Add(time.Duration(s.refreshTokenExpireDays) * 24 * time.Hour)

    claims := jwt.StandardClaims{
        ExpiresAt: expireTime.Unix(),
        IssuedAt:  nowTime.Unix(),
        Id:        fmt.Sprintf("refresh_%d_%s", user.ID, uuid.New().String()),
        Subject:   user.Username,
    }

    tokenClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    token, err := tokenClaims.SignedString([]byte(s.jwtRefreshSecret))
    return token, err
}
```

### 令牌撤销

为了支持令牌撤销功能，系统采用以下策略：

1. **令牌黑名单机制**

   - 使用 Redis 存储已撤销的令牌标识
   - 撤销的令牌会被登记到黑名单中直到原本过期时间
   - 每次验证令牌时检查是否在黑名单中

2. **Redis 存储结构**

   - 使用 Redis 的 Hash 结构存储撤销的令牌
   - 使用令牌 ID 作为键，过期时间作为值
   - 自动设置过期时间减少存储开销

**实现细节**：

```go
// 撤销令牌
func (s *AuthService) RevokeToken(token string) error {
    claims, err := s.parseToken(token)
    if err != nil {
        return err
    }

    // 获取令牌 ID
    jti := claims.Id
    exp := claims.ExpiresAt

    // 将令牌加入黑名单
    key := fmt.Sprintf("revoked_token:%s", jti)
    return s.redisClient.Set(key, "1", time.Unix(exp, 0).Sub(time.Now())).Err()
}

// 验证令牌是否被撤销
func (s *AuthService) isTokenRevoked(jti string) bool {
    key := fmt.Sprintf("revoked_token:%s", jti)
    exists, err := s.redisClient.Exists(key).Result()
    if err != nil {
        return false
    }
    return exists > 0
}
```

### 令牌内容优化

为降低安全风险，优化令牌内容：

1. **精简令牌内容**

   - Access Token 中仅存储用户 ID (userID)
   - 不存储敏感信息如权限列表、个人资料等
   - 敏感信息通过专用 API 获取

2. **分离关注点**

   - 认证信息与用户资料分离
   - 令牌仅用于身份验证
   - 业务数据从数据库或缓存获取

**实现示例**：

```go
// 验证令牌并获取用户ID
func (s *AuthService) ValidateToken(token string) (int64, error) {
    claims, err := s.parseToken(token)
    if (err != nil) {
        return 0, err
    }

    // 检查令牌是否被撤销
    if s.isTokenRevoked(claims.Id) {
        return 0, errors.New("token has been revoked")
    }

    // 从令牌中提取用户ID
    userID, err := strconv.ParseInt(claims.Id, 10, 64)
    if (err != nil) {
        return 0, err
    }

    return userID, nil
}

// 根据用户ID获取用户权限（单独API）
func (s *AuthService) GetUserPermissions(userID int64) ([]string, error) {
    // 从数据库或缓存获取用户权限
    return s.userRepository.GetUserPermissions(userID)
}
```

### 令牌传输安全

提高令牌传输和存储的安全性：

1. **HttpOnly Cookie**

   - Refresh Token 通过 HttpOnly Cookie 传输
   - 防止 JavaScript 访问令牌，抵御 XSS 攻击
   - 设置 Secure 标志确保仅通过 HTTPS 传输

2. **Access Token 传输**

   - Access Token 通过 Authorization 头传输
   - 避免在 URL 参数中传输令牌
   - 前端避免在 localStorage 中存储令牌

**服务端实现**：

```go
func (h *AuthHandler) Login(c *gin.Context) {
    // 验证用户凭据...

    // 生成令牌
    accessToken, refreshToken, err := h.authService.GenerateTokens(user)
    if (err != nil) {
        response.Fail(c, response.CodeInternalError, err.Error())
        return
    }

    // 设置 Refresh Token 为 HttpOnly Cookie
    c.SetCookie(
        "refresh_token",
        refreshToken,
        int(time.Duration(h.config.Auth.RefreshTokenExpireDays)*24*time.Hour/time.Second),
        "/v1/auth",  // 仅用于认证路径
        h.config.Server.Domain,
        h.config.Server.SecureCookie,  // 在生产环境设为 true
        true,  // HttpOnly
    )

    // 仅返回 Access Token 给客户端
    response.Success(c, gin.H{
        "token": accessToken,
        "user":  userInfo,
    })
}
```

### 令牌刷新流程

优化令牌刷新逻辑，提高安全性：

1. **标准刷新流程**

   - 清理前端中的令牌刷新队列机制
   - 当 Access Token 过期时，使用 Refresh Token 获取新令牌
   - 每次刷新同时更新 Access Token 和 Refresh Token
   - 实现令牌轮转机制，增加安全性

2. **严格的验证机制**

   - 刷新令牌时验证用户的设备信息
   - 监控异常的刷新请求模式
   - 支持多设备登录但各自使用独立的令牌对

**前端实现**：

```typescript
// API 拦截器处理令牌刷新
const apiClient = axios.create({
  baseURL: "/v1",
  timeout: 10000,
});

// 响应拦截器处理令牌过期情况
apiClient.interceptors.response.use(
  (response) => response,
  async (error) => {
    const originalRequest = error.config;

    // 如果是因为令牌过期导致的 401 错误且未尝试过刷新令牌
    if (error.response.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true;

      try {
        // 调用刷新令牌接口，此处无需手动发送刷新令牌
        // 服务端会从 HttpOnly Cookie 中读取
        const { data } = await axios.post("/v1/auth/refresh");

        // 更新访问令牌
        setToken(data.token);

        // 重新发送之前失败的请求
        originalRequest.headers["Authorization"] = "Bearer " + data.token;
        return apiClient(originalRequest);
      } catch (refreshError) {
        // 刷新令牌也失效，需要重新登录
        removeToken();
        router.push("/login");
        return Promise.reject(refreshError);
      }
    }

    return Promise.reject(error);
  }
);
```

**服务端实现**：

```go
func (h *AuthHandler) RefreshToken(c *gin.Context) {
    // 从 Cookie 中获取刷新令牌
    refreshToken, err := c.Cookie("refresh_token")
    if (err != nil) {
        response.Fail(c, response.CodeUnauthorized, "refresh token not found")
        return
    }

    // 验证刷新令牌
    userID, err := h.authService.ValidateRefreshToken(refreshToken)
    if (err != nil) {
        response.Fail(c, response.CodeUnauthorized, "invalid refresh token")
        return
    }

    // 获取用户信息
    user, err := h.userService.GetUserByID(userID)
    if (err != nil) {
        response.Fail(c, response.CodeInternalError, err.Error())
        return
    }

    // 生成新的令牌对
    accessToken, newRefreshToken, err := h.authService.GenerateTokens(user)
    if (err != nil) {
        response.Fail(c, response.CodeInternalError, err.Error())
        return
    }

    // 撤销旧的刷新令牌
    h.authService.RevokeToken(refreshToken)

    // 设置新的刷新令牌 Cookie
    c.SetCookie(
        "refresh_token",
        newRefreshToken,
        int(time.Duration(h.config.Auth.RefreshTokenExpireDays)*24*time.Hour/time.Second),
        "/v1/auth",
        h.config.Server.Domain,
        h.config.Server.SecureCookie,
        true,
    )

    // 返回新的访问令牌
    response.Success(c, gin.H{
        "token": accessToken,
    })
}
```

### 令牌最佳实践

系统实施以下令牌安全最佳实践：

1. **令牌生命周期管理**

   - Access Token: 30 分钟有效期
   - Refresh Token: 7 天有效期
   - 支持手动撤销令牌
   - 支持全局令牌刷新（如密码更改时）

2. **传输安全**

   - 所有令牌操作都通过 HTTPS 进行
   - 令牌不暴露给第三方 JavaScript
   - 不在客户端日志中记录令牌信息

3. **防护措施**

   - 实现速率限制，防止暴力攻击
   - 监控异常的令牌使用模式
   - 记录关键令牌操作的审计日志

4. **应急响应**

   - 支持用户会话强制终止
   - 支持所有活跃令牌的批量撤销
   - 提供可疑活动的实时警报

## 角色与权限管理系统(RBAC)

### RBAC 模型概述

DistributedJob 实现了基于部门-角色-权限的三层访问控制模型，提供精细的权限管理：

1. **部门(Department)**：组织的基本单位，用于资源隔离
2. **角色(Role)**：职责的抽象，如管理员、普通用户等
3. **权限(Permission)**：具体操作的权限项，如创建任务、查看报表等

### 权限设计

系统权限采用"资源:操作"的命名模式，例如：

- `task:create` - 创建任务的权限
- `user:read` - 查看用户信息的权限
- `system:admin` - 系统管理权限

权限级别划分为：

1. **系统级权限**：影响整个系统的操作权限
2. **部门级权限**：限定在特定部门内的操作权限
3. **资源级权限**：针对特定资源的操作权限

### 角色权限关系

角色与权限是多对多关系，通过 role_permission 表建立关联。系统预定义了几个基本角色：

1. **超级管理员**：拥有所有权限
2. **部门管理员**：管理部门内的用户和任务
3. **普通用户**：使用系统基本功能
4. **只读用户**：只能查看而无法修改数据

下面是角色权限关系的表示：

```go
// 角色实体
type Role struct {
    ID          int64     `json:"id" gorm:"primaryKey;autoIncrement"`
    Name        string    `json:"name" gorm:"type:varchar(50);not null;uniqueIndex:idx_name"`
    Description string    `json:"description" gorm:"type:varchar(255)"`
    Status      int8      `json:"status" gorm:"type:tinyint(4);not null;default:1"`
    CreateTime  time.Time `json:"createTime" gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP"`
    UpdateTime  time.Time `json:"updateTime" gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP;autoUpdateTime"`
    Permissions []Permission `json:"permissions" gorm:"many2many:role_permission"`
}

// 权限实体
type Permission struct {
    ID          int64     `json:"id" gorm:"primaryKey;autoIncrement"`
    Name        string    `json:"name" gorm:"type:varchar(50);not null"`
    Code        string    `json:"code" gorm:"type:varchar(50);not null;uniqueIndex:idx_code"`
    Description string    `json:"description" gorm:"type:varchar(255)"`
    Status      int8      `json:"status" gorm:"type:tinyint(4);not null;default:1"`
    CreateTime  time.Time `json:"createTime" gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP"`
    UpdateTime  time.Time `json:"updateTime" gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP;autoUpdateTime"`
}

// 角色权限关联表
type RolePermission struct {
    ID           int64     `json:"id" gorm:"primaryKey;autoIncrement"`
    RoleID       int64     `json:"roleId" gorm:"column:role_id;not null;uniqueIndex:idx_role_perm"`
    PermissionID int64     `json:"permissionId" gorm:"column:permission_id;not null;uniqueIndex:idx_role_perm"`
    CreateTime   time.Time `json:"createTime" gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP"`
    UpdateTime   time.Time `json:"updateTime" gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP;autoUpdateTime"`
}
```

### 权限检查流程

1. **API 层检查**：通过中间件对 API 请求进行权限检查
2. **服务层检查**：在关键业务逻辑中二次验证权限
3. **资源隔离**：确保用户只能访问所属部门的资源
4. **权限缓存**：缓存用户权限，减少数据库查询

具体实现如下：

```go
// JWT认证中间件
func JWTAuthMiddleware(authService service.AuthService) gin.HandlerFunc {
    return func(c *gin.Context) {
        token := c.GetHeader("Authorization")
        if token == "" {
            response.Fail(c, response.CodeUnauthorized, "请先登录")
            c.Abort()
            return
        }

        // Bearer Token格式处理
        if len(token) > 7 && token[0:7] == "Bearer " {
            token = token[7:]
        }

        // 验证令牌
        userID, err := authService.ValidateToken(token)
        if err != nil {
            response.Fail(c, response.CodeUnauthorized, "无效的令牌")
            c.Abort()
            return
        }

        // 设置用户ID到上下文
        c.Set("userID", userID)
        c.Next()
    }
}

// 权限检查中间件
func PermissionMiddleware(authService service.AuthService, requiredPermissions ...string) gin.HandlerFunc {
    return func(c *gin.Context) {
        userID, exists := c.Get("userID")
        if !exists {
            response.Fail(c, response.CodeUnauthorized, "请先登录")
            c.Abort()
            return
        }

        // 获取用户权限
        permissions, err := authService.GetUserPermissions(userID.(int64))
        if err != nil {
            response.Fail(c, response.CodeInternalError, "获取权限失败")
            c.Abort()
            return
        }

        // 检查是否有超级管理员权限
        if contains(permissions, "system:admin") {
            c.Next()
            return
        }

        // 检查是否有所需权限
        hasPermission := false
        for _, required := range requiredPermissions {
            if contains(permissions, required) {
                hasPermission = true
                break
            }
        }

        if !hasPermission {
            response.Fail(c, response.CodeForbidden, "权限不足")
            c.Abort()
            return
        }

        c.Next()
    }
}
```

### 部门资源隔离

系统严格实施基于部门的资源隔离，确保多租户安全：

1. 任务资源按部门隔离，只有同部门用户可查看和管理
2. 用户账号与部门绑定，限制跨部门访问
3. 报表和统计数据按部门筛选
4. 超级管理员可以跨部门管理资源

实现示例：

```go
// 获取任务列表的服务层实现
func (s *taskService) GetTaskList(userID int64, page, size int) ([]*entity.Task, int64, error) {
    // 获取用户信息
    user, err := s.userRepo.GetUserByID(userID)
    if err != nil {
        return nil, 0, err
    }

    // 获取用户权限
    permissions, err := s.authService.GetUserPermissions(userID)
    if err != nil {
        return nil, 0, err
    }

    // 如果具有超级管理员权限，可以查看所有部门的任务
    if contains(permissions, "system:admin") {
        return s.taskRepo.GetAllTasks(page, size)
    }

    // 普通用户只能查看本部门任务
    return s.taskRepo.GetTasksByDepartmentID(user.DepartmentID, page, size)
}
```

### 权限模型最佳实践

系统采用以下权限管理最佳实践：

1. **最小权限原则**：默认分配最小权限集合，遵循最小权限原则
2. **职责分离**：将敏感操作权限分配给不同角色，实现职责分离
3. **权限审计**：记录关键权限变更操作，便于安全审计
4. **权限模板**：预设角色权限模板，便于批量权限管理
5. **动态权限调整**：支持根据业务需要动态调整权限策略
6. **权限可视化**：前端提供权限树形结构可视化管理界面
