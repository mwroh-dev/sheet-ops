package workbookcase

import (
	"fmt"
	"path/filepath"
	"strings"

	runtimetaskspec "github.com/mwroh/sheet-ops/runtime/taskspec"
	"github.com/mwroh/sheet-ops/runtime/templateclass"
)

type OrganismStep struct {
	AtomID string
	Build  func(input, output string) runtimetaskspec.TaskSpec
}

type OrganismRunRequest struct {
	ScenarioID  string
	RequestText string
	InputFile   string
	OutputFile  string
	Steps       []OrganismStep
}

type OrganismRunResult struct {
	OrganismID              string
	StepResults             []RunResult
	ExecutedAtomIDs         []string
	TemplateClassEvaluation templateclass.EvaluationResult
}

func RunOrganismPlan(req OrganismRunRequest) (OrganismRunResult, error) {
	var result OrganismRunResult
	if strings.TrimSpace(req.ScenarioID) == "" {
		return result, fmt.Errorf("scenario id must not be empty")
	}
	if strings.TrimSpace(req.InputFile) == "" || strings.TrimSpace(req.OutputFile) == "" {
		return result, fmt.Errorf("input and output files must not be empty")
	}
	if len(req.Steps) == 0 {
		return result, fmt.Errorf("organism run requires at least one step")
	}
	plan, ok := templateclass.PlanForRequest(req.RequestText)
	if !ok {
		return result, fmt.Errorf("no template class plan for request")
	}
	result.OrganismID = plan.OrganismID
	if err := validateOrganismStepsMatchPlan(plan, req.Steps); err != nil {
		return result, err
	}

	currentInput := req.InputFile
	executed := make([]string, 0, len(req.Steps))
	for index, step := range req.Steps {
		if strings.TrimSpace(step.AtomID) == "" || step.Build == nil {
			return result, fmt.Errorf("step %d is incomplete", index)
		}
		stepOutput := req.OutputFile
		if index < len(req.Steps)-1 {
			stepOutput = intermediateOutputPath(req.OutputFile, index, step.AtomID)
		}
		taskSpec := step.Build(currentInput, stepOutput)
		stepResult, err := Run(Request{
			ScenarioID: fmt.Sprintf("%s-step-%02d-%s", req.ScenarioID, index+1, step.AtomID),
			TaskSpec:   taskSpec,
		})
		if err != nil {
			return result, err
		}
		if !stepResult.Verification.Pass {
			return result, fmt.Errorf("step %d %s verification failed: %v", index+1, step.AtomID, stepResult.Verification.Reasons)
		}
		result.StepResults = append(result.StepResults, stepResult)
		executed = append(executed, step.AtomID)
		currentInput = stepOutput
	}
	result.ExecutedAtomIDs = append([]string(nil), executed...)
	result.TemplateClassEvaluation = templateclass.EvaluateEvidence(plan, templateclass.Evidence{
		ExecutedAtomIDs: executed,
		VerifierPasses: map[string]bool{
			plan.RequiredVerifierSpecs[0]: true,
		},
	})
	if !result.TemplateClassEvaluation.Pass {
		return result, fmt.Errorf("template class evaluation failed: %v", result.TemplateClassEvaluation.Reasons)
	}
	return result, nil
}

func validateOrganismStepsMatchPlan(plan templateclass.Plan, steps []OrganismStep) error {
	if len(steps) != len(plan.OperationSequence) {
		return fmt.Errorf("step count %d does not match template class plan count %d", len(steps), len(plan.OperationSequence))
	}
	for index, step := range steps {
		if step.AtomID != plan.OperationSequence[index] {
			return fmt.Errorf("step %d atom %q does not match template class plan atom %q", index+1, step.AtomID, plan.OperationSequence[index])
		}
	}
	return nil
}

func intermediateOutputPath(outputFile string, index int, atomID string) string {
	ext := filepath.Ext(outputFile)
	stem := strings.TrimSuffix(outputFile, ext)
	return fmt.Sprintf("%s.step-%02d-%s%s", stem, index+1, atomID, ext)
}
