package validate

import "strings"

func AdmitIntent(intent NormalizedIntent, decision CompilerDecision) (admittedIntent, *Result, error) {
	switch decision.Status {
	case StatusNeedsHumanCheckpoint:
		checkpoint := checkpointFromDecision(decision, intent)
		result, err := finalizeResult(Result{
			Status:     StatusNeedsHumanCheckpoint,
			Checkpoint: &checkpoint,
		})
		return admittedIntent{}, &result, err
	case StatusBlocked:
		result, err := finalizeResult(Result{
			Status: StatusBlocked,
			Blocked: &Blocked{
				FailedStage: StageIntentAdmission,
				ReasonCodes: blockedReasonCodes(decision, intent),
			},
		})
		return admittedIntent{}, &result, err
	}

	switch decision.SelectedOperation {
	case "create_summary_sheet":
		return admitSummaryIntent(intent, decision)
	case "highlight_threshold_rows":
		return admitHighlightIntent(intent, decision)
	case "create_join_lookup_result_sheet":
		return admitJoinLookupIntent(intent, decision)
	case "append_structured_rows":
		return admitAppendRowsIntent(intent, decision)
	default:
		result, err := finalizeResult(Result{
			Status: StatusBlocked,
			Blocked: &Blocked{
				FailedStage: StageIntentAdmission,
				ReasonCodes: blockedReasonCodes(decision, intent, "unsupported_request"),
			},
		})
		return admittedIntent{}, &result, err
	}
}

func admitSummaryIntent(intent NormalizedIntent, decision CompilerDecision) (admittedIntent, *Result, error) {
	if !intent.Materialization.PreserveOriginal ||
		intent.Materialization.OutputDestinationMode != "new_workbook" ||
		intent.Materialization.WriteShape != "new_sheet" {
		result, err := finalizeResult(Result{
			Status: StatusBlocked,
			Blocked: &Blocked{
				FailedStage: StageIntentAdmission,
				ReasonCodes: []string{"invalid_materialization"},
			},
		})
		return admittedIntent{}, &result, err
	}

	groupBy := summaryGroupBy(intent)
	intentMetrics := summaryMetrics(intent)
	if len(groupBy) == 0 || len(intentMetrics) == 0 {
		result, err := finalizeResult(Result{
			Status: StatusBlocked,
			Blocked: &Blocked{
				FailedStage: StageIntentAdmission,
				ReasonCodes: []string{"incomplete_summary_intent"},
			},
		})
		return admittedIntent{}, &result, err
	}

	metrics, metricErrorCode := supportedSummaryMetrics(intentMetrics)
	if metricErrorCode != "" {
		result, err := finalizeResult(Result{
			Status: StatusBlocked,
			Blocked: &Blocked{
				FailedStage: StageIntentAdmission,
				ReasonCodes: []string{metricErrorCode},
			},
		})
		return admittedIntent{}, &result, err
	}

	return admittedIntent{
		decision:        decision,
		compositionKind: "group_summary",
		sourceSheets:    preferredSourceSheets(intent, decision),
		targetSheet:     coalesceString(intent.Summary.TargetSheet, "요약"),
		summaryMode:     coalesceString(intent.Summary.SummaryMode, "values"),
		filters:         cloneFilters(intent.Summary.Filters),
		groupBy:         groupBy,
		metrics:         metrics,
	}, nil, nil
}

func admitHighlightIntent(intent NormalizedIntent, decision CompilerDecision) (admittedIntent, *Result, error) {
	if !intent.Materialization.PreserveOriginal ||
		intent.Materialization.OutputDestinationMode != "new_workbook" ||
		intent.Materialization.WriteShape != "in_place_cells" {
		result, err := finalizeResult(Result{
			Status: StatusBlocked,
			Blocked: &Blocked{
				FailedStage: StageIntentAdmission,
				ReasonCodes: []string{"invalid_materialization"},
			},
		})
		return admittedIntent{}, &result, err
	}

	if strings.TrimSpace(intent.Highlight.TargetColumn) == "" ||
		strings.TrimSpace(intent.Highlight.Operator) == "" ||
		intent.Highlight.Threshold == nil {
		result, err := finalizeResult(Result{
			Status: StatusBlocked,
			Blocked: &Blocked{
				FailedStage: StageIntentAdmission,
				ReasonCodes: []string{"incomplete_highlight_intent"},
			},
		})
		return admittedIntent{}, &result, err
	}

	return admittedIntent{
		decision:        decision,
		compositionKind: "threshold_highlight",
		sourceSheets:    preferredSourceSheets(intent, decision),
		targetColumn:    intent.Highlight.TargetColumn,
		operator:        intent.Highlight.Operator,
		threshold:       intent.Highlight.Threshold,
		highlightColor:  coalesceString(intent.Highlight.HighlightColor, "#FFF59D"),
	}, nil, nil
}

func admitJoinLookupIntent(intent NormalizedIntent, decision CompilerDecision) (admittedIntent, *Result, error) {
	if !intent.Materialization.PreserveOriginal ||
		intent.Materialization.OutputDestinationMode != "new_workbook" ||
		intent.Materialization.WriteShape != "new_sheet" {
		result, err := finalizeResult(Result{
			Status: StatusBlocked,
			Blocked: &Blocked{
				FailedStage: StageIntentAdmission,
				ReasonCodes: []string{"invalid_materialization"},
			},
		})
		return admittedIntent{}, &result, err
	}

	if strings.TrimSpace(intent.JoinLookup.JoinKey) == "" || len(intent.JoinLookup.AppendLookupColumns) == 0 {
		result, err := finalizeResult(Result{
			Status: StatusBlocked,
			Blocked: &Blocked{
				FailedStage: StageIntentAdmission,
				ReasonCodes: []string{"incomplete_join_lookup_intent"},
			},
		})
		return admittedIntent{}, &result, err
	}

	return admittedIntent{
		decision:             decision,
		compositionKind:      "join_lookup",
		sourceSheets:         preferredSourceSheets(intent, decision),
		lookupSheets:         preferredLookupSheets(intent),
		targetSheet:          coalesceString(intent.JoinLookup.TargetSheet, "조회결과"),
		joinKey:              intent.JoinLookup.JoinKey,
		includeSourceColumns: append([]string(nil), intent.JoinLookup.IncludeSourceColumns...),
		appendLookupColumns:  append([]string(nil), intent.JoinLookup.AppendLookupColumns...),
	}, nil, nil
}

