package http

import (
	"net/http"
	"strconv"

	grpc "go-job/api/grpc"
	"go-job/internal/permission"
	"go-job/internal/role"
	"go-job/internal/user"

	"github.com/gin-gonic/gin"
)

// UserHandler 用户处理器
type UserHandler struct {
	service *user.UserService
}

func NewUserHandler(service *user.UserService) *UserHandler {
	return &UserHandler{service: service}
}

func (h *UserHandler) CreateUser(c *gin.Context) {
	var req struct {
		Username     string `json:"username" binding:"required"`
		Email        string `json:"email"`
		Password     string `json:"password" binding:"required"`
		RealName     string `json:"real_name"`
		Phone        string `json:"phone"`
		DepartmentID string `json:"department_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	grpcReq := &grpc.CreateUserRequest{
		Username:     req.Username,
		Email:        req.Email,
		Password:     req.Password,
		RealName:     req.RealName,
		Phone:        req.Phone,
		DepartmentId: req.DepartmentID,
	}

	resp, err := h.service.CreateUser(c.Request.Context(), grpcReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, resp.User)
}

func (h *UserHandler) ListUsers(c *gin.Context) {
	departmentID := c.Query("department_id")
	status := c.Query("status")
	keyword := c.Query("keyword")
	pageStr := c.DefaultQuery("page", "1")
	sizeStr := c.DefaultQuery("size", "20")

	page, _ := strconv.ParseInt(pageStr, 10, 32)
	size, _ := strconv.ParseInt(sizeStr, 10, 32)

	grpcReq := &grpc.ListUsersRequest{
		Page:         int32(page),
		Size:         int32(size),
		Keyword:      keyword,
		DepartmentId: departmentID,
		Status:       status,
	}

	resp, err := h.service.ListUsers(c.Request.Context(), grpcReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"users": resp.Users,
		"total": resp.Total,
	})
}

func (h *UserHandler) GetUser(c *gin.Context) {
	id := c.Param("id")

	grpcReq := &grpc.GetUserRequest{Id: id}

	resp, err := h.service.GetUser(c.Request.Context(), grpcReq)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp.User)
}

func (h *UserHandler) UpdateUser(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		Email        string   `json:"email"`
		RealName     string   `json:"real_name"`
		Phone        string   `json:"phone"`
		Avatar       string   `json:"avatar"`
		DepartmentID string   `json:"department_id"`
		Status       string   `json:"status"`
		RoleIDs      []string `json:"role_ids"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	grpcReq := &grpc.UpdateUserRequest{
		Id:           id,
		Email:        req.Email,
		RealName:     req.RealName,
		Phone:        req.Phone,
		Avatar:       req.Avatar,
		DepartmentId: req.DepartmentID,
		Status:       req.Status,
		RoleIds:      req.RoleIDs,
	}

	resp, err := h.service.UpdateUser(c.Request.Context(), grpcReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp.User)
}

func (h *UserHandler) DeleteUser(c *gin.Context) {
	id := c.Param("id")

	grpcReq := &grpc.DeleteUserRequest{Id: id}

	_, err := h.service.DeleteUser(c.Request.Context(), grpcReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

func (h *UserHandler) AssignUserRoles(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		RoleIDs []string `json:"role_ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	grpcReq := &grpc.AssignUserRolesRequest{
		UserId:  id,
		RoleIds: req.RoleIDs,
	}

	_, err := h.service.AssignUserRoles(c.Request.Context(), grpcReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "用户角色分配成功"})
}

// RoleHandler 角色处理器
type RoleHandler struct {
	service *role.RoleService
}

func NewRoleHandler(service *role.RoleService) *RoleHandler {
	return &RoleHandler{service: service}
}

func (h *RoleHandler) CreateRole(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Code        string `json:"code" binding:"required"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	grpcReq := &grpc.CreateRoleRequest{
		Name:        req.Name,
		Code:        req.Code,
		Description: req.Description,
	}

	resp, err := h.service.CreateRole(c.Request.Context(), grpcReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, resp.Role)
}

func (h *RoleHandler) ListRoles(c *gin.Context) {
	status := c.Query("status")
	keyword := c.Query("keyword")
	pageStr := c.DefaultQuery("page", "1")
	sizeStr := c.DefaultQuery("size", "20")

	page, _ := strconv.ParseInt(pageStr, 10, 32)
	size, _ := strconv.ParseInt(sizeStr, 10, 32)

	grpcReq := &grpc.ListRolesRequest{
		Page:    int32(page),
		Size:    int32(size),
		Keyword: keyword,
		Status:  status,
	}

	resp, err := h.service.ListRoles(c.Request.Context(), grpcReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"roles": resp.Roles,
		"total": resp.Total,
	})
}

func (h *RoleHandler) GetRole(c *gin.Context) {
	id := c.Param("id")

	grpcReq := &grpc.GetRoleRequest{Id: id}

	resp, err := h.service.GetRole(c.Request.Context(), grpcReq)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp.Role)
}

func (h *RoleHandler) UpdateRole(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		Name        string `json:"name" binding:"required"`
		Code        string `json:"code" binding:"required"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	grpcReq := &grpc.UpdateRoleRequest{
		Id:          id,
		Name:        req.Name,
		Code:        req.Code,
		Description: req.Description,
	}

	resp, err := h.service.UpdateRole(c.Request.Context(), grpcReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp.Role)
}

func (h *RoleHandler) DeleteRole(c *gin.Context) {
	id := c.Param("id")

	grpcReq := &grpc.DeleteRoleRequest{Id: id}

	_, err := h.service.DeleteRole(c.Request.Context(), grpcReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Role deleted successfully"})
}

func (h *RoleHandler) AssignRolePermissions(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		PermissionIds []string `json:"permission_ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	grpcReq := &grpc.AssignPermissionsRequest{
		RoleId:        id,
		PermissionIds: req.PermissionIds,
	}

	_, err := h.service.AssignPermissions(c.Request.Context(), grpcReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Permissions assigned successfully"})
}

// PermissionHandler 权限处理器
type PermissionHandler struct {
	service *permission.PermissionService
}

func NewPermissionHandler(service *permission.PermissionService) *PermissionHandler {
	return &PermissionHandler{service: service}
}

func (h *PermissionHandler) CreatePermission(c *gin.Context) {
	var req struct {
		Name     string `json:"name" binding:"required"`
		Code     string `json:"code" binding:"required"`
		Type     string `json:"type" binding:"required"`
		Resource string `json:"resource"`
		Action   string `json:"action"`
		ParentID string `json:"parent_id"`
		Path     string `json:"path"`
		Icon     string `json:"icon"`
		Sort     int32  `json:"sort"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	grpcReq := &grpc.CreatePermissionRequest{
		Name:     req.Name,
		Code:     req.Code,
		Type:     req.Type,
		Resource: req.Resource,
		Action:   req.Action,
		ParentId: req.ParentID,
		Path:     req.Path,
		Icon:     req.Icon,
		Sort:     req.Sort,
	}

	resp, err := h.service.CreatePermission(c.Request.Context(), grpcReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, resp.Permission)
}

func (h *PermissionHandler) ListPermissions(c *gin.Context) {
	permType := c.Query("type")
	status := c.Query("status")
	keyword := c.Query("keyword")
	pageStr := c.DefaultQuery("page", "1")
	sizeStr := c.DefaultQuery("size", "20")

	page, _ := strconv.ParseInt(pageStr, 10, 32)
	size, _ := strconv.ParseInt(sizeStr, 10, 32)

	grpcReq := &grpc.ListPermissionsRequest{
		Page:    int32(page),
		Size:    int32(size),
		Keyword: keyword,
		Type:    permType,
		Status:  status,
	}

	resp, err := h.service.ListPermissions(c.Request.Context(), grpcReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"permissions": resp.Permissions,
		"total":       resp.Total,
	})
}

func (h *PermissionHandler) GetPermission(c *gin.Context) {
	id := c.Param("id")

	grpcReq := &grpc.GetPermissionRequest{Id: id}

	resp, err := h.service.GetPermission(c.Request.Context(), grpcReq)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp.Permission)
}

func (h *PermissionHandler) UpdatePermission(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		Name     string `json:"name" binding:"required"`
		Code     string `json:"code" binding:"required"`
		Type     string `json:"type" binding:"required"`
		Resource string `json:"resource"`
		Action   string `json:"action"`
		ParentID string `json:"parent_id"`
		Path     string `json:"path"`
		Icon     string `json:"icon"`
		Sort     int32  `json:"sort"`
		Status   string `json:"status"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	grpcReq := &grpc.UpdatePermissionRequest{
		Id:       id,
		Name:     req.Name,
		Code:     req.Code,
		Type:     req.Type,
		Resource: req.Resource,
		Action:   req.Action,
		ParentId: req.ParentID,
		Path:     req.Path,
		Icon:     req.Icon,
		Sort:     req.Sort,
		Status:   req.Status,
	}

	resp, err := h.service.UpdatePermission(c.Request.Context(), grpcReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp.Permission)
}

func (h *PermissionHandler) DeletePermission(c *gin.Context) {
	id := c.Param("id")

	grpcReq := &grpc.DeletePermissionRequest{Id: id}

	_, err := h.service.DeletePermission(c.Request.Context(), grpcReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Permission deleted successfully"})
}

func (h *PermissionHandler) GetPermissionTree(c *gin.Context) {
	grpcReq := &grpc.GetPermissionTreeRequest{}

	resp, err := h.service.GetPermissionTree(c.Request.Context(), grpcReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"permissions": resp.Permissions,
	})
}
