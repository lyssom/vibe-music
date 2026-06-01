package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/lyssom/vibe-music/agent/song"
	"github.com/lyssom/vibe-music/core/audio"
	"github.com/lyssom/vibe-music/core/pattern"
	"github.com/lyssom/vibe-music/core/playback"
)

func main() {
	fmt.Println("=== Song Save & Load Test ===")
	
	// Generate song
	composer := song.NewComposer(nil)
	composer.StartSession("写一首爵士歌曲")
	
	answers := []string{"爵士", "快乐", "快", "钢琴和鼓", "无", "无", "120", "无", "主歌副歌主歌副歌桥副歌", "无"}
	for _, ans := range answers {
		_, _, _ = composer.ProcessResponse(context.Background(), ans)
	}
	
	composer.ProcessStructureSelection(-1)
	err := composer.GenerateAllSections(context.Background())
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	
	s := composer.GetSong()
	if s == nil {
		fmt.Println("Song is nil!")
		return
	}
	
	elements := composer.GetSessionState().Elements
	
	fmt.Printf("Generated: %d sections, %d bars, BPM: %d\n", len(s.Sections), s.TotalBars, elements.BPM)
	
	// Save to file - WITHOUT comments since parser doesn't support them
	var dsl strings.Builder
	for i, sec := range s.Sections {
		if i > 0 {
			dsl.WriteString("\n")
		}
		dsl.WriteString(sec.DSLCode)
	}
	
	filename := "test_song.dsl"
	err = os.WriteFile(filename, []byte(dsl.String()), 0644)
	if err != nil {
		fmt.Printf("Save error: %v\n", err)
		return
	}
	fmt.Printf("Saved to: %s\n", filename)
	
	// Load and verify
	data, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("Load error: %v\n", err)
		return
	}
	fmt.Printf("Loaded %d bytes\n", len(data))
	
	// Parse
	ast, err := pattern.Parse(string(data))
	if err != nil {
		fmt.Printf("Parse error: %v\n", err)
		return
	}
	fmt.Printf("Parsed %d commands\n", len(ast.Commands))
	
	// Play
	audioEng, err := audio.NewWinMMEngine()
	if err != nil {
		fmt.Printf("WinMM init error: %v\n", err)
		return
	}
	defer audioEng.Close()
	
	play := playback.New(audioEng)
	play.LoadAST(ast, elements.BPM)
	play.Start()
	
	fmt.Println("Playing for 5 seconds...")
	for i := 0; i < 5; i++ {
		fmt.Printf("\rPlaying... %d/5", i+1)
		os.Stdout.Sync()
	}
	fmt.Println("\nDone!")
	
	play.Stop()
	os.Remove(filename)
}
