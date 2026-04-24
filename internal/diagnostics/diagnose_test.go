package diagnostics

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rajeshradhakrishnanmvk/aOS/internal/config"
)

func TestDiagnosePodReturnsOllamaOutput(t *testing.T) {
	// fake Ollama server that returns a simple chat response
	h := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/chat" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"message":{"content":"fake diagnosis"}}`))
	}

	srv := httptest.NewServer(http.HandlerFunc(h))
	defer srv.Close()

	cfg := &config.Config{
		OllamaURL:      srv.URL,
		Model:          "test-model",
		RequestTimeout: 2,
		MaxLogLines:    10,
		OllamaRetries:  1,
	}

	s := NewDiagnosisService(cfg)
	out, err := s.DiagnosePod("default", "mypod")
	if err != nil {
		t.Fatalf("DiagnosePod returned error: %v", err)
	}
	if out != "fake diagnosis" {
		t.Fatalf("unexpected output: %q", out)
	}
	t.Logf("Ollama output: %s", out)
}
