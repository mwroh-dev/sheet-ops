package requestcompiler

import (
	"testing"

	runtimeinspect "github.com/mwroh/sheet-ops/runtime/inspect"
	runtimevalidate "github.com/mwroh/sheet-ops/runtime/validate"
)

func TestFormulaProtectionIntentCompilesThroughValidation(t *testing.T) {
	intent := normalizeIntent(NormalizedIntent{
		SourceSheetCandidates: []string{"LineItems"},
		CompositionCandidates: []string{CompositionCandidateFormulaProtection},
		ProtectFormulaCells: ProtectFormulaCellsIntent{
			ProtectionRule: FormulaProtectionRule{
				FormulaRanges: []string{"D2"},
				InputRanges:   []string{"A2:C10"},
			},
		},
		Materialization: MaterializationIntent{
			PreserveOriginal:      true,
			OutputDestinationMode: OutputDestinationModeNewWorkbook,
			WriteShape:            WriteShapeInPlaceCells,
		},
		Ambiguity: AmbiguityIntent{Markers: []string{}, UnresolvedFields: []string{}, CheckpointHints: []string{}},
	})

	decision := compileDecision(intent)
	if decision.SelectedOperation != "protect_formula_cells" {
		t.Fatalf("selected_operation=%q want protect_formula_cells", decision.SelectedOperation)
	}

	validator := runtimevalidate.Validator{
		RequestContext: runtimevalidate.RequestContext{ScenarioID: "formula-protection", RequestKind: "structured_use_request", OutputFile: "out.xlsx"},
	}
	result, err := validator.Validate(toRuntimeIntent(intent), toRuntimeDecision(decision), runtimeinspect.WorkbookFacts{
		InputFile: "in.xlsx",
		Sheets:    []runtimeinspect.SheetFacts{{Name: "LineItems", Columns: []string{"sku", "quantity", "unit_price", "line_total"}}},
	})
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if result.ValidatedExecutionRequest == nil || result.ValidatedExecutionRequest.CompositionKind != "formula_protection" {
		t.Fatalf("validated request=%+v want formula_protection", result.ValidatedExecutionRequest)
	}
}
