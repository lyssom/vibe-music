package pattern

import "fmt"

// parser implements a recursive descent parser for the pattern DSL.
type parser struct {
	lex       *lexer
	beatCount int // used for beat offset assignment
}

// Parse compiles pattern source code into an AST.
func Parse(src string) (*AST, error) {
	p := &parser{}
	p.lex = newLexer(src)
	p.lex.next() // prime first token
	return p.parsePattern()
}

// parsePattern: command ((dot | newline) command)*
func (p *parser) parsePattern() (*AST, error) {
	ast := &AST{}

	_, err := p.parseCommand(ast)
	if err != nil {
		return nil, err
	}

	// Accept dot-separated or newline-separated commands.
	for p.lex.token.Type == TokDot || p.lex.token.Type == TokIdent {
		if p.lex.token.Type == TokDot {
			p.lex.next()
		}
		_, err := p.parseCommand(ast)
		if err != nil {
			return nil, err
		}
	}

	if p.lex.token.Type != TokEOF {
		return nil, fmt.Errorf("unexpected token %s at position %d", tokenName(p.lex.token.Type), p.lex.token.Pos)
	}

	return ast, nil
}

// parseCommand: ident ( args ) [.modifier]*.
// Modifiers (.every, .slow, .fast, .swing) chain onto the current command.
func (p *parser) parseCommand(ast *AST) (bool, error) {
	if p.lex.token.Type != TokIdent {
		return false, fmt.Errorf("expected function name, got %s at position %d", tokenName(p.lex.token.Type), p.lex.token.Pos)
	}
	cmd := Command{Name: p.lex.token.Value, Every: 1}
	p.lex.next()

	if p.lex.token.Type != TokLParen {
		return false, fmt.Errorf("expected '(' after %s, got %s at position %d", cmd.Name, tokenName(p.lex.token.Type), p.lex.token.Pos)
	}
	p.lex.next()

	if p.lex.token.Type != TokRParen {
		args, err := p.parseArgs()
		if err != nil {
			return false, err
		}
		cmd.Args = args
	}

	if p.lex.token.Type != TokRParen {
		return false, fmt.Errorf("expected ')' after arguments, got %s at position %d", tokenName(p.lex.token.Type), p.lex.token.Pos)
	}
	p.lex.next()

	p.assignBeat(&cmd, ast)
	ast.Commands = append(ast.Commands, cmd)

	// Chainable modifiers: .every(N) .swing(N) .slow(N) .fast(N) .iter(N, "...")
	for p.lex.token.Type == TokDot {
		p.lex.next()
		if p.lex.token.Type != TokIdent {
			return false, fmt.Errorf("expected modifier name after '.', got %s", tokenName(p.lex.token.Type))
		}
		modName := p.lex.token.Value
		p.lex.next()

		if p.lex.token.Type != TokLParen {
			return false, fmt.Errorf("expected '(' after .%s, got %s", modName, tokenName(p.lex.token.Type))
		}
		p.lex.next()

		// Get the command being modified (last one in Commands)
		lastIdx := len(ast.Commands) - 1
		targ := &ast.Commands[lastIdx]

		switch modName {
		case "every":
			every, err := p.parseIntArg()
			if err != nil {
				return false, err
			}
			if every < 1 {
				every = 1
			}
			targ.Every *= every
			targ.BeatOffset = p.beatCount % targ.Every

		case "swing":
			swing, err := p.parseIntArg()
			if err != nil {
				return false, err
			}
			targ.Args = append(targ.Args, Arg{Type: ArgNumber, Value: fmt.Sprintf("%d", swing)})

		case "slow":
			slow, err := p.parseIntArg()
			if err != nil {
				return false, err
			}
			if slow < 1 {
				slow = 1
			}
			targ.Every *= slow
			targ.BeatOffset = p.beatCount % targ.Every

		case "fast":
			fast, err := p.parseIntArg()
			if err != nil {
				return false, err
			}
			if fast < 1 {
				fast = 1
			}
			targ.Every /= fast
			if targ.Every < 1 {
				targ.Every = 1
			}

		case "iter":
			iter, err := p.parseIntArg()
			if err != nil {
				return false, err
			}
			if p.lex.token.Type == TokComma {
				p.lex.next()
			}
			subPattern, err := p.parseStringArg()
			if err != nil {
				return false, err
			}
			for i := 0; i < iter; i++ {
				sub, err := Parse(subPattern)
				if err != nil {
					return false, err
				}
				for _, sc := range sub.Commands {
					p.assignBeat(&sc, ast)
					ast.SubCommands = append(ast.SubCommands, sc)
				}
			}

		default:
			return false, fmt.Errorf("unknown modifier '.%s'", modName)
		}

		if p.lex.token.Type != TokRParen {
			return false, fmt.Errorf("expected ')' after modifier arguments, got %s", tokenName(p.lex.token.Type))
		}
		p.lex.next()
	}

	p.beatCount++
	return true, nil
}

// assignBeat assigns beat offset.
func (p *parser) assignBeat(cmd *Command, ast *AST) {
	if cmd.Every == 0 {
		cmd.Every = 1
	}
	cmd.BeatOffset = p.beatCount % cmd.Every
}

// parseArgs: arg (',' arg)*
func (p *parser) parseArgs() ([]Arg, error) {
	var args []Arg
	arg, err := p.parseArg()
	if err != nil {
		return nil, err
	}
	args = append(args, arg)

	for p.lex.token.Type == TokComma {
		p.lex.next()
		arg, err = p.parseArg()
		if err != nil {
			return nil, err
		}
		args = append(args, arg)
	}
	return args, nil
}

// parseArg: string | number
func (p *parser) parseArg() (Arg, error) {
	switch p.lex.token.Type {
	case TokString:
		arg := Arg{Type: ArgString, Value: p.lex.token.Value}
		p.lex.next()
		return arg, nil
	case TokNumber, TokMinus:
		arg := Arg{Type: ArgNumber, Value: p.lex.token.Value}
		p.lex.next()
		return arg, nil
	default:
		return Arg{}, fmt.Errorf("expected string or number argument, got %s at position %d", tokenName(p.lex.token.Type), p.lex.token.Pos)
	}
}

// parseIntArg parses a numeric argument as an integer.
func (p *parser) parseIntArg() (int, error) {
	if p.lex.token.Type == TokMinus {
		p.lex.next()
		return 0, nil
	}
	if p.lex.token.Type != TokNumber {
		return 0, fmt.Errorf("expected number for modifier argument, got %s", tokenName(p.lex.token.Type))
	}
	val := p.lex.token.Value
	p.lex.next()
	var n int
	fmt.Sscanf(val, "%d", &n)
	return n, nil
}

// parseStringArg parses a string argument.
func (p *parser) parseStringArg() (string, error) {
	if p.lex.token.Type != TokString {
		return "", fmt.Errorf("expected string argument, got %s", tokenName(p.lex.token.Type))
	}
	val := p.lex.token.Value
	p.lex.next()
	return val, nil
}