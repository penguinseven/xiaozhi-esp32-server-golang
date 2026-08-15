package llm

import "context"

// LLMExtraErrorKey 错误透传约定：ResponseWithContext 失败时在 Message.Extra 中使用的 key
const LLMExtraErrorKey = "error"

// IsLLMErrorMessage 判断是否为 LLM 透传的错误消息（Extra 中含 error）
func IsLLMErrorMessage(msg *Message) bool {
	if msg == nil || msg.Extra == nil {
		return false
	}
	v, ok := msg.Extra[LLMExtraErrorKey]
	if !ok || v == nil {
		return false
	}
	_, ok = v.(string)
	return ok
}

// LLMErrorMessage 从 Message.Extra 中解析出错误文案（若为错误消息）
func LLMErrorMessage(msg *Message) string {
	if msg == nil || msg.Extra == nil {
		return ""
	}
	v, ok := msg.Extra[LLMExtraErrorKey].(string)
	if !ok {
		return ""
	}
	return v
}

// LLMProvider 大语言模型提供者接口
// 使用领域类型（Message/Tool），eino 等具体形状只出现在 eino_llm 边缘 adapter 内。
type LLMProvider interface {
	// ResponseWithContext 带有上下文控制的响应，支持取消操作
	// ctx: 上下文，可用于取消长时间运行的请求
	// sessionID: 会话标识符
	// dialogue: 对话历史（领域消息类型）
	// functions: 工具定义（领域类型）
	ResponseWithContext(ctx context.Context, sessionID string, dialogue []*Message, functions []*Tool) <-chan *Message

	// ResponseWithVllm 视觉（图片识别）路径，直接读取 vision.vllm.* 配置
	ResponseWithVllm(ctx context.Context, file []byte, text string, mimeType string) (string, error)

	// Close 关闭资源，释放连接等
	Close() error
	// IsValid 检查资源是否有效
	IsValid() bool
}
