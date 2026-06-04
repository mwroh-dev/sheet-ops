package inspect

import (
	"fmt"

	"github.com/xuri/excelize/v2"
)

type surfaceProbeLimitSet struct {
	MaxWorkbookBytes int64
	MaxSheets        int
	MaxHeaderColumns int
}

var surfaceProbeLimits = surfaceProbeLimitSet{
	MaxWorkbookBytes: limits.MaxWorkbookBytes,
	MaxSheets:        limits.MaxSheets,
	MaxHeaderColumns: limits.MaxColumnsPerSheet,
}

type WorkbookSurfaceFacts struct {
	VisibleSheetNames []string            `json:"visible_sheet_names"`
	HeaderRowBySheet  map[string][]string `json:"header_row_by_sheet"`
}

func InspectWorkbookSurfaceFacts(inputFile string) (WorkbookSurfaceFacts, error) {
	if err := checkWorkbookFileSizeWithin(inputFile, surfaceProbeLimits.MaxWorkbookBytes); err != nil {
		return WorkbookSurfaceFacts{}, err
	}

	file, err := excelize.OpenFile(inputFile)
	if err != nil {
		return WorkbookSurfaceFacts{}, fmt.Errorf("open input workbook %s: %w", inputFile, err)
	}
	defer func() { _ = file.Close() }()

	sheets := file.GetSheetList()
	if len(sheets) > surfaceProbeLimits.MaxSheets {
		return WorkbookSurfaceFacts{}, fmt.Errorf("workbook exceeds sheet limit: %d > %d", len(sheets), surfaceProbeLimits.MaxSheets)
	}

	headers := make(map[string][]string, len(sheets))
	for _, sheet := range sheets {
		header, err := readSheetHeader(file, sheet)
		if err != nil {
			return WorkbookSurfaceFacts{}, fmt.Errorf("read header row for sheet %s: %w", sheet, err)
		}
		if len(header) > surfaceProbeLimits.MaxHeaderColumns {
			return WorkbookSurfaceFacts{}, fmt.Errorf("sheet %s exceeds header limit: %d > %d", sheet, len(header), surfaceProbeLimits.MaxHeaderColumns)
		}
		headers[sheet] = append([]string(nil), header...)
	}

	return WorkbookSurfaceFacts{
		VisibleSheetNames: append([]string(nil), sheets...),
		HeaderRowBySheet:  headers,
	}, nil
}

func readSheetHeader(file *excelize.File, sheet string) ([]string, error) {
	rows, err := file.Rows(sheet)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	if !rows.Next() {
		if err := rows.Error(); err != nil {
			return nil, err
		}
		return []string{}, nil
	}

	header, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	return append([]string(nil), header...), nil
}
