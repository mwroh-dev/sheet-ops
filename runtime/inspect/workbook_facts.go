package inspect

import (
	"fmt"
	"os"
	"path/filepath"
	goruntime "runtime"
	"sort"
	"strings"

	runtimeschema "github.com/mwroh/sheet-ops/runtime/schema"
	"github.com/xuri/excelize/v2"
)

type resourceLimits struct {
	MaxWorkbookBytes    int64
	MaxSheets           int
	MaxRowsPerSheet     int
	MaxColumnsPerSheet  int
	MaxFormulaCellScans int
}

var limits = resourceLimits{
	MaxWorkbookBytes:    50 * 1024 * 1024,
	MaxSheets:           50,
	MaxRowsPerSheet:     100000,
	MaxColumnsPerSheet:  1024,
	MaxFormulaCellScans: 250000,
}

type WorkbookFacts struct {
	InputFile string       `json:"input_file"`
	Sheets    []SheetFacts `json:"sheets"`
}

type SheetFacts struct {
	Name               string                  `json:"name"`
	Hidden             bool                    `json:"hidden,omitempty"`
	RowCount           int                     `json:"row_count,omitempty"`
	ColumnCount        int                     `json:"column_count,omitempty"`
	UsedRange          *UsedRange              `json:"used_range,omitempty"`
	Columns            []string                `json:"columns,omitempty"`
	Formulas           []FormulaFact           `json:"formulas,omitempty"`
	Tables             []TableFact             `json:"tables,omitempty"`
	DefinedNames       []DefinedNameFact       `json:"defined_names,omitempty"`
	MergedCells        []MergedCellFact        `json:"merged_cells,omitempty"`
	Validations        []ValidationFact        `json:"validations,omitempty"`
	ConditionalFormats []ConditionalFormatFact `json:"conditional_formats,omitempty"`
	HiddenRows         []int                   `json:"hidden_rows,omitempty"`
	HiddenColumns      []string                `json:"hidden_columns,omitempty"`
}

type UsedRange struct {
	StartCell string `json:"start_cell"`
	EndCell   string `json:"end_cell"`
}

type FormulaFact struct {
	Cell    string `json:"cell"`
	Formula string `json:"formula"`
}

type TableFact struct {
	Name  string `json:"name"`
	Range string `json:"range"`
}

type DefinedNameFact struct {
	Name     string `json:"name"`
	RefersTo string `json:"refers_to"`
	Scope    string `json:"scope,omitempty"`
}

type MergedCellFact struct {
	Range string `json:"range"`
}

type ValidationFact struct {
	Range    string `json:"range"`
	Type     string `json:"type,omitempty"`
	Operator string `json:"operator,omitempty"`
	Formula1 string `json:"formula1,omitempty"`
	Formula2 string `json:"formula2,omitempty"`
}

type ConditionalFormatFact struct {
	Range string                      `json:"range"`
	Rules []ConditionalFormatRuleFact `json:"rules,omitempty"`
}

type ConditionalFormatRuleFact struct {
	Type     string `json:"type,omitempty"`
	Criteria string `json:"criteria,omitempty"`
	Value    string `json:"value,omitempty"`
}

func InspectWorkbookFacts(inputFile string) (WorkbookFacts, error) {
	if err := checkWorkbookFileSize(inputFile); err != nil {
		return WorkbookFacts{}, err
	}
	file, err := excelize.OpenFile(inputFile)
	if err != nil {
		return WorkbookFacts{}, err
	}
	defer func() { _ = file.Close() }()

	sheets := file.GetSheetList()
	if len(sheets) > limits.MaxSheets {
		return WorkbookFacts{}, fmt.Errorf("workbook exceeds sheet limit: %d > %d", len(sheets), limits.MaxSheets)
	}
	definedNames := file.GetDefinedName()
	facts := WorkbookFacts{
		InputFile: inputFile,
		Sheets:    make([]SheetFacts, 0, len(sheets)),
	}

	for _, name := range sheets {
		rows, err := file.GetRows(name)
		if err != nil {
			return WorkbookFacts{}, err
		}

		sheetFacts := SheetFacts{Name: name}
		visible, err := file.GetSheetVisible(name)
		if err != nil {
			return WorkbookFacts{}, err
		}
		sheetFacts.Hidden = !visible
		if len(rows) > 0 {
			sheetFacts.Columns = compactColumns(rows[0])
			sheetFacts.RowCount, sheetFacts.ColumnCount = sheetDimensions(rows)
			if err := checkSheetDimensions(name, sheetFacts.RowCount, sheetFacts.ColumnCount); err != nil {
				return WorkbookFacts{}, err
			}
			if sheetFacts.RowCount > 0 && sheetFacts.ColumnCount > 0 {
				endCell, err := excelize.CoordinatesToCellName(sheetFacts.ColumnCount, sheetFacts.RowCount)
				if err != nil {
					return WorkbookFacts{}, err
				}
				sheetFacts.UsedRange = &UsedRange{StartCell: "A1", EndCell: endCell}
			}
		}
		if err := enrichStructuralFacts(file, name, definedNames, &sheetFacts); err != nil {
			return WorkbookFacts{}, err
		}
		facts.Sheets = append(facts.Sheets, sheetFacts)
	}

	if err := runtimeschema.ValidateStruct(workbookFactsSchemaPath(), facts); err != nil {
		return WorkbookFacts{}, err
	}
	return facts, nil
}

