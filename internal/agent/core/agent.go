package core

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"distributedJob/internal/agent/types"
	"distributedJob/internal/mcp/client"
	"distributedJob/internal/mcp/protocol"
)

// Agent 智能代理结构体
type Agent struct {
	id           string
	name         string
	description  string
	model        string
	tools        map[string]types.Tool
	memory       types.MemoryManager
	mcpClient    client.Client
	systemPrompt string

	mu sync.Mutex
}

// AgentConfig 代理配置
type AgentConfig struct {
	Name         string             `json:"name"`
	Description  string             `json:"description"`
	Model        string             `json:"model"`
	SystemPrompt string             `json:"system_prompt"`
	Tools        []string           `json:"tools"`
	Memory       types.MemoryConfig `json:"memory"`
	MCPConfig    client.Config      `json:"mcp_config"`
}

// NewAgent 创建一个新的智能代理
func NewAgent(config AgentConfig, mcpClient client.Client, tools []types.Tool) (*Agent, error) {
	// 创建记忆管理器
	var memory types.MemoryManager

	switch config.Memory.Type {
	case types.BufferMemory:
		memory = NewBufferMemory(config.Memory.Capacity)
	case types.VectorMemory:
		// 向量记忆在实际应用中需要实现
		memory = NewBufferMemory(config.Memory.Capacity) // 暂时使用缓冲区记忆
	default:
		memory = NewBufferMemory(100) // 默认使用缓冲区记忆
	}

	// 如果未提供MCP客户端，则根据配置创建一个
	if mcpClient == nil {
		mcpClient = client.NewClient(config.MCPConfig)
	}

	// 将工具映射到名称
	toolMap := make(map[string]types.Tool)
	for _, tool := range tools {
		toolMap[tool.Name()] = tool
	}

	return &Agent{
		id:           uuid.New().String(),
		name:         config.Name,
		description:  config.Description,
		model:        config.Model,
		tools:        toolMap,
		memory:       memory,
		mcpClient:    mcpClient,
		systemPrompt: config.SystemPrompt,
	}, nil
}

// Process 处理用户输入，生成响应
func (a *Agent) Process(ctx context.Context, input string) (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	// 记录用户输入
	a.memory.AddMemory(types.MemoryItem{
		Key:   "user_input_" + time.Now().Format(time.RFC3339),
		Value: input,
		Metadata: map[string]interface{}{
			"type": "user_input",
		},
	})

	// 执行思考-计划-行动循环
	plan, err := a.createPlan(ctx, input)
	if err != nil {
		return "", fmt.Errorf("failed to create plan: %w", err)
	}

	// 记录计划
	a.memory.AddMemory(types.MemoryItem{
		Key:   "plan_" + time.Now().Format(time.RFC3339),
		Value: plan,
		Metadata: map[string]interface{}{
			"type": "plan",
		},
	})

	// 执行计划
	result, err := a.executePlan(ctx, plan)
	if err != nil {
		return "", fmt.Errorf("failed to execute plan: %w", err)
	}

	// 记录结果
	a.memory.AddMemory(types.MemoryItem{
		Key:   "result_" + time.Now().Format(time.RFC3339),
		Value: result,
		Metadata: map[string]interface{}{
			"type": "result",
		},
	})

	return result, nil
}

// createPlan 根据用户输入创建执行计划
func (a *Agent) createPlan(ctx context.Context, input string) (*types.Plan, error) {
	// 获取最近的记忆，以提供上下文
	recentMemories := a.memory.GetRecentMemories(5)

	// 构建用于规划的提示
	planningPrompt := a.buildPlanningPrompt(input, recentMemories)

	// 调用MCP生成计划
	response, err := a.mcpClient.Chat(ctx, &protocol.ChatRequest{
		Messages: []protocol.Message{
			{Role: "system", Content: a.systemPrompt + "\n" + a.getToolDescriptions()},
			{Role: "user", Content: planningPrompt},
		},
	})

	if err != nil {
		return nil, fmt.Errorf("failed to generate plan: %w", err)
	}

	// 解析响应中的计划
	plan, err := a.parsePlanFromResponse(response.Content)
	if err != nil {
		// 如果解析失败，返回简单计划
		return &types.Plan{
			Goal:        "回应用户问题",
			Description: "直接回应用户的问题",
			Steps: []types.PlanStep{
				{
					ID:          uuid.New().String(),
					Description: "生成回应",
					Tool:        "direct_response",
					Parameters: map[string]interface{}{
						"query": input,
					},
					Completed: false,
				},
			},
		}, nil
	}

	return plan, nil
}

