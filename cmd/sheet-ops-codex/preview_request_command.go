package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
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
	Planner          string                   `json:"planner"`
	PlanConfidence   string                   `json:"plan_confidence"`
	Operation        string                   `json:"operation"`
	WouldMutate      bool                     `json:"would_mutate"`
	MutationSummary  previewMutationSummary   `json:"mutation_summary"`
	Fingerprints     previewFingerprints      `json:"fingerprints"`
	ScenarioID       string                   `json:"scenario_id"`
	InputFile        string                   `json:"input_file"`
	OutputFile       string                   `json:"output_file"`
	StateRoot        string                   `json:"state_root"`
	PlannedReads     []previewPlannedArtifact `json:"planned_reads"`
	PlannedWrites    []previewPlannedArtifact `json:"planned_writes"`
	PlannedArtifacts []previewPlannedArtifact `json:"planned_artifacts"`
	Limitations      []string                 `json:"limitations"`
}

type previewMutationSummary struct {
	Workbook          string `json:"workbook"`
	Artifacts         string `json:"artifacts"`
	State             string `json:"state"`
	ExecutionRequired bool   `json:"execution_required"`
}

type previewFingerprints struct {
	NormalizedIntentSHA256 string `json:"normalized_intent_sha256"`
	InputWorkbookSHA256    string `json:"input_workbook_sha256"`
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
				var typed *cliError
				if errors.As(err, &typed) {
					return writeCLIErrorJSONAndReturn(cmd.OutOrStdout(), typed)
				}
				classified := classifyCLIError(err)
				return writeCLIErrorJSONAndReturn(cmd.OutOrStdout(), &classified)
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
	intent, err := loadPreviewNormalizedIntent(intentFile)
	if err != nil {
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
	compilerResult, err := requestcompiler.ValidateIntent(requestcompiler.Input{
		RequestSource: requestcompiler.RequestSource{
			Kind: requestcompiler.RequestSourceDirectText,
			Text: "preview-request",
		},
		WorkspaceRoot:  workspaceRoot,
		WorkingDir:     workspaceRoot,
		ScenarioSlug:   scenarioID,
		InputWorkbooks: []requestcompiler.WorkbookInput{{Path: inputFile, Role: "primary_input"}},
		OutputFile:     outputFile,
	}, intent)
	if err != nil {
		return previewRequestResult{}, err
	}
	operation := strings.TrimSpace(compilerResult.Decision.SelectedOperation)
	if operation == "" {
		operation = "unknown"
	}
	wouldMutate := compilerResult.Decision.Status == requestcompiler.StatusCompiled
	mutationSummary := previewMutationSummary{
		Workbook:          "none",
		Artifacts:         "none",
		State:             "none",
		ExecutionRequired: false,
	}
	plannedWrites := []previewPlannedArtifact{}
	plannedArtifacts := []previewPlannedArtifact{}
	if wouldMutate {
		mutationSummary = previewMutationSummary{
			Workbook:          "planned_output_workbook",
			Artifacts:         "planned_runtime_artifacts",
			State:             "planned_state_root",
			ExecutionRequired: true,
		}
		plannedWrites = []previewPlannedArtifact{
			{Kind: "output_workbook", Path: outputFile, Required: true},
		}
		plannedArtifacts = []previewPlannedArtifact{
			{Kind: "state_root", Path: stateRoot, Required: true},
			{Kind: "request_compiler_artifacts", Path: filepath.Join(stateRoot, "artifacts", "work"), Required: true},
			{Kind: "runtime_evidence", Path: filepath.Join(stateRoot, "artifacts"), Required: true},
		}
	}
	fingerprints, err := inputFingerprints(intentFile, inputFile)
	if err != nil {
		return previewRequestResult{}, err
	}

	return previewRequestResult{
		SchemaVersion:   cliContractSchemaVersion,
		Command:         "preview-request",
		OK:              true,
		ReadOnly:        true,
		DryRun:          false,
		Planner:         "requestcompiler_validate_intent",
		PlanConfidence:  "compiler_validated_boundary",
		Operation:       operation,
		WouldMutate:     wouldMutate,
		MutationSummary: mutationSummary,
		Fingerprints:    fingerprints,
		ScenarioID:      scenarioID,
		InputFile:       inputFile,
		OutputFile:      outputFile,
		StateRoot:       stateRoot,
		PlannedReads: []previewPlannedArtifact{
			{Kind: "normalized_intent", Path: intentFile, Required: true},
			{Kind: "input_workbook", Path: inputFile, Required: true},
		},
		PlannedWrites:    plannedWrites,
		PlannedArtifacts: plannedArtifacts,
		Limitations: []string{
			"Preview runs non-persisting request-compiler validation over the normalized intent and input workbook boundary.",
			"Preview does not compile a persisted request, execute runtime orchestration, create .sheet-ops-state, or write the output workbook.",
			"Preview is not a dry-run and must not be used as evidence of execution success.",
		},
	}, nil
}

func inputFingerprints(intentFile string, inputFile string) (previewFingerprints, error) {
	intentFingerprint, err := fileSHA256(intentFile)
	if err != nil {
		return previewFingerprints{}, err
	}
	workbookFingerprint, err := fileSHA256(inputFile)
	if err != nil {
		return previewFingerprints{}, err
	}
	return previewFingerprints{
		NormalizedIntentSHA256: intentFingerprint,
		InputWorkbookSHA256:    workbookFingerprint,
	}, nil
}

func executedFingerprints(inputIdentity previewFingerprints, outputFile string, verificationOutputFingerprint string) (executionFingerprints, error) {
	outputFingerprint := strings.TrimSpace(verificationOutputFingerprint)
	if outputFingerprint == "" {
		var err error
		outputFingerprint, err = fileSHA256(outputFile)
		if err != nil {
			return executionFingerprints{}, err
		}
	}
	return executionFingerprints{
		NormalizedIntentSHA256: inputIdentity.NormalizedIntentSHA256,
		InputWorkbookSHA256:    inputIdentity.InputWorkbookSHA256,
		OutputWorkbookSHA256:   outputFingerprint,
	}, nil
}

func loadPreviewNormalizedIntent(path string) (requestcompiler.NormalizedIntent, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return requestcompiler.NormalizedIntent{}, err
	}
	intent, err := requestcompiler.LoadNormalizedIntentFromBytes(raw)
	if err != nil {
		return requestcompiler.NormalizedIntent{}, newInvalidDataError(fmt.Sprintf("load normalized intent %q: %v", path, err))
	}
	return intent, nil
}

func fileSHA256(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
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
