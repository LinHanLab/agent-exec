package display

import (
	"fmt"
	"os"
	"strings"
	"time"

	"golang.org/x/term"
)

type TextFormatter interface {
	IndentContent(content string) string
	FormatContentWithFrame(content string, useBorder ...bool) string
	FormatContentWithFrameAndColor(content string, color string, useBorder ...bool) string
	FormatDuration(d time.Duration) string
	FormatTime() string
	ApplyReverseVideo(text string, color string) string
	TerminalWidth() int
}

type PlainTextFormatter struct {
	terminalWidth int
}

func NewTextFormatter() *PlainTextFormatter {
	return &PlainTextFormatter{
		terminalWidth: GetTerminalWidth(),
	}
}

func GetTerminalWidth() int {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || width <= 0 {
		return DefaultTerminalWidth
	}
	return width
}

func (tf *PlainTextFormatter) IndentContent(content string) string {
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

func (tf *PlainTextFormatter) FormatContentWithFrame(content string, useBorder ...bool) string {
	return tf.FormatContentWithFrameAndColor(content, "", useBorder...)
}

func (tf *PlainTextFormatter) FormatContentWithFrameAndColor(content string, color string, useBorder ...bool) string {
	if content == "" {
		return ""
	}

	drawBorder := false
	if len(useBorder) > 0 {
		drawBorder = useBorder[0]
	}

	lines := strings.Split(content, "\n")
	maxLineLen := 0
	for _, line := range lines {
		if len(line) > maxLineLen {
			maxLineLen = len(line)
		}
	}

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

func (tf *PlainTextFormatter) FormatDuration(d time.Duration) string {
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

func (tf *PlainTextFormatter) FormatTime() string {
	return time.Now().Format("15:04:05")
}

func (tf *PlainTextFormatter) ApplyReverseVideo(text string, color string) string {
	if color == "" {
		return fmt.Sprintf("%s%s%s", ReverseVideo, text, Reset)
	}
	return fmt.Sprintf("%s%s%s%s", color, ReverseVideo, text, Reset)
}

func (tf *PlainTextFormatter) TerminalWidth() int {
	return tf.terminalWidth
}

var _ TextFormatter = (*PlainTextFormatter)(nil)
