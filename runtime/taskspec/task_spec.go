package taskspec

type FilterSpec struct {
	Column string `json:"column"`
	Op     string `json:"op"`
	Value  string `json:"value"`
}

type MetricSpec struct {
	Column string `json:"column"`
	Op     string `json:"op"`
	As     string `json:"as"`
}

type ThresholdRule struct {
	Column         string  `json:"column"`
	Operator       string  `json:"operator"`
	Threshold      float64 `json:"threshold"`
	HighlightColor string  `json:"highlight_color"`
}

type JoinLookupSpec struct {
	LookupSheet          string   `json:"lookup_sheet"`
	JoinKey              string   `json:"join_key"`
	IncludeSourceColumns []string `json:"include_source_columns"`
	AppendLookupColumns  []string `json:"append_lookup_columns"`
}

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

type HeaderMapping struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type SourceSpec struct {
	Kind string `json:"kind"`
	Path string `json:"path,omitempty"`
}

type TaskSpec struct {
	RequestKind          string                 `json:"request_kind"`
	ExecutionKind        string                 `json:"execution_kind"`
	CompositionKind      string                 `json:"composition_kind,omitempty"`
	Source               SourceSpec             `json:"source"`
	RequestText          string                 `json:"request_text,omitempty"`
	InputWorkbook        string                 `json:"input_workbook"`
	OutputWorkbook       string                 `json:"output_workbook"`
	Operation            string                 `json:"operation"`
	SourceSheet          string                 `json:"source_sheet"`
	TargetSheet          string                 `json:"target_sheet,omitempty"`
	Filters              []FilterSpec           `json:"filters,omitempty"`
	GroupBy              []string               `json:"group_by,omitempty"`
	Metrics              []MetricSpec           `json:"metrics,omitempty"`
	SummaryMode          string                 `json:"summary_mode,omitempty"`
	ThresholdRule        *ThresholdRule         `json:"threshold_rule,omitempty"`
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
}

type GroupSummarizeTask struct {
	TaskSpec         TaskSpec
	SourceSheet      string
	TargetSheet      string
	Filters          []FilterSpec
	GroupBy          []string
	Metrics          []MetricSpec
	PreserveOriginal bool
}

type HighlightThresholdTask struct {
	TaskSpec         TaskSpec
	SourceSheet      string
	ThresholdRule    ThresholdRule
	PreserveOriginal bool
}

type JoinLookupTask struct {
	TaskSpec             TaskSpec
	SourceSheet          string
	LookupSheet          string
	JoinKey              string
	TargetSheet          string
	IncludeSourceColumns []string
	AppendLookupColumns  []string
	PreserveOriginal     bool
}

type AppendStructuredRowsTask struct {
	TaskSpec             TaskSpec
	SourceSheet          string
	IncludeSourceColumns []string
	Values               []CellValue
	PreserveOriginal     bool
}

type ExtendTableFormulasTask struct {
	TaskSpec         TaskSpec
	SourceSheet      string
	FormulaSourceRow int
	TargetRows       []int
	FormulaColumns   []string
	PreserveOriginal bool
}

type CopyPeriodSheetTask struct {
	TaskSpec         TaskSpec
	SourceSheet      string
	TargetSheet      string
	PreserveOriginal bool
}

type AddDataValidationTask struct {
	TaskSpec         TaskSpec
	SourceSheet      string
	ValidationRule   DataValidationRule
	PreserveOriginal bool
}

type ProtectFormulaCellsTask struct {
	TaskSpec         TaskSpec
	SourceSheet      string
	ProtectionRule   FormulaProtectionRule
	PreserveOriginal bool
}

type NormalizeHeadersTask struct {
	TaskSpec         TaskSpec
	SourceSheet      string
	HeaderRow        int
	HeaderMappings   []HeaderMapping
	PreserveOriginal bool
}

type UseRequest struct {
	RequestText string
	InputFile   string
	SourceSheet string
	OutputFile  string
	TargetSheet string
	SummaryMode string
	Filters     []FilterSpec
	GroupBy     []string
	Metrics     []MetricSpec
}

type HighlightThresholdRequest struct {
	RequestText    string
	InputFile      string
	SourceSheet    string
	OutputFile     string
	Column         string
	Operator       string
	Threshold      *float64
	HighlightColor string
}

type JoinLookupRequest struct {
	RequestText          string
	InputFile            string
	SourceSheet          string
	LookupSheet          string
	OutputFile           string
	TargetSheet          string
	JoinKey              string
	IncludeSourceColumns []string
	AppendLookupColumns  []string
}

type AppendStructuredRowsRequest struct {
	RequestText          string
	InputFile            string
	SourceSheet          string
	OutputFile           string
	IncludeSourceColumns []string
	Values               []CellValue
}

