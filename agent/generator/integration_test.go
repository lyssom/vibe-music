package generator_test

import (
	"context"
	"testing"

	"github.com/lyssom/vibe-music/agent/generator"
	"github.com/lyssom/vibe-music/core/pattern"
)

func TestIntegration_AgentToParser(t *testing.T) {
	client := &mockLLMClient{response: `sound("bd sd hh")`}
	gen := generator.NewLLMGenerator(client)

	code, err := gen.Generate(context.Background(), "make a beat", generator.PromptContext{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// The generated code should be parseable
	ast, err := pattern.Parse(code)
	if err != nil {
		t.Fatalf("generated code should be valid pattern: %v", err)
	}
	if len(ast.Commands) != 1 {
		t.Errorf("expected 1 command, got %d", len(ast.Commands))
	}
	if ast.Commands[0].Name != "sound" {
		t.Errorf("expected 'sound' command, got %q", ast.Commands[0].Name)
	}
}

func TestIntegration_AgentWithErrorContext(t *testing.T) {
	client := &mockLLMClient{response: `sound("bd sd")`}
	gen := generator.NewLLMGenerator(client)

	code, err := gen.Generate(context.Background(), "fix it", generator.PromptContext{
		CurrentCode:  `sound("invalid_drum")`,
		ErrorMessage: "unknown drum: invalid_drum",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = pattern.Parse(code)
	if err != nil {
		t.Fatalf("generated code should be valid: %v", err)
	}
}

func TestIntegration_GeneratorSatisfiesInterface(t *testing.T) {
	client := &mockLLMClient{response: `sound("bd")`}
	gen := generator.NewLLMGenerator(client)

	var _ generator.Generator = gen
}

func TestIntegration_SystemPromptProducesValidDrumCode(t *testing.T) {
	client := &mockLLMClient{
		response: `sound("bd sd hh oh lt ht")`,
	}
	gen := generator.NewLLMGenerator(client)

	code, err := gen.Generate(context.Background(), "all drums", generator.PromptContext{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ast, err := pattern.Parse(code)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	if len(ast.Commands) > 0 && ast.Commands[0].Name == "sound" {
		drumStr := ast.Commands[0].Args[0].Value
		if drumStr == "" {
			t.Error("expected non-empty drum string")
		}
	}
}

func TestIntegration_StreamingConversation(t *testing.T) {
	client := &mockLLMClient{response: `sound("bd")`}
	gen := generator.NewLLMGenerator(client)

	// First question
	gen.Generate(context.Background(), "make a kick", generator.PromptContext{})

	// Change mock response for follow-up
	client.response = `sound("bd sd")`

	// Follow-up should include history and produce valid code
	code, err := gen.Generate(context.Background(), "add snare", generator.PromptContext{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = pattern.Parse(code)
	if err != nil {
		t.Fatalf("follow-up code should be valid: %v", err)
	}
}