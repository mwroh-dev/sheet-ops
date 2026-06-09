package compiler

import (
	"testing"

	"github.com/mwroh/sheet-ops/runtime/taskspec"
)

func TestCompileGeneratePrintableFormOperationBuildsPrintableFormIR(t *testing.T) {
	task := taskspec.BuildGeneratePrintableFormTask(taskspec.GeneratePrintableFormRequest{
		RequestText: "Generate invoice printable form",
		InputFile:   "invoice.xlsx",
		SourceSheet: "InvoiceData",
		TargetSheet: "InvoicePrint",
		OutputFile:  "invoice-output.xlsx",
		FormTitle:   "Invoice",
		PrintArea:   "A1:D8",
		FieldBindings: []taskspec.FormFieldBinding{
			{Label: "Invoice No", SourceSheet: "InvoiceData", SourceCell: "B2", LabelCell: "A2", ValueCell: "B2"},
		},
		TableBinding: &taskspec.FormTableBinding{
			SourceSheet:   "LineItems",
			SourceColumns: []string{"sku", "quantity"},
			HeaderStart:   "A5",
			DataStart:     "A6",
		},
	})

	ir, err := CompileGeneratePrintableFormOperation(task)
	if err != nil {
		t.Fatalf("CompileGeneratePrintableFormOperation: %v", err)
	}
	if ir.CompositionKind != taskspec.CompositionKindPrintableForm {
		t.Fatalf("composition_kind=%q want printable_form", ir.CompositionKind)
	}
	if ir.OperationFamily != taskspec.OperationFamilyGeneratePrintableForm {
		t.Fatalf("operation_family=%q want generate_printable_form", ir.OperationFamily)
	}
	if len(ir.FieldBindings) != 1 {
		t.Fatalf("field_bindings=%v want 1 binding", ir.FieldBindings)
	}
	if ir.TableBinding == nil || len(ir.TableBinding.SourceColumns) != 2 {
		t.Fatalf("table_binding=%+v want 2 columns", ir.TableBinding)
	}
}

func TestCompileGeneratePrintableFormOperationRejectsMissingTableBinding(t *testing.T) {
	task := taskspec.BuildGeneratePrintableFormTask(taskspec.GeneratePrintableFormRequest{
		InputFile:   "invoice.xlsx",
		SourceSheet: "InvoiceData",
		TargetSheet: "InvoicePrint",
		OutputFile:  "invoice-output.xlsx",
		FormTitle:   "Invoice",
		PrintArea:   "A1:D8",
		FieldBindings: []taskspec.FormFieldBinding{
			{Label: "Invoice No", SourceSheet: "InvoiceData", SourceCell: "B2", LabelCell: "A2", ValueCell: "B2"},
		},
	})

	if _, err := CompileGeneratePrintableFormOperation(task); err == nil {
		t.Fatalf("CompileGeneratePrintableFormOperation succeeded; want missing table binding error")
	}
}
