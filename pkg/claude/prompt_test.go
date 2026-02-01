package claude

import (
	"reflect"
	"testing"
)

func TestPromptOptions_BuildClaudeArgs(t *testing.T) {
	tests := []struct {
		name   string
		opts   *PromptOptions
		prompt string
		want   []string
	}{
		{
			name:   "no system prompts",
			opts:   &PromptOptions{},
			prompt: "test prompt",
			want:   []string{"--verbose", "--output-format", "stream-json", "-p", "test prompt"},
		},
		{
			name: "with system prompt",
			opts: &PromptOptions{
				SystemPrompt: "You are a helpful assistant",
			},
			prompt: "test prompt",
			want:   []string{"--verbose", "--output-format", "stream-json", "-p", "test prompt", "--system-prompt", "You are a helpful assistant"},
		},
		{
			name: "with append system prompt",
			opts: &PromptOptions{
				AppendSystemPrompt: "Focus on security",
			},
			prompt: "test prompt",
			want:   []string{"--verbose", "--output-format", "stream-json", "-p", "test prompt", "--append-system-prompt", "Focus on security"},
		},
		{
			name: "with both system prompts",
			opts: &PromptOptions{
				SystemPrompt:       "You are a helpful assistant",
				AppendSystemPrompt: "Focus on security",
			},
			prompt: "test prompt",
			want:   []string{"--verbose", "--output-format", "stream-json", "-p", "test prompt", "--system-prompt", "You are a helpful assistant", "--append-system-prompt", "Focus on security"},
		},
		{
			name: "empty strings are ignored",
			opts: &PromptOptions{
				SystemPrompt:       "",
				AppendSystemPrompt: "",
			},
			prompt: "test prompt",
			want:   []string{"--verbose", "--output-format", "stream-json", "-p", "test prompt"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.opts.BuildClaudeArgs(tt.prompt)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BuildClaudeArgs() = %v, want %v", got, tt.want)
			}
		})
	}
}
