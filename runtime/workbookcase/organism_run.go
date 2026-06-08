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
	case "attendance_register":
		return verifyAttendanceWorkbook(outputFile)
	case "timesheet_hours_log":
		return verifyTimesheetWorkbook(outputFile)
	case "project_timeline_tracker":
		return verifyProjectTimelineWorkbook(outputFile)
	case "shift_roster_planner":
		return verifyShiftRosterWorkbook(outputFile)
	case "construction_cost_tracker":
		return verifyConstructionCostWorkbook(outputFile)
	case "inventory_movement_log":
		return verifyInventoryMovementWorkbook(outputFile)
	case "procurement_reconciliation":
		return verifyProcurementReconciliationWorkbook(outputFile)
	case "warehouse_reorder_tracker":
		return verifyWarehouseReorderWorkbook(outputFile)
	case "student_gradebook":
		return verifyStudentGradebookWorkbook(outputFile)
	case "training_completion_matrix":
		return verifyTrainingCompletionWorkbook(outputFile)
	case "service_ticket_queue":
		return verifyServiceTicketWorkbook(outputFile)
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

func verifyAttendanceWorkbook(outputFile string) []string {
	handle, err := excelize.OpenFile(outputFile)
	if err != nil {
		return []string{fmt.Sprintf("attendance workbook semantic check failed to open output: %v", err)}
	}
	defer func() { _ = handle.Close() }()

	var reasons []string
	if got, err := handle.GetCellValue("NextAttendance", "A2"); err != nil {
		reasons = append(reasons, fmt.Sprintf("attendance period-copy semantic check missing NextAttendance!A2: %v", err))
	} else if strings.TrimSpace(got) == "" {
		reasons = append(reasons, "attendance period-copy semantic check missing NextAttendance!A2 student value")
	}
	if got, err := handle.GetCellFormula("NextAttendance", "D2"); err != nil {
		reasons = append(reasons, fmt.Sprintf("attendance formula semantic check missing NextAttendance!D2 formula: %v", err))
	} else if strings.TrimSpace(got) == "" {
		reasons = append(reasons, "attendance formula semantic check missing NextAttendance!D2 formula")
	}
	if validations, err := handle.GetDataValidations("NextAttendance"); err != nil {
		reasons = append(reasons, fmt.Sprintf("attendance validation semantic check missing NextAttendance data validation: %v", err))
	} else if len(validations) == 0 {
		reasons = append(reasons, "attendance validation semantic check missing NextAttendance data validation")
	}
	if protection, err := handle.GetSheetProtection("NextAttendance"); err != nil {
		reasons = append(reasons, fmt.Sprintf("attendance protection semantic check missing NextAttendance protection: %v", err))
	} else if !protection.SelectLockedCells || !protection.SelectUnlockedCells {
		reasons = append(reasons, "attendance protection semantic check missing NextAttendance protection options")
	}
	return reasons
}

func verifyTimesheetWorkbook(outputFile string) []string {
	handle, err := excelize.OpenFile(outputFile)
	if err != nil {
		return []string{fmt.Sprintf("timesheet workbook semantic check failed to open output: %v", err)}
	}
	defer func() { _ = handle.Close() }()

	var reasons []string
	if got, err := handle.GetCellValue("Week2", "B3"); err != nil {
		reasons = append(reasons, fmt.Sprintf("timesheet append semantic check missing Week2!B3: %v", err))
	} else if strings.TrimSpace(got) == "" {
		reasons = append(reasons, "timesheet append semantic check missing Week2!B3 employee value")
	}
	if got, err := handle.GetCellFormula("Week2", "F3"); err != nil {
		reasons = append(reasons, fmt.Sprintf("timesheet pay formula semantic check missing Week2!F3 formula: %v", err))
	} else if strings.TrimSpace(got) == "" {
		reasons = append(reasons, "timesheet pay formula semantic check missing Week2!F3 formula")
	}
	if validations, err := handle.GetDataValidations("Week2"); err != nil {
		reasons = append(reasons, fmt.Sprintf("timesheet validation semantic check missing Week2 data validation: %v", err))
	} else if len(validations) == 0 {
		reasons = append(reasons, "timesheet validation semantic check missing Week2 data validation")
	}
	if protection, err := handle.GetSheetProtection("Week2"); err != nil {
		reasons = append(reasons, fmt.Sprintf("timesheet protection semantic check missing Week2 protection: %v", err))
	} else if !protection.SelectLockedCells || !protection.SelectUnlockedCells {
		reasons = append(reasons, "timesheet protection semantic check missing Week2 protection options")
	}
	return reasons
}

