package requestcompiler

import (
	"bufio"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	runtimeinspect "github.com/mwroh/sheet-ops/runtime/inspect"
)

var blankLineSplitPattern = regexp.MustCompile(`\n\s*\n+`)

func applyPreferenceMemory(input Input, requestText string, facts runtimeinspect.WorkbookFacts, intent NormalizedIntent) (*MemoryMatchSummary, NormalizedIntent, error) {
	summary := &MemoryMatchSummary{
		AppliedDefaults:    []string{},
		IgnoredPreferences: []string{},
		ConflictNotes:      []string{},
	}

	paragraphs, err := loadPreferenceMemoryParagraphs(input.WorkspaceRoot)
	if err != nil {
		return nil, NormalizedIntent{}, err
	}
	if len(paragraphs) == 0 {
		return summary, intent, nil
	}

	matched := matchPreferenceParagraphs(requestText, facts, paragraphs)
	summary.MatchedParagraphCount = len(matched)
	if len(matched) == 0 {
		return summary, intent, nil
	}

	hints := MemoryPreferenceHints{}
	for _, paragraph := range matched {
		hints = mergeMemoryHints(hints, extractMemoryHints(paragraph))
	}

	intent = applyMemoryHintsToIntent(requestText, intent, hints, summary)
	return summary, intent, nil
}

func loadPreferenceMemoryParagraphs(workspaceRoot string) ([]string, error) {
	raw, err := os.ReadFile(preferenceMemoryPath(workspaceRoot))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	return splitMemoryParagraphs(string(raw)), nil
}

func splitMemoryParagraphs(raw string) []string {
	raw = strings.ReplaceAll(raw, "\r\n", "\n")
	parts := blankLineSplitPattern.Split(strings.TrimSpace(raw), -1)
	paragraphs := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		paragraphs = append(paragraphs, part)
	}
	return paragraphs
}

func matchPreferenceParagraphs(requestText string, facts runtimeinspect.WorkbookFacts, paragraphs []string) []string {
	terms := memoryMatchTerms(requestText, facts)
	matched := make([]string, 0, len(paragraphs))
	for _, paragraph := range paragraphs {
		if !paragraphContainsRecognizedPreference(paragraph) {
			continue
		}
		if containsAny(paragraph, "엑셀", "excel", "workbook", "워크북", "sheet", "시트", "파일") {
			matched = append(matched, paragraph)
			continue
		}
		for _, term := range terms {
			if term != "" && strings.Contains(strings.ToLower(paragraph), strings.ToLower(term)) {
				matched = append(matched, paragraph)
				break
			}
		}
	}
	return matched
}

func memoryMatchTerms(requestText string, facts runtimeinspect.WorkbookFacts) []string {
	terms := make([]string, 0, 32)
	terms = append(terms, tokenizeMemoryText(requestText)...)
	if isSummaryRequestText(requestText) {
		terms = append(terms, "요약", "summary", "집계", "정리")
	}
	if isHighlightRequestText(requestText) {
		terms = append(terms, "표시", "강조", "highlight")
	}
	if isJoinLookupRequestText(requestText) {
		terms = append(terms, "조인", "lookup", "매칭")
	}
	for _, sheet := range facts.Sheets {
		terms = append(terms, sheet.Name)
		terms = append(terms, sheet.Columns...)
	}
	return dedupeTrimmedStrings(terms)
}

func tokenizeMemoryText(raw string) []string {
	fields := strings.FieldsFunc(strings.TrimSpace(raw), func(r rune) bool {
		switch {
		case r >= '0' && r <= '9':
			return false
		case r >= 'A' && r <= 'Z':
			return false
		case r >= 'a' && r <= 'z':
			return false
		case r >= '가' && r <= '힣':
			return false
		case r == '_':
			return false
		default:
			return true
		}
	})
	return dedupeTrimmedStrings(fields)
}

