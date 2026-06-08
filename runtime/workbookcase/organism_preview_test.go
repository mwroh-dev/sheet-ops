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

func TestInventoryMovementLogPreviewComposesHeaderNormalization(t *testing.T) {
	setRuntimeRoots(t)

	tempDir := t.TempDir()
	inputFile := filepath.Join(tempDir, "inventory.xlsx")
	outputFile := filepath.Join(tempDir, "inventory-normalized.xlsx")

	file := excelize.NewFile()
	defer func() { _ = file.Close() }()
	defaultSheet := file.GetSheetName(0)
	if err := file.SetSheetName(defaultSheet, "Movements"); err != nil {
		t.Fatalf("SetSheetName: %v", err)
	}
	header := []any{"SKU ID", "Qty In", "Qty Out", "Balance"}
	if err := file.SetSheetRow("Movements", "A1", &header); err != nil {
		t.Fatalf("SetSheetRow(header): %v", err)
	}
	row := []any{"A001", 5, 1}
	if err := file.SetSheetRow("Movements", "A2", &row); err != nil {
		t.Fatalf("SetSheetRow(row): %v", err)
	}
	if err := file.SetCellFormula("Movements", "D2", "=B2-C2"); err != nil {
		t.Fatalf("SetCellFormula(D2): %v", err)
	}
	if err := file.SaveAs(inputFile); err != nil {
		t.Fatalf("SaveAs: %v", err)
	}

	normalizeTask := runtimetaskspec.BuildNormalizeHeadersTask(runtimetaskspec.NormalizeHeadersRequest{
		RequestText: "inventory movement log 헤더를 reconciliation 전에 표준 필드명으로 정규화한다.",
		InputFile:   inputFile,
		SourceSheet: "Movements",
		OutputFile:  outputFile,
		HeaderRow:   1,
		HeaderMappings: []runtimetaskspec.HeaderMapping{
			{From: "SKU ID", To: "sku"},
			{From: "Qty In", To: "quantity_in"},
			{From: "Qty Out", To: "quantity_out"},
			{From: "Balance", To: "balance"},
		},
	})
	result, err := Run(Request{ScenarioID: "inventory-preview-normalize-headers", TaskSpec: normalizeTask.TaskSpec})
	if err != nil {
		t.Fatalf("normalize Run: %v", err)
	}
	if !result.Verification.Pass {
		t.Fatalf("normalize verification failed: %+v", result.Verification)
	}

	outputHandle, err := excelize.OpenFile(outputFile)
	if err != nil {
		t.Fatalf("Open output: %v", err)
	}
	defer func() { _ = outputHandle.Close() }()
	if got, err := outputHandle.GetCellValue("Movements", "A1"); err != nil || got != "sku" {
		t.Fatalf("Movements!A1=%q err=%v want sku", got, err)
	}
	if got, err := outputHandle.GetCellValue("Movements", "A2"); err != nil || got != "A001" {
		t.Fatalf("Movements!A2=%q err=%v want A001", got, err)
	}
	if got, err := outputHandle.GetCellFormula("Movements", "D2"); err != nil || got != "=B2-C2" {
		t.Fatalf("Movements!D2 formula=%q err=%v want =B2-C2", got, err)
	}
}

func TestInventoryMovementLogPreviewComposesTableReconciliation(t *testing.T) {
	setRuntimeRoots(t)

	tempDir := t.TempDir()
	inputFile := filepath.Join(tempDir, "inventory.xlsx")
	outputFile := filepath.Join(tempDir, "inventory-reconciled.xlsx")

	file := excelize.NewFile()
	defer func() { _ = file.Close() }()
	defaultSheet := file.GetSheetName(0)
	if err := file.SetSheetName(defaultSheet, "Movements"); err != nil {
		t.Fatalf("SetSheetName: %v", err)
	}
	movementHeader := []any{"sku", "balance"}
	if err := file.SetSheetRow("Movements", "A1", &movementHeader); err != nil {
		t.Fatalf("SetSheetRow movement header: %v", err)
	}
	for idx, row := range [][]any{{"A001", 10}, {"B002", 5}, {"C003", 2}} {
		cell, _ := excelize.CoordinatesToCellName(1, idx+2)
		if err := file.SetSheetRow("Movements", cell, &row); err != nil {
			t.Fatalf("SetSheetRow movement %d: %v", idx, err)
		}
	}
	if _, err := file.NewSheet("StockMaster"); err != nil {
		t.Fatalf("NewSheet StockMaster: %v", err)
	}
	masterHeader := []any{"sku", "on_hand"}
	if err := file.SetSheetRow("StockMaster", "A1", &masterHeader); err != nil {
		t.Fatalf("SetSheetRow master header: %v", err)
	}
	for idx, row := range [][]any{{"A001", 10}, {"B002", 7}, {"D004", 1}} {
		cell, _ := excelize.CoordinatesToCellName(1, idx+2)
		if err := file.SetSheetRow("StockMaster", cell, &row); err != nil {
			t.Fatalf("SetSheetRow master %d: %v", idx, err)
		}
	}
	if err := file.SaveAs(inputFile); err != nil {
		t.Fatalf("SaveAs: %v", err)
	}

	reconcileTask := runtimetaskspec.BuildReconcileTablesTask(runtimetaskspec.ReconcileTablesRequest{
		RequestText: "inventory movement log의 balance를 stock master on_hand와 대조한다.",
		InputFile:   inputFile,
		SourceSheet: "Movements",
		LookupSheet: "StockMaster",
		TargetSheet: "Reconciliation",
		OutputFile:  outputFile,
		LeftKey:     "sku",
		RightKey:    "sku",
		CompareMappings: []runtimetaskspec.CompareMapping{
			{LeftColumn: "balance", RightColumn: "on_hand", As: "balance"},
		},
	})
	result, err := Run(Request{ScenarioID: "inventory-preview-reconcile-tables", TaskSpec: reconcileTask.TaskSpec})
	if err != nil {
		t.Fatalf("reconcile Run: %v", err)
	}
	if !result.Verification.Pass {
		t.Fatalf("reconcile verification failed: %+v", result.Verification)
	}
	if result.Verification.SummaryRows != 4 {
		t.Fatalf("summary rows=%d want 4", result.Verification.SummaryRows)
	}
}

