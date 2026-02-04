package display

import (
	"fmt"
	"os"
	"strings"
	"time"

	"golang.org/x/term"
)

// TextFormatter is the interface for text formatting operations.
type TextFormatter interface {
	IndentContent(content string) string
	FormatContentWithFrame(content string, useBorder ...bool) string
	FormatContentWithFrameAndColor(content string, color string, useBorder ...bool) string
	FormatDuration(d time.Duration) string
	FormatTime() string
	ApplyReverseVideo(text string, color string) string
	TerminalWidth() int
}

// DefaultTextFormatter handles text formatting operations
type DefaultTextFormatter struct {
	terminalWidth int
}

// NewTextFormatter creates a new DefaultTextFormatter
func NewTextFormatter() *DefaultTextFormatter {
	return &DefaultTextFormatter{
		terminalWidth: GetTerminalWidth(),
	}
}

// GetTerminalWidth returns the current terminal width, or default if detection fails
func GetTerminalWidth() int {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || width <= 0 {
		return DefaultTerminalWidth
	}
	return width
}

// IndentContent adds ContentIndent prefix to each line of content
func (tf *DefaultTextFormatter) IndentContent(content string) string {
	if content == "" {
		return content
	}

	lines := strings.Split(content, "\n")
	indented := make([]string, len(lines))
	for i, line := range lines {
		indented[i] = ContentIndent + line
	}
	return strings.Join(indented, "\n")
}

// FormatContentWithFrame wraps content in a frame with optional box drawing characters
// By default (useBorder=false), uses whitespace for borders (invisible frame)
// When useBorder=true, uses box drawing characters (┌─┐│└┘) for visible borders
// Optional color parameter can be provided to colorize the content
func (tf *DefaultTextFormatter) FormatContentWithFrame(content string, useBorder ...bool) string {
	return tf.FormatContentWithFrameAndColor(content, "", useBorder...)
}

// FormatContentWithFrameAndColor wraps content in a frame with optional color
func (tf *DefaultTextFormatter) FormatContentWithFrameAndColor(content string, color string, useBorder ...bool) string {
	if content == "" {
		return ""
	}

	// Determine if we should use box drawing characters (default: false)
	drawBorder := false
	if len(useBorder) > 0 {
		drawBorder = useBorder[0]
	}

	// Split content into lines to calculate max line length
	lines := strings.Split(content, "\n")
	maxLineLen := 0
	for _, line := range lines {
		if len(line) > maxLineLen {
			maxLineLen = len(line)
		}
	}

	// Calculate content width with reasonable bounds
	minContentWidth := 40
	maxContentWidth := tf.terminalWidth - len(ContentIndent) - 6
	if maxContentWidth < minContentWidth {
		maxContentWidth = minContentWidth
	}

	contentWidth := maxLineLen + 2
	if contentWidth < minContentWidth {
		contentWidth = minContentWidth
	}
	if contentWidth > maxContentWidth {
		contentWidth = maxContentWidth
	}

	// Build frame using FrameBuilder
	opts := []FrameOption{
		WithContentWidth(contentWidth),
		WithIndent(ContentIndent),
	}
	if color != "" {
		opts = append(opts, WithTextColor(color))
	}
	if drawBorder {
		opts = append(opts, WithBoxDrawing())
	}

	fb := NewFrameBuilder(opts...)
	return fb.Build(content)
}

// FormatDuration formats duration in human-readable format
func (tf *DefaultTextFormatter) FormatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%.1fms", float64(d.Milliseconds()))
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	minutes := int(d.Minutes())
	seconds := int(d.Seconds()) % 60
	return fmt.Sprintf("%dm %ds", minutes, seconds)
}

// FormatTime returns current time in HH:MM:SS format
func (tf *DefaultTextFormatter) FormatTime() string {
	return time.Now().Format("15:04:05")
}

// ApplyReverseVideo wraps text with reverse video effect
// The color parameter should be the existing color code (e.g., BoldYellow)
func (tf *DefaultTextFormatter) ApplyReverseVideo(text string, color string) string {
	if color == "" {
		return fmt.Sprintf("%s%s%s", ReverseVideo, text, Reset)
	}
	return fmt.Sprintf("%s%s%s%s", color, ReverseVideo, text, Reset)
}

// TerminalWidth returns the terminal width
func (tf *DefaultTextFormatter) TerminalWidth() int {
	return tf.terminalWidth
}

// Ensure DefaultTextFormatter implements TextFormatter interface
var _ TextFormatter = (*DefaultTextFormatter)(nil)
