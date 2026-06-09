package compiler

import (
	"testing"

	"github.com/mwroh/sheet-ops/runtime/taskspec"
)

func TestCompileNormalizeHeadersOperationBuildsHeaderNormalizationIR(t *testing.T) {
	task := taskspec.BuildNormalizeHeadersTask(taskspec.NormalizeHeadersRequest{
		RequestText: "Normalize line item headers",
		InputFile:   "input.xlsx",
		SourceSheet: "LineItems",
		OutputFile:  "output.xlsx",
		HeaderRow:   1,
		HeaderMappings: []taskspec.HeaderMapping{
			{From: "SKU ID", To: "sku"},
			{From: "Qty", To: "quantity"},
		},
	})

	ir, err := CompileNormalizeHeadersOperation(task)
	if err != nil {
		t.Fatalf("CompileNormalizeHeadersOperation: %v", err)
	}
	if ir.ExecutionKind != taskspec.ExecutionKindComposition {
		t.Fatalf("execution_kind=%q want composition", ir.ExecutionKind)
	}
	if ir.CompositionKind != taskspec.CompositionKindHeaderNormalization {
		t.Fatalf("composition_kind=%q want header_normalization", ir.CompositionKind)
	}
	if ir.OperationFamily != taskspec.OperationFamilyNormalizeHeaders {
		t.Fatalf("operation_family=%q want normalize_headers", ir.OperationFamily)
	}
	if ir.HeaderRow != 1 {
		t.Fatalf("header_row=%d want 1", ir.HeaderRow)
	}
	if got, want := len(ir.HeaderMappings), 2; got != want {
		t.Fatalf("header_mappings=%d want %d", got, want)
	}
	if ir.HeaderMappings[0].From != "SKU ID" || ir.HeaderMappings[0].To != "sku" {
		t.Fatalf("first mapping=%+v", ir.HeaderMappings[0])
	}
	if !ir.PreserveOriginal {
		t.Fatalf("preserve_original=false want true")
	}
}

func TestCompileNormalizeHeadersOperationRejectsDuplicateTargets(t *testing.T) {
	task := taskspec.BuildNormalizeHeadersTask(taskspec.NormalizeHeadersRequest{
		InputFile:   "input.xlsx",
		SourceSheet: "LineItems",
		OutputFile:  "output.xlsx",
		HeaderRow:   1,
		HeaderMappings: []taskspec.HeaderMapping{
			{From: "SKU ID", To: "sku"},
			{From: "Item Code", To: "sku"},
		},
	})

	if _, err := CompileNormalizeHeadersOperation(task); err == nil {
		t.Fatalf("CompileNormalizeHeadersOperation succeeded; want duplicate target error")
	}
}
