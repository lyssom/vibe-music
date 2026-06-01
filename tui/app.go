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
	// Track pending AI response for multi-round conversation
	currentQuestion string
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

	// Keywords that trigger multi-round song composition mode
	songKeywords = []string{"歌曲", "写歌", "主歌", "副歌", "桥段", "bridge", "verse", "chorus", "完整", "完整歌曲", "jazz", "爵士", "pop", "流行", "rock", "摇滚"}
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

// shouldStartComposeMode checks if input indicates user wants multi-round song composition
func shouldStartComposeMode(input string) bool {
	inputLower := strings.ToLower(input)
	for _, keyword := range songKeywords {
		if strings.Contains(inputLower, strings.ToLower(keyword)) {
			return true
		}
	}
	return false
}

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
		case "ctrl+/":
			m.startComposeMode()
		case " ":
			m.togglePlay()
		case "ctrl+r":
			m.runCode()
		case "enter":
			input := strings.TrimSpace(m.editor.Value())
			if input == "" {
				return m, nil
			}

			// Handle explicit commands
			if strings.HasPrefix(input, "/") {
				m.handleCommand(input)
				m.editor.SetValue("")
				return m, nil
			}

			// Auto-detect song composition mode based on keywords
			if shouldStartComposeMode(input) {
				m.startComposeModeWithPrompt(input)
				m.editor.SetValue("")
				return m, nil
			}

			// Process input based on current mode
			if m.composeMode {
				m.processComposeInput(input)
				m.editor.SetValue("")
				return m, nil
			}

			// Normal single-shot generation
			_, cmd := m.runAgent(input)
			m.editor.SetValue("")
			return m, cmd

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

// handleCommand processes slash commands
func (m *AppModel) handleCommand(cmd string) {
	switch strings.ToLower(cmd) {
	case "/compose":
		m.startComposeMode()
	case "/quit", "/exit":
		m.exitComposeMode()
	case "/back":
		if m.composer != nil {
			if err := m.composer.Rollback(); err != nil {
				m.errorMsg = err.Error()
			}
		}
	case "/help":
		m.agentStatus = "help"
		m.errorMsg = "/compose - 开启多轮歌曲创作  /quit - 退出创作  /back - 回退一步"
	}
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
	case "composing": statusStr = cyan.Render("COMPOSING")
	case "complete": statusStr = accent.Render("COMPLETE")
	case "help": statusStr = muted.Render(m.errorMsg)
	default: statusStr = muted.Render("STANDBY")
	}

	left := play + " " + statusStr
	if m.errorMsg != "" && m.agentStatus != "help" {
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
	m.currentQuestion = question.Text
	m.agentStatus = "composing"
}

// startComposeModeWithPrompt starts composition with user's initial prompt
func (m *AppModel) startComposeModeWithPrompt(prompt string) {
	if m.generator == nil {
		m.agentStatus = "no key"
		return
	}
	m.composer = song.NewComposer(m.generator)
	m.composeMode = true
	
	// Start session with the user's initial prompt
	question := m.composer.StartSession(prompt)
	
	// Store the AI's first question
	m.currentQuestion = question.Text
	
	// Add AI's question to display
	_ = question
	
	m.agentStatus = "composing"
}

