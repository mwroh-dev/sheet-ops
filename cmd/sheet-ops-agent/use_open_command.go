package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	runtimebundle "github.com/mwroh/sheet-ops/runtime/bundle"
	runtimeinspect "github.com/mwroh/sheet-ops/runtime/inspect"
	requestcompiler "github.com/mwroh/sheet-ops/runtime/openlayer/requestcompiler"
	requestmode "github.com/mwroh/sheet-ops/runtime/openlayer/requestmode"
	"github.com/mwroh/sheet-ops/runtime/openlayer/useorchestrator"
	runtimesession "github.com/mwroh/sheet-ops/runtime/session"
	runtimesubagent "github.com/mwroh/sheet-ops/runtime/subagent"
	"github.com/spf13/cobra"
)

func newUseOpenCommand() *cobra.Command {
	var envelopePath string

	cmd := &cobra.Command{
		Use:   "use-open",
		Short: "Compile an open-layer use request, execute it, and verify with the model-gated path",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUseCommandWithSetup(cmd.OutOrStdout(), envelopePath, executeUseOpen)
		},
	}

	cmd.Flags().StringVar(&envelopePath, "envelope", "", "Path to a use-envelope JSON file with request.kind=prompt_text")
	if err := cmd.MarkFlagRequired("envelope"); err != nil {
		panic(err)
	}
	return cmd
}

func executeUseOpen(output io.Writer, ctx useCommandContext) error {
	if ctx.Envelope.Request.Kind != requestmode.ModePromptText {
		return writeRunFailureResult(
			output,
			ctx.RunSession,
			ctx.EnvelopePath,
			ctx.BundleRoot,
			ctx.Envelope,
			"REQUEST_INVALID",
			fmt.Sprintf("use-open requires request.kind %q, got %q", requestmode.ModePromptText, ctx.Envelope.Request.Kind),
		)
	}

	decision, err := loadOrRunOrchestratorDecision(ctx.CaseRoot, ctx.BundleRoot, ctx.RunSession, ctx.Envelope, ctx.Contract)
	if err != nil {
		return writeRunFailureResult(
			output,
			ctx.RunSession,
			ctx.EnvelopePath,
			ctx.BundleRoot,
			ctx.Envelope,
			orchestratorFailureState(err),
			err.Error(),
		)
	}

	if _, err := runtimesession.WriteProof(ctx.RunSession, "orchestrator-decision.json", decision); err != nil {
		return err
	}
	if err := runtimesubagent.RequireSpecialistLoopState(decision.RequestCompilerLoopState); err != nil {
		return writeRunFailureResult(output, ctx.RunSession, ctx.EnvelopePath, ctx.BundleRoot, ctx.Envelope, "BLOCKED_AT_REQUEST_COMPILER", err.Error())
	}
	if err := ensureLoopStateMatchesRoleContract(decision.RequestCompilerLoopState, "request-compiler", ctx.Contract.RequestCompiler); err != nil {
		return writeRunFailureResult(output, ctx.RunSession, ctx.EnvelopePath, ctx.BundleRoot, ctx.Envelope, "BLOCKED_AT_REQUEST_COMPILER", err.Error())
	}
	if decision.Decision != "execute" || decision.ValidatedExecutionRequest == nil {
		return writeRunFailureResult(
			output,
			ctx.RunSession,
			ctx.EnvelopePath,
			ctx.BundleRoot,
			ctx.Envelope,
			"BLOCKED_AT_REQUEST_COMPILER",
			blockedCompilerMessage(decision),
		)
	}

	runtimeResult, err := runCaseLocalRuntime(ctx.CaseRoot, *decision.ValidatedExecutionRequest)
	if err != nil {
		return writeRuntimeFailureResult(output, ctx.RunSession, ctx.EnvelopePath, ctx.BundleRoot, ctx.Envelope, runtimeResult, "RUNTIME_FAILED", err.Error())
	}

	verifierOutcome, err := loadOrRunResultVerifierOutcome(ctx.CaseRoot, ctx.BundleRoot, ctx.RunSession, ctx.Envelope, ctx.Contract, runtimeResult)
	if err != nil {
		return writeRuntimeFailureResult(output, ctx.RunSession, ctx.EnvelopePath, ctx.BundleRoot, ctx.Envelope, runtimeResult, "BLOCKED_AT_RESULT_VERIFIER", err.Error())
	}
	if _, err := runtimesession.WriteProof(ctx.RunSession, "result-verifier-review.json", verifierOutcome.Review); err != nil {
		return err
	}
	if _, err := runtimesession.WriteProof(ctx.RunSession, "result-verifier-loop-state.json", verifierOutcome.LoopState); err != nil {
		return err
	}
	if err := runtimesubagent.RequireSpecialistLoopState(verifierOutcome.LoopState); err != nil {
		return writeRuntimeFailureResult(output, ctx.RunSession, ctx.EnvelopePath, ctx.BundleRoot, ctx.Envelope, runtimeResult, "BLOCKED_AT_RESULT_VERIFIER", err.Error())
	}
	if err := ensureLoopStateMatchesRoleContract(verifierOutcome.LoopState, "result-verifier", ctx.Contract.ResultVerifier); err != nil {
		return writeRuntimeFailureResult(output, ctx.RunSession, ctx.EnvelopePath, ctx.BundleRoot, ctx.Envelope, runtimeResult, "BLOCKED_AT_RESULT_VERIFIER", err.Error())
	}
	if verifierOutcome.Review.Status != "pass" {
		return writeRuntimeFailureResult(
			output,
			ctx.RunSession,
			ctx.EnvelopePath,
			ctx.BundleRoot,
			ctx.Envelope,
			runtimeResult,
			"BLOCKED_AT_RESULT_VERIFIER",
			failureReviewMessage(verifierOutcome.Review.Status, verifierOutcome.Review.Summary),
		)
	}

	result, err := buildRunSuccessResult(ctx.RunSession, ctx.EnvelopePath, ctx.BundleRoot, ctx.Envelope, runtimeResult)
	if err != nil {
		return err
	}
	return writeUseTerminalResultJSON(output, result)
}

