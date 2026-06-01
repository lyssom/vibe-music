package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/lyssom/vibe-music/agent/generator"
	"github.com/lyssom/vibe-music/agent/song"
	"github.com/lyssom/vibe-music/agent/llm"
)

// Mock generator
type mockGen struct{}
func (m *mockGen) Generate(ctx context.Context, prompt string, pctx generator.PromptContext) (string, error) {
	return "// mock generated code", nil
}
func (m *mockGen) GenerateStream(ctx context.Context, prompt string, pctx generator.PromptContext) (<-chan llm.StreamEvent, error) {
	ch := make(chan llm.StreamEvent)
	close(ch)
	return ch, nil
}

func main() {
	fmt.Println("=== Testing TUI flow ===")
	
	// Simulate what TUI does
	gen := &mockGen{}
	composer := song.NewComposer(gen)
	
	// Step 1: User enters "写一首爵士歌曲"
	prompt := "写一首爵士歌曲"
	fmt.Printf("\n1. User input: %q\n", prompt)
	
	// Check if shouldStartComposeMode would return true
	songKeywords := []string{"歌曲", "写歌", "主歌", "副歌", "桥段", "bridge", "verse", "chorus", "完整", "完整歌曲", "jazz", "爵士", "pop", "流行", "rock", "摇滚"}
	promptLower := strings.ToLower(prompt)
	shouldCompose := false
	for _, kw := range songKeywords {
		if strings.Contains(promptLower, strings.ToLower(kw)) {
			shouldCompose = true
			fmt.Printf("   Matched keyword: %q\n", kw)
			break
		}
	}
	fmt.Printf("   shouldStartComposeMode: %v\n", shouldCompose)
	
	// Start session
	q1 := composer.StartSession(prompt)
	fmt.Printf("   First question: %s\n", q1.Text)
	fmt.Printf("   Question kind: %s\n", q1.Kind)
	
	state := composer.GetSessionState()
	fmt.Printf("   Mode: %d (ModeFull=%d)\n", state.Mode, song.ModeFull)
	
	// Step 2: User enters "爵士"
	fmt.Printf("\n2. User input: %q\n", "爵士")
	
	// This is what processComposeInput does
	_, q2, err := composer.ProcessResponse(context.Background(), "爵士")
	if err != nil {
		fmt.Printf("   ERROR: %v\n", err)
	} else {
		fmt.Printf("   Next question: %s\n", q2.Text)
		fmt.Printf("   Question kind: %s\n", q2.Kind)
	}
	
	// Check elements
	state = composer.GetSessionState()
	fmt.Printf("   Genre: %q, Emotion: %q\n", state.Elements.Genre, state.Elements.Emotion)
	
	fmt.Println("\n=== Test complete ===")
}
