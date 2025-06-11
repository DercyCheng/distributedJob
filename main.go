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
	"go-job/pkg/broadcaster"
	"go-job/pkg/config"
	"go-job/pkg/database"
	"go-job/pkg/logger"
	"go-job/pkg/redis"
	"go-job/pkg/websocket"
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

	// 初始化默认数据
	if err := database.InitDefaultData(); err != nil {
		logrus.Fatalf("初始化默认数据失败: %v", err)
	}
	logrus.Info("默认数据初始化完成")

	// 初始化WebSocket Hub
	wsHub := websocket.NewHub()
	go wsHub.Run()
	logrus.Info("WebSocket Hub启动完成")

	// 初始化服务
	services, schedulerService := initServices(cfg, wsHub)
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
		startGRPCServer(ctx, cfg, services, schedulerService)
	}()

	// 启动调度器
	wg.Add(1)
	go func() {
		defer wg.Done()
		startScheduler(ctx, cfg, schedulerService)
	}()

	// 启动统计数据广播服务
	wg.Add(1)
	go func() {
		defer wg.Done()
		startStatsBroadcaster(ctx, services)
	}()

	// 等待退出信号
	waitForShutdown(cancel)

	// 等待所有协程结束
	wg.Wait()
	logrus.Info("服务已停止")
}

