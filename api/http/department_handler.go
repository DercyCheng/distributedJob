package http

import (
	"net/http"
	"strconv"

	grpc "go-job/api/grpc"
	"go-job/internal/department"

	"github.com/gin-gonic/gin"
)

type DepartmentHandler struct {
	service *department.DepartmentService
}

func NewDepartmentHandler(service *department.DepartmentService) *DepartmentHandler {
	return &DepartmentHandler{
		service: service,
	}
}

type CreateDepartmentRequest struct {
	Name        string `json:"name" binding:"required"`
	Code        string `json:"code" binding:"required"`
	Description string `json:"description"`
	ParentID    string `json:"parent_id"`
	Sort        int32  `json:"sort"`
}

type UpdateDepartmentRequest struct {
	Name        string `json:"name" binding:"required"`
	Code        string `json:"code" binding:"required"`
	Description string `json:"description"`
	ParentID    string `json:"parent_id"`
	Sort        int32  `json:"sort"`
}

func (h *DepartmentHandler) CreateDepartment(c *gin.Context) {
	var req CreateDepartmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	grpcReq := &grpc.CreateDepartmentRequest{
		Name:        req.Name,
		Code:        req.Code,
		Description: req.Description,
		ParentId:    req.ParentID,
		Sort:        req.Sort,
	}

	resp, err := h.service.CreateDepartment(c.Request.Context(), grpcReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, resp.Department)
}

func (h *DepartmentHandler) ListDepartments(c *gin.Context) {
	keyword := c.Query("keyword")
	status := c.Query("status")
	pageStr := c.DefaultQuery("page", "1")
	sizeStr := c.DefaultQuery("size", "20")

	page, _ := strconv.ParseInt(pageStr, 10, 32)
	size, _ := strconv.ParseInt(sizeStr, 10, 32)

	grpcReq := &grpc.ListDepartmentsRequest{
		Page:    int32(page),
		Size:    int32(size),
		Keyword: keyword,
		Status:  status,
	}

	resp, err := h.service.ListDepartments(c.Request.Context(), grpcReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"departments": resp.Departments,
		"total":       resp.Total,
	})
}

func (h *DepartmentHandler) GetDepartment(c *gin.Context) {
	id := c.Param("id")

	grpcReq := &grpc.GetDepartmentRequest{
		Id: id,
	}

	resp, err := h.service.GetDepartment(c.Request.Context(), grpcReq)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp.Department)
}

func (h *DepartmentHandler) UpdateDepartment(c *gin.Context) {
	id := c.Param("id")

	var req UpdateDepartmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	grpcReq := &grpc.UpdateDepartmentRequest{
		Id:          id,
		Name:        req.Name,
		Code:        req.Code,
		Description: req.Description,
		ParentId:    req.ParentID,
		Sort:        req.Sort,
	}

	resp, err := h.service.UpdateDepartment(c.Request.Context(), grpcReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp.Department)
}

func (h *DepartmentHandler) DeleteDepartment(c *gin.Context) {
	id := c.Param("id")

	grpcReq := &grpc.DeleteDepartmentRequest{
		Id: id,
	}

	_, err := h.service.DeleteDepartment(c.Request.Context(), grpcReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Department deleted successfully"})
}

func (h *DepartmentHandler) GetDepartmentTree(c *gin.Context) {
	grpcReq := &grpc.GetDepartmentTreeRequest{}

	resp, err := h.service.GetDepartmentTree(c.Request.Context(), grpcReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"departments": resp.Departments,
	})
}