func loadOrRunOrchestratorDecision(
	caseRoot string,
	bundleRoot string,
	runSession runtimesession.RunSession,
	envelope UseEnvelope,
	contract runtimebundle.ModelContract,
) (useorchestrator.OrchestratorDecision, error) {
	if fixturePath := strings.TrimSpace(os.Getenv(orchestratorDecisionFixtureEnv)); fixturePath != "" {
		decision, err := useorchestrator.LoadOrchestratorDecision(fixturePath)
		if err != nil {
			return decision, err
		}
		return decision, bindValidatedExecutionRequestToEnvelope(&decision, envelope)
	}

	outputPath := filepath.Join(runSession.OutputsDir, "orchestrator-decision.json")
	schemaPath := filepath.Join(runSession.OutputsDir, "orchestrator-decision-output.schema.json")
	if err := writeOrchestratorDecisionOutputSchema(schemaPath); err != nil {
		return useorchestrator.OrchestratorDecision{}, err
	}
	requestText, err := loadRequestText(envelope.Request.Path)
	if err != nil {
		return useorchestrator.OrchestratorDecision{}, err
	}
	workbookFacts, err := loadWorkbookFacts(envelope.InputFile)
	if err != nil {
		return useorchestrator.OrchestratorDecision{}, err
	}
	orchestratorPrompt, err := loadRolePrompt(bundleRoot, contract.Orchestrator.Prompt)
	if err != nil {
		return useorchestrator.OrchestratorDecision{}, err
	}
	requestCompilerPrompt, err := loadRolePrompt(bundleRoot, contract.RequestCompiler.Prompt)
	if err != nil {
		return useorchestrator.OrchestratorDecision{}, err
	}
	prompt := buildBundledOrchestratorPrompt(bundleRoot, envelope, contract, orchestratorPrompt, requestCompilerPrompt, requestText, workbookFacts)
	requestCompilerLaunchConfig := codexLaunchConfigFromRole(contract.RequestCompiler)
	jsonlPath, err := runCodexPrompt(caseRoot, outputPath, schemaPath, contract.Orchestrator, prompt)
	if err != nil {
		return useorchestrator.OrchestratorDecision{}, err
	}
	decision, err := useorchestrator.LoadOrchestratorDecision(outputPath)
	if err != nil {
		decision, err := recoverOrchestratorDecision(outputPath, envelope, requestText)
		if err != nil {
			return useorchestrator.OrchestratorDecision{}, err
		}
		decision.RequestCompilerLoopState = observedLoopStateFromJSONL(jsonlPath, envelope.ScenarioID, "request-compiler", contract.RequestCompiler, requestCompilerLaunchConfig, decision.Decision == "execute")
		return decision, nil
	}
	if err := bindValidatedExecutionRequestToEnvelope(&decision, envelope); err != nil {
		return useorchestrator.OrchestratorDecision{}, err
	}
	if decision.ValidatedExecutionRequest != nil && strings.TrimSpace(decision.ValidatedExecutionRequest.RequestText) == "" {
		decision.ValidatedExecutionRequest.RequestText = requestText
	}
	decision.RequestCompilerLoopState = observedLoopStateFromJSONL(jsonlPath, envelope.ScenarioID, "request-compiler", contract.RequestCompiler, requestCompilerLaunchConfig, decision.Decision == "execute")
	return decision, nil
}

