package tui_test

import (
	"sync"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/lyssom/vibe-music/core/audio"
	"github.com/lyssom/vibe-music/tui"
)

var (
	testAudioEng audio.Engine
	testAudioErr error
	testEngOnce  sync.Once
)

func getTestAudioEngine(t *testing.T) audio.Engine {
	t.Helper()
	testEngOnce.Do(func() {
		testAudioEng, testAudioErr = audio.NewOtoEngine()
	})
	if testAudioErr != nil {
		t.Skipf("audio device not available: %v", testAudioErr)
	}
	return testAudioEng
}

func TestAppModelInit(t *testing.T) {
	eng := getTestAudioEngine(t)

	m := tui.NewAppModel(eng, nil)
	cmd := m.Init()
	if cmd == nil {
		t.Fatal("expected non-nil init command")
	}
}

func TestAppModelQuitOnCtrlC(t *testing.T) {
	eng := getTestAudioEngine(t)

	m := tui.NewAppModel(eng, nil)
	msg := tea.KeyMsg{Type: tea.KeyCtrlC}
	_, cmd := m.Update(msg)
	if cmd == nil {
		t.Fatal("expected quit command on ctrl+c")
	}
}

func TestAppModelViewNotEmpty(t *testing.T) {
	eng := getTestAudioEngine(t)

	m := tui.NewAppModel(eng, nil)
	v := m.View()
	if v == "" {
		t.Error("view should render content")
	}
}

func TestAppModelCommandMode(t *testing.T) {
	eng := getTestAudioEngine(t)

	m := tui.NewAppModel(eng, nil)

	// Enter command mode with /
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}, Alt: false}
	updated, _ := m.Update(msg)
	m2 := updated.(tui.AppModel)

	view := m2.View()
	if view == "" {
		t.Error("view should render in command mode")
	}

	// Exit command mode with esc
	escMsg := tea.KeyMsg{Type: tea.KeyEsc}
	updated2, _ := m2.Update(escMsg)
	m3 := updated2.(tui.AppModel)

	view2 := m3.View()
	if view2 == "" {
		t.Error("view should render after exiting command mode")
	}
}