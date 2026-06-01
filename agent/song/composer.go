package song

import (
	"context"
	"fmt"
	"strings"

	"github.com/lyssom/vibe-music/agent/generator"
	"github.com/lyssom/vibe-music/pkg/logger"
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
func (c *Composer) ProcessResponse(ctx context.Context, response string) (*DialogTurn, Question, error) {
	logger.Debug("[Composer] ProcessResponse: response=%q", response)
	
	c.session.AddDialogTurn("user", response)

	// Get the current question BEFORE parsing
	historyLen := len(c.session.GetState().History) - 1 // -1 because we just added user turn
	currentQuestion := c.explorer.NextQuestion(historyLen)
	
	// Parse response and update elements using the current question
	c.parseAndSetElement(response, currentQuestion)

	elements := c.session.GetElements()
	logger.Debug("[Composer] After parseAndSetElement: Genre=%q, Emotion=%q, Rhythm=%q", 
		elements.Genre, elements.Emotion, elements.Rhythm)

	// Save checkpoint
	c.session.SaveHistoryNode()

	phase := c.session.GetState().CurrentPhase
	logger.Debug("[Composer] Current phase: %v", phase)

	switch phase {
	case PhaseExplore:
		if c.explorer.IsComplete() {
			logger.Debug("[Composer] Exploration complete, transitioning to PhaseStructure")
			c.session.SetPhase(PhaseStructure)
			return c.handleStructurePhase()
		}
		// Get NEXT question using updated history length
		nextQuestion := c.explorer.NextQuestion(len(c.session.GetState().History))
		logger.Debug("[Composer] Next question: kind=%q", nextQuestion.Kind)
		return c.buildAssistantTurn(nextQuestion), nextQuestion, nil

	case PhaseStructure:
		return c.handleStructurePhase()

	case PhaseGenerate, PhaseRefine:
		return c.handleGeneratePhase(ctx)

	case PhaseComplete:
		return nil, Question{Kind: "done", Text: "创作完成！"}, nil
	}

	return nil, Question{Kind: "error", Text: "Unknown phase"}, nil
}

// ProcessStructureSelection handles user's structure choice
func (c *Composer) ProcessStructureSelection(index int) error {
	err := c.structurer.SelectStructure(index)
	if err != nil {
		return err
	}
	// Transition to generate phase
	c.session.SetPhase(PhaseGenerate)
	
	// Build the song from the selected structure
	elements := c.session.GetElements()
	bpm := elements.BPM
	if bpm == 0 {
		bpm = 120
	}
	
	proposals := c.session.GetProposedStructures()
	var proposal StructureProposal
	for _, p := range proposals {
		if len(p.Sections) == len(c.session.GetSelectedStructure()) {
			proposal = p
			break
		}
	}
	if proposal.Name == "" && len(proposals) > 0 {
		proposal = proposals[0]
	}
	
	song := c.structurer.BuildSongFromStructure("Untitled Song", proposal, bpm)
	c.session.SetSong(song)
	
	return nil
}

// parseAndSetElement extracts and sets element from response
func (c *Composer) parseAndSetElement(response string, currentQuestion Question) {
	responseLower := strings.ToLower(response)
	kind := currentQuestion.Kind

	if kind == "done" {
		return
	}

	// Try to match options first
	matched := false
	for _, opt := range currentQuestion.Options {
		optLower := strings.ToLower(opt)

		// Strategy 1: exact match after lowercase
		if responseLower == optLower {
			c.session.SetElement(kind, opt)
			matched = true
			break
		}

		// Strategy 2: extract Chinese part for "爵士 Jazz" format
		if strings.Contains(optLower, " ") {
			chinesePart := strings.SplitN(optLower, " ", 2)[0]
			if responseLower == chinesePart || strings.Contains(responseLower, chinesePart) {
				c.session.SetElement(kind, opt)
				matched = true
				break
			}
		}

		// Strategy 3: option contains response
		if strings.Contains(optLower, responseLower) && responseLower != "" {
			c.session.SetElement(kind, opt)
			matched = true
			break
		}
	}

	// Fallback: use response as direct value
	if !matched {
		c.session.SetElement(kind, response)
	}
}

// handleStructurePhase handles structure negotiation
func (c *Composer) handleStructurePhase() (*DialogTurn, Question, error) {
	logger.Info("[Composer] handleStructurePhase called")
	
	proposals := c.structurer.ProposeStructures(3)
	c.session.SetProposedStructures(proposals)

	var sb strings.Builder
	sb.WriteString("根据你的音乐偏好，我推荐以下歌曲结构：\n\n")

	for i, p := range proposals {
		sb.WriteString(fmt.Sprintf("%d. %s (%d 小节)\n", i+1, p.Name, p.TotalBars))
		for _, cfg := range p.Sections {
			sb.WriteString(fmt.Sprintf("   - %s: %d 小节\n", sectionTypeName(cfg.Type), cfg.Bars))
		}
		sb.WriteString("\n")
	}
	sb.WriteString("请选择结构编号（输入 1、2 或 3），或告诉我你的自定义结构。")

	question := Question{
		Kind:    "structure",
		Text:    sb.String(),
		Options: []string{"1", "2", "3"},
	}

	logger.Info("[Composer] handleStructurePhase returning structure question with %d proposals", len(proposals))
	return c.buildAssistantTurn(question), question, nil
}

// handleGeneratePhase handles song generation
func (c *Composer) handleGeneratePhase(ctx context.Context) (*DialogTurn, Question, error) {
	elements := c.session.GetElements()
	bpm := elements.BPM
	if bpm == 0 {
		bpm = 120
	}

	// Get selected structure or use first proposal
	structure := c.session.GetSelectedStructure()
	var proposal StructureProposal
	
	if len(structure) > 0 {
		// Build proposal from selected structure
		proposals := c.session.GetProposedStructures()
		for _, p := range proposals {
			if len(p.Sections) == len(structure) {
				proposal = p
				break
			}
		}
	}
	
	if proposal.Name == "" {
		proposals := c.session.GetProposedStructures()
		if len(proposals) > 0 {
			proposal = proposals[0]
		} else {
			proposal = defaultStructures[0]
		}
	}

	song := c.structurer.BuildSongFromStructure("Untitled Song", proposal, bpm)
	c.session.SetSong(song)

	// Generate description
	desc := c.generateSongDescription(song, elements)
	song.Description = desc

	// Return early - actual DSL generation happens in GenerateAllSections
	return c.buildAssistantTurn(Question{
		Kind: "generating",
		Text: fmt.Sprintf("正在生成歌曲「%s」，共 %d 个段落...", song.Title, len(song.Sections)),
	}), Question{Kind: "generating", Text: "歌曲结构已创建。"}, nil
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

	// Check if the internal generator is available
	// c.gen is *SongGenerator, c.gen.gen is generator.Generator
	hasGenerator := c.gen != nil && c.gen.gen != nil
	
	if !hasGenerator {
		// No LLM generator available, use template-based generation
		fmt.Println("[Composer] No LLM generator, using template fallback")
		for i := range song.Sections {
			song.Sections[i].DSLCode = generateTemplateDSL(song.Sections[i])
		}
	} else {
		for i := range song.Sections {
			code, err := c.gen.GenerateSection(ctx, song.Sections[i], c.session.GetElements(), song)
			if err != nil {
				return err
			}
			song.Sections[i].DSLCode = code
		}
	}

	c.session.SetSong(song)
	c.session.SetPhase(PhaseComplete)
	c.session.SaveHistoryNode()
	return nil
}

// generateTemplateDSL generates DSL code from templates (fallback when no LLM)
func generateTemplateDSL(section Section) string {
	// Simple template-based DSL generation using valid DSL syntax
	switch section.Type {
	case SectionIntro:
		return fmt.Sprintf(`// Intro - %d bars at %d BPM
bass("c2", "2n", 0.7)
chord("c3 e3 g3", "4n", 0.4).every(4)`, section.Bars, section.BPM)
	case SectionVerse:
		return fmt.Sprintf(`// Verse - %d bars
bass("c2", "2n", 0.8)
chord("c3 e3 g3", "4n", 0.5).every(4)
sound("hh").every(4).every(2)`, section.Bars)
	case SectionPreChorus:
		return fmt.Sprintf(`// Pre-Chorus - %d bars
bass("c2", "2n", 0.85)
chord("c3 e3 g3", "4n", 0.55).every(4)
sound("hh").every(4).every(2)`, section.Bars)
	case SectionChorus:
		return fmt.Sprintf(`// Chorus - %d bars
bass("c2", "2n", 0.9)
chord("c3 e3 g3", "4n", 0.6).every(4)
sound("bd sd").every(4).every(2)
sound("hh").every(4).every(2)`, section.Bars)
	case SectionBridge:
		return fmt.Sprintf(`// Bridge - %d bars
bass("eb2", "2n", 0.85)
chord("eb3 gb3 bb3", "4n", 0.5).every(4)
sound("hh").every(4).every(3)`, section.Bars)
	case SectionOutro:
		return fmt.Sprintf(`// Outro - %d bars
bass("c2", "2n", 0.7)
chord("c3 e3 g3", "4n", 0.4).every(4)`, section.Bars)
	default:
		return fmt.Sprintf(`// Section - %d bars
bass("c2", "2n", 0.8)
chord("c3 e3 g3", "4n", 0.5).every(4)`, section.Bars)
	}
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

	// Generate modified code using the song generator
	newCode, err := c.gen.GenerateSection(ctx, *section, c.session.GetElements(), song)
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