type ExtendTableFormulasRequest struct {
	RequestText      string
	InputFile        string
	SourceSheet      string
	OutputFile       string
	FormulaSourceRow int
	TargetRows       []int
	FormulaColumns   []string
}

type CopyPeriodSheetRequest struct {
	RequestText string
	InputFile   string
	SourceSheet string
	TargetSheet string
	OutputFile  string
}

type AddDataValidationRequest struct {
	RequestText    string
	InputFile      string
	SourceSheet    string
	OutputFile     string
	ValidationRule DataValidationRule
}

type ProtectFormulaCellsRequest struct {
	RequestText    string
	InputFile      string
	SourceSheet    string
	OutputFile     string
	ProtectionRule FormulaProtectionRule
}

type NormalizeHeadersRequest struct {
	RequestText    string
	InputFile      string
	SourceSheet    string
	OutputFile     string
	HeaderRow      int
	HeaderMappings []HeaderMapping
}

const (
	ExecutionKindComposition             = "composition"
	CompositionKindGroupSummary          = "group_summary"
	CompositionKindThresholdHighlight    = "threshold_highlight"
	CompositionKindJoinLookup            = "join_lookup"
	CompositionKindStructuredRowAppend   = "structured_row_append"
	CompositionKindFormulaExtension      = "formula_extension"
	CompositionKindPeriodCopy            = "period_copy"
	CompositionKindDataValidation        = "data_validation"
	CompositionKindFormulaProtection     = "formula_protection"
	CompositionKindHeaderNormalization   = "header_normalization"
	OperationCreateSummarySheet          = "create_summary_sheet"
	OperationHighlightThresholdRows      = "highlight_threshold_rows"
	OperationCreateJoinLookupResultSheet = "create_join_lookup_result_sheet"
	OperationAppendStructuredRows        = "append_structured_rows"
	OperationExtendTableFormulas         = "extend_table_formulas"
	OperationCopyPeriodSheet             = "copy_period_sheet"
	OperationAddDataValidation           = "add_data_validation"
	OperationProtectFormulaCells         = "protect_formula_cells"
	OperationNormalizeHeaders            = "normalize_headers"
	OperationFamilyGroupSummarize        = "group_summarize"
	OperationFamilyHighlightThreshold    = "highlight_threshold"
	OperationFamilyJoinLookup            = "join_lookup"
	OperationFamilyAppendStructuredRows  = "append_structured_rows"
	OperationFamilyExtendTableFormulas   = "extend_table_formulas"
	OperationFamilyCopyPeriodSheet       = "copy_period_sheet"
	OperationFamilyAddDataValidation     = "add_data_validation"
	OperationFamilyProtectFormulaCells   = "protect_formula_cells"
	OperationFamilyNormalizeHeaders      = "normalize_headers"

	defaultHighlightColumn    = "amount"
	defaultHighlightOperator  = ">"
	defaultHighlightThreshold = 100000.0
	defaultHighlightColor     = "#FFF59D"
)

func FromUseRequest(req UseRequest) TaskSpec {
	return BuildGroupSummarizeTask(req).TaskSpec
}

func BuildGroupSummarizeTask(req UseRequest) GroupSummarizeTask {
	summaryMode := req.SummaryMode
	if summaryMode == "" {
		summaryMode = "values"
	}

	return GroupSummarizeTask{
		TaskSpec: TaskSpec{
			RequestKind:     "workbook_case",
			ExecutionKind:   ExecutionKindComposition,
			CompositionKind: CompositionKindGroupSummary,
			Source:          SourceSpec{Kind: "natural_language"},
			RequestText:     req.RequestText,
			InputWorkbook:   req.InputFile,
			OutputWorkbook:  req.OutputFile,
			Operation:       OperationCreateSummarySheet,
			SourceSheet:     req.SourceSheet,
			TargetSheet:     req.TargetSheet,
			Filters:         cloneFilters(req.Filters),
			GroupBy:         cloneStrings(req.GroupBy),
			Metrics:         cloneMetrics(req.Metrics),
			SummaryMode:     summaryMode,
		},
		SourceSheet:      req.SourceSheet,
		TargetSheet:      req.TargetSheet,
		Filters:          cloneFilters(req.Filters),
		GroupBy:          cloneStrings(req.GroupBy),
		Metrics:          cloneMetrics(req.Metrics),
		PreserveOriginal: true,
	}
}

