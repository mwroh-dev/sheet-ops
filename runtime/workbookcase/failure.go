package workbookcase

import (
	"errors"
	"fmt"
	"strings"

	runtimetaskspec "github.com/mwroh/sheet-ops/runtime/taskspec"
)

const (
	failureClassRequestValidation = "request_validation_failure"
	failureClassRequestResolution = "request_resolution_failure"
	failureClassInspection        = "inspection_failure"
	failureClassPlanning          = "planning_failure"
	failureClassPolicy            = "policy_failure"
	failureClassExecution         = "execution_failure"
	failureClassVerification      = "verification_failure"
	failureClassArtifactEmission  = "artifact_emission_failure"
)

const (
	domainCodeVerificationFailed         = "verify_failed"
	domainCodeVerificationOutputMismatch = "verification_output_mismatch"
)

type phaseError struct {
	failure FailureDetails
	err     error
}

func (err *phaseError) Error() string {
	return err.err.Error()
}

func (err *phaseError) Unwrap() error {
	return err.err
}

func wrapFailure(phase, failureClass, domainCode, criticalStep string, err error) error {
	if err == nil {
		return nil
	}
	if existing := extractFailure(err); existing != nil {
		return err
	}
	return errorFromFailure(FailureDetails{
		Phase:        phase,
		FailureClass: normalizeFailureClass(phase, failureClass),
		DomainCode:   normalizeDomainCode(phase, failureClass, domainCode),
		CriticalStep: criticalStep,
		Message:      err.Error(),
	}, err)
}

func newFailure(phase, failureClass, domainCode, criticalStep, message string) *FailureDetails {
	return &FailureDetails{
		Phase:        phase,
		FailureClass: normalizeFailureClass(phase, failureClass),
		DomainCode:   normalizeDomainCode(phase, failureClass, domainCode),
		CriticalStep: criticalStep,
		Message:      message,
	}
}

func normalizeFailureClass(phase, failureClass string) string {
	switch failureClass {
	case "routing_failure":
		return failureClassRequestResolution
	case failureClassRequestValidation,
		failureClassRequestResolution,
		failureClassInspection,
		failureClassPlanning,
		failureClassPolicy,
		failureClassExecution,
		failureClassVerification,
		failureClassArtifactEmission:
		return failureClass
	default:
		return fallbackFailureClassForPhase(phase)
	}
}

func fallbackFailureClassForPhase(phase string) string {
	switch phase {
	case "request_validation":
		return failureClassRequestValidation
	case "request_resolution", "knowledge_read":
		return failureClassRequestResolution
	case "inspection":
		return failureClassInspection
	case "planning":
		return failureClassPlanning
	case "policy":
		return failureClassPolicy
	case "execution":
		return failureClassExecution
	case "verification":
		return failureClassVerification
	case "artifact_emission", "knowledge_write", "finalization":
		return failureClassArtifactEmission
	default:
		return failureClassArtifactEmission
	}
}

func normalizeDomainCode(phase, failureClass, domainCode string) string {
	if normalizeFailureClass(phase, failureClass) == failureClassVerification && phase == "verification" && domainCode == domainCodeVerificationOutputMismatch {
		return domainCodeVerificationOutputMismatch
	}
	if normalizeFailureClass(phase, failureClass) == failureClassVerification && phase == "verification" && domainCode == domainCodeVerificationFailed {
		return domainCodeVerificationFailed
	}
	if normalizeFailureClass(phase, failureClass) == failureClassVerification && phase == "verification" && domainCode == "verification_failed" {
		return domainCodeVerificationOutputMismatch
	}
	return domainCode
}

func errorFromFailure(failure FailureDetails, err error) error {
	if err == nil {
		err = errors.New(failure.Message)
	}
	return &phaseError{
		failure: failure,
		err:     err,
	}
}

func extractFailure(err error) *FailureDetails {
	var phaseErr *phaseError
	if !errors.As(err, &phaseErr) {
		return nil
	}
	failure := phaseErr.failure
	if strings.TrimSpace(failure.Message) == "" {
		failure.Message = phaseErr.err.Error()
	}
	return &failure
}

func bestEffortRequestText(taskSpec runtimetaskspec.TaskSpec, resolved string) string {
	if strings.TrimSpace(resolved) != "" {
		return resolved
	}
	if strings.TrimSpace(taskSpec.RequestText) != "" {
		return taskSpec.RequestText
	}

	lines := make([]string, 0, 5)
	if strings.TrimSpace(taskSpec.Source.Kind) != "" {
		lines = append(lines, fmt.Sprintf("request unavailable; source kind: %s", taskSpec.Source.Kind))
	}
	if strings.TrimSpace(taskSpec.Source.Path) != "" {
		lines = append(lines, fmt.Sprintf("source path: %s", taskSpec.Source.Path))
	}
	if strings.TrimSpace(taskSpec.Operation) != "" {
		lines = append(lines, fmt.Sprintf("operation: %s", taskSpec.Operation))
	}
	if strings.TrimSpace(taskSpec.InputWorkbook) != "" {
		lines = append(lines, fmt.Sprintf("input workbook: %s", taskSpec.InputWorkbook))
	}
	if strings.TrimSpace(taskSpec.OutputWorkbook) != "" {
		lines = append(lines, fmt.Sprintf("output workbook: %s", taskSpec.OutputWorkbook))
	}
	if len(lines) == 0 {
		return "request unavailable"
	}
	return strings.Join(lines, "\n") + "\n"
}

func failureFromVerification(result RunResult) *FailureDetails {
	if result.Verification.Operation == "" || result.Verification.Pass {
		return nil
	}
	message := strings.Join(result.Verification.Reasons, "; ")
	if strings.TrimSpace(message) == "" {
		message = "verification reported a failed outcome"
	}
	return newFailure(
		"verification",
		failureClassVerification,
		domainCodeVerificationOutputMismatch,
		"verify emitted workbook artifacts against the compiled runtime plan",
		message,
	)
}

func outcomeReasons(result RunResult, failure *FailureDetails) []string {
	if len(result.Verification.Reasons) > 0 {
		return append([]string(nil), result.Verification.Reasons...)
	}
	if len(result.Policy.Reasons) > 0 {
		return append([]string(nil), result.Policy.Reasons...)
	}
	if failure != nil && strings.TrimSpace(failure.Message) != "" {
		return []string{failure.Message}
	}
	return nil
}
