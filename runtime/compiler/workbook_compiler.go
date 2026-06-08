package compiler

import (
	"fmt"
	"strconv"
	"strings"

	runtimeinspect "github.com/mwroh/sheet-ops/runtime/inspect"
	"github.com/mwroh/sheet-ops/runtime/taskspec"
)

type FilterSpec = taskspec.FilterSpec
type MetricSpec = taskspec.MetricSpec

type CellValue struct {
	Cell  string `json:"cell"`
	Value any    `json:"value"`
}

type DataValidationRule struct {
	Ranges        []string `json:"ranges"`
	RuleType      string   `json:"rule_type"`
	AllowedValues []string `json:"allowed_values,omitempty"`
	AllowBlank    bool     `json:"allow_blank"`
}

type FormulaProtectionRule struct {
	FormulaRanges []string `json:"formula_ranges"`
	InputRanges   []string `json:"input_ranges,omitempty"`
	Password      string   `json:"password,omitempty"`
}

type WorkbookInspection struct {
	InputWorkbook     string                  `json:"input_workbook"`
	SheetNames        []string                `json:"sheet_names"`
	SourceSheet       string                  `json:"source_sheet"`
	HeaderNames       []string                `json:"header_names"`
	HeaderIndex       map[string]int          `json:"-"`
	TargetColumn      string                  `json:"target_column,omitempty"`
	TargetColumnIndex int                     `json:"target_column_index,omitempty"`
	HeaderRow         int                     `json:"header_row,omitempty"`
	RowCount          int                     `json:"row_count"`
	Authority         GroupSummarizeAuthority `json:"-"`
}

type WorkbookOperationIR struct {
	ExecutionKind        string                 `json:"execution_kind"`
	CompositionKind      string                 `json:"composition_kind,omitempty"`
	OperationFamily      string                 `json:"operation_family"`
	SourceSheet          string                 `json:"source_sheet"`
	Filters              []FilterSpec           `json:"filters,omitempty"`
	GroupBy              []string               `json:"group_by,omitempty"`
	Metrics              []MetricSpec           `json:"metrics,omitempty"`
	TargetSheet          string                 `json:"target_sheet,omitempty"`
	SummaryMode          string                 `json:"summary_mode,omitempty"`
	TargetColumn         string                 `json:"target_column,omitempty"`
	Operator             string                 `json:"operator,omitempty"`
	Threshold            *float64               `json:"threshold,omitempty"`
	HighlightColor       string                 `json:"highlight_color,omitempty"`
	LookupSheet          string                 `json:"lookup_sheet,omitempty"`
	JoinKey              string                 `json:"join_key,omitempty"`
	IncludeSourceColumns []string               `json:"include_source_columns,omitempty"`
	AppendLookupColumns  []string               `json:"append_lookup_columns,omitempty"`
	Values               []CellValue            `json:"values,omitempty"`
	FormulaSourceRow     int                    `json:"formula_source_row,omitempty"`
	TargetRows           []int                  `json:"target_rows,omitempty"`
	FormulaColumns       []string               `json:"formula_columns,omitempty"`
	ValidationRule       *DataValidationRule    `json:"validation_rule,omitempty"`
	ProtectionRule       *FormulaProtectionRule `json:"protection_rule,omitempty"`
	PreserveOriginal     bool                   `json:"preserve_original"`
}

func InspectWorkbook(inputWorkbook, sourceSheet string, filters []FilterSpec, groupBy []string, metrics []MetricSpec) (WorkbookInspection, error) {
	inspection, err := runtimeinspect.GroupSummarizeWorkbook(inputWorkbook, sourceSheet, filters, groupBy, metrics)
	if err != nil {
		return WorkbookInspection{}, err
	}

	return WorkbookInspection{
		InputWorkbook: inspection.InputWorkbook,
		SheetNames:    append([]string(nil), inspection.SheetNames...),
		SourceSheet:   inspection.SourceSheet,
		HeaderNames:   append([]string(nil), inspection.HeaderNames...),
		HeaderIndex:   cloneHeaderIndex(inspection.HeaderIndex),
		HeaderRow:     1,
		RowCount:      inspection.RowCount,
		Authority:     groupSummarizeAuthorityFromInspection(inputWorkbook, sourceSheet, filters, groupBy, metrics),
	}, nil
}

