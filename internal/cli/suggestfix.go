package cli

import (
	"fmt"

	"github.com/rajeshradhakrishnanmvk/aOS/internal/config"
	"github.com/rajeshradhakrishnanmvk/aOS/internal/diagnostics"
	"github.com/rajeshradhakrishnanmvk/aOS/internal/output"
	"github.com/spf13/cobra"
)

var suggestFixCmd = &cobra.Command{
	Use:   "suggest-fix",
	Short: "Suggest fixes for Kubernetes resources",
}

var suggestFixPodCmd = &cobra.Command{
	Use:   "pod <pod-name>",
	Short: "Suggest fixes for a pod",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runSuggestFix("pod", args[0])
	},
}

var suggestFixDeploymentCmd = &cobra.Command{
	Use:   "deployment <deployment-name>",
	Short: "Suggest fixes for a deployment",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runSuggestFix("deployment", args[0])
	},
}

func runSuggestFix(resourceType, resourceName string) error {
	cfgCopy := *config.GetConfig()
	cfg := &cfgCopy
	cfg.Model = model
	cfg.OllamaURL = ollamaURL

	formatter := output.NewFormatter()
	formatter.PrintProgress(fmt.Sprintf("Analyzing %s %s in namespace %s", resourceType, resourceName, namespace))
	fmt.Println("\n⚠️  NOTE: These are suggestions only. Review before applying any changes.")

	svc := diagnostics.NewDiagnosisService(cfg)
	if stream {
		err := svc.SuggestFixStream(resourceType, resourceName, namespace, func(chunk string) {
			fmt.Print(chunk)
		}, func() {
			fmt.Println()
		})
		if err != nil {
			formatter.PrintError(err)
			return err
		}
		return nil
	}

	result, err := svc.SuggestFix(resourceType, resourceName, namespace)
	if err != nil {
		formatter.PrintError(err)
		return err
	}

	formatter.PrintSuggestion(result)
	return nil
}

func init() {
	suggestFixCmd.AddCommand(suggestFixPodCmd)
	suggestFixCmd.AddCommand(suggestFixDeploymentCmd)
	rootCmd.AddCommand(suggestFixCmd)
}
