package display

import (
	"strings"
)

// FrameOption configures a FrameBuilder
type FrameOption func(*FrameBuilder)

// FrameBuilder constructs framed content with borders and wrapping.
type FrameBuilder struct {
	contentWidth  int
	borderChar    string
	indent        string
	textColor     string
	useBoxDrawing bool
}

// NewFrameBuilder creates a FrameBuilder with the given options
func NewFrameBuilder(opts ...FrameOption) *FrameBuilder {
	fb := &FrameBuilder{
		contentWidth:  60, // default
		borderChar:    " ",
		indent:        ContentIndent,
		textColor:     "",
		useBoxDrawing: false,
	}
	for _, opt := range opts {
		opt(fb)
	}
	return fb
}

// WithContentWidth sets the content width for the frame
func WithContentWidth(width int) FrameOption {
	return func(fb *FrameBuilder) {
		fb.contentWidth = width
	}
}

// WithBoxDrawing enables box drawing characters for borders
func WithBoxDrawing() FrameOption {
	return func(fb *FrameBuilder) {
		fb.useBoxDrawing = true
		fb.borderChar = "│"
	}
}

// WithIndent sets the indentation for the frame
func WithIndent(indent string) FrameOption {
	return func(fb *FrameBuilder) {
		fb.indent = indent
	}
}

// WithTextColor sets the text color for the frame
func WithTextColor(color string) FrameOption {
	return func(fb *FrameBuilder) {
		fb.textColor = color
	}
}

// Build constructs the framed content
func (fb *FrameBuilder) Build(content string) string {
	if content == "" {
		return ""
	}

	// Define border characters based on mode
	var topLeft, topRight, bottomLeft, bottomRight, horizontal, vertical string
	if fb.useBoxDrawing {
		topLeft = "┌"
		topRight = "┐"
		bottomLeft = "└"
		bottomRight = "┘"
		horizontal = "─"
		vertical = "│"
	} else {
		topLeft = " "
		topRight = " "
		bottomLeft = " "
		bottomRight = " "
		horizontal = " "
		vertical = " "
	}

	// Split content into lines
	lines := strings.Split(content, "\n")

	// Build the frame
	var result strings.Builder

	// Top border
	result.WriteString("\n")
	result.WriteString(fb.indent)
	result.WriteString(Gray)
	result.WriteString(topLeft)
	result.WriteString(strings.Repeat(horizontal, fb.contentWidth))
	result.WriteString(topRight)
	result.WriteString(Reset)
	result.WriteString("\n")

	// Content lines
	for _, line := range lines {
		fb.writeLine(&result, line, vertical)
	}

	// Bottom border
	result.WriteString(fb.indent)
	result.WriteString(Gray)
	result.WriteString(bottomLeft)
	result.WriteString(strings.Repeat(horizontal, fb.contentWidth))
	result.WriteString(bottomRight)
	result.WriteString(Reset)
	result.WriteString("\n")

	return result.String()
}

// writeLine writes a single line with proper padding and wrapping
func (fb *FrameBuilder) writeLine(result *strings.Builder, line string, vertical string) {
	contentWidth := fb.contentWidth

	// Handle lines that are too long by wrapping them
	if len(line) > contentWidth {
		fb.wrapLine(result, line, vertical, contentWidth)
	} else {
		fb.writePaddedLine(result, line, vertical, contentWidth)
	}
}

// wrapLine handles wrapping long lines at natural boundaries
func (fb *FrameBuilder) wrapLine(result *strings.Builder, line string, vertical string, contentWidth int) {
	remaining := line
	for len(remaining) > 0 {
		if len(remaining) <= contentWidth {
			fb.writePaddedLine(result, remaining, vertical, contentWidth)
			break
		}

		// Find break point at natural boundaries
		breakPoint := -1
		for i := contentWidth - 1; i > contentWidth/2 && i < len(remaining); i-- {
			if remaining[i] == ' ' || remaining[i] == ',' || remaining[i] == '-' {
				breakPoint = i + 1
				break
			}
		}

		if breakPoint == -1 {
			fb.writePaddedLine(result, remaining, vertical, contentWidth)
			break
		}

		chunk := remaining[:breakPoint]
		fb.writePaddedLine(result, chunk, vertical, contentWidth)
		remaining = strings.TrimLeft(remaining[breakPoint:], " ")
	}
}

// writePaddedLine writes a line with proper padding
func (fb *FrameBuilder) writePaddedLine(result *strings.Builder, line string, vertical string, contentWidth int) {
	// Ensure padding is never negative
	padding := contentWidth - len(line)
	if padding < 0 {
		padding = 0
	}

	result.WriteString(fb.indent)
	result.WriteString(Gray)
	result.WriteString(vertical)
	result.WriteString(Reset)
	result.WriteString(" ")
	if fb.textColor != "" {
		result.WriteString(fb.textColor)
	}
	result.WriteString(line)
	if fb.textColor != "" {
		result.WriteString(Reset)
	}
	result.WriteString(strings.Repeat(" ", padding))
	result.WriteString(" ")
	result.WriteString(Gray)
	result.WriteString(vertical)
	result.WriteString(Reset)
	result.WriteString("\n")
}
