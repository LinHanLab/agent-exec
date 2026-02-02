package events

import "time"

// Emitter is the interface for emitting events
type Emitter interface {
	// Emit sends an event to all subscribers
	Emit(eventType EventType, data interface{})
	// Subscribe returns a channel that receives all events
	Subscribe() <-chan Event
	// Close closes the emitter and all subscriber channels
	Close()
}

// ChannelEmitter implements Emitter using Go channels
type ChannelEmitter struct {
	ch     chan Event
	closed bool
}

// NewChannelEmitter creates a new ChannelEmitter with the specified buffer size
func NewChannelEmitter(bufferSize int) *ChannelEmitter {
	return &ChannelEmitter{
		ch:     make(chan Event, bufferSize),
		closed: false,
	}
}

// Emit sends an event to all subscribers
func (e *ChannelEmitter) Emit(eventType EventType, data interface{}) {
	if e.closed {
		return
	}
	e.ch <- Event{
		Type:      eventType,
		Timestamp: time.Now(),
		Data:      data,
	}
}

// Subscribe returns a channel that receives all events
func (e *ChannelEmitter) Subscribe() <-chan Event {
	return e.ch
}

// Close closes the emitter and all subscriber channels
func (e *ChannelEmitter) Close() {
	if !e.closed {
		e.closed = true
		close(e.ch)
	}
}

// NullEmitter is an emitter that discards all events (for testing)
type NullEmitter struct{}

// NewNullEmitter creates a new NullEmitter
func NewNullEmitter() *NullEmitter {
	return &NullEmitter{}
}

// Emit discards the event
func (e *NullEmitter) Emit(eventType EventType, data interface{}) {
	// Do nothing
}

// Subscribe returns a closed channel
func (e *NullEmitter) Subscribe() <-chan Event {
	ch := make(chan Event)
	close(ch)
	return ch
}

// Close does nothing
func (e *NullEmitter) Close() {
	// Do nothing
}
