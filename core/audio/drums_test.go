package audio_test

import (
	"math"
	"testing"

	"github.com/lyssom/vibe-music/core/audio"
)

func TestLookupDrum(t *testing.T) {
	tests := []struct {
		name   string
		expect audio.DrumType
		found  bool
	}{
		{"bd", audio.DrumKick, true},
		{"sd", audio.DrumSnare, true},
		{"hh", audio.DrumHihat, true},
		{"oh", audio.DrumOpenHat, true},
		{"lt", audio.DrumLowTom, true},
		{"ht", audio.DrumHighTom, true},
		{"xyz", 0, false},
	}

	for _, tt := range tests {
		dt, ok := audio.LookupDrum(tt.name)
		if ok != tt.found {
			t.Errorf("LookupDrum(%q) found=%v, want %v", tt.name, ok, tt.found)
		}
		if ok && dt != tt.expect {
			t.Errorf("LookupDrum(%q) type=%v, want %v", tt.name, dt, tt.expect)
		}
	}
}

func TestSynthesizeDrum_ProducesSamples(t *testing.T) {
	n := 22050 // 0.5 second
	for _, dt := range []audio.DrumType{
		audio.DrumKick, audio.DrumSnare, audio.DrumHihat,
		audio.DrumOpenHat, audio.DrumLowTom, audio.DrumHighTom,
	} {
		samples := audio.SynthesizeDrum(dt, 0.8, n)
		if len(samples) == 0 {
			t.Errorf("SynthesizeDrum(%d) returned 0 samples", dt)
		}

		// Verify samples are within [-1.0, 1.0]
		for i, s := range samples {
			if s < -1.0 || s > 1.0 {
				t.Errorf("SynthesizeDrum(%d) sample[%d]=%f out of range [-1,1]", dt, i, s)
				break // one failure is enough
			}
		}

		// Verify at least one non-zero sample (sound was generated)
		hasSound := false
		for _, s := range samples {
			if math.Abs(s) > 0.001 {
				hasSound = true
				break
			}
		}
		if !hasSound {
			t.Errorf("SynthesizeDrum(%d) produced only silence", dt)
		}
	}
}

func TestSynthesizeDrum_ZeroVelocityIsSilent(t *testing.T) {
	samples := audio.SynthesizeDrum(audio.DrumKick, 0.0, 1000)
	for _, s := range samples {
		if math.Abs(s) > 0.001 {
			t.Error("zero velocity should produce silence")
			break
		}
	}
}