package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	runtimebundle "github.com/mwroh/sheet-ops/runtime/bundle"
	requestmode "github.com/mwroh/sheet-ops/runtime/openlayer/requestmode"
	"github.com/mwroh/sheet-ops/runtime/openlayer/useorchestrator"
	runtimeschema "github.com/mwroh/sheet-ops/runtime/schema"
	runtimesession "github.com/mwroh/sheet-ops/runtime/session"
	"github.com/spf13/cobra"
)

const (
	caseLocalBundleRoot             = ".codex/skills/sheet-ops/agent-system"
	orchestratorDecisionFixtureEnv  = "SHEET_OPS_AGENT_ORCHESTRATOR_DECISION_FIXTURE"
	resultVerifierOutcomeFixtureEnv = "SHEET_OPS_AGENT_RESULT_VERIFIER_FIXTURE"
	requestRefSchemaRel             = "contracts/requests/request_ref.schema.json"
	legacyUseEnvelopeSchemaRel      = "contracts/requests/use_envelope.schema.json"
	useEnvelopeV2SchemaRel          = "contracts/requests/use_envelope_v2.schema.json"
)

var errLegacyStructuredRequestRef = errors.New("legacy request_ref envelopes must not point to structured request JSON; use the typed envelope instead")

type UseEnvelope struct {
	ScenarioID string                 `json:"scenario_id"`
	Request    requestmode.RequestRef `json:"request"`
	InputFile  string                 `json:"input_file"`
	OutputFile string                 `json:"output_file"`
}

type legacyUseEnvelope struct {
	ScenarioID string `json:"scenario_id"`
	RequestRef string `json:"request_ref"`
	InputFile  string `json:"input_file"`
	OutputFile string `json:"output_file"`
}

type useTerminalEvidence struct {
	Entry                string `json:"entry"`
	Status               string `json:"status"`
	ScenarioID           string `json:"scenario_id"`
	RunID                string `json:"run_id"`
	SessionID            string `json:"session_id"`
	TerminalState        string `json:"terminal_state"`
	Message              string `json:"message"`
	EnvelopePath         string `json:"envelope_path"`
	BundleRoot           string `json:"bundle_root"`
	RequestRef           string `json:"request_ref"`
	InputFile            string `json:"input_file"`
	OutputFile           string `json:"output_file"`
	ScenarioEvidencePath string `json:"scenario_evidence_path,omitempty"`
}

type codexLaunchConfig struct {
	Model           string
	ReasoningEffort string
}

func newUseCommand() *cobra.Command {
	var envelopePath string

	cmd := &cobra.Command{
		Use:   "use",
		Short: "Compatibility dispatcher for Sheet Ops use requests",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUseCommandWithSetup(cmd.OutOrStdout(), envelopePath, dispatchUseCompatibility)
		},
	}

	cmd.Flags().StringVar(&envelopePath, "envelope", "", "Path to a use-envelope JSON file")
	if err := cmd.MarkFlagRequired("envelope"); err != nil {
		panic(err)
	}

	return cmd
}

type useCommandContext struct {
	CaseRoot     string
	BundleRoot   string
	EnvelopePath string
	Envelope     UseEnvelope
	RunSession   runtimesession.RunSession
	Contract     runtimebundle.ModelContract
}

type useCommandExecutor func(io.Writer, useCommandContext) error

func runUseCommandWithSetup(output io.Writer, envelopePath string, executor useCommandExecutor) error {
	caseRoot, err := os.Getwd()
	if err != nil {
		return err
	}

	bundleRoot := filepath.Join(caseRoot, caseLocalBundleRoot)
	terminalState, message := classifyBundleRoot(caseRoot, bundleRoot)
	if terminalState != "" {
		envelope := bestEffortEnvelopeForFailure(envelopePath)
		result, err := buildInstallFailureResult(envelopePath, bundleRoot, envelope, terminalState, message)
		if err != nil {
			return err
		}
		return writeUseTerminalResultJSON(output, result)
	}

	envelope, err := loadUseEnvelope(envelopePath)
	if err != nil {
		return err
	}

	runSession, err := runtimesession.OpenSession(caseRoot, envelope.ScenarioID)
	if err != nil {
		return err
	}

	if err := runtimebundle.RequireInstallRoot(caseRoot, bundleRoot); err != nil {
		return writeRunFailureResult(output, runSession, envelopePath, bundleRoot, envelope, "INSTALL_INVALID", err.Error())
	}

	contract, err := runtimebundle.LoadModelContract(bundleModelContractPath(bundleRoot))
	if err != nil {
		return writeRunFailureResult(output, runSession, envelopePath, bundleRoot, envelope, "INSTALL_INVALID", err.Error())
	}

	if _, err := runtimesession.WriteProof(runSession, "install-proof.json", runtimebundle.InstallProof{
		BundleRoot:        bundleRoot,
		ManifestPath:      filepath.Join(bundleRoot, "bundle.manifest.json"),
		ModelContractPath: bundleModelContractPath(bundleRoot),
		Status:            "validated",
	}); err != nil {
		return err
	}

	return executor(output, useCommandContext{
		CaseRoot:     caseRoot,
		BundleRoot:   bundleRoot,
		EnvelopePath: envelopePath,
		Envelope:     envelope,
		RunSession:   runSession,
		Contract:     contract,
	})
}

