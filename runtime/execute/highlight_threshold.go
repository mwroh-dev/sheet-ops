package execute

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/mwroh/sheet-ops/runtime/compiler"
	"github.com/xuri/excelize/v2"
)

func RunHighlightThreshold(ir compiler.WorkbookOperationIR, inputWorkbook, outputWorkbook string) (ExecutionResult, error) {
	inspection, err := compiler.InspectHighlightThresholdWorkbook(inputWorkbook, ir.SourceSheet, ir.TargetColumn)
	if err != nil {
		return ExecutionResult{}, err
	}
	return RunHighlightThresholdWithInspection(inspection, ir, outputWorkbook)
}

func RunHighlightThresholdWithInspection(inspection compiler.WorkbookInspection, ir compiler.WorkbookOperationIR, outputWorkbook string) (ExecutionResult, error) {
	if err := validateHighlightBoundary(inspection.InputWorkbook, outputWorkbook, ir.PreserveOriginal); err != nil {
		return ExecutionResult{}, err
	}

	sourceBefore, err := hashFile(inspection.InputWorkbook)
	if err != nil {
		return ExecutionResult{}, err
	}

	file, err := excelize.OpenFile(inspection.InputWorkbook)
	if err != nil {
		return ExecutionResult{}, err
	}
	defer func() { _ = file.Close() }()

	rows, err := file.GetRows(inspection.SourceSheet)
	if err != nil {
		return ExecutionResult{}, err
	}
	styleID, err := file.NewStyle(&excelize.Style{
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{ir.HighlightColor},
			Pattern: 1,
		},
	})
	if err != nil {
		return ExecutionResult{}, err
	}

	threshold := 0.0
	if ir.Threshold != nil {
		threshold = *ir.Threshold
	}
	highlightedRows := []int{}
	highlightedCells := []string{}
	for rowIdx := 1; rowIdx < len(rows); rowIdx++ {
		row := rows[rowIdx]
		if inspection.TargetColumnIndex-1 >= len(row) {
			continue
		}
		value := strings.ReplaceAll(strings.TrimSpace(row[inspection.TargetColumnIndex-1]), ",", "")
		number, err := strconv.ParseFloat(value, 64)
		if err != nil {
			continue
		}
		if !compareHighlightNumber(number, ir.Operator, threshold) {
			continue
		}
		excelRow := rowIdx + 1
		highlightedRows = append(highlightedRows, excelRow)
		for colIdx := 0; colIdx < len(inspection.HeaderNames); colIdx++ {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, excelRow)
			if err := file.SetCellStyle(inspection.SourceSheet, cell, cell, styleID); err != nil {
				return ExecutionResult{}, err
			}
			highlightedCells = append(highlightedCells, cell)
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

	sourceAfter, err := hashFile(inspection.InputWorkbook)
	if err != nil {
		return ExecutionResult{}, err
	}

	return ExecutionResult{
		OperationFamily:    ir.OperationFamily,
		OutputWorkbook:     outputWorkbook,
		HighlightedRows:    highlightedRows,
		HighlightedCells:   highlightedCells,
		SourceSHA256Before: sourceBefore,
		SourceSHA256After:  sourceAfter,
	}, nil
}

func validateHighlightBoundary(inputWorkbook, outputWorkbook string, preserveOriginal bool) error {
	if outputWorkbook == "" {
		return fmt.Errorf("output workbook must not be empty")
	}
	if !preserveOriginal {
		return fmt.Errorf("highlight_threshold execution requires preserve_original=true")
	}
	if inputWorkbook == "" {
		return fmt.Errorf("input workbook must not be empty")
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

func compareHighlightNumber(number float64, operator string, threshold float64) bool {
	switch operator {
	case ">":
		return number > threshold
	case ">=":
		return number >= threshold
	case "<":
		return number < threshold
	case "<=":
		return number <= threshold
	case "=":
		return number == threshold
	default:
		return false
	}
}
