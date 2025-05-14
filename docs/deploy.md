# DistributedJob - 部署文档

## Docker 容器化部署

DistributedJob 采用 Docker 容器化部署方案，通过 `docker-compose` 管理基础设施组件，简化部署和运维流程。所有基础设施如本地模型、MySQL、Redis 等均采用 docker-compose 部署，并使用阿里云镜像源加速拉取。本文档详细介绍如何使用 Docker 部署 DistributedJob 的各个组件，包括基础设施服务和本地 AI 模型。

## 基础设施概览

DistributedJob 的基础设施组件包括：

1. **数据库服务**

   - MySQL - 主数据存储
   - Redis - 缓存和消息队列
   - ETCD - 分布式协调和服务发现

2. **AI 服务**

   - 本地 AI 模型服务 (Qwen3、DeepseekV3 系列)
   - 向量数据库 (用于 RAG)

3. **消息系统**

   - Kafka - 事件流处理

4. **监控系统**

   - Prometheus - 指标收集
   - Grafana - 可视化监控面板

## 镜像源配置

为提高国内网络环境下的部署速度，我们默认使用阿里云镜像源。请在 Docker 配置中添加以下镜像源设置：

### Linux 系统

创建或编辑 `/etc/docker/daemon.json` 文件：

```json
{
  "registry-mirrors": ["https://registry.cn-hangzhou.aliyuncs.com"]
}
```

然后重启 Docker 服务：

```bash
sudo systemctl restart docker
```

### macOS 系统

在 Docker Desktop 的 "Settings" -> "Docker Engine" 中添加以下配置：

```json
{
  "registry-mirrors": ["https://registry.cn-hangzhou.aliyuncs.com"]
}
```

## Docker Compose 配置

所有基础设施组件通过 `docker-compose.yml` 文件进行统一管理。以下是示例配置：

```yaml
# docker-compose.yml
version: "3.8"

services:
  # 数据库服务
  mysql:
    image: registry.cn-hangzhou.aliyuncs.com/aliyun_mysql/mysql:8.0
    container_name: distributed_job_mysql
    environment:
      MYSQL_ROOT_PASSWORD: ${MYSQL_ROOT_PASSWORD}
      MYSQL_DATABASE: distributed_job
      MYSQL_USER: ${MYSQL_USER}
      MYSQL_PASSWORD: ${MYSQL_PASSWORD}
    ports:
      - "3306:3306"
    volumes:
      - mysql_data:/var/lib/mysql
      - ./scripts/init.sql:/docker-entrypoint-initdb.d/init.sql
    networks:
      - distributed_job_network
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-h", "localhost"]
      interval: 10s
      timeout: 5s
      retries: 5

  redis:
    image: registry.cn-hangzhou.aliyuncs.com/aliyun_redis/redis:7.0-alpine
    container_name: distributed_job_redis
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    networks:
      - distributed_job_network
    restart: unless-stopped
    command: redis-server --appendonly yes
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5

  etcd:
    image: registry.cn-hangzhou.aliyuncs.com/aliyun_etcd/bitnami-etcd:3.5
    container_name: distributed_job_etcd
    environment:
      - ALLOW_NONE_AUTHENTICATION=yes
    ports:
      - "2379:2379"
      - "2380:2380"
    volumes:
      - etcd_data:/bitnami/etcd
    networks:
      - distributed_job_network
    restart: unless-stopped

  # 消息系统
  kafka:
    image: registry.cn-hangzhou.aliyuncs.com/aliyun_kafka/kafka:3.0
    container_name: distributed_job_kafka
    depends_on:
      - zookeeper
    ports:
      - "9092:9092"
    environment:
      KAFKA_BROKER_ID: 1
      KAFKA_ZOOKEEPER_CONNECT: zookeeper:2181
      KAFKA_ADVERTISED_LISTENERS: PLAINTEXT://kafka:29092,PLAINTEXT_HOST://localhost:9092
      KAFKA_LISTENER_SECURITY_PROTOCOL_MAP: PLAINTEXT:PLAINTEXT,PLAINTEXT_HOST:PLAINTEXT
      KAFKA_INTER_BROKER_LISTENER_NAME: PLAINTEXT
      KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR: 1
    volumes:
      - kafka_data:/var/lib/kafka/data
    networks:
      - distributed_job_network
    restart: unless-stopped

  zookeeper:
    image: registry.cn-hangzhou.aliyuncs.com/aliyun_zookeeper/zookeeper:3.8
    container_name: distributed_job_zookeeper
    ports:
      - "2181:2181"
    environment:
      ZOOKEEPER_CLIENT_PORT: 2181
    volumes:
      - zookeeper_data:/var/lib/zookeeper/data
    networks:
      - distributed_job_network
    restart: unless-stopped

  # AI 服务 - 本地模型
  llm-server:
    image: distributedjob/llm-server:latest
    container_name: distributed_job_llm_server
    ports:
      - "8080:8080"
    volumes:
      - ./models:/app/models
    environment:
      - MODEL_PATH=/app/models
      - THREADS=8
      - GPU_LAYERS=0
    deploy:
      resources:
        reservations:
          devices:
            - driver: nvidia
              count: 1
              capabilities: [gpu]
    networks:
      - distributed_job_network
    restart: unless-stopped

  # 向量数据库
  qdrant:
    image: registry.cn-hangzhou.aliyuncs.com/aliyun_qdrant/qdrant:v1.7.0
    container_name: distributed_job_qdrant
    ports:
      - "6333:6333"
    volumes:
      - qdrant_data:/qdrant/storage
    networks:
      - distributed_job_network
    restart: unless-stopped

  # 监控系统
  prometheus:
    image: registry.cn-hangzhou.aliyuncs.com/aliyun_prometheus/prometheus:v2.45.0
    container_name: distributed_job_prometheus
    ports:
      - "9090:9090"
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml
      - prometheus_data:/prometheus
    networks:
      - distributed_job_network
    restart: unless-stopped

  grafana:
    image: registry.cn-hangzhou.aliyuncs.com/aliyun_grafana/grafana:10.0.0
    container_name: distributed_job_grafana
    ports:
      - "3000:3000"
    volumes:
      - grafana_data:/var/lib/grafana
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=${GRAFANA_PASSWORD}
    depends_on:
      - prometheus
    networks:
      - distributed_job_network
    restart: unless-stopped

  # 应用服务
  app:
    image: distributedjob/app:latest
    container_name: distributed_job_app
    depends_on:
      - mysql
      - redis
      - etcd
      - kafka
      - llm-server
      - qdrant
    ports:
      - "8000:8000"
    volumes:
      - ./configs:/app/configs
    networks:
      - distributed_job_network
    restart: unless-stopped

volumes:
  mysql_data:
  redis_data:
  etcd_data:
  kafka_data:
  zookeeper_data:
  qdrant_data:
  prometheus_data:
  grafana_data:

networks:
  distributed_job_network:
    driver: bridge
```

