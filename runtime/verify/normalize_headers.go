package verify

import (
	"fmt"

	"github.com/mwroh/sheet-ops/runtime/compiler"
	"github.com/xuri/excelize/v2"
)

func VerifyNormalizeHeaders(ir compiler.WorkbookOperationIR, inputWorkbook, outputWorkbook, sourceSHA256Before, sourceSHA256After string) (VerificationResult, error) {
	if sourceSHA256Before == "" || sourceSHA256After == "" {
		return normalizeHeadersFailure(ir, outputWorkbook, "execution hash evidence is missing"), nil
	}
	if sourceSHA256Before != sourceSHA256After {
		return normalizeHeadersFailure(ir, outputWorkbook, "source workbook changed during execution"), nil
	}
	currentHash, err := hashFile(inputWorkbook)
	if err != nil {
		return VerificationResult{}, err
	}
	if currentHash != sourceSHA256Before {
		return normalizeHeadersFailure(ir, outputWorkbook, "source workbook hash differs from execution evidence"), nil
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
	if len(inputRows) < ir.HeaderRow || len(outputRows) < ir.HeaderRow {
		return normalizeHeadersFailure(ir, outputWorkbook, fmt.Sprintf("header row %d is outside source or output sheet", ir.HeaderRow)), nil
	}
	if err := compareRowsExceptHeader(ir.SourceSheet, inputRows, outputRows, ir.HeaderRow); err != nil {
		return normalizeHeadersFailure(ir, outputWorkbook, err.Error()), nil
	}
	if err := compareFormulasExceptHeader(inputFile, outputFile, ir.SourceSheet, len(inputRows), ir.HeaderRow); err != nil {
		return normalizeHeadersFailure(ir, outputWorkbook, err.Error()), nil
	}

	headerIndex := stringHeaderIndex(inputRows[ir.HeaderRow-1])
	writtenCells := make([]string, 0, len(ir.HeaderMappings))
	for _, mapping := range ir.HeaderMappings {
		columnIndex, ok := headerIndex[mapping.From]
		if !ok {
			return normalizeHeadersFailure(ir, outputWorkbook, fmt.Sprintf("header %q missing from input row %d", mapping.From, ir.HeaderRow)), nil
		}
		cell, _ := excelize.CoordinatesToCellName(columnIndex+1, ir.HeaderRow)
		got, err := outputFile.GetCellValue(ir.SourceSheet, cell)
		if err != nil {
			return VerificationResult{}, err
		}
		if got != mapping.To {
			return normalizeHeadersFailure(ir, outputWorkbook, fmt.Sprintf("%s!%s=%q want %q", ir.SourceSheet, cell, got, mapping.To)), nil
		}
		writtenCells = append(writtenCells, ir.SourceSheet+"!"+cell)
	}

	return VerificationResult{
		Pass:                true,
		OperationFamily:     ir.OperationFamily,
		OutputWorkbook:      outputWorkbook,
		SummarySheet:        ir.SourceSheet,
		WrittenCells:        writtenCells,
		SourceSHA256Checked: true,
		Reasons:             []string{"output workbook contains normalized headers and non-header rows are unchanged"},
	}, nil
}

func compareRowsExceptHeader(sheetName string, expected, actual [][]string, headerRow int) error {
	if len(expected) != len(actual) {
		return fmt.Errorf("sheet %q row count=%d want %d", sheetName, len(actual), len(expected))
	}
	for rowIndex := range expected {
		if rowIndex == headerRow-1 {
			continue
		}
		if err := compareStringRows(sheetName, expected[rowIndex:rowIndex+1], actual[rowIndex:rowIndex+1]); err != nil {
			return err
		}
	}
	return nil
}

func compareFormulasExceptHeader(inputFile, outputFile *excelize.File, sheetName string, rowCount int, headerRow int) error {
	for row := 1; row <= rowCount; row++ {
		if row == headerRow {
			continue
		}
		inputCols, err := inputFile.GetCols(sheetName)
		if err != nil {
			return err
		}
		outputCols, err := outputFile.GetCols(sheetName)
		if err != nil {
			return err
		}
		colCount := len(inputCols)
		if len(outputCols) > colCount {
			colCount = len(outputCols)
		}
		for col := 1; col <= colCount; col++ {
			cell, _ := excelize.CoordinatesToCellName(col, row)
			inputFormula, err := inputFile.GetCellFormula(sheetName, cell)
			if err != nil {
				return err
			}
			outputFormula, err := outputFile.GetCellFormula(sheetName, cell)
			if err != nil {
				return err
			}
			if outputFormula != inputFormula {
				return fmt.Errorf("%s!%s formula=%q want %q", sheetName, cell, outputFormula, inputFormula)
			}
		}
	}
	return nil
}

func normalizeHeadersFailure(ir compiler.WorkbookOperationIR, outputWorkbook string, reason string) VerificationResult {
	return VerificationResult{
		Pass:            false,
		OperationFamily: ir.OperationFamily,
		OutputWorkbook:  outputWorkbook,
		SummarySheet:    ir.SourceSheet,
		Reasons:         []string{reason},
	}
}
