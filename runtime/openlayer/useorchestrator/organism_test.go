package useorchestrator

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/xuri/excelize/v2"
)

func TestOrchestrateOrganismExecutesExplicitInvoiceSteps(t *testing.T) {
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

	result, err := OrchestrateOrganism(OrganismExecutionRequest{
		ScenarioID:  "invoice-organism-openlayer",
		RequestText: "Add invoice line items, extend totals, protect formulas, and create a printable invoice",
		InputFile:   inputFile,
		OutputFile:  outputFile,
		Steps: []OrganismExecutionStep{
			{
				AtomID:               "append_structured_rows",
				CompositionKind:      "structured_row_append",
				SourceSheet:          "LineItems",
				IncludeSourceColumns: []string{"sku", "quantity", "unit_price"},
				Values: []CellValue{
					{Cell: "sku", Value: "B002"},
					{Cell: "quantity", Value: 3},
					{Cell: "unit_price", Value: 15},
				},
			},
			{
				AtomID:           "extend_table_formulas",
				CompositionKind:  "formula_extension",
				SourceSheet:      "LineItems",
				FormulaSourceRow: 2,
				TargetRows:       []int{3},
				FormulaColumns:   []string{"D", "E"},
			},
			{
				AtomID:          "add_data_validation",
				CompositionKind: "data_validation",
				SourceSheet:     "LineItems",
				ValidationRule: &DataValidationRule{
					Ranges:        []string{"A2:A10"},
					RuleType:      "list",
					AllowedValues: []string{"A001", "B002"},
					AllowBlank:    false,
				},
			},
			{
				AtomID:          "protect_formula_cells",
				CompositionKind: "formula_protection",
				SourceSheet:     "LineItems",
				ProtectionRule: &FormulaProtectionRule{
					FormulaRanges: []string{"D2:E3"},
					InputRanges:   []string{"A2:C10"},
				},
			},
			{
				AtomID:          "generate_printable_form",
				CompositionKind: "printable_form",
				SourceSheet:     "LineItems",
				TargetSheet:     "InvoicePrint",
				FormTitle:       "Invoice",
				PrintArea:       "A1:E8",
				FieldBindings: []FormFieldBinding{
					{Label: "First SKU", SourceSheet: "LineItems", SourceCell: "A2", LabelCell: "A2", ValueCell: "B2"},
				},
				TableBinding: &FormTableBinding{
					SourceSheet:   "LineItems",
					SourceColumns: []string{"sku", "quantity", "unit_price", "line_total", "tax"},
					HeaderStart:   "A4",
					DataStart:     "A5",
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("OrchestrateOrganism: %v", err)
	}
	if !result.TemplateClassEvaluation.Pass {
		t.Fatalf("template class evaluation failed: %+v", result.TemplateClassEvaluation)
	}
	if !result.OrganismVerification.Pass {
		t.Fatalf("organism verification failed: %+v", result.OrganismVerification)
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

func TestOrchestrateOrganismRejectsStepsOutsideTemplateClassPlan(t *testing.T) {
	_, err := OrchestrateOrganism(OrganismExecutionRequest{
		ScenarioID:  "invoice-organism-mismatch",
		RequestText: "Build an invoice line billing template",
		InputFile:   "/tmp/input.xlsx",
		OutputFile:  "/tmp/output.xlsx",
		Steps: []OrganismExecutionStep{
			{
				AtomID:          "copy_period_sheet",
				CompositionKind: "period_copy",
				SourceSheet:     "Jan",
				TargetSheet:     "Feb",
			},
		},
	})
	if err == nil {
		t.Fatal("OrchestrateOrganism err=nil want mismatch error")
	}
	if !strings.Contains(err.Error(), "does not match template class plan") {
		t.Fatalf("err=%q want template class plan mismatch", err)
	}
}
