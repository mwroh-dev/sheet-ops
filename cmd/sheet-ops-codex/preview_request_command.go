package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	runtimeknowledge "github.com/mwroh/sheet-ops/runtime/knowledge"
	requestcompiler "github.com/mwroh/sheet-ops/runtime/openlayer/requestcompiler"
	runtimeworkbookcase "github.com/mwroh/sheet-ops/runtime/workbookcase"
	"github.com/spf13/cobra"
)

type previewRequestOptions struct {
	IntentFile string
	InputFile  string
	OutputFile string
	ScenarioID string
}

type previewRequestResult struct {
	SchemaVersion    string                   `json:"schema_version"`
	Command          string                   `json:"command"`
	OK               bool                     `json:"ok"`
	ReadOnly         bool                     `json:"read_only"`
	DryRun           bool                     `json:"dry_run"`
	ScenarioID       string                   `json:"scenario_id"`
	InputFile        string                   `json:"input_file"`
	OutputFile       string                   `json:"output_file"`
	StateRoot        string                   `json:"state_root"`
	PlannedReads     []previewPlannedArtifact `json:"planned_reads"`
	PlannedWrites    []previewPlannedArtifact `json:"planned_writes"`
	PlannedArtifacts []previewPlannedArtifact `json:"planned_artifacts"`
	Limitations      []string                 `json:"limitations"`
}

type previewPlannedArtifact struct {
	Kind     string `json:"kind"`
	Path     string `json:"path"`
	Required bool   `json:"required"`
}

func newPreviewRequestCommand() *cobra.Command {
	var jsonOutput bool
	var options previewRequestOptions

	cmd := &cobra.Command{
		Use:   "preview-request",
		Short: "Inspect normalized-intent impact without runtime execution",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !jsonOutput {
				return newCLIError(cliErrorInvalidUsage, "preview-request requires --json", true, cliExitUsage, "preview-request --json")
			}
			result, err := runPreviewRequest(options)
			if err != nil {
				return err
			}
			return writePreviewRequestJSON(cmd.OutOrStdout(), result)
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Emit machine-readable JSON on stdout")
	cmd.Flags().StringVar(&options.IntentFile, "intent-file", "", "Path to a normalized intent JSON file")
	cmd.Flags().StringVar(&options.InputFile, "input-file", "", "Path to an input workbook")
	cmd.Flags().StringVar(&options.OutputFile, "output-file", "", "Path to the planned output workbook")
	cmd.Flags().StringVar(&options.ScenarioID, "scenario-id", "", "Optional scenario id override")
	if err := cmd.MarkFlagRequired("intent-file"); err != nil {
		panic(err)
	}
	if err := cmd.MarkFlagRequired("input-file"); err != nil {
		panic(err)
	}
	if err := cmd.MarkFlagRequired("output-file"); err != nil {
		panic(err)
	}
	return cmd
}

func runPreviewRequest(options previewRequestOptions) (previewRequestResult, error) {
	intentFile, err := filepath.Abs(options.IntentFile)
	if err != nil {
		return previewRequestResult{}, err
	}
	inputFile, err := filepath.Abs(options.InputFile)
	if err != nil {
		return previewRequestResult{}, err
	}
	outputFile, err := filepath.Abs(options.OutputFile)
	if err != nil {
		return previewRequestResult{}, err
	}
	if _, err := requestcompiler.LoadNormalizedIntent(intentFile); err != nil {
		return previewRequestResult{}, err
	}
	if info, err := os.Stat(inputFile); err != nil {
		return previewRequestResult{}, err
	} else if info.IsDir() {
		return previewRequestResult{}, newCLIError(cliErrorInvalidData, "input-file must be a workbook file, got directory", true, cliExitData, "preflight --json --input-file <workbook>")
	}

	workspaceRoot := resolveWorkspaceRoot(inputFile)
	stateRoot, err := previewStateRoot(workspaceRoot)
	if err != nil {
		return previewRequestResult{}, err
	}
	scenarioID := strings.TrimSpace(options.ScenarioID)
	if scenarioID == "" {
		scenarioID = defaultScenarioSlug("", intentFile)
	}

	return previewRequestResult{
		SchemaVersion: cliContractSchemaVersion,
		Command:       "preview-request",
		OK:            true,
		ReadOnly:      true,
		DryRun:        false,
		ScenarioID:    scenarioID,
		InputFile:     inputFile,
		OutputFile:    outputFile,
		StateRoot:     stateRoot,
		PlannedReads: []previewPlannedArtifact{
			{Kind: "normalized_intent", Path: intentFile, Required: true},
			{Kind: "input_workbook", Path: inputFile, Required: true},
		},
		PlannedWrites: []previewPlannedArtifact{
			{Kind: "output_workbook", Path: outputFile, Required: true},
		},
		PlannedArtifacts: []previewPlannedArtifact{
			{Kind: "state_root", Path: stateRoot, Required: true},
			{Kind: "request_compiler_artifacts", Path: filepath.Join(stateRoot, "artifacts", "work"), Required: true},
			{Kind: "runtime_evidence", Path: filepath.Join(stateRoot, "artifacts"), Required: true},
		},
		Limitations: []string{
			"Preview validates the normalized intent JSON and input workbook boundary only.",
			"Preview does not compile a persisted request, execute runtime orchestration, create .sheet-ops-state, or write the output workbook.",
			"Preview is not a dry-run and must not be used as evidence of execution success.",
		},
	}, nil
}

func previewStateRoot(workspaceRoot string) (string, error) {
	expectedStateRoot := filepath.Join(workspaceRoot, ".sheet-ops-state")
	stateRoot := strings.TrimSpace(os.Getenv(stateRootEnv))
	if stateRoot == "" {
		stateRoot = expectedStateRoot
	} else {
		stateRoot = filepath.Clean(stateRoot)
	}
	if !samePath(stateRoot, expectedStateRoot) {
		return "", newStateRootMismatchError(fmt.Sprintf(
			"%s=%q is not supported for preview-request; expected %q so planned artifacts stay under one workbook-case state root",
			stateRootEnv,
			stateRoot,
			expectedStateRoot,
		))
	}
	if err := validateDerivedStateRootEnv(runtimeworkbookcase.ArtifactRootEnv, filepath.Join(stateRoot, "artifacts")); err != nil {
		return "", err
	}
	if err := validateDerivedStateRootEnv(runtimeknowledge.KnowledgeRootEnv, filepath.Join(stateRoot, "knowledge")); err != nil {
		return "", err
	}
	return stateRoot, nil
}

func writePreviewRequestJSON(output io.Writer, result previewRequestResult) error {
	encoder := json.NewEncoder(output)
	encoder.SetIndent("", "  ")
	return encoder.Encode(result)
}
