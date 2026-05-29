// Package llm defines interfaces and implementations for LLM providers.
package llm

import "context"

// Message represents a chat message in an LLM conversation.
type Message struct {
	Role    string // "system", "user", or "assistant"
	Content string
}

// StreamEvent is a chunk of a streaming response.
type StreamEvent struct {
	Content string // incremental text
	Done    bool   // true when the stream is complete
	Err     error  // any error that occurred
}

// Client abstracts an LLM provider.
type Client interface {
	// Chat sends messages and returns the complete response.
	Chat(ctx context.Context, messages []Message) (string, error)

	// ChatStream sends messages and streams the response through the returned channel.
	// The caller must read until StreamEvent.Done == true, then close is guaranteed.
	ChatStream(ctx context.Context, messages []Message) (<-chan StreamEvent, error)
}