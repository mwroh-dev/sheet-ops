package workbookcase

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	runtimetaskspec "github.com/mwroh/sheet-ops/runtime/taskspec"
	"github.com/mwroh/sheet-ops/runtime/templateclass"
	"github.com/xuri/excelize/v2"
)

func TestRunOrganismPlanExecutesInvoiceSequenceAndEvaluatesTemplateClassEvidence(t *testing.T) {
	setRuntimeRoots(t)

	tempDir := t.TempDir()
	inputFile := filepath.Join(tempDir, "invoice.xlsx")
	outputFile := filepath.Join(tempDir, "invoice-organism.xlsx")

	file := excelize.NewFile()
	defer func() { _ = file.Close() }()
	defaultSheet := file.GetSheetName(0)
	if err := file.SetSheetName(defaultSheet, "LineItems"); err != nil {
		t.Fatalf("SetSheetName: %v", err)
	}
	header := []any{"sku", "quantity", "unit_price", "line_total", "tax"}
	if err := file.SetSheetRow("LineItems", "A1", &header); err != nil {
		t.Fatalf("SetSheetRow(header): %v", err)
	}
	row := []any{"A001", 2, 10}
	if err := file.SetSheetRow("LineItems", "A2", &row); err != nil {
		t.Fatalf("SetSheetRow(row): %v", err)
	}
	if err := file.SetCellFormula("LineItems", "D2", "=B2*C2"); err != nil {
		t.Fatalf("SetCellFormula(D2): %v", err)
	}
	if err := file.SetCellFormula("LineItems", "E2", "=D2*0.1"); err != nil {
		t.Fatalf("SetCellFormula(E2): %v", err)
	}
	if err := file.SaveAs(inputFile); err != nil {
		t.Fatalf("SaveAs: %v", err)
	}

	result, err := RunOrganismPlan(OrganismRunRequest{
		ScenarioID:  "invoice-organism-run",
		RequestText: "Add invoice line items, extend totals, protect formulas, and create a printable invoice",
		InputFile:   inputFile,
		OutputFile:  outputFile,
		Steps: []OrganismStep{
			{
				AtomID: "append_structured_rows",
				Build: func(input, output string) runtimetaskspec.TaskSpec {
					return runtimetaskspec.BuildAppendStructuredRowsTask(runtimetaskspec.AppendStructuredRowsRequest{
						RequestText:          "청구서에 새 line item 행을 추가한다.",
						InputFile:            input,
						SourceSheet:          "LineItems",
						OutputFile:           output,
						IncludeSourceColumns: []string{"sku", "quantity", "unit_price"},
						Values: []runtimetaskspec.CellValue{
							{Cell: "sku", Value: "B002"},
							{Cell: "quantity", Value: 3},
							{Cell: "unit_price", Value: 15},
						},
					}).TaskSpec
				},
			},
			{
				AtomID: "extend_table_formulas",
				Build: func(input, output string) runtimetaskspec.TaskSpec {
					return runtimetaskspec.BuildExtendTableFormulasTask(runtimetaskspec.ExtendTableFormulasRequest{
						RequestText:      "새 line item 행에 계산 수식을 확장한다.",
						InputFile:        input,
						SourceSheet:      "LineItems",
						OutputFile:       output,
						FormulaSourceRow: 2,
						TargetRows:       []int{3},
						FormulaColumns:   []string{"D", "E"},
					}).TaskSpec
				},
			},
			{
				AtomID: "add_data_validation",
				Build: func(input, output string) runtimetaskspec.TaskSpec {
					return runtimetaskspec.BuildAddDataValidationTask(runtimetaskspec.AddDataValidationRequest{
						RequestText: "청구서 line item SKU 입력 범위를 허용된 SKU dropdown으로 제한한다.",
						InputFile:   input,
						SourceSheet: "LineItems",
						OutputFile:  output,
						ValidationRule: runtimetaskspec.DataValidationRule{
							Ranges:        []string{"A2:A10"},
							RuleType:      "list",
							AllowedValues: []string{"A001", "B002"},
							AllowBlank:    false,
						},
					}).TaskSpec
				},
			},
			{
				AtomID: "protect_formula_cells",
				Build: func(input, output string) runtimetaskspec.TaskSpec {
					return runtimetaskspec.BuildProtectFormulaCellsTask(runtimetaskspec.ProtectFormulaCellsRequest{
						RequestText: "청구서 계산 수식은 보호하고 line item 입력 범위는 편집 가능하게 둔다.",
						InputFile:   input,
						SourceSheet: "LineItems",
						OutputFile:  output,
						ProtectionRule: runtimetaskspec.FormulaProtectionRule{
							FormulaRanges: []string{"D2:E3"},
							InputRanges:   []string{"A2:C10"},
						},
					}).TaskSpec
				},
			},
			{
				AtomID: "generate_printable_form",
				Build: func(input, output string) runtimetaskspec.TaskSpec {
					return runtimetaskspec.BuildGeneratePrintableFormTask(runtimetaskspec.GeneratePrintableFormRequest{
						RequestText: "청구서 line item 데이터를 printable invoice form으로 생성한다.",
						InputFile:   input,
						SourceSheet: "LineItems",
						TargetSheet: "InvoicePrint",
						OutputFile:  output,
						FormTitle:   "Invoice",
						PrintArea:   "A1:E8",
						FieldBindings: []runtimetaskspec.FormFieldBinding{
							{Label: "First SKU", SourceSheet: "LineItems", SourceCell: "A2", LabelCell: "A2", ValueCell: "B2"},
						},
						TableBinding: &runtimetaskspec.FormTableBinding{
							SourceSheet:   "LineItems",
							SourceColumns: []string{"sku", "quantity", "unit_price", "line_total", "tax"},
							HeaderStart:   "A4",
							DataStart:     "A5",
						},
					}).TaskSpec
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("RunOrganismPlan: %v", err)
	}
	if !result.TemplateClassEvaluation.Pass {
		t.Fatalf("template class evaluation failed: %+v", result.TemplateClassEvaluation)
	}
	if result.TemplateClassEvaluation.RuntimeClaim != "template_class_plan_verified" {
		t.Fatalf("runtime claim=%q want template_class_plan_verified", result.TemplateClassEvaluation.RuntimeClaim)
	}
	if len(result.StepResults) != 5 {
		t.Fatalf("step results=%d want 5", len(result.StepResults))
	}
	if !result.OrganismVerification.Pass {
		t.Fatalf("organism verification failed: %+v", result.OrganismVerification)
	}
	if result.OrganismVerification.VerifierSpecID != "invoice_line_item_billing_verifier" {
		t.Fatalf("organism verifier spec=%q want invoice_line_item_billing_verifier", result.OrganismVerification.VerifierSpecID)
	}
	if len(result.OrganismVerification.ExecutedAtomIDs) != 5 {
		t.Fatalf("organism executed atoms=%v want 5 atoms", result.OrganismVerification.ExecutedAtomIDs)
	}
	if _, err := os.Stat(result.OrganismVerificationPath); err != nil {
		t.Fatalf("organism verification artifact missing: %v", err)
	}

	outputHandle, err := excelize.OpenFile(outputFile)
	if err != nil {
		t.Fatalf("Open output: %v", err)
	}
	defer func() { _ = outputHandle.Close() }()
	if got, err := outputHandle.GetCellValue("LineItems", "A3"); err != nil || got != "B002" {
		t.Fatalf("LineItems!A3=%q err=%v want B002", got, err)
	}
	if got, err := outputHandle.GetCellFormula("LineItems", "D3"); err != nil || got != "=B3*C3" {
		t.Fatalf("LineItems!D3 formula=%q err=%v want =B3*C3", got, err)
	}
	if got, err := outputHandle.GetCellValue("InvoicePrint", "A1"); err != nil || got != "Invoice" {
		t.Fatalf("InvoicePrint!A1=%q err=%v want Invoice", got, err)
	}
}

func TestRunOrganismPlanRejectsStepsThatDoNotMatchTemplateClassPlan(t *testing.T) {
	setRuntimeRoots(t)

	tempDir := t.TempDir()
	inputFile := filepath.Join(tempDir, "invoice.xlsx")
	outputFile := filepath.Join(tempDir, "invoice-organism.xlsx")
	file := excelize.NewFile()
	defer func() { _ = file.Close() }()
	if err := file.SaveAs(inputFile); err != nil {
		t.Fatalf("SaveAs: %v", err)
	}

	_, err := RunOrganismPlan(OrganismRunRequest{
		ScenarioID:  "invoice-organism-run-mismatch",
		RequestText: "Add invoice line items, extend totals, protect formulas, and create a printable invoice",
		InputFile:   inputFile,
		OutputFile:  outputFile,
		Steps: []OrganismStep{
			{
				AtomID: "generate_printable_form",
				Build: func(input, output string) runtimetaskspec.TaskSpec {
					return runtimetaskspec.TaskSpec{InputWorkbook: input, OutputWorkbook: output}
				},
			},
		},
	})
	if err == nil {
		t.Fatal("RunOrganismPlan err=nil want step mismatch error")
	}
	if !strings.Contains(err.Error(), "does not match template class plan") {
		t.Fatalf("err=%q want template class plan mismatch", err)
	}
}

func TestVerifyOrganismPlanExecutionFailsWhenAtomEvidenceIsMissing(t *testing.T) {
	verification := verifyOrganismPlanExecution(templateclass.Plan{
		OrganismID:            "invoice_line_item_billing",
		OperationSequence:     []string{"append_structured_rows", "extend_table_formulas"},
		RequiredVerifierSpecs: []string{"invoice_line_item_billing_verifier"},
	}, []string{"append_structured_rows"}, []RunResult{
		{Verification: VerificationResult{Pass: true}},
	}, "invoice.xlsx")

	if verification.Pass {
		t.Fatalf("organism verification pass=true want false")
	}
	if !strings.Contains(strings.Join(verification.Reasons, " "), "missing executed atom") {
		t.Fatalf("organism verification reasons=%v want missing executed atom", verification.Reasons)
	}
}

func TestRunOrganismPlanExecutesLoanSequenceWithNonClaims(t *testing.T) {
	setRuntimeRoots(t)

	tempDir := t.TempDir()
	inputFile := filepath.Join(tempDir, "loan.xlsx")
	outputFile := filepath.Join(tempDir, "loan-organism.xlsx")

	file := excelize.NewFile()
	defer func() { _ = file.Close() }()
	defaultSheet := file.GetSheetName(0)
	if err := file.SetSheetName(defaultSheet, "Schedule"); err != nil {
		t.Fatalf("SetSheetName: %v", err)
	}
	if err := file.SetSheetRow("Schedule", "A1", &[]any{"period", "payment", "interest", "principal", "balance"}); err != nil {
		t.Fatalf("SetSheetRow header: %v", err)
	}
	if err := file.SetSheetRow("Schedule", "A2", &[]any{1, 100, nil, nil, 1000}); err != nil {
		t.Fatalf("SetSheetRow row2: %v", err)
	}
	if err := file.SetCellFormula("Schedule", "C2", "=E2*0.01"); err != nil {
		t.Fatalf("SetCellFormula(C2): %v", err)
	}
	if err := file.SetCellFormula("Schedule", "D2", "=B2-C2"); err != nil {
		t.Fatalf("SetCellFormula(D2): %v", err)
	}
	if err := file.SetCellFormula("Schedule", "E2", "=1000-D2"); err != nil {
		t.Fatalf("SetCellFormula(E2): %v", err)
	}
	if err := file.SetSheetRow("Schedule", "A3", &[]any{2, 100}); err != nil {
		t.Fatalf("SetSheetRow row3: %v", err)
	}
	if err := file.SaveAs(inputFile); err != nil {
		t.Fatalf("SaveAs: %v", err)
	}

	result, err := RunOrganismPlan(OrganismRunRequest{
		ScenarioID:  "loan-organism-run",
		RequestText: "Extend loan repayment schedule formulas and protect calculated balance cells",
		InputFile:   inputFile,
		OutputFile:  outputFile,
		Steps: []OrganismStep{
			{
				AtomID: "extend_table_formulas",
				Build: func(input, output string) runtimetaskspec.TaskSpec {
					return runtimetaskspec.BuildExtendTableFormulasTask(runtimetaskspec.ExtendTableFormulasRequest{
						RequestText:      "loan repayment schedule 계산 공식을 다음 payment row로 확장한다.",
						InputFile:        input,
						SourceSheet:      "Schedule",
						OutputFile:       output,
						FormulaSourceRow: 2,
						TargetRows:       []int{3},
						FormulaColumns:   []string{"C", "D", "E"},
					}).TaskSpec
				},
			},
			{
				AtomID: "protect_formula_cells",
				Build: func(input, output string) runtimetaskspec.TaskSpec {
					return runtimetaskspec.BuildProtectFormulaCellsTask(runtimetaskspec.ProtectFormulaCellsRequest{
						RequestText: "loan repayment schedule 계산 cells를 보호하고 payment 입력은 열어둔다.",
						InputFile:   input,
						SourceSheet: "Schedule",
						OutputFile:  output,
						ProtectionRule: runtimetaskspec.FormulaProtectionRule{
							FormulaRanges: []string{"C2:E3"},
							InputRanges:   []string{"B2:B20"},
						},
					}).TaskSpec
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("RunOrganismPlan: %v", err)
	}
	if !result.TemplateClassEvaluation.Pass {
		t.Fatalf("template class evaluation failed: %+v", result.TemplateClassEvaluation)
	}
	if result.TemplateClassEvaluation.RuntimeClaim != "template_class_plan_verified_with_non_claims" {
		t.Fatalf("runtime claim=%q want template_class_plan_verified_with_non_claims", result.TemplateClassEvaluation.RuntimeClaim)
	}
	if len(result.TemplateClassEvaluation.NonClaims) == 0 {
		t.Fatalf("expected loan non-claims")
	}

	outputHandle, err := excelize.OpenFile(outputFile)
	if err != nil {
		t.Fatalf("Open output: %v", err)
	}
	defer func() { _ = outputHandle.Close() }()
	if got, err := outputHandle.GetCellFormula("Schedule", "C3"); err != nil || got != "=E3*0.01" {
		t.Fatalf("Schedule!C3 formula=%q err=%v want =E3*0.01", got, err)
	}
	if got, err := outputHandle.GetCellFormula("Schedule", "D3"); err != nil || got != "=B3-C3" {
		t.Fatalf("Schedule!D3 formula=%q err=%v want =B3-C3", got, err)
	}
}

func TestRunOrganismPlanExecutesGradebookSequenceWithVerifiedClassNonClaims(t *testing.T) {
	setRuntimeRoots(t)

	tempDir := t.TempDir()
	inputFile := filepath.Join(tempDir, "gradebook.xlsx")
	outputFile := filepath.Join(tempDir, "gradebook-organism.xlsx")

	file := excelize.NewFile()
	defer func() { _ = file.Close() }()
	defaultSheet := file.GetSheetName(0)
	if err := file.SetSheetName(defaultSheet, "Grades"); err != nil {
		t.Fatalf("SetSheetName: %v", err)
	}
	if err := file.SetSheetRow("Grades", "A1", &[]any{"student", "assignment", "score", "status", "weighted_score"}); err != nil {
		t.Fatalf("SetSheetRow header: %v", err)
	}
	for idx, row := range [][]any{{"Ada", "Quiz 1", 90, "complete"}, {"Ben", "Quiz 1", 70, "missing"}, {"Ada", "Quiz 2", 80, "complete"}} {
		cell, _ := excelize.CoordinatesToCellName(1, idx+2)
		if err := file.SetSheetRow("Grades", cell, &row); err != nil {
			t.Fatalf("SetSheetRow grade %d: %v", idx, err)
		}
	}
	if err := file.SetCellFormula("Grades", "E2", "=C2"); err != nil {
		t.Fatalf("SetCellFormula(E2): %v", err)
	}
	if err := file.SetCellFormula("Grades", "E3", "=C3"); err != nil {
		t.Fatalf("SetCellFormula(E3): %v", err)
	}
	if err := file.SaveAs(inputFile); err != nil {
		t.Fatalf("SaveAs: %v", err)
	}

	result, err := RunOrganismPlan(OrganismRunRequest{
		ScenarioID:  "gradebook-organism-run",
		RequestText: "Summarize student scores in a gradebook, validate completion status, extend formulas, and protect calculated cells",
		InputFile:   inputFile,
		OutputFile:  outputFile,
		Steps: []OrganismStep{
			{
				AtomID: "group_summarize",
				Build: func(input, output string) runtimetaskspec.TaskSpec {
					return runtimetaskspec.BuildGroupSummarizeTask(runtimetaskspec.UseRequest{
						RequestText: "student gradebook 점수를 학생별로 요약한다.",
						InputFile:   input,
						SourceSheet: "Grades",
						OutputFile:  output,
						TargetSheet: "GradeSummary",
						SummaryMode: "values",
						GroupBy:     []string{"student"},
						Metrics:     []runtimetaskspec.MetricSpec{{Column: "score", Op: "sum", As: "score_total"}},
					}).TaskSpec
				},
			},
			{
				AtomID: "add_data_validation",
				Build: func(input, output string) runtimetaskspec.TaskSpec {
					return runtimetaskspec.BuildAddDataValidationTask(runtimetaskspec.AddDataValidationRequest{
						RequestText: "grade status를 허용된 상태로 제한한다.",
						InputFile:   input,
						SourceSheet: "Grades",
						OutputFile:  output,
						ValidationRule: runtimetaskspec.DataValidationRule{
							Ranges:        []string{"D2:D20"},
							RuleType:      "list",
							AllowedValues: []string{"complete", "missing", "excused"},
							AllowBlank:    false,
						},
					}).TaskSpec
				},
			},
			{
				AtomID: "extend_table_formulas",
				Build: func(input, output string) runtimetaskspec.TaskSpec {
					return runtimetaskspec.BuildExtendTableFormulasTask(runtimetaskspec.ExtendTableFormulasRequest{
						RequestText:      "새 grade row에 weighted score 수식을 확장한다.",
						InputFile:        input,
						SourceSheet:      "Grades",
						OutputFile:       output,
						FormulaSourceRow: 2,
						TargetRows:       []int{4},
						FormulaColumns:   []string{"E"},
					}).TaskSpec
				},
			},
			{
				AtomID: "protect_formula_cells",
				Build: func(input, output string) runtimetaskspec.TaskSpec {
					return runtimetaskspec.BuildProtectFormulaCellsTask(runtimetaskspec.ProtectFormulaCellsRequest{
						RequestText: "gradebook 계산 cells를 보호하고 score/status 입력은 열어둔다.",
						InputFile:   input,
						SourceSheet: "Grades",
						OutputFile:  output,
						ProtectionRule: runtimetaskspec.FormulaProtectionRule{
							FormulaRanges: []string{"E2:E4"},
							InputRanges:   []string{"C2:D20"},
						},
					}).TaskSpec
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("RunOrganismPlan: %v", err)
	}
	if !result.TemplateClassEvaluation.Pass {
		t.Fatalf("template class evaluation failed: %+v", result.TemplateClassEvaluation)
	}
	if result.TemplateClassEvaluation.RuntimeClaim != "template_class_plan_verified_with_non_claims" {
		t.Fatalf("runtime claim=%q want template_class_plan_verified_with_non_claims", result.TemplateClassEvaluation.RuntimeClaim)
	}

	outputHandle, err := excelize.OpenFile(outputFile)
	if err != nil {
		t.Fatalf("Open output: %v", err)
	}
	defer func() { _ = outputHandle.Close() }()
	if got, err := outputHandle.GetCellValue("GradeSummary", "B2"); err != nil || got != "170" {
		t.Fatalf("GradeSummary!B2=%q err=%v want 170", got, err)
	}
	if got, err := outputHandle.GetCellFormula("Grades", "E4"); err != nil || got != "=C4" {
		t.Fatalf("Grades!E4 formula=%q err=%v want =C4", got, err)
	}
}

func TestRunOrganismPlanExecutesBudgetSequence(t *testing.T) {
	setRuntimeRoots(t)

	tempDir := t.TempDir()
	inputFile := filepath.Join(tempDir, "budget.xlsx")
	outputFile := filepath.Join(tempDir, "budget-organism.xlsx")

	file := excelize.NewFile()
	defer func() { _ = file.Close() }()
	defaultSheet := file.GetSheetName(0)
	if err := file.SetSheetName(defaultSheet, "Budget"); err != nil {
		t.Fatalf("SetSheetName: %v", err)
	}
	if err := file.SetSheetRow("Budget", "A1", &[]any{"category", "actual", "budget", "variance", "review_total", "closing"}); err != nil {
		t.Fatalf("SetSheetRow header: %v", err)
	}
	for idx, row := range [][]any{{"Travel", 120, 100, 20}, {"Meals", 80, 90, -10}} {
		cell, _ := excelize.CoordinatesToCellName(1, idx+2)
		if err := file.SetSheetRow("Budget", cell, &row); err != nil {
			t.Fatalf("SetSheetRow budget %d: %v", idx, err)
		}
		rowNum := idx + 2
		if err := file.SetCellFormula("Budget", "E"+cellRow(rowNum), "=B"+cellRow(rowNum)+"+C"+cellRow(rowNum)); err != nil {
			t.Fatalf("SetCellFormula review_total row %d: %v", rowNum, err)
		}
		if err := file.SetCellFormula("Budget", "F"+cellRow(rowNum), "=B"+cellRow(rowNum)+"-C"+cellRow(rowNum)); err != nil {
			t.Fatalf("SetCellFormula closing row %d: %v", rowNum, err)
		}
	}
	if err := file.SaveAs(inputFile); err != nil {
		t.Fatalf("SaveAs: %v", err)
	}

	result, err := RunOrganismPlan(OrganismRunRequest{
		ScenarioID:  "budget-organism-run",
		RequestText: "Summarize monthly budget actuals, flag variance, copy period, roll forward closing, and protect formulas",
		InputFile:   inputFile,
		OutputFile:  outputFile,
		Steps: []OrganismStep{
			{
				AtomID: "group_summarize",
				Build: func(input, output string) runtimetaskspec.TaskSpec {
					return runtimetaskspec.BuildGroupSummarizeTask(runtimetaskspec.UseRequest{
						RequestText: "월 예산 실제 지출을 카테고리별로 요약한다.",
						InputFile:   input,
						SourceSheet: "Budget",
						OutputFile:  output,
						TargetSheet: "BudgetSummary",
						SummaryMode: "values",
						GroupBy:     []string{"category"},
						Metrics:     []runtimetaskspec.MetricSpec{{Column: "actual", Op: "sum", As: "actual_total"}},
					}).TaskSpec
				},
			},
			{
				AtomID: "highlight_threshold",
				Build: func(input, output string) runtimetaskspec.TaskSpec {
					threshold := 0.0
					return runtimetaskspec.BuildHighlightThresholdTask(runtimetaskspec.HighlightThresholdRequest{
						RequestText:    "예산 초과 variance 행을 표시한다.",
						InputFile:      input,
						SourceSheet:    "Budget",
						OutputFile:     output,
						Column:         "variance",
						Operator:       ">",
						Threshold:      &threshold,
						HighlightColor: "#FFF59D",
					}).TaskSpec
				},
			},
			{
				AtomID: "copy_period_sheet",
				Build: func(input, output string) runtimetaskspec.TaskSpec {
					return runtimetaskspec.BuildCopyPeriodSheetTask(runtimetaskspec.CopyPeriodSheetRequest{
						RequestText: "Budget 시트를 NextBudget 기간으로 복사한다.",
						InputFile:   input,
						SourceSheet: "Budget",
						TargetSheet: "NextBudget",
						OutputFile:  output,
					}).TaskSpec
				},
			},
			{
				AtomID: "roll_forward_period",
				Build: func(input, output string) runtimetaskspec.TaskSpec {
					return runtimetaskspec.BuildRollForwardPeriodTask(runtimetaskspec.RollForwardPeriodRequest{
						RequestText: "Budget closing 값을 NextBudget opening input으로 이월한다.",
						InputFile:   input,
						SourceSheet: "Budget",
						TargetSheet: "NextBudget",
						OutputFile:  output,
						CarryForwardMappings: []runtimetaskspec.CarryForwardMapping{
							{FromSheet: "Budget", FromCell: "F2", ToSheet: "NextBudget", ToCell: "B3"},
						},
					}).TaskSpec
				},
			},
			{
				AtomID: "protect_formula_cells",
				Build: func(input, output string) runtimetaskspec.TaskSpec {
					return runtimetaskspec.BuildProtectFormulaCellsTask(runtimetaskspec.ProtectFormulaCellsRequest{
						RequestText: "NextBudget 계산 cells를 보호하고 budget 입력은 열어둔다.",
						InputFile:   input,
						SourceSheet: "NextBudget",
						OutputFile:  output,
						ProtectionRule: runtimetaskspec.FormulaProtectionRule{
							FormulaRanges: []string{"E2:F3"},
							InputRanges:   []string{"B2:D10"},
						},
					}).TaskSpec
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("RunOrganismPlan: %v", err)
	}
	if !result.TemplateClassEvaluation.Pass {
		t.Fatalf("template class evaluation failed: %+v", result.TemplateClassEvaluation)
	}
	if result.TemplateClassEvaluation.RuntimeClaim != "template_class_plan_verified_with_non_claims" {
		t.Fatalf("runtime claim=%q want template_class_plan_verified_with_non_claims", result.TemplateClassEvaluation.RuntimeClaim)
	}
	if len(result.StepResults) != 5 {
		t.Fatalf("step results=%d want 5", len(result.StepResults))
	}

	outputHandle, err := excelize.OpenFile(outputFile)
	if err != nil {
		t.Fatalf("Open output: %v", err)
	}
	defer func() { _ = outputHandle.Close() }()
	if got, err := outputHandle.GetCellValue("BudgetSummary", "B2"); err != nil || got != "120" {
		t.Fatalf("BudgetSummary!B2=%q err=%v want 120", got, err)
	}
	if got, err := outputHandle.GetCellValue("NextBudget", "B3"); err != nil || got != "20" {
		t.Fatalf("NextBudget!B3=%q err=%v want 20", got, err)
	}
	if got, err := outputHandle.GetCellFormula("NextBudget", "E2"); err != nil || got != "=B2+C2" {
		t.Fatalf("NextBudget!E2 formula=%q err=%v want =B2+C2", got, err)
	}
}

func TestRunOrganismPlanExecutesInventorySequence(t *testing.T) {
	setRuntimeRoots(t)

	tempDir := t.TempDir()
	inputFile := filepath.Join(tempDir, "inventory.xlsx")
	outputFile := filepath.Join(tempDir, "inventory-organism.xlsx")

	file := excelize.NewFile()
	defer func() { _ = file.Close() }()
	defaultSheet := file.GetSheetName(0)
	if err := file.SetSheetName(defaultSheet, "Movements"); err != nil {
		t.Fatalf("SetSheetName: %v", err)
	}
	if err := file.SetSheetRow("Movements", "A1", &[]any{"SKU ID", "Qty In", "Qty Out", "Balance"}); err != nil {
		t.Fatalf("SetSheetRow movements header: %v", err)
	}
	for idx, row := range [][]any{{"A001", 10, 0}, {"B002", 3, 1}} {
		cell, _ := excelize.CoordinatesToCellName(1, idx+2)
		if err := file.SetSheetRow("Movements", cell, &row); err != nil {
			t.Fatalf("SetSheetRow movement %d: %v", idx, err)
		}
		rowNum := idx + 2
		if err := file.SetCellFormula("Movements", "D"+cellRow(rowNum), "=B"+cellRow(rowNum)+"-C"+cellRow(rowNum)); err != nil {
			t.Fatalf("SetCellFormula balance row %d: %v", rowNum, err)
		}
	}
	if _, err := file.NewSheet("SKU"); err != nil {
		t.Fatalf("NewSheet SKU: %v", err)
	}
	if err := file.SetSheetRow("SKU", "A1", &[]any{"sku", "location"}); err != nil {
		t.Fatalf("SetSheetRow SKU header: %v", err)
	}
	for idx, row := range [][]any{{"A001", "Aisle 1"}, {"B002", "Aisle 2"}, {"C003", "Aisle 3"}} {
		cell, _ := excelize.CoordinatesToCellName(1, idx+2)
		if err := file.SetSheetRow("SKU", cell, &row); err != nil {
			t.Fatalf("SetSheetRow SKU %d: %v", idx, err)
		}
	}
	if _, err := file.NewSheet("StockMaster"); err != nil {
		t.Fatalf("NewSheet StockMaster: %v", err)
	}
	if err := file.SetSheetRow("StockMaster", "A1", &[]any{"sku", "on_hand"}); err != nil {
		t.Fatalf("SetSheetRow stock header: %v", err)
	}
	for idx, row := range [][]any{{"A001", 10}, {"B002", 2}, {"C003", 2}} {
		cell, _ := excelize.CoordinatesToCellName(1, idx+2)
		if err := file.SetSheetRow("StockMaster", cell, &row); err != nil {
			t.Fatalf("SetSheetRow stock %d: %v", idx, err)
		}
	}
	if err := file.SaveAs(inputFile); err != nil {
		t.Fatalf("SaveAs: %v", err)
	}

	result, err := RunOrganismPlan(OrganismRunRequest{
		ScenarioID:  "inventory-organism-run",
		RequestText: "Normalize inventory movement headers, append SKU movement, lookup SKU metadata, protect balance formulas, and reconcile stock master",
		InputFile:   inputFile,
		OutputFile:  outputFile,
		Steps: []OrganismStep{
			{
				AtomID: "normalize_headers",
				Build: func(input, output string) runtimetaskspec.TaskSpec {
					return runtimetaskspec.BuildNormalizeHeadersTask(runtimetaskspec.NormalizeHeadersRequest{
						RequestText: "inventory movement log 헤더를 표준 필드명으로 정규화한다.",
						InputFile:   input,
						SourceSheet: "Movements",
						OutputFile:  output,
						HeaderRow:   1,
						HeaderMappings: []runtimetaskspec.HeaderMapping{
							{From: "SKU ID", To: "sku"},
							{From: "Qty In", To: "quantity_in"},
							{From: "Qty Out", To: "quantity_out"},
							{From: "Balance", To: "balance"},
						},
					}).TaskSpec
				},
			},
			{
				AtomID: "append_structured_rows",
				Build: func(input, output string) runtimetaskspec.TaskSpec {
					return runtimetaskspec.BuildAppendStructuredRowsTask(runtimetaskspec.AppendStructuredRowsRequest{
						RequestText:          "inventory movement log에 새 SKU movement를 추가한다.",
						InputFile:            input,
						SourceSheet:          "Movements",
						OutputFile:           output,
						IncludeSourceColumns: []string{"sku", "quantity_in", "quantity_out", "balance"},
						Values: []runtimetaskspec.CellValue{
							{Cell: "sku", Value: "C003"},
							{Cell: "quantity_in", Value: 2},
							{Cell: "quantity_out", Value: 0},
							{Cell: "balance", Value: 2},
						},
					}).TaskSpec
				},
			},
			{
				AtomID: "join_lookup",
				Build: func(input, output string) runtimetaskspec.TaskSpec {
					return runtimetaskspec.BuildJoinLookupTask(runtimetaskspec.JoinLookupRequest{
						RequestText:          "inventory movement log에 SKU location을 보강한다.",
						InputFile:            input,
						SourceSheet:          "Movements",
						LookupSheet:          "SKU",
						TargetSheet:          "MovementsEnriched",
						OutputFile:           output,
						JoinKey:              "sku",
						IncludeSourceColumns: []string{"sku", "quantity_in", "quantity_out", "balance"},
						AppendLookupColumns:  []string{"location"},
					}).TaskSpec
				},
			},
			{
				AtomID: "protect_formula_cells",
				Build: func(input, output string) runtimetaskspec.TaskSpec {
					return runtimetaskspec.BuildProtectFormulaCellsTask(runtimetaskspec.ProtectFormulaCellsRequest{
						RequestText: "inventory movement balance formula cells를 보호한다.",
						InputFile:   input,
						SourceSheet: "Movements",
						OutputFile:  output,
						ProtectionRule: runtimetaskspec.FormulaProtectionRule{
							FormulaRanges: []string{"D2:D3"},
							InputRanges:   []string{"A2:C20"},
						},
					}).TaskSpec
				},
			},
			{
				AtomID: "reconcile_tables",
				Build: func(input, output string) runtimetaskspec.TaskSpec {
					return runtimetaskspec.BuildReconcileTablesTask(runtimetaskspec.ReconcileTablesRequest{
						RequestText: "inventory movement balance를 stock master on_hand와 대조한다.",
						InputFile:   input,
						SourceSheet: "MovementsEnriched",
						LookupSheet: "StockMaster",
						TargetSheet: "InventoryReconciliation",
						OutputFile:  output,
						LeftKey:     "sku",
						RightKey:    "sku",
						CompareMappings: []runtimetaskspec.CompareMapping{
							{LeftColumn: "balance", RightColumn: "on_hand", As: "balance"},
						},
					}).TaskSpec
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("RunOrganismPlan: %v", err)
	}
	if !result.TemplateClassEvaluation.Pass {
		t.Fatalf("template class evaluation failed: %+v", result.TemplateClassEvaluation)
	}
	if result.TemplateClassEvaluation.RuntimeClaim != "template_class_plan_verified_with_non_claims" {
		t.Fatalf("runtime claim=%q want template_class_plan_verified_with_non_claims", result.TemplateClassEvaluation.RuntimeClaim)
	}
	if len(result.StepResults) != 5 {
		t.Fatalf("step results=%d want 5", len(result.StepResults))
	}

	outputHandle, err := excelize.OpenFile(outputFile)
	if err != nil {
		t.Fatalf("Open output: %v", err)
	}
	defer func() { _ = outputHandle.Close() }()
	if got, err := outputHandle.GetCellValue("Movements", "A1"); err != nil || got != "sku" {
		t.Fatalf("Movements!A1=%q err=%v want sku", got, err)
	}
	if got, err := outputHandle.GetCellValue("MovementsEnriched", "E4"); err != nil || got != "Aisle 3" {
		t.Fatalf("MovementsEnriched!E4=%q err=%v want Aisle 3", got, err)
	}
	if got, err := outputHandle.GetCellValue("InventoryReconciliation", "B4"); err != nil || got != "C003" {
		t.Fatalf("InventoryReconciliation!B4=%q err=%v want C003", got, err)
	}
}

func cellRow(row int) string {
	return strconv.Itoa(row)
}