func InspectHighlightThresholdWorkbook(inputWorkbook, sourceSheet, targetColumn string) (WorkbookInspection, error) {
	inspection, err := runtimeinspect.HighlightThresholdWorkbook(inputWorkbook, sourceSheet, targetColumn)
	if err != nil {
		return WorkbookInspection{}, err
	}

	return WorkbookInspection{
		InputWorkbook:     inspection.InputWorkbook,
		SheetNames:        append([]string(nil), inspection.SheetNames...),
		SourceSheet:       inspection.SourceSheet,
		HeaderNames:       append([]string(nil), inspection.HeaderNames...),
		HeaderIndex:       cloneHeaderIndex(inspection.HeaderIndex),
		TargetColumn:      inspection.TargetColumn,
		TargetColumnIndex: inspection.TargetColumnIndex,
		HeaderRow:         inspection.HeaderRow,
		RowCount:          inspection.RowCount,
	}, nil
}

func CompileWorkbookOperation(task taskspec.GroupSummarizeTask, inspection WorkbookInspection) (WorkbookOperationIR, error) {
	if task.SourceSheet == "" {
		return WorkbookOperationIR{}, fmt.Errorf("source sheet must not be empty")
	}
	if task.TargetSheet == "" {
		return WorkbookOperationIR{}, fmt.Errorf("target sheet must not be empty")
	}
	if task.TargetSheet == inspection.SourceSheet {
		return WorkbookOperationIR{}, fmt.Errorf("target sheet must differ from source sheet")
	}
	if len(task.GroupBy) == 0 {
		return WorkbookOperationIR{}, fmt.Errorf("group by columns must not be empty")
	}
	if len(task.Metrics) == 0 {
		return WorkbookOperationIR{}, fmt.Errorf("metrics must not be empty")
	}
	if err := validateGroupSummarizeAuthority(task, inspection); err != nil {
		return WorkbookOperationIR{}, err
	}

	summaryMode := task.TaskSpec.SummaryMode
	if summaryMode == "" {
		summaryMode = "values"
	}

	return WorkbookOperationIR{
		ExecutionKind:    taskspec.ExecutionKindComposition,
		CompositionKind:  taskspec.CompositionKindGroupSummary,
		OperationFamily:  taskspec.OperationFamilyGroupSummarize,
		SourceSheet:      inspection.SourceSheet,
		Filters:          cloneFilters(task.Filters),
		GroupBy:          cloneStrings(task.GroupBy),
		Metrics:          cloneMetrics(task.Metrics),
		TargetSheet:      task.TargetSheet,
		SummaryMode:      summaryMode,
		PreserveOriginal: task.PreserveOriginal,
	}, nil
}

