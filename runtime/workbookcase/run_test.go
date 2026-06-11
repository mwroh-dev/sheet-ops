package workbookcase

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/mwroh/sheet-ops/internal/testfixtures"
	runtimecompiler "github.com/mwroh/sheet-ops/runtime/compiler"
	runtimeexecute "github.com/mwroh/sheet-ops/runtime/execute"
	runtimeschema "github.com/mwroh/sheet-ops/runtime/schema"
	runtimetaskspec "github.com/mwroh/sheet-ops/runtime/taskspec"
	runtimeverify "github.com/mwroh/sheet-ops/runtime/verify"
	"github.com/xuri/excelize/v2"
)

func setRuntimeRoots(t *testing.T) {
	t.Helper()

	root := t.TempDir()
	t.Setenv(ArtifactRootEnv, filepath.Join(root, "artifacts"))
	t.Setenv("SHEET_OPS_KNOWLEDGE_ROOT", filepath.Join(root, "knowledge"))
}

func TestRunUsesClosedArtifactsForSummaryAndHighlight(t *testing.T) {
	originalRunSummary := runSummaryPlanWithArtifacts
	originalVerifySummary := verifySummaryExecutionWithArtifacts
	originalRunHighlight := runHighlightWithArtifacts
	originalVerifyHighlight := verifyHighlightExecutionWithArtifacts
	defer func() {
		runSummaryPlanWithArtifacts = originalRunSummary
		verifySummaryExecutionWithArtifacts = originalVerifySummary
		runHighlightWithArtifacts = originalRunHighlight
		verifyHighlightExecutionWithArtifacts = originalVerifyHighlight
	}()

	var summaryExecuteCalled bool
	var summaryVerifyCalled bool
	var highlightExecuteCalled bool
	var highlightVerifyCalled bool

	runSummaryPlanWithArtifacts = func(plan runtimecompiler.GroupSummarizePlan, outputWorkbook string) (runtimeexecute.ExecutionResult, error) {
		summaryExecuteCalled = true
		return originalRunSummary(plan, outputWorkbook)
	}
	verifySummaryExecutionWithArtifacts = func(plan runtimecompiler.GroupSummarizePlan, outputWorkbook, before, after string) (runtimeverify.VerificationResult, error) {
		summaryVerifyCalled = true
		return originalVerifySummary(plan, outputWorkbook, before, after)
	}
	runHighlightWithArtifacts = func(inspection runtimecompiler.WorkbookInspection, ir runtimecompiler.WorkbookOperationIR, outputWorkbook string) (runtimeexecute.ExecutionResult, error) {
		highlightExecuteCalled = true
		return originalRunHighlight(inspection, ir, outputWorkbook)
	}
	verifyHighlightExecutionWithArtifacts = func(inspection runtimecompiler.WorkbookInspection, ir runtimecompiler.WorkbookOperationIR, outputWorkbook, before, after string) (runtimeverify.VerificationResult, error) {
		highlightVerifyCalled = true
		return originalVerifyHighlight(inspection, ir, outputWorkbook, before, after)
	}

	setRuntimeRoots(t)

	summaryInput := filepath.Join(t.TempDir(), "orders-summary.xlsx")
	summaryOutput := filepath.Join(t.TempDir(), "orders-summary-output.xlsx")
	if err := testfixtures.CreateSummaryOrdersFixture(summaryInput); err != nil {
		t.Fatalf("CreateOrdersFixture(summary): %v", err)
	}
	summaryTask := runtimetaskspec.BuildGroupSummarizeTask(runtimetaskspec.UseRequest{
		RequestText: "주문내역에서 취소 건을 제외하고 상품별 매출 합계와 주문 수를 요약 시트로 만든다.",
		InputFile:   summaryInput,
		SourceSheet: "주문내역",
		OutputFile:  summaryOutput,
		TargetSheet: "상품별_요약",
		SummaryMode: "values",
		Filters: []runtimetaskspec.FilterSpec{
			{Column: "배송상태", Op: "!=", Value: "취소"},
		},
		GroupBy: []string{"상품명"},
		Metrics: []runtimetaskspec.MetricSpec{
			{Column: "결제금액", Op: "sum", As: "총매출"},
			{Column: "주문번호", Op: "count", As: "주문수"},
		},
	})
	summaryResult, err := Run(Request{ScenarioID: "workbookcase-summary-artifacts", TaskSpec: summaryTask.TaskSpec})
	if err != nil {
		t.Fatalf("Run(summary): %v", err)
	}
	if !summaryResult.Verification.Pass {
		t.Fatalf("summary verification failed: %+v", summaryResult.Verification)
	}
	if !summaryExecuteCalled || !summaryVerifyCalled {
		t.Fatalf("summary artifact seams not used execute=%v verify=%v", summaryExecuteCalled, summaryVerifyCalled)
	}

	highlightInput := filepath.Join(t.TempDir(), "orders-highlight.xlsx")
	highlightOutput := filepath.Join(t.TempDir(), "orders-highlight-output.xlsx")
	if err := testfixtures.CreateOrdersFixture(highlightInput); err != nil {
		t.Fatalf("CreateOrdersFixture(highlight): %v", err)
	}
	threshold := 100000.0
	highlightTask := runtimetaskspec.BuildHighlightThresholdTask(runtimetaskspec.HighlightThresholdRequest{
		RequestText:    "orders.xlsx의 Orders 시트에서 amount 열 값이 100000보다 큰 행을 찾아 노란색으로 표시한다.",
		InputFile:      highlightInput,
		SourceSheet:    "Orders",
		OutputFile:     highlightOutput,
		Column:         "amount",
		Operator:       ">",
		Threshold:      &threshold,
		HighlightColor: "#FFF59D",
	})
	highlightResult, err := Run(Request{ScenarioID: "workbookcase-highlight-artifacts", TaskSpec: highlightTask.TaskSpec})
	if err != nil {
		t.Fatalf("Run(highlight): %v", err)
	}
	if !highlightResult.Verification.Pass {
		t.Fatalf("highlight verification failed: %+v", highlightResult.Verification)
	}
	if !highlightExecuteCalled || !highlightVerifyCalled {
		t.Fatalf("highlight artifact seams not used execute=%v verify=%v", highlightExecuteCalled, highlightVerifyCalled)
	}
}

