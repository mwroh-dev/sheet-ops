package useorchestrator

import (
	"fmt"
	"path/filepath"
	"strings"

	runtimeknowledge "github.com/mwroh/sheet-ops/runtime/knowledge"
	runtimeworkbookcase "github.com/mwroh/sheet-ops/runtime/workbookcase"
)

type routingContext struct {
	selectedOperation string
	routingAuthority  string
	note              string
	knowledge         runtimeknowledge.OrchestratorRoutingKnowledge
}

func NormalizeUseRequest(req UseRequest) UseRequest {
	effectiveOperation := req.Operation
	if effectiveOperation == "" && isSummaryRequest(req) {
		effectiveOperation = runtimeworkbookcase.SummaryOperationName
	}
	if effectiveOperation == runtimeworkbookcase.SummaryOperationName {
		if req.TargetSheet == "" {
			req.TargetSheet = "요약"
		}
		if req.SummaryMode == "" {
			req.SummaryMode = "values"
		}
	}
	if effectiveOperation == runtimeworkbookcase.JoinLookupOperationName {
		if req.TargetSheet == "" {
			req.TargetSheet = "조회결과"
		}
	}
	return req
}

func loadRoutingContext(req UseRequest) routingContext {
	knowledge := advisoryRoutingKnowledge()
	selectedOperation, routingAuthority := resolveRequestedOperation(req)
	return routingContext{
		selectedOperation: selectedOperation,
		routingAuthority:  routingAuthority,
		note:              buildRoutingNote(selectedOperation, routingAuthority, knowledge),
		knowledge:         knowledge,
	}
}

func advisoryRoutingKnowledge() runtimeknowledge.OrchestratorRoutingKnowledge {
	knowledge, err := runtimeknowledge.LoadOrchestratorRoutingKnowledge()
	if err == nil {
		return knowledge
	}
	return runtimeknowledge.OrchestratorRoutingKnowledge{
		KnowledgePath:  filepath.Join(runtimeknowledge.KnowledgeRoot(), "orchestrator", "episodic", "records.jsonl"),
		UsedFallback:   true,
		FallbackReason: fmt.Sprintf("Unable to read orchestrator knowledge (%v); routing will use request signals only.", err),
	}
}

func resolveRequestedOperation(req UseRequest) (string, string) {
	if req.Operation != "" {
		return req.Operation, "explicit_operation"
	}
	if isSummaryRequest(req) {
		return runtimeworkbookcase.SummaryOperationName, "summary_request_signals"
	}
	if isAppendRowsRequest(req) {
		return runtimeworkbookcase.AppendRowsOperationName, "append_rows_request_signals"
	}
	if isFormulaExtensionRequest(req) {
		return runtimeworkbookcase.ExtendFormulasOperationName, "formula_extension_request_signals"
	}
	if isPeriodCopyRequest(req) {
		return runtimeworkbookcase.CopyPeriodSheetOperationName, "period_copy_request_signals"
	}
	if isPeriodRollForwardRequest(req) {
		return runtimeworkbookcase.RollForwardPeriodOperationName, "period_roll_forward_request_signals"
	}
	if isReconcileTablesRequest(req) {
		return runtimeworkbookcase.ReconcileTablesOperationName, "table_reconciliation_request_signals"
	}
	if isPrintableFormRequest(req) {
		return runtimeworkbookcase.GeneratePrintableFormOperationName, "printable_form_request_signals"
	}
	if isDataValidationRequest(req) {
		return runtimeworkbookcase.AddDataValidationOperationName, "data_validation_request_signals"
	}
	if isFormulaProtectionRequest(req) {
		return runtimeworkbookcase.ProtectFormulaCellsOperationName, "formula_protection_request_signals"
	}
	if isHeaderNormalizationRequest(req) {
		return runtimeworkbookcase.NormalizeHeadersOperationName, "header_normalization_request_signals"
	}
	if isJoinLookupRequest(req) {
		return runtimeworkbookcase.JoinLookupOperationName, "join_lookup_request_signals"
	}
	if isHighlightRequest(req) {
		return runtimeworkbookcase.HighlightOperationName, "highlight_request_signals"
	}
	return "", "unsupported_request_signals"
}

func isAppendRowsRequest(req UseRequest) bool {
	return len(req.Values) > 0 ||
		strings.Contains(req.RequestText, "행 추가") ||
		strings.Contains(req.RequestText, "append")
}

