package cli

import (
	"fmt"

	"github.com/rajeshradhakrishnanmvk/aOS/internal/config"
	"github.com/rajeshradhakrishnanmvk/aOS/internal/diagnostics"
	"github.com/rajeshradhakrishnanmvk/aOS/internal/output"
	"github.com/spf13/cobra"
)

var diagnoseCmd = &cobra.Command{
	Use:   "diagnose",
	Short: "Diagnose Kubernetes resources",
}

var diagnosePodCmd = &cobra.Command{
	Use:   "pod <pod-name>",
	Short: "Diagnose a Kubernetes pod",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		podName := args[0]
		cfgCopy := *config.GetConfig()
		cfg := &cfgCopy
		cfg.Model = model
		cfg.OllamaURL = ollamaURL

		formatter := output.NewFormatter()
		formatter.PrintProgress(fmt.Sprintf("Collecting pod diagnostics for %s in namespace %s", podName, namespace))

		svc := diagnostics.NewDiagnosisService(cfg)
		if stream {
			formatter.PrintProgress(fmt.Sprintf("Collecting pod diagnostics for %s in namespace %s", podName, namespace))
			svcStream := diagnostics.NewDiagnosisService(cfg)
			err := svcStream.DiagnosePodStream(namespace, podName, func(chunk string) {
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
			result, err := svc.DiagnosePod(namespace, podName)
			if err != nil {
				formatter.PrintError(err)
				return err
			}

			formatter.PrintDiagnosis(result)
		}
		return nil
	},
}

func init() {
	diagnoseCmd.AddCommand(diagnosePodCmd)
	rootCmd.AddCommand(diagnoseCmd)
}
