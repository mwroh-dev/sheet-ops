package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	requestcompiler "github.com/mwroh/sheet-ops/runtime/openlayer/requestcompiler"
	runtimeschema "github.com/mwroh/sheet-ops/runtime/schema"
	runtimevalidate "github.com/mwroh/sheet-ops/runtime/validate"
	runtimeworkbookcase "github.com/mwroh/sheet-ops/runtime/workbookcase"
)

func TestExecutedPublicEntryResultExposesAgentEnvelope(t *testing.T) {
	envelope := newTestExecutedPublicEntryResult(t)

	if envelope.SchemaVersion != cliContractSchemaVersion {
		t.Fatalf("schema_version = %q, want %q", envelope.SchemaVersion, cliContractSchemaVersion)
	}
	if !envelope.OK {
		t.Fatalf("ok = false, want true")
	}
	if envelope.Command != "run-intent" {
		t.Fatalf("command = %q, want run-intent", envelope.Command)
	}
	if envelope.Recoverable {
		t.Fatalf("recoverable = true, want false for executed success")
	}
	if len(envelope.NextActions) == 0 {
		t.Fatalf("next_actions is empty")
	}
	assertPublicResultArtifact(t, envelope.Artifacts, "output_workbook", true, "primary_success")
	assertPublicResultArtifact(t, envelope.Artifacts, "verification", true, "success_evidence")
	assertPublicResultArtifact(t, envelope.Artifacts, "evidence_dir", true, "audit_trail")
}

func TestExecutedPublicEntryResultExposesInputFingerprints(t *testing.T) {
	envelope := newTestExecutedPublicEntryResult(t)
	document := publicEntryResultDocument(t, envelope)
	fingerprints, ok := document["fingerprints"].(map[string]any)
	if !ok {
		t.Fatalf("fingerprints has type %T, want object", document["fingerprints"])
	}

	if envelope.Fingerprints == nil {
		t.Fatalf("fingerprints is nil")
	}
	if len(envelope.Fingerprints.NormalizedIntentSHA256) != 64 {
		t.Fatalf("normalized_intent_sha256 length = %d, want 64", len(envelope.Fingerprints.NormalizedIntentSHA256))
	}
	if len(envelope.Fingerprints.InputWorkbookSHA256) != 64 {
		t.Fatalf("input_workbook_sha256 length = %d, want 64", len(envelope.Fingerprints.InputWorkbookSHA256))
	}
	outputFingerprint, ok := fingerprints["output_workbook_sha256"].(string)
	if !ok {
		t.Fatalf("output_workbook_sha256 missing or non-string in %+v", fingerprints)
	}
	if len(outputFingerprint) != 64 {
		t.Fatalf("output_workbook_sha256 length = %d, want 64", len(outputFingerprint))
	}
}

func TestExecutedPublicEntryResultValidatesAgainstPublicSchema(t *testing.T) {
	envelope := newTestExecutedPublicEntryResult(t)

	if err := runtimeschema.ValidateStruct(publicEntryResultSchemaPath(), envelope); err != nil {
		t.Fatalf("public entry result schema rejected executed envelope: %v", err)
	}
}

func TestPublicEntryResultGoldenValidatesAgainstPublicSchema(t *testing.T) {
	goldenPath := filepath.Join("testdata", "public_entry_result_executed.golden.json")
	raw, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("ReadFile(%s): %v", goldenPath, err)
	}
	var document any
	if err := json.Unmarshal(raw, &document); err != nil {
		t.Fatalf("json.Unmarshal(%s): %v", goldenPath, err)
	}
	if err := runtimeschema.ValidateStruct(publicEntryResultSchemaPath(), document); err != nil {
		t.Fatalf("public entry result golden failed schema validation: %v", err)
	}
}

func TestExecutedPublicEntryResultSchemaRejectsMissingSuccessArtifacts(t *testing.T) {
	document := publicEntryResultDocument(t, newTestExecutedPublicEntryResult(t))
	document["artifacts"] = []any{}

	if err := runtimeschema.ValidateStruct(publicEntryResultSchemaPath(), document); err == nil {
		t.Fatalf("public entry result schema accepted executed envelope without success artifacts")
	}
}

