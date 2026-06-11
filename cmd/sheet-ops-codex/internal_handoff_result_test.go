package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	runtimeschema "github.com/mwroh/sheet-ops/runtime/schema"
	runtimeworkbookcase "github.com/mwroh/sheet-ops/runtime/workbookcase"
)

func TestInternalHandoffRunResultValidatesAgainstSchema(t *testing.T) {
	envelope := newTestInternalHandoffRunResult(t, true)

	if err := runtimeschema.ValidateStruct(internalHandoffResultSchemaPath(), envelope); err != nil {
		t.Fatalf("internal handoff schema rejected run-validated envelope: %v", err)
	}
}

func TestInternalHandoffResultGoldenValidatesAgainstSchema(t *testing.T) {
	goldenPath := filepath.Join("testdata", "internal_handoff_run_validated.golden.json")
	raw, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("ReadFile(%s): %v", goldenPath, err)
	}
	var document any
	if err := json.Unmarshal(raw, &document); err != nil {
		t.Fatalf("json.Unmarshal(%s): %v", goldenPath, err)
	}
	if err := runtimeschema.ValidateStruct(internalHandoffResultSchemaPath(), document); err != nil {
		t.Fatalf("internal handoff golden failed schema validation: %v", err)
	}
}

func TestInternalHandoffResultSchemaRejectsMissingClassification(t *testing.T) {
	document := internalHandoffResultDocument(t, newTestInternalHandoffRunResult(t, true))
	delete(document, "classification")

	if err := runtimeschema.ValidateStruct(internalHandoffResultSchemaPath(), document); err == nil {
		t.Fatalf("internal handoff schema accepted envelope without classification")
	}
}

func TestInternalHandoffResultSchemaRejectsPublicEntryCommand(t *testing.T) {
	document := internalHandoffResultDocument(t, newTestInternalHandoffRunResult(t, true))
	document["command"] = "run-intent"

	if err := runtimeschema.ValidateStruct(internalHandoffResultSchemaPath(), document); err == nil {
		t.Fatalf("internal handoff schema accepted public entry command")
	}
}

func TestInternalHandoffResultSchemaRejectsFailedOutputMarkedAsPrimarySuccess(t *testing.T) {
	document := internalHandoffResultDocument(t, newTestInternalHandoffRunResult(t, false))
	artifact := publicResultArtifact(t, document, "output_workbook")
	artifact["required"] = true
	artifact["success_role"] = "primary_success"

	if err := runtimeschema.ValidateStruct(internalHandoffResultSchemaPath(), document); err == nil {
		t.Fatalf("internal handoff schema accepted failed output workbook marked as primary_success")
	}
}

func TestInternalHandoffResultSuppressesUnreadableOutputFileWithoutHash(t *testing.T) {
	envelope := newTestInternalHandoffRunResultWithVerification(t, runtimeworkbookcase.VerificationResult{
		Pass:       false,
		Operation:  runtimeworkbookcase.AppendRowsOperationName,
		OutputFile: filepath.Join(t.TempDir(), "missing-output.xlsx"),
		Reasons:    []string{"verification stopped before output workbook was readable"},
	})
	document := internalHandoffResultDocument(t, envelope)
	assertNoPublicResultArtifact(t, document, "output_workbook")

	runtimeDoc, ok := document["runtime"].(map[string]any)
	if !ok {
		t.Fatalf("runtime has type %T, want object", document["runtime"])
	}
	verification, ok := runtimeDoc["verification"].(map[string]any)
	if !ok {
		t.Fatalf("runtime.verification has type %T, want object", runtimeDoc["verification"])
	}
	if outputFile, ok := verification["output_file"].(string); ok && outputFile != "" {
		t.Fatalf("runtime.verification.output_file = %q, want omitted or empty without output_workbook_sha256", outputFile)
	}
	if err := runtimeschema.ValidateStruct(internalHandoffResultSchemaPath(), document); err != nil {
		t.Fatalf("internal handoff schema rejected degraded failure envelope without output hash: %v", err)
	}
}

func TestPublicEntryResultSchemaRejectsRunValidatedHandoff(t *testing.T) {
	document := internalHandoffResultDocument(t, newTestInternalHandoffRunResult(t, true))

	if err := runtimeschema.ValidateStruct(publicEntryResultSchemaPath(), document); err == nil {
		t.Fatalf("public entry result schema accepted internal handoff run-validated envelope")
	}
}

func newTestInternalHandoffRunResult(t *testing.T, pass bool) InternalHandoffRunResult {
	t.Helper()

	tempDir := t.TempDir()
	outputFile := filepath.Join(tempDir, "line-items-output.xlsx")
	verification := runtimeworkbookcase.VerificationResult{
		Pass:                 pass,
		Operation:            runtimeworkbookcase.AppendRowsOperationName,
		OutputFile:           outputFile,
		OutputWorkbookSHA256: "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc",
		Reasons:              []string{"verification evidence"},
	}
	return newTestInternalHandoffRunResultWithVerification(t, verification)
}

func newTestInternalHandoffRunResultWithVerification(t *testing.T, verification runtimeworkbookcase.VerificationResult) InternalHandoffRunResult {
	t.Helper()

	tempDir := t.TempDir()
	outputFile := verification.OutputFile
	if outputFile == "" {
		outputFile = filepath.Join(tempDir, "line-items-output.xlsx")
	}
	result := runtimeworkbookcase.RunResult{
		IDs: runtimeworkbookcase.RunIDs{
			RunID:   "run-validated-20260611",
			TraceID: "trace-validated-20260611",
		},
		Paths: runtimeworkbookcase.RunPaths{
			EvidenceDir:      filepath.Join(tempDir, "evidence"),
			ReportDir:        filepath.Join(tempDir, "reports"),
			ExecutionPath:    filepath.Join(tempDir, "evidence", "execution.json"),
			VerificationPath: filepath.Join(tempDir, "evidence", "verification.json"),
			OutcomePath:      filepath.Join(tempDir, "reports", "outcome.json"),
			ReportPath:       filepath.Join(tempDir, "reports", "report.md"),
		},
		Execution: runtimeworkbookcase.ExecutionSummary{
			Operation:  runtimeworkbookcase.AppendRowsOperationName,
			OutputFile: outputFile,
		},
		Verification: verification,
	}
	return newInternalHandoffRunResult("run-validated", result)
}

func internalHandoffResultDocument(t *testing.T, envelope InternalHandoffRunResult) map[string]any {
	t.Helper()

	raw, err := json.Marshal(envelope)
	if err != nil {
		t.Fatalf("Marshal internal handoff result: %v", err)
	}
	var document map[string]any
	if err := json.Unmarshal(raw, &document); err != nil {
		t.Fatalf("Unmarshal internal handoff result: %v", err)
	}
	return document
}

func internalHandoffResultSchemaPath() string {
	return filepath.Join("..", "..", "contracts", "cli", "internal_handoff_result.schema.json")
}
