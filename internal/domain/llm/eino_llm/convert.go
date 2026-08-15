package eino_llm

import (
	"encoding/json"
	"strings"

	"github.com/cloudwego/eino/schema"
	"github.com/getkin/kin-openapi/openapi3"

	"xiaozhi-esp32-server-golang/internal/domain/llm"
)

// toEinoMessages 领域消息 → eino schema.Message（MultiContent 在 domain 已扁平化为 Content）
func toEinoMessages(dialogue []*llm.Message) []*schema.Message {
	if len(dialogue) == 0 {
		return nil
	}
	out := make([]*schema.Message, 0, len(dialogue))
	for _, m := range dialogue {
		if m == nil {
			continue
		}
		e := &schema.Message{
			Role:       toEinoRole(m.Role),
			Content:    m.Content,
			Name:       m.Name,
			ToolCallID: m.ToolCallID,
			Extra:      m.Extra,
		}
		if len(m.ToolCalls) > 0 {
			e.ToolCalls = make([]schema.ToolCall, 0, len(m.ToolCalls))
			for _, tc := range m.ToolCalls {
				e.ToolCalls = append(e.ToolCalls, toEinoToolCall(tc))
			}
		}
		out = append(out, e)
	}
	return out
}

// fromEinoMessage eino schema.Message → 领域消息；MultiContent 扁平化为 Content
func fromEinoMessage(m *schema.Message) *llm.Message {
	if m == nil {
		return nil
	}
	out := &llm.Message{
		Role:       fromEinoRole(m.Role),
		Content:    m.Content,
		Name:       m.Name,
		ToolCallID: m.ToolCallID,
		Extra:      m.Extra,
	}
	if out.Content == "" && len(m.MultiContent) > 0 {
		out.Content = flattenMultiContent(m.MultiContent)
	}
	if len(m.ToolCalls) > 0 {
		out.ToolCalls = make([]llm.ToolCall, 0, len(m.ToolCalls))
		for _, tc := range m.ToolCalls {
			out.ToolCalls = append(out.ToolCalls, fromEinoToolCall(tc))
		}
	}
	return out
}

func toEinoRole(r llm.MessageRole) schema.RoleType {
	switch r {
	case llm.RoleSystem:
		return schema.System
	case llm.RoleUser:
		return schema.User
	case llm.RoleAssistant:
		return schema.Assistant
	case llm.RoleTool:
		return schema.Tool
	default:
		return schema.User
	}
}

func fromEinoRole(r schema.RoleType) llm.MessageRole {
	switch r {
	case schema.System:
		return llm.RoleSystem
	case schema.Assistant:
		return llm.RoleAssistant
	case schema.Tool:
		return llm.RoleTool
	default:
		return llm.RoleUser
	}
}

func toEinoToolCall(tc llm.ToolCall) schema.ToolCall {
	return schema.ToolCall{
		ID:    tc.ID,
		Type:  tc.Type,
		Extra: tc.Extra,
		Function: schema.FunctionCall{
			Name:      tc.Function.Name,
			Arguments: tc.Function.Arguments,
		},
	}
}

func fromEinoToolCall(tc schema.ToolCall) llm.ToolCall {
	return llm.ToolCall{
		ID:    tc.ID,
		Type:  tc.Type,
		Extra: tc.Extra,
		Function: llm.ToolCallFunction{
			Name:      tc.Function.Name,
			Arguments: tc.Function.Arguments,
		},
	}
}

// toEinoTools 领域工具 → eino schema.ToolInfo（Params 为 JSON Schema map）
func toEinoTools(tools []*llm.Tool) []*schema.ToolInfo {
	if len(tools) == 0 {
		return nil
	}
	out := make([]*schema.ToolInfo, 0, len(tools))
	for _, t := range tools {
		if t == nil {
			continue
		}
		info := &schema.ToolInfo{
			Name: t.Name,
			Desc: t.Desc,
		}
		if len(t.Params) > 0 {
			if raw, err := json.Marshal(t.Params); err == nil {
				inputSchema := &openapi3.Schema{}
				if err := json.Unmarshal(raw, inputSchema); err == nil {
					info.ParamsOneOf = schema.NewParamsOneOfByOpenAPIV3(inputSchema)
				}
			}
		}
		out = append(out, info)
	}
	return out
}

func flattenMultiContent(parts []schema.ChatMessagePart) string {
	var sb strings.Builder
	for _, p := range parts {
		text := strings.TrimSpace(p.Text)
		if text == "" {
			continue
		}
		if sb.Len() > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(text)
	}
	return sb.String()
}
