package useorchestrator

import (
	"testing"

	runtimeworkbookcase "github.com/mwroh/sheet-ops/runtime/workbookcase"
)

func TestReconcileTablesUseRequestBuildsTableReconciliationTaskSpec(t *testing.T) {
	req := UseRequest{
		ScenarioID:  "reconcile-tables",
		RequestText: "Movements balance와 StockMaster on_hand를 sku 기준으로 대조한다.",
		InputFile:   "in.xlsx",
		SheetName:   "Movements",
		LookupSheet: "StockMaster",
		TargetSheet: "Reconciliation",
		OutputFile:  "out.xlsx",
		Operation:   runtimeworkbookcase.ReconcileTablesOperationName,
		LeftKey:     "sku",
		RightKey:    "sku",
		CompareMappings: []CompareMapping{
			{LeftColumn: "balance", RightColumn: "on_hand", As: "balance"},
		},
	}

	spec := taskSpecFromUseRequest(req)
	if spec.Operation != runtimeworkbookcase.ReconcileTablesOperationName {
		t.Fatalf("operation=%q want reconcile_tables", spec.Operation)
	}
	if spec.CompositionKind != "table_reconciliation" {
		t.Fatalf("composition_kind=%q want table_reconciliation", spec.CompositionKind)
	}
	if len(spec.CompareMappings) != 1 {
		t.Fatalf("compare_mappings=%v want 1 mapping", spec.CompareMappings)
	}
}

func TestReconcileTablesValidatedRequestBuildsTableReconciliationTaskSpec(t *testing.T) {
	req := ValidatedExecutionRequest{
		ScenarioID:      "reconcile-tables",
		RequestKind:     "structured_use_request",
		InputFile:       "in.xlsx",
		SourceSheet:     "Movements",
		LookupSheet:     "StockMaster",
		TargetSheet:     "Reconciliation",
		OutputFile:      "out.xlsx",
		ExecutionKind:   "composition",
		CompositionKind: "table_reconciliation",
		LeftKey:         "sku",
		RightKey:        "sku",
		CompareMappings: []CompareMapping{
			{LeftColumn: "balance", RightColumn: "on_hand", As: "balance"},
		},
	}

	spec, err := taskSpecFromValidatedExecutionRequest(req)
	if err != nil {
		t.Fatalf("taskSpecFromValidatedExecutionRequest: %v", err)
	}
	if spec.Operation != runtimeworkbookcase.ReconcileTablesOperationName {
		t.Fatalf("operation=%q want reconcile_tables", spec.Operation)
	}
	if spec.CompositionKind != "table_reconciliation" {
		t.Fatalf("composition_kind=%q want table_reconciliation", spec.CompositionKind)
	}
	if len(spec.CompareMappings) != 1 {
		t.Fatalf("compare_mappings=%v want 1 mapping", spec.CompareMappings)
	}
}