func TestExecutedPublicEntryResultSchemaRejectsMissingFingerprints(t *testing.T) {
	document := publicEntryResultDocument(t, newTestExecutedPublicEntryResult(t))
	delete(document, "fingerprints")

	if err := runtimeschema.ValidateStruct(publicEntryResultSchemaPath(), document); err == nil {
		t.Fatalf("public entry result schema accepted executed envelope without input fingerprints")
	}
}

func TestExecutedPublicEntryResultSchemaRejectsMissingOutputFingerprint(t *testing.T) {
	document := publicEntryResultDocument(t, newTestExecutedPublicEntryResult(t))
	fingerprints, ok := document["fingerprints"].(map[string]any)
	if !ok {
		t.Fatalf("fingerprints has type %T, want object", document["fingerprints"])
	}
	delete(fingerprints, "output_workbook_sha256")

	if err := runtimeschema.ValidateStruct(publicEntryResultSchemaPath(), document); err == nil {
		t.Fatalf("public entry result schema accepted executed envelope without output workbook fingerprint")
	}
}

func TestExecutedPublicEntryResultSchemaRejectsVerificationOutputFileWithoutHash(t *testing.T) {
	document := publicEntryResultDocument(t, newTestExecutedPublicEntryResult(t))
	verification := publicRuntimeVerification(t, document)
	verification["pass"] = false
	verification["output_file"] = "/tmp/output.xlsx"
	delete(verification, "output_workbook_sha256")

	if err := runtimeschema.ValidateStruct(publicEntryResultSchemaPath(), document); err == nil {
		t.Fatalf("public entry result schema accepted runtime verification output_file without output_workbook_sha256")
	}
}

func TestExecutedPublicEntryResultSchemaAllowsVerificationFailureWithOutputFileAndHash(t *testing.T) {
	document := publicEntryResultDocument(t, newTestExecutedFailurePublicEntryResult(t, true))

	if err := runtimeschema.ValidateStruct(publicEntryResultSchemaPath(), document); err != nil {
		t.Fatalf("public entry result schema rejected executed verification failure with output file and hash: %v", err)
	}
}

func TestExecutedPublicEntryResultSchemaAllowsEarlyVerificationFailureWithoutOutputFileOrHash(t *testing.T) {
	document := publicEntryResultDocument(t, newTestExecutedFailurePublicEntryResult(t, false))
	assertNoPublicResultArtifact(t, document, "output_workbook")

	if err := runtimeschema.ValidateStruct(publicEntryResultSchemaPath(), document); err != nil {
		t.Fatalf("public entry result schema rejected early runtime verification failure without output file or hash: %v", err)
	}
}

func TestExecutedPublicEntryResultSchemaAllowsEarlyFailureWithoutOutputFingerprint(t *testing.T) {
	document := publicEntryResultDocument(t, newTestExecutedFailurePublicEntryResult(t, false))
	fingerprints, ok := document["fingerprints"].(map[string]any)
	if !ok {
		t.Fatalf("fingerprints has type %T, want object", document["fingerprints"])
	}
	if _, ok := fingerprints["output_workbook_sha256"]; ok {
		t.Fatalf("early failure fingerprints unexpectedly include output_workbook_sha256: %+v", fingerprints)
	}

	if err := runtimeschema.ValidateStruct(publicEntryResultSchemaPath(), document); err != nil {
		t.Fatalf("public entry result schema rejected early failure without output workbook fingerprint: %v", err)
	}
}

func TestExecutedFingerprintsAllowsMissingOutputForEarlyFailure(t *testing.T) {
	fingerprints, err := executedFingerprints(previewFingerprints{
		NormalizedIntentSHA256: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		InputWorkbookSHA256:    "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
	}, "", "")
	if err != nil {
		t.Fatalf("executedFingerprints returned error for missing output identity: %v", err)
	}
	if fingerprints.NormalizedIntentSHA256 == "" || fingerprints.InputWorkbookSHA256 == "" {
		t.Fatalf("executedFingerprints lost input identity: %+v", fingerprints)
	}
	if fingerprints.OutputWorkbookSHA256 != "" {
		t.Fatalf("output_workbook_sha256 = %q, want empty for early failure", fingerprints.OutputWorkbookSHA256)
	}
}

