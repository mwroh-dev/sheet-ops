package main

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	runtimeconfig "github.com/mwroh/sheet-ops/runtime/runtimeconfig"
	runtimeworkbookcase "github.com/mwroh/sheet-ops/runtime/workbookcase"
	"github.com/xuri/excelize/v2"
)

type installedRunIntentResult struct {
	SchemaVersion string                   `json:"schema_version"`
	OK            bool                     `json:"ok"`
	Command       string                   `json:"command"`
	Recoverable   bool                     `json:"recoverable"`
	Artifacts     []PublicResultArtifact   `json:"artifacts"`
	Fingerprints  installedRunFingerprints `json:"fingerprints"`
	Entry         string                   `json:"entry"`
	Status        string                   `json:"status"`
	Runtime       installedRuntimePayload  `json:"runtime"`
}

type installedRunFingerprints struct {
	NormalizedIntentSHA256 string `json:"normalized_intent_sha256"`
	InputWorkbookSHA256    string `json:"input_workbook_sha256"`
	OutputWorkbookSHA256   string `json:"output_workbook_sha256"`
}

type installedRuntimePayload struct {
	Paths        installedRunPaths        `json:"paths"`
	Verification installedVerification    `json:"verification"`
	Execution    installedExecutionResult `json:"execution"`
}

type installedRunPaths struct {
	EvidenceDir            string `json:"evidence_dir"`
	VerificationPath       string `json:"verification_path"`
	ExecutionPath          string `json:"execution_path"`
	OutcomePath            string `json:"outcome_path"`
	RepairAdvicePath       string `json:"repair_advice_path"`
	VerificationReviewPath string `json:"verification_review_path"`
}

type installedVerification struct {
	Pass                 bool     `json:"pass"`
	Operation            string   `json:"operation"`
	OutputFile           string   `json:"output_file"`
	OutputWorkbookSHA256 string   `json:"output_workbook_sha256"`
	WrittenCells         []string `json:"written_cells"`
	Reasons              []string `json:"reasons"`
}

type installedExecutionResult struct {
	Operation    string   `json:"operation"`
	OutputFile   string   `json:"output_file"`
	WrittenCells []string `json:"written_cells"`
}

