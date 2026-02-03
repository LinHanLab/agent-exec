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

// FormatContentWithFrame wraps content in a frame with optional box drawing characters
// By default (useBorder=false), uses whitespace for borders (invisible frame)
// When useBorder=true, uses box drawing characters (┌─┐│└┘) for visible borders
func (tf *TextFormatter) FormatContentWithFrame(content string, useBorder ...bool) string {
	if content == "" {
		return ""
	}

	// Determine if we should use box drawing characters (default: false)
	drawBorder := false
	if len(useBorder) > 0 {
		drawBorder = useBorder[0]
	}

	// Define border characters based on mode
	var topLeft, topRight, bottomLeft, bottomRight, horizontal, vertical string
	if drawBorder {
		topLeft = "┌"
		topRight = "┐"
		bottomLeft = "└"
		bottomRight = "┘"
		horizontal = "─"
		vertical = "│"
	} else {
		// Use whitespace for invisible borders
		topLeft = " "
		topRight = " "
		bottomLeft = " "
		bottomRight = " "
		horizontal = " "
		vertical = " "
	}

	// Split content into lines
	lines := strings.Split(content, "\n")

	// Calculate the maximum line length to determine frame width
	maxLineLen := 0
	for _, line := range lines {
		if len(line) > maxLineLen {
			maxLineLen = len(line)
		}
	}

	// Set frame width based on content, with reasonable bounds
	// Frame width is the content width (we'll add spaces on both sides separately)
	// Min: 40 chars, Max: terminal width - indent - borders (│ │) - spaces ( content )
	minFrameWidth := 40
	maxFrameWidth := tf.terminalWidth - len(ContentIndent) - 4 - 2
	if maxFrameWidth < minFrameWidth {
		maxFrameWidth = minFrameWidth
	}

	// Frame width should fit the content (the dashes between ┌ and ┐)
	// This is the content width plus 2 spaces (one on each side)
	frameWidth := maxLineLen + 2
	if frameWidth < minFrameWidth {
		frameWidth = minFrameWidth
	}
	if frameWidth > maxFrameWidth {
		frameWidth = maxFrameWidth
	}

	// Content width is frame width minus the 2 spaces
	contentWidth := frameWidth - 2

	// Build the frame
	var result strings.Builder

	// Top border (with gray color)
	result.WriteString("\n")
	result.WriteString(ContentIndent)
	result.WriteString(Gray)
	result.WriteString(topLeft)
	result.WriteString(strings.Repeat(horizontal, frameWidth))
	result.WriteString(topRight)
	result.WriteString(Reset)
	result.WriteString("\n")

	// Content lines
	for _, line := range lines {
		// Handle lines that are too long by wrapping them
		if len(line) > contentWidth {
			// Wrap long lines
			remaining := line
			for len(remaining) > 0 {
				if len(remaining) <= contentWidth {
					result.WriteString(ContentIndent)
					result.WriteString(Gray)
					result.WriteString(vertical)
					result.WriteString(Reset)
					result.WriteString(" ")
					result.WriteString(remaining)
					result.WriteString(strings.Repeat(" ", contentWidth-len(remaining)))
					result.WriteString(" ")
					result.WriteString(Gray)
					result.WriteString(vertical)
					result.WriteString(Reset)
					result.WriteString("\n")
					break
				}

				// Find break point (only at natural boundaries)
				breakPoint := -1
				for i := contentWidth - 1; i > contentWidth/2 && i < len(remaining); i-- {
					if remaining[i] == ' ' || remaining[i] == ',' || remaining[i] == '-' {
						breakPoint = i + 1
						break
					}
				}

				// If no natural break point found, don't wrap - keep the line as-is
				if breakPoint == -1 {
					result.WriteString(ContentIndent)
					result.WriteString(Gray)
					result.WriteString(vertical)
					result.WriteString(Reset)
					result.WriteString(" ")
					result.WriteString(remaining)
					result.WriteString(strings.Repeat(" ", max(0, contentWidth-len(remaining))))
					result.WriteString(" ")
					result.WriteString(Gray)
					result.WriteString(vertical)
					result.WriteString(Reset)
					result.WriteString("\n")
					break
				}

				chunk := remaining[:breakPoint]
				result.WriteString(ContentIndent)
				result.WriteString(Gray)
				result.WriteString(vertical)
				result.WriteString(Reset)
				result.WriteString(" ")
				result.WriteString(chunk)
				result.WriteString(strings.Repeat(" ", contentWidth-len(chunk)))
				result.WriteString(" ")
				result.WriteString(Gray)
				result.WriteString(vertical)
				result.WriteString(Reset)
				result.WriteString("\n")

				remaining = strings.TrimLeft(remaining[breakPoint:], " ")
			}
		} else {
			// Line fits within frame
			result.WriteString(ContentIndent)
			result.WriteString(Gray)
			result.WriteString(vertical)
			result.WriteString(Reset)
			result.WriteString(" ")
			result.WriteString(line)
			result.WriteString(strings.Repeat(" ", contentWidth-len(line)))
			result.WriteString(" ")
			result.WriteString(Gray)
			result.WriteString(vertical)
			result.WriteString(Reset)
			result.WriteString("\n")
		}
	}

	// Bottom border (with gray color)
	result.WriteString(ContentIndent)
	result.WriteString(Gray)
	result.WriteString(bottomLeft)
	result.WriteString(strings.Repeat(horizontal, frameWidth))
	result.WriteString(bottomRight)
	result.WriteString(Reset)
	result.WriteString("\n")

	return result.String()
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
