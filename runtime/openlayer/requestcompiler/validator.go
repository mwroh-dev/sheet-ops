package requestcompiler

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	runtimeinspect "github.com/mwroh/sheet-ops/runtime/inspect"
	runtimevalidate "github.com/mwroh/sheet-ops/runtime/validate"
)

func ValidateIntent(input Input, intent NormalizedIntent) (Result, error) {
	intent = normalizeIntent(intent)
	if err := validateNormalizedIntent(intent); err != nil {
		return Result{}, err
	}

	primaryInput, err := primaryWorkbookPath(input.InputWorkbooks)
	if err != nil {
		return Result{}, err
	}
	facts, err := runtimeinspect.InspectWorkbookFacts(primaryInput)
	if err != nil {
		return Result{}, err
	}
	requestText, err := requestTextForValidation(input.RequestSource)
	if err != nil {
		return Result{}, err
	}

	decision := compileDecision(intent)
	validation, err := runtimevalidate.Validator{
		RequestContext: runtimevalidate.RequestContext{
			ScenarioID:  input.ScenarioSlug,
			RequestKind: requestKindForValidation(input.RequestSource),
			RequestText: requestText,
			OutputFile:  input.OutputFile,
		},
	}.Validate(toRuntimeIntent(intent), toRuntimeDecision(decision), facts)
	if err != nil {
		return Result{}, err
	}

	return Result{
		Decision:                  decisionFromValidation(intent, decision, validation),
		Validation:                validation,
		ValidatedExecutionRequest: validation.ValidatedExecutionRequest,
	}, nil
}

func requestKindForValidation(source RequestSource) string {
	switch source.Kind {
	case RequestSourcePromptFile, RequestSourceDirectText:
		return "prompt_text"
	default:
		return string(source.Kind)
	}
}

func ValidateIntentAndPersist(input Input, intent NormalizedIntent) (PersistedResult, error) {
	workUnitID, err := NewWorkUnitID(input.ScenarioSlug, time.Now().UTC())
	if err != nil {
		return PersistedResult{}, err
	}
	return validateIntentAndPersist(input, intent, workUnitID)
}

func validateIntentAndPersist(input Input, intent NormalizedIntent, workUnitID string) (PersistedResult, error) {
	intent = normalizeIntent(intent)

	result, err := ValidateIntent(input, intent)
	if err != nil {
		return PersistedResult{}, err
	}

	requestDir, err := RequestArtifactDir(input.WorkspaceRoot, workUnitID)
	if err != nil {
		return PersistedResult{}, err
	}
	requestParentDir := filepath.Dir(requestDir)
	if err := os.MkdirAll(requestParentDir, 0o755); err != nil {
		return PersistedResult{}, err
	}
	stagingDir, err := os.MkdirTemp(requestParentDir, ".request-compiler-staging-*")
	if err != nil {
		return PersistedResult{}, err
	}
	persistSucceeded := false
	defer func() {
		if !persistSucceeded {
			_ = os.RemoveAll(stagingDir)
		}
	}()

	if shouldPersistSensitiveCompilerArtifacts() {
		if err := writeJSON(filepath.Join(stagingDir, "compiler_input.json"), compilerInputArtifact(input, workUnitID)); err != nil {
			return PersistedResult{}, err
		}
		if err := writeFileSync(filepath.Join(stagingDir, "report.md"), []byte(buildIntentValidationReport(result))); err != nil {
			return PersistedResult{}, err
		}
	}
	if err := writeJSON(filepath.Join(stagingDir, "normalized_intent.json"), intent); err != nil {
		return PersistedResult{}, err
	}
	if err := writeJSON(filepath.Join(stagingDir, "compiler_decision.json"), compilerDecisionArtifact(result)); err != nil {
		return PersistedResult{}, err
	}
	if result.Validation.Status == runtimevalidate.StatusCompiled && result.ValidatedExecutionRequest != nil {
		validationResult := any(result.Validation)
		validatedExecutionRequest := any(result.ValidatedExecutionRequest)
		if !shouldPersistSensitiveCompilerArtifacts() {
			validationResult = sanitizeValidationResult(result.Validation)
			validatedExecutionRequest = sanitizeValidatedExecutionRequest(result.ValidatedExecutionRequest)
		}
		if err := writeJSON(filepath.Join(stagingDir, "validation_result.json"), validationResult); err != nil {
			return PersistedResult{}, err
		}
		if err := writeJSON(filepath.Join(stagingDir, "validated_execution_request.json"), validatedExecutionRequest); err != nil {
			return PersistedResult{}, err
		}
	}
	if err := syncDir(stagingDir); err != nil {
		return PersistedResult{}, err
	}
	if err := os.Rename(stagingDir, requestDir); err != nil {
		return PersistedResult{}, err
	}
	if err := syncDir(requestDir); err != nil {
		return PersistedResult{}, err
	}
	if err := syncDir(requestParentDir); err != nil {
		return PersistedResult{}, err
	}
	persistSucceeded = true

	return PersistedResult{
		Result:     result,
		WorkUnitID: workUnitID,
		RequestDir: requestDir,
	}, nil
}

