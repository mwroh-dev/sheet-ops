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
	case "expense_reimbursement":
		return verifyExpenseReimbursementWorkbook(outputFile)
	case "purchase_order_control":
		return verifyPurchaseOrderWorkbook(outputFile)
	case "monthly_budget_control":
		return verifyMonthlyBudgetWorkbook(outputFile)
	case "cash_flow_monitor":
		return verifyCashFlowWorkbook(outputFile)
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

func verifyExpenseReimbursementWorkbook(outputFile string) []string {
	handle, err := excelize.OpenFile(outputFile)
	if err != nil {
		return []string{fmt.Sprintf("expense workbook semantic check failed to open output: %v", err)}
	}
	defer func() { _ = handle.Close() }()

	var reasons []string
	if got, err := handle.GetCellValue("ExpenseItems", "A3"); err != nil {
		reasons = append(reasons, fmt.Sprintf("expense reimbursement semantic check failed reading ExpenseItems!A3: %v", err))
	} else if strings.TrimSpace(got) == "" {
		reasons = append(reasons, "expense reimbursement semantic check missing appended ExpenseItems!A3 value")
	}
	if got, err := handle.GetCellFormula("ExpenseItems", "E3"); err != nil {
		reasons = append(reasons, fmt.Sprintf("expense reimbursement semantic check failed reading ExpenseItems!E3 formula: %v", err))
	} else if strings.TrimSpace(got) == "" {
		reasons = append(reasons, "expense reimbursement semantic check missing ExpenseItems!E3 formula")
	}
	if got, err := handle.GetCellValue("ExpenseClaim", "A1"); err != nil {
		reasons = append(reasons, fmt.Sprintf("expense printable semantic check missing ExpenseClaim!A1: %v", err))
	} else if strings.TrimSpace(got) == "" {
		reasons = append(reasons, "expense printable semantic check missing ExpenseClaim!A1 title")
	}
	return reasons
}

func verifyPurchaseOrderWorkbook(outputFile string) []string {
	handle, err := excelize.OpenFile(outputFile)
	if err != nil {
		return []string{fmt.Sprintf("purchase order workbook semantic check failed to open output: %v", err)}
	}
	defer func() { _ = handle.Close() }()

	var reasons []string
	if got, err := handle.GetCellValue("PurchaseOrder", "A3"); err != nil {
		reasons = append(reasons, fmt.Sprintf("purchase order semantic check failed reading PurchaseOrder!A3: %v", err))
	} else if strings.TrimSpace(got) == "" {
		reasons = append(reasons, "purchase order semantic check missing appended PurchaseOrder!A3 value")
	}
	if got, err := handle.GetCellFormula("PurchaseOrder", "E3"); err != nil {
		reasons = append(reasons, fmt.Sprintf("purchase order semantic check failed reading PurchaseOrder!E3 formula: %v", err))
	} else if strings.TrimSpace(got) == "" {
		reasons = append(reasons, "purchase order semantic check missing PurchaseOrder!E3 formula")
	}
	if got, err := handle.GetCellValue("PurchaseOrderPrint", "A1"); err != nil {
		reasons = append(reasons, fmt.Sprintf("purchase order printable semantic check missing PurchaseOrderPrint!A1: %v", err))
	} else if strings.TrimSpace(got) == "" {
		reasons = append(reasons, "purchase order printable semantic check missing PurchaseOrderPrint!A1 title")
	}
	return reasons
}

func verifyMonthlyBudgetWorkbook(outputFile string) []string {
	handle, err := excelize.OpenFile(outputFile)
	if err != nil {
		return []string{fmt.Sprintf("monthly budget workbook semantic check failed to open output: %v", err)}
	}
	defer func() { _ = handle.Close() }()

	var reasons []string
	if got, err := handle.GetCellValue("BudgetSummary", "B2"); err != nil {
		reasons = append(reasons, fmt.Sprintf("monthly budget semantic check missing BudgetSummary!B2: %v", err))
	} else if strings.TrimSpace(got) == "" {
		reasons = append(reasons, "monthly budget semantic check missing BudgetSummary!B2 summary value")
	}
	if got, err := handle.GetCellValue("NextBudget", "B3"); err != nil {
		reasons = append(reasons, fmt.Sprintf("monthly budget roll-forward semantic check missing NextBudget!B3: %v", err))
	} else if strings.TrimSpace(got) == "" {
		reasons = append(reasons, "monthly budget roll-forward semantic check missing NextBudget!B3 value")
	}
	if got, err := handle.GetCellFormula("NextBudget", "E2"); err != nil {
		reasons = append(reasons, fmt.Sprintf("monthly budget formula semantic check missing NextBudget!E2 formula: %v", err))
	} else if strings.TrimSpace(got) == "" {
		reasons = append(reasons, "monthly budget formula semantic check missing NextBudget!E2 formula")
	}
	return reasons
}

func verifyCashFlowWorkbook(outputFile string) []string {
	handle, err := excelize.OpenFile(outputFile)
	if err != nil {
		return []string{fmt.Sprintf("cash flow workbook semantic check failed to open output: %v", err)}
	}
	defer func() { _ = handle.Close() }()

	var reasons []string
	if got, err := handle.GetCellValue("CashFlowSummary", "B2"); err != nil {
		reasons = append(reasons, fmt.Sprintf("cash flow semantic check missing CashFlowSummary!B2: %v", err))
	} else if strings.TrimSpace(got) == "" {
		reasons = append(reasons, "cash flow semantic check missing CashFlowSummary!B2 inflow summary value")
	}
	if got, err := handle.GetCellValue("NextCashFlow", "B3"); err != nil {
		reasons = append(reasons, fmt.Sprintf("cash flow roll-forward semantic check missing NextCashFlow!B3: %v", err))
	} else if strings.TrimSpace(got) == "" {
		reasons = append(reasons, "cash flow roll-forward semantic check missing NextCashFlow!B3 opening value")
	}
	if got, err := handle.GetCellFormula("NextCashFlow", "E4"); err != nil {
		reasons = append(reasons, fmt.Sprintf("cash flow formula semantic check missing NextCashFlow!E4 formula: %v", err))
	} else if strings.TrimSpace(got) == "" {
		reasons = append(reasons, "cash flow formula semantic check missing NextCashFlow!E4 formula")
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
