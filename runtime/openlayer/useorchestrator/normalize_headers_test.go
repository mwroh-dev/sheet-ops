package useorchestrator

import (
	"testing"

	runtimeworkbookcase "github.com/mwroh/sheet-ops/runtime/workbookcase"
)

func TestNormalizeHeadersUseRequestBuildsHeaderNormalizationTaskSpec(t *testing.T) {
	req := UseRequest{
		ScenarioID:  "normalize-headers",
		RequestText: "Movements 시트 헤더를 표준 필드명으로 정규화한다.",
		InputFile:   "in.xlsx",
		SheetName:   "Movements",
		OutputFile:  "out.xlsx",
		Operation:   runtimeworkbookcase.NormalizeHeadersOperationName,
		HeaderRow:   1,
		HeaderMappings: []HeaderMapping{
			{From: "SKU ID", To: "sku"},
			{From: "Qty In", To: "quantity_in"},
		},
	}

	spec := taskSpecFromUseRequest(req)
	if spec.Operation != runtimeworkbookcase.NormalizeHeadersOperationName {
		t.Fatalf("operation=%q want normalize_headers", spec.Operation)
	}
	if spec.CompositionKind != "header_normalization" {
		t.Fatalf("composition_kind=%q want header_normalization", spec.CompositionKind)
	}
	if spec.HeaderRow != 1 || len(spec.HeaderMappings) != 2 {
		t.Fatalf("header row/mappings=%d/%v", spec.HeaderRow, spec.HeaderMappings)
	}
}

func TestNormalizeHeadersValidatedRequestBuildsHeaderNormalizationTaskSpec(t *testing.T) {
	req := ValidatedExecutionRequest{
		ScenarioID:      "normalize-headers",
		RequestKind:     "structured_use_request",
		InputFile:       "in.xlsx",
		SourceSheet:     "Movements",
		OutputFile:      "out.xlsx",
		ExecutionKind:   "composition",
		CompositionKind: "header_normalization",
		HeaderRow:       1,
		HeaderMappings: []HeaderMapping{
			{From: "SKU ID", To: "sku"},
			{From: "Qty In", To: "quantity_in"},
		},
	}

	spec, err := taskSpecFromValidatedExecutionRequest(req)
	if err != nil {
		t.Fatalf("taskSpecFromValidatedExecutionRequest: %v", err)
	}
	if spec.Operation != runtimeworkbookcase.NormalizeHeadersOperationName {
		t.Fatalf("operation=%q want normalize_headers", spec.Operation)
	}
	if spec.CompositionKind != "header_normalization" {
		t.Fatalf("composition_kind=%q want header_normalization", spec.CompositionKind)
	}
	if spec.HeaderRow != 1 || len(spec.HeaderMappings) != 2 {
		t.Fatalf("header row/mappings=%d/%v", spec.HeaderRow, spec.HeaderMappings)
	}
}