func dispatchUseCompatibility(output io.Writer, ctx useCommandContext) error {
	switch ctx.Envelope.Request.Kind {
	case requestmode.ModeStructuredUseRequest:
		structuredRequest, err := loadStructuredUseRequest(ctx.Envelope)
		if err != nil {
			return writeRunFailureResult(
				output,
				ctx.RunSession,
				ctx.EnvelopePath,
				ctx.BundleRoot,
				ctx.Envelope,
				"REQUEST_INVALID",
				err.Error(),
			)
		}
		return executeUseStructured(output, ctx, structuredRequest)
	case requestmode.ModeOrganismExecutionRequest:
		organismRequest, err := loadOrganismExecutionRequest(ctx.Envelope)
		if err != nil {
			return writeRunFailureResult(
				output,
				ctx.RunSession,
				ctx.EnvelopePath,
				ctx.BundleRoot,
				ctx.Envelope,
				"REQUEST_INVALID",
				err.Error(),
			)
		}
		return executeUseOrganism(output, ctx, organismRequest)
	case requestmode.ModePromptText:
		return executeUseOpen(output, ctx)
	default:
		return writeRunFailureResult(
			output,
			ctx.RunSession,
			ctx.EnvelopePath,
			ctx.BundleRoot,
			ctx.Envelope,
			"REQUEST_INVALID",
			fmt.Sprintf("unsupported request.kind %q", ctx.Envelope.Request.Kind),
		)
	}
}

func loadUseEnvelope(path string) (UseEnvelope, error) {
	var envelope UseEnvelope

	raw, err := os.ReadFile(path)
	if err != nil {
		return envelope, err
	}

	var document any
	if err := json.Unmarshal(raw, &document); err != nil {
		return envelope, err
	}
	if err := validateUseEnvelopeValue(document); err != nil {
		legacyEnvelope, legacyErr := loadLegacyUseEnvelopeCompatibility(raw)
		if legacyErr == nil {
			return legacyEnvelope, nil
		}
		if errors.Is(legacyErr, errLegacyStructuredRequestRef) {
			return envelope, legacyErr
		}
		return envelope, err
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return envelope, err
	}

	return envelope, nil
}

func loadUseEnvelopeBestEffort(path string) (UseEnvelope, error) {
	var envelope UseEnvelope

	raw, err := os.ReadFile(path)
	if err != nil {
		return envelope, err
	}

	if err := json.Unmarshal(raw, &envelope); err != nil {
		return UseEnvelope{}, err
	}
	if envelope.ScenarioID != "" && envelope.Request.Path != "" {
		return envelope, nil
	}
	if legacyEnvelope, err := loadLegacyUseEnvelopeCompatibility(raw); err == nil {
		return legacyEnvelope, nil
	}

	return envelope, nil
}

func loadLegacyUseEnvelopeCompatibility(raw []byte) (UseEnvelope, error) {
	var document any
	if err := json.Unmarshal(raw, &document); err != nil {
		return UseEnvelope{}, err
	}
	if err := validateLegacyUseEnvelopeValue(document); err != nil {
		return UseEnvelope{}, err
	}

	var legacy legacyUseEnvelope
	if err := json.Unmarshal(raw, &legacy); err != nil {
		return UseEnvelope{}, err
	}
	if err := rejectLegacyStructuredRequestRef(legacy.RequestRef); err != nil {
		return UseEnvelope{}, err
	}
	return UseEnvelope{
		ScenarioID: legacy.ScenarioID,
		Request: requestmode.RequestRef{
			Kind: requestmode.ModePromptText,
			Path: legacy.RequestRef,
		},
		InputFile:  legacy.InputFile,
		OutputFile: legacy.OutputFile,
	}, nil
}

