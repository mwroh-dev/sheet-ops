package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	runtimeschema "github.com/mwroh/sheet-ops/runtime/schema"
)

type operationListDocument struct {
	SchemaVersion string                 `json:"schema_version"`
	Command       string                 `json:"command"`
	OK            bool                   `json:"ok"`
	Operations    []operationListItemDoc `json:"operations"`
}

type operationListItemDoc struct {
	Name           string   `json:"name"`
	TaskOperations []string `json:"task_operations"`
	Exposure       string   `json:"exposure"`
	Status         string   `json:"status"`
}

type operationSchemaDocument struct {
	SchemaVersion string             `json:"schema_version"`
	Command       string             `json:"command"`
	OK            bool               `json:"ok"`
	Operation     operationSchemaDoc `json:"operation"`
}

type operationSchemaDoc struct {
	Name             string   `json:"name"`
	TaskOperations   []string `json:"task_operations"`
	Inputs           []string `json:"inputs"`
	Outputs          []string `json:"outputs"`
	RequiredFacts    []string `json:"required_facts"`
	Verifies         []string `json:"verifies"`
	ContractPaths    []string `json:"contract_paths"`
	RuntimePaths     []string `json:"runtime_paths"`
	Examples         []string `json:"examples"`
	SchemaReferences []string `json:"schema_references"`
	Status           string   `json:"status"`
	Exposure         string   `json:"exposure"`
}

type operationExampleDocument struct {
	SchemaVersion string         `json:"schema_version"`
	Command       string         `json:"command"`
	OK            bool           `json:"ok"`
	Operation     string         `json:"operation"`
	Example       map[string]any `json:"example"`
	SourcePath    string         `json:"source_path"`
}

func TestOperationListJSONReportsPublicAgentOperations(t *testing.T) {
	var doc operationListDocument
	executeCLIJSON(t, &doc, "operation", "list", "--json")

	if doc.SchemaVersion != cliContractSchemaVersion {
		t.Fatalf("schema_version = %q, want %q", doc.SchemaVersion, cliContractSchemaVersion)
	}
	if doc.Command != "operation list" {
		t.Fatalf("command = %q, want operation list", doc.Command)
	}
	if !doc.OK {
		t.Fatalf("ok = false, want true")
	}
	if len(doc.Operations) < 10 {
		t.Fatalf("operations count = %d, want at least 10", len(doc.Operations))
	}
	names := make([]string, 0, len(doc.Operations))
	for _, operation := range doc.Operations {
		names = append(names, operation.Name)
		if operation.Status != "supported" {
			t.Fatalf("%s status = %q, want supported", operation.Name, operation.Status)
		}
		if operation.Exposure != "public_agent_capability" {
			t.Fatalf("%s exposure = %q, want public_agent_capability", operation.Name, operation.Exposure)
		}
	}
	for _, want := range []string{"append_structured_rows", "group_summarize", "highlight_threshold", "join_lookup"} {
		if !slices.Contains(names, want) {
			t.Fatalf("operation list missing %q: %v", want, names)
		}
	}
	if slices.Contains(names, "write_values") {
		t.Fatalf("operation list exposed runtime primitive write_values: %v", names)
	}
}

func TestOperationListUsesPackageRootWhenWorkingDirectoryIsProject(t *testing.T) {
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	repoRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("filepath.Abs(repo root): %v", err)
	}
	projectDir := t.TempDir()
	t.Setenv("SHEET_OPS_PACKAGE_ROOT", repoRoot)
	if err := os.Chdir(projectDir); err != nil {
		t.Fatalf("Chdir(%s): %v", projectDir, err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Fatalf("restore chdir(%s): %v", originalDir, err)
		}
	})

	var doc operationListDocument
	executeCLIJSON(t, &doc, "operation", "list", "--json")

	if !doc.OK {
		t.Fatalf("ok = false, want true")
	}
	if len(doc.Operations) == 0 {
		t.Fatalf("operations is empty")
	}
}

func TestOperationListAcceptsRelativePackageRoot(t *testing.T) {
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	repoRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("filepath.Abs(repo root): %v", err)
	}
	parentDir := filepath.Dir(repoRoot)
	relativeRoot, err := filepath.Rel(parentDir, repoRoot)
	if err != nil {
		t.Fatalf("filepath.Rel(repo root): %v", err)
	}
	t.Setenv("SHEET_OPS_PACKAGE_ROOT", relativeRoot)
	if err := os.Chdir(parentDir); err != nil {
		t.Fatalf("Chdir(%s): %v", parentDir, err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Fatalf("restore chdir(%s): %v", originalDir, err)
		}
	})

	var doc operationListDocument
	executeCLIJSON(t, &doc, "operation", "list", "--json")

	if !doc.OK {
		t.Fatalf("ok = false, want true")
	}
	if len(doc.Operations) == 0 {
		t.Fatalf("operations is empty")
	}
}

func TestOperationListIgnoresBlankPackageRootEnv(t *testing.T) {
	t.Setenv("SHEET_OPS_PACKAGE_ROOT", "   ")

	var doc operationListDocument
	executeCLIJSON(t, &doc, "operation", "list", "--json")

	if !doc.OK {
		t.Fatalf("ok = false, want true")
	}
	if len(doc.Operations) == 0 {
		t.Fatalf("operations is empty")
	}
}

