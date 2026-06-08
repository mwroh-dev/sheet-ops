package requestcompiler

import (
	"testing"

	runtimeinspect "github.com/mwroh/sheet-ops/runtime/inspect"
	runtimevalidate "github.com/mwroh/sheet-ops/runtime/validate"
)

func TestDataValidationIntentCompilesThroughValidation(t *testing.T) {
	intent := normalizeIntent(NormalizedIntent{
		SourceSheetCandidates: []string{"LineItems"},
		CompositionCandidates: []string{CompositionCandidateDataValidation},
		AddDataValidation: AddDataValidationIntent{
			ValidationRule: DataValidationRule{
				Ranges:        []string{"D2:D10"},
				RuleType:      "list",
				AllowedValues: []string{"draft", "sent", "paid"},
				AllowBlank:    false,
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
	if decision.SelectedOperation != "add_data_validation" {
		t.Fatalf("selected_operation=%q want add_data_validation", decision.SelectedOperation)
	}
	validator := runtimevalidate.Validator{
		RequestContext: runtimevalidate.RequestContext{ScenarioID: "data-validation", RequestKind: "structured_use_request", OutputFile: "out.xlsx"},
	}
	result, err := validator.Validate(toRuntimeIntent(intent), toRuntimeDecision(decision), runtimeinspect.WorkbookFacts{
		InputFile: "in.xlsx",
		Sheets:    []runtimeinspect.SheetFacts{{Name: "LineItems", Columns: []string{"sku", "status"}}},
	})
	if err != nil {
		t.Fatalf("validate data validation intent: %v", err)
	}
	if result.ValidatedExecutionRequest == nil || result.ValidatedExecutionRequest.CompositionKind != "data_validation" {
		t.Fatalf("validated request=%+v want data_validation", result.ValidatedExecutionRequest)
	}
}
