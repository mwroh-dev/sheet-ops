package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/mwroh/sheet-ops/runtime/openlayer/useorchestrator"
	runtimeworkbookcase "github.com/mwroh/sheet-ops/runtime/workbookcase"
)

func TestRunValidatedCommandWritesInternalHandoffEnvelope(t *testing.T) {
	resetRunIntentTestEnv(t)
	tempDir := t.TempDir()
	inputFile := filepath.Join(tempDir, "line-items.xlsx")
	outputFile := filepath.Join(tempDir, "line-items-output.xlsx")
	requestFile := filepath.Join(tempDir, "validated-request.json")
	writeLineItemsWorkbook(t, inputFile)
	writeValidatedAppendRowsRequest(t, requestFile, inputFile, outputFile)
	stubRunValidatedSuccess(t, tempDir, outputFile)

	cmd := newRootCommand()
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{"run-validated", "--request", requestFile})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("run-validated command returned error: %v", err)
	}
	var document map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &document); err != nil {
		t.Fatalf("run-validated stdout is not JSON: %v\nstdout:\n%s", err, stdout.String())
	}
	assertStringField(t, document, "schema_version", cliContractSchemaVersion)
	assertBoolField(t, document, "ok", true)
	assertStringField(t, document, "command", "run-validated")
	assertStringField(t, document, "status", "executed")
	assertStringField(t, document, "classification", "internal_handoff")
	assertBoolField(t, document, "recoverable", false)
	if _, ok := document["runtime"].(map[string]any); !ok {
		t.Fatalf("runtime has type %T, want object", document["runtime"])
	}
	artifacts, ok := document["artifacts"].([]any)
	if !ok || len(artifacts) == 0 {
		t.Fatalf("artifacts = %+v, want non-empty array", document["artifacts"])
	}
	assertDocumentArtifact(t, artifacts, "output_workbook", true, "primary_success")
	assertDocumentArtifact(t, artifacts, "verification", true, "success_evidence")
	assertDocumentArtifact(t, artifacts, "evidence_dir", true, "audit_trail")
}

func TestRunRequestCommandWritesInternalHandoffEnvelope(t *testing.T) {
	resetRunIntentTestEnv(t)
	tempDir := t.TempDir()
	inputFile := filepath.Join(tempDir, "line-items.xlsx")
	outputFile := filepath.Join(tempDir, "line-items-output.xlsx")
	requestFile := filepath.Join(tempDir, "validated-request.json")
	writeLineItemsWorkbook(t, inputFile)
	writeValidatedAppendRowsRequest(t, requestFile, inputFile, outputFile)
	stubRunValidatedSuccess(t, tempDir, outputFile)

	cmd := newRootCommand()
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{"run-request", "--file", requestFile})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("run-request command returned error: %v", err)
	}
	var document map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &document); err != nil {
		t.Fatalf("run-request stdout is not JSON: %v\nstdout:\n%s", err, stdout.String())
	}
	assertStringField(t, document, "schema_version", cliContractSchemaVersion)
	assertBoolField(t, document, "ok", true)
	assertStringField(t, document, "command", "run-request")
	assertStringField(t, document, "status", "executed")
	assertStringField(t, document, "classification", "internal_handoff")
	assertBoolField(t, document, "recoverable", false)
	if _, ok := document["runtime"].(map[string]any); !ok {
		t.Fatalf("runtime has type %T, want object", document["runtime"])
	}
	artifacts, ok := document["artifacts"].([]any)
	if !ok || len(artifacts) == 0 {
		t.Fatalf("artifacts = %+v, want non-empty array", document["artifacts"])
	}
	assertDocumentArtifact(t, artifacts, "output_workbook", true, "primary_success")
	assertDocumentArtifact(t, artifacts, "verification", true, "success_evidence")
	assertDocumentArtifact(t, artifacts, "evidence_dir", true, "audit_trail")
	assertValidatesAgainstSchema(t, document, "contracts/cli/internal_handoff_result.schema.json")
}

