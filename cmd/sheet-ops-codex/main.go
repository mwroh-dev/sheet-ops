package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	runtimeknowledge "github.com/mwroh/sheet-ops/runtime/knowledge"
	requestcompiler "github.com/mwroh/sheet-ops/runtime/openlayer/requestcompiler"
	"github.com/mwroh/sheet-ops/runtime/openlayer/useorchestrator"
	runtimeconfig "github.com/mwroh/sheet-ops/runtime/runtimeconfig"
	runtimeworkbookcase "github.com/mwroh/sheet-ops/runtime/workbookcase"
	"github.com/spf13/cobra"
)

const stateRootEnv = "SHEET_OPS_STATE_ROOT"

func main() {
	if err := newRootCommand().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func newRootCommand() *cobra.Command {
	use := os.Getenv("SHEET_OPS_CLI_NAME")
	if use == "" {
		use = "sheet-ops-codex"
	}
	rootCmd := &cobra.Command{
		Use:           use,
		Short:         "Thin public CLI adapter for Sheet Ops use requests",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	rootCmd.AddCommand(newPrepareUseCommand())
	rootCmd.AddCommand(newRunValidatedCommand())
	rootCmd.AddCommand(newRunRequestCommand())
	rootCmd.AddCommand(newRunIntentCommand())
	rootCmd.AddCommand(newInstallSkillCommand())
	applyCLIContracts(rootCmd)
	return rootCmd
}

func newRunValidatedCommand() *cobra.Command {
	var requestFile string

	cmd := &cobra.Command{
		Use:   "run-validated",
		Short: "Run a validated execution request",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runValidatedExecutionRequest(cmd, requestFile)
		},
	}

	cmd.Flags().StringVar(&requestFile, "request", "", "Path to a validated execution request JSON file")
	if err := cmd.MarkFlagRequired("request"); err != nil {
		panic(err)
	}

	return cmd
}

func newRunRequestCommand() *cobra.Command {
	var requestFile string

	cmd := &cobra.Command{
		Use:    "run-request",
		Hidden: true,
		Short:  "Compatibility alias for run-validated",
		Args:   cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runValidatedExecutionRequest(cmd, requestFile)
		},
	}

	cmd.Flags().StringVar(&requestFile, "file", "", "Path to a validated execution request JSON file")
	if err := cmd.MarkFlagRequired("file"); err != nil {
		panic(err)
	}

	return cmd
}

func runValidatedExecutionRequest(cmd *cobra.Command, requestFile string) error {
	validatedReq, err := useorchestrator.LoadValidatedExecutionRequest(requestFile)
	if err != nil {
		return err
	}
	validatedReq, err = prepareValidatedExecutionRequest(requestFile, validatedReq)
	if err != nil {
		return err
	}
	result, orchestrateErr := useorchestrator.OrchestrateValidated(validatedReq)
	if err := writeResultJSON(cmd.OutOrStdout(), result); err != nil {
		return errors.Join(orchestrateErr, err)
	}
	return orchestrateErr
}

func prepareValidatedExecutionRequest(requestFile string, req useorchestrator.ValidatedExecutionRequest) (useorchestrator.ValidatedExecutionRequest, error) {
	resolvedReq, err := resolveValidatedRequestWorkbookPaths(requestFile, req)
	if err != nil {
		return req, err
	}
	if _, err := configureStateRootsFromStateRoot(resolveWorkspaceRoot(resolvedReq.InputFile)); err != nil {
		return req, err
	}
	if err := ensureSafeRuntimeDefaults(); err != nil {
		return req, err
	}
	return resolvedReq, nil
}

func resolveValidatedRequestWorkbookPaths(requestFile string, req useorchestrator.ValidatedExecutionRequest) (useorchestrator.ValidatedExecutionRequest, error) {
	requestPath, err := filepath.Abs(requestFile)
	if err != nil {
		return req, err
	}
	requestDir := filepath.Dir(requestPath)
	req.InputFile = resolveRelativePathFromRequestDir(requestDir, req.InputFile)
	req.OutputFile = resolveRelativePathFromRequestDir(requestDir, req.OutputFile)
	return req, nil
}

func resolveRelativePathFromRequestDir(requestDir, path string) string {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" {
		return ""
	}
	if filepath.IsAbs(trimmed) {
		return filepath.Clean(trimmed)
	}
	return filepath.Clean(filepath.Join(requestDir, trimmed))
}

