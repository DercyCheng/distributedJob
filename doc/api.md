# API 文档

本文档详细说明 DistributedJob 对外提供的 API 接口。

## API 概述

DistributedJob 提供了一套 RESTful API，用于管理定时任务、部门权限和查询执行记录。所有 API 都使用 JSON 格式进行数据交换，并返回统一的响应格式。

### 基础 URL

所有 API 的基础路径为：`http://<host>:<port>/v1`

### 统一响应格式

```json
{
  "code": 0,       // 0 表示成功，非 0 表示错误
  "message": "",   // 响应消息，成功时为 "success"，失败时为错误信息
  "data": null     // 响应数据，可能是对象、数组或 null
}
```

### 错误码说明

| 错误码 | 说明           |
| ------ | -------------- |
| 0      | 成功           |
| 4001   | 参数错误       |
| 4003   | 权限不足       |
| 4004   | 资源不存在     |
| 5000   | 服务器内部错误 |

### 认证鉴权

除了健康检查接口外，所有 API 都需要通过认证鉴权。认证方式为基于 Token 的认证，Token 通过登录接口获取。

请求时需要在 HTTP Header 中添加 `Authorization` 字段：

```
Authorization: Bearer <token>
```

## 用户认证 API

### 用户登录

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

### 刷新 Token

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

### 获取当前用户信息

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
    "permissions": ["task:create", "task:update", "task:delete", "task:view", "department:manage"]
  }
}
```

## 部门管理 API

### 获取部门列表

```
GET /departments?keyword={keyword}
Authorization: Bearer <token>
```

参数说明：

- `keyword` : 部门名称关键字（可选）

响应示例：

```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": 1,
      "name": "总部",
      "description": "总部",
      "parentId": null,
      "children": [
        {
          "id": 2,
          "name": "技术部",
          "description": "负责系统研发和维护",
          "parentId": 1,
          "status": 1,
          "createTime": "2023-01-01T10:00:00Z",
          "updateTime": "2023-01-02T15:30:00Z"
        },
        {
          "id": 3,
          "name": "运营部",
          "description": "负责系统运营",
          "parentId": 1,
          "status": 1,
          "createTime": "2023-01-01T10:00:00Z",
          "updateTime": "2023-01-02T15:30:00Z"
        }
      ],
      "status": 1,
      "createTime": "2023-01-01T10:00:00Z",
      "updateTime": "2023-01-02T15:30:00Z"
    }
  ]
}
```

### 获取部门详情

```
GET /departments/{id}
Authorization: Bearer <token>
```

参数说明：

- `id` : 部门 ID

响应示例：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 2,
    "name": "技术部",
    "description": "负责系统研发和维护",
    "parentId": 1,
    "parentName": "总部",
    "status": 1,
    "createTime": "2023-01-01T10:00:00Z",
    "updateTime": "2023-01-02T15:30:00Z"
  }
}
```

### 创建部门

```
POST /departments
Content-Type: application/json
Authorization: Bearer <token>
```

请求参数：

| 参数名      | 类型      | 是否必填 | 说明                   |
| ----------- | --------- | -------- | ---------------------- |
| name        | string    | 是       | 部门名称               |
| description | string    | 否       | 部门描述               |
| parentId    | number    | 否       | 父部门ID，顶级部门为空 |
| status      | number    | 是       | 状态：0-禁用，1-启用   |

请求示例：

```json
{
  "name": "测试部",
  "description": "负责系统测试",
  "parentId": 1,
  "status": 1
}
```

响应示例：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 4
  }
}
```

### 更新部门

```
PUT /departments/{id}
Content-Type: application/json
Authorization: Bearer <token>
```

参数说明：

- `id` : 部门 ID

请求参数：同创建部门

请求示例：

```json
{
  "name": "测试部",
  "description": "负责系统测试和质量保证",
  "parentId": 1,
  "status": 1
}
```

响应示例：

```json
{
  "code": 0,
  "message": "success",
  "data": null
}
```

### 删除部门

```
DELETE /departments/{id}
Authorization: Bearer <token>
```

参数说明：

- `id` : 部门 ID

响应示例：

```json
{
  "code": 0,
  "message": "success",
  "data": null
}
```

## 用户管理 API

### 获取用户列表

```
GET /users?page={page}&size={size}&departmentId={departmentId}&keyword={keyword}
Authorization: Bearer <token>
```

参数说明：

- `page` : 页码，从 1 开始（可选，默认为 1）
- `size` : 每页记录数（可选，默认为 10）
- `departmentId` : 部门ID（可选）
- `keyword` : 用户名或真实姓名关键字（可选）

响应示例：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "users": [
      {
        "id": 1,
        "username": "admin",
        "realName": "系统管理员",
        "email": "admin@example.com",
        "phone": "13800138000",
        "departmentId": 1,
        "departmentName": "技术部",
        "roleId": 1,
        "roleName": "管理员",
        "status": 1,
        "createTime": "2023-01-01T10:00:00Z",
        "updateTime": "2023-01-02T15:30:00Z"
      }
    ],
    "total": 100,
    "page": 1,
    "size": 10
  }
}
```

