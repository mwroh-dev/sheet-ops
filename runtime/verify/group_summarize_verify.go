package verify

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/mwroh/sheet-ops/runtime/compiler"
	"github.com/xuri/excelize/v2"
)

type VerificationResult struct {
	Pass                bool          `json:"pass"`
	OperationFamily     string        `json:"operation_family"`
	OutputWorkbook      string        `json:"output_workbook"`
	SummarySheet        string        `json:"summary_sheet,omitempty"`
	SummaryMode         string        `json:"summary_mode,omitempty"`
	SummaryRows         int           `json:"summary_rows,omitempty"`
	FormulaCells        []string      `json:"formula_cells,omitempty"`
	HighlightedRows     []int         `json:"highlighted_rows,omitempty"`
	WrittenCells        []string      `json:"written_cells,omitempty"`
	Layers              []LayerResult `json:"layers,omitempty"`
	Reasons             []string      `json:"reasons"`
	SourceSHA256Checked bool          `json:"-"`
}

type LayerResult struct {
	Level   int      `json:"level"`
	Name    string   `json:"name"`
	Pass    bool     `json:"pass"`
	Reasons []string `json:"reasons,omitempty"`
}

func VerifyWorkbookOperation(ir compiler.WorkbookOperationIR, inputWorkbook, outputWorkbook, sourceSHA256Before, sourceSHA256After string) (VerificationResult, error) {
	switch ir.OperationFamily {
	case "group_summarize":
		result, err := VerifyGroupSummarize(ir, inputWorkbook, outputWorkbook, sourceSHA256Before, sourceSHA256After)
		if err != nil {
			return VerificationResult{}, err
		}
		return WithDefaultLayers(ir, result), nil
	case "highlight_threshold":
		result, err := VerifyHighlightThreshold(ir, inputWorkbook, outputWorkbook, sourceSHA256Before, sourceSHA256After)
		if err != nil {
			return VerificationResult{}, err
		}
		return WithDefaultLayers(ir, result), nil
	case "join_lookup":
		result, err := VerifyJoinLookup(ir, inputWorkbook, outputWorkbook, sourceSHA256Before, sourceSHA256After)
		if err != nil {
			return VerificationResult{}, err
		}
		return WithDefaultLayers(ir, result), nil
	case "append_structured_rows":
		result, err := VerifyAppendStructuredRows(ir, inputWorkbook, outputWorkbook, sourceSHA256Before, sourceSHA256After)
		if err != nil {
			return VerificationResult{}, err
		}
		return WithDefaultLayers(ir, result), nil
	case "write_values":
		result, err := VerifyWriteValues(ir, inputWorkbook, outputWorkbook, sourceSHA256Before, sourceSHA256After)
		if err != nil {
			return VerificationResult{}, err
		}
		return WithDefaultLayers(ir, result), nil
	default:
		return VerificationResult{}, fmt.Errorf("unsupported operation family %q", ir.OperationFamily)
	}
}

func VerifyGroupSummarize(ir compiler.WorkbookOperationIR, inputWorkbook, outputWorkbook, sourceSHA256Before, sourceSHA256After string) (VerificationResult, error) {
	inspection, err := compiler.InspectWorkbook(inputWorkbook, ir.SourceSheet, ir.Filters, ir.GroupBy, ir.Metrics)
	if err != nil {
		return VerificationResult{}, err
	}
	plan, err := compiler.BuildGroupSummarizePlan(inspection, ir)
	if err != nil {
		return VerificationResult{}, err
	}
	return VerifyGroupSummarizeExecution(plan, outputWorkbook, sourceSHA256Before, sourceSHA256After)
}

