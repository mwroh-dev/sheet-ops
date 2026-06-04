package workbookcase

import (
	"fmt"
	"strings"
	"time"

	runtimecompiler "github.com/mwroh/sheet-ops/runtime/compiler"
	runtimeepisode "github.com/mwroh/sheet-ops/runtime/episode"
	runtimeexecute "github.com/mwroh/sheet-ops/runtime/execute"
	runtimepolicy "github.com/mwroh/sheet-ops/runtime/policy"
	runtimeschema "github.com/mwroh/sheet-ops/runtime/schema"
	runtimetaskspec "github.com/mwroh/sheet-ops/runtime/taskspec"
	runtimeverify "github.com/mwroh/sheet-ops/runtime/verify"
)

var (
	runSummaryPlanWithArtifacts           = runtimeexecute.RunGroupSummarizePlan
	verifySummaryExecutionWithArtifacts   = runtimeverify.VerifyGroupSummarizeExecution
	runHighlightWithArtifacts             = runtimeexecute.RunHighlightThresholdWithInspection
	verifyHighlightExecutionWithArtifacts = runtimeverify.VerifyHighlightThresholdExecution
)

func Run(req Request) (result RunResult, err error) {
	if strings.TrimSpace(req.ScenarioID) == "" {
		return result, fmt.Errorf("scenario id must not be empty")
	}

	ids := NewRunIDs()
	paths, err := BuildRunPaths(req.ScenarioID, ids)
	if err != nil {
		return result, err
	}
	result.IDs = ids
	result.Paths = paths
	if err := EnsureRunPaths(paths); err != nil {
		return result, err
	}

	requestText := bestEffortRequestText(req.TaskSpec, "")
	defer func() {
		err = finalizeRun(req, requestText, &result, err)
	}()

	mustLog := func(category, eventType, component string, payload map[string]any) {
		event := TelemetryEvent{
			RunID:     ids.RunID,
			TraceID:   ids.TraceID,
			EventType: eventType,
			Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
			Component: component,
			Payload:   payload,
		}
		_ = AppendTelemetry(paths, category, event)
	}

	routing := req.Routing
	if routing == nil {
		routing = &RoutingContext{
			SelectedOperation: req.TaskSpec.Operation,
			RoutingAuthority:  "task_spec",
			RoutingNote:       "Closed runtime received a validated task spec directly.",
		}
	}

	mustLog("trace", "run_started", "orchestrator", map[string]any{
		"scenario_id": req.ScenarioID,
		"input_file":  req.TaskSpec.InputWorkbook,
		"output_file": req.TaskSpec.OutputWorkbook,
	})
	mustLog("judgment", "routing_context_loaded", "request-router", routing.Payload())
	mustLog("agent", "agent_started", "use-orchestrator", map[string]any{"role": "orchestrator"})

	if err := runtimeschema.ValidateStruct(repoJoin("contracts", "task", "task_spec.schema.json"), req.TaskSpec); err != nil {
		return result, wrapFailure(
			"request_validation",
			"request_validation_failure",
			"task_spec_invalid",
			"validate task spec against the workbook-case contract",
			err,
		)
	}
	requestText, err = resolveRequestText(req.TaskSpec)
	if err != nil {
		return result, wrapFailure(
			"request_resolution",
			"request_resolution_failure",
			"case_markdown_read_failed",
			"read the case markdown request source",
			err,
		)
	}
	requestText = bestEffortRequestText(req.TaskSpec, requestText)

	switch req.TaskSpec.Operation {
	case SummaryOperationName:
		if err := orchestrateSummary(req.TaskSpec, paths, &result, mustLog); err != nil {
			return result, err
		}
	case HighlightOperationName:
		if err := orchestrateHighlight(req.TaskSpec, paths, &result, mustLog); err != nil {
			return result, err
		}
	case JoinLookupOperationName:
		if err := orchestrateJoinLookup(req.TaskSpec, paths, &result, mustLog); err != nil {
			return result, err
		}
	default:
		return result, wrapFailure(
			"request_validation",
			"routing_failure",
			"unsupported_operation",
			"select a supported workbookcase operation",
			fmt.Errorf("unsupported operation %q", req.TaskSpec.Operation),
		)
	}
	mustLog("agent", "agent_completed", "use-orchestrator", map[string]any{"pass": result.Verification.Pass})
	return result, nil
}

