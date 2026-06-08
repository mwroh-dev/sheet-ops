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

func TestRunExtendTableFormulasPreservesSourceAndVerifiesFormulas(t *testing.T) {
	tempDir := t.TempDir()
	inputFile := filepath.Join(tempDir, "line-items.xlsx")
	outputFile := filepath.Join(tempDir, "line-items-output.xlsx")

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
	row3 := []any{"B002", 3, 15}
	if err := file.SetSheetRow("LineItems", "A3", &row3); err != nil {
		t.Fatalf("SetSheetRow(row3): %v", err)
	}
	if err := file.SaveAs(inputFile); err != nil {
		t.Fatalf("SaveAs: %v", err)
	}

	ir := compiler.WorkbookOperationIR{
		ExecutionKind:    "composition",
		CompositionKind:  "formula_extension",
		OperationFamily:  "extend_table_formulas",
		SourceSheet:      "LineItems",
		FormulaSourceRow: 2,
		TargetRows:       []int{3},
		FormulaColumns:   []string{"D", "E"},
		PreserveOriginal: true,
	}

	result, err := RunWorkbookOperation(ir, inputFile, outputFile)
	if err != nil {
		t.Fatalf("RunWorkbookOperation: %v", err)
	}
	if result.OperationFamily != "extend_table_formulas" {
		t.Fatalf("operation_family=%q want extend_table_formulas", result.OperationFamily)
	}
	if len(result.FormulaCells) != 2 {
		t.Fatalf("formula cells=%v want 2", result.FormulaCells)
	}
	if result.SourceSHA256Before != result.SourceSHA256After {
		t.Fatal("source hash changed during execution")
	}

	outputHandle, err := excelize.OpenFile(outputFile)
	if err != nil {
		t.Fatalf("Open output: %v", err)
	}
	defer func() { _ = outputHandle.Close() }()
	if got, err := outputHandle.GetCellFormula("LineItems", "D3"); err != nil || got != "=B3*C3" {
		t.Fatalf("LineItems!D3 formula=%q err=%v want =B3*C3", got, err)
	}
	if got, err := outputHandle.GetCellFormula("LineItems", "E3"); err != nil || got != "=D3*0.1" {
		t.Fatalf("LineItems!E3 formula=%q err=%v want =D3*0.1", got, err)
	}

	verification, err := runtimeverify.VerifyWorkbookOperation(ir, inputFile, outputFile, result.SourceSHA256Before, result.SourceSHA256After)
	if err != nil {
		t.Fatalf("VerifyWorkbookOperation: %v", err)
	}
	if !verification.Pass {
		t.Fatalf("verification failed: %+v", verification)
	}
}

func TestRunAddDataValidationPreservesSourceAndVerifiesRule(t *testing.T) {
	tempDir := t.TempDir()
	inputFile := filepath.Join(tempDir, "invoice.xlsx")
	outputFile := filepath.Join(tempDir, "invoice-output.xlsx")

	file := excelize.NewFile()
	defer func() { _ = file.Close() }()
	defaultSheet := file.GetSheetName(0)
	if err := file.SetSheetName(defaultSheet, "LineItems"); err != nil {
		t.Fatalf("SetSheetName: %v", err)
	}
	header := []any{"sku", "quantity", "unit_price", "status"}
	if err := file.SetSheetRow("LineItems", "A1", &header); err != nil {
		t.Fatalf("SetSheetRow(header): %v", err)
	}
	row := []any{"A001", 2, 10, "draft"}
	if err := file.SetSheetRow("LineItems", "A2", &row); err != nil {
		t.Fatalf("SetSheetRow(row): %v", err)
	}
	if err := file.SaveAs(inputFile); err != nil {
		t.Fatalf("SaveAs: %v", err)
	}

	ir := compiler.WorkbookOperationIR{
		ExecutionKind:   "composition",
		CompositionKind: "data_validation",
		OperationFamily: "add_data_validation",
		SourceSheet:     "LineItems",
		ValidationRule: &compiler.DataValidationRule{
			Ranges:        []string{"D2:D10"},
			RuleType:      "list",
			AllowedValues: []string{"draft", "sent", "paid"},
			AllowBlank:    false,
		},
		PreserveOriginal: true,
	}

	result, err := RunWorkbookOperation(ir, inputFile, outputFile)
	if err != nil {
		t.Fatalf("RunWorkbookOperation: %v", err)
	}
	if result.OperationFamily != "add_data_validation" {
		t.Fatalf("operation_family=%q want add_data_validation", result.OperationFamily)
	}
	if result.SourceSHA256Before != result.SourceSHA256After {
		t.Fatal("source hash changed during execution")
	}

	verification, err := runtimeverify.VerifyWorkbookOperation(ir, inputFile, outputFile, result.SourceSHA256Before, result.SourceSHA256After)
	if err != nil {
		t.Fatalf("VerifyWorkbookOperation: %v", err)
	}
	if !verification.Pass {
		t.Fatalf("verification failed: %+v", verification)
	}
}

