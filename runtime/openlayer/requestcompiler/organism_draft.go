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
	HeaderRow            int                    `json:"header_row,omitempty"`
	HeaderMappings       []HeaderMapping        `json:"header_mappings,omitempty"`
	CarryForwardMappings []CarryForwardMapping  `json:"carry_forward_mappings,omitempty"`
	LeftKey              string                 `json:"left_key,omitempty"`
	RightKey             string                 `json:"right_key,omitempty"`
	CompareMappings      []CompareMapping       `json:"compare_mappings,omitempty"`
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
	case "expense_reimbursement":
		return draftExpenseReimbursement(input)
	case "purchase_order_control":
		return draftPurchaseOrderControl(input)
	case "monthly_budget_control":
		return draftMonthlyBudgetControl(input)
	case "cash_flow_monitor":
		return draftCashFlowMonitor(input)
	case "attendance_register":
		return draftAttendanceRegister(input)
	case "project_timeline_tracker":
		return draftProjectTimelineTracker(input)
	case "shift_roster_planner":
		return draftShiftRosterPlanner(input)
	case "timesheet_hours_log":
		return draftTimesheetHoursLog(input)
	case "warehouse_reorder_tracker":
		return draftWarehouseReorderTracker(input)
	case "inventory_movement_log":
		return draftInventoryMovementLog(input)
	case "student_gradebook":
		return draftStudentGradebook(input)
	case "loan_repayment_calculator":
		return draftLoanRepaymentCalculator(input)
	default:
		return OrganismExecutionRequestDraft{}, false
	}
}

func draftShiftRosterPlanner(input OrganismDraftInput) (OrganismExecutionRequestDraft, bool) {
	sheet, ok := findShiftRosterSheet(input.WorkbookFacts)
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
	shiftColumnLetter, ok := columnLetterForHeader(sheet.Columns, "shift")
	if !ok {
		return OrganismExecutionRequestDraft{}, false
	}
	nextSheet := "Week2"

	return OrganismExecutionRequestDraft{
		ScenarioID:  input.ScenarioID,
		RequestText: input.RequestText,
		InputFile:   input.InputFile,
		OutputFile:  input.OutputFile,
		Steps: []OrganismExecutionStepDraft{
			{
				AtomID:          "copy_period_sheet",
				CompositionKind: "period_copy",
				SourceSheet:     sheet.Name,
				TargetSheet:     nextSheet,
			},
			{
				AtomID:          "add_data_validation",
				CompositionKind: "data_validation",
				SourceSheet:     nextSheet,
				ValidationRule: &DataValidationRule{
					Ranges:        []string{shiftColumnLetter + "2:" + shiftColumnLetter + "20"},
					RuleType:      "list",
					AllowedValues: []string{"AM", "PM", "OFF"},
					AllowBlank:    false,
				},
			},
			{
				AtomID:          "protect_formula_cells",
				CompositionKind: "formula_protection",
				SourceSheet:     nextSheet,
				ProtectionRule: &FormulaProtectionRule{
					FormulaRanges: []string{formulaRange(formulaColumnLetters, minFormulaRow, maxFormulaRow)},
					InputRanges:   []string{"A2:C20"},
				},
			},
		},
	}, true
}

