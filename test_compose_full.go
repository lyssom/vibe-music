package main

import (
	"context"
	"fmt"
	
	"github.com/lyssom/vibe-music/agent/song"
)

func main() {
	fmt.Println("=== Full Composition Flow Test ===")
	
	// Create composer with nil generator (uses template fallback)
	composer := song.NewComposer(nil)
	
	// Step 1: Start session
	fmt.Println("\n1. Start session with '写一首爵士歌曲'")
	q := composer.StartSession("写一首爵士歌曲")
	fmt.Printf("   Q: %s\n", q.Text)
	fmt.Printf("   Mode: %d (ModeFull=%d)\n", composer.GetSessionState().Mode, song.ModeFull)
	
	// Helper to answer question
	answer := func(input string) {
		_, q, _ := composer.ProcessResponse(context.Background(), input)
		fmt.Printf("   A: %s\n", input)
		fmt.Printf("   Next Q: %s\n", q.Text)
	}
	
	// Steps 2-11: Answer all 10 questions
	fmt.Println("\n2-11. Answer all 10 questions")
	answer("爵士")
	answer("浪漫")
	answer("慢节奏")
	answer("钢琴三重奏")
	answer("小调")
	answer("功能和弦")
	answer("中速")
	answer("柔和-激昂")
	answer("由你推荐")  // structure selection
	answer("即兴")     // techniques - last element
	
	// Step 12: Check if exploration is complete
	state := composer.GetSessionState()
	fmt.Printf("\n   Exploration complete: %v\n", state.CurrentPhase == song.PhaseStructure)
	
	// Step 13: Select structure
	fmt.Println("\n12. Select structure 1")
	err := composer.ProcessStructureSelection(0)
	if err != nil {
		fmt.Printf("   ERROR: %v\n", err)
	} else {
		fmt.Println("   Structure selected successfully")
	}
	
	// Step 14: Generate all sections
	fmt.Println("\n13. Generate all sections")
	err = composer.GenerateAllSections(context.Background())
	if err != nil {
		fmt.Printf("   ERROR: %v\n", err)
	} else {
		fmt.Println("   Generation complete!")
	}
	
	// Step 15: Check result
	s := composer.GetSong()
	if s != nil {
		fmt.Printf("\n   Song: %s\n", s.Title)
		fmt.Printf("   Sections: %d\n", len(s.Sections))
		fmt.Printf("   Total bars: %d\n", s.TotalBars)
		fmt.Printf("   Duration: %s\n", composer.GetDurationEstimate())
		
		for _, sec := range s.Sections {
			fmt.Printf("\n   --- %s (%d bars) ---\n", sec.ID, sec.Bars)
			fmt.Println(sec.DSLCode)
		}
	} else {
		fmt.Println("\n   ERROR: Song is nil!")
	}
	
	fmt.Println("\n=== Test Complete ===")
}
