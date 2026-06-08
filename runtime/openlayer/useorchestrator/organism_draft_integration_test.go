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