type intentCompileInput struct {
	IntentFile string
	InputFile  string
	OutputFile string
	ScenarioID string
}

func newRunIntentCommand() *cobra.Command {
	var input intentCompileInput

	cmd := &cobra.Command{
		Use:   "run-intent",
		Short: "Validate a normalized intent document and run it",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			intent, err := requestcompiler.LoadNormalizedIntent(input.IntentFile)
			if err != nil {
				return err
			}

			runtimeResult, publicEntryResult, orchestrateErr := runIntentEntry(intent, input)
			if publicEntryResult != nil {
				if err := writePublicEntryResultJSON(cmd.OutOrStdout(), *publicEntryResult); err != nil {
					return errors.Join(orchestrateErr, err)
				}
				return orchestrateErr
			}
			if orchestrateErr != nil {
				return orchestrateErr
			}
			_ = runtimeResult
			return fmt.Errorf("run-intent did not emit a public entry envelope")
		},
	}

	cmd.Flags().StringVar(&input.IntentFile, "intent-file", "", "Path to a normalized intent JSON file")
	cmd.Flags().StringVar(&input.InputFile, "input-file", "", "Path to an input workbook")
	cmd.Flags().StringVar(&input.OutputFile, "output-file", "", "Path to an output workbook")
	cmd.Flags().StringVar(&input.ScenarioID, "scenario-id", "", "Optional scenario id override")
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

func runIntentEntry(intent requestcompiler.NormalizedIntent, input intentCompileInput) (*useorchestrator.RunResult, *PublicEntryResult, error) {
	processWorkingDir, err := os.Getwd()
	if err != nil {
		return nil, nil, err
	}

	workspaceRoot := resolveWorkspaceRoot(input.InputFile)
	workspaceRoot, err = configureStateRootsFromStateRoot(workspaceRoot)
	if err != nil {
		return nil, nil, err
	}

	compiled, err := requestcompiler.ValidateIntentAndPersist(requestcompiler.Input{
		RequestSource: requestcompiler.RequestSource{
			Kind: requestcompiler.RequestSourceDirectText,
			Text: normalizedIntentRequestText(intent),
		},
		WorkspaceRoot: workspaceRoot,
		WorkingDir:    processWorkingDir,
		ScenarioSlug:  defaultScenarioSlug(input.ScenarioID, input.IntentFile),
		InputWorkbooks: []requestcompiler.WorkbookInput{
			{Path: input.InputFile, Role: "primary_input"},
		},
		OutputFile: input.OutputFile,
	}, intent)
	if err != nil {
		return nil, nil, err
	}
	if compiled.ValidatedExecutionRequest == nil {
		switch compiled.Decision.Status {
		case requestcompiler.StatusBlocked, requestcompiler.StatusNeedsHumanCheckpoint:
			result, stopErr := newTerminalCompilerResult("run-intent", compiled)
			return nil, &result, stopErr
		default:
			return nil, nil, fmt.Errorf("normalized intent did not compile (status=%s)", compiled.Decision.Status)
		}
	}

	result, orchestrateErr := useorchestrator.OrchestrateValidated(*compiled.ValidatedExecutionRequest)
	if orchestrateErr != nil {
		if runtimeStarted(result) {
			return &result, nil, orchestrateErr
		}
		return nil, nil, orchestrateErr
	}
	publicEntryResult := newExecutedPublicEntryResult("run-intent", compiled, result)
	return &result, &publicEntryResult, orchestrateErr
}

