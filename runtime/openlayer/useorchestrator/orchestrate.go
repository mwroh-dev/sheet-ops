package useorchestrator

import (
	"fmt"

	runtimeinspect "github.com/mwroh/sheet-ops/runtime/inspect"
	runtimetaskspec "github.com/mwroh/sheet-ops/runtime/taskspec"
	runtimevalidate "github.com/mwroh/sheet-ops/runtime/validate"
	runtimeworkbookcase "github.com/mwroh/sheet-ops/runtime/workbookcase"
)

func Orchestrate(req UseRequest) (RunResult, error) {
	if req.Operation != "" && !isSupportedOperation(req.Operation) {
		return RunResult{}, fmt.Errorf("unsupported operation %q", req.Operation)
	}

	normalizedReq := NormalizeUseRequest(req)
	if err := validateRequestValue(normalizedReq); err != nil {
		return RunResult{}, err
	}

	routing := loadRoutingContext(normalizedReq)
	if routing.selectedOperation == "" {
		return RunResult{}, fmt.Errorf("could not infer a supported operation from request signals")
	}
	return runTaskSpec(normalizedReq.ScenarioID, taskSpecFromUseRequest(normalizedReq), routing)
}

func OrchestrateValidated(req ValidatedExecutionRequest) (RunResult, error) {
	if err := validateSupportedValidatedExecutionRequest(req); err != nil {
		return RunResult{}, err
	}
	if err := validateValidatedExecutionRequestValue(req); err != nil {
		return RunResult{}, err
	}
	taskSpec, err := taskSpecFromValidatedExecutionRequest(req)
	if err != nil {
		return RunResult{}, err
	}
	routing := loadRoutingContext(validatedRoutingUseRequest(req))
	return runTaskSpec(req.ScenarioID, taskSpec, routing)
}

func validateSupportedValidatedExecutionRequest(req ValidatedExecutionRequest) error {
	if req.ExecutionKind != "composition" {
		return fmt.Errorf("unsupported validated execution kind %q", req.ExecutionKind)
	}
	switch req.CompositionKind {
	case "structured_row_append":
		return nil
	case "group_summary", "threshold_highlight", "join_lookup":
		return nil
	default:
		return fmt.Errorf("unsupported validated composition kind %q", req.CompositionKind)
	}
}

func runTaskSpec(scenarioID string, spec runtimetaskspec.TaskSpec, routing routingContext) (RunResult, error) {
	return runtimeworkbookcase.Run(runtimeworkbookcase.Request{
		ScenarioID: scenarioID,
		TaskSpec:   spec,
		Routing: &runtimeworkbookcase.RoutingContext{
			SelectedOperation:      routing.selectedOperation,
			RoutingAuthority:       routing.routingAuthority,
			KnowledgeRecordsLoaded: routing.knowledge.RecordCount,
			KnowledgeFallback:      routing.knowledge.UsedFallback,
			KnowledgePath:          routing.knowledge.KnowledgePath,
			RoutingNote:            routing.note,
			FallbackReason:         routing.knowledge.FallbackReason,
			RecentOperations:       append([]string(nil), routing.knowledge.RecentOperations...),
		},
	})
}

func taskSpecFromUseRequest(req UseRequest) runtimetaskspec.TaskSpec {
	effectiveOperation, _ := resolveRequestedOperation(req)
	switch effectiveOperation {
	case runtimeworkbookcase.AppendRowsOperationName:
		return runtimetaskspec.BuildAppendStructuredRowsTask(runtimetaskspec.AppendStructuredRowsRequest{
			RequestText:          coalesceValidatedRequestText(req.RequestText, "structured append rows request"),
			InputFile:            req.InputFile,
			SourceSheet:          req.SheetName,
			OutputFile:           req.OutputFile,
			IncludeSourceColumns: append([]string(nil), req.IncludeSourceColumns...),
			Values:               toTaskSpecCellValues(req.Values),
		}).TaskSpec
	case runtimeworkbookcase.SummaryOperationName:
		return runtimetaskspec.BuildGroupSummarizeTask(runtimetaskspec.UseRequest{
			RequestText: req.RequestText,
			InputFile:   req.InputFile,
			SourceSheet: req.SheetName,
			OutputFile:  req.OutputFile,
			TargetSheet: req.TargetSheet,
			SummaryMode: req.SummaryMode,
			Filters:     toTaskSpecFilters(req.Filters),
			GroupBy:     append([]string(nil), req.GroupBy...),
			Metrics:     toTaskSpecMetrics(req.Metrics),
		}).TaskSpec
	case runtimeworkbookcase.JoinLookupOperationName:
		return runtimetaskspec.BuildJoinLookupTask(runtimetaskspec.JoinLookupRequest{
			RequestText:          coalesceValidatedRequestText(req.RequestText, "structured join lookup request"),
			InputFile:            req.InputFile,
			SourceSheet:          req.SheetName,
			LookupSheet:          req.LookupSheet,
			OutputFile:           req.OutputFile,
			TargetSheet:          req.TargetSheet,
			JoinKey:              req.JoinKey,
			IncludeSourceColumns: append([]string(nil), req.IncludeSourceColumns...),
			AppendLookupColumns:  append([]string(nil), req.AppendLookupColumns...),
		}).TaskSpec
	default:
		return runtimetaskspec.BuildHighlightThresholdTask(runtimetaskspec.HighlightThresholdRequest{
			RequestText:    coalesceValidatedRequestText(req.RequestText, "structured threshold highlight request"),
			InputFile:      req.InputFile,
			SourceSheet:    req.SheetName,
			OutputFile:     req.OutputFile,
			Column:         req.TargetColumn,
			Operator:       req.Operator,
			Threshold:      req.Threshold,
			HighlightColor: req.HighlightColor,
		}).TaskSpec
	}
}

