package schema

import (
	"path/filepath"
	"runtime"
	"testing"
)

func TestOperationSchemasRejectForeignFields(t *testing.T) {
	tests := []struct {
		name   string
		schema string
		value  map[string]any
	}{
		{
			name:   "use request",
			schema: filepath.Join("contracts", "requests", "use_request.schema.json"),
			value: map[string]any{
				"scenario_id":        "summary-foreign-field",
				"request_text":       "summarize data",
				"input_file":         "input.xlsx",
				"sheet_name":         "Data",
				"output_file":        "output.xlsx",
				"operation":          "create_summary_sheet",
				"target_sheet":       "Summary",
				"group_by":           []any{"category"},
				"metrics":            []any{map[string]any{"column": "amount", "op": "sum", "as": "total"}},
				"summary_mode":       "values",
				"formula_source_row": 2,
			},
		},
		{
			name:   "validated execution request",
			schema: filepath.Join("contracts", "requests", "validated_execution_request.schema.json"),
			value: map[string]any{
				"scenario_id":        "summary-foreign-field",
				"execution_kind":     "composition",
				"composition_kind":   "group_summary",
				"request_kind":       "structured_use_request",
				"input_file":         "input.xlsx",
				"source_sheet":       "Data",
				"output_file":        "output.xlsx",
				"target_sheet":       "Summary",
				"group_by":           []any{"category"},
				"metrics":            []any{map[string]any{"column": "amount", "op": "sum", "as": "total"}},
				"summary_mode":       "values",
				"formula_source_row": 2,
			},
		},
		{
			name:   "workbook operation IR",
			schema: filepath.Join("contracts", "ir", "workbook_operation_ir.schema.json"),
			value: map[string]any{
				"execution_kind":     "composition",
				"operation_family":   "group_summarize",
				"composition_kind":   "group_summary",
				"preserve_original":  true,
				"source_sheet":       "Data",
				"target_sheet":       "Summary",
				"group_by":           []any{"category"},
				"metrics":            []any{map[string]any{"column": "amount", "op": "sum", "as": "total"}},
				"summary_mode":       "values",
				"formula_source_row": 2,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateStruct(repoSchemaPath(t, tt.schema), tt.value); err == nil {
				t.Fatalf("ValidateStruct passed with foreign formula_source_row; want schema rejection")
			}
		})
	}
}

func repoSchemaPath(t *testing.T, rel string) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller ok=false")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", rel))
}
