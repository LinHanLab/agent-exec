package claude

import (
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
	tests := []struct {
		name        string
		iterations  int
		prompt      string
		wantErr     bool
		expectedErr string
	}{
		{
			name:        "valid arguments",
			iterations:  5,
			prompt:      "test prompt",
			wantErr:     false,
			expectedErr: "",
		},
		{
			name:        "zero iterations",
			iterations:  0,
			prompt:      "test prompt",
			wantErr:     true,
			expectedErr: "iterations must be a positive number",
		},
		{
			name:        "negative iterations",
			iterations:  -1,
			prompt:      "test prompt",
			wantErr:     true,
			expectedErr: "iterations must be a positive number",
		},
		{
			name:        "single iteration",
			iterations:  1,
			prompt:      "test prompt",
			wantErr:     false,
			expectedErr: "",
		},
		{
			name:        "empty prompt",
			iterations:  5,
			prompt:      "",
			wantErr:     true,
			expectedErr: "prompt cannot be empty",
		},
		{
			name:        "whitespace-only prompt",
			iterations:  5,
			prompt:      "   ",
			wantErr:     true,
			expectedErr: "prompt cannot be whitespace-only",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateLoopArgs(tt.iterations, tt.prompt)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateLoopArgs(%d, %q) = nil; want error", tt.iterations, tt.prompt)
					return
				}
				if !strings.Contains(err.Error(), tt.expectedErr) {
					t.Errorf("ValidateLoopArgs(%d, %q) error = %q; want to contain %q", tt.iterations, tt.prompt, err.Error(), tt.expectedErr)
				}
			} else if err != nil {
				t.Errorf("ValidateLoopArgs(%d, %q) = %v; want nil", tt.iterations, tt.prompt, err)
			}
		})
	}
}
