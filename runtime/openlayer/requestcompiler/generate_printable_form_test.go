package requestcompiler

import (
	"testing"

	runtimeinspect "github.com/mwroh/sheet-ops/runtime/inspect"
	runtimevalidate "github.com/mwroh/sheet-ops/runtime/validate"
)

func TestGeneratePrintableFormIntentCompilesThroughValidation(t *testing.T) {
	intent := normalizeIntent(NormalizedIntent{
		SourceSheetCandidates: []string{"InvoiceData"},
		CompositionCandidates: []string{CompositionCandidatePrintableForm},
		GeneratePrintableForm: GeneratePrintableFormIntent{
			TargetSheet: "InvoicePrint",
			FormTitle:   "Invoice",
			PrintArea:   "A1:D8",
			FieldBindings: []FormFieldBinding{
				{Label: "Invoice No", SourceSheet: "InvoiceData", SourceCell: "B2", LabelCell: "A2", ValueCell: "B2"},
			},
			TableBinding: &FormTableBinding{
				SourceSheet:   "LineItems",
				SourceColumns: []string{"sku", "quantity"},
				HeaderStart:   "A5",
				DataStart:     "A6",
			},
		},
		Materialization: MaterializationIntent{
			PreserveOriginal:      true,
			OutputDestinationMode: OutputDestinationModeNewWorkbook,
			WriteShape:            WriteShapeNewSheet,
		},
		Ambiguity: AmbiguityIntent{Markers: []string{}, UnresolvedFields: []string{}, CheckpointHints: []string{}},
	})
	decision := compileDecision(intent)
	if decision.SelectedOperation != "generate_printable_form" {
		t.Fatalf("selected_operation=%q want generate_printable_form", decision.SelectedOperation)
	}
	validator := runtimevalidate.Validator{
		RequestContext: runtimevalidate.RequestContext{ScenarioID: "generate-printable-form", RequestKind: "structured_use_request", OutputFile: "out.xlsx"},
	}
	result, err := validator.Validate(toRuntimeIntent(intent), toRuntimeDecision(decision), runtimeinspect.WorkbookFacts{
		InputFile: "in.xlsx",
		Sheets: []runtimeinspect.SheetFacts{
			{Name: "InvoiceData", Columns: []string{"field", "value"}},
			{Name: "LineItems", Columns: []string{"sku", "quantity"}},
		},
	})
	if err != nil {
		t.Fatalf("validate printable form intent: %v", err)
	}
	if result.ValidatedExecutionRequest == nil || result.ValidatedExecutionRequest.CompositionKind != "printable_form" {
		t.Fatalf("validated request=%+v want printable_form", result.ValidatedExecutionRequest)
	}
	if result.ValidatedExecutionRequest.TableBinding == nil {
		t.Fatalf("table binding missing from validated request")
	}
}
