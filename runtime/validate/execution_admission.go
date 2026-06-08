package validate

func (validator Validator) AdmitExecution(bound boundIntent) (Result, error) {
	request := &ValidatedExecutionRequest{
		ScenarioID:      validator.RequestContext.ScenarioID,
		RequestKind:     validator.RequestContext.RequestKind,
		RequestText:     validator.RequestContext.RequestText,
		InputFile:       bound.facts.InputFile,
		SourceSheet:     bound.sourceSheet,
		OutputFile:      validator.RequestContext.OutputFile,
		ExecutionKind:   "composition",
		CompositionKind: bound.compositionKind,
	}

	switch bound.compositionKind {
	case "group_summary":
		request.TargetSheet = bound.targetSheet
		request.SummaryMode = bound.summaryMode
		request.Filters = append([]FilterSpec(nil), bound.filters...)
		request.GroupBy = append([]string(nil), bound.groupBy...)
		request.Metrics = append([]MetricSpec(nil), bound.metrics...)
	case "threshold_highlight":
		request.TargetColumn = bound.targetColumn
		request.Operator = bound.operator
		request.Threshold = bound.threshold
		request.HighlightColor = bound.highlightColor
	case "join_lookup":
		request.TargetSheet = bound.targetSheet
		request.LookupSheet = bound.lookupSheet
		request.JoinKey = bound.joinKey
		if len(bound.includeSourceColumns) > 0 && len(bound.admittedIntent.includeSourceColumns) > 0 {
			request.IncludeSourceColumns = append([]string(nil), bound.includeSourceColumns...)
		}
		request.AppendLookupColumns = append([]string(nil), bound.appendLookupColumns...)
	case "structured_row_append":
		request.IncludeSourceColumns = append([]string(nil), bound.includeSourceColumns...)
		request.Values = cloneCellValues(bound.values)
	case "formula_extension":
		request.FormulaSourceRow = bound.formulaSourceRow
		request.TargetRows = cloneInts(bound.targetRows)
		request.FormulaColumns = append([]string(nil), bound.formulaColumns...)
	case "data_validation":
		rule := cloneDataValidationRule(bound.validationRule)
		request.ValidationRule = &rule
	}

	return finalizeResult(Result{
		Status:                    StatusCompiled,
		ValidatedExecutionRequest: request,
	})
}