// initServices 初始化所有服务
func initServices(cfg *config.Config, wsHub *websocket.Hub) (*httpapi.Services, *scheduler.Service) {
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

	// 初始化调度器服务
	schedulerService := scheduler.NewService(cfg)

	// 初始化AI相关服务
	aiScheduler := mcp.NewAISchedulerService(db, cfg)
	mcpService := mcp.NewMCPService(db, aiScheduler)

	services := &httpapi.Services{
		AuthService:       authService,
		UserService:       userService,
		DepartmentService: departmentService,
		RoleService:       roleService,
		PermissionService: permissionService,
		JobService:        jobService,
		AIScheduler:       aiScheduler,
		MCPService:        mcpService,
		WSHub:             wsHub,
	}

	return services, schedulerService
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
func startGRPCServer(ctx context.Context, cfg *config.Config, services *httpapi.Services, schedulerService *scheduler.Service) {
	// 创建gRPC服务器
	s := grpc.NewServer()

	// 注册服务
	grpcapi.RegisterJobServiceServer(s, &grpcJobServer{jobService: services.JobService})
	grpcapi.RegisterSchedulerServiceServer(s, &grpcSchedulerServer{schedulerService: schedulerService})
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
func startScheduler(ctx context.Context, _ *config.Config, schedulerService *scheduler.Service) {
	logrus.Info("启动任务调度器...")
	if err := schedulerService.Start(ctx); err != nil {
		logrus.Errorf("调度器启动失败: %v", err)
	}
}

// startStatsBroadcaster 启动统计数据广播服务
func startStatsBroadcaster(ctx context.Context, services *httpapi.Services) {
	logrus.Info("启动统计数据广播服务...")
	broadcaster := broadcaster.NewStatsBroadcaster(services.WSHub)
	broadcaster.Start(ctx)
}

// waitForShutdown 等待关闭信号
func waitForShutdown(cancel context.CancelFunc) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	sig := <-c
	logrus.Infof("收到信号 %v，开始关闭服务...", sig)
	cancel()
}

// 添加调度器服务器
type grpcSchedulerServer struct {
	grpcapi.UnimplementedSchedulerServiceServer
	schedulerService *scheduler.Service
}

// gRPC服务器实现结构体
type grpcJobServer struct {
	grpcapi.UnimplementedJobServiceServer
	jobService *job.Service
}

// JobService gRPC 方法实现
func (s *grpcJobServer) CreateJob(ctx context.Context, req *grpcapi.CreateJobRequest) (*grpcapi.CreateJobResponse, error) {
	return s.jobService.CreateJob(ctx, req)
}

func (s *grpcJobServer) GetJob(ctx context.Context, req *grpcapi.GetJobRequest) (*grpcapi.GetJobResponse, error) {
	return s.jobService.GetJob(ctx, req)
}

func (s *grpcJobServer) ListJobs(ctx context.Context, req *grpcapi.ListJobsRequest) (*grpcapi.ListJobsResponse, error) {
	return s.jobService.ListJobs(ctx, req)
}

func (s *grpcJobServer) UpdateJob(ctx context.Context, req *grpcapi.UpdateJobRequest) (*grpcapi.UpdateJobResponse, error) {
	return s.jobService.UpdateJob(ctx, req)
}

func (s *grpcJobServer) DeleteJob(ctx context.Context, req *grpcapi.DeleteJobRequest) (*grpcapi.DeleteJobResponse, error) {
	return s.jobService.DeleteJob(ctx, req)
}

func (s *grpcJobServer) TriggerJob(ctx context.Context, req *grpcapi.TriggerJobRequest) (*grpcapi.TriggerJobResponse, error) {
	return s.jobService.TriggerJob(ctx, req)
}

// SchedulerService gRPC 方法实现
func (s *grpcSchedulerServer) RegisterWorker(ctx context.Context, req *grpcapi.RegisterWorkerRequest) (*grpcapi.RegisterWorkerResponse, error) {
	return s.schedulerService.RegisterWorker(ctx, req)
}

func (s *grpcSchedulerServer) Heartbeat(ctx context.Context, req *grpcapi.HeartbeatRequest) (*grpcapi.HeartbeatResponse, error) {
	return s.schedulerService.Heartbeat(ctx, req)
}

func (s *grpcSchedulerServer) GetTask(ctx context.Context, req *grpcapi.GetTaskRequest) (*grpcapi.GetTaskResponse, error) {
	return s.schedulerService.GetTask(ctx, req)
}

func (s *grpcSchedulerServer) ReportTaskResult(ctx context.Context, req *grpcapi.ReportTaskResultRequest) (*grpcapi.ReportTaskResultResponse, error) {
	return s.schedulerService.ReportTaskResult(ctx, req)
}

type grpcAuthServer struct {
	grpcapi.UnimplementedAuthServiceServer
	authService *authservice.AuthService
}

// AuthService gRPC 方法实现
func (s *grpcAuthServer) Login(ctx context.Context, req *grpcapi.LoginRequest) (*grpcapi.LoginResponse, error) {
	return s.authService.Login(ctx, req)
}

func (s *grpcAuthServer) Logout(ctx context.Context, req *grpcapi.LogoutRequest) (*grpcapi.LogoutResponse, error) {
	return s.authService.Logout(ctx, req)
}

func (s *grpcAuthServer) RefreshToken(ctx context.Context, req *grpcapi.RefreshTokenRequest) (*grpcapi.RefreshTokenResponse, error) {
	return s.authService.RefreshToken(ctx, req)
}

func (s *grpcAuthServer) GetUserInfo(ctx context.Context, req *grpcapi.GetUserInfoRequest) (*grpcapi.GetUserInfoResponse, error) {
	return s.authService.GetUserInfo(ctx, req)
}

func (s *grpcAuthServer) GetUserPermissions(ctx context.Context, req *grpcapi.GetUserPermissionsRequest) (*grpcapi.GetUserPermissionsResponse, error) {
	return s.authService.GetUserPermissions(ctx, req)
}

type grpcUserServer struct {
	grpcapi.UnimplementedUserServiceServer
	userService *user.UserService
}

// UserService gRPC 方法实现
func (s *grpcUserServer) CreateUser(ctx context.Context, req *grpcapi.CreateUserRequest) (*grpcapi.CreateUserResponse, error) {
	return s.userService.CreateUser(ctx, req)
}

func (s *grpcUserServer) GetUser(ctx context.Context, req *grpcapi.GetUserRequest) (*grpcapi.GetUserResponse, error) {
	return s.userService.GetUser(ctx, req)
}

func (s *grpcUserServer) ListUsers(ctx context.Context, req *grpcapi.ListUsersRequest) (*grpcapi.ListUsersResponse, error) {
	return s.userService.ListUsers(ctx, req)
}

func (s *grpcUserServer) UpdateUser(ctx context.Context, req *grpcapi.UpdateUserRequest) (*grpcapi.UpdateUserResponse, error) {
	return s.userService.UpdateUser(ctx, req)
}

func (s *grpcUserServer) DeleteUser(ctx context.Context, req *grpcapi.DeleteUserRequest) (*grpcapi.DeleteUserResponse, error) {
	return s.userService.DeleteUser(ctx, req)
}

func (s *grpcUserServer) ChangePassword(ctx context.Context, req *grpcapi.ChangePasswordRequest) (*grpcapi.ChangePasswordResponse, error) {
	return s.userService.ChangePassword(ctx, req)
}

func (s *grpcUserServer) AssignUserRoles(ctx context.Context, req *grpcapi.AssignUserRolesRequest) (*grpcapi.AssignUserRolesResponse, error) {
	return s.userService.AssignUserRoles(ctx, req)
}

type grpcDepartmentServer struct {
	grpcapi.UnimplementedDepartmentServiceServer
	departmentService *department.DepartmentService
}

// DepartmentService gRPC 方法实现
func (s *grpcDepartmentServer) CreateDepartment(ctx context.Context, req *grpcapi.CreateDepartmentRequest) (*grpcapi.CreateDepartmentResponse, error) {
	return s.departmentService.CreateDepartment(ctx, req)
}

func (s *grpcDepartmentServer) GetDepartment(ctx context.Context, req *grpcapi.GetDepartmentRequest) (*grpcapi.GetDepartmentResponse, error) {
	return s.departmentService.GetDepartment(ctx, req)
}

func (s *grpcDepartmentServer) ListDepartments(ctx context.Context, req *grpcapi.ListDepartmentsRequest) (*grpcapi.ListDepartmentsResponse, error) {
	return s.departmentService.ListDepartments(ctx, req)
}

func (s *grpcDepartmentServer) UpdateDepartment(ctx context.Context, req *grpcapi.UpdateDepartmentRequest) (*grpcapi.UpdateDepartmentResponse, error) {
	return s.departmentService.UpdateDepartment(ctx, req)
}

func (s *grpcDepartmentServer) DeleteDepartment(ctx context.Context, req *grpcapi.DeleteDepartmentRequest) (*grpcapi.DeleteDepartmentResponse, error) {
	return s.departmentService.DeleteDepartment(ctx, req)
}

func (s *grpcDepartmentServer) GetDepartmentTree(ctx context.Context, req *grpcapi.GetDepartmentTreeRequest) (*grpcapi.GetDepartmentTreeResponse, error) {
	return s.departmentService.GetDepartmentTree(ctx, req)
}

type grpcRoleServer struct {
	grpcapi.UnimplementedRoleServiceServer
	roleService *role.RoleService
}

// RoleService gRPC 方法实现
func (s *grpcRoleServer) CreateRole(ctx context.Context, req *grpcapi.CreateRoleRequest) (*grpcapi.CreateRoleResponse, error) {
	return s.roleService.CreateRole(ctx, req)
}

func (s *grpcRoleServer) GetRole(ctx context.Context, req *grpcapi.GetRoleRequest) (*grpcapi.GetRoleResponse, error) {
	return s.roleService.GetRole(ctx, req)
}

func (s *grpcRoleServer) ListRoles(ctx context.Context, req *grpcapi.ListRolesRequest) (*grpcapi.ListRolesResponse, error) {
	return s.roleService.ListRoles(ctx, req)
}

func (s *grpcRoleServer) UpdateRole(ctx context.Context, req *grpcapi.UpdateRoleRequest) (*grpcapi.UpdateRoleResponse, error) {
	return s.roleService.UpdateRole(ctx, req)
}

func (s *grpcRoleServer) DeleteRole(ctx context.Context, req *grpcapi.DeleteRoleRequest) (*grpcapi.DeleteRoleResponse, error) {
	return s.roleService.DeleteRole(ctx, req)
}

func (s *grpcRoleServer) AssignPermissions(ctx context.Context, req *grpcapi.AssignPermissionsRequest) (*grpcapi.AssignPermissionsResponse, error) {
	return s.roleService.AssignPermissions(ctx, req)
}

type grpcPermissionServer struct {
	grpcapi.UnimplementedPermissionServiceServer
	permissionService *permission.PermissionService
}

// PermissionService gRPC 方法实现
func (s *grpcPermissionServer) CreatePermission(ctx context.Context, req *grpcapi.CreatePermissionRequest) (*grpcapi.CreatePermissionResponse, error) {
	return s.permissionService.CreatePermission(ctx, req)
}

func (s *grpcPermissionServer) GetPermission(ctx context.Context, req *grpcapi.GetPermissionRequest) (*grpcapi.GetPermissionResponse, error) {
	return s.permissionService.GetPermission(ctx, req)
}

func (s *grpcPermissionServer) ListPermissions(ctx context.Context, req *grpcapi.ListPermissionsRequest) (*grpcapi.ListPermissionsResponse, error) {
	return s.permissionService.ListPermissions(ctx, req)
}

func (s *grpcPermissionServer) UpdatePermission(ctx context.Context, req *grpcapi.UpdatePermissionRequest) (*grpcapi.UpdatePermissionResponse, error) {
	return s.permissionService.UpdatePermission(ctx, req)
}

func (s *grpcPermissionServer) DeletePermission(ctx context.Context, req *grpcapi.DeletePermissionRequest) (*grpcapi.DeletePermissionResponse, error) {
	return s.permissionService.DeletePermission(ctx, req)
}

func (s *grpcPermissionServer) GetPermissionTree(ctx context.Context, req *grpcapi.GetPermissionTreeRequest) (*grpcapi.GetPermissionTreeResponse, error) {
	return s.permissionService.GetPermissionTree(ctx, req)
}

type grpcMCPServer struct {
	grpcapi.UnimplementedMCPServiceServer
	mcpService *mcp.MCPService
}

// MCPService gRPC 方法实现
func (s *grpcMCPServer) ListTools(ctx context.Context, req *grpcapi.ListToolsRequest) (*grpcapi.ListToolsResponse, error) {
	return s.mcpService.ListTools(ctx, req)
}

func (s *grpcMCPServer) CallTool(ctx context.Context, req *grpcapi.CallToolRequest) (*grpcapi.CallToolResponse, error) {
	return s.mcpService.CallTool(ctx, req)
}

func (s *grpcMCPServer) GetResources(ctx context.Context, req *grpcapi.GetResourcesRequest) (*grpcapi.GetResourcesResponse, error) {
	return s.mcpService.GetResources(ctx, req)
}

type grpcAISchedulerServer struct {
	grpcapi.UnimplementedAISchedulerServiceServer
	aiScheduler *mcp.AISchedulerService
}

// AISchedulerService gRPC 方法实现
func (s *grpcAISchedulerServer) AnalyzeJob(ctx context.Context, req *grpcapi.AnalyzeJobRequest) (*grpcapi.AnalyzeJobResponse, error) {
	return s.aiScheduler.AnalyzeJob(ctx, req)
}

func (s *grpcAISchedulerServer) OptimizeSchedule(ctx context.Context, req *grpcapi.OptimizeScheduleRequest) (*grpcapi.OptimizeScheduleResponse, error) {
	return s.aiScheduler.OptimizeSchedule(ctx, req)
}

func (s *grpcAISchedulerServer) GetAIRecommendations(ctx context.Context, req *grpcapi.GetAIRecommendationsRequest) (*grpcapi.GetAIRecommendationsResponse, error) {
	return s.aiScheduler.GetAIRecommendations(ctx, req)
}