func verifyProjectTimelineWorkbook(outputFile string) []string {
	handle, err := excelize.OpenFile(outputFile)
	if err != nil {
		return []string{fmt.Sprintf("project timeline workbook semantic check failed to open output: %v", err)}
	}
	defer func() { _ = handle.Close() }()

	var reasons []string
	if got, err := handle.GetCellValue("Sprint2", "A4"); err != nil {
		reasons = append(reasons, fmt.Sprintf("project timeline period-copy semantic check missing Sprint2!A4: %v", err))
	} else if strings.TrimSpace(got) == "" {
		reasons = append(reasons, "project timeline period-copy semantic check missing Sprint2!A4 task value")
	}
	if got, err := handle.GetCellFormula("Sprint2", "F4"); err != nil {
		reasons = append(reasons, fmt.Sprintf("project timeline formula semantic check missing Sprint2!F4 formula: %v", err))
	} else if strings.TrimSpace(got) == "" {
		reasons = append(reasons, "project timeline formula semantic check missing Sprint2!F4 formula")
	}
	if got, err := handle.GetCellValue("TimelineSummary", "A2"); err != nil {
		reasons = append(reasons, fmt.Sprintf("project timeline summary semantic check missing TimelineSummary!A2: %v", err))
	} else if strings.TrimSpace(got) == "" {
		reasons = append(reasons, "project timeline summary semantic check missing TimelineSummary!A2 status value")
	}
	if validations, err := handle.GetDataValidations("Sprint2"); err != nil {
		reasons = append(reasons, fmt.Sprintf("project timeline validation semantic check missing Sprint2 data validation: %v", err))
	} else if len(validations) == 0 {
		reasons = append(reasons, "project timeline validation semantic check missing Sprint2 data validation")
	}
	if protection, err := handle.GetSheetProtection("Sprint2"); err != nil {
		reasons = append(reasons, fmt.Sprintf("project timeline protection semantic check missing Sprint2 protection: %v", err))
	} else if !protection.SelectLockedCells || !protection.SelectUnlockedCells {
		reasons = append(reasons, "project timeline protection semantic check missing Sprint2 protection options")
	}
	return reasons
}

func verifyShiftRosterWorkbook(outputFile string) []string {
	handle, err := excelize.OpenFile(outputFile)
	if err != nil {
		return []string{fmt.Sprintf("shift roster workbook semantic check failed to open output: %v", err)}
	}
	defer func() { _ = handle.Close() }()

	var reasons []string
	if got, err := handle.GetCellValue("Week2", "A2"); err != nil {
		reasons = append(reasons, fmt.Sprintf("shift roster period-copy semantic check missing Week2!A2: %v", err))
	} else if strings.TrimSpace(got) == "" {
		reasons = append(reasons, "shift roster period-copy semantic check missing Week2!A2 employee value")
	}
	if got, err := handle.GetCellFormula("Week2", "D2"); err != nil {
		reasons = append(reasons, fmt.Sprintf("shift roster coverage formula semantic check missing Week2!D2 formula: %v", err))
	} else if strings.TrimSpace(got) == "" {
		reasons = append(reasons, "shift roster coverage formula semantic check missing Week2!D2 formula")
	}
	if validations, err := handle.GetDataValidations("Week2"); err != nil {
		reasons = append(reasons, fmt.Sprintf("shift roster validation semantic check missing Week2 data validation: %v", err))
	} else if len(validations) == 0 {
		reasons = append(reasons, "shift roster validation semantic check missing Week2 data validation")
	}
	if protection, err := handle.GetSheetProtection("Week2"); err != nil {
		reasons = append(reasons, fmt.Sprintf("shift roster protection semantic check missing Week2 protection: %v", err))
	} else if !protection.SelectLockedCells || !protection.SelectUnlockedCells {
		reasons = append(reasons, "shift roster protection semantic check missing Week2 protection options")
	}
	return reasons
}

