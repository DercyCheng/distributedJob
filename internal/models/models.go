package models

import (
	"time"

	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	ID           string         `gorm:"primaryKey;type:varchar(36)" json:"id"`
	Username     string         `gorm:"type:varchar(100);uniqueIndex;not null" json:"username"`
	Email        string         `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	Password     string         `gorm:"type:varchar(255);not null" json:"-"`
	RealName     string         `gorm:"type:varchar(100)" json:"real_name"`
	Phone        string         `gorm:"type:varchar(20)" json:"phone"`
	Avatar       string         `gorm:"type:varchar(500)" json:"avatar"`
	Status       UserStatus     `gorm:"type:varchar(20);default:'active'" json:"status"`
	DepartmentID string         `gorm:"type:varchar(36);index" json:"department_id"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"deleted_at"`
	LastLoginAt  *time.Time     `json:"last_login_at"`

	// 关联
	Department *Department `gorm:"foreignKey:DepartmentID" json:"department,omitempty"`
	UserRoles  []UserRole  `gorm:"foreignKey:UserID" json:"user_roles,omitempty"`
	Roles      []Role      `gorm:"many2many:user_roles;" json:"roles,omitempty"`
}

// Department 部门模型
type Department struct {
	ID          string         `gorm:"primaryKey;type:varchar(36)" json:"id"`
	Name        string         `gorm:"type:varchar(100);not null;index" json:"name"`
	Code        string         `gorm:"type:varchar(50);uniqueIndex" json:"code"`
	Description string         `gorm:"type:text" json:"description"`
	ParentID    *string        `gorm:"type:varchar(36);index" json:"parent_id"`
	Status      DeptStatus     `gorm:"type:varchar(20);default:'active'" json:"status"`
	Sort        int            `gorm:"default:0" json:"sort"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at"`

	// 关联
	Parent   *Department  `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	Children []Department `gorm:"foreignKey:ParentID" json:"children,omitempty"`
	Users    []User       `gorm:"foreignKey:DepartmentID" json:"users,omitempty"`
}

// Role 角色模型
type Role struct {
	ID          string         `gorm:"primaryKey;type:varchar(36)" json:"id"`
	Name        string         `gorm:"type:varchar(100);not null;index" json:"name"`
	Code        string         `gorm:"type:varchar(50);uniqueIndex" json:"code"`
	Description string         `gorm:"type:text" json:"description"`
	Status      RoleStatus     `gorm:"type:varchar(20);default:'active'" json:"status"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at"`

	// 关联
	UserRoles       []UserRole       `gorm:"foreignKey:RoleID" json:"user_roles,omitempty"`
	Users           []User           `gorm:"many2many:user_roles;" json:"users,omitempty"`
	RolePermissions []RolePermission `gorm:"foreignKey:RoleID" json:"role_permissions,omitempty"`
	Permissions     []Permission     `gorm:"many2many:role_permissions;" json:"permissions,omitempty"`
}

// Permission 权限模型
type Permission struct {
	ID        string         `gorm:"primaryKey;type:varchar(36)" json:"id"`
	Name      string         `gorm:"type:varchar(100);not null;index" json:"name"`
	Code      string         `gorm:"type:varchar(100);uniqueIndex" json:"code"`
	Type      PermType       `gorm:"type:varchar(20);default:'menu'" json:"type"`
	Resource  string         `gorm:"type:varchar(100)" json:"resource"`
	Action    string         `gorm:"type:varchar(50)" json:"action"`
	ParentID  *string        `gorm:"type:varchar(36);index" json:"parent_id"`
	Path      string         `gorm:"type:varchar(200)" json:"path"`
	Icon      string         `gorm:"type:varchar(100)" json:"icon"`
	Sort      int            `gorm:"default:0" json:"sort"`
	Status    PermStatus     `gorm:"type:varchar(20);default:'active'" json:"status"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`

	// 关联
	Parent          *Permission      `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	Children        []Permission     `gorm:"foreignKey:ParentID" json:"children,omitempty"`
	RolePermissions []RolePermission `gorm:"foreignKey:PermissionID" json:"role_permissions,omitempty"`
	Roles           []Role           `gorm:"many2many:role_permissions;" json:"roles,omitempty"`
}

// UserRole 用户角色关联表
type UserRole struct {
	ID        string         `gorm:"primaryKey;type:varchar(36)" json:"id"`
	UserID    string         `gorm:"type:varchar(36);not null;index" json:"user_id"`
	RoleID    string         `gorm:"type:varchar(36);not null;index" json:"role_id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`

	// 关联
	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Role *Role `gorm:"foreignKey:RoleID" json:"role,omitempty"`
}