func CompileHighlightThresholdOperation(task taskspec.HighlightThresholdTask, inspection WorkbookInspection) (WorkbookOperationIR, error) {
	if task.SourceSheet == "" {
		return WorkbookOperationIR{}, fmt.Errorf("source sheet must not be empty")
	}
	if task.TaskSpec.OutputWorkbook == "" {
		return WorkbookOperationIR{}, fmt.Errorf("output workbook must not be empty")
	}
	if task.ThresholdRule.Column == "" {
		return WorkbookOperationIR{}, fmt.Errorf("threshold column must not be empty")
	}
	if inspection.SourceSheet != task.SourceSheet {
		return WorkbookOperationIR{}, fmt.Errorf("inspection source sheet %q does not match task source sheet %q", inspection.SourceSheet, task.SourceSheet)
	}
	if inspection.TargetColumn == "" {
		return WorkbookOperationIR{}, fmt.Errorf("inspection target column must not be empty")
	}
	if !strings.EqualFold(strings.TrimSpace(inspection.TargetColumn), strings.TrimSpace(task.ThresholdRule.Column)) {
		return WorkbookOperationIR{}, fmt.Errorf("inspection target column %q does not match task column %q", inspection.TargetColumn, task.ThresholdRule.Column)
	}
	if err := validateHighlightTaskCompositionBoundary(task.TaskSpec); err != nil {
		return WorkbookOperationIR{}, err
	}
	if err := validateHighlightOperator(task.ThresholdRule.Operator); err != nil {
		return WorkbookOperationIR{}, err
	}

	threshold := task.ThresholdRule.Threshold
	return WorkbookOperationIR{
		ExecutionKind:    taskspec.ExecutionKindComposition,
		CompositionKind:  taskspec.CompositionKindThresholdHighlight,
		OperationFamily:  taskspec.OperationFamilyHighlightThreshold,
		SourceSheet:      inspection.SourceSheet,
		TargetColumn:     inspection.TargetColumn,
		Operator:         task.ThresholdRule.Operator,
		Threshold:        &threshold,
		HighlightColor:   task.ThresholdRule.HighlightColor,
		PreserveOriginal: task.PreserveOriginal,
	}, nil
}

func CompileJoinLookupOperation(task taskspec.JoinLookupTask) (WorkbookOperationIR, error) {
	if task.SourceSheet == "" {
		return WorkbookOperationIR{}, fmt.Errorf("source sheet must not be empty")
	}
	if task.LookupSheet == "" {
		return WorkbookOperationIR{}, fmt.Errorf("lookup sheet must not be empty")
	}
	if task.JoinKey == "" {
		return WorkbookOperationIR{}, fmt.Errorf("join key must not be empty")
	}
	if task.TargetSheet == "" {
		return WorkbookOperationIR{}, fmt.Errorf("target sheet must not be empty")
	}
	if task.TargetSheet == task.SourceSheet || task.TargetSheet == task.LookupSheet {
		return WorkbookOperationIR{}, fmt.Errorf("target sheet must differ from source and lookup sheets")
	}
	if len(task.IncludeSourceColumns) == 0 {
		return WorkbookOperationIR{}, fmt.Errorf("include source columns must not be empty")
	}
	if len(task.AppendLookupColumns) == 0 {
		return WorkbookOperationIR{}, fmt.Errorf("append lookup columns must not be empty")
	}
	if err := validateJoinLookupTaskCompositionBoundary(task.TaskSpec); err != nil {
		return WorkbookOperationIR{}, err
	}

	return WorkbookOperationIR{
		ExecutionKind:        taskspec.ExecutionKindComposition,
		CompositionKind:      taskspec.CompositionKindJoinLookup,
		OperationFamily:      taskspec.OperationFamilyJoinLookup,
		SourceSheet:          task.SourceSheet,
		LookupSheet:          task.LookupSheet,
		JoinKey:              task.JoinKey,
		TargetSheet:          task.TargetSheet,
		IncludeSourceColumns: cloneStrings(task.IncludeSourceColumns),
		AppendLookupColumns:  cloneStrings(task.AppendLookupColumns),
		PreserveOriginal:     task.PreserveOriginal,
	}, nil
}

func CompileAppendStructuredRowsOperation(task taskspec.AppendStructuredRowsTask) (WorkbookOperationIR, error) {
	if task.SourceSheet == "" {
		return WorkbookOperationIR{}, fmt.Errorf("source sheet must not be empty")
	}
	if len(task.IncludeSourceColumns) == 0 {
		return WorkbookOperationIR{}, fmt.Errorf("include source columns must not be empty")
	}
	if len(task.Values) == 0 {
		return WorkbookOperationIR{}, fmt.Errorf("values must not be empty")
	}
	if err := validateAppendStructuredRowsTaskCompositionBoundary(task.TaskSpec); err != nil {
		return WorkbookOperationIR{}, err
	}
	values := make([]CellValue, 0, len(task.Values))
	for _, value := range task.Values {
		values = append(values, CellValue{Cell: value.Cell, Value: value.Value})
	}

	return WorkbookOperationIR{
		ExecutionKind:        taskspec.ExecutionKindComposition,
		CompositionKind:      taskspec.CompositionKindStructuredRowAppend,
		OperationFamily:      taskspec.OperationFamilyAppendStructuredRows,
		SourceSheet:          task.SourceSheet,
		IncludeSourceColumns: cloneStrings(task.IncludeSourceColumns),
		Values:               values,
		PreserveOriginal:     task.PreserveOriginal,
	}, nil
}

