package display

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/LinHanLab/agent-exec/pkg/events"
	"golang.org/x/term"
)

// StatusLineFormatter wraps another formatter and adds a 4-line status block
type StatusLineFormatter struct {
	wrapped       Formatter
	writer        io.Writer
	mu            sync.Mutex
	enabled       bool
	isTTY         bool
	terminalWidth int

	// Status block state
	statusVisible bool
	statusLines   int // Always 4

	// Context for status line
	iteration int
	total     int
	cwd       string
	branch    string
	baseURL   string
	prompt    string
	isEvolve  bool // Track if we're in evolve mode (use "Round" instead of "Iteration")
	startTime time.Time
}

// NewStatusLineFormatter creates a new status line formatter
func NewStatusLineFormatter(wrapped Formatter, writer io.Writer, enabled bool) *StatusLineFormatter {
	f := &StatusLineFormatter{
		wrapped:     wrapped,
		writer:      writer,
		enabled:     enabled,
		statusLines: 4,
		startTime:   time.Now(),
	}

	// Detect TTY
	if file, ok := writer.(*os.File); ok {
		f.isTTY = term.IsTerminal(int(file.Fd()))
	}

	// Disable if not TTY or not enabled
	if !f.isTTY || !enabled {
		f.enabled = false
		return f
	}

	// Get terminal width
	if file, ok := writer.(*os.File); ok {
		width, _, err := term.GetSize(int(file.Fd()))
		if err == nil {
			f.terminalWidth = width
		} else {
			f.terminalWidth = 80 // Default fallback
		}
	} else {
		f.terminalWidth = 80
	}

	// Get current working directory
	if cwd, err := os.Getwd(); err == nil {
		f.cwd = cwd
	}

	// Get base URL from environment
	f.baseURL = os.Getenv("ANTHROPIC_BASE_URL")

	return f
}

// Format processes an event and updates the status line
func (f *StatusLineFormatter) Format(event events.Event) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	// Update internal state based on event type
	f.updateState(event)

	// Clear status block if visible
	if f.statusVisible {
		f.clearStatusBlock()
	}

	// Call wrapped formatter for normal output
	if err := f.wrapped.Format(event); err != nil {
		return err
	}

	// Redraw status block
	f.updateStatusBlock()

	return nil
}

// updateState updates internal state based on event type
func (f *StatusLineFormatter) updateState(event events.Event) {
	switch event.Type {
	case events.EventRunPromptStarted:
		if data, ok := event.Data.(events.RunPromptStartedData); ok {
			f.cwd = data.Cwd
			f.prompt = data.Prompt
		}

	case events.EventIterationStarted:
		if data, ok := event.Data.(events.IterationStartedData); ok {
			f.iteration = data.Current
			f.total = data.Total
			f.isEvolve = false
		}

	case events.EventRoundStarted:
		if data, ok := event.Data.(events.RoundStartedData); ok {
			f.iteration = data.Round
			f.total = data.Total
			f.isEvolve = true
		}

	case events.EventGitBranchCreated:
		if data, ok := event.Data.(events.BranchCreatedData); ok {
			f.branch = data.BranchName
		}

	case events.EventGitBranchCheckedOut:
		if data, ok := event.Data.(events.BranchCheckedOutData); ok {
			f.branch = data.BranchName
		}
	}
}

// buildStatusBlock builds the 4-line status block
func (f *StatusLineFormatter) buildStatusBlock() []string {
	// Line 1: Empty (visual divider)
	line1 := ""

	// Line 2: [Iter 3/10 or Round 2/5], CWD: folder-name, Git Branch: branch-name, Time: 1h30m3s
	var parts []string

	// Add iteration/round progress at the start
	if f.iteration > 0 && f.total > 0 {
		if f.isEvolve {
			parts = append(parts, fmt.Sprintf("Round %d/%d", f.iteration, f.total))
		} else {
			parts = append(parts, fmt.Sprintf("Iter %d/%d", f.iteration, f.total))
		}
	}

	if f.cwd != "" {
		folderName := filepath.Base(f.cwd)
		parts = append(parts, fmt.Sprintf("CWD: %s", folderName))
	}

	if f.branch != "" {
		parts = append(parts, fmt.Sprintf("Git Branch: %s", f.branch))
	}

	// Add elapsed time
	elapsed := time.Since(f.startTime)
	timeStr := formatDuration(elapsed)
	parts = append(parts, fmt.Sprintf("Time: %s", timeStr))

	line2 := ""
	if len(parts) > 0 {
		line2 = strings.Join(parts, ", ")
	}

	// Line 3: Base URL: https://example.org
	line3 := ""
	if f.baseURL != "" {
		line3 = fmt.Sprintf("Base URL: %s", f.baseURL)
	}

	// Line 4: Prompt: "text..." (with literal \n instead of newlines)
	line4 := ""
	if f.prompt != "" {
		// Replace actual newlines with literal \n
		prompt := strings.ReplaceAll(f.prompt, "\n", "\\n")
		prompt = strings.ReplaceAll(prompt, "\r", "\\r")
		prompt = strings.ReplaceAll(prompt, "\t", "\\t")

		if len(prompt) > 80 {
			prompt = prompt[:80] + "..."
		}
		line4 = fmt.Sprintf("Prompt: \"%s\"", prompt)
	}

	return []string{line1, line2, line3, line4}
}

// formatDuration formats a duration in a human-readable format like "1h30m3s"
func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second

	if h > 0 {
		return fmt.Sprintf("%dh%dm%ds", h, m, s)
	}
	if m > 0 {
		return fmt.Sprintf("%dm%ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
}

// updateStatusBlock renders the status block
func (f *StatusLineFormatter) updateStatusBlock() {
	if !f.enabled || !f.isTTY {
		return
	}

	lines := f.buildStatusBlock()

	for _, line := range lines {
		// Truncate to terminal width
		if len(line) > f.terminalWidth {
			if f.terminalWidth > 3 {
				line = line[:f.terminalWidth-3] + "..."
			} else {
				line = line[:f.terminalWidth]
			}
		}
		_, _ = fmt.Fprintf(f.writer, "%s\n", line)
	}

	f.statusVisible = true
}

// clearStatusBlock clears the 4-line status block
func (f *StatusLineFormatter) clearStatusBlock() {
	if !f.statusVisible {
		return
	}

	// Move cursor up 4 lines to the start of the status block
	_, _ = fmt.Fprintf(f.writer, "\033[4A")

	// Clear each line and move down
	for i := 0; i < 4; i++ {
		_, _ = fmt.Fprintf(f.writer, "\r\033[K")
		if i < 3 {
			_, _ = fmt.Fprint(f.writer, "\n")
		}
	}

	// Move cursor back up to the first line of where the status block was
	_, _ = fmt.Fprintf(f.writer, "\r\033[3A")

	f.statusVisible = false
}

// Flush clears the status block and flushes the wrapped formatter
func (f *StatusLineFormatter) Flush() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	// Clear status block if visible
	if f.statusVisible {
		f.clearStatusBlock()
	}

	// Flush wrapped formatter
	return f.wrapped.Flush()
}
