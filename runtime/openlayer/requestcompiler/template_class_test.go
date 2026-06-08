package requestcompiler

import "testing"

func TestTemplateClassPlanHintForRequestResolvesInvoiceOrganismBeforeOperationValidation(t *testing.T) {
	plan := TemplateClassPlanHintForRequest("Build an invoice line billing template")
	if plan == nil {
		t.Fatal("missing template class plan hint")
	}
	if plan.OrganismID != "invoice_line_item_billing" {
		t.Fatalf("organism_id=%q want invoice_line_item_billing", plan.OrganismID)
	}
	want := []string{"append_structured_rows", "extend_table_formulas", "add_data_validation", "protect_formula_cells", "generate_printable_form"}
	if !sameStringSlice(plan.OperationSequence, want) {
		t.Fatalf("operation sequence=%v want %v", plan.OperationSequence, want)
	}
	if len(plan.RequiredVerifierSpecs) != 1 || plan.RequiredVerifierSpecs[0] != "invoice_line_item_billing_verifier" {
		t.Fatalf("verifier specs=%v want invoice_line_item_billing_verifier", plan.RequiredVerifierSpecs)
	}
}

func sameStringSlice(got, want []string) bool {
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
