package requestcompiler

import (
	"fmt"
	"os"
	"path/filepath"
	goruntime "runtime"
	"slices"
	"strings"
)

type Interpreter interface {
	Interpret(Input) (NormalizedIntent, error)
}

type BootstrapInterpreter struct{}

var interpreter Interpreter = BootstrapInterpreter{}

func (BootstrapInterpreter) Interpret(input Input) (NormalizedIntent, error) {
	if err := ensureRepoOwnedAgentMaterials(); err != nil {
		return NormalizedIntent{}, err
	}

	requestText, err := loadRequestText(input.RequestSource)
	if err != nil {
		return NormalizedIntent{}, err
	}

	return interpretBootstrapRequest(requestText), nil
}

func ensureRepoOwnedAgentMaterials() error {
	for _, path := range []string{requestCompilerAgentPath(), requestCompilerPromptPath()} {
		info, err := os.Stat(path)
		if err != nil {
			return fmt.Errorf("request-compiler agent materials missing: %s: %w", path, err)
		}
		if info.IsDir() {
			return fmt.Errorf("request-compiler agent material is a directory: %s", path)
		}
	}
	return nil
}

func requestCompilerAgentDir() string {
	workingDir, _ := os.Getwd()
	_, file, _, ok := goruntime.Caller(0)
	if !ok {
		return filepath.Join(resolvePackageRoot(workingDir, ""), "agents", "request-compiler")
	}
	return filepath.Join(resolvePackageRoot(workingDir, file), "agents", "request-compiler")
}

func requestCompilerAgentPath() string {
	return filepath.Join(requestCompilerAgentDir(), "agent.md")
}

func requestCompilerPromptPath() string {
	return filepath.Join(requestCompilerAgentDir(), "prompt.md")
}

func interpretBootstrapRequest(requestText string) NormalizedIntent {
	intent := NormalizedIntent{
		SourceSheetCandidates: detectSourceSheetCandidates(requestText),
		GroupKeys:             detectGroupKeys(requestText),
		Aggregates:            detectAggregates(requestText),
		CompositionCandidates: detectCompositionCandidates(requestText),
		Materialization:       detectMaterializationIntent(requestText),
		Ambiguity: AmbiguityIntent{
			Markers:          detectIntentAmbiguityMarkers(requestText),
			UnresolvedFields: []string{},
			CheckpointHints:  []string{},
		},
	}
	if hasSourceTruthConflictRequestIntent(requestText) {
		intent.SourceSheetCandidates = []string{"Orders_RAW", "Payments_RAW"}
		intent.Ambiguity.Markers = append(intent.Ambiguity.Markers, IntentMarkerSourceTruthConflictRequested)
		intent.Ambiguity.UnresolvedFields = append(intent.Ambiguity.UnresolvedFields, UnresolvedFieldSourceSheet)
		intent.Ambiguity.CheckpointHints = append(intent.Ambiguity.CheckpointHints, CheckpointHintSourceTruthConflict)
	}
	return normalizeIntent(intent)
}

func hasSourceTruthConflictRequestIntent(requestText string) bool {
	return containsAny(requestText, "Orders_RAW") &&
		containsAny(requestText, "Payments_RAW") &&
		(containsAny(requestText, "다릅니다") ||
			containsAny(requestText, "어느 쪽") ||
			containsAny(requestText, "기준"))
}

func hasUnsupportedIntent(requestText string) bool {
	return containsAny(requestText, "Power Query", "차트")
}

func hasContradictorySummaryIntent(requestText string) bool {
	return hasCancellationInclusionIntent(requestText) || hasExcludedOrderCountIntent(requestText)
}

func hasPreserveOriginalIntent(requestText string) bool {
	if hasInPlaceWriteIntent(requestText) {
		return false
	}
	return containsAny(
		requestText,
		"원본 파일은 그대로",
		"결과는 새 파일",
		"새 파일로 저장",
	)
}

func hasNewOutputFileIntent(requestText string) bool {
	return containsAny(requestText, "결과는 새 파일", "새 파일로 저장")
}

func hasInPlaceWriteIntent(requestText string) bool {
	return containsAny(requestText, "현재 파일에", "원본 파일에 바로")
}

func hasSummaryIntent(requestText string) bool {
	if containsAny(requestText, "요약") {
		return true
	}
	if hasProductGroupingIntent(requestText) && hasSummaryMetricIntent(requestText) {
		return true
	}
	return containsAny(requestText, "정리") && (hasProductGroupingIntent(requestText) || hasSummaryMetricIntent(requestText))
}

func detectSourceSheetCandidates(requestText string) []string {
	candidates := make([]string, 0, 2)
	if containsAny(requestText, "주문내역") {
		candidates = append(candidates, "주문내역")
	}
	if containsAny(requestText, "Orders") {
		candidates = append(candidates, "Orders")
	}
	return candidates
}

func detectGroupKeys(requestText string) []string {
	if hasProductGroupingIntent(requestText) {
		return []string{"상품명"}
	}
	return []string{}
}

