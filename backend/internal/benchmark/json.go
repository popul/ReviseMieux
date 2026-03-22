package benchmark

import "strings"

// StripMarkdownFences removes ```json ... ``` fences from LLM responses
// that wrap their JSON output in markdown code blocks despite being told not to.
func StripMarkdownFences(raw []byte) []byte {
	s := strings.TrimSpace(string(raw))
	if !strings.HasPrefix(s, "```") {
		return raw
	}
	// Remove opening fence (```json or ```)
	lines := strings.SplitN(s, "\n", 2)
	if len(lines) < 2 {
		return raw
	}
	s = lines[1]
	// Remove closing fence
	if idx := strings.LastIndex(s, "```"); idx >= 0 {
		s = s[:idx]
	}
	return []byte(strings.TrimSpace(s))
}
