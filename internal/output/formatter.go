package output

import (
	"fmt"
	"strings"
)

const (
	colorReset  = "\033[0m"
	colorBold   = "\033[1m"
	colorCyan   = "\033[36m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorRed    = "\033[31m"
	colorBlue   = "\033[34m"
	separator   = "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
)

type Formatter struct{}

func NewFormatter() *Formatter {
	return &Formatter{}
}

func (f *Formatter) colorizeOutput(result string) string {
	sectionColors := map[string]string{
		"SUMMARY:":            colorCyan + colorBold,
		"EVIDENCE:":           colorGreen + colorBold,
		"LIKELY_CAUSES:":      colorYellow + colorBold,
		"SAFE_ACTIONS:":       colorBlue + colorBold,
		"SUGGESTED_COMMANDS:": colorGreen + colorBold,
		"CONFIDENCE:":         colorBold,
		"LIMITATIONS:":        colorYellow + colorBold,
	}

	lines := strings.Split(result, "\n")
	var out strings.Builder

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		colored := false
		for section, color := range sectionColors {
			if strings.HasPrefix(trimmed, section) {
				out.WriteString(color + line + colorReset + "\n")
				colored = true
				break
			}
		}
		if !colored {
			out.WriteString(line + "\n")
		}
	}
	return out.String()
}

func (f *Formatter) printResult(header, result string) {
	fmt.Println(colorCyan + separator + colorReset)
	fmt.Println(colorBold + header + colorReset)
	fmt.Println(colorCyan + separator + colorReset)
	fmt.Println(f.colorizeOutput(result))
	fmt.Println(colorCyan + separator + colorReset)
}

func (f *Formatter) PrintDiagnosis(result string) {
	f.printResult("🔍 POD DIAGNOSIS", result)
}

func (f *Formatter) PrintExplanation(result string) {
	f.printResult("📊 DEPLOYMENT EXPLANATION", result)
}

func (f *Formatter) PrintReview(result string) {
	f.printResult("📋 FILE REVIEW", result)
}

func (f *Formatter) PrintSuggestion(result string) {
	f.printResult("💡 FIX SUGGESTIONS", result)
}

func (f *Formatter) PrintProgress(msg string) {
	fmt.Printf("%s⏳ %s...%s\n", colorYellow, msg, colorReset)
}

func (f *Formatter) PrintError(err error) {
	fmt.Printf("%s❌ Error: %v%s\n", colorRed, err, colorReset)
}
