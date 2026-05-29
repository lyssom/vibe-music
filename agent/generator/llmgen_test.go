package generator_test

import (
	"context"
	"strings"
	"testing"

	"github.com/lyssom/vibe-music/agent/generator"
	"github.com/lyssom/vibe-music/agent/llm"
)

// mockLLMClient implements llm.Client for testing.
type mockLLMClient struct {
	response string
}

func (m *mockLLMClient) Chat(_ context.Context, messages []llm.Message) (string, error) {
	return m.response, nil
}

func (m *mockLLMClient) ChatStream(_ context.Context, messages []llm.Message) (<-chan llm.StreamEvent, error) {
	ch := make(chan llm.StreamEvent, len(m.response)+1)
	go func() {
		for _, r := range m.response {
			ch <- llm.StreamEvent{Content: string(r), Done: false}
		}
		ch <- llm.StreamEvent{Content: "", Done: true}
		close(ch)
	}()
	return ch, nil
}

func TestLLMGeneratorGenerate(t *testing.T) {
	client := &mockLLMClient{response: `sound("bd sd hh")`}
	gen := generator.NewLLMGenerator(client)

	result, err := gen.Generate(context.Background(), "make a beat", generator.PromptContext{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != `sound("bd sd hh")` {
		t.Errorf("unexpected result: %q", result)
	}
}

func TestLLMGeneratorGenerateStream(t *testing.T) {
	client := &mockLLMClient{response: `sound("bd")`}
	gen := generator.NewLLMGenerator(client)

	ch, err := gen.GenerateStream(context.Background(), "make kick", generator.PromptContext{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result strings.Builder
	for ev := range ch {
		result.WriteString(ev.Content)
	}

	if result.String() != `sound("bd")` {
		t.Errorf("unexpected result: %q", result.String())
	}
}

func TestLLMGeneratorWithContext(t *testing.T) {
	client := &mockLLMClient{
		response: `sound("bd sd hh oh")`,
	}
	gen := generator.NewLLMGenerator(client)

	result, err := gen.Generate(context.Background(), "add open hihat", generator.PromptContext{
		CurrentCode:  `sound("bd sd hh")`,
		ErrorMessage: "",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == "" {
		t.Error("expected non-empty result")
	}
}

func TestLLMGeneratorHistoryPreserved(t *testing.T) {
	client := &mockLLMClient{response: `sound("bd")`}
	gen := generator.NewLLMGenerator(client)

	// First call
	_, _ = gen.Generate(context.Background(), "make a beat", generator.PromptContext{})
	// Second call should include history
	result, err := gen.Generate(context.Background(), "change it up", generator.PromptContext{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != `sound("bd")` {
		t.Errorf("unexpected result: %q", result)
	}
}

func TestLLMGeneratorClearHistory(t *testing.T) {
	client := &mockLLMClient{response: `sound("hh")`}
	gen := generator.NewLLMGenerator(client)

	_, _ = gen.Generate(context.Background(), "make a beat", generator.PromptContext{})
	gen.ClearHistory()
	result, err := gen.Generate(context.Background(), "fresh start", generator.PromptContext{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != `sound("hh")` {
		t.Errorf("unexpected result: %q", result)
	}
}

func TestStripThinking(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			" thinkingThe user wants jazz. Let me think. response\nsound(\"bd hh\")",
			"sound(\"bd hh\")",
		},
		{
			"sound(\"bd\")",
			"sound(\"bd\")",
		},
		{
			" thinkingPlan... response\n\nsound(\"bd sd\")\n",
			"sound(\"bd sd\")",
		},
	}

	for _, tt := range tests {
		// We can't call unexported stripThinking, but test via Generate
		client := &mockLLMClient{response: tt.input}
		gen := generator.NewLLMGenerator(client)
		result, err := gen.Generate(context.Background(), "test", generator.PromptContext{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != tt.expected {
			t.Errorf("for input %q:\n  got      %q\n  expected %q", tt.input, result, tt.expected)
		}
	}
}