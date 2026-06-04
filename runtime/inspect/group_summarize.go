package inspect

import (
	"fmt"
	"strings"

	"github.com/mwroh/sheet-ops/runtime/taskspec"
	"github.com/xuri/excelize/v2"
)

type GroupSummarizeInspection struct {
	InputWorkbook string
	SheetNames    []string
	SourceSheet   string
	HeaderNames   []string
	HeaderIndex   map[string]int
	RowCount      int
}

func GroupSummarizeWorkbook(inputWorkbook, sourceSheet string, filters []taskspec.FilterSpec, groupBy []string, metrics []taskspec.MetricSpec) (GroupSummarizeInspection, error) {
	file, err := excelize.OpenFile(inputWorkbook)
	if err != nil {
		return GroupSummarizeInspection{}, err
	}
	defer func() { _ = file.Close() }()

	sheets := file.GetSheetList()
	if !containsString(sheets, sourceSheet) {
		return GroupSummarizeInspection{}, fmt.Errorf("sheet %q not found", sourceSheet)
	}
	rows, err := file.GetRows(sourceSheet)
	if err != nil {
		return GroupSummarizeInspection{}, err
	}
	if len(rows) == 0 {
		return GroupSummarizeInspection{}, fmt.Errorf("sheet %q is empty", sourceSheet)
	}

	headers := rows[0]
	headerIndex := buildHeaderIndex(headers)
	missing := []string{}
	for _, filter := range filters {
		if _, ok := headerIndex[filter.Column]; !ok && !containsString(missing, filter.Column) {
			missing = append(missing, filter.Column)
		}
	}
	for _, column := range groupBy {
		if _, ok := headerIndex[column]; !ok && !containsString(missing, column) {
			missing = append(missing, column)
		}
	}
	for _, metric := range metrics {
		if _, ok := headerIndex[metric.Column]; !ok && !containsString(missing, metric.Column) {
			missing = append(missing, metric.Column)
		}
	}
	if len(missing) > 0 {
		return GroupSummarizeInspection{}, fmt.Errorf("missing required summary columns: %s", strings.Join(missing, ", "))
	}

	return GroupSummarizeInspection{
		InputWorkbook: inputWorkbook,
		SheetNames:    append([]string(nil), sheets...),
		SourceSheet:   sourceSheet,
		HeaderNames:   append([]string(nil), headers...),
		HeaderIndex:   cloneHeaderIndex(headerIndex),
		RowCount:      len(rows),
	}, nil
}

func buildHeaderIndex(headers []string) map[string]int {
	index := make(map[string]int, len(headers))
	for idx, header := range headers {
		trimmed := strings.TrimSpace(header)
		if trimmed == "" {
			continue
		}
		index[trimmed] = idx
	}
	return index
}

func cloneHeaderIndex(values map[string]int) map[string]int {
	if len(values) == 0 {
		return map[string]int{}
	}
	cloned := make(map[string]int, len(values))
	for key, value := range values {
		cloned[key] = value
	}
	return cloned
}

func containsString(values []string, needle string) bool {
	for _, value := range values {
		if value == needle {
			return true
		}
	}
	return false
}