### 获取用户详情

```
GET /users/{id}
Authorization: Bearer <token>
```

参数说明：

- `id` : 用户 ID

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
    "status": 1,
    "createTime": "2023-01-01T10:00:00Z",
    "updateTime": "2023-01-02T15:30:00Z"
  }
}
```

### 创建用户

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
| departmentId | number | 是       | 所属部门ID           |
| roleId       | number | 是       | 角色ID               |
| status       | number | 是       | 状态：0-禁用，1-启用 |

请求示例：

```json
{
  "username": "test_user",
  "password": "test123",
  "realName": "测试用户",
  "email": "test@example.com",
  "phone": "13900139000",
  "departmentId": 2,
  "roleId": 2,
  "status": 1
}
```

响应示例：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 2
  }
}
```

### 更新用户

```
PUT /users/{id}
Content-Type: application/json
Authorization: Bearer <token>
```

参数说明：

- `id` : 用户 ID

请求参数：同创建用户，但 `password` 为可选

请求示例：

```json
{
  "username": "test_user",
  "realName": "测试用户",
  "email": "test@example.com",
  "phone": "13900139000",
  "departmentId": 2,
  "roleId": 2,
  "status": 1
}
```

响应示例：

```json
{
  "code": 0,
  "message": "success",
  "data": null
}
```

### 删除用户

```
DELETE /users/{id}
Authorization: Bearer <token>
```

参数说明：

- `id` : 用户 ID

响应示例：

```json
{
  "code": 0,
  "message": "success",
  "data": null
}
```

### 修改用户密码

```
PATCH /users/{id}/password
Content-Type: application/json
Authorization: Bearer <token>
```

参数说明：

- `id` : 用户 ID

请求参数：

| 参数名      | 类型   | 是否必填 | 说明     |
| ----------- | ------ | -------- | -------- |
| newPassword | string | 是       | 新密码   |
| oldPassword | string | 是       | 原密码   |

请求示例：

```json
{
  "oldPassword": "test123",
  "newPassword": "test456"
}
```

响应示例：

```json
{
  "code": 0,
  "message": "success",
  "data": null
}
```

## 角色与权限管理 API

### 获取角色列表

```
GET /roles?page={page}&size={size}&keyword={keyword}
Authorization: Bearer <token>
```

参数说明：

- `page` : 页码，从 1 开始（可选，默认为 1）
- `size` : 每页记录数（可选，默认为 10）
- `keyword` : 角色名称关键字（可选）

响应示例：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "roles": [
      {
        "id": 1,
        "name": "管理员",
        "description": "系统管理员，拥有所有权限",
        "status": 1,
        "createTime": "2023-01-01T10:00:00Z",
        "updateTime": "2023-01-02T15:30:00Z"
      }
    ],
    "total": 10,
    "page": 1,
    "size": 10
  }
}
```

### 获取角色详情

```
GET /roles/{id}
Authorization: Bearer <token>
```

参数说明：

- `id` : 角色 ID

响应示例：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "name": "管理员",
    "description": "系统管理员，拥有所有权限",
    "permissions": [
      {
        "id": 1,
        "name": "任务创建",
        "code": "task:create"
      },
      {
        "id": 2,
        "name": "任务更新",
        "code": "task:update"
      }
    ],
    "status": 1,
    "createTime": "2023-01-01T10:00:00Z",
    "updateTime": "2023-01-02T15:30:00Z"
  }
}
```

### 创建角色

```
POST /roles
Content-Type: application/json
Authorization: Bearer <token>
```

请求参数：

| 参数名       | 类型     | 是否必填 | 说明                 |
| ------------ | -------- | -------- | -------------------- |
| name         | string   | 是       | 角色名称             |
| description  | string   | 否       | 角色描述             |
| permissions  | number[] | 是       | 权限ID数组           |
| status       | number   | 是       | 状态：0-禁用，1-启用 |

