package inspect

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xuri/excelize/v2"
)

func TestInspectWorkbookFactsIncludesStableSheetMap(t *testing.T) {
	inputFile := filepath.Join(t.TempDir(), "map.xlsx")
	file := excelize.NewFile()
	defer func() { _ = file.Close() }()

	defaultSheet := file.GetSheetName(0)
	if err := file.SetSheetName(defaultSheet, "Orders"); err != nil {
		t.Fatalf("SetSheetName: %v", err)
	}
	rows := [][]any{
		{"order_id", "region", "amount"},
		{"A001", "West", 120},
		{"A002", "East", 75},
	}
	for rowIndex, row := range rows {
		for colIndex, value := range row {
			cell, err := excelize.CoordinatesToCellName(colIndex+1, rowIndex+1)
			if err != nil {
				t.Fatalf("CoordinatesToCellName: %v", err)
			}
			if err := file.SetCellValue("Orders", cell, value); err != nil {
				t.Fatalf("SetCellValue(%s): %v", cell, err)
			}
		}
	}
	if err := file.SetCellValue("Orders", "E5", "note"); err != nil {
		t.Fatalf("SetCellValue(E5): %v", err)
	}
	if err := file.SaveAs(inputFile); err != nil {
		t.Fatalf("SaveAs: %v", err)
	}

	facts, err := InspectWorkbookFacts(inputFile)
	if err != nil {
		t.Fatalf("InspectWorkbookFacts: %v", err)
	}
	if len(facts.Sheets) != 1 {
		t.Fatalf("sheet count=%d want 1", len(facts.Sheets))
	}
	sheet := facts.Sheets[0]
	if sheet.Name != "Orders" {
		t.Fatalf("sheet name=%q want Orders", sheet.Name)
	}
	if sheet.RowCount != 5 {
		t.Fatalf("row_count=%d want 5", sheet.RowCount)
	}
	if sheet.ColumnCount != 5 {
		t.Fatalf("column_count=%d want 5", sheet.ColumnCount)
	}
	if sheet.UsedRange.StartCell != "A1" || sheet.UsedRange.EndCell != "E5" {
		t.Fatalf("used_range=%+v want A1:E5", sheet.UsedRange)
	}
}

func TestInspectWorkbookFactsAllowsBlankSheet(t *testing.T) {
	inputFile := filepath.Join(t.TempDir(), "blank.xlsx")
	file := excelize.NewFile()
	defer func() { _ = file.Close() }()
	defaultSheet := file.GetSheetName(0)
	if err := file.SetSheetName(defaultSheet, "Blank"); err != nil {
		t.Fatalf("SetSheetName: %v", err)
	}
	if err := file.SaveAs(inputFile); err != nil {
		t.Fatalf("SaveAs: %v", err)
	}

	facts, err := InspectWorkbookFacts(inputFile)
	if err != nil {
		t.Fatalf("InspectWorkbookFacts: %v", err)
	}
	if len(facts.Sheets) != 1 {
		t.Fatalf("sheet count=%d want 1", len(facts.Sheets))
	}
	sheet := facts.Sheets[0]
	if sheet.RowCount != 0 || sheet.ColumnCount != 0 {
		t.Fatalf("blank dimensions row=%d column=%d want 0,0", sheet.RowCount, sheet.ColumnCount)
	}
	if sheet.UsedRange != nil {
		t.Fatalf("blank used range=%+v want nil", sheet.UsedRange)
	}
}

