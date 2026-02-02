package display

import (
	"testing"
)

func TestTruncate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		suffix   string
		expected string
	}{
		{
			name:     "string shorter than max length",
			input:    "short",
			maxLen:   10,
			suffix:   "[...]",
			expected: "short",
		},
		{
			name:     "string equal to max length",
			input:    "exactlyten",
			maxLen:   10,
			suffix:   "[...]",
			expected: "exactlyten",
		},
		{
			name:     "string longer than max length with suffix",
			input:    "this is a long string",
			maxLen:   10,
			suffix:   "[...]",
			expected: "this [...]",
		},
		{
			name:     "empty string",
			input:    "",
			maxLen:   10,
			suffix:   "[...]",
			expected: "",
		},
		{
			name:     "max length is zero",
			input:    "some text",
			maxLen:   0,
			suffix:   "[...]",
			expected: "",
		},
		{
			name:     "suffix longer than max length",
			input:    "text",
			maxLen:   3,
			suffix:   "verylongsuffix",
			expected: "tex",
		},
		{
			name:     "suffix is empty",
			input:    "this is a long string",
			maxLen:   10,
			suffix:   "",
			expected: "this is a ",
		},
		{
			name:     "single character string",
			input:    "a",
			maxLen:   10,
			suffix:   "[...]",
			expected: "a",
		},
		{
			name:     "long word that needs truncating",
			input:    "supercalifragilisticexpialidocious",
			maxLen:   20,
			suffix:   "...",
			expected: "supercalifragilis...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Truncate(tt.input, tt.maxLen, tt.suffix)
			if result != tt.expected {
				t.Errorf("Truncate(%q, %d, %q) = %q; want %q", tt.input, tt.maxLen, tt.suffix, result, tt.expected)
			}
		})
	}
}

func TestWrap(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		width    int
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			width:    80,
			expected: "",
		},
		{
			name:     "width is zero",
			input:    "some text",
			width:    0,
			expected: "s\no\nm\ne\nt\ne\nx\nt",
		},
		{
			name:     "width is negative",
			input:    "some text",
			width:    -1,
			expected: "s\no\nm\ne\nt\ne\nx\nt",
		},
		{
			name:     "single word shorter than width",
			input:    "hello",
			width:    80,
			expected: "hello",
		},
		{
			name:     "multiple words that fit on one line",
			input:    "hello world",
			width:    80,
			expected: "hello world",
		},
		{
			name:     "text needs wrapping at word boundary",
			input:    "this is a long line of text that needs to be wrapped",
			width:    20,
			expected: "this is a long line\nof text that needs\nto be wrapped",
		},
		{
			name:     "text with explicit newlines",
			input:    "line one\nline two\nline three",
			width:    80,
			expected: "line one\nline two\nline three",
		},
		{
			name:     "text with explicit newlines and wrapping needed",
			input:    "line one with long words\nline two\nline three",
			width:    15,
			expected: "line one with\nlong words\nline two\nline three",
		},
		{
			name:     "single word longer than width",
			input:    "supercalifragilisticexpialidocious",
			width:    10,
			expected: "supercalif\nragilistic\nexpialidoc\nious",
		},
		{
			name:     "word longer than width in context",
			input:    "short supercalifragilisticexpialidocious word",
			width:    15,
			expected: "short\nsupercalifragil\nisticexpialidoc\nious word",
		},
		{
			name:     "multiple spaces between words",
			input:    "word1    word2    word3",
			width:    20,
			expected: "word1 word2 word3",
		},
		{
			name:     "single character words",
			input:    "a b c d e f g h",
			width:    3,
			expected: "a b\nc d\ne f\ng h",
		},
		{
			name:     "width of 1 forces character wrapping",
			input:    "hello",
			width:    1,
			expected: "h\ne\nl\nl\no",
		},
		{
			name:     "preserves leading/trailing spaces",
			input:    "  text with spaces  ",
			width:    80,
			expected: "  text with spaces  ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Wrap(tt.input, tt.width)
			if result != tt.expected {
				t.Errorf("Wrap(%q, %d) = %q; want %q", tt.input, tt.width, result, tt.expected)
			}
		})
	}
}
