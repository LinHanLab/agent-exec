package claude

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"sort"

	"github.com/LinHanLab/agent-exec/pkg/format"
)

const maxResultDisplay = 200

// ParseStreamJSON parses streaming JSON output from claude CLI and returns the final result text
func ParseStreamJSON(reader io.Reader) (string, error) {
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
			fmt.Printf("⚠️ JSON parse error: %v\n", err)
			continue
		}

		switch msg.Type {
		case "assistant":
			for _, content := range msg.Message.Content {
				switch content.Type {
				case "text":
					fmt.Printf("💬 %s\n", content.Text)
				case "tool_use":
					fmt.Printf("🔧 %s%s%s", format.Bold, content.Name, format.Reset)
					if len(content.Input) > 0 {
						fmt.Println()
						FormatToolInputs(content.Input)
					} else {
						fmt.Println()
					}
				}
			}
		case "user":
			for _, content := range msg.Message.Content {
				if content.Type == "tool_result" && content.Content != nil {
					resultStr := contentToString(content.Content)
					if resultStr != "" {
						result := truncateResult(resultStr, maxResultDisplay)
						fmt.Printf("✅ %sResult%s: %s\n", format.Green, format.Reset, result)
					}
				}
			}
		case "result":
			if msg.Result != "" {
				resultText = msg.Result
				fmt.Printf("✅ %s\n", msg.Result)
			}
			if msg.DurationMs > 0 {
				durationSec := float64(msg.DurationMs) / 1000.0
				fmt.Printf("⏱️ Duration: %.2fs\n", durationSec)
			}
		}
	}

	return resultText, scanner.Err()
}

// truncateResult truncates a result string to maxLen if needed
func truncateResult(result string, maxLen int) string {
	if len(result) > maxLen {
		return result[:maxLen] + "..."
	}
	return result
}

// FormatToolInputs formats and prints tool inputs with proper alignment
func FormatToolInputs(inputs map[string]interface{}) {
	keys := make([]string, 0, len(inputs))
	for k := range inputs {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		value := inputs[key]
		valueStr := fmt.Sprintf("%v", value)
		if len(valueStr) > 100 {
			valueStr = valueStr[:100] + "..."
		}
		fmt.Printf("     %s%s%s: %s\n", format.Cyan, key, format.Reset, valueStr)
	}
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
