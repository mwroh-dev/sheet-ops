package templateclass

import "testing"

func TestPlanForRequestClassifiesInvoiceAndRequiresSupportedAtomEvidence(t *testing.T) {
	plan, ok := PlanForRequest("Add invoice line items, extend totals, protect formulas, and create a printable invoice")
	if !ok {
		t.Fatal("PlanForRequest returned ok=false")
	}
	if plan.OrganismID != "invoice_line_item_billing" {
		t.Fatalf("organism=%q want invoice_line_item_billing", plan.OrganismID)
	}
	wantOps := []string{"append_structured_rows", "extend_table_formulas", "add_data_validation", "protect_formula_cells", "generate_printable_form"}
	if !sameStrings(plan.OperationSequence, wantOps) {
		t.Fatalf("operation sequence=%v want %v", plan.OperationSequence, wantOps)
	}

	result := EvaluateEvidence(plan, Evidence{
		ExecutedAtomIDs: []string{"append_structured_rows", "extend_table_formulas", "add_data_validation", "protect_formula_cells", "generate_printable_form"},
		VerifierPasses: map[string]bool{
			"invoice_line_item_billing_verifier": true,
		},
	})
	if !result.Pass {
		t.Fatalf("EvaluateEvidence pass=false reasons=%v", result.Reasons)
	}
	if result.RuntimeClaim != "template_class_plan_verified" {
		t.Fatalf("runtime claim=%q want template_class_plan_verified", result.RuntimeClaim)
	}
}

func TestEvaluateEvidenceFailsWhenOperationOrVerifierEvidenceIsMissing(t *testing.T) {
	plan, ok := PlanForRequest("Summarize student scores in a gradebook and validate completion status")
	if !ok {
		t.Fatal("PlanForRequest returned ok=false")
	}
	result := EvaluateEvidence(plan, Evidence{
		ExecutedAtomIDs: []string{"group_summarize"},
		VerifierPasses:  map[string]bool{},
	})
	if result.Pass {
		t.Fatalf("EvaluateEvidence pass=true want false")
	}
	if len(result.Reasons) < 2 {
		t.Fatalf("reasons=%v want missing operation and verifier reasons", result.Reasons)
	}
}

func TestPlanForRequestClassifiesCashFlowMonitorContinuity(t *testing.T) {
	plan, ok := PlanForRequest("Track cash flow opening balance, summarize inflows, roll forward closing balance, and protect formulas")
	if !ok {
		t.Fatal("PlanForRequest returned ok=false")
	}
	if plan.OrganismID != "cash_flow_monitor" {
		t.Fatalf("organism=%q want cash_flow_monitor", plan.OrganismID)
	}
	wantOps := []string{"group_summarize", "extend_table_formulas", "copy_period_sheet", "roll_forward_period", "protect_formula_cells"}
	if !sameStrings(plan.OperationSequence, wantOps) {
		t.Fatalf("operation sequence=%v want %v", plan.OperationSequence, wantOps)
	}
}

func TestPlanForRequestClassifiesTimesheetHoursLog(t *testing.T) {
	plan, ok := PlanForRequest("Copy weekly timesheet, append employee hours, validate work code, extend pay formulas, and protect totals")
	if !ok {
		t.Fatal("PlanForRequest returned ok=false")
	}
	if plan.OrganismID != "timesheet_hours_log" {
		t.Fatalf("organism=%q want timesheet_hours_log", plan.OrganismID)
	}
	wantOps := []string{"copy_period_sheet", "append_structured_rows", "extend_table_formulas", "add_data_validation", "protect_formula_cells"}
	if !sameStrings(plan.OperationSequence, wantOps) {
		t.Fatalf("operation sequence=%v want %v", plan.OperationSequence, wantOps)
	}
}