func orchestrateJoinLookup(taskSpec runtimetaskspec.TaskSpec, paths RunPaths, result *RunResult, mustLog func(string, string, string, map[string]any)) error {
	task := joinLookupTaskFromSpec(taskSpec)
	inspection, err := runtimecompiler.InspectWorkbook(task.TaskSpec.InputWorkbook, task.SourceSheet, nil, task.IncludeSourceColumns, nil)
	if err != nil {
		return wrapFailure(
			"inspection",
			"inspection_failure",
			"inspect_failed",
			"inspect the source workbook before planning join lookup execution",
			err,
		)
	}

	result.Inspection = summaryInspectionArtifact(inspection, task.TargetSheet)
	record := summaryInspectionRecord(inspection, task.TargetSheet)
	result.SummaryInspection = &record
	mustLog("agent", "agent_completed", "inspect-workbook", map[string]any{
		"source_sheet":         record.SourceSheet,
		"planned_target_sheet": record.PlannedTargetSheet,
	})
	if err := writeJSON(paths.InspectionPath, record); err != nil {
		return wrapFailure(
			"inspection",
			"artifact_emission_failure",
			"inspection_artifact_write_failed",
			"write the join lookup inspection artifact",
			err,
		)
	}

	operationIR, err := runtimecompiler.CompileJoinLookupOperation(task)
	if err != nil {
		return wrapFailure(
			"planning",
			"planning_failure",
			"operation_ir_compile_failed",
			"compile workbook join lookup operation IR",
			err,
		)
	}
	plan := JoinLookupPlan{
		Operation:            JoinLookupOperationName,
		SheetName:            operationIR.SourceSheet,
		LookupSheet:          operationIR.LookupSheet,
		JoinKey:              operationIR.JoinKey,
		TargetSheet:          operationIR.TargetSheet,
		IncludeSourceColumns: append([]string(nil), operationIR.IncludeSourceColumns...),
		AppendLookupColumns:  append([]string(nil), operationIR.AppendLookupColumns...),
		PreserveOriginal:     operationIR.PreserveOriginal,
		OutputFile:           task.TaskSpec.OutputWorkbook,
	}
	if err := runtimeschema.ValidateStruct(repoJoin("contracts", "plans", "use", "operation_plan.schema.json"), plan); err != nil {
		return wrapFailure(
			"planning",
			"planning_failure",
			"plan_invalid",
			"validate the join lookup plan artifact against its contract",
			err,
		)
	}
	result.Plan = plan
	if err := writeJSON(paths.PlanPath, plan); err != nil {
		return wrapFailure(
			"planning",
			"artifact_emission_failure",
			"plan_artifact_write_failed",
			"write the join lookup plan artifact",
			err,
		)
	}
	mustLog("judgment", "plan_built", "request-planner", map[string]any{
		"operation":    plan.Operation,
		"source_sheet": plan.SheetName,
		"lookup_sheet": plan.LookupSheet,
		"target_sheet": plan.TargetSheet,
		"join_key":     plan.JoinKey,
	})
	if err := writeJoinLookupRuntimeInputs(task.TaskSpec, operationIR, paths); err != nil {
		return err
	}

	policy := evaluateWritePolicyOutput(task.TaskSpec.InputWorkbook, task.TaskSpec.OutputWorkbook, task.PreserveOriginal)
	if err := runtimeschema.ValidateStruct(repoJoin("contracts", "policy", "write_policy_decision.schema.json"), policy); err != nil {
		return wrapFailure(
			"policy",
			"policy_failure",
			"policy_decision_invalid",
			"validate the join lookup write policy decision",
			err,
		)
	}
	result.Policy = policy
	mustLog("judgment", "policy_decided", "original-preservation", map[string]any{
		"allow":             policy.Allow,
		"approval_required": policy.ApprovalRequired,
		"preserve_original": policy.PreserveOriginal,
	})
	if err := writeJSON(paths.PolicyPath, policy); err != nil {
		return wrapFailure(
			"policy",
			"artifact_emission_failure",
			"policy_artifact_write_failed",
			"write the join lookup policy decision artifact",
			err,
		)
	}
	if !policy.Allow {
		return wrapFailure(
			"policy",
			"policy_failure",
			"policy_denied",
			"evaluate whether the join lookup run may write a distinct output workbook",
			fmt.Errorf("policy denied execution: %s", strings.Join(policy.Reasons, "; ")),
		)
	}

	execution, err := runtimeexecute.RunWorkbookOperation(operationIR, task.TaskSpec.InputWorkbook, task.TaskSpec.OutputWorkbook)
	if err != nil {
		return wrapFailure(
			"execution",
			"execution_failure",
			"execute_failed",
			"execute the join lookup workbook operation and write the output workbook",
			err,
		)
	}
	result.Execution = executionSummaryFromRuntime(execution)
	mustLog("file", "output_written", result.Execution.Operation, map[string]any{
		"output_file":   result.Execution.OutputFile,
		"summary_sheet": result.Execution.SummarySheet,
		"summary_rows":  result.Execution.SummaryRows,
	})
	if err := writeJSON(paths.ExecutionPath, result.Execution); err != nil {
		return wrapFailure(
			"execution",
			"artifact_emission_failure",
			"execution_artifact_write_failed",
			"write the join lookup execution artifact",
			err,
		)
	}

	verification, err := runtimeverify.VerifyWorkbookOperation(
		operationIR,
		task.TaskSpec.InputWorkbook,
		task.TaskSpec.OutputWorkbook,
		result.Execution.SourceSHA256Before,
		result.Execution.SourceSHA256After,
	)
	if err != nil {
		return wrapFailure(
			"verification",
			"verification_failure",
			"verify_failed",
			"verify the emitted join lookup workbook against execution fingerprints",
			err,
		)
	}
	result.Verification = verificationSummaryFromRuntime(task.TaskSpec.Operation, verification)
	if err := runtimeschema.ValidateStruct(repoJoin("contracts", "verification", "verification_result.schema.json"), result.Verification); err != nil {
		return wrapFailure(
			"verification",
			"verification_failure",
			"verification_result_invalid",
			"validate the join lookup verification artifact against its contract",
			err,
		)
	}
	mustLog("judgment", "verification_completed", "result-verifier", map[string]any{
		"pass":          result.Verification.Pass,
		"operation":     result.Verification.Operation,
		"summary_sheet": result.Verification.SummarySheet,
		"summary_rows":  result.Verification.SummaryRows,
	})
	if err := writeJSON(paths.VerificationPath, result.Verification); err != nil {
		return wrapFailure(
			"verification",
			"artifact_emission_failure",
			"verification_artifact_write_failed",
			"write the join lookup verification artifact",
			err,
		)
	}
	return nil
}

