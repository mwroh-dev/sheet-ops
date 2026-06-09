package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	goruntime "runtime"

	requestmode "github.com/mwroh/sheet-ops/runtime/openlayer/requestmode"
	"github.com/mwroh/sheet-ops/runtime/openlayer/useorchestrator"
	runtimeschema "github.com/mwroh/sheet-ops/runtime/schema"
	runtimesession "github.com/mwroh/sheet-ops/runtime/session"
)

func executeUseOrganism(output io.Writer, ctx useCommandContext, req useorchestrator.OrganismExecutionRequest) error {
	if _, err := runtimesession.WriteProof(ctx.RunSession, "organism-execution-request.json", req); err != nil {
		return err
	}
	runtimeResult, err := runCaseLocalOrganismRequest(ctx.CaseRoot, req)
	if err != nil {
		return writeRunFailureResult(output, ctx.RunSession, ctx.EnvelopePath, ctx.BundleRoot, ctx.Envelope, "RUNTIME_FAILED", err.Error())
	}
	result, err := buildOrganismRunSuccessResult(ctx.RunSession, ctx.EnvelopePath, ctx.BundleRoot, ctx.Envelope, runtimeResult)
	if err != nil {
		return err
	}
	return writeUseTerminalResultJSON(output, result)
}

func runCaseLocalOrganismRequest(caseRoot string, req useorchestrator.OrganismExecutionRequest) (useorchestrator.OrganismRunResult, error) {
	restore, err := forceCaseLocalRuntimeRoots(caseRoot)
	if err != nil {
		return useorchestrator.OrganismRunResult{}, err
	}
	defer restore()
	return useorchestrator.OrchestrateOrganism(req)
}

func loadOrganismExecutionRequest(envelope UseEnvelope) (useorchestrator.OrganismExecutionRequest, error) {
	if envelope.Request.Kind != requestmode.ModeOrganismExecutionRequest {
		return useorchestrator.OrganismExecutionRequest{}, fmt.Errorf("organism execution requires request.kind %q, got %q", requestmode.ModeOrganismExecutionRequest, envelope.Request.Kind)
	}
	raw, err := os.ReadFile(envelope.Request.Path)
	if err != nil {
		return useorchestrator.OrganismExecutionRequest{}, fmt.Errorf("load organism execution request %s: %w", envelope.Request.Path, err)
	}
	var document any
	if err := json.Unmarshal(raw, &document); err != nil {
		return useorchestrator.OrganismExecutionRequest{}, err
	}
	if err := runtimeschema.ValidateStruct(organismExecutionRequestSchemaPath(), document); err != nil {
		return useorchestrator.OrganismExecutionRequest{}, err
	}
	var req useorchestrator.OrganismExecutionRequest
	if err := json.Unmarshal(raw, &req); err != nil {
		return useorchestrator.OrganismExecutionRequest{}, err
	}
	if req.ScenarioID != envelope.ScenarioID {
		return useorchestrator.OrganismExecutionRequest{}, fmt.Errorf("organism request scenario_id %q does not match envelope scenario_id %q", req.ScenarioID, envelope.ScenarioID)
	}
	if !samePath(req.InputFile, envelope.InputFile) {
		return useorchestrator.OrganismExecutionRequest{}, fmt.Errorf("organism request input_file %q does not match envelope input_file %q", req.InputFile, envelope.InputFile)
	}
	if !samePath(req.OutputFile, envelope.OutputFile) {
		return useorchestrator.OrganismExecutionRequest{}, fmt.Errorf("organism request output_file %q does not match envelope output_file %q", req.OutputFile, envelope.OutputFile)
	}
	return req, nil
}

func organismExecutionRequestSchemaPath() string {
	if schemaPath, err := executableRelativeRequestSchemaPath("organism_execution_request.schema.json"); err == nil {
		return schemaPath
	}
	_, file, _, ok := goruntime.Caller(0)
	if !ok {
		return filepath.Join("contracts", "requests", "organism_execution_request.schema.json")
	}
	return runtimeschema.ResolveRepoPath(file, 2, "contracts", "requests", "organism_execution_request.schema.json")
}

func buildOrganismRunSuccessResult(
	runSession runtimesession.RunSession,
	envelopePath string,
	bundleRoot string,
	envelope UseEnvelope,
	runtimeResult useorchestrator.OrganismRunResult,
) (UseTerminalResult, error) {
	evidencePath := filepath.Join(runSession.EvidenceDir, "use-terminal-evidence.json")
	evidence := useTerminalEvidence{
		Entry:                "use",
		Status:               "succeeded",
		ScenarioID:           runSession.ScenarioID,
		RunID:                runSession.RunID,
		SessionID:            runSession.SessionID,
		TerminalState:        "TERMINAL_SUCCESS",
		Message:              "sheet-ops organism run completed",
		EnvelopePath:         envelopePath,
		BundleRoot:           bundleRoot,
		RequestRef:           envelope.Request.Path,
		InputFile:            envelope.InputFile,
		OutputFile:           envelope.OutputFile,
		ScenarioEvidencePath: runtimeResult.OrganismVerificationPath,
	}
	if err := writeUseTerminalEvidence(evidencePath, evidence); err != nil {
		return UseTerminalResult{}, err
	}
	return newSucceededUseTerminalResult(
		runSession.ScenarioID,
		runSession.RunID,
		runSession.SessionID,
		evidence.Message,
		envelope.OutputFile,
		evidencePath,
		"",
	), nil
}