func checkWorkbookFileSize(inputFile string) error {
	return checkWorkbookFileSizeWithin(inputFile, limits.MaxWorkbookBytes)
}

func checkWorkbookFileSizeWithin(inputFile string, maxWorkbookBytes int64) error {
	info, err := os.Stat(inputFile)
	if err != nil {
		return err
	}
	if info.Size() > maxWorkbookBytes {
		return fmt.Errorf("workbook exceeds size limit: %d bytes > %d bytes", info.Size(), maxWorkbookBytes)
	}
	return nil
}

func checkSheetDimensions(sheet string, rows, columns int) error {
	if rows > limits.MaxRowsPerSheet {
		return fmt.Errorf("sheet %s exceeds row limit: %d > %d", sheet, rows, limits.MaxRowsPerSheet)
	}
	if columns > limits.MaxColumnsPerSheet {
		return fmt.Errorf("sheet %s exceeds column limit: %d > %d", sheet, columns, limits.MaxColumnsPerSheet)
	}
	if rows > 0 && columns > 0 && rows*columns > limits.MaxFormulaCellScans {
		return fmt.Errorf("sheet %s exceeds formula inspection cell limit: %d > %d", sheet, rows*columns, limits.MaxFormulaCellScans)
	}
	return nil
}

func enrichStructuralFacts(file *excelize.File, sheet string, definedNames []excelize.DefinedName, facts *SheetFacts) error {
	formulas, err := inspectFormulas(file, sheet, facts.RowCount, facts.ColumnCount)
	if err != nil {
		return err
	}
	facts.Formulas = formulas

	tables, err := file.GetTables(sheet)
	if err != nil {
		return err
	}
	facts.Tables = inspectTables(tables)
	facts.DefinedNames = inspectDefinedNames(sheet, definedNames)

	mergedCells, err := file.GetMergeCells(sheet, true)
	if err != nil {
		return err
	}
	facts.MergedCells = inspectMergedCells(mergedCells)

	validations, err := file.GetDataValidations(sheet)
	if err != nil {
		return err
	}
	facts.Validations = inspectValidations(validations)

	conditionalFormats, err := file.GetConditionalFormats(sheet)
	if err != nil {
		return err
	}
	facts.ConditionalFormats = inspectConditionalFormats(conditionalFormats)

	facts.HiddenRows, err = inspectHiddenRows(file, sheet, facts.RowCount)
	if err != nil {
		return err
	}
	facts.HiddenColumns, err = inspectHiddenColumns(file, sheet, facts.ColumnCount)
	if err != nil {
		return err
	}
	return nil
}

func inspectFormulas(file *excelize.File, sheet string, rowCount, columnCount int) ([]FormulaFact, error) {
	formulas := []FormulaFact{}
	for row := 1; row <= rowCount; row++ {
		for column := 1; column <= columnCount; column++ {
			cell, err := excelize.CoordinatesToCellName(column, row)
			if err != nil {
				return nil, err
			}
			formula, err := file.GetCellFormula(sheet, cell)
			if err != nil {
				return nil, err
			}
			if formula == "" {
				continue
			}
			formulas = append(formulas, FormulaFact{Cell: cell, Formula: normalizeFormula(formula)})
		}
	}
	if len(formulas) == 0 {
		return nil, nil
	}
	return formulas, nil
}

func inspectTables(tables []excelize.Table) []TableFact {
	if len(tables) == 0 {
		return nil
	}
	facts := make([]TableFact, 0, len(tables))
	for _, table := range tables {
		facts = append(facts, TableFact{Name: table.Name, Range: table.Range})
	}
	sort.Slice(facts, func(i, j int) bool {
		if facts[i].Name == facts[j].Name {
			return facts[i].Range < facts[j].Range
		}
		return facts[i].Name < facts[j].Name
	})
	return facts
}

func inspectDefinedNames(sheet string, definedNames []excelize.DefinedName) []DefinedNameFact {
	facts := []DefinedNameFact{}
	for _, name := range definedNames {
		if name.Scope != "" && name.Scope != "Workbook" && name.Scope != sheet {
			continue
		}
		if (name.Scope == "" || name.Scope == "Workbook") && !refersToSheet(name.RefersTo, sheet) {
			continue
		}
		facts = append(facts, DefinedNameFact{Name: name.Name, RefersTo: name.RefersTo, Scope: name.Scope})
	}
	if len(facts) == 0 {
		return nil
	}
	sort.Slice(facts, func(i, j int) bool {
		if facts[i].Name == facts[j].Name {
			return facts[i].RefersTo < facts[j].RefersTo
		}
		return facts[i].Name < facts[j].Name
	})
	return facts
}