func TestTerminalCompilerPublicEntryResultSchemaRejectsStatusMismatch(t *testing.T) {
	compiled := requestcompiler.PersistedResult{
		Result: requestcompiler.Result{
			Decision: requestcompiler.Decision{Status: requestcompiler.StatusNeedsHumanCheckpoint},
		},
		WorkUnitID: "needs-review-20260610",
		RequestDir: filepath.Join(t.TempDir(), "request-compiler"),
	}
	envelope, err := newTerminalCompilerResult("run-intent", compiled)
	if err == nil {
		t.Fatalf("newTerminalCompilerResult returned nil error, want checkpoint error")
	}
	document := publicEntryResultDocument(t, envelope)
	document["status"] = "blocked"

	if err := runtimeschema.ValidateStruct(publicEntryResultSchemaPath(), document); err == nil {
		t.Fatalf("public entry result schema accepted mismatched outer/request_compiler status")
	}
}

func newTestExecutedPublicEntryResult(t *testing.T) PublicEntryResult {
	t.Helper()

	return newTestExecutedPublicEntryResultForVerification(t, runtimeworkbookcase.VerificationResult{
		Pass:                 true,
		Operation:            runtimeworkbookcase.AppendRowsOperationName,
		OutputFile:           filepath.Join(t.TempDir(), "line-items-output.xlsx"),
		OutputWorkbookSHA256: "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc",
		Reasons:              []string{"output workbook contains appended row"},
	})
}

func newTestExecutedFailurePublicEntryResult(t *testing.T, withOutputFile bool) PublicEntryResult {
	t.Helper()

	tempDir := t.TempDir()
	verification := runtimeworkbookcase.VerificationResult{
		Pass:      false,
		Operation: runtimeworkbookcase.AppendRowsOperationName,
		Reasons:   []string{"verification failed"},
	}
	if withOutputFile {
		verification.OutputFile = filepath.Join(tempDir, "line-items-output.xlsx")
		verification.OutputWorkbookSHA256 = "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"
	}
	return newTestExecutedPublicEntryResultForVerification(t, verification)
}

func newTestExecutedPublicEntryResultForVerification(t *testing.T, verification runtimeworkbookcase.VerificationResult) PublicEntryResult {
	t.Helper()

	tempDir := t.TempDir()
	if verification.OutputFile == "" && verification.Pass {
		verification.OutputFile = filepath.Join(tempDir, "line-items-output.xlsx")
	}
	compiled := requestcompiler.PersistedResult{
		Result: requestcompiler.Result{
			Decision:                  requestcompiler.Decision{Status: requestcompiler.StatusCompiled},
			ValidatedExecutionRequest: &runtimevalidate.ValidatedExecutionRequest{},
		},
		WorkUnitID: "append-rows-20260610",
		RequestDir: filepath.Join(tempDir, "request-compiler"),
	}
	result := runtimeworkbookcase.RunResult{
		IDs: runtimeworkbookcase.RunIDs{
			RunID:   "run-append-rows-20260610",
			TraceID: "trace-append-rows-20260610",
		},
		Paths: runtimeworkbookcase.RunPaths{
			EvidenceDir:         filepath.Join(tempDir, "evidence"),
			ReportDir:           filepath.Join(tempDir, "reports"),
			ExecutionPath:       filepath.Join(tempDir, "evidence", "execution.json"),
			VerificationPath:    filepath.Join(tempDir, "evidence", "verification.json"),
			OutcomePath:         filepath.Join(tempDir, "reports", "outcome.json"),
			RepairAdvicePath:    filepath.Join(tempDir, "reports", "repair-advice.json"),
			ReportPath:          filepath.Join(tempDir, "reports", "report.md"),
			EvidenceIndex:       filepath.Join(tempDir, "evidence", "index.json"),
			FailureEvidencePath: filepath.Join(tempDir, "evidence", "failure.json"),
		},
		Execution: runtimeworkbookcase.ExecutionSummary{
			Operation:  runtimeworkbookcase.AppendRowsOperationName,
			OutputFile: verification.OutputFile,
		},
		Verification: verification,
	}

	return newExecutedPublicEntryResult("run-intent", compiled, result, executionFingerprints{
		NormalizedIntentSHA256: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		InputWorkbookSHA256:    "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		OutputWorkbookSHA256:   verification.OutputWorkbookSHA256,
	})
}

