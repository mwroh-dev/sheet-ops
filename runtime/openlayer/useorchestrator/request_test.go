package useorchestrator

import "testing"

func TestLoadOrchestratorDecisionFromBytesPreservesTemplateClassPlan(t *testing.T) {
	raw := []byte(`{
  "scenario_id": "invoice-template-class",
  "decision": "execute",
  "request_compiler_loop_state": {
    "role": "request-compiler",
    "model": "codex-default",
    "reasoning_effort": "high",
    "agent_id": "agent-001",
    "lifecycle": {
      "spawned": true,
      "completed": true,
      "closed": true
    },
    "lifecycle_proof": {
      "runtime_rerun": false,
      "workbook_mutated": false,
      "additional_workspace_files_read": false,
      "evidence_source": "deterministic request compiler",
      "insufficient_evidence": false
    }
  },
  "validated_execution_request": {
    "scenario_id": "invoice-template-class",
    "execution_kind": "composition",
    "composition_kind": "period_copy",
    "request_kind": "prompt_text",
    "request_text": "copy period sheet",
    "input_file": "/tmp/in.xlsx",
    "source_sheet": "Jan",
    "target_sheet": "Feb",
    "output_file": "/tmp/out.xlsx"
  },
  "template_class_plan": {
    "organism_id": "invoice_line_item_billing",
    "operation_sequence": [
      "append_structured_rows",
      "extend_table_formulas",
      "add_data_validation",
      "protect_formula_cells",
      "generate_printable_form"
    ],
    "required_verifier_specs": ["invoice_line_item_billing_verifier"],
    "non_claims": ["does not infer invoice layout"]
  }
}`)

	decision, err := LoadOrchestratorDecisionFromBytes(raw)
	if err != nil {
		t.Fatalf("LoadOrchestratorDecisionFromBytes: %v", err)
	}
	if decision.TemplateClassPlan == nil {
		t.Fatal("missing template class plan")
	}
	if decision.TemplateClassPlan.OrganismID != "invoice_line_item_billing" {
		t.Fatalf("organism_id=%q want invoice_line_item_billing", decision.TemplateClassPlan.OrganismID)
	}
	if len(decision.TemplateClassPlan.OperationSequence) != 5 {
		t.Fatalf("operation sequence=%v want 5 atoms", decision.TemplateClassPlan.OperationSequence)
	}
}

func TestLoadResultVerifierOutcomeFromBytesSupportsVerificationReviewShape(t *testing.T) {
	raw := []byte(`{
  "review": {
    "outcome": "pass",
    "needs_review": false,
    "blocked": false,
    "verification_review": {
      "scenario": "02-join-lookup",
      "output_file": "/tmp/output.xlsx",
      "verification_pass": true,
      "operation": "create_join_lookup_result_sheet",
      "evidence": [
        "Output workbook contains the joined lookup sheet",
        "Source workbook is unchanged"
      ],
      "conclusion": "The deterministic runtime summary supports that the requested joined lookup workbook outcome was satisfied."
    }
  },
  "loop_state": {
    "role": "result-verifier",
    "model": "codex-default",
    "reasoning_effort": "high",
    "agent_id": "agent-001",
    "lifecycle": {
      "spawned": true,
      "completed": true,
      "closed": true
    },
    "lifecycle_proof": {
      "runtime_rerun": false,
      "workbook_mutated": false,
      "additional_workspace_files_read": false,
      "evidence_source": "deterministic runtime summary only",
      "insufficient_evidence": false
    }
  }
}`)

	outcome, err := LoadResultVerifierOutcomeFromBytes(raw)
	if err != nil {
		t.Fatalf("LoadResultVerifierOutcomeFromBytes: %v", err)
	}
	if got, want := outcome.Review.Status, "pass"; got != want {
		t.Fatalf("Review.Status=%q want %q", got, want)
	}
	if got := outcome.Review.Summary; got == "" {
		t.Fatal("Review.Summary is empty")
	}
	if got, want := len(outcome.Review.Reasons), 2; got != want {
		t.Fatalf("len(Review.Reasons)=%d want %d", got, want)
	}
}

func TestLoadResultVerifierOutcomeFromBytesUsesReasonsWhenSummaryMissing(t *testing.T) {
	raw := []byte(`{
  "review": {
    "outcome": "pass",
    "scenario": "02-join-lookup",
    "output_file": "/tmp/output.xlsx",
    "operation": "create_join_lookup_result_sheet",
    "reasons": [
      "output workbook contains the joined lookup sheet",
      "source workbook is unchanged"
    ],
    "limitations": [
      "verified only against the provided deterministic runtime summary"
    ]
  },
  "loop_state": {
    "role": "result-verifier",
    "model": "codex-default",
    "reasoning_effort": "high",
    "started": true,
    "completed": true,
    "closed": true,
    "no_rerun": true,
    "no_mutation": true,
    "evidence_source": "provided deterministic runtime summary"
  }
}`)

	outcome, err := LoadResultVerifierOutcomeFromBytes(raw)
	if err != nil {
		t.Fatalf("LoadResultVerifierOutcomeFromBytes: %v", err)
	}
	if got, want := outcome.Review.Status, "pass"; got != want {
		t.Fatalf("Review.Status=%q want %q", got, want)
	}
	if got := outcome.Review.Summary; got == "" {
		t.Fatal("Review.Summary is empty")
	}
	if got, want := len(outcome.Review.Reasons), 2; got != want {
		t.Fatalf("len(Review.Reasons)=%d want %d", got, want)
	}
}