func TestRunValidatedCommandWritesFailureHandoffEnvelopeThenReturnsFailure(t *testing.T) {
	resetRunIntentTestEnv(t)
	tempDir := t.TempDir()
	inputFile := filepath.Join(tempDir, "line-items.xlsx")
	outputFile := filepath.Join(tempDir, "line-items-output.xlsx")
	requestFile := filepath.Join(tempDir, "validated-request.json")
	writeLineItemsWorkbook(t, inputFile)
	writeValidatedAppendRowsRequest(t, requestFile, inputFile, outputFile)
	stubRuntimeStartedFailure(t, tempDir, outputFile)

	cmd := newRootCommand()
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{"run-validated", "--request", requestFile})

	err := cmd.Execute()
	if err == nil {
		t.Fatalf("run-validated command returned nil error, want orchestration failure")
	}
	var document map[string]any
	if decodeErr := json.Unmarshal(stdout.Bytes(), &document); decodeErr != nil {
		t.Fatalf("run-validated stdout is not JSON: %v\nstdout:\n%s", decodeErr, stdout.String())
	}
	assertStringField(t, document, "schema_version", cliContractSchemaVersion)
	assertBoolField(t, document, "ok", false)
	assertStringField(t, document, "command", "run-validated")
	assertStringField(t, document, "classification", "internal_handoff")
	artifacts, ok := document["artifacts"].([]any)
	if !ok || len(artifacts) == 0 {
		t.Fatalf("artifacts = %+v, want non-empty array", document["artifacts"])
	}
	assertDocumentArtifact(t, artifacts, "output_workbook", false, "failure_evidence")
	assertDocumentArtifact(t, artifacts, "verification", true, "success_evidence")
	assertDocumentArtifact(t, artifacts, "evidence_dir", true, "audit_trail")
}

func TestRunValidatedCommandDoesNotEmitExecutedEnvelopeBeforeRuntimeStarts(t *testing.T) {
	resetRunIntentTestEnv(t)
	tempDir := t.TempDir()
	inputFile := filepath.Join(tempDir, "line-items.xlsx")
	outputFile := filepath.Join(tempDir, "line-items-output.xlsx")
	requestFile := filepath.Join(tempDir, "validated-request.json")
	writeLineItemsWorkbook(t, inputFile)
	writeValidatedAppendRowsRequest(t, requestFile, inputFile, outputFile)

	originalOrchestrateValidated := orchestrateValidated
	t.Cleanup(func() {
		orchestrateValidated = originalOrchestrateValidated
	})
	orchestrateValidated = func(req useorchestrator.ValidatedExecutionRequest) (useorchestrator.RunResult, error) {
		return useorchestrator.RunResult{}, errors.New("orchestration failed before runtime start")
	}

	cmd := newRootCommand()
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{"run-validated", "--request", requestFile})

	err := cmd.Execute()
	if err == nil {
		t.Fatalf("run-validated command returned nil error, want pre-runtime failure")
	}
	if stdout.Len() != 0 {
		t.Fatalf("run-validated stdout = %q, want empty before runtime start", stdout.String())
	}
}

func writeValidatedAppendRowsRequest(t *testing.T, path, inputFile, outputFile string) {
	t.Helper()

	request := useorchestrator.ValidatedExecutionRequest{
		ScenarioID:           "validated-append-rows",
		RequestKind:          "prompt_text",
		RequestText:          "LineItems 시트에 새 품목 행을 추가한다.",
		InputFile:            inputFile,
		SourceSheet:          "LineItems",
		OutputFile:           outputFile,
		ExecutionKind:        "composition",
		CompositionKind:      "structured_row_append",
		IncludeSourceColumns: []string{"sku", "quantity", "unit_price"},
		Values: []useorchestrator.CellValue{
			{Cell: "sku", Value: "B002"},
			{Cell: "quantity", Value: float64(3)},
			{Cell: "unit_price", Value: float64(15)},
		},
	}
	raw, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("Marshal validated request: %v", err)
	}
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		t.Fatalf("WriteFile(%s): %v", path, err)
	}
}

