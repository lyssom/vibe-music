package song

import (
	"testing"
	"time"
)

func TestNewSession(t *testing.T) {
	session := NewSession(ModeFull)
	if session == nil {
		t.Fatal("NewSession returned nil")
	}
	if session.GetState().Mode != ModeFull {
		t.Errorf("Expected ModeFull, got %v", session.GetState().Mode)
	}
	if session.GetState().CurrentPhase != PhaseIntent {
		t.Errorf("Expected PhaseIntent, got %v", session.GetState().CurrentPhase)
	}
}

func TestSessionSetElement(t *testing.T) {
	session := NewSession(ModeFull)

	tests := []struct {
		kind  string
		value string
		check func(SongElements) bool
	}{
		{"genre", "jazz", func(e SongElements) bool { return e.Genre == "jazz" }},
		{"emotion", "melancholic", func(e SongElements) bool { return e.Emotion == "melancholic" }},
		{"rhythm", "upbeat", func(e SongElements) bool { return e.Rhythm == "upbeat" }},
		{"bpm", "120", func(e SongElements) bool { return e.BPM == 120 }},
	}

	for _, tt := range tests {
		session.SetElement(tt.kind, tt.value)
		if !tt.check(session.GetElements()) {
			t.Errorf("SetElement(%s, %s) failed", tt.kind, tt.value)
		}
	}
}

func TestSessionAddDialogTurn(t *testing.T) {
	session := NewSession(ModeFull)

	session.AddDialogTurn("user", "hello")
	session.AddDialogTurn("assistant", "hi there")

	history := session.GetState().History
	if len(history) != 2 {
		t.Errorf("Expected 2 turns, got %d", len(history))
	}
	if history[0].Role != "user" || history[0].Content != "hello" {
		t.Errorf("First turn incorrect: %+v", history[0])
	}
	if history[1].Role != "assistant" || history[1].Content != "hi there" {
		t.Errorf("Second turn incorrect: %+v", history[1])
	}
}

func TestSessionSaveHistoryNode(t *testing.T) {
	session := NewSession(ModeFull)
	session.SetElement("genre", "jazz")

	// Save first node
	session.SaveHistoryNode()
	nodes := session.GetHistoryNodes()
	if len(nodes) != 1 {
		t.Errorf("Expected 1 history node, got %d", len(nodes))
	}

	// Modify and save second node
	session.SetElement("emotion", "happy")
	session.SaveHistoryNode()

	nodes = session.GetHistoryNodes()
	if len(nodes) != 2 {
		t.Errorf("Expected 2 history nodes, got %d", len(nodes))
	}
}

func TestSessionRollback(t *testing.T) {
	session := NewSession(ModeFull)
	session.SetElement("genre", "jazz")
	session.SaveHistoryNode()

	session.SetElement("genre", "pop")
	session.SaveHistoryNode()

	// Rollback to first node
	if err := session.RollbackTo(0); err != nil {
		t.Errorf("RollbackTo failed: %v", err)
	}
	if session.GetElements().Genre != "jazz" {
		t.Errorf("Expected genre 'jazz', got '%s'", session.GetElements().Genre)
	}

	// Rollback last
	session.SetElement("genre", "rock")
	session.SaveHistoryNode()
	if err := session.RollbackLast(); err != nil {
		t.Errorf("RollbackLast failed: %v", err)
	}
	if session.GetElements().Genre != "pop" {
		t.Errorf("Expected genre 'pop', got '%s'", session.GetElements().Genre)
	}
}

func TestSessionInvalidRollback(t *testing.T) {
	session := NewSession(ModeFull)

	if err := session.RollbackTo(-1); err == nil {
		t.Error("Expected error for invalid node index -1")
	}

	if err := session.RollbackTo(100); err == nil {
		t.Error("Expected error for invalid node index 100")
	}

	if err := session.RollbackLast(); err == nil {
		t.Error("Expected error for rollback with no history")
	}
}

func TestDetermineMode(t *testing.T) {
	tests := []struct {
		prompt string
		want   InteractionMode
	}{
		{"generate a drum beat", ModeSimple},
		{"写个循环节拍", ModeSimple},
		{"loop pattern", ModeSimple},
		{"create a nice melody", ModeStandard},
		{"主歌副歌桥段", ModeFull}, // contains Chinese section keywords
		{"write a song", ModeFull}, // contains "song"
	}

	for _, tt := range tests {
		got := DetermineMode(tt.prompt)
		if got != tt.want {
			t.Errorf("DetermineMode(%q) = %v, want %v", tt.prompt, got, tt.want)
		}
	}
}

