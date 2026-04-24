package ollama

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestStreamChat(t *testing.T) {
	// httptest server that streams two JSON objects, then closes
	h := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		flusher, ok := w.(http.Flusher)
		if !ok {
			t.Fatalf("response does not support flushing")
		}

		enc := json.NewEncoder(w)

		// first chunk
		_ = enc.Encode(chatResponse{Message: message{Content: "part1"}})
		flusher.Flush()
		// small delay to simulate streaming
		time.Sleep(5 * time.Millisecond)

		// second chunk with Done=true
		_ = enc.Encode(chatResponse{Message: message{Content: "part2"}, Done: true})
		flusher.Flush()
	}

	srv := httptest.NewServer(http.HandlerFunc(h))
	defer srv.Close()

	client := NewOllamaClient(srv.URL, "test-model", 2)

	var chunks []string
	var doneCalled bool

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := client.StreamChat(ctx, "system", "user", func(s string) {
		chunks = append(chunks, s)
	}, func() {
		doneCalled = true
	})
	if err != nil {
		t.Fatalf("StreamChat returned error: %v", err)
	}

	if len(chunks) != 2 {
		t.Fatalf("expected 2 chunks, got %d: %v", len(chunks), chunks)
	}
	if chunks[0] != "part1" || chunks[1] != "part2" {
		t.Fatalf("unexpected chunk contents: %v", chunks)
	}
	if !doneCalled {
		t.Fatalf("expected onDone to be called")
	}
}
