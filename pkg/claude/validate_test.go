package claude

import (
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
