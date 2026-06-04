package useorchestrator

import (
	"strings"
)

type WorkbookFacts struct {
	VisibleSheetNames []string
	HeaderRowBySheet  map[string][]string
}

type JoinLookupRequest struct {
	ScenarioID           string
	RequestText          string
	InputWorkbookPath    string
	OutputWorkbookPath   string
	SourceSheet          string
	LookupSheet          string
	JoinKey              string
	TargetSheet          string
	IncludeSourceColumns []string
	AppendLookupColumns  []string
	PreserveOriginal     bool
}

func TryResolveJoinLookupRequest(
	document map[string]any,
	scenarioID string,
	inputFile string,
	outputFile string,
	requestText string,
	facts WorkbookFacts,
) (*JoinLookupRequest, bool) {
	rawDecision, ok := document["validated_execution_request"].(map[string]any)
	if !ok {
		rawDecision, ok = document["execution_request_candidate"].(map[string]any)
	}
	if !ok {
		return nil, false
	}

	operation, _ := rawDecision["operation"].(string)
	if operation == "" {
		operation, _ = rawDecision["requested_operation"].(string)
	}
	operationObject, _ := rawDecision["operation"].(map[string]any)
	if operation == "" {
		if operationType, ok := operationObject["type"].(string); ok && strings.TrimSpace(operationType) != "" {
			operation = operationType
		}
	}

	sourceSheet, _ := rawDecision["source_sheet"].(string)
	if sourceSheet == "" {
		sourceSheet, _ = rawDecision["base_sheet"].(string)
	}
	if sourceSheet == "" {
		sourceSheet, _ = operationObject["left_sheet"].(string)
	}
	lookupSheet, _ := rawDecision["lookup_sheet"].(string)
	if lookupSheet == "" {
		lookupSheet, _ = operationObject["right_sheet"].(string)
	}

	joinKey, _ := rawDecision["join_key"].(string)
	rawJoinKeyObject, _ := rawDecision["join_key"].(map[string]any)
	if joinKey == "" {
		joinKey, _ = operationObject["join_key"].(string)
	}
	if strings.TrimSpace(joinKey) == "" {
		if sourceColumn, ok := rawJoinKeyObject["source_column"].(string); ok && strings.TrimSpace(sourceColumn) != "" {
			joinKey = sourceColumn
		}
		if sourceColumn, ok := rawJoinKeyObject["base_column"].(string); strings.TrimSpace(joinKey) == "" && ok && strings.TrimSpace(sourceColumn) != "" {
			joinKey = sourceColumn
		}
	}

	resultSheet, _ := rawDecision["result_sheet"].(map[string]any)
	resultSheetIntent, _ := rawDecision["result_sheet_intent"].(map[string]any)
	resultRequirements, _ := rawDecision["result_requirements"].(map[string]any)

	targetSheet := "주문_원가매칭"
	if sheetName, ok := operationObject["result_sheet"].(string); ok && strings.TrimSpace(sheetName) != "" {
		targetSheet = sheetName
	}
	if sheetName, ok := operationObject["result_sheet_name"].(string); ok && strings.TrimSpace(sheetName) != "" {
		targetSheet = sheetName
	}
	if sheetName, ok := resultSheet["name"].(string); ok && strings.TrimSpace(sheetName) != "" {
		targetSheet = sheetName
	}
	if sheetName, ok := resultSheet["sheet_name"].(string); ok && strings.TrimSpace(sheetName) != "" {
		targetSheet = sheetName
	}
	if sheetName, ok := resultSheet["source_sheet"].(string); strings.TrimSpace(sourceSheet) == "" && ok && strings.TrimSpace(sheetName) != "" {
		sourceSheet = sheetName
	}
	if sheetName, ok := resultSheet["lookup_sheet"].(string); strings.TrimSpace(lookupSheet) == "" && ok && strings.TrimSpace(sheetName) != "" {
		lookupSheet = sheetName
	}

	if strings.TrimSpace(joinKey) == "" {
		if joinKeyObject, ok := resultSheet["join_key"].(map[string]any); ok {
			if sourceColumn, ok := joinKeyObject["source_column"].(string); ok && strings.TrimSpace(sourceColumn) != "" {
				joinKey = sourceColumn
			}
			if sourceColumn, ok := joinKeyObject["base_column"].(string); strings.TrimSpace(joinKey) == "" && ok && strings.TrimSpace(sourceColumn) != "" {
				joinKey = sourceColumn
			}
		}
	}
	if strings.TrimSpace(joinKey) == "" {
		if joinKeyObject, ok := operationObject["join_key"].(map[string]any); ok {
			if sourceColumn, ok := joinKeyObject["source_column"].(string); ok && strings.TrimSpace(sourceColumn) != "" {
				joinKey = sourceColumn
			}
			if sourceColumn, ok := joinKeyObject["base_column"].(string); strings.TrimSpace(joinKey) == "" && ok && strings.TrimSpace(sourceColumn) != "" {
				joinKey = sourceColumn
			}
		}
	}

	includeSourceColumns := joinLookupStringSlice(resultSheet["include_source_columns"])
	if len(includeSourceColumns) == 0 {
		includeSourceColumns = joinLookupStringSlice(resultSheet["include_columns"])
	}
	if len(includeSourceColumns) == 0 {
		includeSourceColumns = joinLookupStringSlice(operationObject["include_left_columns"])
	}
	if len(includeSourceColumns) == 0 {
		includeSourceColumns = joinLookupStringSlice(resultSheet["source_columns"])
	}
	if len(includeSourceColumns) == 0 {
		includeSourceColumns = joinLookupStringSlice(rawDecision["source_columns"])
	}
	if len(includeSourceColumns) == 0 {
		includeSourceColumns = joinLookupStringSlice(rawDecision["available_source_headers"])
	}

	appendLookupColumns := joinLookupStringSlice(resultSheet["append_lookup_columns"])
	if len(appendLookupColumns) == 0 {
		appendLookupColumns = joinLookupStringSlice(resultSheet["lookup_columns"])
	}
	if len(appendLookupColumns) == 0 {
		appendLookupColumns = joinLookupStringSlice(resultSheet["append_columns_from_lookup"])
	}
	if len(appendLookupColumns) == 0 {
		appendLookupColumns = joinLookupStringSlice(operationObject["add_columns_from_right"])
	}
	if len(appendLookupColumns) == 0 {
		appendLookupColumns = joinLookupStringSlice(operationObject["append_lookup_columns"])
	}
	if len(appendLookupColumns) == 0 {
		appendLookupColumns = joinLookupStringSlice(resultRequirements["append_lookup_columns"])
	}
	if len(appendLookupColumns) == 0 {
		appendLookupColumns = joinLookupStringSlice(rawDecision["appended_fields"])
	}
	if len(appendLookupColumns) == 0 {
		appendLookupColumns = joinLookupStringSlice(rawDecision["append_columns"])
	}
	if len(appendLookupColumns) == 0 {
		appendLookupColumns = joinLookupStringSlice(rawDecision["lookup_columns"])
	}
	if len(appendLookupColumns) == 0 {
		appendLookupColumns = joinLookupStringSlice(rawDecision["lookup_columns_to_append"])
	}
	if len(appendLookupColumns) == 0 {
		appendLookupColumns = joinLookupStringSlice(rawDecision["columns_to_append"])
	}

	naturalLanguageContext := strings.TrimSpace(operation + "\n" + requestText)
	sourceSheet, lookupSheet, joinKey, includeSourceColumns, appendLookupColumns = inferJoinLookupFromFacts(
		naturalLanguageContext,
		facts,
		sourceSheet,
		lookupSheet,
		joinKey,
		includeSourceColumns,
		appendLookupColumns,
	)
	if !looksLikeJoinLookupOperation(operation, sourceSheet, lookupSheet, joinKey, appendLookupColumns) {
		return nil, false
	}

	if len(includeSourceColumns) == 0 {
		includeSourceColumns = []string{"주문번호", "상품코드", "상품명", "수량", "판매가"}
	}
	appendLookupColumns = filterLookupAppendColumns(appendLookupColumns, includeSourceColumns, joinKey)
	if len(appendLookupColumns) == 0 {
		allColumns := joinLookupStringSlice(resultSheet["columns"])
		if len(allColumns) > len(includeSourceColumns) {
			appendLookupColumns = append([]string(nil), allColumns[len(includeSourceColumns):]...)
		}
	}

	if joinKey == "" {
		joinKey = "상품코드"
	}
	if lookupSheet == "" {
		lookupSheet, _ = resultRequirements["lookup_columns_from_sheet"].(string)
	}
	if lookupSheet == "" {
		lookupSheet = "상품마스터"
	}
	if sourceSheet == "" {
		sourceSheet = "주문내역"
	}
	if sheetName, ok := rawDecision["result_sheet_name"].(string); ok && strings.TrimSpace(sheetName) != "" {
		targetSheet = sheetName
	}
	if sheetName, ok := resultSheetIntent["sheet_name"].(string); strings.TrimSpace(targetSheet) == "" && ok && strings.TrimSpace(sheetName) != "" {
		targetSheet = sheetName
	}
	if sheetName, ok := resultSheetIntent["name"].(string); strings.TrimSpace(targetSheet) == "" && ok && strings.TrimSpace(sheetName) != "" {
		targetSheet = sheetName
	}

	return &JoinLookupRequest{
		ScenarioID:           scenarioID,
		RequestText:          requestText,
		InputWorkbookPath:    joinLookupCoalesceString(rawDecision["input_workbook_path"], inputFile),
		OutputWorkbookPath:   joinLookupCoalesceString(rawDecision["output_workbook_path"], outputFile),
		SourceSheet:          sourceSheet,
		LookupSheet:          lookupSheet,
		JoinKey:              joinKey,
		TargetSheet:          targetSheet,
		IncludeSourceColumns: includeSourceColumns,
		AppendLookupColumns:  appendLookupColumns,
		PreserveOriginal:     joinLookupCoalesceBool(rawDecision["preserve_original_sheets"], joinLookupCoalesceBool(rawDecision["preserve_original"], joinLookupCoalesceBool(operationObject["preserve_source_sheets"], true))),
	}, true
}

