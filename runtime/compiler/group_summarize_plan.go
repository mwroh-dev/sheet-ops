package compiler

import (
	"fmt"
	"strings"

	"github.com/xuri/excelize/v2"
)

type FormulaExpectation struct {
	Cell     string   `json:"cell"`
	Function string   `json:"function"`
	Args     []string `json:"args"`
}

func (expectation FormulaExpectation) Render() string {
	return "=" + expectation.Function + "(" + strings.Join(expectation.Args, ",") + ")"
}

type GroupSummarizePlan struct {
	Inspection          WorkbookInspection   `json:"inspection"`
	Operation           WorkbookOperationIR  `json:"operation"`
	SourceRows          [][]string           `json:"source_rows"`
	SummaryRows         [][]any              `json:"summary_rows"`
	FormulaExpectations []FormulaExpectation `json:"formula_expectations,omitempty"`
}

func BuildGroupSummarizePlan(inspection WorkbookInspection, operation WorkbookOperationIR) (GroupSummarizePlan, error) {
	file, err := excelize.OpenFile(inspection.InputWorkbook)
	if err != nil {
		return GroupSummarizePlan{}, err
	}
	defer func() { _ = file.Close() }()

	sourceRows, err := file.GetRows(inspection.SourceSheet)
	if err != nil {
		return GroupSummarizePlan{}, err
	}
	summaryRows, err := BuildSummaryRows(sourceRows, inspection.HeaderIndex, operation)
	if err != nil {
		return GroupSummarizePlan{}, err
	}

	plan := GroupSummarizePlan{
		Inspection:  inspection,
		Operation:   operation,
		SourceRows:  cloneStringMatrix(sourceRows),
		SummaryRows: cloneAnyMatrix(summaryRows),
	}
	if operation.SummaryMode == "formulas" {
		expectations, err := buildFormulaExpectations(inspection, operation, summaryRows)
		if err != nil {
			return GroupSummarizePlan{}, err
		}
		plan.FormulaExpectations = expectations
	}
	return plan, nil
}

func buildFormulaExpectations(inspection WorkbookInspection, operation WorkbookOperationIR, summaryRows [][]any) ([]FormulaExpectation, error) {
	expectations := []FormulaExpectation{}
	for rowIndex, row := range summaryRows {
		if rowIndex == 0 {
			continue
		}
		excelRow := rowIndex + 1
		for colIndex := len(operation.GroupBy); colIndex < len(row); colIndex++ {
			cell, _ := excelize.CoordinatesToCellName(colIndex+1, excelRow)
			metricIndex := colIndex - len(operation.GroupBy)
			expectation, err := buildFormulaExpectation(inspection, operation, cell, excelRow, metricIndex)
			if err != nil {
				return nil, err
			}
			expectations = append(expectations, expectation)
		}
	}
	return expectations, nil
}

func buildFormulaExpectation(inspection WorkbookInspection, operation WorkbookOperationIR, cell string, summaryRow, metricIndex int) (FormulaExpectation, error) {
	if metricIndex < 0 || metricIndex >= len(operation.Metrics) {
		return FormulaExpectation{}, fmt.Errorf("invalid metric index %d", metricIndex)
	}

	metric := operation.Metrics[metricIndex]
	endRow := inspection.RowCount
	if endRow < 2 {
		endRow = 2
	}

	args := []string{}
	function := ""
	if metric.Op == "sum" {
		function = "SUMIFS"
		args = append(args, formulaSourceRange(operation.SourceSheet, inspection.HeaderIndex, metric.Column, 2, endRow))
	}
	for groupIndex, groupBy := range operation.GroupBy {
		args = append(args,
			formulaSourceRange(operation.SourceSheet, inspection.HeaderIndex, groupBy, 2, endRow),
			formulaSummaryReference(groupIndex+1, summaryRow),
		)
	}
	for _, filter := range operation.Filters {
		criteria, err := formulaCriteria(filter.Op, filter.Value)
		if err != nil {
			return FormulaExpectation{}, err
		}
		args = append(args,
			formulaSourceRange(operation.SourceSheet, inspection.HeaderIndex, filter.Column, 2, endRow),
			criteria,
		)
	}

	switch metric.Op {
	case "count":
		function = "COUNTIFS"
		args = append(args,
			formulaSourceRange(operation.SourceSheet, inspection.HeaderIndex, metric.Column, 2, endRow),
			"\"<>\"",
		)
	case "sum":
	default:
		return FormulaExpectation{}, fmt.Errorf("unsupported metric op %q", metric.Op)
	}

	return FormulaExpectation{
		Cell:     cell,
		Function: function,
		Args:     append([]string(nil), args...),
	}, nil
}

func formulaSourceRange(sheetName string, headerIndex map[string]int, columnName string, startRow, endRow int) string {
	columnNumber := headerIndex[columnName] + 1
	columnNameExcel, _ := excelize.ColumnNumberToName(columnNumber)
	return fmt.Sprintf("%s!$%s$%d:$%s$%d", formulaEscapeSheetName(sheetName), columnNameExcel, startRow, columnNameExcel, endRow)
}

func formulaSummaryReference(columnNumber, rowNumber int) string {
	columnName, _ := excelize.ColumnNumberToName(columnNumber)
	return fmt.Sprintf("$%s%d", columnName, rowNumber)
}

func formulaCriteria(operator, value string) (string, error) {
	escaped := strings.ReplaceAll(value, `"`, `""`)
	switch operator {
	case "!=", "<>":
		return fmt.Sprintf(`"<>%s"`, escaped), nil
	case "=", "==":
		return fmt.Sprintf(`"%s"`, escaped), nil
	default:
		return "", fmt.Errorf("unsupported filter op %q for formula summary", operator)
	}
}

func formulaEscapeSheetName(sheetName string) string {
	escaped := strings.ReplaceAll(sheetName, "'", "''")
	return fmt.Sprintf("'%s'", escaped)
}

func cloneStringMatrix(values [][]string) [][]string {
	if len(values) == 0 {
		return nil
	}
	cloned := make([][]string, len(values))
	for rowIndex, row := range values {
		cloned[rowIndex] = append([]string(nil), row...)
	}
	return cloned
}

func cloneAnyMatrix(values [][]any) [][]any {
	if len(values) == 0 {
		return nil
	}
	cloned := make([][]any, len(values))
	for rowIndex, row := range values {
		cloned[rowIndex] = append([]any(nil), row...)
	}
	return cloned
}
