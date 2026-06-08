package requestcompiler

import (
	"path/filepath"
	goruntime "runtime"
	"strconv"
	"testing"

	runtimeinspect "github.com/mwroh/sheet-ops/runtime/inspect"
	runtimeschema "github.com/mwroh/sheet-ops/runtime/schema"
	"github.com/xuri/excelize/v2"
)

func TestDraftOrganismExecutionRequestBuildsInvoiceLineItemStepsFromWorkbookFacts(t *testing.T) {
	tempDir := t.TempDir()
	inputFile := filepath.Join(tempDir, "invoice.xlsx")
	outputFile := filepath.Join(tempDir, "invoice-organism.xlsx")

	file := excelize.NewFile()
	defer func() { _ = file.Close() }()
	defaultSheet := file.GetSheetName(0)
	if err := file.SetSheetName(defaultSheet, "LineItems"); err != nil {
		t.Fatalf("SetSheetName: %v", err)
	}
	if err := file.SetSheetRow("LineItems", "A1", &[]any{"sku", "quantity", "unit_price", "line_total", "tax"}); err != nil {
		t.Fatalf("SetSheetRow header: %v", err)
	}
	if err := file.SetSheetRow("LineItems", "A2", &[]any{"A001", 2, 10}); err != nil {
		t.Fatalf("SetSheetRow row: %v", err)
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

	facts, err := runtimeinspect.InspectWorkbookFacts(inputFile)
	if err != nil {
		t.Fatalf("InspectWorkbookFacts: %v", err)
	}
	plan := TemplateClassPlanHintForRequest("Add invoice line items, extend totals, protect formulas, and create a printable invoice")
	if plan == nil {
		t.Fatal("missing plan")
	}

	req, ok := DraftOrganismExecutionRequest(OrganismDraftInput{
		ScenarioID:        "invoice-organism-draft",
		RequestText:       "Add invoice line items, extend totals, protect formulas, and create a printable invoice",
		InputFile:         inputFile,
		OutputFile:        outputFile,
		WorkbookFacts:     facts,
		TemplateClassPlan: *plan,
	})
	if !ok {
		t.Fatal("DraftOrganismExecutionRequest ok=false")
	}
	if req.ScenarioID != "invoice-organism-draft" {
		t.Fatalf("scenario_id=%q want invoice-organism-draft", req.ScenarioID)
	}
	if len(req.Steps) != 5 {
		t.Fatalf("steps=%d want 5: %+v", len(req.Steps), req.Steps)
	}
	if req.Steps[0].AtomID != "append_structured_rows" || req.Steps[0].SourceSheet != "LineItems" {
		t.Fatalf("append step=%+v", req.Steps[0])
	}
	if len(req.Steps[0].Values) != 3 {
		t.Fatalf("append values=%v want 3 default values", req.Steps[0].Values)
	}
	if req.Steps[1].AtomID != "extend_table_formulas" || req.Steps[1].FormulaSourceRow != 2 || len(req.Steps[1].FormulaColumns) != 2 {
		t.Fatalf("formula step=%+v", req.Steps[1])
	}
	if req.Steps[3].ProtectionRule == nil || len(req.Steps[3].ProtectionRule.FormulaRanges) != 1 {
		t.Fatalf("protection step=%+v", req.Steps[3])
	}
	if req.Steps[4].TargetSheet != "InvoicePrint" || req.Steps[4].TableBinding == nil {
		t.Fatalf("printable step=%+v", req.Steps[4])
	}
	if err := runtimeschema.ValidateStruct(organismExecutionRequestSchemaPathForTest(), req); err != nil {
		t.Fatalf("draft organism request schema validation: %v", err)
	}
}

func TestDraftOrganismExecutionRequestBuildsExpenseReimbursementStepsFromWorkbookFacts(t *testing.T) {
	tempDir := t.TempDir()
	inputFile := filepath.Join(tempDir, "expense.xlsx")
	outputFile := filepath.Join(tempDir, "expense-organism.xlsx")

	file := excelize.NewFile()
	defer func() { _ = file.Close() }()
	defaultSheet := file.GetSheetName(0)
	if err := file.SetSheetName(defaultSheet, "ExpenseItems"); err != nil {
		t.Fatalf("SetSheetName: %v", err)
	}
	if err := file.SetSheetRow("ExpenseItems", "A1", &[]any{"item", "amount", "reimbursable_rate", "receipt_status", "reimbursable_total"}); err != nil {
		t.Fatalf("SetSheetRow header: %v", err)
	}
	if err := file.SetSheetRow("ExpenseItems", "A2", &[]any{"Hotel", 200, 1, "attached"}); err != nil {
		t.Fatalf("SetSheetRow row: %v", err)
	}
	if err := file.SetCellFormula("ExpenseItems", "E2", "=B2*C2"); err != nil {
		t.Fatalf("SetCellFormula(E2): %v", err)
	}
	if err := file.SaveAs(inputFile); err != nil {
		t.Fatalf("SaveAs: %v", err)
	}

	facts, err := runtimeinspect.InspectWorkbookFacts(inputFile)
	if err != nil {
		t.Fatalf("InspectWorkbookFacts: %v", err)
	}
	requestText := "Append expense reimbursement items, validate receipt status, extend reimbursable totals, protect formulas, and generate a printable claim"
	plan := TemplateClassPlanHintForRequest(requestText)
	if plan == nil {
		t.Fatal("missing plan")
	}

	req, ok := DraftOrganismExecutionRequest(OrganismDraftInput{
		ScenarioID:        "expense-organism-draft",
		RequestText:       requestText,
		InputFile:         inputFile,
		OutputFile:        outputFile,
		WorkbookFacts:     facts,
		TemplateClassPlan: *plan,
	})
	if !ok {
		t.Fatal("DraftOrganismExecutionRequest ok=false")
	}
	if len(req.Steps) != 5 {
		t.Fatalf("steps=%d want 5", len(req.Steps))
	}
	if req.Steps[0].AtomID != "append_structured_rows" || len(req.Steps[0].Values) != 4 {
		t.Fatalf("append step=%+v", req.Steps[0])
	}
	if req.Steps[1].AtomID != "extend_table_formulas" || req.Steps[1].FormulaSourceRow != 2 || req.Steps[1].TargetRows[0] != 3 {
		t.Fatalf("formula step=%+v", req.Steps[1])
	}
	if req.Steps[2].AtomID != "add_data_validation" || req.Steps[2].ValidationRule == nil || req.Steps[2].ValidationRule.Ranges[0] != "D2:D20" {
		t.Fatalf("validation step=%+v", req.Steps[2])
	}
	if req.Steps[3].AtomID != "protect_formula_cells" || req.Steps[3].ProtectionRule == nil || req.Steps[3].ProtectionRule.FormulaRanges[0] != "E2:E3" {
		t.Fatalf("protection step=%+v", req.Steps[3])
	}
	if req.Steps[4].AtomID != "generate_printable_form" || req.Steps[4].TargetSheet != "ExpenseClaim" || req.Steps[4].TableBinding == nil {
		t.Fatalf("printable step=%+v", req.Steps[4])
	}
	if err := runtimeschema.ValidateStruct(organismExecutionRequestSchemaPathForTest(), req); err != nil {
		t.Fatalf("draft organism request schema validation: %v", err)
	}
}

func TestDraftOrganismExecutionRequestBuildsPurchaseOrderStepsFromWorkbookFacts(t *testing.T) {
	tempDir := t.TempDir()
	inputFile := filepath.Join(tempDir, "purchase-order.xlsx")
	outputFile := filepath.Join(tempDir, "purchase-order-organism.xlsx")

	file := excelize.NewFile()
	defer func() { _ = file.Close() }()
	defaultSheet := file.GetSheetName(0)
	if err := file.SetSheetName(defaultSheet, "PurchaseOrder"); err != nil {
		t.Fatalf("SetSheetName: %v", err)
	}
	if err := file.SetSheetRow("PurchaseOrder", "A1", &[]any{"item", "quantity", "unit_price", "po_status", "line_total"}); err != nil {
		t.Fatalf("SetSheetRow header: %v", err)
	}
	if err := file.SetSheetRow("PurchaseOrder", "A2", &[]any{"Keyboard", 2, 50, "draft"}); err != nil {
		t.Fatalf("SetSheetRow row: %v", err)
	}
	if err := file.SetCellFormula("PurchaseOrder", "E2", "=B2*C2"); err != nil {
		t.Fatalf("SetCellFormula(E2): %v", err)
	}
	if err := file.SaveAs(inputFile); err != nil {
		t.Fatalf("SaveAs: %v", err)
	}

	facts, err := runtimeinspect.InspectWorkbookFacts(inputFile)
	if err != nil {
		t.Fatalf("InspectWorkbookFacts: %v", err)
	}
	requestText := "Append purchase order line items, extend order totals, validate PO status, protect formulas, and generate a printable purchase order"
	plan := TemplateClassPlanHintForRequest(requestText)
	if plan == nil {
		t.Fatal("missing plan")
	}

	req, ok := DraftOrganismExecutionRequest(OrganismDraftInput{
		ScenarioID:        "purchase-order-organism-draft",
		RequestText:       requestText,
		InputFile:         inputFile,
		OutputFile:        outputFile,
		WorkbookFacts:     facts,
		TemplateClassPlan: *plan,
	})
	if !ok {
		t.Fatal("DraftOrganismExecutionRequest ok=false")
	}
	if len(req.Steps) != 5 {
		t.Fatalf("steps=%d want 5", len(req.Steps))
	}
	if req.Steps[0].AtomID != "append_structured_rows" || len(req.Steps[0].Values) != 4 {
		t.Fatalf("append step=%+v", req.Steps[0])
	}
	if req.Steps[1].AtomID != "extend_table_formulas" || req.Steps[1].FormulaSourceRow != 2 || req.Steps[1].TargetRows[0] != 3 {
		t.Fatalf("formula step=%+v", req.Steps[1])
	}
	if req.Steps[2].AtomID != "add_data_validation" || req.Steps[2].ValidationRule == nil || req.Steps[2].ValidationRule.Ranges[0] != "D2:D20" {
		t.Fatalf("validation step=%+v", req.Steps[2])
	}
	if req.Steps[3].AtomID != "protect_formula_cells" || req.Steps[3].ProtectionRule == nil || req.Steps[3].ProtectionRule.FormulaRanges[0] != "E2:E3" {
		t.Fatalf("protection step=%+v", req.Steps[3])
	}
	if req.Steps[4].AtomID != "generate_printable_form" || req.Steps[4].TargetSheet != "PurchaseOrderPrint" || req.Steps[4].FormTitle != "Purchase Order" || req.Steps[4].TableBinding == nil {
		t.Fatalf("printable step=%+v", req.Steps[4])
	}
	if err := runtimeschema.ValidateStruct(organismExecutionRequestSchemaPathForTest(), req); err != nil {
		t.Fatalf("draft organism request schema validation: %v", err)
	}
}

func TestDraftOrganismExecutionRequestBuildsMonthlyBudgetStepsFromWorkbookFacts(t *testing.T) {
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
		if err := file.SetCellFormula("Budget", "E"+cellRowForDraftTest(rowNum), "=B"+cellRowForDraftTest(rowNum)+"+C"+cellRowForDraftTest(rowNum)); err != nil {
			t.Fatalf("SetCellFormula review_total row %d: %v", rowNum, err)
		}
		if err := file.SetCellFormula("Budget", "F"+cellRowForDraftTest(rowNum), "=B"+cellRowForDraftTest(rowNum)+"-C"+cellRowForDraftTest(rowNum)); err != nil {
			t.Fatalf("SetCellFormula closing row %d: %v", rowNum, err)
		}
	}
	if err := file.SaveAs(inputFile); err != nil {
		t.Fatalf("SaveAs: %v", err)
	}

	facts, err := runtimeinspect.InspectWorkbookFacts(inputFile)
	if err != nil {
		t.Fatalf("InspectWorkbookFacts: %v", err)
	}
	requestText := "Summarize monthly budget actuals, flag variance, copy period, roll forward closing, and protect formulas"
	plan := TemplateClassPlanHintForRequest(requestText)
	if plan == nil {
		t.Fatal("missing plan")
	}

	req, ok := DraftOrganismExecutionRequest(OrganismDraftInput{
		ScenarioID:        "budget-organism-draft",
		RequestText:       requestText,
		InputFile:         inputFile,
		OutputFile:        outputFile,
		WorkbookFacts:     facts,
		TemplateClassPlan: *plan,
	})
	if !ok {
		t.Fatal("DraftOrganismExecutionRequest ok=false")
	}
	if req.Steps[0].AtomID != "group_summarize" || req.Steps[0].TargetSheet != "BudgetSummary" {
		t.Fatalf("summary step=%+v", req.Steps[0])
	}
	if req.Steps[0].SummaryMode != "values" || len(req.Steps[0].GroupBy) != 1 || req.Steps[0].GroupBy[0] != "category" {
		t.Fatalf("summary grouping=%+v", req.Steps[0])
	}
	if len(req.Steps[0].Metrics) != 1 || req.Steps[0].Metrics[0].Column != "actual" || req.Steps[0].Metrics[0].As != "actual_total" {
		t.Fatalf("summary metrics=%+v", req.Steps[0].Metrics)
	}
	if req.Steps[1].AtomID != "highlight_threshold" || req.Steps[1].TargetColumn != "variance" || req.Steps[1].Threshold == nil || *req.Steps[1].Threshold != 0 {
		t.Fatalf("threshold step=%+v", req.Steps[1])
	}
	if req.Steps[2].AtomID != "copy_period_sheet" || req.Steps[2].TargetSheet != "NextBudget" {
		t.Fatalf("copy step=%+v", req.Steps[2])
	}
	if req.Steps[3].AtomID != "roll_forward_period" || len(req.Steps[3].CarryForwardMappings) != 1 {
		t.Fatalf("roll-forward step=%+v", req.Steps[3])
	}
	if req.Steps[4].AtomID != "protect_formula_cells" || req.Steps[4].ProtectionRule == nil || req.Steps[4].ProtectionRule.FormulaRanges[0] != "E2:F3" {
		t.Fatalf("protection step=%+v", req.Steps[4])
	}
	if err := runtimeschema.ValidateStruct(organismExecutionRequestSchemaPathForTest(), req); err != nil {
		t.Fatalf("draft organism request schema validation: %v", err)
	}
}

