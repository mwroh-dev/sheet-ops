package verify

import (
	"fmt"

	"github.com/mwroh/sheet-ops/runtime/compiler"
	"github.com/xuri/excelize/v2"
)

func VerifyRollForwardPeriod(ir compiler.WorkbookOperationIR, inputWorkbook, outputWorkbook, sourceSHA256Before, sourceSHA256After string) (VerificationResult, error) {
	if sourceSHA256Before == "" || sourceSHA256After == "" {
		return rollForwardPeriodFailure(ir, outputWorkbook, "execution hash evidence is missing"), nil
	}
	if sourceSHA256Before != sourceSHA256After {
		return rollForwardPeriodFailure(ir, outputWorkbook, "source workbook changed during execution"), nil
	}
	currentHash, err := hashFile(inputWorkbook)
	if err != nil {
		return VerificationResult{}, err
	}
	if currentHash != sourceSHA256Before {
		return rollForwardPeriodFailure(ir, outputWorkbook, "source workbook hash differs from execution evidence"), nil
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

	if err := verifyRollForwardSourceSheetsPreserved(inputFile, outputFile, ir); err != nil {
		return rollForwardPeriodFailure(ir, outputWorkbook, err.Error()), nil
	}

	writtenCells := make([]string, 0, len(ir.CarryForwardMappings))
	for _, mapping := range ir.CarryForwardMappings {
		fromSheet := coalesceVerifySheet(mapping.FromSheet, ir.SourceSheet)
		toSheet := coalesceVerifySheet(mapping.ToSheet, ir.TargetSheet)
		want, err := carryForwardVerifyCellValue(inputFile, fromSheet, mapping.FromCell)
		if err != nil {
			return VerificationResult{}, err
		}
		got, err := outputFile.GetCellValue(toSheet, mapping.ToCell)
		if err != nil {
			return VerificationResult{}, err
		}
		if got != want {
			return rollForwardPeriodFailure(ir, outputWorkbook, fmt.Sprintf("%s!%s=%q want carried value %q from %s!%s", toSheet, mapping.ToCell, got, want, fromSheet, mapping.FromCell)), nil
		}
		writtenCells = append(writtenCells, toSheet+"!"+mapping.ToCell)
	}

	return VerificationResult{
		Pass:                true,
		OperationFamily:     ir.OperationFamily,
		OutputWorkbook:      outputWorkbook,
		SummarySheet:        ir.TargetSheet,
		WrittenCells:        writtenCells,
		SourceSHA256Checked: true,
		Reasons:             []string{"output workbook carries declared closing values into next opening cells and source workbook is unchanged"},
	}, nil
}

func carryForwardVerifyCellValue(file *excelize.File, sheet string, cell string) (string, error) {
	formula, err := file.GetCellFormula(sheet, cell)
	if err != nil {
		return "", err
	}
	if formula != "" {
		return file.CalcCellValue(sheet, cell)
	}
	return file.GetCellValue(sheet, cell)
}

func verifyRollForwardSourceSheetsPreserved(inputFile, outputFile *excelize.File, ir compiler.WorkbookOperationIR) error {
	allowed := map[string]struct{}{}
	for _, mapping := range ir.CarryForwardMappings {
		toSheet := coalesceVerifySheet(mapping.ToSheet, ir.TargetSheet)
		allowed[toSheet+"!"+mapping.ToCell] = struct{}{}
	}
	for _, sheet := range inputFile.GetSheetList() {
		inputRows, err := inputFile.GetRows(sheet)
		if err != nil {
			return err
		}
		outputRows, err := outputFile.GetRows(sheet)
		if err != nil {
			return fmt.Errorf("source sheet %q is missing from output workbook", sheet)
		}
		if len(inputRows) != len(outputRows) {
			return fmt.Errorf("sheet %q row count=%d want %d", sheet, len(outputRows), len(inputRows))
		}
		for rowIndex, inputRow := range inputRows {
			outputRow := outputRows[rowIndex]
			maxCols := len(inputRow)
			if len(outputRow) > maxCols {
				maxCols = len(outputRow)
			}
			for colIndex := 0; colIndex < maxCols; colIndex++ {
				cell, _ := excelize.CoordinatesToCellName(colIndex+1, rowIndex+1)
				if _, ok := allowed[sheet+"!"+cell]; ok {
					continue
				}
				if valueAtString(inputRow, colIndex) != valueAtString(outputRow, colIndex) {
					return fmt.Errorf("%s!%s=%q want %q", sheet, cell, valueAtString(outputRow, colIndex), valueAtString(inputRow, colIndex))
				}
				inputFormula, err := inputFile.GetCellFormula(sheet, cell)
				if err != nil {
					return err
				}
				outputFormula, err := outputFile.GetCellFormula(sheet, cell)
				if err != nil {
					return err
				}
				if outputFormula != inputFormula {
					return fmt.Errorf("%s!%s formula=%q want %q", sheet, cell, outputFormula, inputFormula)
				}
			}
		}
	}
	return nil
}

func valueAtString(row []string, index int) string {
	if index < 0 || index >= len(row) {
		return ""
	}
	return row[index]
}

func coalesceVerifySheet(value, fallback string) string {
	if value != "" {
		return value
	}
	return fallback
}

func rollForwardPeriodFailure(ir compiler.WorkbookOperationIR, outputWorkbook string, reason string) VerificationResult {
	return VerificationResult{
		Pass:            false,
		OperationFamily: ir.OperationFamily,
		OutputWorkbook:  outputWorkbook,
		SummarySheet:    ir.TargetSheet,
		Reasons:         []string{reason},
	}
}