func verifyConstructionCostWorkbook(outputFile string) []string {
	handle, err := excelize.OpenFile(outputFile)
	if err != nil {
		return []string{fmt.Sprintf("construction cost workbook semantic check failed to open output: %v", err)}
	}
	defer func() { _ = handle.Close() }()

	var reasons []string
	if got, err := handle.GetCellValue("Costs", "A3"); err != nil {
		reasons = append(reasons, fmt.Sprintf("construction cost append semantic check missing Costs!A3: %v", err))
	} else if strings.TrimSpace(got) == "" {
		reasons = append(reasons, "construction cost append semantic check missing Costs!A3 cost code value")
	}
	if got, err := handle.GetCellFormula("Costs", "E3"); err != nil {
		reasons = append(reasons, fmt.Sprintf("construction cost variance formula semantic check missing Costs!E3 formula: %v", err))
	} else if strings.TrimSpace(got) == "" {
		reasons = append(reasons, "construction cost variance formula semantic check missing Costs!E3 formula")
	}
	if protection, err := handle.GetSheetProtection("Costs"); err != nil {
		reasons = append(reasons, fmt.Sprintf("construction cost protection semantic check missing Costs protection: %v", err))
	} else if !protection.SelectLockedCells || !protection.SelectUnlockedCells {
		reasons = append(reasons, "construction cost protection semantic check missing Costs protection options")
	}
	return reasons
}

func verifyInventoryMovementWorkbook(outputFile string) []string {
	handle, err := excelize.OpenFile(outputFile)
	if err != nil {
		return []string{fmt.Sprintf("inventory movement workbook semantic check failed to open output: %v", err)}
	}
	defer func() { _ = handle.Close() }()

	var reasons []string
	if got, err := handle.GetCellValue("Movements", "A1"); err != nil {
		reasons = append(reasons, fmt.Sprintf("inventory header semantic check missing Movements!A1: %v", err))
	} else if strings.TrimSpace(got) != "sku" {
		reasons = append(reasons, fmt.Sprintf("inventory header semantic check Movements!A1=%q want sku", got))
	}
	if got, err := handle.GetCellValue("MovementsEnriched", "E4"); err != nil {
		reasons = append(reasons, fmt.Sprintf("inventory lookup semantic check missing MovementsEnriched!E4: %v", err))
	} else if strings.TrimSpace(got) == "" {
		reasons = append(reasons, "inventory lookup semantic check missing MovementsEnriched!E4 location value")
	}
	if got, err := handle.GetCellValue("InventoryReconciliation", "B4"); err != nil {
		reasons = append(reasons, fmt.Sprintf("inventory reconciliation semantic check missing InventoryReconciliation!B4: %v", err))
	} else if strings.TrimSpace(got) == "" {
		reasons = append(reasons, "inventory reconciliation semantic check missing InventoryReconciliation!B4 sku value")
	}
	if protection, err := handle.GetSheetProtection("Movements"); err != nil {
		reasons = append(reasons, fmt.Sprintf("inventory protection semantic check missing Movements protection: %v", err))
	} else if !protection.SelectLockedCells || !protection.SelectUnlockedCells {
		reasons = append(reasons, "inventory protection semantic check missing Movements protection options")
	}
	return reasons
}

func verifyProcurementReconciliationWorkbook(outputFile string) []string {
	handle, err := excelize.OpenFile(outputFile)
	if err != nil {
		return []string{fmt.Sprintf("procurement reconciliation workbook semantic check failed to open output: %v", err)}
	}
	defer func() { _ = handle.Close() }()

	var reasons []string
	if got, err := handle.GetCellValue("ProcurementReconciliation", "A2"); err != nil {
		reasons = append(reasons, fmt.Sprintf("procurement reconciliation semantic check missing ProcurementReconciliation!A2: %v", err))
	} else if strings.TrimSpace(got) == "" {
		reasons = append(reasons, "procurement reconciliation semantic check missing ProcurementReconciliation!A2 result value")
	}
	if got, err := handle.GetCellValue("ProcurementReconciliation", "B2"); err != nil {
		reasons = append(reasons, fmt.Sprintf("procurement reconciliation semantic check missing ProcurementReconciliation!B2: %v", err))
	} else if strings.TrimSpace(got) == "" {
		reasons = append(reasons, "procurement reconciliation semantic check missing ProcurementReconciliation!B2 PO id value")
	}
	if got, err := handle.GetCellFormula("Invoice", "D2"); err != nil {
		reasons = append(reasons, fmt.Sprintf("procurement review formula semantic check missing Invoice!D2 formula: %v", err))
	} else if strings.TrimSpace(got) == "" {
		reasons = append(reasons, "procurement review formula semantic check missing Invoice!D2 formula")
	}
	if validations, err := handle.GetDataValidations("Invoice"); err != nil {
		reasons = append(reasons, fmt.Sprintf("procurement validation semantic check missing Invoice data validation: %v", err))
	} else if len(validations) == 0 {
		reasons = append(reasons, "procurement validation semantic check missing Invoice data validation")
	}
	if protection, err := handle.GetSheetProtection("Invoice"); err != nil {
		reasons = append(reasons, fmt.Sprintf("procurement protection semantic check missing Invoice protection: %v", err))
	} else if !protection.SelectLockedCells || !protection.SelectUnlockedCells {
		reasons = append(reasons, "procurement protection semantic check missing Invoice protection options")
	}
	return reasons
}