func loadOrRunResultVerifierOutcome(
	caseRoot string,
	bundleRoot string,
	runSession runtimesession.RunSession,
	envelope UseEnvelope,
	contract runtimebundle.ModelContract,
	runtimeResult useorchestrator.RunResult,
) (useorchestrator.ResultVerifierOutcome, error) {
	if fixturePath := strings.TrimSpace(os.Getenv(resultVerifierOutcomeFixtureEnv)); fixturePath != "" {
		return useorchestrator.LoadResultVerifierOutcomeWithScenarioID(fixturePath, envelope.ScenarioID)
	}

	outputPath := filepath.Join(runSession.OutputsDir, "result-verifier-outcome.json")
	schemaPath := filepath.Join(runSession.OutputsDir, "result-verifier-output.schema.json")
	if err := writeResultVerifierOutcomeSchema(schemaPath); err != nil {
		return useorchestrator.ResultVerifierOutcome{}, err
	}
	runtimeSummary, err := buildRuntimeSummary(runtimeResult)
	if err != nil {
		return useorchestrator.ResultVerifierOutcome{}, err
	}
	resultVerifierPrompt, err := loadRolePrompt(bundleRoot, contract.ResultVerifier.Prompt)
	if err != nil {
		return useorchestrator.ResultVerifierOutcome{}, err
	}
	prompt := buildResultVerifierPrompt(bundleRoot, envelope, contract, runtimeResult, resultVerifierPrompt, runtimeSummary)
	launchConfig := codexLaunchConfigFromRole(contract.ResultVerifier)
	jsonlPath, err := runCodexPrompt(caseRoot, outputPath, schemaPath, contract.ResultVerifier, prompt)
	if err != nil {
		return useorchestrator.ResultVerifierOutcome{}, err
	}
	outcome, err := useorchestrator.LoadResultVerifierOutcomeWithScenarioID(outputPath, envelope.ScenarioID)
	if err != nil {
		return useorchestrator.ResultVerifierOutcome{}, err
	}
	outcome.LoopState = observedLoopStateFromJSONL(jsonlPath, envelope.ScenarioID, "result-verifier", contract.ResultVerifier, launchConfig, outcome.Review.Status == "pass")
	return outcome, nil
}

func runCodexPrompt(caseRoot string, outputPath string, schemaPath string, contract runtimebundle.RoleContract, prompt string) (string, error) {
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return "", err
	}
	jsonlPath := outputPath + ".jsonl"
	jsonlFile, err := os.Create(jsonlPath)
	if err != nil {
		return "", err
	}
	defer jsonlFile.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	defer cancel()

	args := []string{
		"exec",
		"--json",
		"--enable",
		"multi_agent",
		"--sandbox",
		"read-only",
	}
	if model := explicitModelOverride(contract.Model); model != "" {
		args = append(args, "--model", model)
	}
	args = append(args,
		"--config",
		fmt.Sprintf("model_reasoning_effort=%q", contract.ReasoningEffort),
		"-C",
		caseRoot,
		"--output-schema",
		schemaPath,
		"--output-last-message",
		outputPath,
		prompt,
	)
	cmd := exec.CommandContext(ctx, "codex", args...)
	var stderr strings.Builder
	cmd.Stdout = jsonlFile
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return "", fmt.Errorf("codex exec timed out after 180s")
		}
		return "", fmt.Errorf("codex exec failed: %w: %s", err, strings.TrimSpace(stderr.String()))
	}
	return jsonlPath, nil
}

func codexLaunchConfigFromRole(contract runtimebundle.RoleContract) codexLaunchConfig {
	return codexLaunchConfig{
		Model:           launchModelLabel(contract.Model),
		ReasoningEffort: contract.ReasoningEffort,
	}
}

func explicitModelOverride(model string) string {
	trimmed := strings.TrimSpace(model)
	if trimmed == "" || strings.EqualFold(trimmed, "default") {
		return ""
	}
	return trimmed
}

func launchModelLabel(model string) string {
	if explicit := explicitModelOverride(model); explicit != "" {
		return explicit
	}
	return "codex-default"
}

