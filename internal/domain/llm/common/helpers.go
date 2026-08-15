package common

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"strings"

	"xiaozhi-esp32-server-golang/internal/domain/llm"
)

const (
	defaultStableUserPrefix = "xiaozhi"
	maxStableUserIDLength   = 64
)

// BuildPromptFromDialogue flattens message history into a single text prompt.
func BuildPromptFromDialogue(dialogue []*llm.Message) string {
	if len(dialogue) == 0 {
		return ""
	}

	lines := make([]string, 0, len(dialogue))
	for _, msg := range dialogue {
		if msg == nil {
			continue
		}

		content := strings.TrimSpace(msg.Content)
		if content == "" {
			continue
		}

		lines = append(lines, fmt.Sprintf("%s: %s", formatRole(msg.Role), content))
	}

	return strings.TrimSpace(strings.Join(lines, "\n\n"))
}

// BuildStableUserID creates a deterministic and provider-friendly user id.
func BuildStableUserID(prefix, sessionID string) string {
	safePrefix := sanitizeUserIDPart(prefix)
	if safePrefix == "" {
		safePrefix = defaultStableUserPrefix
	}

	safeSession := sanitizeUserIDPart(sessionID)
	if safeSession == "" {
		safeSession = "anonymous"
	}

	candidate := safePrefix + "_" + safeSession
	if len(candidate) <= maxStableUserIDLength {
		return candidate
	}

	sum := sha1.Sum([]byte(prefix + ":" + sessionID))
	suffix := hex.EncodeToString(sum[:8])

	maxPrefixLen := maxStableUserIDLength - len(suffix) - 1
	if maxPrefixLen <= 0 {
		return suffix
	}
	if len(safePrefix) > maxPrefixLen {
		safePrefix = safePrefix[:maxPrefixLen]
	}
	return safePrefix + "_" + suffix
}

func formatRole(role llm.MessageRole) string {
	switch role {
	case llm.RoleSystem:
		return "System"
	case llm.RoleUser:
		return "User"
	case llm.RoleAssistant:
		return "Assistant"
	case llm.RoleTool:
		return "Tool"
	default:
		return "Message"
	}
}

func sanitizeUserIDPart(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}

	var b strings.Builder
	b.Grow(len(s))

	lastWasSeparator := false
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
			lastWasSeparator = false
		case r >= 'A' && r <= 'Z':
			b.WriteRune(r)
			lastWasSeparator = false
		case r >= '0' && r <= '9':
			b.WriteRune(r)
			lastWasSeparator = false
		case r == '_' || r == '-' || r == '.':
			if !lastWasSeparator {
				b.WriteRune(r)
				lastWasSeparator = true
			}
		default:
			if !lastWasSeparator {
				b.WriteRune('_')
				lastWasSeparator = true
			}
		}
	}

	return strings.Trim(b.String(), "._-")
}