func verifyWarehouseReorderWorkbook(outputFile string) []string {
	handle, err := excelize.OpenFile(outputFile)
	if err != nil {
		return []string{fmt.Sprintf("warehouse reorder workbook semantic check failed to open output: %v", err)}
	}
	defer func() { _ = handle.Close() }()

	var reasons []string
	if got, err := handle.GetCellValue("StockEnriched", "E2"); err != nil {
		reasons = append(reasons, fmt.Sprintf("warehouse reorder lookup semantic check missing StockEnriched!E2: %v", err))
	} else if strings.TrimSpace(got) == "" {
		reasons = append(reasons, "warehouse reorder lookup semantic check missing StockEnriched!E2 location value")
	}
	if got, err := handle.GetCellValue("StockEnriched", "A3"); err != nil {
		reasons = append(reasons, fmt.Sprintf("warehouse reorder append semantic check missing StockEnriched!A3: %v", err))
	} else if strings.TrimSpace(got) == "" {
		reasons = append(reasons, "warehouse reorder append semantic check missing StockEnriched!A3 sku value")
	}
	if got, err := handle.GetCellFormula("Stock", "D2"); err != nil {
		reasons = append(reasons, fmt.Sprintf("warehouse reorder formula semantic check missing Stock!D2 formula: %v", err))
	} else if strings.TrimSpace(got) == "" {
		reasons = append(reasons, "warehouse reorder formula semantic check missing Stock!D2 formula")
	}
	if validations, err := handle.GetDataValidations("StockEnriched"); err != nil {
		reasons = append(reasons, fmt.Sprintf("warehouse reorder validation semantic check missing StockEnriched data validation: %v", err))
	} else if len(validations) == 0 {
		reasons = append(reasons, "warehouse reorder validation semantic check missing StockEnriched data validation")
	}
	if protection, err := handle.GetSheetProtection("Stock"); err != nil {
		reasons = append(reasons, fmt.Sprintf("warehouse reorder protection semantic check missing Stock protection: %v", err))
	} else if !protection.SelectLockedCells || !protection.SelectUnlockedCells {
		reasons = append(reasons, "warehouse reorder protection semantic check missing Stock protection options")
	}
	return reasons
}

func verifyStudentGradebookWorkbook(outputFile string) []string {
	handle, err := excelize.OpenFile(outputFile)
	if err != nil {
		return []string{fmt.Sprintf("student gradebook workbook semantic check failed to open output: %v", err)}
	}
	defer func() { _ = handle.Close() }()

	var reasons []string
	if got, err := handle.GetCellValue("GradeSummary", "B2"); err != nil {
		reasons = append(reasons, fmt.Sprintf("student gradebook summary semantic check missing GradeSummary!B2: %v", err))
	} else if strings.TrimSpace(got) == "" {
		reasons = append(reasons, "student gradebook summary semantic check missing GradeSummary!B2 score summary value")
	}
	if got, err := handle.GetCellFormula("Grades", "E4"); err != nil {
		reasons = append(reasons, fmt.Sprintf("student gradebook formula semantic check missing Grades!E4 formula: %v", err))
	} else if strings.TrimSpace(got) == "" {
		reasons = append(reasons, "student gradebook formula semantic check missing Grades!E4 formula")
	}
	if validations, err := handle.GetDataValidations("Grades"); err != nil {
		reasons = append(reasons, fmt.Sprintf("student gradebook validation semantic check missing Grades data validation: %v", err))
	} else if len(validations) == 0 {
		reasons = append(reasons, "student gradebook validation semantic check missing Grades data validation")
	}
	if protection, err := handle.GetSheetProtection("Grades"); err != nil {
		reasons = append(reasons, fmt.Sprintf("student gradebook protection semantic check missing Grades protection: %v", err))
	} else if !protection.SelectLockedCells || !protection.SelectUnlockedCells {
		reasons = append(reasons, "student gradebook protection semantic check missing Grades protection options")
	}
	return reasons
}

