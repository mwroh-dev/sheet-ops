package requestcompiler

import (
	"regexp"
	"slices"
	"strconv"
	"strings"

	runtimeinspect "github.com/mwroh/sheet-ops/runtime/inspect"
)

func augmentIntentFromWorkbookFacts(requestText string, facts runtimeinspect.WorkbookFacts, intent NormalizedIntent) NormalizedIntent {
	intent = normalizeIntent(intent)

	visibleSheets := visibleSheetNames(facts)
	mentionedSheets := mentionedNames(requestText, visibleSheets)
	headersBySheet := sheetHeadersByName(facts)

	if len(intent.SourceSheetCandidates) == 0 {
		if len(mentionedSheets) > 0 {
			intent.SourceSheetCandidates = appendUnique(intent.SourceSheetCandidates, mentionedSheets[0])
		} else if len(visibleSheets) == 1 {
			intent.SourceSheetCandidates = appendUnique(intent.SourceSheetCandidates, visibleSheets[0])
		}
	}

	if isJoinLookupRequestText(requestText) {
		intent.CompositionCandidates = appendUnique(intent.CompositionCandidates, CompositionCandidateJoinLookup)
		if len(intent.SourceSheetCandidates) == 0 && len(visibleSheets) > 0 {
			intent.SourceSheetCandidates = appendUnique(intent.SourceSheetCandidates, visibleSheets[0])
		}
		if len(intent.LookupSheetCandidates) == 0 {
			switch {
			case len(mentionedSheets) >= 2:
				intent.LookupSheetCandidates = appendUnique(intent.LookupSheetCandidates, mentionedSheets[1])
			case len(visibleSheets) == 2 && len(intent.SourceSheetCandidates) == 1:
				for _, sheet := range visibleSheets {
					if sheet != intent.SourceSheetCandidates[0] {
						intent.LookupSheetCandidates = appendUnique(intent.LookupSheetCandidates, sheet)
					}
				}
			}
		}
		joinSource := firstOrEmpty(intent.SourceSheetCandidates)
		joinLookup := firstOrEmpty(intent.LookupSheetCandidates)
		sharedHeaders := intersectHeaders(headersBySheet[joinSource], headersBySheet[joinLookup])
		if intent.JoinLookup.JoinKey == "" {
			intent.JoinLookup.JoinKey = detectJoinKey(requestText, sharedHeaders)
		}
		if len(intent.JoinLookup.AppendLookupColumns) == 0 {
			intent.JoinLookup.AppendLookupColumns = detectMentionedColumns(requestText, headersBySheet[joinLookup], intent.JoinLookup.JoinKey)
		}
		return intent
	}

	if isHighlightRequestText(requestText) {
		intent.CompositionCandidates = appendUnique(intent.CompositionCandidates, CompositionCandidateThresholdHighlight)
		sourceHeaders := headersBySheet[firstOrEmpty(intent.SourceSheetCandidates)]
		if intent.Highlight.TargetColumn == "" {
			intent.Highlight.TargetColumn = detectMentionedColumn(requestText, sourceHeaders)
		}
		if intent.Highlight.Operator == "" || intent.Highlight.Threshold == nil {
			if operator, threshold, ok := detectThresholdRule(requestText); ok {
				if intent.Highlight.Operator == "" {
					intent.Highlight.Operator = operator
				}
				if intent.Highlight.Threshold == nil {
					intent.Highlight.Threshold = threshold
				}
			}
		}
		if intent.Highlight.HighlightColor == "" && (containsAny(requestText, "노란색", "노란", "yellow")) {
			intent.Highlight.HighlightColor = "#FFF59D"
		}
		return intent
	}

	if isSummaryRequestText(requestText) {
		intent.CompositionCandidates = appendUnique(intent.CompositionCandidates, CompositionCandidateGroupSummary)
		sourceHeaders := headersBySheet[firstOrEmpty(intent.SourceSheetCandidates)]
		if len(intent.Summary.GroupBy) == 0 {
			intent.Summary.GroupBy = detectGroupingColumns(requestText, sourceHeaders)
		}
		if len(intent.Summary.Metrics) == 0 {
			intent.Summary.Metrics = detectSummaryMetrics(requestText, sourceHeaders)
		}
		if len(intent.Summary.Filters) == 0 {
			intent.Summary.Filters = detectSimpleFilters(requestText, sourceHeaders)
		}
		if intent.Summary.SummaryMode == "" {
			if summaryMode := explicitSummaryModeFromRequest(requestText); summaryMode != "" {
				intent.Summary.SummaryMode = summaryMode
			}
		}
	}

	return intent
}

func visibleSheetNames(facts runtimeinspect.WorkbookFacts) []string {
	names := make([]string, 0, len(facts.Sheets))
	for _, sheet := range facts.Sheets {
		if sheet.Hidden {
			continue
		}
		names = append(names, sheet.Name)
	}
	return names
}

func sheetHeadersByName(facts runtimeinspect.WorkbookFacts) map[string][]string {
	byName := make(map[string][]string, len(facts.Sheets))
	for _, sheet := range facts.Sheets {
		byName[sheet.Name] = append([]string(nil), sheet.Columns...)
	}
	return byName
}

