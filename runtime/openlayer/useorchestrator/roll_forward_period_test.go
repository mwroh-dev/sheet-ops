package useorchestrator

import (
	"testing"

	runtimeworkbookcase "github.com/mwroh/sheet-ops/runtime/workbookcase"
)

func TestRollForwardPeriodUseRequestBuildsPeriodRollForwardTaskSpec(t *testing.T) {
	req := UseRequest{
		ScenarioID:  "roll-forward-period",
		RequestText: "Jan closing balance를 Feb opening balance로 이월한다.",
		InputFile:   "in.xlsx",
		SheetName:   "Jan",
		TargetSheet: "Feb",
		OutputFile:  "out.xlsx",
		Operation:   runtimeworkbookcase.RollForwardPeriodOperationName,
		CarryForwardMappings: []CarryForwardMapping{
			{FromSheet: "Jan", FromCell: "D2", ToSheet: "Feb", ToCell: "A2"},
		},
	}

	spec := taskSpecFromUseRequest(req)
	if spec.Operation != runtimeworkbookcase.RollForwardPeriodOperationName {
		t.Fatalf("operation=%q want roll_forward_period", spec.Operation)
	}
	if spec.CompositionKind != "period_roll_forward" {
		t.Fatalf("composition_kind=%q want period_roll_forward", spec.CompositionKind)
	}
	if len(spec.CarryForwardMappings) != 1 {
		t.Fatalf("carry_forward_mappings=%v want 1 mapping", spec.CarryForwardMappings)
	}
}

func TestRollForwardPeriodValidatedRequestBuildsPeriodRollForwardTaskSpec(t *testing.T) {
	req := ValidatedExecutionRequest{
		ScenarioID:      "roll-forward-period",
		RequestKind:     "structured_use_request",
		InputFile:       "in.xlsx",
		SourceSheet:     "Jan",
		TargetSheet:     "Feb",
		OutputFile:      "out.xlsx",
		ExecutionKind:   "composition",
		CompositionKind: "period_roll_forward",
		CarryForwardMappings: []CarryForwardMapping{
			{FromSheet: "Jan", FromCell: "D2", ToSheet: "Feb", ToCell: "A2"},
		},
	}

	spec, err := taskSpecFromValidatedExecutionRequest(req)
	if err != nil {
		t.Fatalf("taskSpecFromValidatedExecutionRequest: %v", err)
	}
	if spec.Operation != runtimeworkbookcase.RollForwardPeriodOperationName {
		t.Fatalf("operation=%q want roll_forward_period", spec.Operation)
	}
	if spec.CompositionKind != "period_roll_forward" {
		t.Fatalf("composition_kind=%q want period_roll_forward", spec.CompositionKind)
	}
	if len(spec.CarryForwardMappings) != 1 {
		t.Fatalf("carry_forward_mappings=%v want 1 mapping", spec.CarryForwardMappings)
	}
}
