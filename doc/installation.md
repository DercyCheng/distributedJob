# 安装指南

<div align="center">
  <h3>DistributedJob 安装指南</h3>
</div>

本指南提供了安装和部署 DistributedJob 的全面说明。

## 目录

- [系统要求](#系统要求)
- [安装方法](#安装方法)
  - [源码安装](#源码安装)
  - [二进制安装](#二进制安装)
  - [Docker 安装](#docker-安装)
- [配置](#配置)
- [数据库设置](#数据库设置)
- [运行服务](#运行服务)
- [验证](#验证)
- [部署选项](#部署选项)
  - [单实例部署](#单实例部署)
  - [多实例部署](#多实例部署)
  - [容器化部署](#容器化部署)

## 系统要求

在安装 DistributedJob 之前，请确保您的系统满足以下要求：

| 组件 | 最低要求 |
|-----------|---------------------|
| Go | 1.16 或更高版本 |
| MySQL | 5.7 或更高版本 |
| 操作系统 | Windows、macOS 或 Linux |
| 内存 | 2GB RAM（推荐） |
| 磁盘空间 | 应用程序 200MB，外加日志和数据空间 |
| gRPC | 需要 gRPC 任务功能 |

## 安装方法

### 源码安装

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

### 二进制安装

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

### Docker 安装

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

## 配置

通过编辑 `config.yaml` 文件配置 DistributedJob：

### 服务器配置

```yaml
server:
  port: 9088              # HTTP 服务端口
  contextPath: /v1        # API 基础路径
  timeout: 10             # HTTP 请求超时（秒）
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
  level: INFO             # 日志级别（DEBUG、INFO、WARN、ERROR）
  maxSize: 100            # 单个日志文件的最大大小（MB）
  maxBackups: 10          # 日志文件备份的最大数量
  maxAge: 30              # 日志文件保留天数
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
  jwtSecret: your-secret-key    # JWT 密钥（请更改此项！）
  jwtExpireHours: 24            # JWT 过期时间（小时）
  adminUsername: admin          # 默认管理员用户名
  adminPassword: admin123       # 默认管理员密码（请更改此项！）
```

## 数据库设置

DistributedJob 在首次启动时会自动创建必要的数据库表。但是，您也可以手动初始化数据库。

### 手动数据库初始化

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

## 运行服务

### Linux/macOS

```bash
# 导航到安装目录
cd /opt/distributedJob

# 运行服务
./distributedJob

# 作为后台服务运行
nohup ./distributedJob > /dev/null 2>&1 &
```

### Windows

```cmd
# 导航到安装目录
cd C:\distributedJob

# 运行服务
distributedJob.exe
```

### 使用 Systemd（Linux）

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

### 使用 Docker

```bash
docker run -d \
  --name distributed-job \
  -p 9088:9088 \
  -v /data/distributed-job/configs:/app/configs \
  -v /data/distributed-job/log:/app/log \
  username/distributed-job:latest
```

## 验证

启动服务后，验证其是否正常运行：

1. **检查健康端点**

   ```bash
   curl http://localhost:9088/v1/health
   ```

   预期响应：
   ```json
   {"code":0,"message":"success","data":{"status":"up","timestamp":"2023-01-01T12:00:00Z"}}
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

## 部署选项

### 单实例部署

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

### 多实例部署

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

### 容器化部署

使用 Docker 和 Docker Compose 部署提供了灵活性和隔离性：

1. **Docker Compose 配置**

   创建 `docker-compose.yml` 文件：

   ```yaml
   version: '3'

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