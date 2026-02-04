package display

import (
	"strings"
)

type FrameOption func(*FrameBuilder)

type FrameBuilder struct {
	contentWidth  int
	borderChar    string
	indent        string
	textColor     string
	useBoxDrawing bool
}

func NewFrameBuilder(opts ...FrameOption) *FrameBuilder {
	fb := &FrameBuilder{
		contentWidth:  60,
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

func WithContentWidth(width int) FrameOption {
	return func(fb *FrameBuilder) {
		fb.contentWidth = width
	}
}

func WithBoxDrawing() FrameOption {
	return func(fb *FrameBuilder) {
		fb.useBoxDrawing = true
		fb.borderChar = "│"
	}
}

func WithIndent(indent string) FrameOption {
	return func(fb *FrameBuilder) {
		fb.indent = indent
	}
}

func WithTextColor(color string) FrameOption {
	return func(fb *FrameBuilder) {
		fb.textColor = color
	}
}

func (fb *FrameBuilder) Build(content string) string {
	if content == "" {
		return ""
	}

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

	lines := strings.Split(content, "\n")

	var result strings.Builder

	result.WriteString("\n")
	result.WriteString(fb.indent)
	result.WriteString(Gray)
	result.WriteString(topLeft)
	result.WriteString(strings.Repeat(horizontal, fb.contentWidth))
	result.WriteString(topRight)
	result.WriteString(Reset)
	result.WriteString("\n")

	for _, line := range lines {
		fb.writeLine(&result, line, vertical)
	}

	result.WriteString(fb.indent)
	result.WriteString(Gray)
	result.WriteString(bottomLeft)
	result.WriteString(strings.Repeat(horizontal, fb.contentWidth))
	result.WriteString(bottomRight)
	result.WriteString(Reset)
	result.WriteString("\n")

	return result.String()
}

func (fb *FrameBuilder) writeLine(result *strings.Builder, line string, vertical string) {
	contentWidth := fb.contentWidth

	if len(line) > contentWidth {
		fb.wrapLine(result, line, vertical, contentWidth)
	} else {
		fb.writePaddedLine(result, line, vertical, contentWidth)
	}
}

func (fb *FrameBuilder) wrapLine(result *strings.Builder, line string, vertical string, contentWidth int) {
	remaining := line
	for len(remaining) > 0 {
		if len(remaining) <= contentWidth {
			fb.writePaddedLine(result, remaining, vertical, contentWidth)
			break
		}

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

func (fb *FrameBuilder) writePaddedLine(result *strings.Builder, line string, vertical string, contentWidth int) {
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
