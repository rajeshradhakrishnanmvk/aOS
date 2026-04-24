package cli

import (
	"fmt"
	"github.com/rajeshradhakrishnanmvk/aOS/internal/config"
	"github.com/rajeshradhakrishnanmvk/aOS/internal/diagnostics"
	"github.com/rajeshradhakrishnanmvk/aOS/internal/output"
	"github.com/spf13/cobra"
)

var reviewFileCmd = &cobra.Command{
	Use:   "review-file <file-path>",
	Short: "Review a Kubernetes manifest file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath := args[0]
		cfgCopy := *config.GetConfig()
		cfg := &cfgCopy
		cfg.Model = model
		cfg.OllamaURL = ollamaURL

		formatter := output.NewFormatter()
		formatter.PrintProgress("Reviewing file " + filePath)

		svc := diagnostics.NewDiagnosisService(cfg)
		if stream {
			formatter.PrintProgress("Reviewing file " + filePath)
			svcStream := diagnostics.NewDiagnosisService(cfg)
			err := svcStream.ReviewFileStream(filePath, func(chunk string) {
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
		result, err := svc.ReviewFile(filePath)
		if err != nil {
			formatter.PrintError(err)
			return err
		}

		formatter.PrintReview(result)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(reviewFileCmd)
}