func orchestrateSummary(taskSpec runtimetaskspec.TaskSpec, paths RunPaths, result *RunResult, mustLog func(string, string, string, map[string]any)) error {
	task := summaryTaskFromSpec(taskSpec)
	inspection, err := runtimecompiler.InspectWorkbook(
		task.TaskSpec.InputWorkbook,
		task.SourceSheet,
		task.Filters,
		task.GroupBy,
		task.Metrics,
	)
	if err != nil {
		return wrapFailure(
			"inspection",
			"inspection_failure",
			"inspect_failed",
			"inspect the source workbook before planning summary execution",
			err,
		)
	}

	result.Inspection = summaryInspectionArtifact(inspection, task.TargetSheet)
	record := summaryInspectionRecord(inspection, task.TargetSheet)
	result.SummaryInspection = &record
	mustLog("agent", "agent_completed", "inspect-workbook", map[string]any{
		"source_sheet":         record.SourceSheet,
		"planned_target_sheet": record.PlannedTargetSheet,
	})
	if err := writeJSON(paths.InspectionPath, record); err != nil {
		return wrapFailure(
			"inspection",
			"artifact_emission_failure",
			"inspection_artifact_write_failed",
			"write the summary inspection artifact",
			err,
		)
	}

	operationIR, err := runtimecompiler.CompileWorkbookOperation(task, inspection)
	if err != nil {
		return wrapFailure(
			"planning",
			"planning_failure",
			"operation_ir_compile_failed",
			"compile workbook summary operation IR from inspection",
			err,
		)
	}
	plan, err := runtimecompiler.BuildGroupSummarizePlan(inspection, operationIR)
	if err != nil {
		return wrapFailure(
			"planning",
			"planning_failure",
			"plan_build_failed",
			"build the summary execution plan from the compiled IR",
			err,
		)
	}

	summaryPlan := SummarySheetPlan{
		Operation:        SummaryOperationName,
		SheetName:        operationIR.SourceSheet,
		SummaryMode:      operationIR.SummaryMode,
		Filters:          nonNilFilters(operationIR.Filters),
		GroupBy:          append([]string(nil), operationIR.GroupBy...),
		Metrics:          cloneMetrics(operationIR.Metrics),
		TargetSheet:      operationIR.TargetSheet,
		PreserveOriginal: task.PreserveOriginal,
		OutputFile:       task.TaskSpec.OutputWorkbook,
	}
	if err := runtimeschema.ValidateStruct(repoJoin("contracts", "plans", "use", "operation_plan.schema.json"), summaryPlan); err != nil {
		return wrapFailure(
			"planning",
			"planning_failure",
			"plan_invalid",
			"validate the summary plan artifact against its contract",
			err,
		)
	}
	result.Plan = summaryPlan
	mustLog("judgment", "plan_built", "request-planner", map[string]any{
		"operation":    summaryPlan.Operation,
		"source_sheet": summaryPlan.SheetName,
		"target_sheet": summaryPlan.TargetSheet,
		"group_by":     summaryPlan.GroupBy,
	})
	if err := writeJSON(paths.PlanPath, summaryPlan); err != nil {
		return wrapFailure(
			"planning",
			"artifact_emission_failure",
			"plan_artifact_write_failed",
			"write the summary plan artifact",
			err,
		)
	}
	if err := writeSummaryRuntimeInputs(task.TaskSpec, operationIR, paths); err != nil {
		return err
	}

	policyDecision := runtimepolicy.EvaluateOriginalPreservation(task)
	policy := WritePolicyDecision{
		Allow:            policyDecision.Allow,
		PreserveOriginal: policyDecision.PreserveOriginal,
		ApprovalRequired: policyDecision.ApprovalRequired,
		Reasons:          append([]string(nil), policyDecision.Reasons...),
	}
	if err := runtimeschema.ValidateStruct(repoJoin("contracts", "policy", "write_policy_decision.schema.json"), policy); err != nil {
		return wrapFailure(
			"policy",
			"policy_failure",
			"policy_decision_invalid",
			"validate the summary write policy decision",
			err,
		)
	}
	result.Policy = policy
	mustLog("judgment", "policy_decided", "original-preservation", map[string]any{
		"allow":             policy.Allow,
		"approval_required": policy.ApprovalRequired,
		"preserve_original": policy.PreserveOriginal,
	})
	if err := writeJSON(paths.PolicyPath, policy); err != nil {
		return wrapFailure(
			"policy",
			"artifact_emission_failure",
			"policy_artifact_write_failed",
			"write the summary policy decision artifact",
			err,
		)
	}
	if !policy.Allow {
		return wrapFailure(
			"policy",
			"policy_failure",
			"policy_denied",
			"evaluate whether the summary run may write a distinct output workbook",
			fmt.Errorf("policy denied execution: %s", strings.Join(policy.Reasons, "; ")),
		)
	}

	execution, err := runSummaryPlanWithArtifacts(plan, task.TaskSpec.OutputWorkbook)
	if err != nil {
		return wrapFailure(
			"execution",
			"execution_failure",
			"execute_failed",
			"execute the summary workbook plan and write the output workbook",
			err,
		)
	}
	result.Execution = executionSummaryFromRuntime(execution)
	mustLog("file", "output_written", result.Execution.Operation, map[string]any{
		"output_file":      result.Execution.OutputFile,
		"highlighted_rows": result.Execution.HighlightedRows,
		"summary_sheet":    result.Execution.SummarySheet,
		"summary_rows":     result.Execution.SummaryRows,
	})
	if err := writeJSON(paths.ExecutionPath, result.Execution); err != nil {
		return wrapFailure(
			"execution",
			"artifact_emission_failure",
			"execution_artifact_write_failed",
			"write the summary execution artifact",
			err,
		)
	}

	verification, err := verifySummaryExecutionWithArtifacts(
		plan,
		result.Execution.OutputFile,
		result.Execution.SourceSHA256Before,
		result.Execution.SourceSHA256After,
	)
	if err != nil {
		return wrapFailure(
			"verification",
			"verification_failure",
			"verify_failed",
			"verify the emitted summary workbook against execution fingerprints",
			err,
		)
	}
	verification = runtimeverify.WithDefaultLayers(plan.Operation, verification)
	result.Verification = summaryVerificationFromRuntime(verification)
	if err := runtimeschema.ValidateStruct(repoJoin("contracts", "verification", "verification_result.schema.json"), result.Verification); err != nil {
		return wrapFailure(
			"verification",
			"verification_failure",
			"verification_result_invalid",
			"validate the summary verification artifact against its contract",
			err,
		)
	}
	mustLog("judgment", "verification_completed", "result-verifier", map[string]any{
		"pass":             result.Verification.Pass,
		"operation":        result.Verification.Operation,
		"highlighted_rows": result.Verification.HighlightedRows,
		"summary_sheet":    result.Verification.SummarySheet,
		"summary_rows":     result.Verification.SummaryRows,
	})
	if err := writeJSON(paths.VerificationPath, result.Verification); err != nil {
		return wrapFailure(
			"verification",
			"artifact_emission_failure",
			"verification_artifact_write_failed",
			"write the summary verification artifact",
			err,
		)
	}

	episodeArtifact := runtimeepisode.NewExecutionEpisodeForPaths(summaryExecutionArtifactPaths(paths), result.Verification.Pass)
	if shouldPersistSensitiveWorkbookArtifacts() {
		if err := runtimeschema.ValidateStruct(repoJoin("contracts", "evidence", "execution_episode.schema.json"), episodeArtifact); err != nil {
			return wrapFailure(
				"verification",
				"artifact_emission_failure",
				"episode_artifact_invalid",
				"validate the summary execution episode artifact",
				err,
			)
		}
		if err := runtimeepisode.WriteExecutionEpisode(paths.EpisodePath, episodeArtifact); err != nil {
			return wrapFailure(
				"verification",
				"artifact_emission_failure",
				"episode_artifact_write_failed",
				"write the summary execution episode artifact",
				err,
			)
		}
		result.Episode = &episodeArtifact
	}
	return nil
}

