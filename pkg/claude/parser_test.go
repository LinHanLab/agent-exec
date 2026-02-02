package claude

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/LinHanLab/agent-exec/pkg/display"
	"github.com/LinHanLab/agent-exec/pkg/events"
)

func TestParseStreamJSON(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedResult string
		checkOutput    func(t *testing.T, output string)
		expectError    bool
	}{
		{
			name:           "empty input",
			input:          "",
			expectedResult: "",
			checkOutput: func(t *testing.T, output string) {
				if output != "" {
					t.Errorf("Expected empty output, got %q", output)
				}
			},
			expectError: false,
		},
		{
			name:           "empty lines only",
			input:          "\n\n\n",
			expectedResult: "",
			checkOutput: func(t *testing.T, output string) {
				if output != "" {
					t.Errorf("Expected empty output, got %q", output)
				}
			},
			expectError: false,
		},
		{
			name:           "assistant text message",
			input:          `{"type":"assistant","message":{"content":[{"type":"text","text":"Hello world"}]}}`,
			expectedResult: "",
			checkOutput: func(t *testing.T, output string) {
				checkJSONOutput(t, output, map[string]interface{}{
					"type": "claude_assistant_message",
					"text": "Hello world",
				})
			},
			expectError: false,
		},
		{
			name:           "assistant tool_use without inputs",
			input:          `{"type":"assistant","message":{"content":[{"type":"tool_use","name":"Read"}]}}`,
			expectedResult: "",
			checkOutput: func(t *testing.T, output string) {
				checkJSONOutput(t, output, map[string]interface{}{
					"type": "claude_tool_use",
					"name": "Read",
				})
			},
			expectError: false,
		},
		{
			name:           "assistant tool_use with inputs",
			input:          `{"type":"assistant","message":{"content":[{"type":"tool_use","name":"Read","input":{"file_path":"/test/file.go","limit":10}}]}}`,
			expectedResult: "",
			checkOutput: func(t *testing.T, output string) {
				data := checkJSONOutput(t, output, map[string]interface{}{
					"type": "claude_tool_use",
					"name": "Read",
				})
				// Check nested input object
				input, ok := data["input"].(map[string]interface{})
				if !ok {
					t.Fatal("Expected input to be a map")
				}
				if input["file_path"] != "/test/file.go" {
					t.Errorf("Expected file_path '/test/file.go', got %v", input["file_path"])
				}
				if input["limit"] != float64(10) {
					t.Errorf("Expected limit 10, got %v", input["limit"])
				}
			},
			expectError: false,
		},
		{
			name:           "user tool_result with string content",
			input:          `{"type":"user","message":{"content":[{"type":"tool_result","content":"File read successfully"}]}}`,
			expectedResult: "",
			checkOutput: func(t *testing.T, output string) {
				checkJSONOutput(t, output, map[string]interface{}{
					"type":    "claude_tool_result",
					"content": "File read successfully",
				})
			},
			expectError: false,
		},
		{
			name:           "user tool_result with empty content",
			input:          `{"type":"user","message":{"content":[{"type":"tool_result","content":""}]}}`,
			expectedResult: "",
			checkOutput: func(t *testing.T, output string) {
				if output != "" {
					t.Errorf("Expected empty output for empty content, got %q", output)
				}
			},
			expectError: false,
		},
		{
			name:           "user tool_result with nil content",
			input:          `{"type":"user","message":{"content":[{"type":"tool_result"}]}}`,
			expectedResult: "",
			checkOutput: func(t *testing.T, output string) {
				if output != "" {
					t.Errorf("Expected empty output for nil content, got %q", output)
				}
			},
			expectError: false,
		},
		{
			name:           "result message with duration",
			input:          `{"type":"result","duration_ms":1500}`,
			expectedResult: "",
			checkOutput: func(t *testing.T, output string) {
				data := checkJSONOutput(t, output, map[string]interface{}{
					"type": "claude_execution_result",
				})
				if data["duration_sec"] != float64(1.5) {
					t.Errorf("Expected duration_sec 1.5, got %v", data["duration_sec"])
				}
			},
			expectError: false,
		},
		{
			name: "malformed JSON line",
			input: `{"type":"assistant","message":{"content":[{"type":"text","text":"Valid"}]}}
{invalid json}`,
			expectedResult: "",
			checkOutput: func(t *testing.T, output string) {
				// Should have one valid line
				checkJSONOutput(t, output, map[string]interface{}{
					"type": "claude_assistant_message",
					"text": "Valid",
				})
			},
			expectError: true,
		},
		{
			name: "multiple messages in sequence",
			input: `{"type":"assistant","message":{"content":[{"type":"text","text":"Starting task"}]}}
{"type":"assistant","message":{"content":[{"type":"tool_use","name":"Bash","input":{"command":"ls -la"}}]}}
{"type":"user","message":{"content":[{"type":"tool_result","content":"total 0"}]}}
{"type":"result","duration_ms":500}`,
			expectedResult: "",
			checkOutput: func(t *testing.T, output string) {
				lines := strings.Split(strings.TrimSpace(output), "\n")
				if len(lines) != 4 {
					t.Errorf("Expected 4 lines of output, got %d", len(lines))
				}
				// Check each line contains expected content
				if !strings.Contains(output, "Starting task") {
					t.Error("Expected output to contain 'Starting task'")
				}
				if !strings.Contains(output, "Bash") {
					t.Error("Expected output to contain 'Bash'")
				}
				if !strings.Contains(output, "total 0") {
					t.Error("Expected output to contain 'total 0'")
				}
				if !strings.Contains(output, "duration_sec") {
					t.Error("Expected output to contain 'duration_sec'")
				}
			},
			expectError: false,
		},
		{
			name:           "tool_result with long content is NOT truncated",
			input:          `{"type":"user","message":{"content":[{"type":"tool_result","content":"` + strings.Repeat("a", 250) + `"}]}}`,
			expectedResult: "",
			checkOutput: func(t *testing.T, output string) {
				// JSON output should NOT truncate content
				data := checkJSONOutput(t, output, map[string]interface{}{
					"type": "claude_tool_result",
				})
				content := data["content"].(string)
				if len(content) != 250 {
					t.Errorf("Expected content length 250, got %d (content should not be truncated)", len(content))
				}
			},
			expectError: false,
		},
		{
			name:           "tool input with long value is NOT truncated",
			input:          `{"type":"assistant","message":{"content":[{"type":"tool_use","name":"Test","input":{"long_value":"` + strings.Repeat("x", 150) + `"}}]}}`,
			expectedResult: "",
			checkOutput: func(t *testing.T, output string) {
				// JSON output should NOT truncate input values
				data := checkJSONOutput(t, output, map[string]interface{}{
					"type": "claude_tool_use",
					"name": "Test",
				})
				input := data["input"].(map[string]interface{})
				longValue := input["long_value"].(string)
				if len(longValue) != 150 {
					t.Errorf("Expected long_value length 150, got %d (input should not be truncated)", len(longValue))
				}
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.input)
			emitter := events.NewChannelEmitter(100)
			writer := &bytes.Buffer{}
			formatter := display.NewConsoleFormatter(writer)
			disp := display.NewDisplay(formatter, emitter)
			disp.Start()

			result, err := ParseStreamJSON(reader, emitter)

			emitter.Close()
			disp.Wait()

			if tt.expectError && err == nil {
				t.Errorf("ParseStreamJSON() expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("ParseStreamJSON() unexpected error: %v", err)
			}

			if result != tt.expectedResult {
				t.Errorf("ParseStreamJSON() result = %q; want %q", result, tt.expectedResult)
			}

			output := writer.String()
			if tt.checkOutput != nil {
				tt.checkOutput(t, output)
			}
		})
	}
}

