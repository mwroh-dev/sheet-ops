package requestcompiler

import (
	"testing"

	runtimeinspect "github.com/mwroh/sheet-ops/runtime/inspect"
	runtimevalidate "github.com/mwroh/sheet-ops/runtime/validate"
)

func TestRollForwardPeriodIntentCompilesThroughValidation(t *testing.T) {
	intent := normalizeIntent(NormalizedIntent{
		SourceSheetCandidates: []string{"Jan"},
		CompositionCandidates: []string{CompositionCandidatePeriodRollForward},
		RollForwardPeriod: RollForwardPeriodIntent{
			TargetSheet: "Feb",
			CarryForwardMappings: []CarryForwardMapping{
				{FromSheet: "Jan", FromCell: "D2", ToSheet: "Feb", ToCell: "A2"},
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
	if decision.SelectedOperation != "roll_forward_period" {
		t.Fatalf("selected_operation=%q want roll_forward_period", decision.SelectedOperation)
	}
	validator := runtimevalidate.Validator{
		RequestContext: runtimevalidate.RequestContext{ScenarioID: "roll-forward-period", RequestKind: "structured_use_request", OutputFile: "out.xlsx"},
	}
	result, err := validator.Validate(toRuntimeIntent(intent), toRuntimeDecision(decision), runtimeinspect.WorkbookFacts{
		InputFile: "in.xlsx",
		Sheets: []runtimeinspect.SheetFacts{
			{Name: "Jan", Columns: []string{"opening", "inflow", "outflow", "closing"}},
			{Name: "Feb", Columns: []string{"opening", "inflow", "outflow", "closing"}},
		},
	})
	if err != nil {
		t.Fatalf("validate roll forward intent: %v", err)
	}
	if result.ValidatedExecutionRequest == nil || result.ValidatedExecutionRequest.CompositionKind != "period_roll_forward" {
		t.Fatalf("validated request=%+v want period_roll_forward", result.ValidatedExecutionRequest)
	}
	if len(result.ValidatedExecutionRequest.CarryForwardMappings) != 1 {
		t.Fatalf("carry forward mappings=%v want 1 mapping", result.ValidatedExecutionRequest.CarryForwardMappings)
	}
}
