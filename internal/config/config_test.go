package config

import (
	"testing"
)

func TestDefaultsApplied(t *testing.T) {
	// Reset cfg to nil to force reload
	cfg = nil
	c := GetConfig()

	if c.OllamaURL == "" {
		t.Error("OllamaURL should have a default value")
	}
	if c.Model == "" {
		t.Error("Model should have a default value")
	}
	if c.Namespace == "" {
		t.Error("Namespace should have a default value")
	}
	if c.LogLevel == "" {
		t.Error("LogLevel should have a default value")
	}
	if c.MaxLogLines == 0 {
		t.Error("MaxLogLines should have a non-zero default")
	}
	if c.RequestTimeout == 0 {
		t.Error("RequestTimeout should have a non-zero default")
	}
}

func TestDefaultValues(t *testing.T) {
	cfg = nil
	c := GetConfig()

	if c.OllamaURL != "http://localhost:11434" {
		t.Errorf("expected OllamaURL 'http://localhost:11434', got %q", c.OllamaURL)
	}
	if c.Model != "gemma4:e4b" {
		t.Errorf("expected Model 'gemma4:e4b', got %q", c.Model)
	}
	if c.Namespace != "default" {
		t.Errorf("expected Namespace 'default', got %q", c.Namespace)
	}
	if c.MaxLogLines != 100 {
		t.Errorf("expected MaxLogLines 100, got %d", c.MaxLogLines)
	}
	if c.RequestTimeout != 60 {
		t.Errorf("expected RequestTimeout 60, got %d", c.RequestTimeout)
	}
}
