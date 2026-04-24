//go:build integration
// +build integration

package diagnostics

import (
	"testing"

	"github.com/rajeshradhakrishnanmvk/aOS/internal/config"
)

// Integration test: streams a real diagnosis from the local Ollama/kubectl setup.
// Run with: go test -v -tags=integration ./internal/diagnostics -run TestDiagnosePodStreamIntegration
func TestDiagnosePodStreamIntegration(t *testing.T) {
	cfg := config.GetConfig()
	// ensure reasonable timeout for integration
	cfg.RequestTimeout = 120

	s := NewDiagnosisService(cfg)

	var chunks []string
	err := s.DiagnosePodStream(cfg.Namespace, "moonlight-identity-5fdf64959-826pp", func(chunk string) {
		t.Logf("chunk: %s", chunk)
		chunks = append(chunks, chunk)
	}, func() {
		t.Log("stream done")
	})
	if err != nil {
		t.Fatalf("DiagnosePodStream returned error: %v", err)
	}
	if len(chunks) == 0 {
		t.Fatalf("no chunks received from stream")
	}
}
