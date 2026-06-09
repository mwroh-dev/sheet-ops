package validate

import (
	"path/filepath"
	goruntime "runtime"

	runtimeinspect "github.com/mwroh/sheet-ops/runtime/inspect"
	runtimeschema "github.com/mwroh/sheet-ops/runtime/schema"
)

type Status string

const (
	StatusCompiled             Status = "compiled"
	StatusNeedsHumanCheckpoint Status = "needs_human_checkpoint"
	StatusBlocked              Status = "blocked"
)

type FailedStage string

const (
	StageIntentAdmission    FailedStage = "intent_admission"
	StageFactBinding        FailedStage = "fact_binding"
	StageExecutionAdmission FailedStage = "execution_admission"
)

type AggregateIntent struct {
	Kind   string
	Column string
}

type SummaryIntent struct {
	TargetSheet string
	SummaryMode string
	Filters     []FilterSpec
	GroupBy     []string
	Metrics     []AggregateIntent
}

type HighlightIntent struct {
	TargetColumn   string
	Operator       string
	Threshold      *float64
	HighlightColor string
}

type JoinLookupIntent struct {
	TargetSheet          string
	JoinKey              string
	IncludeSourceColumns []string
	AppendLookupColumns  []string
}

type AppendRowsIntent struct {
	IncludeSourceColumns []string
	Values               []CellValue
}

type ExtendFormulasIntent struct {
	FormulaSourceRow int
	TargetRows       []int
	FormulaColumns   []string
}

type PeriodCopyIntent struct {
	TargetSheet string
}

type AddDataValidationIntent struct {
	ValidationRule DataValidationRule
}

type ProtectFormulaCellsIntent struct {
	ProtectionRule FormulaProtectionRule
}

type NormalizeHeadersIntent struct {
	HeaderRow      int
	HeaderMappings []HeaderMapping
}

type RollForwardPeriodIntent struct {
	TargetSheet          string
	CarryForwardMappings []CarryForwardMapping
}

type ReconcileTablesIntent struct {
	TargetSheet     string
	LeftKey         string
	RightKey        string
	CompareMappings []CompareMapping
}

