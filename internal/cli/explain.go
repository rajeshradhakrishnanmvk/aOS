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
		cfgCopy := *config.GetConfig()
		cfg := &cfgCopy
		cfg.Model = model
		cfg.OllamaURL = ollamaURL

		formatter := output.NewFormatter()
		formatter.PrintProgress(fmt.Sprintf("Collecting deployment info for %s in namespace %s", deployName, namespace))

		svc := diagnostics.NewDiagnosisService(cfg)
		if stream {
			formatter.PrintProgress(fmt.Sprintf("Collecting deployment info for %s in namespace %s", deployName, namespace))
			svcStream := diagnostics.NewDiagnosisService(cfg)
			err := svcStream.ExplainDeploymentStream(namespace, deployName, func(chunk string) {
				fmt.Print(chunk)
			}, func() {
				fmt.Println()
			})
			if err != nil {
				formatter.PrintError(err)
				return err
			}
			return nil
		} else {
			result, err := svc.ExplainDeployment(namespace, deployName)
			if err != nil {
				formatter.PrintError(err)
				return err
			}

			formatter.PrintExplanation(result)
		}
		return nil
	},
}

func init() {
	explainCmd.AddCommand(explainDeploymentCmd)
	rootCmd.AddCommand(explainCmd)
}
