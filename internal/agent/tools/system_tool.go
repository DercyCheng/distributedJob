package tools

import (
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"distributedJob/internal/agent/types"
)

// SystemTool 系统操作工具实现
type SystemTool struct {
	// 可能需要的依赖，如系统管理客户端等
}

// NewSystemTool 创建一个新的系统工具
func NewSystemTool() *SystemTool {
	return &SystemTool{}
}

// Name 实现Tool接口，返回工具名称
func (t *SystemTool) Name() string {
	return "system_tool"
}

// Description 实现Tool接口，返回工具描述
func (t *SystemTool) Description() string {
	return "用于监控和管理系统资源的工具。支持查询系统状态、服务管理、资源分配等操作。"
}

// Parameters 实现Tool接口，返回工具参数定义
func (t *SystemTool) Parameters() map[string]types.Parameter {
	return map[string]types.Parameter{
		"action": {
			Type:        "string",
			Description: "执行的操作：status（系统状态）、service（服务管理）、resource（资源管理）、log（日志查询）",
			Required:    true,
		},
		"target": {
			Type:        "string",
			Description: "操作的目标，如服务名称、资源类型等",
			Required:    false,
		},
		"command": {
			Type:        "string",
			Description: "命令，如'start'、'stop'、'restart'等",
			Required:    false,
		},
		"params": {
			Type:        "object",
			Description: "其他参数，根据具体操作类型不同而不同",
			Required:    false,
		},
	}
}

// Execute 实现Tool接口，执行工具功能
func (t *SystemTool) Execute(args map[string]interface{}) (interface{}, error) {
	// 获取并验证action参数
	actionRaw, ok := args["action"]
	if !ok {
		return nil, errors.New("missing required parameter: action")
	}

	action, ok := actionRaw.(string)
	if !ok {
		return nil, errors.New("action must be a string")
	}

	// 根据操作类型执行不同操作
	switch action {
	case "status":
		return t.getSystemStatus(args)
	case "service":
		return t.manageService(args)
	case "resource":
		return t.manageResource(args)
	case "log":
		return t.queryLog(args)
	default:
		return nil, fmt.Errorf("unknown action: %s", action)
	}
}

// getSystemStatus 获取系统状态
func (t *SystemTool) getSystemStatus(args map[string]interface{}) (interface{}, error) {
	// 获取可选的目标参数
	var target string
	if targetRaw, ok := args["target"]; ok {
		target, _ = targetRaw.(string)
	}

	// 在实际应用中，这里会查询实际的系统状态
	// 目前返回模拟结果

	// 根据目标返回不同类型的状态
	if target == "cpu" || target == "" {
		return map[string]interface{}{
			"cpu": map[string]interface{}{
				"usage":       "45.2%",
				"temperature": "65°C",
				"cores":       8,
				"load_avg": map[string]string{
					"1min":  "0.98",
					"5min":  "1.20",
					"15min": "1.15",
				},
			},
		}, nil
	} else if target == "memory" || target == "" {
		return map[string]interface{}{
			"memory": map[string]interface{}{
				"total":     "32.0GB",
				"used":      "18.5GB",
				"free":      "13.5GB",
				"usage":     "57.8%",
				"swap_used": "2.1GB",
				"swap_free": "5.9GB",
			},
		}, nil
	} else if target == "disk" || target == "" {
		return map[string]interface{}{
			"disk": map[string]interface{}{
				"total": "512.0GB",
				"used":  "342.8GB",
				"free":  "169.2GB",
				"usage": "67.0%",
				"io": map[string]string{
					"read_speed":  "25MB/s",
					"write_speed": "18MB/s",
				},
			},
		}, nil
	} else if target == "network" || target == "" {
		return map[string]interface{}{
			"network": map[string]interface{}{
				"interfaces": []map[string]interface{}{
					{
						"name":       "eth0",
						"ip":         "192.168.1.100",
						"mac":        "00:1A:2B:3C:4D:5E",
						"rx_speed":   "12.5MB/s",
						"tx_speed":   "8.2MB/s",
						"rx_packets": "1250/s",
						"tx_packets": "980/s",
					},
					{
						"name":       "eth1",
						"ip":         "10.0.0.5",
						"mac":        "00:5E:4D:3C:2B:1A",
						"rx_speed":   "5.2MB/s",
						"tx_speed":   "3.1MB/s",
						"rx_packets": "520/s",
						"tx_packets": "310/s",
					},
				},
			},
		}, nil
	} else if target == "services" || target == "" {
		return map[string]interface{}{
			"services": map[string]interface{}{
				"total":   15,
				"running": 12,
				"stopped": 2,
				"failed":  1,
				"top_services": []map[string]interface{}{
					{
						"name":   "api-gateway",
						"status": "running",
						"uptime": "5d 12h 30m",
						"cpu":    "12.5%",
						"memory": "1.2GB",
					},
					{
						"name":   "database",
						"status": "running",
						"uptime": "15d 8h 45m",
						"cpu":    "28.3%",
						"memory": "4.5GB",
					},
					{
						"name":   "cache",
						"status": "running",
						"uptime": "10d 6h 15m",
						"cpu":    "15.8%",
						"memory": "2.8GB",
					},
				},
			},
		}, nil
	} else {
		return nil, fmt.Errorf("unknown status target: %s", target)
	}
}

