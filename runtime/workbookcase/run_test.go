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
