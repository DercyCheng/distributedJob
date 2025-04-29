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
5. [前端开发](#前端开发)
   - [技术栈](#技术栈)
   - [项目结构](#前端项目结构)
   - [开发指南](#开发指南)
   - [构建与部署](#构建与部署)

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
│   └── main.go       # 客户端入口点
├── internal/             # 私有应用程序和库代码
│   ├── api/              # API 相关代码
│   │   ├── handler/      # HTTP 处理器
│   │   ├── middleware/   # HTTP 中间件
│   │   ├── router.go     # HTTP 路由定义
│   │   └── server.go     # API 服务器
│   ├── rpc/              # RPC 服务相关代码
│   │   ├── proto/        # Protocol Buffers 定义
│   │   ├── server/       # RPC 服务器实现
│   │   └── client/       # RPC 客户端实现
│   ├── auth/             # 认证和授权
│   │   ├── department.go # 部门管理
│   │   ├── permission.go # 权限控制
│   │   └── user.go       # 用户管理
│   ├── config/           # 配置管理
│   │   └── config.go     # 配置结构和加载逻辑
│   ├── job/              # 核心任务调度模块
│   │   ├── scheduler.go  # 任务调度器
│   │   ├── task.go       # 任务定义和管理
│   │   ├── http_worker.go # HTTP 任务执行器
│   │   ├── grpc_worker.go # gRPC 任务执行器
│   │   └── history.go    # 任务执行历史记录
│   ├── model/            # 数据模型
│   │   ├── dto/          # 数据传输对象
│   │   └── entity/       # 业务实体对象
│   ├── store/            # 存储层
│   │   ├── mysql/        # MySQL 实现
│   │   │   └── repository/  # 数据访问对象
│   │   └── repository.go # 存储接口定义
│   └── service/          # 业务逻辑服务
│       ├── task_service.go  # 任务服务实现
│       ├── department_service.go  # 部门服务实现
│       └── history_service.go  # 历史记录服务实现
├── pkg/                  # 可被外部应用程序使用的库
│   ├── logger/           # 日志工具
│   └── utils/            # 通用工具函数
├── web/                  # 前端应用 (Vite 构建)
│   ├── src/              # 源代码
│   ├── public/           # 静态资源
│   ├── index.html        # 入口 HTML
│   ├── vite.config.js    # Vite 配置
│   └── package.json      # 依赖配置
├── configs/              # 配置文件目录
│   └── config.yaml       # 主配置文件
├── scripts/              # 构建和部署脚本
├── docs/                 # 文档目录
├── go.mod                # Go 模块依赖
└── go.sum                # Go 模块校验和
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
   - 根据 Cron 表达式组织任务
   - 当任务到期时，分发给相应的执行器

3. **任务执行**

   - 执行器（HTTP 或 gRPC）执行任务
   - 结果记录在执行历史中
   - 根据配置，失败任务可能会重试
   - 如果主要执行失败，则触发备用机制

4. **用户交互**

   - 用户通过基于 Vite 构建的 Web 控制台或 API 与系统交互
   - 强制执行身份验证和授权
   - 用户可以管理任务、查看历史记录并配置部门/权限

### 设计原则

- **模块化**：组件设计具有明确的边界和接口
- **可扩展性**：无状态设计允许水平扩展
- **弹性**：重试机制和备用方案确保可靠性
- **安全性**：基于角色的访问控制模型，实现精细权限控制
- **可观测性**：全面的日志记录和执行历史
- **高性能通信**：使用 gRPC 实现高效的内部服务通信

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

DistributedJob 使用 Protocol Buffers 来定义 RPC 服务接口。以下是主要服务定义示例：

```protobuf
syntax = "proto3";
package scheduler;

option go_package = "github.com/username/distributedJob/internal/rpc/proto;schedulerpb";

service TaskScheduler {
  rpc ScheduleTask(ScheduleTaskRequest) returns (ScheduleTaskResponse);
  rpc PauseTask(TaskRequest) returns (TaskResponse);
  rpc ResumeTask(TaskRequest) returns (TaskResponse);
  rpc GetTaskStatus(TaskRequest) returns (TaskStatusResponse);
}

message ScheduleTaskRequest {
  string name = 1;
  string cron_expression = 2;
  string handler = 3;
  bytes params = 4;
  int32 max_retry = 5;
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
}
```

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
  if err != nil {
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

  if err != nil {
    log.Fatalf("Could not schedule task: %v", err)
  }

  log.Printf("Task scheduled with ID: %d, Success: %v", resp.TaskId, resp.Success)
}
```

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
  if err != nil {
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
| email           |              |          | update_time        |
| phone           |              |          +--------------------+
| department_id(FK)--------------|                   ^
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

## 前端开发

### 技术栈

DistributedJob 前端应用使用现代化技术栈，基于 Vite 构建：

- **构建工具**: Vite
- **框架**: Vue 3 / React
- **UI 组件库**: Element Plus / Ant Design
- **状态管理**: Pinia / Redux
- **HTTP 客户端**: Axios
- **CSS 预处理器**: SCSS / Less
- **打包工具**: Rollup (由 Vite 内置)
- **代码规范**: ESLint + Prettier

### 前端项目结构

```
web/
├── public/                  # 静态资源
│   ├── favicon.ico          # 网站图标
│   └── assets/              # 其他静态资源
├── src/                     # 源代码
│   ├── api/                 # API 请求
│   │   ├── index.js         # API 导出
│   │   ├── request.js       # Axios 配置
│   │   ├── task.js          # 任务相关 API
│   │   ├── user.js          # 用户相关 API
│   │   └── department.js    # 部门相关 API
│   ├── assets/              # 资源文件
│   │   ├── images/          # 图片资源
│   │   └── styles/          # 样式文件
│   ├── components/          # 通用组件
│   │   ├── common/          # 公共组件
│   │   ├── layout/          # 布局组件
│   │   └── widgets/         # 功能组件
│   ├── hooks/               # 自定义 Hooks
│   ├── pages/               # 页面组件
│   │   ├── dashboard/       # 控制台页面
│   │   ├── task/            # 任务管理页面
│   │   ├── user/            # 用户管理页面
│   │   └── department/      # 部门管理页面
│   ├── router/              # 路由配置
│   │   └── index.js         # 路由定义
│   ├── store/               # 状态管理
│   │   ├── modules/         # 状态模块
│   │   └── index.js         # 状态入口
│   ├── utils/               # 工具函数
│   │   ├── auth.js          # 认证相关
│   │   └── formatter.js     # 格式化工具
│   ├── App.vue              # 应用入口组件
│   └── main.js              # 应用入口 JS
├── index.html               # HTML 入口文件
├── vite.config.js           # Vite 配置文件
├── package.json             # 依赖配置
├── .eslintrc.js             # ESLint 配置
└── .prettierrc.js           # Prettier 配置
```

### 开发指南

#### 环境准备

1. **安装 Node.js**

   确保安装了 Node.js 16.0 或更高版本。

2. **安装依赖**

   ```bash
   cd web
   npm install
   ```

#### 开发

1. **启动开发服务器**

   ```bash
   npm run dev
   ```

   这将启动 Vite 开发服务器，通常在 http://localhost:3000 上运行。

2. **API 配置**

   在 `src/api/request.js` 中配置 API 基础 URL：

   ```javascript
   import axios from "axios";
   import { getToken } from "../utils/auth";

   const request = axios.create({
     baseURL: "/v1",
     timeout: 10000,
   });

   request.interceptors.request.use(
     (config) => {
       const token = getToken();
       if (token) {
         config.headers["Authorization"] = `Bearer ${token}`;
       }
       return config;
     },
     (error) => {
       return Promise.reject(error);
     }
   );

   // 响应拦截器...

   export default request;
   ```

3. **开发新页面**

   - 在 `src/pages` 目录中创建新的页面组件
   - 在 `src/router/index.js` 中添加路由配置
   - 在 `src/api` 中添加相关 API 请求方法

### 构建与部署

1. **构建生产版本**

   ```bash
   npm run build
   ```

   构建结果将输出到 `dist` 目录。

2. **预览构建结果**

   ```bash
   npm run preview
   ```

3. **部署配置**

   Vite 项目支持基本 URL 配置，方便部署到子路径：

   ```javascript
   // vite.config.js
   import { defineConfig } from "vite";
   import vue from "@vitejs/plugin-vue";

   export default defineConfig({
     plugins: [vue()],
     base: "/v1/web/", // 部署到子路径
     server: {
       proxy: {
         "/v1/api": {
           target: "http://localhost:9088",
           changeOrigin: true,
         },
       },
     },
     build: {
       outDir: "dist",
       assetsDir: "assets",
       sourcemap: false,
     },
   });
   ```

4. **与后端集成**

   构建完成后，可以将生成的静态文件复制到 Go 应用程序中，并通过 Web 服务器提供。

   ```go
   // 在 Go 应用中提供静态文件
   router.Static("/v1/web", "./web/dist")
   ```

5. **Docker 部署**

   可以使用多阶段构建来创建包含前端和后端的单一 Docker 镜像：

   ```dockerfile
   # 构建前端
   FROM node:16 AS frontend-builder
   WORKDIR /app/web
   COPY web/package*.json ./
   RUN npm install
   COPY web .
   RUN npm run build

   # 构建后端
   FROM golang:1.16 AS backend-builder
   WORKDIR /app
   COPY go.* ./
   RUN go mod download
   COPY . .
   COPY --from=frontend-builder /app/web/dist ./web/dist
   RUN CGO_ENABLED=0 GOOS=linux go build -o distributedjob ./cmd/server/main.go

   # 最终镜像
   FROM alpine:3.14
   RUN apk --no-cache add ca-certificates tzdata
   WORKDIR /app
   COPY --from=backend-builder /app/distributedjob .
   COPY --from=backend-builder /app/web/dist ./web/dist
   COPY configs ./configs
   EXPOSE 9088 9090
   ENTRYPOINT ["/app/distributedjob"]
   ```
