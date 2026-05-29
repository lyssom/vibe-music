package generator

import (
	"context"
	"fmt"
	"strings"

	"github.com/lyssom/vibe-music/agent/llm"
)

const systemPrompt = `You are a music pattern generator for Vibe Music. You write code in a pattern DSL.

## DSL Syntax

### Drum sounds
  sound("bd sd hh")        — trigger built-in drum samples
                              bd=kick, sd=snare, hh=closed hihat, oh=open hihat, lt=low tom, ht=high tom
  sound("bd").every(2)     — trigger every 2nd beat
  sound("bd").every(3)     — trigger every 3rd beat

### Pitched instruments
  note("c3 e3 g3", "4n", 0.7)   — piano/lead notes (pitch, duration, velocity)
                                  duration: "32n" "16n" "8n" "4n" "2n" "1n"
  chord("c3 e3 g3", "4n", 0.5)  — same as note but for simultaneous notes
  bass("c2", "2n", 0.9)         — single bass note (lower register)

### Timing modifiers
  sound("bd sd").every(N)   — fire every N beats (1=every beat, 2=every 2 beats, etc.)
  .slow(N)                  — slow down: N× longer notes (every N beats)
  .fast(N)                 — speed up: divide into N sub-beats
  .swing(N)                — swing feel (0-100, 0=none, 100=full swing)
  .iter(N, "pattern")      — repeat pattern N times

## Examples
User: "jazz swing beat"
You:
sound("bd").every(4)
sound("sd").every(4).every(2)
sound("hh").every(4)

User: "Cmaj piano chords with bass"
You:
chord("c3 e3 g3", "2n", 0.6)
bass("c2", "2n", 0.8)

User: "slow bossa nova"
You:
sound("bd").every(4)
sound("sd").every(4).every(2)
chord("c3 e3 g3", "4n", 0.4).every(2)
bass("c2", "2n", 0.7)

User: "latin groove"
You:
sound("lt").every(4)
sound("hh").every(4)
sound("bd").every(4).every(3)
sound("sd").every(4).every(2)

## Rules
- Always output ONLY valid DSL code, no explanations, no markdown fences.
- For 16th-note swing feel, use .swing(60) on a sound command.
- Keep patterns concise. For long patterns, use multiple sound() commands with .every().
- You may chain: sound("bd sd").every(2) — fires every 2nd beat.
- note/chord/bass take standard note notation: c3, eb4, #f5, gb5, a3, etc.`

// NewLLMGenerator creates a generator backed by an LLM client.
func NewLLMGenerator(client llm.Client) *LLMGenerator {
	return &LLMGenerator{client: client}
}

// NewLLMGeneratorNoSystem creates a generator that converts system messages to user messages.
func NewLLMGeneratorNoSystem(client llm.Client) *LLMGenerator {
	return &LLMGenerator{client: client, noSystem: true}
}

// LLMGenerator implements Generator using an LLM client.
type LLMGenerator struct {
	client   llm.Client
	history  []llm.Message
	noSystem bool
}

// Generate creates pattern code from a user prompt.
func (g *LLMGenerator) Generate(ctx context.Context, prompt string, pctx PromptContext) (string, error) {
	messages := g.buildMessages(prompt, pctx)
	result, err := g.client.Chat(ctx, messages)
	if err != nil {
		return "", fmt.Errorf("generate: %w", err)
	}

	result = StripThinking(result)

	g.history = append(g.history,
		llm.Message{Role: "user", Content: prompt},
		llm.Message{Role: "assistant", Content: result},
	)

	return result, nil
}

// GenerateStream streams generated pattern code token by token.
func (g *LLMGenerator) GenerateStream(ctx context.Context, prompt string, pctx PromptContext) (<-chan llm.StreamEvent, error) {
	messages := g.buildMessages(prompt, pctx)

	ch, err := g.client.ChatStream(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("generate stream: %w", err)
	}

	out := make(chan llm.StreamEvent, 10)
	go func() {
		defer close(out)
		var full strings.Builder
		for ev := range ch {
			full.WriteString(ev.Content)
			out <- ev
		}
		g.history = append(g.history,
			llm.Message{Role: "user", Content: prompt},
			llm.Message{Role: "assistant", Content: StripThinking(full.String())},
		)
	}()

	return out, nil
}

// buildMessages constructs the LLM message list.
func (g *LLMGenerator) buildMessages(prompt string, pctx PromptContext) []llm.Message {
	role := func(r string) string {
		if g.noSystem && r == "system" {
			return "user"
		}
		return r
	}

	messages := []llm.Message{
		{Role: role("system"), Content: systemPrompt},
	}

	start := 0
	if len(g.history) > 10 {
		start = len(g.history) - 10
	}
	messages = append(messages, g.history[start:]...)

	if pctx.CurrentCode != "" {
		codeCtx := fmt.Sprintf("Current code in the editor:\n```\n%s\n```", pctx.CurrentCode)
		messages = append(messages, llm.Message{Role: role("system"), Content: codeCtx})
	}

	if pctx.ErrorMessage != "" {
		errCtx := fmt.Sprintf("The last execution produced this error: %s\nPlease fix the code.", pctx.ErrorMessage)
		messages = append(messages, llm.Message{Role: role("system"), Content: errCtx})
	}

	messages = append(messages, llm.Message{Role: "user", Content: prompt})
	return messages
}

// StripThinking removes  thinking blocks from thinking model responses.
func StripThinking(text string) string {
	// Strategy: find the LAST DSL function call in the response.
	for _, marker := range []string{"sound(", "note(", "chord(", "bass(", "stop(", "slow(", "fast(", "every(", "iter("} {
		if idx := strings.LastIndex(text, marker); idx != -1 {
			return strings.TrimSpace(text[idx:])
		}
	}
	// Fallback: strip after last " response" marker
	if idx := strings.LastIndex(text, " response"); idx != -1 {
		result := strings.TrimSpace(text[idx+len(" response"):])
		if result != "" {
			return result
		}
	}
	return strings.TrimSpace(text)
}

// ClearHistory resets the conversation history.
func (g *LLMGenerator) ClearHistory() {
	g.history = nil
}