func nonNilFilters(values []runtimecompiler.FilterSpec) []FilterSpec {
	filters := cloneFilters(values)
	if filters == nil {
		return []FilterSpec{}
	}
	return filters
}

func orchestrateHighlight(taskSpec runtimetaskspec.TaskSpec, paths RunPaths, result *RunResult, mustLog func(string, string, string, map[string]any)) error {
	task := highlightTaskFromSpec(taskSpec)
	inspection, err := runtimecompiler.InspectHighlightThresholdWorkbook(
		task.TaskSpec.InputWorkbook,
		task.SourceSheet,
		task.ThresholdRule.Column,
	)
	if err != nil {
		return wrapFailure(
			"inspection",
			"inspection_failure",
			"inspect_failed",
			"inspect the source workbook before planning highlight execution",
			err,
		)
	}

	result.Inspection = highlightInspectionArtifact(inspection)
	mustLog("agent", "agent_completed", "inspect-workbook", map[string]any{"sheet": inspection.SourceSheet})
	if err := writeJSON(paths.InspectionPath, result.Inspection); err != nil {
		return wrapFailure(
			"inspection",
			"artifact_emission_failure",
			"inspection_artifact_write_failed",
			"write the highlight inspection artifact",
			err,
		)
	}

	operationIR, err := runtimecompiler.CompileHighlightThresholdOperation(task, inspection)
	if err != nil {
		return wrapFailure(
			"planning",
			"planning_failure",
			"operation_ir_compile_failed",
			"compile workbook highlight operation IR from inspection",
			err,
		)
	}
	highlightPlan := HighlightThresholdPlan{
		Operation:        HighlightOperationName,
		SheetName:        operationIR.SourceSheet,
		TargetColumn:     operationIR.TargetColumn,
		Operator:         operationIR.Operator,
		Threshold:        derefThreshold(operationIR.Threshold),
		HighlightColor:   operationIR.HighlightColor,
		PreserveOriginal: operationIR.PreserveOriginal,
		OutputFile:       task.TaskSpec.OutputWorkbook,
	}
	if err := runtimeschema.ValidateStruct(repoJoin("contracts", "plans", "use", "operation_plan.schema.json"), highlightPlan); err != nil {
		return wrapFailure(
			"planning",
			"planning_failure",
			"plan_invalid",
			"validate the highlight plan artifact against its contract",
			err,
		)
	}
	result.Plan = highlightPlan
	if err := writeJSON(paths.PlanPath, highlightPlan); err != nil {
		return wrapFailure(
			"planning",
			"artifact_emission_failure",
			"plan_artifact_write_failed",
			"write the highlight plan artifact",
			err,
		)
	}
	mustLog("judgment", "plan_built", "request-planner", map[string]any{
		"operation":     highlightPlan.Operation,
		"target_column": highlightPlan.TargetColumn,
		"threshold":     highlightPlan.Threshold,
	})
	if err := writeHighlightRuntimeInputs(task.TaskSpec, operationIR, paths); err != nil {
		return err
	}

	policy := evaluateWritePolicyOutput(task.TaskSpec.InputWorkbook, task.TaskSpec.OutputWorkbook, task.PreserveOriginal)
	if err := runtimeschema.ValidateStruct(repoJoin("contracts", "policy", "write_policy_decision.schema.json"), policy); err != nil {
		return wrapFailure(
			"policy",
			"policy_failure",
			"policy_decision_invalid",
			"validate the highlight write policy decision",
			err,
		)
	}
	result.Policy = policy
	mustLog("judgment", "policy_decided", "original-preservation", map[string]any{
		"allow":             policy.Allow,
		"approval_required": policy.ApprovalRequired,
		"preserve_original": policy.PreserveOriginal,
	})
	if err := writeJSON(paths.PolicyPath, policy); err != nil {
		return wrapFailure(
			"policy",
			"artifact_emission_failure",
			"policy_artifact_write_failed",
			"write the highlight policy decision artifact",
			err,
		)
	}
	if !policy.Allow {
		return wrapFailure(
			"policy",
			"policy_failure",
			"policy_denied",
			"evaluate whether the highlight run may write a distinct output workbook",
			fmt.Errorf("policy denied execution: %s", strings.Join(policy.Reasons, "; ")),
		)
	}

	execution, err := runHighlightWithArtifacts(inspection, operationIR, task.TaskSpec.OutputWorkbook)
	if err != nil {
		return wrapFailure(
			"execution",
			"execution_failure",
			"execute_failed",
			"execute the highlight workbook plan and write the output workbook",
			err,
		)
	}
	result.Execution = executionSummaryFromRuntime(execution)
	mustLog("file", "output_written", result.Execution.Operation, map[string]any{
		"output_file":      result.Execution.OutputFile,
		"highlighted_rows": result.Execution.HighlightedRows,
		"summary_sheet":    result.Execution.SummarySheet,
		"summary_rows":     result.Execution.SummaryRows,
	})
	if err := writeJSON(paths.ExecutionPath, result.Execution); err != nil {
		return wrapFailure(
			"execution",
			"artifact_emission_failure",
			"execution_artifact_write_failed",
			"write the highlight execution artifact",
			err,
		)
	}

	verification, err := verifyHighlightExecutionWithArtifacts(
		inspection,
		operationIR,
		task.TaskSpec.OutputWorkbook,
		result.Execution.SourceSHA256Before,
		result.Execution.SourceSHA256After,
	)
	if err != nil {
		return wrapFailure(
			"verification",
			"verification_failure",
			"verify_failed",
			"verify the emitted highlight workbook against execution fingerprints",
			err,
		)
	}
	verification = runtimeverify.WithDefaultLayers(operationIR, verification)
	result.Verification = verificationSummaryFromRuntime(task.TaskSpec.Operation, verification)
	if err := runtimeschema.ValidateStruct(repoJoin("contracts", "verification", "verification_result.schema.json"), result.Verification); err != nil {
		return wrapFailure(
			"verification",
			"verification_failure",
			"verification_result_invalid",
			"validate the highlight verification artifact against its contract",
			err,
		)
	}
	mustLog("judgment", "verification_completed", "result-verifier", map[string]any{
		"pass":             result.Verification.Pass,
		"operation":        result.Verification.Operation,
		"highlighted_rows": result.Verification.HighlightedRows,
		"summary_sheet":    result.Verification.SummarySheet,
		"summary_rows":     result.Verification.SummaryRows,
	})
	if err := writeJSON(paths.VerificationPath, result.Verification); err != nil {
		return wrapFailure(
			"verification",
			"artifact_emission_failure",
			"verification_artifact_write_failed",
			"write the highlight verification artifact",
			err,
		)
	}
	return nil
}

