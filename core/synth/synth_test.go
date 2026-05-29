package synth_test

import (
	"testing"

	"github.com/lyssom/vibe-music/core/synth"
)

type mockVoice struct{ active bool }

func (m *mockVoice) Trigger(_, _ float64) {}
func (m *mockVoice) Release()             {}
func (m *mockVoice) Process() float64     { return 0.0 }
func (m *mockVoice) IsActive() bool       { return m.active }

var _ synth.Voice = (*mockVoice)(nil)

type mockEngine struct{}

func (m *mockEngine) NewVoice() synth.Voice { return &mockVoice{active: true} }
func (m *mockEngine) Process() float64      { return 0.0 }

var _ synth.Engine = (*mockEngine)(nil)

func TestVoiceInterface(t *testing.T) {
	v := &mockVoice{active: true}
	if !v.IsActive() {
		t.Error("voice should be active")
	}
	v.Release()
}

func TestEngineCreatesVoice(t *testing.T) {
	e := &mockEngine{}
	voice := e.NewVoice()
	if voice == nil {
		t.Fatal("expected non-nil voice")
	}
	if !voice.IsActive() {
		t.Error("new voice should be active")
	}
}