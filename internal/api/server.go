package api

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"distributedJob/internal/api/middleware"
	"distributedJobonfig"
	"distributedJobob"
	"distributedJobodel/entity"
	"distributedJobervice"
	"distributedJobtore"
	"github.com/gin-gonic/gin"
)

// Server 表示API服务器
type Server struct {
	config       *config.Config
	router       *gin.Engine
	scheduler    *job.Scheduler
	taskService  service.TaskService
	authService  service.AuthService
	tokenRevoker store.TokenRevoker
}

// ResponseBody API响应体
type ResponseBody struct {
	Code       int         `json:"code"`
	Message    string      `json:"message"`
	Data       interface{} `json:"data"`
	ErrorField string      `json:"errorField,omitempty"` // 用于标识哪个字段出错
	ErrorType  string      `json:"errorType,omitempty"`  // 错误类型
	RequestId  string      `json:"requestId,omitempty"`  // 请求ID，用于日志跟踪
}

// ApiError represents structured API error
type ApiError struct {
	Code       int    `json:"code"`
	Message    string `json:"message"`
	ErrorField string `json:"errorField,omitempty"`
	ErrorType  string `json:"errorType,omitempty"`
	Details    string `json:"details,omitempty"`
}

// ErrorDetails 定义错误详情常量
var ErrorDetails = struct {
	FieldRequired         string
	FieldInvalidFormat    string
	FieldInvalidValue     string
	FieldTooLong          string
	FieldTooShort         string
	ResourceNotFound      string
	ResourceAlreadyExists string
	AuthFailed            string
	PermissionDenied      string
	SystemError           string
	DatabaseError         string
	NetworkError          string
	ValidationError       string
}{
	FieldRequired:         "FIELD_REQUIRED",
	FieldInvalidFormat:    "FIELD_INVALID_FORMAT",
	FieldInvalidValue:     "FIELD_INVALID_VALUE",
	FieldTooLong:          "FIELD_TOO_LONG",
	FieldTooShort:         "FIELD_TOO_SHORT",
	ResourceNotFound:      "RESOURCE_NOT_FOUND",
	ResourceAlreadyExists: "RESOURCE_ALREADY_EXISTS",
	AuthFailed:            "AUTH_FAILED",
	PermissionDenied:      "PERMISSION_DENIED",
	SystemError:           "SYSTEM_ERROR",
	DatabaseError:         "DATABASE_ERROR",
	NetworkError:          "NETWORK_ERROR",
	ValidationError:       "VALIDATION_ERROR",
}

// ErrorCodes defines standard API error codes
var ErrorCodes = struct {
	// Success
	Success int

	// Client errors (4000-4999)
	BadRequest            int
	Unauthorized          int
	Forbidden             int
	NotFound              int
	MethodNotAllowed      int
	Conflict              int
	TooManyRequests       int
	ValidationFailed      int
	ResourceAlreadyExists int

	// Server errors (5000-5999)
	InternalServerError  int
	ServiceUnavailable   int
	DatabaseError        int
	ExternalServiceError int

	// Business logic errors (6000-6999)
	TaskExecutionFailed  int
	ScheduleError        int
	AuthenticationFailed int
}{
	// Success
	Success: 0,

	// Client errors
	BadRequest:            4000,
	Unauthorized:          4001,
	Forbidden:             4003,
	NotFound:              4004,
	MethodNotAllowed:      4005,
	Conflict:              4009,
	TooManyRequests:       4029,
	ValidationFailed:      4400,
	ResourceAlreadyExists: 4409,

	// Server errors
	InternalServerError:  5000,
	ServiceUnavailable:   5003,
	DatabaseError:        5100,
	ExternalServiceError: 5400,

	// Business logic errors
	TaskExecutionFailed:  6000,
	ScheduleError:        6100,
	AuthenticationFailed: 6200,
}

// NewServer 创建一个新的API服务器
func NewServer(
	config *config.Config,
	scheduler *job.Scheduler,
	repoManager store.RepositoryManager,
	authService service.AuthService,
	tokenRevoker store.TokenRevoker,
) *Server {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	// 添加中间件
	router.Use(middleware.Logger())
	router.Use(middleware.CORS()) // 添加CORS中间件
	router.Use(gin.Recovery())

	// 创建服务
	taskService := service.NewTaskService(repoManager.Task(), scheduler)

	s := &Server{
		config:       config,
		router:       router,
		scheduler:    scheduler,
		taskService:  taskService,
		authService:  authService,
		tokenRevoker: tokenRevoker,
	}

	s.setupRoutes()
	return s
}

// Router 返回服务器的HTTP处理器
func (s *Server) Router() http.Handler {
	return s.router
}

// setupRoutes 设置API路由
func (s *Server) setupRoutes() {
	// API基础路径
	base := s.router.Group(s.config.Server.ContextPath)

	// 健康检查不需要验证
	base.GET("/health", s.healthCheck)

	// 服务关闭API仅限本地访问
	base.GET("/shutdown", s.localOnly(), s.shutdown)

	// 认证API不需要JWT验证
	authGroup := base.Group("/auth")
	{
		authGroup.POST("/login", s.login)
		// 使用专门的刷新令牌中间件
		authGroup.POST("/refresh", middleware.RefreshAuth(s.config, s.tokenRevoker), s.refreshToken)
		authGroup.POST("/logout", s.logout)
	}

	// 所有其他API需要JWT验证
	auth := base.Group("")
	auth.Use(middleware.JWTAuth(s.config, s.tokenRevoker))

	// 用户信息API
	auth.GET("/auth/userinfo", s.getUserInfo)

	// 部门管理API
	deptGroup := auth.Group("/departments")
	{
		deptGroup.GET("", s.getDepartmentList)
		deptGroup.GET("/:id", s.getDepartment)
		deptGroup.POST("", s.createDepartment)
		deptGroup.PUT("/:id", s.updateDepartment)
		deptGroup.DELETE("/:id", s.deleteDepartment)
	}

	// 用户管理API
	userGroup := auth.Group("/users")
	{
		userGroup.GET("", s.getUserList)
		userGroup.GET("/:id", s.getUser)
		userGroup.POST("", s.createUser)
		userGroup.PUT("/:id", s.updateUser)
		userGroup.DELETE("/:id", s.deleteUser)
		userGroup.PATCH("/:id/password", s.updateUserPassword)
	}

	// 角色和权限管理API
	roleGroup := auth.Group("/roles")
	{
		roleGroup.GET("", s.getRoleList)
		roleGroup.GET("/:id", s.getRole)
		roleGroup.POST("", s.createRole)
		roleGroup.PUT("/:id", s.updateRole)
		roleGroup.DELETE("/:id", s.deleteRole)
	}

	auth.GET("/permissions", s.getPermissionList)

	// 任务管理API
	taskGroup := auth.Group("/tasks")
	{
		taskGroup.GET("", s.getTaskList)
		taskGroup.GET("/:id", s.getTask)
		taskGroup.POST("/http", s.createHTTPTask)
		taskGroup.POST("/grpc", s.createGRPCTask)
		taskGroup.PUT("/http/:id", s.updateHTTPTask)
		taskGroup.PUT("/grpc/:id", s.updateGRPCTask)
		taskGroup.DELETE("/:id", s.deleteTask)
		taskGroup.PATCH("/:id/status", s.updateTaskStatus)
	}

	// 执行记录查询API
	recordGroup := auth.Group("/records")
	{
		recordGroup.GET("", s.getRecordList)
		recordGroup.GET("/:id", s.getRecord)
		recordGroup.GET("/stats", s.getRecordStats)
	}

	// 为所有API路径添加全局OPTIONS请求处理
	base.OPTIONS("/*path", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
}

// 健康检查API
func (s *Server) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, ResponseBody{
		Code:    0,
		Message: "success",
		Data: map[string]interface{}{
			"status":    "up",
			"timestamp": fmt.Sprintf("%v", c.MustGet("requestTime")),
		},
	})
}

