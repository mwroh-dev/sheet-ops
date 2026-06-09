package verify

import (
	"fmt"

	"github.com/mwroh/sheet-ops/runtime/compiler"
	"github.com/xuri/excelize/v2"
)

func VerifyAppendStructuredRows(ir compiler.WorkbookOperationIR, inputWorkbook, outputWorkbook, sourceSHA256Before, sourceSHA256After string) (VerificationResult, error) {
	if sourceSHA256Before == "" || sourceSHA256After == "" {
		return appendStructuredRowsFailure(ir, outputWorkbook, "execution hash evidence is missing"), nil
	}
	if sourceSHA256Before != sourceSHA256After {
		return appendStructuredRowsFailure(ir, outputWorkbook, "source workbook changed during execution"), nil
	}
	currentHash, err := hashFile(inputWorkbook)
	if err != nil {
		return VerificationResult{}, err
	}
	if currentHash != sourceSHA256Before {
		return appendStructuredRowsFailure(ir, outputWorkbook, "source workbook hash differs from execution evidence"), nil
	}

	inputFile, err := excelize.OpenFile(inputWorkbook)
	if err != nil {
		return VerificationResult{}, err
	}
	defer func() { _ = inputFile.Close() }()
	outputFile, err := excelize.OpenFile(outputWorkbook)
	if err != nil {
		return VerificationResult{}, err
	}
	defer func() { _ = outputFile.Close() }()

	inputRows, err := inputFile.GetRows(ir.SourceSheet)
	if err != nil {
		return VerificationResult{}, err
	}
	outputRows, err := outputFile.GetRows(ir.SourceSheet)
	if err != nil {
		return VerificationResult{}, err
	}
	if len(outputRows) < len(inputRows) {
		return appendStructuredRowsFailure(ir, outputWorkbook, fmt.Sprintf("output row count=%d less than input row count=%d", len(outputRows), len(inputRows))), nil
	}
	if err := compareStringRows(ir.SourceSheet, inputRows, outputRows[:len(inputRows)]); err != nil {
		return appendStructuredRowsFailure(ir, outputWorkbook, err.Error()), nil
	}

	headerIndex := stringHeaderIndex(inputRows[0])
	appendRows, err := appendRowsFromCellValues(ir.IncludeSourceColumns, ir.Values)
	if err != nil {
		return VerificationResult{}, err
	}
	if got, want := len(outputRows)-len(inputRows), len(appendRows); got != want {
		return appendStructuredRowsFailure(ir, outputWorkbook, fmt.Sprintf("appended row count=%d want %d", got, want)), nil
	}

	writtenCells := []string{}
	for rowOffset, rowValues := range appendRows {
		excelRow := len(inputRows) + 1 + rowOffset
		for _, column := range ir.IncludeSourceColumns {
			columnIndex, ok := headerIndex[column]
			if !ok {
				return appendStructuredRowsFailure(ir, outputWorkbook, fmt.Sprintf("source column %q missing", column)), nil
			}
			cell, _ := excelize.CoordinatesToCellName(columnIndex+1, excelRow)
			got, err := outputFile.GetCellValue(ir.SourceSheet, cell)
			if err != nil {
				return VerificationResult{}, err
			}
			want := expectedWriteValueString(rowValues[column])
			if got != want {
				return appendStructuredRowsFailure(ir, outputWorkbook, fmt.Sprintf("%s!%s=%q want %q", ir.SourceSheet, cell, got, want)), nil
			}
			writtenCells = append(writtenCells, ir.SourceSheet+"!"+cell)
		}
	}

	return VerificationResult{
		Pass:                true,
		OperationFamily:     ir.OperationFamily,
		OutputWorkbook:      outputWorkbook,
		SummarySheet:        ir.SourceSheet,
		SummaryRows:         len(appendRows),
		WrittenCells:        writtenCells,
		SourceSHA256Checked: true,
		Reasons:             []string{"output workbook contains appended structured rows and source workbook is unchanged"},
	}, nil
}

func appendStructuredRowsFailure(ir compiler.WorkbookOperationIR, outputWorkbook string, reason string) VerificationResult {
	return VerificationResult{
		Pass:            false,
		OperationFamily: ir.OperationFamily,
		OutputWorkbook:  outputWorkbook,
		SummarySheet:    ir.SourceSheet,
		Reasons:         []string{reason},
	}
}

func appendRowsFromCellValues(columns []string, values []compiler.CellValue) ([]map[string]any, error) {
	if len(columns) == 0 {
		return nil, fmt.Errorf("include source columns must not be empty")
	}
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

func stringHeaderIndex(header []string) map[string]int {
	index := make(map[string]int, len(header))
	for idx, value := range header {
		index[value] = idx
	}
	return index
}
