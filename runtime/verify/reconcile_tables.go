package verify

import (
	"fmt"

	"github.com/mwroh/sheet-ops/runtime/compiler"
	"github.com/xuri/excelize/v2"
)

type reconciliationRowRef struct {
	rowNumber int
	values    []string
}

func VerifyReconcileTables(ir compiler.WorkbookOperationIR, inputWorkbook, outputWorkbook, sourceSHA256Before, sourceSHA256After string) (VerificationResult, error) {
	if sourceSHA256Before == "" || sourceSHA256After == "" {
		return VerificationResult{
			Pass:            false,
			OperationFamily: ir.OperationFamily,
			OutputWorkbook:  outputWorkbook,
			SummarySheet:    ir.TargetSheet,
			Reasons:         []string{"execution hash evidence is missing"},
		}, nil
	}
	if sourceSHA256Before != sourceSHA256After {
		return VerificationResult{
			Pass:            false,
			OperationFamily: ir.OperationFamily,
			OutputWorkbook:  outputWorkbook,
			SummarySheet:    ir.TargetSheet,
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
			SummarySheet:    ir.TargetSheet,
			Reasons:         []string{"source workbook hash differs from execution evidence"},
		}, nil
	}

	inputFile, err := excelize.OpenFile(inputWorkbook)
	if err != nil {
		return VerificationResult{}, err
	}
	defer func() { _ = inputFile.Close() }()
	expectedRows, err := buildExpectedReconciliationRows(inputFile, ir)
	if err != nil {
		return VerificationResult{}, err
	}

	outputFile, err := excelize.OpenFile(outputWorkbook)
	if err != nil {
		return VerificationResult{}, err
	}
	defer func() { _ = outputFile.Close() }()
	if result := verifyReconciliationSourceSheetPreserved(inputFile, outputFile, ir.SourceSheet, ir); !result.Pass {
		return result, nil
	}
	if result := verifyReconciliationSourceSheetPreserved(inputFile, outputFile, ir.LookupSheet, ir); !result.Pass {
		return result, nil
	}
	if !containsString(outputFile.GetSheetList(), ir.TargetSheet) {
		return VerificationResult{
			Pass:            false,
			OperationFamily: ir.OperationFamily,
			OutputWorkbook:  outputWorkbook,
			SummarySheet:    ir.TargetSheet,
			Reasons:         []string{fmt.Sprintf("reconciliation sheet %q was not created", ir.TargetSheet)},
		}, nil
	}
	actualRows, err := outputFile.GetRows(ir.TargetSheet)
	if err != nil {
		return VerificationResult{}, err
	}
	if err := compareReconciliationRows(ir.TargetSheet, expectedRows, actualRows); err != nil {
		return VerificationResult{
			Pass:            false,
			OperationFamily: ir.OperationFamily,
			OutputWorkbook:  outputWorkbook,
			SummarySheet:    ir.TargetSheet,
			SummaryRows:     summaryRowCount(actualRows),
			Reasons:         []string{err.Error()},
		}, nil
	}
	return VerificationResult{
		Pass:            true,
		OperationFamily: ir.OperationFamily,
		OutputWorkbook:  outputWorkbook,
		SummarySheet:    ir.TargetSheet,
		SummaryRows:     summaryRowCount(actualRows),
		Reasons:         []string{"output workbook contains the expected reconciliation sheet and source workbook is unchanged"},
	}, nil
}

func compareReconciliationRows(sheetName string, expected, actual [][]string) error {
	if len(expected) != len(actual) {
		return fmt.Errorf("source sheet %q row count=%d want %d", sheetName, len(actual), len(expected))
	}
	for rowIndex := range expected {
		for colIndex := range expected[rowIndex] {
			if expected[rowIndex][colIndex] != reconciliationValueAt(actual[rowIndex], colIndex) {
				return fmt.Errorf("source sheet %q cell[%d][%d]=%q want %q", sheetName, rowIndex, colIndex, reconciliationValueAt(actual[rowIndex], colIndex), expected[rowIndex][colIndex])
			}
		}
		for colIndex := len(expected[rowIndex]); colIndex < len(actual[rowIndex]); colIndex++ {
			if actual[rowIndex][colIndex] != "" {
				return fmt.Errorf("source sheet %q cell[%d][%d]=%q want blank", sheetName, rowIndex, colIndex, actual[rowIndex][colIndex])
			}
		}
	}
	return nil
}

