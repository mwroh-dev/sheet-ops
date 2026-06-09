package execute

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mwroh/sheet-ops/runtime/compiler"
	"github.com/xuri/excelize/v2"
)

type ExecutionResult struct {
	OperationFamily    string   `json:"operation_family"`
	OutputWorkbook     string   `json:"output_workbook"`
	SummarySheet       string   `json:"summary_sheet,omitempty"`
	SummaryMode        string   `json:"summary_mode,omitempty"`
	SummaryRows        int      `json:"summary_rows,omitempty"`
	FormulaCells       []string `json:"formula_cells,omitempty"`
	HighlightedRows    []int    `json:"highlighted_rows,omitempty"`
	HighlightedCells   []string `json:"highlighted_cells,omitempty"`
	WrittenCells       []string `json:"written_cells,omitempty"`
	SourceSHA256Before string   `json:"source_sha256_before"`
	SourceSHA256After  string   `json:"source_sha256_after"`
}

type summarySheetWriter func(*excelize.File, compiler.GroupSummarizePlan) ([]string, error)

func RunWorkbookOperation(ir compiler.WorkbookOperationIR, inputWorkbook, outputWorkbook string) (ExecutionResult, error) {
	switch ir.OperationFamily {
	case "group_summarize":
		switch ir.SummaryMode {
		case "values":
			return RunGroupSummarizeValues(ir, inputWorkbook, outputWorkbook)
		case "formulas":
			return RunGroupSummarizeFormulas(ir, inputWorkbook, outputWorkbook)
		default:
			return ExecutionResult{}, fmt.Errorf("unsupported summary mode %q", ir.SummaryMode)
		}
	case "highlight_threshold":
		return RunHighlightThreshold(ir, inputWorkbook, outputWorkbook)
	case "join_lookup":
		return RunJoinLookup(ir, inputWorkbook, outputWorkbook)
	case "append_structured_rows":
		return RunAppendStructuredRows(ir, inputWorkbook, outputWorkbook)
	case "extend_table_formulas":
		return RunExtendTableFormulas(ir, inputWorkbook, outputWorkbook)
	case "add_data_validation":
		return RunAddDataValidation(ir, inputWorkbook, outputWorkbook)
	case "protect_formula_cells":
		return RunProtectFormulaCells(ir, inputWorkbook, outputWorkbook)
	case "copy_period_sheet":
		return RunCopyPeriodSheet(ir, inputWorkbook, outputWorkbook)
	case "normalize_headers":
		return RunNormalizeHeaders(ir, inputWorkbook, outputWorkbook)
	case "roll_forward_period":
		return RunRollForwardPeriod(ir, inputWorkbook, outputWorkbook)
	case "reconcile_tables":
		return RunReconcileTables(ir, inputWorkbook, outputWorkbook)
	case "generate_printable_form":
		return RunGeneratePrintableForm(ir, inputWorkbook, outputWorkbook)
	case "write_values":
		return RunWriteValues(ir, inputWorkbook, outputWorkbook)
	default:
		return ExecutionResult{}, fmt.Errorf("unsupported operation family %q", ir.OperationFamily)
	}
}

func RunGroupSummarizeValues(ir compiler.WorkbookOperationIR, inputWorkbook, outputWorkbook string) (ExecutionResult, error) {
	inspection, err := compiler.InspectWorkbook(inputWorkbook, ir.SourceSheet, ir.Filters, ir.GroupBy, ir.Metrics)
	if err != nil {
		return ExecutionResult{}, err
	}
	plan, err := compiler.BuildGroupSummarizePlan(inspection, ir)
	if err != nil {
		return ExecutionResult{}, err
	}
	return RunGroupSummarizePlan(plan, outputWorkbook)
}

func RunGroupSummarizePlan(plan compiler.GroupSummarizePlan, outputWorkbook string) (ExecutionResult, error) {
	if err := validateExecutionBoundary(plan, outputWorkbook); err != nil {
		return ExecutionResult{}, err
	}

	var writer summarySheetWriter
	switch plan.Operation.SummaryMode {
	case "values":
		writer = writeValueSummarySheet
	case "formulas":
		writer = writeFormulaSummarySheet
	default:
		return ExecutionResult{}, fmt.Errorf("unsupported summary mode %q", plan.Operation.SummaryMode)
	}
	return runGroupSummarizePlan(plan, outputWorkbook, writer)
}

