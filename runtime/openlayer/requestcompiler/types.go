package requestcompiler

const (
	CompositionCandidateGroupSummary        = "group_summary"
	CompositionCandidateThresholdHighlight  = "threshold_highlight"
	CompositionCandidateJoinLookup          = "join_lookup"
	CompositionCandidateStructuredRowAppend = "structured_row_append"
	CompositionCandidateFormulaExtension    = "formula_extension"
	CompositionCandidatePeriodCopy          = "period_copy"
	CompositionCandidateDataValidation      = "data_validation"
	CompositionCandidateFormulaProtection   = "formula_protection"
	CompositionCandidateHeaderNormalization = "header_normalization"
	CompositionCandidatePeriodRollForward   = "period_roll_forward"
	CompositionCandidateTableReconciliation = "table_reconciliation"
	CompositionCandidatePrintableForm       = "printable_form"

	OutputDestinationModeNewWorkbook  = "new_workbook"
	OutputDestinationModeSameWorkbook = "same_workbook"

	WriteShapeNewSheet     = "new_sheet"
	WriteShapeInPlaceCells = "in_place_cells"

	IntentMarkerUnsupportedRequest           = "unsupported_request"
	IntentMarkerContradictorySummaryIntent   = "contradictory_summary_intent"
	IntentMarkerInPlaceWriteRequested        = "in_place_write_requested"
	IntentMarkerMissingCancellationExclusion = "missing_cancellation_exclusion"
	IntentMarkerSourceTruthConflictRequested = "source_truth_conflict_requested"
	IntentMarkerSourceTruthConflict          = "source_truth_conflict"

	UnresolvedFieldSourceSheet        = "source_sheet"
	CheckpointHintSourceTruthConflict = "source_truth_conflict"
)

type AggregateIntent struct {
	Kind   string `json:"kind"`
	Column string `json:"column"`
}

type SummaryIntent struct {
	TargetSheet string            `json:"target_sheet,omitempty"`
	SummaryMode string            `json:"summary_mode,omitempty"`
	Filters     []FilterIntent    `json:"filters,omitempty"`
	GroupBy     []string          `json:"group_by,omitempty"`
	Metrics     []AggregateIntent `json:"metrics,omitempty"`
}

type HighlightIntent struct {
	TargetColumn   string   `json:"target_column,omitempty"`
	Operator       string   `json:"operator,omitempty"`
	Threshold      *float64 `json:"threshold,omitempty"`
	HighlightColor string   `json:"highlight_color,omitempty"`
}

type JoinLookupIntent struct {
	TargetSheet          string   `json:"target_sheet,omitempty"`
	JoinKey              string   `json:"join_key,omitempty"`
	IncludeSourceColumns []string `json:"include_source_columns,omitempty"`
	AppendLookupColumns  []string `json:"append_lookup_columns,omitempty"`
}

type AppendRowsIntent struct {
	IncludeSourceColumns []string    `json:"include_source_columns,omitempty"`
	Values               []CellValue `json:"values,omitempty"`
}

type ExtendFormulasIntent struct {
	FormulaSourceRow int      `json:"formula_source_row,omitempty"`
	TargetRows       []int    `json:"target_rows,omitempty"`
	FormulaColumns   []string `json:"formula_columns,omitempty"`
}

type PeriodCopyIntent struct {
	TargetSheet string `json:"target_sheet,omitempty"`
}

type AddDataValidationIntent struct {
	ValidationRule DataValidationRule `json:"validation_rule,omitempty"`
}

type ProtectFormulaCellsIntent struct {
	ProtectionRule FormulaProtectionRule `json:"protection_rule,omitempty"`
}

type NormalizeHeadersIntent struct {
	HeaderRow      int             `json:"header_row,omitempty"`
	HeaderMappings []HeaderMapping `json:"header_mappings,omitempty"`
}

type RollForwardPeriodIntent struct {
	TargetSheet          string                `json:"target_sheet,omitempty"`
	CarryForwardMappings []CarryForwardMapping `json:"carry_forward_mappings,omitempty"`
}

