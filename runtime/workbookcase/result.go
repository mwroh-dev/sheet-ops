package workbookcase

import (
	"strings"

	runtimeepisode "github.com/mwroh/sheet-ops/runtime/episode"
	runtimetaskspec "github.com/mwroh/sheet-ops/runtime/taskspec"
)

const (
	SummaryOperationName               = "create_summary_sheet"
	HighlightOperationName             = "highlight_threshold_rows"
	JoinLookupOperationName            = "create_join_lookup_result_sheet"
	AppendRowsOperationName            = "append_structured_rows"
	ExtendFormulasOperationName        = "extend_table_formulas"
	CopyPeriodSheetOperationName       = "copy_period_sheet"
	AddDataValidationOperationName     = "add_data_validation"
	ProtectFormulaCellsOperationName   = "protect_formula_cells"
	NormalizeHeadersOperationName      = "normalize_headers"
	RollForwardPeriodOperationName     = "roll_forward_period"
	ReconcileTablesOperationName       = "reconcile_tables"
	GeneratePrintableFormOperationName = "generate_printable_form"
	WriteValuesOperationName           = "write_values"
)

type FilterSpec = runtimetaskspec.FilterSpec
type MetricSpec = runtimetaskspec.MetricSpec
type ExecutionEpisode = runtimeepisode.ExecutionEpisode
type RuntimeVersion = runtimeepisode.RuntimeVersion

type Request struct {
	ScenarioID string
	TaskSpec   runtimetaskspec.TaskSpec
	Routing    *RoutingContext
}

type RoutingContext struct {
	SelectedOperation      string
	RoutingAuthority       string
	KnowledgeRecordsLoaded int
	KnowledgeFallback      bool
	KnowledgePath          string
	RoutingNote            string
	FallbackReason         string
	RecentOperations       []string
}

func (context RoutingContext) Payload() map[string]any {
	payload := map[string]any{
		"selected_operation":       context.SelectedOperation,
		"routing_authority":        context.RoutingAuthority,
		"knowledge_records_loaded": context.KnowledgeRecordsLoaded,
		"knowledge_fallback":       context.KnowledgeFallback,
		"knowledge_path":           context.KnowledgePath,
		"routing_note":             context.RoutingNote,
	}
	if context.FallbackReason != "" {
		payload["knowledge_fallback_reason"] = context.FallbackReason
	}
	if len(context.RecentOperations) > 0 {
		payload["knowledge_recent_operations"] = append([]string(nil), context.RecentOperations...)
	}
	return payload
}

type InspectionArtifact struct {
	InputFile    string         `json:"input_file"`
	SheetNames   []string       `json:"sheet_names"`
	TargetSheet  string         `json:"target_sheet"`
	HeaderNames  []string       `json:"header_names"`
	HeaderIndex  map[string]int `json:"header_index,omitempty"`
	TargetColumn string         `json:"target_column,omitempty"`
	HeaderRow    int            `json:"header_row"`
	AmountColumn int            `json:"amount_column,omitempty"`
	RowCount     int            `json:"row_count"`
}

type SummaryInspectionRecord struct {
	InputFile          string         `json:"input_file"`
	SourceSheet        string         `json:"source_sheet"`
	PlannedTargetSheet string         `json:"planned_target_sheet,omitempty"`
	SheetNames         []string       `json:"sheet_names"`
	HeaderNames        []string       `json:"header_names"`
	HeaderIndex        map[string]int `json:"header_index,omitempty"`
	RowCount           int            `json:"row_count"`
}

type HighlightThresholdPlan struct {
	Operation        string  `json:"operation"`
	SheetName        string  `json:"sheet_name"`
	TargetColumn     string  `json:"target_column"`
	Operator         string  `json:"operator"`
	Threshold        float64 `json:"threshold"`
	HighlightColor   string  `json:"highlight_color"`
	PreserveOriginal bool    `json:"preserve_original"`
	OutputFile       string  `json:"output_file"`
}

type SummarySheetPlan struct {
	Operation        string       `json:"operation"`
	SheetName        string       `json:"sheet_name"`
	SummaryMode      string       `json:"summary_mode"`
	Filters          []FilterSpec `json:"filters"`
	GroupBy          []string     `json:"group_by"`
	Metrics          []MetricSpec `json:"metrics"`
	TargetSheet      string       `json:"target_sheet"`
	PreserveOriginal bool         `json:"preserve_original"`
	OutputFile       string       `json:"output_file"`
}

type JoinLookupPlan struct {
	Operation            string   `json:"operation"`
	SheetName            string   `json:"sheet_name"`
	LookupSheet          string   `json:"lookup_sheet"`
	JoinKey              string   `json:"join_key"`
	TargetSheet          string   `json:"target_sheet"`
	IncludeSourceColumns []string `json:"include_source_columns"`
	AppendLookupColumns  []string `json:"append_lookup_columns"`
	PreserveOriginal     bool     `json:"preserve_original"`
	OutputFile           string   `json:"output_file"`
}

