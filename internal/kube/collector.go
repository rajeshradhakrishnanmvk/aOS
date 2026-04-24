package kube

import (
	"fmt"
	"os"
	"os/exec"
)

type KubeCollector struct{}

func NewKubeCollector() *KubeCollector {
	return &KubeCollector{}
}

func runKubectl(args ...string) (string, error) {
	cmd := exec.Command("kubectl", args...)
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("kubectl error: %s", string(exitErr.Stderr))
		}
		return "", err
	}
	return string(out), nil
}

func (k *KubeCollector) GetPodYAML(namespace, podName string) (string, error) {
	return runKubectl("get", "pod", podName, "-n", namespace, "-o", "yaml")
}

func (k *KubeCollector) DescribePod(namespace, podName string) (string, error) {
	return runKubectl("describe", "pod", podName, "-n", namespace)
}

func (k *KubeCollector) GetPodLogs(namespace, podName string, lines int) (string, error) {
	return runKubectl("logs", podName, "-n", namespace, fmt.Sprintf("--tail=%d", lines))
}

func (k *KubeCollector) GetPodEvents(namespace, podName string) (string, error) {
	return runKubectl("get", "events", "-n", namespace,
		fmt.Sprintf("--field-selector=involvedObject.name=%s", podName))
}

func (k *KubeCollector) GetDeploymentYAML(namespace, deployName string) (string, error) {
	return runKubectl("get", "deployment", deployName, "-n", namespace, "-o", "yaml")
}

func (k *KubeCollector) DescribeDeployment(namespace, deployName string) (string, error) {
	return runKubectl("describe", "deployment", deployName, "-n", namespace)
}

func (k *KubeCollector) GetRolloutStatus(namespace, deployName string) (string, error) {
	return runKubectl("rollout", "status", fmt.Sprintf("deployment/%s", deployName), "-n", namespace)
}

func (k *KubeCollector) GetDeploymentPods(namespace, deployName string) (string, error) {
	return runKubectl("get", "pods", "-n", namespace, fmt.Sprintf("-l=app=%s", deployName))
}

func (k *KubeCollector) ReadFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
