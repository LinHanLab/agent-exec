package display

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/LinHanLab/agent-exec/pkg/events"
)

func TestConsoleFormatter_PromptStarted(t *testing.T) {
	buf := &bytes.Buffer{}
	formatter := NewConsoleFormatter(buf)

	event := events.Event{
		Type: events.EventPromptStarted,
		Data: events.PromptStartedData{
			Prompt:  "test prompt",
			Cwd:     "/test/dir",
			BaseURL: "https://api.anthropic.com",
		},
	}

	err := formatter.Format(event)
	if err != nil {
		t.Errorf("Format failed: %v", err)
	}

	output := buf.String()
	stripped := stripANSI(output)

	// Verify emoji is present
	if !strings.Contains(stripped, "🚀") {
		t.Error("Expected output to contain 🚀 emoji")
	}

	// Verify title formatting with dashes
	if !strings.Contains(stripped, "---") {
		t.Error("Expected output to contain --- separator")
	}

	// Verify prompt is in code block
	if !strings.Contains(stripped, "```") {
		t.Error("Expected output to contain code block markers")
	}

	// Verify prompt content
	if !strings.Contains(stripped, "test prompt") {
		t.Error("Expected output to contain prompt text")
	}

	// Verify optional metadata
	if !strings.Contains(stripped, "/test/dir") {
		t.Error("Expected output to contain working directory")
	}

	if !strings.Contains(stripped, "https://api.anthropic.com") {
		t.Error("Expected output to contain base URL")
	}

	// Verify color code is BoldCyan
	if !strings.HasPrefix(output, BoldCyan) {
		t.Error("Expected output to start with BoldCyan color code")
	}
}

func TestConsoleFormatter_ToolUse(t *testing.T) {
	buf := &bytes.Buffer{}
	formatter := NewConsoleFormatter(buf)

	event := events.Event{
		Type: events.EventClaudeToolUse,
		Data: events.ToolUseData{
			Name: "TestTool",
			Input: map[string]interface{}{
				"param1": "value1",
				"param2": float64(42),
			},
		},
	}

	err := formatter.Format(event)
	if err != nil {
		t.Errorf("Format failed: %v", err)
	}

	output := buf.String()
	stripped := stripANSI(output)

	// Verify emoji is present
	if !strings.Contains(stripped, "🔧") {
		t.Error("Expected output to contain 🔧 emoji")
	}

	// Verify tool name
	if !strings.Contains(stripped, "TestTool") {
		t.Error("Expected output to contain tool name")
	}

	// Verify code block with json language
	if !strings.Contains(stripped, "```json") {
		t.Error("Expected output to contain ```json code block")
	}

	// Verify input parameters are present
	if !strings.Contains(stripped, "param1") {
		t.Error("Expected output to contain param1")
	}

	if !strings.Contains(stripped, "value1") {
		t.Error("Expected output to contain value1")
	}

	if !strings.Contains(stripped, "param2") {
		t.Error("Expected output to contain param2")
	}

	// Verify no color (empty string for tool use)
	if strings.HasPrefix(output, BoldCyan) || strings.HasPrefix(output, BoldYellow) || strings.HasPrefix(output, Magenta) {
		t.Error("Expected no color code for tool use")
	}
}

func TestConsoleFormatter_EvolveStarted(t *testing.T) {
	buf := &bytes.Buffer{}
	formatter := NewConsoleFormatter(buf)

	event := events.Event{
		Type: events.EventEvolveStarted,
		Data: events.EvolveStartedData{
			Prompt:     "test evolution",
			Iterations: 3,
		},
	}

	err := formatter.Format(event)
	if err != nil {
		t.Errorf("Format failed: %v", err)
	}

	output := buf.String()
	stripped := stripANSI(output)

	// Verify emoji is present
	if !strings.Contains(stripped, "🧬") {
		t.Error("Expected output to contain 🧬 emoji")
	}

	// Verify big title formatting with equals
	if !strings.Contains(stripped, "===") {
		t.Error("Expected output to contain === separator for big title")
	}

	// Verify prompt content
	if !strings.Contains(stripped, "test evolution") {
		t.Error("Expected output to contain prompt text")
	}

	// Verify iterations
	if !strings.Contains(stripped, "3") {
		t.Error("Expected output to contain iterations count")
	}

	// Verify color code is BoldYellow
	if !strings.HasPrefix(output, BoldYellow) {
		t.Error("Expected output to start with BoldYellow color code")
	}
}

func TestConsoleFormatter_ExecutionResult(t *testing.T) {
	buf := &bytes.Buffer{}
	formatter := NewConsoleFormatter(buf)

	event := events.Event{
		Type: events.EventClaudeExecutionResult,
		Data: events.ExecutionResultData{
			Duration: 2 * time.Second,
		},
	}

	err := formatter.Format(event)
	if err != nil {
		t.Errorf("Format failed: %v", err)
	}

	output := buf.String()
	stripped := stripANSI(output)

	// Verify emoji is present
	if !strings.Contains(stripped, "⏱️") {
		t.Error("Expected output to contain ⏱️ emoji")
	}

	// Verify duration is formatted
	if !strings.Contains(stripped, "2.0s") {
		t.Error("Expected output to contain formatted duration")
	}

	// Verify color code is BoldGreen
	if !strings.HasPrefix(output, BoldGreen) {
		t.Error("Expected output to start with BoldGreen color code")
	}
}

func TestDisplay_StartAndWait(t *testing.T) {
	buf := &bytes.Buffer{}
	formatter := NewConsoleFormatter(buf)
	emitter := events.NewChannelEmitter(10)

	display := NewDisplay(formatter, emitter)
	display.Start()

	emitter.Emit(events.EventClaudeAssistantMessage, events.AssistantMessageData{
		Text: "test message",
	})

	emitter.Close()
	display.Wait()

	output := buf.String()
	stripped := stripANSI(output)

	// Verify message is present
	if !strings.Contains(stripped, "test message") {
		t.Error("Expected output to contain emitted message")
	}

	// Verify emoji is present
	if !strings.Contains(stripped, "💬") {
		t.Error("Expected output to contain 💬 emoji")
	}
}

func TestDisplay_MultipleEvents(t *testing.T) {
	buf := &bytes.Buffer{}
	formatter := NewConsoleFormatter(buf)
	emitter := events.NewChannelEmitter(10)

	display := NewDisplay(formatter, emitter)
	display.Start()

	emitter.Emit(events.EventPromptStarted, events.PromptStartedData{
		Prompt: "test",
	})
	emitter.Emit(events.EventClaudeAssistantMessage, events.AssistantMessageData{
		Text: "response",
	})
	emitter.Emit(events.EventClaudeExecutionResult, events.ExecutionResultData{
		Duration: 2 * time.Second,
	})

	emitter.Close()
	display.Wait()

	output := buf.String()
	stripped := stripANSI(output)

	// Verify all three events are present
	if !strings.Contains(stripped, "🚀") {
		t.Error("Expected output to contain 🚀 emoji for prompt started")
	}
	if !strings.Contains(stripped, "response") {
		t.Error("Expected output to contain response text")
	}
	if !strings.Contains(stripped, "⏱️") {
		t.Error("Expected output to contain ⏱️ emoji for execution result")
	}

	// Verify multiple lines of output
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) < 3 {
		t.Errorf("Expected at least 3 lines of output, got %d", len(lines))
	}
}

// stripANSI removes ANSI color codes from a string
func stripANSI(s string) string {
	// Remove ANSI escape sequences
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
