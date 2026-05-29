package tui

import (
	"testing"
)

func TestHighlightPattern_NonEmpty(t *testing.T) {
	code := `sound("bd sd hh").slow(2)`
	result := highlightPattern(code)
	if result == "" {
		t.Error("highlight should not be empty for valid code")
	}
}

func TestHighlightPattern_Empty(t *testing.T) {
	result := highlightPattern("")
	if result != "" {
		t.Error("highlight of empty string should be empty")
	}
}

func TestHighlightPattern_Drums(t *testing.T) {
	code := `sound("bd sd hh")`
	result := highlightPattern(code)
	// At minimum, the output should contain the original code (possibly wrapped in ANSI codes)
	if result == "" {
		t.Error("highlight output should not be empty")
	}
}