func TestInstallSkillBundledCLIRunIntentExecutesWorkbookEndToEnd(t *testing.T) {
	if status := inspectGo("go"); status.kind != goStatusReady {
		t.Skipf("go runtime is not ready for installed E2E smoke: %s %v", status.kind, status.err)
	}

	projectDir := t.TempDir()
	installScript := filepath.Join("..", "..", "install-skill.sh")
	cmd := exec.Command(installScript,
		"--project", projectDir,
		"--go-bin", "go",
	)
	installOutput, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("install-skill.sh failed: %v\noutput=%s", err, installOutput)
	}

	workspaceDir := filepath.Join(projectDir, "workspace")
	if err := os.MkdirAll(workspaceDir, 0o755); err != nil {
		t.Fatalf("MkdirAll(workspace): %v", err)
	}
	inputFile := filepath.Join(workspaceDir, "line-items.xlsx")
	outputFile := filepath.Join(workspaceDir, "line-items-output.xlsx")
	writeLineItemsWorkbook(t, inputFile)
	intentFile := filepath.Join(workspaceDir, "append-intent.json")
	writeAppendRowsIntent(t, intentFile)

	installedCLI := filepath.Join(projectDir, ".codex", "skills", "sheet-ops", "bin", "sheet-ops-codex")
	var preview previewRequestDocument
	runInstalledCLIJSONWithEnv(t, installedCLI, &preview, []string{
		runtimeconfig.RetentionModeEnv + "=" + string(runtimeconfig.RetentionModeFull),
		runtimeconfig.RenderModeEnv + "=" + string(runtimeconfig.RenderModeNever),
	}, "preview-request",
		"--json",
		"--intent-file", intentFile,
		"--input-file", inputFile,
		"--output-file", outputFile,
		"--scenario-id", "installed-append-rows-e2e",
	)

	var result installedRunIntentResult
	runInstalledCLIJSONWithEnv(t, installedCLI, &result, []string{
		runtimeconfig.RetentionModeEnv + "=" + string(runtimeconfig.RetentionModeFull),
		runtimeconfig.RenderModeEnv + "=" + string(runtimeconfig.RenderModeNever),
	}, "run-intent",
		"--intent-file", intentFile,
		"--input-file", inputFile,
		"--output-file", outputFile,
		"--scenario-id", "installed-append-rows-e2e",
	)

	if result.SchemaVersion != cliContractSchemaVersion {
		t.Fatalf("schema_version = %q, want %q", result.SchemaVersion, cliContractSchemaVersion)
	}
	if !result.OK {
		t.Fatalf("ok = false, want true: %+v", result)
	}
	if result.Command != "run-intent" || result.Entry != "run-intent" {
		t.Fatalf("command/entry = %q/%q, want run-intent", result.Command, result.Entry)
	}
	if result.Status != "executed" {
		t.Fatalf("status = %q, want executed", result.Status)
	}
	if result.Recoverable {
		t.Fatalf("recoverable = true, want false")
	}
	if result.Fingerprints.NormalizedIntentSHA256 != preview.Fingerprints.NormalizedIntentSHA256 {
		t.Fatalf("normalized_intent_sha256 = %q, want preview fingerprint %q", result.Fingerprints.NormalizedIntentSHA256, preview.Fingerprints.NormalizedIntentSHA256)
	}
	if result.Fingerprints.InputWorkbookSHA256 != preview.Fingerprints.InputWorkbookSHA256 {
		t.Fatalf("input_workbook_sha256 = %q, want preview fingerprint %q", result.Fingerprints.InputWorkbookSHA256, preview.Fingerprints.InputWorkbookSHA256)
	}
	outputFingerprint, err := fileSHA256(outputFile)
	if err != nil {
		t.Fatalf("fileSHA256(%s): %v", outputFile, err)
	}
	if result.Fingerprints.OutputWorkbookSHA256 != outputFingerprint {
		t.Fatalf("output_workbook_sha256 = %q, want output file fingerprint %q", result.Fingerprints.OutputWorkbookSHA256, outputFingerprint)
	}
	if result.Runtime.Verification.OutputWorkbookSHA256 != outputFingerprint {
		t.Fatalf("runtime verification output_workbook_sha256 = %q, want output file fingerprint %q", result.Runtime.Verification.OutputWorkbookSHA256, outputFingerprint)
	}
	var verificationArtifact installedVerification
	readJSONFileLocal(t, result.Runtime.Paths.VerificationPath, &verificationArtifact)
	if verificationArtifact.OutputWorkbookSHA256 != outputFingerprint {
		t.Fatalf("verification artifact output_workbook_sha256 = %q, want output file fingerprint %q", verificationArtifact.OutputWorkbookSHA256, outputFingerprint)
	}
	if verificationArtifact.OutputWorkbookSHA256 != result.Fingerprints.OutputWorkbookSHA256 {
		t.Fatalf("verification artifact output_workbook_sha256 = %q, want public result fingerprint %q", verificationArtifact.OutputWorkbookSHA256, result.Fingerprints.OutputWorkbookSHA256)
	}
	if !result.Runtime.Verification.Pass {
		t.Fatalf("runtime verification pass = false: %+v", result.Runtime.Verification)
	}
	if result.Runtime.Verification.Operation != runtimeworkbookcase.AppendRowsOperationName {
		t.Fatalf("verification operation = %q, want %q", result.Runtime.Verification.Operation, runtimeworkbookcase.AppendRowsOperationName)
	}
	if !slices.Contains(result.Runtime.Verification.WrittenCells, "LineItems!A3") {
		t.Fatalf("written_cells = %v, want LineItems!A3", result.Runtime.Verification.WrittenCells)
	}
	if !slices.Contains(result.Runtime.Verification.WrittenCells, "LineItems!B3") {
		t.Fatalf("written_cells = %v, want LineItems!B3", result.Runtime.Verification.WrittenCells)
	}
	if !slices.Contains(result.Runtime.Verification.WrittenCells, "LineItems!C3") {
		t.Fatalf("written_cells = %v, want LineItems!C3", result.Runtime.Verification.WrittenCells)
	}
	assertInstalledArtifactExists(t, result.Artifacts, "output_workbook", true, "primary_success")
	assertInstalledArtifactExists(t, result.Artifacts, "verification", true, "success_evidence")
	assertInstalledArtifactExists(t, result.Artifacts, "evidence_dir", true, "audit_trail")
	assertFileExistsLocal(t, result.Runtime.Paths.ExecutionPath)
	assertFileExistsLocal(t, result.Runtime.Paths.OutcomePath)
	assertOutputWorkbookRow(t, outputFile, "LineItems", 3, []string{"B002", "3", "15"})
	assertOutputWorkbookCell(t, outputFile, "LineItems", "A3", "B002")
	assertOutputWorkbookCell(t, outputFile, "LineItems", "B3", "3")
	assertOutputWorkbookCell(t, outputFile, "LineItems", "C3", "15")
}

