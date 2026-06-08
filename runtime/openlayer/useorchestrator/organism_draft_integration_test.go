package useorchestrator

import (
	"encoding/json"
	"path/filepath"
	"strconv"
	"testing"

	runtimeinspect "github.com/mwroh/sheet-ops/runtime/inspect"
	requestcompiler "github.com/mwroh/sheet-ops/runtime/openlayer/requestcompiler"
	"github.com/xuri/excelize/v2"
)

func TestOrchestrateOrganismExecutesRequestCompilerInvoiceDraft(t *testing.T) {
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
	requestText := "Add invoice line items, extend totals, protect formulas, and create a printable invoice"
	plan := requestcompiler.TemplateClassPlanHintForRequest(requestText)
	if plan == nil {
		t.Fatal("missing plan")
	}
	draft, ok := requestcompiler.DraftOrganismExecutionRequest(requestcompiler.OrganismDraftInput{
		ScenarioID:        "invoice-organism-draft-integration",
		RequestText:       requestText,
		InputFile:         inputFile,
		OutputFile:        outputFile,
		WorkbookFacts:     facts,
		TemplateClassPlan: *plan,
	})
	if !ok {
		t.Fatal("DraftOrganismExecutionRequest ok=false")
	}
	raw, err := json.Marshal(draft)
	if err != nil {
		t.Fatalf("Marshal draft: %v", err)
	}
	var req OrganismExecutionRequest
	if err := json.Unmarshal(raw, &req); err != nil {
		t.Fatalf("Unmarshal draft: %v", err)
	}

	result, err := OrchestrateOrganism(req)
	if err != nil {
		t.Fatalf("OrchestrateOrganism: %v", err)
	}
	if !result.TemplateClassEvaluation.Pass || !result.OrganismVerification.Pass {
		t.Fatalf("organism evidence failed: eval=%+v verification=%+v", result.TemplateClassEvaluation, result.OrganismVerification)
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

func TestOrchestrateOrganismExecutesRequestCompilerExpenseDraft(t *testing.T) {
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
	plan := requestcompiler.TemplateClassPlanHintForRequest(requestText)
	if plan == nil {
		t.Fatal("missing plan")
	}
	draft, ok := requestcompiler.DraftOrganismExecutionRequest(requestcompiler.OrganismDraftInput{
		ScenarioID:        "expense-organism-draft-integration",
		RequestText:       requestText,
		InputFile:         inputFile,
		OutputFile:        outputFile,
		WorkbookFacts:     facts,
		TemplateClassPlan: *plan,
	})
	if !ok {
		t.Fatal("DraftOrganismExecutionRequest ok=false")
	}
	raw, err := json.Marshal(draft)
	if err != nil {
		t.Fatalf("Marshal draft: %v", err)
	}
	var req OrganismExecutionRequest
	if err := json.Unmarshal(raw, &req); err != nil {
		t.Fatalf("Unmarshal draft: %v", err)
	}

	result, err := OrchestrateOrganism(req)
	if err != nil {
		t.Fatalf("OrchestrateOrganism: %v", err)
	}
	if !result.TemplateClassEvaluation.Pass || !result.OrganismVerification.Pass {
		t.Fatalf("organism evidence failed: eval=%+v verification=%+v", result.TemplateClassEvaluation, result.OrganismVerification)
	}

	outputHandle, err := excelize.OpenFile(outputFile)
	if err != nil {
		t.Fatalf("Open output: %v", err)
	}
	defer func() { _ = outputHandle.Close() }()
	if got, err := outputHandle.GetCellValue("ExpenseItems", "A3"); err != nil || got != "Taxi" {
		t.Fatalf("ExpenseItems!A3=%q err=%v want Taxi", got, err)
	}
	if got, err := outputHandle.GetCellFormula("ExpenseItems", "E3"); err != nil || got != "=B3*C3" {
		t.Fatalf("ExpenseItems!E3 formula=%q err=%v want =B3*C3", got, err)
	}
	if got, err := outputHandle.GetCellValue("ExpenseClaim", "A1"); err != nil || got != "Expense Reimbursement" {
		t.Fatalf("ExpenseClaim!A1=%q err=%v want Expense Reimbursement", got, err)
	}
}

func TestOrchestrateOrganismExecutesRequestCompilerPurchaseOrderDraft(t *testing.T) {
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
	plan := requestcompiler.TemplateClassPlanHintForRequest(requestText)
	if plan == nil {
		t.Fatal("missing plan")
	}
	draft, ok := requestcompiler.DraftOrganismExecutionRequest(requestcompiler.OrganismDraftInput{
		ScenarioID:        "purchase-order-organism-draft-integration",
		RequestText:       requestText,
		InputFile:         inputFile,
		OutputFile:        outputFile,
		WorkbookFacts:     facts,
		TemplateClassPlan: *plan,
	})
	if !ok {
		t.Fatal("DraftOrganismExecutionRequest ok=false")
	}
	raw, err := json.Marshal(draft)
	if err != nil {
		t.Fatalf("Marshal draft: %v", err)
	}
	var req OrganismExecutionRequest
	if err := json.Unmarshal(raw, &req); err != nil {
		t.Fatalf("Unmarshal draft: %v", err)
	}

	result, err := OrchestrateOrganism(req)
	if err != nil {
		t.Fatalf("OrchestrateOrganism: %v", err)
	}
	if !result.TemplateClassEvaluation.Pass || !result.OrganismVerification.Pass {
		t.Fatalf("organism evidence failed: eval=%+v verification=%+v", result.TemplateClassEvaluation, result.OrganismVerification)
	}

	outputHandle, err := excelize.OpenFile(outputFile)
	if err != nil {
		t.Fatalf("Open output: %v", err)
	}
	defer func() { _ = outputHandle.Close() }()
	if got, err := outputHandle.GetCellValue("PurchaseOrder", "A3"); err != nil || got != "Monitor" {
		t.Fatalf("PurchaseOrder!A3=%q err=%v want Monitor", got, err)
	}
	if got, err := outputHandle.GetCellFormula("PurchaseOrder", "E3"); err != nil || got != "=B3*C3" {
		t.Fatalf("PurchaseOrder!E3 formula=%q err=%v want =B3*C3", got, err)
	}
	if got, err := outputHandle.GetCellValue("PurchaseOrderPrint", "A1"); err != nil || got != "Purchase Order" {
		t.Fatalf("PurchaseOrderPrint!A1=%q err=%v want Purchase Order", got, err)
	}
}

func TestOrchestrateOrganismExecutesRequestCompilerBudgetDraft(t *testing.T) {
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
		if err := file.SetCellFormula("Budget", "E"+cellRowForDraftIntegrationTest(rowNum), "=B"+cellRowForDraftIntegrationTest(rowNum)+"+C"+cellRowForDraftIntegrationTest(rowNum)); err != nil {
			t.Fatalf("SetCellFormula review_total row %d: %v", rowNum, err)
		}
		if err := file.SetCellFormula("Budget", "F"+cellRowForDraftIntegrationTest(rowNum), "=B"+cellRowForDraftIntegrationTest(rowNum)+"-C"+cellRowForDraftIntegrationTest(rowNum)); err != nil {
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
	plan := requestcompiler.TemplateClassPlanHintForRequest(requestText)
	if plan == nil {
		t.Fatal("missing plan")
	}
	draft, ok := requestcompiler.DraftOrganismExecutionRequest(requestcompiler.OrganismDraftInput{
		ScenarioID:        "budget-organism-draft-integration",
		RequestText:       requestText,
		InputFile:         inputFile,
		OutputFile:        outputFile,
		WorkbookFacts:     facts,
		TemplateClassPlan: *plan,
	})
	if !ok {
		t.Fatal("DraftOrganismExecutionRequest ok=false")
	}
	raw, err := json.Marshal(draft)
	if err != nil {
		t.Fatalf("Marshal draft: %v", err)
	}
	var req OrganismExecutionRequest
	if err := json.Unmarshal(raw, &req); err != nil {
		t.Fatalf("Unmarshal draft: %v", err)
	}

	result, err := OrchestrateOrganism(req)
	if err != nil {
		t.Fatalf("OrchestrateOrganism: %v", err)
	}
	if !result.TemplateClassEvaluation.Pass || !result.OrganismVerification.Pass {
		t.Fatalf("organism evidence failed: eval=%+v verification=%+v", result.TemplateClassEvaluation, result.OrganismVerification)
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

func TestOrchestrateOrganismExecutesRequestCompilerCashFlowDraft(t *testing.T) {
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
		if err := file.SetCellFormula("Jan", "E"+cellRowForDraftIntegrationTest(rowNum), "=B"+cellRowForDraftIntegrationTest(rowNum)+"+C"+cellRowForDraftIntegrationTest(rowNum)+"-D"+cellRowForDraftIntegrationTest(rowNum)); err != nil {
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
	plan := requestcompiler.TemplateClassPlanHintForRequest(requestText)
	if plan == nil {
		t.Fatal("missing plan")
	}
	draft, ok := requestcompiler.DraftOrganismExecutionRequest(requestcompiler.OrganismDraftInput{
		ScenarioID:        "cash-flow-organism-draft-integration",
		RequestText:       requestText,
		InputFile:         inputFile,
		OutputFile:        outputFile,
		WorkbookFacts:     facts,
		TemplateClassPlan: *plan,
	})
	if !ok {
		t.Fatal("DraftOrganismExecutionRequest ok=false")
	}
	raw, err := json.Marshal(draft)
	if err != nil {
		t.Fatalf("Marshal draft: %v", err)
	}
	var req OrganismExecutionRequest
	if err := json.Unmarshal(raw, &req); err != nil {
		t.Fatalf("Unmarshal draft: %v", err)
	}

	result, err := OrchestrateOrganism(req)
	if err != nil {
		t.Fatalf("OrchestrateOrganism: %v", err)
	}
	if !result.TemplateClassEvaluation.Pass || !result.OrganismVerification.Pass {
		t.Fatalf("organism evidence failed: eval=%+v verification=%+v", result.TemplateClassEvaluation, result.OrganismVerification)
	}

	outputHandle, err := excelize.OpenFile(outputFile)
	if err != nil {
		t.Fatalf("Open output: %v", err)
	}
	defer func() { _ = outputHandle.Close() }()
	if got, err := outputHandle.GetCellValue("CashFlowSummary", "B2"); err != nil || got != "115" {
		t.Fatalf("CashFlowSummary!B2=%q err=%v want 115", got, err)
	}
	if got, err := outputHandle.GetCellFormula("NextCashFlow", "E4"); err != nil || got != "=B4+C4-D4" {
		t.Fatalf("NextCashFlow!E4 formula=%q err=%v want =B4+C4-D4", got, err)
	}
	if got, err := outputHandle.GetCellValue("NextCashFlow", "B3"); err != nil || got != "150" {
		t.Fatalf("NextCashFlow!B3=%q err=%v want 150", got, err)
	}
}

func TestOrchestrateOrganismExecutesRequestCompilerAttendanceDraft(t *testing.T) {
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
	plan := requestcompiler.TemplateClassPlanHintForRequest(requestText)
	if plan == nil {
		t.Fatal("missing plan")
	}
	draft, ok := requestcompiler.DraftOrganismExecutionRequest(requestcompiler.OrganismDraftInput{
		ScenarioID:        "attendance-organism-draft-integration",
		RequestText:       requestText,
		InputFile:         inputFile,
		OutputFile:        outputFile,
		WorkbookFacts:     facts,
		TemplateClassPlan: *plan,
	})
	if !ok {
		t.Fatal("DraftOrganismExecutionRequest ok=false")
	}
	raw, err := json.Marshal(draft)
	if err != nil {
		t.Fatalf("Marshal draft: %v", err)
	}
	var req OrganismExecutionRequest
	if err := json.Unmarshal(raw, &req); err != nil {
		t.Fatalf("Unmarshal draft: %v", err)
	}

	result, err := OrchestrateOrganism(req)
	if err != nil {
		t.Fatalf("OrchestrateOrganism: %v", err)
	}
	if !result.TemplateClassEvaluation.Pass || !result.OrganismVerification.Pass {
		t.Fatalf("organism evidence failed: eval=%+v verification=%+v", result.TemplateClassEvaluation, result.OrganismVerification)
	}

	outputHandle, err := excelize.OpenFile(outputFile)
	if err != nil {
		t.Fatalf("Open output: %v", err)
	}
	defer func() { _ = outputHandle.Close() }()
	if got, err := outputHandle.GetCellValue("NextAttendance", "A2"); err != nil || got != "Alice" {
		t.Fatalf("NextAttendance!A2=%q err=%v want Alice", got, err)
	}
	if got, err := outputHandle.GetCellFormula("NextAttendance", "D2"); err != nil || got != "=IF(C2=\"present\",1,0)" {
		t.Fatalf("NextAttendance!D2 formula=%q err=%v want =IF(C2=\"present\",1,0)", got, err)
	}
}

func TestOrchestrateOrganismExecutesRequestCompilerProjectTimelineDraft(t *testing.T) {
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
	plan := requestcompiler.TemplateClassPlanHintForRequest(requestText)
	if plan == nil {
		t.Fatal("missing plan")
	}
	draft, ok := requestcompiler.DraftOrganismExecutionRequest(requestcompiler.OrganismDraftInput{
		ScenarioID:        "project-timeline-organism-draft-integration",
		RequestText:       requestText,
		InputFile:         inputFile,
		OutputFile:        outputFile,
		WorkbookFacts:     facts,
		TemplateClassPlan: *plan,
	})
	if !ok {
		t.Fatal("DraftOrganismExecutionRequest ok=false")
	}
	raw, err := json.Marshal(draft)
	if err != nil {
		t.Fatalf("Marshal draft: %v", err)
	}
	var req OrganismExecutionRequest
	if err := json.Unmarshal(raw, &req); err != nil {
		t.Fatalf("Unmarshal draft: %v", err)
	}

	result, err := OrchestrateOrganism(req)
	if err != nil {
		t.Fatalf("OrchestrateOrganism: %v", err)
	}
	if !result.TemplateClassEvaluation.Pass || !result.OrganismVerification.Pass {
		t.Fatalf("organism evidence failed: eval=%+v verification=%+v", result.TemplateClassEvaluation, result.OrganismVerification)
	}

	outputHandle, err := excelize.OpenFile(outputFile)
	if err != nil {
		t.Fatalf("Open output: %v", err)
	}
	defer func() { _ = outputHandle.Close() }()
	if got, err := outputHandle.GetCellValue("Sprint2", "A4"); err != nil || got != "Test" {
		t.Fatalf("Sprint2!A4=%q err=%v want Test", got, err)
	}
	if got, err := outputHandle.GetCellFormula("Sprint2", "F4"); err != nil || got != "=IF(D4=\"done\",1,0)" {
		t.Fatalf("Sprint2!F4 formula=%q err=%v want =IF(D4=\"done\",1,0)", got, err)
	}
	if got, err := outputHandle.GetCellValue("TimelineSummary", "A2"); err != nil || got != "todo" {
		t.Fatalf("TimelineSummary!A2=%q err=%v want todo", got, err)
	}
}

func TestOrchestrateOrganismExecutesRequestCompilerShiftRosterDraft(t *testing.T) {
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
	plan := requestcompiler.TemplateClassPlanHintForRequest(requestText)
	if plan == nil {
		t.Fatal("missing plan")
	}
	draft, ok := requestcompiler.DraftOrganismExecutionRequest(requestcompiler.OrganismDraftInput{
		ScenarioID:        "shift-roster-organism-draft-integration",
		RequestText:       requestText,
		InputFile:         inputFile,
		OutputFile:        outputFile,
		WorkbookFacts:     facts,
		TemplateClassPlan: *plan,
	})
	if !ok {
		t.Fatal("DraftOrganismExecutionRequest ok=false")
	}
	raw, err := json.Marshal(draft)
	if err != nil {
		t.Fatalf("Marshal draft: %v", err)
	}
	var req OrganismExecutionRequest
	if err := json.Unmarshal(raw, &req); err != nil {
		t.Fatalf("Unmarshal draft: %v", err)
	}

	result, err := OrchestrateOrganism(req)
	if err != nil {
		t.Fatalf("OrchestrateOrganism: %v", err)
	}
	if !result.TemplateClassEvaluation.Pass || !result.OrganismVerification.Pass {
		t.Fatalf("organism evidence failed: eval=%+v verification=%+v", result.TemplateClassEvaluation, result.OrganismVerification)
	}

	outputHandle, err := excelize.OpenFile(outputFile)
	if err != nil {
		t.Fatalf("Open output: %v", err)
	}
	defer func() { _ = outputHandle.Close() }()
	if got, err := outputHandle.GetCellValue("Week2", "A2"); err != nil || got != "Alex" {
		t.Fatalf("Week2!A2=%q err=%v want Alex", got, err)
	}
	if got, err := outputHandle.GetCellFormula("Week2", "D2"); err != nil || got != "=IF(C2=\"OFF\",0,1)" {
		t.Fatalf("Week2!D2 formula=%q err=%v want =IF(C2=\"OFF\",0,1)", got, err)
	}
}

func TestOrchestrateOrganismExecutesRequestCompilerConstructionCostDraft(t *testing.T) {
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
	plan := requestcompiler.TemplateClassPlanHintForRequest(requestText)
	if plan == nil {
		t.Fatal("missing plan")
	}
	draft, ok := requestcompiler.DraftOrganismExecutionRequest(requestcompiler.OrganismDraftInput{
		ScenarioID:        "construction-cost-organism-draft-integration",
		RequestText:       requestText,
		InputFile:         inputFile,
		OutputFile:        outputFile,
		WorkbookFacts:     facts,
		TemplateClassPlan: *plan,
	})
	if !ok {
		t.Fatal("DraftOrganismExecutionRequest ok=false")
	}
	raw, err := json.Marshal(draft)
	if err != nil {
		t.Fatalf("Marshal draft: %v", err)
	}
	var req OrganismExecutionRequest
	if err := json.Unmarshal(raw, &req); err != nil {
		t.Fatalf("Unmarshal draft: %v", err)
	}

	result, err := OrchestrateOrganism(req)
	if err != nil {
		t.Fatalf("OrchestrateOrganism: %v", err)
	}
	if !result.TemplateClassEvaluation.Pass || !result.OrganismVerification.Pass {
		t.Fatalf("organism evidence failed: eval=%+v verification=%+v", result.TemplateClassEvaluation, result.OrganismVerification)
	}
	if len(result.StepResults) < 3 || len(result.StepResults[2].Verification.HighlightedRows) != 1 || result.StepResults[2].Verification.HighlightedRows[0] != 3 {
		t.Fatalf("highlight evidence=%+v", result.StepResults)
	}

	outputHandle, err := excelize.OpenFile(outputFile)
	if err != nil {
		t.Fatalf("Open output: %v", err)
	}
	defer func() { _ = outputHandle.Close() }()
	if got, err := outputHandle.GetCellValue("Costs", "A3"); err != nil || got != "CHANGE" {
		t.Fatalf("Costs!A3=%q err=%v want CHANGE", got, err)
	}
	if got, err := outputHandle.GetCellFormula("Costs", "E3"); err != nil || got != "=C3-D3" {
		t.Fatalf("Costs!E3 formula=%q err=%v want =C3-D3", got, err)
	}
}

func TestOrchestrateOrganismExecutesRequestCompilerTimesheetDraft(t *testing.T) {
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
	plan := requestcompiler.TemplateClassPlanHintForRequest(requestText)
	if plan == nil {
		t.Fatal("missing plan")
	}
	draft, ok := requestcompiler.DraftOrganismExecutionRequest(requestcompiler.OrganismDraftInput{
		ScenarioID:        "timesheet-organism-draft-integration",
		RequestText:       requestText,
		InputFile:         inputFile,
		OutputFile:        outputFile,
		WorkbookFacts:     facts,
		TemplateClassPlan: *plan,
	})
	if !ok {
		t.Fatal("DraftOrganismExecutionRequest ok=false")
	}
	raw, err := json.Marshal(draft)
	if err != nil {
		t.Fatalf("Marshal draft: %v", err)
	}
	var req OrganismExecutionRequest
	if err := json.Unmarshal(raw, &req); err != nil {
		t.Fatalf("Unmarshal draft: %v", err)
	}

	result, err := OrchestrateOrganism(req)
	if err != nil {
		t.Fatalf("OrchestrateOrganism: %v", err)
	}
	if !result.TemplateClassEvaluation.Pass || !result.OrganismVerification.Pass {
		t.Fatalf("organism evidence failed: eval=%+v verification=%+v", result.TemplateClassEvaluation, result.OrganismVerification)
	}

	outputHandle, err := excelize.OpenFile(outputFile)
	if err != nil {
		t.Fatalf("Open output: %v", err)
	}
	defer func() { _ = outputHandle.Close() }()
	if got, err := outputHandle.GetCellValue("Week2", "B3"); err != nil || got != "Ben" {
		t.Fatalf("Week2!B3=%q err=%v want Ben", got, err)
	}
	if got, err := outputHandle.GetCellFormula("Week2", "F3"); err != nil || got != "=D3*E3" {
		t.Fatalf("Week2!F3 formula=%q err=%v want =D3*E3", got, err)
	}
}

func TestOrchestrateOrganismExecutesRequestCompilerInventoryDraft(t *testing.T) {
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
		if err := file.SetCellFormula("Movements", "D"+cellRowForDraftIntegrationTest(rowNum), "=B"+cellRowForDraftIntegrationTest(rowNum)+"-C"+cellRowForDraftIntegrationTest(rowNum)); err != nil {
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
	plan := requestcompiler.TemplateClassPlanHintForRequest(requestText)
	if plan == nil {
		t.Fatal("missing plan")
	}
	draft, ok := requestcompiler.DraftOrganismExecutionRequest(requestcompiler.OrganismDraftInput{
		ScenarioID:        "inventory-organism-draft-integration",
		RequestText:       requestText,
		InputFile:         inputFile,
		OutputFile:        outputFile,
		WorkbookFacts:     facts,
		TemplateClassPlan: *plan,
	})
	if !ok {
		t.Fatal("DraftOrganismExecutionRequest ok=false")
	}
	raw, err := json.Marshal(draft)
	if err != nil {
		t.Fatalf("Marshal draft: %v", err)
	}
	var req OrganismExecutionRequest
	if err := json.Unmarshal(raw, &req); err != nil {
		t.Fatalf("Unmarshal draft: %v", err)
	}

	result, err := OrchestrateOrganism(req)
	if err != nil {
		t.Fatalf("OrchestrateOrganism: %v", err)
	}
	if !result.TemplateClassEvaluation.Pass || !result.OrganismVerification.Pass {
		t.Fatalf("organism evidence failed: eval=%+v verification=%+v", result.TemplateClassEvaluation, result.OrganismVerification)
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

func TestOrchestrateOrganismExecutesRequestCompilerProcurementDraft(t *testing.T) {
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
	plan := requestcompiler.TemplateClassPlanHintForRequest(requestText)
	if plan == nil {
		t.Fatal("missing plan")
	}
	draft, ok := requestcompiler.DraftOrganismExecutionRequest(requestcompiler.OrganismDraftInput{
		ScenarioID:        "procurement-organism-draft-integration",
		RequestText:       requestText,
		InputFile:         inputFile,
		OutputFile:        outputFile,
		WorkbookFacts:     facts,
		TemplateClassPlan: *plan,
	})
	if !ok {
		t.Fatal("DraftOrganismExecutionRequest ok=false")
	}
	raw, err := json.Marshal(draft)
	if err != nil {
		t.Fatalf("Marshal draft: %v", err)
	}
	var req OrganismExecutionRequest
	if err := json.Unmarshal(raw, &req); err != nil {
		t.Fatalf("Unmarshal draft: %v", err)
	}

	result, err := OrchestrateOrganism(req)
	if err != nil {
		t.Fatalf("OrchestrateOrganism: %v", err)
	}
	if !result.TemplateClassEvaluation.Pass || !result.OrganismVerification.Pass {
		t.Fatalf("organism evidence failed: eval=%+v verification=%+v", result.TemplateClassEvaluation, result.OrganismVerification)
	}

	outputHandle, err := excelize.OpenFile(outputFile)
	if err != nil {
		t.Fatalf("Open output: %v", err)
	}
	defer func() { _ = outputHandle.Close() }()
	if got, err := outputHandle.GetCellValue("ProcurementReconciliation", "A2"); err != nil || got != "value_mismatch" {
		t.Fatalf("ProcurementReconciliation!A2=%q err=%v want value_mismatch", got, err)
	}
	if got, err := outputHandle.GetCellValue("ProcurementReconciliation", "B2"); err != nil || got != "PO-1" {
		t.Fatalf("ProcurementReconciliation!B2=%q err=%v want PO-1", got, err)
	}
	if got, err := outputHandle.GetCellFormula("Invoice", "D2"); err != nil || got != "=IF(B2>0,\"review\",\"missing\")" {
		t.Fatalf("Invoice!D2 formula=%q err=%v want =IF(B2>0,\"review\",\"missing\")", got, err)
	}
}

func TestOrchestrateOrganismExecutesRequestCompilerWarehouseDraft(t *testing.T) {
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
	plan := requestcompiler.TemplateClassPlanHintForRequest(requestText)
	if plan == nil {
		t.Fatal("missing plan")
	}
	draft, ok := requestcompiler.DraftOrganismExecutionRequest(requestcompiler.OrganismDraftInput{
		ScenarioID:        "warehouse-organism-draft-integration",
		RequestText:       requestText,
		InputFile:         inputFile,
		OutputFile:        outputFile,
		WorkbookFacts:     facts,
		TemplateClassPlan: *plan,
	})
	if !ok {
		t.Fatal("DraftOrganismExecutionRequest ok=false")
	}
	raw, err := json.Marshal(draft)
	if err != nil {
		t.Fatalf("Marshal draft: %v", err)
	}
	var req OrganismExecutionRequest
	if err := json.Unmarshal(raw, &req); err != nil {
		t.Fatalf("Unmarshal draft: %v", err)
	}

	result, err := OrchestrateOrganism(req)
	if err != nil {
		t.Fatalf("OrchestrateOrganism: %v", err)
	}
	if !result.TemplateClassEvaluation.Pass || !result.OrganismVerification.Pass {
		t.Fatalf("organism evidence failed: eval=%+v verification=%+v", result.TemplateClassEvaluation, result.OrganismVerification)
	}

	outputHandle, err := excelize.OpenFile(outputFile)
	if err != nil {
		t.Fatalf("Open output: %v", err)
	}
	defer func() { _ = outputHandle.Close() }()
	if got, err := outputHandle.GetCellValue("StockEnriched", "E2"); err != nil || got != "Aisle 1" {
		t.Fatalf("StockEnriched!E2=%q err=%v want Aisle 1", got, err)
	}
	if got, err := outputHandle.GetCellValue("StockEnriched", "A3"); err != nil || got != "B002" {
		t.Fatalf("StockEnriched!A3=%q err=%v want B002", got, err)
	}
	if got, err := outputHandle.GetCellFormula("Stock", "D2"); err != nil || got != "=B2-C2" {
		t.Fatalf("Stock!D2 formula=%q err=%v want =B2-C2", got, err)
	}
}

func TestOrchestrateOrganismExecutesRequestCompilerGradebookDraft(t *testing.T) {
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
	plan := requestcompiler.TemplateClassPlanHintForRequest(requestText)
	if plan == nil {
		t.Fatal("missing plan")
	}
	draft, ok := requestcompiler.DraftOrganismExecutionRequest(requestcompiler.OrganismDraftInput{
		ScenarioID:        "gradebook-organism-draft-integration",
		RequestText:       requestText,
		InputFile:         inputFile,
		OutputFile:        outputFile,
		WorkbookFacts:     facts,
		TemplateClassPlan: *plan,
	})
	if !ok {
		t.Fatal("DraftOrganismExecutionRequest ok=false")
	}
	raw, err := json.Marshal(draft)
	if err != nil {
		t.Fatalf("Marshal draft: %v", err)
	}
	var req OrganismExecutionRequest
	if err := json.Unmarshal(raw, &req); err != nil {
		t.Fatalf("Unmarshal draft: %v", err)
	}

	result, err := OrchestrateOrganism(req)
	if err != nil {
		t.Fatalf("OrchestrateOrganism: %v", err)
	}
	if !result.TemplateClassEvaluation.Pass || !result.OrganismVerification.Pass {
		t.Fatalf("organism evidence failed: eval=%+v verification=%+v", result.TemplateClassEvaluation, result.OrganismVerification)
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

func TestOrchestrateOrganismExecutesRequestCompilerTrainingCompletionDraft(t *testing.T) {
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
		if err := file.SetCellFormula("Training", "E"+cellRowForDraftIntegrationTest(rowNum), "=IF(C"+cellRowForDraftIntegrationTest(rowNum)+"=\"complete\",1,0)"); err != nil {
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
	plan := requestcompiler.TemplateClassPlanHintForRequest(requestText)
	if plan == nil {
		t.Fatal("missing plan")
	}
	draft, ok := requestcompiler.DraftOrganismExecutionRequest(requestcompiler.OrganismDraftInput{
		ScenarioID:        "training-organism-draft-integration",
		RequestText:       requestText,
		InputFile:         inputFile,
		OutputFile:        outputFile,
		WorkbookFacts:     facts,
		TemplateClassPlan: *plan,
	})
	if !ok {
		t.Fatal("DraftOrganismExecutionRequest ok=false")
	}
	raw, err := json.Marshal(draft)
	if err != nil {
		t.Fatalf("Marshal draft: %v", err)
	}
	var req OrganismExecutionRequest
	if err := json.Unmarshal(raw, &req); err != nil {
		t.Fatalf("Unmarshal draft: %v", err)
	}

	result, err := OrchestrateOrganism(req)
	if err != nil {
		t.Fatalf("OrchestrateOrganism: %v", err)
	}
	if !result.TemplateClassEvaluation.Pass || !result.OrganismVerification.Pass {
		t.Fatalf("organism evidence failed: eval=%+v verification=%+v", result.TemplateClassEvaluation, result.OrganismVerification)
	}

	outputHandle, err := excelize.OpenFile(outputFile)
	if err != nil {
		t.Fatalf("Open output: %v", err)
	}
	defer func() { _ = outputHandle.Close() }()
	if got, err := outputHandle.GetCellValue("TrainingSummary", "B2"); err != nil || got != "2" {
		t.Fatalf("TrainingSummary!B2=%q err=%v want 2", got, err)
	}
	if got, err := outputHandle.GetCellValue("TrainingReport", "A1"); err != nil || got != "Training Completion" {
		t.Fatalf("TrainingReport!A1=%q err=%v want Training Completion", got, err)
	}
	if got, err := outputHandle.GetCellFormula("Training", "E2"); err != nil || got != "=IF(C2=\"complete\",1,0)" {
		t.Fatalf("Training!E2 formula=%q err=%v want =IF(C2=\"complete\",1,0)", got, err)
	}
}

func TestOrchestrateOrganismExecutesRequestCompilerServiceTicketDraft(t *testing.T) {
	tempDir := t.TempDir()
	inputFile := filepath.Join(tempDir, "tickets.xlsx")
	outputFile := filepath.Join(tempDir, "tickets-organism.xlsx")

	file := excelize.NewFile()
	defer func() { _ = file.Close() }()
	defaultSheet := file.GetSheetName(0)
	if err := file.SetSheetName(defaultSheet, "Tickets"); err != nil {
		t.Fatalf("SetSheetName: %v", err)
	}
	if err := file.SetSheetRow("Tickets", "A1", &[]any{"ticket_id", "status", "days_open", "sla_breach", "ticket_count"}); err != nil {
		t.Fatalf("SetSheetRow header: %v", err)
	}
	if err := file.SetSheetRow("Tickets", "A2", &[]any{"T-1", "open", 2, nil, 1}); err != nil {
		t.Fatalf("SetSheetRow row: %v", err)
	}
	if err := file.SetCellFormula("Tickets", "D2", "=IF(C2>5,1,0)"); err != nil {
		t.Fatalf("SetCellFormula(D2): %v", err)
	}
	if err := file.SaveAs(inputFile); err != nil {
		t.Fatalf("SaveAs: %v", err)
	}

	facts, err := runtimeinspect.InspectWorkbookFacts(inputFile)
	if err != nil {
		t.Fatalf("InspectWorkbookFacts: %v", err)
	}
	requestText := "Append service ticket, validate ticket status, flag overdue tickets, summarize open tickets, and protect SLA formulas"
	plan := requestcompiler.TemplateClassPlanHintForRequest(requestText)
	if plan == nil {
		t.Fatal("missing plan")
	}
	draft, ok := requestcompiler.DraftOrganismExecutionRequest(requestcompiler.OrganismDraftInput{
		ScenarioID:        "service-ticket-organism-draft-integration",
		RequestText:       requestText,
		InputFile:         inputFile,
		OutputFile:        outputFile,
		WorkbookFacts:     facts,
		TemplateClassPlan: *plan,
	})
	if !ok {
		t.Fatal("DraftOrganismExecutionRequest ok=false")
	}
	raw, err := json.Marshal(draft)
	if err != nil {
		t.Fatalf("Marshal draft: %v", err)
	}
	var req OrganismExecutionRequest
	if err := json.Unmarshal(raw, &req); err != nil {
		t.Fatalf("Unmarshal draft: %v", err)
	}

	result, err := OrchestrateOrganism(req)
	if err != nil {
		t.Fatalf("OrchestrateOrganism: %v", err)
	}
	if !result.TemplateClassEvaluation.Pass || !result.OrganismVerification.Pass {
		t.Fatalf("organism evidence failed: eval=%+v verification=%+v", result.TemplateClassEvaluation, result.OrganismVerification)
	}
	if len(result.StepResults) < 4 || len(result.StepResults[3].Verification.HighlightedRows) != 1 || result.StepResults[3].Verification.HighlightedRows[0] != 3 {
		t.Fatalf("highlight evidence=%+v", result.StepResults)
	}

	outputHandle, err := excelize.OpenFile(outputFile)
	if err != nil {
		t.Fatalf("Open output: %v", err)
	}
	defer func() { _ = outputHandle.Close() }()
	if got, err := outputHandle.GetCellValue("Tickets", "A3"); err != nil || got != "T-2" {
		t.Fatalf("Tickets!A3=%q err=%v want T-2", got, err)
	}
	if got, err := outputHandle.GetCellValue("TicketSummary", "B2"); err != nil || got != "2" {
		t.Fatalf("TicketSummary!B2=%q err=%v want 2", got, err)
	}
	if got, err := outputHandle.GetCellFormula("Tickets", "D3"); err != nil || got != "=IF(C3>5,1,0)" {
		t.Fatalf("Tickets!D3 formula=%q err=%v want =IF(C3>5,1,0)", got, err)
	}
}

func TestOrchestrateOrganismExecutesRequestCompilerSalesPipelineDraft(t *testing.T) {
	tempDir := t.TempDir()
	inputFile := filepath.Join(tempDir, "sales.xlsx")
	outputFile := filepath.Join(tempDir, "sales-organism.xlsx")

	file := excelize.NewFile()
	defer func() { _ = file.Close() }()
	defaultSheet := file.GetSheetName(0)
	if err := file.SetSheetName(defaultSheet, "Pipeline"); err != nil {
		t.Fatalf("SetSheetName: %v", err)
	}
	if err := file.SetSheetRow("Pipeline", "A1", &[]any{"deal_id", "stage", "deal_value", "probability", "weighted_value", "deal_count"}); err != nil {
		t.Fatalf("SetSheetRow header: %v", err)
	}
	if err := file.SetSheetRow("Pipeline", "A2", &[]any{"D-1", "proposal", 5000, 0.5, nil, 1}); err != nil {
		t.Fatalf("SetSheetRow row: %v", err)
	}
	if err := file.SetCellFormula("Pipeline", "E2", "=C2*D2"); err != nil {
		t.Fatalf("SetCellFormula(E2): %v", err)
	}
	if err := file.SaveAs(inputFile); err != nil {
		t.Fatalf("SaveAs: %v", err)
	}

	facts, err := runtimeinspect.InspectWorkbookFacts(inputFile)
	if err != nil {
		t.Fatalf("InspectWorkbookFacts: %v", err)
	}
	requestText := "Append sales pipeline deal, validate sales stage, extend weighted value formulas, flag large deals, summarize pipeline stages, and protect forecast formulas"
	plan := requestcompiler.TemplateClassPlanHintForRequest(requestText)
	if plan == nil {
		t.Fatal("missing plan")
	}
	draft, ok := requestcompiler.DraftOrganismExecutionRequest(requestcompiler.OrganismDraftInput{
		ScenarioID:        "sales-pipeline-organism-draft-integration",
		RequestText:       requestText,
		InputFile:         inputFile,
		OutputFile:        outputFile,
		WorkbookFacts:     facts,
		TemplateClassPlan: *plan,
	})
	if !ok {
		t.Fatal("DraftOrganismExecutionRequest ok=false")
	}
	raw, err := json.Marshal(draft)
	if err != nil {
		t.Fatalf("Marshal draft: %v", err)
	}
	var req OrganismExecutionRequest
	if err := json.Unmarshal(raw, &req); err != nil {
		t.Fatalf("Unmarshal draft: %v", err)
	}

	result, err := OrchestrateOrganism(req)
	if err != nil {
		t.Fatalf("OrchestrateOrganism: %v", err)
	}
	if !result.TemplateClassEvaluation.Pass || !result.OrganismVerification.Pass {
		t.Fatalf("organism evidence failed: eval=%+v verification=%+v", result.TemplateClassEvaluation, result.OrganismVerification)
	}
	if len(result.StepResults) < 4 || len(result.StepResults[3].Verification.HighlightedRows) != 1 || result.StepResults[3].Verification.HighlightedRows[0] != 3 {
		t.Fatalf("highlight evidence=%+v", result.StepResults)
	}

	outputHandle, err := excelize.OpenFile(outputFile)
	if err != nil {
		t.Fatalf("Open output: %v", err)
	}
	defer func() { _ = outputHandle.Close() }()
	if got, err := outputHandle.GetCellValue("Pipeline", "A3"); err != nil || got != "D-2" {
		t.Fatalf("Pipeline!A3=%q err=%v want D-2", got, err)
	}
	if got, err := outputHandle.GetCellValue("PipelineSummary", "B2"); err != nil || got != "2" {
		t.Fatalf("PipelineSummary!B2=%q err=%v want 2", got, err)
	}
	if got, err := outputHandle.GetCellFormula("Pipeline", "E3"); err != nil || got != "=C3*D3" {
		t.Fatalf("Pipeline!E3 formula=%q err=%v want =C3*D3", got, err)
	}
}

func TestOrchestrateOrganismExecutesRequestCompilerMaintenanceIssueDraft(t *testing.T) {
	tempDir := t.TempDir()
	inputFile := filepath.Join(tempDir, "maintenance.xlsx")
	outputFile := filepath.Join(tempDir, "maintenance-organism.xlsx")

	file := excelize.NewFile()
	defer func() { _ = file.Close() }()
	defaultSheet := file.GetSheetName(0)
	if err := file.SetSheetName(defaultSheet, "Maintenance"); err != nil {
		t.Fatalf("SetSheetName: %v", err)
	}
	if err := file.SetSheetRow("Maintenance", "A1", &[]any{"issue_id", "status", "days_open", "risk_score", "action_required", "action_count"}); err != nil {
		t.Fatalf("SetSheetRow header: %v", err)
	}
	if err := file.SetSheetRow("Maintenance", "A2", &[]any{"M-1", "open", 2, 2, nil, 1}); err != nil {
		t.Fatalf("SetSheetRow row: %v", err)
	}
	if err := file.SetCellFormula("Maintenance", "E2", "=IF(D2>=4,1,0)"); err != nil {
		t.Fatalf("SetCellFormula(E2): %v", err)
	}
	if err := file.SaveAs(inputFile); err != nil {
		t.Fatalf("SaveAs: %v", err)
	}

	facts, err := runtimeinspect.InspectWorkbookFacts(inputFile)
	if err != nil {
		t.Fatalf("InspectWorkbookFacts: %v", err)
	}
	requestText := "Append maintenance issue, validate issue status, extend action-required formulas, flag high risk issues, summarize maintenance status, and protect action formulas"
	plan := requestcompiler.TemplateClassPlanHintForRequest(requestText)
	if plan == nil {
		t.Fatal("missing plan")
	}
	draft, ok := requestcompiler.DraftOrganismExecutionRequest(requestcompiler.OrganismDraftInput{
		ScenarioID:        "maintenance-issue-organism-draft-integration",
		RequestText:       requestText,
		InputFile:         inputFile,
		OutputFile:        outputFile,
		WorkbookFacts:     facts,
		TemplateClassPlan: *plan,
	})
	if !ok {
		t.Fatal("DraftOrganismExecutionRequest ok=false")
	}
	raw, err := json.Marshal(draft)
	if err != nil {
		t.Fatalf("Marshal draft: %v", err)
	}
	var req OrganismExecutionRequest
	if err := json.Unmarshal(raw, &req); err != nil {
		t.Fatalf("Unmarshal draft: %v", err)
	}

	result, err := OrchestrateOrganism(req)
	if err != nil {
		t.Fatalf("OrchestrateOrganism: %v", err)
	}
	if !result.TemplateClassEvaluation.Pass || !result.OrganismVerification.Pass {
		t.Fatalf("organism evidence failed: eval=%+v verification=%+v", result.TemplateClassEvaluation, result.OrganismVerification)
	}
	if len(result.StepResults) < 4 || len(result.StepResults[3].Verification.HighlightedRows) != 1 || result.StepResults[3].Verification.HighlightedRows[0] != 3 {
		t.Fatalf("highlight evidence=%+v", result.StepResults)
	}

	outputHandle, err := excelize.OpenFile(outputFile)
	if err != nil {
		t.Fatalf("Open output: %v", err)
	}
	defer func() { _ = outputHandle.Close() }()
	if got, err := outputHandle.GetCellValue("Maintenance", "A3"); err != nil || got != "M-2" {
		t.Fatalf("Maintenance!A3=%q err=%v want M-2", got, err)
	}
	if got, err := outputHandle.GetCellValue("MaintenanceSummary", "B2"); err != nil || got != "2" {
		t.Fatalf("MaintenanceSummary!B2=%q err=%v want 2", got, err)
	}
	if got, err := outputHandle.GetCellFormula("Maintenance", "E3"); err != nil || got != "=IF(D3>=4,1,0)" {
		t.Fatalf("Maintenance!E3 formula=%q err=%v want =IF(D3>=4,1,0)", got, err)
	}
}

func TestOrchestrateOrganismExecutesRequestCompilerComplianceActionDraft(t *testing.T) {
	tempDir := t.TempDir()
	inputFile := filepath.Join(tempDir, "compliance.xlsx")
	outputFile := filepath.Join(tempDir, "compliance-organism.xlsx")

	file := excelize.NewFile()
	defer func() { _ = file.Close() }()
	defaultSheet := file.GetSheetName(0)
	if err := file.SetSheetName(defaultSheet, "Compliance"); err != nil {
		t.Fatalf("SetSheetName: %v", err)
	}
	if err := file.SetSheetRow("Compliance", "A1", &[]any{"action_id", "owner", "status", "days_until_due", "review_required", "action_count"}); err != nil {
		t.Fatalf("SetSheetRow header: %v", err)
	}
	if err := file.SetSheetRow("Compliance", "A2", &[]any{"C-1", "Ops", "open", 5, nil, 1}); err != nil {
		t.Fatalf("SetSheetRow row: %v", err)
	}
	if err := file.SetCellFormula("Compliance", "E2", "=IF(D2<0,1,0)"); err != nil {
		t.Fatalf("SetCellFormula(E2): %v", err)
	}
	if err := file.SaveAs(inputFile); err != nil {
		t.Fatalf("SaveAs: %v", err)
	}

	facts, err := runtimeinspect.InspectWorkbookFacts(inputFile)
	if err != nil {
		t.Fatalf("InspectWorkbookFacts: %v", err)
	}
	requestText := "Append compliance action, validate action status, extend review-required formulas, flag overdue actions, summarize compliance status, protect review formulas, and generate a printable action register"
	plan := requestcompiler.TemplateClassPlanHintForRequest(requestText)
	if plan == nil {
		t.Fatal("missing plan")
	}
	draft, ok := requestcompiler.DraftOrganismExecutionRequest(requestcompiler.OrganismDraftInput{
		ScenarioID:        "compliance-action-organism-draft-integration",
		RequestText:       requestText,
		InputFile:         inputFile,
		OutputFile:        outputFile,
		WorkbookFacts:     facts,
		TemplateClassPlan: *plan,
	})
	if !ok {
		t.Fatal("DraftOrganismExecutionRequest ok=false")
	}
	raw, err := json.Marshal(draft)
	if err != nil {
		t.Fatalf("Marshal draft: %v", err)
	}
	var req OrganismExecutionRequest
	if err := json.Unmarshal(raw, &req); err != nil {
		t.Fatalf("Unmarshal draft: %v", err)
	}

	result, err := OrchestrateOrganism(req)
	if err != nil {
		t.Fatalf("OrchestrateOrganism: %v", err)
	}
	if !result.TemplateClassEvaluation.Pass || !result.OrganismVerification.Pass {
		t.Fatalf("organism evidence failed: eval=%+v verification=%+v", result.TemplateClassEvaluation, result.OrganismVerification)
	}
	if len(result.StepResults) < 4 || len(result.StepResults[3].Verification.HighlightedRows) != 1 || result.StepResults[3].Verification.HighlightedRows[0] != 3 {
		t.Fatalf("highlight evidence=%+v", result.StepResults)
	}

	outputHandle, err := excelize.OpenFile(outputFile)
	if err != nil {
		t.Fatalf("Open output: %v", err)
	}
	defer func() { _ = outputHandle.Close() }()
	if got, err := outputHandle.GetCellValue("Compliance", "A3"); err != nil || got != "C-2" {
		t.Fatalf("Compliance!A3=%q err=%v want C-2", got, err)
	}
	if got, err := outputHandle.GetCellValue("ComplianceSummary", "B2"); err != nil || got != "2" {
		t.Fatalf("ComplianceSummary!B2=%q err=%v want 2", got, err)
	}
	if got, err := outputHandle.GetCellFormula("Compliance", "E3"); err != nil || got != "=IF(D3<0,1,0)" {
		t.Fatalf("Compliance!E3 formula=%q err=%v want =IF(D3<0,1,0)", got, err)
	}
	if got, err := outputHandle.GetCellValue("ComplianceRegister", "A1"); err != nil || got != "Compliance Action Register" {
		t.Fatalf("ComplianceRegister!A1=%q err=%v want Compliance Action Register", got, err)
	}
}

func TestOrchestrateOrganismExecutesRequestCompilerSafetyComplianceDraft(t *testing.T) {
	tempDir := t.TempDir()
	inputFile := filepath.Join(tempDir, "safety.xlsx")
	outputFile := filepath.Join(tempDir, "safety-organism.xlsx")

	file := excelize.NewFile()
	defer func() { _ = file.Close() }()
	defaultSheet := file.GetSheetName(0)
	if err := file.SetSheetName(defaultSheet, "Safety"); err != nil {
		t.Fatalf("SetSheetName: %v", err)
	}
	if err := file.SetSheetRow("Safety", "A1", &[]any{"check_id", "area", "status", "risk_score", "completed", "action_required"}); err != nil {
		t.Fatalf("SetSheetRow header: %v", err)
	}
	if err := file.SetSheetRow("Safety", "A2", &[]any{"S-1", "Warehouse", "complete", 2, 1}); err != nil {
		t.Fatalf("SetSheetRow row: %v", err)
	}
	if err := file.SetCellFormula("Safety", "F2", "=IF(D2>=4,1,0)"); err != nil {
		t.Fatalf("SetCellFormula(F2): %v", err)
	}
	if err := file.SaveAs(inputFile); err != nil {
		t.Fatalf("SaveAs: %v", err)
	}

	facts, err := runtimeinspect.InspectWorkbookFacts(inputFile)
	if err != nil {
		t.Fatalf("InspectWorkbookFacts: %v", err)
	}
	requestText := "Append safety compliance check, validate safety status, extend action-required formulas, flag high risk safety items, summarize completion by area, protect action formulas, and generate a printable safety report"
	plan := requestcompiler.TemplateClassPlanHintForRequest(requestText)
	if plan == nil {
		t.Fatal("missing plan")
	}
	draft, ok := requestcompiler.DraftOrganismExecutionRequest(requestcompiler.OrganismDraftInput{
		ScenarioID:        "safety-compliance-organism-draft-integration",
		RequestText:       requestText,
		InputFile:         inputFile,
		OutputFile:        outputFile,
		WorkbookFacts:     facts,
		TemplateClassPlan: *plan,
	})
	if !ok {
		t.Fatal("DraftOrganismExecutionRequest ok=false")
	}
	raw, err := json.Marshal(draft)
	if err != nil {
		t.Fatalf("Marshal draft: %v", err)
	}
	var req OrganismExecutionRequest
	if err := json.Unmarshal(raw, &req); err != nil {
		t.Fatalf("Unmarshal draft: %v", err)
	}

	result, err := OrchestrateOrganism(req)
	if err != nil {
		t.Fatalf("OrchestrateOrganism: %v", err)
	}
	if !result.TemplateClassEvaluation.Pass || !result.OrganismVerification.Pass {
		t.Fatalf("organism evidence failed: eval=%+v verification=%+v", result.TemplateClassEvaluation, result.OrganismVerification)
	}
	if len(result.StepResults) < 4 || len(result.StepResults[3].Verification.HighlightedRows) != 1 || result.StepResults[3].Verification.HighlightedRows[0] != 3 {
		t.Fatalf("highlight evidence=%+v", result.StepResults)
	}

	outputHandle, err := excelize.OpenFile(outputFile)
	if err != nil {
		t.Fatalf("Open output: %v", err)
	}
	defer func() { _ = outputHandle.Close() }()
	if got, err := outputHandle.GetCellValue("Safety", "A3"); err != nil || got != "S-2" {
		t.Fatalf("Safety!A3=%q err=%v want S-2", got, err)
	}
	if got, err := outputHandle.GetCellValue("SafetySummary", "B2"); err != nil || got != "1" {
		t.Fatalf("SafetySummary!B2=%q err=%v want 1", got, err)
	}
	if got, err := outputHandle.GetCellFormula("Safety", "F3"); err != nil || got != "=IF(D3>=4,1,0)" {
		t.Fatalf("Safety!F3 formula=%q err=%v want =IF(D3>=4,1,0)", got, err)
	}
	if got, err := outputHandle.GetCellValue("SafetyReport", "A1"); err != nil || got != "Safety Compliance Report" {
		t.Fatalf("SafetyReport!A1=%q err=%v want Safety Compliance Report", got, err)
	}
}

func TestOrchestrateOrganismExecutesRequestCompilerLoanDraft(t *testing.T) {
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
	plan := requestcompiler.TemplateClassPlanHintForRequest(requestText)
	if plan == nil {
		t.Fatal("missing plan")
	}
	draft, ok := requestcompiler.DraftOrganismExecutionRequest(requestcompiler.OrganismDraftInput{
		ScenarioID:        "loan-organism-draft-integration",
		RequestText:       requestText,
		InputFile:         inputFile,
		OutputFile:        outputFile,
		WorkbookFacts:     facts,
		TemplateClassPlan: *plan,
	})
	if !ok {
		t.Fatal("DraftOrganismExecutionRequest ok=false")
	}
	raw, err := json.Marshal(draft)
	if err != nil {
		t.Fatalf("Marshal draft: %v", err)
	}
	var req OrganismExecutionRequest
	if err := json.Unmarshal(raw, &req); err != nil {
		t.Fatalf("Unmarshal draft: %v", err)
	}

	result, err := OrchestrateOrganism(req)
	if err != nil {
		t.Fatalf("OrchestrateOrganism: %v", err)
	}
	if !result.TemplateClassEvaluation.Pass || !result.OrganismVerification.Pass {
		t.Fatalf("organism evidence failed: eval=%+v verification=%+v", result.TemplateClassEvaluation, result.OrganismVerification)
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

func cellRowForDraftIntegrationTest(row int) string {
	return strconv.Itoa(row)
}
