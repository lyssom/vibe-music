package song

import "github.com/lyssom/vibe-music/agent/llm"

// Deprecated: Use llm.StructuredResponse instead. This alias is kept for backwards compatibility.
type StructuredResponse = llm.StructuredResponse

// Deprecated: Use llm.SectionSpec instead. This alias is kept for backwards compatibility.
type SectionSpec = llm.SectionSpec

// Deprecated: Use llm.ActionQuestion instead.
const ActionQuestion = llm.ActionQuestion

// Deprecated: Use llm.ActionGenerate instead.
const ActionGenerate = llm.ActionGenerate

// Deprecated: Use llm.ActionDone instead.
const ActionDone = llm.ActionDone

// Deprecated: Use llm.ActionRefine instead.
const ActionRefine = llm.ActionRefine