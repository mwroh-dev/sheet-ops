package useorchestrator

import (
	requestcompiler "github.com/mwroh/sheet-ops/runtime/openlayer/requestcompiler"
	runtimesubagent "github.com/mwroh/sheet-ops/runtime/subagent"
	runtimeworkbookcase "github.com/mwroh/sheet-ops/runtime/workbookcase"
)

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

type TelemetryEvent = runtimeworkbookcase.TelemetryEvent
type RunIDs = runtimeworkbookcase.RunIDs
type RunPaths = runtimeworkbookcase.RunPaths
type RunResult = runtimeworkbookcase.RunResult
type TemplateClassPlanHint = requestcompiler.TemplateClassPlanHint

type OrchestratorDecision struct {
	ScenarioID                string                            `json:"scenario_id"`
	Decision                  string                            `json:"decision"`
	RequestCompilerLoopState  runtimesubagent.SubagentLoopState `json:"request_compiler_loop_state"`
	ValidatedExecutionRequest *ValidatedExecutionRequest        `json:"validated_execution_request,omitempty"`
	TemplateClassPlan         *TemplateClassPlanHint            `json:"template_class_plan,omitempty"`
	RepairAdvice              *runtimeworkbookcase.RepairAdvice `json:"repair_advice,omitempty"`
}

type OrganismExecutionRequest struct {
	ScenarioID  string                  `json:"scenario_id"`
	RequestText string                  `json:"request_text"`
	InputFile   string                  `json:"input_file"`
	OutputFile  string                  `json:"output_file"`
	Steps       []OrganismExecutionStep `json:"steps"`
}

type OrganismExecutionStep struct {
	AtomID               string                 `json:"atom_id"`
	CompositionKind      string                 `json:"composition_kind"`
	SourceSheet          string                 `json:"source_sheet"`
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

type ResultVerifierOutcome struct {
	Review    runtimeworkbookcase.VerificationReview `json:"review"`
	LoopState runtimesubagent.SubagentLoopState      `json:"loop_state"`
}
