package display

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/LinHanLab/agent-exec/pkg/events"
)

func TestConsoleFormatter_RunPromptStarted(t *testing.T) {
	buf := &bytes.Buffer{}
	formatter := NewConsoleFormatter(buf, false)

	event := events.Event{
		Type: events.EventRunPromptStarted,
		Data: events.RunPromptStartedData{
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

	// Verify color code is BoldCyan (no reverse video for this event)
	if !strings.Contains(output, BoldCyan) {
		t.Error("Expected output to contain BoldCyan color code")
	}

	// Verify content is indented with 4 spaces
	lines := strings.Split(stripped, "\n")
	foundIndentedContent := false
	for _, line := range lines {
		if strings.HasPrefix(line, "    ") && strings.Contains(line, "test prompt") {
			foundIndentedContent = true
			break
		}
	}
	if !foundIndentedContent {
		t.Error("Expected prompt content to be indented with 4 spaces")
	}
}

func TestConsoleFormatter_ToolUse(t *testing.T) {
	buf := &bytes.Buffer{}
	formatter := NewConsoleFormatter(buf, false)

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

	// Verify content is indented with 4 spaces
	lines := strings.Split(stripped, "\n")
	foundIndentedContent := false
	for _, line := range lines {
		if strings.HasPrefix(line, "    ") && (strings.Contains(line, "param1") || strings.Contains(line, "param2")) {
			foundIndentedContent = true
			break
		}
	}
	if !foundIndentedContent {
		t.Error("Expected JSON content to be indented with 4 spaces")
	}
}

func TestConsoleFormatter_EvolveStarted(t *testing.T) {
	buf := &bytes.Buffer{}
	formatter := NewConsoleFormatter(buf, false)

	event := events.Event{
		Type: events.EventEvolveStarted,
		Data: events.EvolveStartedData{
			TotalIterations: 3,
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

	// Verify iterations
	if !strings.Contains(stripped, "3") {
		t.Error("Expected output to contain iterations count")
	}

	// Verify color code is BoldYellow with reverse video
	if !strings.Contains(output, BoldYellow) {
		t.Error("Expected output to contain BoldYellow color code")
	}
	if !strings.Contains(output, ReverseVideo) {
		t.Error("Expected output to contain reverse video code")
	}

	// Verify content is indented with 4 spaces
	lines := strings.Split(stripped, "\n")
	foundIndentedContent := false
	for _, line := range lines {
		if strings.HasPrefix(line, "    ") && strings.Contains(line, "Iterations") {
			foundIndentedContent = true
			break
		}
	}
	if !foundIndentedContent {
		t.Error("Expected iteration count to be indented with 4 spaces")
	}
}

func TestConsoleFormatter_ExecutionResult(t *testing.T) {
	buf := &bytes.Buffer{}
	formatter := NewConsoleFormatter(buf, false)

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

	// Verify color code is BoldGreen (no reverse video for this event)
	if !strings.Contains(output, BoldGreen) {
		t.Error("Expected output to contain BoldGreen color code")
	}
}

func TestDisplay_StartAndWait(t *testing.T) {
	buf := &bytes.Buffer{}
	formatter := NewConsoleFormatter(buf, false)
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
	formatter := NewConsoleFormatter(buf, false)
	emitter := events.NewChannelEmitter(10)

	display := NewDisplay(formatter, emitter)
	display.Start()

	emitter.Emit(events.EventRunPromptStarted, events.RunPromptStartedData{
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