type AppendStructuredRowsPlan struct {
	Operation            string                      `json:"operation"`
	SheetName            string                      `json:"sheet_name"`
	IncludeSourceColumns []string                    `json:"include_source_columns"`
	Values               []runtimetaskspec.CellValue `json:"values"`
	PreserveOriginal     bool                        `json:"preserve_original"`
	OutputFile           string                      `json:"output_file"`
}

type ExtendTableFormulasPlan struct {
	Operation        string   `json:"operation"`
	SheetName        string   `json:"sheet_name"`
	FormulaSourceRow int      `json:"formula_source_row"`
	TargetRows       []int    `json:"target_rows"`
	FormulaColumns   []string `json:"formula_columns"`
	PreserveOriginal bool     `json:"preserve_original"`
	OutputFile       string   `json:"output_file"`
}

type CopyPeriodSheetPlan struct {
	Operation        string `json:"operation"`
	SheetName        string `json:"sheet_name"`
	TargetSheet      string `json:"target_sheet"`
	PreserveOriginal bool   `json:"preserve_original"`
	OutputFile       string `json:"output_file"`
}

type AddDataValidationPlan struct {
	Operation        string                             `json:"operation"`
	SheetName        string                             `json:"sheet_name"`
	ValidationRule   runtimetaskspec.DataValidationRule `json:"validation_rule"`
	PreserveOriginal bool                               `json:"preserve_original"`
	OutputFile       string                             `json:"output_file"`
}

type ProtectFormulaCellsPlan struct {
	Operation        string                                `json:"operation"`
	SheetName        string                                `json:"sheet_name"`
	ProtectionRule   runtimetaskspec.FormulaProtectionRule `json:"protection_rule"`
	PreserveOriginal bool                                  `json:"preserve_original"`
	OutputFile       string                                `json:"output_file"`
}

type NormalizeHeadersPlan struct {
	Operation        string                          `json:"operation"`
	SheetName        string                          `json:"sheet_name"`
	HeaderRow        int                             `json:"header_row"`
	HeaderMappings   []runtimetaskspec.HeaderMapping `json:"header_mappings"`
	PreserveOriginal bool                            `json:"preserve_original"`
	OutputFile       string                          `json:"output_file"`
}

type RollForwardPeriodPlan struct {
	Operation            string                                `json:"operation"`
	SheetName            string                                `json:"sheet_name"`
	TargetSheet          string                                `json:"target_sheet"`
	CarryForwardMappings []runtimetaskspec.CarryForwardMapping `json:"carry_forward_mappings"`
	PreserveOriginal     bool                                  `json:"preserve_original"`
	OutputFile           string                                `json:"output_file"`
}

type ReconcileTablesPlan struct {
	Operation        string                           `json:"operation"`
	SheetName        string                           `json:"sheet_name"`
	LookupSheet      string                           `json:"lookup_sheet"`
	TargetSheet      string                           `json:"target_sheet"`
	LeftKey          string                           `json:"left_key"`
	RightKey         string                           `json:"right_key"`
	CompareMappings  []runtimetaskspec.CompareMapping `json:"compare_mappings"`
	PreserveOriginal bool                             `json:"preserve_original"`
	OutputFile       string                           `json:"output_file"`
}

type GeneratePrintableFormPlan struct {
	Operation        string                             `json:"operation"`
	SheetName        string                             `json:"sheet_name"`
	TargetSheet      string                             `json:"target_sheet"`
	FormTitle        string                             `json:"form_title"`
	PrintArea        string                             `json:"print_area"`
	FieldBindings    []runtimetaskspec.FormFieldBinding `json:"field_bindings"`
	TableBinding     *runtimetaskspec.FormTableBinding  `json:"table_binding"`
	PreserveOriginal bool                               `json:"preserve_original"`
	OutputFile       string                             `json:"output_file"`
}

type WritePolicyDecision struct {
	Allow            bool     `json:"allow"`
	PreserveOriginal bool     `json:"preserve_original"`
	ApprovalRequired bool     `json:"approval_required"`
	Reasons          []string `json:"reasons"`
}

