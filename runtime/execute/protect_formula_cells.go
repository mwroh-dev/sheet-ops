package execute

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mwroh/sheet-ops/runtime/compiler"
	"github.com/xuri/excelize/v2"
)

func RunProtectFormulaCells(ir compiler.WorkbookOperationIR, inputWorkbook, outputWorkbook string) (ExecutionResult, error) {
	if err := validateProtectFormulaCellsBoundary(ir, inputWorkbook, outputWorkbook); err != nil {
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

	formulaCells := []string{}
	for _, targetRange := range ir.ProtectionRule.FormulaRanges {
		cells, err := cellsInRange(targetRange)
		if err != nil {
			return ExecutionResult{}, err
		}
		for _, cell := range cells {
			if err := setCellLocked(file, ir.SourceSheet, cell, true); err != nil {
				return ExecutionResult{}, err
			}
			formulaCells = append(formulaCells, cell)
		}
	}
	for _, targetRange := range ir.ProtectionRule.InputRanges {
		cells, err := cellsInRange(targetRange)
		if err != nil {
			return ExecutionResult{}, err
		}
		for _, cell := range cells {
			if err := setCellLocked(file, ir.SourceSheet, cell, false); err != nil {
				return ExecutionResult{}, err
			}
		}
	}

	if err := file.ProtectSheet(ir.SourceSheet, &excelize.SheetProtectionOptions{
		Password:            ir.ProtectionRule.Password,
		SelectLockedCells:   true,
		SelectUnlockedCells: true,
		EditScenarios:       true,
	}); err != nil {
		return ExecutionResult{}, err
	}

	outputDir := filepath.Dir(outputWorkbook)
	if outputDir != "." {
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
		FormulaCells:       formulaCells,
		SourceSHA256Before: sourceBefore,
		SourceSHA256After:  sourceAfter,
	}, nil
}

func validateProtectFormulaCellsBoundary(ir compiler.WorkbookOperationIR, inputWorkbook, outputWorkbook string) error {
	if inputWorkbook == "" {
		return fmt.Errorf("input workbook must not be empty")
	}
	if outputWorkbook == "" {
		return fmt.Errorf("output workbook must not be empty")
	}
	if !ir.PreserveOriginal {
		return fmt.Errorf("protect_formula_cells execution requires preserve_original=true")
	}
	if ir.SourceSheet == "" {
		return fmt.Errorf("source sheet must not be empty")
	}
	if ir.ProtectionRule == nil || len(ir.ProtectionRule.FormulaRanges) == 0 {
		return fmt.Errorf("formula protection ranges must not be empty")
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

func setCellLocked(file *excelize.File, sheet, cell string, locked bool) error {
	styleID, err := file.GetCellStyle(sheet, cell)
	if err != nil {
		return err
	}
	style, err := file.GetStyle(styleID)
	if err != nil {
		return err
	}
	if style.Protection == nil {
		style.Protection = &excelize.Protection{}
	}
	style.Protection.Locked = locked
	newStyleID, err := file.NewStyle(style)
	if err != nil {
		return err
	}
	return file.SetCellStyle(sheet, cell, cell, newStyleID)
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
