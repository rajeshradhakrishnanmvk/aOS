package diagnostics

import (
	"context"
	"fmt"
	"time"

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
	client := ollama.NewOllamaClient(cfg.OllamaURL, cfg.Model, cfg.RequestTimeout)
	// apply config overrides for retries/backoff if provided via exported setters
	if cfg.OllamaRetries > 0 {
		client.SetRetries(cfg.OllamaRetries)
	}
	if cfg.OllamaBackoffMs > 0 {
		client.SetBackoff(cfg.OllamaBackoffMs)
	}
	if cfg.OllamaMaxBackoffMs > 0 {
		client.SetMaxBackoff(cfg.OllamaMaxBackoffMs)
	}

	return &DiagnosisService{
		collector:     kube.NewKubeCollector(),
		ollamaClient:  client,
		promptBuilder: prompt.NewPromptBuilder(),
		cfg:           cfg,
	}
}

func (s *DiagnosisService) DiagnosePod(namespace, podName string) (string, error) {
	// Collect artifacts in parallel to reduce time-to-request
	type result struct {
		val string
		err error
	}

	podCh := make(chan result, 1)
	descCh := make(chan result, 1)
	logsCh := make(chan result, 1)
	evCh := make(chan result, 1)

	go func() {
		v, err := s.collector.GetPodYAML(namespace, podName)
		podCh <- result{val: v, err: err}
	}()
	go func() {
		v, err := s.collector.DescribePod(namespace, podName)
		descCh <- result{val: v, err: err}
	}()
	go func() {
		v, err := s.collector.GetPodLogs(namespace, podName, s.cfg.MaxLogLines)
		logsCh <- result{val: v, err: err}
	}()
	go func() {
		v, err := s.collector.GetPodEvents(namespace, podName)
		evCh <- result{val: v, err: err}
	}()

	podRes := <-podCh
	descRes := <-descCh
	logsRes := <-logsCh
	evRes := <-evCh

	podYAML := podRes.val
	if podRes.err != nil {
		podYAML = fmt.Sprintf("Error collecting pod YAML: %v", podRes.err)
	}
	describeOutput := descRes.val
	if descRes.err != nil {
		describeOutput = fmt.Sprintf("Error collecting describe output: %v", descRes.err)
	}
	logs := logsRes.val
	if logsRes.err != nil {
		logs = fmt.Sprintf("Error collecting logs: %v", logsRes.err)
	}
	events := evRes.val
	if evRes.err != nil {
		events = fmt.Sprintf("Error collecting events: %v", evRes.err)
	}

	userPrompt := s.promptBuilder.BuildDiagnosePrompt(podName, namespace, podYAML, describeOutput, logs, events)
	return s.ollamaClient.Chat(s.promptBuilder.SystemPrompt(), userPrompt)
}

// Streaming variants
func (s *DiagnosisService) DiagnosePodStream(namespace, podName string, onData func(string), onDone func()) error {
	// Parallelize collection like non-stream variant
	type result struct {
		val string
		err error
	}

	podCh := make(chan result, 1)
	descCh := make(chan result, 1)
	logsCh := make(chan result, 1)
	evCh := make(chan result, 1)

	go func() {
		v, err := s.collector.GetPodYAML(namespace, podName)
		podCh <- result{val: v, err: err}
	}()
	go func() {
		v, err := s.collector.DescribePod(namespace, podName)
		descCh <- result{val: v, err: err}
	}()
	go func() {
		v, err := s.collector.GetPodLogs(namespace, podName, s.cfg.MaxLogLines)
		logsCh <- result{val: v, err: err}
	}()
	go func() {
		v, err := s.collector.GetPodEvents(namespace, podName)
		evCh <- result{val: v, err: err}
	}()

	podRes := <-podCh
	descRes := <-descCh
	logsRes := <-logsCh
	evRes := <-evCh

	podYAML := podRes.val
	if podRes.err != nil {
		podYAML = fmt.Sprintf("Error collecting pod YAML: %v", podRes.err)
	}
	describeOutput := descRes.val
	if descRes.err != nil {
		describeOutput = fmt.Sprintf("Error collecting describe output: %v", descRes.err)
	}
	logs := logsRes.val
	if logsRes.err != nil {
		logs = fmt.Sprintf("Error collecting logs: %v", logsRes.err)
	}
	events := evRes.val
	if evRes.err != nil {
		events = fmt.Sprintf("Error collecting events: %v", evRes.err)
	}

	userPrompt := s.promptBuilder.BuildDiagnosePrompt(podName, namespace, podYAML, describeOutput, logs, events)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(s.cfg.RequestTimeout)*time.Second)
	defer cancel()
	return s.ollamaClient.StreamChat(ctx, s.promptBuilder.SystemPrompt(), userPrompt, onData, onDone)
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