// executePlan 执行计划，处理每个步骤
func (a *Agent) executePlan(ctx context.Context, plan *types.Plan) (string, error) {
	var results []string

	// 遍历并执行每个步骤
	for i := range plan.Steps {
		step := &plan.Steps[i]

		// 检查是否存在指定的工具
		tool, exists := a.tools[step.Tool]
		if !exists {
			// 如果工具不存在，记录错误并继续
			step.Result = &types.ToolResult{
				Success: false,
				Error:   fmt.Sprintf("Tool '%s' not found", step.Tool),
				ToolID:  step.Tool,
			}
			continue
		}

		// 执行工具
		result, err := tool.Execute(step.Parameters)
		if err != nil {
			step.Result = &types.ToolResult{
				Success: false,
				Error:   err.Error(),
				ToolID:  step.Tool,
			}
		} else {
			step.Result = &types.ToolResult{
				Success: true,
				Result:  result,
				ToolID:  step.Tool,
			}
		}

		// 标记为已完成
		step.Completed = true

		// 对于成功的结果，添加到结果列表
		if step.Result.Success {
			resultStr := fmt.Sprint(step.Result.Result)
			results = append(results, resultStr)
		}
	}

	// 根据计划执行结果生成最终响应
	finalResponse, err := a.generateFinalResponse(ctx, plan, plan.Goal, results)
	if err != nil {
		return "抱歉，我无法完成您的请求。", err
	}

	return finalResponse, nil
}

// generateFinalResponse 根据计划执行结果生成最终响应
func (a *Agent) generateFinalResponse(ctx context.Context, plan *types.Plan, input string, results []string) (string, error) {
	// 构建提示
	var prompt strings.Builder
	prompt.WriteString("根据以下执行计划和结果，生成给用户的最终响应：\n\n")
	prompt.WriteString("用户输入：" + input + "\n\n")

	prompt.WriteString("计划：\n")
	prompt.WriteString("目标：" + plan.Goal + "\n")
	prompt.WriteString("描述：" + plan.Description + "\n\n")

	prompt.WriteString("步骤结果：\n")
	for i, step := range plan.Steps {
		prompt.WriteString(fmt.Sprintf("步骤 %d: %s (工具: %s)\n", i+1, step.Description, step.Tool))
		if step.Result != nil {
			if step.Result.Success {
				prompt.WriteString("结果：" + fmt.Sprint(step.Result.Result) + "\n")
			} else {
				prompt.WriteString("错误：" + step.Result.Error + "\n")
			}
		} else {
			prompt.WriteString("未执行\n")
		}
		prompt.WriteString("\n")
	}

	prompt.WriteString("请生成用户友好的最终响应，总结执行结果，不要提及计划或步骤细节。")

	// 调用MCP生成最终响应
	response, err := a.mcpClient.Chat(ctx, &protocol.ChatRequest{
		Messages: []protocol.Message{
			{Role: "system", Content: a.systemPrompt},
			{Role: "user", Content: prompt.String()},
		},
	})

	if err != nil {
		return "", fmt.Errorf("failed to generate final response: %w", err)
	}

	return response.Content, nil
}

