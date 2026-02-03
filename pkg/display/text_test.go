package display

import (
	"strings"
	"testing"
	"time"
)

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

func TestTextFormatter_FormatContentWithFrame(t *testing.T) {
	tf := NewTextFormatter()

	t.Run("with box border", func(t *testing.T) {
		tests := []struct {
			name     string
			input    string
			validate func(t *testing.T, result string)
		}{
			{
				name:  "empty string",
				input: "",
				validate: func(t *testing.T, result string) {
					if result != "" {
						t.Errorf("Expected empty result for empty input, got %q", result)
					}
				},
			},
			{
				name:  "single line",
				input: "hello world",
				validate: func(t *testing.T, result string) {
					if !strings.Contains(result, "┌") {
						t.Error("Expected top border with ┌")
					}
					if !strings.Contains(result, "└") {
						t.Error("Expected bottom border with └")
					}
					if !strings.Contains(result, "│") {
						t.Error("Expected side borders with │")
					}
					if !strings.Contains(result, "hello world") {
						t.Error("Expected content to be present")
					}
				},
			},
			{
				name:  "multiple lines",
				input: "line1\nline2\nline3",
				validate: func(t *testing.T, result string) {
					lines := strings.Split(result, "\n")
					// Should have: newline, top border, 3 content lines, bottom border, final newline
					if len(lines) < 6 {
						t.Errorf("Expected at least 6 lines, got %d", len(lines))
					}
					if !strings.Contains(result, "line1") {
						t.Error("Expected line1 to be present")
					}
					if !strings.Contains(result, "line2") {
						t.Error("Expected line2 to be present")
					}
					if !strings.Contains(result, "line3") {
						t.Error("Expected line3 to be present")
					}
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := tf.FormatContentWithFrame(tt.input, true)
				tt.validate(t, result)
			})
		}
	})

	t.Run("without box border (default - whitespace borders)", func(t *testing.T) {
		tests := []struct {
			name     string
			input    string
			validate func(t *testing.T, result string)
		}{
			{
				name:  "empty string",
				input: "",
				validate: func(t *testing.T, result string) {
					if result != "" {
						t.Errorf("Expected empty result for empty input, got %q", result)
					}
				},
			},
			{
				name:  "single line",
				input: "hello world",
				validate: func(t *testing.T, result string) {
					// Should NOT contain box drawing characters
					if strings.Contains(result, "┌") || strings.Contains(result, "└") || strings.Contains(result, "│") {
						t.Error("Should not contain box drawing characters")
					}
					// Should contain the content
					if !strings.Contains(result, "hello world") {
						t.Error("Expected content to be present")
					}
					// Should have frame structure (newlines, padding)
					lines := strings.Split(result, "\n")
					if len(lines) < 3 {
						t.Errorf("Expected at least 3 lines (top, content, bottom), got %d", len(lines))
					}
				},
			},
			{
				name:  "multiple lines",
				input: "line1\nline2\nline3",
				validate: func(t *testing.T, result string) {
					// Should NOT contain box drawing characters
					if strings.Contains(result, "┌") || strings.Contains(result, "└") || strings.Contains(result, "│") {
						t.Error("Should not contain box drawing characters")
					}
					// Should contain all content lines
					if !strings.Contains(result, "line1") {
						t.Error("Expected line1 to be present")
					}
					if !strings.Contains(result, "line2") {
						t.Error("Expected line2 to be present")
					}
					if !strings.Contains(result, "line3") {
						t.Error("Expected line3 to be present")
					}
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := tf.FormatContentWithFrame(tt.input)
				tt.validate(t, result)
			})
		}
	})
}

func TestTextFormatter_FormatContentWithFrameAndColor(t *testing.T) {
	tf := NewTextFormatter()

	t.Run("color preserved across wrapped lines", func(t *testing.T) {
		// Create a long line that will be wrapped
		longText := "I'll help you update the scripts to use the new agent-exec command structure. Let me start by reading the help message for agent-exec to understand the new command structure and available options."
		color := Magenta

		result := tf.FormatContentWithFrameAndColor(longText, color)

		// Count occurrences of the color code and reset code
		colorCount := strings.Count(result, color)
		resetCount := strings.Count(result, Reset)

		// Each content line should have color applied
		// The exact count depends on wrapping, but there should be multiple occurrences
		if colorCount < 2 {
			t.Errorf("Expected color code to appear at least twice (for wrapped lines), got %d", colorCount)
		}

		// Each colored segment should have a corresponding reset
		// We subtract gray color/reset pairs for borders
		grayCount := strings.Count(result, Gray)
		contentColorCount := colorCount
		contentResetCount := resetCount - grayCount

		if contentResetCount < contentColorCount {
			t.Errorf("Expected at least as many reset codes as color codes, got color=%d reset=%d", contentColorCount, contentResetCount)
		}

		// Verify the original text is still present (without color codes)
		if !strings.Contains(result, "I'll help you update") {
			t.Error("Expected original text to be present")
		}
		if !strings.Contains(result, "start by reading") {
			t.Error("Expected wrapped portion to be present")
		}
	})

	t.Run("color not applied when empty", func(t *testing.T) {
		text := "hello world"
		result := tf.FormatContentWithFrameAndColor(text, "")

		// Should not contain any Magenta color codes
		if strings.Contains(result, Magenta) {
			t.Error("Should not contain color when color parameter is empty")
		}
	})
}
