package requestcompiler

import (
	"sort"
	"strconv"
	"strings"

	runtimeinspect "github.com/mwroh/sheet-ops/runtime/inspect"
	"github.com/xuri/excelize/v2"
)

type OrganismExecutionRequestDraft struct {
	ScenarioID  string                       `json:"scenario_id"`
	RequestText string                       `json:"request_text"`
	InputFile   string                       `json:"input_file"`
	OutputFile  string                       `json:"output_file"`
	Steps       []OrganismExecutionStepDraft `json:"steps"`
}

type OrganismExecutionStepDraft struct {
	AtomID               string                 `json:"atom_id"`
	CompositionKind      string                 `json:"composition_kind"`
	SourceSheet          string                 `json:"source_sheet"`
	TargetSheet          string                 `json:"target_sheet,omitempty"`
	SummaryMode          string                 `json:"summary_mode,omitempty"`
	GroupBy              []string               `json:"group_by,omitempty"`
	Metrics              []MetricSpec           `json:"metrics,omitempty"`
	TargetColumn         string                 `json:"target_column,omitempty"`
	Operator             string                 `json:"operator,omitempty"`
	Threshold            *float64               `json:"threshold,omitempty"`
	HighlightColor       string                 `json:"highlight_color,omitempty"`
	IncludeSourceColumns []string               `json:"include_source_columns,omitempty"`
	Values               []CellValue            `json:"values,omitempty"`
	FormulaSourceRow     int                    `json:"formula_source_row,omitempty"`
	TargetRows           []int                  `json:"target_rows,omitempty"`
	FormulaColumns       []string               `json:"formula_columns,omitempty"`
	ValidationRule       *DataValidationRule    `json:"validation_rule,omitempty"`
	ProtectionRule       *FormulaProtectionRule `json:"protection_rule,omitempty"`
	CarryForwardMappings []CarryForwardMapping  `json:"carry_forward_mappings,omitempty"`
	FormTitle            string                 `json:"form_title,omitempty"`
	PrintArea            string                 `json:"print_area,omitempty"`
	FieldBindings        []FormFieldBinding     `json:"field_bindings,omitempty"`
	TableBinding         *FormTableBinding      `json:"table_binding,omitempty"`
}

type MetricSpec struct {
	Column string `json:"column"`
	Op     string `json:"op"`
	As     string `json:"as"`
}

type OrganismDraftInput struct {
	ScenarioID        string
	RequestText       string
	InputFile         string
	OutputFile        string
	WorkbookFacts     runtimeinspect.WorkbookFacts
	TemplateClassPlan TemplateClassPlanHint
}

func DraftOrganismExecutionRequest(input OrganismDraftInput) (OrganismExecutionRequestDraft, bool) {
	switch input.TemplateClassPlan.OrganismID {
	case "invoice_line_item_billing":
		return draftInvoiceLineItemBilling(input)
	case "monthly_budget_control":
		return draftMonthlyBudgetControl(input)
	default:
		return OrganismExecutionRequestDraft{}, false
	}
}

func draftMonthlyBudgetControl(input OrganismDraftInput) (OrganismExecutionRequestDraft, bool) {
	sheet, ok := findMonthlyBudgetSheet(input.WorkbookFacts)
	if !ok {
		return OrganismExecutionRequestDraft{}, false
	}
	formulaColumns := formulaColumnsForSheet(sheet)
	if len(formulaColumns) == 0 {
		return OrganismExecutionRequestDraft{}, false
	}
	formulaColumnLetters := make([]string, 0, len(formulaColumns))
	for _, column := range formulaColumns {
		if letter, ok := columnLetterForHeader(sheet.Columns, column); ok {
			formulaColumnLetters = append(formulaColumnLetters, letter)
		}
	}
	if len(formulaColumnLetters) == 0 {
		return OrganismExecutionRequestDraft{}, false
	}
	minFormulaRow, maxFormulaRow, ok := formulaRowBounds(sheet)
	if !ok {
		return OrganismExecutionRequestDraft{}, false
	}
	closingColumnLetter, ok := columnLetterForHeader(sheet.Columns, "closing")
	if !ok {
		return OrganismExecutionRequestDraft{}, false
	}
	threshold := 0.0
	nextSheet := "NextBudget"
	formulaProtectionRange := formulaRange(formulaColumnLetters, minFormulaRow, maxFormulaRow)

	return OrganismExecutionRequestDraft{
		ScenarioID:  input.ScenarioID,
		RequestText: input.RequestText,
		InputFile:   input.InputFile,
		OutputFile:  input.OutputFile,
		Steps: []OrganismExecutionStepDraft{
			{
				AtomID:          "group_summarize",
				CompositionKind: "group_summary",
				SourceSheet:     sheet.Name,
				TargetSheet:     "BudgetSummary",
				SummaryMode:     "values",
				GroupBy:         []string{"category"},
				Metrics:         []MetricSpec{{Column: "actual", Op: "sum", As: "actual_total"}},
			},
			{
				AtomID:          "highlight_threshold",
				CompositionKind: "threshold_highlight",
				SourceSheet:     sheet.Name,
				TargetColumn:    "variance",
				Operator:        ">",
				Threshold:       &threshold,
				HighlightColor:  "#FFF59D",
			},
			{
				AtomID:          "copy_period_sheet",
				CompositionKind: "period_copy",
				SourceSheet:     sheet.Name,
				TargetSheet:     nextSheet,
			},
			{
				AtomID:          "roll_forward_period",
				CompositionKind: "period_roll_forward",
				SourceSheet:     sheet.Name,
				TargetSheet:     nextSheet,
				CarryForwardMappings: []CarryForwardMapping{
					{FromSheet: sheet.Name, FromCell: closingColumnLetter + strconv.Itoa(minFormulaRow), ToSheet: nextSheet, ToCell: "B" + strconv.Itoa(minFormulaRow+1)},
				},
			},
			{
				AtomID:          "protect_formula_cells",
				CompositionKind: "formula_protection",
				SourceSheet:     nextSheet,
				ProtectionRule: &FormulaProtectionRule{
					FormulaRanges: []string{formulaProtectionRange},
					InputRanges:   []string{"B2:D10"},
				},
			},
		},
	}, true
}

