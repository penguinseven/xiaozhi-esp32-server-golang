package chat

import (
	"testing"

	"xiaozhi-esp32-server-golang/internal/domain/llm"
)

func TestNormalizeToolCallIDsFillsMissingIDs(t *testing.T) {
	calls := []llm.ToolCall{
		{
			Type: "function",
			Function: llm.ToolCallFunction{
				Name:      "get_eth0_ip",
				Arguments: `{}`,
			},
		},
		{
			ID:   "existing-id",
			Type: "function",
			Function: llm.ToolCallFunction{
				Name:      "get_weather",
				Arguments: `{"city":"Shanghai"}`,
			},
		},
	}

	normalized := normalizeToolCallIDs(calls)
	if normalized[0].ID == "" {
		t.Fatalf("expected missing tool call id to be generated")
	}
	if normalized[1].ID != "existing-id" {
		t.Fatalf("expected existing id to remain unchanged, got %q", normalized[1].ID)
	}
}

func TestNormalizeToolCallIDsUsesStableFallbackForSameCall(t *testing.T) {
	call := llm.ToolCall{
		Type: "function",
		Function: llm.ToolCallFunction{
			Name:      "tell_joke",
			Arguments: `{}`,
		},
	}

	first := normalizeToolCallIDs([]llm.ToolCall{call})
	second := normalizeToolCallIDs([]llm.ToolCall{
		{
			Type: "function",
			Function: llm.ToolCallFunction{
				Name:      "other_tool",
				Arguments: `{"ignored":true}`,
			},
		},
		call,
	})

	if first[0].ID == "" {
		t.Fatalf("expected fallback id to be generated")
	}
	if second[1].ID != first[0].ID {
		t.Fatalf("expected same call to use stable fallback id, got %q and %q", first[0].ID, second[1].ID)
	}
}

func TestEnsureToolCallIDMatchesNormalizeFallback(t *testing.T) {
	call := llm.ToolCall{
		Type: "function",
		Function: llm.ToolCallFunction{
			Name:      "tell_joke",
			Arguments: `{}`,
		},
	}

	normalized := normalizeToolCallIDs([]llm.ToolCall{call})
	executed := ensureToolCallID(call)
	if executed.ID != normalized[0].ID {
		t.Fatalf("expected executor and normalize paths to use same fallback id, got %q and %q", executed.ID, normalized[0].ID)
	}
}
