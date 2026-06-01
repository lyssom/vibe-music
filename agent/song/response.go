package song

// Action types for LLM responses
const (
	ActionQuestion = "question"
	ActionGenerate = "generate"
	ActionDone     = "done"
	ActionRefine   = "refine"
)

// StructuredResponse represents LLM's structured response
type StructuredResponse struct {
	Type    string   `json:"type"`              // "question" | "generate" | "done" | "refine"
	Action  string   `json:"action"`            // Same as Type for compatibility
	Message string   `json:"message"`           // Human-readable message
	Options []string `json:"options,omitempty"` // Question options (action=question)

	// Generate parameters (action=generate or refine)
	Structure []SectionSpec `json:"structure,omitempty"`
	BPM       int           `json:"bpm,omitempty"`
	Notes     string        `json:"notes,omitempty"`
}

// SectionSpec defines a section for generation
type SectionSpec struct {
	ID    string `json:"id"`    // "intro", "verse", "chorus", etc.
	Name  string `json:"name"`  // Chinese name: "前奏", "主歌", etc.
	Bars  int    `json:"bars"`  // Number of bars
}