func TestPlanForRequestClassifiesAttendanceRegister(t *testing.T) {
	plan, ok := PlanForRequest("Copy attendance register to the next class period, validate attendance status, and protect total formulas")
	if !ok {
		t.Fatal("PlanForRequest returned ok=false")
	}
	if plan.OrganismID != "attendance_register" {
		t.Fatalf("organism=%q want attendance_register", plan.OrganismID)
	}
	wantOps := []string{"copy_period_sheet", "add_data_validation", "protect_formula_cells"}
	if !sameStrings(plan.OperationSequence, wantOps) {
		t.Fatalf("operation sequence=%v want %v", plan.OperationSequence, wantOps)
	}
}

func TestPlanForRequestClassifiesProjectTimelineTracker(t *testing.T) {
	plan, ok := PlanForRequest("Copy project timeline period, extend task progress formulas, validate task status, summarize timeline status, and protect formulas")
	if !ok {
		t.Fatal("PlanForRequest returned ok=false")
	}
	if plan.OrganismID != "project_timeline_tracker" {
		t.Fatalf("organism=%q want project_timeline_tracker", plan.OrganismID)
	}
	wantOps := []string{"copy_period_sheet", "extend_table_formulas", "add_data_validation", "group_summarize", "protect_formula_cells"}
	if !sameStrings(plan.OperationSequence, wantOps) {
		t.Fatalf("operation sequence=%v want %v", plan.OperationSequence, wantOps)
	}
}

func TestPlanForRequestClassifiesShiftRosterPlanner(t *testing.T) {
	plan, ok := PlanForRequest("Copy weekly shift roster, validate shift codes, and protect coverage formulas")
	if !ok {
		t.Fatal("PlanForRequest returned ok=false")
	}
	if plan.OrganismID != "shift_roster_planner" {
		t.Fatalf("organism=%q want shift_roster_planner", plan.OrganismID)
	}
	wantOps := []string{"copy_period_sheet", "add_data_validation", "protect_formula_cells"}
	if !sameStrings(plan.OperationSequence, wantOps) {
		t.Fatalf("operation sequence=%v want %v", plan.OperationSequence, wantOps)
	}
}

func TestPlanForRequestClassifiesConstructionCostTracker(t *testing.T) {
	plan, ok := PlanForRequest("Append construction cost row, extend variance formulas, highlight budget overrun, and protect forecast formulas")
	if !ok {
		t.Fatal("PlanForRequest returned ok=false")
	}
	if plan.OrganismID != "construction_cost_tracker" {
		t.Fatalf("organism=%q want construction_cost_tracker", plan.OrganismID)
	}
	wantOps := []string{"append_structured_rows", "extend_table_formulas", "highlight_threshold", "protect_formula_cells"}
	if !sameStrings(plan.OperationSequence, wantOps) {
		t.Fatalf("operation sequence=%v want %v", plan.OperationSequence, wantOps)
	}
}

func TestPlanForRequestClassifiesWarehouseReorderTracker(t *testing.T) {
	plan, ok := PlanForRequest("Enrich warehouse reorder stock with SKU location, flag low stock, append reorder candidate, validate SKU, and protect reorder formulas")
	if !ok {
		t.Fatal("PlanForRequest returned ok=false")
	}
	if plan.OrganismID != "warehouse_reorder_tracker" {
		t.Fatalf("organism=%q want warehouse_reorder_tracker", plan.OrganismID)
	}
	wantOps := []string{"join_lookup", "highlight_threshold", "append_structured_rows", "add_data_validation", "protect_formula_cells"}
	if !sameStrings(plan.OperationSequence, wantOps) {
		t.Fatalf("operation sequence=%v want %v", plan.OperationSequence, wantOps)
	}
}