func TestDraftOrganismExecutionRequestBuildsCashFlowMonitorStepsFromWorkbookFacts(t *testing.T) {
	tempDir := t.TempDir()
	inputFile := filepath.Join(tempDir, "cash-flow.xlsx")
	outputFile := filepath.Join(tempDir, "cash-flow-organism.xlsx")

	file := excelize.NewFile()
	defer func() { _ = file.Close() }()
	defaultSheet := file.GetSheetName(0)
	if err := file.SetSheetName(defaultSheet, "Jan"); err != nil {
		t.Fatalf("SetSheetName: %v", err)
	}
	if err := file.SetSheetRow("Jan", "A1", &[]any{"period", "opening", "inflow", "outflow", "closing"}); err != nil {
		t.Fatalf("SetSheetRow header: %v", err)
	}
	for idx, row := range [][]any{{"Jan", 100, 75, 25}, {"Jan", 150, 40, 30}} {
		cell, _ := excelize.CoordinatesToCellName(1, idx+2)
		if err := file.SetSheetRow("Jan", cell, &row); err != nil {
			t.Fatalf("SetSheetRow cash %d: %v", idx, err)
		}
		rowNum := idx + 2
		if err := file.SetCellFormula("Jan", "E"+cellRowForDraftTest(rowNum), "=B"+cellRowForDraftTest(rowNum)+"+C"+cellRowForDraftTest(rowNum)+"-D"+cellRowForDraftTest(rowNum)); err != nil {
			t.Fatalf("SetCellFormula closing row %d: %v", rowNum, err)
		}
	}
	if err := file.SetSheetRow("Jan", "A4", &[]any{"Feb", 0, 20, 10}); err != nil {
		t.Fatalf("SetSheetRow target row: %v", err)
	}
	if err := file.SaveAs(inputFile); err != nil {
		t.Fatalf("SaveAs: %v", err)
	}

	facts, err := runtimeinspect.InspectWorkbookFacts(inputFile)
	if err != nil {
		t.Fatalf("InspectWorkbookFacts: %v", err)
	}
	requestText := "Track cash flow opening balance, summarize inflows, extend closing formulas, copy period, roll forward closing balance, and protect formulas"
	plan := TemplateClassPlanHintForRequest(requestText)
	if plan == nil {
		t.Fatal("missing plan")
	}

	req, ok := DraftOrganismExecutionRequest(OrganismDraftInput{
		ScenarioID:        "cash-flow-organism-draft",
		RequestText:       requestText,
		InputFile:         inputFile,
		OutputFile:        outputFile,
		WorkbookFacts:     facts,
		TemplateClassPlan: *plan,
	})
	if !ok {
		t.Fatal("DraftOrganismExecutionRequest ok=false")
	}
	if len(req.Steps) != 5 {
		t.Fatalf("steps=%d want 5", len(req.Steps))
	}
	if req.Steps[0].AtomID != "group_summarize" || req.Steps[0].TargetSheet != "CashFlowSummary" {
		t.Fatalf("summary step=%+v", req.Steps[0])
	}
	if len(req.Steps[0].Metrics) != 1 || req.Steps[0].Metrics[0].Column != "inflow" || req.Steps[0].Metrics[0].As != "inflow_total" {
		t.Fatalf("summary metrics=%+v", req.Steps[0].Metrics)
	}
	if req.Steps[1].AtomID != "extend_table_formulas" || req.Steps[1].FormulaSourceRow != 2 || req.Steps[1].TargetRows[0] != 4 {
		t.Fatalf("formula step=%+v", req.Steps[1])
	}
	if req.Steps[2].AtomID != "copy_period_sheet" || req.Steps[2].TargetSheet != "NextCashFlow" {
		t.Fatalf("copy step=%+v", req.Steps[2])
	}
	if req.Steps[3].AtomID != "roll_forward_period" || req.Steps[3].CarryForwardMappings[0].FromCell != "E2" || req.Steps[3].CarryForwardMappings[0].ToCell != "B3" {
		t.Fatalf("roll-forward step=%+v", req.Steps[3])
	}
	if req.Steps[4].AtomID != "protect_formula_cells" || req.Steps[4].ProtectionRule == nil || req.Steps[4].ProtectionRule.FormulaRanges[0] != "E2:E4" {
		t.Fatalf("protection step=%+v", req.Steps[4])
	}
	if err := runtimeschema.ValidateStruct(organismExecutionRequestSchemaPathForTest(), req); err != nil {
		t.Fatalf("draft organism request schema validation: %v", err)
	}
}

