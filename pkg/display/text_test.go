package display

import (
	"strings"
	"testing"
	"time"
)

func TestTextFormatter_IndentContent(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "single line",
			input:    "hello",
			expected: "    hello",
		},
		{
			name:     "multiple lines",
			input:    "line1\nline2\nline3",
			expected: "    line1\n    line2\n    line3",
		},
	}

	tf := NewTextFormatter()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tf.IndentContent(tt.input)
			if result != tt.expected {
				t.Errorf("IndentContent() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestTextFormatter_FormatContent(t *testing.T) {
	tf := NewTextFormatter()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "single line",
			input:    "content",
			expected: "\n    content\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tf.FormatContent(tt.input)
			if result != tt.expected {
				t.Errorf("FormatContent() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestTextFormatter_FormatDuration(t *testing.T) {
	tf := NewTextFormatter()

	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{
			name:     "milliseconds",
			duration: 500 * time.Millisecond,
			expected: "500.0ms",
		},
		{
			name:     "seconds",
			duration: 2 * time.Second,
			expected: "2.0s",
		},
		{
			name:     "minutes and seconds",
			duration: 2*time.Minute + 30*time.Second,
			expected: "2m 30s",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tf.FormatDuration(tt.duration)
			if result != tt.expected {
				t.Errorf("FormatDuration() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestTextFormatter_FormatTime(t *testing.T) {
	tf := NewTextFormatter()
	result := tf.FormatTime()

	// Just verify it returns a string in HH:MM:SS format
	parts := strings.Split(result, ":")
	if len(parts) != 3 {
		t.Errorf("FormatTime() = %q, expected HH:MM:SS format", result)
	}
}

func TestTextFormatter_ApplyReverseVideo(t *testing.T) {
	tf := NewTextFormatter()

	tests := []struct {
		name  string
		text  string
		color string
	}{
		{
			name:  "with color",
			text:  "test",
			color: BoldYellow,
		},
		{
			name:  "without color",
			text:  "test",
			color: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tf.ApplyReverseVideo(tt.text, tt.color)
			if !strings.Contains(result, ReverseVideo) {
				t.Error("Expected result to contain ReverseVideo code")
			}
			if !strings.Contains(result, Reset) {
				t.Error("Expected result to contain Reset code")
			}
			if !strings.Contains(result, tt.text) {
				t.Error("Expected result to contain original text")
			}
		})
	}
}