func draftProjectTimelineTracker(input OrganismDraftInput) (OrganismExecutionRequestDraft, bool) {
	sheet, ok := findProjectTimelineSheet(input.WorkbookFacts)
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
	targetRow := maxFormulaRow + 1
	if targetRow > sheet.RowCount {
		return OrganismExecutionRequestDraft{}, false
	}
	statusColumnLetter, ok := columnLetterForHeader(sheet.Columns, "status")
	if !ok {
		return OrganismExecutionRequestDraft{}, false
	}
	nextSheet := "Sprint2"

	return OrganismExecutionRequestDraft{
		ScenarioID:  input.ScenarioID,
		RequestText: input.RequestText,
		InputFile:   input.InputFile,
		OutputFile:  input.OutputFile,
		Steps: []OrganismExecutionStepDraft{
			{
				AtomID:          "copy_period_sheet",
				CompositionKind: "period_copy",
				SourceSheet:     sheet.Name,
				TargetSheet:     nextSheet,
			},
			{
				AtomID:           "extend_table_formulas",
				CompositionKind:  "formula_extension",
				SourceSheet:      nextSheet,
				FormulaSourceRow: minFormulaRow,
				TargetRows:       []int{targetRow},
				FormulaColumns:   formulaColumnLetters,
			},
			{
				AtomID:          "add_data_validation",
				CompositionKind: "data_validation",
				SourceSheet:     nextSheet,
				ValidationRule: &DataValidationRule{
					Ranges:        []string{statusColumnLetter + "2:" + statusColumnLetter + "20"},
					RuleType:      "list",
					AllowedValues: []string{"todo", "in_progress", "done"},
					AllowBlank:    false,
				},
			},
			{
				AtomID:          "group_summarize",
				CompositionKind: "group_summary",
				SourceSheet:     nextSheet,
				TargetSheet:     "TimelineSummary",
				SummaryMode:     "values",
				GroupBy:         []string{"status"},
				Metrics:         []MetricSpec{{Column: "task_count", Op: "sum", As: "task_total"}},
			},
			{
				AtomID:          "protect_formula_cells",
				CompositionKind: "formula_protection",
				SourceSheet:     nextSheet,
				ProtectionRule: &FormulaProtectionRule{
					FormulaRanges: []string{formulaRange(formulaColumnLetters, minFormulaRow, targetRow)},
					InputRanges:   []string{"A2:D20"},
				},
			},
		},
	}, true
}

func draftAttendanceRegister(input OrganismDraftInput) (OrganismExecutionRequestDraft, bool) {
	sheet, ok := findAttendanceRegisterSheet(input.WorkbookFacts)
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
	statusColumnLetter, ok := columnLetterForHeader(sheet.Columns, "status")
	if !ok {
		return OrganismExecutionRequestDraft{}, false
	}
	nextSheet := "NextAttendance"

	return OrganismExecutionRequestDraft{
		ScenarioID:  input.ScenarioID,
		RequestText: input.RequestText,
		InputFile:   input.InputFile,
		OutputFile:  input.OutputFile,
		Steps: []OrganismExecutionStepDraft{
			{
				AtomID:          "copy_period_sheet",
				CompositionKind: "period_copy",
				SourceSheet:     sheet.Name,
				TargetSheet:     nextSheet,
			},
			{
				AtomID:          "add_data_validation",
				CompositionKind: "data_validation",
				SourceSheet:     nextSheet,
				ValidationRule: &DataValidationRule{
					Ranges:        []string{statusColumnLetter + "2:" + statusColumnLetter + "20"},
					RuleType:      "list",
					AllowedValues: []string{"present", "absent", "excused"},
					AllowBlank:    false,
				},
			},
			{
				AtomID:          "protect_formula_cells",
				CompositionKind: "formula_protection",
				SourceSheet:     nextSheet,
				ProtectionRule: &FormulaProtectionRule{
					FormulaRanges: []string{formulaRange(formulaColumnLetters, minFormulaRow, maxFormulaRow)},
					InputRanges:   []string{"A2:C20"},
				},
			},
		},
	}, true
}

