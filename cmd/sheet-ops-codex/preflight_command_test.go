package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type preflightDocumentView struct {
	SchemaVersion string               `json:"schema_version"`
	Command       string               `json:"command"`
	OK            bool                 `json:"ok"`
	ReadOnly      bool                 `json:"read_only"`
	Checks        []preflightCheckView `json:"checks"`
}

type preflightCheckView struct {
	Name        string `json:"name"`
	Status      string `json:"status"`
	Recoverable bool   `json:"recoverable"`
	Message     string `json:"message"`
}

func TestPreflightJSONReportsReadOnlyChecks(t *testing.T) {
	projectDir := t.TempDir()
	inputFile := filepath.Join(projectDir, "input.xlsx")
	outputFile := filepath.Join(projectDir, "out", "output.xlsx")
	writeTestFile(t, inputFile, "placeholder workbook")
	t.Setenv(stateRootEnv, filepath.Join(projectDir, ".sheet-ops-state"))

	var doc preflightDocumentView
	executeCLIJSON(t, &doc,
		"preflight",
		"--json",
		"--project", projectDir,
		"--input-file", inputFile,
		"--output-file", outputFile,
	)

	if doc.SchemaVersion != cliContractSchemaVersion {
		t.Fatalf("schema_version = %q, want %q", doc.SchemaVersion, cliContractSchemaVersion)
	}
	if doc.Command != "preflight" {
		t.Fatalf("command = %q, want preflight", doc.Command)
	}
	if !doc.OK {
		t.Fatalf("ok = false, checks = %+v", doc.Checks)
	}
	if !doc.ReadOnly {
		t.Fatalf("read_only = false, want true")
	}
	for _, name := range []string{"project_directory", "go_runtime", "package_manifest", "state_root", "input_file", "output_parent"} {
		check := findPreflightCheck(t, doc.Checks, name)
		if check.Status != "pass" {
			t.Fatalf("%s status = %q, want pass: %+v", name, check.Status, check)
		}
	}
}

func TestPreflightJSONReportsRecoverableStateRootMismatch(t *testing.T) {
	projectDir := t.TempDir()
	inputFile := filepath.Join(projectDir, "input.xlsx")
	writeTestFile(t, inputFile, "placeholder workbook")
	t.Setenv(stateRootEnv, filepath.Join(t.TempDir(), ".sheet-ops-state"))

	var doc preflightDocumentView
	executeCLIJSON(t, &doc,
		"preflight",
		"--json",
		"--project", projectDir,
		"--input-file", inputFile,
	)

	if doc.OK {
		t.Fatalf("ok = true, want false for state root mismatch")
	}
	check := findPreflightCheck(t, doc.Checks, "state_root")
	if check.Status != "fail" {
		t.Fatalf("state_root status = %q, want fail", check.Status)
	}
	if !check.Recoverable {
		t.Fatalf("state_root recoverable = false, want true")
	}
}

func TestPreflightInputFileRejectsDirectory(t *testing.T) {
	projectDir := t.TempDir()
	inputDir := filepath.Join(projectDir, "input-as-directory.xlsx")
	if err := os.MkdirAll(inputDir, 0o755); err != nil {
		t.Fatalf("MkdirAll(%s): %v", inputDir, err)
	}

	var doc preflightDocumentView
	executeCLIJSON(t, &doc,
		"preflight",
		"--json",
		"--project", projectDir,
		"--input-file", inputDir,
	)

	if doc.OK {
		t.Fatalf("ok = true, want false for directory input file: %+v", doc.Checks)
	}
	check := findPreflightCheck(t, doc.Checks, "input_file")
	if check.Status != preflightStatusFail {
		t.Fatalf("input_file status = %q, want %q", check.Status, preflightStatusFail)
	}
	if !check.Recoverable {
		t.Fatalf("input_file recoverable = false, want true")
	}
	if !strings.Contains(check.Message, "directory") {
		t.Fatalf("input_file message = %q, want directory", check.Message)
	}
}

func TestPreflightRequiresJSON(t *testing.T) {
	root := newRootCommand()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&stderr)
	root.SetArgs([]string{"preflight"})

	err := root.Execute()
	if err == nil {
		t.Fatalf("Execute returned nil, want error")
	}
	cliErr := classifyCLIError(err)
	if cliErr.Code != cliErrorInvalidUsage {
		t.Fatalf("code = %q, want %q", cliErr.Code, cliErrorInvalidUsage)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
}

func findPreflightCheck(t *testing.T, checks []preflightCheckView, name string) preflightCheckView {
	t.Helper()
	for _, check := range checks {
		if check.Name == name {
			return check
		}
	}
	t.Fatalf("missing preflight check %q: %+v", name, checks)
	return preflightCheckView{}
}

func writeTestFile(t *testing.T, path string, contents string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
}