func (s *DiagnosisService) ExplainDeploymentStream(namespace, deployName string, onData func(string), onDone func()) error {
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
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(s.cfg.RequestTimeout)*time.Second)
	defer cancel()
	return s.ollamaClient.StreamChat(ctx, s.promptBuilder.SystemPrompt(), userPrompt, onData, onDone)
}

func (s *DiagnosisService) ReviewFile(filePath string) (string, error) {
	content, err := s.collector.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	userPrompt := s.promptBuilder.BuildReviewFilePrompt(filePath, content)
	return s.ollamaClient.Chat(s.promptBuilder.SystemPrompt(), userPrompt)
}

func (s *DiagnosisService) ReviewFileStream(filePath string, onData func(string), onDone func()) error {
	content, err := s.collector.ReadFile(filePath)
	if err != nil {
		// forward the error content as a small message and continue
		return fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	userPrompt := s.promptBuilder.BuildReviewFilePrompt(filePath, content)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(s.cfg.RequestTimeout)*time.Second)
	defer cancel()
	return s.ollamaClient.StreamChat(ctx, s.promptBuilder.SystemPrompt(), userPrompt, onData, onDone)
}

func (s *DiagnosisService) SuggestFix(resourceType, resourceName, namespace string) (string, error) {
	var context string

	switch resourceType {
	case "pod":
		// parallelize collection
		type result struct {
			val string
			err error
		}
		podCh := make(chan result, 1)
		descCh := make(chan result, 1)
		evCh := make(chan result, 1)

		go func() { v, err := s.collector.GetPodYAML(namespace, resourceName); podCh <- result{val: v, err: err} }()
		go func() { v, err := s.collector.DescribePod(namespace, resourceName); descCh <- result{val: v, err: err} }()
		go func() { v, err := s.collector.GetPodEvents(namespace, resourceName); evCh <- result{val: v, err: err} }()

		podRes := <-podCh
		descRes := <-descCh
		evRes := <-evCh

		podYAML := podRes.val
		if podRes.err != nil {
			podYAML = fmt.Sprintf("Error: %v", podRes.err)
		}
		describeOutput := descRes.val
		if descRes.err != nil {
			describeOutput = fmt.Sprintf("Error: %v", descRes.err)
		}
		events := evRes.val
		if evRes.err != nil {
			events = fmt.Sprintf("Error: %v", evRes.err)
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

func (s *DiagnosisService) SuggestFixStream(resourceType, resourceName, namespace string, onData func(string), onDone func()) error {
	var contextStr string

	switch resourceType {
	case "pod":
		type result struct {
			val string
			err error
		}
		podCh := make(chan result, 1)
		descCh := make(chan result, 1)
		evCh := make(chan result, 1)

		go func() { v, err := s.collector.GetPodYAML(namespace, resourceName); podCh <- result{val: v, err: err} }()
		go func() { v, err := s.collector.DescribePod(namespace, resourceName); descCh <- result{val: v, err: err} }()
		go func() { v, err := s.collector.GetPodEvents(namespace, resourceName); evCh <- result{val: v, err: err} }()

		podRes := <-podCh
		descRes := <-descCh
		evRes := <-evCh

		podYAML := podRes.val
		if podRes.err != nil {
			podYAML = fmt.Sprintf("Error: %v", podRes.err)
		}
		describeOutput := descRes.val
		if descRes.err != nil {
			describeOutput = fmt.Sprintf("Error: %v", descRes.err)
		}
		events := evRes.val
		if evRes.err != nil {
			events = fmt.Sprintf("Error: %v", evRes.err)
		}

		contextStr = fmt.Sprintf("Pod YAML:\n%s\n\nDescribe:\n%s\n\nEvents:\n%s", podYAML, describeOutput, events)
	case "deployment":
		deployYAML, err := s.collector.GetDeploymentYAML(namespace, resourceName)
		if err != nil {
			deployYAML = fmt.Sprintf("Error: %v", err)
		}
		describeOutput, err := s.collector.DescribeDeployment(namespace, resourceName)
		if err != nil {
			describeOutput = fmt.Sprintf("Error: %v", err)
		}
		contextStr = fmt.Sprintf("Deployment YAML:\n%s\n\nDescribe:\n%s", deployYAML, describeOutput)
	default:
		contextStr = fmt.Sprintf("Resource type: %s, Name: %s", resourceType, resourceName)
	}

	userPrompt := s.promptBuilder.BuildSuggestFixPrompt(resourceType, resourceName, namespace, contextStr)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(s.cfg.RequestTimeout)*time.Second)
	defer cancel()
	return s.ollamaClient.StreamChat(ctx, s.promptBuilder.SystemPrompt(), userPrompt, onData, onDone)
}
