package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	requestcompiler "github.com/mwroh/sheet-ops/runtime/openlayer/requestcompiler"
	runtimeconfig "github.com/mwroh/sheet-ops/runtime/runtimeconfig"
	runtimeworkbookcase "github.com/mwroh/sheet-ops/runtime/workbookcase"
)

type PublicEntryResult struct {
	SchemaVersion   string                 `json:"schema_version"`
	OK              bool                   `json:"ok"`
	Command         string                 `json:"command"`
	Recoverable     bool                   `json:"recoverable"`
	Artifacts       []PublicResultArtifact `json:"artifacts"`
	NextActions     []string               `json:"next_actions"`
	Fingerprints    *executionFingerprints `json:"fingerprints,omitempty"`
	Entry           string                 `json:"entry"`
	Status          string                 `json:"status"`
	WorkUnitID      string                 `json:"work_unit_id"`
	RequestCompiler PublicEntryCompilerRef `json:"request_compiler"`
	Runtime         any                    `json:"runtime"`
}

type PublicResultArtifact struct {
	Kind        string `json:"kind"`
	Path        string `json:"path"`
	Required    bool   `json:"required"`
	SuccessRole string `json:"success_role"`
}

type PublicEntryCompilerRef struct {
	Status               string `json:"status"`
	RequestDir           string `json:"request_dir"`
	DecisionPath         string `json:"decision_path"`
	GeneratedRequestPath string `json:"generated_request_path"`
}

type executionFingerprints struct {
	NormalizedIntentSHA256 string `json:"normalized_intent_sha256"`
	InputWorkbookSHA256    string `json:"input_workbook_sha256"`
	OutputWorkbookSHA256   string `json:"output_workbook_sha256,omitempty"`
}

type PublicRuntimeResult struct {
	IDs                PublicRunIDs                                 `json:"ids"`
	Paths              PublicRunPaths                               `json:"paths"`
	Inspection         runtimeworkbookcase.InspectionArtifact       `json:"inspection"`
	SummaryInspection  *runtimeworkbookcase.SummaryInspectionRecord `json:"summary_inspection"`
	Plan               any                                          `json:"plan"`
	Policy             runtimeworkbookcase.WritePolicyDecision      `json:"policy"`
	Execution          runtimeworkbookcase.ExecutionSummary         `json:"execution"`
	Episode            *runtimeworkbookcase.ExecutionEpisode        `json:"episode"`
	Verification       runtimeworkbookcase.VerificationResult       `json:"verification"`
	VerificationReview *runtimeworkbookcase.VerificationReview      `json:"verification_review,omitempty"`
	RepairAdvice       *runtimeworkbookcase.RepairAdvice            `json:"repair_advice,omitempty"`
	Failure            *runtimeworkbookcase.FailureDetails          `json:"failure"`
}

type PublicRunIDs struct {
	RunID   string `json:"run_id"`
	TraceID string `json:"trace_id"`
}

type PublicRunPaths struct {
	TelemetryDir           string `json:"telemetry_dir"`
	EvidenceDir            string `json:"evidence_dir"`
	ReportDir              string `json:"report_dir"`
	RequestPath            string `json:"request_path"`
	InspectionPath         string `json:"inspection_path"`
	PlanPath               string `json:"plan_path"`
	PolicyPath             string `json:"policy_path"`
	TaskSpecPath           string `json:"task_spec_path"`
	OperationIRPath        string `json:"operation_ir_path"`
	EpisodePath            string `json:"episode_path"`
	ExecutionPath          string `json:"execution_path"`
	VerificationPath       string `json:"verification_path"`
	VerificationReviewPath string `json:"verification_review_path"`
	OutcomePath            string `json:"outcome_path"`
	FailureEvidencePath    string `json:"failure_evidence_path"`
	RepairAdvicePath       string `json:"repair_advice_path"`
	EvidenceIndex          string `json:"evidence_index"`
	ReportPath             string `json:"report_path"`
}