// RolePermission 角色权限关联表
type RolePermission struct {
	ID           string         `gorm:"primaryKey;type:varchar(36)" json:"id"`
	RoleID       string         `gorm:"type:varchar(36);not null;index" json:"role_id"`
	PermissionID string         `gorm:"type:varchar(36);not null;index" json:"permission_id"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"deleted_at"`

	// 关联
	Role       *Role       `gorm:"foreignKey:RoleID" json:"role,omitempty"`
	Permission *Permission `gorm:"foreignKey:PermissionID" json:"permission,omitempty"`
}

// AISchedule 智能调度记录
type AISchedule struct {
	ID             string         `gorm:"primaryKey;type:varchar(36)" json:"id"`
	JobID          string         `gorm:"type:varchar(36);not null;index" json:"job_id"`
	PromptTemplate string         `gorm:"type:text" json:"prompt_template"`
	AIResponse     string         `gorm:"type:longtext" json:"ai_response"`
	Strategy       string         `gorm:"type:varchar(100)" json:"strategy"`
	Priority       int            `gorm:"default:0" json:"priority"`
	Context        string         `gorm:"type:json" json:"context"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"deleted_at"`

	// 关联
	Job *Job `gorm:"foreignKey:JobID" json:"job,omitempty"`
}

// Job 任务模型
type Job struct {
	ID            string         `gorm:"primaryKey;type:varchar(36)" json:"id"`
	Name          string         `gorm:"type:varchar(255);not null;index" json:"name"`
	Description   string         `gorm:"type:text" json:"description"`
	Cron          string         `gorm:"type:varchar(100);not null" json:"cron"`
	Command       string         `gorm:"type:text;not null" json:"command"`
	Params        string         `gorm:"type:json" json:"params"` // JSON 字符串
	Enabled       bool           `gorm:"default:true" json:"enabled"`
	RetryAttempts int            `gorm:"default:3" json:"retry_attempts"`
	Timeout       int            `gorm:"default:300" json:"timeout"` // 秒
	Priority      int            `gorm:"default:0" json:"priority"`
	DepartmentID  string         `gorm:"type:varchar(36);index" json:"department_id"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"deleted_at"`
	CreatedBy     string         `gorm:"type:varchar(100)" json:"created_by"`

	// 关联
	Department  *Department  `gorm:"foreignKey:DepartmentID" json:"department,omitempty"`
	Creator     *User        `gorm:"foreignKey:CreatedBy;references:Username" json:"creator,omitempty"`
	AISchedules []AISchedule `gorm:"foreignKey:JobID" json:"ai_schedules,omitempty"`
}

// JobExecution 任务执行记录
type JobExecution struct {
	ID         string             `gorm:"primaryKey;type:varchar(36)" json:"id"`
	JobID      string             `gorm:"type:varchar(36);not null;index" json:"job_id"`
	WorkerID   string             `gorm:"type:varchar(36);index" json:"worker_id"`
	Status     JobExecutionStatus `gorm:"type:varchar(20);default:'pending'" json:"status"`
	StartedAt  *time.Time         `json:"started_at"`
	FinishedAt *time.Time         `json:"finished_at"`
	Output     string             `gorm:"type:longtext" json:"output"`
	Error      string             `gorm:"type:longtext" json:"error"`
	ExitCode   int                `json:"exit_code"`
	CreatedAt  time.Time          `json:"created_at"`
	UpdatedAt  time.Time          `json:"updated_at"`
	DeletedAt  gorm.DeletedAt     `gorm:"index" json:"deleted_at"`

	// 关联
	Job    Job    `gorm:"foreignKey:JobID" json:"job,omitempty"`
	Worker Worker `gorm:"foreignKey:WorkerID" json:"worker,omitempty"`
}

