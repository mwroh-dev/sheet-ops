package execute

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mwroh/sheet-ops/runtime/compiler"
	"github.com/xuri/excelize/v2"
)

func RunRollForwardPeriod(ir compiler.WorkbookOperationIR, inputWorkbook, outputWorkbook string) (ExecutionResult, error) {
	if err := validateRollForwardPeriodBoundary(ir, inputWorkbook, outputWorkbook); err != nil {
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

	writtenCells := make([]string, 0, len(ir.CarryForwardMappings))
	for _, mapping := range ir.CarryForwardMappings {
		fromSheet := coalesceSheet(mapping.FromSheet, ir.SourceSheet)
		toSheet := coalesceSheet(mapping.ToSheet, ir.TargetSheet)
		value, err := carryForwardCellValue(file, fromSheet, mapping.FromCell)
		if err != nil {
			return ExecutionResult{}, err
		}
		if err := file.SetCellValue(toSheet, mapping.ToCell, value); err != nil {
			return ExecutionResult{}, err
		}
		writtenCells = append(writtenCells, toSheet+"!"+mapping.ToCell)
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
		SummarySheet:       ir.TargetSheet,
		WrittenCells:       writtenCells,
		SourceSHA256Before: sourceBefore,
		SourceSHA256After:  sourceAfter,
	}, nil
}

func carryForwardCellValue(file *excelize.File, sheet string, cell string) (string, error) {
	formula, err := file.GetCellFormula(sheet, cell)
	if err != nil {
		return "", err
	}
	if formula != "" {
		return file.CalcCellValue(sheet, cell)
	}
	return file.GetCellValue(sheet, cell)
}

func validateRollForwardPeriodBoundary(ir compiler.WorkbookOperationIR, inputWorkbook, outputWorkbook string) error {
	if inputWorkbook == "" {
		return fmt.Errorf("input workbook must not be empty")
	}
	if outputWorkbook == "" {
		return fmt.Errorf("output workbook must not be empty")
	}
	if !ir.PreserveOriginal {
		return fmt.Errorf("roll_forward_period execution requires preserve_original=true")
	}
	if ir.SourceSheet == "" {
		return fmt.Errorf("source sheet must not be empty")
	}
	if ir.TargetSheet == "" {
		return fmt.Errorf("target sheet must not be empty")
	}
	if len(ir.CarryForwardMappings) == 0 {
		return fmt.Errorf("carry forward mappings must not be empty")
	}
	for _, mapping := range ir.CarryForwardMappings {
		if mapping.FromCell == "" {
			return fmt.Errorf("carry forward mapping from_cell must not be empty")
		}
		if mapping.ToCell == "" {
			return fmt.Errorf("carry forward mapping to_cell must not be empty")
		}
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

func coalesceSheet(value, fallback string) string {
	if value != "" {
		return value
	}
	return fallback
}
