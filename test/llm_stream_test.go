package test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/ZyqAlwaysCool/anvil/internal/platform/config"
	"github.com/ZyqAlwaysCool/anvil/internal/platform/llm"
)

func TestGenerateStreamReturnsIncrementalChunks(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/chat/completions") {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = io.WriteString(w, "data: {\"choices\":[{\"delta\":{\"content\":\"Hel\"}}]}\n\n")
		_, _ = io.WriteString(w, "data: {\"choices\":[{\"delta\":{\"content\":\"lo\"}}]}\n\n")
		_, _ = io.WriteString(w, "data: {\"choices\":[{\"delta\":{}}],\"usage\":{\"prompt_tokens\":1,\"completion_tokens\":2,\"total_tokens\":3}}\n\n")
		_, _ = io.WriteString(w, "data: [DONE]\n\n")
	}))
	defer server.Close()

	client, err := llm.New(config.LLMConfig{
		Enabled: true,
		APIKey:  "test-key",
		Model:   "gpt-4o-mini",
		BaseURL: server.URL,
	})
	if err != nil {
		t.Fatalf("new llm client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	chunkCh, errCh := client.GenerateStream(ctx, &llm.GenerateRequest{
		Model:    "gpt-4o-mini",
		Messages: []llm.Message{{Role: "user", Content: "hi"}},
	})

	var deltas []string
	var done bool
	for chunk := range chunkCh {
		if chunk.Delta != "" {
			deltas = append(deltas, chunk.Delta)
		}
		if chunk.Done {
			done = true
			if chunk.Usage == nil || chunk.Usage.TotalTokens != 3 {
				t.Fatalf("unexpected usage on done chunk: %+v", chunk.Usage)
			}
		}
	}
	if err := <-errCh; err != nil {
		t.Fatalf("stream error: %v", err)
	}
	if len(deltas) < 2 {
		t.Fatalf("expected incremental deltas, got %v", deltas)
	}
	if !done {
		t.Fatal("expected done chunk")
	}
	if strings.Join(deltas, "") != "Hello" {
		t.Fatalf("unexpected merged content: %q", strings.Join(deltas, ""))
	}
}