func CompileExtendTableFormulasOperation(task taskspec.ExtendTableFormulasTask) (WorkbookOperationIR, error) {
	if task.SourceSheet == "" {
		return WorkbookOperationIR{}, fmt.Errorf("source sheet must not be empty")
	}
	if task.FormulaSourceRow < 1 {
		return WorkbookOperationIR{}, fmt.Errorf("formula source row must be positive")
	}
	if len(task.TargetRows) == 0 {
		return WorkbookOperationIR{}, fmt.Errorf("target rows must not be empty")
	}
	for _, row := range task.TargetRows {
		if row < 1 {
			return WorkbookOperationIR{}, fmt.Errorf("target row must be positive")
		}
		if row == task.FormulaSourceRow {
			return WorkbookOperationIR{}, fmt.Errorf("target row must differ from formula source row")
		}
	}
	if len(task.FormulaColumns) == 0 {
		return WorkbookOperationIR{}, fmt.Errorf("formula columns must not be empty")
	}
	if err := validateExtendTableFormulasTaskCompositionBoundary(task.TaskSpec); err != nil {
		return WorkbookOperationIR{}, err
	}

	return WorkbookOperationIR{
		ExecutionKind:    taskspec.ExecutionKindComposition,
		CompositionKind:  taskspec.CompositionKindFormulaExtension,
		OperationFamily:  taskspec.OperationFamilyExtendTableFormulas,
		SourceSheet:      task.SourceSheet,
		FormulaSourceRow: task.FormulaSourceRow,
		TargetRows:       cloneInts(task.TargetRows),
		FormulaColumns:   cloneStrings(task.FormulaColumns),
		PreserveOriginal: task.PreserveOriginal,
	}, nil
}

func CompileAddDataValidationOperation(task taskspec.AddDataValidationTask) (WorkbookOperationIR, error) {
	if task.SourceSheet == "" {
		return WorkbookOperationIR{}, fmt.Errorf("source sheet must not be empty")
	}
	if len(task.ValidationRule.Ranges) == 0 {
		return WorkbookOperationIR{}, fmt.Errorf("validation ranges must not be empty")
	}
	if task.ValidationRule.RuleType != "list" {
		return WorkbookOperationIR{}, fmt.Errorf("unsupported validation rule type %q", task.ValidationRule.RuleType)
	}
	if len(task.ValidationRule.AllowedValues) == 0 {
		return WorkbookOperationIR{}, fmt.Errorf("allowed values must not be empty for list validation")
	}
	if err := validateAddDataValidationTaskCompositionBoundary(task.TaskSpec); err != nil {
		return WorkbookOperationIR{}, err
	}
	return WorkbookOperationIR{
		ExecutionKind:    taskspec.ExecutionKindComposition,
		CompositionKind:  taskspec.CompositionKindDataValidation,
		OperationFamily:  taskspec.OperationFamilyAddDataValidation,
		SourceSheet:      task.SourceSheet,
		ValidationRule:   cloneDataValidationRule(task.ValidationRule),
		PreserveOriginal: task.PreserveOriginal,
	}, nil
}