func rejectLegacyStructuredRequestRef(path string) error {
	judgment, err := requestmode.JudgeRequestRef(requestmode.RequestRef{
		Kind: requestmode.ModeStructuredUseRequest,
		Path: path,
	})
	if err != nil {
		return nil
	}
	if judgment.RequestMode == requestmode.ModeStructuredUseRequest {
		return fmt.Errorf("%w: %s", errLegacyStructuredRequestRef, path)
	}
	return nil
}

func bestEffortEnvelopeForFailure(path string) UseEnvelope {
	envelope, err := loadUseEnvelopeBestEffort(path)
	if err == nil && envelope.ScenarioID != "" {
		return envelope
	}
	return UseEnvelope{
		ScenarioID: "unknown-scenario",
	}
}

func validateUseEnvelopeValue(value any) error {
	schemaPath, err := executableRelativeUseEnvelopeSchemaPath()
	if err != nil {
		return err
	}
	return runtimeschema.ValidateStruct(schemaPath, value)
}

func validateLegacyUseEnvelopeValue(value any) error {
	schemaPath, err := executableRelativeLegacyUseEnvelopeSchemaPath()
	if err != nil {
		return err
	}
	return runtimeschema.ValidateStruct(schemaPath, value)
}

func executableRelativeUseEnvelopeSchemaPath() (string, error) {
	return executableRelativeRequestSchemaPath(filepath.Base(useEnvelopeV2SchemaRel))
}

func executableRelativeLegacyUseEnvelopeSchemaPath() (string, error) {
	return executableRelativeRequestSchemaPath(filepath.Base(legacyUseEnvelopeSchemaRel))
}

func executableRelativeRequestSchemaPath(name string) (string, error) {
	executablePath, err := os.Executable()
	if err != nil {
		return "", err
	}
	candidate := filepath.Clean(filepath.Join(filepath.Dir(executablePath), "..", "contracts", "requests", name))
	if _, err := os.Stat(candidate); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("bundled use-envelope contract is missing at %s", candidate)
		}
		return "", fmt.Errorf("could not inspect bundled use-envelope contract at %s: %v", candidate, err)
	}
	return candidate, nil
}

func classifyBundleRoot(caseRoot string, path string) (string, string) {
	info, err := os.Lstat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "INSTALL_INVALID", fmt.Sprintf("case-local bundle is missing at %s; install state is not available", path)
		}
		return "INSTALL_UNAVAILABLE", fmt.Sprintf("could not inspect case-local bundle at %s: %v", path, err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return "INSTALL_INVALID", fmt.Sprintf("case-local bundle path %s exists but is a symlink; symlink bundle roots are not allowed", path)
	}
	if !info.IsDir() {
		return "INSTALL_INVALID", fmt.Sprintf("case-local bundle path %s exists but is not a directory", path)
	}

	resolvedCaseRoot, err := filepath.EvalSymlinks(caseRoot)
	if err != nil {
		return "INSTALL_UNAVAILABLE", fmt.Sprintf("could not resolve case root %s: %v", caseRoot, err)
	}
	resolvedBundleRoot, err := filepath.EvalSymlinks(path)
	if err != nil {
		return "INSTALL_UNAVAILABLE", fmt.Sprintf("could not resolve case-local bundle at %s: %v", path, err)
	}

	expectedBundleRoot := comparablePath(filepath.Join(resolvedCaseRoot, caseLocalBundleRoot))
	if resolvedBundleRoot != expectedBundleRoot {
		return "INSTALL_INVALID", fmt.Sprintf("case-local bundle path %s resolves to %s; expected %s and does not allow symlink escape", path, resolvedBundleRoot, expectedBundleRoot)
	}
	requiredBundleEntries := []string{
		"bundle.manifest.json",
		"model-contract.json",
		filepath.FromSlash(requestRefSchemaRel),
		filepath.Join("contracts", "requests", "organism_execution_request.schema.json"),
		filepath.FromSlash(legacyUseEnvelopeSchemaRel),
		filepath.FromSlash(useEnvelopeV2SchemaRel),
		filepath.Join("prompts", "orchestrator-use.md"),
		filepath.Join("prompts", "request-compiler.md"),
		filepath.Join("prompts", "result-verifier.md"),
		filepath.Join("prompts", "repair-advisor.md"),
	}
	for _, entry := range requiredBundleEntries {
		entryPath := filepath.Join(path, entry)
		entryInfo, err := os.Lstat(entryPath)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return "INSTALL_INVALID", fmt.Sprintf("case-local bundle is incomplete at %s; missing %s", path, entryPath)
			}
			return "INSTALL_UNAVAILABLE", fmt.Sprintf("could not inspect required bundle entry %s: %v", entryPath, err)
		}
		if entryInfo.Mode()&os.ModeSymlink != 0 {
			return "INSTALL_INVALID", fmt.Sprintf("case-local bundle entry %s is a symlink; symlink bundle entries are not allowed", entryPath)
		}
		if entryInfo.IsDir() {
			return "INSTALL_INVALID", fmt.Sprintf("case-local bundle entry %s must be a file", entryPath)
		}
		resolvedEntryPath, err := filepath.EvalSymlinks(entryPath)
		if err != nil {
			return "INSTALL_UNAVAILABLE", fmt.Sprintf("could not resolve required bundle entry %s: %v", entryPath, err)
		}
		expectedEntryPath := comparablePath(entryPath)
		if resolvedEntryPath != expectedEntryPath {
			return "INSTALL_INVALID", fmt.Sprintf("case-local bundle entry %s resolves to %s; expected %s and does not allow symlink escape", entryPath, resolvedEntryPath, expectedEntryPath)
		}
	}

	return "", ""
}

