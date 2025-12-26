package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sort"
)

// ANSI color codes
const (
	Bold   = "\033[1m"
	Reset  = "\033[0m"
	Cyan   = "\033[36m"
	Green  = "\033[32m"
)

// JSON structures for parsing claude CLI output
type ClaudeMessage struct {
	Type    string        `json:"type"`
	Message MessageDetail `json:"message,omitempty"`
	Result  string        `json:"result,omitempty"`
}

type MessageDetail struct {
	Content []ContentItem `json:"content,omitempty"`
}

type ContentItem struct {
	Type  string                 `json:"type"`
	Text  string                 `json:"text,omitempty"`
	Name  string                 `json:"name,omitempty"`
	Input map[string]interface{} `json:"input,omitempty"`
}

type ToolResultMessage struct {
	Type   string `json:"type"`
	Result string `json:"result,omitempty"`
}

// Run a single prompt with claude CLI
func runOneShot(prompt string) error {
	if err := validatePrompt(prompt); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	fmt.Println("▐ 🪄PROMPT")
	fmt.Println("▐ ─────────")

	displayPrompt := truncateString(prompt, PromptMaxLen, TruncateSuffix)
	printPrefixed(displayPrompt, "▐ ", DisplayWidth)

	fmt.Println()

	if baseURL := os.Getenv("ANTHROPIC_BASE_URL"); baseURL != "" {
		fmt.Printf("🌐 ANTHROPIC_BASE_URL: %s\n", baseURL)
		fmt.Println()
	}

	fmt.Println("🚀 Starting...")
	fmt.Println()

	cmd := exec.Command("claude", "--verbose", "--output-format", "stream-json", "-p", prompt)
	cmd.Stderr = os.Stderr

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start claude CLI: %w", err)
	}

	if err := parseClaudeOutput(stdout); err != nil {
		cmd.Wait()
		return fmt.Errorf("failed to parse output: %w", err)
	}

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("claude CLI failed: %w", err)
	}

	return nil
}

// Parse streaming JSON output from claude CLI
func parseClaudeOutput(reader io.Reader) error {
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
					fmt.Printf("🔧 %s%s%s", Bold, content.Name, Reset)
					if len(content.Input) > 0 {
						fmt.Println()
						printToolInputs(content.Input)
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
				fmt.Printf("✅ %sResult%s: %s\n", Green, Reset, result)
			}
		case "result":
			if msg.Result != "" {
				fmt.Printf("✅ %s\n", msg.Result)
			}
		}
	}

	return scanner.Err()
}

// Print tool inputs with formatting
func printToolInputs(inputs map[string]interface{}) {
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
		fmt.Printf("     %s%s%s: %s\n", Cyan, key, Reset, valueStr)
	}
}