// 服务关闭API
func (s *Server) shutdown(c *gin.Context) {
	// 实现服务平滑关闭
	c.JSON(http.StatusOK, ResponseBody{
		Code:    0,
		Message: "service will shutdown",
		Data:    nil,
	})
}

// 用户登录API
func (s *Server) login(c *gin.Context) {
	// 解析请求参数
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		// 详细解析验证错误信息
		validationErrors := []map[string]string{}

		if strings.Contains(err.Error(), "username") {
			validationErrors = append(validationErrors, map[string]string{
				"field": "username",
				"error": "用户名不能为空",
			})
		}

		if strings.Contains(err.Error(), "password") {
			validationErrors = append(validationErrors, map[string]string{
				"field": "password",
				"error": "密码不能为空",
			})
		}

		c.JSON(http.StatusBadRequest, ResponseBody{
			Code:      ErrorCodes.ValidationFailed,
			Message:   "请求参数验证失败",
			Data:      validationErrors,
			ErrorType: ErrorDetails.ValidationError,
			RequestId: c.GetString("requestId"),
		})
		return
	}

	// 参数额外验证
	if len(req.Username) < 3 {
		c.JSON(http.StatusBadRequest, ResponseBody{
			Code:       ErrorCodes.ValidationFailed,
			Message:    "用户名长度不能少于3个字符",
			ErrorField: "username",
			ErrorType:  ErrorDetails.FieldTooShort,
			RequestId:  c.GetString("requestId"),
		})
		return
	}

	if len(req.Password) < 6 {
		c.JSON(http.StatusBadRequest, ResponseBody{
			Code:       ErrorCodes.ValidationFailed,
			Message:    "密码长度不能少于6个字符",
			ErrorField: "password",
			ErrorType:  ErrorDetails.FieldTooShort,
			RequestId:  c.GetString("requestId"),
		})
		return
	}

	// 调用认证服务登录
	accessToken, refreshToken, user, err := s.authService.Login(req.Username, req.Password)
	if err != nil {
		errorMsg := err.Error()
		errorType := ErrorDetails.AuthFailed

		// 根据具体错误类型提供更友好的错误消息
		if strings.Contains(errorMsg, "not found") {
			errorMsg = "用户不存在"
			errorType = ErrorDetails.ResourceNotFound
			errorField := "username"
			c.JSON(http.StatusUnauthorized, ResponseBody{
				Code:       ErrorCodes.Unauthorized,
				Message:    errorMsg,
				Data:       nil,
				ErrorType:  errorType,
				ErrorField: errorField,
				RequestId:  c.GetString("requestId"),
			})
			return
		} else if strings.Contains(errorMsg, "password") {
			errorMsg = "密码错误"
			errorField := "password"
			c.JSON(http.StatusUnauthorized, ResponseBody{
				Code:       ErrorCodes.Unauthorized,
				Message:    errorMsg,
				Data:       nil,
				ErrorType:  errorType,
				ErrorField: errorField,
				RequestId:  c.GetString("requestId"),
			})
			return
		}

		c.JSON(http.StatusUnauthorized, ResponseBody{
			Code:      ErrorCodes.Unauthorized,
			Message:   errorMsg,
			Data:      nil,
			ErrorType: errorType,
			RequestId: c.GetString("requestId"),
		})
		return
	}

	c.JSON(http.StatusOK, ResponseBody{
		Code:      ErrorCodes.Success,
		Message:   "登录成功",
		RequestId: c.GetString("requestId"),
		Data: map[string]interface{}{
			"accessToken":  accessToken,
			"refreshToken": refreshToken,
			"userId":       user.ID,
			"username":     user.Username,
			"realName":     user.RealName,
			"departmentId": user.DepartmentID,
			"roleId":       user.RoleID,
			"tokenType":    "Bearer",
			"expiresIn":    s.config.Auth.JwtExpireMinutes * 60, // 过期时间(秒)
		},
	})
}

// 刷新Token API
func (s *Server) refreshToken(c *gin.Context) {
	// 从上下文获取用户ID（已通过中间件验证刷新令牌）
	userID, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, ResponseBody{
			Code:    ErrorCodes.Unauthorized,
			Message: "Invalid refresh token",
			Data:    nil,
		})
		return
	}

	// 生成新的令牌对
	accessToken, refreshToken, err := s.authService.GenerateTokens(&entity.User{
		ID: userID.(int64),
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, ResponseBody{
			Code:    ErrorCodes.InternalServerError,
			Message: "Failed to generate tokens: " + err.Error(),
			Data:    nil,
		})
		return
	}

	c.JSON(http.StatusOK, ResponseBody{
		Code:    ErrorCodes.Success,
		Message: "Token refreshed successfully",
		Data: map[string]interface{}{
			"accessToken":  accessToken,
			"refreshToken": refreshToken,
			"tokenType":    "Bearer",
			"expiresIn":    s.config.Auth.JwtExpireMinutes * 60, // 过期时间(秒)
		},
	})
}