func publicEntryResultDocument(t *testing.T, envelope PublicEntryResult) map[string]any {
	t.Helper()

	raw, err := json.Marshal(envelope)
	if err != nil {
		t.Fatalf("Marshal public entry result: %v", err)
	}
	var document map[string]any
	if err := json.Unmarshal(raw, &document); err != nil {
		t.Fatalf("Unmarshal public entry result: %v", err)
	}
	return document
}

func publicRuntimeVerification(t *testing.T, document map[string]any) map[string]any {
	t.Helper()

	runtime, ok := document["runtime"].(map[string]any)
	if !ok {
		t.Fatalf("runtime has type %T, want object", document["runtime"])
	}
	verification, ok := runtime["verification"].(map[string]any)
	if !ok {
		t.Fatalf("runtime.verification has type %T, want object", runtime["verification"])
	}
	return verification
}

func TestPublicEntryResultTerminalCompilerExposesRecoverableAgentEnvelope(t *testing.T) {
	compiled := requestcompiler.PersistedResult{
		Result: requestcompiler.Result{
			Decision: requestcompiler.Decision{Status: requestcompiler.StatusNeedsHumanCheckpoint},
		},
		WorkUnitID: "needs-review-20260610",
		RequestDir: filepath.Join(t.TempDir(), "request-compiler"),
	}

	envelope, err := newTerminalCompilerResult("run-intent", compiled)
	if err == nil {
		t.Fatalf("newTerminalCompilerResult returned nil error, want checkpoint error")
	}
	if envelope.SchemaVersion != cliContractSchemaVersion {
		t.Fatalf("schema_version = %q, want %q", envelope.SchemaVersion, cliContractSchemaVersion)
	}
	if envelope.OK {
		t.Fatalf("ok = true, want false")
	}
	if envelope.Command != "run-intent" {
		t.Fatalf("command = %q, want run-intent", envelope.Command)
	}
	if !envelope.Recoverable {
		t.Fatalf("recoverable = false, want true")
	}
	if len(envelope.NextActions) == 0 {
		t.Fatalf("next_actions is empty")
	}
	assertPublicResultArtifact(t, envelope.Artifacts, "compiler_decision", true, "failure_context")
}

func TestTerminalCompilerPublicEntryResultValidatesAgainstPublicSchema(t *testing.T) {
	compiled := requestcompiler.PersistedResult{
		Result: requestcompiler.Result{
			Decision: requestcompiler.Decision{Status: requestcompiler.StatusNeedsHumanCheckpoint},
		},
		WorkUnitID: "needs-review-20260610",
		RequestDir: filepath.Join(t.TempDir(), "request-compiler"),
	}

	envelope, err := newTerminalCompilerResult("run-intent", compiled)
	if err == nil {
		t.Fatalf("newTerminalCompilerResult returned nil error, want checkpoint error")
	}
	if err := runtimeschema.ValidateStruct(publicEntryResultSchemaPath(), envelope); err != nil {
		t.Fatalf("public entry result schema rejected terminal compiler envelope: %v", err)
	}
}

func assertPublicResultArtifact(t *testing.T, artifacts []PublicResultArtifact, kind string, required bool, role string) {
	t.Helper()

	for _, artifact := range artifacts {
		if artifact.Kind == kind {
			if artifact.Required != required {
				t.Fatalf("artifact %q required=%v want %v", kind, artifact.Required, required)
			}
			if artifact.SuccessRole != role {
				t.Fatalf("artifact %q success_role=%q want %q", kind, artifact.SuccessRole, role)
			}
			if artifact.Path == "" {
				t.Fatalf("artifact %q path is empty", kind)
			}
			return
		}
	}
	t.Fatalf("missing artifact kind %q in %+v", kind, artifacts)
}

func assertNoPublicResultArtifact(t *testing.T, document map[string]any, kind string) {
	t.Helper()

	artifacts, ok := document["artifacts"].([]any)
	if !ok {
		t.Fatalf("artifacts has type %T, want array", document["artifacts"])
	}
	for _, value := range artifacts {
		artifact, ok := value.(map[string]any)
		if !ok {
			t.Fatalf("artifact has type %T, want object", value)
		}
		if artifact["kind"] == kind {
			t.Fatalf("unexpected artifact kind %q in %+v", kind, artifacts)
		}
	}
}

func publicEntryResultSchemaPath() string {
	return filepath.Join("..", "..", "contracts", "results", "public_entry_result.schema.json")
}
