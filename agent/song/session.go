package song

import (
	"fmt"
	"time"
)

// Session manages the composition session state
type Session struct {
	state       SessionState
	turnCounter int
}

// NewSession creates a new composition session
func NewSession(mode InteractionMode) *Session {
	return &Session{
		state: SessionState{
			Mode:         mode,
			CurrentPhase: PhaseIntent,
			CurrentNode:  -1,
		},
		turnCounter: 0,
	}
}

// GetState returns the current session state
func (s *Session) GetState() SessionState {
	return s.state
}

// GetElements returns the current song elements
func (s *Session) GetElements() SongElements {
	return s.state.Elements
}

// SetElement sets a specific element value
func (s *Session) SetElement(kind, value string) {
	switch kind {
	case "genre":
		s.state.Elements.Genre = value
	case "emotion":
		s.state.Elements.Emotion = value
	case "rhythm":
		s.state.Elements.Rhythm = value
	case "instrumentation":
		s.state.Elements.Instrumentation = value
	case "scale":
		s.state.Elements.Scale = value
	case "harmony":
		s.state.Elements.Harmony = value
	case "bpm":
		s.state.Elements.BPM = parseBPM(value)
	case "dynamic":
		s.state.Elements.Dynamic = value
	case "structure":
		s.state.Elements.Structure = value
	case "techniques":
		s.state.Elements.Techniques = value
	}
}

// AddDialogTurn adds a conversation turn
func (s *Session) AddDialogTurn(role, content string) {
	s.turnCounter++
	s.state.History = append(s.state.History, DialogTurn{
		Role:    role,
		Content: content,
		Turn:    s.turnCounter,
	})
}

// SaveHistoryNode creates a checkpoint
func (s *Session) SaveHistoryNode() {
	node := HistoryNode{
		Timestamp:  time.Now(),
		Phase:      s.state.CurrentPhase,
		Elements:   s.state.Elements,
		Structure:  s.state.SelectedStructure,
		Song:       s.state.Song,
	}
	s.state.HistoryNodes = append(s.state.HistoryNodes, node)
	s.state.CurrentNode = len(s.state.HistoryNodes) - 1
}

// RollbackTo reverts to a previous checkpoint
func (s *Session) RollbackTo(nodeIndex int) error {
	if nodeIndex < 0 || nodeIndex >= len(s.state.HistoryNodes) {
		return fmt.Errorf("invalid node index: %d", nodeIndex)
	}
	node := s.state.HistoryNodes[nodeIndex]
	s.state.Elements = node.Elements
	s.state.SelectedStructure = node.Structure
	s.state.Song = node.Song
	s.state.CurrentNode = nodeIndex
	return nil
}

// RollbackLast reverts to the previous checkpoint
func (s *Session) RollbackLast() error {
	if s.state.CurrentNode < 0 {
		return fmt.Errorf("no history to rollback")
	}
	return s.RollbackTo(s.state.CurrentNode)
}

// GetHistoryNodes returns all history nodes for UI
func (s *Session) GetHistoryNodes() []HistoryNode {
	return s.state.HistoryNodes
}

// SetPhase sets the current session phase
func (s *Session) SetPhase(phase SessionPhase) {
	s.state.CurrentPhase = phase
}

// SetSong sets the current song
func (s *Session) SetSong(song *Song) {
	s.state.Song = song
}

// GetSong returns the current song
func (s *Session) GetSong() *Song {
	return s.state.Song
}

// SetStructure sets the selected structure
func (s *Session) SetStructure(structure []SectionType) {
	s.state.SelectedStructure = structure
}

// parseBPM converts string to BPM
func parseBPM(value string) int {
	var bpm int
	fmt.Sscanf(value, "%d", &bpm)
	if bpm <= 0 {
		bpm = 120 // default
	}
	return bpm
}