// 登出API
func (s *Server) logout(c *gin.Context) {
	// 解析请求参数
	var req struct {
		AccessToken  string `json:"accessToken"`
		RefreshToken string `json:"refreshToken"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, ResponseBody{
			Code:    ErrorCodes.Success,
			Message: "Logged out successfully",
			Data:    nil,
		})
		return
	}

	// 撤销访问令牌和刷新令牌
	if req.AccessToken != "" {
		_ = s.authService.RevokeToken(req.AccessToken)
	}

	if req.RefreshToken != "" {
		_ = s.authService.RevokeToken(req.RefreshToken)
	}

	c.JSON(http.StatusOK, ResponseBody{
		Code:    ErrorCodes.Success,
		Message: "Logged out successfully",
		Data:    nil,
	})
}

// 获取用户信息API
func (s *Server) getUserInfo(c *gin.Context) {
	// 从上下文获取用户ID
	userID, _ := c.Get("userId")

	// 调用认证服务获取用户信息
	user, err := s.authService.GetUserByID(userID.(int64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, ResponseBody{
			Code:    ErrorCodes.InternalServerError,
			Message: err.Error(),
			Data:    nil,
		})
		return
	}

	// 获取用户权限
	// AuthService没有GetUserPermissions方法，使用HasPermission检查权限
	// 这里可以获取所有权限，然后检查用户是否拥有每个权限
	permissions, err := s.authService.GetPermissionList()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ResponseBody{
			Code:    ErrorCodes.InternalServerError,
			Message: "Failed to get permissions",
			Data:    nil,
		})
		return
	}

	// 创建用户权限列表
	userPermissions := []string{}
	for _, p := range permissions {
		hasPermission, _ := s.authService.HasPermission(userID.(int64), p.Code)
		if hasPermission {
			userPermissions = append(userPermissions, p.Code)
		}
	}

	c.JSON(http.StatusOK, ResponseBody{
		Code:    ErrorCodes.Success,
		Message: "success",
		Data: map[string]interface{}{
			"user":        user,
			"permissions": userPermissions,
		},
	})
}

// 获取任务列表API
func (s *Server) getTaskList(c *gin.Context) {
	// 解析请求参数
	departmentID, err := strconv.ParseInt(c.DefaultQuery("departmentId", "0"), 10, 64)
	if err != nil {
		departmentID = 0
	}

	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	size, err := strconv.Atoi(c.DefaultQuery("size", "10"))
	if err != nil || size < 1 || size > 100 {
		size = 10
	}

	// 调用任务服务获取任务列表
	tasks, total, err := s.taskService.GetTaskList(departmentID, page, size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ResponseBody{
			Code:    ErrorCodes.InternalServerError,
			Message: err.Error(),
			Data:    nil,
		})
		return
	}

	c.JSON(http.StatusOK, ResponseBody{
		Code:    ErrorCodes.Success,
		Message: "success",
		Data: map[string]interface{}{
			"tasks": tasks,
			"total": total,
		},
	})
}

// 获取任务详情API
func (s *Server) getTask(c *gin.Context) {
	// 解析任务ID
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, ResponseBody{
			Code:    ErrorCodes.BadRequest,
			Message: "invalid task id",
			Data:    nil,
		})
		return
	}

	// 调用任务服务获取任务详情
	task, err := s.taskService.GetTaskByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, ResponseBody{
			Code:    ErrorCodes.NotFound,
			Message: err.Error(),
			Data:    nil,
		})
		return
	}

	c.JSON(http.StatusOK, ResponseBody{
		Code:    ErrorCodes.Success,
		Message: "success",
		Data:    task,
	})
}

// 创建HTTP任务API
func (s *Server) createHTTPTask(c *gin.Context) {
	// 解析请求参数
	var task entity.Task
	if err := c.ShouldBindJSON(&task); err != nil {
		// 详细解析验证错误信息
		validationErrors := s.parseValidationError(err)
		c.JSON(http.StatusBadRequest, ResponseBody{
			Code:      ErrorCodes.ValidationFailed,
			Message:   "任务参数验证失败",
			Data:      validationErrors,
			ErrorType: ErrorDetails.ValidationError,
			RequestId: c.GetString("requestId"),
		})
		return
	}

	// 进行额外的业务逻辑验证
	if task.Name == "" {
		c.JSON(http.StatusBadRequest, ResponseBody{
			Code:       ErrorCodes.ValidationFailed,
			Message:    "任务名称不能为空",
			ErrorField: "name",
			ErrorType:  ErrorDetails.FieldRequired,
			RequestId:  c.GetString("requestId"),
		})
		return
	}

	if task.Cron == "" {
		c.JSON(http.StatusBadRequest, ResponseBody{
			Code:       ErrorCodes.ValidationFailed,
			Message:    "CRON 表达式不能为空",
			ErrorField: "cron",
			ErrorType:  ErrorDetails.FieldRequired,
			RequestId:  c.GetString("requestId"),
		})
		return
	}

	// 验证 HTTP 任务特有字段
	if task.TaskType == "HTTP" {
		if task.URL == "" {
			c.JSON(http.StatusBadRequest, ResponseBody{
				Code:       ErrorCodes.ValidationFailed,
				Message:    "HTTP URL 不能为空",
				ErrorField: "url",
				ErrorType:  ErrorDetails.FieldRequired,
				RequestId:  c.GetString("requestId"),
			})
			return
		}
	}

	// 调用任务服务创建HTTP任务
	id, err := s.taskService.CreateHTTPTask(&task)
	if err != nil {
		// 根据不同错误类型，提供不同的响应
		if strings.Contains(err.Error(), "already exists") {
			c.JSON(http.StatusConflict, ResponseBody{
				Code:       ErrorCodes.ResourceAlreadyExists,
				Message:    "任务名称已存在",
				ErrorField: "name",
				ErrorType:  ErrorDetails.ResourceAlreadyExists,
				RequestId:  c.GetString("requestId"),
			})
			return
		}

		if strings.Contains(err.Error(), "invalid cron") {
			c.JSON(http.StatusBadRequest, ResponseBody{
				Code:       ErrorCodes.ValidationFailed,
				Message:    "CRON 表达式格式不正确",
				ErrorField: "cron",
				ErrorType:  ErrorDetails.FieldInvalidFormat,
				RequestId:  c.GetString("requestId"),
			})
			return
		}

		if strings.Contains(err.Error(), "department") {
			c.JSON(http.StatusBadRequest, ResponseBody{
				Code:       ErrorCodes.ValidationFailed,
				Message:    "部门不存在或无效",
				ErrorField: "departmentId",
				ErrorType:  ErrorDetails.FieldInvalidValue,
			})
			return
		}

		c.JSON(http.StatusInternalServerError, ResponseBody{
			Code:    ErrorCodes.InternalServerError,
			Message: err.Error(),
			Data:    nil,
		})
		return
	}

	c.JSON(http.StatusOK, ResponseBody{
		Code:    ErrorCodes.Success,
		Message: "success",
		Data: map[string]interface{}{
			"id": id,
		},
	})
}

// 创建GRPC任务API
func (s *Server) createGRPCTask(c *gin.Context) {
	// 解析请求参数
	var task entity.Task
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, ResponseBody{
			Code:    ErrorCodes.BadRequest,
			Message: "invalid request parameters",
			Data:    nil,
		})
		return
	}

	// 验证 gRPC 任务特有字段
	if task.TaskType == "GRPC" {
		if task.GrpcService == "" {
			c.JSON(http.StatusBadRequest, ResponseBody{
				Code:       ErrorCodes.ValidationFailed,
				Message:    "gRPC 服务名称不能为空",
				ErrorField: "grpcService",
				ErrorType:  ErrorDetails.FieldRequired,
				RequestId:  c.GetString("requestId"),
			})
			return
		}

		if task.GrpcMethod == "" {
			c.JSON(http.StatusBadRequest, ResponseBody{
				Code:       ErrorCodes.ValidationFailed,
				Message:    "gRPC 方法名称不能为空",
				ErrorField: "grpcMethod",
				ErrorType:  ErrorDetails.FieldRequired,
				RequestId:  c.GetString("requestId"),
			})
			return
		}
	}

	// 调用任务服务创建GRPC任务
	id, err := s.taskService.CreateGRPCTask(&task)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ResponseBody{
			Code:    ErrorCodes.InternalServerError,
			Message: err.Error(),
			Data:    nil,
		})
		return
	}

	c.JSON(http.StatusOK, ResponseBody{
		Code:    ErrorCodes.Success,
		Message: "success",
		Data: map[string]interface{}{
			"id": id,
		},
	})
}

// 更新HTTP任务API
func (s *Server) updateHTTPTask(c *gin.Context) {
	// 解析任务ID
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, ResponseBody{
			Code:    ErrorCodes.BadRequest,
			Message: "invalid task id",
			Data:    nil,
		})
		return
	}

	// 解析请求参数
	var task entity.Task
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, ResponseBody{
			Code:    ErrorCodes.BadRequest,
			Message: "invalid request parameters",
			Data:    nil,
		})
		return
	}

	// 设置任务ID
	task.ID = id

	// 调用任务服务更新HTTP任务
	err = s.taskService.UpdateHTTPTask(&task)
	if err != nil {
		status := http.StatusInternalServerError
		code := ErrorCodes.InternalServerError

		if err == service.ErrTaskNotFound {
			status = http.StatusNotFound
			code = ErrorCodes.NotFound
		} else if err == service.ErrInvalidParameters {
			status = http.StatusBadRequest
			code = ErrorCodes.BadRequest
		}

		c.JSON(status, ResponseBody{
			Code:    code,
			Message: err.Error(),
			Data:    nil,
		})
		return
	}

	c.JSON(http.StatusOK, ResponseBody{
		Code:    ErrorCodes.Success,
		Message: "success",
		Data:    nil,
	})
}

// 更新GRPC任务API
func (s *Server) updateGRPCTask(c *gin.Context) {
	// 解析任务ID
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, ResponseBody{
			Code:       ErrorCodes.BadRequest,
			Message:    "任务ID格式不正确",
			ErrorField: "id",
			ErrorType:  ErrorDetails.FieldInvalidFormat,
			RequestId:  c.GetString("requestId"),
		})
		return
	}

	// 解析请求参数
	var task entity.Task
	if err := c.ShouldBindJSON(&task); err != nil {
		// 详细解析验证错误信息
		validationErrors := s.parseValidationError(err)
		c.JSON(http.StatusBadRequest, ResponseBody{
			Code:      ErrorCodes.ValidationFailed,
			Message:   "任务参数验证失败",
			Data:      validationErrors,
			ErrorType: ErrorDetails.ValidationError,
			RequestId: c.GetString("requestId"),
		})
		return
	}

	// 设置任务ID
	task.ID = id

	// 验证 gRPC 任务特有字段
	if task.TaskType == "GRPC" {
		if task.GrpcService == "" {
			c.JSON(http.StatusBadRequest, ResponseBody{
				Code:       ErrorCodes.ValidationFailed,
				Message:    "gRPC 服务名称不能为空",
				ErrorField: "grpcService",
				ErrorType:  ErrorDetails.FieldRequired,
				RequestId:  c.GetString("requestId"),
			})
			return
		}

		if task.GrpcMethod == "" {
			c.JSON(http.StatusBadRequest, ResponseBody{
				Code:       ErrorCodes.ValidationFailed,
				Message:    "gRPC 方法名称不能为空",
				ErrorField: "grpcMethod",
				ErrorType:  ErrorDetails.FieldRequired,
				RequestId:  c.GetString("requestId"),
			})
			return
		}
	}

	// 调用任务服务更新GRPCTask
	err = s.taskService.UpdateGRPCTask(&task)
	if err != nil {
		status := http.StatusInternalServerError
		code := ErrorCodes.InternalServerError
		errorType := ErrorDetails.SystemError
		errorField := ""

		if err == service.ErrTaskNotFound {
			status = http.StatusNotFound
			code = ErrorCodes.NotFound
			errorType = ErrorDetails.ResourceNotFound
			c.JSON(status, ResponseBody{
				Code:      code,
				Message:   "任务不存在",
				ErrorType: errorType,
				RequestId: c.GetString("requestId"),
			})
			return
		} else if err == service.ErrInvalidParameters {
			status = http.StatusBadRequest
			code = ErrorCodes.BadRequest
			errorType = ErrorDetails.ValidationError
			c.JSON(status, ResponseBody{
				Code:      code,
				Message:   "参数无效，请检查输入",
				ErrorType: errorType,
				RequestId: c.GetString("requestId"),
			})
			return
		}

		// 根据错误消息进一步细分错误
		errorMsg := err.Error()
		if strings.Contains(errorMsg, "already exists") {
			status = http.StatusConflict
			code = ErrorCodes.ResourceAlreadyExists
			errorType = ErrorDetails.ResourceAlreadyExists
			errorField = "name"
			c.JSON(status, ResponseBody{
				Code:       code,
				Message:    "任务名称已存在",
				ErrorField: errorField,
				ErrorType:  errorType,
				RequestId:  c.GetString("requestId"),
			})
			return
		}

		// 其他错误的通用处理
		c.JSON(status, ResponseBody{
			Code:      code,
			Message:   "更新任务失败: " + errorMsg,
			ErrorType: errorType,
			RequestId: c.GetString("requestId"),
		})
		return
	}

	c.JSON(http.StatusOK, ResponseBody{
		Code:      ErrorCodes.Success,
		Message:   "任务更新成功",
		RequestId: c.GetString("requestId"),
	})
}

// 删除任务API
func (s *Server) deleteTask(c *gin.Context) {
	// 解析任务ID
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, ResponseBody{
			Code:    ErrorCodes.BadRequest,
			Message: "invalid task id",
			Data:    nil,
		})
		return
	}

	// 调用任务服务删除任务
	err = s.taskService.DeleteTask(id)
	if err != nil {
		status := http.StatusInternalServerError
		code := ErrorCodes.InternalServerError

		if err == service.ErrTaskNotFound {
			status = http.StatusNotFound
			code = ErrorCodes.NotFound
		}

		c.JSON(status, ResponseBody{
			Code:    code,
			Message: err.Error(),
			Data:    nil,
		})
		return
	}

	c.JSON(http.StatusOK, ResponseBody{
		Code:    ErrorCodes.Success,
		Message: "success",
		Data:    nil,
	})
}

// 更新任务状态API
func (s *Server) updateTaskStatus(c *gin.Context) {
	// 解析任务ID
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, ResponseBody{
			Code:    ErrorCodes.BadRequest,
			Message: "invalid task id",
			Data:    nil,
		})
		return
	}

	// 解析请求参数
	var req struct {
		Status int8 `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ResponseBody{
			Code:    ErrorCodes.BadRequest,
			Message: "invalid request parameters",
			Data:    nil,
		})
		return
	}

	// 调用任务服务更新任务状态
	err = s.taskService.UpdateTaskStatus(id, req.Status)
	if err != nil {
		status := http.StatusInternalServerError
		code := ErrorCodes.InternalServerError

		if err == service.ErrTaskNotFound {
			status = http.StatusNotFound
			code = ErrorCodes.NotFound
		} else if err == service.ErrInvalidParameters {
			status = http.StatusBadRequest
			code = ErrorCodes.BadRequest
		}

		c.JSON(status, ResponseBody{
			Code:    code,
			Message: err.Error(),
			Data:    nil,
		})
		return
	}

	c.JSON(http.StatusOK, ResponseBody{
		Code:    ErrorCodes.Success,
		Message: "success",
		Data:    nil,
	})
}

