package display

import (
	"strings"
	"testing"
)

func TestContentFilter_ApplyToolInputFilters(t *testing.T) {
	tests := []struct {
		name     string
		verbose  bool
		toolName string
		input    map[string]interface{}
		wantHide bool
	}{
		{
			name:     "Write tool non-verbose hides content",
			verbose:  false,
			toolName: "Write",
			input: map[string]interface{}{
				"file_path": "/test/file.txt",
				"content":   "secret content",
			},
			wantHide: true,
		},
		{
			name:     "Write tool verbose shows content",
			verbose:  true,
			toolName: "Write",
			input: map[string]interface{}{
				"file_path": "/test/file.txt",
				"content":   "secret content",
			},
			wantHide: false,
		},
		{
			name:     "Edit tool non-verbose hides strings",
			verbose:  false,
			toolName: "Edit",
			input: map[string]interface{}{
				"file_path":  "/test/file.txt",
				"old_string": "old",
				"new_string": "new",
			},
			wantHide: true,
		},
		{
			name:     "Other tool not filtered",
			verbose:  false,
			toolName: "Read",
			input: map[string]interface{}{
				"file_path": "/test/file.txt",
			},
			wantHide: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cf := NewContentFilter(tt.verbose)
			result := cf.ApplyToolInputFilters(tt.toolName, tt.input)

			if tt.wantHide {
				// Check that sensitive fields are hidden
				for key, val := range result {
					if key == "content" || key == "old_string" || key == "new_string" {
						if !strings.Contains(val.(string), "hidden") {
							t.Errorf("Expected field %s to be hidden, got %v", key, val)
						}
					}
				}
			} else {
				// Check that original values are preserved
				for key, val := range tt.input {
					if result[key] != val {
						t.Errorf("Expected field %s to be %v, got %v", key, val, result[key])
					}
				}
			}
		})
	}
}

func TestContentFilter_LimitCodeBlock(t *testing.T) {
	tests := []struct {
		name      string
		verbose   bool
		content   string
		wantLimit bool
		limitType string // "lines" or "chars"
	}{
		{
			name:      "short content not limited",
			verbose:   false,
			content:   "line1\nline2\nline3",
			wantLimit: false,
		},
		{
			name:      "long content limited by lines",
			verbose:   false,
			content:   strings.Repeat("line\n", 20),
			wantLimit: true,
			limitType: "lines",
		},
		{
			name:      "long content not limited in verbose mode",
			verbose:   true,
			content:   strings.Repeat("line\n", 20),
			wantLimit: false,
		},
		{
			name:      "long content limited by chars",
			verbose:   false,
			content:   strings.Repeat("x", 6000),
			wantLimit: true,
			limitType: "chars",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cf := NewContentFilter(tt.verbose)
			result := cf.LimitCodeBlock(tt.content)

			if tt.wantLimit {
				if !strings.Contains(result, "hidden") {
					t.Error("Expected content to be limited with 'hidden' message")
				}
				switch tt.limitType {
				case "lines":
					// For line-limited content, check that we have at most MaxCodeBlockLines + 1 (for the message)
					resultLines := strings.Split(result, "\n")
					if len(resultLines) > MaxCodeBlockLines+1 {
						t.Errorf("Expected result to have at most %d lines, got %d", MaxCodeBlockLines+1, len(resultLines))
					}
				case "chars":
					// For char-limited content, check that result is shorter
					if len(result) >= len(tt.content) {
						t.Error("Expected result to be shorter than input for char-limited content")
					}
				}
			} else if result != tt.content {
				t.Error("Expected content to be unchanged")
			}
		})
	}
}