func TestWorkbookcaseSummariesCarryWrittenCells(t *testing.T) {
	execution := executionSummaryFromRuntime(runtimeexecute.ExecutionResult{
		OperationFamily: "write_values",
		OutputWorkbook:  "/tmp/output.xlsx",
		SummarySheet:    "Notes",
		WrittenCells:    []string{"Notes!A1", "Notes!B1"},
	})
	if execution.Operation != WriteValuesOperationName {
		t.Fatalf("execution operation=%q want %s", execution.Operation, WriteValuesOperationName)
	}
	if got, want := execution.WrittenCells, []string{"Notes!A1", "Notes!B1"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("execution written cells=%v want %v", got, want)
	}

	verification := verificationSummaryFromRuntime(WriteValuesOperationName, runtimeverify.VerificationResult{
		Pass:            true,
		OperationFamily: "write_values",
		OutputWorkbook:  "/tmp/output.xlsx",
		SummarySheet:    "Notes",
		WrittenCells:    []string{"Notes!A1", "Notes!B1"},
		Layers: []runtimeverify.LayerResult{
			{Level: 1, Name: "file_opens", Pass: true},
		},
		Reasons: []string{"pass"},
	})
	if verification.Operation != WriteValuesOperationName {
		t.Fatalf("verification operation=%q want %s", verification.Operation, WriteValuesOperationName)
	}
	if got, want := verification.WrittenCells, []string{"Notes!A1", "Notes!B1"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("verification written cells=%v want %v", got, want)
	}
	if len(verification.Layers) != 1 || verification.Layers[0].Name != "file_opens" {
		t.Fatalf("verification layers=%+v want file_opens", verification.Layers)
	}
}

func TestVerificationSummaryCarriesOutputWorkbookFingerprint(t *testing.T) {
	outputFile := filepath.Join(t.TempDir(), "output.xlsx")
	if err := os.WriteFile(outputFile, []byte("workbook bytes"), 0o644); err != nil {
		t.Fatalf("WriteFile(%s): %v", outputFile, err)
	}

	verification := verificationSummaryFromRuntime(WriteValuesOperationName, runtimeverify.VerificationResult{
		Pass:           true,
		OutputWorkbook: outputFile,
		Reasons:        []string{"pass"},
	})

	if verification.OutputWorkbookSHA256 != "92860184f82a31ed8824ef2f06029d344c6e806f6d8922bbbbd6feca4a01551e" {
		t.Fatalf("output_workbook_sha256 = %q, want SHA-256 of output file", verification.OutputWorkbookSHA256)
	}
}