func buildIntentValidationReport(result Result) string {
	files := []string{
		"- `normalized_intent.json`",
		"- `compiler_decision.json`",
	}
	if result.Validation.Status == runtimevalidate.StatusCompiled && result.ValidatedExecutionRequest != nil {
		files = append(files, "- `validation_result.json`", "- `validated_execution_request.json`")
	}
	report := fmt.Sprintf(
		"# Request Compiler Artifacts\n\nStatus: %s\n\n## Files\n%s\n",
		result.Decision.Status,
		stringsJoin(files, "\n"),
	)
	if summary := result.MemoryMatchSummary; summary != nil && summary.hasEntries() {
		report += fmt.Sprintf("\n## Memory Influence\nMatched paragraphs: %d\n", summary.MatchedParagraphCount)
		if len(summary.AppliedDefaults) > 0 {
			report += "\nApplied defaults:\n" + stringsJoin(prefixSummaryLines(summary.AppliedDefaults), "\n") + "\n"
		}
		if len(summary.IgnoredPreferences) > 0 {
			report += "\nIgnored preferences:\n" + stringsJoin(prefixSummaryLines(summary.IgnoredPreferences), "\n") + "\n"
		}
		if len(summary.ConflictNotes) > 0 {
			report += "\nConflict notes:\n" + stringsJoin(prefixSummaryLines(summary.ConflictNotes), "\n") + "\n"
		}
	}
	return report
}

func prefixSummaryLines(values []string) []string {
	lines := make([]string, 0, len(values))
	for _, value := range values {
		lines = append(lines, "- "+value)
	}
	return lines
}

