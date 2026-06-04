package workbookcase

import (
	"errors"
	"path/filepath"
	"strings"
	"time"

	runtimeepisode "github.com/mwroh/sheet-ops/runtime/episode"
	runtimeknowledge "github.com/mwroh/sheet-ops/runtime/knowledge"
	requestcompiler "github.com/mwroh/sheet-ops/runtime/openlayer/requestcompiler"
	runtimerender "github.com/mwroh/sheet-ops/runtime/render"
	runtimeschema "github.com/mwroh/sheet-ops/runtime/schema"
)

var (
	writeFailureEvidenceArtifact = writeFailureEvidence
	writeEvidenceIndexArtifact   = writeEvidenceIndex
	writeReportArtifact          = writeReport
)

func finalizeRun(req Request, requestText string, result *RunResult, runErr error) error {
	requestText = bestEffortRequestText(req.TaskSpec, requestText)
	failure := extractFailure(runErr)
	if failure == nil {
		failure = failureFromVerification(*result)
	}
	result.Failure = failure

	outcome := "pass"
	pass := true
	if failure != nil {
		outcome = "fail"
		pass = false
	}
	operation := result.Operation(req.TaskSpec.Operation)
	reasons := outcomeReasons(*result, failure)
	outcomeArtifact := OutcomeArtifact{
		Outcome:   outcome,
		Pass:      pass,
		Operation: operation,
		Reasons:   reasons,
		Failure:   failure,
	}

	var finalizeErrs []error
	recordFinalizeError := func(err error) {
		if err == nil {
			return
		}
		finalizeErrs = append(finalizeErrs, err)
	}

	if shouldPersistSensitiveWorkbookArtifacts() {
		recordFinalizeError(writeText(result.Paths.RequestPath, requestText))
	}
	recordFinalizeError(writeJSON(result.Paths.OutcomePath, outcomeArtifact))
	recordFinalizeError(appendFinalOutcomeEvent(result.IDs, result.Paths, outcomeArtifact))
	if review, ok := buildVerificationReview(*result, failure); ok {
		recordFinalizeError(writeVerificationReviewArtifact(result.Paths, review))
	}
	if result.Verification.Operation != "" && fileExists(result.OutputFile(req.TaskSpec.OutputWorkbook)) && shouldWriteRenderArtifactForRun(failure) {
		_ = writeRenderArtifact(result.Paths, result.OutputFile(req.TaskSpec.OutputWorkbook))
	}

	if failure != nil {
		recordFinalizeError(writeRepairAdviceArtifact(result.Paths, buildRepairAdvice(*result, failure)))
		if shouldPersistSensitiveWorkbookArtifacts() {
			recordFinalizeError(writeFailureEvidenceArtifact(req.ScenarioID, operation, result.Paths, *result, failure))
		}
	}
	if shouldPersistSensitiveWorkbookArtifacts() {
		recordFinalizeError(writeEvidenceIndexArtifact(req.ScenarioID, req.TaskSpec, requestText, result.Paths, *result, failure, outcome))
		recordFinalizeError(writeReportArtifact(req.ScenarioID, req.TaskSpec, requestText, result.Paths, *result, failure, outcome))
	}
	if len(finalizeErrs) == 0 && shouldAppendVerificationFailureKnowledge(failure) && fileExists(result.Paths.FailureEvidencePath) {
		recordFinalizeError(runtimeknowledge.AppendVerificationFailureEpisodicRecord(runtimeknowledge.VerificationFailureAppendInput{
			RunID:        result.IDs.RunID,
			ScenarioID:   req.ScenarioID,
			Operation:    operation,
			Outcome:      outcome,
			Phase:        failure.Phase,
			FailureClass: failure.FailureClass,
			DomainCode:   failure.DomainCode,
			CriticalStep: failure.CriticalStep,
			RepairHint:   buildRepairAdvice(*result, failure).RepairHint,
		}))
	}
	if len(finalizeErrs) == 0 && runErr == nil && result.Verification.Operation != "" {
		recordFinalizeError(runtimeknowledge.AppendOrchestratorEpisodicRecord(runtimeknowledge.OrchestratorAppendInput{
			RunID:      result.IDs.RunID,
			ScenarioID: req.ScenarioID,
			Operation:  result.Verification.Operation,
			Outcome:    outcome,
		}))
		recordFinalizeError(requestcompiler.PersistPreferenceObservations(
			requestcompiler.InferWorkspaceRoot(req.TaskSpec.InputWorkbook),
			preferenceObservationsForSuccessfulRun(req, *result),
		))
	}

	if len(finalizeErrs) == 0 {
		if runErr == nil && failure != nil {
			return errorFromFailure(*failure, nil)
		}
		return runErr
	}
	emissionErr := finalizeErrs[0]
	finalFailure := newFailure(
		"finalization",
		"artifact_emission_failure",
		"artifact_emission_failed",
		"emit final workbookcase run artifacts",
		emissionErr.Error(),
	)
	result.Failure = finalFailure

	cause := error(emissionErr)
	switch {
	case runErr != nil:
		cause = errors.Join(cause, runErr)
	case failure != nil:
		cause = errors.Join(cause, errorFromFailure(*failure, nil))
	}
	return errorFromFailure(*finalFailure, cause)
}

