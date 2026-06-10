package main

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

type previewRequestDocument struct {
	SchemaVersion    string                   `json:"schema_version"`
	Command          string                   `json:"command"`
	OK               bool                     `json:"ok"`
	ReadOnly         bool                     `json:"read_only"`
	DryRun           bool                     `json:"dry_run"`
	ScenarioID       string                   `json:"scenario_id"`
	InputFile        string                   `json:"input_file"`
	OutputFile       string                   `json:"output_file"`
	StateRoot        string                   `json:"state_root"`
	PlannedReads     []previewPlannedArtifact `json:"planned_reads"`
	PlannedWrites    []previewPlannedArtifact `json:"planned_writes"`
	PlannedArtifacts []previewPlannedArtifact `json:"planned_artifacts"`
	Limitations      []string                 `json:"limitations"`
}

func TestPreviewRequestReportsImpactWithoutMutating(t *testing.T) {
	projectDir := t.TempDir()
	inputFile := filepath.Join(projectDir, "line-items.xlsx")
	outputFile := filepath.Join(projectDir, "line-items-output.xlsx")
	intentFile := filepath.Join(projectDir, "append-intent.json")
	writeLineItemsWorkbook(t, inputFile)
	writeAppendRowsIntent(t, intentFile)

	var doc previewRequestDocument
	executeCLIJSON(t, &doc,
		"preview-request",
		"--json",
		"--intent-file", intentFile,
		"--input-file", inputFile,
		"--output-file", outputFile,
		"--scenario-id", "preview-append-rows",
	)
	assertValidatesAgainstSchema(t, doc, "contracts/cli/preview_request_result.schema.json")

	if doc.SchemaVersion != cliContractSchemaVersion {
		t.Fatalf("schema_version = %q, want %q", doc.SchemaVersion, cliContractSchemaVersion)
	}
	if !doc.OK {
		t.Fatalf("ok = false, want true")
	}
	if doc.Command != "preview-request" {
		t.Fatalf("command = %q, want preview-request", doc.Command)
	}
	if !doc.ReadOnly {
		t.Fatalf("read_only = false, want true")
	}
	if doc.DryRun {
		t.Fatalf("dry_run = true, want false because preview is not runtime dry-run")
	}
	if doc.ScenarioID != "preview-append-rows" {
		t.Fatalf("scenario_id = %q, want preview-append-rows", doc.ScenarioID)
	}
	if doc.InputFile != inputFile {
		t.Fatalf("input_file = %q, want %q", doc.InputFile, inputFile)
	}
	if doc.OutputFile != outputFile {
		t.Fatalf("output_file = %q, want %q", doc.OutputFile, outputFile)
	}
	if doc.StateRoot != filepath.Join(projectDir, ".sheet-ops-state") {
		t.Fatalf("state_root = %q, want project state root", doc.StateRoot)
	}
	assertPreviewArtifact(t, doc.PlannedReads, "normalized_intent", intentFile, true)
	assertPreviewArtifact(t, doc.PlannedReads, "input_workbook", inputFile, true)
	assertPreviewArtifact(t, doc.PlannedWrites, "output_workbook", outputFile, true)
	assertPreviewArtifact(t, doc.PlannedArtifacts, "state_root", filepath.Join(projectDir, ".sheet-ops-state"), true)
	if len(doc.Limitations) == 0 {
		t.Fatalf("limitations is empty")
	}
	if _, err := os.Stat(outputFile); !os.IsNotExist(err) {
		t.Fatalf("output workbook was created during preview: %v", err)
	}
	if _, err := os.Stat(filepath.Join(projectDir, ".sheet-ops-state")); !os.IsNotExist(err) {
		t.Fatalf("state root was created during preview: %v", err)
	}
}

func TestPreviewRequestIsExposedAsReadOnlyNonDryRunCommand(t *testing.T) {
	var capabilities cliCapabilitiesDocument
	executeCLIJSON(t, &capabilities, "capabilities", "--json")
	agentGroup := findCapabilityGroup(t, capabilities.CommandGroups, "agent_contract")
	if !slices.Contains(agentGroup.Commands, "preview-request") {
		t.Fatalf("agent_contract commands missing preview-request: %v", agentGroup.Commands)
	}

	var schema cliCommandSchemaEnvelope
	executeCLIJSON(t, &schema, "schema", "command", "preview-request", "--json")
	if schema.Command.Name != "preview-request" {
		t.Fatalf("command.name = %q, want preview-request", schema.Command.Name)
	}
	if schema.Command.Classification != cliClassificationAgentContract {
		t.Fatalf("classification = %q, want %q", schema.Command.Classification, cliClassificationAgentContract)
	}
	if schema.Command.Mutating {
		t.Fatalf("preview-request mutating = true")
	}
	if !schema.Command.ReadOnly {
		t.Fatalf("preview-request read_only = false")
	}
	if schema.Command.DryRunCapable {
		t.Fatalf("preview-request dry_run_capable = true, want false")
	}
}

func assertPreviewArtifact(t *testing.T, artifacts []previewPlannedArtifact, kind string, path string, required bool) {
	t.Helper()

	for _, artifact := range artifacts {
		if artifact.Kind != kind {
			continue
		}
		if artifact.Path != path {
			t.Fatalf("artifact %q path = %q, want %q", kind, artifact.Path, path)
		}
		if artifact.Required != required {
			t.Fatalf("artifact %q required = %v, want %v", kind, artifact.Required, required)
		}
		return
	}
	t.Fatalf("missing artifact kind %q in %+v", kind, artifacts)
}