func taskSpecFromValidatedExecutionRequest(req ValidatedExecutionRequest) (runtimetaskspec.TaskSpec, error) {
	switch req.CompositionKind {
	case "structured_row_append":
		return runtimetaskspec.BuildAppendStructuredRowsTask(runtimetaskspec.AppendStructuredRowsRequest{
			RequestText:          coalesceValidatedRequestText(req.RequestText, "validated append rows request"),
			InputFile:            req.InputFile,
			SourceSheet:          req.SourceSheet,
			OutputFile:           req.OutputFile,
			IncludeSourceColumns: append([]string(nil), req.IncludeSourceColumns...),
			Values:               toTaskSpecCellValues(req.Values),
		}).TaskSpec, nil
	case "join_lookup":
		includeSourceColumns, err := defaultJoinSourceColumns(req)
		if err != nil {
			return runtimetaskspec.TaskSpec{}, err
		}
		return runtimetaskspec.BuildJoinLookupTask(runtimetaskspec.JoinLookupRequest{
			RequestText:          coalesceValidatedRequestText(req.RequestText, "validated join lookup request"),
			InputFile:            req.InputFile,
			SourceSheet:          req.SourceSheet,
			LookupSheet:          req.LookupSheet,
			OutputFile:           req.OutputFile,
			TargetSheet:          req.TargetSheet,
			JoinKey:              req.JoinKey,
			IncludeSourceColumns: includeSourceColumns,
			AppendLookupColumns:  append([]string(nil), req.AppendLookupColumns...),
		}).TaskSpec, nil
	case "threshold_highlight":
		return runtimetaskspec.BuildHighlightThresholdTask(runtimetaskspec.HighlightThresholdRequest{
			RequestText:    coalesceValidatedRequestText(req.RequestText, "validated threshold highlight request"),
			InputFile:      req.InputFile,
			SourceSheet:    req.SourceSheet,
			OutputFile:     req.OutputFile,
			Column:         req.TargetColumn,
			Operator:       req.Operator,
			Threshold:      req.Threshold,
			HighlightColor: req.HighlightColor,
		}).TaskSpec, nil
	default:
		return taskSpecFromUseRequest(validatedSummaryUseRequest(req)), nil
	}
}

func validatedSummaryUseRequest(req ValidatedExecutionRequest) UseRequest {
	return UseRequest{
		ScenarioID:  req.ScenarioID,
		RequestText: coalesceValidatedRequestText(req.RequestText, "validated summary request"),
		InputFile:   req.InputFile,
		SheetName:   req.SourceSheet,
		OutputFile:  req.OutputFile,
		Operation:   runtimeworkbookcase.SummaryOperationName,
		TargetSheet: req.TargetSheet,
		SummaryMode: req.SummaryMode,
		Filters:     toUseRequestFilters(req.Filters),
		GroupBy:     append([]string(nil), req.GroupBy...),
		Metrics:     toUseRequestMetrics(req.Metrics),
	}
}

func validatedRoutingUseRequest(req ValidatedExecutionRequest) UseRequest {
	return UseRequest{
		ScenarioID:  req.ScenarioID,
		RequestText: validatedRoutingRequestText(req),
		InputFile:   req.InputFile,
		SheetName:   req.SourceSheet,
		OutputFile:  req.OutputFile,
		Operation:   validatedOperation(req),
	}
}