func TestInstallSkillBundledCLIRunValidatedEmitsHandoffEnvelope(t *testing.T) {
	runInstalledInternalHandoffSmoke(t, "run-validated", "--request")
}

func TestInstallSkillBundledCLIRunRequestEmitsHandoffEnvelope(t *testing.T) {
	runInstalledInternalHandoffSmoke(t, "run-request", "--file")
}

func TestInstallSkillBundledCLIRunValidatedFailureEmitsHandoffEnvelope(t *testing.T) {
	if status := inspectGo("go"); status.kind != goStatusReady {
		t.Skipf("go runtime is not ready for installed run-validated failure smoke: %s %v", status.kind, status.err)
	}

	projectDir := t.TempDir()
	installScript := filepath.Join("..", "..", "install-skill.sh")
	cmd := exec.Command(installScript,
		"--project", projectDir,
		"--go-bin", "go",
	)
	installOutput, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("install-skill.sh failed: %v\noutput=%s", err, installOutput)
	}

	workspaceDir := filepath.Join(projectDir, "workspace")
	if err := os.MkdirAll(workspaceDir, 0o755); err != nil {
		t.Fatalf("MkdirAll(workspace): %v", err)
	}
	inputFile := filepath.Join(workspaceDir, "formula-protection.xlsx")
	outputFile := filepath.Join(workspaceDir, "formula-protection-output.xlsx")
	requestFile := filepath.Join(workspaceDir, "validated-failure-request.json")
	writeFormulaProtectionFailureWorkbook(t, inputFile)
	writeFormulaProtectionFailureRequest(t, requestFile, inputFile, outputFile)

	installedCLI := filepath.Join(projectDir, ".codex", "skills", "sheet-ops", "bin", "sheet-ops-codex")
	var result InternalHandoffRunResult
	runInstalledCLIJSONExpectFailureWithEnv(t, installedCLI, &result, []string{
		runtimeconfig.RetentionModeEnv + "=" + string(runtimeconfig.RetentionModeFull),
		runtimeconfig.RenderModeEnv + "=" + string(runtimeconfig.RenderModeNever),
	}, "run-validated",
		"--request", requestFile,
	)

	if result.SchemaVersion != cliContractSchemaVersion {
		t.Fatalf("schema_version = %q, want %q", result.SchemaVersion, cliContractSchemaVersion)
	}
	if result.OK {
		t.Fatalf("ok = true, want false: %+v", result)
	}
	if result.Command != "run-validated" {
		t.Fatalf("command = %q, want run-validated", result.Command)
	}
	if result.Classification != cliClassificationInternalHandoff {
		t.Fatalf("classification = %q, want %q", result.Classification, cliClassificationInternalHandoff)
	}
	if result.Status != "executed" {
		t.Fatalf("status = %q, want executed", result.Status)
	}
	if result.Recoverable {
		t.Fatalf("recoverable = true, want false")
	}
	assertValidatesAgainstSchema(t, result, "contracts/cli/internal_handoff_result.schema.json")
	if result.Runtime.Verification.Pass {
		t.Fatalf("runtime verification pass = true, want false: %+v", result.Runtime.Verification)
	}
	if result.Runtime.Verification.Operation != runtimeworkbookcase.ProtectFormulaCellsOperationName {
		t.Fatalf("verification operation = %q, want %q", result.Runtime.Verification.Operation, runtimeworkbookcase.ProtectFormulaCellsOperationName)
	}
	if !slices.Contains(result.Runtime.Verification.Reasons, "D2 has no formula to protect") {
		t.Fatalf("verification reasons = %v, want D2 has no formula to protect", result.Runtime.Verification.Reasons)
	}
	outputFingerprint, err := fileSHA256(outputFile)
	if err != nil {
		t.Fatalf("fileSHA256(%s): %v", outputFile, err)
	}
	if result.Runtime.Verification.OutputWorkbookSHA256 != outputFingerprint {
		t.Fatalf("runtime verification output_workbook_sha256 = %q, want output file fingerprint %q", result.Runtime.Verification.OutputWorkbookSHA256, outputFingerprint)
	}
	var verificationArtifact installedVerification
	readJSONFileLocal(t, result.Runtime.Paths.VerificationPath, &verificationArtifact)
	if verificationArtifact.Pass {
		t.Fatalf("verification artifact pass = true, want false")
	}
	if verificationArtifact.OutputWorkbookSHA256 != outputFingerprint {
		t.Fatalf("verification artifact output_workbook_sha256 = %q, want output file fingerprint %q", verificationArtifact.OutputWorkbookSHA256, outputFingerprint)
	}
	assertInstalledArtifactExists(t, result.Artifacts, "output_workbook", false, "failure_evidence")
	assertInstalledArtifactExists(t, result.Artifacts, "verification", true, "success_evidence")
	assertInstalledArtifactExists(t, result.Artifacts, "evidence_dir", true, "audit_trail")
	assertFileExistsLocal(t, result.Runtime.Paths.ExecutionPath)
	assertFileExistsLocal(t, result.Runtime.Paths.OutcomePath)
	assertFileExistsLocal(t, result.Runtime.Paths.RepairAdvicePath)
	assertFileExistsLocal(t, result.Runtime.Paths.VerificationReviewPath)
	assertOutputWorkbookCell(t, outputFile, "LineItems", "D2", "not a formula")
}

