package pattern

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/lyssom/vibe-music/core/audio"
)

// Scheduler evaluates pattern ASTs and dispatches note events in time.
type Scheduler struct {
	mu       sync.Mutex
	bpm      int
	notes    chan Note
	ctx      context.Context
	cancel   context.CancelFunc
	beat     int // global beat count
	subBeat  int // 16th note sub-divisions (0-15 per beat)
	swing    int // swing strength 0-100
	subdiv   int // subdivision multiplier (4=16th notes)
	interval time.Duration
}

// schedCmd pairs a command with its beat timing.
type schedCmd struct {
	cmd   Command
	every int // beat cycle length in beats
	phase int // which beat in the cycle fires (0-based)
}

// NewScheduler creates a pattern scheduler with the given BPM and output channel.
func NewScheduler(bpm int, notes chan Note) *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())
	s := &Scheduler{
		bpm:     bpm,
		notes:   notes,
		ctx:     ctx,
		cancel:  cancel,
		swing:   0,
		subdiv:  4,
		subBeat: 0,
		beat:    0,
	}
	s.interval = time.Minute / time.Duration(bpm*s.subdiv)
	return s
}

// SetBPM changes the tempo.
func (s *Scheduler) SetBPM(bpm int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.bpm = bpm
	s.interval = time.Minute / time.Duration(bpm*s.subdiv)
}

// SetSwing sets the swing strength (0-100, 0=none, 100=full swing).
func (s *Scheduler) SetSwing(swing int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.swing = swing
}

// Stop halts the scheduler.
func (s *Scheduler) Stop() {
	s.cancel()
}

// Run evaluates the AST in a loop at sub-beat resolution (16th notes).
func (s *Scheduler) Run(ast *AST) {
	s.mu.Lock()
	swing := s.swing
	interval := s.interval
	s.mu.Unlock()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Pre-process: collect commands with their beat timing.
	var sched []schedCmd
	for _, c := range ast.Commands {
		every := c.Every
		if every < 1 {
			every = 1
		}
		sched = append(sched, schedCmd{
			cmd:   c,
			every: every,
			phase: c.BeatOffset % every,
		})
	}

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.evaluateSubBeat(sched, swing)
		}
	}
}

func (s *Scheduler) evaluateSubBeat(sched []schedCmd, swing int) {
	// Only process on main beat positions (subBeat 0,4,8,12 = beats 1,2,3,4)
	if s.subBeat%4 != 0 {
		s.advance()
		return
	}

	// Advance main beat counter every 4 sub-beats
	if s.subBeat > 0 {
		s.beat++
	}

	for _, sc := range sched {
		beatInCycle := s.beat % sc.every
		if beatInCycle != sc.phase {
			continue
		}

		switch sc.cmd.Name {
		case "sound":
			s.execSound(sc.cmd)
		case "note":
			s.execNote(sc.cmd)
		case "chord":
			s.execChord(sc.cmd)
		case "bass":
			s.execBass(sc.cmd)
		}
	}

	s.advance()
}

func (s *Scheduler) advance() {
	s.subBeat++
	if s.subBeat >= 16 {
		s.subBeat = 0
	}
}

// execSound triggers drum samples.
func (s *Scheduler) execSound(cmd Command) {
	if len(cmd.Args) == 0 {
		return
	}

	drumStr := cmd.Args[0].Value
	drumNames := strings.Fields(drumStr)

	for _, name := range drumNames {
		if drum, ok := audio.LookupDrum(name); ok {
			note := Note{
				Sample:   name,
				Velocity: 0.8,
				Duration: 100 * time.Millisecond,
				Pitch:    float64(drum),
			}
			select {
			case s.notes <- note:
			default:
				// Channel full, drop note
			}
		}
	}
}

// execNote handles the "note" command for pitched instruments.
// Syntax: note("c3 eb4 g5", "8n", 0.7) — notes, duration, velocity
func (s *Scheduler) execNote(cmd Command) {
	if len(cmd.Args) == 0 {
		return
	}

	noteStr := cmd.Args[0].Value
	noteDur := "4n"
	velocity := 0.7

	if len(cmd.Args) > 1 {
		noteDur = cmd.Args[1].Value
	}
	if len(cmd.Args) > 2 {
		var v float64
		fmt.Sscanf(cmd.Args[2].Value, "%f", &v)
		if v > 0 && v <= 1 {
			velocity = v
		}
	}

	duration := NoteDuration(noteDur, s.bpm)
	if duration == 0 {
		duration = time.Minute / time.Duration(s.bpm)
	}

	noteNames := strings.Fields(noteStr)
	for _, name := range noteNames {
		freq := NoteToFreq(name)
		note := Note{
			Freq:     freq,
			Velocity: velocity,
			Duration: duration,
		}
		select {
		case s.notes <- note:
		default:
		}
	}
}

// execChord handles the "chord" command (multiple notes at once).
// Syntax: chord("c3 e3 g3", "4n", 0.5)
func (s *Scheduler) execChord(cmd Command) {
	s.execNote(cmd) // Same logic, just different name
}

// execBass handles the "bass" command (single bass note).
// Syntax: bass("c2", "2n", 0.9)
func (s *Scheduler) execBass(cmd Command) {
	if len(cmd.Args) == 0 {
		return
	}
	noteStr := cmd.Args[0].Value
	noteDur := "2n"
	velocity := 0.9

	if len(cmd.Args) > 1 {
		noteDur = cmd.Args[1].Value
	}
	if len(cmd.Args) > 2 {
		var v float64
		fmt.Sscanf(cmd.Args[2].Value, "%f", &v)
		if v > 0 && v <= 1 {
			velocity = v
		}
	}

	duration := NoteDuration(noteDur, s.bpm)
	freq := NoteToFreq(noteStr)
	note := Note{
		Freq:     freq,
		Velocity: velocity,
		Duration: duration,
	}
	select {
	case s.notes <- note:
	default:
	}
}