package execute

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mwroh/sheet-ops/runtime/compiler"
	runtimeformula "github.com/mwroh/sheet-ops/runtime/formula"
	"github.com/xuri/excelize/v2"
)

func RunExtendTableFormulas(ir compiler.WorkbookOperationIR, inputWorkbook, outputWorkbook string) (ExecutionResult, error) {
	if err := validateExtendTableFormulasBoundary(ir, inputWorkbook, outputWorkbook); err != nil {
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

	formulaCells := []string{}
	for _, column := range ir.FormulaColumns {
		sourceCell, err := formulaCell(column, ir.FormulaSourceRow)
		if err != nil {
			return ExecutionResult{}, err
		}
		sourceFormula, err := file.GetCellFormula(ir.SourceSheet, sourceCell)
		if err != nil {
			return ExecutionResult{}, err
		}
		if strings.TrimSpace(sourceFormula) == "" {
			return ExecutionResult{}, fmt.Errorf("%s!%s has no formula to extend", ir.SourceSheet, sourceCell)
		}
		if runtimeformula.IsFormulaWithUnsupportedExtensionSyntax(sourceFormula) {
			return ExecutionResult{}, fmt.Errorf("%s!%s uses unsupported formula extension syntax", ir.SourceSheet, sourceCell)
		}
		for _, targetRow := range ir.TargetRows {
			targetCell, err := formulaCell(column, targetRow)
			if err != nil {
				return ExecutionResult{}, err
			}
			translated, err := runtimeformula.TranslateFormulaRows(sourceFormula, targetRow-ir.FormulaSourceRow)
			if err != nil {
				return ExecutionResult{}, err
			}
			if err := file.SetCellFormula(ir.SourceSheet, targetCell, translated); err != nil {
				return ExecutionResult{}, err
			}
			formulaCells = append(formulaCells, targetCell)
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
		SummarySheet:       ir.SourceSheet,
		FormulaCells:       append([]string(nil), formulaCells...),
		SourceSHA256Before: sourceBefore,
		SourceSHA256After:  sourceAfter,
	}, nil
}

func validateExtendTableFormulasBoundary(ir compiler.WorkbookOperationIR, inputWorkbook, outputWorkbook string) error {
	if inputWorkbook == "" {
		return fmt.Errorf("input workbook must not be empty")
	}
	if outputWorkbook == "" {
		return fmt.Errorf("output workbook must not be empty")
	}
	if !ir.PreserveOriginal {
		return fmt.Errorf("extend_table_formulas execution requires preserve_original=true")
	}
	if ir.SourceSheet == "" {
		return fmt.Errorf("source sheet must not be empty")
	}
	if ir.FormulaSourceRow < 1 {
		return fmt.Errorf("formula source row must be positive")
	}
	if len(ir.TargetRows) == 0 {
		return fmt.Errorf("target rows must not be empty")
	}
	if len(ir.FormulaColumns) == 0 {
		return fmt.Errorf("formula columns must not be empty")
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

func formulaCell(column string, row int) (string, error) {
	if row < 1 {
		return "", fmt.Errorf("row must be positive")
	}
	normalizedColumn := strings.ToUpper(strings.TrimSpace(column))
	if normalizedColumn == "" {
		return "", fmt.Errorf("formula column must not be empty")
	}
	columnNumber, err := excelize.ColumnNameToNumber(normalizedColumn)
	if err != nil {
		return "", err
	}
	return excelize.CoordinatesToCellName(columnNumber, row)
}