请求示例：

```json
{
  "name": "运维人员",
  "description": "负责系统运维",
  "permissions": [1, 2, 3, 4],
  "status": 1
}
```

响应示例：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 2
  }
}
```

### 更新角色

```
PUT /roles/{id}
Content-Type: application/json
Authorization: Bearer <token>
```

参数说明：

- `id` : 角色 ID

请求参数：同创建角色

请求示例：

```json
{
  "name": "运维人员",
  "description": "负责系统运维和监控",
  "permissions": [1, 2, 3, 4, 5],
  "status": 1
}
```

响应示例：

```json
{
  "code": 0,
  "message": "success",
  "data": null
}
```

### 删除角色

```
DELETE /roles/{id}
Authorization: Bearer <token>
```

参数说明：

- `id` : 角色 ID

响应示例：

```json
{
  "code": 0,
  "message": "success",
  "data": null
}
```

### 获取所有权限列表

```
GET /permissions
Authorization: Bearer <token>
```

响应示例：

```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": 1,
      "name": "任务创建",
      "code": "task:create",
      "description": "创建定时任务的权限"
    },
    {
      "id": 2,
      "name": "任务更新",
      "code": "task:update",
      "description": "更新定时任务的权限"
    }
  ]
}
```

## 健康检查 API

### 获取服务健康状态

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

正在等待关闭服务：HTTP 状态码返回 `400`

```json
{
  "code": 4000,
  "message": "service is shutting down",
  "data": {
    "status": "shutting_down",
    "timestamp": "2023-01-01T12:00:00Z"
  }
}
```

## 服务关闭 API

### 平滑关闭服务实例

```
GET /shutdown?wait={wait}
```

参数说明：

- `wait` : 等待关闭时间（单位-秒），必须大于 0

**注意**：该接口仅限本机调用（只能使用 `localhost`、`127.0.0.1`、`0.0.0.0` 这三个 hostname 访问）

响应示例：

```json
{
  "code": 0,
  "message": "service will shutdown after 10 seconds",
  "data": null
}
```

## 任务管理 API

### 获取任务列表

```
GET /tasks?page={page}&size={size}&keyword={keyword}&departmentId={departmentId}&taskType={taskType}
Authorization: Bearer <token>
```

参数说明：

- `page` : 页码，从 1 开始（可选，默认为 1）
- `size` : 每页记录数（可选，默认为 10）
- `keyword` : 任务名称关键字（可选）
- `departmentId` : 部门ID（可选）
- `taskType` : 任务类型，HTTP 或 GRPC（可选）

响应示例：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "tasks": [
      {
        "id": 1,
        "name": "每日数据统计",
        "departmentId": 2,
        "departmentName": "技术部",
        "taskType": "HTTP",
        "cron": "0 1 * * *",
        "url": "http://example.com/api/stats",
        "httpMethod": "GET",
        "body": "",
        "headers": "{\"Authorization\":\"Bearer token123\"}",
        "retryCount": 3,
        "retryInterval": 60,
        "fallbackUrl": "http://backup.example.com/api/stats",
        "status": 1,
        "createTime": "2023-01-01T10:00:00Z",
        "updateTime": "2023-01-02T15:30:00Z",
        "createBy": "admin",
        "updateBy": "admin"
      },
      {
        "id": 2,
        "name": "用户数据同步",
        "departmentId": 3,
        "departmentName": "运营部",
        "taskType": "GRPC",
        "cron": "0 0 * * *",
        "grpcService": "user.UserService",
        "grpcMethod": "SyncUserData",
        "grpcParams": "{\"source\":\"main_db\"}",
        "retryCount": 3,
        "retryInterval": 60,
        "fallbackGrpcService": "user.BackupUserService",
        "fallbackGrpcMethod": "SyncUserData",
        "status": 1,
        "createTime": "2023-01-01T10:00:00Z",
        "updateTime": "2023-01-02T15:30:00Z",
        "createBy": "admin",
        "updateBy": "admin"
      }
    ],
    "total": 100,
    "page": 1,
    "size": 10
  }
}
```

### 获取任务详情

```
GET /tasks/{id}
Authorization: Bearer <token>
```

参数说明：

- `id` : 任务 ID