func draftInvoiceLineItemBilling(input OrganismDraftInput) (OrganismExecutionRequestDraft, bool) {
	sheet, ok := findInvoiceLineItemSheet(input.WorkbookFacts)
	if !ok {
		return OrganismExecutionRequestDraft{}, false
	}
	inputColumns, formulaColumns := splitInputAndFormulaColumns(sheet)
	if len(inputColumns) == 0 || len(formulaColumns) == 0 {
		return OrganismExecutionRequestDraft{}, false
	}
	formulaSourceRow, targetRow, ok := formulaRows(sheet)
	if !ok {
		return OrganismExecutionRequestDraft{}, false
	}
	formulaColumnLetters := make([]string, 0, len(formulaColumns))
	for _, column := range formulaColumns {
		if letter, ok := columnLetterForHeader(sheet.Columns, column); ok {
			formulaColumnLetters = append(formulaColumnLetters, letter)
		}
	}
	if len(formulaColumnLetters) == 0 {
		return OrganismExecutionRequestDraft{}, false
	}

	return OrganismExecutionRequestDraft{
		ScenarioID:  input.ScenarioID,
		RequestText: input.RequestText,
		InputFile:   input.InputFile,
		OutputFile:  input.OutputFile,
		Steps: []OrganismExecutionStepDraft{
			{
				AtomID:               "append_structured_rows",
				CompositionKind:      "structured_row_append",
				SourceSheet:          sheet.Name,
				IncludeSourceColumns: append([]string(nil), inputColumns...),
				Values:               defaultInvoiceValues(inputColumns),
			},
			{
				AtomID:           "extend_table_formulas",
				CompositionKind:  "formula_extension",
				SourceSheet:      sheet.Name,
				FormulaSourceRow: formulaSourceRow,
				TargetRows:       []int{targetRow},
				FormulaColumns:   formulaColumnLetters,
			},
			{
				AtomID:          "add_data_validation",
				CompositionKind: "data_validation",
				SourceSheet:     sheet.Name,
				ValidationRule: &DataValidationRule{
					Ranges:        []string{"A2:A10"},
					RuleType:      "list",
					AllowedValues: []string{"A001", "B002"},
					AllowBlank:    false,
				},
			},
			{
				AtomID:          "protect_formula_cells",
				CompositionKind: "formula_protection",
				SourceSheet:     sheet.Name,
				ProtectionRule: &FormulaProtectionRule{
					FormulaRanges: []string{formulaRange(formulaColumnLetters, formulaSourceRow, targetRow)},
					InputRanges:   []string{inputRange(inputColumns, targetRow)},
				},
			},
			{
				AtomID:          "generate_printable_form",
				CompositionKind: "printable_form",
				SourceSheet:     sheet.Name,
				TargetSheet:     "InvoicePrint",
				FormTitle:       "Invoice",
				PrintArea:       "A1:" + lastColumnLetter(sheet.Columns) + strconv.Itoa(targetRow+5),
				FieldBindings: []FormFieldBinding{
					{Label: "First SKU", SourceSheet: sheet.Name, SourceCell: "A2", LabelCell: "A2", ValueCell: "B2"},
				},
				TableBinding: &FormTableBinding{
					SourceSheet:   sheet.Name,
					SourceColumns: append([]string(nil), sheet.Columns...),
					HeaderStart:   "A4",
					DataStart:     "A5",
				},
			},
		},
	}, true
}

func findInvoiceLineItemSheet(facts runtimeinspect.WorkbookFacts) (runtimeinspect.SheetFacts, bool) {
	for _, sheet := range facts.Sheets {
		headers := lowerSet(sheet.Columns)
		if headers["sku"] && headers["quantity"] && headers["unit_price"] {
			return sheet, true
		}
	}
	return runtimeinspect.SheetFacts{}, false
}

