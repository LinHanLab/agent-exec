package display

import (
	"fmt"
	"os"
	"strings"
	"time"

	"golang.org/x/term"
)

// TextFormatter handles text formatting operations
type TextFormatter struct {
	terminalWidth int
}

// NewTextFormatter creates a new TextFormatter
func NewTextFormatter() *TextFormatter {
	return &TextFormatter{
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

// WrapText wraps text to fit within terminal width, respecting indentation
// Only wraps at natural break points (spaces, commas, hyphens)
func (tf *TextFormatter) WrapText(text string, indent string) string {
	if text == "" {
		return ""
	}

	lines := strings.Split(text, "\n")
	wrapped := make([]string, 0, len(lines))

	// Calculate available width for content (terminal width - indent length)
	availableWidth := tf.terminalWidth - len(indent)
	if availableWidth <= 0 {
		availableWidth = DefaultTerminalWidth - len(indent)
	}

	for _, line := range lines {
		// If line fits within available width, keep it as is
		if len(line) <= availableWidth {
			wrapped = append(wrapped, indent+line)
			continue
		}

		// For long lines, try to wrap at natural break points
		remaining := line
		for len(remaining) > 0 {
			if len(remaining) <= availableWidth {
				wrapped = append(wrapped, indent+remaining)
				break
			}

			// Find a good break point (space, comma, etc.)
			breakPoint := -1
			for i := availableWidth - 1; i > availableWidth/2 && i < len(remaining); i-- {
				if remaining[i] == ' ' || remaining[i] == ',' || remaining[i] == '-' {
					breakPoint = i + 1
					break
				}
			}

			// If no break point found, don't wrap (keep line as is to preserve content)
			if breakPoint == -1 {
				wrapped = append(wrapped, indent+remaining)
				break
			}

			wrapped = append(wrapped, indent+remaining[:breakPoint])
			remaining = strings.TrimLeft(remaining[breakPoint:], " ")
		}
	}

	return strings.Join(wrapped, "\n")
}

// IndentContent adds ContentIndent prefix to each line of content
func (tf *TextFormatter) IndentContent(content string) string {
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

// FormatContent formats content with indentation (no wrapping for JSON)
func (tf *TextFormatter) FormatContent(content string) string {
	if content == "" {
		return ""
	}
	return "\n" + tf.IndentContent(content) + "\n"
}

// FormatWrappedContent formats content with indentation and wrapping for plain text
func (tf *TextFormatter) FormatWrappedContent(content string) string {
	if content == "" {
		return ""
	}
	return "\n" + tf.WrapText(content, ContentIndent) + "\n"
}

// FormatDuration formats duration in human-readable format
func (tf *TextFormatter) FormatDuration(d time.Duration) string {
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
func (tf *TextFormatter) FormatTime() string {
	return time.Now().Format("15:04:05")
}

// ApplyReverseVideo wraps text with reverse video effect
// The color parameter should be the existing color code (e.g., BoldYellow)
func (tf *TextFormatter) ApplyReverseVideo(text string, color string) string {
	if color == "" {
		return fmt.Sprintf("%s%s%s", ReverseVideo, text, Reset)
	}
	return fmt.Sprintf("%s%s%s%s", color, ReverseVideo, text, Reset)
}
