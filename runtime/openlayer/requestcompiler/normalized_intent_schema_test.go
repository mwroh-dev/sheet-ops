package requestcompiler

import (
	"encoding/json"
	"testing"
)

func TestNormalizedIntentSchemaRejectsForeignPrintableFields(t *testing.T) {
	tests := []struct {
		name  string
		patch func(map[string]any)
	}{
		{
			name: "summary rejects printable form fields",
			patch: func(value map[string]any) {
				value["summary"] = map[string]any{
					"target_sheet": "Summary",
					"summary_mode": "values",
					"metrics":      []any{map[string]any{"kind": "sum", "column": "amount"}},
					"form_title":   "Invoice",
				}
			},
		},
		{
			name: "aggregate rejects printable form fields",
			patch: func(value map[string]any) {
				value["aggregates"] = []any{map[string]any{"kind": "sum", "column": "amount", "print_area": "A1:D8"}}
			},
		},
		{
			name: "validation rule rejects printable form fields",
			patch: func(value map[string]any) {
				value["add_data_validation"] = map[string]any{
					"validation_rule": map[string]any{
						"ranges":         []any{"A2:A10"},
						"rule_type":      "list",
						"allowed_values": []any{"Open", "Closed"},
						"allow_blank":    false,
						"table_binding":  map[string]any{"source_sheet": "Data", "source_columns": []any{"Status"}, "header_start": "A1", "data_start": "A2"},
					},
				}
			},
		},
		{
			name: "compare mapping rejects printable form fields",
			patch: func(value map[string]any) {
				value["reconcile_tables"] = map[string]any{
					"target_sheet": "Recon",
					"left_key":     "sku",
					"right_key":    "sku",
					"compare_mappings": []any{
						map[string]any{"left_column": "qty", "right_column": "qty", "field_bindings": []any{}},
					},
				}
			},
		},
		{
			name: "materialization rejects printable form fields",
			patch: func(value map[string]any) {
				value["materialization"].(map[string]any)["form_title"] = "Invoice"
			},
		},
		{
			name: "root rejects printable form fields",
			patch: func(value map[string]any) {
				value["print_area"] = "A1:D8"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value := minimalNormalizedIntentDocument()
			tt.patch(value)
			raw, err := json.Marshal(value)
			if err != nil {
				t.Fatalf("Marshal normalized intent: %v", err)
			}
			if _, err := LoadNormalizedIntentFromBytes(raw); err == nil {
				t.Fatalf("LoadNormalizedIntentFromBytes passed with foreign printable field; want schema rejection")
			}
		})
	}
}

func TestNormalizedIntentSchemaAcceptsPrintableFormSurface(t *testing.T) {
	value := minimalNormalizedIntentDocument()
	value["source_sheet_candidates"] = []any{"InvoiceData"}
	value["composition_candidates"] = []any{"printable_form"}
	value["generate_printable_form"] = map[string]any{
		"target_sheet": "InvoicePrint",
		"form_title":   "Invoice",
		"print_area":   "A1:D8",
		"field_bindings": []any{
			map[string]any{
				"label":        "Invoice No",
				"source_sheet": "InvoiceData",
				"source_cell":  "B2",
				"label_cell":   "A2",
				"value_cell":   "B2",
			},
		},
		"table_binding": map[string]any{
			"source_sheet":   "LineItems",
			"source_columns": []any{"sku", "quantity"},
			"header_start":   "A5",
			"data_start":     "A6",
		},
	}
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("Marshal printable normalized intent: %v", err)
	}
	if _, err := LoadNormalizedIntentFromBytes(raw); err != nil {
		t.Fatalf("LoadNormalizedIntentFromBytes rejected printable form surface: %v", err)
	}
}

func TestNormalizedIntentSchemaAcceptsReconcileTablesSurface(t *testing.T) {
	value := minimalNormalizedIntentDocument()
	value["source_sheet_candidates"] = []any{"Actual"}
	value["lookup_sheet_candidates"] = []any{"Expected"}
	value["composition_candidates"] = []any{"table_reconciliation"}
	value["reconcile_tables"] = map[string]any{
		"target_sheet": "Recon",
		"left_key":     "sku",
		"right_key":    "sku",
		"compare_mappings": []any{
			map[string]any{"left_column": "qty", "right_column": "qty", "as": "quantity_delta"},
		},
	}
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("Marshal reconcile normalized intent: %v", err)
	}
	if _, err := LoadNormalizedIntentFromBytes(raw); err != nil {
		t.Fatalf("LoadNormalizedIntentFromBytes rejected reconcile surface: %v", err)
	}
}

func minimalNormalizedIntentDocument() map[string]any {
	return map[string]any{
		"materialization": map[string]any{
			"preserve_original":       true,
			"output_destination_mode": "new_workbook",
			"write_shape":             "new_sheet",
		},
		"ambiguity": map[string]any{
			"markers":           []any{},
			"unresolved_fields": []any{},
			"checkpoint_hints":  []any{},
		},
	}
}