func compileDecision(intent NormalizedIntent) Decision {
	if slices.Contains(intent.Ambiguity.Markers, IntentMarkerSourceTruthConflict) {
		return Decision{
			Status:            StatusNeedsHumanCheckpoint,
			SelectedOperation: "unresolved",
			Notes: []string{
				"the request contains a source-of-truth conflict that must be resolved before execution",
			},
			Checkpoint: Checkpoint{
				Required: true,
				Kind:     "source_truth_conflict",
				Question: "어느 시트를 원천 진실로 볼 것인지 먼저 결정해야 합니다.",
				Options:  []string{"Orders_RAW", "Payments_RAW", "block"},
			},
			SheetCandidates:   checkpointSheetCandidates(intent),
			StructuralSignals: []string{IntentMarkerSourceTruthConflict},
		}
	}

	if containsBlockedMarker(intent.Ambiguity.Markers) {
		return blockedDecision(intent, blockedSignals(intent), "the normalized intent does not map to a supported runtime operation")
	}

	if slices.Contains(intent.CompositionCandidates, CompositionCandidateStructuredRowAppend) && appendRowsIntentReady(intent) {
		return Decision{
			Status:            StatusCompiled,
			SelectedOperation: "append_structured_rows",
			Confidence:        0.98,
			Notes: []string{
				"the normalized intent maps cleanly to the supported structured row append runtime operation",
			},
			SheetCandidates:   supportedSheetCandidates(intent),
			StructuralSignals: []string{"append_structured_rows_request", "preserve_original"},
		}
	}

	if slices.Contains(intent.CompositionCandidates, CompositionCandidateJoinLookup) && joinLookupIntentReady(intent) {
		return Decision{
			Status:            StatusCompiled,
			SelectedOperation: "create_join_lookup_result_sheet",
			Confidence:        0.98,
			Notes: []string{
				"the normalized intent maps cleanly to the supported join lookup runtime operation",
			},
			SheetCandidates:   supportedSheetCandidates(intent),
			StructuralSignals: []string{"join_lookup_request", "preserve_original"},
		}
	}

	if slices.Contains(intent.CompositionCandidates, CompositionCandidateThresholdHighlight) && highlightIntentReady(intent) {
		return Decision{
			Status:            StatusCompiled,
			SelectedOperation: "highlight_threshold_rows",
			Confidence:        0.98,
			Notes: []string{
				"the normalized intent maps cleanly to the supported threshold highlight runtime operation",
			},
			SheetCandidates:   supportedSheetCandidates(intent),
			StructuralSignals: []string{"highlight_threshold_request"},
		}
	}

	if slices.Contains(intent.CompositionCandidates, CompositionCandidateGroupSummary) && summaryIntentReady(intent) {
		return Decision{
			Status:            StatusCompiled,
			SelectedOperation: "create_summary_sheet",
			Confidence:        0.98,
			Notes: []string{
				"the normalized intent maps cleanly to the supported grouped summary runtime operation",
			},
			SheetCandidates:   supportedSheetCandidates(intent),
			StructuralSignals: []string{"summary_request"},
		}
	}

	return blockedDecision(intent, blockedSignals(intent), "the normalized intent does not map to a supported runtime operation")
}

func containsBlockedMarker(markers []string) bool {
	for _, marker := range markers {
		switch marker {
		case IntentMarkerUnsupportedRequest, IntentMarkerInPlaceWriteRequested:
			return true
		}
	}
	return false
}

func blockedDecision(intent NormalizedIntent, structuralSignals []string, note string) Decision {
	if len(structuralSignals) == 0 {
		structuralSignals = []string{IntentMarkerUnsupportedRequest}
	}
	return Decision{
		Status:            StatusBlocked,
		SelectedOperation: "unresolved",
		Notes:             []string{note},
		SheetCandidates:   supportedSheetCandidates(intent),
		StructuralSignals: append([]string(nil), structuralSignals...),
	}
}

func requestTextForValidation(source RequestSource) (string, error) {
	switch source.Kind {
	case "":
		return "", nil
	case RequestSourcePromptFile, RequestSourceDirectText:
		return loadRequestText(source)
	default:
		return "", fmt.Errorf("unknown request source kind %q", source.Kind)
	}
}

func checkpointSheetCandidates(intent NormalizedIntent) []string {
	if len(intent.SourceSheetCandidates) > 0 {
		return append([]string(nil), intent.SourceSheetCandidates...)
	}
	return []string{"Orders_RAW", "Payments_RAW"}
}

func supportedSheetCandidates(intent NormalizedIntent) []string {
	if len(intent.SourceSheetCandidates) > 0 {
		return append([]string(nil), intent.SourceSheetCandidates...)
	}
	return []string{}
}

func blockedSignals(intent NormalizedIntent) []string {
	if slices.Contains(intent.Ambiguity.Markers, IntentMarkerSourceTruthConflict) {
		return []string{IntentMarkerSourceTruthConflict}
	}
	if len(intent.Ambiguity.Markers) > 0 {
		return append([]string(nil), intent.Ambiguity.Markers...)
	}
	return []string{IntentMarkerUnsupportedRequest}
}

