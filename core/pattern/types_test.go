package pattern_test

import (
	"testing"
	"time"

	"github.com/lyssom/vibe-music/core/pattern"
)

func TestNoteCreation(t *testing.T) {
	n := pattern.Note{
		Pitch:    60.0, // middle C
		Velocity: 0.8,
		Duration: 100 * time.Millisecond,
		Sample:   "bd",
	}

	if n.Pitch != 60.0 {
		t.Errorf("expected pitch 60.0, got %f", n.Pitch)
	}
	if n.Velocity != 0.8 {
		t.Errorf("expected velocity 0.8, got %f", n.Velocity)
	}
	if n.Duration != 100*time.Millisecond {
		t.Errorf("expected duration 100ms, got %v", n.Duration)
	}
	if n.Sample != "bd" {
		t.Errorf("expected sample 'bd', got %q", n.Sample)
	}
}

func TestPatternCreation(t *testing.T) {
	p := pattern.Pattern{
		Notes: []pattern.Note{
			{Pitch: 60, Velocity: 1.0, Duration: 100 * time.Millisecond},
		},
		Cycle: 2 * time.Second,
	}

	if len(p.Notes) != 1 {
		t.Errorf("expected 1 note, got %d", len(p.Notes))
	}
}

// mockExecutor verifies the interface at compile time.
type mockExecutor struct{}

func (m *mockExecutor) Parse(_ string) (*pattern.Pattern, error)            { return nil, nil }
func (m *mockExecutor) Evaluate(_ *pattern.Pattern, _ time.Time) []pattern.Note { return nil }

var _ pattern.Executor = (*mockExecutor)(nil)