func admitAppendRowsIntent(intent NormalizedIntent, decision CompilerDecision) (admittedIntent, *Result, error) {
	if !intent.Materialization.PreserveOriginal ||
		intent.Materialization.OutputDestinationMode != "new_workbook" ||
		intent.Materialization.WriteShape != "in_place_cells" {
		result, err := finalizeResult(Result{
			Status: StatusBlocked,
			Blocked: &Blocked{
				FailedStage: StageIntentAdmission,
				ReasonCodes: []string{"invalid_materialization"},
			},
		})
		return admittedIntent{}, &result, err
	}

	if len(intent.AppendRows.IncludeSourceColumns) == 0 || len(intent.AppendRows.Values) == 0 {
		result, err := finalizeResult(Result{
			Status: StatusBlocked,
			Blocked: &Blocked{
				FailedStage: StageIntentAdmission,
				ReasonCodes: []string{"incomplete_append_rows_intent"},
			},
		})
		return admittedIntent{}, &result, err
	}

	return admittedIntent{
		decision:             decision,
		compositionKind:      "structured_row_append",
		sourceSheets:         preferredSourceSheets(intent, decision),
		includeSourceColumns: append([]string(nil), intent.AppendRows.IncludeSourceColumns...),
		values:               cloneCellValues(intent.AppendRows.Values),
	}, nil, nil
}

func preferredSourceSheets(intent NormalizedIntent, decision CompilerDecision) []string {
	if len(intent.SourceSheetCandidates) > 0 {
		return append([]string(nil), intent.SourceSheetCandidates...)
	}
	if len(decision.SheetCandidates) > 0 {
		return append([]string(nil), decision.SheetCandidates...)
	}
	return []string{}
}

func preferredLookupSheets(intent NormalizedIntent) []string {
	if len(intent.LookupSheetCandidates) > 0 {
		return append([]string(nil), intent.LookupSheetCandidates...)
	}
	return []string{}
}

func blockedReasonCodes(decision CompilerDecision, intent NormalizedIntent, fallback ...string) []string {
	if len(decision.StructuralSignals) > 0 {
		return append([]string(nil), decision.StructuralSignals...)
	}
	if len(intent.Ambiguity.Markers) > 0 {
		return append([]string(nil), intent.Ambiguity.Markers...)
	}
	if len(fallback) > 0 {
		return append([]string(nil), fallback...)
	}
	return []string{"unsupported_request"}
}

func checkpointFromDecision(decision CompilerDecision, intent NormalizedIntent) Checkpoint {
	if decision.Checkpoint != nil {
		return Checkpoint{
			Kind:             decision.Checkpoint.Kind,
			Question:         decision.Checkpoint.Question,
			Options:          append([]string(nil), decision.Checkpoint.Options...),
			UnresolvedFields: checkpointFields(decision.Checkpoint.UnresolvedFields, intent),
		}
	}
	return Checkpoint{
		Kind:             "source_truth_conflict",
		Question:         "어느 시트를 원천 진실로 볼 것인지 먼저 결정해야 합니다.",
		Options:          []string{"Orders_RAW", "Payments_RAW", "block"},
		UnresolvedFields: checkpointFields(nil, intent),
	}
}

func checkpointFields(fields []string, intent NormalizedIntent) []string {
	if len(fields) > 0 {
		return append([]string(nil), fields...)
	}
	if len(intent.Ambiguity.UnresolvedFields) > 0 {
		return append([]string(nil), intent.Ambiguity.UnresolvedFields...)
	}
	return []string{"source_sheet"}
}

func supportedSummaryMetrics(intentMetrics []AggregateIntent) ([]MetricSpec, string) {
	metrics := make([]MetricSpec, 0, len(intentMetrics))
	seen := make(map[string]struct{}, len(intentMetrics))
	for _, metric := range intentMetrics {
		key := metric.Kind + ":" + metric.Column
		if _, ok := seen[key]; ok {
			return nil, "duplicate_metrics"
		}
		seen[key] = struct{}{}

		switch metric.Kind {
		case "sum", "count":
			metrics = append(metrics, MetricSpec{
				Column: metric.Column,
				Op:     metric.Kind,
				As:     metricAlias(metric),
			})
		default:
			return nil, "unsupported_metric_set"
		}
	}
	return metrics, ""
}

func metricAlias(metric AggregateIntent) string {
	switch metric {
	case AggregateIntent{Kind: "sum", Column: "결제금액"}:
		return "총매출"
	case AggregateIntent{Kind: "count", Column: "주문번호"}:
		return "주문수"
	}
	switch metric.Kind {
	case "sum":
		return metric.Column + "_합계"
	default:
		return metric.Column + "_수"
	}
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

func cloneFilters(values []FilterSpec) []FilterSpec {
	if len(values) == 0 {
		return nil
	}
	cloned := make([]FilterSpec, len(values))
	copy(cloned, values)
	return cloned
}

func coalesceString(value string, fallback string) string {
	if strings.TrimSpace(value) != "" {
		return value
	}
	return fallback
}
