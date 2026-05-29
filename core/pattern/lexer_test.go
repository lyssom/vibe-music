package pattern

import (
	"testing"
)

func lexAll(src string) []Token {
	l := newLexer(src)
	var tokens []Token
	for {
		tok := l.next()
		tokens = append(tokens, tok)
		if tok.Type == TokEOF {
			break
		}
	}
	return tokens
}

func TestLexer_BasicDrumPattern(t *testing.T) {
	src := `sound("bd sd hh")`
	tokens := lexAll(src)

	expected := []struct {
		typ   TokenType
		value string
	}{
		{TokIdent, "sound"},
		{TokLParen, "("},
		{TokString, "bd sd hh"},
		{TokRParen, ")"},
		{TokEOF, ""},
	}

	if len(tokens) != len(expected) {
		t.Fatalf("got %d tokens, want %d: %v", len(tokens), len(expected), tokens)
	}

	for i, exp := range expected {
		if tokens[i].Type != exp.typ {
			t.Errorf("token[%d].Type = %s, want %s (value=%q)", i, tokenName(tokens[i].Type), tokenName(exp.typ), tokens[i].Value)
		}
		if tokens[i].Value != exp.value {
			t.Errorf("token[%d].Value = %q, want %q", i, tokens[i].Value, exp.value)
		}
	}
}

func TestLexer_MethodChain(t *testing.T) {
	src := `sound("bd sd").note("c2 e2").every(2, "fast 2")`
	tokens := lexAll(src)

	types := []TokenType{
		TokIdent, TokLParen, TokString, TokRParen,
		TokDot, TokIdent, TokLParen, TokString, TokRParen,
		TokDot, TokIdent, TokLParen, TokNumber, TokComma, TokString, TokRParen,
		TokEOF,
	}

	if len(tokens) != len(types) {
		t.Fatalf("got %d tokens, want %d", len(tokens), len(types))
	}

	for i, wantType := range types {
		if tokens[i].Type != wantType {
			t.Errorf("token[%d] type = %s, want %s (value=%q)", i, tokenName(tokens[i].Type), tokenName(wantType), tokens[i].Value)
		}
	}
}

func TestLexer_WhitespaceHandling(t *testing.T) {
	src := "sound  (  \"bd\" , 120 )"
	tokens := lexAll(src)
	if len(tokens) != 7 { // ident, (, string, comma, number, ), EOF
		t.Fatalf("got %d tokens: %v", len(tokens), tokens)
	}
	if tokens[2].Value != "bd" {
		t.Errorf("string value = %q, want 'bd'", tokens[2].Value)
	}
}

func TestLexer_EscapedString(t *testing.T) {
	src := `sound("bd\\")`
	tokens := lexAll(src)
	if tokens[2].Value != `bd\` {
		t.Errorf("escaped string value = %q, want 'bd\\'", tokens[2].Value)
	}
}

func TestLexer_EmptyInput(t *testing.T) {
	tokens := lexAll("")
	if len(tokens) != 1 || tokens[0].Type != TokEOF {
		t.Errorf("empty input should produce single EOF token, got %v", tokens)
	}
}