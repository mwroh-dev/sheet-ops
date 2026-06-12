package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	runtimeknowledge "github.com/mwroh/sheet-ops/runtime/knowledge"
	requestcompiler "github.com/mwroh/sheet-ops/runtime/openlayer/requestcompiler"
	"github.com/mwroh/sheet-ops/runtime/openlayer/useorchestrator"
	runtimeschema "github.com/mwroh/sheet-ops/runtime/schema"
	runtimeworkbookcase "github.com/mwroh/sheet-ops/runtime/workbookcase"
)

func TestRunIntentEntryEmitsPublicEnvelopeWhenRuntimeStartedThenFails(t *testing.T) {
	resetRunIntentTestEnv(t)
	tempDir, intentFile, inputFile, outputFile, intent := writeRunIntentTestInputs(t)
	stubRuntimeStartedFailure(t, tempDir, outputFile)

	runtimeResult, publicEntryResult, err := runIntentEntry(intent, intentCompileInput{
		IntentFile: intentFile,
		InputFile:  inputFile,
		OutputFile: outputFile,
		ScenarioID: "append-rows-failure",
	})
	if err == nil {
		t.Fatalf("runIntentEntry returned nil error, want orchestration failure")
	}
	if runtimeResult == nil {
		t.Fatalf("runIntentEntry returned nil runtime result")
	}
	if publicEntryResult == nil {
		t.Fatalf("runIntentEntry returned nil public entry result after runtime-started failure")
	}
	if publicEntryResult.OK {
		t.Fatalf("public entry ok = true, want false")
	}
	assertPublicResultArtifact(t, publicEntryResult.Artifacts, "output_workbook", false, "failure_evidence")
	if publicEntryResult.Fingerprints == nil || publicEntryResult.Fingerprints.OutputWorkbookSHA256 == "" {
		t.Fatalf("missing output workbook fingerprint in failed public entry: %+v", publicEntryResult.Fingerprints)
	}
	if err := runtimeschema.ValidateStruct(publicEntryResultSchemaPath(), *publicEntryResult); err != nil {
		t.Fatalf("public entry result schema rejected runtime-started failure envelope: %v", err)
	}
}

func TestRunIntentCommandWritesPublicEnvelopeThenReturnsFailure(t *testing.T) {
	resetRunIntentTestEnv(t)
	tempDir, intentFile, inputFile, outputFile, _ := writeRunIntentTestInputs(t)
	stubRuntimeStartedFailure(t, tempDir, outputFile)

	cmd := newRootCommand()
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{
		"run-intent",
		"--intent-file", intentFile,
		"--input-file", inputFile,
		"--output-file", outputFile,
		"--scenario-id", "append-rows-failure",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatalf("run-intent command returned nil error, want orchestration failure")
	}
	var result PublicEntryResult
	if decodeErr := json.Unmarshal(stdout.Bytes(), &result); decodeErr != nil {
		t.Fatalf("run-intent stdout is not a public entry result JSON document: %v\nstdout:\n%s", decodeErr, stdout.String())
	}
	if result.OK {
		t.Fatalf("public entry ok = true, want false")
	}
	assertPublicResultArtifact(t, result.Artifacts, "output_workbook", false, "failure_evidence")
	if err := runtimeschema.ValidateStruct(publicEntryResultSchemaPath(), result); err != nil {
		t.Fatalf("public entry result schema rejected command stdout envelope: %v", err)
	}
}

func TestRunIntentCommandReadsNormalizedIntentFromStdin(t *testing.T) {
	resetRunIntentTestEnv(t)
	tempDir, intentFile, inputFile, outputFile, _ := writeRunIntentTestInputs(t)
	stubRuntimeStartedFailure(t, tempDir, outputFile)
	intentRaw, err := os.ReadFile(intentFile)
	if err != nil {
		t.Fatalf("ReadFile(intent): %v", err)
	}

	cmd := newRootCommand()
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetIn(bytes.NewReader(intentRaw))
	cmd.SetArgs([]string{
		"run-intent",
		"--intent-file", "-",
		"--input-file", inputFile,
		"--output-file", outputFile,
		"--scenario-id", "stdin-append-rows-failure",
	})

	err = cmd.Execute()
	if err == nil {
		t.Fatalf("run-intent command returned nil error, want orchestration failure")
	}
	var result PublicEntryResult
	if decodeErr := json.Unmarshal(stdout.Bytes(), &result); decodeErr != nil {
		t.Fatalf("run-intent stdout is not a public entry result JSON document: %v\nstdout:\n%s", decodeErr, stdout.String())
	}
	if result.Fingerprints == nil {
		t.Fatalf("fingerprints is nil")
	}
	wantIntentSHA := sha256.Sum256(intentRaw)
	if result.Fingerprints.NormalizedIntentSHA256 != hex.EncodeToString(wantIntentSHA[:]) {
		t.Fatalf("normalized_intent_sha256 = %q, want stdin bytes hash", result.Fingerprints.NormalizedIntentSHA256)
	}
	if result.Fingerprints.OutputWorkbookSHA256 == "" {
		t.Fatalf("output_workbook_sha256 is empty")
	}
	if err := runtimeschema.ValidateStruct(publicEntryResultSchemaPath(), result); err != nil {
		t.Fatalf("public entry result schema rejected stdin command stdout envelope: %v", err)
	}
}

