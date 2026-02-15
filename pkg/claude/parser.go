package claude

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/LinHanLab/agent-exec/pkg/events"
)

// ParseStreamJSON parses streaming JSON output from claude CLI and returns the final result text
func ParseStreamJSON(reader io.Reader, emitter events.Emitter) (string, error) {
	scanner := bufio.NewScanner(reader)
	buf := make([]byte, 0, 1024*1024)
	scanner.Buffer(buf, 10*1024*1024)

	var resultText string

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var msg ClaudeMessage
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			return "", fmt.Errorf("failed to parse JSON output: %w", err)
		}

		switch msg.Type {
		case "assistant":
			for _, content := range msg.Message.Content {
				switch content.Type {
				case "text":
					emitter.Emit(events.EventClaudeAssistantMessage, events.AssistantMessageData{
						Text: content.Text,
					})
				case "tool_use":
					emitter.Emit(events.EventClaudeToolUse, events.ToolUseData{
						Name:  content.Name,
						Input: content.Input,
					})
				}
			}
		case "user":
			for _, content := range msg.Message.Content {
				if content.Type == "tool_result" && content.Content != nil {
					resultStr := contentToString(content.Content)
					if resultStr != "" {
						emitter.Emit(events.EventClaudeToolResult, events.ToolResultData{
							Content: resultStr,
						})
					}
				}
			}
		case "result":
			if msg.Result != "" {
				resultText = msg.Result
			}
			if msg.DurationMs > 0 {
				duration := time.Duration(msg.DurationMs) * time.Millisecond
				emitter.Emit(events.EventClaudeExecutionResult, events.ExecutionResultData{
					Duration: duration,
				})
			}
		}
	}

	return resultText, scanner.Err()
}

// contentToString converts content (string or array) to string
func contentToString(content interface{}) string {
	if str, ok := content.(string); ok {
		return str
	}
	if arr, ok := content.([]interface{}); ok {
		var result string
		for _, item := range arr {
			result += fmt.Sprintf("%v", item)
		}
		return result
	}
	return ""
}