func draftPurchaseOrderControl(input OrganismDraftInput) (OrganismExecutionRequestDraft, bool) {
	sheet, ok := findPurchaseOrderSheet(input.WorkbookFacts)
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
	statusColumnLetter, ok := columnLetterForHeader(sheet.Columns, "po_status")
	if !ok {
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
				Values:               defaultPurchaseOrderValues(inputColumns),
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
					Ranges:        []string{statusColumnLetter + "2:" + statusColumnLetter + "20"},
					RuleType:      "list",
					AllowedValues: []string{"draft", "submitted", "approved"},
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
				TargetSheet:     "PurchaseOrderPrint",
				FormTitle:       "Purchase Order",
				PrintArea:       "A1:" + lastColumnLetter(sheet.Columns) + strconv.Itoa(targetRow+5),
				FieldBindings: []FormFieldBinding{
					{Label: "First Item", SourceSheet: sheet.Name, SourceCell: "A2", LabelCell: "A2", ValueCell: "B2"},
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

func draftExpenseReimbursement(input OrganismDraftInput) (OrganismExecutionRequestDraft, bool) {
	sheet, ok := findExpenseReimbursementSheet(input.WorkbookFacts)
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
	receiptColumnLetter, ok := columnLetterForHeader(sheet.Columns, "receipt_status")
	if !ok {
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
				Values:               defaultExpenseValues(inputColumns),
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
					Ranges:        []string{receiptColumnLetter + "2:" + receiptColumnLetter + "20"},
					RuleType:      "list",
					AllowedValues: []string{"attached", "missing", "not_required"},
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
				TargetSheet:     "ExpenseClaim",
				FormTitle:       "Expense Reimbursement",
				PrintArea:       "A1:" + lastColumnLetter(sheet.Columns) + strconv.Itoa(targetRow+5),
				FieldBindings: []FormFieldBinding{
					{Label: "First Item", SourceSheet: sheet.Name, SourceCell: "A2", LabelCell: "A2", ValueCell: "B2"},
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

func draftWarehouseReorderTracker(input OrganismDraftInput) (OrganismExecutionRequestDraft, bool) {
	stockSheet, ok := findWarehouseReorderSheet(input.WorkbookFacts)
	if !ok {
		return OrganismExecutionRequestDraft{}, false
	}
	if _, ok := findSheetWithHeaders(input.WorkbookFacts, "sku", "location"); !ok {
		return OrganismExecutionRequestDraft{}, false
	}
	formulaColumns := formulaColumnsForSheet(stockSheet)
	if len(formulaColumns) == 0 {
		return OrganismExecutionRequestDraft{}, false
	}
	formulaColumnLetters := make([]string, 0, len(formulaColumns))
	for _, column := range formulaColumns {
		if letter, ok := columnLetterForHeader(stockSheet.Columns, column); ok {
			formulaColumnLetters = append(formulaColumnLetters, letter)
		}
	}
	if len(formulaColumnLetters) == 0 {
		return OrganismExecutionRequestDraft{}, false
	}
	minFormulaRow, maxFormulaRow, ok := formulaRowBounds(stockSheet)
	if !ok {
		return OrganismExecutionRequestDraft{}, false
	}
	threshold := 5.0
	targetSheet := "StockEnriched"
	sourceColumns := []string{"sku", "quantity", "reorder_level", "reorder_gap"}

	return OrganismExecutionRequestDraft{
		ScenarioID:  input.ScenarioID,
		RequestText: input.RequestText,
		InputFile:   input.InputFile,
		OutputFile:  input.OutputFile,
		Steps: []OrganismExecutionStepDraft{
			{
				AtomID:               "join_lookup",
				CompositionKind:      "join_lookup",
				SourceSheet:          stockSheet.Name,
				LookupSheet:          "SKU",
				TargetSheet:          targetSheet,
				JoinKey:              "sku",
				IncludeSourceColumns: append([]string(nil), sourceColumns...),
				AppendLookupColumns:  []string{"location"},
			},
			{
				AtomID:          "highlight_threshold",
				CompositionKind: "threshold_highlight",
				SourceSheet:     targetSheet,
				TargetColumn:    "quantity",
				Operator:        "<",
				Threshold:       &threshold,
				HighlightColor:  "#FFF59D",
			},
			{
				AtomID:               "append_structured_rows",
				CompositionKind:      "structured_row_append",
				SourceSheet:          targetSheet,
				IncludeSourceColumns: []string{"sku", "quantity", "reorder_level", "reorder_gap", "location"},
				Values: []CellValue{
					{Cell: "sku", Value: "B002"},
					{Cell: "quantity", Value: 2},
					{Cell: "reorder_level", Value: 5},
					{Cell: "reorder_gap", Value: -3},
					{Cell: "location", Value: "Aisle 2"},
				},
			},
			{
				AtomID:          "add_data_validation",
				CompositionKind: "data_validation",
				SourceSheet:     targetSheet,
				ValidationRule: &DataValidationRule{
					Ranges:        []string{"A2:A20"},
					RuleType:      "list",
					AllowedValues: []string{"A001", "B002"},
					AllowBlank:    false,
				},
			},
			{
				AtomID:          "protect_formula_cells",
				CompositionKind: "formula_protection",
				SourceSheet:     stockSheet.Name,
				ProtectionRule: &FormulaProtectionRule{
					FormulaRanges: []string{formulaRange(formulaColumnLetters, minFormulaRow, maxFormulaRow)},
					InputRanges:   []string{"A2:C20"},
				},
			},
		},
	}, true
}

func draftTimesheetHoursLog(input OrganismDraftInput) (OrganismExecutionRequestDraft, bool) {
	sheet, ok := findTimesheetHoursSheet(input.WorkbookFacts)
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
	sourceRow, targetRow, ok := formulaRows(sheet)
	if !ok {
		return OrganismExecutionRequestDraft{}, false
	}
	nextSheet := "Week2"
	inputColumns, _ := splitInputAndFormulaColumns(sheet)

	return OrganismExecutionRequestDraft{
		ScenarioID:  input.ScenarioID,
		RequestText: input.RequestText,
		InputFile:   input.InputFile,
		OutputFile:  input.OutputFile,
		Steps: []OrganismExecutionStepDraft{
			{
				AtomID:          "copy_period_sheet",
				CompositionKind: "period_copy",
				SourceSheet:     sheet.Name,
				TargetSheet:     nextSheet,
			},
			{
				AtomID:               "append_structured_rows",
				CompositionKind:      "structured_row_append",
				SourceSheet:          nextSheet,
				IncludeSourceColumns: append([]string(nil), inputColumns...),
				Values:               defaultTimesheetValues(inputColumns),
			},
			{
				AtomID:           "extend_table_formulas",
				CompositionKind:  "formula_extension",
				SourceSheet:      nextSheet,
				FormulaSourceRow: sourceRow,
				TargetRows:       []int{targetRow},
				FormulaColumns:   formulaColumnLetters,
			},
			{
				AtomID:          "add_data_validation",
				CompositionKind: "data_validation",
				SourceSheet:     nextSheet,
				ValidationRule: &DataValidationRule{
					Ranges:        []string{"C2:C20"},
					RuleType:      "list",
					AllowedValues: []string{"DEV", "OPS", "ADMIN"},
					AllowBlank:    false,
				},
			},
			{
				AtomID:          "protect_formula_cells",
				CompositionKind: "formula_protection",
				SourceSheet:     nextSheet,
				ProtectionRule: &FormulaProtectionRule{
					FormulaRanges: []string{formulaRange(formulaColumnLetters, sourceRow, targetRow)},
					InputRanges:   []string{"A2:E20"},
				},
			},
		},
	}, true
}

func draftCashFlowMonitor(input OrganismDraftInput) (OrganismExecutionRequestDraft, bool) {
	sheet, ok := findCashFlowSheet(input.WorkbookFacts)
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
	targetRow := maxFormulaRow + 1
	if targetRow > sheet.RowCount {
		return OrganismExecutionRequestDraft{}, false
	}
	closingColumnLetter, ok := columnLetterForHeader(sheet.Columns, "closing")
	if !ok {
		return OrganismExecutionRequestDraft{}, false
	}
	nextSheet := "NextCashFlow"

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
				TargetSheet:     "CashFlowSummary",
				SummaryMode:     "values",
				GroupBy:         []string{"period"},
				Metrics:         []MetricSpec{{Column: "inflow", Op: "sum", As: "inflow_total"}},
			},
			{
				AtomID:           "extend_table_formulas",
				CompositionKind:  "formula_extension",
				SourceSheet:      sheet.Name,
				FormulaSourceRow: minFormulaRow,
				TargetRows:       []int{targetRow},
				FormulaColumns:   formulaColumnLetters,
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
					FormulaRanges: []string{formulaRange(formulaColumnLetters, minFormulaRow, targetRow)},
					InputRanges:   []string{"B2:D20"},
				},
			},
		},
	}, true
}

func draftLoanRepaymentCalculator(input OrganismDraftInput) (OrganismExecutionRequestDraft, bool) {
	sheet, ok := findLoanRepaymentSheet(input.WorkbookFacts)
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
	targetRow := maxFormulaRow + 1
	if targetRow > sheet.RowCount {
		return OrganismExecutionRequestDraft{}, false
	}

	return OrganismExecutionRequestDraft{
		ScenarioID:  input.ScenarioID,
		RequestText: input.RequestText,
		InputFile:   input.InputFile,
		OutputFile:  input.OutputFile,
		Steps: []OrganismExecutionStepDraft{
			{
				AtomID:           "extend_table_formulas",
				CompositionKind:  "formula_extension",
				SourceSheet:      sheet.Name,
				FormulaSourceRow: minFormulaRow,
				TargetRows:       []int{targetRow},
				FormulaColumns:   formulaColumnLetters,
			},
			{
				AtomID:          "protect_formula_cells",
				CompositionKind: "formula_protection",
				SourceSheet:     sheet.Name,
				ProtectionRule: &FormulaProtectionRule{
					FormulaRanges: []string{formulaRange(formulaColumnLetters, minFormulaRow, targetRow)},
					InputRanges:   []string{"B2:B20"},
				},
			},
		},
	}, true
}

func draftStudentGradebook(input OrganismDraftInput) (OrganismExecutionRequestDraft, bool) {
	sheet, ok := findStudentGradebookSheet(input.WorkbookFacts)
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
	targetRow := maxFormulaRow + 1
	if targetRow > sheet.RowCount {
		return OrganismExecutionRequestDraft{}, false
	}
	statusColumnLetter, ok := columnLetterForHeader(sheet.Columns, "status")
	if !ok {
		return OrganismExecutionRequestDraft{}, false
	}

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
				TargetSheet:     "GradeSummary",
				SummaryMode:     "values",
				GroupBy:         []string{"student"},
				Metrics:         []MetricSpec{{Column: "score", Op: "sum", As: "score_total"}},
			},
			{
				AtomID:          "add_data_validation",
				CompositionKind: "data_validation",
				SourceSheet:     sheet.Name,
				ValidationRule: &DataValidationRule{
					Ranges:        []string{statusColumnLetter + "2:" + statusColumnLetter + "20"},
					RuleType:      "list",
					AllowedValues: []string{"complete", "missing", "excused"},
					AllowBlank:    false,
				},
			},
			{
				AtomID:           "extend_table_formulas",
				CompositionKind:  "formula_extension",
				SourceSheet:      sheet.Name,
				FormulaSourceRow: minFormulaRow,
				TargetRows:       []int{targetRow},
				FormulaColumns:   formulaColumnLetters,
			},
			{
				AtomID:          "protect_formula_cells",
				CompositionKind: "formula_protection",
				SourceSheet:     sheet.Name,
				ProtectionRule: &FormulaProtectionRule{
					FormulaRanges: []string{formulaRange(formulaColumnLetters, minFormulaRow, targetRow)},
					InputRanges:   []string{"C2:D20"},
				},
			},
		},
	}, true
}

