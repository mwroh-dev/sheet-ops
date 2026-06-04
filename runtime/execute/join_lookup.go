package execute

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mwroh/sheet-ops/runtime/compiler"
	"github.com/xuri/excelize/v2"
)

func RunJoinLookup(ir compiler.WorkbookOperationIR, inputWorkbook, outputWorkbook string) (ExecutionResult, error) {
	if err := validateJoinLookupExecutionBoundary(ir, inputWorkbook, outputWorkbook); err != nil {
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

	outRows, err := buildJoinLookupRows(file, ir)
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

func validateJoinLookupExecutionBoundary(ir compiler.WorkbookOperationIR, inputWorkbook, outputWorkbook string) error {
	if inputWorkbook == "" {
		return fmt.Errorf("input workbook must not be empty")
	}
	if outputWorkbook == "" {
		return fmt.Errorf("output workbook must not be empty")
	}
	if !ir.PreserveOriginal {
		return fmt.Errorf("join_lookup execution requires preserve_original=true")
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

func buildJoinLookupRows(file *excelize.File, ir compiler.WorkbookOperationIR) ([][]string, error) {
	leftRows, err := file.GetRows(ir.SourceSheet)
	if err != nil {
		return nil, fmt.Errorf("read source sheet %q: %w", ir.SourceSheet, err)
	}
	rightRows, err := file.GetRows(ir.LookupSheet)
	if err != nil {
		return nil, fmt.Errorf("read lookup sheet %q: %w", ir.LookupSheet, err)
	}
	if len(leftRows) == 0 || len(rightRows) == 0 {
		return nil, fmt.Errorf("join lookup sheets must not be empty")
	}

	leftIndex := stringHeaderIndex(leftRows[0])
	rightIndex := stringHeaderIndex(rightRows[0])
	leftKeyIndex, ok := leftIndex[ir.JoinKey]
	if !ok {
		return nil, fmt.Errorf("source join key %q missing", ir.JoinKey)
	}
	rightKeyIndex, ok := rightIndex[ir.JoinKey]
	if !ok {
		return nil, fmt.Errorf("lookup join key %q missing", ir.JoinKey)
	}
	for _, column := range ir.IncludeSourceColumns {
		if _, ok := leftIndex[column]; !ok {
			return nil, fmt.Errorf("source column %q missing", column)
		}
	}
	for _, column := range ir.AppendLookupColumns {
		if _, ok := rightIndex[column]; !ok {
			return nil, fmt.Errorf("lookup column %q missing", column)
		}
	}

	lookup := make(map[string][]string, len(rightRows)-1)
	for _, row := range rightRows[1:] {
		key := valueAt(row, rightKeyIndex)
		if key != "" {
			lookup[key] = row
		}
	}

	header := append([]string(nil), ir.IncludeSourceColumns...)
	header = append(header, ir.AppendLookupColumns...)
	outRows := [][]string{header}
	for _, row := range leftRows[1:] {
		out := make([]string, 0, len(header))
		for _, column := range ir.IncludeSourceColumns {
			out = append(out, valueAt(row, leftIndex[column]))
		}
		right := lookup[valueAt(row, leftKeyIndex)]
		for _, column := range ir.AppendLookupColumns {
			out = append(out, valueAt(right, rightIndex[column]))
		}
		outRows = append(outRows, out)
	}
	return outRows, nil
}

func stringHeaderIndex(header []string) map[string]int {
	index := make(map[string]int, len(header))
	for idx, value := range header {
		index[value] = idx
	}
	return index
}

func valueAt(row []string, index int) string {
	if index < 0 || index >= len(row) {
		return ""
	}
	return row[index]
}
