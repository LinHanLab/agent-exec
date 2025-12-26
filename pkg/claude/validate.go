package claude

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// ValidatePrompt checks if a prompt is valid
func ValidatePrompt(prompt string) error {
	if prompt == "" {
		return errors.New("prompt cannot be empty")
	}

	if strings.TrimSpace(prompt) == "" {
		return errors.New("prompt cannot be whitespace-only")
	}

	return nil
}

// ValidateLoopArgs validates iteration arguments and returns prompt contents
// Returns validated prompt contents to avoid duplicate file reads
func ValidateLoopArgs(iterations int, sleepSeconds int, promptFiles []string) ([]string, error) {
	if iterations < 1 {
		return nil, errors.New("iterations must be a positive number")
	}

	if sleepSeconds < 0 {
		return nil, errors.New("sleep seconds must be a non-negative number")
	}

	if len(promptFiles) == 0 {
		return nil, errors.New("at least one prompt file is required")
	}

	promptContents := make([]string, 0, len(promptFiles))

	for _, promptFile := range promptFiles {
		info, err := os.Stat(promptFile)
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("prompt file not found: %s", promptFile)
		}
		if err != nil {
			return nil, fmt.Errorf("error accessing prompt file: %s: %w", promptFile, err)
		}
		if info.Size() == 0 {
			return nil, fmt.Errorf("prompt file is empty: %s", promptFile)
		}

		content, err := os.ReadFile(promptFile)
		if err != nil {
			return nil, fmt.Errorf("error reading prompt file: %s: %w", promptFile, err)
		}
		if strings.TrimSpace(string(content)) == "" {
			return nil, fmt.Errorf("prompt file contains only whitespace: %s", promptFile)
		}

		promptContents = append(promptContents, string(content))
	}

	return promptContents, nil
}