func CompileProtectFormulaCellsOperation(task taskspec.ProtectFormulaCellsTask) (WorkbookOperationIR, error) {
	if task.SourceSheet == "" {
		return WorkbookOperationIR{}, fmt.Errorf("source sheet must not be empty")
	}
	if len(task.ProtectionRule.FormulaRanges) == 0 {
		return WorkbookOperationIR{}, fmt.Errorf("formula ranges must not be empty")
	}
	if err := validateProtectFormulaCellsTaskCompositionBoundary(task.TaskSpec); err != nil {
		return WorkbookOperationIR{}, err
	}
	return WorkbookOperationIR{
		ExecutionKind:    taskspec.ExecutionKindComposition,
		CompositionKind:  taskspec.CompositionKindFormulaProtection,
		OperationFamily:  taskspec.OperationFamilyProtectFormulaCells,
		SourceSheet:      task.SourceSheet,
		ProtectionRule:   cloneFormulaProtectionRule(task.ProtectionRule),
		PreserveOriginal: task.PreserveOriginal,
	}, nil
}

func validateHighlightTaskCompositionBoundary(spec taskspec.TaskSpec) error {
	if spec.ExecutionKind != taskspec.ExecutionKindComposition {
		return fmt.Errorf("highlight task execution_kind=%q want %q", spec.ExecutionKind, taskspec.ExecutionKindComposition)
	}
	if spec.CompositionKind != taskspec.CompositionKindThresholdHighlight {
		return fmt.Errorf("highlight task composition_kind=%q want %q", spec.CompositionKind, taskspec.CompositionKindThresholdHighlight)
	}
	if spec.Operation != taskspec.OperationHighlightThresholdRows {
		return fmt.Errorf("highlight task operation=%q want %q", spec.Operation, taskspec.OperationHighlightThresholdRows)
	}
	return nil
}

func validateJoinLookupTaskCompositionBoundary(spec taskspec.TaskSpec) error {
	if spec.ExecutionKind != taskspec.ExecutionKindComposition {
		return fmt.Errorf("join lookup task execution_kind=%q want %q", spec.ExecutionKind, taskspec.ExecutionKindComposition)
	}
	if spec.CompositionKind != taskspec.CompositionKindJoinLookup {
		return fmt.Errorf("join lookup task composition_kind=%q want %q", spec.CompositionKind, taskspec.CompositionKindJoinLookup)
	}
	if spec.Operation != taskspec.OperationCreateJoinLookupResultSheet {
		return fmt.Errorf("join lookup task operation=%q want %q", spec.Operation, taskspec.OperationCreateJoinLookupResultSheet)
	}
	return nil
}

func validateAppendStructuredRowsTaskCompositionBoundary(spec taskspec.TaskSpec) error {
	if spec.ExecutionKind != taskspec.ExecutionKindComposition {
		return fmt.Errorf("append structured rows task execution_kind=%q want %q", spec.ExecutionKind, taskspec.ExecutionKindComposition)
	}
	if spec.CompositionKind != taskspec.CompositionKindStructuredRowAppend {
		return fmt.Errorf("append structured rows task composition_kind=%q want %q", spec.CompositionKind, taskspec.CompositionKindStructuredRowAppend)
	}
	if spec.Operation != taskspec.OperationAppendStructuredRows {
		return fmt.Errorf("append structured rows task operation=%q want %q", spec.Operation, taskspec.OperationAppendStructuredRows)
	}
	return nil
}

func validateExtendTableFormulasTaskCompositionBoundary(spec taskspec.TaskSpec) error {
	if spec.ExecutionKind != taskspec.ExecutionKindComposition {
		return fmt.Errorf("extend table formulas task execution_kind=%q want %q", spec.ExecutionKind, taskspec.ExecutionKindComposition)
	}
	if spec.CompositionKind != taskspec.CompositionKindFormulaExtension {
		return fmt.Errorf("extend table formulas task composition_kind=%q want %q", spec.CompositionKind, taskspec.CompositionKindFormulaExtension)
	}
	if spec.Operation != taskspec.OperationExtendTableFormulas {
		return fmt.Errorf("extend table formulas task operation=%q want %q", spec.Operation, taskspec.OperationExtendTableFormulas)
	}
	return nil
}

