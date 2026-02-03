package display

import (
	"fmt"
	"strings"
)

const (
	// Content limits for non-verbose mode
	MaxCodeBlockLines = 10
	MaxCodeBlockChars = 5000
)

// ContentFilter handles content filtering and limiting
type ContentFilter struct {
	verbose bool
}

// NewContentFilter creates a new ContentFilter
func NewContentFilter(verbose bool) *ContentFilter {
	return &ContentFilter{
		verbose: verbose,
	}
}

// ToolInputFilter defines which fields to hide for specific tools
type ToolInputFilter struct {
	ToolName string
	Fields   []string
}

// Default filters for non-verbose mode
var defaultToolInputFilters = []ToolInputFilter{
	{
		ToolName: "Write",
		Fields:   []string{"content"},
	},
	{
		ToolName: "Edit",
		Fields:   []string{"new_string", "old_string"},
	},
}

// ApplyToolInputFilters replaces filtered field values with placeholder text
func (cf *ContentFilter) ApplyToolInputFilters(toolName string, input map[string]interface{}) map[string]interface{} {
	if cf.verbose {
		return input
	}

	// Create a copy to avoid modifying original
	filtered := make(map[string]interface{})
	for k, v := range input {
		filtered[k] = v
	}

	// Apply filters by replacing values with placeholder
	for _, filter := range defaultToolInputFilters {
		if filter.ToolName == toolName {
			for _, field := range filter.Fields {
				if _, exists := filtered[field]; exists {
					filtered[field] = "<hidden, use --verbose to see>"
				}
			}
		}
	}

	return filtered
}

// LimitCodeBlock truncates content to max lines and chars
func (cf *ContentFilter) LimitCodeBlock(content string) string {
	if cf.verbose {
		return content
	}

	lines := strings.Split(content, "\n")

	// Limit by line count
	if len(lines) > MaxCodeBlockLines {
		hiddenLines := len(lines) - MaxCodeBlockLines
		lines = lines[:MaxCodeBlockLines]
		lines = append(lines, fmt.Sprintf("... (%d more lines hidden, use --verbose to see all)", hiddenLines))
	}

	result := strings.Join(lines, "\n")

	// Limit by character count
	if len(result) > MaxCodeBlockChars {
		hiddenChars := len(content) - MaxCodeBlockChars
		result = result[:MaxCodeBlockChars] + fmt.Sprintf("\n... (%d more characters hidden, use --verbose to see all)", hiddenChars)
	}

	return result
}
