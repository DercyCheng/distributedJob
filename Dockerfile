# 第一阶段：构建应用
FROM golang:1.22 AS builder

# 设置工作目录
WORKDIR /app

# 设置Go代理为国内源
ENV GOPROXY=https://goproxy.cn,direct

# 复制go.mod和go.sum文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制项目源码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o distributed_job ./cmd/server

# 第二阶段：创建最终镜像
FROM alpine:latest

# 安装依赖
RUN apk --no-cache add ca-certificates tzdata

# 设置工作目录
WORKDIR /app

# 复制配置文件和必要文件
COPY --from=builder /app/configs/config.yaml /app/configs/
COPY --from=builder /app/distributed_job /app/

# 创建日志目录
RUN mkdir -p /app/logs

# 暴露端口
EXPOSE 8080

# 运行应用
CMD ["/app/distributed_job"]