func writeSummaryRuntimeInputs(taskSpec runtimetaskspec.TaskSpec, operationIR runtimecompiler.WorkbookOperationIR, paths RunPaths) error {
	if !shouldPersistSensitiveWorkbookArtifacts() {
		return nil
	}
	if err := runtimeschema.ValidateStruct(repoJoin("contracts", "task", "task_spec.schema.json"), taskSpec); err != nil {
		return wrapFailure(
			"planning",
			"planning_failure",
			"task_spec_invalid",
			"validate the summary runtime task spec artifact",
			err,
		)
	}
	if err := writeJSON(paths.TaskSpecPath, taskSpec); err != nil {
		return wrapFailure(
			"planning",
			"artifact_emission_failure",
			"task_spec_artifact_write_failed",
			"write the summary runtime task spec artifact",
			err,
		)
	}
	if err := runtimeschema.ValidateStruct(repoJoin("contracts", "ir", "workbook_operation_ir.schema.json"), operationIR); err != nil {
		return wrapFailure(
			"planning",
			"planning_failure",
			"operation_ir_invalid",
			"validate the summary operation IR artifact",
			err,
		)
	}
	if err := writeJSON(paths.OperationIRPath, operationIR); err != nil {
		return wrapFailure(
			"planning",
			"artifact_emission_failure",
			"operation_ir_artifact_write_failed",
			"write the summary operation IR artifact",
			err,
		)
	}
	return nil
}

