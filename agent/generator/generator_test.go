package generator_test

import (
	"context"
	"testing"

	"github.com/lyssom/vibe-music/agent/generator"
	"github.com/lyssom/vibe-music/agent/llm"
)

type mockGenerator struct{}

func (m *mockGenerator) Generate(_ context.Context, _ string, _ generator.PromptContext) (string, error) {
	return `sound("bd sd hh")`, nil
}

func (m *mockGenerator) GenerateStream(_ context.Context, _ string, _ generator.PromptContext) (<-chan llm.StreamEvent, error) {
	ch := make(chan llm.StreamEvent, 1)
	go func() {
		ch <- llm.StreamEvent{Content: `sound("bd sd hh")`, Done: true}
		close(ch)
	}()
	return ch, nil
}

var _ generator.Generator = (*mockGenerator)(nil)

func TestGeneratorProducesCode(t *testing.T) {
	g := &mockGenerator{}
	code, err := g.Generate(context.Background(), "make a beat", generator.PromptContext{
		CurrentCode: "",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != `sound("bd sd hh")` {
		t.Errorf("unexpected code: %q", code)
	}
}

func TestPromptContextFields(t *testing.T) {
	pc := generator.PromptContext{
		CurrentCode:  `sound("bd")`,
		History:      []llm.Message{{Role: "user", Content: "add snare"}},
		ErrorMessage: "unknown sample: bd",
	}
	if pc.CurrentCode != `sound("bd")` {
		t.Error("CurrentCode mismatch")
	}
	if len(pc.History) != 1 {
		t.Error("History length mismatch")
	}
	if pc.ErrorMessage != "unknown sample: bd" {
		t.Error("ErrorMessage mismatch")
	}
}