func mentionedNames(requestText string, names []string) []string {
	mentioned := make([]string, 0, len(names))
	for _, name := range names {
		if strings.Contains(strings.TrimSpace(requestText), name) {
			mentioned = append(mentioned, name)
		}
	}
	return mentioned
}

func detectGroupingColumns(requestText string, headers []string) []string {
	groupBy := make([]string, 0, 1)
	for _, header := range headers {
		if containsAny(requestText, header+"별", header+"별로", header+" 기준") {
			groupBy = append(groupBy, header)
		}
	}
	if len(groupBy) == 0 && hasProductGroupingIntent(requestText) && slices.Contains(headers, "상품명") {
		groupBy = append(groupBy, "상품명")
	}
	return groupBy
}

func detectSummaryMetrics(requestText string, headers []string) []AggregateIntent {
	metrics := make([]AggregateIntent, 0, 2)
	for _, header := range headers {
		if isSumMetricRequest(requestText, header) {
			metrics = append(metrics, AggregateIntent{Kind: "sum", Column: header})
		}
		if isCountMetricRequest(requestText, header) {
			metrics = append(metrics, AggregateIntent{Kind: "count", Column: header})
		}
	}
	return metrics
}

func detectSimpleFilters(requestText string, headers []string) []FilterIntent {
	if slices.Contains(headers, "배송상태") &&
		containsAny(requestText, "취소된 주문은 제외", "취소 주문은 제외", "취소된 주문을 제외", "취소 주문을 제외", "취소된 주문은 빼고", "취소 주문은 빼고") {
		return []FilterIntent{{Column: "배송상태", Op: "!=", Value: "취소"}}
	}
	return nil
}

func detectMentionedColumn(requestText string, headers []string) string {
	for _, header := range headers {
		if strings.Contains(strings.TrimSpace(requestText), header) {
			return header
		}
	}
	return ""
}

func detectMentionedColumns(requestText string, headers []string, exclude string) []string {
	columns := make([]string, 0, len(headers))
	for _, header := range headers {
		if header == exclude {
			continue
		}
		if strings.Contains(strings.TrimSpace(requestText), header) {
			columns = append(columns, header)
		}
	}
	return columns
}

func detectJoinKey(requestText string, sharedHeaders []string) string {
	for _, header := range sharedHeaders {
		if strings.Contains(strings.TrimSpace(requestText), header) {
			return header
		}
	}
	if len(sharedHeaders) == 1 {
		return sharedHeaders[0]
	}
	return ""
}

func intersectHeaders(left []string, right []string) []string {
	if len(left) == 0 || len(right) == 0 {
		return nil
	}
	rightSet := make(map[string]struct{}, len(right))
	for _, value := range right {
		rightSet[value] = struct{}{}
	}
	shared := make([]string, 0, len(left))
	for _, value := range left {
		if _, ok := rightSet[value]; ok {
			shared = append(shared, value)
		}
	}
	return shared
}

func detectThresholdRule(requestText string) (string, *float64, bool) {
	type operatorPattern struct {
		operator string
		re       *regexp.Regexp
	}
	for _, pattern := range []operatorPattern{
		{operator: ">=", re: regexp.MustCompile(`([0-9][0-9,]*)\s*(이상|or more)`)},
		{operator: ">", re: regexp.MustCompile(`([0-9][0-9,]*)\s*(보다 큰|초과|greater than)`)},
		{operator: "<=", re: regexp.MustCompile(`([0-9][0-9,]*)\s*(이하)`)},
		{operator: "<", re: regexp.MustCompile(`([0-9][0-9,]*)\s*(보다 작은|미만|less than)`)},
		{operator: "=", re: regexp.MustCompile(`([0-9][0-9,]*)\s*(와 같은|와 동일한|같은)`)},
	} {
		match := pattern.re.FindStringSubmatch(strings.TrimSpace(requestText))
		if len(match) < 2 {
			continue
		}
		value, err := strconv.ParseFloat(strings.ReplaceAll(match[1], ",", ""), 64)
		if err != nil {
			continue
		}
		return pattern.operator, &value, true
	}
	return "", nil, false
}

func isSummaryRequestText(requestText string) bool {
	return containsAny(requestText, "요약", "집계", "정리")
}

func isHighlightRequestText(requestText string) bool {
	return containsAny(requestText, "표시", "강조", "highlight")
}

func isJoinLookupRequestText(requestText string) bool {
	return containsAny(requestText, "조인", "매칭", "lookup", "붙인", "붙여")
}

func isSumMetricRequest(requestText string, header string) bool {
	if header == "결제금액" {
		return hasRevenueMetricIntent(requestText)
	}
	escaped := regexp.QuoteMeta(header)
	return regexp.MustCompile(escaped+`\s*(의|를|은|는)?\s*(합계|합산|총합|sum)`).MatchString(strings.TrimSpace(requestText)) ||
		regexp.MustCompile(`(합계|합산|총합|sum)\s*(의|를|은|는)?\s*`+escaped).MatchString(strings.TrimSpace(requestText))
}

func isCountMetricRequest(requestText string, header string) bool {
	if header == "주문번호" {
		return hasOrderCountIntent(requestText)
	}
	return strings.Contains(strings.TrimSpace(requestText), header) &&
		containsAny(requestText, "건수", "개수", "수", "count")
}

func firstOrEmpty(values []string) string {
	if len(values) == 0 {
		return ""
	}
	return values[0]
}
