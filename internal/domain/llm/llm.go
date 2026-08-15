package llm

import (
	"context"

	log "xiaozhi-esp32-server-golang/internal/pkg/logger"
)

// ConvertMCPToolsToLLMTools 将MCP工具转换为领域 Tool 列表
func ConvertMCPToolsToLLMTools(ctx context.Context, mcpTools map[string]interface{}) ([]*Tool, error) {
	var tools []*Tool

	for toolName, mcpTool := range mcpTools {
		// 仅需要工具的元信息（Info），由实现方返回领域 Tool
		if invokableTool, ok := mcpTool.(InvokableTool); ok {
			tool, err := invokableTool.Info(ctx)
			if err != nil {
				log.Errorf("获取工具 %s 信息失败: %v", toolName, err)
				continue
			}
			if tool != nil {
				tools = append(tools, tool)
			}
		} else {
			log.Warnf("工具 %s 不支持 Info 接口，跳过转换", toolName)
		}
	}

	log.Infof("成功转换了 %d 个MCP工具为LLM工具", len(tools))
	return tools, nil
}