func RunJoinLookup(req JoinLookupRequest) (RunResult, error) {
	return OrchestrateValidated(req.ValidatedExecutionRequest())
}

func (req JoinLookupRequest) ValidatedExecutionRequest() ValidatedExecutionRequest {
	return ValidatedExecutionRequest{
		ScenarioID:           req.ScenarioID,
		RequestKind:          "prompt_text",
		RequestText:          req.RequestText,
		InputFile:            req.InputWorkbookPath,
		SourceSheet:          req.SourceSheet,
		OutputFile:           req.OutputWorkbookPath,
		ExecutionKind:        "composition",
		CompositionKind:      "join_lookup",
		TargetSheet:          req.TargetSheet,
		LookupSheet:          req.LookupSheet,
		JoinKey:              req.JoinKey,
		IncludeSourceColumns: append([]string(nil), req.IncludeSourceColumns...),
		AppendLookupColumns:  append([]string(nil), req.AppendLookupColumns...),
	}
}

func inferJoinLookupFromFacts(
	contextText string,
	facts WorkbookFacts,
	sourceSheet string,
	lookupSheet string,
	joinKey string,
	includeSourceColumns []string,
	appendLookupColumns []string,
) (string, string, string, []string, []string) {
	if len(facts.VisibleSheetNames) == 0 {
		return sourceSheet, lookupSheet, joinKey, includeSourceColumns, appendLookupColumns
	}

	mentioned := mentionedSheetNames(contextText, facts.VisibleSheetNames)
	if strings.TrimSpace(sourceSheet) == "" && len(mentioned) >= 1 {
		sourceSheet = mentioned[0]
	}
	if strings.TrimSpace(lookupSheet) == "" {
		if len(mentioned) >= 2 {
			lookupSheet = mentioned[1]
		} else if len(mentioned) == 1 && len(facts.VisibleSheetNames) == 2 {
			for _, candidate := range facts.VisibleSheetNames {
				if candidate != mentioned[0] {
					lookupSheet = candidate
				}
			}
		}
	}

	if strings.TrimSpace(joinKey) == "" && sourceSheet != "" && lookupSheet != "" {
		sharedHeaders := intersectHeaders(facts.HeaderRowBySheet[sourceSheet], facts.HeaderRowBySheet[lookupSheet])
		for _, header := range sharedHeaders {
			if strings.Contains(contextText, header) {
				joinKey = header
				break
			}
		}
		if joinKey == "" && len(sharedHeaders) == 1 {
			joinKey = sharedHeaders[0]
		}
	}

	if len(includeSourceColumns) == 0 && sourceSheet != "" {
		includeSourceColumns = append([]string(nil), facts.HeaderRowBySheet[sourceSheet]...)
	}

	if len(appendLookupColumns) == 0 && sourceSheet != "" && lookupSheet != "" {
		sourceHeaders := facts.HeaderRowBySheet[sourceSheet]
		lookupHeaders := facts.HeaderRowBySheet[lookupSheet]
		for _, header := range lookupHeaders {
			if strings.TrimSpace(header) == "" || header == joinKey || containsString(sourceHeaders, header) {
				continue
			}
			if strings.Contains(contextText, header) {
				appendLookupColumns = append(appendLookupColumns, header)
			}
		}
		if len(appendLookupColumns) == 0 {
			for _, header := range lookupHeaders {
				if strings.TrimSpace(header) == "" || header == joinKey || containsString(sourceHeaders, header) {
					continue
				}
				appendLookupColumns = append(appendLookupColumns, header)
			}
		}
	}

	return sourceSheet, lookupSheet, joinKey, includeSourceColumns, appendLookupColumns
}

