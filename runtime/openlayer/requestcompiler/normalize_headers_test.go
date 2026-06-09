package requestcompiler

import (
	"testing"

	runtimeinspect "github.com/mwroh/sheet-ops/runtime/inspect"
	runtimevalidate "github.com/mwroh/sheet-ops/runtime/validate"
)

func TestNormalizeHeadersIntentCompilesThroughValidation(t *testing.T) {
	intent := normalizeIntent(NormalizedIntent{
		SourceSheetCandidates: []string{"Movements"},
		CompositionCandidates: []string{CompositionCandidateHeaderNormalization},
		NormalizeHeaders: NormalizeHeadersIntent{
			HeaderRow: 1,
			HeaderMappings: []HeaderMapping{
				{From: "SKU ID", To: "sku"},
				{From: "Qty In", To: "quantity_in"},
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
	if decision.SelectedOperation != "normalize_headers" {
		t.Fatalf("selected_operation=%q want normalize_headers", decision.SelectedOperation)
	}
	validator := runtimevalidate.Validator{
		RequestContext: runtimevalidate.RequestContext{ScenarioID: "normalize-headers", RequestKind: "structured_use_request", OutputFile: "out.xlsx"},
	}
	result, err := validator.Validate(toRuntimeIntent(intent), toRuntimeDecision(decision), runtimeinspect.WorkbookFacts{
		InputFile: "in.xlsx",
		Sheets:    []runtimeinspect.SheetFacts{{Name: "Movements", Columns: []string{"SKU ID", "Qty In", "Qty Out"}}},
	})
	if err != nil {
		t.Fatalf("validate normalize headers intent: %v", err)
	}
	if result.ValidatedExecutionRequest == nil || result.ValidatedExecutionRequest.CompositionKind != "header_normalization" {
		t.Fatalf("validated request=%+v want header_normalization", result.ValidatedExecutionRequest)
	}
	if len(result.ValidatedExecutionRequest.HeaderMappings) != 2 {
		t.Fatalf("header mappings=%v want 2 mappings", result.ValidatedExecutionRequest.HeaderMappings)
	}
}