func TestInspectWorkbookFactsIncludesStructuralWorkbookMap(t *testing.T) {
	inputFile := filepath.Join(t.TempDir(), "structural.xlsx")
	file := excelize.NewFile()
	defer func() { _ = file.Close() }()

	defaultSheet := file.GetSheetName(0)
	if err := file.SetSheetName(defaultSheet, "Orders"); err != nil {
		t.Fatalf("SetSheetName: %v", err)
	}
	rows := [][]any{
		{"order_id", "region", "amount", "double_amount"},
		{"A001", "West", 120, nil},
		{"A002", "East", 75, nil},
	}
	for rowIndex, row := range rows {
		for colIndex, value := range row {
			if value == nil {
				continue
			}
			cell, err := excelize.CoordinatesToCellName(colIndex+1, rowIndex+1)
			if err != nil {
				t.Fatalf("CoordinatesToCellName: %v", err)
			}
			if err := file.SetCellValue("Orders", cell, value); err != nil {
				t.Fatalf("SetCellValue(%s): %v", cell, err)
			}
		}
	}
	if err := file.SetCellFormula("Orders", "D2", "=C2*2"); err != nil {
		t.Fatalf("SetCellFormula(D2): %v", err)
	}
	if err := file.SetCellFormula("Orders", "D3", "=SUM(C2:C3)"); err != nil {
		t.Fatalf("SetCellFormula(D3): %v", err)
	}
	if err := file.AddTable("Orders", &excelize.Table{Name: "OrdersTable", Range: "A1:D3", StyleName: "TableStyleMedium2"}); err != nil {
		t.Fatalf("AddTable: %v", err)
	}
	if err := file.SetDefinedName(&excelize.DefinedName{Name: "OrdersAmount", RefersTo: "Orders!$C$2:$C$3"}); err != nil {
		t.Fatalf("SetDefinedName: %v", err)
	}
	if err := file.SetCellValue("Orders", "F1", "Merged heading"); err != nil {
		t.Fatalf("SetCellValue(F1): %v", err)
	}
	if err := file.MergeCell("Orders", "F1", "G1"); err != nil {
		t.Fatalf("MergeCell: %v", err)
	}
	validation := excelize.NewDataValidation(true)
	validation.SetSqref("B2:B3")
	if err := validation.SetDropList([]string{"West", "East"}); err != nil {
		t.Fatalf("SetDropList: %v", err)
	}
	if err := file.AddDataValidation("Orders", validation); err != nil {
		t.Fatalf("AddDataValidation: %v", err)
	}
	style, err := file.NewConditionalStyle(&excelize.Style{Fill: excelize.Fill{Type: "pattern", Color: []string{"FEC7CE"}, Pattern: 1}})
	if err != nil {
		t.Fatalf("NewConditionalStyle: %v", err)
	}
	if err := file.SetConditionalFormat("Orders", "C2:C3", []excelize.ConditionalFormatOptions{{
		Type:     "cell",
		Format:   &style,
		Criteria: "greater than",
		Value:    "100",
	}}); err != nil {
		t.Fatalf("SetConditionalFormat: %v", err)
	}
	if err := file.SetRowVisible("Orders", 3, false); err != nil {
		t.Fatalf("SetRowVisible: %v", err)
	}
	if err := file.SetColVisible("Orders", "F", false); err != nil {
		t.Fatalf("SetColVisible: %v", err)
	}
	if _, err := file.NewSheet("Archive"); err != nil {
		t.Fatalf("NewSheet: %v", err)
	}
	if err := file.SetSheetVisible("Archive", false); err != nil {
		t.Fatalf("SetSheetVisible: %v", err)
	}
	if err := file.SaveAs(inputFile); err != nil {
		t.Fatalf("SaveAs: %v", err)
	}

	facts, err := InspectWorkbookFacts(inputFile)
	if err != nil {
		t.Fatalf("InspectWorkbookFacts: %v", err)
	}
	orders := facts.Sheets[0]
	if orders.Hidden {
		t.Fatalf("Orders hidden=true want false")
	}
	if len(orders.Formulas) != 2 {
		t.Fatalf("formula count=%d want 2: %+v", len(orders.Formulas), orders.Formulas)
	}
	if orders.Formulas[0].Cell != "D2" || orders.Formulas[0].Formula != "=C2*2" {
		t.Fatalf("first formula=%+v want D2 =C2*2", orders.Formulas[0])
	}
	if len(orders.Tables) != 1 || orders.Tables[0].Name != "OrdersTable" || orders.Tables[0].Range != "A1:D3" {
		t.Fatalf("tables=%+v want OrdersTable A1:D3", orders.Tables)
	}
	if len(orders.DefinedNames) != 1 || orders.DefinedNames[0].Name != "OrdersAmount" || orders.DefinedNames[0].RefersTo != "Orders!$C$2:$C$3" {
		t.Fatalf("defined_names=%+v want OrdersAmount", orders.DefinedNames)
	}
	if len(orders.MergedCells) != 1 || orders.MergedCells[0].Range != "F1:G1" {
		t.Fatalf("merged_cells=%+v want F1:G1", orders.MergedCells)
	}
	if len(orders.Validations) != 1 || orders.Validations[0].Range != "B2:B3" || orders.Validations[0].Type != "list" {
		t.Fatalf("validations=%+v want list on B2:B3", orders.Validations)
	}
	if len(orders.ConditionalFormats) != 1 || orders.ConditionalFormats[0].Range != "C2:C3" || len(orders.ConditionalFormats[0].Rules) != 1 {
		t.Fatalf("conditional_formats=%+v want one rule on C2:C3", orders.ConditionalFormats)
	}
	if orders.ConditionalFormats[0].Rules[0].Type != "cell" || orders.ConditionalFormats[0].Rules[0].Value != "100" {
		t.Fatalf("conditional format rule=%+v want cell > 100", orders.ConditionalFormats[0].Rules[0])
	}
	if len(orders.HiddenRows) != 1 || orders.HiddenRows[0] != 3 {
		t.Fatalf("hidden_rows=%+v want [3]", orders.HiddenRows)
	}
	if len(orders.HiddenColumns) != 1 || orders.HiddenColumns[0] != "F" {
		t.Fatalf("hidden_columns=%+v want [F]", orders.HiddenColumns)
	}
	archive := facts.Sheets[1]
	if archive.Name != "Archive" || !archive.Hidden {
		t.Fatalf("archive=%+v want hidden Archive sheet", archive)
	}
}