func TestDraftOrganismExecutionRequestBuildsAttendanceRegisterStepsFromWorkbookFacts(t *testing.T) {
	tempDir := t.TempDir()
	inputFile := filepath.Join(tempDir, "attendance.xlsx")
	outputFile := filepath.Join(tempDir, "attendance-organism.xlsx")

	file := excelize.NewFile()
	defer func() { _ = file.Close() }()
	defaultSheet := file.GetSheetName(0)
	if err := file.SetSheetName(defaultSheet, "Attendance"); err != nil {
		t.Fatalf("SetSheetName: %v", err)
	}
	if err := file.SetSheetRow("Attendance", "A1", &[]any{"student", "date", "status", "attendance_total"}); err != nil {
		t.Fatalf("SetSheetRow header: %v", err)
	}
	if err := file.SetSheetRow("Attendance", "A2", &[]any{"Alice", "2026-06-08", "present"}); err != nil {
		t.Fatalf("SetSheetRow row2: %v", err)
	}
	if err := file.SetSheetRow("Attendance", "A3", &[]any{"Bob", "2026-06-08", "absent"}); err != nil {
		t.Fatalf("SetSheetRow row3: %v", err)
	}
	if err := file.SetCellFormula("Attendance", "D2", "=IF(C2=\"present\",1,0)"); err != nil {
		t.Fatalf("SetCellFormula(D2): %v", err)
	}
	if err := file.SetCellFormula("Attendance", "D3", "=IF(C3=\"present\",1,0)"); err != nil {
		t.Fatalf("SetCellFormula(D3): %v", err)
	}
	if err := file.SaveAs(inputFile); err != nil {
		t.Fatalf("SaveAs: %v", err)
	}

	facts, err := runtimeinspect.InspectWorkbookFacts(inputFile)
	if err != nil {
		t.Fatalf("InspectWorkbookFacts: %v", err)
	}
	requestText := "Copy attendance register to the next class period, validate attendance status, and protect total formulas"
	plan := TemplateClassPlanHintForRequest(requestText)
	if plan == nil {
		t.Fatal("missing plan")
	}

	req, ok := DraftOrganismExecutionRequest(OrganismDraftInput{
		ScenarioID:        "attendance-organism-draft",
		RequestText:       requestText,
		InputFile:         inputFile,
		OutputFile:        outputFile,
		WorkbookFacts:     facts,
		TemplateClassPlan: *plan,
	})
	if !ok {
		t.Fatal("DraftOrganismExecutionRequest ok=false")
	}
	if len(req.Steps) != 3 {
		t.Fatalf("steps=%d want 3", len(req.Steps))
	}
	if req.Steps[0].AtomID != "copy_period_sheet" || req.Steps[0].SourceSheet != "Attendance" || req.Steps[0].TargetSheet != "NextAttendance" {
		t.Fatalf("copy step=%+v", req.Steps[0])
	}
	if req.Steps[1].AtomID != "add_data_validation" || req.Steps[1].SourceSheet != "NextAttendance" || req.Steps[1].ValidationRule == nil || req.Steps[1].ValidationRule.Ranges[0] != "C2:C20" {
		t.Fatalf("validation step=%+v", req.Steps[1])
	}
	if req.Steps[2].AtomID != "protect_formula_cells" || req.Steps[2].SourceSheet != "NextAttendance" || req.Steps[2].ProtectionRule == nil || req.Steps[2].ProtectionRule.FormulaRanges[0] != "D2:D3" {
		t.Fatalf("protection step=%+v", req.Steps[2])
	}
	if err := runtimeschema.ValidateStruct(organismExecutionRequestSchemaPathForTest(), req); err != nil {
		t.Fatalf("draft organism request schema validation: %v", err)
	}
}

