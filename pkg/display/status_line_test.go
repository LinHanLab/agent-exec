package display

import (
	"bytes"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/LinHanLab/agent-exec/pkg/events"
)

// mockFormatter is a simple formatter for testing
type mockFormatter struct {
	formatted []events.Event
	mu        sync.Mutex
}

func (m *mockFormatter) Format(event events.Event) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.formatted = append(m.formatted, event)
	return nil
}

func (m *mockFormatter) Flush() error {
	return nil
}

func (m *mockFormatter) getFormatted() []events.Event {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]events.Event{}, m.formatted...)
}

func TestStatusLineFormatter_TTYDetection(t *testing.T) {
	wrapped := &mockFormatter{}
	buf := &bytes.Buffer{}

	// Create with enabled=true but non-TTY writer
	f := NewStatusLineFormatter(wrapped, buf, true)

	// Should be disabled because buf is not a TTY
	if f.enabled {
		t.Error("Expected formatter to be disabled for non-TTY writer")
	}
}

func TestStatusLineFormatter_StatusBlockRendering(t *testing.T) {
	wrapped := &mockFormatter{}
	buf := &bytes.Buffer{}

	f := &StatusLineFormatter{
		wrapped:       wrapped,
		writer:        buf,
		enabled:       true,
		isTTY:         true,
		terminalWidth: 120,
		statusLines:   4,
		iteration:     3,
		total:         10,
		cwd:           "/path/to/cwd",
		branch:        "feat/branch-name",
		baseURL:       "https://api.example.com",
		prompt:        "improve the authentication system to use JWT tokens",
		isEvolve:      false,
		startTime:     time.Now().Add(-5 * time.Minute),
	}

	lines := f.buildStatusBlock()

	// Verify 4 lines
	if len(lines) != 4 {
		t.Fatalf("Expected 4 lines, got %d", len(lines))
	}

	// Line 1: Empty
	if lines[0] != "" {
		t.Errorf("Expected line 1 to be empty, got %q", lines[0])
	}

	// Line 2: Should contain Iter, CWD, branch, and time
	if !strings.Contains(lines[1], "Iter") {
		t.Errorf("Line 2 should contain 'Iter', got %q", lines[1])
	}
	if !strings.Contains(lines[1], "3/10") {
		t.Errorf("Line 2 should contain '3/10', got %q", lines[1])
	}
	if !strings.Contains(lines[1], "CWD: cwd") {
		t.Errorf("Line 2 should contain 'CWD: cwd', got %q", lines[1])
	}
	if !strings.Contains(lines[1], "Git Branch: feat/branch-name") {
		t.Errorf("Line 2 should contain 'Git Branch: feat/branch-name', got %q", lines[1])
	}
	if !strings.Contains(lines[1], "Time:") {
		t.Errorf("Line 2 should contain 'Time:', got %q", lines[1])
	}

	// Line 3: Should contain Base URL
	if !strings.Contains(lines[2], "Base URL: https://api.example.com") {
		t.Errorf("Line 3 should contain 'Base URL: https://api.example.com', got %q", lines[2])
	}

	// Line 4: Should contain prompt with Prompt: prefix
	if !strings.Contains(lines[3], "Prompt: \"improve the authentication system") {
		t.Errorf("Line 4 should contain prompt with prefix, got %q", lines[3])
	}
}

func TestStatusLineFormatter_EvolveMode(t *testing.T) {
	wrapped := &mockFormatter{}
	buf := &bytes.Buffer{}

	f := &StatusLineFormatter{
		wrapped:       wrapped,
		writer:        buf,
		enabled:       true,
		isTTY:         true,
		terminalWidth: 80,
		statusLines:   4,
		iteration:     2,
		total:         5,
		isEvolve:      true,
		cwd:           "/test/path",
		branch:        "test-branch",
	}

	lines := f.buildStatusBlock()

	// Verify 4 lines
	if len(lines) != 4 {
		t.Fatalf("Expected 4 lines, got %d", len(lines))
	}

	// Line 2: Should contain Round progress and CWD
	if !strings.Contains(lines[1], "Round") {
		t.Errorf("Line 2 should contain 'Round', got %q", lines[1])
	}
	if !strings.Contains(lines[1], "2/5") {
		t.Errorf("Line 2 should contain '2/5', got %q", lines[1])
	}
	if !strings.Contains(lines[1], "CWD: path") {
		t.Errorf("Expected CWD in line 2, got %q", lines[1])
	}
}

