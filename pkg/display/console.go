package display

import (
	"fmt"
	"io"

	"github.com/LinHanLab/agent-exec/pkg/events"
)

const (
	// Default terminal width if detection fails
	DefaultTerminalWidth = 80
	// Indentation for content blocks
	ContentIndent = "    "
)

// JSONFormatter formats events as human-readable text with color highlighting
type JSONFormatter struct {
	writer        io.Writer
	verbose       bool
	textFormatter TextFormatter
	contentFilter *ContentFilter
}

// NewConsoleFormatter creates a new JSONFormatter
func NewConsoleFormatter(writer io.Writer, verbose bool) *JSONFormatter {
	return &JSONFormatter{
		writer:        writer,
		verbose:       verbose,
		textFormatter: NewTextFormatter(),
		contentFilter: NewContentFilter(verbose),
	}
}

// Ensure JSONFormatter implements Formatter interface
var _ Formatter = (*JSONFormatter)(nil)

// Format processes an event and outputs it as human-readable colored text
func (f *JSONFormatter) Format(event events.Event) error {
	// Look up the formatter for this event type
	formatter, ok := eventFormatters[event.Type]
	if !ok {
		return fmt.Errorf("unknown event type: %s", event.Type)
	}

	// Create format context
	ctx := &FormatContext{
		TextFormatter: f.textFormatter,
		ContentFilter: f.contentFilter,
		Verbose:       f.verbose,
	}

	// Format the event
	output, err := formatter(event, ctx)
	if err != nil {
		return err
	}

	// Write spacing first, then output
	_, err = fmt.Fprint(f.writer, "\n")
	if err != nil {
		return fmt.Errorf("failed to write spacing to console: %w", err)
	}

	// Write output (color already applied by formatter)
	_, err = fmt.Fprintf(f.writer, "%s\n", output)
	if err != nil {
		return fmt.Errorf("failed to write to console: %w", err)
	}

	return nil
}

// Flush ensures all buffered output is written
func (f *JSONFormatter) Flush() error {
	return nil
}