func inspectMergedCells(mergedCells []excelize.MergeCell) []MergedCellFact {
	if len(mergedCells) == 0 {
		return nil
	}
	facts := make([]MergedCellFact, 0, len(mergedCells))
	for _, mergedCell := range mergedCells {
		if len(mergedCell) == 0 || mergedCell[0] == "" {
			continue
		}
		facts = append(facts, MergedCellFact{Range: mergedCell[0]})
	}
	if len(facts) == 0 {
		return nil
	}
	sort.Slice(facts, func(i, j int) bool { return facts[i].Range < facts[j].Range })
	return facts
}

func inspectValidations(validations []*excelize.DataValidation) []ValidationFact {
	if len(validations) == 0 {
		return nil
	}
	facts := make([]ValidationFact, 0, len(validations))
	for _, validation := range validations {
		if validation == nil || validation.Sqref == "" {
			continue
		}
		facts = append(facts, ValidationFact{
			Range:    validation.Sqref,
			Type:     validation.Type,
			Operator: validation.Operator,
			Formula1: validation.Formula1,
			Formula2: validation.Formula2,
		})
	}
	if len(facts) == 0 {
		return nil
	}
	sort.Slice(facts, func(i, j int) bool { return facts[i].Range < facts[j].Range })
	return facts
}

func inspectConditionalFormats(formats map[string][]excelize.ConditionalFormatOptions) []ConditionalFormatFact {
	if len(formats) == 0 {
		return nil
	}
	ranges := make([]string, 0, len(formats))
	for rangeRef := range formats {
		ranges = append(ranges, rangeRef)
	}
	sort.Strings(ranges)

	facts := make([]ConditionalFormatFact, 0, len(ranges))
	for _, rangeRef := range ranges {
		rules := make([]ConditionalFormatRuleFact, 0, len(formats[rangeRef]))
		for _, rule := range formats[rangeRef] {
			rules = append(rules, ConditionalFormatRuleFact{
				Type:     rule.Type,
				Criteria: rule.Criteria,
				Value:    rule.Value,
			})
		}
		facts = append(facts, ConditionalFormatFact{Range: rangeRef, Rules: rules})
	}
	return facts
}

func inspectHiddenRows(file *excelize.File, sheet string, rowCount int) ([]int, error) {
	hiddenRows := []int{}
	for row := 1; row <= rowCount; row++ {
		visible, err := file.GetRowVisible(sheet, row)
		if err != nil {
			return nil, err
		}
		if !visible {
			hiddenRows = append(hiddenRows, row)
		}
	}
	if len(hiddenRows) == 0 {
		return nil, nil
	}
	return hiddenRows, nil
}

func inspectHiddenColumns(file *excelize.File, sheet string, columnCount int) ([]string, error) {
	hiddenColumns := []string{}
	for column := 1; column <= columnCount; column++ {
		name, err := excelize.ColumnNumberToName(column)
		if err != nil {
			return nil, err
		}
		visible, err := file.GetColVisible(sheet, name)
		if err != nil {
			return nil, err
		}
		if !visible {
			hiddenColumns = append(hiddenColumns, name)
		}
	}
	if len(hiddenColumns) == 0 {
		return nil, nil
	}
	return hiddenColumns, nil
}

func sheetDimensions(rows [][]string) (int, int) {
	rowCount := 0
	columnCount := 0
	for rowIndex, row := range rows {
		lastColumn := lastNonEmptyColumn(row)
		if lastColumn == 0 {
			continue
		}
		rowCount = rowIndex + 1
		if lastColumn > columnCount {
			columnCount = lastColumn
		}
	}
	return rowCount, columnCount
}

func lastNonEmptyColumn(row []string) int {
	for index := len(row) - 1; index >= 0; index-- {
		if strings.TrimSpace(row[index]) != "" {
			return index + 1
		}
	}
	return 0
}

func compactColumns(values []string) []string {
	if len(values) == 0 {
		return nil
	}

	columns := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		columns = append(columns, trimmed)
	}
	if len(columns) == 0 {
		return nil
	}
	return columns
}

func normalizeFormula(formula string) string {
	if strings.HasPrefix(formula, "=") {
		return formula
	}
	return "=" + formula
}

func refersToSheet(refersTo, sheet string) bool {
	refersTo = strings.TrimPrefix(refersTo, "=")
	prefixes := []string{
		sheet + "!",
		"'" + sheet + "'!",
	}
	for _, prefix := range prefixes {
		if strings.HasPrefix(refersTo, prefix) {
			return true
		}
	}
	return false
}

func workbookFactsSchemaPath() string {
	_, file, _, ok := goruntime.Caller(0)
	if !ok {
		return filepath.Join("contracts", "inspection", "workbook_facts.schema.json")
	}
	return runtimeschema.ResolveRepoPath(file, 2, "contracts", "inspection", "workbook_facts.schema.json")
}