func paragraphContainsRecognizedPreference(paragraph string) bool {
	return containsAny(
		paragraph,
		"수식",
		"formula",
		"formulas",
		"값으로",
		"value",
		"values",
		"원본",
		"새 파일",
		"new file",
		"new workbook",
		"현재 시트",
		"덮어쓴",
		"same workbook",
		"in-place",
		"in place",
		"새 시트",
		"new sheet",
	)
}

func extractMemoryHints(paragraph string) MemoryPreferenceHints {
	hints := MemoryPreferenceHints{}

	switch {
	case containsAny(paragraph, "수식", "formula", "formulas"):
		hints.SummaryMode = "formulas"
	case containsAny(paragraph, "값으로", "값만", "value", "values"):
		hints.SummaryMode = "values"
	}

	if containsAny(paragraph, "원본을 보존", "원본 보존", "원본 파일은 그대로", "원본은 그대로", "원본 유지") {
		value := true
		hints.PreserveOriginal = &value
	}

	switch {
	case containsAny(paragraph, "새 파일", "new file", "새 워크북", "new workbook"):
		hints.OutputDestinationMode = OutputDestinationModeNewWorkbook
	case containsAny(paragraph, "같은 워크북", "same workbook", "현재 워크북", "같은 파일"):
		hints.OutputDestinationMode = OutputDestinationModeSameWorkbook
	}

	switch {
	case containsAny(paragraph, "새 시트", "new sheet", "요약 시트"):
		hints.WriteShape = WriteShapeNewSheet
	case containsAny(paragraph, "현재 시트", "셀에 바로", "덮어쓴", "in-place", "in place", "같은 셀"):
		hints.WriteShape = WriteShapeInPlaceCells
	}

	return hints
}

func mergeMemoryHints(base MemoryPreferenceHints, extra MemoryPreferenceHints) MemoryPreferenceHints {
	if extra.SummaryMode != "" {
		base.SummaryMode = extra.SummaryMode
	}
	if extra.PreserveOriginal != nil {
		value := *extra.PreserveOriginal
		base.PreserveOriginal = &value
	}
	if extra.OutputDestinationMode != "" {
		base.OutputDestinationMode = extra.OutputDestinationMode
	}
	if extra.WriteShape != "" {
		base.WriteShape = extra.WriteShape
	}
	return base
}

func applyMemoryHintsToIntent(requestText string, intent NormalizedIntent, hints MemoryPreferenceHints, summary *MemoryMatchSummary) NormalizedIntent {
	intent = normalizeIntent(intent)

	if hints.SummaryMode != "" && isSummaryRequestText(requestText) {
		explicit := explicitSummaryModeFromRequest(requestText)
		switch {
		case explicit != "" && explicit != hints.SummaryMode:
			summary.ConflictNotes = appendUnique(summary.ConflictNotes, "summary.summary_mode request_overrides_memory")
		case explicit == "" && intent.Summary.SummaryMode == "":
			intent.Summary.SummaryMode = hints.SummaryMode
			summary.AppliedDefaults = appendUnique(summary.AppliedDefaults, "summary.summary_mode="+hints.SummaryMode)
		}
	}

	if hints.PreserveOriginal != nil && *hints.PreserveOriginal {
		switch {
		case hasInPlaceWriteIntent(requestText):
			summary.ConflictNotes = appendUnique(summary.ConflictNotes, "materialization.preserve_original request_overrides_memory")
		case !intent.Materialization.PreserveOriginal:
			intent.Materialization.PreserveOriginal = true
			summary.AppliedDefaults = appendUnique(summary.AppliedDefaults, "materialization.preserve_original=true")
		}
	}

	switch hints.OutputDestinationMode {
	case OutputDestinationModeNewWorkbook:
		switch {
		case hasInPlaceWriteIntent(requestText):
			summary.ConflictNotes = appendUnique(summary.ConflictNotes, "materialization.output_destination_mode request_overrides_memory")
		case !hasNewOutputFileIntent(requestText) && intent.Materialization.OutputDestinationMode != OutputDestinationModeNewWorkbook:
			intent.Materialization.OutputDestinationMode = OutputDestinationModeNewWorkbook
			summary.AppliedDefaults = appendUnique(summary.AppliedDefaults, "materialization.output_destination_mode=new_workbook")
		}
	case OutputDestinationModeSameWorkbook:
		summary.IgnoredPreferences = appendUnique(summary.IgnoredPreferences, "materialization.output_destination_mode=same_workbook")
	}

	switch hints.WriteShape {
	case WriteShapeNewSheet:
		switch {
		case hasInPlaceWriteIntent(requestText):
			summary.ConflictNotes = appendUnique(summary.ConflictNotes, "materialization.write_shape request_overrides_memory")
		case (isSummaryRequestText(requestText) || isJoinLookupRequestText(requestText)) && intent.Materialization.WriteShape != WriteShapeNewSheet:
			intent.Materialization.WriteShape = WriteShapeNewSheet
			summary.AppliedDefaults = appendUnique(summary.AppliedDefaults, "materialization.write_shape=new_sheet")
		}
	case WriteShapeInPlaceCells:
		summary.IgnoredPreferences = appendUnique(summary.IgnoredPreferences, "materialization.write_shape=in_place_cells")
	}

	return intent
}