func TestRunProtectFormulaCellsPreservesSourceAndVerifiesProtection(t *testing.T) {
	tempDir := t.TempDir()
	inputFile := filepath.Join(tempDir, "invoice.xlsx")
	outputFile := filepath.Join(tempDir, "invoice-output.xlsx")

	file := excelize.NewFile()
	defer func() { _ = file.Close() }()
	defaultSheet := file.GetSheetName(0)
	if err := file.SetSheetName(defaultSheet, "LineItems"); err != nil {
		t.Fatalf("SetSheetName: %v", err)
	}
	header := []any{"sku", "quantity", "unit_price", "line_total"}
	if err := file.SetSheetRow("LineItems", "A1", &header); err != nil {
		t.Fatalf("SetSheetRow(header): %v", err)
	}
	row := []any{"A001", 2, 10}
	if err := file.SetSheetRow("LineItems", "A2", &row); err != nil {
		t.Fatalf("SetSheetRow(row): %v", err)
	}
	if err := file.SetCellFormula("LineItems", "D2", "=B2*C2"); err != nil {
		t.Fatalf("SetCellFormula(D2): %v", err)
	}
	if err := file.SaveAs(inputFile); err != nil {
		t.Fatalf("SaveAs: %v", err)
	}

	ir := compiler.WorkbookOperationIR{
		ExecutionKind:   "composition",
		CompositionKind: "formula_protection",
		OperationFamily: "protect_formula_cells",
		SourceSheet:     "LineItems",
		ProtectionRule: &compiler.FormulaProtectionRule{
			FormulaRanges: []string{"D2"},
			InputRanges:   []string{"A2:C10"},
		},
		PreserveOriginal: true,
	}

	result, err := RunWorkbookOperation(ir, inputFile, outputFile)
	if err != nil {
		t.Fatalf("RunWorkbookOperation: %v", err)
	}
	if result.OperationFamily != "protect_formula_cells" {
		t.Fatalf("operation_family=%q want protect_formula_cells", result.OperationFamily)
	}
	if result.SourceSHA256Before != result.SourceSHA256After {
		t.Fatal("source hash changed during execution")
	}

	outputHandle, err := excelize.OpenFile(outputFile)
	if err != nil {
		t.Fatalf("Open output: %v", err)
	}
	defer func() { _ = outputHandle.Close() }()
	protection, err := outputHandle.GetSheetProtection("LineItems")
	if err != nil {
		t.Fatalf("GetSheetProtection: %v", err)
	}
	if !protection.SelectLockedCells || !protection.SelectUnlockedCells {
		t.Fatalf("sheet protection not enabled enough: %+v", protection)
	}
	if locked, err := testCellLocked(outputHandle, "LineItems", "D2"); err != nil || !locked {
		t.Fatalf("D2 locked=%v err=%v want true", locked, err)
	}
	if locked, err := testCellLocked(outputHandle, "LineItems", "A2"); err != nil || locked {
		t.Fatalf("A2 locked=%v err=%v want false", locked, err)
	}

	verification, err := runtimeverify.VerifyWorkbookOperation(ir, inputFile, outputFile, result.SourceSHA256Before, result.SourceSHA256After)
	if err != nil {
		t.Fatalf("VerifyWorkbookOperation: %v", err)
	}
	if !verification.Pass {
		t.Fatalf("verification failed: %+v", verification)
	}
}

func testCellLocked(file *excelize.File, sheet, cell string) (bool, error) {
	styleID, err := file.GetCellStyle(sheet, cell)
	if err != nil {
		return false, err
	}
	style, err := file.GetStyle(styleID)
	if err != nil {
		return false, err
	}
	if style.Protection == nil {
		return true, nil
	}
	return style.Protection.Locked, nil
}
