package workbookcase

import (
	"path/filepath"
	"testing"

	runtimetaskspec "github.com/mwroh/sheet-ops/runtime/taskspec"
	"github.com/xuri/excelize/v2"
)

func TestInvoiceLineItemBillingPreviewComposesAppendFormulaExtensionValidationAndProtection(t *testing.T) {
	setRuntimeRoots(t)

	tempDir := t.TempDir()
	inputFile := filepath.Join(tempDir, "invoice.xlsx")
	appendOutput := filepath.Join(tempDir, "invoice-appended.xlsx")
	finalOutput := filepath.Join(tempDir, "invoice-final.xlsx")
	validationOutput := filepath.Join(tempDir, "invoice-validated.xlsx")
	protectedOutput := filepath.Join(tempDir, "invoice-protected.xlsx")

	file := excelize.NewFile()
	defer func() { _ = file.Close() }()
	defaultSheet := file.GetSheetName(0)
	if err := file.SetSheetName(defaultSheet, "LineItems"); err != nil {
		t.Fatalf("SetSheetName: %v", err)
	}
	header := []any{"sku", "quantity", "unit_price", "line_total", "tax"}
	if err := file.SetSheetRow("LineItems", "A1", &header); err != nil {
		t.Fatalf("SetSheetRow(header): %v", err)
	}
	row2 := []any{"A001", 2, 10}
	if err := file.SetSheetRow("LineItems", "A2", &row2); err != nil {
		t.Fatalf("SetSheetRow(row2): %v", err)
	}
	if err := file.SetCellFormula("LineItems", "D2", "=B2*C2"); err != nil {
		t.Fatalf("SetCellFormula(D2): %v", err)
	}
	if err := file.SetCellFormula("LineItems", "E2", "=D2*0.1"); err != nil {
		t.Fatalf("SetCellFormula(E2): %v", err)
	}
	if err := file.SaveAs(inputFile); err != nil {
		t.Fatalf("SaveAs: %v", err)
	}

	appendTask := runtimetaskspec.BuildAppendStructuredRowsTask(runtimetaskspec.AppendStructuredRowsRequest{
		RequestText:          "청구서에 새 line item 행을 추가한다.",
		InputFile:            inputFile,
		SourceSheet:          "LineItems",
		OutputFile:           appendOutput,
		IncludeSourceColumns: []string{"sku", "quantity", "unit_price"},
		Values: []runtimetaskspec.CellValue{
			{Cell: "sku", Value: "B002"},
			{Cell: "quantity", Value: 3},
			{Cell: "unit_price", Value: 15},
		},
	})
	appendResult, err := Run(Request{ScenarioID: "invoice-preview-append", TaskSpec: appendTask.TaskSpec})
	if err != nil {
		t.Fatalf("append Run: %v", err)
	}
	if !appendResult.Verification.Pass {
		t.Fatalf("append verification failed: %+v", appendResult.Verification)
	}

	formulaTask := runtimetaskspec.BuildExtendTableFormulasTask(runtimetaskspec.ExtendTableFormulasRequest{
		RequestText:      "새 line item 행에 계산 수식을 확장한다.",
		InputFile:        appendOutput,
		SourceSheet:      "LineItems",
		OutputFile:       finalOutput,
		FormulaSourceRow: 2,
		TargetRows:       []int{3},
		FormulaColumns:   []string{"D", "E"},
	})
	formulaResult, err := Run(Request{ScenarioID: "invoice-preview-formulas", TaskSpec: formulaTask.TaskSpec})
	if err != nil {
		t.Fatalf("formula Run: %v", err)
	}
	if !formulaResult.Verification.Pass {
		t.Fatalf("formula verification failed: %+v", formulaResult.Verification)
	}

	validationTask := runtimetaskspec.BuildAddDataValidationTask(runtimetaskspec.AddDataValidationRequest{
		RequestText: "청구서 line item SKU 입력 범위를 허용된 SKU dropdown으로 제한한다.",
		InputFile:   finalOutput,
		SourceSheet: "LineItems",
		OutputFile:  validationOutput,
		ValidationRule: runtimetaskspec.DataValidationRule{
			Ranges:        []string{"A2:A10"},
			RuleType:      "list",
			AllowedValues: []string{"A001", "B002"},
			AllowBlank:    false,
		},
	})
	validationResult, err := Run(Request{ScenarioID: "invoice-preview-validation", TaskSpec: validationTask.TaskSpec})
	if err != nil {
		t.Fatalf("validation Run: %v", err)
	}
	if !validationResult.Verification.Pass {
		t.Fatalf("validation verification failed: %+v", validationResult.Verification)
	}

	protectionTask := runtimetaskspec.BuildProtectFormulaCellsTask(runtimetaskspec.ProtectFormulaCellsRequest{
		RequestText: "청구서 계산 수식은 보호하고 line item 입력 범위는 편집 가능하게 둔다.",
		InputFile:   validationOutput,
		SourceSheet: "LineItems",
		OutputFile:  protectedOutput,
		ProtectionRule: runtimetaskspec.FormulaProtectionRule{
			FormulaRanges: []string{"D2:E3"},
			InputRanges:   []string{"A2:C10"},
		},
	})
	protectionResult, err := Run(Request{ScenarioID: "invoice-preview-protection", TaskSpec: protectionTask.TaskSpec})
	if err != nil {
		t.Fatalf("protection Run: %v", err)
	}
	if !protectionResult.Verification.Pass {
		t.Fatalf("protection verification failed: %+v", protectionResult.Verification)
	}

	outputHandle, err := excelize.OpenFile(protectedOutput)
	if err != nil {
		t.Fatalf("Open protected output: %v", err)
	}
	defer func() { _ = outputHandle.Close() }()
	if got, err := outputHandle.GetCellValue("LineItems", "A3"); err != nil || got != "B002" {
		t.Fatalf("LineItems!A3=%q err=%v want B002", got, err)
	}
	if got, err := outputHandle.GetCellFormula("LineItems", "D3"); err != nil || got != "=B3*C3" {
		t.Fatalf("LineItems!D3 formula=%q err=%v want =B3*C3", got, err)
	}
	if got, err := outputHandle.GetCellFormula("LineItems", "E3"); err != nil || got != "=D3*0.1" {
		t.Fatalf("LineItems!E3 formula=%q err=%v want =D3*0.1", got, err)
	}
	validations, err := outputHandle.GetDataValidations("LineItems")
	if err != nil {
		t.Fatalf("GetDataValidations: %v", err)
	}
	foundValidation := false
	for _, validation := range validations {
		if validation != nil && validation.Sqref == "A2:A10" && validation.Type == "list" {
			foundValidation = true
			break
		}
	}
	if !foundValidation {
		t.Fatalf("expected list validation on LineItems!A2:A10, got %+v", validations)
	}
	protection, err := outputHandle.GetSheetProtection("LineItems")
	if err != nil {
		t.Fatalf("GetSheetProtection: %v", err)
	}
	if !protection.SelectLockedCells || !protection.SelectUnlockedCells {
		t.Fatalf("expected sheet protection options, got %+v", protection)
	}
}