// checkJSONOutput strips ANSI codes, parses JSON, and checks expected fields
func checkJSONOutput(t *testing.T, output string, expectedFields map[string]interface{}) map[string]interface{} {
	t.Helper()

	// Strip ANSI color codes
	jsonStr := stripANSI(output)

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		t.Fatalf("Failed to parse JSON: %v\nOutput: %s", err, jsonStr)
	}

	// Check expected fields
	for key, expectedValue := range expectedFields {
		actualValue, ok := data[key]
		if !ok {
			t.Errorf("Expected field %q to be present", key)
			continue
		}
		if actualValue != expectedValue {
			t.Errorf("Expected %q to be %v, got %v", key, expectedValue, actualValue)
		}
	}

	// Verify time field is present
	if _, ok := data["time"]; !ok {
		t.Error("Expected time field to be present")
	}

	return data
}

// stripANSI removes ANSI color codes from a string
func stripANSI(s string) string {
	result := ""
	inEscape := false
	for _, r := range s {
		switch {
		case r == '\033':
			inEscape = true
		case inEscape && r == 'm':
			inEscape = false
		case !inEscape:
			result += string(r)
		}
	}
	return strings.TrimSpace(result)
}

func TestContentToString(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{
			name:     "string content",
			input:    "hello world",
			expected: "hello world",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "array of strings",
			input:    []interface{}{"hello", " ", "world"},
			expected: "hello world",
		},
		{
			name:     "array of mixed types",
			input:    []interface{}{"text", 123, true},
			expected: "text123true",
		},
		{
			name:     "empty array",
			input:    []interface{}{},
			expected: "",
		},
		{
			name:     "nil value",
			input:    nil,
			expected: "",
		},
		{
			name:     "number",
			input:    42,
			expected: "",
		},
		{
			name:     "boolean",
			input:    true,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := contentToString(tt.input)
			if result != tt.expected {
				t.Errorf("contentToString() = %q; want %q", result, tt.expected)
			}
		})
	}
}