// 获取执行记录列表API
func (s *Server) getRecordList(c *gin.Context) {
	// 解析请求参数
	year, err := strconv.Atoi(c.DefaultQuery("year", "0"))
	if err != nil || year == 0 {
		// 默认为当前年份
		year = time.Now().Year()
	}

	month, err := strconv.Atoi(c.DefaultQuery("month", "0"))
	if err != nil || month == 0 {
		// 默认为当前月份
		month = int(time.Now().Month())
	}

	taskIDStr := c.Query("taskId")
	var taskID *int64
	if taskIDStr != "" {
		id, err := strconv.ParseInt(taskIDStr, 10, 64)
		if err == nil {
			taskID = &id
		}
	}

	departmentIDStr := c.Query("departmentId")
	var departmentID *int64
	if departmentIDStr != "" {
		id, err := strconv.ParseInt(departmentIDStr, 10, 64)
		if err == nil {
			departmentID = &id
		}
	}

	successStr := c.Query("success")
	var success *int8
	if successStr != "" {
		val, err := strconv.ParseInt(successStr, 10, 8)
		if err == nil {
			s := int8(val)
			success = &s
		}
	}

	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	size, err := strconv.Atoi(c.DefaultQuery("size", "10"))
	if err != nil || size < 1 || size > 100 {
		size = 10
	}

	// 调用任务服务获取执行记录列表
	records, total, err := s.taskService.GetRecordList(year, month, taskID, departmentID, success, page, size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ResponseBody{
			Code:    ErrorCodes.InternalServerError,
			Message: err.Error(),
			Data:    nil,
		})
		return
	}

	c.JSON(http.StatusOK, ResponseBody{
		Code:    ErrorCodes.Success,
		Message: "success",
		Data: map[string]interface{}{
			"records": records,
			"total":   total,
		},
	})
}

