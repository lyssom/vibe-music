package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/lyssom/vibe-music/agent/generator"
	"github.com/lyssom/vibe-music/agent/song"
	"github.com/lyssom/vibe-music/core/audio"
	"github.com/lyssom/vibe-music/core/pattern"
	"github.com/lyssom/vibe-music/core/playback"
)

type AppModel struct {
	editor         textarea.Model
	codeContent    string
	playback       *playback.Engine
	generator      generator.Generator
	bpm            int
	playing        bool
	agentStatus    string
	agentStreaming bool
	errorMsg       string
	width          int
	height         int
	beat           int
	frame          int // for scrolling animation
	// new fields for composition mode
	composer    *song.Composer
	composeMode bool
}

const (
	cAccent = lipgloss.Color("46")
	cPink   = lipgloss.Color("201")
	cCyan   = lipgloss.Color("51")
	cOrange = lipgloss.Color("208")
	cMuted  = lipgloss.Color("239")
	cDim    = lipgloss.Color("240")
	cBright = lipgloss.Color("255")
	cBorder = lipgloss.Color("57")
)

var (
	accent = lipgloss.NewStyle().Foreground(cAccent)
	pink   = lipgloss.NewStyle().Foreground(cPink)
	cyan   = lipgloss.NewStyle().Foreground(cCyan)
	orange = lipgloss.NewStyle().Foreground(cOrange)
	muted  = lipgloss.NewStyle().Foreground(cDim)
	bright = lipgloss.NewStyle().Foreground(cBright)
	danger = lipgloss.NewStyle().Foreground(cOrange)

	lineNum    = lipgloss.NewStyle().Foreground(cMuted).Width(4).Align(lipgloss.Right)
	codeText   = lipgloss.NewStyle().Foreground(cBright)
	codeBorder = lipgloss.NewStyle().Foreground(cBorder)

	inputBox = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(cBorder).
			Padding(0, 1)

	// Beat colors — cycle per beat
	beatColors = []lipgloss.Color{
		lipgloss.Color("46"),  // green
		lipgloss.Color("51"),  // cyan
		lipgloss.Color("201"), // pink
		lipgloss.Color("208"), // orange
		lipgloss.Color("75"),  // blue
		lipgloss.Color("141"), // purple
		lipgloss.Color("220"), // gold
		lipgloss.Color("81"),  // teal
	}
)

type agentResultMsg struct{ response string; err error }

func NewAppModel(audioEng audio.Engine, gen generator.Generator) AppModel {
	ta := textarea.New()
	ta.Placeholder = "describe your beat..."
	ta.SetHeight(1)
	ta.Focus()
	ta.Prompt = ""
	ta.CharLimit = 300

	return AppModel{
		editor:      ta,
		playback:    playback.New(audioEng),
		generator:   gen,
		bpm:         120,
		agentStatus: "standby",
	}
}

func (m AppModel) Init() tea.Cmd {
	return textarea.Blink
}

func tickCmd() tea.Cmd {
	return tea.Tick(50*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg{}
	})
}

type tickMsg struct{}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tickMsg:
		// Refreshes beat position while playing

	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.editor.SetWidth(msg.Width - 4)

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			if m.playing { m.playback.Stop() }
			return m, tea.Quit
		case "/compose", "ctrl+/":
			m.startComposeMode()
		case " ":
			m.togglePlay()
		case "ctrl+r":
			m.runCode()
		case "enter":
			if input := strings.TrimSpace(m.editor.Value()); input != "" {
				if m.composeMode {
					m.processComposeInput(input)
					m.editor.SetValue("")
				} else {
					_, cmd := m.runAgent(input)
					m.editor.SetValue("")
					return m, cmd
				}
			}
		case "ctrl+d":
			m.editor.SetValue("")
		default:
			var cmd tea.Cmd
			m.editor, cmd = m.editor.Update(msg)
			return m, cmd
		}

	case agentResultMsg:
		if msg.err != nil {
			m.agentStatus, m.errorMsg, m.agentStreaming = "err", msg.err.Error(), false
		} else if code := strings.TrimSpace(generator.StripThinking(msg.response)); code != "" {
			m.codeContent, m.agentStatus, m.agentStreaming = code, "compiled", false
			m.runCode()
		}
	}

	// Track beat position
	if m.playing && m.playback != nil {
		m.beat = m.playback.Beat()
	}
	m.frame++

	cmds = append(cmds, tickCmd())
	return m, tea.Batch(cmds...)
}

