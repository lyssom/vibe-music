package pattern_test

import (
	"testing"
	"time"

	"github.com/lyssom/vibe-music/core/pattern"
)

func TestSchedulerSoundCommand(t *testing.T) {
	notes := make(chan pattern.Note, 16)
	s := pattern.NewScheduler(120, notes)
	defer s.Stop()

	ast, err := pattern.Parse(`sound("bd sd")`)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	go s.Run(ast)

	// Wait for a beat to fire
	select {
	case n := <-notes:
		if n.Sample != "bd" && n.Sample != "sd" {
			t.Errorf("unexpected note: %+v", n)
		}
		// Should receive a second note
		select {
		case n2 := <-notes:
			if n2.Sample == n.Sample {
				t.Errorf("expected different drum, got same: %s", n2.Sample)
			}
		case <-time.After(2 * time.Second):
			t.Fatal("timed out waiting for second note")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for first note")
	}
}

func TestSchedulerBPMChange(t *testing.T) {
	notes := make(chan pattern.Note, 16)
	s := pattern.NewScheduler(300, notes)
	defer s.Stop()

	ast, err := pattern.Parse(`sound("bd")`)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	go s.Run(ast)

	// Should receive a note quickly at 300 BPM (200ms per beat)
	select {
	case <-notes:
		// success
	case <-time.After(1 * time.Second):
		t.Fatal("timed out at 300 BPM")
	}

	s.SetBPM(60)
}

func TestSchedulerStop(t *testing.T) {
	notes := make(chan pattern.Note, 16)
	s := pattern.NewScheduler(600, notes)

	ast, _ := pattern.Parse(`sound("bd")`)
	go s.Run(ast)

	// Wait for at least one note
	select {
	case <-notes:
	case <-time.After(1 * time.Second):
	}

	s.Stop()

	// After stop, no more notes should arrive
	time.Sleep(200 * time.Millisecond)
	select {
	case <-notes:
		t.Error("received note after Stop()")
	default:
		// expected - channel is empty
	}
}

func TestSchedulerIgnoreUnknownCommands(t *testing.T) {
	notes := make(chan pattern.Note, 16)
	s := pattern.NewScheduler(600, notes)
	defer s.Stop()

	ast, err := pattern.Parse(`unknown("xyz")
sound("hh")
other("x")`)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	go s.Run(ast)

	// Should receive hihat note, but no crash from unknown commands
	select {
	case n := <-notes:
		if n.Sample != "hh" {
			t.Errorf("expected 'hh', got %q", n.Sample)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timed out waiting for hihat note")
	}
}