func runInstalledInternalHandoffSmoke(t *testing.T, commandName string, requestFlag string) {
	t.Helper()

	if status := inspectGo("go"); status.kind != goStatusReady {
		t.Skipf("go runtime is not ready for installed %s smoke: %s %v", commandName, status.kind, status.err)
	}

	projectDir := t.TempDir()
	installScript := filepath.Join("..", "..", "install-skill.sh")
	cmd := exec.Command(installScript,
		"--project", projectDir,
		"--go-bin", "go",
	)
	installOutput, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("install-skill.sh failed: %v\noutput=%s", err, installOutput)
	}

	workspaceDir := filepath.Join(projectDir, "workspace")
	if err := os.MkdirAll(workspaceDir, 0o755); err != nil {
		t.Fatalf("MkdirAll(workspace): %v", err)
	}
	inputFile := filepath.Join(workspaceDir, "line-items.xlsx")
	outputFile := filepath.Join(workspaceDir, "line-items-output.xlsx")
	requestFile := filepath.Join(workspaceDir, "validated-request.json")
	writeLineItemsWorkbook(t, inputFile)
	writeValidatedAppendRowsRequest(t, requestFile, inputFile, outputFile)

	installedCLI := filepath.Join(projectDir, ".codex", "skills", "sheet-ops", "bin", "sheet-ops-codex")
	var result InternalHandoffRunResult
	runInstalledCLIJSONWithEnv(t, installedCLI, &result, []string{
		runtimeconfig.RetentionModeEnv + "=" + string(runtimeconfig.RetentionModeFull),
		runtimeconfig.RenderModeEnv + "=" + string(runtimeconfig.RenderModeNever),
	}, commandName,
		requestFlag, requestFile,
	)

	if result.SchemaVersion != cliContractSchemaVersion {
		t.Fatalf("schema_version = %q, want %q", result.SchemaVersion, cliContractSchemaVersion)
	}
	if !result.OK {
		t.Fatalf("ok = false, want true: %+v", result)
	}
	if result.Command != commandName {
		t.Fatalf("command = %q, want %s", result.Command, commandName)
	}
	if result.Classification != cliClassificationInternalHandoff {
		t.Fatalf("classification = %q, want %q", result.Classification, cliClassificationInternalHandoff)
	}
	if result.Status != "executed" {
		t.Fatalf("status = %q, want executed", result.Status)
	}
	if result.Recoverable {
		t.Fatalf("recoverable = true, want false")
	}
	assertValidatesAgainstSchema(t, result, "contracts/cli/internal_handoff_result.schema.json")
	if !result.Runtime.Verification.Pass {
		t.Fatalf("runtime verification pass = false: %+v", result.Runtime.Verification)
	}
	if result.Runtime.Verification.Operation != runtimeworkbookcase.AppendRowsOperationName {
		t.Fatalf("verification operation = %q, want %q", result.Runtime.Verification.Operation, runtimeworkbookcase.AppendRowsOperationName)
	}
	if !slices.Contains(result.Runtime.Verification.WrittenCells, "LineItems!A3") {
		t.Fatalf("written_cells = %v, want LineItems!A3", result.Runtime.Verification.WrittenCells)
	}
	outputFingerprint, err := fileSHA256(outputFile)
	if err != nil {
		t.Fatalf("fileSHA256(%s): %v", outputFile, err)
	}
	if result.Runtime.Verification.OutputWorkbookSHA256 != outputFingerprint {
		t.Fatalf("runtime verification output_workbook_sha256 = %q, want output file fingerprint %q", result.Runtime.Verification.OutputWorkbookSHA256, outputFingerprint)
	}
	var verificationArtifact installedVerification
	readJSONFileLocal(t, result.Runtime.Paths.VerificationPath, &verificationArtifact)
	if verificationArtifact.OutputWorkbookSHA256 != outputFingerprint {
		t.Fatalf("verification artifact output_workbook_sha256 = %q, want output file fingerprint %q", verificationArtifact.OutputWorkbookSHA256, outputFingerprint)
	}
	assertInstalledArtifactExists(t, result.Artifacts, "output_workbook", true, "primary_success")
	assertInstalledArtifactExists(t, result.Artifacts, "verification", true, "success_evidence")
	assertInstalledArtifactExists(t, result.Artifacts, "evidence_dir", true, "audit_trail")
	assertFileExistsLocal(t, result.Runtime.Paths.ExecutionPath)
	assertFileExistsLocal(t, result.Runtime.Paths.OutcomePath)
	assertOutputWorkbookRow(t, outputFile, "LineItems", 3, []string{"B002", "3", "15"})
	assertOutputWorkbookCell(t, outputFile, "LineItems", "A3", "B002")
	assertOutputWorkbookCell(t, outputFile, "LineItems", "B3", "3")
	assertOutputWorkbookCell(t, outputFile, "LineItems", "C3", "15")
}

