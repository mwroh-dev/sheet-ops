package main

import (
	"path/filepath"
	"testing"

	runtimeschema "github.com/mwroh/sheet-ops/runtime/schema"
)

func TestWriteOrchestratorDecisionOutputSchemaAllowsNullTemplateClassPlan(t *testing.T) {
	schemaPath := filepath.Join(t.TempDir(), "orchestrator-decision-output.schema.json")
	if err := writeOrchestratorDecisionOutputSchema(schemaPath); err != nil {
		t.Fatalf("writeOrchestratorDecisionOutputSchema: %v", err)
	}

	decision := map[string]any{
		"scenario_id": "blocked-template-plan",
		"decision":    "blocked",
		"request_compiler_loop_state": map[string]any{
			"scenario_id":  "blocked-template-plan",
			"outcome":      "blocked",
			"parent_state": "BLOCKED",
			"specialists": []any{
				map[string]any{
					"role":                  "request-compiler",
					"carrier":               "default",
					"model":                 "codex-default",
					"reasoning_effort":      "high",
					"state":                 "BLOCKED",
					"session_id":            "request-compiler-001",
					"spawn_completed_count": 1,
					"wait_completed_count":  1,
					"close_completed_count": 1,
					"last_message":          "blocked",
				},
			},
		},
		"validated_execution_request": nil,
		"template_class_plan":         nil,
		"repair_advice": map[string]any{
			"summary":           "not enough evidence to choose a supported operation",
			"suggested_actions": []any{"provide a more specific workbook request"},
			"assumptions":       []any{},
		},
	}

	if err := runtimeschema.ValidateStruct(schemaPath, decision); err != nil {
		t.Fatalf("ValidateStruct with null template_class_plan: %v", err)
	}
}
