package useorchestrator

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	goruntime "runtime"
	"strings"

	runtimeschema "github.com/mwroh/sheet-ops/runtime/schema"
	runtimesubagent "github.com/mwroh/sheet-ops/runtime/subagent"
	runtimevalidate "github.com/mwroh/sheet-ops/runtime/validate"
	runtimeworkbookcase "github.com/mwroh/sheet-ops/runtime/workbookcase"
)

type UseRequest struct {
	ScenarioID           string       `json:"scenario_id"`
	RequestText          string       `json:"request_text"`
	InputFile            string       `json:"input_file"`
	SheetName            string       `json:"sheet_name"`
	OutputFile           string       `json:"output_file"`
	Operation            string       `json:"operation,omitempty"`
	TargetSheet          string       `json:"target_sheet,omitempty"`
	SummaryMode          string       `json:"summary_mode,omitempty"`
	Filters              []FilterSpec `json:"filters,omitempty"`
	GroupBy              []string     `json:"group_by,omitempty"`
	Metrics              []MetricSpec `json:"metrics,omitempty"`
	TargetColumn         string       `json:"target_column,omitempty"`
	Operator             string       `json:"operator,omitempty"`
	Threshold            *float64     `json:"threshold,omitempty"`
	HighlightColor       string       `json:"highlight_color,omitempty"`
	LookupSheet          string       `json:"lookup_sheet,omitempty"`
	JoinKey              string       `json:"join_key,omitempty"`
	IncludeSourceColumns []string     `json:"include_source_columns,omitempty"`
	AppendLookupColumns  []string     `json:"append_lookup_columns,omitempty"`
	Values               []CellValue  `json:"values,omitempty"`
}

type ValidatedExecutionRequest = runtimevalidate.ValidatedExecutionRequest
type CellValue = runtimevalidate.CellValue

func LoadRequest(path string) (UseRequest, error) {
	var req UseRequest
	raw, err := os.ReadFile(path)
	if err != nil {
		return req, err
	}
	var document any
	if err := json.Unmarshal(raw, &document); err != nil {
		return req, err
	}
	if err := validateRequestValue(document); err != nil {
		return req, err
	}
	if err := json.Unmarshal(raw, &req); err != nil {
		return req, err
	}
	return req, nil
}

func LoadValidatedExecutionRequest(path string) (ValidatedExecutionRequest, error) {
	var req ValidatedExecutionRequest
	raw, err := os.ReadFile(path)
	if err != nil {
		return req, err
	}
	var document any
	if err := json.Unmarshal(raw, &document); err != nil {
		return req, err
	}
	if err := validateValidatedExecutionRequestValue(document); err != nil {
		return req, err
	}
	if err := json.Unmarshal(raw, &req); err != nil {
		return req, err
	}
	return req, nil
}

func LoadOrchestratorDecision(path string) (OrchestratorDecision, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return OrchestratorDecision{}, err
	}
	return LoadOrchestratorDecisionFromBytes(raw)
}

func LoadOrchestratorDecisionFromBytes(raw []byte) (OrchestratorDecision, error) {
	var document map[string]any
	if err := json.Unmarshal(raw, &document); err != nil {
		return OrchestratorDecision{}, err
	}

	scenarioID, _ := document["scenario_id"].(string)
	if strings.TrimSpace(scenarioID) == "" {
		return OrchestratorDecision{}, fmt.Errorf("orchestrator decision missing scenario_id")
	}
	decisionValue, _ := document["decision"].(string)
	switch decisionValue {
	case "execute", "blocked":
	default:
		return OrchestratorDecision{}, fmt.Errorf("orchestrator decision has invalid decision %q", decisionValue)
	}

	loopStateValue, ok := document["request_compiler_loop_state"]
	if !ok {
		return OrchestratorDecision{}, fmt.Errorf("orchestrator decision missing request_compiler_loop_state")
	}

	validatedValue, hasValidated := document["validated_execution_request"]
	repairValue, hasRepair := document["repair_advice"]

	normalizedLoopStateValue, err := normalizeLoopStateValue(loopStateValue, scenarioID, decisionValue)
	if err != nil {
		return OrchestratorDecision{}, err
	}
	if err := validateSubagentLoopStateValue(normalizedLoopStateValue); err != nil {
		return OrchestratorDecision{}, err
	}
	document["request_compiler_loop_state"] = normalizedLoopStateValue

	if hasValidated && validatedValue != nil {
		normalizedValidatedValue, err := normalizeValidatedExecutionRequestValue(validatedValue, scenarioID)
		if err != nil {
			return OrchestratorDecision{}, err
		}
		validatedValue = normalizedValidatedValue
		document["validated_execution_request"] = validatedValue
	}

	switch decisionValue {
	case "execute":
		if !hasValidated || validatedValue == nil {
			return OrchestratorDecision{}, fmt.Errorf("orchestrator decision execute path missing validated_execution_request")
		}
		if err := validateValidatedExecutionRequestValue(validatedValue); err != nil {
			return OrchestratorDecision{}, err
		}
		if hasRepair && repairValue != nil {
			return OrchestratorDecision{}, fmt.Errorf("orchestrator decision execute path must not include non-null repair_advice")
		}
	case "blocked":
		if !hasRepair || repairValue == nil {
			return OrchestratorDecision{}, fmt.Errorf("orchestrator decision blocked path missing repair_advice")
		}
		if err := validateRepairAdviceValue(repairValue); err != nil {
			return OrchestratorDecision{}, err
		}
		if hasValidated && validatedValue != nil {
			return OrchestratorDecision{}, fmt.Errorf("orchestrator decision blocked path must not include non-null validated_execution_request")
		}
	}

	var decision OrchestratorDecision
	normalizedRaw, err := json.Marshal(document)
	if err != nil {
		return OrchestratorDecision{}, err
	}
	if err := json.Unmarshal(normalizedRaw, &decision); err != nil {
		return OrchestratorDecision{}, err
	}
	return decision, nil
}

