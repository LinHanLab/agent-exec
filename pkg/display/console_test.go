package display

import (
	"bytes"
	"encoding/json"
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

	// Strip ANSI color codes for JSON parsing
	jsonStr := stripANSI(output)

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v\nOutput: %s", err, jsonStr)
	}

	// Verify type field is present and correct
	if result["type"] != "prompt_started" {
		t.Errorf("Expected type 'prompt_started', got %v", result["type"])
	}

	// Verify time field is present
	if _, ok := result["time"]; !ok {
		t.Error("Expected time field to be present")
	}

	// Verify prompt text
	if result["prompt"] != "test prompt" {
		t.Errorf("Expected prompt 'test prompt', got %v", result["prompt"])
	}

	// Verify cwd
	if result["cwd"] != "/test/dir" {
		t.Errorf("Expected cwd '/test/dir', got %v", result["cwd"])
	}

	// Verify base_url
	if result["base_url"] != "https://api.anthropic.com" {
		t.Errorf("Expected base_url 'https://api.anthropic.com', got %v", result["base_url"])
	}

	// Verify color code is BoldCyan
	if !strings.HasPrefix(output, BoldCyan) {
		t.Error("Expected output to start with BoldCyan color code")
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
	jsonStr := stripANSI(output)

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if result["type"] != "claude_assistant_message" {
		t.Errorf("Expected type 'claude_assistant_message', got %v", result["type"])
	}

	if result["text"] != "Hello, world!" {
		t.Errorf("Expected text 'Hello, world!', got %v", result["text"])
	}

	// Verify no color (empty string for assistant messages)
	if strings.HasPrefix(output, BoldCyan) || strings.HasPrefix(output, BoldYellow) {
		t.Error("Expected no color code for assistant message")
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
	jsonStr := stripANSI(output)

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if result["type"] != "claude_tool_use" {
		t.Errorf("Expected type 'claude_tool_use', got %v", result["type"])
	}

	if result["name"] != "TestTool" {
		t.Errorf("Expected name 'TestTool', got %v", result["name"])
	}

	// Verify input is a nested object
	input, ok := result["input"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected input to be a map")
	}

	if input["param1"] != "value1" {
		t.Errorf("Expected param1 'value1', got %v", input["param1"])
	}

	if input["param2"] != float64(42) {
		t.Errorf("Expected param2 42, got %v", input["param2"])
	}

	// Verify color code is Magenta
	if !strings.HasPrefix(output, Magenta) {
		t.Error("Expected output to start with Magenta color code")
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
	jsonStr := stripANSI(output)

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if result["type"] != "iteration_started" {
		t.Errorf("Expected type 'iteration_started', got %v", result["type"])
	}

	if result["current"] != float64(2) {
		t.Errorf("Expected current 2, got %v", result["current"])
	}

	if result["total"] != float64(5) {
		t.Errorf("Expected total 5, got %v", result["total"])
	}

	// Verify color code is BoldYellow
	if !strings.HasPrefix(output, BoldYellow) {
		t.Error("Expected output to start with BoldYellow color code")
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
	jsonStr := stripANSI(output)

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if result["type"] != "evolve_started" {
		t.Errorf("Expected type 'evolve_started', got %v", result["type"])
	}

	if result["prompt"] != "test evolution" {
		t.Errorf("Expected prompt 'test evolution', got %v", result["prompt"])
	}

	if result["iterations"] != float64(3) {
		t.Errorf("Expected iterations 3, got %v", result["iterations"])
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
	jsonStr := stripANSI(output)

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if result["type"] != "claude_execution_result" {
		t.Errorf("Expected type 'claude_execution_result', got %v", result["type"])
	}

	if result["duration_sec"] != float64(2) {
		t.Errorf("Expected duration_sec 2, got %v", result["duration_sec"])
	}

	// Verify color code is BoldGreen
	if !strings.HasPrefix(output, BoldGreen) {
		t.Error("Expected output to start with BoldGreen color code")
	}
}

func TestConsoleFormatter_GitBranchCreated(t *testing.T) {
	buf := &bytes.Buffer{}
	formatter := NewConsoleFormatter(buf)

	event := events.Event{
		Type: events.EventGitBranchCreated,
		Data: events.BranchCreatedData{
			BranchName: "feature-branch",
			Base:       "main",
		},
	}

	err := formatter.Format(event)
	if err != nil {
		t.Errorf("Format failed: %v", err)
	}

	output := buf.String()
	jsonStr := stripANSI(output)

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if result["type"] != "git_branch_created" {
		t.Errorf("Expected type 'git_branch_created', got %v", result["type"])
	}

	if result["branch"] != "feature-branch" {
		t.Errorf("Expected branch 'feature-branch', got %v", result["branch"])
	}

	if result["base"] != "main" {
		t.Errorf("Expected base 'main', got %v", result["base"])
	}

	// Verify color code is Cyan
	if !strings.HasPrefix(output, Cyan) {
		t.Error("Expected output to start with Cyan color code")
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

	// Verify all three events are present
	if !strings.Contains(output, "prompt_started") {
		t.Error("Expected output to contain prompt_started event")
	}
	if !strings.Contains(output, "response") {
		t.Error("Expected output to contain response")
	}
	if !strings.Contains(output, "duration_sec") {
		t.Error("Expected output to contain duration_sec field")
	}

	// Verify JSON structure by parsing each line
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 3 {
		t.Errorf("Expected 3 lines of output, got %d", len(lines))
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
