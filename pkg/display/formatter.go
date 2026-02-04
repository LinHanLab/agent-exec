package display

import (
	"fmt"
	"os"
	"sync"

	"github.com/LinHanLab/agent-exec/pkg/events"
)

// Formatter is the interface for formatting and displaying events
type Formatter interface {
	// Format processes an event and outputs it
	Format(event events.Event) error
	// Flush ensures all buffered output is written
	Flush() error
}

// Display coordinates event processing with a formatter
type Display struct {
	formatter Formatter
	emitter   events.Emitter
	wg        sync.WaitGroup
}

// NewDisplay creates a new Display
func NewDisplay(formatter Formatter, emitter events.Emitter) *Display {
	return &Display{
		formatter: formatter,
		emitter:   emitter,
	}
}

// Start begins processing events in a goroutine
func (d *Display) Start() {
	d.wg.Add(1)
	go func() {
		defer d.wg.Done()
		ch := d.emitter.Subscribe()
		for event := range ch {
			if err := d.formatter.Format(event); err != nil {
				fmt.Fprintf(os.Stderr, "[display] format error: %v\n", err)
			}
		}
		_ = d.formatter.Flush()
	}()
}

// Wait blocks until all events have been processed
func (d *Display) Wait() {
	d.wg.Wait()
}