func BuildHighlightThresholdTask(req HighlightThresholdRequest) HighlightThresholdTask {
	column := req.Column
	if column == "" {
		column = defaultHighlightColumn
	}

	operator := req.Operator
	if operator == "" {
		operator = defaultHighlightOperator
	}

	threshold := defaultHighlightThreshold
	if req.Threshold != nil {
		threshold = *req.Threshold
	}

	highlightColor := req.HighlightColor
	if highlightColor == "" {
		highlightColor = defaultHighlightColor
	}

	thresholdRule := ThresholdRule{
		Column:         column,
		Operator:       operator,
		Threshold:      threshold,
		HighlightColor: highlightColor,
	}

	return HighlightThresholdTask{
		TaskSpec: TaskSpec{
			RequestKind:     "workbook_case",
			ExecutionKind:   ExecutionKindComposition,
			CompositionKind: CompositionKindThresholdHighlight,
			Source:          SourceSpec{Kind: "natural_language"},
			RequestText:     req.RequestText,
			InputWorkbook:   req.InputFile,
			OutputWorkbook:  req.OutputFile,
			Operation:       OperationHighlightThresholdRows,
			SourceSheet:     req.SourceSheet,
			ThresholdRule:   cloneThresholdRule(thresholdRule),
		},
		SourceSheet:      req.SourceSheet,
		ThresholdRule:    thresholdRule,
		PreserveOriginal: true,
	}
}

func BuildJoinLookupTask(req JoinLookupRequest) JoinLookupTask {
	return JoinLookupTask{
		TaskSpec: TaskSpec{
			RequestKind:          "workbook_case",
			ExecutionKind:        ExecutionKindComposition,
			CompositionKind:      CompositionKindJoinLookup,
			Source:               SourceSpec{Kind: "natural_language"},
			RequestText:          req.RequestText,
			InputWorkbook:        req.InputFile,
			OutputWorkbook:       req.OutputFile,
			Operation:            OperationCreateJoinLookupResultSheet,
			SourceSheet:          req.SourceSheet,
			TargetSheet:          req.TargetSheet,
			LookupSheet:          req.LookupSheet,
			JoinKey:              req.JoinKey,
			IncludeSourceColumns: cloneStrings(req.IncludeSourceColumns),
			AppendLookupColumns:  cloneStrings(req.AppendLookupColumns),
		},
		SourceSheet:          req.SourceSheet,
		LookupSheet:          req.LookupSheet,
		JoinKey:              req.JoinKey,
		TargetSheet:          req.TargetSheet,
		IncludeSourceColumns: cloneStrings(req.IncludeSourceColumns),
		AppendLookupColumns:  cloneStrings(req.AppendLookupColumns),
		PreserveOriginal:     true,
	}
}

func BuildAppendStructuredRowsTask(req AppendStructuredRowsRequest) AppendStructuredRowsTask {
	return AppendStructuredRowsTask{
		TaskSpec: TaskSpec{
			RequestKind:          "workbook_case",
			ExecutionKind:        ExecutionKindComposition,
			CompositionKind:      CompositionKindStructuredRowAppend,
			Source:               SourceSpec{Kind: "natural_language"},
			RequestText:          req.RequestText,
			InputWorkbook:        req.InputFile,
			OutputWorkbook:       req.OutputFile,
			Operation:            OperationAppendStructuredRows,
			SourceSheet:          req.SourceSheet,
			IncludeSourceColumns: cloneStrings(req.IncludeSourceColumns),
			Values:               cloneCellValues(req.Values),
		},
		SourceSheet:          req.SourceSheet,
		IncludeSourceColumns: cloneStrings(req.IncludeSourceColumns),
		Values:               cloneCellValues(req.Values),
		PreserveOriginal:     true,
	}
}

func BuildExtendTableFormulasTask(req ExtendTableFormulasRequest) ExtendTableFormulasTask {
	return ExtendTableFormulasTask{
		TaskSpec: TaskSpec{
			RequestKind:      "workbook_case",
			ExecutionKind:    ExecutionKindComposition,
			CompositionKind:  CompositionKindFormulaExtension,
			Source:           SourceSpec{Kind: "natural_language"},
			RequestText:      req.RequestText,
			InputWorkbook:    req.InputFile,
			OutputWorkbook:   req.OutputFile,
			Operation:        OperationExtendTableFormulas,
			SourceSheet:      req.SourceSheet,
			FormulaSourceRow: req.FormulaSourceRow,
			TargetRows:       cloneInts(req.TargetRows),
			FormulaColumns:   cloneStrings(req.FormulaColumns),
		},
		SourceSheet:      req.SourceSheet,
		FormulaSourceRow: req.FormulaSourceRow,
		TargetRows:       cloneInts(req.TargetRows),
		FormulaColumns:   cloneStrings(req.FormulaColumns),
		PreserveOriginal: true,
	}
}