func TestDraftOrganismExecutionRequestBuildsProjectTimelineStepsFromWorkbookFacts(t *testing.T) {
	tempDir := t.TempDir()
	inputFile := filepath.Join(tempDir, "timeline.xlsx")
	outputFile := filepath.Join(tempDir, "timeline-organism.xlsx")

	file := excelize.NewFile()
	defer func() { _ = file.Close() }()
	defaultSheet := file.GetSheetName(0)
	if err := file.SetSheetName(defaultSheet, "Sprint1"); err != nil {
		t.Fatalf("SetSheetName: %v", err)
	}
	if err := file.SetSheetRow("Sprint1", "A1", &[]any{"task", "start", "end", "status", "task_count", "progress_pct"}); err != nil {
		t.Fatalf("SetSheetRow header: %v", err)
	}
	if err := file.SetSheetRow("Sprint1", "A2", &[]any{"Design", "2026-06-01", "2026-06-05", "todo", 1}); err != nil {
		t.Fatalf("SetSheetRow row2: %v", err)
	}
	if err := file.SetSheetRow("Sprint1", "A3", &[]any{"Build", "2026-06-06", "2026-06-12", "in_progress", 1}); err != nil {
		t.Fatalf("SetSheetRow row3: %v", err)
	}
	if err := file.SetSheetRow("Sprint1", "A4", &[]any{"Test", "2026-06-13", "2026-06-15", "done", 1}); err != nil {
		t.Fatalf("SetSheetRow row4: %v", err)
	}
	if err := file.SetCellFormula("Sprint1", "F2", "=IF(D2=\"done\",1,0)"); err != nil {
		t.Fatalf("SetCellFormula(F2): %v", err)
	}
	if err := file.SetCellFormula("Sprint1", "F3", "=IF(D3=\"done\",1,0)"); err != nil {
		t.Fatalf("SetCellFormula(F3): %v", err)
	}
	if err := file.SaveAs(inputFile); err != nil {
		t.Fatalf("SaveAs: %v", err)
	}

	facts, err := runtimeinspect.InspectWorkbookFacts(inputFile)
	if err != nil {
		t.Fatalf("InspectWorkbookFacts: %v", err)
	}
	requestText := "Copy project timeline period, extend task progress formulas, validate task status, summarize timeline status, and protect formulas"
	plan := TemplateClassPlanHintForRequest(requestText)
	if plan == nil {
		t.Fatal("missing plan")
	}

	req, ok := DraftOrganismExecutionRequest(OrganismDraftInput{
		ScenarioID:        "project-timeline-organism-draft",
		RequestText:       requestText,
		InputFile:         inputFile,
		OutputFile:        outputFile,
		WorkbookFacts:     facts,
		TemplateClassPlan: *plan,
	})
	if !ok {
		t.Fatal("DraftOrganismExecutionRequest ok=false")
	}
	if len(req.Steps) != 5 {
		t.Fatalf("steps=%d want 5", len(req.Steps))
	}
	if req.Steps[0].AtomID != "copy_period_sheet" || req.Steps[0].SourceSheet != "Sprint1" || req.Steps[0].TargetSheet != "Sprint2" {
		t.Fatalf("copy step=%+v", req.Steps[0])
	}
	if req.Steps[1].AtomID != "extend_table_formulas" || req.Steps[1].SourceSheet != "Sprint2" || req.Steps[1].FormulaSourceRow != 2 || req.Steps[1].TargetRows[0] != 4 {
		t.Fatalf("formula step=%+v", req.Steps[1])
	}
	if req.Steps[2].AtomID != "add_data_validation" || req.Steps[2].SourceSheet != "Sprint2" || req.Steps[2].ValidationRule == nil || req.Steps[2].ValidationRule.Ranges[0] != "D2:D20" {
		t.Fatalf("validation step=%+v", req.Steps[2])
	}
	if req.Steps[3].AtomID != "group_summarize" || req.Steps[3].SourceSheet != "Sprint2" || req.Steps[3].TargetSheet != "TimelineSummary" {
		t.Fatalf("summary step=%+v", req.Steps[3])
	}
	if req.Steps[4].AtomID != "protect_formula_cells" || req.Steps[4].SourceSheet != "Sprint2" || req.Steps[4].ProtectionRule == nil || req.Steps[4].ProtectionRule.FormulaRanges[0] != "F2:F4" {
		t.Fatalf("protection step=%+v", req.Steps[4])
	}
	if err := runtimeschema.ValidateStruct(organismExecutionRequestSchemaPathForTest(), req); err != nil {
		t.Fatalf("draft organism request schema validation: %v", err)
	}
}

