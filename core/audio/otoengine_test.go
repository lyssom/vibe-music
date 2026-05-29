package audio_test

import (
	"sync"
	"testing"

	"github.com/lyssom/vibe-music/core/audio"
)

func TestOtoEngine(t *testing.T) {
	e, err := audio.NewOtoEngine()
	if err != nil {
		t.Skipf("audio device not available: %v", err)
	}
	defer e.Close()

	// Test SetBPM
	e.SetBPM(140)

	// Test LoadSample (no-op)
	err = e.LoadSample("bd", "/nonexistent/path")
	if err != nil {
		t.Errorf("LoadSample should be no-op, got error: %v", err)
	}

	// Test Play - kick drum for 0.1 seconds
	drumSamples := audio.SynthesizeDrum(audio.DrumKick, 0.8, 4410)
	ch := make(chan float64, 128)
	var wg sync.WaitGroup
	wg.Add(1)

	var playErr error
	go func() {
		defer wg.Done()
		playErr = e.Play(ch)
	}()

	for _, s := range drumSamples {
		ch <- s
	}
	close(ch)
	wg.Wait()

	if playErr != nil {
		t.Errorf("Play returned error: %v", playErr)
	}
}