func toRuntimeIntent(intent NormalizedIntent) runtimevalidate.NormalizedIntent {
	aggregates := make([]runtimevalidate.AggregateIntent, 0, len(intent.Aggregates))
	for _, aggregate := range intent.Aggregates {
		aggregates = append(aggregates, runtimevalidate.AggregateIntent{
			Kind:   aggregate.Kind,
			Column: aggregate.Column,
		})
	}
	summaryMetrics := make([]runtimevalidate.AggregateIntent, 0, len(intent.Summary.Metrics))
	for _, aggregate := range intent.Summary.Metrics {
		summaryMetrics = append(summaryMetrics, runtimevalidate.AggregateIntent{
			Kind:   aggregate.Kind,
			Column: aggregate.Column,
		})
	}
	summaryFilters := make([]runtimevalidate.FilterSpec, 0, len(intent.Summary.Filters))
	for _, filter := range intent.Summary.Filters {
		summaryFilters = append(summaryFilters, runtimevalidate.FilterSpec{
			Column: filter.Column,
			Op:     filter.Op,
			Value:  filter.Value,
		})
	}

	return runtimevalidate.NormalizedIntent{
		SourceSheetCandidates: append([]string(nil), intent.SourceSheetCandidates...),
		LookupSheetCandidates: append([]string(nil), intent.LookupSheetCandidates...),
		GroupKeys:             append([]string(nil), intent.GroupKeys...),
		Aggregates:            aggregates,
		CompositionCandidates: append([]string(nil), intent.CompositionCandidates...),
		Summary: runtimevalidate.SummaryIntent{
			TargetSheet: intent.Summary.TargetSheet,
			SummaryMode: intent.Summary.SummaryMode,
			Filters:     summaryFilters,
			GroupBy:     append([]string(nil), intent.Summary.GroupBy...),
			Metrics:     summaryMetrics,
		},
		Highlight: runtimevalidate.HighlightIntent{
			TargetColumn:   intent.Highlight.TargetColumn,
			Operator:       intent.Highlight.Operator,
			Threshold:      intent.Highlight.Threshold,
			HighlightColor: intent.Highlight.HighlightColor,
		},
		JoinLookup: runtimevalidate.JoinLookupIntent{
			TargetSheet:          intent.JoinLookup.TargetSheet,
			JoinKey:              intent.JoinLookup.JoinKey,
			IncludeSourceColumns: append([]string(nil), intent.JoinLookup.IncludeSourceColumns...),
			AppendLookupColumns:  append([]string(nil), intent.JoinLookup.AppendLookupColumns...),
		},
		AppendRows: runtimevalidate.AppendRowsIntent{
			IncludeSourceColumns: append([]string(nil), intent.AppendRows.IncludeSourceColumns...),
			Values:               toRuntimeCellValues(intent.AppendRows.Values),
		},
		Materialization: runtimevalidate.MaterializationIntent{
			PreserveOriginal:      intent.Materialization.PreserveOriginal,
			OutputDestinationMode: intent.Materialization.OutputDestinationMode,
			WriteShape:            intent.Materialization.WriteShape,
		},
		Ambiguity: runtimevalidate.AmbiguityIntent{
			Markers:          append([]string(nil), intent.Ambiguity.Markers...),
			UnresolvedFields: append([]string(nil), intent.Ambiguity.UnresolvedFields...),
			CheckpointHints:  append([]string(nil), intent.Ambiguity.CheckpointHints...),
		},
	}
}

func toRuntimeDecision(decision Decision) runtimevalidate.CompilerDecision {
	runtimeDecision := runtimevalidate.CompilerDecision{
		Status:            runtimevalidate.Status(decision.Status),
		SelectedOperation: decision.SelectedOperation,
		Notes:             append([]string(nil), decision.Notes...),
		SheetCandidates:   append([]string(nil), decision.SheetCandidates...),
		StructuralSignals: append([]string(nil), decision.StructuralSignals...),
	}
	if decision.Checkpoint.Required {
		runtimeDecision.Checkpoint = &runtimevalidate.Checkpoint{
			Kind:             decision.Checkpoint.Kind,
			Question:         decision.Checkpoint.Question,
			Options:          append([]string(nil), decision.Checkpoint.Options...),
			UnresolvedFields: []string{"source_sheet"},
		}
	}
	return runtimeDecision
}

