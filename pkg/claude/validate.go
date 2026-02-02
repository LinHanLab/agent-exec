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
