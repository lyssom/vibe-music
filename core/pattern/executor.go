package pattern

import "time"

// Executor parses pattern code and evaluates it into note events at a given time.
type Executor interface {
	Parse(code string) (*Pattern, error)
	Evaluate(p *Pattern, t time.Time) []Note
}

// TokenType identifies the kind of lexer token.
type TokenType int

const (
	TokIdent TokenType = iota // function/method name
	TokString                  // "quoted string"
	TokNumber                  // 123 or 1.5
	TokDot                     // .
	TokLParen                  // (
	TokRParen                  // )
	TokComma                   // ,
	TokMinus                   // -
	TokTimes                   // *
	TokEOF
)

// Token is a single lexical token.
type Token struct {
	Type  TokenType
	Value string
	Pos   int // byte offset in source
}

// Command is a parsed function/method call.
type Command struct {
	Name         string
	Args         []Arg
	Every        int // 0 = every beat, 1 = every 2 beats, etc.
	BeatOffset   int // which beat within the pattern (0-based)
	ChainNext    *Command
}

// Arg is a single argument to a command.
type Arg struct {
	Type  ArgType
	Value string
}

// ArgType distinguishes argument types.
type ArgType int

const (
	ArgString ArgType = iota
	ArgNumber
)

// AST is the parsed abstract syntax tree.
type AST struct {
	Commands    []Command
	SubCommands []Command // .every() commands that don't go in Commands (beat-filtered)
}

// NewCommand creates a command with the given name.
func NewCommand(name string, args ...string) Command {
	c := Command{Name: name}
	for _, a := range args {
		c.Args = append(c.Args, Arg{Type: ArgString, Value: a})
	}
	return c
}

// AddArg adds a string argument.
func (c *Command) AddArg(v string) {
	c.Args = append(c.Args, Arg{Type: ArgString, Value: v})
}

// SetEvery sets the beat divisor (every N beats).
func (c *Command) SetEvery(n int) {
	if n < 1 {
		n = 1
	}
	c.Every = n
}