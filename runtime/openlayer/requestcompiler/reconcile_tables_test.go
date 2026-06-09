package requestcompiler

import (
	"testing"

	runtimeinspect "github.com/mwroh/sheet-ops/runtime/inspect"
	runtimevalidate "github.com/mwroh/sheet-ops/runtime/validate"
)

func TestReconcileTablesIntentCompilesThroughValidation(t *testing.T) {
	intent := normalizeIntent(NormalizedIntent{
		SourceSheetCandidates: []string{"Movements"},
		LookupSheetCandidates: []string{"StockMaster"},
		CompositionCandidates: []string{CompositionCandidateTableReconciliation},
		ReconcileTables: ReconcileTablesIntent{
			TargetSheet: "Reconciliation",
			LeftKey:     "sku",
			RightKey:    "sku",
			CompareMappings: []CompareMapping{
				{LeftColumn: "balance", RightColumn: "on_hand", As: "balance"},
			},
		},
		Materialization: MaterializationIntent{
			PreserveOriginal:      true,
			OutputDestinationMode: OutputDestinationModeNewWorkbook,
			WriteShape:            WriteShapeNewSheet,
		},
		Ambiguity: AmbiguityIntent{Markers: []string{}, UnresolvedFields: []string{}, CheckpointHints: []string{}},
	})
	decision := compileDecision(intent)
	if decision.SelectedOperation != "reconcile_tables" {
		t.Fatalf("selected_operation=%q want reconcile_tables", decision.SelectedOperation)
	}
	validator := runtimevalidate.Validator{
		RequestContext: runtimevalidate.RequestContext{ScenarioID: "reconcile-tables", RequestKind: "structured_use_request", OutputFile: "out.xlsx"},
	}
	result, err := validator.Validate(toRuntimeIntent(intent), toRuntimeDecision(decision), runtimeinspect.WorkbookFacts{
		InputFile: "in.xlsx",
		Sheets: []runtimeinspect.SheetFacts{
			{Name: "Movements", Columns: []string{"sku", "balance"}},
			{Name: "StockMaster", Columns: []string{"sku", "on_hand"}},
		},
	})
	if err != nil {
		t.Fatalf("validate reconciliation intent: %v", err)
	}
	if result.ValidatedExecutionRequest == nil || result.ValidatedExecutionRequest.CompositionKind != "table_reconciliation" {
		t.Fatalf("validated request=%+v want table_reconciliation", result.ValidatedExecutionRequest)
	}
	if len(result.ValidatedExecutionRequest.CompareMappings) != 1 {
		t.Fatalf("compare mappings=%v want 1 mapping", result.ValidatedExecutionRequest.CompareMappings)
	}
}
