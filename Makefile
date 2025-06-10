# Makefile for go-job project

.PHONY: help build run dev test clean docker docker-build docker-up docker-down proto deps lint fmt

# 默认目标
help:
	@echo "可用的命令:"
	@echo "  build       - 编译应用"
	@echo "  run         - 运行应用"
	@echo "  dev         - 开发模式运行"
	@echo "  test        - 运行测试"
	@echo "  clean       - 清理构建文件"
	@echo "  docker      - 构建并启动 Docker 容器"
	@echo "  docker-build - 构建 Docker 镜像"
	@echo "  docker-up   - 启动 Docker 容器"
	@echo "  docker-down - 停止 Docker 容器"
	@echo "  proto       - 生成 protobuf 文件"
	@echo "  deps        - 下载依赖"
	@echo "  lint        - 代码检查"
	@echo "  fmt         - 格式化代码"

# 应用相关
build:
	@echo "编译应用..."
	go build -o bin/go-job .

run: build
	@echo "运行应用..."
	./bin/go-job

dev:
	@echo "开发模式运行..."
	go run main.go

test:
	@echo "运行测试..."
	go test -v ./...

clean:
	@echo "清理构建文件..."
	rm -rf bin/
	go clean

# Docker 相关
docker: docker-build docker-up

docker-build:
	@echo "构建 Docker 镜像..."
	docker-compose build

docker-up:
	@echo "启动 Docker 容器..."
	docker-compose up -d

docker-down:
	@echo "停止 Docker 容器..."
	docker-compose down

docker-logs:
	@echo "查看日志..."
	docker-compose logs -f go-job

# 开发工具
proto:
	@echo "生成 protobuf 文件..."
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		api/grpc/*.proto

deps:
	@echo "下载依赖..."
	go mod download
	go mod tidy

lint:
	@echo "代码检查..."
	golangci-lint run

fmt:
	@echo "格式化代码..."
	go fmt ./...
	goimports -w .

# 数据库相关
db-migrate:
	@echo "数据库迁移..."
	go run main.go -migrate

db-seed:
	@echo "初始化数据..."
	go run main.go -seed

# 监控相关
monitoring-up:
	@echo "启动监控服务..."
	docker-compose --profile monitoring up -d

monitoring-down:
	@echo "停止监控服务..."
	docker-compose --profile monitoring down

# 日志相关
logging-up:
	@echo "启动日志服务..."
	docker-compose --profile logging up -d

logging-down:
	@echo "停止日志服务..."
	docker-compose --profile logging down

# 完整环境
full-up:
	@echo "启动完整环境..."
	docker-compose --profile monitoring --profile logging up -d

full-down:
	@echo "停止完整环境..."
	docker-compose --profile monitoring --profile logging down

# 前端相关
web-install:
	@echo "安装前端依赖..."
	cd web && npm install

web-build:
	@echo "构建前端..."
	cd web && npm run build

web-dev:
	@echo "前端开发模式..."
	cd web && npm run dev
