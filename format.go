package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

const (
	DisplayWidth   = 76
	PromptMaxLen   = 270
	TruncateSuffix = "[...Truncated]"
)

// Truncate string to max length with optional suffix (total output <= maxLen)
func truncateString(s string, maxLen int, suffix string) string {
	if len(s) <= maxLen {
		return s
	}

	suffixLen := len(suffix)
	if suffixLen >= maxLen {
		return s[:maxLen]
	}

	truncateAt := maxLen - suffixLen
	return s[:truncateAt] + suffix
}

// Print text with left bar prefix, handles line wrapping
func printPrefixed(text string, prefix string, totalWidth int) {
	prefixLen := len(prefix)
	contentWidth := totalWidth - prefixLen
	if contentWidth < 1 {
		contentWidth = 1
	}

	wrapped := wrapText(text, contentWidth)
	lines := strings.Split(wrapped, "\n")
	for _, line := range lines {
		fmt.Println(prefix + line)
	}
}

// Wrap text at word boundaries to fit within width
func wrapText(text string, width int) string {
	if width <= 0 {
		width = 1
	}

	var result strings.Builder
	lines := strings.Split(text, "\n")

	for i, line := range lines {
		if i > 0 {
			result.WriteString("\n")
		}

		if len(line) <= width {
			result.WriteString(line)
			continue
		}

		words := strings.Fields(line)
		currentLine := ""

		for _, word := range words {
			switch {
			case currentLine == "":
				if len(word) > width {
					// Word is longer than width, force break
					for len(word) > width {
						if currentLine != "" {
							result.WriteString(currentLine)
							result.WriteString("\n")
						}
						result.WriteString(word[:width])
						result.WriteString("\n")
						word = word[width:]
					}
					currentLine = word
				} else {
					currentLine = word
				}
			case len(currentLine)+1+len(word) <= width:
				currentLine += " " + word
			default:
				result.WriteString(currentLine)
				result.WriteString("\n")
				if len(word) > width {
					for len(word) > width {
						result.WriteString(word[:width])
						result.WriteString("\n")
						word = word[width:]
					}
					currentLine = word
				} else {
					currentLine = word
				}
			}
		}

		if currentLine != "" {
			result.WriteString(currentLine)
		}
	}

	return result.String()
}

// Validate prompt argument for runOneShot
func validatePrompt(prompt string) error {
	if prompt == "" {
		return errors.New("prompt cannot be empty")
	}

	if strings.TrimSpace(prompt) == "" {
		return errors.New("prompt cannot be whitespace-only")
	}

	return nil
}

// Validate arguments for runIterations
func validateIterationArgs(iterations int, sleepSeconds int, promptFiles []string) error {
	if iterations < 1 {
		return errors.New("iterations must be a positive number")
	}

	if sleepSeconds < 0 {
		return errors.New("sleep seconds must be a non-negative number")
	}

	if len(promptFiles) == 0 {
		return errors.New("at least one prompt file is required")
	}

	for _, promptFile := range promptFiles {
		info, err := os.Stat(promptFile)
		if os.IsNotExist(err) {
			return fmt.Errorf("prompt file not found: %s", promptFile)
		}
		if err != nil {
			return fmt.Errorf("error accessing prompt file: %s: %w", promptFile, err)
		}
		if info.Size() == 0 {
			return fmt.Errorf("prompt file is empty: %s", promptFile)
		}

		content, err := os.ReadFile(promptFile)
		if err != nil {
			return fmt.Errorf("error reading prompt file: %s: %w", promptFile, err)
		}
		if strings.TrimSpace(string(content)) == "" {
			return fmt.Errorf("prompt file contains only whitespace: %s", promptFile)
		}
	}

	return nil
}