func writeLineItemsWorkbook(t *testing.T, path string) {
	t.Helper()

	file := excelize.NewFile()
	defer func() { _ = file.Close() }()
	defaultSheet := file.GetSheetName(0)
	if err := file.SetSheetName(defaultSheet, "LineItems"); err != nil {
		t.Fatalf("SetSheetName: %v", err)
	}
	header := []any{"sku", "quantity", "unit_price"}
	if err := file.SetSheetRow("LineItems", "A1", &header); err != nil {
		t.Fatalf("SetSheetRow(header): %v", err)
	}
	existing := []any{"A001", 2, 10}
	if err := file.SetSheetRow("LineItems", "A2", &existing); err != nil {
		t.Fatalf("SetSheetRow(existing): %v", err)
	}
	if err := file.SaveAs(path); err != nil {
		t.Fatalf("SaveAs(%s): %v", path, err)
	}
}

func writeAppendRowsIntent(t *testing.T, path string) {
	t.Helper()

	intent := map[string]any{
		"source_sheet_candidates": []string{"LineItems"},
		"composition_candidates":  []string{"structured_row_append"},
		"append_rows": map[string]any{
			"include_source_columns": []string{"sku", "quantity", "unit_price"},
			"values": []map[string]any{
				{"cell": "sku", "value": "B002"},
				{"cell": "quantity", "value": 3},
				{"cell": "unit_price", "value": 15},
			},
		},
		"add_data_validation": map[string]any{
			"validation_rule": map[string]any{
				"ranges":         []string{},
				"rule_type":      "list",
				"allowed_values": []string{},
				"allow_blank":    false,
			},
		},
		"materialization": map[string]any{
			"preserve_original":       true,
			"output_destination_mode": "new_workbook",
			"write_shape":             "in_place_cells",
		},
		"ambiguity": map[string]any{
			"markers":           []string{},
			"unresolved_fields": []string{},
			"checkpoint_hints":  []string{},
		},
	}
	raw, err := json.MarshalIndent(intent, "", "  ")
	if err != nil {
		t.Fatalf("Marshal intent: %v", err)
	}
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		t.Fatalf("WriteFile(%s): %v", path, err)
	}
}

