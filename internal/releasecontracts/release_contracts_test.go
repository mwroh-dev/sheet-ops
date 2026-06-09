package releasecontracts

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"

	"github.com/mwroh/sheet-ops/runtime/capabilities"
	runtimeschema "github.com/mwroh/sheet-ops/runtime/schema"
	"github.com/santhosh-tekuri/jsonschema/v5"
)

func TestPublicExamplesValidateAgainstDeclaredContracts(t *testing.T) {
	root := repoRoot(t)
	cases := map[string]string{
		"examples/group_summarize/formulas_request.json":        "contracts/requests/structured_use_request.schema.json",
		"examples/group_summarize/values_request.json":          "contracts/requests/structured_use_request.schema.json",
		"examples/highlight_threshold/request.json":             "contracts/requests/structured_use_request.schema.json",
		"examples/join_lookup/structured_use_request.json":      "contracts/requests/structured_use_request.schema.json",
		"examples/join_lookup/validated_execution_request.json": "contracts/requests/validated_execution_request.schema.json",
		"examples/write_values/workbook_operation_ir.json":      "contracts/ir/workbook_operation_ir.schema.json",
	}

	examplePaths, err := filepath.Glob(filepath.Join(root, "examples", "*", "*.json"))
	if err != nil {
		t.Fatalf("glob examples: %v", err)
	}
	if len(examplePaths) != len(cases) {
		t.Fatalf("validated example count=%d want %d", len(cases), len(examplePaths))
	}

	for _, examplePath := range examplePaths {
		rel := filepath.ToSlash(mustRel(t, root, examplePath))
		schemaRel, ok := cases[rel]
		if !ok {
			t.Fatalf("%s is not mapped to a contract schema", rel)
		}
		t.Run(rel, func(t *testing.T) {
			document := readJSON(t, examplePath)
			schemaPath := filepath.Join(root, filepath.FromSlash(schemaRel))
			if err := runtimeschema.ValidateStruct(schemaPath, document); err != nil {
				t.Fatalf("validate %s against %s: %v", rel, schemaRel, err)
			}
		})
	}
}

func TestCapabilityRecordsOnlyReferenceValidatedPublicExamples(t *testing.T) {
	root := repoRoot(t)
	registry, err := capabilities.LoadRegistry(root)
	if err != nil {
		t.Fatalf("LoadRegistry: %v", err)
	}

	validatedExamples := map[string]struct{}{}
	examplePaths, err := filepath.Glob(filepath.Join(root, "examples", "*", "*.json"))
	if err != nil {
		t.Fatalf("glob examples: %v", err)
	}
	for _, examplePath := range examplePaths {
		validatedExamples[filepath.ToSlash(mustRel(t, root, examplePath))] = struct{}{}
	}

	for _, capability := range registry.Capabilities {
		if len(capability.Examples) == 0 {
			t.Fatalf("%s has no public examples", capability.Name)
		}
		for _, example := range capability.Examples {
			if _, ok := validatedExamples[example]; !ok {
				t.Fatalf("%s references unvalidated example %s", capability.Name, example)
			}
		}
	}
}

func TestReleaseSchemasCompile(t *testing.T) {
	root := repoRoot(t)
	schemaPaths := collectSchemaPaths(t, root, "contracts")
	schemaPaths = append(schemaPaths, collectSchemaPaths(t, root, "agents")...)
	sort.Strings(schemaPaths)

	if len(schemaPaths) == 0 {
		t.Fatal("no release schemas found")
	}
	for _, schemaPath := range schemaPaths {
		rel := filepath.ToSlash(mustRel(t, root, schemaPath))
		t.Run(rel, func(t *testing.T) {
			compiler := compilerWithLocalSchemaIDs(t, schemaPaths)
			if _, err := compiler.Compile(schemaPath); err != nil {
				t.Fatalf("compile schema: %v", err)
			}
		})
	}
}

func collectSchemaPaths(t *testing.T, root, relDir string) []string {
	t.Helper()
	var schemaPaths []string
	base := filepath.Join(root, relDir)
	if err := filepath.WalkDir(base, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() && entry.Name() == "node_modules" {
			return filepath.SkipDir
		}
		if entry.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".schema.json") {
			schemaPaths = append(schemaPaths, path)
		}
		return nil
	}); err != nil {
		t.Fatalf("walk %s schemas: %v", relDir, err)
	}
	return schemaPaths
}

func readJSON(t *testing.T, path string) any {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	var document any
	if err := json.Unmarshal(raw, &document); err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
	return document
}

func compilerWithLocalSchemaIDs(t *testing.T, schemaPaths []string) *jsonschema.Compiler {
	t.Helper()
	compiler := jsonschema.NewCompiler()
	for _, schemaPath := range schemaPaths {
		raw, err := os.ReadFile(schemaPath)
		if err != nil {
			t.Fatalf("read schema %s: %v", schemaPath, err)
		}
		var header struct {
			ID string `json:"$id"`
		}
		if err := json.Unmarshal(raw, &header); err != nil {
			t.Fatalf("parse schema header %s: %v", schemaPath, err)
		}
		if header.ID == "" {
			continue
		}
		if err := compiler.AddResource(header.ID, bytes.NewReader(raw)); err != nil {
			t.Fatalf("register schema %s as %s: %v", schemaPath, header.ID, err)
		}
	}
	return compiler
}

func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}

func mustRel(t *testing.T, base, target string) string {
	t.Helper()
	rel, err := filepath.Rel(base, target)
	if err != nil {
		t.Fatalf("rel(%s, %s): %v", base, target, err)
	}
	return rel
}