// 获取执行记录详情API
func (s *Server) getRecord(c *gin.Context) {
	// 解析记录ID
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, ResponseBody{
			Code:    ErrorCodes.BadRequest,
			Message: "invalid record id",
			Data:    nil,
		})
		return
	}

	// 解析年月
	year, err := strconv.Atoi(c.DefaultQuery("year", "0"))
	if err != nil || year == 0 {
		// 默认为当前年份
		year = time.Now().Year()
	}

	month, err := strconv.Atoi(c.DefaultQuery("month", "0"))
	if err != nil || month == 0 {
		// 默认为当前月份
		month = int(time.Now().Month())
	}

	// 调用任务服务获取执行记录详情
	record, err := s.taskService.GetRecordByID(id, year, month)
	if err != nil {
		c.JSON(http.StatusNotFound, ResponseBody{
			Code:    ErrorCodes.NotFound,
			Message: err.Error(),
			Data:    nil,
		})
		return
	}

	c.JSON(http.StatusOK, ResponseBody{
		Code:    ErrorCodes.Success,
		Message: "success",
		Data:    record,
	})
}

// 获取执行记录统计API
func (s *Server) getRecordStats(c *gin.Context) {
	// 解析请求参数
	year, err := strconv.Atoi(c.DefaultQuery("year", "0"))
	if err != nil || year == 0 {
		// 默认为当前年份
		year = time.Now().Year()
	}

	month, err := strconv.Atoi(c.DefaultQuery("month", "0"))
	if err != nil || month == 0 {
		// 默认为当前月份
		month = int(time.Now().Month())
	}

	taskIDStr := c.Query("taskId")
	var taskID *int64
	if taskIDStr != "" {
		id, err := strconv.ParseInt(taskIDStr, 10, 64)
		if err == nil {
			taskID = &id
		}
	}

	departmentIDStr := c.Query("departmentId")
	var departmentID *int64
	if departmentIDStr != "" {
		id, err := strconv.ParseInt(departmentIDStr, 10, 64)
		if err == nil {
			departmentID = &id
		}
	}

	// 调用任务服务获取执行记录统计
	stats, err := s.taskService.GetRecordStats(year, month, taskID, departmentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ResponseBody{
			Code:    ErrorCodes.InternalServerError,
			Message: err.Error(),
			Data:    nil,
		})
		return
	}

	c.JSON(http.StatusOK, ResponseBody{
		Code:    ErrorCodes.Success,
		Message: "success",
		Data:    stats,
	})
}

