package display

import (
	"fmt"
	"strings"
)

const (
	MaxCodeBlockLines = 10
	MaxCodeBlockChars = 5000
)

type ContentFilter struct {
	verbose bool
}

func NewContentFilter(verbose bool) *ContentFilter {
	return &ContentFilter{
		verbose: verbose,
	}
}

type ToolInputFilter struct {
	ToolName string
	Fields   []string
}

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

func (cf *ContentFilter) ApplyToolInputFilters(toolName string, input map[string]interface{}) map[string]interface{} {
	if cf.verbose {
		return input
	}

	filtered := make(map[string]interface{})
	for k, v := range input {
		filtered[k] = v
	}

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

func (cf *ContentFilter) LimitCodeBlock(content string) string {
	if cf.verbose {
		return content
	}

	lines := strings.Split(content, "\n")

	if len(lines) > MaxCodeBlockLines {
		hiddenLines := len(lines) - MaxCodeBlockLines
		lines = lines[:MaxCodeBlockLines]
		lines = append(lines, fmt.Sprintf("... (%d more lines hidden, use --verbose to see all)", hiddenLines))
	}

	result := strings.Join(lines, "\n")

	if len(result) > MaxCodeBlockChars {
		hiddenChars := len(content) - MaxCodeBlockChars
		result = result[:MaxCodeBlockChars] + fmt.Sprintf("\n... (%d more characters hidden, use --verbose to see all)", hiddenChars)
	}

	return result
}
