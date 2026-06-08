package verify

import (
	"fmt"
	"strings"

	"github.com/mwroh/sheet-ops/runtime/compiler"
	"github.com/xuri/excelize/v2"
)

func VerifyProtectFormulaCells(ir compiler.WorkbookOperationIR, inputWorkbook, outputWorkbook, sourceSHA256Before, sourceSHA256After string) (VerificationResult, error) {
	if sourceSHA256Before == "" || sourceSHA256After == "" {
		return VerificationResult{Pass: false, OperationFamily: ir.OperationFamily, OutputWorkbook: outputWorkbook, SummarySheet: ir.SourceSheet, Reasons: []string{"execution hash evidence is missing"}}, nil
	}
	if sourceSHA256Before != sourceSHA256After {
		return VerificationResult{Pass: false, OperationFamily: ir.OperationFamily, OutputWorkbook: outputWorkbook, SummarySheet: ir.SourceSheet, Reasons: []string{"source workbook changed during execution"}}, nil
	}
	currentHash, err := hashFile(inputWorkbook)
	if err != nil {
		return VerificationResult{}, err
	}
	if currentHash != sourceSHA256Before {
		return VerificationResult{Pass: false, OperationFamily: ir.OperationFamily, OutputWorkbook: outputWorkbook, SummarySheet: ir.SourceSheet, Reasons: []string{"source workbook hash differs from execution evidence"}}, nil
	}

	outputHandle, err := excelize.OpenFile(outputWorkbook)
	if err != nil {
		return VerificationResult{}, err
	}
	defer func() { _ = outputHandle.Close() }()

	protection, err := outputHandle.GetSheetProtection(ir.SourceSheet)
	if err != nil {
		return VerificationResult{}, err
	}
	if !protection.SelectLockedCells || !protection.SelectUnlockedCells || !protection.EditScenarios {
		return VerificationResult{Pass: false, OperationFamily: ir.OperationFamily, OutputWorkbook: outputWorkbook, SummarySheet: ir.SourceSheet, Reasons: []string{"sheet protection options are not present"}}, nil
	}

	formulaCells := []string{}
	for _, targetRange := range ir.ProtectionRule.FormulaRanges {
		cells, err := cellsInRange(targetRange)
		if err != nil {
			return VerificationResult{}, err
		}
		for _, cell := range cells {
			formula, err := outputHandle.GetCellFormula(ir.SourceSheet, cell)
			if err != nil {
				return VerificationResult{}, err
			}
			if strings.TrimSpace(formula) == "" {
				return VerificationResult{Pass: false, OperationFamily: ir.OperationFamily, OutputWorkbook: outputWorkbook, SummarySheet: ir.SourceSheet, FormulaCells: formulaCells, Reasons: []string{fmt.Sprintf("%s has no formula to protect", cell)}}, nil
			}
			locked, err := cellLocked(outputHandle, ir.SourceSheet, cell)
			if err != nil {
				return VerificationResult{}, err
			}
			if !locked {
				return VerificationResult{Pass: false, OperationFamily: ir.OperationFamily, OutputWorkbook: outputWorkbook, SummarySheet: ir.SourceSheet, FormulaCells: formulaCells, Reasons: []string{fmt.Sprintf("%s is not locked", cell)}}, nil
			}
			formulaCells = append(formulaCells, cell)
		}
	}
	for _, targetRange := range ir.ProtectionRule.InputRanges {
		cells, err := cellsInRange(targetRange)
		if err != nil {
			return VerificationResult{}, err
		}
		for _, cell := range cells {
			locked, err := cellLocked(outputHandle, ir.SourceSheet, cell)
			if err != nil {
				return VerificationResult{}, err
			}
			if locked {
				return VerificationResult{Pass: false, OperationFamily: ir.OperationFamily, OutputWorkbook: outputWorkbook, SummarySheet: ir.SourceSheet, FormulaCells: formulaCells, Reasons: []string{fmt.Sprintf("%s should remain editable", cell)}}, nil
			}
		}
	}

	return VerificationResult{
		Pass:                true,
		OperationFamily:     ir.OperationFamily,
		OutputWorkbook:      outputWorkbook,
		SummarySheet:        ir.SourceSheet,
		FormulaCells:        formulaCells,
		Reasons:             []string{"formula cells are locked, input cells are editable, sheet protection is enabled, and source workbook is unchanged"},
		SourceSHA256Checked: true,
	}, nil
}

func cellLocked(file *excelize.File, sheet, cell string) (bool, error) {
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

func cellsInRange(targetRange string) ([]string, error) {
	parts := strings.Split(strings.TrimSpace(targetRange), ":")
	if len(parts) == 1 {
		if _, _, err := excelize.SplitCellName(parts[0]); err != nil {
			return nil, err
		}
		return []string{strings.ToUpper(parts[0])}, nil
	}
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid cell range %q", targetRange)
	}
	startColName, startRow, err := excelize.SplitCellName(parts[0])
	if err != nil {
		return nil, err
	}
	endColName, endRow, err := excelize.SplitCellName(parts[1])
	if err != nil {
		return nil, err
	}
	startCol, err := excelize.ColumnNameToNumber(startColName)
	if err != nil {
		return nil, err
	}
	endCol, err := excelize.ColumnNameToNumber(endColName)
	if err != nil {
		return nil, err
	}
	if startCol > endCol || startRow > endRow {
		return nil, fmt.Errorf("invalid descending range %q", targetRange)
	}
	cells := []string{}
	for row := startRow; row <= endRow; row++ {
		for col := startCol; col <= endCol; col++ {
			cell, err := excelize.CoordinatesToCellName(col, row)
			if err != nil {
				return nil, err
			}
			cells = append(cells, cell)
		}
	}
	return cells, nil
}
