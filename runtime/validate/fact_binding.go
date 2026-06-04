package validate

import runtimeinspect "github.com/mwroh/sheet-ops/runtime/inspect"

func BindFacts(admitted admittedIntent, facts runtimeinspect.WorkbookFacts) (boundIntent, *Result, error) {
	switch admitted.compositionKind {
	case "group_summary":
		sourceSheet, ok := firstExistingSheet(facts, admitted.sourceSheets)
		if !ok {
			return blockedFactBinding("unsupported_source_sheet")
		}
		if missingColumns := missingWorkbookColumns(facts, sourceSheet, requiredSummaryColumns(admitted)); len(missingColumns) > 0 {
			return blockedFactBinding("missing_required_columns")
		}
		return boundIntent{
			admittedIntent: admitted,
			facts:          facts,
			sourceSheet:    sourceSheet,
		}, nil, nil
	case "threshold_highlight":
		sourceSheet, ok := firstExistingSheet(facts, admitted.sourceSheets)
		if !ok {
			return blockedFactBinding("unsupported_source_sheet")
		}
		if missingColumns := missingWorkbookColumns(facts, sourceSheet, []string{admitted.targetColumn}); len(missingColumns) > 0 {
			return blockedFactBinding("missing_required_columns")
		}
		return boundIntent{
			admittedIntent: admitted,
			facts:          facts,
			sourceSheet:    sourceSheet,
		}, nil, nil
	case "join_lookup":
		sourceSheet, ok := firstExistingSheet(facts, admitted.sourceSheets)
		if !ok {
			return blockedFactBinding("unsupported_source_sheet")
		}
		lookupSheet, ok := firstExistingLookupSheet(facts, admitted.lookupSheets, sourceSheet)
		if !ok {
			return blockedFactBinding("unsupported_lookup_sheet")
		}
		if missingColumns := missingWorkbookColumns(facts, sourceSheet, requiredJoinSourceColumns(admitted)); len(missingColumns) > 0 {
			return blockedFactBinding("missing_required_columns")
		}
		if missingColumns := missingWorkbookColumns(facts, lookupSheet, requiredJoinLookupColumns(admitted)); len(missingColumns) > 0 {
			return blockedFactBinding("missing_required_columns")
		}
		includeSourceColumns := append([]string(nil), admitted.includeSourceColumns...)
		if len(includeSourceColumns) == 0 {
			sheet, _ := factsSheet(facts, sourceSheet)
			includeSourceColumns = append([]string(nil), sheet.Columns...)
		}
		return boundIntent{
			admittedIntent:       admitted,
			facts:                facts,
			sourceSheet:          sourceSheet,
			lookupSheet:          lookupSheet,
			includeSourceColumns: includeSourceColumns,
		}, nil, nil
	default:
		return blockedFactBinding("unsupported_request")
	}
}

func blockedFactBinding(reasonCode string) (boundIntent, *Result, error) {
	result, err := finalizeResult(Result{
		Status: StatusBlocked,
		Blocked: &Blocked{
			FailedStage: StageFactBinding,
			ReasonCodes: []string{reasonCode},
		},
	})
	return boundIntent{}, &result, err
}

func firstExistingSheet(facts runtimeinspect.WorkbookFacts, candidates []string) (string, bool) {
	for _, candidate := range candidates {
		if _, ok := factsSheet(facts, candidate); ok {
			return candidate, true
		}
	}
	return "", false
}

func firstExistingLookupSheet(facts runtimeinspect.WorkbookFacts, candidates []string, sourceSheet string) (string, bool) {
	for _, candidate := range candidates {
		if candidate == sourceSheet {
			continue
		}
		if _, ok := factsSheet(facts, candidate); ok {
			return candidate, true
		}
	}
	for _, sheet := range facts.Sheets {
		if sheet.Name != sourceSheet {
			return sheet.Name, true
		}
	}
	return "", false
}

func requiredSummaryColumns(admitted admittedIntent) []string {
	columns := make([]string, 0, len(admitted.groupBy)+len(admitted.metrics)+len(admitted.filters))
	columns = append(columns, admitted.groupBy...)
	for _, metric := range admitted.metrics {
		columns = append(columns, metric.Column)
	}
	for _, filter := range admitted.filters {
		columns = append(columns, filter.Column)
	}
	return columns
}

func requiredJoinSourceColumns(admitted admittedIntent) []string {
	columns := []string{admitted.joinKey}
	columns = append(columns, admitted.includeSourceColumns...)
	return columns
}

func requiredJoinLookupColumns(admitted admittedIntent) []string {
	columns := []string{admitted.joinKey}
	columns = append(columns, admitted.appendLookupColumns...)
	return columns
}

func missingWorkbookColumns(facts runtimeinspect.WorkbookFacts, sheetName string, required []string) []string {
	sheet, ok := factsSheet(facts, sheetName)
	if !ok {
		return append([]string(nil), required...)
	}

	available := make(map[string]struct{}, len(sheet.Columns))
	for _, column := range sheet.Columns {
		available[column] = struct{}{}
	}

	seen := make(map[string]struct{}, len(required))
	missing := make([]string, 0, len(required))
	for _, column := range required {
		if _, ok := seen[column]; ok {
			continue
		}
		seen[column] = struct{}{}
		if _, ok := available[column]; !ok {
			missing = append(missing, column)
		}
	}
	return missing
}

func factsSheet(facts runtimeinspect.WorkbookFacts, name string) (runtimeinspect.SheetFacts, bool) {
	for _, sheet := range facts.Sheets {
		if sheet.Name == name {
			return sheet, true
		}
	}
	return runtimeinspect.SheetFacts{}, false
}
