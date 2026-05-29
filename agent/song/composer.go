package song

import (
	"context"
	"fmt"
	"strings"

	"github.com/lyssom/vibe-music/agent/generator"
)

// Composer orchestrates the entire composition flow
type Composer struct {
	session    *Session
	explorer   *Explorer
	structurer *Structurer
	gen        *SongGenerator
}

// NewComposer creates a new composer
func NewComposer(llmGen generator.Generator) *Composer {
	session := NewSession(ModeFull) // Default to full mode
	return &Composer{
		session:    session,
		explorer:   NewExplorer(session),
		structurer: NewStructurer(session),
		gen:        NewSongGenerator(llmGen),
	}
}

// StartSession initializes a new composition session
func (c *Composer) StartSession(initialPrompt string) Question {
	mode := DetermineMode(initialPrompt)
	c.session = NewSession(mode)
	c.explorer = NewExplorer(c.session)
	c.structurer = NewStructurer(c.session)

	c.session.AddDialogTurn("user", initialPrompt)

	if mode == ModeSimple {
		// Skip exploration, go directly to generation
		c.session.SetPhase(PhaseGenerate)
		return Question{Kind: "done", Text: "好的，让我为你生成一个节拍循环。"}
	}

	c.session.SetPhase(PhaseExplore)
	return c.explorer.NextQuestion(0)
}

// ProcessResponse handles user response to a question
func (c *Composer) ProcessResponse(response string) (*DialogTurn, Question, error) {
	c.session.AddDialogTurn("user", response)

	// Parse response and update elements
	c.parseAndSetElement(response)

	// Save checkpoint
	c.session.SaveHistoryNode()

	phase := c.session.GetState().CurrentPhase

	switch phase {
	case PhaseExplore:
		if c.explorer.IsComplete() {
			c.session.SetPhase(PhaseStructure)
			return c.handleStructurePhase()
		}
		turn := len(c.session.GetState().History)
		question := c.explorer.NextQuestion(turn)
		return c.buildAssistantTurn(question), question, nil

	case PhaseStructure:
		return c.handleStructurePhase()

	case PhaseGenerate, PhaseRefine:
		return c.handleGeneratePhase()

	case PhaseComplete:
		return nil, Question{Kind: "done", Text: "创作完成！"}, nil
	}

	return nil, Question{Kind: "error", Text: "Unknown phase"}, nil
}

// parseAndSetElement extracts and sets element from response
func (c *Composer) parseAndSetElement(response string) {
	response = strings.ToLower(response)
	state := c.session.GetState()
	question := c.explorer.NextQuestion(len(state.History))
	kind := question.Kind

	if kind == "done" {
		return
	}

	// Try to match options first
	for _, opt := range question.Options {
		if strings.Contains(strings.ToLower(response), strings.ToLower(opt)) {
			c.session.SetElement(kind, opt)
			return
		}
	}

	// Use response as direct value
	c.session.SetElement(kind, response)
}

// handleStructurePhase handles structure negotiation
func (c *Composer) handleStructurePhase() (*DialogTurn, Question, error) {
	proposals := c.structurer.ProposeStructures(3)
	c.session.state.ProposedStructures = proposals

	var sb strings.Builder
	sb.WriteString("根据你的音乐偏好，我推荐以下歌曲结构：\n\n")

	for i, p := range proposals {
		sb.WriteString(fmt.Sprintf("%d. %s (%d 小节)\n", i+1, p.Name, p.TotalBars))
		for _, cfg := range p.Sections {
			sb.WriteString(fmt.Sprintf("   - %s: %d 小节\n", sectionTypeName(cfg.Type), cfg.Bars))
		}
		sb.WriteString("\n")
	}
	sb.WriteString("请选择结构编号，或告诉我你的自定义结构。")

	question := Question{
		Kind:    "structure",
		Text:    sb.String(),
		Options: []string{"1", "2", "3"},
	}

	return c.buildAssistantTurn(question), question, nil
}

// handleGeneratePhase handles song generation
func (c *Composer) handleGeneratePhase() (*DialogTurn, Question, error) {
	elements := c.session.GetElements()
	bpm := elements.BPM
	if bpm == 0 {
		bpm = 120
	}

	// Get selected structure proposal or use first
	proposals := c.session.state.ProposedStructures
	var proposal StructureProposal
	if len(proposals) > 0 {
		proposal = proposals[0]
	} else {
		proposal = defaultStructures[0]
	}

	song := c.structurer.BuildSongFromStructure("Untitled Song", proposal, bpm)
	c.session.SetSong(song)

	// Generate description
	desc := c.generateSongDescription(song, elements)
	song.Description = desc

	return c.buildAssistantTurn(Question{
		Kind:    "generating",
		Text:    fmt.Sprintf("正在生成歌曲「%s」...", song.Title),
	}), Question{Kind: "done", Text: "歌曲生成完成！"}, nil
}

