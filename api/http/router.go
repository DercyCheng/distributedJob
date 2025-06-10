package http

import (
	grpc "go-job/api/grpc"
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
			// AI分析任务处理器
			ai.POST("/analyze-job", requirePermission("ai:analyze"), func(c *gin.Context) {
				var req struct {
					JobID   string `json:"job_id" binding:"required"`
					Context string `json:"context"`
				}

				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}

				if services.AIScheduler == nil {
					c.JSON(http.StatusServiceUnavailable, gin.H{"error": "AI service not available"})
					return
				}

				grpcReq := &grpc.AnalyzeJobRequest{
					JobId:   req.JobID,
					Context: req.Context,
				}

				resp, err := services.AIScheduler.AnalyzeJob(c.Request.Context(), grpcReq)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				c.JSON(http.StatusOK, gin.H{
					"message": "Job analysis completed",
					"data": gin.H{
						"analysis":        resp.Analysis,
						"strategy":        resp.Strategy,
						"priority":        resp.Priority,
						"recommendations": resp.Recommendations,
						"analyzed_at":     time.Now().Unix(),
					},
				})
			})

			// AI优化调度处理器
			ai.POST("/optimize-schedule", requirePermission("ai:optimize"), func(c *gin.Context) {
				var req struct {
					JobIDs      []string `json:"job_ids" binding:"required"`
					Constraints string   `json:"constraints"`
				}

				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}

				if services.AIScheduler == nil {
					c.JSON(http.StatusServiceUnavailable, gin.H{"error": "AI service not available"})
					return
				}
				grpcReq := &grpc.OptimizeScheduleRequest{
					JobIds:           req.JobIDs,
					OptimizationGoal: "performance", // 默认优化目标
					Constraints:      map[string]string{"custom": req.Constraints},
				}

				resp, err := services.AIScheduler.OptimizeSchedule(c.Request.Context(), grpcReq)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				c.JSON(http.StatusOK, gin.H{
					"message": "Schedule optimization completed",
					"data": gin.H{
						"optimizations": resp.Optimizations,
						"summary":       resp.Summary,
						"optimized_at":  time.Now().Unix(),
					},
				})
			}) // AI推荐处理器
			ai.GET("/recommendations", requirePermission("ai:read"), func(c *gin.Context) {
				if services.AIScheduler == nil {
					c.JSON(http.StatusServiceUnavailable, gin.H{"error": "AI service not available"})
					return
				}

				grpcReq := &grpc.GetAIRecommendationsRequest{
					Type: c.Query("type"), // performance, reliability, cost, general
					Context: map[string]string{
						"source": "http_api",
					},
				}

				// 添加额外的上下文参数
				if category := c.Query("category"); category != "" {
					grpcReq.Context["category"] = category
				}

				resp, err := services.AIScheduler.GetAIRecommendations(c.Request.Context(), grpcReq)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				c.JSON(http.StatusOK, gin.H{
					"message": "AI recommendations retrieved",
					"data": gin.H{
						"recommendations": resp.Recommendations,
						"total":           len(resp.Recommendations),
						"generated_at":    time.Now().Unix(),
					},
				})
			})
		}
		// MCP工具
		mcp := private.Group("/mcp")
		{
			// 获取MCP工具列表
			mcp.GET("/tools", requirePermission("mcp:read"), func(c *gin.Context) {
				if services.MCPService == nil {
					c.JSON(http.StatusServiceUnavailable, gin.H{"error": "MCP service not available"})
					return
				}

				grpcReq := &grpc.ListToolsRequest{
					Category: c.Query("category"), // 可选的分类过滤
				}

				resp, err := services.MCPService.ListTools(c.Request.Context(), grpcReq)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				c.JSON(http.StatusOK, gin.H{
					"message": "MCP tools retrieved",
					"data": gin.H{
						"tools": resp.Tools,
						"total": len(resp.Tools),
					},
				})
			})

			// 调用MCP工具
			mcp.POST("/tools/:name", requirePermission("mcp:execute"), func(c *gin.Context) {
				if services.MCPService == nil {
					c.JSON(http.StatusServiceUnavailable, gin.H{"error": "MCP service not available"})
					return
				}

				toolName := c.Param("name")

				var req struct {
					Parameters map[string]string `json:"parameters"`
				}

				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}
				grpcReq := &grpc.CallToolRequest{
					ToolName:  toolName,
					Arguments: req.Parameters,
				}

				resp, err := services.MCPService.CallTool(c.Request.Context(), grpcReq)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				c.JSON(http.StatusOK, gin.H{
					"message": "MCP tool executed successfully",
					"data": gin.H{
						"result":      resp.Result,
						"success":     resp.Success,
						"executed_at": time.Now().Unix(),
					},
				})
			})
			// 获取MCP资源
			mcp.GET("/resources", requirePermission("mcp:read"), func(c *gin.Context) {
				if services.MCPService == nil {
					c.JSON(http.StatusServiceUnavailable, gin.H{"error": "MCP service not available"})
					return
				}

				grpcReq := &grpc.GetResourcesRequest{
					Filter: c.Query("type"), // 可选的类型过滤
				}

				resp, err := services.MCPService.GetResources(c.Request.Context(), grpcReq)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}

				c.JSON(http.StatusOK, gin.H{
					"message": "MCP resources retrieved",
					"data": gin.H{
						"resources": resp.Resources,
						"total":     len(resp.Resources),
					},
				})
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
