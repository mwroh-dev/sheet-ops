package execute

import (
	"path/filepath"
	"testing"

	"github.com/mwroh/sheet-ops/runtime/compiler"
	runtimeverify "github.com/mwroh/sheet-ops/runtime/verify"
	"github.com/xuri/excelize/v2"
)

func TestRunWriteValuesPreservesSourceAndVerifiesCells(t *testing.T) {
	tempDir := t.TempDir()
	inputFile := filepath.Join(tempDir, "input.xlsx")
	outputFile := filepath.Join(tempDir, "output.xlsx")

	file := excelize.NewFile()
	defer func() { _ = file.Close() }()
	defaultSheet := file.GetSheetName(0)
	if err := file.SetSheetName(defaultSheet, "Orders"); err != nil {
		t.Fatalf("SetSheetName: %v", err)
	}
	if err := file.SetCellValue("Orders", "A1", "order_id"); err != nil {
		t.Fatalf("SetCellValue(A1): %v", err)
	}
	if err := file.SetCellValue("Orders", "A2", "A001"); err != nil {
		t.Fatalf("SetCellValue(A2): %v", err)
	}
	if err := file.SaveAs(inputFile); err != nil {
		t.Fatalf("SaveAs: %v", err)
	}

	ir := compiler.WorkbookOperationIR{
		ExecutionKind:    "primitive",
		OperationFamily:  "write_values",
		TargetSheet:      "Notes",
		PreserveOriginal: true,
		Values: []compiler.CellValue{
			{Cell: "A1", Value: "status"},
			{Cell: "B1", Value: "ready"},
			{Cell: "C1", Value: true},
			{Cell: "D1", Value: nil},
		},
	}
	result, err := RunWorkbookOperation(ir, inputFile, outputFile)
	if err != nil {
		t.Fatalf("RunWorkbookOperation: %v", err)
	}
	if result.OperationFamily != "write_values" {
		t.Fatalf("operation_family=%q want write_values", result.OperationFamily)
	}
	if result.WrittenCells == nil || len(result.WrittenCells) != 4 {
		t.Fatalf("written_cells=%v want 4 cells", result.WrittenCells)
	}

	verification, err := runtimeverify.VerifyWorkbookOperation(ir, inputFile, outputFile, result.SourceSHA256Before, result.SourceSHA256After)
	if err != nil {
		t.Fatalf("VerifyWorkbookOperation: %v", err)
	}
	if !verification.Pass {
		t.Fatalf("verification failed: %+v", verification)
	}

	inputHandle, err := excelize.OpenFile(inputFile)
	if err != nil {
		t.Fatalf("Open input: %v", err)
	}
	defer func() { _ = inputHandle.Close() }()
	gotSource, err := inputHandle.GetCellValue("Orders", "A2")
	if err != nil {
		t.Fatalf("GetCellValue input Orders!A2: %v", err)
	}
	if gotSource != "A001" {
		t.Fatalf("input Orders!A2=%q want A001", gotSource)
	}

	outputHandle, err := excelize.OpenFile(outputFile)
	if err != nil {
		t.Fatalf("Open output: %v", err)
	}
	defer func() { _ = outputHandle.Close() }()
	got, err := outputHandle.GetCellValue("Notes", "B1")
	if err != nil {
		t.Fatalf("GetCellValue output Notes!B1: %v", err)
	}
	if got != "ready" {
		t.Fatalf("output Notes!B1=%q want ready", got)
	}
	gotBool, err := outputHandle.GetCellValue("Notes", "C1")
	if err != nil {
		t.Fatalf("GetCellValue output Notes!C1: %v", err)
	}
	if gotBool != "TRUE" {
		t.Fatalf("output Notes!C1=%q want TRUE", gotBool)
	}
	gotBlank, err := outputHandle.GetCellValue("Notes", "D1")
	if err != nil {
		t.Fatalf("GetCellValue output Notes!D1: %v", err)
	}
	if gotBlank != "" {
		t.Fatalf("output Notes!D1=%q want blank", gotBlank)
	}
}

