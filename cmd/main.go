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

	"distributedJob/internal/api"
	"distributedJob/internal/config"
	"distributedJob/internal/infrastructure"
	"distributedJob/internal/job"
	"distributedJob/internal/rpc/server"
	"distributedJob/internal/service"
	"distributedJob/internal/store"
	"distributedJob/pkg/memory"
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

	ctx := context.Background()

	// 初始化基础设施
	infra := infrastructure.New()
	if err := infra.Initialize(ctx, cfg); err != nil {
		fmt.Printf("Failed to initialize infrastructure: %v\n", err)
		os.Exit(1)
	}
	defer infra.Shutdown(ctx)
	// 使用基础设施中的日志系统
	logger := infra.Logger
	logger.Infof("Starting DistributedJob service, version: %s", version)

	// 获取仓库管理器
	repoManager := infra.DB

	// 初始化令牌撤销器
	var tokenRevoker store.TokenRevoker
	if cfg.Auth.TokenRevocationStrategy == "redis" {
		tokenRevoker = infra.Redis.CreateTokenRevoker()
	} else {
		tokenRevoker = memory.NewMemoryTokenRevoker()
	}

	// 开启链路追踪
	tracer := infra.Tracer
	ctx, span := tracer.StartSpan(ctx, "application_startup")
	defer span.End()
	// 系统初始化（添加默认管理员用户等）
	if err := service.InitializeSystem(
		repoManager.User(),
		repoManager.Role(),
		repoManager.Department(),
		repoManager.Permission()); err != nil {
		logger.Warn("Failed to initialize system", "error", err)
	}

	// 初始化任务调度器，集成Kafka和指标
	schedulerOpts := []job.SchedulerOption{
		job.WithMetrics(infra.Metrics),
		job.WithTracer(infra.Tracer),
	}

	// 如果Kafka可用，配置为Kafka后端
	if infra.Kafka != nil {
		schedulerOpts = append(schedulerOpts, job.WithKafka(infra.Kafka))
	}

	// 如果Etcd可用，配置为分布式锁
	if infra.Etcd != nil {
		schedulerOpts = append(schedulerOpts, job.WithEtcd(infra.Etcd))
	}

	scheduler, err := job.NewScheduler(cfg, schedulerOpts...)
	if err != nil {
		logger.Fatalf("Failed to initialize scheduler: %v", err)
	}

	// 设置任务存储库
	scheduler.SetTaskRepository(repoManager.Task())

	// 创建服务
	authService := service.NewAuthService(
		repoManager.User(),
		repoManager.Role(),
		repoManager.Department(),
		repoManager.Permission(),
		cfg.Auth.JwtSecret,
		cfg.Auth.JwtRefreshSecret,
		time.Duration(cfg.Auth.JwtExpireMinutes)*time.Minute,
		time.Duration(cfg.Auth.JwtRefreshExpireDays)*24*time.Hour,
		tokenRevoker,
	)

	// 增加链路追踪
	authService.SetTracer(infra.Tracer)

	taskService := service.NewTaskService(repoManager.Task(), scheduler)
	taskService.SetTracer(infra.Tracer)
	taskService.SetMetrics(infra.Metrics)
	// 启动调度器
	if err := scheduler.Start(); err != nil {
		logger.Fatal("Failed to start scheduler", "error", err)
	}
	// 创建API服务器
	apiServer := api.NewServer(cfg, scheduler, repoManager, authService, tokenRevoker, infra)

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
		logger.Info("HTTP server listening", "address", addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("HTTP server failed", "error", err)
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
		logger.Error("HTTP server shutdown error", "error", err)
	}

	// 关闭数据库连接
	if manager, ok := repoManager.(interface{ Close() error }); ok {
		if err := manager.Close(); err != nil {
			logger.Error("Database connection close error", "error", err)
		}
	}

	logger.Info("Server gracefully stopped")
}