type ReconcileTablesIntent struct {
	TargetSheet     string           `json:"target_sheet,omitempty"`
	LeftKey         string           `json:"left_key,omitempty"`
	RightKey        string           `json:"right_key,omitempty"`
	CompareMappings []CompareMapping `json:"compare_mappings,omitempty"`
}

type GeneratePrintableFormIntent struct {
	TargetSheet   string             `json:"target_sheet,omitempty"`
	FormTitle     string             `json:"form_title,omitempty"`
	PrintArea     string             `json:"print_area,omitempty"`
	FieldBindings []FormFieldBinding `json:"field_bindings,omitempty"`
	TableBinding  *FormTableBinding  `json:"table_binding,omitempty"`
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

type CellValue struct {
	Cell  string `json:"cell"`
	Value any    `json:"value"`
}

type FilterIntent struct {
	Column string `json:"column"`
	Op     string `json:"op"`
	Value  string `json:"value"`
}

type MaterializationIntent struct {
	PreserveOriginal      bool   `json:"preserve_original"`
	OutputDestinationMode string `json:"output_destination_mode"`
	WriteShape            string `json:"write_shape"`
}

type AmbiguityIntent struct {
	Markers          []string `json:"markers"`
	UnresolvedFields []string `json:"unresolved_fields"`
	CheckpointHints  []string `json:"checkpoint_hints"`
}

type MemoryPreferenceHints struct {
	SummaryMode           string `json:"summary_mode_preference,omitempty"`
	PreserveOriginal      *bool  `json:"preserve_original_preference,omitempty"`
	OutputDestinationMode string `json:"output_destination_preference,omitempty"`
	WriteShape            string `json:"write_shape_preference,omitempty"`
}

type PreferenceObservation struct {
	Key        string `json:"key"`
	Value      string `json:"value"`
	RunID      string `json:"run_id,omitempty"`
	ScenarioID string `json:"scenario_id,omitempty"`
	Operation  string `json:"operation,omitempty"`
	RecordedAt string `json:"recorded_at,omitempty"`
}

type MemoryMatchSummary struct {
	MatchedParagraphCount int      `json:"matched_paragraph_count"`
	AppliedDefaults       []string `json:"applied_defaults,omitempty"`
	IgnoredPreferences    []string `json:"ignored_preferences,omitempty"`
	ConflictNotes         []string `json:"conflict_notes,omitempty"`
}

type NormalizedIntent struct {
	SourceSheetCandidates []string                    `json:"source_sheet_candidates,omitempty"`
	LookupSheetCandidates []string                    `json:"lookup_sheet_candidates,omitempty"`
	GroupKeys             []string                    `json:"group_keys,omitempty"`
	Aggregates            []AggregateIntent           `json:"aggregates,omitempty"`
	CompositionCandidates []string                    `json:"composition_candidates,omitempty"`
	Summary               SummaryIntent               `json:"summary,omitempty"`
	Highlight             HighlightIntent             `json:"highlight,omitempty"`
	JoinLookup            JoinLookupIntent            `json:"join_lookup,omitempty"`
	AppendRows            AppendRowsIntent            `json:"append_rows,omitempty"`
	ExtendFormulas        ExtendFormulasIntent        `json:"extend_formulas,omitempty"`
	PeriodCopy            PeriodCopyIntent            `json:"period_copy,omitempty"`
	AddDataValidation     AddDataValidationIntent     `json:"add_data_validation,omitempty"`
	ProtectFormulaCells   ProtectFormulaCellsIntent   `json:"protect_formula_cells,omitempty"`
	NormalizeHeaders      NormalizeHeadersIntent      `json:"normalize_headers,omitempty"`
	RollForwardPeriod     RollForwardPeriodIntent     `json:"roll_forward_period,omitempty"`
	ReconcileTables       ReconcileTablesIntent       `json:"reconcile_tables,omitempty"`
	GeneratePrintableForm GeneratePrintableFormIntent `json:"generate_printable_form,omitempty"`
	Materialization       MaterializationIntent       `json:"materialization"`
	Ambiguity             AmbiguityIntent             `json:"ambiguity"`
	Ambiguities           []string                    `json:"-"`
}