响应示例：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "name": "每日数据统计",
    "departmentId": 2,
    "departmentName": "技术部",
    "taskType": "HTTP",
    "cron": "0 1 * * *",
    "url": "http://example.com/api/stats",
    "httpMethod": "GET",
    "body": "",
    "headers": "{\"Authorization\":\"Bearer token123\"}",
    "retryCount": 3,
    "retryInterval": 60,
    "fallbackUrl": "http://backup.example.com/api/stats",
    "status": 1,
    "createTime": "2023-01-01T10:00:00Z",
    "updateTime": "2023-01-02T15:30:00Z",
    "createBy": "admin",
    "updateBy": "admin"
  }
}
```

### 创建 HTTP 任务

```
POST /tasks/http
Content-Type: application/json
Authorization: Bearer <token>
```

请求参数：

| 参数名        | 类型   | 是否必填 | 说明                                       |
| ------------- | ------ | -------- | ------------------------------------------ |
| name          | string | 是       | 任务名称                                   |
| departmentId  | number | 是       | 所属部门ID                                 |
| cron          | string | 是       | cron 表达式                                |
| url           | string | 是       | 调度 URL                                   |
| httpMethod    | string | 是       | HTTP 方法（GET、POST、PUT、PATCH、DELETE） |
| body          | string | 否       | 请求体                                     |
| headers       | string | 否       | 请求头（JSON 格式字符串）                  |
| retryCount    | number | 否       | 最大重试次数                               |
| retryInterval | number | 否       | 重试间隔（秒）                             |
| fallbackUrl   | string | 否       | 备用 URL                                   |
| status        | number | 是       | 状态：0-禁用，1-启用                       |

请求示例：

```json
{
  "name": "每日数据统计",
  "departmentId": 2,
  "cron": "0 1 * * *",
  "url": "http://example.com/api/stats",
  "httpMethod": "GET",
  "body": "",
  "headers": "{\"Authorization\":\"Bearer token123\"}",
  "retryCount": 3,
  "retryInterval": 60,
  "fallbackUrl": "http://backup.example.com/api/stats",
  "status": 1
}
```

响应示例：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1
  }
}
```

### 创建 gRPC 任务

```
POST /tasks/grpc
Content-Type: application/json
Authorization: Bearer <token>
```

请求参数：

| 参数名                | 类型   | 是否必填 | 说明                 |
| --------------------- | ------ | -------- | -------------------- |
| name                  | string | 是       | 任务名称             |
| departmentId          | number | 是       | 所属部门ID           |
| cron                  | string | 是       | cron 表达式          |
| grpcService           | string | 是       | gRPC 服务名          |
| grpcMethod            | string | 是       | gRPC 方法名          |
| grpcParams            | string | 否       | gRPC 参数(JSON字符串) |
| retryCount            | number | 否       | 最大重试次数         |
| retryInterval         | number | 否       | 重试间隔（秒）       |
| fallbackGrpcService   | string | 否       | 备用 gRPC 服务名     |
| fallbackGrpcMethod    | string | 否       | 备用 gRPC 方法名     |
| status                | number | 是       | 状态：0-禁用，1-启用 |

请求示例：

```json
{
  "name": "用户数据同步",
  "departmentId": 3,
  "cron": "0 0 * * *",
  "grpcService": "user.UserService",
  "grpcMethod": "SyncUserData",
  "grpcParams": "{\"source\":\"main_db\"}",
  "retryCount": 3,
  "retryInterval": 60,
  "fallbackGrpcService": "user.BackupUserService",
  "fallbackGrpcMethod": "SyncUserData",
  "status": 1
}
```

