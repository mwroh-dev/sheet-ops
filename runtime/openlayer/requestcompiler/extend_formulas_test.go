package requestcompiler

import (
	"testing"

	runtimeinspect "github.com/mwroh/sheet-ops/runtime/inspect"
	runtimevalidate "github.com/mwroh/sheet-ops/runtime/validate"
)

func TestExtendFormulasIntentCompilesThroughValidation(t *testing.T) {
	intent := normalizeIntent(NormalizedIntent{
		SourceSheetCandidates: []string{"LineItems"},
		CompositionCandidates: []string{CompositionCandidateFormulaExtension},
		ExtendFormulas: ExtendFormulasIntent{
			FormulaSourceRow: 2,
			TargetRows:       []int{3},
			FormulaColumns:   []string{"D"},
		},
		Materialization: MaterializationIntent{
			PreserveOriginal:      true,
			OutputDestinationMode: OutputDestinationModeNewWorkbook,
			WriteShape:            WriteShapeInPlaceCells,
		},
		Ambiguity: AmbiguityIntent{
			Markers:          []string{},
			UnresolvedFields: []string{},
			CheckpointHints:  []string{},
		},
	})

	decision := compileDecision(intent)
	if decision.SelectedOperation != "extend_table_formulas" {
		t.Fatalf("selected_operation=%q want extend_table_formulas", decision.SelectedOperation)
	}

	validator := runtimevalidate.Validator{
		RequestContext: runtimevalidate.RequestContext{
			ScenarioID:  "extend-formulas",
			RequestKind: "structured_use_request",
			OutputFile:  "out.xlsx",
		},
	}
	result, err := validator.Validate(toRuntimeIntent(intent), toRuntimeDecision(decision), runtimeinspect.WorkbookFacts{
		InputFile: "in.xlsx",
		Sheets: []runtimeinspect.SheetFacts{
			{Name: "LineItems", Columns: []string{"sku", "quantity", "unit_price", "line_total"}},
		},
	})
	if err != nil {
		t.Fatalf("validate formula extension intent: %v", err)
	}
	if result.ValidatedExecutionRequest == nil {
		t.Fatal("missing validated execution request")
	}
	request := result.ValidatedExecutionRequest
	if request.CompositionKind != "formula_extension" {
		t.Fatalf("composition_kind=%q want formula_extension", request.CompositionKind)
	}
	if len(request.TargetRows) != 1 || request.TargetRows[0] != 3 {
		t.Fatalf("target_rows=%v want [3]", request.TargetRows)
	}
}