func TestPlanForRequestClassifiesExpenseReimbursement(t *testing.T) {
	plan, ok := PlanForRequest("Append expense reimbursement items, validate receipt status, extend reimbursable totals, protect formulas, and generate a printable claim")
	if !ok {
		t.Fatal("PlanForRequest returned ok=false")
	}
	if plan.OrganismID != "expense_reimbursement" {
		t.Fatalf("organism=%q want expense_reimbursement", plan.OrganismID)
	}
	wantOps := []string{"append_structured_rows", "extend_table_formulas", "add_data_validation", "protect_formula_cells", "generate_printable_form"}
	if !sameStrings(plan.OperationSequence, wantOps) {
		t.Fatalf("operation sequence=%v want %v", plan.OperationSequence, wantOps)
	}
}

func TestPlanForRequestClassifiesPurchaseOrderControl(t *testing.T) {
	plan, ok := PlanForRequest("Append purchase order line items, extend order totals, validate PO status, protect formulas, and generate a printable purchase order")
	if !ok {
		t.Fatal("PlanForRequest returned ok=false")
	}
	if plan.OrganismID != "purchase_order_control" {
		t.Fatalf("organism=%q want purchase_order_control", plan.OrganismID)
	}
	wantOps := []string{"append_structured_rows", "extend_table_formulas", "add_data_validation", "protect_formula_cells", "generate_printable_form"}
	if !sameStrings(plan.OperationSequence, wantOps) {
		t.Fatalf("operation sequence=%v want %v", plan.OperationSequence, wantOps)
	}
}

func TestPlanForRequestClassifiesProcurementReconciliation(t *testing.T) {
	plan, ok := PlanForRequest("Validate invoice status, reconcile procurement PO and invoice amounts, and protect review formulas")
	if !ok {
		t.Fatal("PlanForRequest returned ok=false")
	}
	if plan.OrganismID != "procurement_reconciliation" {
		t.Fatalf("organism=%q want procurement_reconciliation", plan.OrganismID)
	}
	wantOps := []string{"add_data_validation", "reconcile_tables", "protect_formula_cells"}
	if !sameStrings(plan.OperationSequence, wantOps) {
		t.Fatalf("operation sequence=%v want %v", plan.OperationSequence, wantOps)
	}
}

func TestPlanForRequestClassifiesTrainingCompletionMatrix(t *testing.T) {
	plan, ok := PlanForRequest("Summarize training completion by employee, validate training status, protect completion formulas, and generate a printable training report")
	if !ok {
		t.Fatal("PlanForRequest returned ok=false")
	}
	if plan.OrganismID != "training_completion_matrix" {
		t.Fatalf("organism=%q want training_completion_matrix", plan.OrganismID)
	}
	wantOps := []string{"group_summarize", "add_data_validation", "protect_formula_cells", "generate_printable_form"}
	if !sameStrings(plan.OperationSequence, wantOps) {
		t.Fatalf("operation sequence=%v want %v", plan.OperationSequence, wantOps)
	}
}

func TestPlanForRequestRejectsFinancialAdviceClaim(t *testing.T) {
	plan, ok := PlanForRequest("Give financial advice and certify the amortization correctness of this loan repayment schedule")
	if !ok {
		t.Fatal("PlanForRequest returned ok=false for loan schedule mechanics")
	}
	if plan.OrganismID != "loan_repayment_calculator" {
		t.Fatalf("organism=%q want loan_repayment_calculator", plan.OrganismID)
	}
	result := EvaluateEvidence(plan, Evidence{
		ExecutedAtomIDs: []string{"extend_table_formulas", "protect_formula_cells"},
		VerifierPasses: map[string]bool{
			"loan_repayment_calculator_verifier": true,
		},
	})
	if !result.Pass {
		t.Fatalf("EvaluateEvidence pass=false reasons=%v", result.Reasons)
	}
	if result.RuntimeClaim != "template_class_plan_verified_with_non_claims" {
		t.Fatalf("runtime claim=%q want template_class_plan_verified_with_non_claims", result.RuntimeClaim)
	}
	if len(result.NonClaims) == 0 {
		t.Fatalf("non-claims missing")
	}
}

func sameStrings(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for index := range got {
		if got[index] != want[index] {
			return false
		}
	}
	return true
}
