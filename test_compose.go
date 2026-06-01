package main

import (
	"context"
	"fmt"

	"github.com/lyssom/vibe-music/agent/song"
)

func main() {
	// Create a mock LLM generator (for testing purposes)
	// In production, this would be a real LLM client

	composer := song.NewComposer(nil)

	fmt.Println("=== Testing Multi-Round Composition Flow ===\n")

	// Step 1: User starts composition
	fmt.Println("1. User: '写一首爵士歌曲'")
	q1 := composer.StartSession("写一首爵士歌曲")
	fmt.Printf("   AI question: %s\n", q1.Text)
	fmt.Printf("   Question kind: %s\n", q1.Kind)
	fmt.Printf("   Options: %v\n", q1.Options)

	// Check what mode was set
	state := composer.GetSessionState()
	fmt.Printf("   Mode: %d (ModeFull=%d)\n\n", state.Mode, song.ModeFull)

	// Step 2: User responds with genre preference
	fmt.Println("2. User: '爵士'")
	_, q2, _ := composer.ProcessResponse(context.Background(), "爵士")
	fmt.Printf("   AI question: %s\n", q2.Text)
	fmt.Printf("   Question kind: %s\n", q2.Kind)

	// Check if genre was set
	state = composer.GetSessionState()
	fmt.Printf("   Genre: %q, Emotion: %q\n\n", state.Elements.Genre, state.Elements.Emotion)

	// Step 3: User responds with emotion preference
	fmt.Println("3. User: '浪漫'")
	_, q3, _ := composer.ProcessResponse(context.Background(), "浪漫")
	fmt.Printf("   AI question: %s\n", q3.Text)
	fmt.Printf("   Question kind: %s\n", q3.Kind)

	state = composer.GetSessionState()
	fmt.Printf("   Genre: %q, Emotion: %q, Rhythm: %q\n\n", 
		state.Elements.Genre, state.Elements.Emotion, state.Elements.Rhythm)

	// Step 4: User responds with rhythm preference
	fmt.Println("4. User: '慢节奏'")
	_, q4, _ := composer.ProcessResponse(context.Background(), "慢节奏")
	fmt.Printf("   AI question: %s\n", q4.Text)
	fmt.Printf("   Question kind: %s\n\n", q4.Kind)

	// Continue a few more rounds...
	fmt.Println("5. User: '钢琴三重奏'")
	_, q5, _ := composer.ProcessResponse(context.Background(), "钢琴三重奏")
	fmt.Printf("   AI question: %s\n", q5.Text)
	fmt.Printf("   Question kind: %s\n\n", q5.Kind)

	fmt.Println("6. User: '小调'")
	_, q6, _ := composer.ProcessResponse(context.Background(), "小调")
	fmt.Printf("   AI question: %s\n", q6.Text)
	fmt.Printf("   Question kind: %s\n\n", q6.Kind)

	fmt.Println("7. User: '功能和弦'")
	_, q7, _ := composer.ProcessResponse(context.Background(), "功能和弦")
	fmt.Printf("   AI question: %s\n", q7.Text)
	fmt.Printf("   Question kind: %s\n\n", q7.Kind)

	fmt.Println("8. User: '中速'")
	_, q8, _ := composer.ProcessResponse(context.Background(), "中速")
	fmt.Printf("   AI question: %s\n", q8.Text)
	fmt.Printf("   Question kind: %s\n\n", q8.Kind)

	fmt.Println("9. User: '柔和-激昂'")
	_, q9, _ := composer.ProcessResponse(context.Background(), "柔和-激昂")
	fmt.Printf("   AI question: %s\n", q9.Text)
	fmt.Printf("   Question kind: %s\n\n", q9.Kind)

	fmt.Println("10. User: '由你推荐'")
	_, q10, _ := composer.ProcessResponse(context.Background(), "由你推荐")
	fmt.Printf("   AI question: %s\n", q10.Text)
	fmt.Printf("   Question kind: %s\n\n", q10.Kind)

	fmt.Println("=== Test Complete ===")
}
