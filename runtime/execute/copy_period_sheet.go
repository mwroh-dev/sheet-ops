package execute

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mwroh/sheet-ops/runtime/compiler"
	"github.com/xuri/excelize/v2"
)

func RunCopyPeriodSheet(ir compiler.WorkbookOperationIR, inputWorkbook, outputWorkbook string) (ExecutionResult, error) {
	if err := validateCopyPeriodSheetBoundary(ir, inputWorkbook, outputWorkbook); err != nil {
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

	sourceIndex, err := file.GetSheetIndex(ir.SourceSheet)
	if err != nil {
		return ExecutionResult{}, err
	}
	if sourceIndex < 0 {
		return ExecutionResult{}, fmt.Errorf("source sheet %q does not exist", ir.SourceSheet)
	}
	if targetIndex, err := file.GetSheetIndex(ir.TargetSheet); err != nil {
		return ExecutionResult{}, err
	} else if targetIndex >= 0 {
		return ExecutionResult{}, fmt.Errorf("target sheet %q already exists", ir.TargetSheet)
	}
	targetIndex, err := file.NewSheet(ir.TargetSheet)
	if err != nil {
		return ExecutionResult{}, err
	}
	if err := file.CopySheet(sourceIndex, targetIndex); err != nil {
		return ExecutionResult{}, err
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
		SourceSHA256Before: sourceBefore,
		SourceSHA256After:  sourceAfter,
	}, nil
}

func validateCopyPeriodSheetBoundary(ir compiler.WorkbookOperationIR, inputWorkbook, outputWorkbook string) error {
	if inputWorkbook == "" {
		return fmt.Errorf("input workbook must not be empty")
	}
	if outputWorkbook == "" {
		return fmt.Errorf("output workbook must not be empty")
	}
	if !ir.PreserveOriginal {
		return fmt.Errorf("copy_period_sheet execution requires preserve_original=true")
	}
	if ir.SourceSheet == "" {
		return fmt.Errorf("source sheet must not be empty")
	}
	if ir.TargetSheet == "" {
		return fmt.Errorf("target sheet must not be empty")
	}
	if ir.SourceSheet == ir.TargetSheet {
		return fmt.Errorf("target sheet must differ from source sheet")
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
