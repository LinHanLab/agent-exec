package display

import (
	"strings"
)

// Truncate string to max length with optional suffix (total output <= maxLen)
func Truncate(s string, maxLen int, suffix string) string {
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

// Wrap text at word boundaries to fit within width
func Wrap(text string, width int) string {
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
