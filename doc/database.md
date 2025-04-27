# 数据库设计

本文档详细说明 DistributedJob 的数据库设计。

## 数据库概述

DistributedJob 使用 MySQL 数据库存储任务配置和执行记录。数据库设计遵循以下原则：

- 简单实用：只设计必要的表结构，减少复杂度
- 良好性能：合理的索引设计，优化查询性能
- 数据分区：按年月对执行记录进行分表，提高大数据量下的查询效率
- 权限分离：清晰的权限模型，支持多部门管理和权限控制

## 表结构设计

### 部门表 (department)

部门表存储系统中的部门信息。

| 字段名 | 数据类型 | 是否为空 | 默认值 | 说明 |
|-------|---------|---------|-------|------|
| id | bigint(20) | 否 | 自增 | 主键 |
| name | varchar(255) | 否 | 无 | 部门名称 |
| description | varchar(500) | 是 | NULL | 部门描述 |
| parent_id | bigint(20) | 是 | NULL | 父部门ID，顶级部门为NULL |
| status | tinyint(4) | 否 | 1 | 状态：0-禁用，1-启用 |
| create_time | datetime | 否 | CURRENT_TIMESTAMP | 创建时间 |
| update_time | datetime | 否 | CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP | 更新时间 |

索引：
- PRIMARY KEY (`id`)
- KEY `idx_parent_id` (`parent_id`)
- KEY `idx_status` (`status`)

建表 SQL：

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

### 用户表 (user)

用户表存储系统用户信息。

| 字段名 | 数据类型 | 是否为空 | 默认值 | 说明 |
|-------|---------|---------|-------|------|
| id | bigint(20) | 否 | 自增 | 主键 |
| username | varchar(50) | 否 | 无 | 用户名 |
| password | varchar(100) | 否 | 无 | 密码（加密存储） |
| real_name | varchar(50) | 否 | 无 | 真实姓名 |
| email | varchar(100) | 是 | NULL | 电子邮箱 |
| phone | varchar(20) | 是 | NULL | 手机号码 |
| department_id | bigint(20) | 否 | 无 | 所属部门ID |
| role_id | bigint(20) | 否 | 无 | 角色ID |
| status | tinyint(4) | 否 | 1 | 状态：0-禁用，1-启用 |
| create_time | datetime | 否 | CURRENT_TIMESTAMP | 创建时间 |
| update_time | datetime | 否 | CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP | 更新时间 |

索引：
- PRIMARY KEY (`id`)
- UNIQUE KEY `idx_username` (`username`)
- KEY `idx_department_id` (`department_id`)
- KEY `idx_role_id` (`role_id`)
- KEY `idx_status` (`status`)

建表 SQL：

```sql
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
```

### 角色表 (role)

角色表存储系统角色信息。

| 字段名 | 数据类型 | 是否为空 | 默认值 | 说明 |
|-------|---------|---------|-------|------|
| id | bigint(20) | 否 | 自增 | 主键 |
| name | varchar(50) | 否 | 无 | 角色名称 |
| description | varchar(255) | 是 | NULL | 角色描述 |
| status | tinyint(4) | 否 | 1 | 状态：0-禁用，1-启用 |
| create_time | datetime | 否 | CURRENT_TIMESTAMP | 创建时间 |
| update_time | datetime | 否 | CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP | 更新时间 |

索引：
- PRIMARY KEY (`id`)
- UNIQUE KEY `idx_name` (`name`)

建表 SQL：

```sql
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
```

### 权限表 (permission)

权限表存储系统权限信息。

| 字段名 | 数据类型 | 是否为空 | 默认值 | 说明 |
|-------|---------|---------|-------|------|
| id | bigint(20) | 否 | 自增 | 主键 |
| name | varchar(50) | 否 | 无 | 权限名称 |
| code | varchar(50) | 否 | 无 | 权限编码 |
| description | varchar(255) | 是 | NULL | 权限描述 |
| status | tinyint(4) | 否 | 1 | 状态：0-禁用，1-启用 |
| create_time | datetime | 否 | CURRENT_TIMESTAMP | 创建时间 |
| update_time | datetime | 否 | CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP | 更新时间 |

索引：
- PRIMARY KEY (`id`)
- UNIQUE KEY `idx_code` (`code`)

建表 SQL：

```sql
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
```

### 角色权限关联表 (role_permission)

角色权限关联表存储角色与权限的多对多关系。

| 字段名 | 数据类型 | 是否为空 | 默认值 | 说明 |
|-------|---------|---------|-------|------|
| id | bigint(20) | 否 | 自增 | 主键 |
| role_id | bigint(20) | 否 | 无 | 角色ID |
| permission_id | bigint(20) | 否 | 无 | 权限ID |
| create_time | datetime | 否 | CURRENT_TIMESTAMP | 创建时间 |

