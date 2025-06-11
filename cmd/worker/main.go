package main

import (
	"context"
	"flag"
	"go-job/internal/worker"
	"go-job/pkg/config"
	"go-job/pkg/logger"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
)

func main() {
	var configPath = flag.String("config", "configs/config.yaml", "配置文件路径")
	var workerName = flag.String("name", "", "工作节点名称")
	flag.Parse()

	// 加载配置
	cfg, err := config.Load(*configPath)
	if err != nil {
		logrus.Fatalf("加载配置失败: %v", err)
	}

	// 初始化日志
	if err := logger.Init(cfg); err != nil {
		logrus.Fatalf("初始化日志失败: %v", err)
	}

	logrus.Info("启动 Go-Job Worker 节点...")

	// 创建工作节点
	workerInstance := worker.NewWorker(cfg)
	if *workerName != "" {
		workerInstance.SetName(*workerName)
	}

	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())

	// 启动工作节点
	go func() {
		if err := workerInstance.Start(ctx); err != nil {
			logrus.WithError(err).Fatal("工作节点启动失败")
		}
	}()

	// 等待退出信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logrus.Info("正在停止工作节点...")
	cancel()
	workerInstance.Stop()
	logrus.Info("工作节点已停止")
}
