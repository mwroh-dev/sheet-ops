package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	requestmode "github.com/mwroh/sheet-ops/runtime/openlayer/requestmode"
	requestpacker "github.com/mwroh/sheet-ops/runtime/openlayer/requestpacker"
	"github.com/spf13/cobra"
)

type prepareUseInput struct {
	ScenarioID   string
	RequestRef   string
	RequestKind  string
	InputFile    string
	OutputFile   string
	EnvelopeFile string
}

type prepareUseResult struct {
	Entry        string `json:"entry"`
	Status       string `json:"status"`
	ScenarioID   string `json:"scenario_id"`
	EnvelopePath string `json:"envelope_path"`
	RequestRef   string `json:"request_ref"`
	RequestKind  string `json:"request_kind"`
	InputFile    string `json:"input_file"`
	OutputFile   string `json:"output_file"`
}

func newPrepareUseCommand() *cobra.Command {
	var input prepareUseInput

	cmd := &cobra.Command{
		Use:   "prepare-use",
		Short: "Deterministically write a use envelope for the internal agent-system boundary",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := prepareUse(input)
			if err != nil {
				return err
			}
			encoder := json.NewEncoder(cmd.OutOrStdout())
			encoder.SetIndent("", "  ")
			return encoder.Encode(result)
		},
	}

	cmd.Flags().StringVar(&input.ScenarioID, "scenario-id", "", "Scenario identifier for the use envelope")
	cmd.Flags().StringVar(&input.RequestRef, "request-ref", "", "Path to the request reference file")
	cmd.Flags().StringVar(&input.RequestKind, "request-kind", requestmode.ModePromptText, "Typed request kind: prompt_text, structured_use_request, or organism_execution_request")
	cmd.Flags().StringVar(&input.InputFile, "input-file", "", "Path to the input workbook")
	cmd.Flags().StringVar(&input.OutputFile, "output-file", "", "Path to the output workbook")
	cmd.Flags().StringVar(&input.EnvelopeFile, "envelope-file", "", "Path to write the use-envelope JSON file")
	for _, name := range []string{"scenario-id", "request-ref", "input-file", "output-file", "envelope-file"} {
		if err := cmd.MarkFlagRequired(name); err != nil {
			panic(err)
		}
	}
	return cmd
}

func prepareUse(input prepareUseInput) (prepareUseResult, error) {
	scenarioID := strings.TrimSpace(input.ScenarioID)
	requestPath := cleanRequiredPath(input.RequestRef)
	requestKind := strings.TrimSpace(input.RequestKind)
	inputFile := cleanRequiredPath(input.InputFile)
	outputFile := cleanRequiredPath(input.OutputFile)
	envelopePath := cleanRequiredPath(input.EnvelopeFile)
	if scenarioID == "" {
		return prepareUseResult{}, fmt.Errorf("scenario-id is required")
	}
	if requestPath == "" {
		return prepareUseResult{}, fmt.Errorf("request-ref is required")
	}
	if requestKind == "" {
		return prepareUseResult{}, fmt.Errorf("request-kind is required")
	}
	if inputFile == "" {
		return prepareUseResult{}, fmt.Errorf("input-file is required")
	}
	if outputFile == "" {
		return prepareUseResult{}, fmt.Errorf("output-file is required")
	}
	if envelopePath == "" {
		return prepareUseResult{}, fmt.Errorf("envelope-file is required")
	}

	judgment, err := requestmode.JudgeRequestRef(requestmode.RequestRef{
		Kind: requestKind,
		Path: requestPath,
	})
	if err != nil {
		return prepareUseResult{}, err
	}
	envelope, err := requestpacker.Pack(requestpacker.Input{
		ScenarioID:  scenarioID,
		RequestPath: requestPath,
		InputFile:   inputFile,
		OutputFile:  outputFile,
		Judgment:    judgment,
	})
	if err != nil {
		return prepareUseResult{}, err
	}

	raw, err := json.MarshalIndent(envelope, "", "  ")
	if err != nil {
		return prepareUseResult{}, err
	}
	if err := os.MkdirAll(filepath.Dir(envelopePath), 0o755); err != nil {
		return prepareUseResult{}, err
	}
	if err := os.WriteFile(envelopePath, append(raw, '\n'), 0o644); err != nil {
		return prepareUseResult{}, err
	}
	return prepareUseResult{
		Entry:        "prepare-use",
		Status:       "prepared",
		ScenarioID:   envelope.ScenarioID,
		EnvelopePath: envelopePath,
		RequestRef:   envelope.Request.Path,
		RequestKind:  envelope.Request.Kind,
		InputFile:    envelope.InputFile,
		OutputFile:   envelope.OutputFile,
	}, nil
}

func cleanRequiredPath(path string) string {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" {
		return ""
	}
	return filepath.Clean(trimmed)
}
