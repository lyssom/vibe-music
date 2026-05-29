package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/lyssom/vibe-music/pkg/logger"
)

// OpenAIClient implements Client using the OpenAI-compatible API.
type OpenAIClient struct {
	baseURL string
	model   string
	apiKey  string
	client  *http.Client
	log     *logger.Logger
}

// Option configures an OpenAIClient.
type Option func(*OpenAIClient)

// WithBaseURL sets the base URL for the API.
func WithBaseURL(url string) Option {
	return func(c *OpenAIClient) {
		c.baseURL = url
	}
}

// WithModel sets the model name.
func WithModel(model string) Option {
	return func(c *OpenAIClient) {
		c.model = model
	}
}

// NewOpenAIClient creates a new OpenAI-compatible client.
func NewOpenAIClient(apiKey string, opts ...Option) *OpenAIClient {
	c := &OpenAIClient{
		baseURL: "https://api.openai.com/v1",
		model:   "gpt-4",
		apiKey:  apiKey,
		log:     logger.New("llm", logger.DEBUG),
	}

	for _, opt := range opts {
		opt(c)
	}

	c.client = &http.Client{
		Timeout: 60 * time.Second,
	}

	c.log.Info("OpenAIClient created: baseURL=%s model=%s", c.baseURL, c.model)
	return c
}

// Chat sends a non-streaming chat request.
func (c *OpenAIClient) Chat(ctx context.Context, messages []Message) (string, error) {
	c.log.Debug("Chat: %d messages", len(messages))

	payload := map[string]interface{}{
		"model":    c.model,
		"messages": messages,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	c.log.Debug("Request: POST %s", c.baseURL+"/chat/completions")

	resp, err := c.client.Do(req)
	if err != nil {
		c.log.Error("Request failed: %v", err)
		return "", fmt.Errorf("request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		c.log.Error("Non-200 response: %d %s", resp.StatusCode, string(respBody))
		return "", fmt.Errorf("api error: status %d", resp.StatusCode)
	}

	var result struct {
		Choices []struct {
			Message Message `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode: %w", err)
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	content := result.Choices[0].Message.Content
	c.log.Debug("Response: %d chars", len(content))

	return content, nil
}

// ChatStream sends a streaming chat request.
func (c *OpenAIClient) ChatStream(ctx context.Context, messages []Message) (<-chan StreamEvent, error) {
	c.log.Debug("ChatStream: %d messages", len(messages))

	payload := map[string]interface{}{
		"model":    c.model,
		"messages": messages,
		"stream":   true,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	c.log.Debug("Stream request: POST %s", c.baseURL+"/chat/completions")

	resp, err := c.client.Do(req)
	if err != nil {
		c.log.Error("Stream request failed: %v", err)
		return nil, fmt.Errorf("request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		c.log.Error("Stream non-200: %d %s", resp.StatusCode, string(respBody))
		resp.Body.Close()
		return nil, fmt.Errorf("api error: status %d", resp.StatusCode)
	}

	ch := make(chan StreamEvent, 10)

	go func() {
		defer close(ch)
		defer resp.Body.Close()

		reader := NewSSEReader(resp.Body)
		for {
			event, err := reader.Read()
			if err == io.EOF {
				ch <- StreamEvent{Done: true}
				c.log.Debug("Stream complete")
				return
			}
			if err != nil {
				ch <- StreamEvent{Err: err}
				c.log.Error("Stream error: %v", err)
				return
			}

			if event.Data == "[DONE]" {
				ch <- StreamEvent{Done: true}
				c.log.Debug("Stream done marker")
				return
			}

			var chunk struct {
				Choices []struct {
					Delta struct {
						Content string `json:"content"`
					} `json:"delta"`
				} `json:"choices"`
			}

			if err := json.Unmarshal([]byte(event.Data), &chunk); err != nil {
				c.log.Warn("Failed to parse chunk: %v", err)
				continue
			}

			if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
				ch <- StreamEvent{Content: chunk.Choices[0].Delta.Content}
			}
		}
	}()

	return ch, nil
}

// SSEReader reads Server-Sent Events from a response body.
type SSEReader struct {
	reader *strings.Reader
}

// NewSSEReader creates a new SSE reader.
func NewSSEReader(r io.Reader) *SSEReader {
	data, _ := io.ReadAll(r)
	return &SSEReader{
		reader: strings.NewReader(string(data)),
	}
}

// SSEEvent represents a server-sent event.
type SSEEvent struct {
	Data string
}

// Read reads the next SSE event.
func (s *SSEReader) Read() (SSEEvent, error) {
	line, err := s.readLine()
	if err != nil {
		return SSEEvent{}, err
	}

	// Skip non-data lines
	if !strings.HasPrefix(line, "data:") {
		return s.Read() // Skip and try next
	}

	data := strings.TrimPrefix(line, "data: ")
	return SSEEvent{Data: data}, nil
}

func (s *SSEReader) readLine() (string, error) {
	var buf strings.Builder
	for {
		r, _, err := s.reader.ReadRune()
		if err != nil {
			if buf.Len() == 0 {
				return "", err
			}
			return buf.String(), nil
		}
		if r == '\n' {
			return buf.String(), nil
		}
		buf.WriteRune(r)
	}
}