func validatedRoutingRequestText(req ValidatedExecutionRequest) string {
	switch req.CompositionKind {
	case "structured_row_append":
		return coalesceValidatedRequestText(req.RequestText, "validated append rows request")
	case "join_lookup":
		return coalesceValidatedRequestText(req.RequestText, "validated join lookup request")
	case "threshold_highlight":
		return coalesceValidatedRequestText(req.RequestText, "validated threshold highlight request")
	default:
		return coalesceValidatedRequestText(req.RequestText, "validated summary request")
	}
}

func validatedOperation(req ValidatedExecutionRequest) string {
	switch req.CompositionKind {
	case "structured_row_append":
		return runtimeworkbookcase.AppendRowsOperationName
	case "join_lookup":
		return runtimeworkbookcase.JoinLookupOperationName
	case "threshold_highlight":
		return runtimeworkbookcase.HighlightOperationName
	default:
		return runtimeworkbookcase.SummaryOperationName
	}
}

func coalesceValidatedRequestText(value string, fallback string) string {
	if value != "" {
		return value
	}
	return fallback
}

func defaultJoinSourceColumns(req ValidatedExecutionRequest) ([]string, error) {
	if len(req.IncludeSourceColumns) > 0 {
		return append([]string(nil), req.IncludeSourceColumns...), nil
	}
	facts, err := runtimeinspect.InspectWorkbookFacts(req.InputFile)
	if err != nil {
		return nil, fmt.Errorf("default join lookup include_source_columns: inspect workbook %q: %w", req.InputFile, err)
	}
	for _, sheet := range facts.Sheets {
		if sheet.Name == req.SourceSheet {
			if len(sheet.Columns) == 0 {
				return nil, fmt.Errorf("default join lookup include_source_columns: source sheet %q in workbook %q has no header columns", req.SourceSheet, req.InputFile)
			}
			return append([]string(nil), sheet.Columns...), nil
		}
	}
	return nil, fmt.Errorf("default join lookup include_source_columns: source sheet %q not found in workbook %q", req.SourceSheet, req.InputFile)
}

func isSupportedOperation(operation string) bool {
	switch operation {
	case runtimeworkbookcase.SummaryOperationName, runtimeworkbookcase.HighlightOperationName, runtimeworkbookcase.JoinLookupOperationName, runtimeworkbookcase.AppendRowsOperationName:
		return true
	default:
		return false
	}
}

func toTaskSpecCellValues(values []CellValue) []runtimetaskspec.CellValue {
	if len(values) == 0 {
		return nil
	}
	converted := make([]runtimetaskspec.CellValue, 0, len(values))
	for _, value := range values {
		converted = append(converted, runtimetaskspec.CellValue{Cell: value.Cell, Value: value.Value})
	}
	return converted
}

func toTaskSpecFilters(filters []FilterSpec) []runtimetaskspec.FilterSpec {
	if len(filters) == 0 {
		return nil
	}
	converted := make([]runtimetaskspec.FilterSpec, 0, len(filters))
	for _, filter := range filters {
		converted = append(converted, runtimetaskspec.FilterSpec{
			Column: filter.Column,
			Op:     filter.Op,
			Value:  filter.Value,
		})
	}
	return converted
}

func toTaskSpecMetrics(metrics []MetricSpec) []runtimetaskspec.MetricSpec {
	if len(metrics) == 0 {
		return nil
	}
	converted := make([]runtimetaskspec.MetricSpec, 0, len(metrics))
	for _, metric := range metrics {
		converted = append(converted, runtimetaskspec.MetricSpec{
			Column: metric.Column,
			Op:     metric.Op,
			As:     metric.As,
		})
	}
	return converted
}

func toUseRequestFilters(filters []runtimevalidate.FilterSpec) []FilterSpec {
	if len(filters) == 0 {
		return nil
	}
	converted := make([]FilterSpec, 0, len(filters))
	for _, filter := range filters {
		converted = append(converted, FilterSpec{
			Column: filter.Column,
			Op:     filter.Op,
			Value:  filter.Value,
		})
	}
	return converted
}

func toUseRequestMetrics(metrics []runtimevalidate.MetricSpec) []MetricSpec {
	if len(metrics) == 0 {
		return nil
	}
	converted := make([]MetricSpec, 0, len(metrics))
	for _, metric := range metrics {
		converted = append(converted, MetricSpec{
			Column: metric.Column,
			Op:     metric.Op,
			As:     metric.As,
		})
	}
	return converted
}
