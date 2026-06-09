package useorchestrator

import (
	"testing"

	runtimeworkbookcase "github.com/mwroh/sheet-ops/runtime/workbookcase"
)

func TestGeneratePrintableFormUseRequestBuildsPrintableFormTaskSpec(t *testing.T) {
	req := UseRequest{
		ScenarioID:  "generate-printable-form",
		RequestText: "invoice printable form을 생성한다.",
		InputFile:   "in.xlsx",
		SheetName:   "InvoiceData",
		TargetSheet: "InvoicePrint",
		OutputFile:  "out.xlsx",
		Operation:   runtimeworkbookcase.GeneratePrintableFormOperationName,
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
	}

	spec := taskSpecFromUseRequest(req)
	if spec.Operation != runtimeworkbookcase.GeneratePrintableFormOperationName {
		t.Fatalf("operation=%q want generate_printable_form", spec.Operation)
	}
	if spec.CompositionKind != "printable_form" {
		t.Fatalf("composition_kind=%q want printable_form", spec.CompositionKind)
	}
	if spec.TableBinding == nil || len(spec.FieldBindings) != 1 {
		t.Fatalf("form bindings missing: fields=%v table=%+v", spec.FieldBindings, spec.TableBinding)
	}
}

func TestGeneratePrintableFormValidatedRequestBuildsPrintableFormTaskSpec(t *testing.T) {
	req := ValidatedExecutionRequest{
		ScenarioID:      "generate-printable-form",
		RequestKind:     "structured_use_request",
		InputFile:       "in.xlsx",
		SourceSheet:     "InvoiceData",
		TargetSheet:     "InvoicePrint",
		OutputFile:      "out.xlsx",
		ExecutionKind:   "composition",
		CompositionKind: "printable_form",
		FormTitle:       "Invoice",
		PrintArea:       "A1:D8",
		FieldBindings: []FormFieldBinding{
			{Label: "Invoice No", SourceSheet: "InvoiceData", SourceCell: "B2", LabelCell: "A2", ValueCell: "B2"},
		},
		TableBinding: &FormTableBinding{
			SourceSheet:   "LineItems",
			SourceColumns: []string{"sku", "quantity"},
			HeaderStart:   "A5",
			DataStart:     "A6",
		},
	}

	spec, err := taskSpecFromValidatedExecutionRequest(req)
	if err != nil {
		t.Fatalf("taskSpecFromValidatedExecutionRequest: %v", err)
	}
	if spec.Operation != runtimeworkbookcase.GeneratePrintableFormOperationName {
		t.Fatalf("operation=%q want generate_printable_form", spec.Operation)
	}
	if spec.CompositionKind != "printable_form" {
		t.Fatalf("composition_kind=%q want printable_form", spec.CompositionKind)
	}
	if spec.TableBinding == nil || len(spec.FieldBindings) != 1 {
		t.Fatalf("form bindings missing: fields=%v table=%+v", spec.FieldBindings, spec.TableBinding)
	}
}