func decisionFromValidation(intent NormalizedIntent, provisional Decision, validation runtimevalidate.Result) Decision {
	switch validation.Status {
	case runtimevalidate.StatusCompiled:
		compiled := provisional
		compiled.Status = StatusCompiled
		compiled.Checkpoint = Checkpoint{}
		if validation.ValidatedExecutionRequest != nil {
			compiled.SheetCandidates = []string{validation.ValidatedExecutionRequest.SourceSheet}
		}
		return compiled
	case runtimevalidate.StatusNeedsHumanCheckpoint:
		decision := provisional
		decision.Status = StatusNeedsHumanCheckpoint
		decision.SelectedOperation = "unresolved"
		decision.Confidence = 0
		decision.StructuralSignals = []string{IntentMarkerSourceTruthConflict}
		if validation.Checkpoint != nil {
			decision.Checkpoint = Checkpoint{
				Required: true,
				Kind:     validation.Checkpoint.Kind,
				Question: validation.Checkpoint.Question,
				Options:  append([]string(nil), validation.Checkpoint.Options...),
			}
		}
		return decision
	default:
		reasonCodes := []string{IntentMarkerUnsupportedRequest}
		if validation.Blocked != nil && len(validation.Blocked.ReasonCodes) > 0 {
			reasonCodes = append([]string(nil), validation.Blocked.ReasonCodes...)
		}
		return blockedDecision(intent, reasonCodes, blockedDecisionNote(validation))
	}
}

func blockedDecisionNote(validation runtimevalidate.Result) string {
	if validation.Blocked == nil {
		return "the normalized intent does not map to a supported runtime operation"
	}
	switch validation.Blocked.FailedStage {
	case runtimevalidate.StageFactBinding:
		if slices.Contains(validation.Blocked.ReasonCodes, "missing_required_columns") {
			return "the source sheet is missing required columns for the supported summary operation"
		}
		return "the workbook facts do not support the requested summary operation"
	case runtimevalidate.StageExecutionAdmission:
		return "the validated execution request could not be admitted"
	default:
		return "the normalized intent does not map to a supported runtime operation"
	}
}

func summaryIntentReady(intent NormalizedIntent) bool {
	return len(summaryGroupBy(intent)) > 0 && len(summaryMetrics(intent)) > 0
}

func highlightIntentReady(intent NormalizedIntent) bool {
	return strings.TrimSpace(intent.Highlight.TargetColumn) != "" &&
		strings.TrimSpace(intent.Highlight.Operator) != "" &&
		intent.Highlight.Threshold != nil
}

func joinLookupIntentReady(intent NormalizedIntent) bool {
	return len(intent.SourceSheetCandidates) > 0 &&
		len(intent.LookupSheetCandidates) > 0 &&
		strings.TrimSpace(intent.JoinLookup.JoinKey) != "" &&
		len(intent.JoinLookup.AppendLookupColumns) > 0
}

func appendRowsIntentReady(intent NormalizedIntent) bool {
	return len(intent.SourceSheetCandidates) > 0 &&
		len(intent.AppendRows.IncludeSourceColumns) > 0 &&
		len(intent.AppendRows.Values) > 0
}

func toRuntimeCellValues(values []CellValue) []runtimevalidate.CellValue {
	if len(values) == 0 {
		return []runtimevalidate.CellValue{}
	}
	converted := make([]runtimevalidate.CellValue, 0, len(values))
	for _, value := range values {
		converted = append(converted, runtimevalidate.CellValue{
			Cell:  value.Cell,
			Value: value.Value,
		})
	}
	return converted
}

func summaryGroupBy(intent NormalizedIntent) []string {
	if len(intent.Summary.GroupBy) > 0 {
		return append([]string(nil), intent.Summary.GroupBy...)
	}
	return append([]string(nil), intent.GroupKeys...)
}

func summaryMetrics(intent NormalizedIntent) []AggregateIntent {
	if len(intent.Summary.Metrics) > 0 {
		return append([]AggregateIntent(nil), intent.Summary.Metrics...)
	}
	return append([]AggregateIntent(nil), intent.Aggregates...)
}

func stringsJoin(values []string, separator string) string {
	switch len(values) {
	case 0:
		return ""
	case 1:
		return values[0]
	}

	total := 0
	for _, value := range values {
		total += len(value)
	}
	total += len(separator) * (len(values) - 1)

	builder := make([]byte, 0, total)
	for index, value := range values {
		if index > 0 {
			builder = append(builder, separator...)
		}
		builder = append(builder, value...)
	}
	return string(builder)
}
