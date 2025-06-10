package http

import (
	"go-job/internal/auth"
	"go-job/internal/department"
	"go-job/internal/job"
	"go-job/internal/mcp"
	"go-job/internal/permission"
	"go-job/internal/role"
	"go-job/internal/user"
	"go-job/pkg/config"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Services 服务容器
type Services struct {
	AuthService       *auth.AuthService
	UserService       *user.UserService
	DepartmentService *department.DepartmentService
	RoleService       *role.RoleService
	PermissionService *permission.PermissionService
	JobService        *job.Service
	AIScheduler       *mcp.AISchedulerService
	MCPService        *mcp.MCPService
}

// NewRouter 创建路由
func NewRouter(cfg *config.Config, services *Services) *gin.Engine {
	router := gin.Default()

	// 中间件
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(corsMiddleware())

	// 健康检查
	router.GET("/health", healthCheck)

	// 公共API（无需认证）
	public := router.Group("/api/v1")
	{
		auth := public.Group("/auth")
		{
			auth.POST("/login", loginHandler)
			auth.POST("/refresh", refreshTokenHandler)
		}
	}

	// 私有API（需要认证）
	private := router.Group("/api/v1")
	private.Use(authMiddleware())
	{
		// 认证相关
		auth := private.Group("/auth")
		{
			auth.POST("/logout", logoutHandler)
			auth.GET("/profile", getProfileHandler)
			auth.PUT("/profile", updateProfileHandler)
			auth.POST("/change-password", changePasswordHandler)
		}

		// 部门管理
		departments := private.Group("/departments")
		{
			departmentHandler := NewDepartmentHandler(services.DepartmentService)
			departments.POST("", requirePermission("department:create"), departmentHandler.CreateDepartment)
			departments.GET("", requirePermission("department:read"), departmentHandler.ListDepartments)
			departments.GET("/tree", requirePermission("department:read"), departmentHandler.GetDepartmentTree)
			departments.GET("/:id", requirePermission("department:read"), departmentHandler.GetDepartment)
			departments.PUT("/:id", requirePermission("department:update"), departmentHandler.UpdateDepartment)
			departments.DELETE("/:id", requirePermission("department:delete"), departmentHandler.DeleteDepartment)
		}

		// 用户管理
		users := private.Group("/users")
		{
			userHandler := NewUserHandler(services.UserService)
			users.POST("", requirePermission("user:create"), userHandler.CreateUser)
			users.GET("", requirePermission("user:read"), userHandler.ListUsers)
			users.GET("/:id", requirePermission("user:read"), userHandler.GetUser)
			users.PUT("/:id", requirePermission("user:update"), userHandler.UpdateUser)
			users.DELETE("/:id", requirePermission("user:delete"), userHandler.DeleteUser)
			users.POST("/:id/roles", requirePermission("user:assign_role"), userHandler.AssignUserRoles)
		}

		// 角色管理
		roles := private.Group("/roles")
		{
			roleHandler := NewRoleHandler(services.RoleService)
			roles.POST("", requirePermission("role:create"), roleHandler.CreateRole)
			roles.GET("", requirePermission("role:read"), roleHandler.ListRoles)
			roles.GET("/:id", requirePermission("role:read"), roleHandler.GetRole)
			roles.PUT("/:id", requirePermission("role:update"), roleHandler.UpdateRole)
			roles.DELETE("/:id", requirePermission("role:delete"), roleHandler.DeleteRole)
			roles.POST("/:id/permissions", requirePermission("role:assign_permission"), roleHandler.AssignRolePermissions)
		}

		// 权限管理
		permissions := private.Group("/permissions")
		{
			permissionHandler := NewPermissionHandler(services.PermissionService)
			permissions.POST("", requirePermission("permission:create"), permissionHandler.CreatePermission)
			permissions.GET("", requirePermission("permission:read"), permissionHandler.ListPermissions)
			permissions.GET("/tree", requirePermission("permission:read"), permissionHandler.GetPermissionTree)
			permissions.GET("/:id", requirePermission("permission:read"), permissionHandler.GetPermission)
			permissions.PUT("/:id", requirePermission("permission:update"), permissionHandler.UpdatePermission)
			permissions.DELETE("/:id", requirePermission("permission:delete"), permissionHandler.DeletePermission)
		}

		// 任务管理
		jobs := private.Group("/jobs")
		{
			jobHandler := NewJobHandler()
			jobs.POST("", requirePermission("job:create"), jobHandler.CreateJob)
			jobs.GET("", requirePermission("job:read"), jobHandler.ListJobs)
			jobs.GET("/:id", requirePermission("job:read"), jobHandler.GetJob)
			jobs.PUT("/:id", requirePermission("job:update"), jobHandler.UpdateJob)
			jobs.DELETE("/:id", requirePermission("job:delete"), jobHandler.DeleteJob)
			jobs.POST("/:id/trigger", requirePermission("job:execute"), jobHandler.TriggerJob)
			jobs.GET("/:id/executions", requirePermission("job:read"), jobHandler.GetJobExecutions)
		}

		// 执行记录
		executions := private.Group("/executions")
		{
			executionHandler := NewExecutionHandler()
			executions.GET("", requirePermission("execution:read"), executionHandler.ListExecutions)
			executions.GET("/:id", requirePermission("execution:read"), executionHandler.GetExecution)
			executions.POST("/:id/cancel", requirePermission("execution:cancel"), executionHandler.CancelExecution)
		}

		// 工作节点管理
		workers := private.Group("/workers")
		{
			workerHandler := NewWorkerHandler()
			workers.GET("", requirePermission("worker:read"), workerHandler.ListWorkers)
			workers.GET("/:id", requirePermission("worker:read"), workerHandler.GetWorker)
			workers.PUT("/:id/status", requirePermission("worker:update"), workerHandler.UpdateWorkerStatus)
		}

		// 统计信息
		stats := private.Group("/stats")
		{
			statsHandler := NewStatsHandler()
			stats.GET("/dashboard", requirePermission("stats:read"), statsHandler.GetDashboard)
			stats.GET("/jobs", requirePermission("stats:read"), statsHandler.GetJobStats)
			stats.GET("/workers", requirePermission("stats:read"), statsHandler.GetWorkerStats)
			stats.GET("/executions", requirePermission("stats:read"), statsHandler.GetExecutionStats)
		}

		// AI调度
		ai := private.Group("/ai")
		{
			// TODO: Implement AI handlers
			ai.POST("/analyze-job", requirePermission("ai:analyze"), func(c *gin.Context) {
				c.JSON(200, gin.H{"message": "AI analyze job endpoint"})
			})
			ai.POST("/optimize-schedule", requirePermission("ai:optimize"), func(c *gin.Context) {
				c.JSON(200, gin.H{"message": "AI optimize schedule endpoint"})
			})
			ai.GET("/recommendations", requirePermission("ai:read"), func(c *gin.Context) {
				c.JSON(200, gin.H{"message": "AI recommendations endpoint"})
			})
		}

		// MCP工具
		mcp := private.Group("/mcp")
		{
			// TODO: Implement MCP handlers
			mcp.GET("/tools", requirePermission("mcp:read"), func(c *gin.Context) {
				c.JSON(200, gin.H{"message": "MCP tools endpoint"})
			})
			mcp.POST("/tools/:name", requirePermission("mcp:execute"), func(c *gin.Context) {
				c.JSON(200, gin.H{"message": "MCP call tool endpoint"})
			})
			mcp.GET("/resources", requirePermission("mcp:read"), func(c *gin.Context) {
				c.JSON(200, gin.H{"message": "MCP resources endpoint"})
			})
		}
	}

	return router
}

// healthCheck 健康检查
func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"time":   time.Now().Unix(),
	})
}

// corsMiddleware CORS 中间件
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
