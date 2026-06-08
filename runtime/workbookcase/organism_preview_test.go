package workbookcase

import (
	"path/filepath"
	"testing"

	runtimetaskspec "github.com/mwroh/sheet-ops/runtime/taskspec"
	"github.com/xuri/excelize/v2"
)

func TestInvoiceLineItemBillingPreviewComposesAppendAndFormulaExtension(t *testing.T) {
	setRuntimeRoots(t)

	tempDir := t.TempDir()
	inputFile := filepath.Join(tempDir, "invoice.xlsx")
	appendOutput := filepath.Join(tempDir, "invoice-appended.xlsx")
	finalOutput := filepath.Join(tempDir, "invoice-final.xlsx")

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

	outputHandle, err := excelize.OpenFile(finalOutput)
	if err != nil {
		t.Fatalf("Open final output: %v", err)
	}
	defer func() { _ = outputHandle.Close() }()
	if got, err := outputHandle.GetCellValue("LineItems", "A3"); err != nil || got != "B002" {
		t.Fatalf("LineItems!A3=%q err=%v want B002", got, err)
	}
	if got, err := outputHandle.GetCellFormula("LineItems", "D3"); err != nil || got != "=B3*C3" {
		t.Fatalf("LineItems!D3 formula=%q err=%v want =B3*C3", got, err)
	}
	if got, err := outputHandle.GetCellFormula("LineItems", "E3"); err != nil || got != "=D3*0.1" {
		t.Fatalf("LineItems!E3 formula=%q err=%v want =D3*0.1", got, err)
	}
}
