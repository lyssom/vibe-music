//go:build windows

package audio

import (
	"fmt"
	"sync"
	"syscall"
	"time"
	"unsafe"
)

// winmmEngine implements Engine using the Windows legacy waveOut API (winmm.dll).
// This is more reliable than WASAPI on some Windows configurations — it's the same
// API used by System.Media.SoundPlayer.
type winmmEngine struct {
	hWaveOut syscall.Handle
	bpm      int32
	mu       sync.Mutex
	playing  bool
}

var (
	winmmDLL                     = syscall.NewLazyDLL("winmm.dll")
	procWaveOutOpen              = winmmDLL.NewProc("waveOutOpen")
	procWaveOutClose             = winmmDLL.NewProc("waveOutClose")
	procWaveOutWrite             = winmmDLL.NewProc("waveOutWrite")
	procWaveOutReset             = winmmDLL.NewProc("waveOutReset")
	procWaveOutUnprepareHeader   = winmmDLL.NewProc("waveOutUnprepareHeader")
	procWaveOutPrepareHeader     = winmmDLL.NewProc("waveOutPrepareHeader")
	procWaveOutPause             = winmmDLL.NewProc("waveOutPause")
	procWaveOutRestart           = winmmDLL.NewProc("waveOutRestart")
)

const (
	_WAVE_FORMAT_PCM = 1
	_WHDR_DONE       = 0x00000001
	_WHDR_PREPARED   = 0x00000002
	_WAVERR_STILLPLAYING = 33
)

type _WAVEFORMATEX struct {
	wFormatTag      uint16
	nChannels       uint16
	nSamplesPerSec  uint32
	nAvgBytesPerSec uint32
	nBlockAlign     uint16
	wBitsPerSample  uint16
	cbSize          uint16
}

type _WAVEHDR struct {
	lpData          uintptr
	dwBufferLength  uint32
	dwBytesRecorded uint32
	dwUser          uintptr
	dwFlags         uint32
	dwLoops         uint32
	lpNext          uintptr
	reserved        uintptr
}

// NewWinMMEngine creates an audio engine using the Windows waveOut API.
func NewWinMMEngine() (*winmmEngine, error) {
	wf := _WAVEFORMATEX{
		wFormatTag:      _WAVE_FORMAT_PCM,
		nChannels:       1,
		nSamplesPerSec:  sampleRate,
		nAvgBytesPerSec: sampleRate * 2, // 16-bit mono
		nBlockAlign:     2,
		wBitsPerSample:  16,
	}

	var hWaveOut syscall.Handle
	ret, _, _ := procWaveOutOpen.Call(
		uintptr(unsafe.Pointer(&hWaveOut)),
		uintptr(0xFFFFFFFF), // WAVE_MAPPER
		uintptr(unsafe.Pointer(&wf)),
		0, 0,
		uintptr(0x00030000), // CALLBACK_FUNCTION | WAVE_ALLOWSYNC
	)
	if ret != 0 {
		return nil, fmt.Errorf("waveOutOpen failed: MMSYSERR=%d", ret)
	}

	return &winmmEngine{
		hWaveOut: hWaveOut,
		bpm:      defaultBPM,
	}, nil
}

func (e *winmmEngine) SetBPM(bpm int) {}

func (e *winmmEngine) LoadSample(name, path string) error { return nil }

func (e *winmmEngine) Play(samples <-chan float64) error {
	e.mu.Lock()
	if e.playing {
		e.mu.Unlock()
		return fmt.Errorf("already playing")
	}
	e.playing = true
	e.mu.Unlock()

	defer func() {
		e.mu.Lock()
		e.playing = false
		e.mu.Unlock()
	}()

	// Simple approach: accumulate ~100ms chunks, write a block, wait for
	// it to finish, then write the next block. This has ~100ms latency
	// but is dead simple and reliable.
	chunkSize := sampleRate / 10 // 100ms at 44100Hz

	for {
		buf := make([]int16, chunkSize)
		n := 0
		done := false

		// Fill buffer
		for n < chunkSize {
			select {
			case s, ok := <-samples:
				if !ok {
					done = true
					goto writeChunk
				}
				buf[n] = clampSample(s)
				n++
			default:
				// Very brief wait if no samples available yet
				if n > 0 {
					goto writeChunk
				}
				s, ok := <-samples
				if !ok {
					done = true
					goto writeChunk
				}
				buf[n] = clampSample(s)
				n++
			}
		}

	writeChunk:
		if n == 0 {
			if done {
				return nil
			}
			continue
		}

		// Pad with silence
		for i := n; i < chunkSize; i++ {
			buf[i] = 0
		}

		hdr := _WAVEHDR{
			lpData:         uintptr(unsafe.Pointer(&buf[0])),
			dwBufferLength: uint32(chunkSize * 2),
		}

		ret, _, _ := procWaveOutPrepareHeader.Call(
			uintptr(e.hWaveOut),
			uintptr(unsafe.Pointer(&hdr)),
			uintptr(unsafe.Sizeof(_WAVEHDR{})),
		)
		if ret != 0 {
			return fmt.Errorf("waveOutPrepareHeader failed: %d", ret)
		}

		ret, _, _ = procWaveOutWrite.Call(
			uintptr(e.hWaveOut),
			uintptr(unsafe.Pointer(&hdr)),
			uintptr(unsafe.Sizeof(_WAVEHDR{})),
		)
		if ret != 0 {
			procWaveOutUnprepareHeader.Call(
				uintptr(e.hWaveOut),
				uintptr(unsafe.Pointer(&hdr)),
				uintptr(unsafe.Sizeof(_WAVEHDR{})),
			)
			return fmt.Errorf("waveOutWrite failed: %d", ret)
		}

		// Wait for this block to finish playing
		for hdr.dwFlags&_WHDR_DONE == 0 {
			time.Sleep(time.Millisecond)
		}

		procWaveOutUnprepareHeader.Call(
			uintptr(e.hWaveOut),
			uintptr(unsafe.Pointer(&hdr)),
			uintptr(unsafe.Sizeof(_WAVEHDR{})),
		)

		if done {
			return nil
		}
	}
}

func (e *winmmEngine) Close() error {
	procWaveOutReset.Call(uintptr(e.hWaveOut))
	procWaveOutClose.Call(uintptr(e.hWaveOut))
	return nil
}
