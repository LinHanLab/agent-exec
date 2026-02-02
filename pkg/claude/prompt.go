package claude

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/LinHanLab/agent-exec/pkg/events"
)

// PromptOptions holds optional configuration for running prompts
type PromptOptions struct {
	SystemPrompt       string // Replace entire system prompt (empty = use defaults)
	AppendSystemPrompt string // Append to default system prompt (empty = use defaults)
}

// BuildClaudeArgs constructs the claude CLI arguments based on options
func (opts *PromptOptions) BuildClaudeArgs(prompt string) []string {
	args := []string{"--verbose", "--output-format", "stream-json", "-p", prompt}

	if opts.SystemPrompt != "" {
		args = append(args, "--system-prompt", opts.SystemPrompt)
	}
	if opts.AppendSystemPrompt != "" {
		args = append(args, "--append-system-prompt", opts.AppendSystemPrompt)
	}

	return args
}

// getCwdInfo retrieves current working directory and file list with error handling
func getCwdInfo(emitter events.Emitter) (cwd, fileList string, err error) {
	cwd, err = os.Getwd()
	if err != nil {
		return "", "", fmt.Errorf("failed to get cwd: %w", err)
	}

	files, err := os.ReadDir(cwd)
	if err != nil {
		return "", "", fmt.Errorf("failed to read cwd files: %w", err)
	}

	var names []string
	for _, f := range files {
		names = append(names, f.Name())
	}
	fileList = " [" + strings.Join(names, ", ") + "]"

	return
}

// RunPrompt executes a single prompt with claude CLI and returns the final result text
func RunPrompt(prompt string, opts *PromptOptions, emitter events.Emitter) (string, error) {
	if err := ValidatePrompt(prompt); err != nil {
		return "", err
	}

	cwd, fileList, err := getCwdInfo(emitter)
	if err != nil {
		return "", err
	}

	emitter.Emit(events.EventRunPromptStarted, events.RunPromptStartedData{
		Prompt:   prompt,
		BaseURL:  os.Getenv("ANTHROPIC_BASE_URL"),
		Cwd:      cwd,
		FileList: fileList,
	})

	if opts == nil {
		opts = &PromptOptions{}
	}
	args := opts.BuildClaudeArgs(prompt)
	cmd := exec.Command("claude", args...)
	cmd.Stderr = os.Stderr

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("failed to start claude CLI: %w", err)
	}

	result, parseErr := ParseStreamJSON(stdout, emitter)
	if parseErr != nil {
		_ = cmd.Wait()
		return "", parseErr
	}

	if err := cmd.Wait(); err != nil {
		return "", fmt.Errorf("claude CLI failed: %w", err)
	}

	return result, nil
}