func isFormulaExtensionRequest(req UseRequest) bool {
	return req.FormulaSourceRow > 0 ||
		len(req.TargetRows) > 0 ||
		len(req.FormulaColumns) > 0 ||
		strings.Contains(req.RequestText, "수식 확장") ||
		strings.Contains(req.RequestText, "수식을 확장") ||
		strings.Contains(req.RequestText, "formula extension") ||
		strings.Contains(req.RequestText, "extend formulas")
}

func isPeriodCopyRequest(req UseRequest) bool {
	return req.TargetSheet != "" && (strings.Contains(req.RequestText, "시트 복사") ||
		strings.Contains(req.RequestText, "period copy") ||
		strings.Contains(req.RequestText, "copy period") ||
		strings.Contains(req.RequestText, "copy sheet"))
}

func isPeriodRollForwardRequest(req UseRequest) bool {
	return len(req.CarryForwardMappings) > 0 ||
		strings.Contains(req.RequestText, "이월") ||
		strings.Contains(req.RequestText, "roll forward") ||
		strings.Contains(req.RequestText, "carry forward")
}

func isDataValidationRequest(req UseRequest) bool {
	return req.ValidationRule != nil ||
		strings.Contains(req.RequestText, "데이터 검증") ||
		strings.Contains(req.RequestText, "dropdown") ||
		strings.Contains(req.RequestText, "data validation")
}

func isReconcileTablesRequest(req UseRequest) bool {
	return len(req.CompareMappings) > 0 ||
		req.LeftKey != "" ||
		req.RightKey != "" ||
		strings.Contains(req.RequestText, "대조") ||
		strings.Contains(req.RequestText, "reconcile") ||
		strings.Contains(req.RequestText, "reconciliation")
}

func isPrintableFormRequest(req UseRequest) bool {
	return len(req.FieldBindings) > 0 ||
		req.TableBinding != nil ||
		req.FormTitle != "" ||
		req.PrintArea != "" ||
		strings.Contains(req.RequestText, "printable") ||
		strings.Contains(req.RequestText, "출력 양식") ||
		strings.Contains(req.RequestText, "인쇄")
}

func isFormulaProtectionRequest(req UseRequest) bool {
	return req.ProtectionRule != nil ||
		strings.Contains(req.RequestText, "수식 보호") ||
		strings.Contains(req.RequestText, "formula protection") ||
		strings.Contains(req.RequestText, "protect formula")
}

func isHeaderNormalizationRequest(req UseRequest) bool {
	return len(req.HeaderMappings) > 0 ||
		strings.Contains(req.RequestText, "헤더 정규화") ||
		strings.Contains(req.RequestText, "header normalization") ||
		strings.Contains(req.RequestText, "normalize headers")
}

func isSummaryRequest(req UseRequest) bool {
	return len(req.Filters) > 0 ||
		len(req.GroupBy) > 0 ||
		len(req.Metrics) > 0 ||
		req.SummaryMode != "" ||
		strings.Contains(req.RequestText, "요약")
}

func isHighlightRequest(req UseRequest) bool {
	return req.TargetColumn != "" ||
		req.Operator != "" ||
		req.Threshold != nil ||
		req.HighlightColor != "" ||
		(strings.Contains(req.RequestText, "표시") && (strings.Contains(req.RequestText, "보다 큰") || strings.Contains(req.RequestText, "초과") || strings.Contains(req.RequestText, "이상") || strings.Contains(req.RequestText, "미만") || strings.Contains(req.RequestText, "이하")))
}

func isJoinLookupRequest(req UseRequest) bool {
	return req.LookupSheet != "" ||
		req.JoinKey != "" ||
		len(req.IncludeSourceColumns) > 0 ||
		len(req.AppendLookupColumns) > 0 ||
		strings.Contains(req.RequestText, "조인") ||
		strings.Contains(req.RequestText, "매칭") ||
		strings.Contains(req.RequestText, "lookup")
}

