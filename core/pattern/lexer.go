package pattern

import (
	"fmt"
	"unicode"
)

// lexer tokenizes pattern source code.
type lexer struct {
	src   []rune
	pos   int
	token Token
}

// newLexer creates a lexer for the given source.
func newLexer(src string) *lexer {
	return &lexer{src: []rune(src)}
}

// next advances to the next token and returns it.
func (l *lexer) next() Token {
	l.skipWhitespace()

	if l.pos >= len(l.src) {
		l.token = Token{Type: TokEOF, Pos: l.pos}
		return l.token
	}

	ch := l.src[l.pos]

	switch {
	case ch == '(':
		l.token = Token{Type: TokLParen, Value: "(", Pos: l.pos}
		l.pos++
	case ch == ')':
		l.token = Token{Type: TokRParen, Value: ")", Pos: l.pos}
		l.pos++
	case ch == '.':
		l.token = Token{Type: TokDot, Value: ".", Pos: l.pos}
		l.pos++
	case ch == ',':
		l.token = Token{Type: TokComma, Value: ",", Pos: l.pos}
		l.pos++
	case ch == '"':
		l.token = l.readString()
	case ch == '-':
		l.token = Token{Type: TokMinus, Value: "-", Pos: l.pos}
		l.pos++
	case unicode.IsDigit(ch):
		l.token = l.readNumber()
	case ch == '#' || (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_' || ch >= 0x80:
		l.token = l.readIdent()
	default:
		l.token = Token{Type: TokEOF, Pos: l.pos}
		l.pos++
	}

	return l.token
}

func (l *lexer) peek() Token {
	saved := l.token
	savedPos := l.pos
	tok := l.next()
	l.token = saved
	l.pos = savedPos
	return tok
}

func (l *lexer) readString() Token {
	start := l.pos
	l.pos++ // skip opening quote
	val := make([]rune, 0)
	for l.pos < len(l.src) && l.src[l.pos] != '"' {
		if l.src[l.pos] == '\\' && l.pos+1 < len(l.src) {
			l.pos++
		}
		val = append(val, l.src[l.pos])
		l.pos++
	}
	if l.pos < len(l.src) {
		l.pos++ // skip closing quote
	}
	return Token{Type: TokString, Value: string(val), Pos: start}
}

func (l *lexer) readNumber() Token {
	start := l.pos
	hasDot := false
	for l.pos < len(l.src) {
		ch := l.src[l.pos]
		if unicode.IsDigit(ch) {
			l.pos++
		} else if ch == '.' && !hasDot {
			hasDot = true
			l.pos++
		} else {
			break
		}
	}
	return Token{Type: TokNumber, Value: string(l.src[start:l.pos]), Pos: start}
}

func (l *lexer) readIdent() Token {
	start := l.pos
	for l.pos < len(l.src) {
		ch := l.src[l.pos]
		if unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '_' || ch >= 0x80 {
			l.pos++
		} else {
			break
		}
	}
	return Token{Type: TokIdent, Value: string(l.src[start:l.pos]), Pos: start}
}

func (l *lexer) skipWhitespace() {
	for l.pos < len(l.src) && (unicode.IsSpace(l.src[l.pos]) || l.src[l.pos] == '|') {
		l.pos++
	}
}

// tokenName returns a human-readable name for a token type.
func tokenName(t TokenType) string {
	switch t {
	case TokIdent:
		return "identifier"
	case TokString:
		return "string"
	case TokNumber:
		return "number"
	case TokDot:
		return "."
	case TokLParen:
		return "("
	case TokRParen:
		return ")"
	case TokComma:
		return ","
	case TokMinus:
		return "-"
	case TokTimes:
		return "*"
	case TokEOF:
		return "end of input"
	default:
		return fmt.Sprintf("unknown(%d)", t)
	}
}