func draftInventoryMovementLog(input OrganismDraftInput) (OrganismExecutionRequestDraft, bool) {
	movementSheet, ok := findInventoryMovementSheet(input.WorkbookFacts)
	if !ok {
		return OrganismExecutionRequestDraft{}, false
	}
	if _, ok := findSheetWithHeaders(input.WorkbookFacts, "sku", "location"); !ok {
		return OrganismExecutionRequestDraft{}, false
	}
	if _, ok := findSheetWithHeaders(input.WorkbookFacts, "sku", "on_hand"); !ok {
		return OrganismExecutionRequestDraft{}, false
	}
	formulaColumns := formulaColumnsForSheet(movementSheet)
	if len(formulaColumns) == 0 {
		return OrganismExecutionRequestDraft{}, false
	}
	formulaColumnLetters := make([]string, 0, len(formulaColumns))
	for _, column := range formulaColumns {
		if letter, ok := columnLetterForHeader(movementSheet.Columns, column); ok {
			formulaColumnLetters = append(formulaColumnLetters, letter)
		}
	}
	if len(formulaColumnLetters) == 0 {
		return OrganismExecutionRequestDraft{}, false
	}
	minFormulaRow, maxFormulaRow, ok := formulaRowBounds(movementSheet)
	if !ok {
		return OrganismExecutionRequestDraft{}, false
	}
	normalizedColumns := []string{"sku", "quantity_in", "quantity_out", "balance"}

	return OrganismExecutionRequestDraft{
		ScenarioID:  input.ScenarioID,
		RequestText: input.RequestText,
		InputFile:   input.InputFile,
		OutputFile:  input.OutputFile,
		Steps: []OrganismExecutionStepDraft{
			{
				AtomID:          "normalize_headers",
				CompositionKind: "header_normalization",
				SourceSheet:     movementSheet.Name,
				HeaderRow:       1,
				HeaderMappings: []HeaderMapping{
					{From: "SKU ID", To: "sku"},
					{From: "Qty In", To: "quantity_in"},
					{From: "Qty Out", To: "quantity_out"},
					{From: "Balance", To: "balance"},
				},
			},
			{
				AtomID:               "append_structured_rows",
				CompositionKind:      "structured_row_append",
				SourceSheet:          movementSheet.Name,
				IncludeSourceColumns: append([]string(nil), normalizedColumns...),
				Values: []CellValue{
					{Cell: "sku", Value: "C003"},
					{Cell: "quantity_in", Value: 2},
					{Cell: "quantity_out", Value: 0},
					{Cell: "balance", Value: 2},
				},
			},
			{
				AtomID:               "join_lookup",
				CompositionKind:      "join_lookup",
				SourceSheet:          movementSheet.Name,
				LookupSheet:          "SKU",
				TargetSheet:          "MovementsEnriched",
				JoinKey:              "sku",
				IncludeSourceColumns: append([]string(nil), normalizedColumns...),
				AppendLookupColumns:  []string{"location"},
			},
			{
				AtomID:          "protect_formula_cells",
				CompositionKind: "formula_protection",
				SourceSheet:     movementSheet.Name,
				ProtectionRule: &FormulaProtectionRule{
					FormulaRanges: []string{formulaRange(formulaColumnLetters, minFormulaRow, maxFormulaRow)},
					InputRanges:   []string{"A2:C20"},
				},
			},
			{
				AtomID:          "reconcile_tables",
				CompositionKind: "table_reconciliation",
				SourceSheet:     "MovementsEnriched",
				LookupSheet:     "StockMaster",
				TargetSheet:     "InventoryReconciliation",
				LeftKey:         "sku",
				RightKey:        "sku",
				CompareMappings: []CompareMapping{
					{LeftColumn: "balance", RightColumn: "on_hand", As: "balance"},
				},
			},
		},
	}, true
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

func findExpenseReimbursementSheet(facts runtimeinspect.WorkbookFacts) (runtimeinspect.SheetFacts, bool) {
	for _, sheet := range facts.Sheets {
		headers := lowerSet(sheet.Columns)
		if headers["item"] && headers["amount"] && headers["reimbursable_rate"] && headers["receipt_status"] && headers["reimbursable_total"] {
			return sheet, true
		}
	}
	return runtimeinspect.SheetFacts{}, false
}

func findPurchaseOrderSheet(facts runtimeinspect.WorkbookFacts) (runtimeinspect.SheetFacts, bool) {
	for _, sheet := range facts.Sheets {
		headers := lowerSet(sheet.Columns)
		if headers["item"] && headers["quantity"] && headers["unit_price"] && headers["po_status"] && headers["line_total"] {
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

func findCashFlowSheet(facts runtimeinspect.WorkbookFacts) (runtimeinspect.SheetFacts, bool) {
	for _, sheet := range facts.Sheets {
		headers := lowerSet(sheet.Columns)
		if headers["period"] && headers["opening"] && headers["inflow"] && headers["outflow"] && headers["closing"] {
			return sheet, true
		}
	}
	return runtimeinspect.SheetFacts{}, false
}

func findTimesheetHoursSheet(facts runtimeinspect.WorkbookFacts) (runtimeinspect.SheetFacts, bool) {
	for _, sheet := range facts.Sheets {
		headers := lowerSet(sheet.Columns)
		if headers["date"] && headers["employee"] && headers["work_code"] && headers["hours"] && headers["rate"] && headers["pay"] {
			return sheet, true
		}
	}
	return runtimeinspect.SheetFacts{}, false
}

func findAttendanceRegisterSheet(facts runtimeinspect.WorkbookFacts) (runtimeinspect.SheetFacts, bool) {
	for _, sheet := range facts.Sheets {
		headers := lowerSet(sheet.Columns)
		if headers["student"] && headers["date"] && headers["status"] && headers["attendance_total"] {
			return sheet, true
		}
	}
	return runtimeinspect.SheetFacts{}, false
}

func findProjectTimelineSheet(facts runtimeinspect.WorkbookFacts) (runtimeinspect.SheetFacts, bool) {
	for _, sheet := range facts.Sheets {
		headers := lowerSet(sheet.Columns)
		if headers["task"] && headers["start"] && headers["end"] && headers["status"] && headers["task_count"] && headers["progress_pct"] {
			return sheet, true
		}
	}
	return runtimeinspect.SheetFacts{}, false
}

func findShiftRosterSheet(facts runtimeinspect.WorkbookFacts) (runtimeinspect.SheetFacts, bool) {
	for _, sheet := range facts.Sheets {
		headers := lowerSet(sheet.Columns)
		if headers["employee"] && headers["date"] && headers["shift"] && headers["coverage_total"] {
			return sheet, true
		}
	}
	return runtimeinspect.SheetFacts{}, false
}

func findWarehouseReorderSheet(facts runtimeinspect.WorkbookFacts) (runtimeinspect.SheetFacts, bool) {
	for _, sheet := range facts.Sheets {
		headers := lowerSet(sheet.Columns)
		if headers["sku"] && headers["quantity"] && headers["reorder_level"] && headers["reorder_gap"] {
			return sheet, true
		}
	}
	return runtimeinspect.SheetFacts{}, false
}

func findInventoryMovementSheet(facts runtimeinspect.WorkbookFacts) (runtimeinspect.SheetFacts, bool) {
	for _, sheet := range facts.Sheets {
		headers := lowerSet(sheet.Columns)
		if headers["sku id"] && headers["qty in"] && headers["qty out"] && headers["balance"] {
			return sheet, true
		}
	}
	return runtimeinspect.SheetFacts{}, false
}

func findStudentGradebookSheet(facts runtimeinspect.WorkbookFacts) (runtimeinspect.SheetFacts, bool) {
	for _, sheet := range facts.Sheets {
		headers := lowerSet(sheet.Columns)
		if headers["student"] && headers["assignment"] && headers["score"] && headers["status"] && headers["weighted_score"] {
			return sheet, true
		}
	}
	return runtimeinspect.SheetFacts{}, false
}

func findLoanRepaymentSheet(facts runtimeinspect.WorkbookFacts) (runtimeinspect.SheetFacts, bool) {
	for _, sheet := range facts.Sheets {
		headers := lowerSet(sheet.Columns)
		if headers["period"] && headers["payment"] && headers["interest"] && headers["principal"] && headers["balance"] {
			return sheet, true
		}
	}
	return runtimeinspect.SheetFacts{}, false
}

func findSheetWithHeaders(facts runtimeinspect.WorkbookFacts, requiredHeaders ...string) (runtimeinspect.SheetFacts, bool) {
	for _, sheet := range facts.Sheets {
		headers := lowerSet(sheet.Columns)
		missing := false
		for _, required := range requiredHeaders {
			if !headers[strings.ToLower(required)] {
				missing = true
				break
			}
		}
		if !missing {
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

func defaultExpenseValues(columns []string) []CellValue {
	values := make([]CellValue, 0, len(columns))
	for _, column := range columns {
		switch strings.ToLower(column) {
		case "item":
			values = append(values, CellValue{Cell: column, Value: "Taxi"})
		case "amount":
			values = append(values, CellValue{Cell: column, Value: 35})
		case "reimbursable_rate":
			values = append(values, CellValue{Cell: column, Value: 1})
		case "receipt_status":
			values = append(values, CellValue{Cell: column, Value: "attached"})
		}
	}
	return values
}

func defaultPurchaseOrderValues(columns []string) []CellValue {
	values := make([]CellValue, 0, len(columns))
	for _, column := range columns {
		switch strings.ToLower(column) {
		case "item":
			values = append(values, CellValue{Cell: column, Value: "Monitor"})
		case "quantity":
			values = append(values, CellValue{Cell: column, Value: 1})
		case "unit_price":
			values = append(values, CellValue{Cell: column, Value: 150})
		case "po_status":
			values = append(values, CellValue{Cell: column, Value: "submitted"})
		}
	}
	return values
}

func defaultTimesheetValues(columns []string) []CellValue {
	values := make([]CellValue, 0, len(columns))
	for _, column := range columns {
		switch strings.ToLower(column) {
		case "date":
			values = append(values, CellValue{Cell: column, Value: "2026-06-08"})
		case "employee":
			values = append(values, CellValue{Cell: column, Value: "Ben"})
		case "work_code":
			values = append(values, CellValue{Cell: column, Value: "DEV"})
		case "hours":
			values = append(values, CellValue{Cell: column, Value: 6})
		case "rate":
			values = append(values, CellValue{Cell: column, Value: 25})
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
