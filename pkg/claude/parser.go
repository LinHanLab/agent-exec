package claude

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"sort"

	"github.com/LinHanLab/agent-exec/pkg/format"
)

// ParseStreamJSON parses streaming JSON output from claude CLI
func ParseStreamJSON(reader io.Reader) error {
	scanner := bufio.NewScanner(reader)
	buf := make([]byte, 0, 1024*1024)
	scanner.Buffer(buf, 10*1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var msg ClaudeMessage
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
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
		case "tool_result":
			var toolResult ToolResultMessage
			if err := json.Unmarshal([]byte(line), &toolResult); err == nil && toolResult.Result != "" {
				result := toolResult.Result
				if len(result) > 200 {
					result = result[:200] + "..."
				}
				fmt.Printf("✅ %sResult%s: %s\n", format.Green, format.Reset, result)
			}
		case "result":
			if msg.Result != "" {
				fmt.Printf("✅ %s\n", msg.Result)
			}
		}
	}

	return scanner.Err()
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
