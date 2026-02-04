package display

import (
	"strings"
	"testing"
)

func TestFrameBuilder_EmptyContent(t *testing.T) {
	fb := NewFrameBuilder()
	result := fb.Build("")
	if result != "" {
		t.Errorf("Expected empty string for empty content, got %q", result)
	}
}

func TestFrameBuilder_NoBorder(t *testing.T) {
	fb := NewFrameBuilder(
		WithContentWidth(20),
		WithIndent("    "),
	)
	result := fb.Build("hello world")

	// Should not contain box drawing characters
	if strings.Contains(result, "┌") || strings.Contains(result, "│") {
		t.Error("Expected no box drawing characters when box drawing is disabled")
	}

	// Should contain the content
	if !strings.Contains(result, "hello world") {
		t.Error("Expected output to contain content")
	}

	// Should be indented (after the initial newline)
	if !strings.Contains(result, "\n    ") {
		t.Error("Expected output to be indented")
	}
}

func TestFrameBuilder_WithBorder(t *testing.T) {
	fb := NewFrameBuilder(
		WithContentWidth(20),
		WithIndent("    "),
		WithBoxDrawing(),
	)
	result := fb.Build("hello world")

	// Should contain box drawing characters
	if !strings.Contains(result, "┌") {
		t.Error("Expected top-left corner character")
	}
	if !strings.Contains(result, "┐") {
		t.Error("Expected top-right corner character")
	}
	if !strings.Contains(result, "│") {
		t.Error("Expected vertical border character")
	}
	if !strings.Contains(result, "└") {
		t.Error("Expected bottom-left corner character")
	}
	if !strings.Contains(result, "┘") {
		t.Error("Expected bottom-right corner character")
	}

	// Should contain the content
	if !strings.Contains(result, "hello world") {
		t.Error("Expected output to contain content")
	}
}

func TestFrameBuilder_LineWrapping(t *testing.T) {
	fb := NewFrameBuilder(
		WithContentWidth(30),
		WithIndent(""),
		WithBoxDrawing(),
	)

	// Create a long string that will wrap
	longContent := "This is a very long line that should wrap"
	result := fb.Build(longContent)

	// The output should contain the full content
	stripped := stripANSI(result)
	// Just verify the content is present and wrapped (multiple lines)
	lines := strings.Split(stripped, "\n")
	// Should have: top border + wrapped content lines + bottom border = 3+ lines
	if len(lines) < 3 {
		t.Errorf("Expected wrapped content to produce multiple lines, got %d lines", len(lines))
	}
}

func TestFrameBuilder_WrappingAtWhitespace(t *testing.T) {
	fb := NewFrameBuilder(
		WithContentWidth(15),
		WithIndent(""),
		WithBoxDrawing(),
	)

	// Content with natural break points
	content := "hello world test"
	result := fb.Build(content)

	// Should wrap at spaces
	stripped := stripANSI(result)
	// The content should be present
	if !strings.Contains(stripped, "hello") || !strings.Contains(stripped, "world") || !strings.Contains(stripped, "test") {
		t.Error("Expected output to contain all words")
	}
}

func TestFrameBuilder_ColorApplied(t *testing.T) {
	fb := NewFrameBuilder(
		WithContentWidth(20),
		WithIndent("    "),
		WithTextColor(BoldCyan),
		WithBoxDrawing(),
	)
	result := fb.Build("hello world")

	// Should contain the color code
	if !strings.Contains(result, BoldCyan) {
		t.Error("Expected output to contain color code")
	}

	// Should contain the content
	if !strings.Contains(result, "hello world") {
		t.Error("Expected output to contain content")
	}
}

func TestFrameBuilder_MultipleLines(t *testing.T) {
	fb := NewFrameBuilder(
		WithContentWidth(30),
		WithIndent("    "),
		WithBoxDrawing(),
	)
	result := fb.Build("line1\nline2\nline3")

	// Should contain all lines
	stripped := stripANSI(result)
	if !strings.Contains(stripped, "line1") {
		t.Error("Expected output to contain line1")
	}
	if !strings.Contains(stripped, "line2") {
		t.Error("Expected output to contain line2")
	}
	if !strings.Contains(stripped, "line3") {
		t.Error("Expected output to contain line3")
	}
}

func TestFrameBuilder_FrameOptions(t *testing.T) {
	// Test that options are applied correctly
	fb := NewFrameBuilder(
		WithContentWidth(50),
		WithIndent("XX"),
		WithBoxDrawing(),
		WithTextColor(Green),
	)

	if fb.contentWidth != 50 {
		t.Errorf("Expected contentWidth 50, got %d", fb.contentWidth)
	}
	if fb.indent != "XX" {
		t.Errorf("Expected indent XX, got %q", fb.indent)
	}
	if fb.textColor != Green {
		t.Errorf("Expected textColor Green, got %q", fb.textColor)
	}
	if !fb.useBoxDrawing {
		t.Error("Expected useBoxDrawing to be true")
	}
}

func TestFrameBuilder_BorderAlignment(t *testing.T) {
	fb := NewFrameBuilder(
		WithContentWidth(50),
		WithIndent(""),
		WithBoxDrawing(),
	)

	testCases := []string{
		"Short",
		"Medium length content here",
		"This is a much longer line of text that should still align properly",
		"- Delete comments (following rules above)",
		"- Rename identifiers (variables, functions, classes, etc.)",
	}

	for _, content := range testCases {
		result := fb.Build(content)
		stripped := stripANSI(result)
		lines := strings.Split(stripped, "\n")

		var borderedLines []string
		for _, line := range lines {
			if strings.Contains(line, "│") || strings.Contains(line, "┌") || strings.Contains(line, "└") {
				borderedLines = append(borderedLines, line)
			}
		}

		if len(borderedLines) < 2 {
			t.Fatalf("Expected at least 2 bordered lines, got %d", len(borderedLines))
		}

		expectedLen := len([]rune(borderedLines[0]))
		for i, line := range borderedLines {
			lineLen := len([]rune(line))
			if lineLen != expectedLen {
				t.Errorf("Line %d length mismatch: expected %d, got %d\nContent: %q\nLine: %q",
					i, expectedLen, lineLen, content, line)
			}
		}
	}
}