func configureStateRootsFromStateRoot(workspaceRoot string) (string, error) {
	expectedStateRoot := filepath.Join(workspaceRoot, ".sheet-ops-state")
	stateRoot := strings.TrimSpace(os.Getenv(stateRootEnv))
	if stateRoot == "" {
		stateRoot = expectedStateRoot
	} else {
		stateRoot = filepath.Clean(stateRoot)
	}
	if !samePath(stateRoot, expectedStateRoot) {
		return "", fmt.Errorf(
			"%s=%q is not supported for public entry; expected %q so request-compiler and runtime artifacts stay under one state root",
			stateRootEnv,
			stateRoot,
			expectedStateRoot,
		)
	}

	expectedArtifactRoot := filepath.Join(stateRoot, "artifacts")
	if err := validateDerivedStateRootEnv(runtimeworkbookcase.ArtifactRootEnv, expectedArtifactRoot); err != nil {
		return "", err
	}
	expectedKnowledgeRoot := filepath.Join(stateRoot, "knowledge")
	if err := validateDerivedStateRootEnv(runtimeknowledge.KnowledgeRootEnv, expectedKnowledgeRoot); err != nil {
		return "", err
	}

	if err := os.Setenv(stateRootEnv, stateRoot); err != nil {
		return "", err
	}
	if err := os.Setenv(runtimeworkbookcase.ArtifactRootEnv, expectedArtifactRoot); err != nil {
		return "", err
	}
	if err := os.Setenv(runtimeknowledge.KnowledgeRootEnv, expectedKnowledgeRoot); err != nil {
		return "", err
	}
	if strings.TrimSpace(os.Getenv(runtimeconfig.RetentionModeEnv)) == "" {
		if err := os.Setenv(runtimeconfig.RetentionModeEnv, string(runtimeconfig.RetentionModeRedacted)); err != nil {
			return "", err
		}
	}
	if strings.TrimSpace(os.Getenv(runtimeconfig.RenderModeEnv)) == "" {
		if err := os.Setenv(runtimeconfig.RenderModeEnv, string(runtimeconfig.RenderModeNever)); err != nil {
			return "", err
		}
	}
	return filepath.Dir(stateRoot), nil
}

func ensureSafeRuntimeDefaults() error {
	if strings.TrimSpace(os.Getenv(runtimeconfig.RetentionModeEnv)) == "" {
		if err := os.Setenv(runtimeconfig.RetentionModeEnv, string(runtimeconfig.RetentionModeRedacted)); err != nil {
			return err
		}
	}
	if strings.TrimSpace(os.Getenv(runtimeconfig.RenderModeEnv)) == "" {
		if err := os.Setenv(runtimeconfig.RenderModeEnv, string(runtimeconfig.RenderModeNever)); err != nil {
			return err
		}
	}
	return nil
}

func resolveWorkspaceRoot(inputFile string) string {
	return filepath.Dir(filepath.Clean(inputFile))
}

func validateDerivedStateRootEnv(envName, expectedValue string) error {
	currentValue := strings.TrimSpace(os.Getenv(envName))
	if currentValue == "" {
		return nil
	}
	if samePath(currentValue, expectedValue) {
		return nil
	}
	return fmt.Errorf(
		"%s=%q conflicts with %s; expected %q",
		envName,
		currentValue,
		stateRootEnv,
		expectedValue,
	)
}

func samePath(left, right string) bool {
	return comparablePath(left) == comparablePath(right)
}

func comparablePath(path string) string {
	cleaned := filepath.Clean(path)
	pending := make([]string, 0, 4)
	current := cleaned

	for {
		resolved, err := filepath.EvalSymlinks(current)
		if err == nil {
			for index := len(pending) - 1; index >= 0; index-- {
				resolved = filepath.Join(resolved, pending[index])
			}
			return resolved
		}
		parent := filepath.Dir(current)
		if parent == current {
			return cleaned
		}
		pending = append(pending, filepath.Base(current))
		current = parent
	}
}

func defaultScenarioSlug(explicit, sourcePath string) string {
	if trimmed := strings.TrimSpace(explicit); trimmed != "" {
		return trimmed
	}
	name := strings.TrimSpace(strings.TrimSuffix(filepath.Base(sourcePath), filepath.Ext(sourcePath)))
	if name != "" {
		return name
	}
	return "run-prompt"
}

func normalizedIntentRequestText(intent requestcompiler.NormalizedIntent) string {
	raw, err := json.Marshal(intent)
	if err != nil {
		return "normalized intent machine-boundary request"
	}
	return string(raw)
}

func currentWorkingDirOrPanic() string {
	workingDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return workingDir
}

func writeResultJSON(output io.Writer, result useorchestrator.RunResult) error {
	encoder := json.NewEncoder(output)
	encoder.SetIndent("", "  ")
	return encoder.Encode(result)
}

func runtimeStarted(result useorchestrator.RunResult) bool {
	return strings.TrimSpace(result.IDs.RunID) != "" || strings.TrimSpace(result.Paths.TelemetryDir) != ""
}
