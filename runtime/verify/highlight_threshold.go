package verify

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/mwroh/sheet-ops/runtime/compiler"
	"github.com/xuri/excelize/v2"
)

func VerifyHighlightThreshold(ir compiler.WorkbookOperationIR, inputWorkbook, outputWorkbook, sourceSHA256Before, sourceSHA256After string) (VerificationResult, error) {
	inspection, err := compiler.InspectHighlightThresholdWorkbook(inputWorkbook, ir.SourceSheet, ir.TargetColumn)
	if err != nil {
		return VerificationResult{}, err
	}
	return VerifyHighlightThresholdExecution(inspection, ir, outputWorkbook, sourceSHA256Before, sourceSHA256After)
}

func VerifyHighlightThresholdExecution(inspection compiler.WorkbookInspection, ir compiler.WorkbookOperationIR, outputWorkbook, sourceSHA256Before, sourceSHA256After string) (VerificationResult, error) {
	if sourceSHA256Before == "" || sourceSHA256After == "" {
		return VerificationResult{
			Pass:            false,
			OperationFamily: ir.OperationFamily,
			OutputWorkbook:  outputWorkbook,
			Reasons:         []string{"execution hash evidence is missing"},
		}, nil
	}
	if sourceSHA256Before != sourceSHA256After {
		return VerificationResult{
			Pass:            false,
			OperationFamily: ir.OperationFamily,
			OutputWorkbook:  outputWorkbook,
			Reasons:         []string{"source workbook changed during execution"},
		}, nil
	}

	currentHash, err := hashFile(inspection.InputWorkbook)
	if err != nil {
		return VerificationResult{}, err
	}
	if currentHash != sourceSHA256Before {
		return VerificationResult{
			Pass:            false,
			OperationFamily: ir.OperationFamily,
			OutputWorkbook:  outputWorkbook,
			Reasons:         []string{"source workbook hash differs from execution evidence"},
		}, nil
	}

	outputFile, err := excelize.OpenFile(outputWorkbook)
	if err != nil {
		return VerificationResult{}, err
	}
	defer func() { _ = outputFile.Close() }()

	inputFile, err := excelize.OpenFile(inspection.InputWorkbook)
	if err != nil {
		return VerificationResult{}, err
	}
	defer func() { _ = inputFile.Close() }()

	expectedRows, err := expectedHighlightedRows(inspection, ir)
	if err != nil {
		return VerificationResult{}, err
	}
	if len(expectedRows) == 0 {
		pass, reasons, err := verifyZeroMatchWorkbook(inputFile, outputFile, inspection)
		if err != nil {
			return VerificationResult{}, err
		}
		return VerificationResult{
			Pass:            pass,
			OperationFamily: ir.OperationFamily,
			OutputWorkbook:  outputWorkbook,
			HighlightedRows: []int{},
			Reasons:         reasons,
		}, nil
	}

	reasons := []string{}
	pass := true
	highlightStyle := -1
	for _, row := range expectedRows {
		styleID, err := rowStyleID(outputFile, inspection.SourceSheet, row, len(inspection.HeaderNames))
		if err != nil {
			return VerificationResult{}, err
		}
		if highlightStyle == -1 {
			highlightStyle = styleID
		}
		if styleID <= 0 {
			pass = false
			reasons = append(reasons, fmt.Sprintf("row %d does not have a row-wide highlight style", row))
			continue
		}
		if styleID != highlightStyle {
			pass = false
			reasons = append(reasons, fmt.Sprintf("row %d does not share the expected highlight style", row))
		}
	}
	if highlightStyle <= 0 {
		pass = false
		reasons = append(reasons, "highlight style was not written to the workbook")
	} else {
		matches, err := styleMatchesHighlightColor(outputFile, highlightStyle, ir.HighlightColor)
		if err != nil {
			return VerificationResult{}, err
		}
		if !matches {
			pass = false
			reasons = append(reasons, fmt.Sprintf("highlight style does not use configured color %s", ir.HighlightColor))
		}
	}
	for row := inspection.HeaderRow + 1; row <= inspection.RowCount; row++ {
		if containsHighlightedRow(expectedRows, row) {
			continue
		}
		rowChanged, err := rowStylesChanged(inputFile, outputFile, inspection.SourceSheet, row, len(inspection.HeaderNames))
		if err != nil {
			return VerificationResult{}, err
		}
		if rowChanged {
			pass = false
			reasons = append(reasons, fmt.Sprintf("row %d unexpectedly changed style in non-highlighted row", row))
		}
	}
	if pass {
		reasons = append(reasons, "output workbook contains the expected highlighted rows and source hash is unchanged")
	}
	return VerificationResult{
		Pass:            pass,
		OperationFamily: ir.OperationFamily,
		OutputWorkbook:  outputWorkbook,
		HighlightedRows: expectedRows,
		Reasons:         reasons,
	}, nil
}

