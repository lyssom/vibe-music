package audio_test

import (
	"testing"

	"github.com/lyssom/vibe-music/core/audio"
)

// mockEngine is a no-op implementation of audio.Engine for compile-time interface verification.
type mockEngine struct {
	bpm int
}

func (m *mockEngine) Play(_ <-chan float64) error  { return nil }
func (m *mockEngine) SetBPM(bpm int)                { m.bpm = bpm }
func (m *mockEngine) LoadSample(_, _ string) error   { return nil }
func (m *mockEngine) Close() error                   { return nil }

var _ audio.Engine = (*mockEngine)(nil)

func TestMockEngineSatisfiesInterface(t *testing.T) {
	var e audio.Engine = &mockEngine{}
	e.SetBPM(120)
	if e == nil {
		t.Fatal("engine should not be nil")
	}
}