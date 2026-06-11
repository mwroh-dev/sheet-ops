package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type cliErrorEnvelopeView struct {
	SchemaVersion string             `json:"schema_version"`
	OK            bool               `json:"ok"`
	Error         cliErrorDetailView `json:"error"`
}

type cliErrorDetailView struct {
	Code              string   `json:"code"`
	Message           string   `json:"message"`
	Recoverable       bool     `json:"recoverable"`
	SuggestedCommands []string `json:"suggested_commands"`
	ExitCode          int      `json:"exit_code"`
}

func TestSchemaCommandUnknownJSONEmitsTypedError(t *testing.T) {
	stdout, stderr, err := executeCLIExpectError(t, "schema", "command", "nope", "--json")
	if err == nil {
		t.Fatalf("Execute returned nil, want error")
	}
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty before main handles the error", stderr)
	}

	var envelope cliErrorEnvelopeView
	if unmarshalErr := json.Unmarshal([]byte(stdout), &envelope); unmarshalErr != nil {
		t.Fatalf("json.Unmarshal error envelope: %v\nstdout:\n%s", unmarshalErr, stdout)
	}
	if envelope.OK {
		t.Fatalf("ok = true, want false")
	}
	if envelope.SchemaVersion != cliContractSchemaVersion {
		t.Fatalf("schema_version = %q, want %q", envelope.SchemaVersion, cliContractSchemaVersion)
	}
	if envelope.Error.Code != "unknown_command" {
		t.Fatalf("error.code = %q, want unknown_command", envelope.Error.Code)
	}
	if !strings.Contains(envelope.Error.Message, "nope") {
		t.Fatalf("error.message = %q, want it to mention unknown command", envelope.Error.Message)
	}
	if !envelope.Error.Recoverable {
		t.Fatalf("error.recoverable = false, want true")
	}
	if envelope.Error.ExitCode != 64 {
		t.Fatalf("error.exit_code = %d, want 64", envelope.Error.ExitCode)
	}
	if !containsString(envelope.Error.SuggestedCommands, "schema --json") {
		t.Fatalf("suggested_commands = %v, want schema --json", envelope.Error.SuggestedCommands)
	}
}

func TestSchemaCommandMissingJSONFlagClassifiesInvalidUsage(t *testing.T) {
	stdout, stderr, err := executeCLIExpectError(t, "schema", "command", "prepare-use")
	if err == nil {
		t.Fatalf("Execute returned nil, want error")
	}
	if stdout != "" {
		t.Fatalf("stdout = %q, want empty for non-json failure", stdout)
	}
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty before main handles the error", stderr)
	}

	cliErr := classifyCLIError(err)
	if cliErr.Code != "invalid_usage" {
		t.Fatalf("code = %q, want invalid_usage", cliErr.Code)
	}
	if cliErr.ExitCode != 64 {
		t.Fatalf("exit code = %d, want 64", cliErr.ExitCode)
	}
	if !cliErr.Recoverable {
		t.Fatalf("recoverable = false, want true")
	}
}

func TestMissingRequiredOptionClassifiesAsUsageError(t *testing.T) {
	_, stderr, err := executeCLIExpectError(t, "prepare-use")
	if err == nil {
		t.Fatalf("Execute returned nil, want error")
	}
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty before main handles the error", stderr)
	}

	cliErr := classifyCLIError(err)
	if cliErr.Code != "missing_required_option" {
		t.Fatalf("code = %q, want missing_required_option", cliErr.Code)
	}
	if cliErr.ExitCode != 64 {
		t.Fatalf("exit code = %d, want 64", cliErr.ExitCode)
	}
	if !containsString(cliErr.SuggestedCommands, "prepare-use --help") {
		t.Fatalf("suggested commands = %v, want prepare-use --help", cliErr.SuggestedCommands)
	}
}

func TestUnknownRootCommandClassifiesAsUsageError(t *testing.T) {
	_, _, err := executeCLIExpectError(t, "definitely-not-a-command")
	if err == nil {
		t.Fatalf("Execute returned nil, want error")
	}

	cliErr := classifyCLIError(err)
	if cliErr.Code != "unknown_command" {
		t.Fatalf("code = %q, want unknown_command", cliErr.Code)
	}
	if cliErr.ExitCode != 64 {
		t.Fatalf("exit code = %d, want 64", cliErr.ExitCode)
	}
	if !containsString(cliErr.SuggestedCommands, "capabilities --json") {
		t.Fatalf("suggested commands = %v, want capabilities --json", cliErr.SuggestedCommands)
	}
}

func TestStateRootMismatchClassifiesAsConfigurationError(t *testing.T) {
	workspaceRoot := t.TempDir()
	t.Setenv(stateRootEnv, filepath.Join(t.TempDir(), "other-state"))

	_, err := configureStateRootsFromStateRoot(workspaceRoot)
	if err == nil {
		t.Fatalf("configureStateRootsFromStateRoot returned nil, want mismatch error")
	}

	cliErr := classifyCLIError(err)
	if cliErr.Code != cliErrorStateRootMismatch {
		t.Fatalf("code = %q, want %q", cliErr.Code, cliErrorStateRootMismatch)
	}
	if cliErr.ExitCode != cliExitConfiguration {
		t.Fatalf("exit code = %d, want %d", cliErr.ExitCode, cliExitConfiguration)
	}
	if !cliErr.Recoverable {
		t.Fatalf("recoverable = false, want true")
	}
	if !containsString(cliErr.SuggestedCommands, "preflight --json --input-file <workbook>") {
		t.Fatalf("suggested commands = %v, want preflight", cliErr.SuggestedCommands)
	}
}

