package claude

// ClaudeMessage represents the main JSON structure from claude CLI
type ClaudeMessage struct {
	Type       string        `json:"type"`
	Message    MessageDetail `json:"message,omitempty"`
	Result     string        `json:"result,omitempty"`
	DurationMs int           `json:"duration_ms,omitempty"`
}

// MessageDetail contains the message content
type MessageDetail struct {
	Content []ContentItem `json:"content,omitempty"`
}

// ContentItem represents individual content parts
type ContentItem struct {
	Type    string                 `json:"type"`
	Text    string                 `json:"text,omitempty"`
	Name    string                 `json:"name,omitempty"`
	Input   map[string]interface{} `json:"input,omitempty"`
	Content string                 `json:"content,omitempty"` // tool_result output
}