func TestVerificationSchemaRejectsSuccessfulResultWithoutOutputFingerprint(t *testing.T) {
	document := map[string]any{
		"pass":        true,
		"operation":   WriteValuesOperationName,
		"output_file": "/tmp/output.xlsx",
		"reasons":     []string{"pass"},
	}

	if err := runtimeschema.ValidateStruct(repoJoin("contracts", "verification", "verification_result.schema.json"), document); err == nil {
		t.Fatalf("verification schema accepted pass=true result without output_workbook_sha256")
	}
}

func TestVerificationSchemaAllowsFailedResultWithoutOutputFileOrFingerprint(t *testing.T) {
	document := map[string]any{
		"pass":        false,
		"operation":   WriteValuesOperationName,
		"output_file": "",
		"reasons":     []string{"failed before output identity was available"},
	}

	if err := runtimeschema.ValidateStruct(repoJoin("contracts", "verification", "verification_result.schema.json"), document); err != nil {
		t.Fatalf("verification schema rejected pass=false result without output_file or output_workbook_sha256: %v", err)
	}
}

func TestVerificationSchemaRejectsFailedResultWithOutputFileWithoutOutputFingerprint(t *testing.T) {
	document := map[string]any{
		"pass":        false,
		"operation":   WriteValuesOperationName,
		"output_file": "/tmp/output.xlsx",
		"reasons":     []string{"failed after output workbook was materialized"},
	}

	if err := runtimeschema.ValidateStruct(repoJoin("contracts", "verification", "verification_result.schema.json"), document); err == nil {
		t.Fatalf("verification schema accepted pass=false result with output_file but without output_workbook_sha256")
	}
}

func TestRunAppendStructuredRowsEndToEnd(t *testing.T) {
	setRuntimeRoots(t)

	tempDir := t.TempDir()
	inputFile := filepath.Join(tempDir, "line-items.xlsx")
	outputFile := filepath.Join(tempDir, "line-items-output.xlsx")
	file := excelize.NewFile()
	defer func() { _ = file.Close() }()
	defaultSheet := file.GetSheetName(0)
	if err := file.SetSheetName(defaultSheet, "LineItems"); err != nil {
		t.Fatalf("SetSheetName: %v", err)
	}
	header := []any{"sku", "quantity", "unit_price"}
	if err := file.SetSheetRow("LineItems", "A1", &header); err != nil {
		t.Fatalf("SetSheetRow(header): %v", err)
	}
	existing := []any{"A001", 2, 10}
	if err := file.SetSheetRow("LineItems", "A2", &existing); err != nil {
		t.Fatalf("SetSheetRow(existing): %v", err)
	}
	if err := file.SaveAs(inputFile); err != nil {
		t.Fatalf("SaveAs: %v", err)
	}

	task := runtimetaskspec.BuildAppendStructuredRowsTask(runtimetaskspec.AppendStructuredRowsRequest{
		RequestText:          "LineItems 시트에 새 품목 행을 추가한다.",
		InputFile:            inputFile,
		SourceSheet:          "LineItems",
		OutputFile:           outputFile,
		IncludeSourceColumns: []string{"sku", "quantity", "unit_price"},
		Values: []runtimetaskspec.CellValue{
			{Cell: "sku", Value: "B002"},
			{Cell: "quantity", Value: 3},
			{Cell: "unit_price", Value: 15},
		},
	})
	result, err := Run(Request{ScenarioID: "workbookcase-append-rows", TaskSpec: task.TaskSpec})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !result.Verification.Pass {
		t.Fatalf("verification failed: %+v", result.Verification)
	}
	if result.Verification.Operation != AppendRowsOperationName {
		t.Fatalf("verification operation=%q want %s", result.Verification.Operation, AppendRowsOperationName)
	}
	if len(result.Verification.WrittenCells) != 3 {
		t.Fatalf("written cells=%v want 3", result.Verification.WrittenCells)
	}

	outputHandle, err := excelize.OpenFile(outputFile)
	if err != nil {
		t.Fatalf("Open output: %v", err)
	}
	defer func() { _ = outputHandle.Close() }()
	got, err := outputHandle.GetCellValue("LineItems", "A3")
	if err != nil {
		t.Fatalf("GetCellValue A3: %v", err)
	}
	if got != "B002" {
		t.Fatalf("LineItems!A3=%q want B002", got)
	}
}

