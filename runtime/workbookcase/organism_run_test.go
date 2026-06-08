package workbookcase

import (
	"path/filepath"
	"strings"
	"testing"

	runtimetaskspec "github.com/mwroh/sheet-ops/runtime/taskspec"
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
