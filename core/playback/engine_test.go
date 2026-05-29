package playback_test

import (
	"sync"
	"testing"
	"time"

	"github.com/lyssom/vibe-music/core/audio"
	"github.com/lyssom/vibe-music/core/pattern"
	"github.com/lyssom/vibe-music/core/playback"
)

var (
	testAudioEng audio.Engine
	testAudioErr error
	testEngOnce  sync.Once
)

func getTestAudioEngine() (audio.Engine, error) {
	testEngOnce.Do(func() {
		testAudioEng, testAudioErr = audio.NewOtoEngine()
	})
	return testAudioEng, testAudioErr
}

func TestEngineStartStop(t *testing.T) {
	e, err := getTestAudioEngine()
	if err != nil {
		t.Skipf("audio device not available: %v", err)
	}

	ast, err := pattern.Parse(`sound("bd")`)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	pb := playback.New(e)
	pb.LoadAST(ast, 120)

	if pb.IsPlaying() {
		t.Error("engine should not be playing before Start()")
	}
	if pb.BPM() != 120 {
		t.Errorf("expected BPM 120, got %d", pb.BPM())
	}
	if pb.Beat() != 0 {
		t.Errorf("expected beat 0, got %d", pb.Beat())
	}

	pb.Start()
	if !pb.IsPlaying() {
		t.Error("engine should be playing after Start()")
	}

	// Wait for at least one beat (500ms at 120 BPM)
	time.Sleep(600 * time.Millisecond)
	if pb.Beat() < 1 {
		t.Errorf("expected at least 1 beat, got %d", pb.Beat())
	}

	pb.Stop()
	if pb.IsPlaying() {
		t.Error("engine should not be playing after Stop()")
	}
}

func TestEngineHotSwapAST(t *testing.T) {
	e, err := getTestAudioEngine()
	if err != nil {
		t.Skipf("audio device not available: %v", err)
	}

	ast1, _ := pattern.Parse(`sound("bd")`)
	ast2, _ := pattern.Parse(`sound("hh")`)

	pb := playback.New(e)

	pb.LoadAST(ast1, 120)
	pb.Start()
	time.Sleep(100 * time.Millisecond)

	// Hot swap while playing
	pb.LoadAST(ast2, 140)
	if pb.BPM() != 140 {
		t.Errorf("expected BPM 140 after hot swap, got %d", pb.BPM())
	}

	pb.Stop()
}

func TestEngineDoubleStartIsIdempotent(t *testing.T) {
	e, err := getTestAudioEngine()
	if err != nil {
		t.Skipf("audio device not available: %v", err)
	}

	ast, _ := pattern.Parse(`sound("bd")`)
	pb := playback.New(e)
	pb.LoadAST(ast, 300)

	pb.Start()
	pb.Start() // second Start should be no-op
	if !pb.IsPlaying() {
		t.Error("should still be playing")
	}

	pb.Stop()
	pb.Stop() // second Stop should be no-op
	if pb.IsPlaying() {
		t.Error("should not be playing")
	}
}