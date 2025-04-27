# DistributedJob

[![Go Report Card](https://goreportcard.com/badge/github.com/username/distributedJob)](https://goreportcard.com/report/github.com/username/distributedJob)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/go-1.16%2B-blue.svg)](https://golang.org/dl/)
[![MySQL](https://img.shields.io/badge/mysql-5.7%2B-blue.svg)](https://www.mysql.com/)

一个轻量级、分布式的定时任务调度平台，支持 HTTP 和 gRPC 任务类型。

## 📖 概述

DistributedJob 是一个基于 Go 和 MySQL 构建的分布式调度平台。它为跨分布式系统的定时任务调度和管理提供了简单可靠的解决方案，具备自动故障转移和全面的监控能力。

## ✨ 特性

- **分布式架构** - 无状态分布式服务设计，使用乐观锁确保每个任务只在一个实例上执行
- **多种任务类型** - 支持 HTTP 钩子和 gRPC 调用
- **Web 控制台** - 内置可视化管理界面，用于配置和监控任务
- **强大的重试机制** - 可配置的重试策略和备用端点
- **执行历史记录** - 全面的执行记录，支持自动表分区
- **优雅关闭** - 健康检查和平滑服务终止
- **部门管理** - 按部门组织任务，便于分类
- **精细化权限控制** - 全面的权限系统，控制用户访问
- **响应式界面** - 现代响应式设计，同时支持桌面和移动设备

## 🚀 快速开始

### 前置条件

- Go 1.16+
- MySQL 5.7+
- 任意操作系统（Windows、macOS、Linux）

### 安装

```bash
# 克隆仓库
git clone https://github.com/username/distributedJob.git
cd distributedJob

# 从源代码构建
go build -o distributedJob ./cmd/server/main.go

# 在 config.yaml 中配置数据库
# 启动服务
./distributedJob
```

有关详细的安装说明，请参阅[安装指南](./doc/installation.md)。

## 📚 文档

- [架构设计](./doc/architecture.md) - 系统架构和组件详情
- [安装指南](./doc/installation.md) - 详细的安装和部署指南
- [使用指南](./doc/usage.md) - 系统使用方法和示例
- [API 参考](./doc/api.md) - API 文档
- [数据库设计](./doc/database.md) - 数据库架构和数据模型
- [前端设计](./doc/ui.md) - 界面实现详情

## 💡 使用场景

- 定时数据处理和 ETL 任务
- 定期健康检查和系统监控
- 周期性数据同步
- 自动化报表生成和分发
- 定时清理和维护任务

## 🔧 技术栈

### 后端

- **语言**：Go
- **数据库**：MySQL
- **API**：RESTful API

### 前端

- **构建工具**：Vite
- **语言**：TypeScript
- **框架**：Vue 3 (Composition API)
- **UI 库**：Element Plus
- **状态管理**：Pinia
- **HTTP 客户端**：Axios

## 📝 许可证

本项目采用 MIT 许可证 - 详情请查看 [LICENSE](LICENSE) 文件。

## 🤝 贡献

欢迎贡献！请随时提交 Pull Request。
