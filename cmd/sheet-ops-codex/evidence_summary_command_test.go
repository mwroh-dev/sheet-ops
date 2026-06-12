package main

import (
	"path/filepath"
	"slices"
	"testing"
)

type evidenceSummaryDocument struct {
	SchemaVersion string                     `json:"schema_version"`
	Command       string                     `json:"command"`
	OK            bool                       `json:"ok"`
	ReadOnly      bool                       `json:"read_only"`
	EvidenceDir   string                     `json:"evidence_dir"`
	Status        string                     `json:"status"`
	Checks        []evidenceSummaryCheckView `json:"checks"`
	NextActions   []string                   `json:"next_actions"`
}

type evidenceSummaryCheckView struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

func TestEvidenceSummaryReportsVerifiedSuccess(t *testing.T) {
	evidenceDir := t.TempDir()
	writeJSONFile(t, filepath.Join(evidenceDir, "verification.json"), map[string]any{
		"pass":                   true,
		"output_workbook_sha256": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	})
	writeJSONFile(t, filepath.Join(evidenceDir, "execution.json"), map[string]any{
		"operation": "append_structured_rows",
	})

	var doc evidenceSummaryDocument
	executeCLIJSON(t, &doc, "evidence-summary", "--json", "--evidence-dir", evidenceDir)

	if doc.SchemaVersion != cliContractSchemaVersion {
		t.Fatalf("schema_version = %q, want %q", doc.SchemaVersion, cliContractSchemaVersion)
	}
	if doc.Command != "evidence-summary" {
		t.Fatalf("command = %q, want evidence-summary", doc.Command)
	}
	if !doc.OK || !doc.ReadOnly {
		t.Fatalf("ok/read_only = %v/%v, want true/true", doc.OK, doc.ReadOnly)
	}
	if doc.Status != "verified_success" {
		t.Fatalf("status = %q, want verified_success", doc.Status)
	}
	assertEvidenceCheck(t, doc.Checks, "verification_json", "pass")
	assertEvidenceCheck(t, doc.Checks, "execution_json", "present")
	assertEvidenceCheck(t, doc.Checks, "failure_json", "absent")
	assertEvidenceCheck(t, doc.Checks, "output_workbook_sha256", "present")
	if !slices.Contains(doc.NextActions, "safe_to_report_success_with_output_workbook_sha256") {
		t.Fatalf("next_actions missing success action: %v", doc.NextActions)
	}
}

func TestEvidenceSummaryReportsFailedVerification(t *testing.T) {
	evidenceDir := t.TempDir()
	writeJSONFile(t, filepath.Join(evidenceDir, "verification.json"), map[string]any{
		"pass":    false,
		"reasons": []string{"row count mismatch"},
	})
	writeJSONFile(t, filepath.Join(evidenceDir, "execution.json"), map[string]any{
		"operation": "append_structured_rows",
	})
	writeJSONFile(t, filepath.Join(evidenceDir, "failure.json"), map[string]any{
		"classification": "verification_failed",
	})

	var doc evidenceSummaryDocument
	executeCLIJSON(t, &doc, "evidence-summary", "--json", "--evidence-dir", evidenceDir)

	if doc.Status != "verification_failed" {
		t.Fatalf("status = %q, want verification_failed", doc.Status)
	}
	assertEvidenceCheck(t, doc.Checks, "verification_json", "fail")
	assertEvidenceCheck(t, doc.Checks, "failure_json", "present")
	assertEvidenceCheck(t, doc.Checks, "output_workbook_sha256", "missing")
	if !slices.Contains(doc.NextActions, "inspect_failure_evidence_and_repair_advice") {
		t.Fatalf("next_actions missing repair action: %v", doc.NextActions)
	}
}

func TestEvidenceSummaryCommandIsExposedAsReadOnlyAgentContract(t *testing.T) {
	var capabilities cliCapabilitiesDocument
	executeCLIJSON(t, &capabilities, "capabilities", "--json")
	if !slices.Contains(capabilities.ReadOnlyCommands, "evidence-summary") {
		t.Fatalf("read_only_commands missing evidence-summary: %v", capabilities.ReadOnlyCommands)
	}
	command := findCapabilityCommand(t, capabilities.Commands, "evidence-summary")
	if command.Classification != cliClassificationAgentContract {
		t.Fatalf("classification = %q, want %q", command.Classification, cliClassificationAgentContract)
	}
	if command.Mutating || !command.ReadOnly {
		t.Fatalf("command flags = %+v, want read-only non-mutating", command)
	}

	var schema cliCommandSchemaEnvelope
	executeCLIJSON(t, &schema, "schema", "command", "evidence-summary", "--json")
	if schema.Command.Name != "evidence-summary" {
		t.Fatalf("schema command name = %q, want evidence-summary", schema.Command.Name)
	}
	if schema.Command.Mutating || !schema.Command.ReadOnly {
		t.Fatalf("schema flags = %+v, want read-only non-mutating", schema.Command)
	}
}

func assertEvidenceCheck(t *testing.T, checks []evidenceSummaryCheckView, name, status string) {
	t.Helper()

	for _, check := range checks {
		if check.Name == name {
			if check.Status != status {
				t.Fatalf("check %q status = %q, want %q: %+v", name, check.Status, status, checks)
			}
			return
		}
	}
	t.Fatalf("missing check %q: %+v", name, checks)
}