// manageService 管理服务
func (t *SystemTool) manageService(args map[string]interface{}) (interface{}, error) {
	// 获取必要参数
	targetRaw, ok := args["target"]
	if !ok {
		return nil, errors.New("missing required parameter for service: target")
	}

	target, ok := targetRaw.(string)
	if !ok {
		return nil, errors.New("target must be a string")
	}

	commandRaw, ok := args["command"]
	if !ok {
		return nil, errors.New("missing required parameter for service: command")
	}

	command, ok := commandRaw.(string)
	if !ok {
		return nil, errors.New("command must be a string")
	}

	// 在实际应用中，这里会执行实际的服务管理命令
	// 目前返回模拟结果

	validCommands := map[string]bool{
		"start":   true,
		"stop":    true,
		"restart": true,
		"status":  true,
	}

	if !validCommands[command] {
		return nil, fmt.Errorf("invalid command: %s", command)
	}

	// 模拟服务状态
	var status string
	if command == "start" || command == "restart" {
		status = "running"
	} else if command == "stop" {
		status = "stopped"
	} else { // status command
		// 根据服务名随机选择状态
		statuses := []string{"running", "stopped", "failed"}
		hash := 0
		for _, c := range target {
			hash += int(c)
		}
		rand.Seed(int64(hash))
		status = statuses[rand.Intn(len(statuses))]
	}

	return map[string]interface{}{
		"service":   target,
		"command":   command,
		"status":    status,
		"message":   fmt.Sprintf("Service '%s' %s successfully", target, getCommandPast(command)),
		"timestamp": time.Now().Format(time.RFC3339),
	}, nil
}

// getCommandPast 获取命令的过去式
func getCommandPast(command string) string {
	switch command {
	case "start":
		return "started"
	case "stop":
		return "stopped"
	case "restart":
		return "restarted"
	case "status":
		return "status checked"
	default:
		return command + "ed"
	}
}

// manageResource 管理资源
func (t *SystemTool) manageResource(args map[string]interface{}) (interface{}, error) {
	// 获取必要参数
	targetRaw, ok := args["target"]
	if !ok {
		return nil, errors.New("missing required parameter for resource: target")
	}

	target, ok := targetRaw.(string)
	if !ok {
		return nil, errors.New("target must be a string")
	}

	// 获取其他可选参数
	var params map[string]interface{}
	if paramsRaw, ok := args["params"]; ok {
		params, _ = paramsRaw.(map[string]interface{})
	}

	// 在实际应用中，这里会执行实际的资源管理
	// 目前返回模拟结果

	if target == "cpu" {
		// CPU资源分配
		return map[string]interface{}{
			"resource": "cpu",
			"action":   "allocation",
			"result": map[string]interface{}{
				"cores_allocated": params["cores"],
				"priority":        params["priority"],
				"message":         "CPU resource allocation updated",
				"status":          "success",
			},
		}, nil
	} else if target == "memory" {
		// 内存资源分配
		return map[string]interface{}{
			"resource": "memory",
			"action":   "allocation",
			"result": map[string]interface{}{
				"memory_allocated": params["memory"],
				"priority":         params["priority"],
				"message":          "Memory resource allocation updated",
				"status":           "success",
			},
		}, nil
	} else if target == "disk" {
		// 磁盘资源分配
		return map[string]interface{}{
			"resource": "disk",
			"action":   "allocation",
			"result": map[string]interface{}{
				"disk_allocated": params["disk"],
				"path":           params["path"],
				"message":        "Disk resource allocation updated",
				"status":         "success",
			},
		}, nil
	} else if target == "network" {
		// 网络资源分配
		return map[string]interface{}{
			"resource": "network",
			"action":   "allocation",
			"result": map[string]interface{}{
				"bandwidth_allocated": params["bandwidth"],
				"interface":           params["interface"],
				"message":             "Network resource allocation updated",
				"status":              "success",
			},
		}, nil
	} else {
		return nil, fmt.Errorf("unknown resource target: %s", target)
	}
}