func explicitSummaryModeFromRequest(requestText string) string {
	if !isSummaryRequestText(requestText) {
		return ""
	}
	switch {
	case containsAny(requestText, "수식이 아니라 값", "값으로 남", "값만 남", "values", "value only"):
		return "values"
	case containsAny(requestText, "수식", "formula", "formulas"):
		return "formulas"
	default:
		return ""
	}
}

func dedupeTrimmedStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	out := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		key := strings.ToLower(value)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, value)
	}
	return out
}

func (summary MemoryMatchSummary) hasEntries() bool {
	return summary.MatchedParagraphCount > 0 ||
		len(summary.AppliedDefaults) > 0 ||
		len(summary.IgnoredPreferences) > 0 ||
		len(summary.ConflictNotes) > 0
}

func InferWorkspaceRoot(inputPath string) string {
	start := strings.TrimSpace(inputPath)
	if start == "" {
		return "."
	}
	dir := start
	if info, err := os.Stat(start); err == nil && !info.IsDir() {
		dir = filepath.Dir(start)
	}
	dir = filepath.Clean(dir)
	fallback := dir
	for {
		if pathExists(filepath.Join(dir, ".sheet-ops-state")) || pathExists(filepath.Join(dir, ".codex")) || pathExists(filepath.Join(dir, ".git")) || pathExists(filepath.Join(dir, "go.mod")) {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return fallback
		}
		dir = parent
	}
}

func PersistPreferenceObservations(workspaceRoot string, observations []PreferenceObservation) error {
	if len(observations) == 0 {
		return nil
	}

	path := preferenceObservationPath(workspaceRoot)
	for _, observation := range observations {
		if strings.TrimSpace(observation.Key) == "" || strings.TrimSpace(observation.Value) == "" {
			continue
		}
		if strings.TrimSpace(observation.RecordedAt) == "" {
			observation.RecordedAt = time.Now().UTC().Format(time.RFC3339Nano)
		}
		if err := appendPreferenceObservation(path, observation); err != nil {
			return err
		}
	}
	return promotePreferenceMemory(workspaceRoot)
}

func appendPreferenceObservation(path string, observation PreferenceObservation) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	raw, err := json.Marshal(observation)
	if err != nil {
		return err
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()
	if _, err := file.Write(append(raw, '\n')); err != nil {
		return err
	}
	return nil
}

func promotePreferenceMemory(workspaceRoot string) error {
	records, err := loadPreferenceObservations(preferenceObservationPath(workspaceRoot))
	if err != nil {
		return err
	}

	stableParagraphs := stablePreferenceParagraphs(records)
	memoryPath := preferenceMemoryPath(workspaceRoot)
	existingParagraphs, err := loadExistingMemoryParagraphs(memoryPath)
	if err != nil {
		return err
	}
	userParagraphs := make([]string, 0, len(existingParagraphs))
	for _, paragraph := range existingParagraphs {
		if isManagedPreferenceParagraph(paragraph) {
			continue
		}
		userParagraphs = append(userParagraphs, paragraph)
	}
	finalParagraphs := append(userParagraphs, stableParagraphs...)
	finalParagraphs = dedupeTrimmedStrings(finalParagraphs)
	if len(finalParagraphs) == 0 {
		return nil
	}
	return writePreferenceMemory(memoryPath, finalParagraphs)
}

