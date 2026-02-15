package display

import (
	"sync"

	"github.com/LinHanLab/agent-exec/pkg/events"
)

type MockFormatter struct {
	mu     sync.Mutex
	Events []events.Event
}

func NewMockFormatter() *MockFormatter {
	return &MockFormatter{}
}

func (m *MockFormatter) Format(event events.Event) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Events = append(m.Events, event)
	return nil
}

func (m *MockFormatter) Flush() error {
	return nil
}

func (m *MockFormatter) GetEvents() []events.Event {
	m.mu.Lock()
	defer m.mu.Unlock()
	eventsCopy := make([]events.Event, len(m.Events))
	copy(eventsCopy, m.Events)
	return eventsCopy
}

func (m *MockFormatter) EventCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.Events)
}

func (m *MockFormatter) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Events = nil
}

var _ Formatter = (*MockFormatter)(nil)
