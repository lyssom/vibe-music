package playback

import (
	"math"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/lyssom/vibe-music/core/audio"
	"github.com/lyssom/vibe-music/core/pattern"
	"github.com/sinshu/go-meltysynth/meltysynth"
)

const sampleRate = 44100

// Engine connects the pattern scheduler to the audio engine, converting Notes to audio samples.
type Engine struct {
	mu        sync.Mutex
	audioEng  audio.Engine
	scheduler *pattern.Scheduler
	ast       *pattern.AST

	notes   chan pattern.Note
	samples chan float64

	// meltysynth for piano/lead
	synth *meltysynth.Synthesizer

	playing atomic.Bool
	bpm     atomic.Int32
	beat    atomic.Int32
	stopCh  chan struct{}
	doneCh  chan struct{}
}

// New creates a playback engine.
func New(audioEng audio.Engine) *Engine {
	e := &Engine{
		audioEng:    audioEng,
		notes:       make(chan pattern.Note, 256),
		samples:     make(chan float64, sampleRate),
		stopCh:      make(chan struct{}),
	}
	e.bpm.Store(120)

	// Load SoundFont for high-quality piano/lead sounds.
	sf2Path := "TimGM6mb.sf2"
	if sf2, err := os.Open(sf2Path); err == nil {
		defer sf2.Close()
		if sf, err := meltysynth.NewSoundFont(sf2); err == nil {
			settings := meltysynth.NewSynthesizerSettings(sampleRate)
			settings.BlockSize = 512
			settings.MaximumPolyphony = 32
			settings.EnableReverbAndChorus = true
			if synth, err := meltysynth.NewSynthesizer(sf, settings); err == nil {
				synth.MasterVolume = 0.4
				e.synth = synth
			}
		}
	}

	return e
}

// BPM returns the current BPM.
func (e *Engine) BPM() int {
	return int(e.bpm.Load())
}

// Beat returns the current beat position.
func (e *Engine) Beat() int {
	return int(e.beat.Load())
}

// IsPlaying returns true if the engine is currently playing.
func (e *Engine) IsPlaying() bool {
	return e.playing.Load()
}

// Synth returns the underlying SoundFont synthesizer (nil if not loaded).
func (e *Engine) Synth() *meltysynth.Synthesizer {
	return e.synth
}

// LoadAST sets the pattern to play. Can be called while playing to hot-swap.
func (e *Engine) LoadAST(ast *pattern.AST, bpm int) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.ast = ast
	e.bpm.Store(int32(bpm))
}

// Start begins playback. Idempotent if already playing.
func (e *Engine) Start() {
	if e.playing.Load() {
		return
	}

	e.mu.Lock()
	ast := e.ast
	bpm := int(e.bpm.Load())
	e.mu.Unlock()

	if bpm <= 0 {
		bpm = 120
	}

	e.playing.Store(true)
	e.scheduler = pattern.NewScheduler(bpm, e.notes)

	e.stopCh = make(chan struct{})
	e.doneCh = make(chan struct{})

	go e.runAudio()
	go e.runSampleGen()

	go func() {
		e.scheduler.Run(ast)
	}()

	go e.runBeatCounter(bpm)
}

// Stop halts playback.
func (e *Engine) Stop() {
	if !e.playing.Load() {
		return
	}
	e.playing.Store(false)

	if e.scheduler != nil {
		e.scheduler.Stop()
	}
	close(e.stopCh)
	if e.doneCh != nil {
		<-e.doneCh
	}
}

// runAudio reads samples and sends them to the audio engine.
func (e *Engine) runAudio() {
	defer close(e.doneCh)

	if e.audioEng == nil {
		for range e.samples {
			select {
			case <-e.stopCh:
				return
			default:
			}
		}
		return
	}

	ch := make(chan float64, sampleRate/10)

	go func() {
		_ = e.audioEng.Play(ch)
	}()

	for {
		select {
		case <-e.stopCh:
			close(ch)
			return
		case s, ok := <-e.samples:
			if !ok {
				close(ch)
				return
			}
			ch <- s
		}
	}
}

// noteOffTimer holds scheduled note-off events.
type noteOffTimer struct {
	absTime int64 // absolute sample time when to call NoteOff
	ch      int32
	key     int32
}