func VerifyGroupSummarizeExecution(plan compiler.GroupSummarizePlan, outputWorkbook, sourceSHA256Before, sourceSHA256After string) (VerificationResult, error) {
	if sourceSHA256Before == "" || sourceSHA256After == "" {
		return VerificationResult{
			Pass:            false,
			OperationFamily: plan.Operation.OperationFamily,
			OutputWorkbook:  outputWorkbook,
			SummarySheet:    plan.Operation.TargetSheet,
			SummaryMode:     plan.Operation.SummaryMode,
			Reasons:         []string{"execution hash evidence is missing"},
		}, nil
	}
	if sourceSHA256Before != sourceSHA256After {
		return VerificationResult{
			Pass:            false,
			OperationFamily: plan.Operation.OperationFamily,
			OutputWorkbook:  outputWorkbook,
			SummarySheet:    plan.Operation.TargetSheet,
			SummaryMode:     plan.Operation.SummaryMode,
			Reasons:         []string{"source workbook changed during execution"},
		}, nil
	}

	currentHash, err := hashFile(plan.Inspection.InputWorkbook)
	if err != nil {
		return VerificationResult{}, err
	}
	if currentHash != sourceSHA256Before {
		return VerificationResult{
			Pass:            false,
			OperationFamily: plan.Operation.OperationFamily,
			OutputWorkbook:  outputWorkbook,
			SummarySheet:    plan.Operation.TargetSheet,
			SummaryMode:     plan.Operation.SummaryMode,
			Reasons:         []string{"source workbook hash differs from execution evidence"},
		}, nil
	}

	return VerifyGroupSummarizePlan(plan, outputWorkbook)
}

func VerifyGroupSummarizePlan(plan compiler.GroupSummarizePlan, outputWorkbook string) (VerificationResult, error) {
	outputFileHandle, err := excelize.OpenFile(outputWorkbook)
	if err != nil {
		return VerificationResult{}, err
	}
	defer func() { _ = outputFileHandle.Close() }()

	if !containsString(outputFileHandle.GetSheetList(), plan.Operation.TargetSheet) {
		return VerificationResult{
			Pass:            false,
			OperationFamily: plan.Operation.OperationFamily,
			OutputWorkbook:  outputWorkbook,
			SummarySheet:    plan.Operation.TargetSheet,
			SummaryMode:     plan.Operation.SummaryMode,
			Reasons:         []string{fmt.Sprintf("summary sheet %q was not created", plan.Operation.TargetSheet)},
		}, nil
	}

	outputSourceRows, err := outputFileHandle.GetRows(plan.Operation.SourceSheet)
	if err != nil {
		return VerificationResult{}, err
	}
	if err := compareStringRows(plan.Operation.SourceSheet, plan.SourceRows, outputSourceRows); err != nil {
		return VerificationResult{
			Pass:            false,
			OperationFamily: plan.Operation.OperationFamily,
			OutputWorkbook:  outputWorkbook,
			SummarySheet:    plan.Operation.TargetSheet,
			SummaryMode:     plan.Operation.SummaryMode,
			Reasons:         []string{err.Error()},
		}, nil
	}

	actualRows, err := outputFileHandle.GetRows(plan.Operation.TargetSheet)
	if err != nil {
		return VerificationResult{}, err
	}

	switch plan.Operation.SummaryMode {
	case "values":
		if err := verifyValueSummaryHasNoMetricFormulas(outputFileHandle, plan); err != nil {
			return VerificationResult{
				Pass:            false,
				OperationFamily: plan.Operation.OperationFamily,
				OutputWorkbook:  outputWorkbook,
				SummarySheet:    plan.Operation.TargetSheet,
				SummaryMode:     plan.Operation.SummaryMode,
				SummaryRows:     summaryRowCount(actualRows),
				Reasons:         []string{err.Error()},
			}, nil
		}
		if err := compiler.CompareSummaryRows(plan.SummaryRows, actualRows); err != nil {
			return VerificationResult{
				Pass:            false,
				OperationFamily: plan.Operation.OperationFamily,
				OutputWorkbook:  outputWorkbook,
				SummarySheet:    plan.Operation.TargetSheet,
				SummaryMode:     plan.Operation.SummaryMode,
				SummaryRows:     summaryRowCount(actualRows),
				Reasons:         []string{err.Error()},
			}, nil
		}
		return VerificationResult{
			Pass:            true,
			OperationFamily: plan.Operation.OperationFamily,
			OutputWorkbook:  outputWorkbook,
			SummarySheet:    plan.Operation.TargetSheet,
			SummaryMode:     plan.Operation.SummaryMode,
			SummaryRows:     summaryRowCount(actualRows),
			Reasons:         []string{"output workbook contains the expected summary sheet and source worksheet is unchanged"},
		}, nil
	case "formulas":
		formulaCells, err := verifyFormulaSummary(outputFileHandle, plan)
		if err != nil {
			return VerificationResult{
				Pass:            false,
				OperationFamily: plan.Operation.OperationFamily,
				OutputWorkbook:  outputWorkbook,
				SummarySheet:    plan.Operation.TargetSheet,
				SummaryMode:     plan.Operation.SummaryMode,
				SummaryRows:     summaryRowCount(actualRows),
				Reasons:         []string{err.Error()},
			}, nil
		}
		return VerificationResult{
			Pass:            true,
			OperationFamily: plan.Operation.OperationFamily,
			OutputWorkbook:  outputWorkbook,
			SummarySheet:    plan.Operation.TargetSheet,
			SummaryMode:     plan.Operation.SummaryMode,
			SummaryRows:     summaryRowCount(actualRows),
			FormulaCells:    append([]string(nil), formulaCells...),
			Reasons:         []string{"output workbook contains the expected formula-linked summary sheet and source worksheet is unchanged"},
		}, nil
	default:
		return VerificationResult{}, fmt.Errorf("unsupported summary mode %q", plan.Operation.SummaryMode)
	}
}

