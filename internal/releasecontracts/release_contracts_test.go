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
	cases := collectDeclaredExampleSchemas(t, root)

	examplePaths := collectExamplePaths(t, root)
	if len(examplePaths) != len(cases) {
		t.Fatalf("validated example count=%d want %d", len(cases), len(examplePaths))
	}

	for _, examplePath := range examplePaths {
		rel := filepath.ToSlash(mustRel(t, root, examplePath))
		t.Run(rel, func(t *testing.T) {
			schemaRel, ok := cases[rel]
			if !ok {
				t.Fatalf("%s is not mapped to a contract schema", rel)
			}
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

	validatedExamples := collectDeclaredExampleSchemas(t, root)

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

	for _, schemaPath := range schemaPaths {
		rel := filepath.ToSlash(mustRel(t, root, schemaPath))
		t.Run(rel, func(t *testing.T) {
			if _, err := compiler.Compile(schemaPath); err != nil {
				t.Fatalf("compile schema: %v", err)
			}
		})
	}
}

func collectDeclaredExampleSchemas(t *testing.T, root string) map[string]string {
	t.Helper()
	registry, err := capabilities.LoadRegistry(root)
	if err != nil {
		t.Fatalf("LoadRegistry: %v", err)
	}

	cases := map[string]string{}
	for _, capability := range registry.Capabilities {
		for _, example := range capability.Examples {
			examplePath := filepath.Join(root, filepath.FromSlash(example))
			schemaRel := inferExampleSchema(t, examplePath)
			if existing, ok := cases[example]; ok && existing != schemaRel {
				t.Fatalf("%s maps to both %s and %s", example, existing, schemaRel)
			}
			cases[example] = schemaRel
		}
	}
	return cases
}

func inferExampleSchema(t *testing.T, examplePath string) string {
	t.Helper()
	document := readJSON(t, examplePath)
	object, ok := document.(map[string]any)
	if !ok {
		t.Fatalf("%s is not a JSON object", examplePath)
	}

	if _, ok := object["operation_family"]; ok {
		return "contracts/ir/workbook_operation_ir.schema.json"
	}
	if _, ok := object["composition_kind"]; ok {
		return "contracts/requests/validated_execution_request.schema.json"
	}
	if _, ok := object["operation"]; ok {
		return "contracts/requests/structured_use_request.schema.json"
	}
	t.Fatalf("%s does not match a known public example shape", examplePath)
	return ""
}

func collectExamplePaths(t *testing.T, root string) []string {
	t.Helper()
	var examplePaths []string
	base := filepath.Join(root, "examples")
	if err := filepath.WalkDir(base, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".json") {
			examplePaths = append(examplePaths, path)
		}
		return nil
	}); err != nil {
		t.Fatalf("walk examples: %v", err)
	}
	sort.Strings(examplePaths)
	return examplePaths
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