func TestPreviewRequestInvalidIntentJSONEmitsInvalidDataError(t *testing.T) {
	projectDir := t.TempDir()
	inputFile := filepath.Join(projectDir, "line-items.xlsx")
	outputFile := filepath.Join(projectDir, "line-items-output.xlsx")
	intentFile := filepath.Join(projectDir, "broken-intent.json")
	writeLineItemsWorkbook(t, inputFile)
	if err := os.WriteFile(intentFile, []byte(`{"composition_candidates": [`), 0o644); err != nil {
		t.Fatalf("WriteFile(%s): %v", intentFile, err)
	}

	stdout, stderr, err := executeCLIExpectError(t,
		"preview-request",
		"--json",
		"--intent-file", intentFile,
		"--input-file", inputFile,
		"--output-file", outputFile,
	)
	if err == nil {
		t.Fatalf("Execute returned nil, want invalid data error")
	}
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty for JSON error", stderr)
	}

	var envelope cliErrorEnvelopeView
	if unmarshalErr := json.Unmarshal([]byte(stdout), &envelope); unmarshalErr != nil {
		t.Fatalf("json.Unmarshal error envelope: %v\nstdout:\n%s", unmarshalErr, stdout)
	}
	if envelope.OK {
		t.Fatalf("ok = true, want false")
	}
	if envelope.Error.Code != cliErrorInvalidData {
		t.Fatalf("error.code = %q, want %q\nstdout:\n%s", envelope.Error.Code, cliErrorInvalidData, stdout)
	}
	if envelope.Error.ExitCode != cliExitData {
		t.Fatalf("error.exit_code = %d, want %d", envelope.Error.ExitCode, cliExitData)
	}
	if !envelope.Error.Recoverable {
		t.Fatalf("error.recoverable = false, want true")
	}
	if !containsString(envelope.Error.SuggestedCommands, "schema command preview-request --json") {
		t.Fatalf("suggested_commands = %v, want preview schema command", envelope.Error.SuggestedCommands)
	}
}

func TestPreviewRequestMissingIntentFileClassifiesAsRecoverableUsageError(t *testing.T) {
	projectDir := t.TempDir()
	inputFile := filepath.Join(projectDir, "line-items.xlsx")
	outputFile := filepath.Join(projectDir, "line-items-output.xlsx")
	intentFile := filepath.Join(projectDir, "missing-intent.json")
	writeLineItemsWorkbook(t, inputFile)

	stdout, stderr, err := executeCLIExpectError(t,
		"preview-request",
		"--json",
		"--intent-file", intentFile,
		"--input-file", inputFile,
		"--output-file", outputFile,
	)
	if err == nil {
		t.Fatalf("Execute returned nil, want missing file error")
	}
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty for JSON error", stderr)
	}

	var envelope cliErrorEnvelopeView
	if unmarshalErr := json.Unmarshal([]byte(stdout), &envelope); unmarshalErr != nil {
		t.Fatalf("json.Unmarshal error envelope: %v\nstdout:\n%s", unmarshalErr, stdout)
	}
	if envelope.Error.Code == cliErrorInvalidData {
		t.Fatalf("error.code = %q, want non-invalid-data classification for missing file\nstdout:\n%s", envelope.Error.Code, stdout)
	}
	if envelope.Error.Code != cliErrorInvalidUsage {
		t.Fatalf("error.code = %q, want %q\nstdout:\n%s", envelope.Error.Code, cliErrorInvalidUsage, stdout)
	}
	if envelope.Error.ExitCode != cliExitUsage {
		t.Fatalf("error.exit_code = %d, want %d", envelope.Error.ExitCode, cliExitUsage)
	}
	if !envelope.Error.Recoverable {
		t.Fatalf("error.recoverable = false, want true")
	}
	if !containsString(envelope.Error.SuggestedCommands, "preflight --json --input-file <workbook>") {
		t.Fatalf("suggested_commands = %v, want preflight suggestion", envelope.Error.SuggestedCommands)
	}
}

func TestWriteCLIErrorJSONAndReturnNilErrorReturnsNil(t *testing.T) {
	var stdout bytes.Buffer

	err := writeCLIErrorJSONAndReturn(&stdout, nil)

	if err != nil {
		t.Fatalf("writeCLIErrorJSONAndReturn returned non-nil error for nil *cliError: %T %[1]v", err)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty for nil *cliError", stdout.String())
	}
}

func executeCLIExpectError(t *testing.T, args ...string) (string, string, error) {
	t.Helper()

	root := newRootCommand()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&stderr)
	root.SetArgs(args)
	err := root.Execute()
	return stdout.String(), stderr.String(), err
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