func looksLikeJoinLookupOperation(operation string, sourceSheet string, lookupSheet string, joinKey string, appendLookupColumns []string) bool {
	switch operation {
	case "create_join_lookup_result_sheet",
		"join_lookup",
		"join_lookup_create_result_sheet",
		"join_lookup_append_fields",
		"create_new_result_sheet_with_lookup",
		"left_join_lookup":
		return true
	}
	if strings.TrimSpace(sourceSheet) == "" || strings.TrimSpace(lookupSheet) == "" || strings.TrimSpace(joinKey) == "" {
		return false
	}
	if len(appendLookupColumns) == 0 {
		return false
	}

	normalizedOperation := strings.ToLower(strings.TrimSpace(operation))
	if normalizedOperation == "" {
		return true
	}
	return strings.Contains(normalizedOperation, "lookup") ||
		strings.Contains(normalizedOperation, "result sheet") ||
		strings.Contains(normalizedOperation, "match rows") ||
		strings.Contains(normalizedOperation, "join")
}

func joinLookupStringSlice(value any) []string {
	rawValues, ok := value.([]any)
	if !ok {
		return nil
	}
	values := make([]string, 0, len(rawValues))
	for _, rawValue := range rawValues {
		if text, ok := rawValue.(string); ok && strings.TrimSpace(text) != "" {
			values = append(values, text)
		}
	}
	return values
}

