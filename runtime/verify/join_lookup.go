package verify

import (
	"fmt"

	"github.com/mwroh/sheet-ops/runtime/compiler"
	"github.com/xuri/excelize/v2"
)

func VerifyJoinLookup(ir compiler.WorkbookOperationIR, inputWorkbook, outputWorkbook, sourceSHA256Before, sourceSHA256After string) (VerificationResult, error) {
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
	expectedRows, err := buildExpectedJoinLookupRows(inputFile, ir)
	if err != nil {
		return VerificationResult{}, err
	}

	outputFile, err := excelize.OpenFile(outputWorkbook)
	if err != nil {
		return VerificationResult{}, err
	}
	defer func() { _ = outputFile.Close() }()
	inputSourceRows, err := inputFile.GetRows(ir.SourceSheet)
	if err != nil {
		return VerificationResult{}, err
	}
	outputSourceRows, err := outputFile.GetRows(ir.SourceSheet)
	if err != nil {
		return VerificationResult{
			Pass:            false,
			OperationFamily: ir.OperationFamily,
			OutputWorkbook:  outputWorkbook,
			SummarySheet:    ir.TargetSheet,
			Reasons:         []string{fmt.Sprintf("source sheet %q is missing from output workbook", ir.SourceSheet)},
		}, nil
	}
	if err := compareStringRows(ir.SourceSheet, inputSourceRows, outputSourceRows); err != nil {
		return VerificationResult{
			Pass:            false,
			OperationFamily: ir.OperationFamily,
			OutputWorkbook:  outputWorkbook,
			SummarySheet:    ir.TargetSheet,
			Reasons:         []string{err.Error()},
		}, nil
	}
	inputLookupRows, err := inputFile.GetRows(ir.LookupSheet)
	if err != nil {
		return VerificationResult{}, err
	}
	outputLookupRows, err := outputFile.GetRows(ir.LookupSheet)
	if err != nil {
		return VerificationResult{
			Pass:            false,
			OperationFamily: ir.OperationFamily,
			OutputWorkbook:  outputWorkbook,
			SummarySheet:    ir.TargetSheet,
			Reasons:         []string{fmt.Sprintf("lookup sheet %q is missing from output workbook", ir.LookupSheet)},
		}, nil
	}
	if err := compareStringRows(ir.LookupSheet, inputLookupRows, outputLookupRows); err != nil {
		return VerificationResult{
			Pass:            false,
			OperationFamily: ir.OperationFamily,
			OutputWorkbook:  outputWorkbook,
			SummarySheet:    ir.TargetSheet,
			Reasons:         []string{err.Error()},
		}, nil
	}
	if !containsString(outputFile.GetSheetList(), ir.TargetSheet) {
		return VerificationResult{
			Pass:            false,
			OperationFamily: ir.OperationFamily,
			OutputWorkbook:  outputWorkbook,
			SummarySheet:    ir.TargetSheet,
			Reasons:         []string{fmt.Sprintf("join lookup sheet %q was not created", ir.TargetSheet)},
		}, nil
	}
	actualRows, err := outputFile.GetRows(ir.TargetSheet)
	if err != nil {
		return VerificationResult{}, err
	}
	if err := compareStringRows(ir.TargetSheet, expectedRows, actualRows); err != nil {
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
		Reasons:         []string{"output workbook contains the expected joined lookup sheet and source workbook is unchanged"},
	}, nil
}

func buildExpectedJoinLookupRows(file *excelize.File, ir compiler.WorkbookOperationIR) ([][]string, error) {
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

	leftIndex := joinLookupHeaderIndex(leftRows[0])
	rightIndex := joinLookupHeaderIndex(rightRows[0])
	leftKeyIndex, ok := leftIndex[ir.JoinKey]
	if !ok {
		return nil, fmt.Errorf("source join key %q missing", ir.JoinKey)
	}
	rightKeyIndex, ok := rightIndex[ir.JoinKey]
	if !ok {
		return nil, fmt.Errorf("lookup join key %q missing", ir.JoinKey)
	}

	lookup := make(map[string][]string, len(rightRows)-1)
	for _, row := range rightRows[1:] {
		key := joinLookupValueAt(row, rightKeyIndex)
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
			out = append(out, joinLookupValueAt(row, leftIndex[column]))
		}
		right := lookup[joinLookupValueAt(row, leftKeyIndex)]
		for _, column := range ir.AppendLookupColumns {
			out = append(out, joinLookupValueAt(right, rightIndex[column]))
		}
		outRows = append(outRows, out)
	}
	return outRows, nil
}

func joinLookupHeaderIndex(header []string) map[string]int {
	index := make(map[string]int, len(header))
	for idx, value := range header {
		index[value] = idx
	}
	return index
}

func joinLookupValueAt(row []string, index int) string {
	if index < 0 || index >= len(row) {
		return ""
	}
	return row[index]
}
