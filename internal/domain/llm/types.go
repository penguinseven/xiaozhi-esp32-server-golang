package llm

import "context"

// MessageRole 消息角色（领域类型，与 eino schema.RoleType 对应）
type MessageRole string

const (
	RoleAssistant MessageRole = "assistant"
	RoleUser      MessageRole = "user"
	RoleSystem    MessageRole = "system"
	RoleTool      MessageRole = "tool"
)

// ToolCallFunction 工具调用中的函数信息
type ToolCallFunction struct {
	Name      string
	Arguments string
}

// ToolCall 工具调用（领域类型，对应 eino schema.ToolCall）
type ToolCall struct {
	ID       string
	Type     string
	Function ToolCallFunction
	Extra    map[string]any
}

// Message 对话消息（领域类型，对应 eino schema.Message）
// 仅包含消费端实际使用的字段；MultiContent 在转换层扁平化为 Content。
type Message struct {
	Role       MessageRole
	Content    string
	Name       string
	ToolCalls  []ToolCall
	ToolCallID string
	Extra      map[string]any
}

// Tool 工具定义（领域类型，对应 eino schema.ToolInfo）
// Params 以 JSON Schema（map）形式描述参数；转换层负责与 eino ParamsOneOf 互转。
type Tool struct {
	Name   string
	Desc   string
	Params map[string]any
}

// LLMResponse 流式响应片段（领域类型）
type LLMResponse struct {
	Text      string
	IsStart   bool
	IsEnd     bool
	ToolCalls []ToolCall
}

// InvokableTool 工具执行接口（领域类型，替代 eino components/tool.InvokableTool）。
// Info 返回领域 Tool 元数据；InvokableRun 以 JSON 参数串调用工具。
type InvokableTool interface {
	Info(ctx context.Context) (*Tool, error)
	InvokableRun(ctx context.Context, argumentsInJSON string) (string, error)
}
