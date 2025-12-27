package claude

import (
	"errors"
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

// ValidateLoopArgs validates iteration arguments
func ValidateLoopArgs(iterations int, prompt string) error {
	if iterations < 1 {
		return errors.New("iterations must be a positive number")
	}

	return ValidatePrompt(prompt)
}