// 仅需实现的几个为完善示例的API方法
func (s *Server) getDepartmentList(c *gin.Context) {
	// 解析请求参数
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	size, err := strconv.Atoi(c.DefaultQuery("size", "10"))
	if err != nil || size < 1 || size > 100 {
		size = 10
	}

	// 调用部门服务获取部门列表
	departments, total, err := s.authService.GetDepartmentList(page, size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ResponseBody{
			Code:    ErrorCodes.InternalServerError,
			Message: err.Error(),
			Data:    nil,
		})
		return
	}

	c.JSON(http.StatusOK, ResponseBody{
		Code:    ErrorCodes.Success,
		Message: "success",
		Data: map[string]interface{}{
			"departments": departments,
			"total":       total,
			"page":        page,
			"size":        size,
		},
	})
}

func (s *Server) getDepartment(c *gin.Context) {
	// 解析部门ID
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, ResponseBody{
			Code:    ErrorCodes.BadRequest,
			Message: "invalid department id",
			Data:    nil,
		})
		return
	}

	// 调用部门服务获取部门详情
	department, err := s.authService.GetDepartmentByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, ResponseBody{
			Code:    ErrorCodes.NotFound,
			Message: err.Error(),
			Data:    nil,
		})
		return
	}

	c.JSON(http.StatusOK, ResponseBody{
		Code:    ErrorCodes.Success,
		Message: "success",
		Data:    department,
	})
}

func (s *Server) createDepartment(c *gin.Context) {
	// 解析请求参数
	var department entity.Department
	if err := c.ShouldBindJSON(&department); err != nil {
		c.JSON(http.StatusBadRequest, ResponseBody{
			Code:    ErrorCodes.BadRequest,
			Message: "invalid request parameters",
			Data:    nil,
		})
		return
	}

	// 调用部门服务创建部门
	id, err := s.authService.CreateDepartment(&department)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ResponseBody{
			Code:    ErrorCodes.InternalServerError,
			Message: err.Error(),
			Data:    nil,
		})
		return
	}

	c.JSON(http.StatusOK, ResponseBody{
		Code:    ErrorCodes.Success,
		Message: "success",
		Data:    map[string]interface{}{"id": id},
	})
}

func (s *Server) updateDepartment(c *gin.Context) {
	// 解析部门ID
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, ResponseBody{
			Code:    ErrorCodes.BadRequest,
			Message: "invalid department id",
			Data:    nil,
		})
		return
	}

	// 解析请求参数
	var department entity.Department
	if err := c.ShouldBindJSON(&department); err != nil {
		c.JSON(http.StatusBadRequest, ResponseBody{
			Code:    ErrorCodes.BadRequest,
			Message: "invalid request parameters",
			Data:    nil,
		})
		return
	}

	// 设置部门ID
	department.ID = id

	// 调用部门服务更新部门
	err = s.authService.UpdateDepartment(&department)
	if err != nil {
		status := http.StatusInternalServerError
		code := ErrorCodes.InternalServerError

		// 根据错误类型设置不同的状态码
		if err.Error() == "department not found" {
			status = http.StatusNotFound
			code = ErrorCodes.NotFound
		}

		c.JSON(status, ResponseBody{
			Code:    code,
			Message: err.Error(),
			Data:    nil,
		})
		return
	}

	c.JSON(http.StatusOK, ResponseBody{
		Code:    ErrorCodes.Success,
		Message: "success",
		Data:    nil,
	})
}

func (s *Server) deleteDepartment(c *gin.Context) {
	// 解析部门ID
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, ResponseBody{
			Code:    ErrorCodes.BadRequest,
			Message: "invalid department id",
			Data:    nil,
		})
		return
	}

	// 调用部门服务删除部门
	err = s.authService.DeleteDepartment(id)
	if err != nil {
		status := http.StatusInternalServerError
		code := ErrorCodes.InternalServerError

		// 根据错误类型设置不同的状态码
		if err.Error() == "department not found" {
			status = http.StatusNotFound
			code = ErrorCodes.NotFound
		} else if err.Error() == "cannot delete department with users" {
			status = http.StatusBadRequest
			code = ErrorCodes.BadRequest
			c.JSON(status, ResponseBody{
				Code:    code,
				Message: err.Error(),
				Data:    nil,
			})
			return
		}

		c.JSON(status, ResponseBody{
			Code:    code,
			Message: err.Error(),
			Data:    nil,
		})
		return
	}

	c.JSON(http.StatusOK, ResponseBody{
		Code:    ErrorCodes.Success,
		Message: "success",
		Data:    nil,
	})
}

func (s *Server) getUserList(c *gin.Context) {
	// 解析请求参数
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	size, err := strconv.Atoi(c.DefaultQuery("size", "10"))
	if err != nil || size < 1 || size > 100 {
		size = 10
	}

	departmentIDStr := c.Query("departmentId")
	var departmentID int64 = 0
	if departmentIDStr != "" {
		id, err := strconv.ParseInt(departmentIDStr, 10, 64)
		if err == nil {
			departmentID = id
		}
	}

	// 调用用户服务获取用户列表
	users, total, err := s.authService.GetUserList(departmentID, page, size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ResponseBody{
			Code:    ErrorCodes.InternalServerError,
			Message: err.Error(),
			Data:    nil,
		})
		return
	}

	c.JSON(http.StatusOK, ResponseBody{
		Code:    ErrorCodes.Success,
		Message: "success",
		Data: map[string]interface{}{
			"users": users,
			"total": total,
			"page":  page,
			"size":  size,
		},
	})
}

