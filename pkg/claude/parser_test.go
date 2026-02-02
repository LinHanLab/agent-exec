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
		expectedOutput string
		expectError    bool
	}{
		{
			name:           "empty input",
			input:          "",
			expectedResult: "",
			expectedOutput: "",
			expectError:    false,
		},
		{
			name:           "empty lines only",
			input:          "\n\n\n",
			expectedResult: "",
			expectedOutput: "",
			expectError:    false,
		},
		{
			name:           "assistant text message",
			input:          `{"type":"assistant","message":{"content":[{"type":"text","text":"Hello world"}]}}`,
			expectedResult: "",
			expectedOutput: "💬 Hello world\n",
			expectError:    false,
		},
		{
			name:           "assistant tool_use without inputs",
			input:          `{"type":"assistant","message":{"content":[{"type":"tool_use","name":"Read"}]}}`,
			expectedResult: "",
			expectedOutput: "🔧 \x1b[1mRead\x1b[0m\n",
			expectError:    false,
		},
		{
			name:           "assistant tool_use with inputs",
			input:          `{"type":"assistant","message":{"content":[{"type":"tool_use","name":"Read","input":{"file_path":"/test/file.go","limit":10}}]}}`,
			expectedResult: "",
			expectedOutput: "🔧 \x1b[1mRead\x1b[0m\n     \x1b[36mfile_path\x1b[0m: /test/file.go\n     \x1b[36mlimit\x1b[0m: 10\n",
			expectError:    false,
		},
		{
			name:           "user tool_result with string content",
			input:          `{"type":"user","message":{"content":[{"type":"tool_result","content":"File read successfully"}]}}`,
			expectedResult: "",
			expectedOutput: "✅ \x1b[32mResult\x1b[0m: File read successfully\n",
			expectError:    false,
		},
		{
			name:           "user tool_result with empty content",
			input:          `{"type":"user","message":{"content":[{"type":"tool_result","content":""}]}}`,
			expectedResult: "",
			expectedOutput: "",
			expectError:    false,
		},
		{
			name:           "user tool_result with nil content",
			input:          `{"type":"user","message":{"content":[{"type":"tool_result"}]}}`,
			expectedResult: "",
			expectedOutput: "",
			expectError:    false,
		},
		{
			name:           "result message with duration",
			input:          `{"type":"result","duration_ms":1500}`,
			expectedResult: "",
			expectedOutput: "⏱️ Duration: 1.50s\n",
			expectError:    false,
		},
		{
			name: "malformed JSON line",
			input: `{"type":"assistant","message":{"content":[{"type":"text","text":"Valid"}]}}
{invalid json}`,
			expectedResult: "",
			expectedOutput: "💬 Valid\n",
			expectError:    true,
		},
		{
			name: "multiple messages in sequence",
			input: `{"type":"assistant","message":{"content":[{"type":"text","text":"Starting task"}]}}
{"type":"assistant","message":{"content":[{"type":"tool_use","name":"Bash","input":{"command":"ls -la"}}]}}
{"type":"user","message":{"content":[{"type":"tool_result","content":"total 0"}]}}
{"type":"result","duration_ms":500}`,
			expectedResult: "",
			expectedOutput: "💬 Starting task\n🔧 \x1b[1mBash\x1b[0m\n     \x1b[36mcommand\x1b[0m: ls -la\n✅ \x1b[32mResult\x1b[0m: total 0\n⏱️ Duration: 0.50s\n",
			expectError:    false,
		},
		{
			name:           "tool_result with long content gets truncated",
			input:          `{"type":"user","message":{"content":[{"type":"tool_result","content":"` + strings.Repeat("a", 250) + `"}]}}`,
			expectedResult: "",
			expectedOutput: "✅ \x1b[32mResult\x1b[0m: " + strings.Repeat("a", 200) + "...\n",
			expectError:    false,
		},
		{
			name:           "tool input with long value gets truncated",
			input:          `{"type":"assistant","message":{"content":[{"type":"tool_use","name":"Test","input":{"long_value":"` + strings.Repeat("x", 150) + `"}}]}}`,
			expectedResult: "",
			expectedOutput: "🔧 \x1b[1mTest\x1b[0m\n     \x1b[36mlong_value\x1b[0m: " + strings.Repeat("x", 100) + "...\n",
			expectError:    false,
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
			if output != tt.expectedOutput {
				t.Errorf("ParseStreamJSON() output = %q; want %q", output, tt.expectedOutput)
			}
		})
	}
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
				t.Errorf("contentToString(%v) = %q; want %q", tt.input, result, tt.expected)
			}
		})
	}
}
