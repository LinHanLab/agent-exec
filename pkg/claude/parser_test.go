package claude

import (
	"bytes"
	"strings"
	"testing"
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
			name:           "result message with text and duration",
			input:          `{"type":"result","result":"Task completed","duration_ms":1500}`,
			expectedResult: "Task completed",
			expectedOutput: "✅ Task completed\n⏱️ Duration: 1.50s\n",
			expectError:    false,
		},
		{
			name:           "result message with only duration",
			input:          `{"type":"result","duration_ms":2500}`,
			expectedResult: "",
			expectedOutput: "⏱️ Duration: 2.50s\n",
			expectError:    false,
		},
		{
			name:           "result message with only text",
			input:          `{"type":"result","result":"Done"}`,
			expectedResult: "Done",
			expectedOutput: "✅ Done\n",
			expectError:    false,
		},
		{
			name: "malformed JSON line",
			input: `{"type":"assistant","message":{"content":[{"type":"text","text":"Valid"}]}}
{invalid json}
{"type":"result","result":"Still works"}`,
			expectedResult: "Still works",
			expectedOutput: "💬 Valid\n⚠️ JSON parse error: invalid character 'i' looking for beginning of object key string\n✅ Still works\n",
			expectError:    false,
		},
		{
			name: "multiple messages in sequence",
			input: `{"type":"assistant","message":{"content":[{"type":"text","text":"Starting task"}]}}
{"type":"assistant","message":{"content":[{"type":"tool_use","name":"Bash","input":{"command":"ls -la"}}]}}
{"type":"user","message":{"content":[{"type":"tool_result","content":"total 0"}]}}
{"type":"result","result":"Complete","duration_ms":500}`,
			expectedResult: "Complete",
			expectedOutput: "💬 Starting task\n🔧 \x1b[1mBash\x1b[0m\n     \x1b[36mcommand\x1b[0m: ls -la\n✅ \x1b[32mResult\x1b[0m: total 0\n✅ Complete\n⏱️ Duration: 0.50s\n",
			expectError:    false,
		},
		{
			name:           "assistant with multiple content items",
			input:          `{"type":"assistant","message":{"content":[{"type":"text","text":"First message"},{"type":"text","text":"Second message"},{"type":"tool_use","name":"Read"}]}}`,
			expectedResult: "",
			expectedOutput: "💬 First message\n💬 Second message\n🔧 \x1b[1mRead\x1b[0m\n",
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
			writer := &bytes.Buffer{}

			result, err := ParseStreamJSON(reader, writer)

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

func TestTruncateResult(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{
			name:     "string shorter than max",
			input:    "short",
			maxLen:   10,
			expected: "short",
		},
		{
			name:     "string equal to max",
			input:    "exactly10c",
			maxLen:   10,
			expected: "exactly10c",
		},
		{
			name:     "string longer than max",
			input:    "this is a long string",
			maxLen:   10,
			expected: "this is a ...",
		},
		{
			name:     "empty string",
			input:    "",
			maxLen:   10,
			expected: "",
		},
		{
			name:     "max length zero",
			input:    "text",
			maxLen:   0,
			expected: "...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateResult(tt.input, tt.maxLen)
			if result != tt.expected {
				t.Errorf("truncateResult(%q, %d) = %q; want %q", tt.input, tt.maxLen, result, tt.expected)
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

func TestFormatToolInputs(t *testing.T) {
	tests := []struct {
		name     string
		inputs   map[string]interface{}
		expected string
	}{
		{
			name:     "empty inputs",
			inputs:   map[string]interface{}{},
			expected: "",
		},
		{
			name: "single input",
			inputs: map[string]interface{}{
				"file_path": "/test/file.go",
			},
			expected: "     \x1b[36mfile_path\x1b[0m: /test/file.go\n",
		},
		{
			name: "multiple inputs sorted alphabetically",
			inputs: map[string]interface{}{
				"zebra":  "last",
				"alpha":  "first",
				"middle": "second",
			},
			expected: "     \x1b[36malpha\x1b[0m: first\n     \x1b[36mmiddle\x1b[0m: second\n     \x1b[36mzebra\x1b[0m: last\n",
		},
		{
			name: "input with long value gets truncated",
			inputs: map[string]interface{}{
				"long": strings.Repeat("x", 150),
			},
			expected: "     \x1b[36mlong\x1b[0m: " + strings.Repeat("x", 100) + "...\n",
		},
		{
			name: "input with various types",
			inputs: map[string]interface{}{
				"string":  "text",
				"number":  42,
				"boolean": true,
			},
			expected: "     \x1b[36mboolean\x1b[0m: true\n     \x1b[36mnumber\x1b[0m: 42\n     \x1b[36mstring\x1b[0m: text\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writer := &bytes.Buffer{}
			FormatToolInputs(writer, tt.inputs)
			result := writer.String()
			if result != tt.expected {
				t.Errorf("FormatToolInputs() = %q; want %q", result, tt.expected)
			}
		})
	}
}