func TestDraftOrganismExecutionRequestBuildsShiftRosterStepsFromWorkbookFacts(t *testing.T) {
	tempDir := t.TempDir()
	inputFile := filepath.Join(tempDir, "roster.xlsx")
	outputFile := filepath.Join(tempDir, "roster-organism.xlsx")

	file := excelize.NewFile()
	defer func() { _ = file.Close() }()
	defaultSheet := file.GetSheetName(0)
	if err := file.SetSheetName(defaultSheet, "Week1"); err != nil {
		t.Fatalf("SetSheetName: %v", err)
	}
	if err := file.SetSheetRow("Week1", "A1", &[]any{"employee", "date", "shift", "coverage_total"}); err != nil {
		t.Fatalf("SetSheetRow header: %v", err)
	}
	if err := file.SetSheetRow("Week1", "A2", &[]any{"Alex", "2026-06-01", "AM"}); err != nil {
		t.Fatalf("SetSheetRow row2: %v", err)
	}
	if err := file.SetSheetRow("Week1", "A3", &[]any{"Blair", "2026-06-01", "PM"}); err != nil {
		t.Fatalf("SetSheetRow row3: %v", err)
	}
	if err := file.SetCellFormula("Week1", "D2", "=IF(C2=\"OFF\",0,1)"); err != nil {
		t.Fatalf("SetCellFormula(D2): %v", err)
	}
	if err := file.SetCellFormula("Week1", "D3", "=IF(C3=\"OFF\",0,1)"); err != nil {
		t.Fatalf("SetCellFormula(D3): %v", err)
	}
	if err := file.SaveAs(inputFile); err != nil {
		t.Fatalf("SaveAs: %v", err)
	}

	facts, err := runtimeinspect.InspectWorkbookFacts(inputFile)
	if err != nil {
		t.Fatalf("InspectWorkbookFacts: %v", err)
	}
	requestText := "Copy weekly shift roster, validate shift codes, and protect coverage formulas"
	plan := TemplateClassPlanHintForRequest(requestText)
	if plan == nil {
		t.Fatal("missing plan")
	}

	req, ok := DraftOrganismExecutionRequest(OrganismDraftInput{
		ScenarioID:        "shift-roster-organism-draft",
		RequestText:       requestText,
		InputFile:         inputFile,
		OutputFile:        outputFile,
		WorkbookFacts:     facts,
		TemplateClassPlan: *plan,
	})
	if !ok {
		t.Fatal("DraftOrganismExecutionRequest ok=false")
	}
	if len(req.Steps) != 3 {
		t.Fatalf("steps=%d want 3", len(req.Steps))
	}
	if req.Steps[0].AtomID != "copy_period_sheet" || req.Steps[0].SourceSheet != "Week1" || req.Steps[0].TargetSheet != "Week2" {
		t.Fatalf("copy step=%+v", req.Steps[0])
	}
	if req.Steps[1].AtomID != "add_data_validation" || req.Steps[1].SourceSheet != "Week2" || req.Steps[1].ValidationRule == nil || req.Steps[1].ValidationRule.Ranges[0] != "C2:C20" {
		t.Fatalf("validation step=%+v", req.Steps[1])
	}
	if req.Steps[2].AtomID != "protect_formula_cells" || req.Steps[2].SourceSheet != "Week2" || req.Steps[2].ProtectionRule == nil || req.Steps[2].ProtectionRule.FormulaRanges[0] != "D2:D3" {
		t.Fatalf("protection step=%+v", req.Steps[2])
	}
	if err := runtimeschema.ValidateStruct(organismExecutionRequestSchemaPathForTest(), req); err != nil {
		t.Fatalf("draft organism request schema validation: %v", err)
	}
}

func TestDraftOrganismExecutionRequestBuildsConstructionCostStepsFromWorkbookFacts(t *testing.T) {
	tempDir := t.TempDir()
	inputFile := filepath.Join(tempDir, "construction-cost.xlsx")
	outputFile := filepath.Join(tempDir, "construction-cost-organism.xlsx")

	file := excelize.NewFile()
	defer func() { _ = file.Close() }()
	defaultSheet := file.GetSheetName(0)
	if err := file.SetSheetName(defaultSheet, "Costs"); err != nil {
		t.Fatalf("SetSheetName: %v", err)
	}
	if err := file.SetSheetRow("Costs", "A1", &[]any{"cost_code", "phase", "actual", "budget", "variance"}); err != nil {
		t.Fatalf("SetSheetRow header: %v", err)
	}
	if err := file.SetSheetRow("Costs", "A2", &[]any{"LABOR", "framing", 90, 100}); err != nil {
		t.Fatalf("SetSheetRow row: %v", err)
	}
	if err := file.SetCellFormula("Costs", "E2", "=C2-D2"); err != nil {
		t.Fatalf("SetCellFormula(E2): %v", err)
	}
	if err := file.SaveAs(inputFile); err != nil {
		t.Fatalf("SaveAs: %v", err)
	}

	facts, err := runtimeinspect.InspectWorkbookFacts(inputFile)
	if err != nil {
		t.Fatalf("InspectWorkbookFacts: %v", err)
	}
	requestText := "Append construction cost row, extend variance formulas, highlight budget overrun, and protect forecast formulas"
	plan := TemplateClassPlanHintForRequest(requestText)
	if plan == nil {
		t.Fatal("missing plan")
	}

	req, ok := DraftOrganismExecutionRequest(OrganismDraftInput{
		ScenarioID:        "construction-cost-organism-draft",
		RequestText:       requestText,
		InputFile:         inputFile,
		OutputFile:        outputFile,
		WorkbookFacts:     facts,
		TemplateClassPlan: *plan,
	})
	if !ok {
		t.Fatal("DraftOrganismExecutionRequest ok=false")
	}
	if len(req.Steps) != 4 {
		t.Fatalf("steps=%d want 4", len(req.Steps))
	}
	if req.Steps[0].AtomID != "append_structured_rows" || req.Steps[0].SourceSheet != "Costs" || len(req.Steps[0].Values) != 4 {
		t.Fatalf("append step=%+v", req.Steps[0])
	}
	if req.Steps[1].AtomID != "extend_table_formulas" || req.Steps[1].FormulaSourceRow != 2 || req.Steps[1].TargetRows[0] != 3 || req.Steps[1].FormulaColumns[0] != "E" {
		t.Fatalf("formula step=%+v", req.Steps[1])
	}
	if req.Steps[2].AtomID != "highlight_threshold" || req.Steps[2].TargetColumn != "actual" || req.Steps[2].Threshold == nil || *req.Steps[2].Threshold != 100 {
		t.Fatalf("threshold step=%+v", req.Steps[2])
	}
	if req.Steps[3].AtomID != "protect_formula_cells" || req.Steps[3].ProtectionRule == nil || req.Steps[3].ProtectionRule.FormulaRanges[0] != "E2:E3" {
		t.Fatalf("protection step=%+v", req.Steps[3])
	}
	if err := runtimeschema.ValidateStruct(organismExecutionRequestSchemaPathForTest(), req); err != nil {
		t.Fatalf("draft organism request schema validation: %v", err)
	}
}

