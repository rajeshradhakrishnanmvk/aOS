package prompt

import (
	"strings"
	"testing"
)

func TestTrimLogs(t *testing.T) {
	logs := "line1\nline2\nline3\nline4\nline5"
	result := trimLogs(logs, 3)
	lines := strings.Split(result, "\n")
	if len(lines) != 3 {
		t.Errorf("expected 3 lines, got %d", len(lines))
	}
	if !strings.Contains(result, "line3") {
		t.Errorf("expected last lines to be kept")
	}
}

func TestTrimLogsNoTruncation(t *testing.T) {
	logs := "line1\nline2"
	result := trimLogs(logs, 10)
	if result != logs {
		t.Errorf("expected no truncation, got %q", result)
	}
}

func TestTrimYAML(t *testing.T) {
	yaml := "line1\nline2\nline3\nline4\nline5"
	result := trimYAML(yaml, 3)
	if !strings.Contains(result, "truncated") {
		t.Errorf("expected truncated marker")
	}
	lines := strings.Split(result, "\n")
	if lines[0] != "line1" {
		t.Errorf("expected first line to be line1")
	}
}

func TestTrimYAMLNoTruncation(t *testing.T) {
	yaml := "line1\nline2"
	result := trimYAML(yaml, 10)
	if result != yaml {
		t.Errorf("expected no truncation")
	}
}

func TestBuildDiagnosePromptContainsSections(t *testing.T) {
	p := NewPromptBuilder()
	result := p.BuildDiagnosePrompt("my-pod", "default", "podYAML", "describeOut", "log line1\nlog line2", "event1")

	sections := []string{"Pod:", "Namespace:", "POD STATUS", "DESCRIBE OUTPUT", "RECENT LOGS", "EVENTS"}
	for _, section := range sections {
		if !strings.Contains(result, section) {
			t.Errorf("expected section %q in prompt, got:\n%s", section, result)
		}
	}
}

func TestSystemPromptContainsSchema(t *testing.T) {
	p := NewPromptBuilder()
	result := p.SystemPrompt()

	sections := []string{"SUMMARY", "EVIDENCE", "LIKELY_CAUSES", "SAFE_ACTIONS", "SUGGESTED_COMMANDS", "CONFIDENCE", "LIMITATIONS"}
	for _, section := range sections {
		if !strings.Contains(result, section) {
			t.Errorf("expected %q in system prompt", section)
		}
	}
}

func TestBuildSuggestFixPromptNoUnsafeCommands(t *testing.T) {
	p := NewPromptBuilder()
	result := p.BuildSuggestFixPrompt("pod", "my-pod", "default", "some context")

	unsafeVerbs := []string{"kubectl apply", "kubectl delete", "kubectl patch", "kubectl scale", "kubectl exec"}
	for _, verb := range unsafeVerbs {
		if strings.Contains(result, verb) {
			t.Errorf("suggest fix prompt should not contain %q", verb)
		}
	}

	if !strings.Contains(result, "DO NOT") {
		t.Errorf("expected explicit restriction against unsafe commands")
	}
}