func comparablePath(path string) string {
	cleaned := filepath.Clean(path)
	resolved, err := filepath.EvalSymlinks(cleaned)
	if err == nil {
		return resolved
	}
	return cleaned
}

func samePath(left string, right string) bool {
	return comparablePath(left) == comparablePath(right)
}

func bundleModelContractPath(bundleRoot string) string {
	return filepath.Join(bundleRoot, "model-contract.json")
}

func buildInstallFailureResult(
	envelopePath string,
	bundleRoot string,
	envelope UseEnvelope,
	terminalState string,
	message string,
) (UseTerminalResult, error) {
	suffix := time.Now().UTC().Format("20060102T150405.000000000Z")
	runID := "run-" + suffix
	sessionID := "session-" + suffix
	evidencePath := filepath.Join(filepath.Dir(envelopePath), runID+"__use-terminal-evidence.json")

	evidence := useTerminalEvidence{
		Entry:         "use",
		Status:        "failed",
		ScenarioID:    envelope.ScenarioID,
		RunID:         runID,
		SessionID:     sessionID,
		TerminalState: terminalState,
		Message:       message,
		EnvelopePath:  envelopePath,
		BundleRoot:    bundleRoot,
		RequestRef:    envelope.Request.Path,
		InputFile:     envelope.InputFile,
		OutputFile:    envelope.OutputFile,
	}
	if err := writeUseTerminalEvidence(evidencePath, evidence); err != nil {
		return UseTerminalResult{}, err
	}

	return newFailedUseTerminalResult(
		envelope.ScenarioID,
		runID,
		sessionID,
		terminalState,
		message,
		evidencePath,
	), nil
}

func buildRunSuccessResult(
	runSession runtimesession.RunSession,
	envelopePath string,
	bundleRoot string,
	envelope UseEnvelope,
	runtimeResult useorchestrator.RunResult,
) (UseTerminalResult, error) {
	evidencePath := filepath.Join(runSession.EvidenceDir, "use-terminal-evidence.json")
	outputFile := runtimeResult.OutputFile(envelope.OutputFile)
	evidence := useTerminalEvidence{
		Entry:                "use",
		Status:               "succeeded",
		ScenarioID:           runSession.ScenarioID,
		RunID:                runSession.RunID,
		SessionID:            runSession.SessionID,
		TerminalState:        "TERMINAL_SUCCESS",
		Message:              "sheet-ops agent-system run completed",
		EnvelopePath:         envelopePath,
		BundleRoot:           bundleRoot,
		RequestRef:           envelope.Request.Path,
		InputFile:            envelope.InputFile,
		OutputFile:           outputFile,
		ScenarioEvidencePath: runtimeResult.Paths.EvidenceIndex,
	}
	if err := writeUseTerminalEvidence(evidencePath, evidence); err != nil {
		return UseTerminalResult{}, err
	}
	return newSucceededUseTerminalResult(
		runSession.ScenarioID,
		runSession.RunID,
		runSession.SessionID,
		evidence.Message,
		outputFile,
		evidencePath,
		runtimeResult.IDs.RunID,
	), nil
}