func TestInspectWorkbookFactsRejectsOversizedWorkbookFile(t *testing.T) {
	originalLimits := limits
	limits.MaxWorkbookBytes = 4
	defer func() { limits = originalLimits }()

	inputFile := filepath.Join(t.TempDir(), "oversized.xlsx")
	if err := os.WriteFile(inputFile, []byte("12345"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	_, err := InspectWorkbookFacts(inputFile)
	if err == nil {
		t.Fatalf("InspectWorkbookFacts error=nil want workbook size limit")
	}
	if !strings.Contains(err.Error(), "workbook exceeds size limit") {
		t.Fatalf("InspectWorkbookFacts error=%q want size limit", err.Error())
	}
}

func TestInspectWorkbookFactsRejectsTooManySheets(t *testing.T) {
	originalLimits := limits
	limits.MaxSheets = 1
	defer func() { limits = originalLimits }()

	inputFile := filepath.Join(t.TempDir(), "too-many-sheets.xlsx")
	file := excelize.NewFile()
	defer func() { _ = file.Close() }()
	if _, err := file.NewSheet("Extra"); err != nil {
		t.Fatalf("NewSheet: %v", err)
	}
	if err := file.SaveAs(inputFile); err != nil {
		t.Fatalf("SaveAs: %v", err)
	}

	_, err := InspectWorkbookFacts(inputFile)
	if err == nil {
		t.Fatalf("InspectWorkbookFacts error=nil want sheet count limit")
	}
	if !strings.Contains(err.Error(), "workbook exceeds sheet limit") {
		t.Fatalf("InspectWorkbookFacts error=%q want sheet limit", err.Error())
	}
}

func TestInspectWorkbookFactsRejectsTooManyColumns(t *testing.T) {
	originalLimits := limits
	limits.MaxColumnsPerSheet = 1
	defer func() { limits = originalLimits }()

	inputFile := filepath.Join(t.TempDir(), "too-many-columns.xlsx")
	file := excelize.NewFile()
	defer func() { _ = file.Close() }()
	if err := file.SetCellValue("Sheet1", "B1", "overflow"); err != nil {
		t.Fatalf("SetCellValue: %v", err)
	}
	if err := file.SaveAs(inputFile); err != nil {
		t.Fatalf("SaveAs: %v", err)
	}

	_, err := InspectWorkbookFacts(inputFile)
	if err == nil {
		t.Fatalf("InspectWorkbookFacts error=nil want column count limit")
	}
	if !strings.Contains(err.Error(), "sheet Sheet1 exceeds column limit") {
		t.Fatalf("InspectWorkbookFacts error=%q want column limit", err.Error())
	}
}
