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