## 本地 AI 模型配置

### 1. 构建 LLM 服务容器

我们提供了基于 llama.cpp 的本地模型服务容器，支持 Qwen3、DeepseekV3 等多种模型。

```dockerfile
# Dockerfile for LLM Server
FROM ubuntu:22.04

RUN apt-get update && apt-get install -y \
    git \
    build-essential \
    cmake \
    python3 \
    python3-pip \
    curl \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# 克隆并构建 llama.cpp
RUN git clone https://github.com/ggerganov/llama.cpp && \
    cd llama.cpp && \
    make && \
    cp -r * /app/ && \
    cd /app && \
    rm -rf llama.cpp

# 安装 API 服务器依赖
RUN pip3 install fastapi uvicorn pydantic sse-starlette -i https://mirrors.aliyun.com/pypi/simple/

# 添加 API 服务器代码
COPY server.py /app/

# 创建模型目录
RUN mkdir -p /app/models

# 设置启动命令
CMD ["python3", "server.py"]

EXPOSE 8080
```

### 2. AI 模型配置

本系统采用以下 AI 模型：

1. **Qwen3 系列**

   - Qwen3-1.8B [下载链接](https://modelscope.cn/models/qwen/Qwen1.5-1.8B-Chat/files)
   - Qwen3-4B [下载链接](https://modelscope.cn/models/qwen/Qwen1.5-4B-Chat/files)
   - Qwen3-7B [下载链接](https://modelscope.cn/models/qwen/Qwen1.5-7B-Chat/files)

2. **DeepseekV3 系列** (采用低尺寸版本)

   - DeepseekV3-7B [下载链接](https://www.deepseek.com/research/deepseek-v3)
   - DeepseekV3-Coder-7B [下载链接](https://www.deepseek.com/research/deepseek-v3)

下载后，请转换为 GGUF 格式并放置在 `./models` 目录中：

```bash
# 安装转换工具
pip install llama-cpp-python -i https://mirrors.aliyun.com/pypi/simple/

# 转换模型
python -m llama_cpp.convert /path/to/model /path/to/output/model.gguf
```

### 3. 模型服务配置

在 `configs/config.yaml` 中配置本地模型，完整配置可在 `docs/config-example.yaml` 中找到，或参考 [config-example.yaml](config-example.yaml)。该示例配置文件包含所有必要的配置项和详细说明，强烈建议在部署前仔细阅读：

```yaml
ai:
  agent:
    enabled: true
    defaultLLM: deepseekv3-7b # 使用本地 DeepseekV3 低尺寸模型
  mcp:
    provider: local # 使用本地模型
    defaultModel: deepseekv3-7b
    local_models:
      enabled: true
      api_server:
        url: "http://llm-server:8080" # docker-compose 服务名称
        compatible_with: "openai"
      models:
        - name: "qwen3-1.8b" # qwen3 低尺寸模型
          path: "qwen3-1.8b-q4_k_m.gguf"
        - name: "qwen3-4b"
          path: "qwen3-4b-q5_k_m.gguf"
        - name: "qwen3-7b"
          path: "qwen3-7b-q5_k_m.gguf"
        - name: "deepseekv3-7b" # deepseekv3 低尺寸模型
          path: "deepseekv3-7b-q5_k_m.gguf"
        - name: "deepseekv3-coder-7b"
          path: "deepseekv3-coder-7b-q5_k_m.gguf"
```

## 部署步骤

### 1. 准备环境变量

创建 `.env` 文件，设置敏感配置：

```
MYSQL_ROOT_PASSWORD=your_root_password
MYSQL_USER=distributed_job
MYSQL_PASSWORD=your_password
GRAFANA_PASSWORD=admin_password
```

### 2. 准备模型文件

下载并转换模型文件，放入 `./models` 目录：

```bash
# 创建模型目录
mkdir -p models

# 下载并转换模型
# 这里是简化示例，实际下载和转换命令请参考模型提供商文档
curl -L -o models/qwen3-4b-q5_k_m.gguf https://huggingface.co/path/to/model.gguf
curl -L -o models/deepseekv3-7b-q5_k_m.gguf https://huggingface.co/path/to/model.gguf
```

### 3. 构建本地镜像

```bash
# 构建 LLM 服务镜像
docker build -t distributedjob/llm-server:latest -f Dockerfile.llm-server .

# 构建应用镜像
docker build -t distributedjob/app:latest .
```

### 4. 启动服务

```bash
# 启动所有服务
docker-compose up -d

# 查看服务状态
docker-compose ps
```

### 5. 查看日志

```bash
# 查看特定服务日志
docker-compose logs -f llm-server

# 查看应用日志
docker-compose logs -f app
```

## 系统访问

- 应用服务：http://localhost:8000
- Grafana 监控：http://localhost:3000 (默认用户名: admin, 密码: 在.env 中设置)
- Prometheus：http://localhost:9090

## 高级配置

### 1. GPU 加速

如果有 NVIDIA GPU 可用，可以启用 GPU 加速：

1. 安装 NVIDIA Container Toolkit
2. 在 docker-compose.yml 中启用 GPU：

```yaml
llm-server:
  # ...其他配置
  deploy:
    resources:
      reservations:
        devices:
          - driver: nvidia
            count: 1
            capabilities: [gpu]
  environment:
    - GPU_LAYERS=35 # 设置 GPU 层数
```

### 2. 扩展配置

对于生产环境，可以调整资源配置：

```yaml
llm-server:
  # ...其他配置
  deploy:
    resources:
      limits:
        cpus: "4"
        memory: 8G
```

### 3. 高可用配置

生产环境中，建议配置服务的高可用，如数据库主从复制、服务冗余等：

```yaml
mysql:
  # ...其他配置
  command: --default-authentication-plugin=mysql_native_password --character-set-server=utf8mb4 --collation-server=utf8mb4_unicode_ci --max_connections=1000
```

## 常见问题解决

1. **Docker 镜像拉取慢**

   - 确认已配置阿里云镜像源
   - 尝试先手动拉取大型镜像: `docker pull registry.cn-hangzhou.aliyuncs.com/aliyun_mysql/mysql:8.0`

2. **模型加载失败**

   - 检查模型格式是否为 GGUF
   - 检查模型路径配置是否正确
   - 查看 llm-server 容器日志寻找错误信息

3. **内存不足**

   - 对于小内存设备，尝试使用更小的模型如 Qwen3-1.8B
   - 调整量化级别，使用 Q4_K_M 减少内存占用

## 维护和更新

### 更新模型

1. 下载新模型到 `./models` 目录
2. 更新 `configs/config.yaml` 中的模型配置
3. 重启 llm-server: `docker-compose restart llm-server`

### 备份数据

定期备份重要数据：

```bash
# 备份 MySQL 数据
docker exec -it distributed_job_mysql mysqldump -u root -p distributed_job > backup.sql

# 备份 Redis 数据
docker exec -it distributed_job_redis redis-cli SAVE
```

### 更新系统

系统更新流程：

```bash
# 拉取最新代码
git pull

# 重新构建镜像
docker-compose build

# 更新服务
docker-compose up -d
```
