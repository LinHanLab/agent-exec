package display

import (
	"fmt"
	"os"
	"sync"

	"github.com/LinHanLab/agent-exec/pkg/events"
)

type Formatter interface {
	Format(event events.Event) error
	Flush() error
}

type Display struct {
	formatter Formatter
	emitter   events.Emitter
	wg        sync.WaitGroup
}

func NewDisplay(formatter Formatter, emitter events.Emitter) *Display {
	return &Display{
		formatter: formatter,
		emitter:   emitter,
	}
}

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

func (d *Display) Wait() {
	d.wg.Wait()
}