func buildBundledOrchestratorPrompt(
	bundleRoot string,
	envelope UseEnvelope,
	contract runtimebundle.ModelContract,
	orchestratorPrompt string,
	requestCompilerPrompt string,
	requestText string,
	workbookFacts useorchestrator.WorkbookFacts,
) string {
	return fmt.Sprintf(
		"Spawn one built-in %s subagent for request compilation. Give it this role contract summary:\n%s\n\nUse only these supplied facts; do not read additional workspace files.\nScenario: %q\nInput workbook path: %q\nOutput workbook path: %q\nRequest text: %s\nVisible sheets: %v\nHeader rows by sheet: %v\n\nAfter the child returns, wait for it and close it. Then return only one JSON object with keys:\n- scenario_id\n- decision\n- request_compiler_loop_state\n- validated_execution_request\n- repair_advice\n\nRules:\n- scenario_id must equal %q\n- if executable, decision=\"execute\" and validated_execution_request must use the exact provided input/output workbook paths\n- for summary requests, validated_execution_request.composition_kind=\"group_summary\" and include target_sheet, summary_mode, optional filters, group_by, metrics\n- for highlight requests, validated_execution_request.composition_kind=\"threshold_highlight\" and include target_column, operator, threshold, and optional highlight_color\n- for lookup/join requests, validated_execution_request.composition_kind=\"join_lookup\" and include target_sheet, lookup_sheet, join_key, optional include_source_columns, append_lookup_columns\n- if blocked or ambiguous, decision=\"blocked\" and repair_advice must be non-null with summary, suggested_actions, and assumptions\n- request_compiler_loop_state may be either the full loop-state object or a single specialist lifecycle object for the request-compiler role\n- return JSON only",
		contract.RequestCompiler.Carrier,
		roleContractSummary(contract.RequestCompiler, requestCompilerPrompt),
		envelope.ScenarioID,
		envelope.InputFile,
		envelope.OutputFile,
		requestText,
		workbookFacts.VisibleSheetNames,
		workbookFacts.HeaderRowBySheet,
		envelope.ScenarioID,
	)
}

func buildResultVerifierPrompt(
	bundleRoot string,
	envelope UseEnvelope,
	contract runtimebundle.ModelContract,
	runtimeResult useorchestrator.RunResult,
	resultVerifierPrompt string,
	runtimeSummary runtimeSummary,
) string {
	return fmt.Sprintf(
		"Spawn one built-in %s subagent for result verification. Give it this role contract summary:\n%s\n\nUse only this deterministic runtime summary; do not read additional workspace files.\nScenario: %q\nOutput file: %q\nVerification pass: %t\nVerification operation: %q\nVerification reasons: %v\n\nAfter the child returns, wait for it and close it. Then return only one JSON object with keys `review` and `loop_state`. `loop_state` may be either the full loop-state object or a single specialist lifecycle object for the result-verifier role. Return JSON only.",
		contract.ResultVerifier.Carrier,
		roleContractSummary(contract.ResultVerifier, resultVerifierPrompt),
		envelope.ScenarioID,
		runtimeSummary.OutputFile,
		runtimeSummary.VerificationPass,
		runtimeSummary.VerificationOperation,
		runtimeSummary.VerificationReasons,
	)
}

func roleContractSummary(contract runtimebundle.RoleContract, prompt string) string {
	return fmt.Sprintf(
		"model: %s\nreasoning_effort: %s\ncarrier: %s\nfallback: %s\nprompt:\n%s",
		contract.Model,
		contract.ReasoningEffort,
		contract.Carrier,
		contract.Fallback,
		strings.TrimSpace(prompt),
	)
}