func TestRunExtendTableFormulasEndToEnd(t *testing.T) {
	setRuntimeRoots(t)

	tempDir := t.TempDir()
	inputFile := filepath.Join(tempDir, "line-items.xlsx")
	outputFile := filepath.Join(tempDir, "line-items-output.xlsx")
	file := excelize.NewFile()
	defer func() { _ = file.Close() }()
	defaultSheet := file.GetSheetName(0)
	if err := file.SetSheetName(defaultSheet, "LineItems"); err != nil {
		t.Fatalf("SetSheetName: %v", err)
	}
	header := []any{"sku", "quantity", "unit_price", "line_total"}
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
	row3 := []any{"B002", 3, 15}
	if err := file.SetSheetRow("LineItems", "A3", &row3); err != nil {
		t.Fatalf("SetSheetRow(row3): %v", err)
	}
	if err := file.SaveAs(inputFile); err != nil {
		t.Fatalf("SaveAs: %v", err)
	}

	task := runtimetaskspec.BuildExtendTableFormulasTask(runtimetaskspec.ExtendTableFormulasRequest{
		RequestText:      "LineItems 시트에서 2행 수식을 3행으로 확장한다.",
		InputFile:        inputFile,
		SourceSheet:      "LineItems",
		OutputFile:       outputFile,
		FormulaSourceRow: 2,
		TargetRows:       []int{3},
		FormulaColumns:   []string{"D"},
	})
	result, err := Run(Request{ScenarioID: "workbookcase-extend-formulas", TaskSpec: task.TaskSpec})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !result.Verification.Pass {
		t.Fatalf("verification failed: %+v", result.Verification)
	}
	if result.Verification.Operation != ExtendFormulasOperationName {
		t.Fatalf("verification operation=%q want %s", result.Verification.Operation, ExtendFormulasOperationName)
	}
	if len(result.Verification.FormulaCells) != 1 {
		t.Fatalf("formula cells=%v want 1", result.Verification.FormulaCells)
	}

	outputHandle, err := excelize.OpenFile(outputFile)
	if err != nil {
		t.Fatalf("Open output: %v", err)
	}
	defer func() { _ = outputHandle.Close() }()
	got, err := outputHandle.GetCellFormula("LineItems", "D3")
	if err != nil {
		t.Fatalf("GetCellFormula D3: %v", err)
	}
	if got != "=B3*C3" {
		t.Fatalf("LineItems!D3 formula=%q want =B3*C3", got)
	}
}

func TestRunAddDataValidationEndToEnd(t *testing.T) {
	setRuntimeRoots(t)

	tempDir := t.TempDir()
	inputFile := filepath.Join(tempDir, "invoice.xlsx")
	outputFile := filepath.Join(tempDir, "invoice-output.xlsx")
	file := excelize.NewFile()
	defer func() { _ = file.Close() }()
	defaultSheet := file.GetSheetName(0)
	if err := file.SetSheetName(defaultSheet, "LineItems"); err != nil {
		t.Fatalf("SetSheetName: %v", err)
	}
	header := []any{"sku", "status"}
	if err := file.SetSheetRow("LineItems", "A1", &header); err != nil {
		t.Fatalf("SetSheetRow(header): %v", err)
	}
	row := []any{"A001", "draft"}
	if err := file.SetSheetRow("LineItems", "A2", &row); err != nil {
		t.Fatalf("SetSheetRow(row): %v", err)
	}
	if err := file.SaveAs(inputFile); err != nil {
		t.Fatalf("SaveAs: %v", err)
	}

	task := runtimetaskspec.BuildAddDataValidationTask(runtimetaskspec.AddDataValidationRequest{
		RequestText: "LineItems 상태 열에 dropdown 검증을 추가한다.",
		InputFile:   inputFile,
		SourceSheet: "LineItems",
		OutputFile:  outputFile,
		ValidationRule: runtimetaskspec.DataValidationRule{
			Ranges:        []string{"B2:B10"},
			RuleType:      "list",
			AllowedValues: []string{"draft", "sent", "paid"},
			AllowBlank:    false,
		},
	})
	result, err := Run(Request{ScenarioID: "workbookcase-add-validation", TaskSpec: task.TaskSpec})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !result.Verification.Pass {
		t.Fatalf("verification failed: %+v", result.Verification)
	}
	if result.Verification.Operation != AddDataValidationOperationName {
		t.Fatalf("verification operation=%q want %s", result.Verification.Operation, AddDataValidationOperationName)
	}
	if len(result.Verification.WrittenCells) != 1 || result.Verification.WrittenCells[0] != "B2:B10" {
		t.Fatalf("written cells=%v want [B2:B10]", result.Verification.WrittenCells)
	}
}

