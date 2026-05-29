package pattern

import (
	"strings"
	"testing"
)

func TestParse_BasicSound(t *testing.T) {
	ast, err := Parse(`sound("bd sd hh")`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(ast.Commands) != 1 {
		t.Fatalf("expected 1 command, got %d", len(ast.Commands))
	}

	cmd := ast.Commands[0]
	if cmd.Name != "sound" {
		t.Errorf("command name = %q, want 'sound'", cmd.Name)
	}
	if len(cmd.Args) != 1 {
		t.Fatalf("expected 1 arg, got %d", len(cmd.Args))
	}
	if cmd.Args[0].Type != ArgString || cmd.Args[0].Value != "bd sd hh" {
		t.Errorf("arg = {%v %q}, want {string 'bd sd hh'}", cmd.Args[0].Type, cmd.Args[0].Value)
	}
}

func TestParse_MethodChain(t *testing.T) {
	// New DSL: .every chains onto the previous command (same line)
	src := `sound("bd sd").every(2)`
	ast, err := Parse(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(ast.Commands) != 1 {
		t.Fatalf("expected 1 command, got %d", len(ast.Commands))
	}

	cmd := ast.Commands[0]
	if cmd.Name != "sound" {
		t.Errorf("cmd.Name = %q, want 'sound'", cmd.Name)
	}
	if cmd.Args[0].Value != "bd sd" {
		t.Errorf("cmd.Args[0] = %q, want 'bd sd'", cmd.Args[0].Value)
	}
	if cmd.Every != 2 {
		t.Errorf("cmd.Every = %d, want 2", cmd.Every)
	}
}

func TestParse_MultilineCommands(t *testing.T) {
	src := "sound(\"bd sd\")\nnote(\"c3 e3\")\nbass(\"c2\")"
	ast, err := Parse(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(ast.Commands) != 3 {
		t.Fatalf("expected 3 commands, got %d", len(ast.Commands))
	}

	names := []string{"sound", "note", "bass"}
	for i, name := range names {
		if ast.Commands[i].Name != name {
			t.Errorf("cmd[%d].Name = %q, want %q", i, ast.Commands[i].Name, name)
		}
	}
}

func TestParse_NoArgs(t *testing.T) {
	ast, err := Parse(`stop()`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ast.Commands) != 1 {
		t.Fatalf("expected 1 command, got %d", len(ast.Commands))
	}
	if ast.Commands[0].Name != "stop" {
		t.Errorf("name = %q, want 'stop'", ast.Commands[0].Name)
	}
	if len(ast.Commands[0].Args) != 0 {
		t.Errorf("expected 0 args, got %d", len(ast.Commands[0].Args))
	}
}

func TestParse_SyntaxErrors(t *testing.T) {
	tests := []struct {
		src  string
		desc string
	}{
		{`123`, "number alone"},
		{`sound(`, "unclosed paren"},
		{`sound("bd"`, "missing closing paren"},
		{`sound("bd") extra`, "trailing tokens"},
		{`.sound("bd")`, "dot before first command"},
	}

	for _, tt := range tests {
		_, err := Parse(tt.src)
		if err == nil {
			t.Errorf("expected error for %q (%s), got nil", tt.src, tt.desc)
		}
	}
}

func TestParse_ComplexPattern(t *testing.T) {
	// New DSL: separate commands per line or chained with .every/.slow
	src := "sound(\"bd sd hh\").every(4)\nnote(\"c3 e3 g3\").slow(2)\nbass(\"c2\")"
	ast, err := Parse(src)
	if err != nil {
		t.Fatalf("unexpected error parsing complex pattern: %v", err)
	}

	if len(ast.Commands) != 3 {
		t.Fatalf("expected 3 commands, got %d", len(ast.Commands))
	}

	if ast.Commands[0].Name != "sound" {
		t.Errorf("cmd[0].Name = %q, want 'sound'", ast.Commands[0].Name)
	}
	if ast.Commands[0].Every != 4 {
		t.Errorf("cmd[0].Every = %d, want 4", ast.Commands[0].Every)
	}

	if ast.Commands[1].Name != "note" {
		t.Errorf("cmd[1].Name = %q, want 'note'", ast.Commands[1].Name)
	}
	if ast.Commands[1].Every != 2 {
		t.Errorf("cmd[1].Every = %d, want 2 (slow halves beat divisor)", ast.Commands[1].Every)
	}

	if ast.Commands[2].Name != "bass" {
		t.Errorf("cmd[2].Name = %q, want 'bass'", ast.Commands[2].Name)
	}
}

func TestParse_WhitespaceResilience(t *testing.T) {
	ast, err := Parse(`  sound  (  "bd"  ,  120  )  `)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cmd := ast.Commands[0]
	if cmd.Name != "sound" {
		t.Errorf("name = %q", cmd.Name)
	}
	if len(cmd.Args) != 2 {
		t.Fatalf("expected 2 args, got %d", len(cmd.Args))
	}
	if cmd.Args[0].Value != "bd" {
		t.Errorf("arg[0] = %q", cmd.Args[0].Value)
	}
	if cmd.Args[1].Value != "120" {
		t.Errorf("arg[1] = %q", cmd.Args[1].Value)
	}
}

func TestParse_Multiline(t *testing.T) {
	src := "sound(\"bd sd hh\")\nsound(\"bd\")\n"
	ast, err := Parse(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ast.Commands) != 2 {
		t.Fatalf("expected 2 commands, got %d", len(ast.Commands))
	}
	if ast.Commands[0].Name != "sound" || ast.Commands[0].Args[0].Value != "bd sd hh" {
		t.Errorf("cmd[0] = %s(%q)", ast.Commands[0].Name, ast.Commands[0].Args[0].Value)
	}
	if ast.Commands[1].Name != "sound" || ast.Commands[1].Args[0].Value != "bd" {
		t.Errorf("cmd[1] = %s(%q)", ast.Commands[1].Name, ast.Commands[1].Args[0].Value)
	}
}

func TestLexerAndParser_BasicIntegration(t *testing.T) {
	for _, src := range []string{
		`sound("bd")`,
		`sound("bd sd hh").every(2)`,
		`note("c3 e3 g3")`,
		`bass("c2")`,
		`chord("c3 e3 g3", "4n", 0.5)`,
		`sound("bd").every(4)\nsound("sd")`,
	} {
		ast, err := Parse(src)
		if err != nil {
			t.Errorf("Parse(%q) failed: %v", src, err)
			continue
		}
		if ast == nil {
			t.Errorf("Parse(%q) returned nil AST without error", src)
		}
	}

	// Empty input should fail
	_, err := Parse("")
	if err == nil || !strings.Contains(err.Error(), "expected function name") {
		t.Errorf("empty input should fail with 'expected function name', got: %v", err)
	}
}

func TestParse_ChordWithDuration(t *testing.T) {
	src := `chord("c3 e3 g3", "4n", 0.5)`
	ast, err := Parse(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cmd := ast.Commands[0]
	if cmd.Name != "chord" {
		t.Errorf("cmd.Name = %q, want 'chord'", cmd.Name)
	}
	if len(cmd.Args) != 3 {
		t.Fatalf("expected 3 args, got %d", len(cmd.Args))
	}
	if cmd.Args[0].Value != "c3 e3 g3" {
		t.Errorf("cmd.Args[0] = %q, want 'c3 e3 g3'", cmd.Args[0].Value)
	}
	if cmd.Args[1].Value != "4n" {
		t.Errorf("cmd.Args[1] = %q, want '4n'", cmd.Args[1].Value)
	}
	if cmd.Args[2].Value != "0.5" {
		t.Errorf("cmd.Args[2] = %q, want '0.5'", cmd.Args[2].Value)
	}
}

func TestParse_Iter(t *testing.T) {
	src := `sound("bd sd").every(4).iter(2, "sound(\"hh\")")`
	ast, err := Parse(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(ast.SubCommands) == 0 {
		t.Fatalf("expected SubCommands from iter(), got none")
	}
}

func TestNoteToFreq(t *testing.T) {
	tests := []struct {
		note string
		freq float64
		tol  float64
	}{
		{"c4", 261.63, 0.5},
		{"a4", 440.0, 0.5},
		{"eb4", 311.13, 0.5},
		{"b3", 246.94, 0.5},
		{"c0", 16.35, 0.5},
	}

	for _, tt := range tests {
		f := NoteToFreq(tt.note)
		diff := f - tt.freq
		if diff < 0 {
			diff = -diff
		}
		if diff > tt.tol {
			t.Errorf("NoteToFreq(%q) = %f, want ~%f", tt.note, f, tt.freq)
		}
	}
}