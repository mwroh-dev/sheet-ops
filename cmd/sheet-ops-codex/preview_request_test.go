package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	runtimeschema "github.com/mwroh/sheet-ops/runtime/schema"
)

type previewRequestDocument struct {
	SchemaVersion    string                   `json:"schema_version"`
	Command          string                   `json:"command"`
	OK               bool                     `json:"ok"`
	ReadOnly         bool                     `json:"read_only"`
	DryRun           bool                     `json:"dry_run"`
	Planner          string                   `json:"planner"`
	PlanConfidence   string                   `json:"plan_confidence"`
	Operation        string                   `json:"operation"`
	WouldMutate      bool                     `json:"would_mutate"`
	MutationSummary  previewMutationSummary   `json:"mutation_summary"`
	Fingerprints     previewFingerprints      `json:"fingerprints"`
	ScenarioID       string                   `json:"scenario_id"`
	InputFile        string                   `json:"input_file"`
	OutputFile       string                   `json:"output_file"`
	StateRoot        string                   `json:"state_root"`
	PlannedReads     []previewPlannedArtifact `json:"planned_reads"`
	PlannedWrites    []previewPlannedArtifact `json:"planned_writes"`
	PlannedArtifacts []previewPlannedArtifact `json:"planned_artifacts"`
	Limitations      []string                 `json:"limitations"`
}

func TestPreviewRequestReadsNormalizedIntentFromStdin(t *testing.T) {
	projectDir := t.TempDir()
	inputFile := filepath.Join(projectDir, "line-items.xlsx")
	outputFile := filepath.Join(projectDir, "line-items-output.xlsx")
	intentFile := filepath.Join(projectDir, "append-intent.json")
	writeLineItemsWorkbook(t, inputFile)
	writeAppendRowsIntent(t, intentFile)
	intentRaw, err := os.ReadFile(intentFile)
	if err != nil {
		t.Fatalf("ReadFile(intent): %v", err)
	}

	root := newRootCommand()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&stderr)
	root.SetIn(bytes.NewReader(intentRaw))
	root.SetArgs([]string{
		"preview-request",
		"--json",
		"--intent-file", "-",
		"--input-file", inputFile,
		"--output-file", outputFile,
		"--scenario-id", "preview-stdin-append-rows",
	})
	if err := root.Execute(); err != nil {
		t.Fatalf("Execute(preview-request stdin): %v\nstderr:\n%s\nstdout:\n%s", err, stderr.String(), stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}

	var doc previewRequestDocument
	if err := json.Unmarshal(stdout.Bytes(), &doc); err != nil {
		t.Fatalf("json.Unmarshal(stdout): %v\nstdout:\n%s", err, stdout.String())
	}
	assertValidatesAgainstSchema(t, doc, "contracts/cli/preview_request_result.schema.json")
	if doc.Operation != "append_structured_rows" {
		t.Fatalf("operation = %q, want append_structured_rows", doc.Operation)
	}
	wantIntentSHA := sha256.Sum256(intentRaw)
	if doc.Fingerprints.NormalizedIntentSHA256 != hex.EncodeToString(wantIntentSHA[:]) {
		t.Fatalf("normalized_intent_sha256 = %q, want stdin bytes hash", doc.Fingerprints.NormalizedIntentSHA256)
	}
	assertPreviewArtifact(t, doc.PlannedReads, "normalized_intent", "stdin", true)
	if _, err := os.Stat(outputFile); !os.IsNotExist(err) {
		t.Fatalf("output workbook was created during stdin preview: %v", err)
	}
}