func findMonthlyBudgetSheet(facts runtimeinspect.WorkbookFacts) (runtimeinspect.SheetFacts, bool) {
	for _, sheet := range facts.Sheets {
		headers := lowerSet(sheet.Columns)
		if headers["category"] && headers["actual"] && headers["budget"] && headers["variance"] && headers["closing"] {
			return sheet, true
		}
	}
	return runtimeinspect.SheetFacts{}, false
}

func splitInputAndFormulaColumns(sheet runtimeinspect.SheetFacts) ([]string, []string) {
	formulaHeaders := map[string]bool{}
	for _, formula := range sheet.Formulas {
		if header, ok := headerForCell(sheet.Columns, formula.Cell); ok {
			formulaHeaders[header] = true
		}
	}
	inputColumns := []string{}
	formulaColumns := []string{}
	for _, column := range sheet.Columns {
		if formulaHeaders[column] {
			formulaColumns = append(formulaColumns, column)
			continue
		}
		inputColumns = append(inputColumns, column)
	}
	return inputColumns, formulaColumns
}

func formulaColumnsForSheet(sheet runtimeinspect.SheetFacts) []string {
	seen := map[string]bool{}
	columns := []string{}
	for _, formula := range sheet.Formulas {
		header, ok := headerForCell(sheet.Columns, formula.Cell)
		if !ok || seen[header] {
			continue
		}
		seen[header] = true
		columns = append(columns, header)
	}
	return columns
}

func formulaRows(sheet runtimeinspect.SheetFacts) (int, int, bool) {
	rows := map[int]bool{}
	for _, formula := range sheet.Formulas {
		_, row, err := excelize.CellNameToCoordinates(formula.Cell)
		if err != nil {
			return 0, 0, false
		}
		rows[row] = true
	}
	if len(rows) == 0 {
		return 0, 0, false
	}
	ordered := make([]int, 0, len(rows))
	for row := range rows {
		ordered = append(ordered, row)
	}
	sort.Ints(ordered)
	sourceRow := ordered[0]
	targetRow := sheet.RowCount + 1
	if targetRow <= sourceRow {
		targetRow = sourceRow + 1
	}
	return sourceRow, targetRow, true
}

func formulaRowBounds(sheet runtimeinspect.SheetFacts) (int, int, bool) {
	rows := make([]int, 0, len(sheet.Formulas))
	seen := map[int]bool{}
	for _, formula := range sheet.Formulas {
		_, row, err := excelize.CellNameToCoordinates(formula.Cell)
		if err != nil {
			return 0, 0, false
		}
		if seen[row] {
			continue
		}
		seen[row] = true
		rows = append(rows, row)
	}
	if len(rows) == 0 {
		return 0, 0, false
	}
	sort.Ints(rows)
	return rows[0], rows[len(rows)-1], true
}

func defaultInvoiceValues(columns []string) []CellValue {
	values := make([]CellValue, 0, len(columns))
	for _, column := range columns {
		switch strings.ToLower(column) {
		case "sku":
			values = append(values, CellValue{Cell: column, Value: "B002"})
		case "quantity":
			values = append(values, CellValue{Cell: column, Value: 3})
		case "unit_price":
			values = append(values, CellValue{Cell: column, Value: 15})
		}
	}
	return values
}

func lowerSet(values []string) map[string]bool {
	out := map[string]bool{}
	for _, value := range values {
		out[strings.ToLower(strings.TrimSpace(value))] = true
	}
	return out
}

func headerForCell(headers []string, cell string) (string, bool) {
	column, _, err := excelize.CellNameToCoordinates(cell)
	if err != nil || column < 1 || column > len(headers) {
		return "", false
	}
	return headers[column-1], true
}

func columnLetterForHeader(headers []string, header string) (string, bool) {
	for index, candidate := range headers {
		if candidate != header {
			continue
		}
		letter, err := excelize.ColumnNumberToName(index + 1)
		if err != nil {
			return "", false
		}
		return letter, true
	}
	return "", false
}

func formulaRange(columns []string, sourceRow, targetRow int) string {
	if len(columns) == 0 {
		return ""
	}
	return columns[0] + strconv.Itoa(sourceRow) + ":" + columns[len(columns)-1] + strconv.Itoa(targetRow)
}

func inputRange(inputColumns []string, targetRow int) string {
	if len(inputColumns) == 0 {
		return ""
	}
	letter, err := excelize.ColumnNumberToName(len(inputColumns))
	if err != nil {
		letter = "A"
	}
	return "A2:" + letter + strconv.Itoa(targetRow+7)
}

func lastColumnLetter(headers []string) string {
	letter, err := excelize.ColumnNumberToName(len(headers))
	if err != nil || letter == "" {
		return "A"
	}
	return letter
}