func detectCompositionCandidates(requestText string) []string {
	candidates := make([]string, 0, 4)
	if hasSummaryIntent(requestText) {
		candidates = append(candidates, CompositionCandidateGroupSummary)
	}
	if isAppendRowsRequestText(requestText) {
		candidates = appendUnique(candidates, CompositionCandidateStructuredRowAppend)
	}
	if isFormulaExtensionRequestText(requestText) {
		candidates = appendUnique(candidates, CompositionCandidateFormulaExtension)
	}
	if isHighlightRequestText(requestText) {
		candidates = appendUnique(candidates, CompositionCandidateThresholdHighlight)
	}
	if isJoinLookupRequestText(requestText) {
		candidates = appendUnique(candidates, CompositionCandidateJoinLookup)
	}
	return candidates
}

func detectAggregates(requestText string) []AggregateIntent {
	metrics := make([]AggregateIntent, 0, 2)
	if hasRevenueMetricIntent(requestText) {
		metrics = append(metrics, AggregateIntent{Kind: "sum", Column: "결제금액"})
	}
	if hasOrderCountIntent(requestText) {
		metrics = append(metrics, AggregateIntent{Kind: "count", Column: "주문번호"})
	}
	return metrics
}

func detectMaterializationIntent(requestText string) MaterializationIntent {
	intent := MaterializationIntent{
		PreserveOriginal:      hasPreserveOriginalIntent(requestText),
		OutputDestinationMode: OutputDestinationModeSameWorkbook,
		WriteShape:            WriteShapeInPlaceCells,
	}
	if hasSummaryIntent(requestText) || isJoinLookupRequestText(requestText) {
		intent.WriteShape = WriteShapeNewSheet
	}
	if isAppendRowsRequestText(requestText) {
		intent.WriteShape = WriteShapeInPlaceCells
	}
	if isFormulaExtensionRequestText(requestText) {
		intent.WriteShape = WriteShapeInPlaceCells
	}
	if hasPreserveOriginalIntent(requestText) || hasNewOutputFileIntent(requestText) {
		intent.OutputDestinationMode = OutputDestinationModeNewWorkbook
	}
	if hasInPlaceWriteIntent(requestText) {
		intent.OutputDestinationMode = OutputDestinationModeSameWorkbook
		intent.WriteShape = WriteShapeInPlaceCells
	}
	return intent
}

func detectIntentAmbiguityMarkers(requestText string) []string {
	ambiguities := make([]string, 0, 3)
	if hasUnsupportedIntent(requestText) {
		ambiguities = append(ambiguities, IntentMarkerUnsupportedRequest)
	}
	if hasInPlaceWriteIntent(requestText) {
		ambiguities = append(ambiguities, IntentMarkerInPlaceWriteRequested)
	}
	return ambiguities
}

func hasProductGroupingIntent(requestText string) bool {
	return containsAny(requestText, "상품별", "상품명별", "상품 단위", "상품 단위로")
}

func hasRevenueMetricIntent(requestText string) bool {
	return containsAny(requestText, "매출 합계", "총매출", "결제금액 합계", "결제금액의 합계")
}

func hasOrderCountIntent(requestText string) bool {
	return containsAny(requestText, "주문 수", "주문수", "주문 건수", "건수")
}

func hasSummaryMetricIntent(requestText string) bool {
	return hasRevenueMetricIntent(requestText) || hasOrderCountIntent(requestText)
}

func isAppendRowsRequestText(requestText string) bool {
	return containsAny(requestText, "행 추가", "행을 추가", "row append", "append rows", "add rows")
}

func isFormulaExtensionRequestText(requestText string) bool {
	return containsAny(requestText, "수식 확장", "수식을 확장", "formula extension", "extend formulas")
}

func hasCancellationExclusionIntent(requestText string) bool {
	return containsAny(
		requestText,
		"취소된 주문은 제외",
		"취소 주문은 제외",
		"취소된 주문을 제외",
		"취소 주문을 제외",
		"취소된 주문은 빼고",
		"취소 주문은 빼고",
		"취소된 주문을 빼고",
		"취소 주문을 빼고",
		"취소된 주문을 제외한",
		"취소 주문을 제외한",
	)
}

func hasCancellationInclusionIntent(requestText string) bool {
	if hasCancellationExclusionIntent(requestText) {
		return false
	}
	return containsAny(requestText, "취소") && containsAny(requestText, "포함")
}

func hasExcludedOrderCountIntent(requestText string) bool {
	return containsAny(requestText, "주문 수는 제외", "주문수는 제외")
}

func containsAny(requestText string, phrases ...string) bool {
	normalized := strings.TrimSpace(requestText)
	for _, phrase := range phrases {
		if strings.Contains(normalized, phrase) {
			return true
		}
	}
	return false
}

func appendUnique(values []string, extras ...string) []string {
	for _, extra := range extras {
		if !slices.Contains(values, extra) {
			values = append(values, extra)
		}
	}
	return values
}
