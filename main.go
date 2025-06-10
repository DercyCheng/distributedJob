package main

import (
	"context"
	"fmt"
	grpcapi "go-job/api/grpc"
	httpapi "go-job/api/http"
	authservice "go-job/internal/auth"
	"go-job/internal/department"
	"go-job/internal/job"
	"go-job/internal/mcp"
	"go-job/internal/permission"
	"go-job/internal/role"
	"go-job/internal/scheduler"
	"go-job/internal/user"
	"go-job/pkg/auth"
	"go-job/pkg/config"
	"go-job/pkg/database"
	"go-job/pkg/logger"
	"go-job/pkg/redis"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

func main() {
	// 加载配置
	cfg, err := config.Load("configs/config.yaml")
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 初始化日志
	if err := logger.Init(cfg); err != nil {
		log.Fatalf("初始化日志失败: %v", err)
	}

	logrus.Info("启动 go-job 服务...")

	// 初始化数据库
	if err := database.Init(cfg); err != nil {
		logrus.Fatalf("初始化数据库失败: %v", err)
	}
	logrus.Info("数据库连接成功")

	// 初始化Redis
	if err := redis.Init(cfg); err != nil {
		logrus.Fatalf("初始化Redis失败: %v", err)
	}
	logrus.Info("Redis连接成功")

	// 初始化服务
	services := initServices(cfg)
	logrus.Info("服务初始化完成")

	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 启动服务
	var wg sync.WaitGroup

	// 启动HTTP服务器
	wg.Add(1)
	go func() {
		defer wg.Done()
		startHTTPServer(ctx, cfg, services)
	}()

	// 启动gRPC服务器
	wg.Add(1)
	go func() {
		defer wg.Done()
		startGRPCServer(ctx, cfg, services)
	}()

	// 启动调度器
	wg.Add(1)
	go func() {
		defer wg.Done()
		startScheduler(ctx, cfg, services.JobService)
	}()

	// 等待退出信号
	waitForShutdown(cancel)

	// 等待所有协程结束
	wg.Wait()
	logrus.Info("服务已停止")
}

// initServices 初始化所有服务
func initServices(cfg *config.Config) *httpapi.Services {
	db := database.GetDB()

	// 创建JWT管理器
	jwtManager := auth.NewJWTManager(cfg.JWT.Secret, cfg.JWT.Expire)

	// 初始化基础服务
	authService := authservice.NewAuthService(db, jwtManager)
	userService := user.NewUserService(db)
	departmentService := department.NewDepartmentService(db)
	roleService := role.NewRoleService(db)
	permissionService := permission.NewPermissionService(db)
	jobService := job.NewService()

	// 初始化AI相关服务
	aiScheduler := mcp.NewAISchedulerService(db, cfg)
	mcpService := mcp.NewMCPService(db, aiScheduler)

	return &httpapi.Services{
		AuthService:       authService,
		UserService:       userService,
		DepartmentService: departmentService,
		RoleService:       roleService,
		PermissionService: permissionService,
		JobService:        jobService,
		AIScheduler:       aiScheduler,
		MCPService:        mcpService,
	}
}

// startHTTPServer 启动HTTP服务器
func startHTTPServer(ctx context.Context, cfg *config.Config, services *httpapi.Services) {
	// 设置Gin模式
	if cfg.Logger.Level == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建路由
	router := httpapi.NewRouter(cfg, services)

	// 配置HTTP服务器
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", cfg.Server.HTTP.Host, cfg.Server.HTTP.Port),
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 启动服务器
	go func() {
		logrus.Infof("HTTP服务器启动在 %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Fatalf("HTTP服务器启动失败: %v", err)
		}
	}()

	// 等待关闭信号
	<-ctx.Done()

	// 优雅关闭
	logrus.Info("正在关闭HTTP服务器...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logrus.Errorf("HTTP服务器关闭失败: %v", err)
	} else {
		logrus.Info("HTTP服务器已关闭")
	}
}

// startGRPCServer 启动gRPC服务器
func startGRPCServer(ctx context.Context, cfg *config.Config, services *httpapi.Services) {
	// 创建gRPC服务器
	s := grpc.NewServer()

	// 注册服务
	grpcapi.RegisterJobServiceServer(s, &grpcJobServer{jobService: services.JobService})
	grpcapi.RegisterAuthServiceServer(s, &grpcAuthServer{authService: services.AuthService})
	grpcapi.RegisterUserServiceServer(s, &grpcUserServer{userService: services.UserService})
	grpcapi.RegisterDepartmentServiceServer(s, &grpcDepartmentServer{departmentService: services.DepartmentService})
	grpcapi.RegisterRoleServiceServer(s, &grpcRoleServer{roleService: services.RoleService})
	grpcapi.RegisterPermissionServiceServer(s, &grpcPermissionServer{permissionService: services.PermissionService})
	grpcapi.RegisterMCPServiceServer(s, &grpcMCPServer{mcpService: services.MCPService})
	grpcapi.RegisterAISchedulerServiceServer(s, &grpcAISchedulerServer{aiScheduler: services.AIScheduler})

	// 监听端口
	addr := fmt.Sprintf("%s:%s", cfg.Server.GRPC.Host, cfg.Server.GRPC.Port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		logrus.Fatalf("gRPC服务器监听失败: %v", err)
	}

	// 启动服务器
	go func() {
		logrus.Infof("gRPC服务器启动在 %s", addr)
		if err := s.Serve(lis); err != nil {
			logrus.Fatalf("gRPC服务器启动失败: %v", err)
		}
	}()

	// 等待关闭信号
	<-ctx.Done()

	// 优雅关闭
	logrus.Info("正在关闭gRPC服务器...")
	s.GracefulStop()
	logrus.Info("gRPC服务器已关闭")
}

// startScheduler 启动调度器
func startScheduler(ctx context.Context, cfg *config.Config, jobService *job.Service) {
	schedulerService := scheduler.NewService(cfg)

	logrus.Info("启动任务调度器...")
	if err := schedulerService.Start(ctx); err != nil {
		logrus.Errorf("调度器启动失败: %v", err)
	}
}

// waitForShutdown 等待关闭信号
func waitForShutdown(cancel context.CancelFunc) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	sig := <-c
	logrus.Infof("收到信号 %v，开始关闭服务...", sig)
	cancel()
}

// gRPC服务器实现结构体
type grpcJobServer struct {
	grpcapi.UnimplementedJobServiceServer
	jobService *job.Service
}

type grpcAuthServer struct {
	grpcapi.UnimplementedAuthServiceServer
	authService *authservice.AuthService
}

type grpcUserServer struct {
	grpcapi.UnimplementedUserServiceServer
	userService *user.UserService
}

type grpcDepartmentServer struct {
	grpcapi.UnimplementedDepartmentServiceServer
	departmentService *department.DepartmentService
}

type grpcRoleServer struct {
	grpcapi.UnimplementedRoleServiceServer
	roleService *role.RoleService
}

type grpcPermissionServer struct {
	grpcapi.UnimplementedPermissionServiceServer
	permissionService *permission.PermissionService
}

type grpcMCPServer struct {
	grpcapi.UnimplementedMCPServiceServer
	mcpService *mcp.MCPService
}

type grpcAISchedulerServer struct {
	grpcapi.UnimplementedAISchedulerServiceServer
	aiScheduler *mcp.AISchedulerService
}
