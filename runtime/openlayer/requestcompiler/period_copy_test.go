package requestcompiler

import (
	"testing"

	runtimeinspect "github.com/mwroh/sheet-ops/runtime/inspect"
	runtimevalidate "github.com/mwroh/sheet-ops/runtime/validate"
)

func TestPeriodCopyIntentCompilesThroughValidation(t *testing.T) {
	intent := normalizeIntent(NormalizedIntent{
		SourceSheetCandidates: []string{"Jan"},
		CompositionCandidates: []string{CompositionCandidatePeriodCopy},
		PeriodCopy: PeriodCopyIntent{
			TargetSheet: "Feb",
		},
		Materialization: MaterializationIntent{
			PreserveOriginal:      true,
			OutputDestinationMode: OutputDestinationModeNewWorkbook,
			WriteShape:            WriteShapeInPlaceCells,
		},
		Ambiguity: AmbiguityIntent{Markers: []string{}, UnresolvedFields: []string{}, CheckpointHints: []string{}},
	})

	decision := compileDecision(intent)
	if decision.SelectedOperation != "copy_period_sheet" {
		t.Fatalf("selected_operation=%q want copy_period_sheet", decision.SelectedOperation)
	}

	validator := runtimevalidate.Validator{
		RequestContext: runtimevalidate.RequestContext{ScenarioID: "period-copy", RequestKind: "structured_use_request", OutputFile: "out.xlsx"},
	}
	result, err := validator.Validate(toRuntimeIntent(intent), toRuntimeDecision(decision), runtimeinspect.WorkbookFacts{
		InputFile: "in.xlsx",
		Sheets:    []runtimeinspect.SheetFacts{{Name: "Jan", Columns: []string{"period", "amount", "total"}}},
	})
	if err != nil {
		t.Fatalf("validate period copy intent: %v", err)
	}
	if result.ValidatedExecutionRequest == nil || result.ValidatedExecutionRequest.CompositionKind != "period_copy" {
		t.Fatalf("validated request=%+v want period_copy", result.ValidatedExecutionRequest)
	}
	if result.ValidatedExecutionRequest.TargetSheet != "Feb" {
		t.Fatalf("target_sheet=%q want Feb", result.ValidatedExecutionRequest.TargetSheet)
	}
}