func TestExplorerNextQuestion(t *testing.T) {
	session := NewSession(ModeFull)
	explorer := NewExplorer(session)

	q := explorer.NextQuestion(0)
	if q.Kind != "genre" {
		t.Errorf("First question should be genre, got %s", q.Kind)
	}
	if len(q.Options) == 0 {
		t.Error("genre question should have options")
	}
}

func TestExplorerIsComplete(t *testing.T) {
	session := NewSession(ModeFull)
	explorer := NewExplorer(session)

	// Not complete without genre
	if explorer.IsComplete() {
		t.Error("Should not be complete without elements")
	}

	// Set genre
	session.SetElement("genre", "jazz")
	if explorer.IsComplete() {
		t.Error("Should not be complete with only genre")
	}

	// Set all required for ModeFull
	session.SetElement("emotion", "melancholic")
	session.SetElement("rhythm", "upbeat")
	session.SetElement("instrumentation", "piano")
	session.SetElement("scale", "minor")
	session.SetElement("harmony", "functional")
	session.SetElement("bpm", "120")
	session.SetElement("dynamic", "soft-loud")
	session.SetElement("structure", "AABA")
	session.SetElement("techniques", "solo")

	if !explorer.IsComplete() {
		t.Error("Should be complete with all elements set")
	}
}

func TestStructurerProposeStructures(t *testing.T) {
	session := NewSession(ModeFull)
	session.SetElement("genre", "jazz")
	structurer := NewStructurer(session)

	proposals := structurer.ProposeStructures(3)
	if len(proposals) == 0 {
		t.Error("Expected proposals for jazz genre")
	}

	// Check first proposal is AABA style
	if proposals[0].Name == "" {
		t.Error("Proposal name should not be empty")
	}
}

func TestStructurerBuildSongFromStructure(t *testing.T) {
	session := NewSession(ModeFull)
	session.SetElement("genre", "jazz")
	structurer := NewStructurer(session)

	proposals := structurer.ProposeStructures(1)
	proposal := proposals[0]

	song := structurer.BuildSongFromStructure("Test Song", proposal, 120)
	if song.Title != "Test Song" {
		t.Errorf("Expected title 'Test Song', got '%s'", song.Title)
	}
	if len(song.Sections) == 0 {
		t.Error("Expected at least one section")
	}
	if song.EstimatedDuration == 0 {
		t.Error("Expected non-zero duration")
	}
}

func TestBuildSongSectionIDs(t *testing.T) {
	session := NewSession(ModeFull)
	session.SetElement("genre", "jazz")
	structurer := NewStructurer(session)

	proposals := structurer.ProposeStructures(1)
	song := structurer.BuildSongFromStructure("Test", proposals[0], 120)

	// Check that section IDs are unique
	ids := make(map[string]bool)
	for _, s := range song.Sections {
		if ids[s.ID] {
			t.Errorf("Duplicate section ID: %s", s.ID)
		}
		ids[s.ID] = true
	}
}

func TestSessionStateGetters(t *testing.T) {
	session := NewSession(ModeFull)

	// Test ProposedStructures getter/setter
	session.SetProposedStructures([]StructureProposal{
		{Name: "Test 1", TotalBars: 32},
	})
	if len(session.GetProposedStructures()) != 1 {
		t.Error("GetProposedStructures failed")
	}

	// Test SelectedStructure getter/setter
	session.SetStructure([]SectionType{SectionIntro, SectionVerse})
	structure := session.GetSelectedStructure()
	if len(structure) != 2 || structure[0] != SectionIntro {
		t.Error("GetSelectedStructure failed")
	}
}

func TestHistoryNodeTimestamp(t *testing.T) {
	session := NewSession(ModeFull)
	before := time.Now()

	session.SaveHistoryNode()

	after := time.Now()
	nodes := session.GetHistoryNodes()
	node := nodes[0]

	if node.Timestamp.Before(before) || node.Timestamp.After(after) {
		t.Error("HistoryNode timestamp out of range")
	}
}