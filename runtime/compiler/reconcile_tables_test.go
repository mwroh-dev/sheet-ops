package compiler

import (
	"testing"

	"github.com/mwroh/sheet-ops/runtime/taskspec"
)

func TestCompileReconcileTablesOperationBuildsTableReconciliationIR(t *testing.T) {
	task := taskspec.BuildReconcileTablesTask(taskspec.ReconcileTablesRequest{
		RequestText: "Reconcile movement balances with stock master",
		InputFile:   "inventory.xlsx",
		SourceSheet: "Movements",
		LookupSheet: "StockMaster",
		TargetSheet: "Reconciliation",
		OutputFile:  "inventory-output.xlsx",
		LeftKey:     "sku",
		RightKey:    "sku",
		CompareMappings: []taskspec.CompareMapping{
			{LeftColumn: "balance", RightColumn: "on_hand", As: "balance"},
		},
	})

	ir, err := CompileReconcileTablesOperation(task)
	if err != nil {
		t.Fatalf("CompileReconcileTablesOperation: %v", err)
	}
	if ir.ExecutionKind != taskspec.ExecutionKindComposition {
		t.Fatalf("execution_kind=%q want composition", ir.ExecutionKind)
	}
	if ir.CompositionKind != taskspec.CompositionKindTableReconciliation {
		t.Fatalf("composition_kind=%q want table_reconciliation", ir.CompositionKind)
	}
	if ir.OperationFamily != taskspec.OperationFamilyReconcileTables {
		t.Fatalf("operation_family=%q want reconcile_tables", ir.OperationFamily)
	}
	if ir.LeftKey != "sku" || ir.RightKey != "sku" {
		t.Fatalf("keys left=%q right=%q want sku/sku", ir.LeftKey, ir.RightKey)
	}
	if len(ir.CompareMappings) != 1 {
		t.Fatalf("compare_mappings=%v want 1 mapping", ir.CompareMappings)
	}
	if ir.CompareMappings[0].LeftColumn != "balance" || ir.CompareMappings[0].RightColumn != "on_hand" {
		t.Fatalf("mapping=%+v", ir.CompareMappings[0])
	}
	if !ir.PreserveOriginal {
		t.Fatalf("preserve_original=false want true")
	}
}

func TestCompileReconcileTablesOperationRejectsMissingMappings(t *testing.T) {
	task := taskspec.BuildReconcileTablesTask(taskspec.ReconcileTablesRequest{
		InputFile:   "inventory.xlsx",
		SourceSheet: "Movements",
		LookupSheet: "StockMaster",
		TargetSheet: "Reconciliation",
		OutputFile:  "inventory-output.xlsx",
		LeftKey:     "sku",
		RightKey:    "sku",
	})

	if _, err := CompileReconcileTablesOperation(task); err == nil {
		t.Fatalf("CompileReconcileTablesOperation succeeded; want missing mappings error")
	}
}