func loadRolePrompt(bundleRoot string, relativePath string) (string, error) {
	cleanRelativePath := filepath.Clean(relativePath)
	if filepath.IsAbs(cleanRelativePath) || cleanRelativePath == ".." || strings.HasPrefix(cleanRelativePath, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("role prompt path %q must stay within bundle root", relativePath)
	}
	path := filepath.Join(bundleRoot, cleanRelativePath)
	info, err := os.Lstat(path)
	if err != nil {
		return "", fmt.Errorf("inspect role prompt %s: %w", path, err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return "", fmt.Errorf("role prompt %s must not be a symlink", path)
	}
	resolvedPath, err := filepath.EvalSymlinks(path)
	if err != nil {
		return "", fmt.Errorf("resolve role prompt %s: %w", path, err)
	}
	expectedPath := comparablePath(path)
	if resolvedPath != expectedPath {
		return "", fmt.Errorf("role prompt %s resolves to %s; expected %s", path, resolvedPath, expectedPath)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read role prompt %s: %w", path, err)
	}
	return strings.TrimSpace(string(raw)), nil
}

func ensureLoopStateMatchesRoleContract(
	loopState runtimesubagent.SubagentLoopState,
	role string,
	contract runtimebundle.RoleContract,
) error {
	for _, specialist := range loopState.Specialists {
		if specialist.Role != role {
			continue
		}
		if specialist.Carrier != contract.Carrier {
			return fmt.Errorf("%s carrier %q does not match contract %q", role, specialist.Carrier, contract.Carrier)
		}
		if explicit := explicitModelOverride(contract.Model); explicit != "" && specialist.Model != explicit {
			return fmt.Errorf("%s model %q does not match contract %q", role, specialist.Model, contract.Model)
		}
		if specialist.ReasoningEffort != contract.ReasoningEffort {
			return fmt.Errorf("%s reasoning_effort %q does not match contract %q", role, specialist.ReasoningEffort, contract.ReasoningEffort)
		}
		return nil
	}
	return fmt.Errorf("%s lifecycle proof is missing matching specialist entry", role)
}

func writeOrchestratorDecisionOutputSchema(path string) error {
	return writeSchemaFile(path, map[string]any{
		"type": "object",
		"required": []string{
			"scenario_id",
			"decision",
			"request_compiler_loop_state",
			"validated_execution_request",
			"repair_advice",
		},
		"properties": map[string]any{
			"scenario_id":                 map[string]any{"type": "string", "minLength": 1},
			"decision":                    map[string]any{"enum": []string{"execute", "blocked"}},
			"request_compiler_loop_state": requestCompilerLoopStateSchema(),
			"validated_execution_request": map[string]any{
				"anyOf": []any{
					validatedExecutionRequestSchema(),
					map[string]any{"type": "null"},
				},
			},
			"repair_advice": map[string]any{
				"anyOf": []any{
					repairAdviceSchema(),
					map[string]any{"type": "null"},
				},
			},
		},
		"additionalProperties": false,
	})
}

func writeResultVerifierOutcomeSchema(path string) error {
	return writeSchemaFile(path, map[string]any{
		"type":     "object",
		"required": []string{"review", "loop_state"},
		"properties": map[string]any{
			"review":     verificationReviewSchema(),
			"loop_state": requestCompilerLoopStateSchema(),
		},
		"additionalProperties": false,
	})
}

func writeSchemaFile(path string, value any) error {
	raw, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, append(raw, '\n'), 0o644)
}

type runtimeSummary struct {
	OutputFile            string
	VerificationPass      bool
	VerificationOperation string
	VerificationReasons   []string
}

func loadRequestText(path string) (string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read request text %s: %w", path, err)
	}
	return strings.TrimSpace(string(raw)), nil
}

func loadWorkbookFacts(inputFile string) (useorchestrator.WorkbookFacts, error) {
	facts, err := runtimeinspect.InspectWorkbookSurfaceFacts(inputFile)
	if err != nil {
		return useorchestrator.WorkbookFacts{}, err
	}

	return useorchestrator.WorkbookFacts{
		VisibleSheetNames: append([]string(nil), facts.VisibleSheetNames...),
		HeaderRowBySheet:  cloneHeaderRows(facts.HeaderRowBySheet),
	}, nil
}

func cloneHeaderRows(input map[string][]string) map[string][]string {
	if len(input) == 0 {
		return map[string][]string{}
	}
	out := make(map[string][]string, len(input))
	for sheet, header := range input {
		out[sheet] = append([]string(nil), header...)
	}
	return out
}

func buildRuntimeSummary(result useorchestrator.RunResult) (runtimeSummary, error) {
	return runtimeSummary{
		OutputFile:            result.OutputFile(""),
		VerificationPass:      result.Verification.Pass,
		VerificationOperation: result.Verification.Operation,
		VerificationReasons:   append([]string(nil), result.Verification.Reasons...),
	}, nil
}