func verifyZeroMatchWorkbook(inputFile, outputFile *excelize.File, inspection compiler.WorkbookInspection) (bool, []string, error) {
	for row := inspection.HeaderRow + 1; row <= inspection.RowCount; row++ {
		for col := 1; col <= len(inspection.HeaderNames); col++ {
			cell, _ := excelize.CoordinatesToCellName(col, row)
			inputStyle, err := inputFile.GetCellStyle(inspection.SourceSheet, cell)
			if err != nil {
				return false, nil, err
			}
			outputStyle, err := outputFile.GetCellStyle(inspection.SourceSheet, cell)
			if err != nil {
				return false, nil, err
			}
			if inputStyle != outputStyle {
				return false, []string{fmt.Sprintf("row %d unexpectedly changed style in zero-match run", row)}, nil
			}
		}
	}
	return true, []string{"no rows matched the highlight threshold and workbook row styles are unchanged"}, nil
}

func expectedHighlightedRows(inspection compiler.WorkbookInspection, ir compiler.WorkbookOperationIR) ([]int, error) {
	file, err := excelize.OpenFile(inspection.InputWorkbook)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	rows, err := file.GetRows(inspection.SourceSheet)
	if err != nil {
		return nil, err
	}
	threshold := 0.0
	if ir.Threshold != nil {
		threshold = *ir.Threshold
	}
	highlightedRows := []int{}
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
		if compareHighlightNumber(number, ir.Operator, threshold) {
			highlightedRows = append(highlightedRows, rowIdx+1)
		}
	}
	return highlightedRows, nil
}

func rowStyleID(file *excelize.File, sheet string, row, headerColumnCount int) (int, error) {
	if headerColumnCount == 0 {
		return -1, nil
	}
	firstCell, _ := excelize.CoordinatesToCellName(1, row)
	styleID, err := file.GetCellStyle(sheet, firstCell)
	if err != nil {
		return -1, err
	}
	for col := 2; col <= headerColumnCount; col++ {
		cell, _ := excelize.CoordinatesToCellName(col, row)
		currentStyle, err := file.GetCellStyle(sheet, cell)
		if err != nil {
			return -1, err
		}
		if currentStyle != styleID {
			return -1, nil
		}
	}
	return styleID, nil
}

func rowStylesChanged(inputFile, outputFile *excelize.File, sheet string, row, headerColumnCount int) (bool, error) {
	for col := 1; col <= headerColumnCount; col++ {
		cell, _ := excelize.CoordinatesToCellName(col, row)
		inputStyle, err := inputFile.GetCellStyle(sheet, cell)
		if err != nil {
			return false, err
		}
		outputStyle, err := outputFile.GetCellStyle(sheet, cell)
		if err != nil {
			return false, err
		}
		if inputStyle != outputStyle {
			return true, nil
		}
	}
	return false, nil
}

func styleMatchesHighlightColor(file *excelize.File, styleID int, configuredColor string) (bool, error) {
	style, err := file.GetStyle(styleID)
	if err != nil {
		return false, err
	}
	want := normalizeHexColor(configuredColor)
	for _, color := range style.Fill.Color {
		if normalizeHexColor(color) == want {
			return true, nil
		}
	}
	return false, nil
}

func normalizeHexColor(color string) string {
	return strings.TrimPrefix(strings.ToUpper(strings.TrimSpace(color)), "#")
}

func containsHighlightedRow(rows []int, needle int) bool {
	for _, row := range rows {
		if row == needle {
			return true
		}
	}
	return false
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
