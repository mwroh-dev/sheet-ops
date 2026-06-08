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

type AddDataValidationIntent struct {
	ValidationRule DataValidationRule
}

type DataValidationRule struct {
	Ranges        []string `json:"ranges"`
	RuleType      string   `json:"rule_type"`
	AllowedValues []string `json:"allowed_values,omitempty"`
	AllowBlank    bool     `json:"allow_blank"`
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
	AddDataValidation     AddDataValidationIntent
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
	ScenarioID           string              `json:"scenario_id"`
	RequestKind          string              `json:"request_kind"`
	RequestText          string              `json:"request_text,omitempty"`
	InputFile            string              `json:"input_file"`
	SourceSheet          string              `json:"source_sheet"`
	OutputFile           string              `json:"output_file"`
	ExecutionKind        string              `json:"execution_kind"`
	CompositionKind      string              `json:"composition_kind,omitempty"`
	PrimitiveKind        string              `json:"primitive_kind,omitempty"`
	TargetSheet          string              `json:"target_sheet,omitempty"`
	SummaryMode          string              `json:"summary_mode,omitempty"`
	Filters              []FilterSpec        `json:"filters,omitempty"`
	GroupBy              []string            `json:"group_by,omitempty"`
	Metrics              []MetricSpec        `json:"metrics,omitempty"`
	TargetColumn         string              `json:"target_column,omitempty"`
	Operator             string              `json:"operator,omitempty"`
	Threshold            *float64            `json:"threshold,omitempty"`
	HighlightColor       string              `json:"highlight_color,omitempty"`
	LookupSheet          string              `json:"lookup_sheet,omitempty"`
	JoinKey              string              `json:"join_key,omitempty"`
	IncludeSourceColumns []string            `json:"include_source_columns,omitempty"`
	AppendLookupColumns  []string            `json:"append_lookup_columns,omitempty"`
	Values               []CellValue         `json:"values,omitempty"`
	FormulaSourceRow     int                 `json:"formula_source_row,omitempty"`
	TargetRows           []int               `json:"target_rows,omitempty"`
	FormulaColumns       []string            `json:"formula_columns,omitempty"`
	ValidationRule       *DataValidationRule `json:"validation_rule,omitempty"`
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
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", "contracts", "validation", "validation_result.schema.json"))
}

func validatedExecutionRequestSchemaPath() string {
	_, file, _, ok := goruntime.Caller(0)
	if !ok {
		return filepath.Join("contracts", "requests", "validated_execution_request.schema.json")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", "contracts", "requests", "validated_execution_request.schema.json"))
}