type ExecutionSummary struct {
	Operation          string   `json:"operation"`
	OutputFile         string   `json:"output_file"`
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

type VerificationResult struct {
	Pass                 bool                `json:"pass"`
	Operation            string              `json:"operation"`
	OutputFile           string              `json:"output_file"`
	OutputWorkbookSHA256 string              `json:"output_workbook_sha256,omitempty"`
	SummarySheet         string              `json:"summary_sheet,omitempty"`
	SummaryMode          string              `json:"summary_mode,omitempty"`
	SummaryRows          int                 `json:"summary_rows,omitempty"`
	FormulaCells         []string            `json:"formula_cells,omitempty"`
	HighlightedRows      []int               `json:"highlighted_rows,omitempty"`
	WrittenCells         []string            `json:"written_cells,omitempty"`
	Layers               []VerificationLayer `json:"layers,omitempty"`
	Reasons              []string            `json:"reasons"`
}

type VerificationLayer struct {
	Level   int      `json:"level"`
	Name    string   `json:"name"`
	Pass    bool     `json:"pass"`
	Reasons []string `json:"reasons,omitempty"`
}

type VerificationReview struct {
	Status  string   `json:"status"`
	Summary string   `json:"summary"`
	Reasons []string `json:"reasons,omitempty"`
}

type RepairAdvice struct {
	Summary          string   `json:"summary"`
	FailureClass     string   `json:"failure_class"`
	DomainCode       string   `json:"domain_code"`
	RepairHint       string   `json:"repair_hint"`
	SuggestedActions []string `json:"suggested_actions"`
	Assumptions      []string `json:"assumptions,omitempty"`
}

type TelemetryEvent struct {
	RunID     string         `json:"run_id"`
	TraceID   string         `json:"trace_id"`
	EventType string         `json:"event_type"`
	Timestamp string         `json:"timestamp"`
	Component string         `json:"component"`
	Payload   map[string]any `json:"payload,omitempty"`
}

type FailureDetails struct {
	Phase        string `json:"phase"`
	FailureClass string `json:"failure_class"`
	DomainCode   string `json:"domain_code"`
	CriticalStep string `json:"critical_step"`
	Message      string `json:"message"`
}

type FailureEvidence struct {
	ScenarioID             string         `json:"scenario_id"`
	RunID                  string         `json:"run_id"`
	Operation              string         `json:"operation"`
	Phase                  string         `json:"phase"`
	FailureClass           string         `json:"failure_class"`
	DomainCode             string         `json:"domain_code"`
	Message                string         `json:"message"`
	Outcome                string         `json:"outcome"`
	CriticalStep           string         `json:"critical_step"`
	Telemetry              []string       `json:"telemetry"`
	RequestPath            string         `json:"request_path"`
	TaskSpecPath           string         `json:"task_spec_path,omitempty"`
	OperationIRPath        string         `json:"operation_ir_path,omitempty"`
	PlanPath               string         `json:"plan_path,omitempty"`
	PolicyPath             string         `json:"policy_path,omitempty"`
	VerificationPath       string         `json:"verification_path,omitempty"`
	VerificationReviewPath string         `json:"verification_review_path,omitempty"`
	ExecutionPath          string         `json:"execution_path,omitempty"`
	OutcomePath            string         `json:"outcome_path,omitempty"`
	RepairAdvicePath       string         `json:"repair_advice_path,omitempty"`
	RuntimeVersion         RuntimeVersion `json:"runtime_version"`
}

type OutcomeArtifact struct {
	Outcome   string          `json:"outcome"`
	Pass      bool            `json:"pass"`
	Operation string          `json:"operation"`
	Reasons   []string        `json:"reasons,omitempty"`
	Failure   *FailureDetails `json:"failure,omitempty"`
}

type ScenarioEvidence struct {
	ScenarioID             string         `json:"scenario_id"`
	RunID                  string         `json:"run_id"`
	Operation              string         `json:"operation"`
	Request                string         `json:"request"`
	PlanPath               string         `json:"plan_path,omitempty"`
	PolicyPath             string         `json:"policy_path,omitempty"`
	EpisodePath            string         `json:"episode_path,omitempty"`
	VerificationPath       string         `json:"verification_path,omitempty"`
	VerificationReviewPath string         `json:"verification_review_path,omitempty"`
	RenderPath             string         `json:"render_path,omitempty"`
	FailureEvidencePath    string         `json:"failure_evidence_path,omitempty"`
	RepairAdvicePath       string         `json:"repair_advice_path,omitempty"`
	Outcome                string         `json:"outcome"`
	Telemetry              []string       `json:"telemetry"`
	RuntimeVersion         RuntimeVersion `json:"runtime_version"`
}

type RunIDs struct {
	RunID   string
	TraceID string
}

type RunPaths struct {
	TelemetryDir           string
	EvidenceDir            string
	ReportDir              string
	RequestPath            string
	InspectionPath         string
	PlanPath               string
	PolicyPath             string
	TaskSpecPath           string
	OperationIRPath        string
	EpisodePath            string
	ExecutionPath          string
	VerificationPath       string
	VerificationReviewPath string
	RenderPath             string
	OutcomePath            string
	FailureEvidencePath    string
	RepairAdvicePath       string
	EvidenceIndex          string
	ReportPath             string
}

type RunResult struct {
	IDs               RunIDs
	Paths             RunPaths
	Inspection        InspectionArtifact
	SummaryInspection *SummaryInspectionRecord
	Plan              any
	Policy            WritePolicyDecision
	Execution         ExecutionSummary
	Episode           *ExecutionEpisode
	Verification      VerificationResult
	Failure           *FailureDetails
}

func (result RunResult) Operation(fallback string) string {
	for _, candidate := range []string{result.Verification.Operation, result.Execution.Operation, fallback} {
		if strings.TrimSpace(candidate) != "" {
			return candidate
		}
	}
	return ""
}

func (result RunResult) OutputFile(fallback string) string {
	for _, candidate := range []string{result.Verification.OutputFile, result.Execution.OutputFile, fallback} {
		if strings.TrimSpace(candidate) != "" {
			return candidate
		}
	}
	return ""
}