func TestRunAppendStructuredRowsPreservesSourceAndVerifiesRows(t *testing.T) {
	tempDir := t.TempDir()
	inputFile := filepath.Join(tempDir, "input.xlsx")
	outputFile := filepath.Join(tempDir, "output.xlsx")

	file := excelize.NewFile()
	defer func() { _ = file.Close() }()
	defaultSheet := file.GetSheetName(0)
	if err := file.SetSheetName(defaultSheet, "LineItems"); err != nil {
		t.Fatalf("SetSheetName: %v", err)
	}
	headers := []any{"sku", "quantity", "unit_price"}
	if err := file.SetSheetRow("LineItems", "A1", &headers); err != nil {
		t.Fatalf("SetSheetRow(header): %v", err)
	}
	row := []any{"A001", 2, 10}
	if err := file.SetSheetRow("LineItems", "A2", &row); err != nil {
		t.Fatalf("SetSheetRow(row): %v", err)
	}
	if err := file.SaveAs(inputFile); err != nil {
		t.Fatalf("SaveAs: %v", err)
	}

	ir := compiler.WorkbookOperationIR{
		ExecutionKind:        "composition",
		CompositionKind:      "structured_row_append",
		OperationFamily:      "append_structured_rows",
		SourceSheet:          "LineItems",
		IncludeSourceColumns: []string{"sku", "quantity", "unit_price"},
		PreserveOriginal:     true,
		Values: []compiler.CellValue{
			{Cell: "sku", Value: "B002"},
			{Cell: "quantity", Value: 3},
			{Cell: "unit_price", Value: 15},
			{Cell: "sku", Value: "C003"},
			{Cell: "quantity", Value: 1},
			{Cell: "unit_price", Value: 25},
		},
	}
	result, err := RunWorkbookOperation(ir, inputFile, outputFile)
	if err != nil {
		t.Fatalf("RunWorkbookOperation: %v", err)
	}
	if result.OperationFamily != "append_structured_rows" {
		t.Fatalf("operation_family=%q want append_structured_rows", result.OperationFamily)
	}
	if got, want := result.SummaryRows, 2; got != want {
		t.Fatalf("summary_rows=%d want appended rows %d", got, want)
	}
	if len(result.WrittenCells) != 6 {
		t.Fatalf("written_cells=%v want 6 cells", result.WrittenCells)
	}

	verification, err := runtimeverify.VerifyWorkbookOperation(ir, inputFile, outputFile, result.SourceSHA256Before, result.SourceSHA256After)
	if err != nil {
		t.Fatalf("VerifyWorkbookOperation: %v", err)
	}
	if !verification.Pass {
		t.Fatalf("verification failed: %+v", verification)
	}

	inputHandle, err := excelize.OpenFile(inputFile)
	if err != nil {
		t.Fatalf("Open input: %v", err)
	}
	defer func() { _ = inputHandle.Close() }()
	inputRows, err := inputHandle.GetRows("LineItems")
	if err != nil {
		t.Fatalf("GetRows input: %v", err)
	}
	if len(inputRows) != 2 {
		t.Fatalf("input row count=%d want unchanged 2", len(inputRows))
	}

	outputHandle, err := excelize.OpenFile(outputFile)
	if err != nil {
		t.Fatalf("Open output: %v", err)
	}
	defer func() { _ = outputHandle.Close() }()
	gotSKU, err := outputHandle.GetCellValue("LineItems", "A3")
	if err != nil {
		t.Fatalf("GetCellValue output A3: %v", err)
	}
	if gotSKU != "B002" {
		t.Fatalf("output LineItems!A3=%q want B002", gotSKU)
	}
	gotQty, err := outputHandle.GetCellValue("LineItems", "B4")
	if err != nil {
		t.Fatalf("GetCellValue output B4: %v", err)
	}
	if gotQty != "1" {
		t.Fatalf("output LineItems!B4=%q want 1", gotQty)
	}
}