func writeHighlightRuntimeInputs(taskSpec runtimetaskspec.TaskSpec, operationIR runtimecompiler.WorkbookOperationIR, paths RunPaths) error {
	if !shouldPersistSensitiveWorkbookArtifacts() {
		return nil
	}
	if err := runtimeschema.ValidateStruct(repoJoin("contracts", "task", "task_spec.schema.json"), taskSpec); err != nil {
		return wrapFailure(
			"planning",
			"planning_failure",
			"task_spec_invalid",
			"validate the highlight runtime task spec artifact",
			err,
		)
	}
	if err := writeJSON(paths.TaskSpecPath, taskSpec); err != nil {
		return wrapFailure(
			"planning",
			"artifact_emission_failure",
			"task_spec_artifact_write_failed",
			"write the highlight runtime task spec artifact",
			err,
		)
	}
	if err := runtimeschema.ValidateStruct(repoJoin("contracts", "ir", "workbook_operation_ir.schema.json"), operationIR); err != nil {
		return wrapFailure(
			"planning",
			"planning_failure",
			"operation_ir_invalid",
			"validate the highlight operation IR artifact",
			err,
		)
	}
	if err := writeJSON(paths.OperationIRPath, operationIR); err != nil {
		return wrapFailure(
			"planning",
			"artifact_emission_failure",
			"operation_ir_artifact_write_failed",
			"write the highlight operation IR artifact",
			err,
		)
	}
	return nil
}

