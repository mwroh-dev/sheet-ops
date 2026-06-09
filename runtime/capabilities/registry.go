package capabilities

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	runtimeschema "github.com/mwroh/sheet-ops/runtime/schema"
)

type Registry struct {
	Capabilities []Capability
}

type Capability struct {
	Name            string   `json:"name"`
	OperationFamily string   `json:"operation_family"`
	TaskOperations  []string `json:"task_operations"`
	Inputs          []string `json:"inputs"`
	Outputs         []string `json:"outputs"`
	RequiredFacts   []string `json:"required_facts"`
	Verifies        []string `json:"verifies"`
	ContractPaths   []string `json:"contract_paths"`
	RuntimePaths    []string `json:"runtime_paths"`
	Examples        []string `json:"examples"`
	Status          string   `json:"status"`
	Exposure        string   `json:"exposure"`
}

type HelpDocument struct {
	Entry        string           `json:"entry"`
	Capabilities []HelpCapability `json:"capabilities"`
}

type HelpCapability struct {
	Name            string   `json:"name"`
	OperationFamily string   `json:"operation_family"`
	TaskOperations  []string `json:"task_operations"`
	Inputs          []string `json:"inputs"`
	Outputs         []string `json:"outputs"`
	RequiredFacts   []string `json:"required_facts"`
	Verifies        []string `json:"verifies"`
	Status          string   `json:"status"`
	Exposure        string   `json:"exposure"`
}

func LoadRegistry(repoRoot string) (Registry, error) {
	root := filepath.Clean(repoRoot)
	schemaPath := filepath.Join(root, "contracts", "capabilities", "capability.schema.json")
	recordsDir := filepath.Join(root, "contracts", "capabilities", "records")

	entries, err := os.ReadDir(recordsDir)
	if err != nil {
		return Registry{}, err
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	registry := Registry{Capabilities: make([]Capability, 0, len(entries))}
	seen := map[string]struct{}{}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			return Registry{}, fmt.Errorf("unexpected capability record entry %s", entry.Name())
		}
		recordPath := filepath.Join(recordsDir, entry.Name())
		raw, err := os.ReadFile(recordPath)
		if err != nil {
			return Registry{}, err
		}

		var document any
		if err := json.Unmarshal(raw, &document); err != nil {
			return Registry{}, fmt.Errorf("parse %s: %w", recordPath, err)
		}
		if err := runtimeschema.ValidateStruct(schemaPath, document); err != nil {
			return Registry{}, fmt.Errorf("validate %s: %w", recordPath, err)
		}

		var capability Capability
		if err := json.Unmarshal(raw, &capability); err != nil {
			return Registry{}, fmt.Errorf("decode %s: %w", recordPath, err)
		}
		capability.OperationFamily = capability.Name
		if _, ok := seen[capability.OperationFamily]; ok {
			return Registry{}, fmt.Errorf("duplicate capability operation family %q", capability.OperationFamily)
		}
		seen[capability.OperationFamily] = struct{}{}

		if err := requireReferencedPaths(root, capability.ContractPaths); err != nil {
			return Registry{}, fmt.Errorf("%s contract_paths: %w", recordPath, err)
		}
		if err := requireReferencedPaths(root, capability.RuntimePaths); err != nil {
			return Registry{}, fmt.Errorf("%s runtime_paths: %w", recordPath, err)
		}
		if err := requireReferencedPaths(root, capability.Examples); err != nil {
			return Registry{}, fmt.Errorf("%s examples: %w", recordPath, err)
		}

		registry.Capabilities = append(registry.Capabilities, capability)
	}

	return registry, nil
}

func (registry Registry) ByOperationFamily(operationFamily string) (Capability, bool) {
	for _, capability := range registry.Capabilities {
		if capability.OperationFamily == operationFamily {
			return capability, true
		}
	}
	return Capability{}, false
}

func (registry Registry) HelpDocument() HelpDocument {
	capabilities := make([]HelpCapability, 0, len(registry.Capabilities))
	for _, capability := range registry.Capabilities {
		capabilities = append(capabilities, HelpCapability{
			Name:            capability.Name,
			OperationFamily: capability.OperationFamily,
			TaskOperations:  append([]string(nil), capability.TaskOperations...),
			Inputs:          append([]string(nil), capability.Inputs...),
			Outputs:         append([]string(nil), capability.Outputs...),
			RequiredFacts:   append([]string(nil), capability.RequiredFacts...),
			Verifies:        append([]string(nil), capability.Verifies...),
			Status:          capability.Status,
			Exposure:        capability.Exposure,
		})
	}
	return HelpDocument{
		Entry:        "capability_registry",
		Capabilities: capabilities,
	}
}

func requireReferencedPaths(repoRoot string, paths []string) error {
	for _, rel := range paths {
		trimmed := strings.TrimSpace(rel)
		if trimmed == "" {
			return fmt.Errorf("empty path")
		}
		if filepath.IsAbs(trimmed) {
			return fmt.Errorf("%q must be repository-relative", trimmed)
		}
		if _, err := os.Stat(filepath.Join(repoRoot, filepath.FromSlash(trimmed))); err != nil {
			return fmt.Errorf("%q: %w", trimmed, err)
		}
	}
	return nil
}