func newPublicRuntimeResult(result runtimeworkbookcase.RunResult) PublicRuntimeResult {
	paths := PublicRunPaths{
		TelemetryDir:           result.Paths.TelemetryDir,
		EvidenceDir:            result.Paths.EvidenceDir,
		ReportDir:              result.Paths.ReportDir,
		RequestPath:            result.Paths.RequestPath,
		InspectionPath:         result.Paths.InspectionPath,
		PlanPath:               result.Paths.PlanPath,
		PolicyPath:             result.Paths.PolicyPath,
		TaskSpecPath:           result.Paths.TaskSpecPath,
		OperationIRPath:        result.Paths.OperationIRPath,
		EpisodePath:            result.Paths.EpisodePath,
		ExecutionPath:          result.Paths.ExecutionPath,
		VerificationPath:       result.Paths.VerificationPath,
		VerificationReviewPath: result.Paths.VerificationReviewPath,
		OutcomePath:            result.Paths.OutcomePath,
		FailureEvidencePath:    result.Paths.FailureEvidencePath,
		RepairAdvicePath:       result.Paths.RepairAdvicePath,
		EvidenceIndex:          result.Paths.EvidenceIndex,
		ReportPath:             result.Paths.ReportPath,
	}
	inspection := result.Inspection
	summaryInspection := result.SummaryInspection
	plan := result.Plan
	execution := result.Execution
	verification := result.Verification
	if runtimeconfig.UsesRedactedArtifacts() {
		paths = redactPublicRunPaths(result.Paths)
		inspection = redactInspectionArtifact(inspection)
		summaryInspection = redactSummaryInspection(summaryInspection)
		plan = redactPlan(plan)
		execution = redactExecutionSummary(execution)
		verification = redactVerificationResult(verification)
	}

	return PublicRuntimeResult{
		IDs: PublicRunIDs{
			RunID:   result.IDs.RunID,
			TraceID: result.IDs.TraceID,
		},
		Paths:              paths,
		Inspection:         inspection,
		SummaryInspection:  summaryInspection,
		Plan:               plan,
		Policy:             result.Policy,
		Execution:          execution,
		Episode:            result.Episode,
		Verification:       verification,
		VerificationReview: loadVerificationReview(result.Paths.VerificationReviewPath),
		RepairAdvice:       loadRepairAdvice(result.Paths.RepairAdvicePath),
		Failure:            result.Failure,
	}
}

func loadVerificationReview(path string) *runtimeworkbookcase.VerificationReview {
	if !publicResultFileExists(path) {
		return nil
	}
	var review runtimeworkbookcase.VerificationReview
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	if err := json.Unmarshal(raw, &review); err != nil {
		return nil
	}
	return &review
}

func loadRepairAdvice(path string) *runtimeworkbookcase.RepairAdvice {
	if !publicResultFileExists(path) {
		return nil
	}
	var advice runtimeworkbookcase.RepairAdvice
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	if err := json.Unmarshal(raw, &advice); err != nil {
		return nil
	}
	return &advice
}

