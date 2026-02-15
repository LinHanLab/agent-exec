package display

import "regexp"

// stripANSI removes ANSI color codes from a string for testing
func stripANSI(s string) string {
	// ANSI escape sequence pattern (matches \x1B[...m)
	ansiPattern := regexp.MustCompile(`\x1B\[[0-9;]*[mGKHF]`)
	return ansiPattern.ReplaceAllString(s, "")
}