func stubRunValidatedSuccess(t *testing.T, tempDir, outputFile string) {
	t.Helper()

	originalOrchestrateValidated := orchestrateValidated
	t.Cleanup(func() {
		orchestrateValidated = originalOrchestrateValidated
	})
	orchestrateValidated = func(req useorchestrator.ValidatedExecutionRequest) (useorchestrator.RunResult, error) {
		evidenceDir := filepath.Join(tempDir, ".sheet-ops-state", "artifacts", "evidence")
		reportDir := filepath.Join(tempDir, ".sheet-ops-state", "artifacts", "reports")
		verificationPath := filepath.Join(evidenceDir, "verification.json")
		executionPath := filepath.Join(evidenceDir, "execution.json")
		outcomePath := filepath.Join(reportDir, "outcome.json")
		reportPath := filepath.Join(reportDir, "report.md")
		if err := os.MkdirAll(evidenceDir, 0o755); err != nil {
			t.Fatalf("MkdirAll(%s): %v", evidenceDir, err)
		}
		if err := os.MkdirAll(reportDir, 0o755); err != nil {
			t.Fatalf("MkdirAll(%s): %v", reportDir, err)
		}
		for path, content := range map[string]string{
			verificationPath: `{"pass":true}`,
			executionPath:    `{"operation":"append_rows"}`,
			outcomePath:      `{"ok":true}`,
			reportPath:       "# report\n",
		} {
			if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
				t.Fatalf("WriteFile(%s): %v", path, err)
			}
		}
		return useorchestrator.RunResult{
			IDs: runtimeworkbookcase.RunIDs{
				RunID:   "run-validated-20260611",
				TraceID: "trace-validated-20260611",
			},
			Paths: runtimeworkbookcase.RunPaths{
				EvidenceDir:      evidenceDir,
				ReportDir:        reportDir,
				ExecutionPath:    executionPath,
				VerificationPath: verificationPath,
				OutcomePath:      outcomePath,
				ReportPath:       reportPath,
			},
			Execution: runtimeworkbookcase.ExecutionSummary{
				Operation:  runtimeworkbookcase.AppendRowsOperationName,
				OutputFile: outputFile,
			},
			Verification: runtimeworkbookcase.VerificationResult{
				Pass:                 true,
				Operation:            runtimeworkbookcase.AppendRowsOperationName,
				OutputFile:           outputFile,
				OutputWorkbookSHA256: "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc",
				Reasons:              []string{"output workbook contains appended row"},
			},
		}, nil
	}
}

func assertStringField(t *testing.T, document map[string]any, key, want string) {
	t.Helper()

	if got, _ := document[key].(string); got != want {
		t.Fatalf("%s = %q, want %q in %+v", key, got, want, document)
	}
}

func assertBoolField(t *testing.T, document map[string]any, key string, want bool) {
	t.Helper()

	if got, _ := document[key].(bool); got != want {
		t.Fatalf("%s = %v, want %v in %+v", key, got, want, document)
	}
}

func assertDocumentArtifact(t *testing.T, artifacts []any, kind string, required bool, role string) {
	t.Helper()

	for _, value := range artifacts {
		artifact, ok := value.(map[string]any)
		if !ok {
			t.Fatalf("artifact has type %T, want object", value)
		}
		if artifact["kind"] == kind {
			if artifact["required"] != required {
				t.Fatalf("artifact %q required=%v want %v", kind, artifact["required"], required)
			}
			if artifact["success_role"] != role {
				t.Fatalf("artifact %q success_role=%q want %q", kind, artifact["success_role"], role)
			}
			if artifact["path"] == "" {
				t.Fatalf("artifact %q path is empty", kind)
			}
			return
		}
	}
	t.Fatalf("missing artifact kind %q in %+v", kind, artifacts)
}
