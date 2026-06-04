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