func TestRunProtectFormulaCellsEndToEnd(t *testing.T) {
	setRuntimeRoots(t)

	tempDir := t.TempDir()
	inputFile := filepath.Join(tempDir, "invoice.xlsx")
	outputFile := filepath.Join(tempDir, "invoice-output.xlsx")
	file := excelize.NewFile()
	defer func() { _ = file.Close() }()
	defaultSheet := file.GetSheetName(0)
	if err := file.SetSheetName(defaultSheet, "LineItems"); err != nil {
		t.Fatalf("SetSheetName: %v", err)
	}
	header := []any{"sku", "quantity", "unit_price", "line_total"}
	if err := file.SetSheetRow("LineItems", "A1", &header); err != nil {
		t.Fatalf("SetSheetRow(header): %v", err)
	}
	row := []any{"A001", 2, 10}
	if err := file.SetSheetRow("LineItems", "A2", &row); err != nil {
		t.Fatalf("SetSheetRow(row): %v", err)
	}
	if err := file.SetCellFormula("LineItems", "D2", "=B2*C2"); err != nil {
		t.Fatalf("SetCellFormula(D2): %v", err)
	}
	if err := file.SaveAs(inputFile); err != nil {
		t.Fatalf("SaveAs: %v", err)
	}

	task := runtimetaskspec.BuildProtectFormulaCellsTask(runtimetaskspec.ProtectFormulaCellsRequest{
		RequestText: "LineItems 계산 수식 셀은 보호하고 입력 셀은 편집 가능하게 둔다.",
		InputFile:   inputFile,
		SourceSheet: "LineItems",
		OutputFile:  outputFile,
		ProtectionRule: runtimetaskspec.FormulaProtectionRule{
			FormulaRanges: []string{"D2"},
			InputRanges:   []string{"A2:C10"},
		},
	})
	result, err := Run(Request{ScenarioID: "workbookcase-protect-formulas", TaskSpec: task.TaskSpec})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !result.Verification.Pass {
		t.Fatalf("verification failed: %+v", result.Verification)
	}
	if result.Verification.Operation != ProtectFormulaCellsOperationName {
		t.Fatalf("verification operation=%q want %s", result.Verification.Operation, ProtectFormulaCellsOperationName)
	}
	if len(result.Verification.FormulaCells) != 1 || result.Verification.FormulaCells[0] != "D2" {
		t.Fatalf("formula cells=%v want [D2]", result.Verification.FormulaCells)
	}
}

func TestRunCopyPeriodSheetEndToEnd(t *testing.T) {
	setRuntimeRoots(t)

	tempDir := t.TempDir()
	inputFile := filepath.Join(tempDir, "budget.xlsx")
	outputFile := filepath.Join(tempDir, "budget-output.xlsx")
	file := excelize.NewFile()
	defer func() { _ = file.Close() }()
	defaultSheet := file.GetSheetName(0)
	if err := file.SetSheetName(defaultSheet, "Jan"); err != nil {
		t.Fatalf("SetSheetName: %v", err)
	}
	if err := file.SetCellValue("Jan", "A1", "Period"); err != nil {
		t.Fatalf("SetCellValue(A1): %v", err)
	}
	if err := file.SetCellValue("Jan", "B1", "Jan"); err != nil {
		t.Fatalf("SetCellValue(B1): %v", err)
	}
	if err := file.SetCellFormula("Jan", "C2", "=B2*2"); err != nil {
		t.Fatalf("SetCellFormula(C2): %v", err)
	}
	if err := file.SetCellValue("Jan", "B2", 100); err != nil {
		t.Fatalf("SetCellValue(B2): %v", err)
	}
	if err := file.SaveAs(inputFile); err != nil {
		t.Fatalf("SaveAs: %v", err)
	}

	task := runtimetaskspec.BuildCopyPeriodSheetTask(runtimetaskspec.CopyPeriodSheetRequest{
		RequestText: "Jan 시트를 Feb 시트로 복사한다.",
		InputFile:   inputFile,
		SourceSheet: "Jan",
		TargetSheet: "Feb",
		OutputFile:  outputFile,
	})
	result, err := Run(Request{ScenarioID: "workbookcase-copy-period", TaskSpec: task.TaskSpec})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !result.Verification.Pass {
		t.Fatalf("verification failed: %+v", result.Verification)
	}
	if result.Verification.Operation != CopyPeriodSheetOperationName {
		t.Fatalf("verification operation=%q want %s", result.Verification.Operation, CopyPeriodSheetOperationName)
	}
	if result.Verification.SummarySheet != "Feb" {
		t.Fatalf("summary sheet=%q want Feb", result.Verification.SummarySheet)
	}
}

