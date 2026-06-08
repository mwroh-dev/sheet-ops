package verify

import (
	"fmt"
	"strings"

	"github.com/mwroh/sheet-ops/runtime/compiler"
	runtimeformula "github.com/mwroh/sheet-ops/runtime/formula"
	"github.com/xuri/excelize/v2"
)

func VerifyExtendTableFormulas(ir compiler.WorkbookOperationIR, inputWorkbook, outputWorkbook, sourceSHA256Before, sourceSHA256After string) (VerificationResult, error) {
	if sourceSHA256Before == "" || sourceSHA256After == "" {
		return VerificationResult{
			Pass:            false,
			OperationFamily: ir.OperationFamily,
			OutputWorkbook:  outputWorkbook,
			SummarySheet:    ir.SourceSheet,
			Reasons:         []string{"execution hash evidence is missing"},
		}, nil
	}
	if sourceSHA256Before != sourceSHA256After {
		return VerificationResult{
			Pass:            false,
			OperationFamily: ir.OperationFamily,
			OutputWorkbook:  outputWorkbook,
			SummarySheet:    ir.SourceSheet,
			Reasons:         []string{"source workbook changed during execution"},
		}, nil
	}
	currentHash, err := hashFile(inputWorkbook)
	if err != nil {
		return VerificationResult{}, err
	}
	if currentHash != sourceSHA256Before {
		return VerificationResult{
			Pass:            false,
			OperationFamily: ir.OperationFamily,
			OutputWorkbook:  outputWorkbook,
			SummarySheet:    ir.SourceSheet,
			Reasons:         []string{"source workbook hash differs from execution evidence"},
		}, nil
	}

	inputHandle, err := excelize.OpenFile(inputWorkbook)
	if err != nil {
		return VerificationResult{}, err
	}
	defer func() { _ = inputHandle.Close() }()
	outputHandle, err := excelize.OpenFile(outputWorkbook)
	if err != nil {
		return VerificationResult{}, err
	}
	defer func() { _ = outputHandle.Close() }()

	formulaCells := []string{}
	for _, column := range ir.FormulaColumns {
		sourceCell, err := formulaCell(column, ir.FormulaSourceRow)
		if err != nil {
			return VerificationResult{}, err
		}
		sourceFormula, err := inputHandle.GetCellFormula(ir.SourceSheet, sourceCell)
		if err != nil {
			return VerificationResult{}, err
		}
		if strings.TrimSpace(sourceFormula) == "" {
			return VerificationResult{
				Pass:            false,
				OperationFamily: ir.OperationFamily,
				OutputWorkbook:  outputWorkbook,
				SummarySheet:    ir.SourceSheet,
				Reasons:         []string{fmt.Sprintf("%s has no source formula", sourceCell)},
			}, nil
		}
		for _, targetRow := range ir.TargetRows {
			targetCell, err := formulaCell(column, targetRow)
			if err != nil {
				return VerificationResult{}, err
			}
			expected, err := runtimeformula.TranslateFormulaRows(sourceFormula, targetRow-ir.FormulaSourceRow)
			if err != nil {
				return VerificationResult{}, err
			}
			actual, err := outputHandle.GetCellFormula(ir.SourceSheet, targetCell)
			if err != nil {
				return VerificationResult{}, err
			}
			if normalizeFormula(actual) != normalizeFormula(expected) {
				return VerificationResult{
					Pass:            false,
					OperationFamily: ir.OperationFamily,
					OutputWorkbook:  outputWorkbook,
					SummarySheet:    ir.SourceSheet,
					FormulaCells:    formulaCells,
					Reasons:         []string{fmt.Sprintf("%s formula=%q want %q", targetCell, actual, expected)},
				}, nil
			}
			formulaCells = append(formulaCells, targetCell)
		}
	}

	return VerificationResult{
		Pass:                true,
		OperationFamily:     ir.OperationFamily,
		OutputWorkbook:      outputWorkbook,
		SummarySheet:        ir.SourceSheet,
		FormulaCells:        formulaCells,
		Reasons:             []string{"target rows contain expected translated formulas and source workbook is unchanged"},
		SourceSHA256Checked: true,
	}, nil
}

func formulaCell(column string, row int) (string, error) {
	if row < 1 {
		return "", fmt.Errorf("row must be positive")
	}
	normalizedColumn := strings.ToUpper(strings.TrimSpace(column))
	if normalizedColumn == "" {
		return "", fmt.Errorf("formula column must not be empty")
	}
	columnNumber, err := excelize.ColumnNameToNumber(normalizedColumn)
	if err != nil {
		return "", err
	}
	return excelize.CoordinatesToCellName(columnNumber, row)
}

func normalizeFormula(value string) string {
	if strings.HasPrefix(value, "=") {
		return value
	}
	return "=" + value
}
