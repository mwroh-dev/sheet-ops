package workbookcase

import (
	"path/filepath"

	runtimeconfig "github.com/mwroh/sheet-ops/runtime/runtimeconfig"
	runtimetaskspec "github.com/mwroh/sheet-ops/runtime/taskspec"
)

func shouldPersistSensitiveWorkbookArtifacts() bool {
	return !runtimeconfig.UsesRedactedArtifacts()
}

func shouldWriteRenderArtifactForRun(failure *FailureDetails) bool {
	switch runtimeconfig.CurrentRenderMode() {
	case runtimeconfig.RenderModeNever:
		return false
	case runtimeconfig.RenderModeOnFailure:
		return failure != nil
	default:
		return true
	}
}

func sanitizeWorkbookArtifactValue(value any) any {
	if !runtimeconfig.UsesRedactedArtifacts() {
		return value
	}

	switch typed := value.(type) {
	case InspectionArtifact:
		clone := typed
		clone.InputFile = runtimeconfig.UseBasename(clone.InputFile)
		return clone
	case SummaryInspectionRecord:
		clone := typed
		clone.InputFile = runtimeconfig.UseBasename(clone.InputFile)
		return clone
	case SummarySheetPlan:
		clone := typed
		clone.OutputFile = runtimeconfig.UseBasename(clone.OutputFile)
		return clone
	case HighlightThresholdPlan:
		clone := typed
		clone.OutputFile = runtimeconfig.UseBasename(clone.OutputFile)
		return clone
	case JoinLookupPlan:
		clone := typed
		clone.OutputFile = runtimeconfig.UseBasename(clone.OutputFile)
		return clone
	case ExecutionSummary:
		clone := typed
		clone.OutputFile = runtimeconfig.UseBasename(clone.OutputFile)
		return clone
	case VerificationResult:
		clone := typed
		clone.OutputFile = runtimeconfig.UseBasename(clone.OutputFile)
		return clone
	case runtimetaskspec.TaskSpec:
		clone := typed
		clone.RequestText = runtimeconfig.RedactedString(clone.RequestText)
		clone.InputWorkbook = runtimeconfig.UseBasename(clone.InputWorkbook)
		clone.OutputWorkbook = runtimeconfig.UseBasename(clone.OutputWorkbook)
		clone.Source.Path = filepath.Base(clone.Source.Path)
		return clone
	default:
		return value
	}
}
