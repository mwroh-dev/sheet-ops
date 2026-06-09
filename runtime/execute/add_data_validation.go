package execute

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mwroh/sheet-ops/runtime/compiler"
	"github.com/xuri/excelize/v2"
)

func RunAddDataValidation(ir compiler.WorkbookOperationIR, inputWorkbook, outputWorkbook string) (ExecutionResult, error) {
	if err := validateAddDataValidationBoundary(ir, inputWorkbook, outputWorkbook); err != nil {
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

	writtenCells := []string{}
	for _, targetRange := range ir.ValidationRule.Ranges {
		validation := excelize.NewDataValidation(ir.ValidationRule.AllowBlank)
		validation.SetSqref(targetRange)
		if err := validation.SetDropList(ir.ValidationRule.AllowedValues); err != nil {
			return ExecutionResult{}, err
		}
		if err := file.AddDataValidation(ir.SourceSheet, validation); err != nil {
			return ExecutionResult{}, err
		}
		writtenCells = append(writtenCells, targetRange)
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
		WrittenCells:       writtenCells,
		SourceSHA256Before: sourceBefore,
		SourceSHA256After:  sourceAfter,
	}, nil
}

func validateAddDataValidationBoundary(ir compiler.WorkbookOperationIR, inputWorkbook, outputWorkbook string) error {
	if inputWorkbook == "" {
		return fmt.Errorf("input workbook must not be empty")
	}
	if outputWorkbook == "" {
		return fmt.Errorf("output workbook must not be empty")
	}
	if !ir.PreserveOriginal {
		return fmt.Errorf("add_data_validation execution requires preserve_original=true")
	}
	if ir.SourceSheet == "" {
		return fmt.Errorf("source sheet must not be empty")
	}
	if ir.ValidationRule == nil || len(ir.ValidationRule.Ranges) == 0 {
		return fmt.Errorf("validation rule ranges must not be empty")
	}
	if ir.ValidationRule.RuleType != "list" {
		return fmt.Errorf("unsupported validation rule type %q", ir.ValidationRule.RuleType)
	}
	if len(ir.ValidationRule.AllowedValues) == 0 {
		return fmt.Errorf("allowed values must not be empty")
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
