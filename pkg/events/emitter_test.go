package events

import (
	"testing"
	"time"
)

func TestChannelEmitter_EmitAndSubscribe(t *testing.T) {
	emitter := NewChannelEmitter(10)
	defer emitter.Close()

	ch := emitter.Subscribe()

	testData := PromptStartedData{Prompt: "test prompt"}
	emitter.Emit(EventPromptStarted, testData)

	select {
	case event := <-ch:
		if event.Type != EventPromptStarted {
			t.Errorf("Expected event type %s, got %s", EventPromptStarted, event.Type)
		}
		data, ok := event.Data.(PromptStartedData)
		if !ok {
			t.Errorf("Expected PromptStartedData, got %T", event.Data)
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
		EventPromptStarted,
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

	emitter.Emit(EventPromptStarted, nil)
}

func TestChannelEmitter_EmitAfterClose(t *testing.T) {
	emitter := NewChannelEmitter(10)
	emitter.Close()

	emitter.Emit(EventPromptStarted, nil)
}

func TestNullEmitter_Emit(t *testing.T) {
	emitter := NewNullEmitter()
	emitter.Emit(EventPromptStarted, PromptStartedData{Prompt: "test"})
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
	emitter.Emit(EventPromptStarted, nil)
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