func TestDraftOrganismExecutionRequestBuildsTimesheetHoursLogStepsFromWorkbookFacts(t *testing.T) {
	tempDir := t.TempDir()
	inputFile := filepath.Join(tempDir, "timesheet.xlsx")
	outputFile := filepath.Join(tempDir, "timesheet-organism.xlsx")

	file := excelize.NewFile()
	defer func() { _ = file.Close() }()
	defaultSheet := file.GetSheetName(0)
	if err := file.SetSheetName(defaultSheet, "Week1"); err != nil {
		t.Fatalf("SetSheetName: %v", err)
	}
	if err := file.SetSheetRow("Week1", "A1", &[]any{"date", "employee", "work_code", "hours", "rate", "pay"}); err != nil {
		t.Fatalf("SetSheetRow header: %v", err)
	}
	if err := file.SetSheetRow("Week1", "A2", &[]any{"2026-06-01", "Ada", "DEV", 8, 25}); err != nil {
		t.Fatalf("SetSheetRow row: %v", err)
	}
	if err := file.SetCellFormula("Week1", "F2", "=D2*E2"); err != nil {
		t.Fatalf("SetCellFormula(F2): %v", err)
	}
	if err := file.SaveAs(inputFile); err != nil {
		t.Fatalf("SaveAs: %v", err)
	}

	facts, err := runtimeinspect.InspectWorkbookFacts(inputFile)
	if err != nil {
		t.Fatalf("InspectWorkbookFacts: %v", err)
	}
	requestText := "Copy weekly timesheet, append employee hours, validate work code, extend pay formulas, and protect totals"
	plan := TemplateClassPlanHintForRequest(requestText)
	if plan == nil {
		t.Fatal("missing plan")
	}

	req, ok := DraftOrganismExecutionRequest(OrganismDraftInput{
		ScenarioID:        "timesheet-organism-draft",
		RequestText:       requestText,
		InputFile:         inputFile,
		OutputFile:        outputFile,
		WorkbookFacts:     facts,
		TemplateClassPlan: *plan,
	})
	if !ok {
		t.Fatal("DraftOrganismExecutionRequest ok=false")
	}
	if len(req.Steps) != 5 {
		t.Fatalf("steps=%d want 5", len(req.Steps))
	}
	if req.Steps[0].AtomID != "copy_period_sheet" || req.Steps[0].TargetSheet != "Week2" {
		t.Fatalf("copy step=%+v", req.Steps[0])
	}
	if req.Steps[1].AtomID != "append_structured_rows" || req.Steps[1].SourceSheet != "Week2" || len(req.Steps[1].Values) != 5 {
		t.Fatalf("append step=%+v", req.Steps[1])
	}
	if req.Steps[2].AtomID != "extend_table_formulas" || req.Steps[2].FormulaSourceRow != 2 || req.Steps[2].TargetRows[0] != 3 {
		t.Fatalf("formula step=%+v", req.Steps[2])
	}
	if req.Steps[3].AtomID != "add_data_validation" || req.Steps[3].ValidationRule == nil || req.Steps[3].ValidationRule.Ranges[0] != "C2:C20" {
		t.Fatalf("validation step=%+v", req.Steps[3])
	}
	if req.Steps[4].AtomID != "protect_formula_cells" || req.Steps[4].ProtectionRule == nil || req.Steps[4].ProtectionRule.FormulaRanges[0] != "F2:F3" {
		t.Fatalf("protection step=%+v", req.Steps[4])
	}
	if err := runtimeschema.ValidateStruct(organismExecutionRequestSchemaPathForTest(), req); err != nil {
		t.Fatalf("draft organism request schema validation: %v", err)
	}
}

