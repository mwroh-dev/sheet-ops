package compiler

import (
	"testing"

	"github.com/mwroh/sheet-ops/runtime/taskspec"
)

func TestCompileRollForwardPeriodOperationBuildsCarryForwardIR(t *testing.T) {
	task := taskspec.BuildRollForwardPeriodTask(taskspec.RollForwardPeriodRequest{
		RequestText: "Carry Jan closing balance into Feb opening balance",
		InputFile:   "input.xlsx",
		SourceSheet: "Jan",
		TargetSheet: "Feb",
		OutputFile:  "output.xlsx",
		CarryForwardMappings: []taskspec.CarryForwardMapping{
			{FromSheet: "Jan", FromCell: "D2", ToSheet: "Feb", ToCell: "A2"},
		},
	})

	ir, err := CompileRollForwardPeriodOperation(task)
	if err != nil {
		t.Fatalf("CompileRollForwardPeriodOperation: %v", err)
	}
	if ir.ExecutionKind != taskspec.ExecutionKindComposition {
		t.Fatalf("execution_kind=%q want composition", ir.ExecutionKind)
	}
	if ir.CompositionKind != taskspec.CompositionKindPeriodRollForward {
		t.Fatalf("composition_kind=%q want period_roll_forward", ir.CompositionKind)
	}
	if ir.OperationFamily != taskspec.OperationFamilyRollForwardPeriod {
		t.Fatalf("operation_family=%q want roll_forward_period", ir.OperationFamily)
	}
	if len(ir.CarryForwardMappings) != 1 {
		t.Fatalf("carry_forward_mappings=%v want 1 mapping", ir.CarryForwardMappings)
	}
	if ir.CarryForwardMappings[0].FromCell != "D2" || ir.CarryForwardMappings[0].ToCell != "A2" {
		t.Fatalf("mapping=%+v", ir.CarryForwardMappings[0])
	}
}

func TestCompileRollForwardPeriodOperationRejectsMissingMappings(t *testing.T) {
	task := taskspec.BuildRollForwardPeriodTask(taskspec.RollForwardPeriodRequest{
		InputFile:   "input.xlsx",
		SourceSheet: "Jan",
		TargetSheet: "Feb",
		OutputFile:  "output.xlsx",
	})

	if _, err := CompileRollForwardPeriodOperation(task); err == nil {
		t.Fatalf("CompileRollForwardPeriodOperation succeeded; want missing mappings error")
	}
}