func TestRunNormalizeHeadersEndToEnd(t *testing.T) {
	setRuntimeRoots(t)

	tempDir := t.TempDir()
	inputFile := filepath.Join(tempDir, "inventory.xlsx")
	outputFile := filepath.Join(tempDir, "inventory-output.xlsx")
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

	task := runtimetaskspec.BuildNormalizeHeadersTask(runtimetaskspec.NormalizeHeadersRequest{
		RequestText: "Movements 시트 헤더를 표준 필드명으로 정규화한다.",
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
	result, err := Run(Request{ScenarioID: "workbookcase-normalize-headers", TaskSpec: task.TaskSpec})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !result.Verification.Pass {
		t.Fatalf("verification failed: %+v", result.Verification)
	}
	if result.Verification.Operation != NormalizeHeadersOperationName {
		t.Fatalf("verification operation=%q want %s", result.Verification.Operation, NormalizeHeadersOperationName)
	}
	if len(result.Verification.WrittenCells) != 4 {
		t.Fatalf("written cells=%v want 4", result.Verification.WrittenCells)
	}

	outputHandle, err := excelize.OpenFile(outputFile)
	if err != nil {
		t.Fatalf("Open output: %v", err)
	}
	defer func() { _ = outputHandle.Close() }()
	got, err := outputHandle.GetCellValue("Movements", "B1")
	if err != nil {
		t.Fatalf("GetCellValue B1: %v", err)
	}
	if got != "quantity_in" {
		t.Fatalf("Movements!B1=%q want quantity_in", got)
	}
	formula, err := outputHandle.GetCellFormula("Movements", "D2")
	if err != nil {
		t.Fatalf("GetCellFormula D2: %v", err)
	}
	if formula != "=B2-C2" {
		t.Fatalf("Movements!D2 formula=%q want =B2-C2", formula)
	}
}

func TestRunRollForwardPeriodEndToEnd(t *testing.T) {
	setRuntimeRoots(t)

	tempDir := t.TempDir()
	inputFile := filepath.Join(tempDir, "cash-flow.xlsx")
	outputFile := filepath.Join(tempDir, "cash-flow-output.xlsx")
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

	task := runtimetaskspec.BuildRollForwardPeriodTask(runtimetaskspec.RollForwardPeriodRequest{
		RequestText: "Jan closing balance를 Feb opening balance로 이월한다.",
		InputFile:   inputFile,
		SourceSheet: "Jan",
		TargetSheet: "Feb",
		OutputFile:  outputFile,
		CarryForwardMappings: []runtimetaskspec.CarryForwardMapping{
			{FromSheet: "Jan", FromCell: "D2", ToSheet: "Feb", ToCell: "A2"},
		},
	})
	result, err := Run(Request{ScenarioID: "workbookcase-roll-forward-period", TaskSpec: task.TaskSpec})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !result.Verification.Pass {
		t.Fatalf("verification failed: %+v", result.Verification)
	}
	if result.Verification.Operation != RollForwardPeriodOperationName {
		t.Fatalf("verification operation=%q want %s", result.Verification.Operation, RollForwardPeriodOperationName)
	}
	if len(result.Verification.WrittenCells) != 1 || result.Verification.WrittenCells[0] != "Feb!A2" {
		t.Fatalf("written cells=%v want [Feb!A2]", result.Verification.WrittenCells)
	}

	outputHandle, err := excelize.OpenFile(outputFile)
	if err != nil {
		t.Fatalf("Open output: %v", err)
	}
	defer func() { _ = outputHandle.Close() }()
	got, err := outputHandle.GetCellValue("Feb", "A2")
	if err != nil {
		t.Fatalf("GetCellValue Feb A2: %v", err)
	}
	if got != "150" {
		t.Fatalf("Feb!A2=%q want 150", got)
	}
	formula, err := outputHandle.GetCellFormula("Feb", "D2")
	if err != nil {
		t.Fatalf("GetCellFormula Feb D2: %v", err)
	}
	if formula != "=A2+B2-C2" {
		t.Fatalf("Feb!D2 formula=%q want =A2+B2-C2", formula)
	}
}

func TestRunReconcileTablesEndToEnd(t *testing.T) {
	setRuntimeRoots(t)

	tempDir := t.TempDir()
	inputFile := filepath.Join(tempDir, "inventory.xlsx")
	outputFile := filepath.Join(tempDir, "inventory-output.xlsx")
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

	task := runtimetaskspec.BuildReconcileTablesTask(runtimetaskspec.ReconcileTablesRequest{
		RequestText: "Movements balance와 StockMaster on_hand를 sku 기준으로 대조한다.",
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
	result, err := Run(Request{ScenarioID: "workbookcase-reconcile-tables", TaskSpec: task.TaskSpec})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !result.Verification.Pass {
		t.Fatalf("verification failed: %+v", result.Verification)
	}
	if result.Verification.Operation != ReconcileTablesOperationName {
		t.Fatalf("verification operation=%q want %s", result.Verification.Operation, ReconcileTablesOperationName)
	}
	if result.Verification.SummaryRows != 4 {
		t.Fatalf("summary rows=%d want 4", result.Verification.SummaryRows)
	}
}

func TestRunGeneratePrintableFormEndToEnd(t *testing.T) {
	setRuntimeRoots(t)

	tempDir := t.TempDir()
	inputFile := filepath.Join(tempDir, "invoice.xlsx")
	outputFile := filepath.Join(tempDir, "invoice-output.xlsx")
	file := excelize.NewFile()
	defer func() { _ = file.Close() }()
	defaultSheet := file.GetSheetName(0)
	if err := file.SetSheetName(defaultSheet, "InvoiceData"); err != nil {
		t.Fatalf("SetSheetName: %v", err)
	}
	if err := file.SetCellValue("InvoiceData", "B2", "INV-001"); err != nil {
		t.Fatalf("SetCellValue invoice: %v", err)
	}
	if err := file.SetCellValue("InvoiceData", "B3", "Acme Co"); err != nil {
		t.Fatalf("SetCellValue customer: %v", err)
	}
	if _, err := file.NewSheet("LineItems"); err != nil {
		t.Fatalf("NewSheet LineItems: %v", err)
	}
	header := []any{"sku", "quantity", "unit_price", "line_total"}
	if err := file.SetSheetRow("LineItems", "A1", &header); err != nil {
		t.Fatalf("SetSheetRow header: %v", err)
	}
	row := []any{"A001", 2, 10, 20}
	if err := file.SetSheetRow("LineItems", "A2", &row); err != nil {
		t.Fatalf("SetSheetRow row: %v", err)
	}
	if err := file.SaveAs(inputFile); err != nil {
		t.Fatalf("SaveAs: %v", err)
	}

	task := runtimetaskspec.BuildGeneratePrintableFormTask(runtimetaskspec.GeneratePrintableFormRequest{
		RequestText: "invoice printable form을 생성한다.",
		InputFile:   inputFile,
		SourceSheet: "InvoiceData",
		TargetSheet: "InvoicePrint",
		OutputFile:  outputFile,
		FormTitle:   "Invoice",
		PrintArea:   "A1:D8",
		FieldBindings: []runtimetaskspec.FormFieldBinding{
			{Label: "Invoice No", SourceSheet: "InvoiceData", SourceCell: "B2", LabelCell: "A2", ValueCell: "B2"},
			{Label: "Customer", SourceSheet: "InvoiceData", SourceCell: "B3", LabelCell: "A3", ValueCell: "B3"},
		},
		TableBinding: &runtimetaskspec.FormTableBinding{
			SourceSheet:   "LineItems",
			SourceColumns: []string{"sku", "quantity", "unit_price", "line_total"},
			HeaderStart:   "A5",
			DataStart:     "A6",
		},
	})
	result, err := Run(Request{ScenarioID: "workbookcase-generate-printable-form", TaskSpec: task.TaskSpec})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !result.Verification.Pass {
		t.Fatalf("verification failed: %+v", result.Verification)
	}
	if result.Verification.Operation != GeneratePrintableFormOperationName {
		t.Fatalf("verification operation=%q want %s", result.Verification.Operation, GeneratePrintableFormOperationName)
	}
	if result.Verification.SummarySheet != "InvoicePrint" {
		t.Fatalf("summary sheet=%q want InvoicePrint", result.Verification.SummarySheet)
	}
}

func TestRunRejectsUnsafeScenarioIDBeforeCreatingArtifacts(t *testing.T) {
	artifactRoot := t.TempDir()
	t.Setenv(ArtifactRootEnv, artifactRoot)

	_, err := Run(Request{ScenarioID: "../escape", TaskSpec: runtimetaskspec.TaskSpec{}})
	if err == nil {
		t.Fatalf("Run error=nil want invalid scenario id rejection")
	}
	if !strings.Contains(err.Error(), "scenario id") {
		t.Fatalf("Run error=%q want scenario id rejection", err.Error())
	}
	if _, statErr := os.Stat(filepath.Join(artifactRoot, "evidence")); !os.IsNotExist(statErr) {
		t.Fatalf("artifact evidence dir stat err=%v want not exist", statErr)
	}
}

func TestRunAcceptsCaseMarkdownTaskSpec(t *testing.T) {
	setRuntimeRoots(t)

	tempDir := t.TempDir()
	casePath := filepath.Join(tempDir, "case.md")
	requestText := "주문내역에서 취소 건을 제외하고 상품별 매출 합계와 주문 수를 요약 시트로 만든다."
	if err := os.WriteFile(casePath, []byte(requestText+"\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(case): %v", err)
	}

	inputFile := filepath.Join(tempDir, "orders.xlsx")
	outputFile := filepath.Join(tempDir, "orders-summary-output.xlsx")
	if err := testfixtures.CreateSummaryOrdersFixture(inputFile); err != nil {
		t.Fatalf("CreateOrdersFixture: %v", err)
	}

	taskSpec := runtimetaskspec.TaskSpec{
		RequestKind:     "workbook_case",
		ExecutionKind:   runtimetaskspec.ExecutionKindComposition,
		CompositionKind: runtimetaskspec.CompositionKindGroupSummary,
		Source:          runtimetaskspec.SourceSpec{Kind: "case_markdown", Path: casePath},
		InputWorkbook:   inputFile,
		OutputWorkbook:  outputFile,
		Operation:       SummaryOperationName,
		SourceSheet:     "주문내역",
		TargetSheet:     "상품별_요약",
		Filters: []runtimetaskspec.FilterSpec{
			{Column: "배송상태", Op: "!=", Value: "취소"},
		},
		GroupBy: []string{"상품명"},
		Metrics: []runtimetaskspec.MetricSpec{
			{Column: "결제금액", Op: "sum", As: "총매출"},
			{Column: "주문번호", Op: "count", As: "주문수"},
		},
		SummaryMode: "values",
	}

	result, err := Run(Request{
		ScenarioID: "workbookcase-case-markdown",
		TaskSpec:   taskSpec,
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !result.Verification.Pass {
		t.Fatalf("verification failed: %+v", result.Verification)
	}

	requestArtifact, err := os.ReadFile(result.Paths.RequestPath)
	if err != nil {
		t.Fatalf("ReadFile(request.md): %v", err)
	}
	if got := strings.TrimSpace(string(requestArtifact)); got != requestText {
		t.Fatalf("request.md=%q want %q", got, requestText)
	}

	evidenceRaw, err := os.ReadFile(result.Paths.EvidenceIndex)
	if err != nil {
		t.Fatalf("ReadFile(evidence): %v", err)
	}
	var evidence ScenarioEvidence
	if err := json.Unmarshal(evidenceRaw, &evidence); err != nil {
		t.Fatalf("Unmarshal(evidence): %v", err)
	}
	if evidence.Request == "" {
		t.Fatalf("scenario evidence request should not be empty")
	}
	if got := strings.TrimSpace(evidence.Request); got != requestText {
		t.Fatalf("scenario evidence request=%q want %q", got, requestText)
	}
	if evidence.RenderPath == "" {
		t.Fatalf("scenario evidence render_path should not be empty")
	}
	if _, err := os.Stat(evidence.RenderPath); err != nil {
		t.Fatalf("Stat(render_path): %v", err)
	}
}
