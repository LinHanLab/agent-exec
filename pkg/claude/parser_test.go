package claude

import (
	"bytes"
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
				stripped := stripANSI(output)
				if !strings.Contains(stripped, "💬") {
					t.Error("Expected output to contain 💬 emoji")
				}
				if !strings.Contains(stripped, "Hello world") {
					t.Error("Expected output to contain 'Hello world'")
				}
			},
			expectError: false,
		},
		{
			name:           "assistant tool_use without inputs",
			input:          `{"type":"assistant","message":{"content":[{"type":"tool_use","name":"Read"}]}}`,
			expectedResult: "",
			checkOutput: func(t *testing.T, output string) {
				stripped := stripANSI(output)
				if !strings.Contains(stripped, "🔧") {
					t.Error("Expected output to contain 🔧 emoji")
				}
				if !strings.Contains(stripped, "Read") {
					t.Error("Expected output to contain 'Read'")
				}
			},
			expectError: false,
		},
		{
			name:           "assistant tool_use with inputs",
			input:          `{"type":"assistant","message":{"content":[{"type":"tool_use","name":"Read","input":{"file_path":"/test/file.go","limit":10}}]}}`,
			expectedResult: "",
			checkOutput: func(t *testing.T, output string) {
				stripped := stripANSI(output)
				if !strings.Contains(stripped, "🔧") {
					t.Error("Expected output to contain 🔧 emoji")
				}
				if !strings.Contains(stripped, "Read") {
					t.Error("Expected output to contain 'Read'")
				}
				if !strings.Contains(stripped, "/test/file.go") {
					t.Error("Expected output to contain '/test/file.go'")
				}
				if !strings.Contains(stripped, "10") {
					t.Error("Expected output to contain limit value")
				}
			},
			expectError: false,
		},
		{
			name:           "user tool_result with string content",
			input:          `{"type":"user","message":{"content":[{"type":"tool_result","content":"File read successfully"}]}}`,
			expectedResult: "",
			checkOutput: func(t *testing.T, output string) {
				stripped := stripANSI(output)
				if !strings.Contains(stripped, "📋") {
					t.Error("Expected output to contain 📋 emoji")
				}
				if !strings.Contains(stripped, "File read successfully") {
					t.Error("Expected output to contain 'File read successfully'")
				}
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
				stripped := stripANSI(output)
				if !strings.Contains(stripped, "⏱️") {
					t.Error("Expected output to contain ⏱️ emoji")
				}
				if !strings.Contains(stripped, "1.5s") {
					t.Error("Expected output to contain formatted duration '1.5s'")
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
				stripped := stripANSI(output)
				if !strings.Contains(stripped, "💬") {
					t.Error("Expected output to contain 💬 emoji")
				}
				if !strings.Contains(stripped, "Valid") {
					t.Error("Expected output to contain 'Valid'")
				}
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
				stripped := stripANSI(output)
				// Check each message is present
				if !strings.Contains(stripped, "Starting task") {
					t.Error("Expected output to contain 'Starting task'")
				}
				if !strings.Contains(stripped, "Bash") {
					t.Error("Expected output to contain 'Bash'")
				}
				if !strings.Contains(stripped, "total 0") {
					t.Error("Expected output to contain 'total 0'")
				}
				if !strings.Contains(stripped, "⏱️") {
					t.Error("Expected output to contain ⏱️ emoji for execution result")
				}
			},
			expectError: false,
		},
		{
			name:           "tool_result with long content is NOT truncated",
			input:          `{"type":"user","message":{"content":[{"type":"tool_result","content":"` + strings.Repeat("a", 250) + `"}]}}`,
			expectedResult: "",
			checkOutput: func(t *testing.T, output string) {
				stripped := stripANSI(output)
				if !strings.Contains(stripped, "📋") {
					t.Error("Expected output to contain 📋 emoji")
				}
				// Content should not be truncated
				if !strings.Contains(stripped, strings.Repeat("a", 250)) {
					t.Error("Expected output to contain full content (not truncated)")
				}
			},
			expectError: false,
		},
		{
			name:           "tool input with long value is NOT truncated",
			input:          `{"type":"assistant","message":{"content":[{"type":"tool_use","name":"Test","input":{"long_value":"` + strings.Repeat("x", 150) + `"}}]}}`,
			expectedResult: "",
			checkOutput: func(t *testing.T, output string) {
				stripped := stripANSI(output)
				if !strings.Contains(stripped, "🔧") {
					t.Error("Expected output to contain 🔧 emoji")
				}
				if !strings.Contains(stripped, "Test") {
					t.Error("Expected output to contain 'Test'")
				}
				// Input should not be truncated
				if !strings.Contains(stripped, strings.Repeat("x", 150)) {
					t.Error("Expected output to contain full input value (not truncated)")
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
			formatter := display.NewConsoleFormatter(writer, true) // Use verbose mode for tests
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