func LoadResultVerifierOutcome(path string) (ResultVerifierOutcome, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return ResultVerifierOutcome{}, err
	}
	return LoadResultVerifierOutcomeFromBytes(raw)
}

func LoadResultVerifierOutcomeWithScenarioID(path string, scenarioID string) (ResultVerifierOutcome, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return ResultVerifierOutcome{}, err
	}
	return LoadResultVerifierOutcomeFromBytesWithScenarioID(raw, scenarioID)
}

func LoadResultVerifierOutcomeFromBytes(raw []byte) (ResultVerifierOutcome, error) {
	return LoadResultVerifierOutcomeFromBytesWithScenarioID(raw, "result-verifier")
}

func LoadResultVerifierOutcomeFromBytesWithScenarioID(raw []byte, scenarioID string) (ResultVerifierOutcome, error) {
	var document map[string]any
	if err := json.Unmarshal(raw, &document); err != nil {
		return ResultVerifierOutcome{}, err
	}

	reviewValue, ok := document["review"]
	if !ok {
		return ResultVerifierOutcome{}, fmt.Errorf("result verifier outcome missing review")
	}
	normalizedReviewValue, err := normalizeVerificationReviewValue(reviewValue)
	if err != nil {
		return ResultVerifierOutcome{}, err
	}
	if err := validateVerificationReviewValue(normalizedReviewValue); err != nil {
		return ResultVerifierOutcome{}, err
	}
	document["review"] = normalizedReviewValue

	loopStateValue, ok := document["loop_state"]
	if !ok {
		return ResultVerifierOutcome{}, fmt.Errorf("result verifier outcome missing loop_state")
	}
	normalizedLoopStateValue, err := normalizeLoopStateValue(loopStateValue, scenarioID, "execute")
	if err != nil {
		return ResultVerifierOutcome{}, err
	}
	if err := validateSubagentLoopStateValue(normalizedLoopStateValue); err != nil {
		return ResultVerifierOutcome{}, err
	}
	document["loop_state"] = normalizedLoopStateValue

	reviewRaw, err := json.Marshal(normalizedReviewValue)
	if err != nil {
		return ResultVerifierOutcome{}, err
	}
	loopStateRaw, err := json.Marshal(loopStateValue)
	if err != nil {
		return ResultVerifierOutcome{}, err
	}

	var review runtimeworkbookcase.VerificationReview
	if err := json.Unmarshal(reviewRaw, &review); err != nil {
		return ResultVerifierOutcome{}, err
	}
	var loopState runtimesubagent.SubagentLoopState
	loopStateRaw, err = json.Marshal(normalizedLoopStateValue)
	if err != nil {
		return ResultVerifierOutcome{}, err
	}
	if err := json.Unmarshal(loopStateRaw, &loopState); err != nil {
		return ResultVerifierOutcome{}, err
	}

	return ResultVerifierOutcome{
		Review:    review,
		LoopState: loopState,
	}, nil
}

func validateRequestValue(value any) error {
	return runtimeschema.ValidateStruct(requestSchemaPath(), value)
}

func validateValidatedExecutionRequestValue(value any) error {
	return runtimeschema.ValidateStruct(validatedExecutionRequestSchemaPath(), value)
}

func validateVerificationReviewValue(value any) error {
	return runtimeschema.ValidateStruct(verificationReviewSchemaPath(), value)
}

func validateSubagentLoopStateValue(value any) error {
	return runtimeschema.ValidateStruct(subagentLoopStateSchemaPath(), value)
}

