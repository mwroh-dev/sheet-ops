package templateclass

import "strings"

type Plan struct {
	OrganismID            string
	OperationSequence     []string
	RequiredVerifierSpecs []string
	NonClaims             []string
}

type Evidence struct {
	ExecutedAtomIDs []string
	VerifierPasses  map[string]bool
}

type EvaluationResult struct {
	Pass         bool
	RuntimeClaim string
	Reasons      []string
	NonClaims    []string
}

var plans = []Plan{
	{
		OrganismID:            "invoice_line_item_billing",
		OperationSequence:     []string{"append_structured_rows", "extend_table_formulas", "add_data_validation", "protect_formula_cells", "generate_printable_form"},
		RequiredVerifierSpecs: []string{"invoice_line_item_billing_verifier"},
		NonClaims:             []string{"does not infer invoice layout", "does not render PDF or prove pagination", "does not perform tax or accounting advice"},
	},
	{
		OrganismID:            "monthly_budget_control",
		OperationSequence:     []string{"group_summarize", "highlight_threshold", "copy_period_sheet", "roll_forward_period", "protect_formula_cells"},
		RequiredVerifierSpecs: []string{"monthly_budget_control_verifier"},
		NonClaims:             []string{"does not create financial advice", "does not infer budget policy", "does not create native pivot tables"},
	},
	{
		OrganismID:            "cash_flow_monitor",
		OperationSequence:     []string{"group_summarize", "extend_table_formulas", "copy_period_sheet", "roll_forward_period", "protect_formula_cells"},
		RequiredVerifierSpecs: []string{"cash_flow_monitor_verifier"},
		NonClaims:             []string{"does not create financial advice", "does not infer cash-flow policy", "does not create native pivot tables"},
	},
	{
		OrganismID:            "timesheet_hours_log",
		OperationSequence:     []string{"copy_period_sheet", "append_structured_rows", "extend_table_formulas", "add_data_validation", "protect_formula_cells"},
		RequiredVerifierSpecs: []string{"timesheet_hours_log_verifier"},
		NonClaims:             []string{"does not calculate payroll taxes", "does not infer labor policy", "does not certify billable hours"},
	},
	{
		OrganismID:            "inventory_movement_log",
		OperationSequence:     []string{"normalize_headers", "append_structured_rows", "join_lookup", "protect_formula_cells", "reconcile_tables"},
		RequiredVerifierSpecs: []string{"inventory_movement_log_verifier"},
		NonClaims:             []string{"does not infer warehouse policy", "does not resolve duplicate keys", "does not support fuzzy reconciliation"},
	},
	{
		OrganismID:            "student_gradebook",
		OperationSequence:     []string{"group_summarize", "add_data_validation", "extend_table_formulas", "protect_formula_cells"},
		RequiredVerifierSpecs: []string{"student_gradebook_verifier"},
		NonClaims:             []string{"does not infer grading policy", "does not auto-grow arbitrary matrices", "does not create native Excel pivots"},
	},
	{
		OrganismID:            "loan_repayment_calculator",
		OperationSequence:     []string{"extend_table_formulas", "protect_formula_cells"},
		RequiredVerifierSpecs: []string{"loan_repayment_calculator_verifier"},
		NonClaims:             []string{"does not provide financial advice", "does not prove payoff termination", "does not certify amortization correctness"},
	},
}

func PlanForRequest(requestText string) (Plan, bool) {
	normalized := strings.ToLower(requestText)
	switch {
	case containsAll(normalized, "invoice", "line"):
		return planForOrganism("invoice_line_item_billing")
	case containsAny(normalized, "budget", "variance", "over-budget", "over budget"):
		return planForOrganism("monthly_budget_control")
	case containsAny(normalized, "cash flow", "cash-flow", "opening balance", "closing balance"):
		return planForOrganism("cash_flow_monitor")
	case containsAny(normalized, "timesheet", "time sheet", "employee hours", "billable hours"):
		return planForOrganism("timesheet_hours_log")
	case containsAny(normalized, "inventory", "stock", "sku"):
		return planForOrganism("inventory_movement_log")
	case containsAny(normalized, "gradebook", "student score", "student scores"):
		return planForOrganism("student_gradebook")
	case containsAny(normalized, "loan", "amortization", "repayment"):
		return planForOrganism("loan_repayment_calculator")
	default:
		return Plan{}, false
	}
}

func EvaluateEvidence(plan Plan, evidence Evidence) EvaluationResult {
	executed := stringSet(evidence.ExecutedAtomIDs)
	var reasons []string
	for _, atom := range plan.OperationSequence {
		if !executed[atom] {
			reasons = append(reasons, "missing executed atom "+atom)
		}
	}
	for _, verifier := range plan.RequiredVerifierSpecs {
		if !evidence.VerifierPasses[verifier] {
			reasons = append(reasons, "missing verifier pass "+verifier)
		}
	}
	if len(reasons) > 0 {
		return EvaluationResult{
			Pass:         false,
			RuntimeClaim: "template_class_plan_unverified",
			Reasons:      reasons,
			NonClaims:    append([]string(nil), plan.NonClaims...),
		}
	}
	claim := "template_class_plan_verified"
	if len(plan.NonClaims) > 0 {
		claim = "template_class_plan_verified_with_non_claims"
	}
	if plan.OrganismID == "invoice_line_item_billing" {
		claim = "template_class_plan_verified"
	}
	return EvaluationResult{
		Pass:         true,
		RuntimeClaim: claim,
		NonClaims:    append([]string(nil), plan.NonClaims...),
	}
}

func planForOrganism(organismID string) (Plan, bool) {
	for _, plan := range plans {
		if plan.OrganismID == organismID {
			return clonePlan(plan), true
		}
	}
	return Plan{}, false
}

func clonePlan(plan Plan) Plan {
	plan.OperationSequence = append([]string(nil), plan.OperationSequence...)
	plan.RequiredVerifierSpecs = append([]string(nil), plan.RequiredVerifierSpecs...)
	plan.NonClaims = append([]string(nil), plan.NonClaims...)
	return plan
}

func containsAll(value string, needles ...string) bool {
	for _, needle := range needles {
		if !strings.Contains(value, needle) {
			return false
		}
	}
	return true
}

func containsAny(value string, needles ...string) bool {
	for _, needle := range needles {
		if strings.Contains(value, needle) {
			return true
		}
	}
	return false
}

func stringSet(values []string) map[string]bool {
	set := map[string]bool{}
	for _, value := range values {
		set[value] = true
	}
	return set
}
