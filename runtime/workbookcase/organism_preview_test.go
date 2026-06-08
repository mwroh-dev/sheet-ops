package workbookcase

import (
	"path/filepath"
	"testing"

	runtimetaskspec "github.com/mwroh/sheet-ops/runtime/taskspec"
	"github.com/xuri/excelize/v2"
)

var p0OrganismPreviewCoverage = map[string]string{
	"invoice_line_item_billing": "append + formula extension + validation + protection + printable form",
	"monthly_budget_control":    "group summary + threshold highlight",
	"cash_flow_monitor":         "period roll-forward",
	"attendance_register":       "period copy + validation + formula protection",
	"timesheet_hours_log":       "period copy",
	"inventory_movement_log":    "header normalization + table reconciliation",
}

var p1OrganismPreviewCoverage = map[string]string{
	"expense_reimbursement":      "append + formula extension + printable form",
	"purchase_order_control":     "supplier lookup + status validation",
	"project_timeline_tracker":   "period copy + status validation",
	"shift_roster_planner":       "period copy + shift validation",
	"construction_cost_tracker":  "append + threshold highlight",
	"procurement_reconciliation": "table reconciliation",
	"warehouse_reorder_tracker":  "lookup + threshold highlight",
}

func TestP0OrganismPreviewCoverageIncludesEveryP0Organism(t *testing.T) {
	want := []string{
		"invoice_line_item_billing",
		"monthly_budget_control",
		"cash_flow_monitor",
		"attendance_register",
		"timesheet_hours_log",
		"inventory_movement_log",
	}
	for _, organism := range want {
		if p0OrganismPreviewCoverage[organism] == "" {
			t.Fatalf("missing p0 organism preview coverage for %s", organism)
		}
	}
	if len(p0OrganismPreviewCoverage) != len(want) {
		t.Fatalf("p0 organism preview coverage count=%d want %d", len(p0OrganismPreviewCoverage), len(want))
	}
}

func TestP1OrganismPreviewCoverageIncludesEveryP1Organism(t *testing.T) {
	want := []string{
		"expense_reimbursement",
		"purchase_order_control",
		"project_timeline_tracker",
		"shift_roster_planner",
		"construction_cost_tracker",
		"procurement_reconciliation",
		"warehouse_reorder_tracker",
	}
	for _, organism := range want {
		if p1OrganismPreviewCoverage[organism] == "" {
			t.Fatalf("missing p1 organism preview coverage for %s", organism)
		}
	}
	if len(p1OrganismPreviewCoverage) != len(want) {
		t.Fatalf("p1 organism preview coverage count=%d want %d", len(p1OrganismPreviewCoverage), len(want))
	}
}

func runPreviewTask(t *testing.T, scenarioID string, taskSpec runtimetaskspec.TaskSpec) RunResult {
	t.Helper()
	result, err := Run(Request{ScenarioID: scenarioID, TaskSpec: taskSpec})
	if err != nil {
		t.Fatalf("%s Run: %v", scenarioID, err)
	}
	if !result.Verification.Pass {
		t.Fatalf("%s verification failed: %+v", scenarioID, result.Verification)
	}
	return result
}

