package requestcompiler

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

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

func TestTemplateLikeRequestPersistsTemplateClassPlanEvidence(t *testing.T) {
	tempDir := t.TempDir()
	decisionPath := filepath.Join(tempDir, "compiler_decision.json")
	plan := TemplateClassPlanHintForRequest("Build an invoice line item billing template, but first append this line item.")
	if plan == nil {
		t.Fatal("missing template class plan")
	}
	if err := writeCompilerDecisionJSON(decisionPath, Result{
		Decision: Decision{
			Status:            StatusCompiled,
			SelectedOperation: "append_structured_rows",
			Notes:             []string{"single atom selected, but template class evidence must still be retained"},
		},
		TemplateClassPlan: plan,
	}); err != nil {
		t.Fatalf("writeCompilerDecisionJSON: %v", err)
	}

	raw, err := os.ReadFile(decisionPath)
	if err != nil {
		t.Fatalf("ReadFile(%s): %v", decisionPath, err)
	}
	var decision struct {
		TemplateClassPlan struct {
			OrganismID            string   `json:"organism_id"`
			OperationSequence     []string `json:"operation_sequence"`
			RequiredVerifierSpecs []string `json:"required_verifier_specs"`
			NonClaims             []string `json:"non_claims"`
		} `json:"template_class_plan"`
	}
	if err := json.Unmarshal(raw, &decision); err != nil {
		t.Fatalf("Unmarshal compiler decision: %v", err)
	}
	if decision.TemplateClassPlan.OrganismID != "invoice_line_item_billing" {
		t.Fatalf("organism_id=%q want invoice_line_item_billing", decision.TemplateClassPlan.OrganismID)
	}
	if len(decision.TemplateClassPlan.OperationSequence) == 0 || len(decision.TemplateClassPlan.RequiredVerifierSpecs) == 0 || len(decision.TemplateClassPlan.NonClaims) == 0 {
		t.Fatalf("weak template_class_plan evidence: %+v", decision.TemplateClassPlan)
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