type GeneratePrintableFormIntent struct {
	TargetSheet   string
	FormTitle     string
	PrintArea     string
	FieldBindings []FormFieldBinding
	TableBinding  *FormTableBinding
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

type CarryForwardMapping struct {
	FromSheet string `json:"from_sheet,omitempty"`
	FromCell  string `json:"from_cell"`
	ToSheet   string `json:"to_sheet,omitempty"`
	ToCell    string `json:"to_cell"`
}

type CompareMapping struct {
	LeftColumn  string `json:"left_column"`
	RightColumn string `json:"right_column"`
	As          string `json:"as,omitempty"`
}

type FormFieldBinding struct {
	Label       string `json:"label"`
	SourceSheet string `json:"source_sheet,omitempty"`
	SourceCell  string `json:"source_cell"`
	LabelCell   string `json:"label_cell"`
	ValueCell   string `json:"value_cell"`
}

type FormTableBinding struct {
	SourceSheet   string   `json:"source_sheet"`
	SourceColumns []string `json:"source_columns"`
	HeaderStart   string   `json:"header_start"`
	DataStart     string   `json:"data_start"`
}

type MaterializationIntent struct {
	PreserveOriginal      bool
	OutputDestinationMode string
	WriteShape            string
}

type AmbiguityIntent struct {
	Markers          []string
	UnresolvedFields []string
	CheckpointHints  []string
}

type NormalizedIntent struct {
	SourceSheetCandidates []string
	LookupSheetCandidates []string
	GroupKeys             []string
	Aggregates            []AggregateIntent
	CompositionCandidates []string
	Summary               SummaryIntent
	Highlight             HighlightIntent
	JoinLookup            JoinLookupIntent
	AppendRows            AppendRowsIntent
	ExtendFormulas        ExtendFormulasIntent
	PeriodCopy            PeriodCopyIntent
	AddDataValidation     AddDataValidationIntent
	ProtectFormulaCells   ProtectFormulaCellsIntent
	NormalizeHeaders      NormalizeHeadersIntent
	RollForwardPeriod     RollForwardPeriodIntent
	ReconcileTables       ReconcileTablesIntent
	GeneratePrintableForm GeneratePrintableFormIntent
	Materialization       MaterializationIntent
	Ambiguity             AmbiguityIntent
}

type CompilerDecision struct {
	Status            Status
	SelectedOperation string
	Notes             []string
	SheetCandidates   []string
	StructuralSignals []string
	Checkpoint        *Checkpoint
}

type RequestContext struct {
	ScenarioID  string
	RequestKind string
	RequestText string
	OutputFile  string
}

type Result struct {
	Status                    Status                     `json:"status"`
	ValidatedExecutionRequest *ValidatedExecutionRequest `json:"validated_execution_request,omitempty"`
	Checkpoint                *Checkpoint                `json:"checkpoint,omitempty"`
	Blocked                   *Blocked                   `json:"blocked,omitempty"`
}

type Checkpoint struct {
	Kind             string   `json:"kind"`
	Question         string   `json:"question"`
	Options          []string `json:"options"`
	UnresolvedFields []string `json:"unresolved_fields"`
}

type Blocked struct {
	FailedStage FailedStage `json:"failed_stage"`
	ReasonCodes []string    `json:"reason_codes"`
}

type ValidatedExecutionRequest struct {
	ScenarioID           string                 `json:"scenario_id"`
	RequestKind          string                 `json:"request_kind"`
	RequestText          string                 `json:"request_text,omitempty"`
	InputFile            string                 `json:"input_file"`
	SourceSheet          string                 `json:"source_sheet"`
	OutputFile           string                 `json:"output_file"`
	ExecutionKind        string                 `json:"execution_kind"`
	CompositionKind      string                 `json:"composition_kind,omitempty"`
	PrimitiveKind        string                 `json:"primitive_kind,omitempty"`
	TargetSheet          string                 `json:"target_sheet,omitempty"`
	SummaryMode          string                 `json:"summary_mode,omitempty"`
	Filters              []FilterSpec           `json:"filters,omitempty"`
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

type CellValue struct {
	Cell  string `json:"cell"`
	Value any    `json:"value"`
}

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

type Validator struct {
	RequestContext RequestContext
}

type admittedIntent struct {
	decision             CompilerDecision
	compositionKind      string
	sourceSheets         []string
	lookupSheets         []string
	targetSheet          string
	summaryMode          string
	filters              []FilterSpec
	groupBy              []string
	metrics              []MetricSpec
	targetColumn         string
	operator             string
	threshold            *float64
	highlightColor       string
	joinKey              string
	includeSourceColumns []string
	appendLookupColumns  []string
	values               []CellValue
	formulaSourceRow     int
	targetRows           []int
	formulaColumns       []string
	validationRule       DataValidationRule
	protectionRule       FormulaProtectionRule
	headerRow            int
	headerMappings       []HeaderMapping
	carryForwardMappings []CarryForwardMapping
	leftKey              string
	rightKey             string
	compareMappings      []CompareMapping
	formTitle            string
	printArea            string
	fieldBindings        []FormFieldBinding
	tableBinding         *FormTableBinding
}

type boundIntent struct {
	admittedIntent
	facts                runtimeinspect.WorkbookFacts
	sourceSheet          string
	lookupSheet          string
	includeSourceColumns []string
}

func cloneCellValues(values []CellValue) []CellValue {
	if len(values) == 0 {
		return []CellValue{}
	}
	cloned := make([]CellValue, len(values))
	copy(cloned, values)
	return cloned
}

func cloneInts(values []int) []int {
	if len(values) == 0 {
		return []int{}
	}
	cloned := make([]int, len(values))
	copy(cloned, values)
	return cloned
}

func cloneDataValidationRule(value DataValidationRule) DataValidationRule {
	return DataValidationRule{
		Ranges:        append([]string(nil), value.Ranges...),
		RuleType:      value.RuleType,
		AllowedValues: append([]string(nil), value.AllowedValues...),
		AllowBlank:    value.AllowBlank,
	}
}

func cloneFormulaProtectionRule(value FormulaProtectionRule) FormulaProtectionRule {
	return FormulaProtectionRule{
		FormulaRanges: append([]string(nil), value.FormulaRanges...),
		InputRanges:   append([]string(nil), value.InputRanges...),
		Password:      value.Password,
	}
}

func cloneHeaderMappings(values []HeaderMapping) []HeaderMapping {
	if len(values) == 0 {
		return []HeaderMapping{}
	}
	cloned := make([]HeaderMapping, len(values))
	copy(cloned, values)
	return cloned
}

func cloneCarryForwardMappings(values []CarryForwardMapping) []CarryForwardMapping {
	if len(values) == 0 {
		return []CarryForwardMapping{}
	}
	cloned := make([]CarryForwardMapping, len(values))
	copy(cloned, values)
	return cloned
}

func cloneCompareMappings(values []CompareMapping) []CompareMapping {
	if len(values) == 0 {
		return []CompareMapping{}
	}
	cloned := make([]CompareMapping, len(values))
	copy(cloned, values)
	return cloned
}

func cloneFormFieldBindings(values []FormFieldBinding) []FormFieldBinding {
	if len(values) == 0 {
		return []FormFieldBinding{}
	}
	cloned := make([]FormFieldBinding, len(values))
	copy(cloned, values)
	return cloned
}

func cloneFormTableBinding(value *FormTableBinding) *FormTableBinding {
	if value == nil {
		return nil
	}
	cloned := *value
	cloned.SourceColumns = append([]string(nil), value.SourceColumns...)
	return &cloned
}

func finalizeResult(result Result) (Result, error) {
	if result.ValidatedExecutionRequest != nil {
		if err := runtimeschema.ValidateStruct(validatedExecutionRequestSchemaPath(), result.ValidatedExecutionRequest); err != nil {
			return Result{}, err
		}
	}
	if err := runtimeschema.ValidateStruct(validationResultSchemaPath(), result); err != nil {
		return Result{}, err
	}
	return result, nil
}

func validationResultSchemaPath() string {
	_, file, _, ok := goruntime.Caller(0)
	if !ok {
		return filepath.Join("contracts", "validation", "validation_result.schema.json")
	}
	return runtimeschema.ResolveRepoPath(file, 2, "contracts", "validation", "validation_result.schema.json")
}

func validatedExecutionRequestSchemaPath() string {
	_, file, _, ok := goruntime.Caller(0)
	if !ok {
		return filepath.Join("contracts", "requests", "validated_execution_request.schema.json")
	}
	return runtimeschema.ResolveRepoPath(file, 2, "contracts", "requests", "validated_execution_request.schema.json")
}