func TestInvoiceLineItemBillingPreviewComposesAppendFormulaExtensionValidationAndProtection(t *testing.T) {
	setRuntimeRoots(t)

	tempDir := t.TempDir()
	inputFile := filepath.Join(tempDir, "invoice.xlsx")
	appendOutput := filepath.Join(tempDir, "invoice-appended.xlsx")
	finalOutput := filepath.Join(tempDir, "invoice-final.xlsx")
	validationOutput := filepath.Join(tempDir, "invoice-validated.xlsx")
	protectedOutput := filepath.Join(tempDir, "invoice-protected.xlsx")
	printableOutput := filepath.Join(tempDir, "invoice-printable.xlsx")

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
	row2 := []any{"A001", 2, 10}
	if err := file.SetSheetRow("LineItems", "A2", &row2); err != nil {
		t.Fatalf("SetSheetRow(row2): %v", err)
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

	appendTask := runtimetaskspec.BuildAppendStructuredRowsTask(runtimetaskspec.AppendStructuredRowsRequest{
		RequestText:          "청구서에 새 line item 행을 추가한다.",
		InputFile:            inputFile,
		SourceSheet:          "LineItems",
		OutputFile:           appendOutput,
		IncludeSourceColumns: []string{"sku", "quantity", "unit_price"},
		Values: []runtimetaskspec.CellValue{
			{Cell: "sku", Value: "B002"},
			{Cell: "quantity", Value: 3},
			{Cell: "unit_price", Value: 15},
		},
	})
	appendResult, err := Run(Request{ScenarioID: "invoice-preview-append", TaskSpec: appendTask.TaskSpec})
	if err != nil {
		t.Fatalf("append Run: %v", err)
	}
	if !appendResult.Verification.Pass {
		t.Fatalf("append verification failed: %+v", appendResult.Verification)
	}

	formulaTask := runtimetaskspec.BuildExtendTableFormulasTask(runtimetaskspec.ExtendTableFormulasRequest{
		RequestText:      "새 line item 행에 계산 수식을 확장한다.",
		InputFile:        appendOutput,
		SourceSheet:      "LineItems",
		OutputFile:       finalOutput,
		FormulaSourceRow: 2,
		TargetRows:       []int{3},
		FormulaColumns:   []string{"D", "E"},
	})
	formulaResult, err := Run(Request{ScenarioID: "invoice-preview-formulas", TaskSpec: formulaTask.TaskSpec})
	if err != nil {
		t.Fatalf("formula Run: %v", err)
	}
	if !formulaResult.Verification.Pass {
		t.Fatalf("formula verification failed: %+v", formulaResult.Verification)
	}

	validationTask := runtimetaskspec.BuildAddDataValidationTask(runtimetaskspec.AddDataValidationRequest{
		RequestText: "청구서 line item SKU 입력 범위를 허용된 SKU dropdown으로 제한한다.",
		InputFile:   finalOutput,
		SourceSheet: "LineItems",
		OutputFile:  validationOutput,
		ValidationRule: runtimetaskspec.DataValidationRule{
			Ranges:        []string{"A2:A10"},
			RuleType:      "list",
			AllowedValues: []string{"A001", "B002"},
			AllowBlank:    false,
		},
	})
	validationResult, err := Run(Request{ScenarioID: "invoice-preview-validation", TaskSpec: validationTask.TaskSpec})
	if err != nil {
		t.Fatalf("validation Run: %v", err)
	}
	if !validationResult.Verification.Pass {
		t.Fatalf("validation verification failed: %+v", validationResult.Verification)
	}

	protectionTask := runtimetaskspec.BuildProtectFormulaCellsTask(runtimetaskspec.ProtectFormulaCellsRequest{
		RequestText: "청구서 계산 수식은 보호하고 line item 입력 범위는 편집 가능하게 둔다.",
		InputFile:   validationOutput,
		SourceSheet: "LineItems",
		OutputFile:  protectedOutput,
		ProtectionRule: runtimetaskspec.FormulaProtectionRule{
			FormulaRanges: []string{"D2:E3"},
			InputRanges:   []string{"A2:C10"},
		},
	})
	protectionResult, err := Run(Request{ScenarioID: "invoice-preview-protection", TaskSpec: protectionTask.TaskSpec})
	if err != nil {
		t.Fatalf("protection Run: %v", err)
	}
	if !protectionResult.Verification.Pass {
		t.Fatalf("protection verification failed: %+v", protectionResult.Verification)
	}

	printableTask := runtimetaskspec.BuildGeneratePrintableFormTask(runtimetaskspec.GeneratePrintableFormRequest{
		RequestText: "청구서 line item 데이터를 printable invoice form으로 생성한다.",
		InputFile:   protectedOutput,
		SourceSheet: "LineItems",
		TargetSheet: "InvoicePrint",
		OutputFile:  printableOutput,
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
	})
	printableResult, err := Run(Request{ScenarioID: "invoice-preview-printable-form", TaskSpec: printableTask.TaskSpec})
	if err != nil {
		t.Fatalf("printable Run: %v", err)
	}
	if !printableResult.Verification.Pass {
		t.Fatalf("printable verification failed: %+v", printableResult.Verification)
	}

	outputHandle, err := excelize.OpenFile(printableOutput)
	if err != nil {
		t.Fatalf("Open printable output: %v", err)
	}
	defer func() { _ = outputHandle.Close() }()
	if got, err := outputHandle.GetCellValue("InvoicePrint", "A1"); err != nil || got != "Invoice" {
		t.Fatalf("InvoicePrint!A1=%q err=%v want Invoice", got, err)
	}
	if got, err := outputHandle.GetCellValue("InvoicePrint", "A5"); err != nil || got != "A001" {
		t.Fatalf("InvoicePrint!A5=%q err=%v want A001", got, err)
	}
	if got, err := outputHandle.GetCellValue("LineItems", "A3"); err != nil || got != "B002" {
		t.Fatalf("LineItems!A3=%q err=%v want B002", got, err)
	}
	if got, err := outputHandle.GetCellFormula("LineItems", "D3"); err != nil || got != "=B3*C3" {
		t.Fatalf("LineItems!D3 formula=%q err=%v want =B3*C3", got, err)
	}
	if got, err := outputHandle.GetCellFormula("LineItems", "E3"); err != nil || got != "=D3*0.1" {
		t.Fatalf("LineItems!E3 formula=%q err=%v want =D3*0.1", got, err)
	}
	validations, err := outputHandle.GetDataValidations("LineItems")
	if err != nil {
		t.Fatalf("GetDataValidations: %v", err)
	}
	foundValidation := false
	for _, validation := range validations {
		if validation != nil && validation.Sqref == "A2:A10" && validation.Type == "list" {
			foundValidation = true
			break
		}
	}
	if !foundValidation {
		t.Fatalf("expected list validation on LineItems!A2:A10, got %+v", validations)
	}
	protection, err := outputHandle.GetSheetProtection("LineItems")
	if err != nil {
		t.Fatalf("GetSheetProtection: %v", err)
	}
	if !protection.SelectLockedCells || !protection.SelectUnlockedCells {
		t.Fatalf("expected sheet protection options, got %+v", protection)
	}
}

func TestTimesheetHoursLogPreviewComposesPeriodCopy(t *testing.T) {
	setRuntimeRoots(t)

	tempDir := t.TempDir()
	inputFile := filepath.Join(tempDir, "timesheet.xlsx")
	outputFile := filepath.Join(tempDir, "timesheet-week2.xlsx")

	file := excelize.NewFile()
	defer func() { _ = file.Close() }()
	defaultSheet := file.GetSheetName(0)
	if err := file.SetSheetName(defaultSheet, "Week1"); err != nil {
		t.Fatalf("SetSheetName: %v", err)
	}
	header := []any{"date", "hours", "rate", "pay"}
	if err := file.SetSheetRow("Week1", "A1", &header); err != nil {
		t.Fatalf("SetSheetRow(header): %v", err)
	}
	row := []any{"2026-06-01", 8, 25}
	if err := file.SetSheetRow("Week1", "A2", &row); err != nil {
		t.Fatalf("SetSheetRow(row): %v", err)
	}
	if err := file.SetCellFormula("Week1", "D2", "=B2*C2"); err != nil {
		t.Fatalf("SetCellFormula(D2): %v", err)
	}
	if err := file.SaveAs(inputFile); err != nil {
		t.Fatalf("SaveAs: %v", err)
	}

	copyTask := runtimetaskspec.BuildCopyPeriodSheetTask(runtimetaskspec.CopyPeriodSheetRequest{
		RequestText: "Week1 timesheet 구조를 Week2 시트로 복사한다.",
		InputFile:   inputFile,
		SourceSheet: "Week1",
		TargetSheet: "Week2",
		OutputFile:  outputFile,
	})
	copyResult, err := Run(Request{ScenarioID: "timesheet-preview-period-copy", TaskSpec: copyTask.TaskSpec})
	if err != nil {
		t.Fatalf("copy Run: %v", err)
	}
	if !copyResult.Verification.Pass {
		t.Fatalf("copy verification failed: %+v", copyResult.Verification)
	}

	outputHandle, err := excelize.OpenFile(outputFile)
	if err != nil {
		t.Fatalf("Open output: %v", err)
	}
	defer func() { _ = outputHandle.Close() }()
	if got, err := outputHandle.GetCellValue("Week2", "A2"); err != nil || got != "2026-06-01" {
		t.Fatalf("Week2!A2=%q err=%v want 2026-06-01", got, err)
	}
	if got, err := outputHandle.GetCellFormula("Week2", "D2"); err != nil || got != "=B2*C2" {
		t.Fatalf("Week2!D2 formula=%q err=%v want =B2*C2", got, err)
	}
}

func TestMonthlyBudgetControlPreviewComposesSummaryAndThreshold(t *testing.T) {
	setRuntimeRoots(t)

	tempDir := t.TempDir()
	inputFile := filepath.Join(tempDir, "budget.xlsx")
	summaryOutput := filepath.Join(tempDir, "budget-summary.xlsx")
	finalOutput := filepath.Join(tempDir, "budget-highlighted.xlsx")

	file := excelize.NewFile()
	defer func() { _ = file.Close() }()
	defaultSheet := file.GetSheetName(0)
	if err := file.SetSheetName(defaultSheet, "Budget"); err != nil {
		t.Fatalf("SetSheetName: %v", err)
	}
	header := []any{"category", "actual", "budget", "variance"}
	if err := file.SetSheetRow("Budget", "A1", &header); err != nil {
		t.Fatalf("SetSheetRow(header): %v", err)
	}
	for idx, row := range [][]any{{"Travel", 120, 100, 20}, {"Meals", 80, 90, -10}} {
		cell, _ := excelize.CoordinatesToCellName(1, idx+2)
		if err := file.SetSheetRow("Budget", cell, &row); err != nil {
			t.Fatalf("SetSheetRow budget %d: %v", idx, err)
		}
	}
	if err := file.SaveAs(inputFile); err != nil {
		t.Fatalf("SaveAs: %v", err)
	}

	summaryTask := runtimetaskspec.BuildGroupSummarizeTask(runtimetaskspec.UseRequest{
		RequestText: "월 예산 실제 지출을 카테고리별로 요약한다.",
		InputFile:   inputFile,
		SourceSheet: "Budget",
		OutputFile:  summaryOutput,
		TargetSheet: "BudgetSummary",
		SummaryMode: "values",
		GroupBy:     []string{"category"},
		Metrics:     []runtimetaskspec.MetricSpec{{Column: "actual", Op: "sum", As: "actual_total"}},
	})
	summaryResult, err := Run(Request{ScenarioID: "budget-preview-summary", TaskSpec: summaryTask.TaskSpec})
	if err != nil {
		t.Fatalf("summary Run: %v", err)
	}
	if !summaryResult.Verification.Pass {
		t.Fatalf("summary verification failed: %+v", summaryResult.Verification)
	}

	threshold := 0.0
	highlightTask := runtimetaskspec.BuildHighlightThresholdTask(runtimetaskspec.HighlightThresholdRequest{
		RequestText:    "예산 초과 variance 행을 표시한다.",
		InputFile:      summaryOutput,
		SourceSheet:    "Budget",
		OutputFile:     finalOutput,
		Column:         "variance",
		Operator:       ">",
		Threshold:      &threshold,
		HighlightColor: "#FFF59D",
	})
	highlightResult, err := Run(Request{ScenarioID: "budget-preview-threshold", TaskSpec: highlightTask.TaskSpec})
	if err != nil {
		t.Fatalf("highlight Run: %v", err)
	}
	if !highlightResult.Verification.Pass {
		t.Fatalf("highlight verification failed: %+v", highlightResult.Verification)
	}
	if len(highlightResult.Verification.HighlightedRows) != 1 || highlightResult.Verification.HighlightedRows[0] != 2 {
		t.Fatalf("highlighted rows=%v want [2]", highlightResult.Verification.HighlightedRows)
	}

	outputHandle, err := excelize.OpenFile(finalOutput)
	if err != nil {
		t.Fatalf("Open final output: %v", err)
	}
	defer func() { _ = outputHandle.Close() }()
	if got, err := outputHandle.GetCellValue("BudgetSummary", "B2"); err != nil || got != "120" {
		t.Fatalf("BudgetSummary!B2=%q err=%v want 120", got, err)
	}
}

func TestAttendanceRegisterPreviewComposesCopyValidationAndProtection(t *testing.T) {
	setRuntimeRoots(t)

	tempDir := t.TempDir()
	inputFile := filepath.Join(tempDir, "attendance.xlsx")
	copyOutput := filepath.Join(tempDir, "attendance-week2.xlsx")
	validationOutput := filepath.Join(tempDir, "attendance-validated.xlsx")
	finalOutput := filepath.Join(tempDir, "attendance-protected.xlsx")

	file := excelize.NewFile()
	defer func() { _ = file.Close() }()
	defaultSheet := file.GetSheetName(0)
	if err := file.SetSheetName(defaultSheet, "Week1"); err != nil {
		t.Fatalf("SetSheetName: %v", err)
	}
	header := []any{"student", "mon", "tue", "present_total"}
	if err := file.SetSheetRow("Week1", "A1", &header); err != nil {
		t.Fatalf("SetSheetRow(header): %v", err)
	}
	row := []any{"Student A", "P", "A"}
	if err := file.SetSheetRow("Week1", "A2", &row); err != nil {
		t.Fatalf("SetSheetRow(row): %v", err)
	}
	if err := file.SetCellFormula("Week1", "D2", "=COUNTIF(B2:C2,\"P\")"); err != nil {
		t.Fatalf("SetCellFormula(D2): %v", err)
	}
	if err := file.SaveAs(inputFile); err != nil {
		t.Fatalf("SaveAs: %v", err)
	}

	copyTask := runtimetaskspec.BuildCopyPeriodSheetTask(runtimetaskspec.CopyPeriodSheetRequest{
		RequestText: "Week1 attendance register를 Week2 시트로 복사한다.",
		InputFile:   inputFile,
		SourceSheet: "Week1",
		TargetSheet: "Week2",
		OutputFile:  copyOutput,
	})
	copyResult, err := Run(Request{ScenarioID: "attendance-preview-copy", TaskSpec: copyTask.TaskSpec})
	if err != nil {
		t.Fatalf("copy Run: %v", err)
	}
	if !copyResult.Verification.Pass {
		t.Fatalf("copy verification failed: %+v", copyResult.Verification)
	}

	validationTask := runtimetaskspec.BuildAddDataValidationTask(runtimetaskspec.AddDataValidationRequest{
		RequestText: "Week2 attendance 입력 범위를 P/A dropdown으로 제한한다.",
		InputFile:   copyOutput,
		SourceSheet: "Week2",
		OutputFile:  validationOutput,
		ValidationRule: runtimetaskspec.DataValidationRule{
			Ranges:        []string{"B2:C10"},
			RuleType:      "list",
			AllowedValues: []string{"P", "A"},
			AllowBlank:    false,
		},
	})
	validationResult, err := Run(Request{ScenarioID: "attendance-preview-validation", TaskSpec: validationTask.TaskSpec})
	if err != nil {
		t.Fatalf("validation Run: %v", err)
	}
	if !validationResult.Verification.Pass {
		t.Fatalf("validation verification failed: %+v", validationResult.Verification)
	}

	protectionTask := runtimetaskspec.BuildProtectFormulaCellsTask(runtimetaskspec.ProtectFormulaCellsRequest{
		RequestText: "Week2 attendance formula column을 보호하고 입력 범위는 편집 가능하게 둔다.",
		InputFile:   validationOutput,
		SourceSheet: "Week2",
		OutputFile:  finalOutput,
		ProtectionRule: runtimetaskspec.FormulaProtectionRule{
			FormulaRanges: []string{"D2"},
			InputRanges:   []string{"B2:C10"},
		},
	})
	protectionResult, err := Run(Request{ScenarioID: "attendance-preview-protection", TaskSpec: protectionTask.TaskSpec})
	if err != nil {
		t.Fatalf("protection Run: %v", err)
	}
	if !protectionResult.Verification.Pass {
		t.Fatalf("protection verification failed: %+v", protectionResult.Verification)
	}

	outputHandle, err := excelize.OpenFile(finalOutput)
	if err != nil {
		t.Fatalf("Open final output: %v", err)
	}
	defer func() { _ = outputHandle.Close() }()
	if got, err := outputHandle.GetCellFormula("Week2", "D2"); err != nil || got != "=COUNTIF(B2:C2,\"P\")" {
		t.Fatalf("Week2!D2 formula=%q err=%v want COUNTIF", got, err)
	}
}

func TestExpenseReimbursementPreviewComposesAppendFormulasAndPrintableForm(t *testing.T) {
	setRuntimeRoots(t)

	tempDir := t.TempDir()
	inputFile := filepath.Join(tempDir, "expense.xlsx")
	appendOutput := filepath.Join(tempDir, "expense-appended.xlsx")
	formulaOutput := filepath.Join(tempDir, "expense-formulas.xlsx")
	printableOutput := filepath.Join(tempDir, "expense-printable.xlsx")

	file := excelize.NewFile()
	defer func() { _ = file.Close() }()
	defaultSheet := file.GetSheetName(0)
	if err := file.SetSheetName(defaultSheet, "ExpenseItems"); err != nil {
		t.Fatalf("SetSheetName: %v", err)
	}
	header := []any{"item", "amount", "reimbursable_rate", "reimbursable_total"}
	if err := file.SetSheetRow("ExpenseItems", "A1", &header); err != nil {
		t.Fatalf("SetSheetRow(header): %v", err)
	}
	row := []any{"Hotel", 200, 1}
	if err := file.SetSheetRow("ExpenseItems", "A2", &row); err != nil {
		t.Fatalf("SetSheetRow(row): %v", err)
	}
	if err := file.SetCellFormula("ExpenseItems", "D2", "=B2*C2"); err != nil {
		t.Fatalf("SetCellFormula(D2): %v", err)
	}
	if err := file.SaveAs(inputFile); err != nil {
		t.Fatalf("SaveAs: %v", err)
	}

	appendTask := runtimetaskspec.BuildAppendStructuredRowsTask(runtimetaskspec.AppendStructuredRowsRequest{
		RequestText:          "경비 정산 항목을 추가한다.",
		InputFile:            inputFile,
		SourceSheet:          "ExpenseItems",
		OutputFile:           appendOutput,
		IncludeSourceColumns: []string{"item", "amount", "reimbursable_rate"},
		Values: []runtimetaskspec.CellValue{
			{Cell: "item", Value: "Taxi"},
			{Cell: "amount", Value: 35},
			{Cell: "reimbursable_rate", Value: 1},
		},
	})
	runPreviewTask(t, "expense-preview-append", appendTask.TaskSpec)

	formulaTask := runtimetaskspec.BuildExtendTableFormulasTask(runtimetaskspec.ExtendTableFormulasRequest{
		RequestText:      "새 경비 행에 reimbursable total 수식을 확장한다.",
		InputFile:        appendOutput,
		SourceSheet:      "ExpenseItems",
		OutputFile:       formulaOutput,
		FormulaSourceRow: 2,
		TargetRows:       []int{3},
		FormulaColumns:   []string{"D"},
	})
	runPreviewTask(t, "expense-preview-formulas", formulaTask.TaskSpec)

	printableTask := runtimetaskspec.BuildGeneratePrintableFormTask(runtimetaskspec.GeneratePrintableFormRequest{
		RequestText: "경비 정산 항목을 printable reimbursement claim으로 생성한다.",
		InputFile:   formulaOutput,
		SourceSheet: "ExpenseItems",
		TargetSheet: "ExpenseClaim",
		OutputFile:  printableOutput,
		FormTitle:   "Expense Reimbursement",
		PrintArea:   "A1:D8",
		FieldBindings: []runtimetaskspec.FormFieldBinding{
			{Label: "First Item", SourceSheet: "ExpenseItems", SourceCell: "A2", LabelCell: "A2", ValueCell: "B2"},
		},
		TableBinding: &runtimetaskspec.FormTableBinding{
			SourceSheet:   "ExpenseItems",
			SourceColumns: []string{"item", "amount", "reimbursable_rate", "reimbursable_total"},
			HeaderStart:   "A3",
			DataStart:     "A4",
		},
	})
	runPreviewTask(t, "expense-preview-printable-form", printableTask.TaskSpec)

	outputHandle, err := excelize.OpenFile(printableOutput)
	if err != nil {
		t.Fatalf("Open output: %v", err)
	}
	defer func() { _ = outputHandle.Close() }()
	if got, err := outputHandle.GetCellFormula("ExpenseItems", "D3"); err != nil || got != "=B3*C3" {
		t.Fatalf("ExpenseItems!D3 formula=%q err=%v want =B3*C3", got, err)
	}
	if got, err := outputHandle.GetCellValue("ExpenseClaim", "A4"); err != nil || got != "Hotel" {
		t.Fatalf("ExpenseClaim!A4=%q err=%v want Hotel", got, err)
	}
}

func TestPurchaseOrderControlPreviewComposesLookupAndStatusValidation(t *testing.T) {
	setRuntimeRoots(t)

	tempDir := t.TempDir()
	inputFile := filepath.Join(tempDir, "po.xlsx")
	lookupOutput := filepath.Join(tempDir, "po-lookup.xlsx")
	finalOutput := filepath.Join(tempDir, "po-validated.xlsx")

	file := excelize.NewFile()
	defer func() { _ = file.Close() }()
	defaultSheet := file.GetSheetName(0)
	if err := file.SetSheetName(defaultSheet, "PO"); err != nil {
		t.Fatalf("SetSheetName: %v", err)
	}
	if err := file.SetSheetRow("PO", "A1", &[]any{"supplier_id", "status", "quantity"}); err != nil {
		t.Fatalf("SetSheetRow PO header: %v", err)
	}
	if err := file.SetSheetRow("PO", "A2", &[]any{"S001", "draft", 2}); err != nil {
		t.Fatalf("SetSheetRow PO row: %v", err)
	}
	if _, err := file.NewSheet("Suppliers"); err != nil {
		t.Fatalf("NewSheet Suppliers: %v", err)
	}
	if err := file.SetSheetRow("Suppliers", "A1", &[]any{"supplier_id", "supplier_name"}); err != nil {
		t.Fatalf("SetSheetRow suppliers header: %v", err)
	}
	if err := file.SetSheetRow("Suppliers", "A2", &[]any{"S001", "Acme Supply"}); err != nil {
		t.Fatalf("SetSheetRow suppliers row: %v", err)
	}
	if err := file.SaveAs(inputFile); err != nil {
		t.Fatalf("SaveAs: %v", err)
	}

	lookupTask := runtimetaskspec.BuildJoinLookupTask(runtimetaskspec.JoinLookupRequest{
		RequestText:          "PO 공급자 id를 supplier master로 보강한다.",
		InputFile:            inputFile,
		SourceSheet:          "PO",
		LookupSheet:          "Suppliers",
		TargetSheet:          "POEnriched",
		OutputFile:           lookupOutput,
		JoinKey:              "supplier_id",
		IncludeSourceColumns: []string{"supplier_id", "status", "quantity"},
		AppendLookupColumns:  []string{"supplier_name"},
	})
	runPreviewTask(t, "po-preview-lookup", lookupTask.TaskSpec)

	validationTask := runtimetaskspec.BuildAddDataValidationTask(runtimetaskspec.AddDataValidationRequest{
		RequestText: "PO status를 허용된 pipeline 값으로 제한한다.",
		InputFile:   lookupOutput,
		SourceSheet: "POEnriched",
		OutputFile:  finalOutput,
		ValidationRule: runtimetaskspec.DataValidationRule{
			Ranges:        []string{"B2:B10"},
			RuleType:      "list",
			AllowedValues: []string{"draft", "sent", "received", "closed"},
			AllowBlank:    false,
		},
	})
	runPreviewTask(t, "po-preview-validation", validationTask.TaskSpec)

	outputHandle, err := excelize.OpenFile(finalOutput)
	if err != nil {
		t.Fatalf("Open output: %v", err)
	}
	defer func() { _ = outputHandle.Close() }()
	if got, err := outputHandle.GetCellValue("POEnriched", "D2"); err != nil || got != "Acme Supply" {
		t.Fatalf("POEnriched!D2=%q err=%v want Acme Supply", got, err)
	}
}

func TestProjectTimelineTrackerPreviewComposesPeriodCopyAndStatusValidation(t *testing.T) {
	setRuntimeRoots(t)

	tempDir := t.TempDir()
	inputFile := filepath.Join(tempDir, "timeline.xlsx")
	copyOutput := filepath.Join(tempDir, "timeline-copy.xlsx")
	finalOutput := filepath.Join(tempDir, "timeline-validated.xlsx")

	file := excelize.NewFile()
	defer func() { _ = file.Close() }()
	defaultSheet := file.GetSheetName(0)
	if err := file.SetSheetName(defaultSheet, "Sprint1"); err != nil {
		t.Fatalf("SetSheetName: %v", err)
	}
	if err := file.SetSheetRow("Sprint1", "A1", &[]any{"task", "start", "status"}); err != nil {
		t.Fatalf("SetSheetRow header: %v", err)
	}
	if err := file.SetSheetRow("Sprint1", "A2", &[]any{"Build import", "2026-06-01", "todo"}); err != nil {
		t.Fatalf("SetSheetRow row: %v", err)
	}
	if err := file.SaveAs(inputFile); err != nil {
		t.Fatalf("SaveAs: %v", err)
	}

	copyTask := runtimetaskspec.BuildCopyPeriodSheetTask(runtimetaskspec.CopyPeriodSheetRequest{
		RequestText: "Sprint1 timeline을 Sprint2로 복사한다.",
		InputFile:   inputFile,
		SourceSheet: "Sprint1",
		TargetSheet: "Sprint2",
		OutputFile:  copyOutput,
	})
	runPreviewTask(t, "project-timeline-preview-copy", copyTask.TaskSpec)

	validationTask := runtimetaskspec.BuildAddDataValidationTask(runtimetaskspec.AddDataValidationRequest{
		RequestText: "Sprint2 task status를 pipeline 값으로 제한한다.",
		InputFile:   copyOutput,
		SourceSheet: "Sprint2",
		OutputFile:  finalOutput,
		ValidationRule: runtimetaskspec.DataValidationRule{
			Ranges:        []string{"C2:C20"},
			RuleType:      "list",
			AllowedValues: []string{"todo", "doing", "done"},
			AllowBlank:    false,
		},
	})
	runPreviewTask(t, "project-timeline-preview-validation", validationTask.TaskSpec)
}

func TestShiftRosterPlannerPreviewComposesPeriodCopyAndShiftValidation(t *testing.T) {
	setRuntimeRoots(t)

	tempDir := t.TempDir()
	inputFile := filepath.Join(tempDir, "roster.xlsx")
	copyOutput := filepath.Join(tempDir, "roster-copy.xlsx")
	finalOutput := filepath.Join(tempDir, "roster-validated.xlsx")

	file := excelize.NewFile()
	defer func() { _ = file.Close() }()
	defaultSheet := file.GetSheetName(0)
	if err := file.SetSheetName(defaultSheet, "Week1"); err != nil {
		t.Fatalf("SetSheetName: %v", err)
	}
	if err := file.SetSheetRow("Week1", "A1", &[]any{"employee", "date", "shift"}); err != nil {
		t.Fatalf("SetSheetRow header: %v", err)
	}
	if err := file.SetSheetRow("Week1", "A2", &[]any{"Alex", "2026-06-01", "AM"}); err != nil {
		t.Fatalf("SetSheetRow row: %v", err)
	}
	if err := file.SaveAs(inputFile); err != nil {
		t.Fatalf("SaveAs: %v", err)
	}

	copyTask := runtimetaskspec.BuildCopyPeriodSheetTask(runtimetaskspec.CopyPeriodSheetRequest{
		RequestText: "Week1 shift roster를 Week2로 복사한다.",
		InputFile:   inputFile,
		SourceSheet: "Week1",
		TargetSheet: "Week2",
		OutputFile:  copyOutput,
	})
	runPreviewTask(t, "shift-roster-preview-copy", copyTask.TaskSpec)

	validationTask := runtimetaskspec.BuildAddDataValidationTask(runtimetaskspec.AddDataValidationRequest{
		RequestText: "Week2 shift 값을 허용된 코드로 제한한다.",
		InputFile:   copyOutput,
		SourceSheet: "Week2",
		OutputFile:  finalOutput,
		ValidationRule: runtimetaskspec.DataValidationRule{
			Ranges:        []string{"C2:C20"},
			RuleType:      "list",
			AllowedValues: []string{"AM", "PM", "OFF"},
			AllowBlank:    false,
		},
	})
	runPreviewTask(t, "shift-roster-preview-validation", validationTask.TaskSpec)
}

func TestConstructionCostTrackerPreviewComposesAppendAndThreshold(t *testing.T) {
	setRuntimeRoots(t)

	tempDir := t.TempDir()
	inputFile := filepath.Join(tempDir, "construction-cost.xlsx")
	appendOutput := filepath.Join(tempDir, "construction-cost-appended.xlsx")
	finalOutput := filepath.Join(tempDir, "construction-cost-highlighted.xlsx")

	file := excelize.NewFile()
	defer func() { _ = file.Close() }()
	defaultSheet := file.GetSheetName(0)
	if err := file.SetSheetName(defaultSheet, "Costs"); err != nil {
		t.Fatalf("SetSheetName: %v", err)
	}
	if err := file.SetSheetRow("Costs", "A1", &[]any{"cost_code", "actual", "budget", "variance"}); err != nil {
		t.Fatalf("SetSheetRow header: %v", err)
	}
	if err := file.SetSheetRow("Costs", "A2", &[]any{"LABOR", 90, 100, -10}); err != nil {
		t.Fatalf("SetSheetRow row: %v", err)
	}
	if err := file.SaveAs(inputFile); err != nil {
		t.Fatalf("SaveAs: %v", err)
	}

	appendTask := runtimetaskspec.BuildAppendStructuredRowsTask(runtimetaskspec.AppendStructuredRowsRequest{
		RequestText:          "construction cost tracker에 change cost row를 추가한다.",
		InputFile:            inputFile,
		SourceSheet:          "Costs",
		OutputFile:           appendOutput,
		IncludeSourceColumns: []string{"cost_code", "actual", "budget", "variance"},
		Values: []runtimetaskspec.CellValue{
			{Cell: "cost_code", Value: "CHANGE"},
			{Cell: "actual", Value: 150},
			{Cell: "budget", Value: 100},
			{Cell: "variance", Value: 50},
		},
	})
	runPreviewTask(t, "construction-cost-preview-append", appendTask.TaskSpec)

	threshold := 0.0
	highlightTask := runtimetaskspec.BuildHighlightThresholdTask(runtimetaskspec.HighlightThresholdRequest{
		RequestText:    "budget overrun variance를 표시한다.",
		InputFile:      appendOutput,
		SourceSheet:    "Costs",
		OutputFile:     finalOutput,
		Column:         "variance",
		Operator:       ">",
		Threshold:      &threshold,
		HighlightColor: "#FFC7CE",
	})
	result := runPreviewTask(t, "construction-cost-preview-threshold", highlightTask.TaskSpec)
	if len(result.Verification.HighlightedRows) != 1 || result.Verification.HighlightedRows[0] != 3 {
		t.Fatalf("highlighted rows=%v want [3]", result.Verification.HighlightedRows)
	}
}

func TestProcurementReconciliationPreviewComposesTableReconciliation(t *testing.T) {
	setRuntimeRoots(t)

	tempDir := t.TempDir()
	inputFile := filepath.Join(tempDir, "procurement.xlsx")
	outputFile := filepath.Join(tempDir, "procurement-reconciled.xlsx")

	file := excelize.NewFile()
	defer func() { _ = file.Close() }()
	defaultSheet := file.GetSheetName(0)
	if err := file.SetSheetName(defaultSheet, "PO"); err != nil {
		t.Fatalf("SetSheetName: %v", err)
	}
	if err := file.SetSheetRow("PO", "A1", &[]any{"po_id", "amount"}); err != nil {
		t.Fatalf("SetSheetRow PO header: %v", err)
	}
	if err := file.SetSheetRow("PO", "A2", &[]any{"PO-1", 100}); err != nil {
		t.Fatalf("SetSheetRow PO row: %v", err)
	}
	if _, err := file.NewSheet("Invoice"); err != nil {
		t.Fatalf("NewSheet Invoice: %v", err)
	}
	if err := file.SetSheetRow("Invoice", "A1", &[]any{"po_id", "invoice_amount"}); err != nil {
		t.Fatalf("SetSheetRow invoice header: %v", err)
	}
	if err := file.SetSheetRow("Invoice", "A2", &[]any{"PO-1", 125}); err != nil {
		t.Fatalf("SetSheetRow invoice row: %v", err)
	}
	if err := file.SaveAs(inputFile); err != nil {
		t.Fatalf("SaveAs: %v", err)
	}

	reconcileTask := runtimetaskspec.BuildReconcileTablesTask(runtimetaskspec.ReconcileTablesRequest{
		RequestText: "PO amount와 invoice amount를 대조한다.",
		InputFile:   inputFile,
		SourceSheet: "PO",
		LookupSheet: "Invoice",
		TargetSheet: "ProcurementReconciliation",
		OutputFile:  outputFile,
		LeftKey:     "po_id",
		RightKey:    "po_id",
		CompareMappings: []runtimetaskspec.CompareMapping{
			{LeftColumn: "amount", RightColumn: "invoice_amount", As: "amount"},
		},
	})
	result := runPreviewTask(t, "procurement-preview-reconciliation", reconcileTask.TaskSpec)
	if result.Verification.SummaryRows != 1 {
		t.Fatalf("summary rows=%d want 1", result.Verification.SummaryRows)
	}
}

func TestWarehouseReorderTrackerPreviewComposesLookupAndThreshold(t *testing.T) {
	setRuntimeRoots(t)

	tempDir := t.TempDir()
	inputFile := filepath.Join(tempDir, "warehouse.xlsx")
	lookupOutput := filepath.Join(tempDir, "warehouse-lookup.xlsx")
	finalOutput := filepath.Join(tempDir, "warehouse-highlighted.xlsx")

	file := excelize.NewFile()
	defer func() { _ = file.Close() }()
	defaultSheet := file.GetSheetName(0)
	if err := file.SetSheetName(defaultSheet, "Stock"); err != nil {
		t.Fatalf("SetSheetName: %v", err)
	}
	if err := file.SetSheetRow("Stock", "A1", &[]any{"sku", "quantity", "reorder_level"}); err != nil {
		t.Fatalf("SetSheetRow stock header: %v", err)
	}
	if err := file.SetSheetRow("Stock", "A2", &[]any{"A001", 3, 5}); err != nil {
		t.Fatalf("SetSheetRow stock row: %v", err)
	}
	if _, err := file.NewSheet("SKU"); err != nil {
		t.Fatalf("NewSheet SKU: %v", err)
	}
	if err := file.SetSheetRow("SKU", "A1", &[]any{"sku", "location"}); err != nil {
		t.Fatalf("SetSheetRow SKU header: %v", err)
	}
	if err := file.SetSheetRow("SKU", "A2", &[]any{"A001", "Aisle 1"}); err != nil {
		t.Fatalf("SetSheetRow SKU row: %v", err)
	}
	if err := file.SaveAs(inputFile); err != nil {
		t.Fatalf("SaveAs: %v", err)
	}

	lookupTask := runtimetaskspec.BuildJoinLookupTask(runtimetaskspec.JoinLookupRequest{
		RequestText:          "warehouse reorder tracker에 SKU location을 보강한다.",
		InputFile:            inputFile,
		SourceSheet:          "Stock",
		LookupSheet:          "SKU",
		TargetSheet:          "StockEnriched",
		OutputFile:           lookupOutput,
		JoinKey:              "sku",
		IncludeSourceColumns: []string{"sku", "quantity", "reorder_level"},
		AppendLookupColumns:  []string{"location"},
	})
	runPreviewTask(t, "warehouse-preview-lookup", lookupTask.TaskSpec)

	threshold := 5.0
	highlightTask := runtimetaskspec.BuildHighlightThresholdTask(runtimetaskspec.HighlightThresholdRequest{
		RequestText:    "low stock quantity를 표시한다.",
		InputFile:      lookupOutput,
		SourceSheet:    "StockEnriched",
		OutputFile:     finalOutput,
		Column:         "quantity",
		Operator:       "<",
		Threshold:      &threshold,
		HighlightColor: "#FFF59D",
	})
	result := runPreviewTask(t, "warehouse-preview-threshold", highlightTask.TaskSpec)
	if len(result.Verification.HighlightedRows) != 1 || result.Verification.HighlightedRows[0] != 2 {
		t.Fatalf("highlighted rows=%v want [2]", result.Verification.HighlightedRows)
	}
}

func TestInventoryMovementLogPreviewComposesHeaderNormalization(t *testing.T) {
	setRuntimeRoots(t)

	tempDir := t.TempDir()
	inputFile := filepath.Join(tempDir, "inventory.xlsx")
	outputFile := filepath.Join(tempDir, "inventory-normalized.xlsx")

	file := excelize.NewFile()
	defer func() { _ = file.Close() }()
	defaultSheet := file.GetSheetName(0)
	if err := file.SetSheetName(defaultSheet, "Movements"); err != nil {
		t.Fatalf("SetSheetName: %v", err)
	}
	header := []any{"SKU ID", "Qty In", "Qty Out", "Balance"}
	if err := file.SetSheetRow("Movements", "A1", &header); err != nil {
		t.Fatalf("SetSheetRow(header): %v", err)
	}
	row := []any{"A001", 5, 1}
	if err := file.SetSheetRow("Movements", "A2", &row); err != nil {
		t.Fatalf("SetSheetRow(row): %v", err)
	}
	if err := file.SetCellFormula("Movements", "D2", "=B2-C2"); err != nil {
		t.Fatalf("SetCellFormula(D2): %v", err)
	}
	if err := file.SaveAs(inputFile); err != nil {
		t.Fatalf("SaveAs: %v", err)
	}

	normalizeTask := runtimetaskspec.BuildNormalizeHeadersTask(runtimetaskspec.NormalizeHeadersRequest{
		RequestText: "inventory movement log 헤더를 reconciliation 전에 표준 필드명으로 정규화한다.",
		InputFile:   inputFile,
		SourceSheet: "Movements",
		OutputFile:  outputFile,
		HeaderRow:   1,
		HeaderMappings: []runtimetaskspec.HeaderMapping{
			{From: "SKU ID", To: "sku"},
			{From: "Qty In", To: "quantity_in"},
			{From: "Qty Out", To: "quantity_out"},
			{From: "Balance", To: "balance"},
		},
	})
	result, err := Run(Request{ScenarioID: "inventory-preview-normalize-headers", TaskSpec: normalizeTask.TaskSpec})
	if err != nil {
		t.Fatalf("normalize Run: %v", err)
	}
	if !result.Verification.Pass {
		t.Fatalf("normalize verification failed: %+v", result.Verification)
	}

	outputHandle, err := excelize.OpenFile(outputFile)
	if err != nil {
		t.Fatalf("Open output: %v", err)
	}
	defer func() { _ = outputHandle.Close() }()
	if got, err := outputHandle.GetCellValue("Movements", "A1"); err != nil || got != "sku" {
		t.Fatalf("Movements!A1=%q err=%v want sku", got, err)
	}
	if got, err := outputHandle.GetCellValue("Movements", "A2"); err != nil || got != "A001" {
		t.Fatalf("Movements!A2=%q err=%v want A001", got, err)
	}
	if got, err := outputHandle.GetCellFormula("Movements", "D2"); err != nil || got != "=B2-C2" {
		t.Fatalf("Movements!D2 formula=%q err=%v want =B2-C2", got, err)
	}
}

func TestInventoryMovementLogPreviewComposesTableReconciliation(t *testing.T) {
	setRuntimeRoots(t)

	tempDir := t.TempDir()
	inputFile := filepath.Join(tempDir, "inventory.xlsx")
	outputFile := filepath.Join(tempDir, "inventory-reconciled.xlsx")

	file := excelize.NewFile()
	defer func() { _ = file.Close() }()
	defaultSheet := file.GetSheetName(0)
	if err := file.SetSheetName(defaultSheet, "Movements"); err != nil {
		t.Fatalf("SetSheetName: %v", err)
	}
	movementHeader := []any{"sku", "balance"}
	if err := file.SetSheetRow("Movements", "A1", &movementHeader); err != nil {
		t.Fatalf("SetSheetRow movement header: %v", err)
	}
	for idx, row := range [][]any{{"A001", 10}, {"B002", 5}, {"C003", 2}} {
		cell, _ := excelize.CoordinatesToCellName(1, idx+2)
		if err := file.SetSheetRow("Movements", cell, &row); err != nil {
			t.Fatalf("SetSheetRow movement %d: %v", idx, err)
		}
	}
	if _, err := file.NewSheet("StockMaster"); err != nil {
		t.Fatalf("NewSheet StockMaster: %v", err)
	}
	masterHeader := []any{"sku", "on_hand"}
	if err := file.SetSheetRow("StockMaster", "A1", &masterHeader); err != nil {
		t.Fatalf("SetSheetRow master header: %v", err)
	}
	for idx, row := range [][]any{{"A001", 10}, {"B002", 7}, {"D004", 1}} {
		cell, _ := excelize.CoordinatesToCellName(1, idx+2)
		if err := file.SetSheetRow("StockMaster", cell, &row); err != nil {
			t.Fatalf("SetSheetRow master %d: %v", idx, err)
		}
	}
	if err := file.SaveAs(inputFile); err != nil {
		t.Fatalf("SaveAs: %v", err)
	}

	reconcileTask := runtimetaskspec.BuildReconcileTablesTask(runtimetaskspec.ReconcileTablesRequest{
		RequestText: "inventory movement log의 balance를 stock master on_hand와 대조한다.",
		InputFile:   inputFile,
		SourceSheet: "Movements",
		LookupSheet: "StockMaster",
		TargetSheet: "Reconciliation",
		OutputFile:  outputFile,
		LeftKey:     "sku",
		RightKey:    "sku",
		CompareMappings: []runtimetaskspec.CompareMapping{
			{LeftColumn: "balance", RightColumn: "on_hand", As: "balance"},
		},
	})
	result, err := Run(Request{ScenarioID: "inventory-preview-reconcile-tables", TaskSpec: reconcileTask.TaskSpec})
	if err != nil {
		t.Fatalf("reconcile Run: %v", err)
	}
	if !result.Verification.Pass {
		t.Fatalf("reconcile verification failed: %+v", result.Verification)
	}
	if result.Verification.SummaryRows != 4 {
		t.Fatalf("summary rows=%d want 4", result.Verification.SummaryRows)
	}
}

func TestCashFlowMonitorPreviewComposesRollForwardPeriod(t *testing.T) {
	setRuntimeRoots(t)

	tempDir := t.TempDir()
	inputFile := filepath.Join(tempDir, "cash-flow.xlsx")
	outputFile := filepath.Join(tempDir, "cash-flow-rolled.xlsx")

	file := excelize.NewFile()
	defer func() { _ = file.Close() }()
	defaultSheet := file.GetSheetName(0)
	if err := file.SetSheetName(defaultSheet, "Jan"); err != nil {
		t.Fatalf("SetSheetName(Jan): %v", err)
	}
	if _, err := file.NewSheet("Feb"); err != nil {
		t.Fatalf("NewSheet(Feb): %v", err)
	}
	for _, sheet := range []string{"Jan", "Feb"} {
		header := []any{"opening", "inflow", "outflow", "closing"}
		if err := file.SetSheetRow(sheet, "A1", &header); err != nil {
			t.Fatalf("SetSheetRow(%s header): %v", sheet, err)
		}
	}
	jan := []any{100, 75, 25}
	if err := file.SetSheetRow("Jan", "A2", &jan); err != nil {
		t.Fatalf("SetSheetRow(Jan): %v", err)
	}
	if err := file.SetCellFormula("Jan", "D2", "=A2+B2-C2"); err != nil {
		t.Fatalf("SetCellFormula(Jan D2): %v", err)
	}
	feb := []any{0, 20, 10}
	if err := file.SetSheetRow("Feb", "A2", &feb); err != nil {
		t.Fatalf("SetSheetRow(Feb): %v", err)
	}
	if err := file.SetCellFormula("Feb", "D2", "=A2+B2-C2"); err != nil {
		t.Fatalf("SetCellFormula(Feb D2): %v", err)
	}
	if err := file.SaveAs(inputFile); err != nil {
		t.Fatalf("SaveAs: %v", err)
	}

	rollTask := runtimetaskspec.BuildRollForwardPeriodTask(runtimetaskspec.RollForwardPeriodRequest{
		RequestText: "cash flow monitor에서 Jan closing balance를 Feb opening balance로 이월한다.",
		InputFile:   inputFile,
		SourceSheet: "Jan",
		TargetSheet: "Feb",
		OutputFile:  outputFile,
		CarryForwardMappings: []runtimetaskspec.CarryForwardMapping{
			{FromSheet: "Jan", FromCell: "D2", ToSheet: "Feb", ToCell: "A2"},
		},
	})
	result, err := Run(Request{ScenarioID: "cash-flow-preview-roll-forward", TaskSpec: rollTask.TaskSpec})
	if err != nil {
		t.Fatalf("roll-forward Run: %v", err)
	}
	if !result.Verification.Pass {
		t.Fatalf("roll-forward verification failed: %+v", result.Verification)
	}

	outputHandle, err := excelize.OpenFile(outputFile)
	if err != nil {
		t.Fatalf("Open output: %v", err)
	}
	defer func() { _ = outputHandle.Close() }()
	if got, err := outputHandle.GetCellValue("Feb", "A2"); err != nil || got != "150" {
		t.Fatalf("Feb!A2=%q err=%v want 150", got, err)
	}
	if got, err := outputHandle.GetCellFormula("Feb", "D2"); err != nil || got != "=A2+B2-C2" {
		t.Fatalf("Feb!D2 formula=%q err=%v want =A2+B2-C2", got, err)
	}
}
