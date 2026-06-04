package requestcompiler

const (
	CompositionCandidateGroupSummary       = "group_summary"
	CompositionCandidateThresholdHighlight = "threshold_highlight"
	CompositionCandidateJoinLookup         = "join_lookup"

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
	SourceSheetCandidates []string              `json:"source_sheet_candidates,omitempty"`
	LookupSheetCandidates []string              `json:"lookup_sheet_candidates,omitempty"`
	GroupKeys             []string              `json:"group_keys,omitempty"`
	Aggregates            []AggregateIntent     `json:"aggregates,omitempty"`
	CompositionCandidates []string              `json:"composition_candidates,omitempty"`
	Summary               SummaryIntent         `json:"summary,omitempty"`
	Highlight             HighlightIntent       `json:"highlight,omitempty"`
	JoinLookup            JoinLookupIntent      `json:"join_lookup,omitempty"`
	Materialization       MaterializationIntent `json:"materialization"`
	Ambiguity             AmbiguityIntent       `json:"ambiguity"`
	Ambiguities           []string              `json:"-"`
}