func TestOperationCommandsSerializeRegistryLoadErrorsAsJSON(t *testing.T) {
	for _, tc := range []struct {
		name string
		args []string
	}{
		{name: "list", args: []string{"operation", "list", "--json"}},
		{name: "schema", args: []string{"operation", "schema", "append_structured_rows", "--json"}},
		{name: "example", args: []string{"operation", "example", "append_structured_rows", "--json"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("SHEET_OPS_PACKAGE_ROOT", t.TempDir())

			root := newRootCommand()
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			root.SetOut(&stdout)
			root.SetErr(&stderr)
			root.SetArgs(tc.args)

			err := root.Execute()
			if err == nil {
				t.Fatalf("Execute(%v) returned nil, want registry load error", tc.args)
			}
			if stderr.Len() != 0 {
				t.Fatalf("stderr = %q, want empty", stderr.String())
			}
			var envelope cliErrorEnvelope
			if decodeErr := json.Unmarshal(stdout.Bytes(), &envelope); decodeErr != nil {
				t.Fatalf("json.Unmarshal error envelope: %v\nstdout:\n%s", decodeErr, stdout.String())
			}
			if envelope.OK {
				t.Fatalf("envelope.ok = true, want false")
			}
			if envelope.Error.Code != cliErrorInternal {
				t.Fatalf("error.code = %q, want %q", envelope.Error.Code, cliErrorInternal)
			}
			if !strings.Contains(envelope.Error.Message, "SHEET_OPS_PACKAGE_ROOT") {
				t.Fatalf("error.message = %q, want SHEET_OPS_PACKAGE_ROOT context", envelope.Error.Message)
			}
		})
	}
}

func TestOperationExampleRejectsMissingDetailedExample(t *testing.T) {
	stdout, stderr, err := executeCLIExpectError(t, "operation", "example", "extend_table_formulas", "--json")
	if err == nil {
		t.Fatalf("Execute(operation example extend_table_formulas) returned nil, want missing example error")
	}
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty for JSON error", stderr)
	}
	var envelope cliErrorEnvelope
	if decodeErr := json.Unmarshal([]byte(stdout), &envelope); decodeErr != nil {
		t.Fatalf("json.Unmarshal error envelope: %v\nstdout:\n%s", decodeErr, stdout)
	}
	if envelope.OK {
		t.Fatalf("envelope.ok = true, want false")
	}
	if envelope.Error.Code != cliErrorInvalidData {
		t.Fatalf("error.code = %q, want %q", envelope.Error.Code, cliErrorInvalidData)
	}
	if !strings.Contains(envelope.Error.Message, "no normalized intent example") {
		t.Fatalf("error.message = %q, want missing normalized intent example", envelope.Error.Message)
	}
}

func TestOperationSchemaJSONReportsOneOperationContract(t *testing.T) {
	var doc operationSchemaDocument
	executeCLIJSON(t, &doc, "operation", "schema", "append_structured_rows", "--json")

	if doc.Command != "operation schema" {
		t.Fatalf("command = %q, want operation schema", doc.Command)
	}
	if doc.Operation.Name != "append_structured_rows" {
		t.Fatalf("operation.name = %q, want append_structured_rows", doc.Operation.Name)
	}
	for _, want := range []string{"input_workbook", "source_sheet", "values"} {
		if !slices.Contains(doc.Operation.Inputs, want) {
			t.Fatalf("operation.inputs missing %q: %v", want, doc.Operation.Inputs)
		}
	}
	if !slices.Contains(doc.Operation.SchemaReferences, "contracts/task/task_spec.schema.json") {
		t.Fatalf("schema_references missing task spec: %v", doc.Operation.SchemaReferences)
	}
	if len(doc.Operation.Examples) == 0 {
		t.Fatalf("examples is empty")
	}
}

func TestOperationExampleJSONReportsValidNormalizedIntentExample(t *testing.T) {
	var doc operationExampleDocument
	executeCLIJSON(t, &doc, "operation", "example", "append_structured_rows", "--json")

	if doc.Command != "operation example" {
		t.Fatalf("command = %q, want operation example", doc.Command)
	}
	if doc.Operation != "append_structured_rows" {
		t.Fatalf("operation = %q, want append_structured_rows", doc.Operation)
	}
	if doc.SourcePath == "" {
		t.Fatalf("source_path is empty")
	}
	raw, err := json.Marshal(doc.Example)
	if err != nil {
		t.Fatalf("json.Marshal(example): %v", err)
	}
	var normalizedIntent map[string]any
	if err := json.Unmarshal(raw, &normalizedIntent); err != nil {
		t.Fatalf("json.Unmarshal(example): %v", err)
	}
	if err := runtimeschema.ValidateStruct("../../agents/request-compiler/contract/normalized_intent.schema.json", normalizedIntent); err != nil {
		t.Fatalf("example failed normalized intent schema validation: %v\nexample=%s", err, string(raw))
	}
}

func TestOperationCommandIsExposedAsReadOnlyAgentContract(t *testing.T) {
	var capabilities cliCapabilitiesDocument
	executeCLIJSON(t, &capabilities, "capabilities", "--json")
	if !slices.Contains(capabilities.ReadOnlyCommands, "operation") {
		t.Fatalf("read_only_commands missing operation: %v", capabilities.ReadOnlyCommands)
	}
	operationCommand := findCapabilityCommand(t, capabilities.Commands, "operation")
	if operationCommand.Classification != cliClassificationAgentContract {
		t.Fatalf("operation classification = %q, want %q", operationCommand.Classification, cliClassificationAgentContract)
	}
	if operationCommand.Mutating || !operationCommand.ReadOnly {
		t.Fatalf("operation command flags = %+v, want read-only non-mutating", operationCommand)
	}

	var schema cliCommandSchemaEnvelope
	executeCLIJSON(t, &schema, "schema", "command", "operation", "--json")
	if schema.Command.Name != "operation" {
		t.Fatalf("schema command name = %q, want operation", schema.Command.Name)
	}
	if schema.Command.Mutating || !schema.Command.ReadOnly {
		t.Fatalf("schema operation flags = %+v, want read-only non-mutating", schema.Command)
	}
}
