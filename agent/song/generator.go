package song

import (
	"context"
	"fmt"
	"strings"

	"github.com/lyssom/vibe-music/agent/generator"
	"github.com/lyssom/vibe-music/agent/llm"
)

// SongGenerator generates DSL code for song sections
type SongGenerator struct {
	gen generator.Generator
}

// NewSongGenerator creates a new song generator
func NewSongGenerator(gen generator.Generator) *SongGenerator {
	return &SongGenerator{gen: gen}
}

// GenerateSection generates DSL code for a specific section
func (g *SongGenerator) GenerateSection(ctx context.Context, section Section, elements SongElements, song *Song) (string, error) {
	prompt := g.buildSectionPrompt(section, elements)

	pctx := generator.PromptContext{
		CurrentCode: g.getRelevantContext(section, song),
	}

	code, err := g.gen.Generate(ctx, prompt, pctx)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(code), nil
}

// buildSectionPrompt creates a prompt for section generation
func (g *SongGenerator) buildSectionPrompt(section Section, elements SongElements) string {
	sectionDesc := g.getSectionDescription(section)

	prompt := fmt.Sprintf(`Generate DSL code for a song section with these characteristics:

## Section: %s
%s

## Musical Elements:
- Genre: %s
- Emotion: %s
- Rhythm: %s
- Instrumentation: %s
- Scale: %s
- Harmony: %s
- BPM: %d
- Dynamic: %s
- Techniques: %s

## Requirements:
- Generate valid DSL code for this section
- Duration: %d bars at %d BPM
- Match the emotional character: %s
- Return ONLY the DSL code, no explanations

`, section.ID, sectionDesc, elements.Genre, elements.Emotion, elements.Rhythm,
		elements.Instrumentation, elements.Scale, elements.Harmony, elements.BPM,
		elements.Dynamic, elements.Techniques, section.Bars, section.BPM, elements.Emotion)

	return prompt
}

// getSectionDescription returns human-readable section description
func (g *SongGenerator) getSectionDescription(section Section) string {
	switch section.Type {
	case SectionIntro:
		return "Intro - Sets the mood and introduces the musical theme"
	case SectionVerse:
		return "Verse - Tells the story, usually moderate intensity"
	case SectionPreChorus:
		return "Pre-Chorus - Builds tension before the chorus"
	case SectionChorus:
		return "Chorus - The main hook, highest energy and most memorable"
	case SectionBridge:
		return "Bridge - Provides contrast and variation"
	case SectionOutro:
		return "Outro - Brings the song to a close"
	default:
		return "Song section"
	}
}

// getRelevantContext returns context from previous sections
func (g *SongGenerator) getRelevantContext(section Section, song *Song) string {
	if song == nil {
		return ""
	}
	
	var sb strings.Builder
	foundCurrent := false
	
	for _, s := range song.Sections {
		if s.ID == section.ID {
			foundCurrent = true
			continue
		}
		if !foundCurrent && s.DSLCode != "" {
			// Only include previous sections
			if sb.Len() > 0 {
				sb.WriteString("\n\n")
			}
			sb.WriteString(fmt.Sprintf("// %s:\n%s", s.ID, s.DSLCode))
		}
	}
	
	return sb.String()
}

// GenerateSectionStream streams DSL code for a section
func (g *SongGenerator) GenerateSectionStream(ctx context.Context, section Section, elements SongElements, song *Song) (<-chan llm.StreamEvent, error) {
	prompt := g.buildSectionPrompt(section, elements)

	pctx := generator.PromptContext{
		CurrentCode: g.getRelevantContext(section, song),
	}

	return g.gen.GenerateStream(ctx, prompt, pctx)
}