func BuildCopyPeriodSheetTask(req CopyPeriodSheetRequest) CopyPeriodSheetTask {
	return CopyPeriodSheetTask{
		TaskSpec: TaskSpec{
			RequestKind:     "workbook_case",
			ExecutionKind:   ExecutionKindComposition,
			CompositionKind: CompositionKindPeriodCopy,
			Source:          SourceSpec{Kind: "natural_language"},
			RequestText:     req.RequestText,
			InputWorkbook:   req.InputFile,
			OutputWorkbook:  req.OutputFile,
			Operation:       OperationCopyPeriodSheet,
			SourceSheet:     req.SourceSheet,
			TargetSheet:     req.TargetSheet,
		},
		SourceSheet:      req.SourceSheet,
		TargetSheet:      req.TargetSheet,
		PreserveOriginal: true,
	}
}

func BuildAddDataValidationTask(req AddDataValidationRequest) AddDataValidationTask {
	return AddDataValidationTask{
		TaskSpec: TaskSpec{
			RequestKind:     "workbook_case",
			ExecutionKind:   ExecutionKindComposition,
			CompositionKind: CompositionKindDataValidation,
			Source:          SourceSpec{Kind: "natural_language"},
			RequestText:     req.RequestText,
			InputWorkbook:   req.InputFile,
			OutputWorkbook:  req.OutputFile,
			Operation:       OperationAddDataValidation,
			SourceSheet:     req.SourceSheet,
			ValidationRule:  cloneDataValidationRule(req.ValidationRule),
		},
		SourceSheet:      req.SourceSheet,
		ValidationRule:   *cloneDataValidationRule(req.ValidationRule),
		PreserveOriginal: true,
	}
}

func BuildProtectFormulaCellsTask(req ProtectFormulaCellsRequest) ProtectFormulaCellsTask {
	return ProtectFormulaCellsTask{
		TaskSpec: TaskSpec{
			RequestKind:     "workbook_case",
			ExecutionKind:   ExecutionKindComposition,
			CompositionKind: CompositionKindFormulaProtection,
			Source:          SourceSpec{Kind: "natural_language"},
			RequestText:     req.RequestText,
			InputWorkbook:   req.InputFile,
			OutputWorkbook:  req.OutputFile,
			Operation:       OperationProtectFormulaCells,
			SourceSheet:     req.SourceSheet,
			ProtectionRule:  cloneFormulaProtectionRule(req.ProtectionRule),
		},
		SourceSheet:      req.SourceSheet,
		ProtectionRule:   *cloneFormulaProtectionRule(req.ProtectionRule),
		PreserveOriginal: true,
	}
}

func BuildNormalizeHeadersTask(req NormalizeHeadersRequest) NormalizeHeadersTask {
	headerRow := req.HeaderRow
	if headerRow == 0 {
		headerRow = 1
	}
	return NormalizeHeadersTask{
		TaskSpec: TaskSpec{
			RequestKind:     "workbook_case",
			ExecutionKind:   ExecutionKindComposition,
			CompositionKind: CompositionKindHeaderNormalization,
			Source:          SourceSpec{Kind: "natural_language"},
			RequestText:     req.RequestText,
			InputWorkbook:   req.InputFile,
			OutputWorkbook:  req.OutputFile,
			Operation:       OperationNormalizeHeaders,
			SourceSheet:     req.SourceSheet,
			HeaderRow:       headerRow,
			HeaderMappings:  cloneHeaderMappings(req.HeaderMappings),
		},
		SourceSheet:      req.SourceSheet,
		HeaderRow:        headerRow,
		HeaderMappings:   cloneHeaderMappings(req.HeaderMappings),
		PreserveOriginal: true,
	}
}

func cloneFilters(values []FilterSpec) []FilterSpec {
	if len(values) == 0 {
		return nil
	}
	cloned := make([]FilterSpec, len(values))
	copy(cloned, values)
	return cloned
}

func cloneThresholdRule(value ThresholdRule) *ThresholdRule {
	cloned := value
	return &cloned
}

func cloneCellValues(values []CellValue) []CellValue {
	if len(values) == 0 {
		return nil
	}
	cloned := make([]CellValue, len(values))
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

func cloneDataValidationRule(value DataValidationRule) *DataValidationRule {
	cloned := value
	cloned.Ranges = cloneStrings(value.Ranges)
	cloned.AllowedValues = cloneStrings(value.AllowedValues)
	return &cloned
}

func cloneFormulaProtectionRule(value FormulaProtectionRule) *FormulaProtectionRule {
	cloned := value
	cloned.FormulaRanges = cloneStrings(value.FormulaRanges)
	cloned.InputRanges = cloneStrings(value.InputRanges)
	return &cloned
}

func cloneHeaderMappings(values []HeaderMapping) []HeaderMapping {
	if len(values) == 0 {
		return nil
	}
	cloned := make([]HeaderMapping, len(values))
	copy(cloned, values)
	return cloned
}
