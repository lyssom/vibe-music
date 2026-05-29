package audio

import (
	"fmt"
	"io"
	"os"
	"sync"
	"sync/atomic"
	"unsafe"

	"github.com/ebitengine/oto/v3"
)

// OtoEngine implements Engine using oto for cross-platform audio output.
type OtoEngine struct {
	ctx     *oto.Context
	bpm     int32
	closeCh chan struct{}
}

// NewOtoEngine creates an audio engine with oto output.
func NewOtoEngine() (*OtoEngine, error) {
	op := &oto.NewContextOptions{
		SampleRate:   sampleRate,
		ChannelCount: 1,
		Format:       oto.FormatFloat32LE,
		BufferSize:   0, // use oto default
	}

	ctx, ready, err := oto.NewContext(op)
	if err != nil {
		return nil, err
	}
	<-ready

	// Log any driver initialization errors (e.g., if fallback to null context occurred).
	if driverErr := ctx.Err(); driverErr != nil {
		fmt.Fprintf(os.Stderr, "oto: audio driver warning: %v\n", driverErr)
	}

	return &OtoEngine{
		ctx:     ctx,
		bpm:     defaultBPM,
		closeCh: make(chan struct{}),
	}, nil
}

// SetBPM adjusts the beats-per-minute tempo.
func (e *OtoEngine) SetBPM(bpm int) {
	atomic.StoreInt32(&e.bpm, int32(bpm))
}

// LoadSample is a no-op for the built-in drum engine (no external samples).
func (e *OtoEngine) LoadSample(name, path string) error {
	return nil
}

// Play starts consuming sample frames from the channel and playing them through the audio device.
// This call blocks until the samples channel is closed and fully consumed by the player.
func (e *OtoEngine) Play(samples <-chan float64) error {
	// Use a custom io.Reader that reads float64 samples directly from the
	// channel and converts them to int16 bytes. This avoids io.Pipe which
	// can deadlock when the oto player's internal buffer fills up.
	r := &chanReader{ch: samples, done: make(chan struct{})}
	p := e.ctx.NewPlayer(r)
	p.Play()

	// Wait until the channel is closed and fully consumed by the reader.
	<-r.done
	return nil
}

// chanReader implements io.Reader by pulling float64 samples from a channel
// and converting them to float32 little-endian bytes (IEEE 754).
type chanReader struct {
	ch     <-chan float64
	buf    []byte
	done   chan struct{}
	doneOnce sync.Once
}

func (r *chanReader) Read(p []byte) (int, error) {
	// If we have leftover bytes from a previous read, use them first.
	if len(r.buf) > 0 {
		n := copy(p, r.buf)
		r.buf = r.buf[n:]
		return n, nil
	}

	// Read one sample, convert to float32 little-endian bytes.
	sample, ok := <-r.ch
	if !ok {
		// Channel closed — signal done and return EOF.
		r.doneOnce.Do(func() { close(r.done) })
		return 0, io.EOF
	}
	v := float32(sample)
	r.buf = (*[4]byte)(unsafe.Pointer(&v))[:]

	// Fill as much of p as we can from this sample.
	n := copy(p, r.buf)
	r.buf = r.buf[n:]
	return n, nil
}

// Close releases audio resources.
func (e *OtoEngine) Close() error {
	close(e.closeCh)
	return nil
}

// clampSample converts a float64 sample [-1,1] to int16 range.
func clampSample(s float64) int16 {
	if s > 1.0 {
		s = 1.0
	}
	if s < -1.0 {
		s = -1.0
	}
	return int16(s * 32767)
}