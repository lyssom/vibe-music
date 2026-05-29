package synth

import "math"

const sampleRate = 44100

// Voice represents a single synthesizer voice.
type Voice interface {
	Trigger(freq float64, velocity float64)
	Release()
	Process() float64
	IsActive() bool
}

// IsActive returns true if the voice is still producing sound.
func IsActive(v Voice) bool {
	return v != nil && v.IsActive()
}

// Engine manages multiple synth voices.
type Engine interface {
	NewVoice() Voice
	Process() float64
}

// ADSR envelope parameters.
type ADSR struct {
	Attack  float64 // seconds
	Decay   float64 // seconds
	Sustain float64 // level [0, 1]
	Release float64 // seconds
}

// DefaultPiano returns a warm piano-like ADSR.
func DefaultPiano() ADSR {
	return ADSR{Attack: 0.005, Decay: 0.25, Sustain: 0.5, Release: 0.6}
}

// DefaultBass returns a bass ADSR.
func DefaultBass() ADSR {
	return ADSR{Attack: 0.012, Decay: 0.12, Sustain: 0.75, Release: 0.18}
}

// SimpleVoice is a soft synth voice with one-pole lowpass filter.
type SimpleVoice struct {
	Freq       float64
	Velocity   float64
	Wave       int // 0=sine, 1=triangle, 2=sawtooth, 3=square
	ADSR       ADSR
	Phase      float64
	SampleNum  int
	Released   int
	IsReleased bool
	Active     bool

	// One-pole lowpass filter state (smooths harsh harmonics)
	lpState float64
}

// NewSimpleVoice creates a new voice.
func NewSimpleVoice() *SimpleVoice {
	return &SimpleVoice{
		Wave:   1, // triangle (softer default)
		ADSR:   DefaultPiano(),
		Active: false,
	}
}

// Trigger starts the voice.
func (v *SimpleVoice) Trigger(freq float64, velocity float64) {
	v.Freq = freq
	v.Velocity = velocity
	v.Phase = 0
	v.SampleNum = 0
	v.Released = -1
	v.IsReleased = false
	v.Active = true
	v.lpState = 0
}

// Release begins the release phase.
func (v *SimpleVoice) Release() {
	if v.Active && !v.IsReleased {
		v.Released = v.SampleNum
		v.IsReleased = true
	}
}

// Process generates the next sample with a 1-pole lowpass filter for warmth.
func (v *SimpleVoice) Process() float64 {
	if !v.Active {
		return 0
	}
	v.SampleNum++

	// Envelope
	peak := 0.35 + 0.55*v.Velocity
	dt := float64(v.SampleNum) / sampleRate

	var env float64
	if v.IsReleased {
		releaseStart := float64(v.Released) / sampleRate
		releaseTime := dt - releaseStart
		if v.ADSR.Release > 0 {
			env = peak * v.ADSR.Sustain * maxf(0, 1.0-releaseTime/v.ADSR.Release)
		} else {
			env = 0
		}
	} else {
		if dt < v.ADSR.Attack {
			env = (dt / v.ADSR.Attack) * peak
		} else if dt < v.ADSR.Attack+v.ADSR.Decay {
			dp := (dt - v.ADSR.Attack) / v.ADSR.Decay
			env = peak - (peak-v.ADSR.Sustain*peak)*dp
		} else {
			env = v.ADSR.Sustain * peak
		}
	}

	if env <= 0.00005 {
		v.Active = false
		return 0
	}

	// Basic waveform generation
	phase := v.Phase
	var raw float64

	switch v.Wave {
	case 0: // sine
		raw = math.Sin(2.0 * math.Pi * phase)
	case 1: // triangle (warm, few harmonics)
		t := phase - math.Floor(phase+0.5)
		raw = 4.0*math.Abs(t) - 1.0
	case 2: // sawtooth
		raw = 2.0*phase - 1.0
	case 3: // square
		if phase < 0.5 {
			raw = 1.0
		} else {
			raw = -1.0
		}
	}

	// One-pole lowpass filter: cutoff varies by frequency
	// fc ≈ freq * 2 (very rough), then pole = exp(-2π * fc/fs)
	// For warmth: use fixed pole (~3000Hz cutoff at 44100Hz)
	pole := 0.85 // lower = smoother, 0.0 = no filter, 0.9 = very smooth
	v.lpState = v.lpState + pole*(raw - v.lpState)
	sample := v.lpState

	// Advance phase
	v.Phase += v.Freq / sampleRate
	if v.Phase >= 1.0 {
		v.Phase -= 1.0
	}

	return sample * env
}

// IsActive returns true if the voice is still producing sound.
func (v *SimpleVoice) IsActive() bool {
	return v.Active
}

// SimpleEngine manages polyphonic voices.
type SimpleEngine struct {
	Voices []*SimpleVoice
}

// NewSimpleEngine creates a synth engine.
func NewSimpleEngine(n int) *SimpleEngine {
	e := &SimpleEngine{Voices: make([]*SimpleVoice, n)}
	for i := range e.Voices {
		e.Voices[i] = NewSimpleVoice()
	}
	return e
}

// NewVoice implements Engine.NewVoice (reuses from pool).
func (e *SimpleEngine) NewVoice() Voice {
	for _, v := range e.Voices {
		if !v.Active {
			return v
		}
	}
	// Steal the quietest (oldest with lowest velocity) voice
	quietest := e.Voices[0]
	for _, v := range e.Voices {
		if v.SampleNum < quietest.SampleNum {
			quietest = v
		}
	}
	quietest.Active = false
	return quietest
}

// Process mixes all active voices with gentle limiting.
func (e *SimpleEngine) Process() float64 {
	var out float64
	for _, v := range e.Voices {
		out += v.Process()
	}
	// Gentle soft clip (tanh-like)
	if out > 0.9 {
		out = 0.9 + (1.0-0.9)/((out-0.9)*3.0+1.0)
	} else if out < -0.9 {
		out = -0.9 - (1.0-0.9)/((-out-0.9)*3.0+1.0)
	}
	return out * 0.75
}

// TriggerVoice triggers a voice with specific wave/ADSR.
func (e *SimpleEngine) TriggerVoice(freq float64, velocity float64, wave int, adsr ADSR) {
	v := e.NewVoice().(*SimpleVoice)
	v.Wave = wave
	v.ADSR = adsr
	v.Trigger(freq, velocity)
}

// WaveType constants (for external use)
const (
	WaveSine WaveType = iota
	WaveTriangle
	WaveSawtooth
	WaveSquare
)

// WaveType identifies the waveform shape.
type WaveType int

// NewSimpleEnginePoly is an alias.
func NewSimpleEnginePoly(n int) *SimpleEngine {
	return NewSimpleEngine(n)
}

func maxf(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}