func TestStatusLineFormatter_PromptTruncation(t *testing.T) {
	wrapped := &mockFormatter{}
	buf := &bytes.Buffer{}

	longPrompt := strings.Repeat("a", 150)

	f := &StatusLineFormatter{
		wrapped:       wrapped,
		writer:        buf,
		enabled:       true,
		isTTY:         true,
		terminalWidth: 80,
		statusLines:   4,
		prompt:        longPrompt,
	}

	lines := f.buildStatusBlock()

	// Verify 4 lines
	if len(lines) != 4 {
		t.Fatalf("Expected 4 lines, got %d", len(lines))
	}

	// Line 4 should be truncated to 80 chars + "..." + "Prompt: " prefix and quotes
	// Format is: Prompt: "text..."
	if !strings.HasPrefix(lines[3], "Prompt: \"") {
		t.Errorf("Expected prompt to start with 'Prompt: \"', got %q", lines[3])
	}
	if !strings.Contains(lines[3], "...") {
		t.Errorf("Expected truncated prompt to contain ..., got %q", lines[3])
	}
}

func TestStatusLineFormatter_EventTracking(t *testing.T) {
	wrapped := &mockFormatter{}
	buf := &bytes.Buffer{}

	f := &StatusLineFormatter{
		wrapped:       wrapped,
		writer:        buf,
		enabled:       false, // Disable to avoid TTY issues
		isTTY:         false,
		terminalWidth: 80,
		statusLines:   4,
		startTime:     time.Now(),
	}

	// Test RunPromptStarted event (no longer updates baseURL)
	event1 := events.Event{
		Type: events.EventRunPromptStarted,
		Data: events.RunPromptStartedData{
			Prompt:  "test prompt",
			Cwd:     "/test/dir",
			BaseURL: "https://api.test.com",
		},
	}
	f.updateState(event1)

	if f.prompt != "test prompt" {
		t.Errorf("Expected prompt to be updated, got %q", f.prompt)
	}
	if f.cwd != "/test/dir" {
		t.Errorf("Expected cwd to be updated, got %q", f.cwd)
	}

	// Test IterationStarted event
	event2 := events.Event{
		Type: events.EventIterationStarted,
		Data: events.IterationStartedData{
			Current: 5,
			Total:   10,
		},
	}
	f.updateState(event2)

	if f.iteration != 5 {
		t.Errorf("Expected iteration to be 5, got %d", f.iteration)
	}
	if f.total != 10 {
		t.Errorf("Expected total to be 10, got %d", f.total)
	}
	if f.isEvolve {
		t.Error("Expected isEvolve to be false for iteration event")
	}

	// Test RoundStarted event
	event3 := events.Event{
		Type: events.EventRoundStarted,
		Data: events.RoundStartedData{
			Round: 3,
			Total: 7,
		},
	}
	f.updateState(event3)

	if f.iteration != 3 {
		t.Errorf("Expected iteration to be 3, got %d", f.iteration)
	}
	if f.total != 7 {
		t.Errorf("Expected total to be 7, got %d", f.total)
	}
	if !f.isEvolve {
		t.Error("Expected isEvolve to be true for round event")
	}

	// Test GitBranchCreated event
	event4 := events.Event{
		Type: events.EventGitBranchCreated,
		Data: events.BranchCreatedData{
			BranchName: "test-branch",
		},
	}
	f.updateState(event4)

	if f.branch != "test-branch" {
		t.Errorf("Expected branch to be test-branch, got %q", f.branch)
	}

	// Test GitBranchCheckedOut event
	event5 := events.Event{
		Type: events.EventGitBranchCheckedOut,
		Data: events.BranchCheckedOutData{
			BranchName: "another-branch",
		},
	}
	f.updateState(event5)

	if f.branch != "another-branch" {
		t.Errorf("Expected branch to be another-branch, got %q", f.branch)
	}
}

