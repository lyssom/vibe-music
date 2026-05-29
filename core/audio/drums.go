package audio

import (
	"math"
	"math/rand"
)

const (
	sampleRate   = 44100
	defaultBPM   = 120
	beatDuration = float64(sampleRate*60) / float64(defaultBPM)
)

// DrumType identifies a built-in drum sound.
type DrumType int

const (
	DrumKick DrumType = iota
	DrumSnare
	DrumHihat
	DrumOpenHat
	DrumLowTom
	DrumHighTom
)

// drumNames maps built-in drum names to their type.
var drumNames = map[string]DrumType{
	"bd": DrumKick,
	"sd": DrumSnare,
	"hh": DrumHihat,
	"oh": DrumOpenHat,
	"lt": DrumLowTom,
	"ht": DrumHighTom,
}

// LookupDrum returns the DrumType for a named drum, or false if not found.
func LookupDrum(name string) (DrumType, bool) {
	t, ok := drumNames[name]
	return t, ok
}

// SynthesizeDrum generates samples for a drum sound of the given duration.
func SynthesizeDrum(d DrumType, velocity float64, durSamples int) []float64 {
	switch d {
	case DrumKick:
		return synthKick(velocity, durSamples)
	case DrumSnare:
		return synthSnare(velocity, durSamples)
	case DrumHihat:
		return synthHihat(velocity, durSamples)
	case DrumOpenHat:
		return synthHihat(velocity, durSamples*4)
	case DrumLowTom:
		return synthTom(velocity, durSamples, 100)
	case DrumHighTom:
		return synthTom(velocity, durSamples, 200)
	default:
		return nil
	}
}

// synthKick generates a kick drum: low sine with frequency sweep.
func synthKick(velocity float64, n int) []float64 {
	out := make([]float64, n)
	env := velocity * 0.8
	for i := 0; i < n; i++ {
		t := float64(i) / sampleRate
		// Frequency sweep from 150Hz down to 30Hz
		freq := 150.0 - 120.0*(float64(i)/float64(n))
		amp := env * (1.0 - float64(i)/float64(n))
		out[i] = amp * math.Sin(2.0*math.Pi*freq*t)
	}
	return out
}

// synthSnare generates a snare: mix of tone and noise.
func synthSnare(velocity float64, n int) []float64 {
	out := make([]float64, n)
	for i := 0; i < n; i++ {
		t := float64(i) / sampleRate
		env := velocity * 0.6 * (1.0 - float64(i)/float64(n))
		tone := 0.4 * math.Sin(2.0*math.Pi*180.0*t)
		noise := 0.6 * (rand.Float64()*2.0 - 1.0)
		out[i] = env * (tone + noise)
	}
	return out
}

// synthHihat generates a hi-hat: filtered noise with fast decay.
func synthHihat(velocity float64, n int) []float64 {
	out := make([]float64, n)
	for i := 0; i < n; i++ {
		env := velocity * 0.4 * (1.0 - float64(i)/float64(n))
		out[i] = env * (rand.Float64()*2.0 - 1.0)
	}
	return out
}

// synthTom generates a tom: sine with fast decay at given frequency.
func synthTom(velocity float64, n int, freqHz float64) []float64 {
	out := make([]float64, n)
	for i := 0; i < n; i++ {
		t := float64(i) / sampleRate
		env := velocity * 0.5 * (1.0 - float64(i)/float64(n))
		out[i] = env * math.Sin(2.0*math.Pi*freqHz*t)
	}
	return out
}