func TestDraftOrganismExecutionRequestBuildsInventoryMovementStepsFromWorkbookFacts(t *testing.T) {
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
		if err := file.SetCellFormula("Movements", "D"+cellRowForDraftTest(rowNum), "=B"+cellRowForDraftTest(rowNum)+"-C"+cellRowForDraftTest(rowNum)); err != nil {
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

	facts, err := runtimeinspect.InspectWorkbookFacts(inputFile)
	if err != nil {
		t.Fatalf("InspectWorkbookFacts: %v", err)
	}
	requestText := "Normalize inventory movement headers, append SKU movement, lookup SKU metadata, protect balance formulas, and reconcile stock master"
	plan := TemplateClassPlanHintForRequest(requestText)
	if plan == nil {
		t.Fatal("missing plan")
	}

	req, ok := DraftOrganismExecutionRequest(OrganismDraftInput{
		ScenarioID:        "inventory-organism-draft",
		RequestText:       requestText,
		InputFile:         inputFile,
		OutputFile:        outputFile,
		WorkbookFacts:     facts,
		TemplateClassPlan: *plan,
	})
	if !ok {
		t.Fatal("DraftOrganismExecutionRequest ok=false")
	}
	if req.Steps[0].AtomID != "normalize_headers" || len(req.Steps[0].HeaderMappings) != 4 {
		t.Fatalf("normalize step=%+v", req.Steps[0])
	}
	if req.Steps[1].AtomID != "append_structured_rows" || len(req.Steps[1].Values) != 4 {
		t.Fatalf("append step=%+v", req.Steps[1])
	}
	if req.Steps[2].AtomID != "join_lookup" || req.Steps[2].LookupSheet != "SKU" || req.Steps[2].TargetSheet != "MovementsEnriched" {
		t.Fatalf("lookup step=%+v", req.Steps[2])
	}
	if req.Steps[3].AtomID != "protect_formula_cells" || req.Steps[3].ProtectionRule == nil || req.Steps[3].ProtectionRule.FormulaRanges[0] != "D2:D3" {
		t.Fatalf("protection step=%+v", req.Steps[3])
	}
	if req.Steps[4].AtomID != "reconcile_tables" || req.Steps[4].TargetSheet != "InventoryReconciliation" || len(req.Steps[4].CompareMappings) != 1 {
		t.Fatalf("reconcile step=%+v", req.Steps[4])
	}
	if err := runtimeschema.ValidateStruct(organismExecutionRequestSchemaPathForTest(), req); err != nil {
		t.Fatalf("draft organism request schema validation: %v", err)
	}
}

func TestDraftOrganismExecutionRequestBuildsProcurementReconciliationStepsFromWorkbookFacts(t *testing.T) {
	tempDir := t.TempDir()
	inputFile := filepath.Join(tempDir, "procurement.xlsx")
	outputFile := filepath.Join(tempDir, "procurement-organism.xlsx")

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
	if err := file.SetSheetRow("Invoice", "A1", &[]any{"po_id", "invoice_amount", "status", "review_flag"}); err != nil {
		t.Fatalf("SetSheetRow invoice header: %v", err)
	}
	if err := file.SetSheetRow("Invoice", "A2", &[]any{"PO-1", 125, "received"}); err != nil {
		t.Fatalf("SetSheetRow invoice row: %v", err)
	}
	if err := file.SetCellFormula("Invoice", "D2", "=IF(B2>0,\"review\",\"missing\")"); err != nil {
		t.Fatalf("SetCellFormula(D2): %v", err)
	}
	if err := file.SaveAs(inputFile); err != nil {
		t.Fatalf("SaveAs: %v", err)
	}

	facts, err := runtimeinspect.InspectWorkbookFacts(inputFile)
	if err != nil {
		t.Fatalf("InspectWorkbookFacts: %v", err)
	}
	requestText := "Validate invoice status, reconcile procurement PO and invoice amounts, and protect review formulas"
	plan := TemplateClassPlanHintForRequest(requestText)
	if plan == nil {
		t.Fatal("missing plan")
	}

	req, ok := DraftOrganismExecutionRequest(OrganismDraftInput{
		ScenarioID:        "procurement-organism-draft",
		RequestText:       requestText,
		InputFile:         inputFile,
		OutputFile:        outputFile,
		WorkbookFacts:     facts,
		TemplateClassPlan: *plan,
	})
	if !ok {
		t.Fatal("DraftOrganismExecutionRequest ok=false")
	}
	if len(req.Steps) != 3 {
		t.Fatalf("steps=%d want 3", len(req.Steps))
	}
	if req.Steps[0].AtomID != "add_data_validation" || req.Steps[0].SourceSheet != "Invoice" || req.Steps[0].ValidationRule == nil || req.Steps[0].ValidationRule.Ranges[0] != "C2:C20" {
		t.Fatalf("validation step=%+v", req.Steps[0])
	}
	if req.Steps[1].AtomID != "reconcile_tables" || req.Steps[1].SourceSheet != "PO" || req.Steps[1].LookupSheet != "Invoice" || req.Steps[1].TargetSheet != "ProcurementReconciliation" {
		t.Fatalf("reconcile step=%+v", req.Steps[1])
	}
	if len(req.Steps[1].CompareMappings) != 1 || req.Steps[1].CompareMappings[0].LeftColumn != "amount" || req.Steps[1].CompareMappings[0].RightColumn != "invoice_amount" {
		t.Fatalf("compare mappings=%+v", req.Steps[1].CompareMappings)
	}
	if req.Steps[2].AtomID != "protect_formula_cells" || req.Steps[2].SourceSheet != "Invoice" || req.Steps[2].ProtectionRule == nil || req.Steps[2].ProtectionRule.FormulaRanges[0] != "D2:D2" {
		t.Fatalf("protection step=%+v", req.Steps[2])
	}
	if err := runtimeschema.ValidateStruct(organismExecutionRequestSchemaPathForTest(), req); err != nil {
		t.Fatalf("draft organism request schema validation: %v", err)
	}
}

func TestDraftOrganismExecutionRequestBuildsWarehouseReorderStepsFromWorkbookFacts(t *testing.T) {
	tempDir := t.TempDir()
	inputFile := filepath.Join(tempDir, "warehouse.xlsx")
	outputFile := filepath.Join(tempDir, "warehouse-organism.xlsx")

	file := excelize.NewFile()
	defer func() { _ = file.Close() }()
	defaultSheet := file.GetSheetName(0)
	if err := file.SetSheetName(defaultSheet, "Stock"); err != nil {
		t.Fatalf("SetSheetName: %v", err)
	}
	if err := file.SetSheetRow("Stock", "A1", &[]any{"sku", "quantity", "reorder_level", "reorder_gap"}); err != nil {
		t.Fatalf("SetSheetRow stock header: %v", err)
	}
	if err := file.SetSheetRow("Stock", "A2", &[]any{"A001", 3, 5}); err != nil {
		t.Fatalf("SetSheetRow stock row: %v", err)
	}
	if err := file.SetCellFormula("Stock", "D2", "=B2-C2"); err != nil {
		t.Fatalf("SetCellFormula(D2): %v", err)
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
	if err := file.SetSheetRow("SKU", "A3", &[]any{"B002", "Aisle 2"}); err != nil {
		t.Fatalf("SetSheetRow SKU row2: %v", err)
	}
	if err := file.SaveAs(inputFile); err != nil {
		t.Fatalf("SaveAs: %v", err)
	}

	facts, err := runtimeinspect.InspectWorkbookFacts(inputFile)
	if err != nil {
		t.Fatalf("InspectWorkbookFacts: %v", err)
	}
	requestText := "Enrich warehouse reorder stock with SKU location, flag low stock, append reorder candidate, validate SKU, and protect reorder formulas"
	plan := TemplateClassPlanHintForRequest(requestText)
	if plan == nil {
		t.Fatal("missing plan")
	}

	req, ok := DraftOrganismExecutionRequest(OrganismDraftInput{
		ScenarioID:        "warehouse-organism-draft",
		RequestText:       requestText,
		InputFile:         inputFile,
		OutputFile:        outputFile,
		WorkbookFacts:     facts,
		TemplateClassPlan: *plan,
	})
	if !ok {
		t.Fatal("DraftOrganismExecutionRequest ok=false")
	}
	if len(req.Steps) != 5 {
		t.Fatalf("steps=%d want 5", len(req.Steps))
	}
	if req.Steps[0].AtomID != "join_lookup" || req.Steps[0].TargetSheet != "StockEnriched" || req.Steps[0].LookupSheet != "SKU" {
		t.Fatalf("lookup step=%+v", req.Steps[0])
	}
	if req.Steps[1].AtomID != "highlight_threshold" || req.Steps[1].SourceSheet != "StockEnriched" || req.Steps[1].TargetColumn != "quantity" || req.Steps[1].Threshold == nil || *req.Steps[1].Threshold != 5 {
		t.Fatalf("threshold step=%+v", req.Steps[1])
	}
	if req.Steps[2].AtomID != "append_structured_rows" || req.Steps[2].SourceSheet != "StockEnriched" || len(req.Steps[2].Values) != 5 {
		t.Fatalf("append step=%+v", req.Steps[2])
	}
	if req.Steps[3].AtomID != "add_data_validation" || req.Steps[3].ValidationRule == nil || req.Steps[3].ValidationRule.Ranges[0] != "A2:A20" {
		t.Fatalf("validation step=%+v", req.Steps[3])
	}
	if req.Steps[4].AtomID != "protect_formula_cells" || req.Steps[4].ProtectionRule == nil || req.Steps[4].ProtectionRule.FormulaRanges[0] != "D2:D2" {
		t.Fatalf("protection step=%+v", req.Steps[4])
	}
	if err := runtimeschema.ValidateStruct(organismExecutionRequestSchemaPathForTest(), req); err != nil {
		t.Fatalf("draft organism request schema validation: %v", err)
	}
}

func TestDraftOrganismExecutionRequestBuildsStudentGradebookStepsFromWorkbookFacts(t *testing.T) {
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

	facts, err := runtimeinspect.InspectWorkbookFacts(inputFile)
	if err != nil {
		t.Fatalf("InspectWorkbookFacts: %v", err)
	}
	requestText := "Summarize student scores in a gradebook, validate completion status, extend formulas, and protect calculated cells"
	plan := TemplateClassPlanHintForRequest(requestText)
	if plan == nil {
		t.Fatal("missing plan")
	}

	req, ok := DraftOrganismExecutionRequest(OrganismDraftInput{
		ScenarioID:        "gradebook-organism-draft",
		RequestText:       requestText,
		InputFile:         inputFile,
		OutputFile:        outputFile,
		WorkbookFacts:     facts,
		TemplateClassPlan: *plan,
	})
	if !ok {
		t.Fatal("DraftOrganismExecutionRequest ok=false")
	}
	if req.Steps[0].AtomID != "group_summarize" || req.Steps[0].TargetSheet != "GradeSummary" {
		t.Fatalf("summary step=%+v", req.Steps[0])
	}
	if len(req.Steps[0].Metrics) != 1 || req.Steps[0].Metrics[0].Column != "score" || req.Steps[0].Metrics[0].As != "score_total" {
		t.Fatalf("summary metrics=%+v", req.Steps[0].Metrics)
	}
	if req.Steps[1].AtomID != "add_data_validation" || req.Steps[1].ValidationRule == nil || req.Steps[1].ValidationRule.Ranges[0] != "D2:D20" {
		t.Fatalf("validation step=%+v", req.Steps[1])
	}
	if req.Steps[2].AtomID != "extend_table_formulas" || req.Steps[2].FormulaSourceRow != 2 || len(req.Steps[2].TargetRows) != 1 || req.Steps[2].TargetRows[0] != 4 {
		t.Fatalf("formula step=%+v", req.Steps[2])
	}
	if req.Steps[3].AtomID != "protect_formula_cells" || req.Steps[3].ProtectionRule == nil || req.Steps[3].ProtectionRule.FormulaRanges[0] != "E2:E4" {
		t.Fatalf("protection step=%+v", req.Steps[3])
	}
	if err := runtimeschema.ValidateStruct(organismExecutionRequestSchemaPathForTest(), req); err != nil {
		t.Fatalf("draft organism request schema validation: %v", err)
	}
}

func TestDraftOrganismExecutionRequestBuildsTrainingCompletionStepsFromWorkbookFacts(t *testing.T) {
	tempDir := t.TempDir()
	inputFile := filepath.Join(tempDir, "training.xlsx")
	outputFile := filepath.Join(tempDir, "training-organism.xlsx")

	file := excelize.NewFile()
	defer func() { _ = file.Close() }()
	defaultSheet := file.GetSheetName(0)
	if err := file.SetSheetName(defaultSheet, "Training"); err != nil {
		t.Fatalf("SetSheetName: %v", err)
	}
	if err := file.SetSheetRow("Training", "A1", &[]any{"employee", "course", "status", "completed", "completion_flag"}); err != nil {
		t.Fatalf("SetSheetRow header: %v", err)
	}
	for idx, row := range [][]any{{"Alex", "Safety", "complete", 1}, {"Alex", "Privacy", "complete", 1}, {"Blair", "Safety", "missing", 0}} {
		cell, _ := excelize.CoordinatesToCellName(1, idx+2)
		if err := file.SetSheetRow("Training", cell, &row); err != nil {
			t.Fatalf("SetSheetRow training %d: %v", idx, err)
		}
		rowNum := idx + 2
		if err := file.SetCellFormula("Training", "E"+cellRowForDraftTest(rowNum), "=IF(C"+cellRowForDraftTest(rowNum)+"=\"complete\",1,0)"); err != nil {
			t.Fatalf("SetCellFormula completed row %d: %v", rowNum, err)
		}
	}
	if err := file.SaveAs(inputFile); err != nil {
		t.Fatalf("SaveAs: %v", err)
	}

	facts, err := runtimeinspect.InspectWorkbookFacts(inputFile)
	if err != nil {
		t.Fatalf("InspectWorkbookFacts: %v", err)
	}
	requestText := "Summarize training completion by employee, validate training status, protect completion formulas, and generate a printable training report"
	plan := TemplateClassPlanHintForRequest(requestText)
	if plan == nil {
		t.Fatal("missing plan")
	}

	req, ok := DraftOrganismExecutionRequest(OrganismDraftInput{
		ScenarioID:        "training-organism-draft",
		RequestText:       requestText,
		InputFile:         inputFile,
		OutputFile:        outputFile,
		WorkbookFacts:     facts,
		TemplateClassPlan: *plan,
	})
	if !ok {
		t.Fatal("DraftOrganismExecutionRequest ok=false")
	}
	if len(req.Steps) != 4 {
		t.Fatalf("steps=%d want 4", len(req.Steps))
	}
	if req.Steps[0].AtomID != "group_summarize" || req.Steps[0].TargetSheet != "TrainingSummary" {
		t.Fatalf("summary step=%+v", req.Steps[0])
	}
	if len(req.Steps[0].Metrics) != 1 || req.Steps[0].Metrics[0].Column != "completed" || req.Steps[0].Metrics[0].As != "completed_total" {
		t.Fatalf("summary metrics=%+v", req.Steps[0].Metrics)
	}
	if req.Steps[1].AtomID != "add_data_validation" || req.Steps[1].ValidationRule == nil || req.Steps[1].ValidationRule.Ranges[0] != "C2:C20" {
		t.Fatalf("validation step=%+v", req.Steps[1])
	}
	if req.Steps[2].AtomID != "protect_formula_cells" || req.Steps[2].ProtectionRule == nil || req.Steps[2].ProtectionRule.FormulaRanges[0] != "E2:E4" {
		t.Fatalf("protection step=%+v", req.Steps[2])
	}
	if req.Steps[3].AtomID != "generate_printable_form" || req.Steps[3].TargetSheet != "TrainingReport" || req.Steps[3].TableBinding == nil {
		t.Fatalf("printable step=%+v", req.Steps[3])
	}
	if err := runtimeschema.ValidateStruct(organismExecutionRequestSchemaPathForTest(), req); err != nil {
		t.Fatalf("draft organism request schema validation: %v", err)
	}
}

func TestDraftOrganismExecutionRequestBuildsLoanRepaymentStepsFromWorkbookFacts(t *testing.T) {
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

	facts, err := runtimeinspect.InspectWorkbookFacts(inputFile)
	if err != nil {
		t.Fatalf("InspectWorkbookFacts: %v", err)
	}
	requestText := "Extend loan repayment schedule formulas and protect calculated balance cells"
	plan := TemplateClassPlanHintForRequest(requestText)
	if plan == nil {
		t.Fatal("missing plan")
	}

	req, ok := DraftOrganismExecutionRequest(OrganismDraftInput{
		ScenarioID:        "loan-organism-draft",
		RequestText:       requestText,
		InputFile:         inputFile,
		OutputFile:        outputFile,
		WorkbookFacts:     facts,
		TemplateClassPlan: *plan,
	})
	if !ok {
		t.Fatal("DraftOrganismExecutionRequest ok=false")
	}
	if len(req.Steps) != 2 {
		t.Fatalf("steps=%d want 2", len(req.Steps))
	}
	if req.Steps[0].AtomID != "extend_table_formulas" || req.Steps[0].FormulaSourceRow != 2 || len(req.Steps[0].TargetRows) != 1 || req.Steps[0].TargetRows[0] != 3 {
		t.Fatalf("formula step=%+v", req.Steps[0])
	}
	if len(req.Steps[0].FormulaColumns) != 3 || req.Steps[0].FormulaColumns[0] != "C" {
		t.Fatalf("formula columns=%+v", req.Steps[0].FormulaColumns)
	}
	if req.Steps[1].AtomID != "protect_formula_cells" || req.Steps[1].ProtectionRule == nil || req.Steps[1].ProtectionRule.FormulaRanges[0] != "C2:E3" {
		t.Fatalf("protection step=%+v", req.Steps[1])
	}
	if err := runtimeschema.ValidateStruct(organismExecutionRequestSchemaPathForTest(), req); err != nil {
		t.Fatalf("draft organism request schema validation: %v", err)
	}
}

func cellRowForDraftTest(row int) string {
	return strconv.Itoa(row)
}

func organismExecutionRequestSchemaPathForTest() string {
	_, file, _, ok := goruntime.Caller(0)
	if !ok {
		return filepath.Join("contracts", "requests", "organism_execution_request.schema.json")
	}
	return filepath.Join(resolvePackageRoot("", file), "contracts", "requests", "organism_execution_request.schema.json")
}
