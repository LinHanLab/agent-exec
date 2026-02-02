package loop

import (
	"strings"
	"testing"
)

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
