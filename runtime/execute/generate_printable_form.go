package execute

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mwroh/sheet-ops/runtime/compiler"
	"github.com/xuri/excelize/v2"
)

func RunGeneratePrintableForm(ir compiler.WorkbookOperationIR, inputWorkbook, outputWorkbook string) (ExecutionResult, error) {
	if err := validateGeneratePrintableFormBoundary(ir, inputWorkbook, outputWorkbook); err != nil {
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

	writtenCells, err := writePrintableForm(file, ir)
	if err != nil {
		return ExecutionResult{}, err
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
		SummaryRows:        len(writtenCells),
		WrittenCells:       writtenCells,
		SourceSHA256Before: sourceBefore,
		SourceSHA256After:  sourceAfter,
	}, nil
}

func validateGeneratePrintableFormBoundary(ir compiler.WorkbookOperationIR, inputWorkbook, outputWorkbook string) error {
	if inputWorkbook == "" {
		return fmt.Errorf("input workbook must not be empty")
	}
	if outputWorkbook == "" {
		return fmt.Errorf("output workbook must not be empty")
	}
	if !ir.PreserveOriginal {
		return fmt.Errorf("generate_printable_form execution requires preserve_original=true")
	}
	if ir.TargetSheet == "" {
		return fmt.Errorf("target sheet must not be empty")
	}
	if ir.TargetSheet == ir.SourceSheet {
		return fmt.Errorf("target sheet must differ from source sheet")
	}
	if ir.FormTitle == "" || ir.PrintArea == "" {
		return fmt.Errorf("form_title and print_area must not be empty")
	}
	if len(ir.FieldBindings) == 0 {
		return fmt.Errorf("field bindings must not be empty")
	}
	if ir.TableBinding == nil || len(ir.TableBinding.SourceColumns) == 0 {
		return fmt.Errorf("table binding source columns must not be empty")
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

func writePrintableForm(file *excelize.File, ir compiler.WorkbookOperationIR) ([]string, error) {
	if containsString(file.GetSheetList(), ir.TargetSheet) {
		file.DeleteSheet(ir.TargetSheet)
	}
	if _, err := file.NewSheet(ir.TargetSheet); err != nil {
		return nil, err
	}
	written := []string{}
	if err := file.SetCellValue(ir.TargetSheet, "A1", ir.FormTitle); err != nil {
		return nil, err
	}
	written = append(written, ir.TargetSheet+"!A1")
	for _, binding := range ir.FieldBindings {
		sourceSheet := binding.SourceSheet
		if sourceSheet == "" {
			sourceSheet = ir.SourceSheet
		}
		value, err := file.GetCellValue(sourceSheet, binding.SourceCell)
		if err != nil {
			return nil, err
		}
		if err := file.SetCellValue(ir.TargetSheet, binding.LabelCell, binding.Label); err != nil {
			return nil, err
		}
		if err := file.SetCellValue(ir.TargetSheet, binding.ValueCell, value); err != nil {
			return nil, err
		}
		written = append(written, ir.TargetSheet+"!"+binding.LabelCell, ir.TargetSheet+"!"+binding.ValueCell)
	}
	tableCells, err := writePrintableTable(file, ir)
	if err != nil {
		return nil, err
	}
	written = append(written, tableCells...)
	if err := file.SetDefinedName(&excelize.DefinedName{
		Name:     "_xlnm.Print_Area",
		RefersTo: fmt.Sprintf("'%s'!%s", ir.TargetSheet, absoluteRange(ir.PrintArea)),
		Scope:    ir.TargetSheet,
	}); err != nil {
		return nil, err
	}
	return written, nil
}

func writePrintableTable(file *excelize.File, ir compiler.WorkbookOperationIR) ([]string, error) {
	table := ir.TableBinding
	rows, err := file.GetRows(table.SourceSheet)
	if err != nil {
		return nil, fmt.Errorf("read table source sheet %q: %w", table.SourceSheet, err)
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("table source sheet %q is empty", table.SourceSheet)
	}
	headerIndex := stringHeaderIndex(rows[0])
	for _, column := range table.SourceColumns {
		if _, ok := headerIndex[column]; !ok {
			return nil, fmt.Errorf("table source column %q missing", column)
		}
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
		cell, _ := excelize.CoordinatesToCellName(headerCol+idx, headerRow)
		if err := file.SetCellValue(ir.TargetSheet, cell, column); err != nil {
			return nil, err
		}
		written = append(written, ir.TargetSheet+"!"+cell)
	}
	for rowIndex, row := range rows[1:] {
		for columnIndex, column := range table.SourceColumns {
			cell, _ := excelize.CoordinatesToCellName(dataCol+columnIndex, dataRow+rowIndex)
			if err := file.SetCellValue(ir.TargetSheet, cell, valueAt(row, headerIndex[column])); err != nil {
				return nil, err
			}
			written = append(written, ir.TargetSheet+"!"+cell)
		}
	}
	return written, nil
}

func absoluteRange(value string) string {
	start, end := value, value
	for index, char := range value {
		if char == ':' {
			start = value[:index]
			end = value[index+1:]
			break
		}
	}
	return absoluteCell(start) + ":" + absoluteCell(end)
}

func absoluteCell(value string) string {
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