func preferenceObservationsForSuccessfulRun(req Request, result RunResult) []requestcompiler.PreferenceObservation {
	observations := []requestcompiler.PreferenceObservation{
		{
			Key:        "preserve_original",
			Value:      "true",
			RunID:      result.IDs.RunID,
			ScenarioID: req.ScenarioID,
			Operation:  result.Verification.Operation,
		},
		{
			Key:        "output_destination",
			Value:      "new_workbook",
			RunID:      result.IDs.RunID,
			ScenarioID: req.ScenarioID,
			Operation:  result.Verification.Operation,
		},
	}
	if result.Verification.Operation == SummaryOperationName && strings.TrimSpace(result.Verification.SummaryMode) != "" {
		observations = append(observations, requestcompiler.PreferenceObservation{
			Key:        "summary_mode",
			Value:      result.Verification.SummaryMode,
			RunID:      result.IDs.RunID,
			ScenarioID: req.ScenarioID,
			Operation:  result.Verification.Operation,
		})
	}
	return observations
}

func shouldAppendVerificationFailureKnowledge(failure *FailureDetails) bool {
	return failure != nil && failure.Phase == "verification"
}

func appendFinalOutcomeEvent(ids RunIDs, paths RunPaths, outcome OutcomeArtifact) error {
	payload := map[string]any{
		"operation": outcome.Operation,
		"outcome":   outcome.Outcome,
		"pass":      outcome.Pass,
	}
	if len(outcome.Reasons) > 0 {
		payload["reasons"] = append([]string(nil), outcome.Reasons...)
	}
	if outcome.Failure != nil {
		payload["phase"] = outcome.Failure.Phase
		payload["failure_class"] = outcome.Failure.FailureClass
		payload["domain_code"] = outcome.Failure.DomainCode
		payload["critical_step"] = outcome.Failure.CriticalStep
		payload["message"] = outcome.Failure.Message
	}
	return AppendTelemetry(paths, "outcome", TelemetryEvent{
		RunID:     ids.RunID,
		TraceID:   ids.TraceID,
		EventType: "run_finished",
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
		Component: "use-orchestrator",
		Payload:   payload,
	})
}

func buildVerificationReview(result RunResult, failure *FailureDetails) (VerificationReview, bool) {
	if result.Verification.Operation == "" {
		return VerificationReview{}, false
	}
	status := "pass"
	summary := "deterministic verification passed and the output remained within the expected execution boundary"
	if !result.Verification.Pass || failure != nil {
		status = "fail"
		summary = "deterministic verification did not prove that the emitted workbook matches the expected execution boundary"
	}
	return VerificationReview{
		Status:  status,
		Summary: summary,
		Reasons: append([]string(nil), result.Verification.Reasons...),
	}, true
}

func buildRepairAdvice(result RunResult, failure *FailureDetails) RepairAdvice {
	advice := RepairAdvice{
		Summary:          "the run stopped before a verified success result and needs a narrower follow-up action",
		RepairHint:       "inspect the failure phase and correct the request or runtime preconditions before rerunning",
		SuggestedActions: []string{"inspect the failure phase and correct the request or runtime preconditions before rerunning"},
		Assumptions:      []string{"fail-only advice; no automatic retry was attempted"},
	}
	if failure == nil {
		return advice
	}
	advice.FailureClass = failure.FailureClass
	advice.DomainCode = failure.DomainCode
	advice.RepairHint = repairHintForFailure(result, failure)
	switch failure.Phase {
	case "request_resolution":
		advice.Summary = "the request artifact could not be resolved before runtime execution"
		advice.SuggestedActions = []string{
			"confirm the case markdown request path exists",
			"rerun only after the request artifact is readable",
		}
	case "policy":
		advice.Summary = "the request must preserve the original workbook and write to a distinct output file before execution can continue"
		advice.SuggestedActions = []string{
			"provide a distinct output workbook path",
			"avoid in-place workbook mutation",
		}
	case "inspection":
		advice.Summary = "the requested workbook facts could not be confirmed during inspection"
		advice.SuggestedActions = []string{
			"confirm the source sheet exists",
			"confirm the required workbook columns exist",
		}
	case "verification":
		advice.Summary = "the output workbook was written but verification could not confirm the expected behavior"
		advice.SuggestedActions = []string{
			"inspect the verification summary and emitted workbook",
			"correct the request or runtime expectation before rerunning",
		}
	}
	advice.Assumptions = append(advice.Assumptions, "repair advice is deterministic and phase-derived until late-agent runtime wiring is implemented")
	return advice
}