func TestTimesheetHoursLogPreviewComposesPeriodCopy(t *testing.T) {
	setRuntimeRoots(t)

	tempDir := t.TempDir()
	inputFile := filepath.Join(tempDir, "timesheet.xlsx")
	outputFile := filepath.Join(tempDir, "timesheet-week2.xlsx")

	file := excelize.NewFile()
	defer func() { _ = file.Close() }()
	defaultSheet := file.GetSheetName(0)
	if err := file.SetSheetName(defaultSheet, "Week1"); err != nil {
		t.Fatalf("SetSheetName: %v", err)
	}
	header := []any{"date", "hours", "rate", "pay"}
	if err := file.SetSheetRow("Week1", "A1", &header); err != nil {
		t.Fatalf("SetSheetRow(header): %v", err)
	}
	row := []any{"2026-06-01", 8, 25}
	if err := file.SetSheetRow("Week1", "A2", &row); err != nil {
		t.Fatalf("SetSheetRow(row): %v", err)
	}
	if err := file.SetCellFormula("Week1", "D2", "=B2*C2"); err != nil {
		t.Fatalf("SetCellFormula(D2): %v", err)
	}
	if err := file.SaveAs(inputFile); err != nil {
		t.Fatalf("SaveAs: %v", err)
	}

	copyTask := runtimetaskspec.BuildCopyPeriodSheetTask(runtimetaskspec.CopyPeriodSheetRequest{
		RequestText: "Week1 timesheet 구조를 Week2 시트로 복사한다.",
		InputFile:   inputFile,
		SourceSheet: "Week1",
		TargetSheet: "Week2",
		OutputFile:  outputFile,
	})
	copyResult, err := Run(Request{ScenarioID: "timesheet-preview-period-copy", TaskSpec: copyTask.TaskSpec})
	if err != nil {
		t.Fatalf("copy Run: %v", err)
	}
	if !copyResult.Verification.Pass {
		t.Fatalf("copy verification failed: %+v", copyResult.Verification)
	}

	outputHandle, err := excelize.OpenFile(outputFile)
	if err != nil {
		t.Fatalf("Open output: %v", err)
	}
	defer func() { _ = outputHandle.Close() }()
	if got, err := outputHandle.GetCellValue("Week2", "A2"); err != nil || got != "2026-06-01" {
		t.Fatalf("Week2!A2=%q err=%v want 2026-06-01", got, err)
	}
	if got, err := outputHandle.GetCellFormula("Week2", "D2"); err != nil || got != "=B2*C2" {
		t.Fatalf("Week2!D2 formula=%q err=%v want =B2*C2", got, err)
	}
}
