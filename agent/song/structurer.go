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

// SelectStructure selects a structure by index OR parses custom structure input
func (s *Structurer) SelectStructure(index int) error {
	proposals := s.session.GetProposedStructures()
	
	// If index >= 0, use the proposal at that index
	if index >= 0 && index < len(proposals) {
		proposal := proposals[index]
		var structure []SectionType
		for _, cfg := range proposal.Sections {
			structure = append(structure, cfg.Type)
		}
		s.session.SetStructure(structure)
		s.session.SetPhase(PhaseGenerate)
		return nil
	}
	
	// index == -1 means use the custom structure from elements
	structureStr := s.session.GetElements().Structure
	if structureStr == "" || structureStr == "由你推荐" {
		// Use the first proposal
		if len(proposals) > 0 {
			proposal := proposals[0]
			var structure []SectionType
			for _, cfg := range proposal.Sections {
				structure = append(structure, cfg.Type)
			}
			s.session.SetStructure(structure)
			s.session.SetPhase(PhaseGenerate)
			return nil
		}
	}
	
	// Parse custom structure like "主歌副歌主歌副歌桥副歌"
	structure, err := s.ParseCustomStructure(structureStr)
	if err != nil {
		return fmt.Errorf("invalid structure: %v", err)
	}
	
	s.session.SetStructure(structure)
	s.session.SetPhase(PhaseGenerate)
	return nil
}

// ParseCustomStructure parses a custom structure string like "主歌副歌主歌副歌桥副歌"
func (s *Structurer) ParseCustomStructure(input string) ([]SectionType, error) {
	if input == "" {
		return nil, fmt.Errorf("empty structure")
	}
	
	// Mapping from Chinese/English names to SectionType
	nameMap := map[string]SectionType{
		"intro":      SectionIntro,
		"引":         SectionIntro,
		"引入":       SectionIntro,
		"开场":       SectionIntro,
		"verse":      SectionVerse,
		"主歌":       SectionVerse,
		"a段":        SectionVerse,
		"pre":        SectionPreChorus,
		"prechorus":  SectionPreChorus,
		"预副歌":     SectionPreChorus,
		"预设":       SectionPreChorus,
		"chorus":     SectionChorus,
		"副歌":       SectionChorus,
		"b段":        SectionChorus,
		"高潮":        SectionChorus,
		"bridge":     SectionBridge,
		"桥":         SectionBridge,
		"桥段":       SectionBridge,
		"c段":        SectionBridge,
		"outro":      SectionOutro,
		"尾":         SectionOutro,
		"结尾":       SectionOutro,
		"结束":       SectionOutro,
	}
	
	var result []SectionType
	inputLower := strings.ToLower(input)
	
	// Try to find matches for each known section name
	for name, sectionType := range nameMap {
		if strings.Contains(inputLower, name) {
			result = append(result, sectionType)
		}
	}
	
	// If no matches found, try to count occurrences and guess structure
	if len(result) == 0 {
		// Count Chinese characters
		verseCount := strings.Count(inputLower, "主歌") + strings.Count(inputLower, "verse")
		chorusCount := strings.Count(inputLower, "副歌") + strings.Count(inputLower, "chorus")
		bridgeCount := strings.Count(inputLower, "桥") + strings.Count(inputLower, "bridge")
		
		for i := 0; i < verseCount && i < 2; i++ {
			result = append(result, SectionVerse)
		}
		for i := 0; i < chorusCount && i < 2; i++ {
			result = append(result, SectionChorus)
		}
		if bridgeCount > 0 {
			result = append(result, SectionBridge)
		}
	}
	
	if len(result) == 0 {
		return nil, fmt.Errorf("could not parse structure from: %s", input)
	}
	
	return result, nil
}

// BuildProposalFromStructure creates a StructureProposal from a parsed structure
func (s *Structurer) BuildProposalFromStructure(structure []SectionType, bpm int) StructureProposal {
	// Count section types
	sectionCounts := map[SectionType]int{}
	for _, st := range structure {
		sectionCounts[st]++
	}
	
	// Determine bars per section type
	barsMap := map[SectionType]int{
		SectionIntro: 4,
		SectionOutro: 4,
		SectionPreChorus: 4,
	}
	
	// Default bars based on section type
	defaultBars := 8
	for _, st := range structure {
		if st == SectionIntro || st == SectionOutro {
			defaultBars = 4
			break
		}
	}
	
	// Build section configs
	var sections []SectionConfig
	totalBars := 0
	idCounters := map[SectionType]int{}
	
	for _, st := range structure {
		idCounters[st]++
		bars := barsMap[st]
		if bars == 0 {
			bars = defaultBars
		}
		
		sections = append(sections, SectionConfig{
			Type: st,
			Bars: bars,
			Description: fmt.Sprintf("%s %d", sectionTypeName(st), idCounters[st]),
		})
		totalBars += bars
	}
	
	return StructureProposal{
		Name:      "自定义结构",
		Sections:  sections,
		TotalBars: totalBars,
	}
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