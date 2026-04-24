package output

import (
	"errors"
	"testing"
)

func TestFormatterDoesNotPanic(t *testing.T) {
	f := NewFormatter()

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("formatter panicked: %v", r)
		}
	}()

	f.PrintDiagnosis("test diagnosis result")
	f.PrintExplanation("test explanation result")
	f.PrintReview("test review result")
	f.PrintSuggestion("test suggestion result")
	f.PrintProgress("test progress message")
	f.PrintError(errors.New("test error"))
}

func TestFormatterWithStructuredOutput(t *testing.T) {
	f := NewFormatter()

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("formatter panicked with structured output: %v", r)
		}
	}()

	structured := `SUMMARY:
This is a test summary.

EVIDENCE:
- Evidence item 1
- Evidence item 2

LIKELY_CAUSES:
- Cause 1

SAFE_ACTIONS:
- Action 1

SUGGESTED_COMMANDS:
kubectl get pods -n default

CONFIDENCE: HIGH

LIMITATIONS:
- Limitation 1`

	f.PrintDiagnosis(structured)
}

func TestColorizeOutput(t *testing.T) {
	f := NewFormatter()

	result := f.colorizeOutput("SUMMARY:\nsome text\nEVIDENCE:\nmore text")
	if result == "" {
		t.Error("colorizeOutput returned empty string")
	}
}
