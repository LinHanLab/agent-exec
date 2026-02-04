package display

import (
	"fmt"
	"io"

	"github.com/LinHanLab/agent-exec/pkg/events"
)

const (
	DefaultTerminalWidth = 80
	ContentIndent        = "    "
)

// ConsoleFormatter formats events as human-readable text with color highlighting
type ConsoleFormatter struct {
	writer        io.Writer
	verbose       bool
	textFormatter TextFormatter
	contentFilter *ContentFilter
}

func NewConsoleFormatter(writer io.Writer, verbose bool) *ConsoleFormatter {
	return &ConsoleFormatter{
		writer:        writer,
		verbose:       verbose,
		textFormatter: NewTextFormatter(),
		contentFilter: NewContentFilter(verbose),
	}
}

var _ Formatter = (*ConsoleFormatter)(nil)

// Format processes an event and outputs colored text
func (f *ConsoleFormatter) Format(event events.Event) error {
	formatter, ok := eventFormatters[event.Type]
	if !ok {
		return fmt.Errorf("unknown event type: %s", event.Type)
	}

	ctx := &FormatContext{
		TextFormatter: f.textFormatter,
		ContentFilter: f.contentFilter,
		Verbose:       f.verbose,
	}

	output, err := formatter(event, ctx)
	if err != nil {
		return err
	}

	_, err = fmt.Fprint(f.writer, "\n")
	if err != nil {
		return fmt.Errorf("failed to write spacing to console: %w", err)
	}

	_, err = fmt.Fprintf(f.writer, "%s\n", output)
	if err != nil {
		return fmt.Errorf("failed to write to console: %w", err)
	}

	return nil
}

func (f *ConsoleFormatter) Flush() error {
	return nil
}
