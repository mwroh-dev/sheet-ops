package main

import (
	"encoding/json"
	"os"

	"github.com/mwroh/sheet-ops/runtime/openlayer/useorchestrator"
	"github.com/spf13/cobra"
)

func newRunValidatedCommand() *cobra.Command {
	var requestFile string

	cmd := &cobra.Command{
		Use:   "run-validated",
		Short: "Run a validated execution request through the closed runtime only",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			req, err := useorchestrator.LoadValidatedExecutionRequest(requestFile)
			if err != nil {
				return err
			}
			caseRoot, err := os.Getwd()
			if err != nil {
				return err
			}
			result, err := runCaseLocalRuntime(caseRoot, req)
			encoder := json.NewEncoder(cmd.OutOrStdout())
			encoder.SetIndent("", "  ")
			if writeErr := encoder.Encode(result); writeErr != nil {
				if err != nil {
					return err
				}
				return writeErr
			}
			return err
		},
	}

	cmd.Flags().StringVar(&requestFile, "request", "", "Path to a validated execution request JSON file")
	if err := cmd.MarkFlagRequired("request"); err != nil {
		panic(err)
	}
	return cmd
}

func runCaseLocalRuntime(caseRoot string, req useorchestrator.ValidatedExecutionRequest) (useorchestrator.RunResult, error) {
	restore, err := forceCaseLocalRuntimeRoots(caseRoot)
	if err != nil {
		return useorchestrator.RunResult{}, err
	}
	defer restore()
	return useorchestrator.OrchestrateValidated(req)
}
