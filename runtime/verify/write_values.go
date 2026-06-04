package verify

import (
	"fmt"

	"github.com/mwroh/sheet-ops/runtime/compiler"
	"github.com/xuri/excelize/v2"
)

func VerifyWriteValues(ir compiler.WorkbookOperationIR, inputWorkbook, outputWorkbook, sourceSHA256Before, sourceSHA256After string) (VerificationResult, error) {
	if sourceSHA256Before == "" || sourceSHA256After == "" {
		return writeValuesFailure(ir, outputWorkbook, "execution hash evidence is missing"), nil
	}
	if sourceSHA256Before != sourceSHA256After {
		return writeValuesFailure(ir, outputWorkbook, "source workbook changed during execution"), nil
	}
	currentHash, err := hashFile(inputWorkbook)
	if err != nil {
		return VerificationResult{}, err
	}
	if currentHash != sourceSHA256Before {
		return writeValuesFailure(ir, outputWorkbook, "source workbook hash differs from execution evidence"), nil
	}

	inputFile, err := excelize.OpenFile(inputWorkbook)
	if err != nil {
		return VerificationResult{}, err
	}
	defer func() { _ = inputFile.Close() }()
	outputFile, err := excelize.OpenFile(outputWorkbook)
	if err != nil {
		return VerificationResult{}, err
	}
	defer func() { _ = outputFile.Close() }()

	if err := verifyWriteValuesSourcePreserved(inputFile, outputFile); err != nil {
		return writeValuesFailure(ir, outputWorkbook, err.Error()), nil
	}
	writtenCells := make([]string, 0, len(ir.Values))
	for _, cellValue := range ir.Values {
		got, err := outputFile.GetCellValue(ir.TargetSheet, cellValue.Cell)
		if err != nil {
			return VerificationResult{}, err
		}
		want := expectedWriteValueString(cellValue.Value)
		if got != want {
			return writeValuesFailure(ir, outputWorkbook, fmt.Sprintf("%s!%s=%q want %q", ir.TargetSheet, cellValue.Cell, got, want)), nil
		}
		writtenCells = append(writtenCells, ir.TargetSheet+"!"+cellValue.Cell)
	}

	return VerificationResult{
		Pass:            true,
		OperationFamily: ir.OperationFamily,
		OutputWorkbook:  outputWorkbook,
		SummarySheet:    ir.TargetSheet,
		WrittenCells:    writtenCells,
		Reasons:         []string{"output workbook contains expected written values and source workbook is unchanged"},
	}, nil
}

func expectedWriteValueString(value any) string {
	switch typed := value.(type) {
	case nil:
		return ""
	case bool:
		if typed {
			return "TRUE"
		}
		return "FALSE"
	default:
		return fmt.Sprint(value)
	}
}

func verifyWriteValuesSourcePreserved(inputFile, outputFile *excelize.File) error {
	for _, sheet := range inputFile.GetSheetList() {
		inputRows, err := inputFile.GetRows(sheet)
		if err != nil {
			return err
		}
		outputRows, err := outputFile.GetRows(sheet)
		if err != nil {
			return fmt.Errorf("source sheet %q is missing from output workbook", sheet)
		}
		if err := compareStringRows(sheet, inputRows, outputRows); err != nil {
			return err
		}
	}
	return nil
}

func writeValuesFailure(ir compiler.WorkbookOperationIR, outputWorkbook string, reason string) VerificationResult {
	return VerificationResult{
		Pass:            false,
		OperationFamily: ir.OperationFamily,
		OutputWorkbook:  outputWorkbook,
		SummarySheet:    ir.TargetSheet,
		Reasons:         []string{reason},
	}
}
