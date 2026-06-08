package execute

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mwroh/sheet-ops/runtime/compiler"
	"github.com/xuri/excelize/v2"
)

func RunAppendStructuredRows(ir compiler.WorkbookOperationIR, inputWorkbook, outputWorkbook string) (ExecutionResult, error) {
	if err := validateAppendStructuredRowsBoundary(ir, inputWorkbook, outputWorkbook); err != nil {
		return ExecutionResult{}, err
	}
	sourceBefore, err := hashFile(inputWorkbook)
	if err != nil {
		return ExecutionResult{}, err
	}

	file, err := excelize.OpenFile(inputWorkbook)
	if err != nil {
		return ExecutionResult{}, err
	}
	defer func() { _ = file.Close() }()

	rows, err := file.GetRows(ir.SourceSheet)
	if err != nil {
		return ExecutionResult{}, fmt.Errorf("read source sheet %q: %w", ir.SourceSheet, err)
	}
	if len(rows) == 0 {
		return ExecutionResult{}, fmt.Errorf("source sheet %q is empty", ir.SourceSheet)
	}
	headerIndex := stringHeaderIndex(rows[0])
	for _, column := range ir.IncludeSourceColumns {
		if _, ok := headerIndex[column]; !ok {
			return ExecutionResult{}, fmt.Errorf("source column %q missing", column)
		}
	}

	appendRows, err := appendRowsFromCellValues(ir.IncludeSourceColumns, ir.Values)
	if err != nil {
		return ExecutionResult{}, err
	}
	startRow := len(rows) + 1
	writtenCells := []string{}
	for rowOffset, rowValues := range appendRows {
		excelRow := startRow + rowOffset
		for columnName, value := range rowValues {
			columnIndex := headerIndex[columnName] + 1
			cell, _ := excelize.CoordinatesToCellName(columnIndex, excelRow)
			if err := file.SetCellValue(ir.SourceSheet, cell, value); err != nil {
				return ExecutionResult{}, err
			}
			writtenCells = append(writtenCells, ir.SourceSheet+"!"+cell)
		}
	}

	if outputDir := filepath.Dir(outputWorkbook); outputDir != "." {
		if err := os.MkdirAll(outputDir, 0o755); err != nil {
			return ExecutionResult{}, err
		}
	}
	if err := file.SaveAs(outputWorkbook); err != nil {
		return ExecutionResult{}, err
	}
	sourceAfter, err := hashFile(inputWorkbook)
	if err != nil {
		return ExecutionResult{}, err
	}

	return ExecutionResult{
		OperationFamily:    ir.OperationFamily,
		OutputWorkbook:     outputWorkbook,
		SummarySheet:       ir.SourceSheet,
		SummaryRows:        len(appendRows),
		WrittenCells:       writtenCells,
		SourceSHA256Before: sourceBefore,
		SourceSHA256After:  sourceAfter,
	}, nil
}

func validateAppendStructuredRowsBoundary(ir compiler.WorkbookOperationIR, inputWorkbook, outputWorkbook string) error {
	if inputWorkbook == "" {
		return fmt.Errorf("input workbook must not be empty")
	}
	if outputWorkbook == "" {
		return fmt.Errorf("output workbook must not be empty")
	}
	if !ir.PreserveOriginal {
		return fmt.Errorf("append_structured_rows execution requires preserve_original=true")
	}
	if ir.SourceSheet == "" {
		return fmt.Errorf("source sheet must not be empty")
	}
	if len(ir.IncludeSourceColumns) == 0 {
		return fmt.Errorf("include source columns must not be empty")
	}
	if len(ir.Values) == 0 {
		return fmt.Errorf("values must not be empty")
	}
	same, err := sameWorkbookPath(inputWorkbook, outputWorkbook)
	if err != nil {
		return err
	}
	if same {
		return fmt.Errorf("output workbook must differ from input workbook")
	}
	return nil
}

func appendRowsFromCellValues(columns []string, values []compiler.CellValue) ([]map[string]any, error) {
	if len(values)%len(columns) != 0 {
		return nil, fmt.Errorf("values count %d must be a multiple of column count %d", len(values), len(columns))
	}
	rows := make([]map[string]any, 0, len(values)/len(columns))
	for offset := 0; offset < len(values); offset += len(columns) {
		row := map[string]any{}
		for columnOffset, column := range columns {
			value := values[offset+columnOffset]
			if value.Cell != column {
				return nil, fmt.Errorf("value %d targets column %q want %q", offset+columnOffset, value.Cell, column)
			}
			row[column] = value.Value
		}
		rows = append(rows, row)
	}
	return rows, nil
}