响应示例：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 2
  }
}
```

### 更新 HTTP 任务

```
PUT /tasks/http/{id}
Content-Type: application/json
Authorization: Bearer <token>
```

参数说明：

- `id` : HTTP 任务 ID

请求参数：同创建 HTTP 任务

请求示例：

```json
{
  "name": "每日数据统计",
  "departmentId": 2,
  "cron": "0 2 * * *",
  "url": "http://example.com/api/stats",
  "httpMethod": "GET",
  "body": "",
  "headers": "{\"Authorization\":\"Bearer token123\"}",
  "retryCount": 3,
  "retryInterval": 60,
  "fallbackUrl": "http://backup.example.com/api/stats",
  "status": 1
}
```

响应示例：

```json
{
  "code": 0,
  "message": "success",
  "data": null
}
```

### 更新 gRPC 任务

```
PUT /tasks/grpc/{id}
Content-Type: application/json
Authorization: Bearer <token>
```

参数说明：

- `id` : gRPC 任务 ID

请求参数：同创建 gRPC 任务

请求示例：

```json
{
  "name": "用户数据同步",
  "departmentId": 3,
  "cron": "0 1 * * *",
  "grpcService": "user.UserService",
  "grpcMethod": "SyncUserData",
  "grpcParams": "{\"source\":\"main_db\"}",
  "retryCount": 3,
  "retryInterval": 60,
  "fallbackGrpcService": "user.BackupUserService",
  "fallbackGrpcMethod": "SyncUserData",
  "status": 1
}
```

响应示例：

```json
{
  "code": 0,
  "message": "success",
  "data": null
}
```

### 删除任务

```
DELETE /tasks/{id}
Authorization: Bearer <token>
```

参数说明：

- `id` : 任务 ID

响应示例：

```json
{
  "code": 0,
  "message": "success",
  "data": null
}
```

### 修改任务状态

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

请求示例：

```json
{
  "status": 1
}
```

响应示例：

```json
{
  "code": 0,
  "message": "success",
  "data": null
}
```

## 执行记录查询 API

### 获取任务执行记录

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

响应示例：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "records": [
      {
        "id": 1,
        "taskId": 1,
        "taskName": "每日数据统计",
        "taskType": "HTTP",
        "departmentId": 2,
        "departmentName": "技术部",
        "url": "http://example.com/api/stats",
        "httpMethod": "GET",
        "body": "",
        "headers": "{\"Authorization\":\"Bearer token123\"}",
        "response": "{\"status\":\"success\"}",
        "statusCode": 200,
        "success": true,
        "retryTimes": 0,
        "useFallback": false,
        "costTime": 120,
        "createTime": "2023-01-01T01:00:00Z"
      },
      {
        "id": 2,
        "taskId": 2,
        "taskName": "用户数据同步",
        "taskType": "GRPC",
        "departmentId": 3,
        "departmentName": "运营部",
        "grpcService": "user.UserService",
        "grpcMethod": "SyncUserData",
        "grpcParams": "{\"source\":\"main_db\"}",
        "response": "{\"status\":\"success\",\"count\":1000}",
        "grpcStatus": 0,
        "success": true,
        "retryTimes": 0,
        "useFallback": false,
        "costTime": 250,
        "createTime": "2023-01-01T00:00:00Z"
      }
    ],
    "total": 30,
    "page": 1,
    "size": 10
  }
}
```

### 获取记录详情

```
GET /records/{id}?year={year}&month={month}
Authorization: Bearer <token>
```

参数说明：

- `id` : 记录 ID
- `year` : 年份，如 2025（必填）
- `month` : 月份，如 1-12（必填）

响应示例：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "taskId": 1,
    "taskName": "每日数据统计",
    "taskType": "HTTP",
    "departmentId": 2,
    "departmentName": "技术部",
    "url": "http://example.com/api/stats",
    "httpMethod": "GET",
    "body": "",
    "headers": "{\"Authorization\":\"Bearer token123\"}",
    "response": "{\"status\":\"success\"}",
    "statusCode": 200,
    "success": true,
    "retryTimes": 0,
    "useFallback": false,
    "costTime": 120,
    "createTime": "2023-01-01T01:00:00Z"
  }
}
```

### 获取任务执行历史统计

```
GET /records/stats?taskId={taskId}&departmentId={departmentId}&year={year}&month={month}
Authorization: Bearer <token>
```

参数说明：

- `taskId` : 任务 ID（可选）
- `departmentId` : 部门 ID（可选）
- `year` : 年份，如 2025（必填）
- `month` : 月份，如 1-12（必填）

响应示例：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "totalCount": 300,
    "successCount": 290,
    "failCount": 10,
    "successRate": 96.67,
    "avgCostTime": 135,
    "dailyStats": [
      {
        "date": "2025-01-01",
        "totalCount": 10,
        "successCount": 9,
        "failCount": 1
      },
      {
        "date": "2025-01-02",
        "totalCount": 10,
        "successCount": 10,
        "failCount": 0
      }
    ]
  }
}
```

## 错误示例

参数错误：

```json
{
  "code": 4001,
  "message": "invalid parameter: cron expression is invalid",
  "data": null
}
```

权限不足：

```json
{
  "code": 4003,
  "message": "insufficient permissions: you don't have permission to access this resource",
  "data": null
}
```

资源不存在：

```json
{
  "code": 4004,
  "message": "task not found: id=1",
  "data": null
}
```

服务器错误：

```json
{
  "code": 5000,
  "message": "internal server error",
  "data": null
}
```