// generateSongDescription creates a text description of the song
func (c *Composer) generateSongDescription(song *Song, elements SongElements) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("这是一首%s", elements.Genre))
	if elements.Emotion != "" {
		sb.WriteString(fmt.Sprintf("，情感上%s", elements.Emotion))
	}
	if elements.Rhythm != "" {
		sb.WriteString(fmt.Sprintf("，节奏%s", elements.Rhythm))
	}
	if elements.Instrumentation != "" {
		sb.WriteString(fmt.Sprintf("，采用%s配置", elements.Instrumentation))
	}
	sb.WriteString("。")
	return sb.String()
}

// buildAssistantTurn creates an assistant dialog turn
func (c *Composer) buildAssistantTurn(question Question) *DialogTurn {
	turn := len(c.session.GetState().History)
	return &DialogTurn{
		Role:    "assistant",
		Content: question.Text,
		Turn:    turn,
	}
}

// GenerateAllSections generates DSL code for all sections
func (c *Composer) GenerateAllSections(ctx context.Context) error {
	song := c.session.GetSong()
	if song == nil {
		return fmt.Errorf("no song to generate")
	}

	for i := range song.Sections {
		code, err := c.gen.GenerateSection(ctx, song.Sections[i], c.session.GetElements())
		if err != nil {
			return err
		}
		song.Sections[i].DSLCode = code
	}

	c.session.SetSong(song)
	c.session.SetPhase(PhaseComplete)
	c.session.SaveHistoryNode()
	return nil
}

// RefineSection modifies a specific section
func (c *Composer) RefineSection(ctx context.Context, sectionID string, instruction string) error {
	song := c.session.GetSong()
	if song == nil {
		return fmt.Errorf("no song to refine")
	}

	var section *Section
	for i := range song.Sections {
		if song.Sections[i].ID == sectionID {
			section = &song.Sections[i]
			break
		}
	}

	if section == nil {
		return fmt.Errorf("section not found: %s", sectionID)
	}

	prompt := fmt.Sprintf(`Modify the following DSL code based on this instruction: %s

Original code:
%s

Requirements:
- Return ONLY the modified DSL code
- Keep the same section type and duration`, instruction, section.DSLCode)

	newCode, err := c.gen.gen.Generate(ctx, prompt, generator.PromptContext{})
	if err != nil {
		return err
	}

	section.DSLCode = strings.TrimSpace(newCode)
	c.session.SaveHistoryNode()
	return nil
}

// GetSong returns the current song
func (c *Composer) GetSong() *Song {
	return c.session.GetSong()
}

// GetSessionState returns the current session state
func (c *Composer) GetSessionState() SessionState {
	return c.session.GetState()
}

// Rollback performs a rollback to the previous state
func (c *Composer) Rollback() error {
	return c.session.RollbackLast()
}

// GetDialogHistory returns the conversation history
func (c *Composer) GetDialogHistory() []DialogTurn {
	return c.session.GetState().History
}

// GetDurationEstimate returns estimated song duration
func (c *Composer) GetDurationEstimate() string {
	song := c.session.GetSong()
	if song == nil {
		return "Unknown"
	}
	dur := song.EstimatedDuration
	// Convert time.Duration to minutes and seconds
	mins := int(dur.Minutes())
	secs := int(dur.Seconds()) % 60
	return fmt.Sprintf("%d:%02d", mins, secs)
}

// FormatSongOutput formats the song for display
func (c *Composer) FormatSongOutput() string {
	song := c.session.GetSong()
	if song == nil {
		return "No song generated yet."
	}

	var sb strings.Builder
	sb.WriteString("═══════════════════════════════════════════════════════════\n")
	sb.WriteString(fmt.Sprintf("🎵 Song: %s\n", song.Title))
	sb.WriteString("═══════════════════════════════════════════════════════════\n\n")

	sb.WriteString("📝 整体描述:\n")
	sb.WriteString(song.Description)
	sb.WriteString("\n\n")

	for _, section := range song.Sections {
		sb.WriteString("───────────────────────────────────────────────────────────\n")
		sb.WriteString(fmt.Sprintf("[%s] %d小节 | %s\n",
			strings.ToUpper(section.ID), section.Bars, section.Description))
		sb.WriteString("───────────────────────────────────────────────────────────\n")
		sb.WriteString(section.DSLCode)
		sb.WriteString("\n\n")
	}

	sb.WriteString("═══════════════════════════════════════════════════════════\n")
	sb.WriteString(fmt.Sprintf("总时长: %s | %d 小节 | BPM: %d\n",
		c.GetDurationEstimate(), song.TotalBars, song.Sections[0].BPM))
	sb.WriteString("═══════════════════════════════════════════════════════════\n")

	return sb.String()
}
