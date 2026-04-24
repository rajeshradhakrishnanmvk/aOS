package diagnostics

import (
	"fmt"

	"github.com/rajeshradhakrishnanmvk/aOS/internal/config"
	"github.com/rajeshradhakrishnanmvk/aOS/internal/kube"
	"github.com/rajeshradhakrishnanmvk/aOS/internal/ollama"
	"github.com/rajeshradhakrishnanmvk/aOS/internal/prompt"
)

type DiagnosisService struct {
	collector     *kube.KubeCollector
	ollamaClient  *ollama.OllamaClient
	promptBuilder *prompt.PromptBuilder
	cfg           *config.Config
}

func NewDiagnosisService(cfg *config.Config) *DiagnosisService {
	return &DiagnosisService{
		collector:     kube.NewKubeCollector(),
		ollamaClient:  ollama.NewOllamaClient(cfg.OllamaURL, cfg.Model, cfg.RequestTimeout),
		promptBuilder: prompt.NewPromptBuilder(),
		cfg:           cfg,
	}
}

func (s *DiagnosisService) DiagnosePod(namespace, podName string) (string, error) {
	podYAML, err := s.collector.GetPodYAML(namespace, podName)
	if err != nil {
		podYAML = fmt.Sprintf("Error collecting pod YAML: %v", err)
	}

	describeOutput, err := s.collector.DescribePod(namespace, podName)
	if err != nil {
		describeOutput = fmt.Sprintf("Error collecting describe output: %v", err)
	}

	logs, err := s.collector.GetPodLogs(namespace, podName, s.cfg.MaxLogLines)
	if err != nil {
		logs = fmt.Sprintf("Error collecting logs: %v", err)
	}

	events, err := s.collector.GetPodEvents(namespace, podName)
	if err != nil {
		events = fmt.Sprintf("Error collecting events: %v", err)
	}

	userPrompt := s.promptBuilder.BuildDiagnosePrompt(podName, namespace, podYAML, describeOutput, logs, events)
	return s.ollamaClient.Chat(s.promptBuilder.SystemPrompt(), userPrompt)
}

func (s *DiagnosisService) ExplainDeployment(namespace, deployName string) (string, error) {
	deployYAML, err := s.collector.GetDeploymentYAML(namespace, deployName)
	if err != nil {
		deployYAML = fmt.Sprintf("Error collecting deployment YAML: %v", err)
	}

	describeOutput, err := s.collector.DescribeDeployment(namespace, deployName)
	if err != nil {
		describeOutput = fmt.Sprintf("Error collecting describe output: %v", err)
	}

	rolloutStatus, err := s.collector.GetRolloutStatus(namespace, deployName)
	if err != nil {
		rolloutStatus = fmt.Sprintf("Error collecting rollout status: %v", err)
	}

	pods, err := s.collector.GetDeploymentPods(namespace, deployName)
	if err != nil {
		pods = fmt.Sprintf("Error collecting pods: %v", err)
	}

	userPrompt := s.promptBuilder.BuildExplainPrompt(deployName, namespace, deployYAML, describeOutput, rolloutStatus, pods)
	return s.ollamaClient.Chat(s.promptBuilder.SystemPrompt(), userPrompt)
}

func (s *DiagnosisService) ReviewFile(filePath string) (string, error) {
	content, err := s.collector.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	userPrompt := s.promptBuilder.BuildReviewFilePrompt(filePath, content)
	return s.ollamaClient.Chat(s.promptBuilder.SystemPrompt(), userPrompt)
}

func (s *DiagnosisService) SuggestFix(resourceType, resourceName, namespace string) (string, error) {
	var context string

	switch resourceType {
	case "pod":
		podYAML, err := s.collector.GetPodYAML(namespace, resourceName)
		if err != nil {
			podYAML = fmt.Sprintf("Error: %v", err)
		}
		describeOutput, err := s.collector.DescribePod(namespace, resourceName)
		if err != nil {
			describeOutput = fmt.Sprintf("Error: %v", err)
		}
		events, err := s.collector.GetPodEvents(namespace, resourceName)
		if err != nil {
			events = fmt.Sprintf("Error: %v", err)
		}
		context = fmt.Sprintf("Pod YAML:\n%s\n\nDescribe:\n%s\n\nEvents:\n%s", podYAML, describeOutput, events)
	case "deployment":
		deployYAML, err := s.collector.GetDeploymentYAML(namespace, resourceName)
		if err != nil {
			deployYAML = fmt.Sprintf("Error: %v", err)
		}
		describeOutput, err := s.collector.DescribeDeployment(namespace, resourceName)
		if err != nil {
			describeOutput = fmt.Sprintf("Error: %v", err)
		}
		context = fmt.Sprintf("Deployment YAML:\n%s\n\nDescribe:\n%s", deployYAML, describeOutput)
	default:
		context = fmt.Sprintf("Resource type: %s, Name: %s", resourceType, resourceName)
	}

	userPrompt := s.promptBuilder.BuildSuggestFixPrompt(resourceType, resourceName, namespace, context)
	return s.ollamaClient.Chat(s.promptBuilder.SystemPrompt(), userPrompt)
}
