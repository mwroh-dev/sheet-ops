package useorchestrator

import (
	runtimesubagent "github.com/mwroh/sheet-ops/runtime/subagent"
	runtimeworkbookcase "github.com/mwroh/sheet-ops/runtime/workbookcase"
)

type FilterSpec struct {
	Column string `json:"column"`
	Op     string `json:"op"`
	Value  string `json:"value"`
}

type MetricSpec struct {
	Column string `json:"column"`
	Op     string `json:"op"`
	As     string `json:"as"`
}

type TelemetryEvent = runtimeworkbookcase.TelemetryEvent
type RunIDs = runtimeworkbookcase.RunIDs
type RunPaths = runtimeworkbookcase.RunPaths
type RunResult = runtimeworkbookcase.RunResult

type OrchestratorDecision struct {
	ScenarioID                string                            `json:"scenario_id"`
	Decision                  string                            `json:"decision"`
	RequestCompilerLoopState  runtimesubagent.SubagentLoopState `json:"request_compiler_loop_state"`
	ValidatedExecutionRequest *ValidatedExecutionRequest        `json:"validated_execution_request,omitempty"`
	RepairAdvice              *runtimeworkbookcase.RepairAdvice `json:"repair_advice,omitempty"`
}

type ResultVerifierOutcome struct {
	Review    runtimeworkbookcase.VerificationReview `json:"review"`
	LoopState runtimesubagent.SubagentLoopState      `json:"loop_state"`
}
