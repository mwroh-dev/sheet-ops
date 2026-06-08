package execute

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mwroh/sheet-ops/runtime/compiler"
	"github.com/xuri/excelize/v2"
)

type reconciliationRowRef struct {
	rowNumber int
	values    []string
}

func RunReconcileTables(ir compiler.WorkbookOperationIR, inputWorkbook, outputWorkbook string) (ExecutionResult, error) {
	if err := validateReconcileTablesExecutionBoundary(ir, inputWorkbook, outputWorkbook); err != nil {
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

	outRows, err := BuildReconciliationRows(file, ir)
	if err != nil {
		return ExecutionResult{}, err
	}
	if containsString(file.GetSheetList(), ir.TargetSheet) {
		file.DeleteSheet(ir.TargetSheet)
	}
	if _, err := file.NewSheet(ir.TargetSheet); err != nil {
		return ExecutionResult{}, err
	}
	for rowIndex, row := range outRows {
		values := make([]any, len(row))
		for columnIndex, value := range row {
			values[columnIndex] = value
		}
		cell, _ := excelize.CoordinatesToCellName(1, rowIndex+1)
		if err := file.SetSheetRow(ir.TargetSheet, cell, &values); err != nil {
			return ExecutionResult{}, err
		}
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
		SummarySheet:       ir.TargetSheet,
		SummaryRows:        len(outRows) - 1,
		SourceSHA256Before: sourceBefore,
		SourceSHA256After:  sourceAfter,
	}, nil
}

func validateReconcileTablesExecutionBoundary(ir compiler.WorkbookOperationIR, inputWorkbook, outputWorkbook string) error {
	if inputWorkbook == "" {
		return fmt.Errorf("input workbook must not be empty")
	}
	if outputWorkbook == "" {
		return fmt.Errorf("output workbook must not be empty")
	}
	if !ir.PreserveOriginal {
		return fmt.Errorf("reconcile_tables execution requires preserve_original=true")
	}
	if ir.TargetSheet == ir.SourceSheet || ir.TargetSheet == ir.LookupSheet {
		return fmt.Errorf("target sheet must differ from source and lookup sheets")
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

func BuildReconciliationRows(file *excelize.File, ir compiler.WorkbookOperationIR) ([][]string, error) {
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
	if ir.LeftKey == "" || ir.RightKey == "" {
		return nil, fmt.Errorf("left_key and right_key must not be empty")
	}
	if len(ir.CompareMappings) == 0 {
		return nil, fmt.Errorf("compare mappings must not be empty")
	}

	leftIndex := stringHeaderIndex(leftRows[0])
	rightIndex := stringHeaderIndex(rightRows[0])
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
			leftValue := valueAt(left.values, leftIndex[mapping.LeftColumn])
			rightValue := valueAt(right.values, rightIndex[mapping.RightColumn])
			if leftValue == rightValue {
				continue
			}
			mismatched = true
			field := mapping.As
			if field == "" {
				field = mapping.LeftColumn
			}
			outRows = append(outRows, []string{
				"value_mismatch",
				key,
				fmt.Sprintf("%d", left.rowNumber),
				fmt.Sprintf("%d", right.rowNumber),
				field,
				leftValue,
				rightValue,
			})
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
		key := valueAt(row, keyIndex)
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