func verifyTrainingCompletionWorkbook(outputFile string) []string {
	handle, err := excelize.OpenFile(outputFile)
	if err != nil {
		return []string{fmt.Sprintf("training completion workbook semantic check failed to open output: %v", err)}
	}
	defer func() { _ = handle.Close() }()

	var reasons []string
	if got, err := handle.GetCellValue("TrainingSummary", "B2"); err != nil {
		reasons = append(reasons, fmt.Sprintf("training completion summary semantic check missing TrainingSummary!B2: %v", err))
	} else if strings.TrimSpace(got) == "" {
		reasons = append(reasons, "training completion summary semantic check missing TrainingSummary!B2 completion summary value")
	}
	if got, err := handle.GetCellValue("TrainingReport", "A1"); err != nil {
		reasons = append(reasons, fmt.Sprintf("training completion printable semantic check missing TrainingReport!A1: %v", err))
	} else if strings.TrimSpace(got) == "" {
		reasons = append(reasons, "training completion printable semantic check missing TrainingReport!A1 title")
	}
	if got, err := handle.GetCellFormula("Training", "E2"); err != nil {
		reasons = append(reasons, fmt.Sprintf("training completion formula semantic check missing Training!E2 formula: %v", err))
	} else if strings.TrimSpace(got) == "" {
		reasons = append(reasons, "training completion formula semantic check missing Training!E2 formula")
	}
	if validations, err := handle.GetDataValidations("Training"); err != nil {
		reasons = append(reasons, fmt.Sprintf("training completion validation semantic check missing Training data validation: %v", err))
	} else if len(validations) == 0 {
		reasons = append(reasons, "training completion validation semantic check missing Training data validation")
	}
	if protection, err := handle.GetSheetProtection("Training"); err != nil {
		reasons = append(reasons, fmt.Sprintf("training completion protection semantic check missing Training protection: %v", err))
	} else if !protection.SelectLockedCells || !protection.SelectUnlockedCells {
		reasons = append(reasons, "training completion protection semantic check missing Training protection options")
	}
	return reasons
}

func verifyServiceTicketWorkbook(outputFile string) []string {
	handle, err := excelize.OpenFile(outputFile)
	if err != nil {
		return []string{fmt.Sprintf("service ticket workbook semantic check failed to open output: %v", err)}
	}
	defer func() { _ = handle.Close() }()

	var reasons []string
	if got, err := handle.GetCellValue("Tickets", "A3"); err != nil {
		reasons = append(reasons, fmt.Sprintf("service ticket append semantic check missing Tickets!A3: %v", err))
	} else if strings.TrimSpace(got) == "" {
		reasons = append(reasons, "service ticket append semantic check missing Tickets!A3 ticket id value")
	}
	if got, err := handle.GetCellValue("TicketSummary", "B2"); err != nil {
		reasons = append(reasons, fmt.Sprintf("service ticket summary semantic check missing TicketSummary!B2: %v", err))
	} else if strings.TrimSpace(got) == "" {
		reasons = append(reasons, "service ticket summary semantic check missing TicketSummary!B2 ticket summary value")
	}
	if got, err := handle.GetCellFormula("Tickets", "D3"); err != nil {
		reasons = append(reasons, fmt.Sprintf("service ticket SLA formula semantic check missing Tickets!D3 formula: %v", err))
	} else if strings.TrimSpace(got) == "" {
		reasons = append(reasons, "service ticket SLA formula semantic check missing Tickets!D3 formula")
	}
	if validations, err := handle.GetDataValidations("Tickets"); err != nil {
		reasons = append(reasons, fmt.Sprintf("service ticket validation semantic check missing Tickets data validation: %v", err))
	} else if len(validations) == 0 {
		reasons = append(reasons, "service ticket validation semantic check missing Tickets data validation")
	}
	if protection, err := handle.GetSheetProtection("Tickets"); err != nil {
		reasons = append(reasons, fmt.Sprintf("service ticket protection semantic check missing Tickets protection: %v", err))
	} else if !protection.SelectLockedCells || !protection.SelectUnlockedCells {
		reasons = append(reasons, "service ticket protection semantic check missing Tickets protection options")
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
