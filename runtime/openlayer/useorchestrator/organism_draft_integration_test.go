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

func cellRowForDraftIntegrationTest(row int) string {
	return strconv.Itoa(row)
}