func publicResultFileExists(path string) bool {
	if path == "" {
		return false
	}
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

func newTerminalCompilerResult(entry string, compiled requestcompiler.PersistedResult) (PublicEntryResult, error) {
	status := string(compiled.Decision.Status)
	requestDir := compiled.RequestDir
	decisionPath := compiled.DecisionPath()
	generatedRequestPath := compiled.GeneratedRequestPath()
	if runtimeconfig.UsesRedactedArtifacts() {
		requestDir = "request-compiler"
		decisionPath = filepath.Base(decisionPath)
		if generatedRequestPath != "" {
			generatedRequestPath = filepath.Base(generatedRequestPath)
		}
	}
	return PublicEntryResult{
		SchemaVersion: cliContractSchemaVersion,
		OK:            false,
		Command:       entry,
		Recoverable:   true,
		Artifacts: []PublicResultArtifact{
			{
				Kind:        "compiler_decision",
				Path:        decisionPath,
				Required:    true,
				SuccessRole: "failure_context",
			},
		},
		NextActions: []string{
			"Inspect request_compiler.decision_path.",
			"Resolve the compiler checkpoint or blocked request before retrying.",
		},
		Entry:      entry,
		Status:     status,
		WorkUnitID: compiled.WorkUnitID,
		RequestCompiler: PublicEntryCompilerRef{
			Status:               status,
			RequestDir:           requestDir,
			DecisionPath:         decisionPath,
			GeneratedRequestPath: generatedRequestPath,
		},
		Runtime: nil,
	}, fmt.Errorf("request compiler stopped before runtime execution (status=%s)", compiled.Decision.Status)
}

func newExecutedPublicEntryResult(entry string, compiled requestcompiler.PersistedResult, result runtimeworkbookcase.RunResult, fingerprints executionFingerprints) PublicEntryResult {
	requestDir := compiled.RequestDir
	decisionPath := compiled.DecisionPath()
	generatedRequestPath := compiled.GeneratedRequestPath()
	if runtimeconfig.UsesRedactedArtifacts() {
		requestDir = "request-compiler"
		decisionPath = filepath.Base(decisionPath)
		if generatedRequestPath != "" {
			generatedRequestPath = filepath.Base(generatedRequestPath)
		}
	}
	publicRuntime := newPublicRuntimeResult(result)
	return PublicEntryResult{
		SchemaVersion: cliContractSchemaVersion,
		OK:            result.Verification.Pass,
		Command:       entry,
		Recoverable:   false,
		Artifacts:     publicSuccessArtifacts(publicRuntime),
		NextActions: []string{
			"Inspect runtime.verification before claiming workbook success.",
			"Open the output workbook only after verification pass is true.",
		},
		Fingerprints: &fingerprints,
		Entry:        entry,
		Status:       "executed",
		WorkUnitID:   compiled.WorkUnitID,
		RequestCompiler: PublicEntryCompilerRef{
			Status:               string(requestcompiler.StatusCompiled),
			RequestDir:           requestDir,
			DecisionPath:         decisionPath,
			GeneratedRequestPath: generatedRequestPath,
		},
		Runtime: publicRuntime,
	}
}

func publicSuccessArtifacts(result PublicRuntimeResult) []PublicResultArtifact {
	artifacts := make([]PublicResultArtifact, 0, 6)
	artifacts = appendPublicResultArtifact(artifacts, "output_workbook", result.Verification.OutputFile, true, "primary_success")
	artifacts = appendPublicResultArtifact(artifacts, "verification", result.Paths.VerificationPath, true, "success_evidence")
	artifacts = appendPublicResultArtifact(artifacts, "evidence_dir", result.Paths.EvidenceDir, true, "audit_trail")
	artifacts = appendPublicResultArtifact(artifacts, "execution", result.Paths.ExecutionPath, false, "supporting_evidence")
	artifacts = appendPublicResultArtifact(artifacts, "outcome", result.Paths.OutcomePath, false, "supporting_evidence")
	artifacts = appendPublicResultArtifact(artifacts, "report", result.Paths.ReportPath, false, "human_review")
	return artifacts
}

func appendPublicResultArtifact(artifacts []PublicResultArtifact, kind, path string, required bool, role string) []PublicResultArtifact {
	if path == "" {
		return artifacts
	}
	return append(artifacts, PublicResultArtifact{
		Kind:        kind,
		Path:        path,
		Required:    required,
		SuccessRole: role,
	})
}

func redactPublicRunPaths(paths runtimeworkbookcase.RunPaths) PublicRunPaths {
	return PublicRunPaths{
		TelemetryDir:           "",
		EvidenceDir:            runtimeconfig.UseBasename(paths.EvidenceDir),
		ReportDir:              "",
		RequestPath:            "",
		InspectionPath:         visibleArtifactPath(paths.InspectionPath),
		PlanPath:               visibleArtifactPath(paths.PlanPath),
		PolicyPath:             visibleArtifactPath(paths.PolicyPath),
		TaskSpecPath:           "",
		OperationIRPath:        "",
		EpisodePath:            "",
		ExecutionPath:          visibleArtifactPath(paths.ExecutionPath),
		VerificationPath:       visibleArtifactPath(paths.VerificationPath),
		VerificationReviewPath: visibleArtifactPath(paths.VerificationReviewPath),
		OutcomePath:            visibleArtifactPath(paths.OutcomePath),
		FailureEvidencePath:    "",
		RepairAdvicePath:       visibleArtifactPath(paths.RepairAdvicePath),
		EvidenceIndex:          "",
		ReportPath:             "",
	}
}

func visibleArtifactPath(path string) string {
	if !publicResultFileExists(path) {
		return ""
	}
	return filepath.Base(path)
}

func redactInspectionArtifact(value runtimeworkbookcase.InspectionArtifact) runtimeworkbookcase.InspectionArtifact {
	value.InputFile = runtimeconfig.UseBasename(value.InputFile)
	return value
}

func redactSummaryInspection(value *runtimeworkbookcase.SummaryInspectionRecord) *runtimeworkbookcase.SummaryInspectionRecord {
	if value == nil {
		return nil
	}
	clone := *value
	clone.InputFile = runtimeconfig.UseBasename(clone.InputFile)
	return &clone
}

func redactExecutionSummary(value runtimeworkbookcase.ExecutionSummary) runtimeworkbookcase.ExecutionSummary {
	value.OutputFile = runtimeconfig.UseBasename(value.OutputFile)
	return value
}

func redactVerificationResult(value runtimeworkbookcase.VerificationResult) runtimeworkbookcase.VerificationResult {
	value.OutputFile = runtimeconfig.UseBasename(value.OutputFile)
	return value
}

func redactPlan(value any) any {
	switch typed := value.(type) {
	case runtimeworkbookcase.SummarySheetPlan:
		typed.OutputFile = runtimeconfig.UseBasename(typed.OutputFile)
		return typed
	case runtimeworkbookcase.HighlightThresholdPlan:
		typed.OutputFile = runtimeconfig.UseBasename(typed.OutputFile)
		return typed
	case runtimeworkbookcase.JoinLookupPlan:
		typed.OutputFile = runtimeconfig.UseBasename(typed.OutputFile)
		return typed
	default:
		return value
	}
}

func writePublicEntryResultJSON(output io.Writer, result PublicEntryResult) error {
	encoder := json.NewEncoder(output)
	encoder.SetIndent("", "  ")
	return encoder.Encode(result)
}