func validateAddDataValidationTaskCompositionBoundary(spec taskspec.TaskSpec) error {
	if spec.ExecutionKind != taskspec.ExecutionKindComposition {
		return fmt.Errorf("add data validation task execution_kind=%q want %q", spec.ExecutionKind, taskspec.ExecutionKindComposition)
	}
	if spec.CompositionKind != taskspec.CompositionKindDataValidation {
		return fmt.Errorf("add data validation task composition_kind=%q want %q", spec.CompositionKind, taskspec.CompositionKindDataValidation)
	}
	if spec.Operation != taskspec.OperationAddDataValidation {
		return fmt.Errorf("add data validation task operation=%q want %q", spec.Operation, taskspec.OperationAddDataValidation)
	}
	return nil
}

func validateProtectFormulaCellsTaskCompositionBoundary(spec taskspec.TaskSpec) error {
	if spec.ExecutionKind != taskspec.ExecutionKindComposition {
		return fmt.Errorf("protect formula cells task execution_kind=%q want %q", spec.ExecutionKind, taskspec.ExecutionKindComposition)
	}
	if spec.CompositionKind != taskspec.CompositionKindFormulaProtection {
		return fmt.Errorf("protect formula cells task composition_kind=%q want %q", spec.CompositionKind, taskspec.CompositionKindFormulaProtection)
	}
	if spec.Operation != taskspec.OperationProtectFormulaCells {
		return fmt.Errorf("protect formula cells task operation=%q want %q", spec.Operation, taskspec.OperationProtectFormulaCells)
	}
	return nil
}

func BuildSummaryRows(rows [][]string, headerIndex map[string]int, ir WorkbookOperationIR) ([][]any, error) {
	if len(rows) == 0 {
		return nil, fmt.Errorf("sheet %q is empty", ir.SourceSheet)
	}

	header := make([]any, 0, len(ir.GroupBy)+len(ir.Metrics))
	for _, groupBy := range ir.GroupBy {
		header = append(header, groupBy)
	}
	for _, metric := range ir.Metrics {
		header = append(header, metric.As)
	}

	order := []string{}
	groups := map[string]*summaryGroup{}
	for rowIndex := 1; rowIndex < len(rows); rowIndex++ {
		row := rows[rowIndex]
		matched, err := rowMatchesFilters(row, headerIndex, ir.Filters)
		if err != nil {
			return nil, fmt.Errorf("row %d: %w", rowIndex+1, err)
		}
		if !matched {
			continue
		}

		groupValues := make([]string, 0, len(ir.GroupBy))
		for _, groupBy := range ir.GroupBy {
			groupValues = append(groupValues, valueAt(row, headerIndex[groupBy]))
		}
		groupKey := strings.Join(groupValues, "\x1f")
		group, ok := groups[groupKey]
		if !ok {
			group = &summaryGroup{
				groupValues: groupValues,
				sums:        make([]float64, len(ir.Metrics)),
				counts:      make([]int, len(ir.Metrics)),
			}
			groups[groupKey] = group
			order = append(order, groupKey)
		}

		for metricIndex, metric := range ir.Metrics {
			cell := valueAt(row, headerIndex[metric.Column])
			switch metric.Op {
			case "sum":
				number, err := parseNumber(cell)
				if err != nil {
					return nil, fmt.Errorf("row %d column %q is not numeric", rowIndex+1, metric.Column)
				}
				group.sums[metricIndex] += number
			case "count":
				if strings.TrimSpace(cell) != "" {
					group.counts[metricIndex]++
				}
			default:
				return nil, fmt.Errorf("unsupported metric op %q", metric.Op)
			}
		}
	}

	out := [][]any{header}
	for _, groupKey := range order {
		group := groups[groupKey]
		row := make([]any, 0, len(ir.GroupBy)+len(ir.Metrics))
		for _, value := range group.groupValues {
			row = append(row, value)
		}
		for metricIndex, metric := range ir.Metrics {
			switch metric.Op {
			case "sum":
				row = append(row, excelNumberValue(group.sums[metricIndex]))
			case "count":
				row = append(row, group.counts[metricIndex])
			}
		}
		out = append(out, row)
	}

	return out, nil
}