func (m *AppModel) togglePlay() {
	if m.playing { m.playback.Stop(); m.playing, m.agentStatus = false, "paused" } else { m.runCode() }
}

func (m *AppModel) runCode() {
	if m.codeContent == "" { return }
	if ast, err := pattern.Parse(m.codeContent); err != nil {
		m.errorMsg, m.agentStatus = err.Error(), "parse error"
	} else {
		m.errorMsg = ""
		m.playback.LoadAST(ast, m.bpm)
		m.playback.Start()
		m.playing, m.agentStatus, m.beat = true, "live", 0
	}
}

func (m *AppModel) runAgent(prompt string) (tea.Model, tea.Cmd) {
	if m.generator == nil { m.agentStatus = "no key"; return m, nil }
	m.agentStreaming, m.agentStatus = true, "thinking"
	return m, func() tea.Msg {
		ctx := context.Background()
		ch, err := m.generator.GenerateStream(ctx, prompt, generator.PromptContext{CurrentCode: m.codeContent})
		if err != nil { return agentResultMsg{err: err} }
		var full strings.Builder
		for ev := range ch { full.WriteString(ev.Content) }
		return agentResultMsg{response: full.String()}
	}
}

func (m AppModel) View() string {
	w := max(m.width, 40)
	h := max(m.height, 8)

	if m.composeMode {
		return m.renderComposeView()
	}

	var b strings.Builder

	// ══════════════════════════════════════════════════
	// HEADER
	// ══════════════════════════════════════════════════
	b.WriteString(pink.Render("▓ VIBE ECHO ▓"))
	b.WriteString(" " + cyan.Render("v0.1.0"))
	b.WriteString("  " + muted.Render(fmt.Sprintf("BPM:%d", m.bpm)))
	b.WriteString("\n")
	b.WriteString(m.renderVisualizerBar(w))
	b.WriteString("\n")

	// ══════════════════════════════════════════════════
	// CODE ZONE — full screen, numbered
	// ══════════════════════════════════════════════════
	codeLines := h - 6
	if codeLines < 2 { codeLines = 2 }
	b.WriteString(m.renderCodeZone(w, codeLines))
	b.WriteString("\n")

	// ══════════════════════════════════════════════════
	// INPUT — tiny framed box
	// ══════════════════════════════════════════════════
	b.WriteString(inputBox.Render(m.editor.View()))
	b.WriteString("\n")

	// ══════════════════════════════════════════════════
	// STATUS BAR
	// ══════════════════════════════════════════════════
	b.WriteString(m.renderStatusLine(w))

	return b.String()
}

// renderVisualizerBar — silent line vs dense flowing marquee waveform
func (m AppModel) renderVisualizerBar(w int) string {
	if !m.playing || m.playback == nil {
		return codeBorder.Render(strings.Repeat("─", w))
	}

	beatPos := m.beat % 4
	beatIdx := beatPos % len(beatColors)
	col := beatColors[beatIdx]

	// Scroll offset from frame counter — smooth horizontal movement
	offset := m.frame % 13

	// Dense wave pattern
	wave := "▁▂▃▄▅▆▇█▇▆▅▄▃"

	// Shift by offset
	shifted := wave[offset:] + wave[:offset]
	repeats := w/len(shifted) + 1

	return lipgloss.NewStyle().Foreground(col).Bold(true).Render(
		strings.Repeat(shifted, repeats)[:w],
	)
}

func (m AppModel) renderCodeZone(w int, lines int) string {
	var b strings.Builder

	if m.codeContent == "" {
		// Empty state — hint in center
		for i := 0; i < lines/2-1; i++ {
			b.WriteString(lineNum.Render("") + "\n")
		}
		b.WriteString(muted.Render(fmt.Sprintf("%*s", w/2+10, "type a beat in the box below...")))
		b.WriteString("\n")
		for i := lines / 2; i < lines; i++ {
			b.WriteString(lineNum.Render("") + "\n")
		}
		return b.String()
	}

	codeW := w - 6
	if codeW < 10 { codeW = 10 }
	codeLines := strings.Split(m.codeContent, "\n")

	for i := 0; i < lines; i++ {
		b.WriteString(lineNum.Render(fmt.Sprintf("%d", i+1)))
		b.WriteString(" ")
		if i < len(codeLines) {
			trunc := codeLines[i]
			if len(trunc) > codeW { trunc = trunc[:codeW] }
			b.WriteString(codeText.Render(trunc))
		}
		b.WriteString("\n")
	}
	return b.String()
}