func TestRunIntentEntryDoesNotEmitExecutionEnvelopeBeforeRuntimeStarts(t *testing.T) {
	resetRunIntentTestEnv(t)
	_, intentFile, inputFile, outputFile, intent := writeRunIntentTestInputs(t)
	originalOrchestrateValidated := orchestrateValidated
	t.Cleanup(func() {
		orchestrateValidated = originalOrchestrateValidated
	})
	orchestrateValidated = func(req useorchestrator.ValidatedExecutionRequest) (useorchestrator.RunResult, error) {
		return useorchestrator.RunResult{}, errors.New("orchestration failed before runtime start")
	}

	runtimeResult, publicEntryResult, err := runIntentEntry(intent, intentCompileInput{
		IntentFile: intentFile,
		InputFile:  inputFile,
		OutputFile: outputFile,
		ScenarioID: "append-rows-pre-runtime-failure",
	})
	if err == nil {
		t.Fatalf("runIntentEntry returned nil error, want pre-runtime orchestration failure")
	}
	if runtimeResult != nil {
		t.Fatalf("runtime result = %+v, want nil before runtime start", *runtimeResult)
	}
	if publicEntryResult != nil {
		t.Fatalf("public entry result = %+v, want nil before runtime start", *publicEntryResult)
	}
}

func resetRunIntentTestEnv(t *testing.T) {
	t.Helper()

	t.Setenv(stateRootEnv, "")
	t.Setenv(runtimeworkbookcase.ArtifactRootEnv, "")
	t.Setenv(runtimeknowledge.KnowledgeRootEnv, "")
}

func writeRunIntentTestInputs(t *testing.T) (string, string, string, string, requestcompiler.NormalizedIntent) {
	t.Helper()

	tempDir := t.TempDir()
	inputFile := filepath.Join(tempDir, "line-items.xlsx")
	outputFile := filepath.Join(tempDir, "line-items-output.xlsx")
	intentFile := filepath.Join(tempDir, "append-rows.intent.json")
	writeLineItemsWorkbook(t, inputFile)
	writeAppendRowsIntent(t, intentFile)
	intent, err := requestcompiler.LoadNormalizedIntent(intentFile)
	if err != nil {
		t.Fatalf("LoadNormalizedIntent(%s): %v", intentFile, err)
	}
	return tempDir, intentFile, inputFile, outputFile, intent
}

func stubRuntimeStartedFailure(t *testing.T, tempDir, outputFile string) {
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
		repairAdvicePath := filepath.Join(reportDir, "repair-advice.json")
		if err := os.MkdirAll(evidenceDir, 0o755); err != nil {
			t.Fatalf("MkdirAll(%s): %v", evidenceDir, err)
		}
		if err := os.MkdirAll(reportDir, 0o755); err != nil {
			t.Fatalf("MkdirAll(%s): %v", reportDir, err)
		}
		if err := os.WriteFile(verificationPath, []byte(`{"pass":false}`), 0o644); err != nil {
			t.Fatalf("WriteFile(%s): %v", verificationPath, err)
		}
		if err := os.WriteFile(executionPath, []byte(`{"operation":"append_rows"}`), 0o644); err != nil {
			t.Fatalf("WriteFile(%s): %v", executionPath, err)
		}
		return useorchestrator.RunResult{
			IDs: runtimeworkbookcase.RunIDs{
				RunID:   "run-failed-20260611",
				TraceID: "trace-failed-20260611",
			},
			Paths: runtimeworkbookcase.RunPaths{
				TelemetryDir:        filepath.Join(tempDir, ".sheet-ops-state", "artifacts", "telemetry"),
				EvidenceDir:         evidenceDir,
				ReportDir:           reportDir,
				ExecutionPath:       executionPath,
				VerificationPath:    verificationPath,
				RepairAdvicePath:    repairAdvicePath,
				FailureEvidencePath: filepath.Join(evidenceDir, "failure.json"),
			},
			Execution: runtimeworkbookcase.ExecutionSummary{
				Operation:  runtimeworkbookcase.AppendRowsOperationName,
				OutputFile: outputFile,
			},
			Verification: runtimeworkbookcase.VerificationResult{
				Pass:                 false,
				Operation:            runtimeworkbookcase.AppendRowsOperationName,
				OutputFile:           outputFile,
				OutputWorkbookSHA256: "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc",
				Reasons:              []string{"verification failed"},
			},
		}, errors.New("verification failed")
	}
}