func writeFormulaProtectionFailureWorkbook(t *testing.T, path string) {
	t.Helper()

	file := excelize.NewFile()
	defer func() { _ = file.Close() }()
	defaultSheet := file.GetSheetName(0)
	if err := file.SetSheetName(defaultSheet, "LineItems"); err != nil {
		t.Fatalf("SetSheetName: %v", err)
	}
	header := []any{"sku", "quantity", "unit_price", "line_total"}
	if err := file.SetSheetRow("LineItems", "A1", &header); err != nil {
		t.Fatalf("SetSheetRow(header): %v", err)
	}
	row := []any{"A001", 2, 10, "not a formula"}
	if err := file.SetSheetRow("LineItems", "A2", &row); err != nil {
		t.Fatalf("SetSheetRow(row): %v", err)
	}
	if err := file.SaveAs(path); err != nil {
		t.Fatalf("SaveAs(%s): %v", path, err)
	}
}

func writeFormulaProtectionFailureRequest(t *testing.T, path, inputFile, outputFile string) {
	t.Helper()

	request := map[string]any{
		"scenario_id":      "installed-formula-protection-failure",
		"request_kind":     "prompt_text",
		"request_text":     "LineItems 계산 수식 셀을 보호한다.",
		"input_file":       inputFile,
		"source_sheet":     "LineItems",
		"output_file":      outputFile,
		"execution_kind":   "composition",
		"composition_kind": "formula_protection",
		"protection_rule": map[string]any{
			"formula_ranges": []string{"D2"},
			"input_ranges":   []string{"A2:C10"},
		},
	}
	raw, err := json.MarshalIndent(request, "", "  ")
	if err != nil {
		t.Fatalf("Marshal formula protection failure request: %v", err)
	}
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		t.Fatalf("WriteFile(%s): %v", path, err)
	}
}

func runInstalledCLIJSONWithEnv(t *testing.T, binary string, target any, env []string, args ...string) {
	t.Helper()

	cmd := exec.Command(binary, args...)
	cmd.Env = cleanSheetOpsEnv(os.Environ())
	cmd.Env = append(cmd.Env, env...)
	stdout, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			t.Fatalf("%s %s failed: %v\nstderr:\n%s\nstdout:\n%s", binary, strings.Join(args, " "), err, exitErr.Stderr, stdout)
		}
		t.Fatalf("%s %s failed: %v\nstdout:\n%s", binary, strings.Join(args, " "), err, stdout)
	}
	if err := json.Unmarshal(stdout, target); err != nil {
		t.Fatalf("json.Unmarshal(%s %s): %v\nstdout:\n%s", binary, strings.Join(args, " "), err, stdout)
	}
}

