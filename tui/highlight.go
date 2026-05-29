package tui

import (
	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/charmbracelet/lipgloss"
)

var highlightStyle *chroma.Style

func init() {
	highlightStyle = styles.Get("monokai")
	if highlightStyle == nil {
		highlightStyle = styles.Fallback
	}
}

// highlightPattern applies syntax coloring to pattern source code.
func highlightPattern(code string) string {
	if code == "" {
		return ""
	}

	lexer := lexers.Get("go")
	if lexer == nil {
		return code
	}

	iterator, err := lexer.Tokenise(nil, code)
	if err != nil {
		return code
	}

	var result string
	for _, token := range iterator.Tokens() {
		entry := highlightStyle.Get(token.Type)
		if entry.IsZero() {
			result += token.Value
			continue
		}
		style := lipgloss.NewStyle().
			Foreground(lipgloss.Color(entry.Colour.String())).
			Bold(entry.Bold == chroma.Yes)
		result += style.Render(token.Value)
	}

	return result
}