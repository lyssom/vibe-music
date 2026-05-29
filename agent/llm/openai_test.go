package llm_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lyssom/vibe-music/agent/llm"
)

func TestOpenAIClientChat(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		resp := map[string]interface{}{
			"choices": []map[string]interface{}{
				{
					"message": map[string]string{
						"content": "C D E F G A B",
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	client := llm.NewOpenAIClient(
		"test-key",
		llm.WithBaseURL(ts.URL),
		llm.WithHTTPClient(ts.Client()),
	)

	messages := []llm.Message{
		{Role: "system", Content: "You are a music generator."},
		{Role: "user", Content: "Generate a C major scale."},
	}

	result, err := client.Chat(context.Background(), messages)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != "C D E F G A B" {
		t.Errorf("expected 'C D E F G A B', got %q", result)
	}
}

func TestOpenAIClientChat_ErrorResponse(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		resp := map[string]interface{}{
			"error": map[string]string{
				"message": "model overloaded",
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	client := llm.NewOpenAIClient(
		"test-key",
		llm.WithBaseURL(ts.URL),
		llm.WithHTTPClient(ts.Client()),
	)

	_, err := client.Chat(context.Background(), []llm.Message{
		{Role: "user", Content: "hello"},
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestOpenAIClientChatStream(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		flusher, ok := w.(http.Flusher)
		if !ok {
			t.Fatal("expected http.Flusher")
		}

		chunks := []string{
			`data: {"choices":[{"delta":{"content":"note"},"finish_reason":null}]}`,
			`data: {"choices":[{"delta":{"content":" C"},"finish_reason":null}]}`,
			`data: {"choices":[{"delta":{"content":" D"},"finish_reason":null}]}`,
			`data: {"choices":[{"delta":{"content":""},"finish_reason":"stop"}]}`,
			`data: [DONE]`,
		}
		for _, chunk := range chunks {
			w.Write([]byte(chunk + "\n\n"))
			flusher.Flush()
		}
	}))
	defer ts.Close()

	client := llm.NewOpenAIClient(
		"test-key",
		llm.WithBaseURL(ts.URL),
		llm.WithHTTPClient(ts.Client()),
	)

	ch, err := client.ChatStream(context.Background(), []llm.Message{
		{Role: "user", Content: "make a pattern"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result string
	for ev := range ch {
		result += ev.Content
	}

	if result != "note C D" {
		t.Errorf("expected 'note C D', got %q", result)
	}
}