func writeJoinLookupRuntimeInputs(taskSpec runtimetaskspec.TaskSpec, operationIR runtimecompiler.WorkbookOperationIR, paths RunPaths) error {
	if !shouldPersistSensitiveWorkbookArtifacts() {
		return nil
	}
	if err := runtimeschema.ValidateStruct(repoJoin("contracts", "task", "task_spec.schema.json"), taskSpec); err != nil {
		return wrapFailure(
			"planning",
			"planning_failure",
			"task_spec_invalid",
			"validate the join lookup runtime task spec artifact",
			err,
		)
	}
	if err := writeJSON(paths.TaskSpecPath, taskSpec); err != nil {
		return wrapFailure(
			"planning",
			"artifact_emission_failure",
			"task_spec_artifact_write_failed",
			"write the join lookup runtime task spec artifact",
			err,
		)
	}
	if err := runtimeschema.ValidateStruct(repoJoin("contracts", "ir", "workbook_operation_ir.schema.json"), operationIR); err != nil {
		return wrapFailure(
			"planning",
			"planning_failure",
			"operation_ir_invalid",
			"validate the join lookup operation IR artifact",
			err,
		)
	}
	if err := writeJSON(paths.OperationIRPath, operationIR); err != nil {
		return wrapFailure(
			"planning",
			"artifact_emission_failure",
			"operation_ir_artifact_write_failed",
			"write the join lookup operation IR artifact",
			err,
		)
	}
	return nil
}

func summaryTaskFromSpec(spec runtimetaskspec.TaskSpec) runtimetaskspec.GroupSummarizeTask {
	return runtimetaskspec.GroupSummarizeTask{
		TaskSpec:         spec,
		SourceSheet:      spec.SourceSheet,
		TargetSheet:      spec.TargetSheet,
		Filters:          cloneFilters(spec.Filters),
		GroupBy:          append([]string(nil), spec.GroupBy...),
		Metrics:          cloneMetrics(spec.Metrics),
		PreserveOriginal: true,
	}
}

func highlightTaskFromSpec(spec runtimetaskspec.TaskSpec) runtimetaskspec.HighlightThresholdTask {
	thresholdRule := runtimetaskspec.ThresholdRule{}
	if spec.ThresholdRule != nil {
		thresholdRule = *spec.ThresholdRule
	}
	return runtimetaskspec.HighlightThresholdTask{
		TaskSpec:         spec,
		SourceSheet:      spec.SourceSheet,
		ThresholdRule:    thresholdRule,
		PreserveOriginal: true,
	}
}

func joinLookupTaskFromSpec(spec runtimetaskspec.TaskSpec) runtimetaskspec.JoinLookupTask {
	return runtimetaskspec.JoinLookupTask{
		TaskSpec:             spec,
		SourceSheet:          spec.SourceSheet,
		LookupSheet:          spec.LookupSheet,
		JoinKey:              spec.JoinKey,
		TargetSheet:          spec.TargetSheet,
		IncludeSourceColumns: append([]string(nil), spec.IncludeSourceColumns...),
		AppendLookupColumns:  append([]string(nil), spec.AppendLookupColumns...),
		PreserveOriginal:     true,
	}
}

func summaryInspectionArtifact(inspection runtimecompiler.WorkbookInspection, targetSheet string) InspectionArtifact {
	return InspectionArtifact{
		InputFile:   inspection.InputWorkbook,
		SheetNames:  append([]string(nil), inspection.SheetNames...),
		TargetSheet: targetSheet,
		HeaderNames: append([]string(nil), inspection.HeaderNames...),
		HeaderIndex: cloneHeaderIndex(inspection.HeaderIndex),
		HeaderRow:   1,
		RowCount:    inspection.RowCount,
	}
}

func summaryInspectionRecord(inspection runtimecompiler.WorkbookInspection, targetSheet string) SummaryInspectionRecord {
	return SummaryInspectionRecord{
		InputFile:          inspection.InputWorkbook,
		SourceSheet:        inspection.SourceSheet,
		PlannedTargetSheet: targetSheet,
		SheetNames:         append([]string(nil), inspection.SheetNames...),
		HeaderNames:        append([]string(nil), inspection.HeaderNames...),
		HeaderIndex:        cloneHeaderIndex(inspection.HeaderIndex),
		RowCount:           inspection.RowCount,
	}
}

