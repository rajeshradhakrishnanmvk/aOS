package prompt

import (
	"fmt"
	"strings"
)

type PromptBuilder struct{}

func NewPromptBuilder() *PromptBuilder {
	return &PromptBuilder{}
}

func (p *PromptBuilder) SystemPrompt() string {
	return `You are ClusterBrain, an expert Kubernetes troubleshooting assistant.

Rules:
- Clearly distinguish between FACTS (observed), INFERENCES (likely causes), and RECOMMENDATIONS (suggested actions).
- Never make unsafe assumptions about production workloads.
- Only suggest READ-ONLY kubectl commands unless explicitly discussing remediation.
- Do not suggest kubectl apply, delete, patch, scale, or exec commands.

Always respond using this exact schema:

SUMMARY:
<One paragraph overview of the issue>

EVIDENCE:
<Bullet list of observed facts from the provided data>

LIKELY_CAUSES:
<Bullet list of probable root causes, ranked by likelihood>

SAFE_ACTIONS:
<Bullet list of safe investigation steps>

SUGGESTED_COMMANDS:
<Read-only kubectl commands only, one per line>

CONFIDENCE: <HIGH|MEDIUM|LOW>

LIMITATIONS:
<What you cannot determine from the available data>`
}

func (p *PromptBuilder) BuildDiagnosePrompt(podName, namespace, podYAML, describeOutput, logs, events string) string {
	trimmedYAML := trimYAML(extractKeyFields(podYAML), 80)
	trimmedLogs := trimLogs(logs, 50)
	trimmedDescribe := trimYAML(describeOutput, 100)
	trimmedEvents := trimYAML(events, 50)

	return fmt.Sprintf(`Diagnose the following Kubernetes pod issue.

Pod: %s
Namespace: %s

=== POD STATUS (key fields) ===
%s

=== DESCRIBE OUTPUT ===
%s

=== RECENT LOGS (last 50 lines) ===
%s

=== EVENTS ===
%s

Please diagnose the issue using the structured schema.`,
		podName, namespace, trimmedYAML, trimmedDescribe, trimmedLogs, trimmedEvents)
}

func (p *PromptBuilder) BuildExplainPrompt(deployName, namespace, deployYAML, describeOutput, rolloutStatus, pods string) string {
	trimmedYAML := trimYAML(deployYAML, 80)
	trimmedDescribe := trimYAML(describeOutput, 80)

	return fmt.Sprintf(`Explain the status of the following Kubernetes deployment.

Deployment: %s
Namespace: %s

=== DEPLOYMENT YAML (key fields) ===
%s

=== DESCRIBE OUTPUT ===
%s

=== ROLLOUT STATUS ===
%s

=== PODS ===
%s

Please explain the deployment status using the structured schema.`,
		deployName, namespace, trimmedYAML, trimmedDescribe, rolloutStatus, pods)
}

func (p *PromptBuilder) BuildReviewFilePrompt(filename, content string) string {
	trimmedContent := trimYAML(content, 200)

	return fmt.Sprintf(`Review the following Kubernetes manifest file for issues.

Filename: %s

=== CONTENT ===
%s

Please analyze this manifest and provide:
1. Purpose summary - what does this manifest do?
2. Anti-patterns detected
3. Missing probes (liveness, readiness, startup)
4. Missing resource limits/requests
5. Missing security context
6. Risky defaults or configurations
7. Suggested improvements

Use the structured schema in your response.`,
		filename, trimmedContent)
}

func (p *PromptBuilder) BuildSuggestFixPrompt(resourceType, resourceName, namespace, context string) string {
	return fmt.Sprintf(`Suggest remediation steps for the following Kubernetes %s issue.

Resource Type: %s
Resource Name: %s
Namespace: %s

=== CONTEXT ===
%s

IMPORTANT CONSTRAINTS:
- DO NOT suggest mutating commands (apply, delete, patch, scale, or exec).
- Only provide human-reviewable suggestions and safe read-only commands.
- All suggestions must be reviewed by a human before implementation.
- Focus on root cause analysis and safe remediation strategies.

Please provide safe, human-reviewable remediation suggestions using the structured schema.`,
		resourceType, resourceType, resourceName, namespace, context)
}

func trimYAML(yaml string, maxLines int) string {
	lines := strings.Split(yaml, "\n")
	if len(lines) <= maxLines {
		return yaml
	}
	return strings.Join(lines[:maxLines], "\n") + "\n... (truncated)"
}

func trimLogs(logs string, maxLines int) string {
	lines := strings.Split(logs, "\n")
	if len(lines) <= maxLines {
		return logs
	}
	return strings.Join(lines[len(lines)-maxLines:], "\n")
}

func extractKeyFields(yaml string) string {
	var result strings.Builder
	lines := strings.Split(yaml, "\n")

	inStatus := false
	inConditions := false
	inContainerStatuses := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "status:") {
			inStatus = true
			result.WriteString(line + "\n")
			continue
		}

		if inStatus {
			leadingSpaces := len(line) - len(strings.TrimLeft(line, " "))
			if leadingSpaces == 0 && trimmed != "" && !strings.HasPrefix(trimmed, "#") {
				inStatus = false
				inConditions = false
				inContainerStatuses = false
				continue
			}

			if strings.Contains(trimmed, "conditions:") {
				inConditions = true
			}
			if strings.Contains(trimmed, "containerStatuses:") {
				inContainerStatuses = true
			}

			if inConditions || inContainerStatuses || strings.Contains(trimmed, "phase:") ||
				strings.Contains(trimmed, "reason:") || strings.Contains(trimmed, "message:") ||
				strings.Contains(trimmed, "ready:") || strings.Contains(trimmed, "restartCount:") {
				result.WriteString(line + "\n")
			}
		}

		// Also include metadata name/namespace at top
		if strings.Contains(trimmed, "name:") || strings.Contains(trimmed, "namespace:") {
			if !inStatus {
				result.WriteString(line + "\n")
			}
		}
	}

	if result.Len() == 0 {
		return yaml
	}
	return result.String()
}
