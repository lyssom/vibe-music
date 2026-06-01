// Package generator defines the interface for AI-powered music code generation.
package generator

import (
	"context"

	"github.com/lyssom/vibe-music/agent/llm"
)

// PromptContext holds the information available to the generator when creating code.
type PromptContext struct {
	// CurrentCode is the existing pattern code the user has written.
	CurrentCode string
	// History contains recent conversation turns.
	History []llm.Message
	// ErrorMessage is the last execution error, if any.
	ErrorMessage string
}

// Generator produces music pattern code from natural language prompts.
type Generator interface {
	// Generate creates pattern code based on a user prompt and current context.
	Generate(ctx context.Context, prompt string, pctx PromptContext) (string, error)

	// GenerateStream streams generated pattern code token by token.
	GenerateStream(ctx context.Context, prompt string, pctx PromptContext) (<-chan llm.StreamEvent, error)

	// GenerateWithStructuredResponse returns structured response for conversation flow
	GenerateWithStructuredResponse(ctx context.Context, prompt string, history []llm.Message) (*llm.StructuredResponse, error)
}