func recoverOrchestratorDecision(
	outputPath string,
	envelope UseEnvelope,
	requestText string,
) (useorchestrator.OrchestratorDecision, error) {
	raw, err := os.ReadFile(outputPath)
	if err != nil {
		return useorchestrator.OrchestratorDecision{}, err
	}
	var document map[string]any
	if err := json.Unmarshal(raw, &document); err != nil {
		return useorchestrator.OrchestratorDecision{}, err
	}

	decisionValue, _ := document["decision"].(string)
	if strings.TrimSpace(decisionValue) == "" {
		return useorchestrator.OrchestratorDecision{}, fmt.Errorf("orchestrator decision missing decision")
	}

	decision := useorchestrator.OrchestratorDecision{
		ScenarioID: envelope.ScenarioID,
		Decision:   decisionValue,
	}

	if decision.Decision == "blocked" {
		return decision, fmt.Errorf("orchestrator blocked without canonical contract")
	}

	facts, err := loadWorkbookFacts(envelope.InputFile)
	if err != nil {
		facts = useorchestrator.WorkbookFacts{}
	}
	if joinReq, ok := useorchestrator.TryResolveJoinLookupRequest(document, envelope.ScenarioID, envelope.InputFile, envelope.OutputFile, requestText, facts); ok {
		validated := joinReq.ValidatedExecutionRequest()
		decision.ValidatedExecutionRequest = &validated
		return decision, nil
	}

	compiled, err := requestcompiler.Compile(requestcompiler.Input{
		RequestSource: requestcompiler.RequestSource{
			Kind: requestcompiler.RequestSourceDirectText,
			Text: requestText,
		},
		WorkspaceRoot: filepath.Dir(envelope.InputFile),
		WorkingDir:    filepath.Dir(envelope.InputFile),
		ScenarioSlug:  envelope.ScenarioID,
		InputWorkbooks: []requestcompiler.WorkbookInput{
			{Path: envelope.InputFile, Role: "primary_input"},
		},
		OutputFile: envelope.OutputFile,
	})
	if err != nil {
		return useorchestrator.OrchestratorDecision{}, err
	}
	if compiled.ValidatedExecutionRequest == nil {
		return useorchestrator.OrchestratorDecision{}, fmt.Errorf("deterministic recovery did not emit validated execution request")
	}
	validated := useorchestrator.ValidatedExecutionRequest(*compiled.ValidatedExecutionRequest)
	decision.ValidatedExecutionRequest = &validated
	return decision, bindValidatedExecutionRequestToEnvelope(&decision, envelope)
}

func bindValidatedExecutionRequestToEnvelope(decision *useorchestrator.OrchestratorDecision, envelope UseEnvelope) error {
	if decision == nil || decision.ValidatedExecutionRequest == nil {
		return nil
	}

	req := decision.ValidatedExecutionRequest
	if strings.TrimSpace(req.RequestKind) == "" {
		req.RequestKind = envelope.Request.Kind
	}
	if strings.TrimSpace(req.RequestKind) != envelope.Request.Kind {
		return fmt.Errorf(
			"validated_execution_request request_kind %q does not match envelope request.kind %q",
			req.RequestKind,
			envelope.Request.Kind,
		)
	}
	if strings.TrimSpace(req.ScenarioID) != envelope.ScenarioID {
		return fmt.Errorf(
			"validated_execution_request scenario_id %q does not match envelope scenario_id %q",
			req.ScenarioID,
			envelope.ScenarioID,
		)
	}
	if !samePath(req.InputFile, envelope.InputFile) {
		return fmt.Errorf(
			"validated_execution_request input_file %q does not match envelope input_file %q",
			req.InputFile,
			envelope.InputFile,
		)
	}
	if !samePath(req.OutputFile, envelope.OutputFile) {
		return fmt.Errorf(
			"validated_execution_request output_file %q does not match envelope output_file %q",
			req.OutputFile,
			envelope.OutputFile,
		)
	}

	req.ScenarioID = envelope.ScenarioID
	req.InputFile = envelope.InputFile
	req.OutputFile = envelope.OutputFile
	return nil
}

type codexJSONLEvent struct {
	Type string `json:"type"`
	Item *struct {
		Type              string   `json:"type"`
		Tool              string   `json:"tool"`
		Status            string   `json:"status"`
		ReceiverThreadIDs []string `json:"receiver_thread_ids"`
		AgentsStates      map[string]struct {
			Status  string `json:"status"`
			Message string `json:"message"`
		} `json:"agents_states"`
	} `json:"item,omitempty"`
}

func observedLoopStateFromJSONL(
	jsonlPath string,
	scenarioID string,
	role string,
	contract runtimebundle.RoleContract,
	launchConfig codexLaunchConfig,
	success bool,
) runtimesubagent.SubagentLoopState {
	raw, err := os.ReadFile(jsonlPath)
	if err != nil {
		return runtimesubagent.SubagentLoopState{}
	}

	var spawnCount, waitCount, closeCount int
	var sessionID string
	var lastMessage string
	lines := strings.Split(string(raw), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var event codexJSONLEvent
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			continue
		}
		if event.Item == nil || event.Item.Type != "collab_tool_call" || event.Item.Status != "completed" {
			continue
		}
		if sessionID == "" && len(event.Item.ReceiverThreadIDs) > 0 {
			sessionID = event.Item.ReceiverThreadIDs[0]
		}
		switch event.Item.Tool {
		case "spawn_agent":
			spawnCount++
		case "wait":
			waitCount++
		case "close_agent":
			closeCount++
		}
		for _, threadID := range event.Item.ReceiverThreadIDs {
			if state, ok := event.Item.AgentsStates[threadID]; ok && strings.TrimSpace(state.Message) != "" {
				lastMessage = state.Message
			}
		}
	}

	parentState := "BLOCKED"
	outcome := "blocked"
	if success {
		parentState = "DONE"
		outcome = "pass"
	}

	return runtimesubagent.SubagentLoopState{
		ScenarioID:  scenarioID,
		Outcome:     outcome,
		ParentState: parentState,
		Specialists: []runtimesubagent.SpecialistLoopState{{
			Role:                role,
			Carrier:             contract.Carrier,
			Model:               launchConfig.Model,
			ReasoningEffort:     launchConfig.ReasoningEffort,
			State:               "DONE",
			SessionID:           sessionID,
			SpawnCompletedCount: spawnCount,
			WaitCompletedCount:  waitCount,
			CloseCompletedCount: closeCount,
			LastMessage:         lastMessage,
		}},
	}
}

