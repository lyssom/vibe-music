// Package audio defines the audio playback and sample management interface.
// Implementations handle audio output, mixing, and BPM synchronization.
package audio

// Engine defines the audio playback and sample management interface.
// Implementations handle audio output, mixing, and BPM synchronization.
type Engine interface {
	// Play starts consuming sample frames from the channel and sending them to the audio output.
	// The channel provides float64 samples normalized to [-1.0, 1.0].
	Play(samples <-chan float64) error

	// SetBPM adjusts the beats-per-minute tempo.
	SetBPM(bpm int)

	// LoadSample loads an audio sample from disk and associates it with a name.
	LoadSample(name string, path string) error

	// Close releases audio resources.
	Close() error
}