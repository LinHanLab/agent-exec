package claude

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidatePrompt(t *testing.T) {
	tests := []struct {
		name        string
		prompt      string
		wantErr     bool
		expectedErr string
	}{
		{
			name:        "valid prompt",
			prompt:      "this is a valid prompt",
			wantErr:     false,
			expectedErr: "",
		},
		{
			name:        "empty prompt",
			prompt:      "",
			wantErr:     true,
			expectedErr: "prompt cannot be empty",
		},
		{
			name:        "whitespace-only prompt",
			prompt:      "   ",
			wantErr:     true,
			expectedErr: "prompt cannot be whitespace-only",
		},
		{
			name:        "tabs only",
			prompt:      "\t\t\t",
			wantErr:     true,
			expectedErr: "prompt cannot be whitespace-only",
		},
		{
			name:        "mixed whitespace",
			prompt:      " \t \n ",
			wantErr:     true,
			expectedErr: "prompt cannot be whitespace-only",
		},
		{
			name:        "single character",
			prompt:      "a",
			wantErr:     false,
			expectedErr: "",
		},
		{
			name:        "prompt with newlines",
			prompt:      "line one\nline two\nline three",
			wantErr:     false,
			expectedErr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePrompt(tt.prompt)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidatePrompt(%q) = nil; want error", tt.prompt)
					return
				}
				if err.Error() != tt.expectedErr {
					t.Errorf("ValidatePrompt(%q) error = %q; want %q", tt.prompt, err.Error(), tt.expectedErr)
				}
			} else if err != nil {
				t.Errorf("ValidatePrompt(%q) = %v; want nil", tt.prompt, err)
			}
		})
	}
}

func TestValidateLoopArgs(t *testing.T) {
	// Create temporary test files
	tmpDir := t.TempDir()

	validFile := filepath.Join(tmpDir, "valid.txt")
	if err := os.WriteFile(validFile, []byte("test prompt content"), 0644); err != nil {
		t.Fatal(err)
	}

	emptyFile := filepath.Join(tmpDir, "empty.txt")
	if err := os.WriteFile(emptyFile, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	whitespaceFile := filepath.Join(tmpDir, "whitespace.txt")
	if err := os.WriteFile(whitespaceFile, []byte("   \n\t   "), 0644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name           string
		iterations     int
		sleepSeconds   int
		promptFiles    []string
		wantErr        bool
		expectedErr    string
		expectedLength int
	}{
		{
			name:           "valid arguments",
			iterations:     5,
			sleepSeconds:   10,
			promptFiles:    []string{validFile},
			wantErr:        false,
			expectedErr:    "",
			expectedLength: 1,
		},
		{
			name:         "zero iterations",
			iterations:   0,
			sleepSeconds: 10,
			promptFiles:  []string{validFile},
			wantErr:      true,
			expectedErr:  "iterations must be a positive number",
		},
		{
			name:         "negative iterations",
			iterations:   -1,
			sleepSeconds: 10,
			promptFiles:  []string{validFile},
			wantErr:      true,
			expectedErr:  "iterations must be a positive number",
		},
		{
			name:         "negative sleep seconds",
			iterations:   5,
			sleepSeconds: -1,
			promptFiles:  []string{validFile},
			wantErr:      true,
			expectedErr:  "sleep seconds must be a non-negative number",
		},
		{
			name:         "zero sleep seconds",
			iterations:   5,
			sleepSeconds: 0,
			promptFiles:  []string{validFile},
			wantErr:      false,
			expectedErr:  "",
		},
		{
			name:         "no prompt files",
			iterations:   5,
			sleepSeconds: 10,
			promptFiles:  []string{},
			wantErr:      true,
			expectedErr:  "at least one prompt file is required",
		},
		{
			name:         "non-existent prompt file",
			iterations:   5,
			sleepSeconds: 10,
			promptFiles:  []string{filepath.Join(tmpDir, "nonexistent.txt")},
			wantErr:      true,
			expectedErr:  "prompt file not found:",
		},
		{
			name:         "empty prompt file",
			iterations:   5,
			sleepSeconds: 10,
			promptFiles:  []string{emptyFile},
			wantErr:      true,
			expectedErr:  "prompt file is empty:",
		},
		{
			name:         "whitespace-only prompt file",
			iterations:   5,
			sleepSeconds: 10,
			promptFiles:  []string{whitespaceFile},
			wantErr:      true,
			expectedErr:  "prompt file contains only whitespace:",
		},
		{
			name:           "multiple valid prompt files",
			iterations:     5,
			sleepSeconds:   10,
			promptFiles:    []string{validFile, validFile},
			wantErr:        false,
			expectedErr:    "",
			expectedLength: 2,
		},
		{
			name:         "mix of valid and invalid files",
			iterations:   5,
			sleepSeconds: 10,
			promptFiles:  []string{validFile, emptyFile},
			wantErr:      true,
			expectedErr:  "prompt file is empty:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ValidateLoopArgs(tt.iterations, tt.sleepSeconds, tt.promptFiles)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateLoopArgs(%d, %d, %v) = nil; want error", tt.iterations, tt.sleepSeconds, tt.promptFiles)
					return
				}
				if !strings.Contains(err.Error(), tt.expectedErr) {
					t.Errorf("ValidateLoopArgs(%d, %d, %v) error = %q; want to contain %q", tt.iterations, tt.sleepSeconds, tt.promptFiles, err.Error(), tt.expectedErr)
				}
			} else if err != nil {
				t.Errorf("ValidateLoopArgs(%d, %d, %v) = %v; want nil", tt.iterations, tt.sleepSeconds, tt.promptFiles, err)
			}
			if tt.expectedLength > 0 && len(result) != tt.expectedLength {
				t.Errorf("ValidateLoopArgs(%d, %d, %v) returned %d items; want %d", tt.iterations, tt.sleepSeconds, tt.promptFiles, len(result), tt.expectedLength)
			}
		})
	}
}