func filterLookupAppendColumns(columns []string, sourceColumns []string, joinKey string) []string {
	if len(columns) == 0 {
		return nil
	}
	excluded := make(map[string]struct{}, len(sourceColumns)+1)
	for _, column := range sourceColumns {
		if trimmed := strings.TrimSpace(column); trimmed != "" {
			excluded[trimmed] = struct{}{}
		}
	}
	if trimmed := strings.TrimSpace(joinKey); trimmed != "" {
		excluded[trimmed] = struct{}{}
	}
	filtered := make([]string, 0, len(columns))
	seen := make(map[string]struct{}, len(columns))
	for _, column := range columns {
		trimmed := strings.TrimSpace(column)
		if trimmed == "" {
			continue
		}
		if _, skip := excluded[trimmed]; skip {
			continue
		}
		if _, duplicate := seen[trimmed]; duplicate {
			continue
		}
		seen[trimmed] = struct{}{}
		filtered = append(filtered, trimmed)
	}
	return filtered
}

func joinLookupCoalesceString(value any, fallback string) string {
	if text, ok := value.(string); ok && strings.TrimSpace(text) != "" {
		return text
	}
	return fallback
}

func joinLookupCoalesceBool(value any, fallback bool) bool {
	if typed, ok := value.(bool); ok {
		return typed
	}
	return fallback
}

func mentionedSheetNames(contextText string, sheetNames []string) []string {
	matched := make([]string, 0, len(sheetNames))
	for _, sheetName := range sheetNames {
		if strings.TrimSpace(sheetName) == "" {
			continue
		}
		if strings.Contains(contextText, sheetName) {
			matched = append(matched, sheetName)
		}
	}
	return matched
}

func intersectHeaders(left []string, right []string) []string {
	rightSet := make(map[string]struct{}, len(right))
	for _, value := range right {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			rightSet[trimmed] = struct{}{}
		}
	}
	shared := make([]string, 0, len(left))
	for _, value := range left {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		if _, ok := rightSet[trimmed]; ok {
			shared = append(shared, trimmed)
		}
	}
	return shared
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if strings.TrimSpace(value) == strings.TrimSpace(target) {
			return true
		}
	}
	return false
}