func TestCashFlowMonitorPreviewComposesRollForwardPeriod(t *testing.T) {
	setRuntimeRoots(t)

	tempDir := t.TempDir()
	inputFile := filepath.Join(tempDir, "cash-flow.xlsx")
	outputFile := filepath.Join(tempDir, "cash-flow-rolled.xlsx")

	file := excelize.NewFile()
	defer func() { _ = file.Close() }()
	defaultSheet := file.GetSheetName(0)
	if err := file.SetSheetName(defaultSheet, "Jan"); err != nil {
		t.Fatalf("SetSheetName(Jan): %v", err)
	}
	if _, err := file.NewSheet("Feb"); err != nil {
		t.Fatalf("NewSheet(Feb): %v", err)
	}
	for _, sheet := range []string{"Jan", "Feb"} {
		header := []any{"opening", "inflow", "outflow", "closing"}
		if err := file.SetSheetRow(sheet, "A1", &header); err != nil {
			t.Fatalf("SetSheetRow(%s header): %v", sheet, err)
		}
	}
	jan := []any{100, 75, 25}
	if err := file.SetSheetRow("Jan", "A2", &jan); err != nil {
		t.Fatalf("SetSheetRow(Jan): %v", err)
	}
	if err := file.SetCellFormula("Jan", "D2", "=A2+B2-C2"); err != nil {
		t.Fatalf("SetCellFormula(Jan D2): %v", err)
	}
	feb := []any{0, 20, 10}
	if err := file.SetSheetRow("Feb", "A2", &feb); err != nil {
		t.Fatalf("SetSheetRow(Feb): %v", err)
	}
	if err := file.SetCellFormula("Feb", "D2", "=A2+B2-C2"); err != nil {
		t.Fatalf("SetCellFormula(Feb D2): %v", err)
	}
	if err := file.SaveAs(inputFile); err != nil {
		t.Fatalf("SaveAs: %v", err)
	}

	rollTask := runtimetaskspec.BuildRollForwardPeriodTask(runtimetaskspec.RollForwardPeriodRequest{
		RequestText: "cash flow monitor에서 Jan closing balance를 Feb opening balance로 이월한다.",
		InputFile:   inputFile,
		SourceSheet: "Jan",
		TargetSheet: "Feb",
		OutputFile:  outputFile,
		CarryForwardMappings: []runtimetaskspec.CarryForwardMapping{
			{FromSheet: "Jan", FromCell: "D2", ToSheet: "Feb", ToCell: "A2"},
		},
	})
	result, err := Run(Request{ScenarioID: "cash-flow-preview-roll-forward", TaskSpec: rollTask.TaskSpec})
	if err != nil {
		t.Fatalf("roll-forward Run: %v", err)
	}
	if !result.Verification.Pass {
		t.Fatalf("roll-forward verification failed: %+v", result.Verification)
	}

	outputHandle, err := excelize.OpenFile(outputFile)
	if err != nil {
		t.Fatalf("Open output: %v", err)
	}
	defer func() { _ = outputHandle.Close() }()
	if got, err := outputHandle.GetCellValue("Feb", "A2"); err != nil || got != "150" {
		t.Fatalf("Feb!A2=%q err=%v want 150", got, err)
	}
	if got, err := outputHandle.GetCellFormula("Feb", "D2"); err != nil || got != "=A2+B2-C2" {
		t.Fatalf("Feb!D2 formula=%q err=%v want =A2+B2-C2", got, err)
	}
}