func verifyValueSummaryHasNoMetricFormulas(file *excelize.File, plan compiler.GroupSummarizePlan) error {
	for rowIndex, row := range plan.SummaryRows {
		if rowIndex == 0 {
			continue
		}
		excelRow := rowIndex + 1
		for colIndex := len(plan.Operation.GroupBy); colIndex < len(row); colIndex++ {
			cell, _ := excelize.CoordinatesToCellName(colIndex+1, excelRow)
			formula, err := file.GetCellFormula(plan.Operation.TargetSheet, cell)
			if err != nil {
				return err
			}
			if formula != "" {
				return fmt.Errorf("unexpected formula in values-mode metric cell %s", cell)
			}
		}
	}
	return nil
}

func verifyFormulaSummary(file *excelize.File, plan compiler.GroupSummarizePlan) ([]string, error) {
	expectationByCell := make(map[string]compiler.FormulaExpectation, len(plan.FormulaExpectations))
	for _, expectation := range plan.FormulaExpectations {
		expectationByCell[expectation.Cell] = expectation
	}

	formulaCells := []string{}
	expectedFormulaCount := 0
	for rowIndex, row := range plan.SummaryRows {
		excelRow := rowIndex + 1
		for colIndex, value := range row {
			cell, _ := excelize.CoordinatesToCellName(colIndex+1, excelRow)
			expectFormula := rowIndex > 0 && colIndex >= len(plan.Operation.GroupBy)
			if !expectFormula {
				actualFormula, err := file.GetCellFormula(plan.Operation.TargetSheet, cell)
				if err != nil {
					return nil, err
				}
				if actualFormula != "" {
					return nil, fmt.Errorf("unexpected formula in literal cell %s", cell)
				}
				actual, err := file.GetCellValue(plan.Operation.TargetSheet, cell)
				if err != nil {
					return nil, err
				}
				if actual != stringifyCell(value) {
					return nil, fmt.Errorf("summary cell %s=%q want %q", cell, actual, stringifyCell(value))
				}
				continue
			}
			expectedFormulaCount++

			expectation, hasFormula := expectationByCell[cell]
			if !hasFormula {
				return nil, fmt.Errorf("missing formula expectation for %s", cell)
			}
			actualFormula, err := file.GetCellFormula(plan.Operation.TargetSheet, cell)
			if err != nil {
				return nil, err
			}
			if err := verifyMetricFormula(actualFormula, expectation); err != nil {
				return nil, fmt.Errorf("summary formula %s: %w", cell, err)
			}

			calculated, err := file.CalcCellValue(plan.Operation.TargetSheet, cell)
			if err != nil {
				return nil, err
			}
			if calculated != stringifyCell(value) {
				return nil, fmt.Errorf("summary formula value %s=%q want %q", cell, calculated, stringifyCell(value))
			}
			formulaCells = append(formulaCells, cell)
		}
	}
	if len(plan.FormulaExpectations) != expectedFormulaCount {
		return nil, fmt.Errorf("formula expectation count=%d want %d", len(plan.FormulaExpectations), expectedFormulaCount)
	}
	return formulaCells, nil
}