func verifyReconciliationSourceSheetPreserved(inputFile, outputFile *excelize.File, sheet string, ir compiler.WorkbookOperationIR) VerificationResult {
	inputRows, err := inputFile.GetRows(sheet)
	if err != nil {
		return VerificationResult{
			Pass:            false,
			OperationFamily: ir.OperationFamily,
			SummarySheet:    ir.TargetSheet,
			Reasons:         []string{err.Error()},
		}
	}
	outputRows, err := outputFile.GetRows(sheet)
	if err != nil {
		return VerificationResult{
			Pass:            false,
			OperationFamily: ir.OperationFamily,
			SummarySheet:    ir.TargetSheet,
			Reasons:         []string{fmt.Sprintf("source sheet %q is missing from output workbook", sheet)},
		}
	}
	if err := compareStringRows(sheet, inputRows, outputRows); err != nil {
		return VerificationResult{
			Pass:            false,
			OperationFamily: ir.OperationFamily,
			SummarySheet:    ir.TargetSheet,
			Reasons:         []string{err.Error()},
		}
	}
	return VerificationResult{Pass: true}
}

func buildExpectedReconciliationRows(file *excelize.File, ir compiler.WorkbookOperationIR) ([][]string, error) {
	leftRows, err := file.GetRows(ir.SourceSheet)
	if err != nil {
		return nil, fmt.Errorf("read source sheet %q: %w", ir.SourceSheet, err)
	}
	rightRows, err := file.GetRows(ir.LookupSheet)
	if err != nil {
		return nil, fmt.Errorf("read lookup sheet %q: %w", ir.LookupSheet, err)
	}
	if len(leftRows) == 0 || len(rightRows) == 0 {
		return nil, fmt.Errorf("reconciliation sheets must not be empty")
	}
	leftIndex := reconciliationHeaderIndex(leftRows[0])
	rightIndex := reconciliationHeaderIndex(rightRows[0])
	leftKeyIndex, ok := leftIndex[ir.LeftKey]
	if !ok {
		return nil, fmt.Errorf("source key %q missing", ir.LeftKey)
	}
	rightKeyIndex, ok := rightIndex[ir.RightKey]
	if !ok {
		return nil, fmt.Errorf("lookup key %q missing", ir.RightKey)
	}
	for _, mapping := range ir.CompareMappings {
		if _, ok := leftIndex[mapping.LeftColumn]; !ok {
			return nil, fmt.Errorf("source compare column %q missing", mapping.LeftColumn)
		}
		if _, ok := rightIndex[mapping.RightColumn]; !ok {
			return nil, fmt.Errorf("lookup compare column %q missing", mapping.RightColumn)
		}
	}
	leftByKey, leftOrder, err := reconciliationIndexRows(leftRows[1:], leftKeyIndex, "source")
	if err != nil {
		return nil, err
	}
	rightByKey, rightOrder, err := reconciliationIndexRows(rightRows[1:], rightKeyIndex, "lookup")
	if err != nil {
		return nil, err
	}
	outRows := [][]string{{"status", "key", "left_row", "right_row", "field", "left_value", "right_value"}}
	for _, key := range leftOrder {
		left := leftByKey[key]
		right, ok := rightByKey[key]
		if !ok {
			outRows = append(outRows, []string{"left_only", key, fmt.Sprintf("%d", left.rowNumber), "", "", "", ""})
			continue
		}
		mismatched := false
		for _, mapping := range ir.CompareMappings {
			leftValue := reconciliationValueAt(left.values, leftIndex[mapping.LeftColumn])
			rightValue := reconciliationValueAt(right.values, rightIndex[mapping.RightColumn])
			if leftValue == rightValue {
				continue
			}
			mismatched = true
			field := mapping.As
			if field == "" {
				field = mapping.LeftColumn
			}
			outRows = append(outRows, []string{"value_mismatch", key, fmt.Sprintf("%d", left.rowNumber), fmt.Sprintf("%d", right.rowNumber), field, leftValue, rightValue})
		}
		if !mismatched {
			outRows = append(outRows, []string{"matched", key, fmt.Sprintf("%d", left.rowNumber), fmt.Sprintf("%d", right.rowNumber), "", "", ""})
		}
	}
	for _, key := range rightOrder {
		if _, ok := leftByKey[key]; ok {
			continue
		}
		right := rightByKey[key]
		outRows = append(outRows, []string{"right_only", key, "", fmt.Sprintf("%d", right.rowNumber), "", "", ""})
	}
	return outRows, nil
}

func reconciliationIndexRows(rows [][]string, keyIndex int, side string) (map[string]reconciliationRowRef, []string, error) {
	byKey := make(map[string]reconciliationRowRef, len(rows))
	order := []string{}
	for index, row := range rows {
		key := reconciliationValueAt(row, keyIndex)
		if key == "" {
			continue
		}
		if existing, ok := byKey[key]; ok {
			return nil, nil, fmt.Errorf("%s duplicate key %q at rows %d and %d", side, key, existing.rowNumber, index+2)
		}
		byKey[key] = reconciliationRowRef{rowNumber: index + 2, values: row}
		order = append(order, key)
	}
	return byKey, order, nil
}

func reconciliationHeaderIndex(header []string) map[string]int {
	index := make(map[string]int, len(header))
	for idx, value := range header {
		index[value] = idx
	}
	return index
}

func reconciliationValueAt(row []string, index int) string {
	if index < 0 || index >= len(row) {
		return ""
	}
	return row[index]
}