// queryLog 查询日志
func (t *SystemTool) queryLog(args map[string]interface{}) (interface{}, error) {
	// 获取必要参数
	targetRaw, ok := args["target"]
	if !ok {
		return nil, errors.New("missing required parameter for log: target")
	}

	target, ok := targetRaw.(string)
	if !ok {
		return nil, errors.New("target must be a string")
	}

	// 获取其他可选参数
	var params map[string]interface{}
	if paramsRaw, ok := args["params"]; ok {
		params, _ = paramsRaw.(map[string]interface{})
	}

	// 在实际应用中，这里会查询实际的日志
	// 目前返回模拟结果

	// 模拟不同服务的日志
	var logs []map[string]interface{}

	if target == "system" || target == "kernel" {
		logs = []map[string]interface{}{
			{
				"timestamp": "2023-07-13T08:00:00Z",
				"level":     "INFO",
				"message":   "System boot completed",
				"component": "kernel",
			},
			{
				"timestamp": "2023-07-13T08:01:30Z",
				"level":     "INFO",
				"message":   "All services started",
				"component": "systemd",
			},
			{
				"timestamp": "2023-07-13T09:15:45Z",
				"level":     "WARNING",
				"message":   "High memory usage detected",
				"component": "memory-monitor",
			},
			{
				"timestamp": "2023-07-13T10:30:22Z",
				"level":     "ERROR",
				"message":   "Disk I/O error on /dev/sda2",
				"component": "storage",
			},
			{
				"timestamp": "2023-07-13T11:45:10Z",
				"level":     "INFO",
				"message":   "Scheduled maintenance started",
				"component": "maintenance",
			},
		}
	} else if strings.Contains(target, "api") || strings.Contains(target, "gateway") {
		logs = []map[string]interface{}{
			{
				"timestamp": "2023-07-13T08:30:00Z",
				"level":     "INFO",
				"message":   "API Gateway started on port 8080",
				"component": "api-gateway",
			},
			{
				"timestamp":  "2023-07-13T09:12:34Z",
				"level":      "INFO",
				"message":    "Received request: GET /api/v1/users",
				"component":  "api-gateway",
				"request_id": "req-12345",
			},
			{
				"timestamp":  "2023-07-13T09:15:22Z",
				"level":      "ERROR",
				"message":    "Failed to connect to auth service",
				"component":  "api-gateway",
				"request_id": "req-12346",
			},
			{
				"timestamp": "2023-07-13T10:05:11Z",
				"level":     "INFO",
				"message":   "Rate limit applied for IP 192.168.1.100",
				"component": "api-gateway",
			},
			{
				"timestamp": "2023-07-13T11:30:45Z",
				"level":     "WARNING",
				"message":   "High latency detected in database responses",
				"component": "api-gateway",
			},
		}
	} else if strings.Contains(target, "db") || strings.Contains(target, "database") {
		logs = []map[string]interface{}{
			{
				"timestamp": "2023-07-13T08:15:00Z",
				"level":     "INFO",
				"message":   "Database service started",
				"component": "postgresql",
			},
			{
				"timestamp": "2023-07-13T09:20:35Z",
				"level":     "INFO",
				"message":   "Connection pool initialized with 50 connections",
				"component": "database-manager",
			},
			{
				"timestamp": "2023-07-13T09:45:12Z",
				"level":     "WARNING",
				"message":   "Slow query detected: SELECT * FROM users WHERE...",
				"component": "query-monitor",
				"query_id":  "q-78901",
				"duration":  "2.5s",
			},
			{
				"timestamp": "2023-07-13T10:30:00Z",
				"level":     "INFO",
				"message":   "Scheduled backup started",
				"component": "backup-service",
			},
			{
				"timestamp": "2023-07-13T11:22:33Z",
				"level":     "ERROR",
				"message":   "Failed to write to transaction log: disk space full",
				"component": "postgresql",
			},
		}
	} else {
		logs = []map[string]interface{}{
			{
				"timestamp": "2023-07-13T08:00:00Z",
				"level":     "INFO",
				"message":   fmt.Sprintf("%s service started", target),
				"component": target,
			},
			{
				"timestamp": "2023-07-13T09:30:00Z",
				"level":     "INFO",
				"message":   "Processing request batch #12345",
				"component": target,
			},
			{
				"timestamp": "2023-07-13T10:15:00Z",
				"level":     "WARNING",
				"message":   "Resource usage approaching limits",
				"component": target,
			},
			{
				"timestamp": "2023-07-13T11:45:00Z",
				"level":     "INFO",
				"message":   "Scheduled task completed successfully",
				"component": target,
			},
		}
	}

	// 过滤日志（如果有过滤参数）
	if params != nil {
		if level, ok := params["level"].(string); ok && level != "" {
			filteredLogs := []map[string]interface{}{}
			for _, log := range logs {
				if log["level"].(string) == strings.ToUpper(level) {
					filteredLogs = append(filteredLogs, log)
				}
			}
			logs = filteredLogs
		}
	}

	return map[string]interface{}{
		"target":    target,
		"total":     len(logs),
		"logs":      logs,
		"timestamp": time.Now().Format(time.RFC3339),
	}, nil
}