func CompareSummaryRows(expected [][]any, actual [][]string) error {
	if len(expected) != len(actual) {
		return fmt.Errorf("summary row count=%d want %d rows=%v", len(actual), len(expected), actual)
	}
	for rowIndex := range expected {
		if len(expected[rowIndex]) != len(actual[rowIndex]) {
			return fmt.Errorf("summary row %d column count=%d want %d row=%v", rowIndex, len(actual[rowIndex]), len(expected[rowIndex]), actual[rowIndex])
		}
		for colIndex := range expected[rowIndex] {
			want := stringifyCell(expected[rowIndex][colIndex])
			if actual[rowIndex][colIndex] != want {
				return fmt.Errorf("summary cell[%d][%d]=%q want %q rows=%v", rowIndex, colIndex, actual[rowIndex][colIndex], want, actual)
			}
		}
	}
	return nil
}

type summaryGroup struct {
	groupValues []string
	sums        []float64
	counts      []int
}

func rowMatchesFilters(row []string, headerIndex map[string]int, filters []FilterSpec) (bool, error) {
	for _, filter := range filters {
		columnIndex, ok := headerIndex[filter.Column]
		if !ok {
			return false, fmt.Errorf("missing filter column %q", filter.Column)
		}

		value := valueAt(row, columnIndex)
		switch filter.Op {
		case "!=", "<>":
			if value == filter.Value {
				return false, nil
			}
		case "=", "==":
			if value != filter.Value {
				return false, nil
			}
		default:
			return false, fmt.Errorf("unsupported filter op %q", filter.Op)
		}
	}
	return true, nil
}

func parseNumber(raw string) (float64, error) {
	cleaned := strings.ReplaceAll(strings.TrimSpace(raw), ",", "")
	return strconv.ParseFloat(cleaned, 64)
}

func excelNumberValue(value float64) any {
	if value == float64(int64(value)) {
		return int64(value)
	}
	return value
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

func valueAt(row []string, index int) string {
	if index < 0 || index >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[index])
}

func validateHighlightOperator(operator string) error {
	switch operator {
	case ">", ">=", "<", "<=", "=":
		return nil
	default:
		return fmt.Errorf("unsupported highlight operator %q", operator)
	}
}

func cloneHeaderIndex(values map[string]int) map[string]int {
	if len(values) == 0 {
		return map[string]int{}
	}
	cloned := make(map[string]int, len(values))
	for key, value := range values {
		cloned[key] = value
	}
	return cloned
}

func cloneFilters(values []FilterSpec) []FilterSpec {
	if len(values) == 0 {
		return nil
	}
	cloned := make([]FilterSpec, len(values))
	copy(cloned, values)
	return cloned
}

func cloneMetrics(values []MetricSpec) []MetricSpec {
	if len(values) == 0 {
		return nil
	}
	cloned := make([]MetricSpec, len(values))
	copy(cloned, values)
	return cloned
}

func cloneStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	cloned := make([]string, len(values))
	copy(cloned, values)
	return cloned
}

func cloneInts(values []int) []int {
	if len(values) == 0 {
		return nil
	}
	cloned := make([]int, len(values))
	copy(cloned, values)
	return cloned
}

func cloneDataValidationRule(value taskspec.DataValidationRule) *DataValidationRule {
	return &DataValidationRule{
		Ranges:        cloneStrings(value.Ranges),
		RuleType:      value.RuleType,
		AllowedValues: cloneStrings(value.AllowedValues),
		AllowBlank:    value.AllowBlank,
	}
}

func cloneFormulaProtectionRule(value taskspec.FormulaProtectionRule) *FormulaProtectionRule {
	return &FormulaProtectionRule{
		FormulaRanges: cloneStrings(value.FormulaRanges),
		InputRanges:   cloneStrings(value.InputRanges),
		Password:      value.Password,
	}
}
