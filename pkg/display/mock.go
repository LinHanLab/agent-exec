package display

import (
	"sync"

	"github.com/LinHanLab/agent-exec/pkg/events"
)

// MockFormatter captures events for testing.
type MockFormatter struct {
	mu     sync.Mutex
	Events []events.Event
}

// NewMockFormatter creates a new MockFormatter
func NewMockFormatter() *MockFormatter {
	return &MockFormatter{}
}

// Format captures the event for testing
func (m *MockFormatter) Format(event events.Event) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Events = append(m.Events, event)
	return nil
}

// Flush does nothing (required by Formatter interface)
func (m *MockFormatter) Flush() error {
	return nil
}

// GetEvents returns a copy of captured events (thread-safe)
func (m *MockFormatter) GetEvents() []events.Event {
	m.mu.Lock()
	defer m.mu.Unlock()
	eventsCopy := make([]events.Event, len(m.Events))
	copy(eventsCopy, m.Events)
	return eventsCopy
}

// EventCount returns the number of captured events
func (m *MockFormatter) EventCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.Events)
}

// Clear removes all captured events
func (m *MockFormatter) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Events = nil
}

// Ensure MockFormatter implements Formatter interface
var _ Formatter = (*MockFormatter)(nil)