// runSampleGen converts Note events into audio samples using SoundFont synthesis.
func (e *Engine) runSampleGen() {
	var (
		drumSamples  [][]float64
		noteOffQueue []noteOffTimer
		sampleCount  int64
		blockLeft    []float32
		blockRight   []float32
		blockOff     int
	)

	// Pre-allocate meltysynth render buffers.
	const blockSize = 2048
	if e.synth != nil {
		blockLeft = make([]float32, blockSize)
		blockRight = make([]float32, blockSize)
	}

	// Schedule a note-off at a given sample index.
	scheduleOff := func(ch int32, key int32, samplesFromNow int) {
		noteOffQueue = append(noteOffQueue, noteOffTimer{
			absTime: sampleCount + int64(samplesFromNow),
			ch:      ch,
			key:     key,
		})
	}

	for {
		// Process pending note-off events.
		newQueue := noteOffQueue[:0]
		for _, t := range noteOffQueue {
			if sampleCount >= t.absTime {
				if e.synth != nil {
					e.synth.NoteOff(t.ch, t.key)
				}
			} else {
				newQueue = append(newQueue, t)
			}
		}
		noteOffQueue = newQueue

		// Drain any pending notes.
		for {
			select {
			case <-e.stopCh:
				return
			case note, ok := <-e.notes:
				if !ok {
					return
				}

				if note.Sample != "" {
					// Drum: synthesize via audio package.
					drum, ok := audio.LookupDrum(note.Sample)
					if !ok {
						continue
					}
					dur := int(float64(sampleRate) * note.Duration.Seconds())
					if dur < 100 {
						dur = 100
					}
					samps := audio.SynthesizeDrum(drum, note.Velocity, dur)
					drumSamples = append(drumSamples, samps)
				} else if note.Freq > 0 {
					// Pitched note → SoundFont.
					if e.synth == nil {
						continue
					}
					midiNote := pattern.FreqToMIDI(note.Freq)
					velocity := int32(math.Round(note.Velocity * 127))
					if velocity < 1 {
						velocity = 1
					}
					if velocity > 127 {
						velocity = 127
					}
					// Channel 0 = piano, channel 1 = bass.
					ch := int32(0)
					if note.Freq < 200 {
						ch = 1
					}
					e.synth.NoteOn(ch, midiNote, velocity)

					// Schedule note-off based on duration.
					samples := int(float64(sampleRate) * note.Duration.Seconds())
					if samples < sampleRate/10 {
						samples = sampleRate / 10
					}
					scheduleOff(ch, midiNote, samples)
				}
			default:
				goto mix
			}
		}

	mix:
		// Mix all sources into one output frame.
		var mixed float64

		// SoundFont synthesis.
		if e.synth != nil && blockLeft != nil {
			if blockOff >= blockSize {
				e.synth.Render(blockLeft, blockRight)
				blockOff = 0
			}
			if blockOff < blockSize {
				// Mix left+right channels into mono.
				s := (float64(blockLeft[blockOff]) + float64(blockRight[blockOff])) * 0.5
				mixed += s * 0.7
				blockOff++
			}
		}

		// Drum sample buffers.
		var stillActive [][]float64
		for _, buf := range drumSamples {
			if len(buf) > 0 {
				mixed += buf[0] * 0.8
				if len(buf) > 1 {
					stillActive = append(stillActive, buf[1:])
				}
			}
		}
		drumSamples = stillActive

		// Soft clip to prevent harsh clipping.
		if mixed > 1.0 {
			mixed = 1.0 - 1.0/(mixed+0.5)
		} else if mixed < -1.0 {
			mixed = -1.0 + 1.0/(mixed-0.5)
		}

		select {
		case e.samples <- mixed:
			sampleCount++
		case <-e.stopCh:
			return
		}
	}
}

// runBeatCounter increments the beat counter at BPM intervals.
func (e *Engine) runBeatCounter(bpm int) {
	interval := time.Duration(float64(time.Minute) / float64(bpm))
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-e.stopCh:
			return
		case <-ticker.C:
			e.beat.Add(1)
		}
	}
}