func runInstalledCLIJSONExpectFailureWithEnv(t *testing.T, binary string, target any, env []string, args ...string) {
	t.Helper()

	cmd := exec.Command(binary, args...)
	cmd.Env = cleanSheetOpsEnv(os.Environ())
	cmd.Env = append(cmd.Env, env...)
	stdout, err := cmd.Output()
	if err == nil {
		t.Fatalf("%s %s returned nil error, want failure\nstdout:\n%s", binary, strings.Join(args, " "), stdout)
	}
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("%s %s failed without exit status: %v\nstdout:\n%s", binary, strings.Join(args, " "), err, stdout)
	}
	if len(stdout) == 0 {
		t.Fatalf("%s %s wrote empty stdout on failure\nstderr:\n%s", binary, strings.Join(args, " "), exitErr.Stderr)
	}
	if err := json.Unmarshal(stdout, target); err != nil {
		t.Fatalf("json.Unmarshal(%s %s): %v\nstderr:\n%s\nstdout:\n%s", binary, strings.Join(args, " "), err, exitErr.Stderr, stdout)
	}
}

func cleanSheetOpsEnv(env []string) []string {
	cleaned := make([]string, 0, len(env))
	for _, entry := range env {
		if strings.HasPrefix(entry, "SHEET_OPS_") {
			continue
		}
		cleaned = append(cleaned, entry)
	}
	return cleaned
}

func assertInstalledArtifactExists(t *testing.T, artifacts []PublicResultArtifact, kind string, required bool, role string) {
	t.Helper()

	for _, artifact := range artifacts {
		if artifact.Kind != kind {
			continue
		}
		if artifact.Required != required {
			t.Fatalf("artifact %q required = %v, want %v", kind, artifact.Required, required)
		}
		if artifact.SuccessRole != role {
			t.Fatalf("artifact %q success_role = %q, want %q", kind, artifact.SuccessRole, role)
		}
		assertFileOrDirExistsLocal(t, artifact.Path)
		return
	}
	t.Fatalf("missing artifact kind %q in %+v", kind, artifacts)
}

func assertFileOrDirExistsLocal(t *testing.T, path string) {
	t.Helper()

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("stat %s: %v", path, err)
	}
}

func assertFileExistsLocal(t *testing.T, path string) {
	t.Helper()

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat %s: %v", path, err)
	}
	if info.IsDir() {
		t.Fatalf("%s is a directory, want file", path)
	}
}

func readJSONFileLocal(t *testing.T, path string, target any) {
	t.Helper()

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%s): %v", path, err)
	}
	if err := json.Unmarshal(raw, target); err != nil {
		t.Fatalf("json.Unmarshal(%s): %v", path, err)
	}
}

func assertOutputWorkbookCell(t *testing.T, path string, sheet string, cell string, want string) {
	t.Helper()

	workbook, err := excelize.OpenFile(path)
	if err != nil {
		t.Fatalf("OpenFile(%s): %v", path, err)
	}
	defer func() { _ = workbook.Close() }()
	got, err := workbook.GetCellValue(sheet, cell)
	if err != nil {
		t.Fatalf("GetCellValue(%s!%s): %v", sheet, cell, err)
	}
	if got != want {
		t.Fatalf("%s!%s = %q, want %q", sheet, cell, got, want)
	}
}

func assertOutputWorkbookRow(t *testing.T, path string, sheet string, rowNumber int, want []string) {
	t.Helper()

	workbook, err := excelize.OpenFile(path)
	if err != nil {
		t.Fatalf("OpenFile(%s): %v", path, err)
	}
	defer func() { _ = workbook.Close() }()

	rows, err := workbook.GetRows(sheet)
	if err != nil {
		t.Fatalf("GetRows(%s): %v", sheet, err)
	}
	rowIndex := rowNumber - 1
	if rowIndex < 0 || rowIndex >= len(rows) {
		t.Fatalf("%s row %d missing in %v", sheet, rowNumber, rows)
	}
	got := rows[rowIndex]
	if len(got) < len(want) {
		t.Fatalf("%s row %d = %v, want at least %v", sheet, rowNumber, got, want)
	}
	for index, wantValue := range want {
		if got[index] != wantValue {
			t.Fatalf("%s row %d = %v, want prefix %v", sheet, rowNumber, got, want)
		}
	}
}