// processComposeInput handles user input in composition mode
func (m *AppModel) processComposeInput(input string) {
	// Check for command shortcuts
	inputLower := strings.ToLower(input)
	if inputLower == "/quit" {
		m.exitComposeMode()
		return
	}
	if inputLower == "/back" {
		if m.composer != nil {
			if err := m.composer.Rollback(); err != nil {
				m.errorMsg = err.Error()
			}
		}
		return
	}

	// Check if this is a structure selection (1, 2, or 3)
	phase := m.composer.GetSessionState().CurrentPhase
	if phase == song.PhaseStructure {
		if input == "1" || input == "2" || input == "3" {
			var idx int
			fmt.Sscanf(input, "%d", &idx)
			if err := m.composer.ProcessStructureSelection(idx - 1); err != nil {
				m.errorMsg = err.Error()
			}
		}
	}

	ctx := context.Background()
	_, question, err := m.composer.ProcessResponse(ctx, input)
	if err != nil {
		m.errorMsg = err.Error()
		return
	}

	// Store the AI's next question
	if question.Text != "" && question.Kind != "done" {
		m.currentQuestion = question.Text
	}

	// Check if we should generate the song
	if question.Kind == "done" || phase == song.PhaseGenerate {
		// Generate the DSL code for all sections
		m.agentStatus = "generating"
		if err := m.composer.GenerateAllSections(ctx); err != nil {
			m.errorMsg = err.Error()
			return
		}

		// Combine all section codes
		s := m.composer.GetSong()
		if s != nil && len(s.Sections) > 0 {
			var allCode strings.Builder
			for i, section := range s.Sections {
				if i > 0 {
					allCode.WriteString("\n\n")
				}
				allCode.WriteString(fmt.Sprintf("// === %s (%d bars) ===\n%s", 
					section.ID, section.Bars, section.DSLCode))
			}
			m.codeContent = allCode.String()
		}

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
	m.currentQuestion = ""
	m.agentStatus = "standby"
}

// renderComposeView renders the composition mode UI with multi-round conversation
func (m AppModel) renderComposeView() string {
	w := max(m.width, 40)
	h := max(m.height, 8)

	var b strings.Builder

	// Header with mode indicator
	b.WriteString(pink.Render("▓ SONG COMPOSER ▓"))
	b.WriteString(" " + cyan.Render("多轮头脑风暴"))
	b.WriteString("\n")
	b.WriteString(codeBorder.Render(strings.Repeat("─", w)))
	b.WriteString("\n\n")

	// Display conversation history in a scrollable area
	history := m.composer.GetDialogHistory()
	contentLines := h - 10
	if contentLines < 4 {
		contentLines = 4
	}

	// Build conversation display
	var conv strings.Builder
	
	// Show last few turns
	startIdx := 0
	if len(history) > contentLines/2 {
		startIdx = len(history) - contentLines/2
	}
	
	for i := startIdx; i < len(history); i++ {
		turn := history[i]
		if turn.Role == "user" {
			conv.WriteString(bright.Render("你: "))
			conv.WriteString(turn.Content)
		} else {
			// AI response - may contain question
			conv.WriteString(accent.Render("AI: "))
			conv.WriteString(turn.Content)
		}
		conv.WriteString("\n\n")
	}
	
	b.WriteString(conv.String())

	// Show current question prominently if in explore mode
	phase := m.composer.GetSessionState().CurrentPhase
	if phase == song.PhaseExplore || phase == song.PhaseStructure {
		if m.currentQuestion != "" {
			b.WriteString("\n")
			b.WriteString(cyan.Render("📋 ") + bright.Render("请回答:"))
			b.WriteString("\n")
			b.WriteString(accent.Render(m.currentQuestion))
			b.WriteString("\n")
		}
	}

	// Show progress info
	b.WriteString("\n")
	elements := m.composer.GetSessionState().Elements
	if elements.Genre != "" {
		b.WriteString(muted.Render(fmt.Sprintf("已确定: %s", elements.Genre)))
	}
	b.WriteString(muted.Render(fmt.Sprintf("  预计时长: %s", m.composer.GetDurationEstimate())))
	b.WriteString("\n\n")

	// Input area
	b.WriteString(inputBox.Render(m.editor.View()))
	b.WriteString("\n")

	// Status bar with help
	b.WriteString(m.renderComposeStatusLine())

	return b.String()
}

// renderComposeStatusLine renders the status line for compose mode
func (m AppModel) renderComposeStatusLine() string {
	phase := m.composer.GetSessionState().CurrentPhase
	var phaseName string
	switch phase {
	case song.PhaseExplore:
		phaseName = "探索要素"
	case song.PhaseStructure:
		phaseName = "选择结构"
	case song.PhaseGenerate:
		phaseName = "生成中"
	default:
		phaseName = "创作中"
	}
	
	left := accent.Render(phaseName)
	right := muted.Render("/quit 退出  ·  /back 回退")
	pad := max(0, m.width-len(left)-len(right))
	return left + strings.Repeat(" ", pad) + right
}