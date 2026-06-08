package requestcompiler

import (
	"path/filepath"
	goruntime "runtime"
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

func organismExecutionRequestSchemaPathForTest() string {
	_, file, _, ok := goruntime.Caller(0)
	if !ok {
		return filepath.Join("contracts", "requests", "organism_execution_request.schema.json")
	}
	return filepath.Join(resolvePackageRoot("", file), "contracts", "requests", "organism_execution_request.schema.json")
}
