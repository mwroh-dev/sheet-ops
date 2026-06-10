package main

import (
	"path/filepath"
	"testing"

	requestcompiler "github.com/mwroh/sheet-ops/runtime/openlayer/requestcompiler"
	runtimeworkbookcase "github.com/mwroh/sheet-ops/runtime/workbookcase"
)

func TestExecutedPublicEntryResultExposesAgentEnvelope(t *testing.T) {
	tempDir := t.TempDir()
	compiled := requestcompiler.PersistedResult{
		Result: requestcompiler.Result{
			Decision: requestcompiler.Decision{Status: requestcompiler.StatusCompiled},
		},
		WorkUnitID: "append-rows-20260610",
		RequestDir: filepath.Join(tempDir, "request-compiler"),
	}
	result := runtimeworkbookcase.RunResult{
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
			OutputFile: filepath.Join(tempDir, "line-items-output.xlsx"),
		},
		Verification: runtimeworkbookcase.VerificationResult{
			Pass:       true,
			Operation:  runtimeworkbookcase.AppendRowsOperationName,
			OutputFile: filepath.Join(tempDir, "line-items-output.xlsx"),
			Reasons:    []string{"output workbook contains appended row"},
		},
	}

	envelope := newExecutedPublicEntryResult("run-intent", compiled, result)

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
