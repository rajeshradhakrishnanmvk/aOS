# ClusterBrain Architecture

## Overview

ClusterBrain is a CLI tool built with Go that integrates Kubernetes cluster diagnostics with local LLM inference via Ollama.

```
┌─────────────────────────────────────────────────────────────┐
│                        CLI Layer                             │
│  brain diagnose pod  |  brain explain deployment  |  ...    │
│         (cobra commands — internal/cli/)                     │
└───────────────────────┬─────────────────────────────────────┘
                        │
┌───────────────────────▼─────────────────────────────────────┐
│                   Diagnostics Service                        │
│              (internal/diagnostics/service.go)               │
│   Orchestrates: collect → build prompt → call LLM → output  │
└───────────┬──────────────────────────┬──────────────────────┘
            │                          │
┌───────────▼──────────┐  ┌────────────▼────────────────────┐
│   Kube Collector     │  │        Ollama Client             │
│  (internal/kube/)    │  │      (internal/ollama/)          │
│  Read-only kubectl   │  │  POST /api/chat (local LLM)      │
│  wrapper calls       │  │  No data leaves the machine      │
└──────────────────────┘  └─────────────────────────────────┘
            │
┌───────────▼──────────────────────────────────────────────────┐
│                    Prompt Builder                             │
│                 (internal/prompt/)                            │
│  Builds structured prompts with YAML/logs/events context     │
└──────────────────────────────────────────────────────────────┘
```

## Package Structure

| Package | Responsibility |
|---------|---------------|
| `cmd/brain` | Entry point — calls `cli.Execute()` |
| `internal/cli` | Cobra command definitions and flag parsing |
| `internal/config` | Viper-based config loading with defaults |
| `internal/logging` | Zap-based structured logger (singleton) |
| `internal/kube` | Read-only kubectl wrappers via `os/exec` |
| `internal/ollama` | HTTP client for Ollama `/api/chat` endpoint |
| `internal/prompt` | Prompt construction with context trimming |
| `internal/diagnostics` | Service layer orchestrating the full diagnosis flow |
| `internal/output` | ANSI-colored terminal formatter |

## Data Flow

1. **User** runs `brain diagnose pod my-pod -n production`
2. **CLI** parses flags, updates config, calls `DiagnosisService.DiagnosePod()`
3. **KubeCollector** runs kubectl commands (get pod YAML, describe, logs, events)
4. **PromptBuilder** assembles structured prompt with collected data (truncated to fit context)
5. **OllamaClient** sends prompt to local Ollama instance (`POST /api/chat`)
6. **Formatter** colorizes and prints the structured response

## Design Decisions

### Privacy First
All LLM inference is local. The `OllamaClient` only communicates with `127.0.0.1:11434` by default. No telemetry, no cloud.

### Read-Only Safety
The `KubeCollector` only wraps read-only kubectl subcommands. The system prompt explicitly instructs the LLM not to suggest mutating commands. The `suggest-fix` command prints a prominent warning that suggestions are human-reviewable only.

### Structured Output
The system prompt enforces a consistent schema (SUMMARY, EVIDENCE, LIKELY_CAUSES, SAFE_ACTIONS, SUGGESTED_COMMANDS, CONFIDENCE, LIMITATIONS). The formatter colorizes each section heading.

### Context Trimming
Pod YAMLs and logs can be very large. The `PromptBuilder` trims YAML to key status fields and logs to the last N lines to stay within LLM context windows.
