package song

import (
	"fmt"
	"strings"
	"time"
)

// Structurer handles song structure proposals
type Structurer struct {
	session *Session
}

// NewStructurer creates a new structurer
func NewStructurer(session *Session) *Structurer {
	return &Structurer{session: session}
}

// structureTemplates maps genre to structure templates
var structureTemplates = map[string][]StructureProposal{
	"jazz": {
		{
			Name: "经典 AABA (Jazz Standard)",
			Sections: []SectionConfig{
				{Type: SectionIntro, Bars: 4, Description: "引入主题"},
				{Type: SectionVerse, Bars: 8, Description: "A 段 - 主题陈述"},
				{Type: SectionBridge, Bars: 8, Description: "B 段 - 即兴/变奏"},
				{Type: SectionVerse, Bars: 8, Description: "A 段回归"},
				{Type: SectionOutro, Bars: 4, Description: "渐弱结束"},
			},
			TotalBars: 32,
		},
		{
			Name: "扩展 AABA",
			Sections: []SectionConfig{
				{Type: SectionIntro, Bars: 4},
				{Type: SectionVerse, Bars: 8},
				{Type: SectionBridge, Bars: 8},
				{Type: SectionVerse, Bars: 8},
				{Type: SectionChorus, Bars: 8},
				{Type: SectionOutro, Bars: 4},
			},
			TotalBars: 40,
		},
	},
	"pop": {
		{
			Name: "经典主副结构",
			Sections: []SectionConfig{
				{Type: SectionIntro, Bars: 4, Description: "主歌 1"},
				{Type: SectionVerse, Bars: 8, Description: "主歌 1"},
				{Type: SectionPreChorus, Bars: 4, Description: "预副歌"},
				{Type: SectionChorus, Bars: 8, Description: "副歌 1"},
				{Type: SectionVerse, Bars: 8, Description: "主歌 2"},
				{Type: SectionChorus, Bars: 8, Description: "副歌 2"},
				{Type: SectionBridge, Bars: 8, Description: "桥段"},
				{Type: SectionChorus, Bars: 8, Description: "副歌 3"},
				{Type: SectionOutro, Bars: 4},
			},
			TotalBars: 60,
		},
		{
			Name: "简洁主副结构",
			Sections: []SectionConfig{
				{Type: SectionIntro, Bars: 4},
				{Type: SectionVerse, Bars: 8},
				{Type: SectionChorus, Bars: 8},
				{Type: SectionVerse, Bars: 8},
				{Type: SectionChorus, Bars: 8},
				{Type: SectionOutro, Bars: 4},
			},
			TotalBars: 40,
		},
	},
	"rock": {
		{
			Name: "摇滚标准结构",
			Sections: []SectionConfig{
				{Type: SectionIntro, Bars: 4},
				{Type: SectionVerse, Bars: 8},
				{Type: SectionChorus, Bars: 8},
				{Type: SectionVerse, Bars: 8},
				{Type: SectionChorus, Bars: 8},
				{Type: SectionBridge, Bars: 8},
				{Type: SectionChorus, Bars: 8},
				{Type: SectionOutro, Bars: 4},
			},
			TotalBars: 56,
		},
	},
	"electronic": {
		{
			Name: "EDM 结构",
			Sections: []SectionConfig{
				{Type: SectionIntro, Bars: 8},
				{Type: SectionVerse, Bars: 8},
				{Type: SectionChorus, Bars: 16},
				{Type: SectionBridge, Bars: 8},
				{Type: SectionChorus, Bars: 16},
				{Type: SectionOutro, Bars: 8},
			},
			TotalBars: 64,
		},
	},
	"folk": {
		{
			Name: "民谣叙事结构",
			Sections: []SectionConfig{
				{Type: SectionIntro, Bars: 4},
				{Type: SectionVerse, Bars: 8},
				{Type: SectionChorus, Bars: 8},
				{Type: SectionVerse, Bars: 8},
				{Type: SectionChorus, Bars: 8},
				{Type: SectionVerse, Bars: 8},
				{Type: SectionChorus, Bars: 8},
				{Type: SectionOutro, Bars: 4},
			},
			TotalBars: 56,
		},
	},
	"blues": {
		{
			Name: "12 小节蓝调",
			Sections: []SectionConfig{
				{Type: SectionIntro, Bars: 4},
				{Type: SectionVerse, Bars: 12, Description: "12 小节蓝调进行"},
				{Type: SectionVerse, Bars: 12},
				{Type: SectionBridge, Bars: 12, Description: "即兴 Solo"},
				{Type: SectionVerse, Bars: 12},
				{Type: SectionOutro, Bars: 4},
			},
			TotalBars: 56,
		},
	},
}

