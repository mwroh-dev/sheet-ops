package workbookcase

import (
	"strconv"
	"strings"

	runtimeepisode "github.com/mwroh/sheet-ops/runtime/episode"
	runtimeschema "github.com/mwroh/sheet-ops/runtime/schema"
	runtimetaskspec "github.com/mwroh/sheet-ops/runtime/taskspec"
)

func writeEvidenceIndex(scenarioID string, taskSpec runtimetaskspec.TaskSpec, requestText string, paths RunPaths, result RunResult, failure *FailureDetails, outcome string) error {
	index := ScenarioEvidence{
		ScenarioID:     scenarioID,
		RunID:          result.IDs.RunID,
		Operation:      result.Operation(taskSpec.Operation),
		Request:        requestText,
		Outcome:        outcome,
		Telemetry:      telemetryPaths(paths),
		RuntimeVersion: runtimeepisode.CurrentRuntimeVersion(),
	}
	if fileExists(paths.PlanPath) {
		index.PlanPath = paths.PlanPath
	}
	if fileExists(paths.PolicyPath) {
		index.PolicyPath = paths.PolicyPath
	}
	if fileExists(paths.VerificationPath) {
		index.VerificationPath = paths.VerificationPath
	}
	if fileExists(paths.VerificationReviewPath) {
		index.VerificationReviewPath = paths.VerificationReviewPath
	}
	if fileExists(paths.RenderPath) {
		index.RenderPath = paths.RenderPath
	}
	if result.Episode != nil && fileExists(paths.EpisodePath) {
		index.EpisodePath = paths.EpisodePath
	}
	if failure != nil && fileExists(paths.FailureEvidencePath) {
		index.FailureEvidencePath = paths.FailureEvidencePath
	}
	if failure != nil && fileExists(paths.RepairAdvicePath) {
		index.RepairAdvicePath = paths.RepairAdvicePath
	}
	if err := runtimeschema.ValidateStruct(repoJoin("contracts", "evidence", "scenario_evidence.schema.json"), index); err != nil {
		return err
	}
	return writeJSON(paths.EvidenceIndex, index)
}

func writeReport(scenarioID string, taskSpec runtimetaskspec.TaskSpec, requestText string, paths RunPaths, result RunResult, failure *FailureDetails, outcome string) error {
	details := "- Highlighted Rows: " + intsToString(result.Verification.HighlightedRows)
	if result.Operation(taskSpec.Operation) == SummaryOperationName && result.Verification.Operation != "" {
		details = "- Summary Sheet: " + result.Verification.SummarySheet + "\n- Summary Rows: " + strconv.Itoa(result.Verification.SummaryRows)
	}
	if result.Verification.Operation == "" {
		details = "- Highlighted Rows: []"
	}

	lines := []string{
		"# Run Summary",
		"",
		"- Scenario: " + scenarioID,
		"- Run ID: " + result.IDs.RunID,
		"- Input File: " + taskSpec.InputWorkbook,
		"- Output File: " + result.OutputFile(taskSpec.OutputWorkbook),
		"- Operation: " + result.Operation(taskSpec.Operation),
		"- Outcome: " + outcome,
		"- Pass: " + strconv.FormatBool(outcome == "pass"),
		"- Request: " + strings.TrimSpace(requestText),
		details,
	}
	reasons := outcomeReasons(result, failure)
	lines = append(lines, "- Reasons: "+strings.Join(reasons, "; "))
	if failure != nil {
		lines = append(lines,
			"- Failure Phase: "+failure.Phase,
			"- Failure Class: "+failure.FailureClass,
			"- Domain Code: "+failure.DomainCode,
			"- Critical Step: "+failure.CriticalStep,
			"- Failure Message: "+failure.Message,
		)
	}
	lines = append(lines, "")
	return writeText(paths.ReportPath, strings.Join(lines, "\n"))
}

func boolOutcome(value bool) string {
	if value {
		return "pass"
	}
	return "fail"
}

func summaryExecutionArtifactPaths(paths RunPaths) runtimeepisode.ExecutionArtifactPaths {
	return runtimeepisode.ExecutionArtifactPaths{
		TaskSpecPath:     paths.TaskSpecPath,
		OperationIRPath:  paths.OperationIRPath,
		ExecutionPath:    paths.ExecutionPath,
		VerificationPath: paths.VerificationPath,
	}
}

func intsToString(values []int) string {
	if len(values) == 0 {
		return "[]"
	}
	out := make([]string, 0, len(values))
	for _, value := range values {
		out = append(out, strconv.Itoa(value))
	}
	return "[" + strings.Join(out, ", ") + "]"
}