// buildPlanningPrompt 构建用于规划的提示
func (a *Agent) buildPlanningPrompt(input string, memories []types.MemoryItem) string {
	var prompt strings.Builder

	prompt.WriteString("请为以下用户请求创建一个执行计划。\n\n")
	prompt.WriteString("用户请求：" + input + "\n\n")

	// 添加记忆上下文
	if len(memories) > 0 {
		prompt.WriteString("以下是之前的上下文：\n")
		for _, memory := range memories {
			if value, ok := memory.Value.(string); ok {
				prompt.WriteString("- " + value + "\n")
			}
		}
		prompt.WriteString("\n")
	}

	// 添加可用工具列表
	prompt.WriteString("你可以使用以下工具：\n")
	for name, tool := range a.tools {
		prompt.WriteString("- " + name + ": " + tool.Description() + "\n")

		// 添加参数信息
		params := tool.Parameters()
		if len(params) > 0 {
			prompt.WriteString("  参数：\n")
			for pName, param := range params {
				reqStr := ""
				if param.Required {
					reqStr = "（必须）"
				}
				prompt.WriteString(fmt.Sprintf("  - %s: %s %s\n", pName, param.Description, reqStr))
			}
		}
		prompt.WriteString("\n")
	}

	// 添加计划格式指南
	prompt.WriteString(`
请以以下JSON格式返回计划：
{
  "goal": "计划目标",
  "description": "计划描述",
  "steps": [
    {
      "id": "步骤ID",
      "description": "步骤描述",
      "tool": "工具名称",
      "parameters": {
        "参数名1": "参数值1",
        "参数名2": "参数值2"
      }
    },
    ...
  ]
}

确保每个步骤使用正确的工具名称并提供所有必需参数。
`)

	return prompt.String()
}

// parsePlanFromResponse 从响应文本中解析计划
func (a *Agent) parsePlanFromResponse(response string) (*types.Plan, error) {
	// 提取JSON部分
	jsonStr := extractJSON(response)
	if jsonStr == "" {
		return nil, fmt.Errorf("no valid JSON plan found in response")
	}

	// 解析JSON
	var plan types.Plan
	if err := json.Unmarshal([]byte(jsonStr), &plan); err != nil {
		return nil, fmt.Errorf("failed to parse plan: %w", err)
	}

	// 验证计划
	if len(plan.Steps) == 0 {
		return nil, fmt.Errorf("plan has no steps")
	}

	// 为没有ID的步骤生成ID
	for i := range plan.Steps {
		if plan.Steps[i].ID == "" {
			plan.Steps[i].ID = uuid.New().String()
		}
	}

	return &plan, nil
}

// getToolDescriptions 获取所有工具的描述
func (a *Agent) getToolDescriptions() string {
	var desc strings.Builder

	desc.WriteString("可用工具：\n")
	for name, tool := range a.tools {
		desc.WriteString("- " + name + ": " + tool.Description() + "\n")

		// 添加参数信息
		params := tool.Parameters()
		if len(params) > 0 {
			desc.WriteString("  参数：\n")
			for pName, param := range params {
				reqStr := ""
				if param.Required {
					reqStr = "（必须）"
				}
				desc.WriteString(fmt.Sprintf("  - %s: %s %s\n", pName, param.Description, reqStr))
			}
		}
		desc.WriteString("\n")
	}

	return desc.String()
}

// extractJSON 从文本中提取第一个JSON对象
func extractJSON(text string) string {
	// 查找第一个左大括号
	start := strings.Index(text, "{")
	if start == -1 {
		return ""
	}

	// 查找匹配的右大括号
	depth := 1
	for i := start + 1; i < len(text); i++ {
		switch text[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return text[start : i+1]
			}
		}
	}

	return ""
}

// GetID 获取代理ID
func (a *Agent) GetID() string {
	return a.id
}

// GetName 获取代理名称
func (a *Agent) GetName() string {
	return a.name
}

// GetDescription 获取代理描述
func (a *Agent) GetDescription() string {
	return a.description
}

// GetModel 获取代理模型
func (a *Agent) GetModel() string {
	return a.model
}

// GetTools 获取代理工具列表
func (a *Agent) GetTools() []string {
	tools := make([]string, 0, len(a.tools))
	for name := range a.tools {
		tools = append(tools, name)
	}
	return tools
}
