# 安装部署

本文档提供 DistributedJob 的安装和部署指南。

## 系统要求

- Go 1.16 或更高版本
- MySQL 5.7 或更高版本
- 任意操作系统（Windows, macOS, Linux）
- 如需使用 gRPC 任务功能，需安装相应的 gRPC 依赖

## 安装方式

### 方式一：从源代码编译

1. 克隆或下载源代码到本地
2. 进入项目目录
3. 编译源代码

```bash
go build main.go
```

4. 将编译后生成的可执行文件（main 或 main.exe）与配置文件 config.yaml、页面文件夹 web 放在同一目录下

目录结构应如下所示：
```
运行目录/
├── main (或 main.exe)    # 可执行文件
├── config.yaml           # 配置文件
└── web/                  # 网页控制台文件
    ├── index.html
    ├── css/
    └── js/
```

### 方式二：使用预编译的二进制文件

1. 从项目的 Release 页面下载对应操作系统的预编译二进制文件
2. 解压到合适的目录
3. 确保二进制文件、配置文件和 web 目录在同一目录下

## 配置说明

在启动 DistributedJob 前，需要配置 `config.yaml` 文件。以下是配置文件的主要参数：

### 服务器配置

```yaml
server:
  port: 9088              # HTTP 服务端口
  contextPath: /v1        # API 基础路径
  timeout: 10             # HTTP 请求超时时间 (秒)
```

### 数据库配置

```yaml
database:
  url: localhost:3306     # MySQL 服务器地址和端口
  username: root          # 数据库用户名
  password: 123456        # 数据库密码
  schema: scheduler       # 数据库名称
  maxConn: 10             # 最大连接数
  maxIdle: 5              # 最大空闲连接数
```

### 日志配置

```yaml
log:
  path: ./log             # 日志文件存储路径
  level: INFO             # 日志级别 (DEBUG, INFO, WARN, ERROR)
  maxSize: 100            # 单个日志文件大小上限 (MB)
  maxBackups: 10          # 最大日志文件备份数
  maxAge: 30              # 日志文件保存天数
```

### 任务配置

```yaml
job:
  workers: 5              # 工作线程数
  queueSize: 100          # 任务队列大小
  httpWorkers: 3          # HTTP 任务工作线程数
  grpcWorkers: 2          # gRPC 任务工作线程数
```

### 认证配置

```yaml
auth:
  jwtSecret: your-secret-key    # JWT 密钥
  jwtExpireHours: 24            # JWT 过期时间（小时）
  adminUsername: admin          # 默认管理员用户名
  adminPassword: admin123       # 默认管理员密码
```

## 数据库初始化

DistributedJob 在首次启动时会自动创建所需的数据库表。如果需要手动初始化数据库，请执行以下步骤：

1. 创建数据库

```sql
CREATE DATABASE IF NOT EXISTS scheduler DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

2. 创建部门表

```sql
CREATE TABLE IF NOT EXISTS `department` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `name` varchar(255) NOT NULL COMMENT '部门名称',
  `description` varchar(500) DEFAULT NULL COMMENT '部门描述',
  `parent_id` bigint(20) DEFAULT NULL COMMENT '父部门ID',
  `status` tinyint(4) NOT NULL DEFAULT '1' COMMENT '状态: 0-禁用, 1-启用',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_parent_id` (`parent_id`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='部门表';
```

3. 创建用户、角色和权限表