func verifyMetricFormula(formula string, expected compiler.FormulaExpectation) error {
	name, args, err := parseFormulaCall(formula)
	if err != nil {
		return err
	}
	if name != expected.Function {
		return fmt.Errorf("function=%q want %q", name, expected.Function)
	}
	if len(args) != len(expected.Args) {
		return fmt.Errorf("arg count=%d want %d", len(args), len(expected.Args))
	}
	for index := range expected.Args {
		if args[index] != expected.Args[index] {
			return fmt.Errorf("arg[%d]=%q want %q", index, args[index], expected.Args[index])
		}
	}
	return nil
}

func parseFormulaCall(formula string) (string, []string, error) {
	if !strings.HasPrefix(formula, "=") {
		return "", nil, fmt.Errorf("missing leading '=' in %q", formula)
	}
	openIndex := strings.Index(formula, "(")
	closeIndex := strings.LastIndex(formula, ")")
	if openIndex <= 1 || closeIndex <= openIndex {
		return "", nil, fmt.Errorf("malformed formula %q", formula)
	}

	name := formula[1:openIndex]
	args, err := splitFormulaArgs(formula[openIndex+1 : closeIndex])
	if err != nil {
		return "", nil, err
	}
	return name, args, nil
}

func splitFormulaArgs(raw string) ([]string, error) {
	if raw == "" {
		return nil, nil
	}

	args := []string{}
	var token strings.Builder
	inDouble := false
	inSingle := false

	for index := 0; index < len(raw); index++ {
		ch := raw[index]
		switch ch {
		case '"':
			token.WriteByte(ch)
			if inSingle {
				continue
			}
			if inDouble && index+1 < len(raw) && raw[index+1] == '"' {
				index++
				token.WriteByte(raw[index])
				continue
			}
			inDouble = !inDouble
		case '\'':
			token.WriteByte(ch)
			if inDouble {
				continue
			}
			if inSingle && index+1 < len(raw) && raw[index+1] == '\'' {
				index++
				token.WriteByte(raw[index])
				continue
			}
			inSingle = !inSingle
		case ',':
			if inDouble || inSingle {
				token.WriteByte(ch)
				continue
			}
			args = append(args, strings.TrimSpace(token.String()))
			token.Reset()
		default:
			token.WriteByte(ch)
		}
	}

	if inDouble || inSingle {
		return nil, fmt.Errorf("unterminated quoted argument list %q", raw)
	}
	args = append(args, strings.TrimSpace(token.String()))
	return args, nil
}

func summaryRowCount(rows [][]string) int {
	if len(rows) == 0 {
		return 0
	}
	return len(rows) - 1
}

func compareStringRows(sheetName string, expected, actual [][]string) error {
	if len(expected) != len(actual) {
		return fmt.Errorf("source sheet %q row count=%d want %d", sheetName, len(actual), len(expected))
	}
	for rowIndex := range expected {
		if len(expected[rowIndex]) != len(actual[rowIndex]) {
			return fmt.Errorf("source sheet %q row %d column count=%d want %d", sheetName, rowIndex, len(actual[rowIndex]), len(expected[rowIndex]))
		}
		for colIndex := range expected[rowIndex] {
			if expected[rowIndex][colIndex] != actual[rowIndex][colIndex] {
				return fmt.Errorf("source sheet %q cell[%d][%d]=%q want %q", sheetName, rowIndex, colIndex, actual[rowIndex][colIndex], expected[rowIndex][colIndex])
			}
		}
	}
	return nil
}

func containsString(values []string, needle string) bool {
	for _, value := range values {
		if value == needle {
			return true
		}
	}
	return false
}

func hashFile(path string) (string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:]), nil
}

func stringifyCell(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case int:
		return strconv.Itoa(typed)
	case int64:
		return strconv.FormatInt(typed, 10)
	case float64:
		return strconv.FormatFloat(typed, 'f', -1, 64)
	default:
		return fmt.Sprint(typed)
	}
}