func loadPreferenceObservations(path string) ([]PreferenceObservation, error) {
	file, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	defer func() { _ = file.Close() }()

	var records []PreferenceObservation
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var record PreferenceObservation
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return records, nil
}

func stablePreferenceParagraphs(records []PreferenceObservation) []string {
	if len(records) == 0 {
		return nil
	}

	paragraphs := make([]string, 0, 3)
	if summaryMode, ok := stableObservationValue(records, "summary_mode"); ok {
		switch summaryMode {
		case "formulas":
			paragraphs = append(paragraphs, "요약 작업은 보통 수식으로 남긴다.")
		case "values":
			paragraphs = append(paragraphs, "요약 결과는 값으로 남긴다.")
		}
	}

	preserveStable := observationCount(records, "preserve_original", "true") >= 2
	outputStable := observationCount(records, "output_destination", OutputDestinationModeNewWorkbook) >= 2
	switch {
	case preserveStable && outputStable:
		paragraphs = append(paragraphs, "엑셀 작업은 항상 원본을 보존하고 새 파일로 저장한다.")
	case preserveStable:
		paragraphs = append(paragraphs, "엑셀 작업은 항상 원본을 보존한다.")
	case outputStable:
		paragraphs = append(paragraphs, "엑셀 작업 결과는 새 파일로 저장한다.")
	}

	sort.Strings(paragraphs)
	return paragraphs
}

func stableObservationValue(records []PreferenceObservation, key string) (string, bool) {
	type tally struct {
		count      int
		recordedAt string
	}
	values := make(map[string]tally)
	for _, record := range records {
		if record.Key != key {
			continue
		}
		entry := values[record.Value]
		entry.count++
		if record.RecordedAt > entry.recordedAt {
			entry.recordedAt = record.RecordedAt
		}
		values[record.Value] = entry
	}
	bestValue := ""
	bestCount := 0
	bestRecordedAt := ""
	for value, entry := range values {
		if entry.count < 2 {
			continue
		}
		if entry.count > bestCount || (entry.count == bestCount && entry.recordedAt > bestRecordedAt) {
			bestValue = value
			bestCount = entry.count
			bestRecordedAt = entry.recordedAt
		}
	}
	if bestValue == "" {
		return "", false
	}
	return bestValue, true
}

func observationCount(records []PreferenceObservation, key string, value string) int {
	count := 0
	for _, record := range records {
		if record.Key == key && record.Value == value {
			count++
		}
	}
	return count
}

func loadExistingMemoryParagraphs(path string) ([]string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	return splitMemoryParagraphs(string(raw)), nil
}

func writePreferenceMemory(path string, paragraphs []string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	body := strings.Join(paragraphs, "\n\n") + "\n"
	return os.WriteFile(path, []byte(body), 0o644)
}

func isManagedPreferenceParagraph(paragraph string) bool {
	return containsAny(
		paragraph,
		"요약 작업은 보통 수식으로 남긴다.",
		"요약 결과는 값으로 남긴다.",
		"엑셀 작업은 항상 원본을 보존한다.",
		"엑셀 작업 결과는 새 파일로 저장한다.",
		"엑셀 작업은 항상 원본을 보존하고 새 파일로 저장한다.",
	)
}

func preferenceMemoryPath(workspaceRoot string) string {
	return filepath.Join(filepath.Clean(workspaceRoot), ".sheet-ops-state", "knowledge", "request-compiler", "MEMORY.md")
}

func preferenceObservationPath(workspaceRoot string) string {
	return filepath.Join(filepath.Clean(workspaceRoot), ".sheet-ops-state", "knowledge", "request-compiler", "preference-observations.jsonl")
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