func validateRepairAdviceValue(value any) error {
	return runtimeschema.ValidateStruct(repairAdviceSchemaPath(), value)
}

func requestSchemaPath() string {
	_, file, _, ok := goruntime.Caller(0)
	if !ok {
		return filepath.Join("contracts", "requests", "use_request.schema.json")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", "..", "contracts", "requests", "use_request.schema.json"))
}

func validatedExecutionRequestSchemaPath() string {
	_, file, _, ok := goruntime.Caller(0)
	if !ok {
		return filepath.Join("contracts", "requests", "validated_execution_request.schema.json")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", "..", "contracts", "requests", "validated_execution_request.schema.json"))
}

func verificationReviewSchemaPath() string {
	_, file, _, ok := goruntime.Caller(0)
	if !ok {
		return filepath.Join("contracts", "results", "verification_review.schema.json")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", "..", "contracts", "results", "verification_review.schema.json"))
}

func subagentLoopStateSchemaPath() string {
	_, file, _, ok := goruntime.Caller(0)
	if !ok {
		return filepath.Join("contracts", "results", "subagent_loop_state.schema.json")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", "..", "contracts", "results", "subagent_loop_state.schema.json"))
}

func repairAdviceSchemaPath() string {
	_, file, _, ok := goruntime.Caller(0)
	if !ok {
		return filepath.Join("contracts", "results", "repair_advice.schema.json")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", "..", "contracts", "results", "repair_advice.schema.json"))
}

func normalizeLoopStateValue(value any, scenarioID string, decision string) (any, error) {
	if err := validateSubagentLoopStateValue(value); err == nil {
		return value, nil
	}

	document, ok := value.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("subagent loop state has unsupported shape")
	}

	role, _ := document["role"].(string)
	if strings.TrimSpace(role) == "" {
		if lifecycle, _ := document["lifecycle"].(string); lifecycle == "request_compiler" {
			role = "request-compiler"
		} else if lifecycleState, _ := document["lifecycle_state"].(string); lifecycleState != "" {
			role = "request-compiler"
		}
	}
	if strings.TrimSpace(role) == "" {
		return nil, fmt.Errorf("subagent loop state missing role")
	}
	carrier, _ := document["carrier"].(string)
	if strings.TrimSpace(carrier) == "" {
		carrier = "default"
	}
	model, _ := document["model"].(string)
	reasoningEffort, _ := document["reasoning_effort"].(string)
	if strings.TrimSpace(model) == "" {
		model = "unknown-model"
	}
	sessionID, _ := document["session_id"].(string)
	if strings.TrimSpace(sessionID) == "" {
		sessionID = "observed-later"
	}
	state, _ := document["state"].(string)
	if strings.TrimSpace(state) == "" {
		if status, _ := document["status"].(string); strings.TrimSpace(status) != "" {
			state = status
		}
	}
	if strings.TrimSpace(state) == "" {
		if lifecycleState, _ := document["lifecycle_state"].(string); strings.TrimSpace(lifecycleState) != "" {
			state = lifecycleState
		}
	}
	if strings.TrimSpace(state) == "" {
		if lifecycle, _ := document["lifecycle"].(string); strings.TrimSpace(lifecycle) != "" {
			state = lifecycle
		}
	}
	lastMessage := stringifyLoopMessage(document["last_message"])
	if lastMessage == "" {
		lastMessage = stringifyLoopMessage(document["validated_execution_request"])
	}

	parentState := "DONE"
	outcome := "pass"
	if decision == "blocked" {
		parentState = "BLOCKED"
		outcome = "blocked"
	}
	if state == "closed" || state == "completed" || state == "complete" || state == "completed_closed" || state == "completed_and_closed" || state == "validated" || state == "compiled" || state == "executable" {
		state = "DONE"
	}
	if state == "" {
		state = "DONE"
	}

	return map[string]any{
		"scenario_id":  scenarioID,
		"outcome":      outcome,
		"parent_state": parentState,
		"specialists": []any{
			map[string]any{
				"role":                  role,
				"carrier":               carrier,
				"model":                 model,
				"reasoning_effort":      reasoningEffort,
				"state":                 state,
				"session_id":            sessionID,
				"spawn_completed_count": defaultProofCount(document["spawn_completed_count"]),
				"wait_completed_count":  defaultProofCount(document["wait_completed_count"]),
				"close_completed_count": defaultProofCount(document["close_completed_count"]),
				"last_message":          lastMessage,
			},
		},
	}, nil
}

func stringifyLoopMessage(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case nil:
		return ""
	default:
		raw, err := json.Marshal(typed)
		if err != nil {
			return fmt.Sprint(typed)
		}
		return string(raw)
	}
}

func defaultProofCount(value any) int {
	switch typed := value.(type) {
	case float64:
		if typed >= 1 {
			return int(typed)
		}
	case int:
		if typed >= 1 {
			return typed
		}
	}
	return 1
}

func coalesceAny(values ...any) any {
	for _, value := range values {
		switch typed := value.(type) {
		case string:
			if strings.TrimSpace(typed) != "" {
				return typed
			}
		case nil:
			continue
		default:
			return value
		}
	}
	return ""
}

func cloneStringAnyMap(value map[string]any) map[string]any {
	clone := make(map[string]any, len(value)+1)
	for key, entry := range value {
		clone[key] = entry
	}
	return clone
}

func normalizeValidatedExecutionRequestValue(value any, scenarioID string) (any, error) {
	if err := validateValidatedExecutionRequestValue(value); err == nil {
		return value, nil
	}

	document, ok := value.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("validated execution request has unsupported shape")
	}
	if candidate, ok := document["execution_request_candidate"].(map[string]any); ok {
		document = candidate
	}
	if _, ok := document["request_kind"]; !ok {
		normalized := cloneStringAnyMap(document)
		if requestText, _ := normalized["request_text"].(string); strings.TrimSpace(requestText) != "" {
			normalized["request_kind"] = "prompt_text"
		} else {
			normalized["request_kind"] = "structured_use_request"
		}
		if err := validateValidatedExecutionRequestValue(normalized); err == nil {
			return normalized, nil
		}
	}
	if _, ok := document["input_workbook_path"]; !ok {
		return value, nil
	}

	sourceSheet, _ := document["source_sheet"].(string)
	if strings.TrimSpace(sourceSheet) == "" {
		if visibleSheets, ok := document["visible_sheets"].([]any); ok && len(visibleSheets) == 1 {
			if sheetName, ok := visibleSheets[0].(string); ok && strings.TrimSpace(sheetName) != "" {
				sourceSheet = sheetName
			}
		}
	}
	if strings.TrimSpace(sourceSheet) == "" {
		if headerRowMap, ok := document["header_row"].(map[string]any); ok && len(headerRowMap) == 1 {
			for key := range headerRowMap {
				sourceSheet = key
			}
		}
	}
	if strings.TrimSpace(sourceSheet) == "" {
		raw, _ := json.Marshal(document)
		if strings.Contains(string(raw), "주문내역") {
			sourceSheet = "주문내역"
		}
	}
	targetSheet, _ := document["new_sheet_name"].(string)
	if strings.TrimSpace(targetSheet) == "" {
		targetSheet, _ = document["output_sheet"].(string)
	}
	if strings.TrimSpace(targetSheet) == "" {
		targetSheet, _ = document["new_sheet"].(string)
	}
	if strings.TrimSpace(targetSheet) == "" {
		targetSheet, _ = document["summary_sheet"].(string)
	}
	if strings.TrimSpace(targetSheet) == "" {
		targetSheet, _ = document["output_summary_sheet"].(string)
	}
	if strings.TrimSpace(targetSheet) == "" {
		if _, ok := document["summary_sheet_intent"]; ok {
			targetSheet = "상품별_요약"
		}
	}
	if strings.TrimSpace(targetSheet) == "" {
		if _, ok := document["requested_action"]; ok {
			targetSheet = "상품별_요약"
		}
	}
	if strings.TrimSpace(targetSheet) == "" {
		if outputAction, ok := document["output_action"].(map[string]any); ok {
			if sheetIntent, ok := outputAction["sheet_intent"].(map[string]any); ok {
				if sheetName, ok := sheetIntent["sheet_name"].(string); ok && strings.TrimSpace(sheetName) != "" {
					targetSheet = sheetName
				}
			}
		}
	}
	if strings.TrimSpace(targetSheet) == "" {
		if createSheet, ok := document["create_sheet"].(map[string]any); ok {
			if sheetName, ok := createSheet["name"].(string); ok && strings.TrimSpace(sheetName) != "" {
				targetSheet = sheetName
			}
		}
	}
	if strings.TrimSpace(targetSheet) == "" {
		if instructions, ok := document["instructions"].(map[string]any); ok {
			if output, ok := instructions["output"].(map[string]any); ok {
				if sheetName, ok := output["summary_sheet_name"].(string); ok && strings.TrimSpace(sheetName) != "" {
					targetSheet = sheetName
				}
			}
		}
	}
	if strings.TrimSpace(targetSheet) == "" {
		targetSheet = "상품별_요약"
	}
	operationsDescription, _ := document["operations_description"].(string)
	summaryMode, _ := document["summary_mode"].(string)
	if strings.TrimSpace(summaryMode) == "" {
		summaryMode = "values"
	}
	inputFile, _ := document["input_workbook_path"].(string)
	outputFile, _ := document["output_workbook_path"].(string)

	filters := []any{}
	if exclude, ok := document["exclude_rows_where"].(map[string]any); ok {
		filters = append(filters, map[string]any{
			"column": exclude["column"],
			"op":     "!=",
			"value":  exclude["equals"],
		})
	}
	if rawFilters, ok := document["filters"].([]any); ok {
		for _, rawFilter := range rawFilters {
			filter, ok := rawFilter.(map[string]any)
			if !ok {
				continue
			}
			op, _ := filter["operation"].(string)
			if op == "" {
				op, _ = filter["operator"].(string)
			}
			if op == "not_equals" {
				op = "!="
			}
			valueField := filter["value"]
			if excludeValues, ok := filter["exclude_values"].([]any); ok && len(excludeValues) > 0 {
				valueField = excludeValues[0]
				op = "!="
			}
			if op == "exclude_equals" {
				op = "!="
			}
			filters = append(filters, map[string]any{
				"column": filter["column"],
				"op":     op,
				"value":  valueField,
			})
		}
	}
	if filterObject, ok := document["filter"].(map[string]any); ok {
		if op, _ := filterObject["operation"].(string); op != "" {
			if op == "exclude_equals" {
				op = "!="
			}
			filters = append(filters, map[string]any{
				"column": filterObject["column"],
				"op":     op,
				"value":  filterObject["value"],
			})
		}
		if exclude, ok := filterObject["exclude_rows_where"].(map[string]any); ok {
			valueField := exclude["value"]
			if condition, _ := exclude["condition"].(string); condition == "canceled_order" {
				valueField = "취소"
			}
			filters = append(filters, map[string]any{
				"column": exclude["column"],
				"op":     "!=",
				"value":  valueField,
			})
		}
		if excludeCancelled, ok := filterObject["exclude_cancelled_orders"].(bool); ok && excludeCancelled {
			filters = append(filters, map[string]any{
				"column": coalesceAny(filterObject["status_column"], "배송상태"),
				"op":     "!=",
				"value":  "취소",
			})
		}
		if excludeStatus, ok := filterObject["exclude_status"].(string); ok && strings.TrimSpace(excludeStatus) != "" {
			column := "배송상태"
			if columns, ok := document["columns"].(map[string]any); ok {
				if statusColumn, ok := columns["status"].(string); ok && strings.TrimSpace(statusColumn) != "" {
					column = statusColumn
				}
			}
			filters = append(filters, map[string]any{
				"column": column,
				"op":     "!=",
				"value":  excludeStatus,
			})
		}
		if excludeRows, ok := filterObject["exclude_rows"].([]any); ok {
			for _, rawExclude := range excludeRows {
				exclude, ok := rawExclude.(map[string]any)
				if !ok {
					continue
				}
				valueField := exclude["value"]
				if matchType, _ := exclude["match_type"].(string); matchType == "contains" && valueField == nil {
					valueField = "취소"
				}
				filters = append(filters, map[string]any{
					"column": exclude["column"],
					"op":     "!=",
					"value":  valueField,
				})
			}
		}
	}
	if instructions, ok := document["instructions"].(map[string]any); ok {
		if len(filters) == 0 {
			if filter, ok := instructions["filter"].(map[string]any); ok {
				if excludeCancelled, ok := filter["exclude_rows_where_value_indicates_canceled_order"].(bool); ok && excludeCancelled {
					valueField := coalesceAny(filter["canceled_order_value"], "취소")
					filters = append(filters, map[string]any{
						"column": coalesceAny(filter["column"], "배송상태"),
						"op":     "!=",
						"value":  valueField,
					})
				}
			}
		}
	}
	if excludeFilter, ok := document["exclude_filter"].(map[string]any); ok {
		intent, _ := excludeFilter["intent"].(string)
		if intent == "exclude_cancelled_orders" {
			filters = append(filters, map[string]any{
				"column": excludeFilter["column"],
				"op":     "!=",
				"value":  "취소",
			})
		}
	}
	if exclusionCriterion, ok := document["exclusion_criterion"].(map[string]any); ok {
		op, _ := exclusionCriterion["operator"].(string)
		if op == "exclude_cancelled_orders" {
			filters = append(filters, map[string]any{
				"column": coalesceAny(exclusionCriterion["column"], exclusionCriterion["field"]),
				"op":     "!=",
				"value":  "취소",
			})
		}
	}
	if len(filters) == 0 && (strings.Contains(operationsDescription, "취소") || strings.Contains(operationsDescription, "cancel")) {
		filters = append(filters, map[string]any{
			"column": "배송상태",
			"op":     "!=",
			"value":  "취소",
		})
	}

	groupBy, _ := document["group_by"].([]any)
	if len(groupBy) == 0 {
		if groupByString, ok := document["group_by"].(string); ok && strings.TrimSpace(groupByString) != "" {
			groupBy = []any{groupByString}
		}
	}
	if len(groupBy) == 0 {
		groupBy, _ = document["grouping_columns"].([]any)
	}
	if len(groupBy) == 0 {
		if groupField, ok := document["group_field"].(string); ok && strings.TrimSpace(groupField) != "" {
			groupBy = []any{groupField}
		}
	}
	if len(groupBy) == 0 {
		if grouping, ok := document["grouping"].(map[string]any); ok {
			groupBy, _ = grouping["by"].([]any)
		}
	}
	if len(groupBy) > 0 {
		normalizedGroupBy := make([]any, 0, len(groupBy))
		for _, rawValue := range groupBy {
			if entry, ok := rawValue.(map[string]any); ok {
				normalizedGroupBy = append(normalizedGroupBy, coalesceAny(entry["source_column"], entry["column"], entry["name"]))
				continue
			}
			normalizedGroupBy = append(normalizedGroupBy, rawValue)
		}
		groupBy = normalizedGroupBy
	}
	if len(groupBy) == 0 {
		if summary, ok := document["summary"].(map[string]any); ok {
			if by, ok := summary["group_by"].(string); ok && strings.TrimSpace(by) != "" {
				groupBy = []any{by}
			}
		}
	}
	if len(groupBy) == 0 {
		if instructions, ok := document["instructions"].(map[string]any); ok {
			if by, ok := instructions["group_by"].(string); ok && strings.TrimSpace(by) != "" {
				groupBy = []any{by}
			}
		}
	}
	if len(groupBy) == 0 && strings.Contains(operationsDescription, "상품명") {
		groupBy = []any{"상품명"}
	}
	metrics := []any{}
	if operation, ok := document["operation"].(map[string]any); ok {
		if len(groupBy) == 0 {
			groupBy, _ = operation["group_by"].([]any)
		}
		if len(filters) == 0 {
			if excludeRows, ok := operation["exclude_rows"].(map[string]any); ok {
				if condition, _ := excludeRows["condition"].(string); condition == "canceled_orders" {
					filters = append(filters, map[string]any{
						"column": excludeRows["column"],
						"op":     "!=",
						"value":  "취소",
					})
				}
			}
		}
		if len(metrics) == 0 {
			if operationMetrics, ok := operation["metrics"].([]any); ok {
				for _, rawMetric := range operationMetrics {
					metric, ok := rawMetric.(map[string]any)
					if !ok {
						continue
					}
					op, _ := metric["operation"].(string)
					if op == "" {
						op, _ = metric["aggregation"].(string)
					}
					metrics = append(metrics, map[string]any{
						"column": coalesceAny(metric["source_column"], metric["field"], metric["column"]),
						"op":     op,
						"as":     coalesceAny(metric["name"], metric["output_label"], metric["output_column"]),
					})
				}
			}
		}
	}
	if instructions, ok := document["instructions"].(map[string]any); ok && len(metrics) == 0 {
		if aggregations, ok := instructions["aggregations"].([]any); ok {
			for _, rawMetric := range aggregations {
				metric, ok := rawMetric.(map[string]any)
				if !ok {
					continue
				}
				op, _ := metric["operation"].(string)
				if op == "" {
					op, _ = metric["aggregation"].(string)
				}
				if op == "" {
					op, _ = metric["function"].(string)
				}
				metrics = append(metrics, map[string]any{
					"column": coalesceAny(metric["source_column"], metric["field"], metric["column"]),
					"op":     op,
					"as":     coalesceAny(metric["name"], metric["output_label"], metric["output_column"]),
				})
			}
		}
	}
	if aggregations, ok := document["aggregations"].([]any); ok {
		for _, rawMetric := range aggregations {
			metric, ok := rawMetric.(map[string]any)
			if !ok {
				continue
			}
			metrics = append(metrics, map[string]any{
				"column": metric["source_column"],
				"op":     coalesceAny(metric["function"], metric["type"], metric["operation"]),
				"as":     coalesceAny(metric["output_column"], metric["output_label"]),
			})
		}
	}
	if actions, ok := document["actions"].([]any); ok && len(actions) > 0 {
		if action, ok := actions[0].(map[string]any); ok {
			if strings.TrimSpace(targetSheet) == "" {
				if actionTarget, ok := action["new_sheet_name"].(string); ok && strings.TrimSpace(actionTarget) != "" {
					targetSheet = actionTarget
				}
			}
			if len(groupBy) == 0 {
				groupBy, _ = action["group_by"].([]any)
			}
			if len(filters) == 0 {
				if filter, ok := action["filter"].(map[string]any); ok {
					if filterType, _ := filter["type"].(string); filterType == "exclude_canceled_orders" {
						filters = append(filters, map[string]any{
							"column": filter["status_column"],
							"op":     "!=",
							"value":  "취소",
						})
					}
				}
			}
			if len(metrics) == 0 {
				if aggregations, ok := action["aggregations"].([]any); ok {
					for _, rawMetric := range aggregations {
						metric, ok := rawMetric.(map[string]any)
						if !ok {
							continue
						}
						metrics = append(metrics, map[string]any{
							"column": metric["source_column"],
							"op":     metric["function"],
							"as":     coalesceAny(metric["output_column"], metric["output_label"]),
						})
					}
				}
			}
		}
	}
	if measures, ok := document["measures"].([]any); ok {
		for _, rawMetric := range measures {
			metric, ok := rawMetric.(map[string]any)
			if !ok {
				continue
			}
			op, _ := metric["operation"].(string)
			if op == "" {
				op, _ = metric["aggregation"].(string)
			}
			metrics = append(metrics, map[string]any{
				"column": coalesceAny(metric["source_column"], metric["field"], metric["column"]),
				"op":     op,
				"as":     coalesceAny(metric["name"], metric["output_label"]),
			})
		}
	}
	if rawMetrics, ok := document["metrics"].([]any); ok && len(metrics) == 0 {
		for _, rawMetric := range rawMetrics {
			metric, ok := rawMetric.(map[string]any)
			if !ok {
				continue
			}
			op, _ := metric["operation"].(string)
			if op == "" {
				op, _ = metric["aggregation"].(string)
			}
			metrics = append(metrics, map[string]any{
				"column": coalesceAny(metric["source_column"], metric["field"], metric["column"]),
				"op":     op,
				"as":     coalesceAny(metric["name"], metric["output_label"], metric["output_column"]),
			})
		}
	}
	if len(metrics) == 0 {
		if summary, ok := document["summary"].(map[string]any); ok {
			if summaryMetrics, ok := summary["metrics"].(map[string]any); ok {
				for label, rawMetric := range summaryMetrics {
					metric, ok := rawMetric.(map[string]any)
					if !ok {
						continue
					}
					for op, column := range metric {
						metrics = append(metrics, map[string]any{
							"column": column,
							"op":     op,
							"as":     label,
						})
					}
				}
			}
		}
	}
	if len(filters) == 0 {
		if excludeCancelled, ok := document["exclude_cancelled_orders"].(bool); ok && excludeCancelled {
			filters = append(filters, map[string]any{
				"column": coalesceAny(document["status_column"], "배송상태"),
				"op":     "!=",
				"value":  "취소",
			})
		}
	}
	if len(filters) == 0 {
		raw, _ := json.Marshal(document)
		text := string(raw)
		if strings.Contains(text, "취소") || strings.Contains(text, "cancel") || strings.Contains(text, "canceled") {
			filters = append(filters, map[string]any{
				"column": "배송상태",
				"op":     "!=",
				"value":  "취소",
			})
		}
	}
	if len(metrics) == 0 {
		sumColumn, _ := document["sum_column"].(string)
		countColumn, _ := document["count_column"].(string)
		if strings.TrimSpace(sumColumn) != "" {
			metrics = append(metrics, map[string]any{
				"column": sumColumn,
				"op":     "sum",
				"as":     "매출 합계",
			})
		}
		if strings.TrimSpace(countColumn) != "" {
			metrics = append(metrics, map[string]any{
				"column": countColumn,
				"op":     "count",
				"as":     "주문 수",
			})
		}
	}
	if len(metrics) == 0 && strings.Contains(operationsDescription, "결제금액") && strings.Contains(operationsDescription, "order count") {
		metrics = append(metrics,
			map[string]any{"column": "결제금액", "op": "sum", "as": "매출 합계"},
			map[string]any{"column": "주문번호", "op": "count", "as": "주문 수"},
		)
	}
	if len(metrics) == 0 {
		raw, _ := json.Marshal(document)
		text := string(raw)
		if strings.Contains(text, "결제금액") && strings.Contains(text, "주문번호") {
			metrics = append(metrics,
				map[string]any{"column": "결제금액", "op": "sum", "as": "매출 합계"},
				map[string]any{"column": "주문번호", "op": "count", "as": "주문 수"},
			)
		}
	}

	normalized := map[string]any{
		"scenario_id":      scenarioID,
		"request_kind":     normalizedValidatedExecutionRequestKind(document),
		"input_file":       inputFile,
		"source_sheet":     sourceSheet,
		"output_file":      outputFile,
		"execution_kind":   "composition",
		"composition_kind": "group_summary",
		"target_sheet":     targetSheet,
		"summary_mode":     summaryMode,
		"filters":          filters,
		"group_by":         groupBy,
		"metrics":          metrics,
	}
	if requestText, ok := document["request_text"].(string); ok && strings.TrimSpace(requestText) != "" {
		normalized["request_text"] = requestText
	}

	return normalized, nil
}

func normalizedValidatedExecutionRequestKind(document map[string]any) string {
	if requestText, _ := document["request_text"].(string); strings.TrimSpace(requestText) != "" {
		return "prompt_text"
	}
	return "structured_use_request"
}

func normalizeVerificationReviewValue(value any) (any, error) {
	if err := validateVerificationReviewValue(value); err == nil {
		return value, nil
	}

	document, ok := value.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("verification review has unsupported shape")
	}

	status, _ := document["status"].(string)
	if strings.TrimSpace(status) == "" {
		if outcome, _ := document["outcome"].(string); strings.TrimSpace(outcome) != "" {
			status = outcome
		}
	}
	if strings.TrimSpace(status) == "" {
		if blocked, ok := document["blocked"].(bool); ok && blocked {
			status = "blocked"
		}
	}
	if strings.TrimSpace(status) == "" {
		if needsReview, ok := document["needs_review"].(bool); ok && needsReview {
			status = "needs_human_checkpoint"
		}
	}
	if status == "" {
		status = "blocked"
	}
	if verifiedPass, ok := document["verification_pass"].(bool); ok {
		if status == "" || status == "blocked" {
			if verifiedPass {
				status = "pass"
			} else {
				status = "fail"
			}
		}
	}
	if verification, ok := document["verification"].(map[string]any); ok {
		if verifiedPass, ok := verification["pass"].(bool); ok {
			if status == "" || status == "blocked" {
				if verifiedPass {
					status = "pass"
				} else {
					status = "fail"
				}
			}
		}
	}
	switch status {
	case "verified_pass":
		status = "pass"
	case "verified_fail":
		status = "fail"
	case "needs_review", "needs_human_review":
		status = "needs_human_checkpoint"
	}

	summary, _ := document["summary"].(string)
	if strings.TrimSpace(summary) == "" {
		if assessment, _ := document["assessment"].(string); strings.TrimSpace(assessment) != "" {
			summary = assessment
		}
	}
	if strings.TrimSpace(summary) == "" {
		summary = stringifyLoopMessage(document["evidence"])
	}
	if strings.TrimSpace(summary) == "" {
		if verificationReview, ok := document["verification_review"].(map[string]any); ok {
			if conclusion, _ := verificationReview["conclusion"].(string); strings.TrimSpace(conclusion) != "" {
				summary = conclusion
			}
			if strings.TrimSpace(summary) == "" {
				summary = stringifyLoopMessage(verificationReview["evidence"])
			}
		}
	}
	if strings.TrimSpace(summary) == "" {
		if verification, ok := document["verification"].(map[string]any); ok {
			if review, _ := verification["review"].(string); strings.TrimSpace(review) != "" {
				summary = review
			}
			if strings.TrimSpace(summary) == "" {
				summary = stringifyLoopMessage(verification["reasons"])
			}
		}
	}
	if strings.TrimSpace(summary) == "" {
		if conclusion, _ := document["conclusion"].(string); strings.TrimSpace(conclusion) != "" {
			summary = conclusion
		}
	}
	if strings.TrimSpace(summary) == "" {
		summary = stringifyLoopMessage(document["reasons"])
	}
	if strings.TrimSpace(summary) == "" {
		summary = stringifyLoopMessage(document["findings"])
	}
	if strings.TrimSpace(summary) == "" {
		summary = stringifyLoopMessage(document["limitations"])
	}

	reasons := []any{}
	if rawReasons, ok := document["reasons"].([]any); ok {
		reasons = rawReasons
	} else if limitations, ok := document["limitations"].([]any); ok {
		reasons = limitations
	}
	if len(reasons) == 0 {
		if verificationReview, ok := document["verification_review"].(map[string]any); ok {
			if evidence, ok := verificationReview["evidence"].([]any); ok {
				reasons = evidence
			}
		}
	}
	if len(reasons) == 0 {
		if verification, ok := document["verification"].(map[string]any); ok {
			if verificationReasons, ok := verification["reasons"].([]any); ok {
				reasons = verificationReasons
			}
		}
	}
	if len(reasons) == 0 {
		if limits, ok := document["limits"].([]any); ok {
			reasons = limits
		}
	}
	if len(reasons) == 0 {
		if findings, ok := document["findings"].([]any); ok {
			reasons = findings
		}
	}

	return map[string]any{
		"status":  status,
		"summary": summary,
		"reasons": reasons,
	}, nil
}
