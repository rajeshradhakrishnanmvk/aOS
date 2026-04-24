package cli

import (
	"fmt"
	"os"

	"github.com/rajeshradhakrishnanmvk/aOS/internal/config"
	"github.com/spf13/cobra"
)

var (
	namespace        string
	model            string
	ollamaURL        string
	logLevel         string
	ollamaRetries    int
	ollamaBackoff    int
	ollamaMaxBackoff int
	stream           = true
)

var rootCmd = &cobra.Command{
	Use:   "brain",
	Short: "ClusterBrain - Privacy-first Kubernetes incident copilot",
	Long: `ClusterBrain diagnoses Kubernetes issues locally using Gemma 4 and Ollama.
No cloud. No data leaks. Your cluster data stays on your machine.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cfg := config.GetConfig()

	rootCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "default", "Kubernetes namespace")
	rootCmd.PersistentFlags().StringVar(&model, "model", cfg.Model, "Ollama model to use")
	rootCmd.PersistentFlags().StringVar(&ollamaURL, "ollama-url", cfg.OllamaURL, "Ollama API URL")
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "Log level (debug, info, warn, error)")
	rootCmd.PersistentFlags().IntVar(&ollamaRetries, "ollama-retries", cfg.OllamaRetries, "Number of retries for Ollama requests")
	rootCmd.PersistentFlags().IntVar(&ollamaBackoff, "ollama-backoff-ms", cfg.OllamaBackoffMs, "Base backoff in ms for Ollama retries")
	rootCmd.PersistentFlags().IntVar(&ollamaMaxBackoff, "ollama-max-backoff-ms", cfg.OllamaMaxBackoffMs, "Max backoff in ms for Ollama retries")
	rootCmd.PersistentFlags().BoolVar(&stream, "stream", true, "Stream partial responses from Ollama (default: true)")
}