func (m AppModel) renderStatusLine(w int) string {
	play := accent.Render("▶")
	if !m.playing { play = orange.Render("⏸") }

	var statusStr string
	switch m.agentStatus {
	case "live": statusStr = accent.Render("LIVE")
	case "paused": statusStr = orange.Render("PAUSED")
	case "thinking": statusStr = pink.Render("THINKING")
	case "no key": statusStr = danger.Render("NO KEY")
	default: statusStr = muted.Render("STANDBY")
	}

	left := play + " " + statusStr
	if m.errorMsg != "" {
		left += " " + danger.Render(m.errorMsg)
	}

	right := muted.Render("space play  ·  ^R run  ·  ^D clear  ·  ^C quit")
	pad := max(0, w-len(left)-len(right))
	return left + strings.Repeat(" ", pad) + right
}

// startComposeMode initializes the composition mode
func (m *AppModel) startComposeMode() {
	if m.generator == nil {
		m.agentStatus = "no key"
		return
	}
	m.composer = song.NewComposer(m.generator)
	m.composeMode = true
	question := m.composer.StartSession("")
	// question is displayed in renderComposeView via GetDialogHistory
	_ = question // consumed by StartSession
	m.agentStatus = "composing"
}

// processComposeInput handles user input in composition mode
func (m *AppModel) processComposeInput(input string) {
	// Check for command shortcuts
	inputLower := strings.ToLower(input)
	if inputLower == "/quit" || inputLower == "/done" {
		m.exitComposeMode()
		return
	}
	if inputLower == "/back" {
		if err := m.composer.Rollback(); err != nil {
			m.errorMsg = err.Error()
		}
		return
	}

	turn, question, err := m.composer.ProcessResponse(input)
	if err != nil {
		m.errorMsg = err.Error()
		return
	}

	// Add assistant turn to history
	if turn != nil && turn.Role == "assistant" {
		m.composer.GetDialogHistory() // ensure history is updated
	}

	if question.Kind == "done" {
		// Generation complete
		m.codeContent = m.composer.GetSong().Sections[0].DSLCode
		m.composeMode = false
		m.agentStatus = "complete"
		return
	}

	m.agentStatus = "composing"
}

// exitComposeMode exits composition mode
func (m *AppModel) exitComposeMode() {
	m.composeMode = false
	m.composer = nil
	m.agentStatus = "standby"
}

// renderComposeView renders the composition mode UI
func (m AppModel) renderComposeView() string {
	w := max(m.width, 40)
	h := max(m.height, 8)

	var b strings.Builder

	// Header
	b.WriteString(pink.Render("▓ SONG COMPOSER ▓"))
	b.WriteString(" " + cyan.Render("多轮创作模式"))
	b.WriteString("\n")
	b.WriteString(codeBorder.Render(strings.Repeat("─", w)))
	b.WriteString("\n\n")

	// Dialog history
	history := m.composer.GetDialogHistory()
	contentLines := h - 8
	if contentLines < 4 {
		contentLines = 4
	}

	// Build content area
	var content strings.Builder
	for _, turn := range history {
		if turn.Role == "user" {
			content.WriteString(muted.Render("你: "))
			content.WriteString(turn.Content)
		} else {
			content.WriteString(accent.Render("AI: "))
			content.WriteString(turn.Content)
		}
		content.WriteString("\n\n")
	}

	// Truncate if too long
	contentStr := content.String()
	lines := strings.Split(contentStr, "\n")
	if len(lines) > contentLines {
		lines = lines[len(lines)-contentLines:]
		contentStr = strings.Join(lines, "\n")
		content.Reset()
		content.WriteString(muted.Render("...\n\n"))
		content.WriteString(contentStr)
	}
	b.WriteString(content.String())

	// Duration estimate
	b.WriteString(muted.Render(fmt.Sprintf("总时长: %s", m.composer.GetDurationEstimate())))
	b.WriteString("\n\n")

	// Input area
	b.WriteString(inputBox.Render(m.editor.View()))
	b.WriteString("\n")

	// Status bar
	b.WriteString(m.renderComposeStatusLine())

	return b.String()
}

// renderComposeStatusLine renders the status line for compose mode
func (m AppModel) renderComposeStatusLine() string {
	left := accent.Render("COMPOSING")
	right := muted.Render("/done 完成  ·  /back 回退  ·  /quit 退出")
	pad := max(0, m.width-len(left)-len(right))
	return left + strings.Repeat(" ", pad) + right
}