索引：
- PRIMARY KEY (`id`)
- UNIQUE KEY `idx_role_permission` (`role_id`, `permission_id`)

建表 SQL：

```sql
CREATE TABLE IF NOT EXISTS `role_permission` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `role_id` bigint(20) NOT NULL COMMENT '角色ID',
  `permission_id` bigint(20) NOT NULL COMMENT '权限ID',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_role_permission` (`role_id`, `permission_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='角色权限关联表';
```

### 任务表 (task)

任务表存储所有定时任务的配置信息。

| 字段名 | 数据类型 | 是否为空 | 默认值 | 说明 |
|-------|---------|---------|-------|------|
| id | bigint(20) | 否 | 自增 | 主键 |
| name | varchar(255) | 否 | 无 | 任务名称 |
| department_id | bigint(20) | 否 | 无 | 所属部门ID |
| task_type | varchar(20) | 否 | 'HTTP' | 任务类型：HTTP、GRPC |
| cron | varchar(100) | 否 | 无 | cron 表达式 |
| url | varchar(500) | 是 | NULL | 调度URL（HTTP任务时必填） |
| http_method | varchar(10) | 是 | 'GET' | HTTP方法（HTTP任务时必填） |
| body | text | 是 | NULL | 请求体（HTTP任务时选填） |
| headers | text | 是 | NULL | 请求头（HTTP任务时选填，JSON格式字符串） |
| grpc_service | varchar(255) | 是 | NULL | gRPC服务名（gRPC任务时必填） |
| grpc_method | varchar(255) | 是 | NULL | gRPC方法名（gRPC任务时必填） |
| grpc_params | text | 是 | NULL | gRPC参数（gRPC任务时选填，JSON格式字符串） |
| retry_count | int(11) | 否 | 0 | 最大重试次数 |
| retry_interval | int(11) | 否 | 0 | 重试间隔（秒） |
| fallback_url | varchar(500) | 是 | NULL | 备用URL（HTTP任务时选填） |
| fallback_grpc_service | varchar(255) | 是 | NULL | 备用gRPC服务名（gRPC任务时选填） |
| fallback_grpc_method | varchar(255) | 是 | NULL | 备用gRPC方法名（gRPC任务时选填） |
| status | tinyint(4) | 否 | 1 | 状态：0-禁用，1-启用 |
| create_time | datetime | 否 | CURRENT_TIMESTAMP | 创建时间 |
| update_time | datetime | 否 | CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP | 更新时间 |
| create_by | bigint(20) | 否 | 无 | 创建人ID |
| update_by | bigint(20) | 是 | NULL | 更新人ID |

索引：
- PRIMARY KEY (`id`)
- KEY `idx_department_id` (`department_id`)
- KEY `idx_status` (`status`)
- KEY `idx_task_type` (`task_type`)

建表 SQL：

```sql
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

### 执行记录表 (record_YYYYMM)

执行记录表存储任务的每次执行记录，按年月分表，表名格式为 `record_YYYYMM`，例如 `record_202501` 表示 2025 年 1 月的执行记录。

| 字段名 | 数据类型 | 是否为空 | 默认值 | 说明 |
|-------|---------|---------|-------|------|
| id | bigint(20) | 否 | 自增 | 主键 |
| task_id | bigint(20) | 否 | 无 | 任务 ID，关联 task 表的 id |
| task_name | varchar(255) | 否 | 无 | 任务名称（冗余存储，方便查询） |
| task_type | varchar(20) | 否 | 'HTTP' | 任务类型：HTTP、GRPC |
| department_id | bigint(20) | 否 | 无 | 所属部门ID |
| url | varchar(500) | 是 | NULL | 实际调用的 URL（HTTP任务时必填） |
| http_method | varchar(10) | 是 | NULL | HTTP 方法（HTTP任务时必填） |
| body | text | 是 | NULL | 请求体 |
| headers | text | 是 | NULL | 请求头 |
| grpc_service | varchar(255) | 是 | NULL | gRPC服务名（gRPC任务时必填） |
| grpc_method | varchar(255) | 是 | NULL | gRPC方法名（gRPC任务时必填） |
| grpc_params | text | 是 | NULL | gRPC参数 |
| response | text | 是 | NULL | 响应内容 |
| status_code | int(11) | 是 | NULL | HTTP状态码（HTTP任务时填写） |
| grpc_status | int(11) | 是 | NULL | gRPC状态码（gRPC任务时填写） |
| success | tinyint(1) | 否 | 0 | 是否成功：0-失败，1-成功 |
| retry_times | int(11) | 否 | 0 | 重试次数 |
| use_fallback | tinyint(1) | 否 | 0 | 是否使用了备用调用：0-否，1-是 |
| cost_time | int(11) | 否 | 0 | 耗时（毫秒） |
| create_time | datetime | 否 | CURRENT_TIMESTAMP | 创建时间 |

索引：
- PRIMARY KEY (`id`)
- KEY `idx_task_id` (`task_id`)
- KEY `idx_department_id` (`department_id`)
- KEY `idx_success` (`success`)
- KEY `idx_create_time` (`create_time`)
- KEY `idx_task_type` (`task_type`)

建表 SQL（以 2025 年 1 月为例）：

```sql
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

## 数据分表策略

DistributedJob 对执行记录表采用按年月分表的策略，主要考虑以下几点：

1. **数据量控制**：定时任务执行记录可能会随着时间快速增长，分表可以控制单表数据量
2. **查询性能**：用户通常只关注最近的执行记录，按年月分表可以提高查询效率
3. **数据清理**：便于按时间范围清理旧数据，只需删除整张表即可

### 分表命名规则

分表采用 `record_YYYYMM` 格式命名，其中：
- `YYYY` 表示年份，如 2025
- `MM` 表示月份，如 01 表示一月

例如：
- `record_202501` 表示 2025 年 1 月的记录表
- `record_202502` 表示 2025 年 2 月的记录表

### 表创建规则

系统会在启动时以及每月开始时，自动检查并创建当月的记录表。

## ER 图

下图展示了 DistributedJob 的实体关系图：

```
+-----------------+        +------------------+        +-----------------------+
|   department    |        |       task       |        |    record_YYYYMM     |
+-----------------+        +------------------+        +-----------------------+
| id (PK)         |        | id (PK)          |        | id (PK)              |
| name            |  1   n | department_id (FK)|  1   n| task_id (FK)         |
| description     |------->| name             |------->| task_name            |
| parent_id       |        | task_type        |        | task_type            |
| status          |        | cron             |        | department_id (FK)   |
| create_time     |        | url              |        | url                  |
| update_time     |        | http_method      |        | http_method          |
+-----------------+        | body             |        | body                 |
                           | headers          |        | headers              |
                           | grpc_service     |        | grpc_service         |
+-----------------+        | grpc_method      |        | grpc_method          |
|      user       |        | grpc_params      |        | grpc_params          |
+-----------------+        | retry_count      |        | response             |
| id (PK)         |        | retry_interval   |        | status_code          |
| username        |        | fallback_url     |        | grpc_status          |
| password        |        | fallback_grpc_service|     | success              |
| real_name       |        | fallback_grpc_method |     | retry_times          |
| email           |        | status           |        | use_fallback         |
| phone           |        | create_time      |        | cost_time            |
| department_id (FK)       | update_time      |        | create_time          |
| role_id (FK)    |        | create_by (FK)   |        +-----------------------+
| status          |        | update_by (FK)   |
| create_time     |        +------------------+
| update_time     |
+-----------------+
        |
        |1
        v
        n
+-----------------+        +--------------------+
|      role       |        |  role_permission   |
+-----------------+        +--------------------+
| id (PK)         |  1   n | id (PK)            |
| name            |------->| role_id (FK)       |
| description     |        | permission_id (FK) |
| status          |        | create_time        |
| create_time     |        +--------------------+
| update_time     |               n|
+-----------------+                |
                                   |
                                   v1
                          +-----------------+
                          |   permission    |
                          +-----------------+
                          | id (PK)         |
                          | name            |
                          | code            |
                          | description     |
                          | status          |
                          | create_time     |
                          | update_time     |
                          +-----------------+
```

## 数据库优化建议

### 索引优化

- 任务表 (`task`) 已添加 `department_id`、`status`、`task_type` 字段的索引，用于优化常见查询场景
- 记录表已添加 `task_id`、`department_id`、`success`、`create_time`、`task_type` 字段的索引
- 如果经常按任务名称关键字查询，可考虑在 `task` 表的 `name` 字段上创建索引

### 大数据量优化

- 记录表已按年月分表，但长期运行后仍可能有大量历史数据
- 建议实现自动归档策略，如保留最近 6 个月的记录，将更早的记录归档或清理
- 对于需要长期保存的记录，可导出到其他存储系统或归档数据库

### 并发控制

- 任务调度采用乐观锁控制并发，确保同一任务不会被多个实例同时执行
- 在 MySQL 配置中适当调整 `max_connections` 参数，确保足够的连接数

### 数据备份

- 定期备份数据库，保证数据安全
- 可使用 MySQL 自带的备份工具如 mysqldump 进行备份
- 示例备份命令：

  ```bash
  mysqldump -u username -p scheduler > scheduler_backup_$(date +%Y%m%d).sql
  ```