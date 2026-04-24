package ollama

import (
	"os"
	"strings"
	"testing"
)

// TestLocalOllamaChat is an optional integration test that attempts to
// contact a local Ollama server. It will skip if the server is not
// reachable so running `go test ./...` won't fail in environments
// without Ollama installed.
func TestLocalOllamaChat(t *testing.T) {
	base := os.Getenv("OLLAMA_BASE_URL")
	if base == "" {
		base = "http://127.0.0.1:11434"
	}
	model := os.Getenv("OLLAMA_MODEL")
	if model == "" {
		model = "gemma4:e4b"
	}

	// Use a longer timeout and a few retries for the integration test
	// to accommodate slower local Ollama startups or heavier models.
	client := NewOllamaClient(base, model, 30)
	client.SetRetries(3)
	client.SetBackoff(500)
	client.SetMaxBackoff(2000)

	resp, err := client.Chat("system:you are a test responder", "Hello from integration test")
	if err != nil {
		// If the host isn't reachable, treat as skipped (not a failure).
		e := strings.ToLower(err.Error())
		if strings.Contains(e, "connect") ||
			strings.Contains(e, "connection refused") ||
			strings.Contains(e, "no such host") ||
			strings.Contains(e, "deadline exceeded") ||
			strings.Contains(e, "timeout") {
			t.Skipf("Ollama not reachable at %s: %v", base, err)
		}
		t.Fatalf("Chat returned error: %v", err)
	}

	if strings.TrimSpace(resp) == "" {
		t.Fatalf("empty response from Ollama at %s", base)
	}
}