func repairHintForFailure(result RunResult, failure *FailureDetails) string {
	switch failure.FailureClass {
	case "verification_failure":
		if len(result.Verification.Layers) > 0 {
			for _, layer := range result.Verification.Layers {
				if !layer.Pass && len(layer.Reasons) > 0 {
					return "repair " + layer.Name + ": " + layer.Reasons[0]
				}
			}
		}
		if len(result.Verification.Reasons) > 0 {
			return "re-read output workbook and compare verification expectation: " + result.Verification.Reasons[0]
		}
		return "re-read output workbook and compare emitted artifacts against the compiled runtime plan"
	case "policy_failure":
		return "choose a distinct output workbook path and preserve the source workbook"
	case "inspection_failure":
		return "inspect workbook facts and confirm the requested sheet, headers, and ranges exist"
	case "request_resolution_failure":
		return "repair the request reference so the case markdown can be read before compilation"
	default:
		if failure.CriticalStep != "" {
			return "repair failed step: " + failure.CriticalStep
		}
		return "inspect failure evidence and rerun only after the blocking condition is fixed"
	}
}

func writeVerificationReviewArtifact(paths RunPaths, review VerificationReview) error {
	if err := runtimeschema.ValidateStruct(repoJoin("contracts", "results", "verification_review.schema.json"), review); err != nil {
		return err
	}
	return writeJSON(paths.VerificationReviewPath, review)
}

func writeRepairAdviceArtifact(paths RunPaths, advice RepairAdvice) error {
	if err := runtimeschema.ValidateStruct(repoJoin("contracts", "results", "repair_advice.schema.json"), advice); err != nil {
		return err
	}
	return writeJSON(paths.RepairAdvicePath, advice)
}

func writeRenderArtifact(paths RunPaths, outputWorkbook string) error {
	artifact, err := runtimerender.RenderWorkbookPreview(outputWorkbook, filepath.Join(paths.EvidenceDir, "render"))
	if err != nil {
		return err
	}
	if err := runtimeschema.ValidateStruct(repoJoin("contracts", "verification", "render_artifact.schema.json"), artifact); err != nil {
		return err
	}
	return writeJSON(paths.RenderPath, artifact)
}

func writeFailureEvidence(scenarioID, operation string, paths RunPaths, result RunResult, failure *FailureDetails) error {
	if failure == nil {
		return nil
	}
	telemetry := telemetryPaths(paths)
	if !fileExists(paths.RequestPath) {
		return errors.New("request artifact missing for failure evidence")
	}
	if len(telemetry) == 0 {
		return errors.New("telemetry missing for failure evidence")
	}

	evidence := FailureEvidence{
		ScenarioID:     scenarioID,
		RunID:          result.IDs.RunID,
		Operation:      operation,
		Phase:          failure.Phase,
		FailureClass:   failure.FailureClass,
		DomainCode:     failure.DomainCode,
		Message:        failure.Message,
		Outcome:        "fail",
		CriticalStep:   failure.CriticalStep,
		Telemetry:      telemetry,
		RequestPath:    paths.RequestPath,
		RuntimeVersion: runtimeepisode.CurrentRuntimeVersion(),
	}
	if fileExists(paths.TaskSpecPath) {
		evidence.TaskSpecPath = paths.TaskSpecPath
	}
	if fileExists(paths.OperationIRPath) {
		evidence.OperationIRPath = paths.OperationIRPath
	}
	if fileExists(paths.PlanPath) {
		evidence.PlanPath = paths.PlanPath
	}
	if fileExists(paths.PolicyPath) {
		evidence.PolicyPath = paths.PolicyPath
	}
	if fileExists(paths.VerificationPath) {
		evidence.VerificationPath = paths.VerificationPath
	}
	if fileExists(paths.VerificationReviewPath) {
		evidence.VerificationReviewPath = paths.VerificationReviewPath
	}
	if fileExists(paths.ExecutionPath) {
		evidence.ExecutionPath = paths.ExecutionPath
	}
	if fileExists(paths.OutcomePath) {
		evidence.OutcomePath = paths.OutcomePath
	}
	if fileExists(paths.RepairAdvicePath) {
		evidence.RepairAdvicePath = paths.RepairAdvicePath
	}
	if err := runtimeschema.ValidateStruct(repoJoin("contracts", "evidence", "failure_evidence.schema.json"), evidence); err != nil {
		return err
	}
	return writeJSON(paths.FailureEvidencePath, evidence)
}

func telemetryPaths(paths RunPaths) []string {
	all := []string{
		filepath.Join(paths.TelemetryDir, "trace.jsonl"),
		filepath.Join(paths.TelemetryDir, "agent.jsonl"),
		filepath.Join(paths.TelemetryDir, "judgment.jsonl"),
		filepath.Join(paths.TelemetryDir, "file.jsonl"),
		filepath.Join(paths.TelemetryDir, "outcome.jsonl"),
	}
	telemetry := make([]string, 0, len(all))
	for _, path := range all {
		if fileExists(path) {
			telemetry = append(telemetry, path)
		}
	}
	return telemetry
}
