package execute

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mwroh/sheet-ops/runtime/compiler"
	"github.com/xuri/excelize/v2"
)

func RunNormalizeHeaders(ir compiler.WorkbookOperationIR, inputWorkbook, outputWorkbook string) (ExecutionResult, error) {
	if err := validateNormalizeHeadersBoundary(ir, inputWorkbook, outputWorkbook); err != nil {
		return ExecutionResult{}, err
	}
	sourceBefore, err := hashFile(inputWorkbook)
	if err != nil {
		return ExecutionResult{}, err
	}

	file, err := excelize.OpenFile(inputWorkbook)
	if err != nil {
		return ExecutionResult{}, err
	}
	defer func() { _ = file.Close() }()

	rows, err := file.GetRows(ir.SourceSheet)
	if err != nil {
		return ExecutionResult{}, fmt.Errorf("read source sheet %q: %w", ir.SourceSheet, err)
	}
	if len(rows) < ir.HeaderRow {
		return ExecutionResult{}, fmt.Errorf("header row %d is outside sheet %q", ir.HeaderRow, ir.SourceSheet)
	}
	headerIndex := stringHeaderIndex(rows[ir.HeaderRow-1])
	writtenCells := make([]string, 0, len(ir.HeaderMappings))
	for _, mapping := range ir.HeaderMappings {
		columnIndex, ok := headerIndex[mapping.From]
		if !ok {
			return ExecutionResult{}, fmt.Errorf("header %q missing from row %d", mapping.From, ir.HeaderRow)
		}
		cell, _ := excelize.CoordinatesToCellName(columnIndex+1, ir.HeaderRow)
		if err := file.SetCellValue(ir.SourceSheet, cell, mapping.To); err != nil {
			return ExecutionResult{}, err
		}
		writtenCells = append(writtenCells, ir.SourceSheet+"!"+cell)
	}

	if outputDir := filepath.Dir(outputWorkbook); outputDir != "." {
		if err := os.MkdirAll(outputDir, 0o755); err != nil {
			return ExecutionResult{}, err
		}
	}
	if err := file.SaveAs(outputWorkbook); err != nil {
		return ExecutionResult{}, err
	}
	sourceAfter, err := hashFile(inputWorkbook)
	if err != nil {
		return ExecutionResult{}, err
	}

	return ExecutionResult{
		OperationFamily:    ir.OperationFamily,
		OutputWorkbook:     outputWorkbook,
		SummarySheet:       ir.SourceSheet,
		WrittenCells:       writtenCells,
		SourceSHA256Before: sourceBefore,
		SourceSHA256After:  sourceAfter,
	}, nil
}

func validateNormalizeHeadersBoundary(ir compiler.WorkbookOperationIR, inputWorkbook, outputWorkbook string) error {
	if inputWorkbook == "" {
		return fmt.Errorf("input workbook must not be empty")
	}
	if outputWorkbook == "" {
		return fmt.Errorf("output workbook must not be empty")
	}
	if !ir.PreserveOriginal {
		return fmt.Errorf("normalize_headers execution requires preserve_original=true")
	}
	if ir.SourceSheet == "" {
		return fmt.Errorf("source sheet must not be empty")
	}
	if ir.HeaderRow < 1 {
		return fmt.Errorf("header row must be positive")
	}
	if len(ir.HeaderMappings) == 0 {
		return fmt.Errorf("header mappings must not be empty")
	}
	seenFrom := map[string]struct{}{}
	seenTo := map[string]struct{}{}
	for _, mapping := range ir.HeaderMappings {
		if mapping.From == "" {
			return fmt.Errorf("header mapping from must not be empty")
		}
		if mapping.To == "" {
			return fmt.Errorf("header mapping to must not be empty")
		}
		if _, ok := seenFrom[mapping.From]; ok {
			return fmt.Errorf("duplicate header mapping from %q", mapping.From)
		}
		if _, ok := seenTo[mapping.To]; ok {
			return fmt.Errorf("duplicate header mapping to %q", mapping.To)
		}
		seenFrom[mapping.From] = struct{}{}
		seenTo[mapping.To] = struct{}{}
	}
	same, err := sameWorkbookPath(inputWorkbook, outputWorkbook)
	if err != nil {
		return err
	}
	if same {
		return fmt.Errorf("output workbook must differ from input workbook")
	}
	return nil
}
