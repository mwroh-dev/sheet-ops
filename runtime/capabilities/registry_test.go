package capabilities

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestLoadRegistryReadsCurrentSupportedCapabilities(t *testing.T) {
	registry, err := LoadRegistry(repoRoot(t))
	if err != nil {
		t.Fatalf("LoadRegistry: %v", err)
	}

	recordPaths, err := filepath.Glob(filepath.Join(repoRoot(t), "contracts", "capabilities", "records", "*.yaml"))
	if err != nil {
		t.Fatalf("Glob(capability records): %v", err)
	}
	if len(registry.Capabilities) != len(recordPaths) {
		t.Fatalf("capability count=%d want %d", len(registry.Capabilities), len(recordPaths))
	}

	for _, family := range []string{"group_summarize", "highlight_threshold", "join_lookup", "append_structured_rows", "write_values"} {
		capability, ok := registry.ByOperationFamily(family)
		if !ok {
			t.Fatalf("ByOperationFamily(%q) missing", family)
		}
		if capability.Status != "supported" {
			t.Fatalf("%s status=%q want supported", family, capability.Status)
		}
		wantExposure := "public_agent_capability"
		if family == "write_values" {
			wantExposure = "runtime_primitive"
		}
		if capability.Exposure != wantExposure {
			t.Fatalf("%s exposure=%q want %s", family, capability.Exposure, wantExposure)
		}
		if len(capability.ContractPaths) == 0 || len(capability.RuntimePaths) == 0 || len(capability.Examples) == 0 {
			t.Fatalf("%s did not load authority paths: %+v", family, capability)
		}
	}
}

func TestLoadRegistryRejectsMalformedRecords(t *testing.T) {
	root := t.TempDir()
	recordsDir := filepath.Join(root, "contracts", "capabilities", "records")
	if err := os.MkdirAll(recordsDir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(filepath.Join(recordsDir, "bad.yaml"), []byte(`{"name":"group_summarize"}`), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	if _, err := LoadRegistry(root); err == nil {
		t.Fatal("LoadRegistry unexpectedly accepted malformed record")
	}
}

func TestLoadRegistryRejectsMissingReferencedPaths(t *testing.T) {
	root := t.TempDir()
	recordsDir := filepath.Join(root, "contracts", "capabilities", "records")
	if err := os.MkdirAll(recordsDir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	record := `{
  "name": "group_summarize",
  "operation_family": "group_summarize",
  "task_operations": ["create_summary_sheet"],
  "inputs": ["input_workbook"],
  "outputs": ["output_workbook"],
  "required_facts": ["sheets"],
  "verifies": ["source_preserved"],
  "contract_paths": ["contracts/task/task_spec.schema.json"],
  "runtime_paths": ["runtime/compiler/workbook_compiler.go"],
  "examples": ["examples/group_summarize/values_request.json"],
  "status": "supported",
  "exposure": "public_agent_capability"
}`
	if err := os.WriteFile(filepath.Join(recordsDir, "group_summarize.yaml"), []byte(record), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	if _, err := LoadRegistry(root); err == nil {
		t.Fatal("LoadRegistry unexpectedly accepted missing referenced paths")
	}
}

func TestRegistryHelpDocumentIsStableAndMachineReadable(t *testing.T) {
	registry, err := LoadRegistry(repoRoot(t))
	if err != nil {
		t.Fatalf("LoadRegistry: %v", err)
	}

	help := registry.HelpDocument()
	if help.Entry != "capability_registry" {
		t.Fatalf("help entry=%q want capability_registry", help.Entry)
	}
	recordPaths, err := filepath.Glob(filepath.Join(repoRoot(t), "contracts", "capabilities", "records", "*.yaml"))
	if err != nil {
		t.Fatalf("Glob(capability records): %v", err)
	}
	if len(help.Capabilities) != len(recordPaths) {
		t.Fatalf("help capability count=%d want %d", len(help.Capabilities), len(recordPaths))
	}
	for _, capability := range help.Capabilities {
		if capability.Name == "" || capability.OperationFamily == "" || capability.Status != "supported" || capability.Exposure == "" {
			t.Fatalf("malformed help capability: %+v", capability)
		}
		if len(capability.Inputs) == 0 || len(capability.Outputs) == 0 || len(capability.Verifies) == 0 {
			t.Fatalf("help capability missing input/output/verification summary: %+v", capability)
		}
	}
}

func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
