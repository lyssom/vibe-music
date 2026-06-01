package song

import "time"

// SongElements represents the ten musical elements
type SongElements struct {
	Genre           string // 风格/流派
	Emotion         string // 情感/情绪
	Rhythm          string // 节奏感
	Instrumentation string // 配器/音色
	Scale           string // 调式/音阶
	Harmony         string // 和声走向
	BPM             int    // 速度
	Dynamic         string // 动态范围
	Structure       string // 歌曲结构
	Techniques      string // 特殊技巧
}

// SectionType represents the type of a song section
type SectionType int

const (
	SectionIntro SectionType = iota
	SectionVerse
	SectionPreChorus
	SectionChorus
	SectionBridge
	SectionOutro
)

// Section represents a song section
type Section struct {
	ID          string
	Type        SectionType
	Name        string // Chinese display name (e.g., "前奏", "主歌")
	Bars        int
	BPM         int
	DSLCode     string
	Description string
	Elements    SongElements
}

// Song represents a complete song
type Song struct {
	Title             string
	Sections          []Section
	Description       string
	TotalBars         int
	EstimatedDuration time.Duration
}

// DialogTurn represents a conversation turn
type DialogTurn struct {
	Role    string
	Content string
	Turn    int
}

// HistoryNode represents a checkpoint for rollback
type HistoryNode struct {
	Timestamp   time.Time
	Phase       SessionPhase
	Elements    SongElements
	Structure   []SectionType
	Song        *Song
	Description string
}

// InteractionMode determines the exploration depth
type InteractionMode int

const (
	ModeSimple InteractionMode = iota
	ModeStandard
	ModeFull
)

// SessionPhase represents the current phase
type SessionPhase int

const (
	PhaseIntent SessionPhase = iota
	PhaseExplore
	PhaseStructure
	PhaseGenerate
	PhaseRefine
	PhaseComplete
)

// Question represents an AI question to the user
type Question struct {
	Kind    string // element kind: "genre", "emotion", etc.
	Text    string
	Options []string // optional multiple choice
}

// StructureProposal represents a proposed song structure
type StructureProposal struct {
	Name      string
	Sections  []SectionConfig
	TotalBars int
}

// SectionConfig defines a section in a structure template
type SectionConfig struct {
	Type        SectionType
	Bars        int
	Description string
}

// SessionState holds the session state
type SessionState struct {
	Mode                InteractionMode
	Elements            SongElements
	CurrentPhase        SessionPhase
	History             []DialogTurn
	HistoryNodes        []HistoryNode
	CurrentNode         int
	ProposedStructures  []StructureProposal
	SelectedStructure   []SectionType
	Song                *Song
}