// defaultStructures for unknown genres
var defaultStructures = []StructureProposal{
	{
		Name: "通用结构",
		Sections: []SectionConfig{
			{Type: SectionIntro, Bars: 4},
			{Type: SectionVerse, Bars: 8},
			{Type: SectionChorus, Bars: 8},
			{Type: SectionBridge, Bars: 8},
			{Type: SectionChorus, Bars: 8},
			{Type: SectionOutro, Bars: 4},
		},
		TotalBars: 40,
	},
}

// ProposeStructures returns structure proposals based on genre
func (s *Structurer) ProposeStructures(count int) []StructureProposal {
	genre := s.session.GetElements().Genre
	genreLower := strings.ToLower(genre)

	// Find matching templates
	var templates []StructureProposal

	// Direct genre match
	if templates, ok := structureTemplates[genreLower]; ok {
		return takeFirst(templates, count)
	}

	// Partial match
	for key, vals := range structureTemplates {
		if strings.Contains(genreLower, key) || strings.Contains(key, genreLower) {
			templates = append(templates, vals...)
		}
	}

	if len(templates) > 0 {
		return takeFirst(templates, count)
	}

	// Fallback to default
	return takeFirst(defaultStructures, count)
}

// SelectStructure selects a structure by index
func (s *Structurer) SelectStructure(index int) error {
	proposals := s.session.GetProposedStructures()
	if index < 0 || index >= len(proposals) {
		return fmt.Errorf("invalid structure index: %d", index)
	}

	proposal := proposals[index]
	var structure []SectionType
	for _, cfg := range proposal.Sections {
		structure = append(structure, cfg.Type)
	}

	s.session.SetStructure(structure)
	s.session.SetPhase(PhaseGenerate)
	return nil
}

// BuildSongFromStructure creates Song sections from structure
func (s *Structurer) BuildSongFromStructure(title string, proposal StructureProposal, bpm int) *Song {
	song := &Song{
		Title:       title,
		TotalBars:   proposal.TotalBars,
		Description: "",
		Sections:    make([]Section, 0, len(proposal.Sections)),
	}

	// Estimate duration: 4 beats per bar, at given BPM
	song.EstimatedDuration = time.Duration(estimateDuration(proposal.TotalBars, bpm) * 1e9)

	// Create sections
	idCounters := map[SectionType]int{}
	for _, cfg := range proposal.Sections {
		idCounters[cfg.Type]++
		sectionTypeName := sectionTypeName(cfg.Type)
		id := sectionTypeName + formatInt(idCounters[cfg.Type])

		section := Section{
			ID:          id,
			Type:        cfg.Type,
			Bars:        cfg.Bars,
			BPM:         bpm,
			DSLCode:     "",
			Description: cfg.Description,
			Elements:    s.session.GetElements(),
		}
		song.Sections = append(song.Sections, section)
	}

	return song
}

// sectionTypeName returns readable name for section type
func sectionTypeName(st SectionType) string {
	switch st {
	case SectionIntro:
		return "intro"
	case SectionVerse:
		return "verse"
	case SectionPreChorus:
		return "prechorus"
	case SectionChorus:
		return "chorus"
	case SectionBridge:
		return "bridge"
	case SectionOutro:
		return "outro"
	default:
		return "section"
	}
}

func formatInt(n int) string {
	if n == 1 {
		return ""
	}
	return formatIntImpl(n)
}

func formatIntImpl(n int) string {
	if n == 0 {
		return ""
	}
	return formatIntImpl(n/10) + string(rune('0'+n%10))
}

func estimateDuration(bars, bpm int) int {
	beats := bars * 4
	seconds := float64(beats) * 60.0 / float64(bpm)
	return int(seconds)
}

func takeFirst(slice []StructureProposal, n int) []StructureProposal {
	if len(slice) < n {
		return slice
	}
	return slice[:n]
}