```sql
-- 用户表
CREATE TABLE IF NOT EXISTS `user` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `username` varchar(50) NOT NULL COMMENT '用户名',
  `password` varchar(100) NOT NULL COMMENT '密码',
  `real_name` varchar(50) NOT NULL COMMENT '真实姓名',
  `email` varchar(100) DEFAULT NULL COMMENT '电子邮箱',
  `phone` varchar(20) DEFAULT NULL COMMENT '手机号码',
  `department_id` bigint(20) NOT NULL COMMENT '所属部门ID',
  `role_id` bigint(20) NOT NULL COMMENT '角色ID',
  `status` tinyint(4) NOT NULL DEFAULT '1' COMMENT '状态: 0-禁用, 1-启用',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_username` (`username`),
  KEY `idx_department_id` (`department_id`),
  KEY `idx_role_id` (`role_id`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户表';

-- 角色表
CREATE TABLE IF NOT EXISTS `role` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `name` varchar(50) NOT NULL COMMENT '角色名称',
  `description` varchar(255) DEFAULT NULL COMMENT '角色描述',
  `status` tinyint(4) NOT NULL DEFAULT '1' COMMENT '状态: 0-禁用, 1-启用',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='角色表';

-- 权限表
CREATE TABLE IF NOT EXISTS `permission` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `name` varchar(50) NOT NULL COMMENT '权限名称',
  `code` varchar(50) NOT NULL COMMENT '权限编码',
  `description` varchar(255) DEFAULT NULL COMMENT '权限描述',
  `status` tinyint(4) NOT NULL DEFAULT '1' COMMENT '状态: 0-禁用, 1-启用',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_code` (`code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='权限表';

-- 角色权限关联表
CREATE TABLE IF NOT EXISTS `role_permission` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `role_id` bigint(20) NOT NULL COMMENT '角色ID',
  `permission_id` bigint(20) NOT NULL COMMENT '权限ID',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_role_permission` (`role_id`, `permission_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='角色权限关联表';
```

4. 创建任务表

```sql
-- 任务表
CREATE TABLE IF NOT EXISTS `task` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `name` varchar(255) NOT NULL COMMENT '任务名称',
  `department_id` bigint(20) NOT NULL COMMENT '所属部门ID',
  `task_type` varchar(20) NOT NULL DEFAULT 'HTTP' COMMENT '任务类型: HTTP、GRPC',
  `cron` varchar(100) NOT NULL COMMENT 'cron表达式',
  `url` varchar(500) DEFAULT NULL COMMENT '调度URL',
  `http_method` varchar(10) DEFAULT 'GET' COMMENT 'HTTP方法',
  `body` text COMMENT '请求体',
  `headers` text COMMENT '请求头',
  `grpc_service` varchar(255) DEFAULT NULL COMMENT 'gRPC服务名',
  `grpc_method` varchar(255) DEFAULT NULL COMMENT 'gRPC方法名',
  `grpc_params` text COMMENT 'gRPC参数',
  `retry_count` int(11) NOT NULL DEFAULT '0' COMMENT '最大重试次数',
  `retry_interval` int(11) NOT NULL DEFAULT '0' COMMENT '重试间隔(秒)',
  `fallback_url` varchar(500) DEFAULT NULL COMMENT '备用URL',
  `fallback_grpc_service` varchar(255) DEFAULT NULL COMMENT '备用gRPC服务名',
  `fallback_grpc_method` varchar(255) DEFAULT NULL COMMENT '备用gRPC方法名',
  `status` tinyint(4) NOT NULL DEFAULT '1' COMMENT '状态: 0-禁用, 1-启用',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `create_by` bigint(20) NOT NULL COMMENT '创建人ID',
  `update_by` bigint(20) DEFAULT NULL COMMENT '更新人ID',
  PRIMARY KEY (`id`),
  KEY `idx_department_id` (`department_id`),
  KEY `idx_status` (`status`),
  KEY `idx_task_type` (`task_type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='定时任务表';
```

5. 创建记录表

```sql
-- 记录表 (按年月分表，以202501为例)
CREATE TABLE IF NOT EXISTS `record_202501` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `task_id` bigint(20) NOT NULL COMMENT '任务ID',
  `task_name` varchar(255) NOT NULL COMMENT '任务名称',
  `task_type` varchar(20) NOT NULL DEFAULT 'HTTP' COMMENT '任务类型: HTTP、GRPC',
  `department_id` bigint(20) NOT NULL COMMENT '所属部门ID',
  `url` varchar(500) DEFAULT NULL COMMENT '调度URL',
  `http_method` varchar(10) DEFAULT NULL COMMENT 'HTTP方法',
  `body` text COMMENT '请求体',
  `headers` text COMMENT '请求头',
  `grpc_service` varchar(255) DEFAULT NULL COMMENT 'gRPC服务名',
  `grpc_method` varchar(255) DEFAULT NULL COMMENT 'gRPC方法名',
  `grpc_params` text COMMENT 'gRPC参数',
  `response` text COMMENT '响应内容',
  `status_code` int(11) DEFAULT NULL COMMENT 'HTTP状态码',
  `grpc_status` int(11) DEFAULT NULL COMMENT 'gRPC状态码',
  `success` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否成功',
  `retry_times` int(11) NOT NULL DEFAULT '0' COMMENT '重试次数',
  `use_fallback` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否使用了备用调用',
  `cost_time` int(11) NOT NULL DEFAULT '0' COMMENT '耗时(毫秒)',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_task_id` (`task_id`),
  KEY `idx_department_id` (`department_id`),
  KEY `idx_success` (`success`),
  KEY `idx_create_time` (`create_time`),
  KEY `idx_task_type` (`task_type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='任务执行记录表';
```

6. 初始化权限数据

```sql
-- 插入默认权限
INSERT INTO `permission` (`name`, `code`, `description`) VALUES 
('任务查看', 'task:view', '查看任务的权限'),
('任务创建', 'task:create', '创建任务的权限'),
('任务编辑', 'task:update', '编辑任务的权限'),
('任务删除', 'task:delete', '删除任务的权限'),
('记录查看', 'record:view', '查看执行记录的权限'),
('部门管理', 'department:manage', '管理部门的权限'),
('用户管理', 'user:manage', '管理用户的权限'),
('角色管理', 'role:manage', '管理角色的权限');

-- 插入管理员角色
INSERT INTO `role` (`name`, `description`) VALUES 
('管理员', '系统管理员，拥有所有权限');

-- 关联管理员角色和所有权限
INSERT INTO `role_permission` (`role_id`, `permission_id`) 
SELECT 1, id FROM `permission`;

-- 插入默认部门
INSERT INTO `department` (`name`, `description`, `parent_id`) VALUES 
('总部', '总部', NULL);

-- 插入管理员用户(密码需加密存储，这里使用明文演示)
INSERT INTO `user` (`username`, `password`, `real_name`, `department_id`, `role_id`) VALUES 
('admin', 'admin123', '系统管理员', 1, 1);
```

## 运行服务

### Linux/macOS 环境

```bash
cd <运行目录>
./main
```

### Windows 环境

双击 `main.exe` 运行程序，或在命令行中执行：

```cmd
cd <运行目录>
main.exe
```

## 验证安装

服务启动后，可通过以下方式验证安装是否成功：

1. 访问健康检查接口：http://localhost:9088/v1/health
   - 如果返回 HTTP 200 状态码，表示服务正常运行

2. 访问网页控制台：http://localhost:9088/v1/web/
   - 应该能够看到任务管理界面
   - 使用默认管理员账号 `admin` 和密码 `admin123` 登录系统

## 服务部署

### Docker 部署

1. 创建 Dockerfile

```dockerfile
FROM golang:1.17-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o main main.go

FROM alpine:3.14
WORKDIR /app
COPY --from=builder /app/main .
COPY config.yaml .
COPY web/ web/
EXPOSE 9088
CMD ["./main"]
```

2. 构建镜像

```bash
docker build -t distributed-job:latest .
```

3. 运行容器

```bash
docker run -d -p 9088:9088 --name distributed-job distributed-job:latest
```

### 多实例部署

DistributedJob 支持多实例部署，只需确保所有实例连接到同一个 MySQL 数据库即可。通过乐观锁机制，系统会确保同一时刻同一任务只会被一个服务实例执行。

多实例部署步骤：

1. 在不同服务器上部署多个 DistributedJob 实例
2. 配置所有实例使用相同的 MySQL 数据库
3. 可以通过负载均衡器分发 API 请求（例如 Nginx）

## 常见问题

1. **数据库连接失败**
   - 检查 config.yaml 中的数据库配置是否正确
   - 确认 MySQL 服务是否正常运行
   - 检查防火墙设置是否允许数据库连接

2. **服务无法启动**
   - 检查端口是否被占用：`netstat -ano | findstr 9088`（Windows）或 `lsof -i:9088`（Linux/macOS）
   - 查看日志文件了解详细错误信息

3. **任务不执行**
   - 检查任务状态是否为"启用"
   - 验证 cron 表达式是否正确
   - 确认目标 URL 是否可访问（HTTP任务）或gRPC服务是否可用（gRPC任务）
   - 检查用户是否有对应部门的任务执行权限

4. **gRPC 任务执行失败**
   - 确认 gRPC 服务地址正确且可访问
   - 检查 gRPC 参数格式是否正确
   - 确认目标服务的 gRPC 方法存在且参数类型匹配