func highlightInspectionArtifact(inspection runtimecompiler.WorkbookInspection) InspectionArtifact {
	return InspectionArtifact{
		InputFile:    inspection.InputWorkbook,
		SheetNames:   append([]string(nil), inspection.SheetNames...),
		TargetSheet:  inspection.SourceSheet,
		HeaderNames:  append([]string(nil), inspection.HeaderNames...),
		HeaderIndex:  cloneHeaderIndex(inspection.HeaderIndex),
		TargetColumn: inspection.TargetColumn,
		HeaderRow:    inspection.HeaderRow,
		AmountColumn: inspection.TargetColumnIndex,
		RowCount:     inspection.RowCount,
	}
}

func evaluateWritePolicyOutput(inputFile, outputFile string, preserveOriginal bool) WritePolicyDecision {
	reasons := []string{}
	allow := true
	if outputFile == "" {
		allow = false
		reasons = append(reasons, "output file must not be empty")
	}
	if inputFile == outputFile {
		allow = false
		reasons = append(reasons, "output file must differ from input file")
	}
	if !preserveOriginal {
		allow = false
		reasons = append(reasons, "operation must preserve original workbook")
	}
	if allow {
		reasons = append(reasons, "new output file will be written and source workbook remains untouched")
	}
	return WritePolicyDecision{
		Allow:            allow,
		PreserveOriginal: preserveOriginal,
		ApprovalRequired: false,
		Reasons:          reasons,
	}
}

func executionSummaryFromRuntime(execution runtimeexecute.ExecutionResult) ExecutionSummary {
	return ExecutionSummary{
		Operation:          operationNameForFamily(execution.OperationFamily),
		OutputFile:         execution.OutputWorkbook,
		SummarySheet:       execution.SummarySheet,
		SummaryMode:        execution.SummaryMode,
		SummaryRows:        execution.SummaryRows,
		FormulaCells:       append([]string(nil), execution.FormulaCells...),
		HighlightedRows:    append([]int(nil), execution.HighlightedRows...),
		HighlightedCells:   append([]string(nil), execution.HighlightedCells...),
		WrittenCells:       append([]string(nil), execution.WrittenCells...),
		SourceSHA256Before: execution.SourceSHA256Before,
		SourceSHA256After:  execution.SourceSHA256After,
	}
}

func summaryVerificationFromRuntime(verification runtimeverify.VerificationResult) VerificationResult {
	return verificationSummaryFromRuntime(SummaryOperationName, verification)
}

func verificationSummaryFromRuntime(operation string, verification runtimeverify.VerificationResult) VerificationResult {
	return VerificationResult{
		Pass:            verification.Pass,
		Operation:       operation,
		OutputFile:      verification.OutputWorkbook,
		SummarySheet:    verification.SummarySheet,
		SummaryMode:     verification.SummaryMode,
		SummaryRows:     verification.SummaryRows,
		FormulaCells:    append([]string(nil), verification.FormulaCells...),
		HighlightedRows: append([]int(nil), verification.HighlightedRows...),
		WrittenCells:    append([]string(nil), verification.WrittenCells...),
		Layers:          verificationLayersFromRuntime(verification.Layers),
		Reasons:         append([]string(nil), verification.Reasons...),
	}
}

func verificationLayersFromRuntime(layers []runtimeverify.LayerResult) []VerificationLayer {
	if len(layers) == 0 {
		return nil
	}
	out := make([]VerificationLayer, len(layers))
	for index, layer := range layers {
		out[index] = VerificationLayer{
			Level:   layer.Level,
			Name:    layer.Name,
			Pass:    layer.Pass,
			Reasons: append([]string(nil), layer.Reasons...),
		}
	}
	return out
}

func operationNameForFamily(family string) string {
	switch family {
	case "group_summarize":
		return SummaryOperationName
	case "highlight_threshold":
		return HighlightOperationName
	case "join_lookup":
		return JoinLookupOperationName
	case "write_values":
		return WriteValuesOperationName
	default:
		return family
	}
}

func derefThreshold(value *float64) float64 {
	if value == nil {
		return 0
	}
	return *value
}

func cloneHeaderIndex(values map[string]int) map[string]int {
	if len(values) == 0 {
		return map[string]int{}
	}
	cloned := make(map[string]int, len(values))
	for key, value := range values {
		cloned[key] = value
	}
	return cloned
}

func cloneFilters(values []runtimetaskspec.FilterSpec) []runtimetaskspec.FilterSpec {
	if len(values) == 0 {
		return nil
	}
	cloned := make([]runtimetaskspec.FilterSpec, len(values))
	copy(cloned, values)
	return cloned
}

func cloneMetrics(values []runtimetaskspec.MetricSpec) []runtimetaskspec.MetricSpec {
	if len(values) == 0 {
		return nil
	}
	cloned := make([]runtimetaskspec.MetricSpec, len(values))
	copy(cloned, values)
	return cloned
}
