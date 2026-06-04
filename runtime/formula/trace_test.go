package formula

import (
	"errors"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/xuri/excelize/v2"
)

func TestTraceWorkbookCellExtractsSupportedFormulaReferences(t *testing.T) {
	inputFile := filepath.Join(t.TempDir(), "trace.xlsx")
	file := excelize.NewFile()
	defer func() { _ = file.Close() }()

	defaultSheet := file.GetSheetName(0)
	if err := file.SetSheetName(defaultSheet, "Orders"); err != nil {
		t.Fatalf("SetSheetName: %v", err)
	}
	if _, err := file.NewSheet("Summary"); err != nil {
		t.Fatalf("NewSheet: %v", err)
	}
	values := map[string]any{
		"Orders!A1":  "region",
		"Orders!D1":  "amount",
		"Orders!A2":  "West",
		"Orders!D2":  120,
		"Summary!A5": "West",
	}
	for cell, value := range values {
		sheet, axis, err := splitTestCell(cell)
		if err != nil {
			t.Fatalf("splitTestCell(%s): %v", cell, err)
		}
		if err := file.SetCellValue(sheet, axis, value); err != nil {
			t.Fatalf("SetCellValue(%s): %v", cell, err)
		}
	}
	if err := file.SetCellFormula("Summary", "D5", "=SUMIFS(Orders!D:D,Orders!A:A,A5)"); err != nil {
		t.Fatalf("SetCellFormula: %v", err)
	}
	if err := file.SaveAs(inputFile); err != nil {
		t.Fatalf("SaveAs: %v", err)
	}

	trace, err := TraceWorkbookCell(inputFile, "Summary!D5")
	if err != nil {
		t.Fatalf("TraceWorkbookCell: %v", err)
	}
	if trace.Cell != "Summary!D5" {
		t.Fatalf("cell=%q want Summary!D5", trace.Cell)
	}
	if trace.Formula != "=SUMIFS(Orders!D:D,Orders!A:A,A5)" {
		t.Fatalf("formula=%q", trace.Formula)
	}
	wantReferences := []string{"Orders!D:D", "Orders!A:A", "Summary!A5"}
	if !reflect.DeepEqual(trace.References, wantReferences) {
		t.Fatalf("references=%v want %v", trace.References, wantReferences)
	}
	if !reflect.DeepEqual(trace.Functions, []string{"SUMIFS"}) {
		t.Fatalf("functions=%v want [SUMIFS]", trace.Functions)
	}
	if trace.Status != "supported" {
		t.Fatalf("status=%q want supported", trace.Status)
	}
}

func TestTraceFormulaReferencesSupportedLookupAndIndexMatchFamilies(t *testing.T) {
	tests := []struct {
		name       string
		sheet      string
		formula    string
		references []string
		functions  []string
	}{
		{
			name:       "sum",
			sheet:      "Summary",
			formula:    "=SUM(Orders!D2:D10)",
			references: []string{"Orders!D2:D10"},
			functions:  []string{"SUM"},
		},
		{
			name:       "countifs",
			sheet:      "Summary",
			formula:    "=COUNTIFS(Orders!A:A,A5,Orders!D:D,\">0\")",
			references: []string{"Orders!A:A", "Summary!A5", "Orders!D:D"},
			functions:  []string{"COUNTIFS"},
		},
		{
			name:       "xlookup",
			sheet:      "Summary",
			formula:    "=XLOOKUP(A5,Customers!A:A,Customers!C:C)",
			references: []string{"Summary!A5", "Customers!A:A", "Customers!C:C"},
			functions:  []string{"XLOOKUP"},
		},
		{
			name:       "vlookup",
			sheet:      "Summary",
			formula:    "=VLOOKUP(A5,Customers!A:D,4,FALSE)",
			references: []string{"Summary!A5", "Customers!A:D"},
			functions:  []string{"VLOOKUP"},
		},
		{
			name:       "index match",
			sheet:      "Summary",
			formula:    "=INDEX(Customers!C:C,MATCH(A5,Customers!A:A,0))",
			references: []string{"Customers!C:C", "Summary!A5", "Customers!A:A"},
			functions:  []string{"INDEX", "MATCH"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			trace := TraceFormula(tc.sheet, "D5", tc.formula)
			if !reflect.DeepEqual(trace.References, tc.references) {
				t.Fatalf("references=%v want %v", trace.References, tc.references)
			}
			if !reflect.DeepEqual(trace.Functions, tc.functions) {
				t.Fatalf("functions=%v want %v", trace.Functions, tc.functions)
			}
			if trace.Status != "supported" {
				t.Fatalf("status=%q want supported", trace.Status)
			}
		})
	}
}

func TestTraceFormulaReportsUnsupportedFunctionsWithReferences(t *testing.T) {
	trace := TraceFormula("Summary", "D5", `=FILTER(Orders!A:A,Orders!D:D>100)`)
	if trace.Status != "unsupported" {
		t.Fatalf("status=%q want unsupported", trace.Status)
	}
	if !reflect.DeepEqual(trace.UnsupportedFunctions, []string{"FILTER"}) {
		t.Fatalf("unsupported_functions=%v want [FILTER]", trace.UnsupportedFunctions)
	}
	wantReferences := []string{"Orders!A:A", "Orders!D:D"}
	if !reflect.DeepEqual(trace.References, wantReferences) {
		t.Fatalf("references=%v want %v", trace.References, wantReferences)
	}
}

func TestTraceFormulaDoesNotTreatScientificNotationAsCellReference(t *testing.T) {
	trace := TraceFormula("Summary", "D5", "=SUM(1E3,A1)")
	wantReferences := []string{"Summary!A1"}
	if !reflect.DeepEqual(trace.References, wantReferences) {
		t.Fatalf("references=%v want %v", trace.References, wantReferences)
	}
}

func TestTraceFormulaReportsStructuredReferencesAsUnsupportedSyntax(t *testing.T) {
	trace := TraceFormula("Summary", "D5", "=SUM(Table1[Amount])")
	if trace.Status != "unsupported" {
		t.Fatalf("status=%q want unsupported", trace.Status)
	}
	if !reflect.DeepEqual(trace.UnsupportedSyntax, []string{"structured_reference"}) {
		t.Fatalf("unsupported_syntax=%v want [structured_reference]", trace.UnsupportedSyntax)
	}
}

func splitTestCell(ref string) (string, string, error) {
	parts := strings.SplitN(ref, "!", 2)
	if len(parts) != 2 {
		return "", "", errors.New("invalid test cell")
	}
	return parts[0], parts[1], nil
}
