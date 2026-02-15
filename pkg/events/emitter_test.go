package events

import (
	"testing"
	"time"
)

func TestChannelEmitter_EmitAndSubscribe(t *testing.T) {
	emitter := NewChannelEmitter(10)
	defer emitter.Close()

	ch := emitter.Subscribe()

	testData := RunPromptStartedData{Prompt: "test prompt"}
	emitter.Emit(EventRunPromptStarted, testData)

	select {
	case event := <-ch:
		if event.Type != EventRunPromptStarted {
			t.Errorf("Expected event type %s, got %s", EventRunPromptStarted, event.Type)
		}
		data, ok := event.Data.(RunPromptStartedData)
		if !ok {
			t.Errorf("Expected RunPromptStartedData, got %T", event.Data)
		}
		if data.Prompt != "test prompt" {
			t.Errorf("Expected prompt 'test prompt', got '%s'", data.Prompt)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Timeout waiting for event")
	}
}

func TestChannelEmitter_MultipleEvents(t *testing.T) {
	emitter := NewChannelEmitter(10)
	defer emitter.Close()

	ch := emitter.Subscribe()

	events := []EventType{
		EventRunPromptStarted,
		EventClaudeAssistantMessage,
		EventClaudeToolUse,
	}

	for _, eventType := range events {
		emitter.Emit(eventType, nil)
	}

	for i, expectedType := range events {
		select {
		case event := <-ch:
			if event.Type != expectedType {
				t.Errorf("Event %d: expected type %s, got %s", i, expectedType, event.Type)
			}
		case <-time.After(100 * time.Millisecond):
			t.Errorf("Timeout waiting for event %d", i)
		}
	}
}

func TestChannelEmitter_Close(t *testing.T) {
	emitter := NewChannelEmitter(10)
	ch := emitter.Subscribe()

	emitter.Close()

	_, ok := <-ch
	if ok {
		t.Error("Expected channel to be closed")
	}

	emitter.Emit(EventRunPromptStarted, nil)
}

func TestChannelEmitter_EmitAfterClose(t *testing.T) {
	emitter := NewChannelEmitter(10)
	emitter.Close()

	emitter.Emit(EventRunPromptStarted, nil)
}

func TestNullEmitter_Emit(t *testing.T) {
	emitter := NewNullEmitter()
	emitter.Emit(EventRunPromptStarted, RunPromptStartedData{Prompt: "test"})
}

func TestNullEmitter_Subscribe(t *testing.T) {
	emitter := NewNullEmitter()
	ch := emitter.Subscribe()

	_, ok := <-ch
	if ok {
		t.Error("Expected NullEmitter channel to be closed")
	}
}

func TestNullEmitter_Close(t *testing.T) {
	emitter := NewNullEmitter()
	emitter.Close()
}

func TestEvent_Timestamp(t *testing.T) {
	emitter := NewChannelEmitter(10)
	defer emitter.Close()

	ch := emitter.Subscribe()

	before := time.Now()
	emitter.Emit(EventRunPromptStarted, nil)
	after := time.Now()

	select {
	case event := <-ch:
		if event.Timestamp.Before(before) || event.Timestamp.After(after) {
			t.Errorf("Event timestamp %v is outside expected range [%v, %v]",
				event.Timestamp, before, after)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Timeout waiting for event")
	}
}