func buildRoutingNote(selectedOperation, routingAuthority string, knowledge runtimeknowledge.OrchestratorRoutingKnowledge) string {
	if knowledge.UsedFallback {
		switch routingAuthority {
		case "explicit_operation":
			return fmt.Sprintf("%s Fallback retained explicit operation %s.", knowledge.FallbackReason, selectedOperation)
		case "summary_request_signals":
			return fmt.Sprintf("%s Fallback routed by summary request signals to %s.", knowledge.FallbackReason, selectedOperation)
		case "join_lookup_request_signals":
			return fmt.Sprintf("%s Fallback routed by join lookup request signals to %s.", knowledge.FallbackReason, selectedOperation)
		case "append_rows_request_signals":
			return fmt.Sprintf("%s Fallback routed by append row request signals to %s.", knowledge.FallbackReason, selectedOperation)
		case "formula_extension_request_signals":
			return fmt.Sprintf("%s Fallback routed by formula extension request signals to %s.", knowledge.FallbackReason, selectedOperation)
		case "period_copy_request_signals":
			return fmt.Sprintf("%s Fallback routed by period copy request signals to %s.", knowledge.FallbackReason, selectedOperation)
		case "period_roll_forward_request_signals":
			return fmt.Sprintf("%s Fallback routed by period roll-forward request signals to %s.", knowledge.FallbackReason, selectedOperation)
		case "table_reconciliation_request_signals":
			return fmt.Sprintf("%s Fallback routed by table reconciliation request signals to %s.", knowledge.FallbackReason, selectedOperation)
		case "printable_form_request_signals":
			return fmt.Sprintf("%s Fallback routed by printable form request signals to %s.", knowledge.FallbackReason, selectedOperation)
		case "data_validation_request_signals":
			return fmt.Sprintf("%s Fallback routed by data validation request signals to %s.", knowledge.FallbackReason, selectedOperation)
		case "formula_protection_request_signals":
			return fmt.Sprintf("%s Fallback routed by formula protection request signals to %s.", knowledge.FallbackReason, selectedOperation)
		case "header_normalization_request_signals":
			return fmt.Sprintf("%s Fallback routed by header normalization request signals to %s.", knowledge.FallbackReason, selectedOperation)
		case "highlight_request_signals":
			return fmt.Sprintf("%s Fallback routed by highlight request signals to %s.", knowledge.FallbackReason, selectedOperation)
		default:
			return fmt.Sprintf("%s Fallback could not infer a supported operation from request signals.", knowledge.FallbackReason)
		}
	}

	switch routingAuthority {
	case "explicit_operation":
		return fmt.Sprintf(
			"Loaded %d orchestrator knowledge records; retained explicit operation %s. Knowledge remains advisory only.",
			knowledge.RecordCount,
			selectedOperation,
		)
	case "summary_request_signals":
		return fmt.Sprintf(
			"Loaded %d orchestrator knowledge records; request summary signals still selected %s. Knowledge remains advisory only.",
			knowledge.RecordCount,
			selectedOperation,
		)
	case "join_lookup_request_signals":
		return fmt.Sprintf(
			"Loaded %d orchestrator knowledge records; join lookup request signals still selected %s. Knowledge remains advisory only.",
			knowledge.RecordCount,
			selectedOperation,
		)
	case "append_rows_request_signals":
		return fmt.Sprintf(
			"Loaded %d orchestrator knowledge records; append row request signals still selected %s. Knowledge remains advisory only.",
			knowledge.RecordCount,
			selectedOperation,
		)
	case "period_copy_request_signals":
		return fmt.Sprintf(
			"Loaded %d orchestrator knowledge records; period copy request signals still selected %s. Knowledge remains advisory only.",
			knowledge.RecordCount,
			selectedOperation,
		)
	case "formula_protection_request_signals":
		return fmt.Sprintf(
			"Loaded %d orchestrator knowledge records; formula protection request signals still selected %s. Knowledge remains advisory only.",
			knowledge.RecordCount,
			selectedOperation,
		)
	case "period_roll_forward_request_signals":
		return fmt.Sprintf(
			"Loaded %d orchestrator knowledge records; period roll-forward request signals still selected %s. Knowledge remains advisory only.",
			knowledge.RecordCount,
			selectedOperation,
		)
	case "table_reconciliation_request_signals":
		return fmt.Sprintf(
			"Loaded %d orchestrator knowledge records; table reconciliation request signals still selected %s. Knowledge remains advisory only.",
			knowledge.RecordCount,
			selectedOperation,
		)
	case "printable_form_request_signals":
		return fmt.Sprintf(
			"Loaded %d orchestrator knowledge records; printable form request signals still selected %s. Knowledge remains advisory only.",
			knowledge.RecordCount,
			selectedOperation,
		)
	case "header_normalization_request_signals":
		return fmt.Sprintf(
			"Loaded %d orchestrator knowledge records; header normalization request signals still selected %s. Knowledge remains advisory only.",
			knowledge.RecordCount,
			selectedOperation,
		)
	case "highlight_request_signals":
		return fmt.Sprintf(
			"Loaded %d orchestrator knowledge records; highlight request signals still selected %s. Knowledge remains advisory only.",
			knowledge.RecordCount,
			selectedOperation,
		)
	default:
		return fmt.Sprintf(
			"Loaded %d orchestrator knowledge records; request signals did not map to a supported operation. Knowledge remains advisory only.",
			knowledge.RecordCount,
		)
	}
}
