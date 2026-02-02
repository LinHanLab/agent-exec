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
			Prompt: "test prompt",
		},
	}

	err := formatter.Format(event)
	if err != nil {
		t.Errorf("Format failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "🪄PROMPT") {
		t.Error("Expected output to contain prompt header")
	}
	if !strings.Contains(output, "test prompt") {
		t.Error("Expected output to contain prompt text")
	}
}

func TestConsoleFormatter_AssistantMessage(t *testing.T) {
	buf := &bytes.Buffer{}
	formatter := NewConsoleFormatter(buf)

	event := events.Event{
		Type: events.EventClaudeAssistantMessage,
		Data: events.AssistantMessageData{
			Text: "Hello, world!",
		},
	}

	err := formatter.Format(event)
	if err != nil {
		t.Errorf("Format failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "💬") {
		t.Error("Expected output to contain message emoji")
	}
	if !strings.Contains(output, "Hello, world!") {
		t.Error("Expected output to contain message text")
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
				"param2": 42,
			},
		},
	}

	err := formatter.Format(event)
	if err != nil {
		t.Errorf("Format failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "🔧") {
		t.Error("Expected output to contain tool emoji")
	}
	if !strings.Contains(output, "TestTool") {
		t.Error("Expected output to contain tool name")
	}
	if !strings.Contains(output, "param1") {
		t.Error("Expected output to contain parameter name")
	}
}

func TestConsoleFormatter_IterationStarted(t *testing.T) {
	buf := &bytes.Buffer{}
	formatter := NewConsoleFormatter(buf)

	event := events.Event{
		Type: events.EventIterationStarted,
		Data: events.IterationStartedData{
			Current: 2,
			Total:   5,
		},
	}

	err := formatter.Format(event)
	if err != nil {
		t.Errorf("Format failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Starting iteration 2 of 5") {
		t.Error("Expected output to contain iteration info")
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
	if !strings.Contains(output, "🧬 Starting Evolution") {
		t.Error("Expected output to contain evolution header")
	}
	if !strings.Contains(output, "Iterations: 3") {
		t.Error("Expected output to contain iteration count")
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
	if !strings.Contains(output, "test message") {
		t.Error("Expected output to contain emitted message")
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
	if !strings.Contains(output, "🪄PROMPT") {
		t.Error("Expected output to contain prompt header")
	}
	if !strings.Contains(output, "response") {
		t.Error("Expected output to contain response")
	}
	if !strings.Contains(output, "Duration: 2.00s") {
		t.Error("Expected output to contain duration")
	}
}
