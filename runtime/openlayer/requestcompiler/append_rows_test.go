package requestcompiler

import (
	"testing"

	runtimeinspect "github.com/mwroh/sheet-ops/runtime/inspect"
	runtimevalidate "github.com/mwroh/sheet-ops/runtime/validate"
)

func TestAppendRowsIntentCompilesThroughValidation(t *testing.T) {
	intent := normalizeIntent(NormalizedIntent{
		SourceSheetCandidates: []string{"LineItems"},
		CompositionCandidates: []string{CompositionCandidateStructuredRowAppend},
		AppendRows: AppendRowsIntent{
			IncludeSourceColumns: []string{"sku", "quantity", "unit_price"},
			Values: []CellValue{
				{Cell: "sku", Value: "B002"},
				{Cell: "quantity", Value: float64(2)},
				{Cell: "unit_price", Value: float64(15)},
			},
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
	if decision.SelectedOperation != "append_structured_rows" {
		t.Fatalf("selected_operation=%q want append_structured_rows", decision.SelectedOperation)
	}

	validator := runtimevalidate.Validator{
		RequestContext: runtimevalidate.RequestContext{
			ScenarioID:  "append-rows",
			RequestKind: "structured_use_request",
			OutputFile:  "out.xlsx",
		},
	}
	result, err := validator.Validate(toRuntimeIntent(intent), toRuntimeDecision(decision), runtimeinspect.WorkbookFacts{
		InputFile: "in.xlsx",
		Sheets: []runtimeinspect.SheetFacts{
			{Name: "LineItems", Columns: []string{"sku", "quantity", "unit_price"}},
		},
	})
	if err != nil {
		t.Fatalf("validate append rows intent: %v", err)
	}
	if result.ValidatedExecutionRequest == nil {
		t.Fatal("missing validated execution request")
	}
	request := result.ValidatedExecutionRequest
	if request.CompositionKind != "structured_row_append" {
		t.Fatalf("composition_kind=%q want structured_row_append", request.CompositionKind)
	}
	if len(request.Values) != 3 {
		t.Fatalf("values len=%d want 3", len(request.Values))
	}
}
