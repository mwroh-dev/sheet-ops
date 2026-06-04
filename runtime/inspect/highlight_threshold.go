package inspect

import (
	"fmt"
	"strings"

	"github.com/xuri/excelize/v2"
)

type HighlightThresholdInspection struct {
	InputWorkbook     string
	SheetNames        []string
	SourceSheet       string
	HeaderNames       []string
	HeaderIndex       map[string]int
	TargetColumn      string
	TargetColumnIndex int
	HeaderRow         int
	RowCount          int
}

func HighlightThresholdWorkbook(inputWorkbook, sourceSheet, targetColumn string) (HighlightThresholdInspection, error) {
	file, err := excelize.OpenFile(inputWorkbook)
	if err != nil {
		return HighlightThresholdInspection{}, err
	}
	defer func() { _ = file.Close() }()

	sheets := file.GetSheetList()
	if !containsString(sheets, sourceSheet) {
		return HighlightThresholdInspection{}, fmt.Errorf("sheet %q not found", sourceSheet)
	}
	rows, err := file.GetRows(sourceSheet)
	if err != nil {
		return HighlightThresholdInspection{}, err
	}
	if len(rows) == 0 {
		return HighlightThresholdInspection{}, fmt.Errorf("sheet %q is empty", sourceSheet)
	}

	headers := rows[0]
	headerIndex := buildHeaderIndex(headers)
	normalizedTarget := strings.TrimSpace(targetColumn)
	if normalizedTarget == "" {
		return HighlightThresholdInspection{}, fmt.Errorf("target column must not be empty")
	}

	resolvedColumn := ""
	resolvedIndex := -1
	for index, header := range headers {
		if strings.EqualFold(strings.TrimSpace(header), normalizedTarget) {
			resolvedColumn = strings.TrimSpace(header)
			resolvedIndex = index
			break
		}
	}
	if resolvedIndex < 0 {
		return HighlightThresholdInspection{}, fmt.Errorf("%s column not found", normalizedTarget)
	}

	return HighlightThresholdInspection{
		InputWorkbook:     inputWorkbook,
		SheetNames:        append([]string(nil), sheets...),
		SourceSheet:       sourceSheet,
		HeaderNames:       append([]string(nil), headers...),
		HeaderIndex:       cloneHeaderIndex(headerIndex),
		TargetColumn:      resolvedColumn,
		TargetColumnIndex: resolvedIndex + 1,
		HeaderRow:         1,
		RowCount:          len(rows),
	}, nil
}
