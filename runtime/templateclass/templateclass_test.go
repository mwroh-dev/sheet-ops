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