func (s *Server) getUser(c *gin.Context) {
	// 解析用户ID
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, ResponseBody{
			Code:    ErrorCodes.BadRequest,
			Message: "invalid user id",
			Data:    nil,
		})
		return
	}

	// 调用用户服务获取用户详情
	user, err := s.authService.GetUserByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, ResponseBody{
			Code:    ErrorCodes.NotFound,
			Message: err.Error(),
			Data:    nil,
		})
		return
	}

	c.JSON(http.StatusOK, ResponseBody{
		Code:    ErrorCodes.Success,
		Message: "success",
		Data:    user,
	})
}

func (s *Server) createUser(c *gin.Context) {
	// 解析请求参数
	var user entity.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, ResponseBody{
			Code:    ErrorCodes.BadRequest,
			Message: "invalid request parameters",
			Data:    nil,
		})
		return
	}

	// 调用用户服务创建用户
	id, err := s.authService.CreateUser(&user)
	if err != nil {
		status := http.StatusInternalServerError
		code := ErrorCodes.InternalServerError

		// 根据错误类型设置不同的状态码
		if err.Error() == "username already exists" {
			status = http.StatusBadRequest
			code = ErrorCodes.BadRequest
		}

		c.JSON(status, ResponseBody{
			Code:    code,
			Message: err.Error(),
			Data:    nil,
		})
		return
	}

	c.JSON(http.StatusOK, ResponseBody{
		Code:    ErrorCodes.Success,
		Message: "success",
		Data:    map[string]interface{}{"id": id},
	})
}

func (s *Server) updateUser(c *gin.Context) {
	// 解析用户ID
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, ResponseBody{
			Code:    ErrorCodes.BadRequest,
			Message: "invalid user id",
			Data:    nil,
		})
		return
	}

	// 解析请求参数
	var user entity.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, ResponseBody{
			Code:    ErrorCodes.BadRequest,
			Message: "invalid request parameters",
			Data:    nil,
		})
		return
	}

	// 设置用户ID
	user.ID = id

	// 调用用户服务更新用户
	err = s.authService.UpdateUser(&user)
	if err != nil {
		status := http.StatusInternalServerError
		code := ErrorCodes.InternalServerError

		// 根据错误类型设置不同的状态码
		if err.Error() == "user not found" {
			status = http.StatusNotFound
			code = ErrorCodes.NotFound
		} else if err.Error() == "username already exists" {
			status = http.StatusBadRequest
			code = ErrorCodes.BadRequest
		}

		c.JSON(status, ResponseBody{
			Code:    code,
			Message: err.Error(),
			Data:    nil,
		})
		return
	}

	c.JSON(http.StatusOK, ResponseBody{
		Code:    ErrorCodes.Success,
		Message: "success",
		Data:    nil,
	})
}

func (s *Server) deleteUser(c *gin.Context) {
	// 解析用户ID
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, ResponseBody{
			Code:    ErrorCodes.BadRequest,
			Message: "invalid user id",
			Data:    nil,
		})
		return
	}

	// 调用用户服务删除用户
	err = s.authService.DeleteUser(id)
	if err != nil {
		status := http.StatusInternalServerError
		code := ErrorCodes.InternalServerError

		// 根据错误类型设置不同的状态码
		if err.Error() == "user not found" {
			status = http.StatusNotFound
			code = ErrorCodes.NotFound
		} else if err.Error() == "cannot delete user with tasks" {
			status = http.StatusBadRequest
			code = ErrorCodes.BadRequest
		}

		c.JSON(status, ResponseBody{
			Code:    code,
			Message: err.Error(),
			Data:    nil,
		})
		return
	}

	c.JSON(http.StatusOK, ResponseBody{
		Code:    ErrorCodes.Success,
		Message: "success",
		Data:    nil,
	})
}

// 由于AuthService没有提供UpdateUserPassword方法，我们需要在这里自己实现
func (s *Server) updateUserPassword(c *gin.Context) {
	// 解析用户ID
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, ResponseBody{
			Code:    ErrorCodes.BadRequest,
			Message: "invalid user id",
			Data:    nil,
		})
		return
	}

	// 解析请求参数
	var req struct {
		OldPassword string `json:"oldPassword" binding:"required"`
		NewPassword string `json:"newPassword" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ResponseBody{
			Code:    ErrorCodes.BadRequest,
			Message: "invalid request parameters",
			Data:    nil,
		})
		return
	}

	// 获取用户信息
	user, err := s.authService.GetUserByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, ResponseBody{
			Code:    ErrorCodes.NotFound,
			Message: "user not found",
			Data:    nil,
		})
		return
	}

	// 创建更新用户对象
	updateUser := &entity.User{
		ID:       id,
		Password: req.NewPassword,
		// 其他信息保持不变
		Username:     user.Username,
		RealName:     user.RealName,
		Email:        user.Email,
		Phone:        user.Phone,
		DepartmentID: user.DepartmentID,
		RoleID:       user.RoleID,
		Status:       user.Status,
	}

	// 调用用户服务更新用户
	err = s.authService.UpdateUser(updateUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ResponseBody{
			Code:    ErrorCodes.InternalServerError,
			Message: err.Error(),
			Data:    nil,
		})
		return
	}

	c.JSON(http.StatusOK, ResponseBody{
		Code:    ErrorCodes.Success,
		Message: "success",
		Data:    nil,
	})
}

func (s *Server) getRoleList(c *gin.Context) {
	// 解析请求参数
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	size, err := strconv.Atoi(c.DefaultQuery("size", "10"))
	if err != nil || size < 1 || size > 100 {
		size = 10
	}

	// 调用角色服务获取角色列表
	roles, total, err := s.authService.GetRoleList(page, size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ResponseBody{
			Code:    ErrorCodes.InternalServerError,
			Message: err.Error(),
			Data:    nil,
		})
		return
	}

	c.JSON(http.StatusOK, ResponseBody{
		Code:    ErrorCodes.Success,
		Message: "success",
		Data: map[string]interface{}{
			"roles": roles,
			"total": total,
			"page":  page,
			"size":  size,
		},
	})
}

func (s *Server) getRole(c *gin.Context) {
	// 解析角色ID
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, ResponseBody{
			Code:    ErrorCodes.BadRequest,
			Message: "invalid role id",
			Data:    nil,
		})
		return
	}

	// 调用角色服务获取角色详情，已包含权限
	role, err := s.authService.GetRoleByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, ResponseBody{
			Code:    ErrorCodes.NotFound,
			Message: err.Error(),
			Data:    nil,
		})
		return
	}

	// 组装返回数据
	roleData := map[string]interface{}{
		"id":          role.ID,
		"name":        role.Name,
		"description": role.Description,
		"status":      role.Status,
		"permissions": role.Permissions, // GetRoleByID 方法已经获取了角色权限
		"createTime":  role.CreateTime,
		"updateTime":  role.UpdateTime,
	}

	c.JSON(http.StatusOK, ResponseBody{
		Code:    ErrorCodes.Success,
		Message: "success",
		Data:    roleData,
	})
}

func (s *Server) createRole(c *gin.Context) {
	// 解析请求参数
	var req struct {
		Name        string  `json:"name" binding:"required"`
		Description string  `json:"description"`
		Permissions []int64 `json:"permissions" binding:"required"`
		Status      int8    `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ResponseBody{
			Code:    ErrorCodes.BadRequest,
			Message: "invalid request parameters",
			Data:    nil,
		})
		return
	}

	// 创建角色对象
	role := &entity.Role{
		Name:        req.Name,
		Description: req.Description,
		Status:      req.Status,
	}

	// 创建权限对象数组
	permissions := make([]*entity.Permission, len(req.Permissions))
	for i, permID := range req.Permissions {
		permissions[i] = &entity.Permission{ID: permID}
	}
	role.Permissions = permissions

	// 调用角色服务创建角色
	id, err := s.authService.CreateRole(role)
	if err != nil {
		status := http.StatusInternalServerError
		code := ErrorCodes.InternalServerError

		// 根据错误类型设置不同的状态码
		if err.Error() == "role name already exists" {
			status = http.StatusBadRequest
			code = ErrorCodes.BadRequest
		}

		c.JSON(status, ResponseBody{
			Code:    code,
			Message: err.Error(),
			Data:    nil,
		})
		return
	}

	c.JSON(http.StatusOK, ResponseBody{
		Code:    ErrorCodes.Success,
		Message: "success",
		Data:    map[string]interface{}{"id": id},
	})
}

