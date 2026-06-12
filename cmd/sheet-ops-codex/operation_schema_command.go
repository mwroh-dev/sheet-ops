package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	runtimecapabilities "github.com/mwroh/sheet-ops/runtime/capabilities"
	"github.com/spf13/cobra"
)

const publicAgentCapabilityExposure = "public_agent_capability"

type operationListPayload struct {
	SchemaVersion string                     `json:"schema_version"`
	Command       string                     `json:"command"`
	OK            bool                       `json:"ok"`
	Operations    []operationListItemPayload `json:"operations"`
}

type operationListItemPayload struct {
	Name           string   `json:"name"`
	TaskOperations []string `json:"task_operations"`
	Exposure       string   `json:"exposure"`
	Status         string   `json:"status"`
}

type operationSchemaPayload struct {
	SchemaVersion string                   `json:"schema_version"`
	Command       string                   `json:"command"`
	OK            bool                     `json:"ok"`
	Operation     operationContractPayload `json:"operation"`
}

type operationContractPayload struct {
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

type operationExamplePayload struct {
	SchemaVersion string         `json:"schema_version"`
	Command       string         `json:"command"`
	OK            bool           `json:"ok"`
	Operation     string         `json:"operation"`
	Example       map[string]any `json:"example"`
	SourcePath    string         `json:"source_path"`
}

func newOperationCommand() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "operation",
		Short: "Inspect supported workbook operation schemas and examples",
	}
	cmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "Emit machine-readable JSON on stdout")
	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List supported public workbook operations",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !jsonOutput {
				return writeCLIErrorJSONAndReturn(
					cmd.OutOrStdout(),
					newCLIError(cliErrorInvalidUsage, "operation list requires --json", true, cliExitUsage, "operation list --json"),
				)
			}
			registry, err := loadOperationRegistry()
			if err != nil {
				return err
			}
			return writeOperationJSON(cmd.OutOrStdout(), buildOperationListPayload(registry))
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "schema <name>",
		Short: "Show one supported workbook operation contract",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !jsonOutput {
				return writeCLIErrorJSONAndReturn(
					cmd.OutOrStdout(),
					newCLIError(cliErrorInvalidUsage, "operation schema requires --json", true, cliExitUsage, "operation schema <name> --json"),
				)
			}
			registry, err := loadOperationRegistry()
			if err != nil {
				return err
			}
			capability, cliErr := publicCapabilityByName(registry, args[0])
			if cliErr != nil {
				return writeCLIErrorJSONAndReturn(cmd.OutOrStdout(), cliErr)
			}
			return writeOperationJSON(cmd.OutOrStdout(), operationSchemaPayload{
				SchemaVersion: cliContractSchemaVersion,
				Command:       "operation schema",
				OK:            true,
				Operation:     buildOperationContractPayload(capability),
			})
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "example <name>",
		Short: "Show one normalized-intent example for a workbook operation",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !jsonOutput {
				return writeCLIErrorJSONAndReturn(
					cmd.OutOrStdout(),
					newCLIError(cliErrorInvalidUsage, "operation example requires --json", true, cliExitUsage, "operation example <name> --json"),
				)
			}
			registry, err := loadOperationRegistry()
			if err != nil {
				return err
			}
			capability, cliErr := publicCapabilityByName(registry, args[0])
			if cliErr != nil {
				return writeCLIErrorJSONAndReturn(cmd.OutOrStdout(), cliErr)
			}
			example, ok := normalizedIntentExample(capability.OperationFamily)
			if !ok {
				return writeCLIErrorJSONAndReturn(
					cmd.OutOrStdout(),
					newInvalidDataError(fmt.Sprintf("no normalized intent example for operation %q", capability.OperationFamily), "operation list --json"),
				)
			}
			return writeOperationJSON(cmd.OutOrStdout(), operationExamplePayload{
				SchemaVersion: cliContractSchemaVersion,
				Command:       "operation example",
				OK:            true,
				Operation:     capability.OperationFamily,
				Example:       example,
				SourcePath:    "embedded:" + capability.OperationFamily,
			})
		},
	})
	return cmd
}

func loadOperationRegistry() (runtimecapabilities.Registry, error) {
	repoRoot, err := discoverRepoRoot()
	if err != nil {
		return runtimecapabilities.Registry{}, err
	}
	return runtimecapabilities.LoadRegistry(repoRoot)
}

func discoverRepoRoot() (string, error) {
	workingDir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for dir := filepath.Clean(workingDir); ; dir = filepath.Dir(dir) {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("could not locate repository root from %s", workingDir)
		}
	}
}

