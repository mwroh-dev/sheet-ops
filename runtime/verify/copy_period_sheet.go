package verify

import (
	"fmt"

	"github.com/mwroh/sheet-ops/runtime/compiler"
	"github.com/xuri/excelize/v2"
)

func VerifyCopyPeriodSheet(ir compiler.WorkbookOperationIR, inputWorkbook, outputWorkbook, sourceSHA256Before, sourceSHA256After string) (VerificationResult, error) {
	if sourceSHA256Before == "" || sourceSHA256After == "" {
		return VerificationResult{Pass: false, OperationFamily: ir.OperationFamily, OutputWorkbook: outputWorkbook, SummarySheet: ir.TargetSheet, Reasons: []string{"execution hash evidence is missing"}}, nil
	}
	if sourceSHA256Before != sourceSHA256After {
		return VerificationResult{Pass: false, OperationFamily: ir.OperationFamily, OutputWorkbook: outputWorkbook, SummarySheet: ir.TargetSheet, Reasons: []string{"source workbook changed during execution"}}, nil
	}
	currentHash, err := hashFile(inputWorkbook)
	if err != nil {
		return VerificationResult{}, err
	}
	if currentHash != sourceSHA256Before {
		return VerificationResult{Pass: false, OperationFamily: ir.OperationFamily, OutputWorkbook: outputWorkbook, SummarySheet: ir.TargetSheet, Reasons: []string{"source workbook hash differs from execution evidence"}}, nil
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

	if index, err := outputHandle.GetSheetIndex(ir.TargetSheet); err != nil {
		return VerificationResult{}, err
	} else if index < 0 {
		return VerificationResult{Pass: false, OperationFamily: ir.OperationFamily, OutputWorkbook: outputWorkbook, SummarySheet: ir.TargetSheet, Reasons: []string{fmt.Sprintf("target sheet %q missing", ir.TargetSheet)}}, nil
	}
	sourceRows, err := inputHandle.GetRows(ir.SourceSheet)
	if err != nil {
		return VerificationResult{}, err
	}
	targetRows, err := outputHandle.GetRows(ir.TargetSheet)
	if err != nil {
		return VerificationResult{}, err
	}
	if err := compareCopiedRows(sourceRows, targetRows); err != nil {
		return VerificationResult{Pass: false, OperationFamily: ir.OperationFamily, OutputWorkbook: outputWorkbook, SummarySheet: ir.TargetSheet, Reasons: []string{err.Error()}}, nil
	}

	for rowIndex, row := range sourceRows {
		for colIndex := range row {
			cell, err := excelize.CoordinatesToCellName(colIndex+1, rowIndex+1)
			if err != nil {
				return VerificationResult{}, err
			}
			sourceFormula, err := inputHandle.GetCellFormula(ir.SourceSheet, cell)
			if err != nil {
				return VerificationResult{}, err
			}
			targetFormula, err := outputHandle.GetCellFormula(ir.TargetSheet, cell)
			if err != nil {
				return VerificationResult{}, err
			}
			if sourceFormula != targetFormula {
				return VerificationResult{Pass: false, OperationFamily: ir.OperationFamily, OutputWorkbook: outputWorkbook, SummarySheet: ir.TargetSheet, Reasons: []string{fmt.Sprintf("%s formula=%q want %q", cell, targetFormula, sourceFormula)}}, nil
			}
			sourceStyle, err := inputHandle.GetCellStyle(ir.SourceSheet, cell)
			if err != nil {
				return VerificationResult{}, err
			}
			targetStyle, err := outputHandle.GetCellStyle(ir.TargetSheet, cell)
			if err != nil {
				return VerificationResult{}, err
			}
			if sourceStyle != targetStyle {
				return VerificationResult{Pass: false, OperationFamily: ir.OperationFamily, OutputWorkbook: outputWorkbook, SummarySheet: ir.TargetSheet, Reasons: []string{fmt.Sprintf("%s style=%d want %d", cell, targetStyle, sourceStyle)}}, nil
			}
		}
	}

	return VerificationResult{
		Pass:                true,
		OperationFamily:     ir.OperationFamily,
		OutputWorkbook:      outputWorkbook,
		SummarySheet:        ir.TargetSheet,
		Reasons:             []string{"target sheet preserves source values, formulas, styles, and source workbook is unchanged"},
		SourceSHA256Checked: true,
	}, nil
}

func compareCopiedRows(expected, actual [][]string) error {
	if len(actual) != len(expected) {
		return fmt.Errorf("target row count=%d want %d", len(actual), len(expected))
	}
	for rowIndex := range expected {
		if len(actual[rowIndex]) != len(expected[rowIndex]) {
			return fmt.Errorf("target row %d column count=%d want %d", rowIndex+1, len(actual[rowIndex]), len(expected[rowIndex]))
		}
		for colIndex := range expected[rowIndex] {
			if actual[rowIndex][colIndex] != expected[rowIndex][colIndex] {
				return fmt.Errorf("target cell[%d][%d]=%q want %q", rowIndex+1, colIndex+1, actual[rowIndex][colIndex], expected[rowIndex][colIndex])
			}
		}
	}
	return nil
}