func TestLoadResultVerifierOutcomeFromBytesMapsVerifiedPassOutcome(t *testing.T) {
	raw := []byte(`{
  "review": {
    "outcome": "verified_pass",
    "scenario": "02-join-lookup",
    "output_file": "/tmp/output.xlsx",
    "verification_operation": "create_join_lookup_result_sheet",
    "satisfies_requested_workbook_outcome": true,
    "evidence": [
      "Verification pass is true.",
      "Output workbook contains the joined lookup sheet.",
      "Source workbook is unchanged."
    ],
    "limitations": [
      "No runtime rerun performed."
    ]
  },
  "loop_state": {
    "role": "result-verifier",
    "model": "codex-default",
    "reasoning_effort": "high",
    "entry": "sheet-ops use",
    "state": "complete",
    "runtime_rerun": false,
    "workbook_mutated": false,
    "additional_workspace_files_read": false,
    "deterministic_summary_used": true,
    "blocked": false,
    "needs_review": false
  }
}`)

	outcome, err := LoadResultVerifierOutcomeFromBytes(raw)
	if err != nil {
		t.Fatalf("LoadResultVerifierOutcomeFromBytes: %v", err)
	}
	if got, want := outcome.Review.Status, "pass"; got != want {
		t.Fatalf("Review.Status=%q want %q", got, want)
	}
	if got := outcome.Review.Summary; got == "" {
		t.Fatal("Review.Summary is empty")
	}
}

func TestLoadResultVerifierOutcomeFromBytesSupportsNestedVerificationReview(t *testing.T) {
	raw := []byte(`{
  "review": {
    "outcome": "pass",
    "scenario": "02-join-lookup",
    "output_file": "/tmp/output.xlsx",
    "verification": {
      "pass": true,
      "operation": "create_join_lookup_result_sheet",
      "review": "Runtime summary states the output workbook contains the joined lookup sheet and the source workbook is unchanged, satisfying the requested workbook outcome.",
      "reasons": [
        "output workbook contains the joined lookup sheet",
        "source workbook is unchanged"
      ]
    }
  },
  "loop_state": {
    "role": "result-verifier",
    "model": "codex-default",
    "reasoning_effort": "high",
    "agent_path": "agent-001",
    "status": "completed",
    "closed": true,
    "lifecycle_proof": {
      "evidence_basis": "deterministic runtime summary only",
      "reran_execution": false,
      "read_workspace_files": false,
      "mutated_files": false
    }
  }
}`)

	outcome, err := LoadResultVerifierOutcomeFromBytes(raw)
	if err != nil {
		t.Fatalf("LoadResultVerifierOutcomeFromBytes: %v", err)
	}
	if got, want := outcome.Review.Status, "pass"; got != want {
		t.Fatalf("Review.Status=%q want %q", got, want)
	}
	if got := outcome.Review.Summary; got == "" {
		t.Fatal("Review.Summary is empty")
	}
	if got, want := len(outcome.Review.Reasons), 2; got != want {
		t.Fatalf("len(Review.Reasons)=%d want %d", got, want)
	}
}

func TestLoadResultVerifierOutcomeFromBytesSupportsFindingsLimitsConclusionShape(t *testing.T) {
	raw := []byte(`{
  "review": {
    "outcome": "pass",
    "scenario": "02-join-lookup",
    "output_file": "/tmp/output.xlsx",
    "evidence_used": "deterministic runtime summary only",
    "verification_operation": "create_join_lookup_result_sheet",
    "verification_pass": true,
    "findings": [
      "Output workbook contains the joined lookup sheet.",
      "Source workbook is unchanged."
    ],
    "limits": [
      "Workbook contents were not independently inspected because the task explicitly restricted verification to the deterministic runtime summary."
    ],
    "conclusion": "The runtime result satisfies the requested workbook outcome based on the supplied deterministic verification evidence."
  },
  "loop_state": {
    "role": "result-verifier",
    "model": "codex-default",
    "reasoning_effort": "high",
    "status": "completed",
    "decision": "pass",
    "reran_runtime": false,
    "mutated_workbooks": false,
    "used_only_runtime_summary": true,
    "blocked": false,
    "needs_review": false
  }
}`)

	outcome, err := LoadResultVerifierOutcomeFromBytes(raw)
	if err != nil {
		t.Fatalf("LoadResultVerifierOutcomeFromBytes: %v", err)
	}
	if got, want := outcome.Review.Status, "pass"; got != want {
		t.Fatalf("Review.Status=%q want %q", got, want)
	}
	if got, want := outcome.Review.Summary, "The runtime result satisfies the requested workbook outcome based on the supplied deterministic verification evidence."; got != want {
		t.Fatalf("Review.Summary=%q want %q", got, want)
	}
	if got, want := len(outcome.Review.Reasons), 1; got != want {
		t.Fatalf("len(Review.Reasons)=%d want %d", got, want)
	}
}

func TestLoadResultVerifierOutcomeFromBytesMapsVerificationPassField(t *testing.T) {
	raw := []byte(`{
  "review": {
    "scenario": "01-create-summary-sheet",
    "output_file": "/tmp/output.xlsx",
    "verification_operation": "create_summary_sheet",
    "verification_pass": true,
    "assessment": "The deterministic runtime summary reports that the output workbook contains the expected summary sheet.",
    "limitations": "Review is limited to the deterministic runtime summary."
  },
  "loop_state": {
    "role": "result-verifier",
    "model": "codex-default",
    "reasoning_effort": "high",
    "status": "complete",
    "runtime_was_rerun": false,
    "files_were_mutated": false,
    "evidence_scope": "deterministic runtime summary only",
    "final_outcome": "passed"
  }
}`)

	outcome, err := LoadResultVerifierOutcomeFromBytes(raw)
	if err != nil {
		t.Fatalf("LoadResultVerifierOutcomeFromBytes: %v", err)
	}
	if outcome.Review.Status != "pass" {
		t.Fatalf("review status=%q want pass", outcome.Review.Status)
	}
}
