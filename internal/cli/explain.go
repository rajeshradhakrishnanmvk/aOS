package cli

import (
	"fmt"

	"github.com/rajeshradhakrishnanmvk/aOS/internal/config"
	"github.com/rajeshradhakrishnanmvk/aOS/internal/diagnostics"
	"github.com/rajeshradhakrishnanmvk/aOS/internal/output"
	"github.com/spf13/cobra"
)

var explainCmd = &cobra.Command{
	Use:   "explain",
	Short: "Explain Kubernetes resources",
}

var explainDeploymentCmd = &cobra.Command{
	Use:   "deployment <deployment-name>",
	Short: "Explain a Kubernetes deployment",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		deployName := args[0]
		cfg := config.GetConfig()
		if model != cfg.Model {
			cfg.Model = model
		}
		if ollamaURL != cfg.OllamaURL {
			cfg.OllamaURL = ollamaURL
		}

		formatter := output.NewFormatter()
		formatter.PrintProgress(fmt.Sprintf("Collecting deployment info for %s in namespace %s", deployName, namespace))

		svc := diagnostics.NewDiagnosisService(cfg)
		result, err := svc.ExplainDeployment(namespace, deployName)
		if err != nil {
			formatter.PrintError(err)
			return err
		}

		formatter.PrintExplanation(result)
		return nil
	},
}

func init() {
	explainCmd.AddCommand(explainDeploymentCmd)
	rootCmd.AddCommand(explainCmd)
}
