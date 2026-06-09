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

func TestReleaseSchemaReferencesAreLocallyResolvable(t *testing.T) {
	root := repoRoot(t)
	schemaPaths := collectReleaseSchemaPaths(t, root)

	for _, schemaPath := range schemaPaths {
		rel := filepath.ToSlash(mustRel(t, root, schemaPath))
		t.Run(rel, func(t *testing.T) {
			assertSchemaReferencesResolvable(t, root, schemaPath)
		})
	}
}

func TestBundledSchemasStayInSyncWithSourceContracts(t *testing.T) {
	root := repoRoot(t)
	assertBundledSchemaMatches(t, root, "contracts/requests/organism_execution_request.schema.json")
	assertBundledSchemaMatches(t, root, "contracts/requests/request_ref.schema.json")
	assertBundledSchemaMatches(t, root, "contracts/requests/use_envelope.schema.json")
	assertBundledSchemaMatches(t, root, "contracts/requests/use_envelope_v2.schema.json")
}

func TestSharedSchemaDefinitionsDoNotDrift(t *testing.T) {
	root := repoRoot(t)
	assertSharedDefinitionEqual(t, root, "cell_value",
		"contracts/ir/workbook_operation_ir.schema.json",
		"contracts/task/task_spec.schema.json",
		"contracts/validation/validation_result.schema.json",
		"agents/request-compiler/contract/normalized_intent.schema.json",
		"contracts/requests/organism_execution_request.schema.json",
	)
	assertSharedDefinitionEqual(t, root, "compare_mapping",
		"contracts/ir/workbook_operation_ir.schema.json",
		"contracts/task/task_spec.schema.json",
		"contracts/requests/use_request.schema.json",
		"contracts/requests/validated_execution_request.schema.json",
		"contracts/validation/validation_result.schema.json",
		"agents/request-compiler/contract/normalized_intent.schema.json",
		"contracts/requests/organism_execution_request.schema.json",
	)
	assertSharedDefinitionEqual(t, root, "form_field_binding",
		"contracts/ir/workbook_operation_ir.schema.json",
		"contracts/task/task_spec.schema.json",
		"contracts/requests/use_request.schema.json",
		"contracts/requests/validated_execution_request.schema.json",
		"contracts/validation/validation_result.schema.json",
		"agents/request-compiler/contract/normalized_intent.schema.json",
		"contracts/plans/use/operation_plan.schema.json",
		"contracts/requests/organism_execution_request.schema.json",
	)
	assertSharedDefinitionEqual(t, root, "form_table_binding",
		"contracts/ir/workbook_operation_ir.schema.json",
		"contracts/task/task_spec.schema.json",
		"contracts/requests/use_request.schema.json",
		"contracts/requests/validated_execution_request.schema.json",
		"contracts/validation/validation_result.schema.json",
		"agents/request-compiler/contract/normalized_intent.schema.json",
		"contracts/plans/use/operation_plan.schema.json",
		"contracts/requests/organism_execution_request.schema.json",
	)
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

func collectReleaseSchemaPaths(t *testing.T, root string) []string {
	t.Helper()
	seen := map[string]bool{}
	var schemaPaths []string
	for _, relDir := range []string{
		"contracts",
		"agents",
		filepath.Join("cmd", "sheet-ops-codex", "skill_assets", "agent-system", "contracts"),
	} {
		for _, schemaPath := range collectSchemaPaths(t, root, relDir) {
			if seen[schemaPath] {
				continue
			}
			seen[schemaPath] = true
			schemaPaths = append(schemaPaths, schemaPath)
		}
	}
	sort.Strings(schemaPaths)
	return schemaPaths
}

func assertSchemaReferencesResolvable(t *testing.T, root string, schemaPath string) {
	t.Helper()
	document := readJSON(t, schemaPath)
	for _, ref := range collectSchemaRefs(document) {
		switch {
		case strings.HasPrefix(ref, "#/"):
			if !jsonPointerExists(document, strings.TrimPrefix(ref, "#")) {
				t.Fatalf("%s references missing local pointer %s", schemaPath, ref)
			}
		case strings.HasPrefix(ref, "https://sheet-ops.local/"):
			if !schemaIDExists(t, root, ref) {
				t.Fatalf("%s references unknown schema id %s", schemaPath, ref)
			}
		case strings.Contains(ref, "://"):
			continue
		default:
			refPath, fragment, _ := strings.Cut(ref, "#")
			targetPath := filepath.Clean(filepath.Join(filepath.Dir(schemaPath), filepath.FromSlash(refPath)))
			info, err := os.Stat(targetPath)
			if err != nil || info.IsDir() {
				t.Fatalf("%s references missing relative schema %s resolved to %s", schemaPath, ref, targetPath)
			}
			if strings.HasPrefix(fragment, "/") {
				targetDocument := readJSON(t, targetPath)
				if !jsonPointerExists(targetDocument, fragment) {
					t.Fatalf("%s references missing relative pointer %s in %s", schemaPath, fragment, targetPath)
				}
			}
		}
	}
}

func collectSchemaRefs(value any) []string {
	var refs []string
	var walk func(any)
	walk = func(current any) {
		switch typed := current.(type) {
		case map[string]any:
			if ref, ok := typed["$ref"].(string); ok {
				refs = append(refs, ref)
			}
			for _, child := range typed {
				walk(child)
			}
		case []any:
			for _, child := range typed {
				walk(child)
			}
		}
	}
	walk(value)
	return refs
}

func jsonPointerExists(document any, pointer string) bool {
	if pointer == "" {
		return true
	}
	current := document
	for _, token := range strings.Split(strings.TrimPrefix(pointer, "/"), "/") {
		token = strings.ReplaceAll(strings.ReplaceAll(token, "~1", "/"), "~0", "~")
		object, ok := current.(map[string]any)
		if !ok {
			return false
		}
		next, ok := object[token]
		if !ok {
			return false
		}
		current = next
	}
	return true
}

func schemaIDExists(t *testing.T, root string, id string) bool {
	t.Helper()
	for _, schemaPath := range collectReleaseSchemaPaths(t, root) {
		if schemaDocumentHasID(readJSON(t, schemaPath), id) {
			return true
		}
	}
	return false
}

func schemaDocumentHasID(value any, id string) bool {
	switch typed := value.(type) {
	case map[string]any:
		if currentID, ok := typed["$id"].(string); ok && currentID == id {
			return true
		}
		for _, child := range typed {
			if schemaDocumentHasID(child, id) {
				return true
			}
		}
	case []any:
		for _, child := range typed {
			if schemaDocumentHasID(child, id) {
				return true
			}
		}
	}
	return false
}

func assertBundledSchemaMatches(t *testing.T, root string, sourceRel string) {
	t.Helper()
	sourcePath := filepath.Join(root, filepath.FromSlash(sourceRel))
	bundledPath := filepath.Join(root, "cmd", "sheet-ops-codex", "skill_assets", "agent-system", filepath.FromSlash(sourceRel))
	sourceCanonical := canonicalJSON(t, readJSON(t, sourcePath))
	bundledCanonical := canonicalJSON(t, readJSON(t, bundledPath))
	if !bytes.Equal(sourceCanonical, bundledCanonical) {
		t.Fatalf("bundled schema %s drifted from %s", bundledPath, sourcePath)
	}
}

func assertSharedDefinitionEqual(t *testing.T, root string, definitionName string, schemaRels ...string) {
	t.Helper()
	var baseline []byte
	var baselineRel string
	for _, schemaRel := range schemaRels {
		document := readJSON(t, filepath.Join(root, filepath.FromSlash(schemaRel)))
		definition, ok := schemaDefinition(document, definitionName)
		if !ok {
			t.Fatalf("%s is missing $defs.%s", schemaRel, definitionName)
		}
		current := canonicalJSON(t, normalizeSchemaDefinition(definition))
		if baseline == nil {
			baseline = current
			baselineRel = schemaRel
			continue
		}
		if !bytes.Equal(baseline, current) {
			t.Fatalf("$defs.%s drifted between %s and %s", definitionName, baselineRel, schemaRel)
		}
	}
}

func schemaDefinition(document any, definitionName string) (any, bool) {
	object, ok := document.(map[string]any)
	if !ok {
		return nil, false
	}
	defs, ok := object["$defs"].(map[string]any)
	if !ok {
		return nil, false
	}
	definition, ok := defs[definitionName]
	return definition, ok
}

func normalizeSchemaDefinition(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		if normalized, ok := normalizePrimitiveAnyOf(typed); ok {
			return normalized
		}
		result := make(map[string]any, len(typed))
		for key, child := range typed {
			if key == "type" {
				if typeValues, ok := normalizeTypeArray(child); ok {
					result[key] = typeValues
					continue
				}
			}
			result[key] = normalizeSchemaDefinition(child)
		}
		return result
	case []any:
		result := make([]any, len(typed))
		for index, child := range typed {
			result[index] = normalizeSchemaDefinition(child)
		}
		return result
	default:
		return value
	}
}