func writeRunFailureResult(
	output io.Writer,
	runSession runtimesession.RunSession,
	envelopePath string,
	bundleRoot string,
	envelope UseEnvelope,
	terminalState string,
	message string,
) error {
	evidencePath := filepath.Join(runSession.EvidenceDir, "use-terminal-evidence.json")
	evidence := useTerminalEvidence{
		Entry:         "use",
		Status:        "failed",
		ScenarioID:    runSession.ScenarioID,
		RunID:         runSession.RunID,
		SessionID:     runSession.SessionID,
		TerminalState: terminalState,
		Message:       message,
		EnvelopePath:  envelopePath,
		BundleRoot:    bundleRoot,
		RequestRef:    envelope.Request.Path,
		InputFile:     envelope.InputFile,
		OutputFile:    envelope.OutputFile,
	}
	if err := writeUseTerminalEvidence(evidencePath, evidence); err != nil {
		return err
	}

	return writeUseTerminalResultJSON(output, newFailedUseTerminalResult(
		runSession.ScenarioID,
		runSession.RunID,
		runSession.SessionID,
		terminalState,
		message,
		evidencePath,
	))
}

func writeRuntimeFailureResult(
	output io.Writer,
	runSession runtimesession.RunSession,
	envelopePath string,
	bundleRoot string,
	envelope UseEnvelope,
	runtimeResult useorchestrator.RunResult,
	terminalState string,
	message string,
) error {
	evidencePath := filepath.Join(runSession.EvidenceDir, "use-terminal-evidence.json")
	outputFile := runtimeResult.OutputFile(envelope.OutputFile)
	evidence := useTerminalEvidence{
		Entry:                "use",
		Status:               "failed",
		ScenarioID:           runSession.ScenarioID,
		RunID:                runSession.RunID,
		SessionID:            runSession.SessionID,
		TerminalState:        terminalState,
		Message:              message,
		EnvelopePath:         envelopePath,
		BundleRoot:           bundleRoot,
		RequestRef:           envelope.Request.Path,
		InputFile:            envelope.InputFile,
		OutputFile:           outputFile,
		ScenarioEvidencePath: runtimeResult.Paths.EvidenceIndex,
	}
	if err := writeUseTerminalEvidence(evidencePath, evidence); err != nil {
		return err
	}
	return writeUseTerminalResultJSON(output, newFailedUseTerminalResult(
		runSession.ScenarioID,
		runSession.RunID,
		runSession.SessionID,
		terminalState,
		message,
		evidencePath,
		runtimeResult.IDs.RunID,
	))
}

func writeUseTerminalEvidence(path string, evidence useTerminalEvidence) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	raw, err := json.MarshalIndent(evidence, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, append(raw, '\n'), 0o644)
}

func blockedCompilerMessage(decision useorchestrator.OrchestratorDecision) string {
	if decision.RepairAdvice != nil && strings.TrimSpace(decision.RepairAdvice.Summary) != "" {
		return decision.RepairAdvice.Summary
	}
	return "request-compiler blocked delegated execution"
}

func orchestratorFailureState(err error) string {
	if err == nil {
		return "ORCHESTRATOR_FAILED"
	}
	lower := strings.ToLower(err.Error())
	if strings.Contains(lower, "request_compiler_loop_state") ||
		strings.Contains(lower, "request-compiler") ||
		strings.Contains(lower, "specialist lifecycle proof") ||
		strings.Contains(lower, "validated_execution_request") {
		return "BLOCKED_AT_REQUEST_COMPILER"
	}
	return "ORCHESTRATOR_FAILED"
}

func failureReviewMessage(status string, summary string) string {
	if strings.TrimSpace(summary) != "" {
		return summary
	}
	return "result-verifier returned " + status
}

func forceCaseLocalRuntimeRoots(caseRoot string) (func(), error) {
	artifactKey := "SHEET_OPS_ARTIFACT_ROOT"
	knowledgeKey := "SHEET_OPS_KNOWLEDGE_ROOT"
	previousArtifact, hadArtifact := os.LookupEnv(artifactKey)
	previousKnowledge, hadKnowledge := os.LookupEnv(knowledgeKey)
	restore := func() {
		restoreEnv(artifactKey, previousArtifact, hadArtifact)
		restoreEnv(knowledgeKey, previousKnowledge, hadKnowledge)
	}
	stateRoot := runtimesession.StateRoot(caseRoot)
	if err := os.Setenv(artifactKey, filepath.Join(stateRoot, "artifacts")); err != nil {
		restore()
		return nil, err
	}
	if err := os.Setenv(knowledgeKey, filepath.Join(stateRoot, "knowledge")); err != nil {
		restore()
		return nil, err
	}
	return restore, nil
}

func restoreEnv(key string, value string, hadValue bool) {
	if hadValue {
		_ = os.Setenv(key, value)
		return
	}
	_ = os.Unsetenv(key)
}
