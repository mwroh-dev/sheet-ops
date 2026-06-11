package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	runtimeschema "github.com/mwroh/sheet-ops/runtime/schema"
)

func TestCLIJSONOutputsValidateAgainstPublishedSchemas(t *testing.T) {
	tests := []struct {
		name      string
		schemaRel string
		args      []string
		setup     func(*testing.T) []string
	}{
		{
			name:      "capabilities",
			schemaRel: "contracts/cli/capabilities.schema.json",
			args:      []string{"capabilities", "--json"},
		},
		{
			name:      "schema root",
			schemaRel: "contracts/cli/command_schema.schema.json",
			args:      []string{"schema", "--json"},
		},
		{
			name:      "schema command",
			schemaRel: "contracts/cli/command_schema.schema.json",
			args:      []string{"schema", "command", "preflight", "--json"},
		},
		{
			name:      "preflight",
			schemaRel: "contracts/cli/preflight_result.schema.json",
			args:      []string{"preflight", "--json", "--project", "."},
		},
		{
			name:      "preview request",
			schemaRel: "contracts/cli/preview_request_result.schema.json",
			setup: func(t *testing.T) []string {
				t.Helper()

				projectDir := t.TempDir()
				inputFile := filepath.Join(projectDir, "line-items.xlsx")
				outputFile := filepath.Join(projectDir, "line-items-output.xlsx")
				intentFile := filepath.Join(projectDir, "append-intent.json")
				writeLineItemsWorkbook(t, inputFile)
				writeAppendRowsIntent(t, intentFile)
				return []string{
					"preview-request",
					"--json",
					"--intent-file", intentFile,
					"--input-file", inputFile,
					"--output-file", outputFile,
					"--scenario-id", "contract-preview-append-rows",
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := tt.args
			if tt.setup != nil {
				args = tt.setup(t)
			}
			var document any
			executeCLIJSON(t, &document, args...)
			assertValidatesAgainstSchema(t, document, tt.schemaRel)
		})
	}
}

func TestCLIErrorEnvelopeValidatesAgainstPublishedSchema(t *testing.T) {
	stdout, stderr, err := executeCLIExpectError(t, "schema", "command", "nope", "--json")
	if err == nil {
		t.Fatalf("schema command nope returned nil error")
	}
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty for JSON error", stderr)
	}
	var document any
	if err := json.Unmarshal([]byte(stdout), &document); err != nil {
		t.Fatalf("json.Unmarshal(error envelope): %v\nstdout:\n%s", err, stdout)
	}
	assertValidatesAgainstSchema(t, document, "contracts/cli/error_envelope.schema.json")
}

func TestPublishedCLISchemasPinSchemaVersion(t *testing.T) {
	for _, schemaRel := range []string{
		"contracts/cli/capabilities.schema.json",
		"contracts/cli/command_schema.schema.json",
		"contracts/cli/error_envelope.schema.json",
		"contracts/cli/internal_handoff_result.schema.json",
		"contracts/cli/preflight_result.schema.json",
		"contracts/cli/preview_request_result.schema.json",
		"contracts/results/public_entry_result.schema.json",
	} {
		t.Run(schemaRel, func(t *testing.T) {
			schemaPath := filepath.Join("..", "..", filepath.FromSlash(schemaRel))
			raw, err := os.ReadFile(schemaPath)
			if err != nil {
				t.Fatalf("ReadFile(%s): %v", schemaRel, err)
			}
			var document map[string]any
			if err := json.Unmarshal(raw, &document); err != nil {
				t.Fatalf("json.Unmarshal(%s): %v", schemaRel, err)
			}
			properties, ok := document["properties"].(map[string]any)
			if !ok {
				t.Fatalf("%s missing properties", schemaRel)
			}
			schemaVersion, ok := properties["schema_version"].(map[string]any)
			if !ok {
				t.Fatalf("%s missing properties.schema_version", schemaRel)
			}
			if got := schemaVersion["const"]; got != cliContractSchemaVersion {
				t.Fatalf("%s schema_version const = %v, want %q", schemaRel, got, cliContractSchemaVersion)
			}
		})
	}
}

func assertValidatesAgainstSchema(t *testing.T, document any, schemaRel string) {
	t.Helper()

	schemaPath := filepath.Join("..", "..", filepath.FromSlash(schemaRel))
	if err := runtimeschema.ValidateStruct(schemaPath, document); err != nil {
		t.Fatalf("validate against %s: %v", schemaRel, err)
	}
}
