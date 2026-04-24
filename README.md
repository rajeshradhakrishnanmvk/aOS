# ClusterBrain 🧠

**A Privacy-First Kubernetes Incident Copilot**

ClusterBrain is a local-first CLI tool that diagnoses Kubernetes issues using local LLMs via Ollama. No cloud required. Your cluster data never leaves your machine.

## Features

- 🔍 **Diagnose pods** – Analyze pod failures, OOMKills, CrashLoopBackOffs, and more
- 📊 **Explain deployments** – Understand deployment health, rollout status, and pod states
- 📋 **Review manifests** – Static analysis of Kubernetes YAML files for anti-patterns
- 💡 **Suggest fixes** – Get AI-powered, human-reviewable remediation suggestions
- 🔒 **Privacy-first** – All LLM inference runs locally via Ollama

## Prerequisites

- Go 1.24+
- [kubectl](https://kubernetes.io/docs/tasks/tools/) configured and pointing at your cluster
- [Ollama](https://ollama.ai) running locally with `gemma4:e4b` pulled

## Installation

```bash
git clone https://github.com/rajeshradhakrishnanmvk/aOS.git
cd aOS
make build
make install
```

## Usage

```bash
# Diagnose a failing pod
brain diagnose pod <pod-name> -n <namespace>

# Explain a deployment
brain explain deployment <deployment-name> -n <namespace>

# Review a Kubernetes manifest file
brain review-file ./my-deployment.yaml

# Get fix suggestions for a pod
brain suggest-fix pod <pod-name> -n <namespace>

# Get fix suggestions for a deployment
brain suggest-fix deployment <deployment-name> -n <namespace>
```

## Configuration

Copy `configs/config.example.yaml` to `~/.config/clusterbrain/config.yaml`:

```bash
mkdir -p ~/.config/clusterbrain
cp configs/config.example.yaml ~/.config/clusterbrain/config.yaml
```

Edit to customize:

```yaml
ollama_url: "http://localhost:11434"
model: "gemma4:e4b"
namespace: "default"
log_level: "info"
max_log_lines: 100
request_timeout: 60
```

## Global Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--namespace` / `-n` | `default` | Kubernetes namespace |
| `--model` | `gemma4:e4b` | Ollama model to use |
| `--ollama-url` | `http://localhost:11434` | Ollama API URL |
| `--log-level` | `info` | Log level: debug, info, warn, error |

## Development

```bash
make build   # Build binary
make test    # Run tests
make lint    # Run linter (requires golangci-lint)
make clean   # Remove binary
```

## Architecture

See [ARCHITECTURE.md](ARCHITECTURE.md) for a detailed architecture overview.

## Safety

ClusterBrain only issues **read-only** kubectl commands. It never suggests `kubectl apply`, `delete`, `patch`, `scale`, or `exec`. All suggestions are human-reviewable before any action is taken.

## License

MIT