func publicCapabilityByName(registry runtimecapabilities.Registry, name string) (runtimecapabilities.Capability, *cliError) {
	trimmed := strings.TrimSpace(name)
	capability, ok := registry.ByOperationFamily(trimmed)
	if !ok || capability.Exposure != publicAgentCapabilityExposure {
		return runtimecapabilities.Capability{}, newCLIError(
			cliErrorUnknownCommand,
			fmt.Sprintf("unknown public operation %q", trimmed),
			true,
			cliExitUsage,
			"operation list --json",
		)
	}
	return capability, nil
}

func buildOperationListPayload(registry runtimecapabilities.Registry) operationListPayload {
	operations := make([]operationListItemPayload, 0, len(registry.Capabilities))
	for _, capability := range registry.Capabilities {
		if capability.Exposure != publicAgentCapabilityExposure {
			continue
		}
		operations = append(operations, operationListItemPayload{
			Name:           capability.OperationFamily,
			TaskOperations: append([]string(nil), capability.TaskOperations...),
			Exposure:       capability.Exposure,
			Status:         capability.Status,
		})
	}
	return operationListPayload{
		SchemaVersion: cliContractSchemaVersion,
		Command:       "operation list",
		OK:            true,
		Operations:    operations,
	}
}

func buildOperationContractPayload(capability runtimecapabilities.Capability) operationContractPayload {
	return operationContractPayload{
		Name:             capability.OperationFamily,
		TaskOperations:   append([]string(nil), capability.TaskOperations...),
		Inputs:           append([]string(nil), capability.Inputs...),
		Outputs:          append([]string(nil), capability.Outputs...),
		RequiredFacts:    append([]string(nil), capability.RequiredFacts...),
		Verifies:         append([]string(nil), capability.Verifies...),
		ContractPaths:    append([]string(nil), capability.ContractPaths...),
		RuntimePaths:     append([]string(nil), capability.RuntimePaths...),
		Examples:         append([]string(nil), capability.Examples...),
		SchemaReferences: append([]string(nil), capability.ContractPaths...),
		Status:           capability.Status,
		Exposure:         capability.Exposure,
	}
}

func normalizedIntentExample(operation string) (map[string]any, bool) {
	base := map[string]any{
		"source_sheet_candidates": []any{"Data"},
		"composition_candidates":  []any{compositionCandidate(operation)},
		"materialization": map[string]any{
			"preserve_original":       true,
			"output_destination_mode": "new_workbook",
			"write_shape":             writeShape(operation),
		},
		"ambiguity": map[string]any{
			"markers":           []any{},
			"unresolved_fields": []any{},
			"checkpoint_hints":  []any{},
		},
	}
	switch operation {
	case "append_structured_rows":
		base["append_rows"] = map[string]any{
			"include_source_columns": []any{"Name", "Amount"},
			"values": []any{
				map[string]any{"cell": "A2", "value": "Alice"},
				map[string]any{"cell": "B2", "value": 1200},
			},
		}
	case "group_summarize":
		base["summary"] = map[string]any{
			"target_sheet": "Summary",
			"summary_mode": "values",
			"group_by":     []any{"Region"},
			"metrics": []any{
				map[string]any{"kind": "sum", "column": "Amount"},
			},
		}
	case "highlight_threshold":
		base["highlight"] = map[string]any{
			"target_column":   "Amount",
			"operator":        ">=",
			"threshold":       1000,
			"highlight_color": "FFFF00",
		}
	case "join_lookup":
		base["lookup_sheet_candidates"] = []any{"Lookup"}
		base["join_lookup"] = map[string]any{
			"target_sheet":           "Joined",
			"join_key":               "Customer ID",
			"include_source_columns": []any{"Customer ID", "Amount"},
			"append_lookup_columns":  []any{"Segment"},
		}
	default:
		return base, true
	}
	return base, true
}

func compositionCandidate(operation string) string {
	switch operation {
	case "append_structured_rows":
		return "structured_row_append"
	case "group_summarize":
		return "group_summary"
	case "highlight_threshold":
		return "threshold_highlight"
	case "join_lookup":
		return "join_lookup"
	case "extend_table_formulas":
		return "formula_extension"
	case "copy_period_sheet":
		return "period_copy"
	case "add_data_validation":
		return "data_validation"
	case "protect_formula_cells":
		return "formula_protection"
	case "normalize_headers":
		return "header_normalization"
	case "roll_forward_period":
		return "period_roll_forward"
	case "reconcile_tables":
		return "table_reconciliation"
	case "generate_printable_form":
		return "printable_form"
	default:
		return operation
	}
}

func writeShape(operation string) string {
	switch operation {
	case "append_structured_rows", "highlight_threshold", "extend_table_formulas", "add_data_validation", "protect_formula_cells", "normalize_headers", "roll_forward_period":
		return "in_place_cells"
	default:
		return "new_sheet"
	}
}

func writeOperationJSON(output io.Writer, value any) error {
	encoder := json.NewEncoder(output)
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}