func (s *Server) updateRole(c *gin.Context) {
	// 解析角色ID
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, ResponseBody{
			Code:    ErrorCodes.BadRequest,
			Message: "invalid role id",
			Data:    nil,
		})
		return
	}

	// 解析请求参数
	var req struct {
		Name        string  `json:"name" binding:"required"`
		Description string  `json:"description"`
		Permissions []int64 `json:"permissions" binding:"required"`
		Status      int8    `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ResponseBody{
			Code:    ErrorCodes.BadRequest,
			Message: "invalid request parameters",
			Data:    nil,
		})
		return
	}

	// 创建角色对象
	role := &entity.Role{
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
		Status:      req.Status,
	}

	// 创建权限对象数组
	permissions := make([]*entity.Permission, len(req.Permissions))
	for i, permID := range req.Permissions {
		permissions[i] = &entity.Permission{ID: permID}
	}
	role.Permissions = permissions

	// 调用角色服务更新角色
	err = s.authService.UpdateRole(role)
	if err != nil {
		status := http.StatusInternalServerError
		code := ErrorCodes.InternalServerError

		// 根据错误类型设置不同的状态码
		if err.Error() == "role not found" {
			status = http.StatusNotFound
			code = ErrorCodes.NotFound
		} else if err.Error() == "role name already exists" {
			status = http.StatusBadRequest
			code = ErrorCodes.BadRequest
		}

		c.JSON(status, ResponseBody{
			Code:    code,
			Message: err.Error(),
			Data:    nil,
		})
		return
	}

	c.JSON(http.StatusOK, ResponseBody{
		Code:    ErrorCodes.Success,
		Message: "success",
		Data:    nil,
	})
}

func (s *Server) deleteRole(c *gin.Context) {
	// 解析角色ID
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, ResponseBody{
			Code:    ErrorCodes.BadRequest,
			Message: "invalid role id",
			Data:    nil,
		})
		return
	}

	// 调用角色服务删除角色
	err = s.authService.DeleteRole(id)
	if err != nil {
		status := http.StatusInternalServerError
		code := ErrorCodes.InternalServerError

		// 根据错误类型设置不同的状态码
		if err.Error() == "role not found" {
			status = http.StatusNotFound
			code = ErrorCodes.NotFound
		} else if err.Error() == "cannot delete role with users" {
			status = http.StatusBadRequest
			code = ErrorCodes.BadRequest
		}

		c.JSON(status, ResponseBody{
			Code:    code,
			Message: err.Error(),
			Data:    nil,
		})
		return
	}

	c.JSON(http.StatusOK, ResponseBody{
		Code:    ErrorCodes.Success,
		Message: "success",
		Data:    nil,
	})
}

func (s *Server) getPermissionList(c *gin.Context) {
	// 调用权限服务获取所有权限列表
	permissions, err := s.authService.GetPermissionList()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ResponseBody{
			Code:    ErrorCodes.InternalServerError,
			Message: err.Error(),
			Data:    nil,
		})
		return
	}

	c.JSON(http.StatusOK, ResponseBody{
		Code:    ErrorCodes.Success,
		Message: "success",
		Data:    permissions,
	})
}

// localOnly 只允许本地访问的中间件
func (s *Server) localOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		host := c.Request.Host
		// 只允许本地访问
		if host != "localhost" && host != "127.0.0.1" && host != "0.0.0.0" {
			c.JSON(http.StatusForbidden, ResponseBody{
				Code:    ErrorCodes.Forbidden,
				Message: "this API can only be accessed locally",
				Data:    nil,
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// parseValidationError 解析验证错误并返回详细的错误信息
func (s *Server) parseValidationError(err error) []map[string]string {
	validationErrors := []map[string]string{}
	errorString := err.Error()

	// 解析验证错误信息
	// 根据字段名称检测错误类型
	fieldNames := []string{"name", "cron", "type", "description", "config", "departmentId", "status"}
	for _, field := range fieldNames {
		if strings.Contains(errorString, field) {
			validationErrors = append(validationErrors, map[string]string{
				"field": field,
				"error": fmt.Sprintf("%s 字段验证失败", field),
			})
		}
	}

	// 如果没有找到具体字段错误，添加一个通用错误
	if len(validationErrors) == 0 {
		validationErrors = append(validationErrors, map[string]string{
			"field": "unknown",
			"error": "参数格式不正确",
		})
	}

	return validationErrors
}
