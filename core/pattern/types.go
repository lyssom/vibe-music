package pattern

import (
	"math"
	"time"
)

// Note represents a single musical event with pitch information and timing.
type Note struct {
	// Pitch is the MIDI note number or frequency identifier.
	Pitch float64
	// Velocity is the note intensity [0.0, 1.0].
	Velocity float64
	// Duration is how long the note sustains.
	Duration time.Duration
	// Sample optionally references a named drum sample.
	Sample string
	// Freq is the frequency in Hz for pitched instruments.
	Freq float64
	// IsRest is true if this note is a rest (silence).
	IsRest bool
}

// NoteDuration converts a string like "4n" "8n" "2n" to time.Duration at the given BPM.
func NoteDuration(notation string, bpm int) time.Duration {
	beat := time.Minute / time.Duration(bpm)
	switch notation {
	case "32n":
		return beat / 8
	case "16n":
		return beat / 4
	case "8n":
		return beat / 2
	case "4n":
		return beat
	case "2n":
		return beat * 2
	case "1n":
		return beat * 4
	default:
		return beat
	}
}

// NoteToFreq converts a note name like "c3" "eb4" "#f5" to frequency in Hz.
func NoteToFreq(name string) float64 {
	// MIDI note number mapping
	type noteVal struct{ name string; midi int }
	notes := []noteVal{
		{"c0", 12}, {"#c0", 13}, {"db0", 13}, {"d0", 14}, {"#d0", 15}, {"eb0", 15},
		{"e0", 16}, {"f0", 17}, {"#f0", 18}, {"gb0", 18}, {"g0", 19}, {"#g0", 20},
		{"ab0", 20}, {"a0", 21}, {"#a0", 22}, {"bb0", 22}, {"b0", 23},
		{"c1", 24}, {"#c1", 25}, {"db1", 25}, {"d1", 26}, {"#d1", 27}, {"eb1", 27},
		{"e1", 28}, {"f1", 29}, {"#f1", 30}, {"gb1", 30}, {"g1", 31}, {"#g1", 32},
		{"ab1", 32}, {"a1", 33}, {"#a1", 34}, {"bb1", 34}, {"b1", 35},
		{"c2", 36}, {"#c2", 37}, {"db2", 37}, {"d2", 38}, {"#d2", 39}, {"eb2", 39},
		{"e2", 40}, {"f2", 41}, {"#f2", 42}, {"gb2", 42}, {"g2", 43}, {"#g2", 44},
		{"ab2", 44}, {"a2", 45}, {"#a2", 46}, {"bb2", 46}, {"b2", 47},
		{"c3", 48}, {"#c3", 49}, {"db3", 49}, {"d3", 50}, {"#d3", 51}, {"eb3", 51},
		{"e3", 52}, {"f3", 53}, {"#f3", 54}, {"gb3", 54}, {"g3", 55}, {"#g3", 56},
		{"ab3", 56}, {"a3", 57}, {"#a3", 58}, {"bb3", 58}, {"b3", 59},
		{"c4", 60}, {"#c4", 61}, {"db4", 61}, {"d4", 62}, {"#d4", 63}, {"eb4", 63},
		{"e4", 64}, {"f4", 65}, {"#f4", 66}, {"gb4", 66}, {"g4", 67}, {"#g4", 68},
		{"ab4", 68}, {"a4", 69}, {"#a4", 70}, {"bb4", 70}, {"b4", 71},
		{"c5", 72}, {"#c5", 73}, {"db5", 73}, {"d5", 74}, {"#d5", 75}, {"eb5", 75},
		{"e5", 76}, {"f5", 77}, {"#f5", 78}, {"gb5", 78}, {"g5", 79}, {"#g5", 80},
		{"ab5", 80}, {"a5", 81}, {"#a5", 82}, {"bb5", 82}, {"b5", 83},
		{"c6", 84}, {"#c6", 85}, {"db6", 85}, {"d6", 86}, {"#d6", 87}, {"eb6", 87},
		{"e6", 88}, {"f6", 89}, {"#f6", 90}, {"gb6", 90}, {"g6", 91}, {"#g6", 92},
		{"ab6", 92}, {"a6", 93}, {"#a6", 94}, {"bb6", 94}, {"b6", 95},
		{"c7", 96}, {"#c7", 97}, {"db7", 97}, {"d7", 98}, {"#d7", 99}, {"eb7", 99},
		{"e7", 100}, {"f7", 101}, {"#f7", 102}, {"gb7", 102}, {"g7", 103}, {"#g7", 104},
		{"ab7", 104}, {"a7", 105}, {"#a7", 106}, {"bb7", 106}, {"b7", 107},
		{"c8", 108},
	}

	// Normalize: lower case and handle sharps/flats
	norm := name
	for i, c := range norm {
		if c >= 'A' && c <= 'Z' {
			norm = norm[:i] + string(c+('a'-'A')) + norm[i+1:]
		}
	}

	// Handle flats
	flatToSharp := map[string]string{
		"db": "#c", "eb": "#d", "gb": "#f", "ab": "#g", "bb": "#a",
	}
	for flat, sharp := range flatToSharp {
		if len(norm) >= 2 && norm[:2] == flat {
			norm = sharp + norm[2:]
		}
	}

	for _, n := range notes {
		if n.name == norm {
			return midiToFreq(float64(n.midi))
		}
	}
	return 440.0 // default to A4
}

// FreqToMIDI converts frequency in Hz to MIDI note number (rounded).
func FreqToMIDI(freq float64) int32 {
	midi := 69.0 + 12.0*math.Log(freq/440.0)/math.Log(2)
	return int32(math.Round(midi))
}

// midiToFreq converts a MIDI note number to frequency in Hz.
func midiToFreq(midi float64) float64 {
	return 440.0 * pow2((midi-69.0)/12.0)
}

func pow2(x float64) float64 {
	return math.Pow(2, x)
}

// Pattern is a time-ordered sequence of musical events.
type Pattern struct {
	Notes []Note
	Cycle time.Duration
}