// Worker 工作节点
type Worker struct {
	ID            string         `gorm:"primaryKey;type:varchar(36)" json:"id"`
	Name          string         `gorm:"type:varchar(255);not null" json:"name"`
	IP            string         `gorm:"type:varchar(45);not null" json:"ip"`
	Port          int            `gorm:"not null" json:"port"`
	Status        WorkerStatus   `gorm:"type:varchar(20);default:'offline'" json:"status"`
	Capacity      int            `gorm:"default:10" json:"capacity"`
	CurrentLoad   int            `gorm:"default:0" json:"current_load"`
	LastHeartbeat *time.Time     `json:"last_heartbeat"`
	Metadata      string         `gorm:"type:json" json:"metadata"` // JSON 字符串
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

// JobSchedule 任务调度记录
type JobSchedule struct {
	ID          string         `gorm:"primaryKey;type:varchar(36)" json:"id"`
	JobID       string         `gorm:"type:varchar(36);not null;index" json:"job_id"`
	ScheduledAt time.Time      `gorm:"not null;index" json:"scheduled_at"`
	ExecutedAt  *time.Time     `json:"executed_at"`
	Status      ScheduleStatus `gorm:"type:varchar(20);default:'pending'" json:"status"`
	WorkerID    string         `gorm:"type:varchar(36);index" json:"worker_id"`
	ExecutionID string         `gorm:"type:varchar(36);index" json:"execution_id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at"`

	// 关联
	Job       Job          `gorm:"foreignKey:JobID" json:"job,omitempty"`
	Worker    Worker       `gorm:"foreignKey:WorkerID" json:"worker,omitempty"`
	Execution JobExecution `gorm:"foreignKey:ExecutionID" json:"execution,omitempty"`
}

// 用户状态
type UserStatus string

const (
	UserStatusActive   UserStatus = "active"
	UserStatusInactive UserStatus = "inactive"
	UserStatusLocked   UserStatus = "locked"
)

// 部门状态
type DeptStatus string

const (
	DeptStatusActive   DeptStatus = "active"
	DeptStatusInactive DeptStatus = "inactive"
)

// 角色状态
type RoleStatus string

const (
	RoleStatusActive   RoleStatus = "active"
	RoleStatusInactive RoleStatus = "inactive"
)

// 权限类型
type PermType string

const (
	PermTypeMenu   PermType = "menu"
	PermTypeButton PermType = "button"
	PermTypeAPI    PermType = "api"
)

// 权限状态
type PermStatus string

const (
	PermStatusActive   PermStatus = "active"
	PermStatusInactive PermStatus = "inactive"
)

// 任务执行状态
type JobExecutionStatus string

const (
	ExecutionStatusPending   JobExecutionStatus = "pending"
	ExecutionStatusRunning   JobExecutionStatus = "running"
	ExecutionStatusSuccess   JobExecutionStatus = "success"
	ExecutionStatusFailed    JobExecutionStatus = "failed"
	ExecutionStatusTimeout   JobExecutionStatus = "timeout"
	ExecutionStatusCancelled JobExecutionStatus = "cancelled"
)

// 工作节点状态
type WorkerStatus string

const (
	WorkerStatusOffline     WorkerStatus = "offline"
	WorkerStatusOnline      WorkerStatus = "online"
	WorkerStatusBusy        WorkerStatus = "busy"
	WorkerStatusMaintenance WorkerStatus = "maintenance"
)

// 调度状态
type ScheduleStatus string

const (
	ScheduleStatusPending   ScheduleStatus = "pending"
	ScheduleStatusAssigned  ScheduleStatus = "assigned"
	ScheduleStatusExecuting ScheduleStatus = "executing"
	ScheduleStatusCompleted ScheduleStatus = "completed"
	ScheduleStatusFailed    ScheduleStatus = "failed"
	ScheduleStatusSkipped   ScheduleStatus = "skipped"
)

// TableName 设置表名
func (User) TableName() string {
	return "users"
}

func (Department) TableName() string {
	return "departments"
}

func (Role) TableName() string {
	return "roles"
}

func (Permission) TableName() string {
	return "permissions"
}

func (UserRole) TableName() string {
	return "user_roles"
}

func (RolePermission) TableName() string {
	return "role_permissions"
}

func (AISchedule) TableName() string {
	return "ai_schedules"
}

func (Job) TableName() string {
	return "jobs"
}

func (JobExecution) TableName() string {
	return "job_executions"
}

func (Worker) TableName() string {
	return "workers"
}

func (JobSchedule) TableName() string {
	return "job_schedules"
}