func runGroupSummarizePlan(plan compiler.GroupSummarizePlan, outputWorkbook string, writer summarySheetWriter) (ExecutionResult, error) {
	sourceBefore, err := hashFile(plan.Inspection.InputWorkbook)
	if err != nil {
		return ExecutionResult{}, err
	}

	file, err := excelize.OpenFile(plan.Inspection.InputWorkbook)
	if err != nil {
		return ExecutionResult{}, err
	}
	defer func() { _ = file.Close() }()

	if containsString(file.GetSheetList(), plan.Operation.TargetSheet) {
		file.DeleteSheet(plan.Operation.TargetSheet)
	}
	file.NewSheet(plan.Operation.TargetSheet)

	formulaCells, err := writer(file, plan)
	if err != nil {
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

	sourceAfter, err := hashFile(plan.Inspection.InputWorkbook)
	if err != nil {
		return ExecutionResult{}, err
	}

	return ExecutionResult{
		OperationFamily:    plan.Operation.OperationFamily,
		OutputWorkbook:     outputWorkbook,
		SummarySheet:       plan.Operation.TargetSheet,
		SummaryMode:        plan.Operation.SummaryMode,
		SummaryRows:        summaryRowCount(plan.SummaryRows),
		FormulaCells:       append([]string(nil), formulaCells...),
		SourceSHA256Before: sourceBefore,
		SourceSHA256After:  sourceAfter,
	}, nil
}

func validateExecutionBoundary(plan compiler.GroupSummarizePlan, outputWorkbook string) error {
	if outputWorkbook == "" {
		return fmt.Errorf("output workbook must not be empty")
	}
	if !plan.Operation.PreserveOriginal {
		return fmt.Errorf("group_summarize execution requires preserve_original=true")
	}
	if plan.Inspection.InputWorkbook == "" {
		return fmt.Errorf("input workbook must not be empty")
	}

	same, err := sameWorkbookPath(plan.Inspection.InputWorkbook, outputWorkbook)
	if err != nil {
		return err
	}
	if same {
		return fmt.Errorf("output workbook must differ from input workbook")
	}
	return nil
}

func sameWorkbookPath(left, right string) (bool, error) {
	leftInfo, leftErr := os.Stat(left)
	if leftErr != nil && !errors.Is(leftErr, os.ErrNotExist) {
		return false, leftErr
	}
	rightInfo, rightErr := os.Stat(right)
	if rightErr != nil && !errors.Is(rightErr, os.ErrNotExist) {
		return false, rightErr
	}
	if leftErr == nil && rightErr == nil && os.SameFile(leftInfo, rightInfo) {
		return true, nil
	}

	leftAbs, err := filepath.Abs(left)
	if err != nil {
		return false, err
	}
	rightAbs, err := filepath.Abs(right)
	if err != nil {
		return false, err
	}
	leftEval, err := filepath.EvalSymlinks(leftAbs)
	if err == nil {
		leftAbs = leftEval
	}
	rightEval, err := filepath.EvalSymlinks(rightAbs)
	if err == nil {
		rightAbs = rightEval
	}
	return filepath.Clean(leftAbs) == filepath.Clean(rightAbs), nil
}

func writeValueSummarySheet(file *excelize.File, plan compiler.GroupSummarizePlan) ([]string, error) {
	for rowIndex, row := range plan.SummaryRows {
		for colIndex, value := range row {
			cell, _ := excelize.CoordinatesToCellName(colIndex+1, rowIndex+1)
			if err := file.SetCellValue(plan.Operation.TargetSheet, cell, value); err != nil {
				return nil, err
			}
		}
	}
	return nil, nil
}

func summaryRowCount(summaryRows [][]any) int {
	if len(summaryRows) == 0 {
		return 0
	}
	return len(summaryRows) - 1
}

func hashFile(path string) (string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:]), nil
}

func containsString(values []string, needle string) bool {
	for _, value := range values {
		if value == needle {
			return true
		}
	}
	return false
}