func requestCompilerLoopStateSchema() map[string]any {
	return map[string]any{
		"type":     "object",
		"required": []string{"scenario_id", "outcome", "parent_state", "specialists"},
		"properties": map[string]any{
			"scenario_id":  map[string]any{"type": "string"},
			"outcome":      map[string]any{"enum": []string{"pass", "blocked"}},
			"parent_state": map[string]any{"enum": []string{"DONE", "NEEDS_REVIEW", "BLOCKED", "AMBIGUOUS"}},
			"specialists": map[string]any{
				"type":     "array",
				"minItems": 1,
				"items": map[string]any{
					"type":     "object",
					"required": []string{"role", "carrier", "model", "reasoning_effort", "state", "session_id", "spawn_completed_count", "wait_completed_count", "close_completed_count", "last_message"},
					"properties": map[string]any{
						"role":                  map[string]any{"type": "string"},
						"carrier":               map[string]any{"enum": []string{"default", "explorer", "worker"}},
						"model":                 map[string]any{"type": "string"},
						"reasoning_effort":      map[string]any{"enum": []string{"low", "medium", "high", "xhigh"}},
						"state":                 map[string]any{"enum": []string{"DONE", "NEEDS_REVIEW", "BLOCKED", "AMBIGUOUS"}},
						"session_id":            map[string]any{"type": "string"},
						"spawn_completed_count": map[string]any{"type": "integer"},
						"wait_completed_count":  map[string]any{"type": "integer"},
						"close_completed_count": map[string]any{"type": "integer"},
						"last_message":          map[string]any{"type": "string"},
					},
					"additionalProperties": false,
				},
			},
		},
		"additionalProperties": false,
	}
}

func validatedExecutionRequestSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"required": []string{
			"scenario_id",
			"request_kind",
			"input_file",
			"source_sheet",
			"output_file",
			"execution_kind",
			"composition_kind",
		},
		"properties": map[string]any{
			"scenario_id":      map[string]any{"type": "string"},
			"request_kind":     map[string]any{"enum": []string{"structured_use_request", "prompt_text"}},
			"request_text":     map[string]any{"type": "string"},
			"input_file":       map[string]any{"type": "string"},
			"source_sheet":     map[string]any{"type": "string"},
			"output_file":      map[string]any{"type": "string"},
			"execution_kind":   map[string]any{"const": "composition"},
			"composition_kind": map[string]any{"enum": []string{"group_summary", "threshold_highlight", "join_lookup"}},
			"target_sheet":     map[string]any{"type": "string"},
			"summary_mode":     map[string]any{"enum": []string{"values", "formulas"}},
			"filters": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type":     "object",
					"required": []string{"column", "op", "value"},
					"properties": map[string]any{
						"column": map[string]any{"type": "string"},
						"op":     map[string]any{"type": "string"},
						"value":  map[string]any{"type": "string"},
					},
					"additionalProperties": false,
				},
			},
			"group_by": map[string]any{
				"type":     "array",
				"minItems": 1,
				"items":    map[string]any{"type": "string"},
			},
			"metrics": map[string]any{
				"type":     "array",
				"minItems": 1,
				"items": map[string]any{
					"type":     "object",
					"required": []string{"column", "op", "as"},
					"properties": map[string]any{
						"column": map[string]any{"type": "string"},
						"op":     map[string]any{"enum": []string{"sum", "count"}},
						"as":     map[string]any{"type": "string"},
					},
					"additionalProperties": false,
				},
			},
			"target_column":   map[string]any{"type": "string"},
			"operator":        map[string]any{"enum": []string{">", ">=", "<", "<=", "=", "=="}},
			"threshold":       map[string]any{"type": "number"},
			"highlight_color": map[string]any{"type": "string"},
			"lookup_sheet":    map[string]any{"type": "string"},
			"join_key":        map[string]any{"type": "string"},
			"include_source_columns": map[string]any{
				"type":  "array",
				"items": map[string]any{"type": "string"},
			},
			"append_lookup_columns": map[string]any{
				"type":     "array",
				"minItems": 1,
				"items":    map[string]any{"type": "string"},
			},
		},
		"allOf": []any{
			map[string]any{
				"if": map[string]any{
					"properties": map[string]any{"request_kind": map[string]any{"const": "structured_use_request"}},
					"required":   []string{"request_kind"},
				},
				"then": map[string]any{
					"not": map[string]any{
						"required": []string{"request_text"},
					},
				},
			},
			map[string]any{
				"if": map[string]any{
					"properties": map[string]any{"request_kind": map[string]any{"const": "prompt_text"}},
					"required":   []string{"request_kind"},
				},
				"then": map[string]any{
					"required": []string{"request_text"},
				},
			},
			map[string]any{
				"if": map[string]any{
					"properties": map[string]any{"composition_kind": map[string]any{"const": "group_summary"}},
					"required":   []string{"composition_kind"},
				},
				"then": map[string]any{
					"required": []string{"target_sheet", "summary_mode", "group_by", "metrics"},
					"not": map[string]any{
						"anyOf": []any{
							map[string]any{"required": []string{"target_column"}},
							map[string]any{"required": []string{"operator"}},
							map[string]any{"required": []string{"threshold"}},
							map[string]any{"required": []string{"highlight_color"}},
							map[string]any{"required": []string{"lookup_sheet"}},
							map[string]any{"required": []string{"join_key"}},
							map[string]any{"required": []string{"include_source_columns"}},
							map[string]any{"required": []string{"append_lookup_columns"}},
						},
					},
				},
			},
			map[string]any{
				"if": map[string]any{
					"properties": map[string]any{"composition_kind": map[string]any{"const": "threshold_highlight"}},
					"required":   []string{"composition_kind"},
				},
				"then": map[string]any{
					"required": []string{"target_column", "operator", "threshold"},
					"not": map[string]any{
						"anyOf": []any{
							map[string]any{"required": []string{"target_sheet"}},
							map[string]any{"required": []string{"summary_mode"}},
							map[string]any{"required": []string{"filters"}},
							map[string]any{"required": []string{"group_by"}},
							map[string]any{"required": []string{"metrics"}},
							map[string]any{"required": []string{"lookup_sheet"}},
							map[string]any{"required": []string{"join_key"}},
							map[string]any{"required": []string{"include_source_columns"}},
							map[string]any{"required": []string{"append_lookup_columns"}},
						},
					},
				},
			},
			map[string]any{
				"if": map[string]any{
					"properties": map[string]any{"composition_kind": map[string]any{"const": "join_lookup"}},
					"required":   []string{"composition_kind"},
				},
				"then": map[string]any{
					"required": []string{"target_sheet", "lookup_sheet", "join_key", "append_lookup_columns"},
					"not": map[string]any{
						"anyOf": []any{
							map[string]any{"required": []string{"summary_mode"}},
							map[string]any{"required": []string{"filters"}},
							map[string]any{"required": []string{"group_by"}},
							map[string]any{"required": []string{"metrics"}},
							map[string]any{"required": []string{"target_column"}},
							map[string]any{"required": []string{"operator"}},
							map[string]any{"required": []string{"threshold"}},
							map[string]any{"required": []string{"highlight_color"}},
						},
					},
				},
			},
		},
		"additionalProperties": false,
	}
}

func repairAdviceSchema() map[string]any {
	return map[string]any{
		"type":     "object",
		"required": []string{"summary", "suggested_actions", "assumptions"},
		"properties": map[string]any{
			"summary": map[string]any{"type": "string"},
			"suggested_actions": map[string]any{
				"type":     "array",
				"minItems": 1,
				"items":    map[string]any{"type": "string"},
			},
			"assumptions": map[string]any{
				"type":  "array",
				"items": map[string]any{"type": "string"},
			},
		},
		"additionalProperties": false,
	}
}

func verificationReviewSchema() map[string]any {
	return map[string]any{
		"type":     "object",
		"required": []string{"status", "summary"},
		"properties": map[string]any{
			"status":  map[string]any{"enum": []string{"pass", "fail", "blocked", "needs_human_checkpoint"}},
			"summary": map[string]any{"type": "string"},
			"reasons": map[string]any{
				"type":  "array",
				"items": map[string]any{"type": "string"},
			},
		},
		"additionalProperties": false,
	}
}
