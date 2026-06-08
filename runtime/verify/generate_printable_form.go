package verify

import (
	"fmt"

	"github.com/mwroh/sheet-ops/runtime/compiler"
	"github.com/xuri/excelize/v2"
)

func VerifyGeneratePrintableForm(ir compiler.WorkbookOperationIR, inputWorkbook, outputWorkbook, sourceSHA256Before, sourceSHA256After string) (VerificationResult, error) {
	if sourceSHA256Before == "" || sourceSHA256After == "" {
		return printableFormFailure(ir, outputWorkbook, "execution hash evidence is missing"), nil
	}
	if sourceSHA256Before != sourceSHA256After {
		return printableFormFailure(ir, outputWorkbook, "source workbook changed during execution"), nil
	}
	currentHash, err := hashFile(inputWorkbook)
	if err != nil {
		return VerificationResult{}, err
	}
	if currentHash != sourceSHA256Before {
		return printableFormFailure(ir, outputWorkbook, "source workbook hash differs from execution evidence"), nil
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
		return printableFormFailure(ir, outputWorkbook, err.Error()), nil
	}
	if !containsString(outputFile.GetSheetList(), ir.TargetSheet) {
		return printableFormFailure(ir, outputWorkbook, fmt.Sprintf("printable form sheet %q was not created", ir.TargetSheet)), nil
	}
	writtenCells, err := verifyPrintableFormCells(inputFile, outputFile, ir)
	if err != nil {
		return printableFormFailure(ir, outputWorkbook, err.Error()), nil
	}
	if !printAreaDefined(outputFile.GetDefinedName(), ir.TargetSheet, ir.PrintArea) {
		return printableFormFailure(ir, outputWorkbook, fmt.Sprintf("print area %q missing for %q", ir.PrintArea, ir.TargetSheet)), nil
	}
	return VerificationResult{
		Pass:            true,
		OperationFamily: ir.OperationFamily,
		OutputWorkbook:  outputWorkbook,
		SummarySheet:    ir.TargetSheet,
		SummaryRows:     len(writtenCells),
		WrittenCells:    writtenCells,
		Reasons:         []string{"output workbook contains expected printable form cells, print area, and unchanged source sheets"},
	}, nil
}

func verifyPrintableFormCells(inputFile, outputFile *excelize.File, ir compiler.WorkbookOperationIR) ([]string, error) {
	written := []string{}
	if err := verifyCell(outputFile, ir.TargetSheet, "A1", ir.FormTitle); err != nil {
		return nil, err
	}
	written = append(written, ir.TargetSheet+"!A1")
	for _, binding := range ir.FieldBindings {
		sourceSheet := binding.SourceSheet
		if sourceSheet == "" {
			sourceSheet = ir.SourceSheet
		}
		sourceValue, err := inputFile.GetCellValue(sourceSheet, binding.SourceCell)
		if err != nil {
			return nil, err
		}
		if err := verifyCell(outputFile, ir.TargetSheet, binding.LabelCell, binding.Label); err != nil {
			return nil, err
		}
		if err := verifyCell(outputFile, ir.TargetSheet, binding.ValueCell, sourceValue); err != nil {
			return nil, err
		}
		written = append(written, ir.TargetSheet+"!"+binding.LabelCell, ir.TargetSheet+"!"+binding.ValueCell)
	}
	tableCells, err := verifyPrintableTable(inputFile, outputFile, ir)
	if err != nil {
		return nil, err
	}
	written = append(written, tableCells...)
	return written, nil
}

func verifyPrintableTable(inputFile, outputFile *excelize.File, ir compiler.WorkbookOperationIR) ([]string, error) {
	table := ir.TableBinding
	rows, err := inputFile.GetRows(table.SourceSheet)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("table source sheet %q is empty", table.SourceSheet)
	}
	headerIndex := map[string]int{}
	for idx, column := range rows[0] {
		headerIndex[column] = idx
	}
	headerCol, headerRow, err := excelize.CellNameToCoordinates(table.HeaderStart)
	if err != nil {
		return nil, err
	}
	dataCol, dataRow, err := excelize.CellNameToCoordinates(table.DataStart)
	if err != nil {
		return nil, err
	}
	written := []string{}
	for idx, column := range table.SourceColumns {
		sourceIndex, ok := headerIndex[column]
		if !ok {
			return nil, fmt.Errorf("table source column %q missing", column)
		}
		cell, _ := excelize.CoordinatesToCellName(headerCol+idx, headerRow)
		if err := verifyCell(outputFile, ir.TargetSheet, cell, column); err != nil {
			return nil, err
		}
		written = append(written, ir.TargetSheet+"!"+cell)
		for rowIndex, row := range rows[1:] {
			dataCell, _ := excelize.CoordinatesToCellName(dataCol+idx, dataRow+rowIndex)
			if err := verifyCell(outputFile, ir.TargetSheet, dataCell, printableValueAt(row, sourceIndex)); err != nil {
				return nil, err
			}
			written = append(written, ir.TargetSheet+"!"+dataCell)
		}
	}
	return written, nil
}

func verifyCell(file *excelize.File, sheet, cell, want string) error {
	got, err := file.GetCellValue(sheet, cell)
	if err != nil {
		return err
	}
	if got != want {
		return fmt.Errorf("%s!%s=%q want %q", sheet, cell, got, want)
	}
	return nil
}

func printAreaDefined(values []excelize.DefinedName, sheet, printArea string) bool {
	refersTo := fmt.Sprintf("'%s'!%s", sheet, printableAbsoluteRange(printArea))
	for _, value := range values {
		if value.Name == "_xlnm.Print_Area" && value.Scope == sheet && value.RefersTo == refersTo {
			return true
		}
	}
	return false
}

func printableAbsoluteRange(value string) string {
	start, end := value, value
	for index, char := range value {
		if char == ':' {
			start = value[:index]
			end = value[index+1:]
			break
		}
	}
	return printableAbsoluteCell(start) + ":" + printableAbsoluteCell(end)
}

func printableAbsoluteCell(value string) string {
	col, row, err := excelize.CellNameToCoordinates(value)
	if err != nil {
		return value
	}
	cell, _ := excelize.CoordinatesToCellName(col, row)
	letters := ""
	digits := ""
	for _, char := range cell {
		if char >= '0' && char <= '9' {
			digits += string(char)
		} else {
			letters += string(char)
		}
	}
	return "$" + letters + "$" + digits
}

func printableValueAt(row []string, index int) string {
	if index < 0 || index >= len(row) {
		return ""
	}
	return row[index]
}

func printableFormFailure(ir compiler.WorkbookOperationIR, outputWorkbook string, reason string) VerificationResult {
	return VerificationResult{
		Pass:            false,
		OperationFamily: ir.OperationFamily,
		OutputWorkbook:  outputWorkbook,
		SummarySheet:    ir.TargetSheet,
		Reasons:         []string{reason},
	}
}
