package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/distributedJob/internal/api"
	"github.com/distributedJob/internal/config"
	"github.com/distributedJob/internal/job"
	"github.com/distributedJob/internal/rpc/server"
	"github.com/distributedJob/internal/service"
	"github.com/distributedJob/internal/store/mysql"
	"github.com/distributedJob/pkg/logger"
)

var (
	configFile string
	version    string = "1.0.0"
	wait       bool
)

func init() {
	flag.StringVar(&configFile, "config", "configs/config.yaml", "配置文件路径")
	flag.BoolVar(&wait, "wait", false, "等待所有任务完成后再关闭服务")
}

func main() {
	// 解析命令行参数
	flag.Parse()

	// 加载配置
	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// 初始化日志
	logger.Init(
		cfg.Log.Level,
		cfg.Log.Filename,
		cfg.Log.MaxSize,
		cfg.Log.MaxBackups,
		cfg.Log.MaxAge,
		cfg.Log.Compress,
	)
	defer logger.Close()

	logger.Infof("Starting DistributedJob service, version: %s", version)

	// 初始化数据库连接
	repoManager, err := mysql.NewMySQLManager(cfg)
	if err != nil {
		logger.Fatalf("Failed to initialize database: %v", err)
	}

	// 系统初始化（添加默认管理员用户等）
	if err := service.InitializeSystem(
		repoManager.User(),
		repoManager.Role(),
		repoManager.Department(),
		repoManager.Permission(),
	); err != nil {
		logger.Warnf("Failed to initialize system: %v", err)
	}

	// 初始化任务调度器
	scheduler, err := job.NewScheduler(cfg)
	if err != nil {
		logger.Fatalf("Failed to initialize scheduler: %v", err)
	}

	// 设置任务存储库
	scheduler.SetTaskRepository(repoManager.Task())

	// 创建服务
	taskService := service.NewTaskService(repoManager.Task(), scheduler)
	authService := service.NewAuthService(
		repoManager.User(),
		repoManager.Role(),
		repoManager.Department(),
		repoManager.Permission(),
		cfg.Auth.JwtSecret,
		time.Duration(cfg.Auth.JwtExpireHours)*time.Hour,
	)

	// 启动调度器
	if err := scheduler.Start(); err != nil {
		logger.Fatalf("Failed to start scheduler: %v", err)
	}

	// 创建API服务器
	apiServer := api.NewServer(cfg, scheduler, repoManager)

	// 创建RPC服务器
	rpcServer := server.NewRPCServer(cfg, scheduler, taskService, authService)

	// 启动RPC服务器（异步）
	rpcServer.StartAsync()

	// 创建HTTP服务器
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	httpServer := &http.Server{
		Addr:    addr,
		Handler: apiServer.Router(),
	}

	// 启动HTTP服务器
	go func() {
		logger.Infof("HTTP server listening on %s", addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("HTTP server failed: %v", err)
		}
	}()

	// 等待终止信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// 创建上下文用于优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.Server.ShutdownTimeout)*time.Second)
	defer cancel()

	// 根据wait标志决定关闭方式
	if wait {
		logger.Info("Waiting for running tasks to complete...")
		// 让任务完成当前执行
		time.Sleep(2 * time.Second)
	}

	// 先停止调度器
	scheduler.Stop()

	// 关闭RPC服务器
	rpcServer.Stop()

	// 关闭HTTP服务器
	if err := httpServer.Shutdown(ctx); err != nil {
		logger.Errorf("HTTP server shutdown error: %v", err)
	}

	// 关闭数据库连接
	if manager, ok := repoManager.(interface{ Close() error }); ok {
		if err := manager.Close(); err != nil {
			logger.Errorf("Database connection close error: %v", err)
		}
	}

	logger.Info("Server gracefully stopped")
}