func TestStatusLineFormatter_ConcurrentAccess(t *testing.T) {
	wrapped := &mockFormatter{}
	buf := &bytes.Buffer{}

	f := &StatusLineFormatter{
		wrapped:       wrapped,
		writer:        buf,
		enabled:       false, // Disable to avoid TTY issues
		isTTY:         false,
		terminalWidth: 80,
		statusLines:   3,
	}

	// Run concurrent Format calls
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			event := events.Event{
				Type: events.EventIterationStarted,
				Data: events.IterationStartedData{
					Current: n,
					Total:   100,
				},
			}
			_ = f.Format(event)
		}(i)
	}

	wg.Wait()

	// Should not panic and should have processed all events
	formatted := wrapped.getFormatted()
	if len(formatted) != 100 {
		t.Errorf("Expected 100 formatted events, got %d", len(formatted))
	}
}

func TestStatusLineFormatter_Flush(t *testing.T) {
	wrapped := &mockFormatter{}
	buf := &bytes.Buffer{}

	f := &StatusLineFormatter{
		wrapped:       wrapped,
		writer:        buf,
		enabled:       false, // Disable to avoid TTY issues
		isTTY:         false,
		terminalWidth: 80,
		statusLines:   4,
		statusVisible: true, // Simulate visible status
	}

	err := f.Flush()
	if err != nil {
		t.Errorf("Flush returned error: %v", err)
	}

	// Status should be cleared
	if f.statusVisible {
		t.Error("Expected statusVisible to be false after Flush")
	}
}

func TestStatusLineFormatter_TerminalWidthTruncation(t *testing.T) {
	wrapped := &mockFormatter{}
	buf := &bytes.Buffer{}

	f := &StatusLineFormatter{
		wrapped:       wrapped,
		writer:        buf,
		enabled:       true,
		isTTY:         true,
		terminalWidth: 40, // Small terminal
		statusLines:   4,
		iteration:     1,
		total:         10,
		cwd:           "/very/long/path/to/working/directory",
		branch:        "feat/very-long-branch-name",
		prompt:        "short prompt",
	}

	f.updateStatusBlock()

	output := buf.String()
	lines := strings.Split(strings.TrimRight(output, "\n"), "\n")

	// Each line should not exceed terminal width
	for i, line := range lines {
		if len(line) > f.terminalWidth {
			t.Errorf("Line %d exceeds terminal width: %d > %d", i, len(line), f.terminalWidth)
		}
	}
}

func TestStatusLineFormatter_NewlineReplacement(t *testing.T) {
	wrapped := &mockFormatter{}
	buf := &bytes.Buffer{}

	promptWithNewlines := "Create a snake game\n- Terminal-based UI\n- Use TUI library"

	f := &StatusLineFormatter{
		wrapped:       wrapped,
		writer:        buf,
		enabled:       true,
		isTTY:         true,
		terminalWidth: 120,
		statusLines:   4,
		prompt:        promptWithNewlines,
		startTime:     time.Now(),
	}

	lines := f.buildStatusBlock()

	// Line 4 should have literal \n instead of actual newlines
	if !strings.Contains(lines[3], "\\n") {
		t.Errorf("Expected prompt to contain literal \\n, got %q", lines[3])
	}
	if strings.Contains(lines[3], "\n") {
		t.Errorf("Expected prompt to not contain actual newlines, got %q", lines[3])
	}
	if !strings.Contains(lines[3], "Create a snake game\\n- Terminal-based UI") {
		t.Errorf("Expected prompt with replaced newlines, got %q", lines[3])
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		duration time.Duration
		expected string
	}{
		{45 * time.Second, "45s"},
		{5*time.Minute + 20*time.Second, "5m20s"},
		{1*time.Hour + 30*time.Minute + 3*time.Second, "1h30m3s"},
		{2 * time.Hour, "2h0m0s"},
		{90 * time.Minute, "1h30m0s"},
		{0 * time.Second, "0s"},
	}

	for _, tt := range tests {
		result := formatDuration(tt.duration)
		if result != tt.expected {
			t.Errorf("formatDuration(%v) = %q, expected %q", tt.duration, result, tt.expected)
		}
	}
}