func normalizePrimitiveAnyOf(value map[string]any) (map[string]any, bool) {
	anyOf, ok := value["anyOf"].([]any)
	if !ok || len(anyOf) == 0 {
		return nil, false
	}
	var types []string
	for _, branch := range anyOf {
		branchObject, ok := branch.(map[string]any)
		if !ok || len(branchObject) != 1 {
			return nil, false
		}
		primitiveType, ok := branchObject["type"].(string)
		if !ok {
			return nil, false
		}
		types = append(types, primitiveType)
	}
	sort.Strings(types)
	result := make(map[string]any, len(value))
	for key, child := range value {
		if key == "anyOf" {
			continue
		}
		result[key] = normalizeSchemaDefinition(child)
	}
	typeValues := make([]any, len(types))
	for index, primitiveType := range types {
		typeValues[index] = primitiveType
	}
	result["type"] = typeValues
	return result, true
}

func normalizeTypeArray(value any) ([]any, bool) {
	items, ok := value.([]any)
	if !ok {
		return nil, false
	}
	types := make([]string, 0, len(items))
	for _, item := range items {
		primitiveType, ok := item.(string)
		if !ok {
			return nil, false
		}
		types = append(types, primitiveType)
	}
	sort.Strings(types)
	result := make([]any, len(types))
	for index, primitiveType := range types {
		result[index] = primitiveType
	}
	return result, true
}

func canonicalJSON(t *testing.T, value any) []byte {
	t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("canonicalize JSON: %v", err)
	}
	return raw
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
	if _, ok := object["steps"]; ok {
		return "contracts/requests/organism_execution_request.schema.json"
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
