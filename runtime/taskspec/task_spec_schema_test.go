package taskspec

import (
	"path/filepath"
	"runtime"
	"testing"

	runtimeschema "github.com/mwroh/sheet-ops/runtime/schema"
)

func TestTaskSpecSchemaRejectsForeignOperationFields(t *testing.T) {
	spec := TaskSpec{
		RequestKind:      "workbook_case",
		ExecutionKind:    "composition",
		CompositionKind:  "group_summary",
		Source:           SourceSpec{Kind: "case_markdown", Path: "case.md"},
		InputWorkbook:    "input.xlsx",
		OutputWorkbook:   "output.xlsx",
		Operation:        "create_summary_sheet",
		SourceSheet:      "Data",
		TargetSheet:      "Summary",
		GroupBy:          []string{"category"},
		Metrics:          []MetricSpec{{Column: "amount", Op: "sum", As: "total"}},
		SummaryMode:      "values",
		FormulaSourceRow: 2,
	}

	if err := runtimeschema.ValidateStruct(taskSpecSchemaPath(t), spec); err == nil {
		t.Fatal("ValidateStruct passed with formula_source_row on create_summary_sheet; want schema rejection")
	}
}

func taskSpecSchemaPath(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller ok=false")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", "contracts", "task", "task_spec.schema.json"))
}