func TestLoadPreviewIntentSourceRejectsEmptyPath(t *testing.T) {
	_, err := loadPreviewIntentSource("   ", strings.NewReader(`{}`))
	if err == nil {
		t.Fatalf("loadPreviewIntentSource returned nil, want empty path error")
	}
	if !strings.Contains(err.Error(), "intent-file path cannot be empty") {
		t.Fatalf("error = %q, want explicit empty path message", err.Error())
	}
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
	if doc.Planner != "requestcompiler_validate_intent" {
		t.Fatalf("planner = %q, want requestcompiler_validate_intent", doc.Planner)
	}
	if doc.PlanConfidence != "compiler_validated_boundary" {
		t.Fatalf("plan_confidence = %q, want compiler_validated_boundary", doc.PlanConfidence)
	}
	if doc.Operation != "append_structured_rows" {
		t.Fatalf("operation = %q, want append_structured_rows", doc.Operation)
	}
	if !doc.WouldMutate {
		t.Fatalf("would_mutate = false, want true because run-intent would write workbook/artifacts")
	}
	if doc.MutationSummary.Workbook != "planned_output_workbook" {
		t.Fatalf("mutation_summary.workbook = %q, want planned_output_workbook", doc.MutationSummary.Workbook)
	}
	if doc.MutationSummary.Artifacts != "planned_runtime_artifacts" {
		t.Fatalf("mutation_summary.artifacts = %q, want planned_runtime_artifacts", doc.MutationSummary.Artifacts)
	}
	if doc.MutationSummary.State != "planned_state_root" {
		t.Fatalf("mutation_summary.state = %q, want planned_state_root", doc.MutationSummary.State)
	}
	if !doc.MutationSummary.ExecutionRequired {
		t.Fatalf("mutation_summary.execution_required = false, want true")
	}
	if len(doc.Fingerprints.NormalizedIntentSHA256) != 64 {
		t.Fatalf("normalized_intent_sha256 length = %d, want 64", len(doc.Fingerprints.NormalizedIntentSHA256))
	}
	if len(doc.Fingerprints.InputWorkbookSHA256) != 64 {
		t.Fatalf("input_workbook_sha256 length = %d, want 64", len(doc.Fingerprints.InputWorkbookSHA256))
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

func TestPreviewRequestOperationUsesCompilerSelectionForMultiCandidateIntent(t *testing.T) {
	projectDir := t.TempDir()
	inputFile := filepath.Join(projectDir, "line-items.xlsx")
	outputFile := filepath.Join(projectDir, "line-items-output.xlsx")
	intentFile := filepath.Join(projectDir, "append-intent.json")
	writeLineItemsWorkbook(t, inputFile)
	writeAppendRowsIntent(t, intentFile)
	writeJSONFile(t, intentFile, map[string]any{
		"source_sheet_candidates": []string{"LineItems"},
		"composition_candidates": []string{
			"group_summary",
			"structured_row_append",
		},
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
	})

	var doc previewRequestDocument
	executeCLIJSON(t, &doc,
		"preview-request",
		"--json",
		"--intent-file", intentFile,
		"--input-file", inputFile,
		"--output-file", outputFile,
		"--scenario-id", "preview-multi-candidate",
	)

	if doc.Operation != "append_structured_rows" {
		t.Fatalf("operation = %q, want compiler-selected append_structured_rows", doc.Operation)
	}
}

func TestPreviewRequestUnresolvedDecisionDoesNotClaimMutation(t *testing.T) {
	projectDir := t.TempDir()
	inputFile := filepath.Join(projectDir, "line-items.xlsx")
	outputFile := filepath.Join(projectDir, "line-items-output.xlsx")
	intentFile := filepath.Join(projectDir, "conflict-intent.json")
	writeLineItemsWorkbook(t, inputFile)
	writeJSONFile(t, intentFile, map[string]any{
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
			"markers":           []string{"source_truth_conflict"},
			"unresolved_fields": []string{},
			"checkpoint_hints":  []string{"source_truth_conflict"},
		},
	})

	var doc previewRequestDocument
	executeCLIJSON(t, &doc,
		"preview-request",
		"--json",
		"--intent-file", intentFile,
		"--input-file", inputFile,
		"--output-file", outputFile,
		"--scenario-id", "preview-source-truth-conflict",
	)
	assertValidatesAgainstSchema(t, doc, "contracts/cli/preview_request_result.schema.json")

	if doc.Operation != "unresolved" {
		t.Fatalf("operation = %q, want unresolved", doc.Operation)
	}
	if doc.WouldMutate {
		t.Fatalf("would_mutate = true, want false for unresolved preview decision")
	}
	if doc.MutationSummary.Workbook != "none" {
		t.Fatalf("mutation_summary.workbook = %q, want none", doc.MutationSummary.Workbook)
	}
	if doc.MutationSummary.Artifacts != "none" {
		t.Fatalf("mutation_summary.artifacts = %q, want none", doc.MutationSummary.Artifacts)
	}
	if doc.MutationSummary.State != "none" {
		t.Fatalf("mutation_summary.state = %q, want none", doc.MutationSummary.State)
	}
	if doc.MutationSummary.ExecutionRequired {
		t.Fatalf("mutation_summary.execution_required = true, want false for unresolved preview decision")
	}
	if len(doc.PlannedWrites) != 0 {
		t.Fatalf("planned_writes = %+v, want empty for unresolved preview decision", doc.PlannedWrites)
	}
	if len(doc.PlannedArtifacts) != 0 {
		t.Fatalf("planned_artifacts = %+v, want empty for unresolved preview decision", doc.PlannedArtifacts)
	}
}

func TestPreviewRequestSchemaRejectsContradictoryNoMutationPlan(t *testing.T) {
	projectDir := t.TempDir()
	inputFile := filepath.Join(projectDir, "line-items.xlsx")
	outputFile := filepath.Join(projectDir, "line-items-output.xlsx")
	intentFile := filepath.Join(projectDir, "conflict-intent.json")
	writeLineItemsWorkbook(t, inputFile)
	writeAppendRowsIntent(t, intentFile)

	var doc previewRequestDocument
	executeCLIJSON(t, &doc,
		"preview-request",
		"--json",
		"--intent-file", intentFile,
		"--input-file", inputFile,
		"--output-file", outputFile,
		"--scenario-id", "preview-contradictory-schema",
	)
	document := previewRequestSchemaDocument(t, doc)
	document["would_mutate"] = false
	document["mutation_summary"] = map[string]any{
		"workbook":           "none",
		"artifacts":          "none",
		"state":              "none",
		"execution_required": false,
	}

	if err := assertValidateStruct("contracts/cli/preview_request_result.schema.json", document); err == nil {
		t.Fatalf("preview schema accepted no-mutation result with planned writes/artifacts")
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
	for _, want := range []string{
		"Input workbook facts and bytes",
		"SHEET_OPS_STATE_ROOT",
		"SHEET_OPS_ARTIFACT_ROOT",
		"SHEET_OPS_KNOWLEDGE_ROOT",
	} {
		if !strings.Contains(strings.Join(schema.Command.ReadArtifacts, "\n"), want) {
			t.Fatalf("preview-request read_artifacts = %v, want substring %q", schema.Command.ReadArtifacts, want)
		}
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

func writeJSONFile(t *testing.T, path string, value any) {
	t.Helper()

	raw, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		t.Fatalf("MarshalIndent(%s): %v", path, err)
	}
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		t.Fatalf("WriteFile(%s): %v", path, err)
	}
}

func previewRequestSchemaDocument(t *testing.T, doc previewRequestDocument) map[string]any {
	t.Helper()

	raw, err := json.Marshal(doc)
	if err != nil {
		t.Fatalf("Marshal preview document: %v", err)
	}
	var document map[string]any
	if err := json.Unmarshal(raw, &document); err != nil {
		t.Fatalf("Unmarshal preview document: %v", err)
	}
	return document
}

func assertValidateStruct(schemaRel string, document any) error {
	schemaPath := filepath.Join("..", "..", filepath.FromSlash(schemaRel))
	return runtimeschema.ValidateStruct(schemaPath, document)
}
