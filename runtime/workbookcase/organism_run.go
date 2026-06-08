package workbookcase

import (
	"fmt"
	"path/filepath"
	"strings"

	runtimeschema "github.com/mwroh/sheet-ops/runtime/schema"
	runtimetaskspec "github.com/mwroh/sheet-ops/runtime/taskspec"
	"github.com/mwroh/sheet-ops/runtime/templateclass"
	"github.com/xuri/excelize/v2"
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
	OrganismID               string
	StepResults              []RunResult
	ExecutedAtomIDs          []string
	OrganismVerification     OrganismVerificationResult
	OrganismVerificationPath string
	TemplateClassEvaluation  templateclass.EvaluationResult
}

type OrganismVerificationResult struct {
	Pass            bool     `json:"pass"`
	OrganismID      string   `json:"organism_id"`
	VerifierSpecID  string   `json:"verifier_spec_id"`
	ExpectedAtomIDs []string `json:"expected_atom_ids"`
	ExecutedAtomIDs []string `json:"executed_atom_ids"`
	StepCount       int      `json:"step_count"`
	PassedStepCount int      `json:"passed_step_count"`
	OutputFile      string   `json:"output_file"`
	Reasons         []string `json:"reasons"`
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
	result.OrganismVerification = verifyOrganismPlanExecution(plan, executed, result.StepResults, req.OutputFile)
	result.OrganismVerificationPath = organismVerificationPath(req.OutputFile)
	if err := runtimeschema.ValidateStruct(repoJoin("contracts", "verification", "organism_verification_result.schema.json"), result.OrganismVerification); err != nil {
		return result, err
	}
	if err := writeJSON(result.OrganismVerificationPath, result.OrganismVerification); err != nil {
		return result, err
	}
	result.TemplateClassEvaluation = templateclass.EvaluateEvidence(plan, templateclass.Evidence{
		ExecutedAtomIDs: executed,
		VerifierPasses: map[string]bool{
			result.OrganismVerification.VerifierSpecID: result.OrganismVerification.Pass,
		},
	})
	if !result.TemplateClassEvaluation.Pass {
		return result, fmt.Errorf("template class evaluation failed: %v", result.TemplateClassEvaluation.Reasons)
	}
	return result, nil
}

func verifyOrganismPlanExecution(plan templateclass.Plan, executed []string, stepResults []RunResult, outputFile string) OrganismVerificationResult {
	verifierSpecID := ""
	if len(plan.RequiredVerifierSpecs) > 0 {
		verifierSpecID = plan.RequiredVerifierSpecs[0]
	}
	result := OrganismVerificationResult{
		Pass:            true,
		OrganismID:      plan.OrganismID,
		VerifierSpecID:  verifierSpecID,
		ExpectedAtomIDs: append([]string(nil), plan.OperationSequence...),
		ExecutedAtomIDs: append([]string(nil), executed...),
		StepCount:       len(stepResults),
		OutputFile:      outputFile,
		Reasons:         []string{},
	}
	if verifierSpecID == "" {
		result.Reasons = append(result.Reasons, "missing organism verifier spec")
	}
	if len(executed) != len(plan.OperationSequence) {
		result.Reasons = append(result.Reasons, fmt.Sprintf("executed atom count %d does not match expected count %d", len(executed), len(plan.OperationSequence)))
	}
	for index, expected := range plan.OperationSequence {
		if index >= len(executed) {
			result.Reasons = append(result.Reasons, fmt.Sprintf("missing executed atom %s", expected))
			continue
		}
		if executed[index] != expected {
			result.Reasons = append(result.Reasons, fmt.Sprintf("executed atom %d %q does not match expected %q", index+1, executed[index], expected))
		}
	}
	for index, stepResult := range stepResults {
		if stepResult.Verification.Pass {
			result.PassedStepCount++
			continue
		}
		result.Reasons = append(result.Reasons, fmt.Sprintf("step %d verification failed", index+1))
	}
	if result.PassedStepCount != len(plan.OperationSequence) {
		result.Reasons = append(result.Reasons, fmt.Sprintf("passed step count %d does not match expected count %d", result.PassedStepCount, len(plan.OperationSequence)))
	}
	result.Reasons = append(result.Reasons, verifyOrganismWorkbookSemantics(plan.OrganismID, outputFile)...)
	result.Pass = len(result.Reasons) == 0
	return result
}

func verifyOrganismWorkbookSemantics(organismID, outputFile string) []string {
	switch organismID {
	case "invoice_line_item_billing":
		return verifyInvoiceLineItemWorkbook(outputFile)
	default:
		return nil
	}
}

func verifyInvoiceLineItemWorkbook(outputFile string) []string {
	handle, err := excelize.OpenFile(outputFile)
	if err != nil {
		return []string{fmt.Sprintf("invoice workbook semantic check failed to open output: %v", err)}
	}
	defer func() { _ = handle.Close() }()

	var reasons []string
	if got, err := handle.GetCellValue("LineItems", "A3"); err != nil {
		reasons = append(reasons, fmt.Sprintf("invoice line item semantic check failed reading LineItems!A3: %v", err))
	} else if strings.TrimSpace(got) == "" {
		reasons = append(reasons, "invoice line item semantic check missing appended LineItems!A3 value")
	}
	if got, err := handle.GetCellFormula("LineItems", "D3"); err != nil {
		reasons = append(reasons, fmt.Sprintf("invoice line item semantic check failed reading LineItems!D3 formula: %v", err))
	} else if strings.TrimSpace(got) == "" {
		reasons = append(reasons, "invoice line item semantic check missing LineItems!D3 formula")
	}
	if got, err := handle.GetCellValue("InvoicePrint", "A1"); err != nil {
		reasons = append(reasons, fmt.Sprintf("invoice printable semantic check missing InvoicePrint!A1: %v", err))
	} else if strings.TrimSpace(got) == "" {
		reasons = append(reasons, "invoice printable semantic check missing InvoicePrint!A1 title")
	}
	return reasons
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

func organismVerificationPath(outputFile string) string {
	ext := filepath.Ext(outputFile)
	stem := strings.TrimSuffix(outputFile, ext)
	return stem + ".organism-verification.json"
}
