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
