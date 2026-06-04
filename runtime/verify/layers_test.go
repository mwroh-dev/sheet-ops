package verify

import (
	"path/filepath"
	"testing"

	"github.com/mwroh/sheet-ops/runtime/compiler"
	"github.com/xuri/excelize/v2"
)

func TestLayeredVerificationPreservesOperationSpecificFailure(t *testing.T) {
	result := VerificationResult{
		Pass:            false,
		OperationFamily: "group_summarize",
		OutputWorkbook:  "/tmp/output.xlsx",
		Reasons:         []string{"summary sheet \"Summary\" was not created"},
	}

	layered := WithVerificationLayers(result, []LayerResult{
		{Level: 1, Name: "file_opens", Pass: true},
		{Level: 2, Name: "expected_ranges_exist", Pass: false, Reasons: []string{"summary sheet missing"}},
	})

	if layered.Pass {
		t.Fatalf("pass=true want false")
	}
	if len(layered.Reasons) != 1 || layered.Reasons[0] != "summary sheet \"Summary\" was not created" {
		t.Fatalf("reasons=%v want original operation-specific failure", layered.Reasons)
	}
	if len(layered.Layers) != 2 {
		t.Fatalf("layers=%+v want 2 layers", layered.Layers)
	}
	if layered.Layers[1].Name != "expected_ranges_exist" || layered.Layers[1].Pass {
		t.Fatalf("second layer=%+v want expected_ranges_exist fail", layered.Layers[1])
	}
}

func TestDefaultVerificationLayersIncludeFileAndSourcePreservation(t *testing.T) {
	outputFile := filepath.Join(t.TempDir(), "output.xlsx")
	writeMinimalWorkbook(t, outputFile)
	ir := compiler.WorkbookOperationIR{OperationFamily: "write_values", PreserveOriginal: true}
	result := VerificationResult{
		Pass:                true,
		OperationFamily:     "write_values",
		OutputWorkbook:      outputFile,
		SourceSHA256Checked: true,
		Reasons:             []string{"pass"},
	}

	layered := WithDefaultLayers(ir, result)
	if len(layered.Layers) < 2 {
		t.Fatalf("layers=%+v want default layers", layered.Layers)
	}
	assertLayer(t, layered.Layers, "file_opens", true)
	assertLayer(t, layered.Layers, "expected_sheets_ranges", true)
	assertLayer(t, layered.Layers, "values_formulas_match", true)
	assertLayer(t, layered.Layers, "formulas_calculate", true)
	assertLayer(t, layered.Layers, "source_preserved", true)
	assertLayer(t, layered.Layers, "render_artifact_emission", false)
	assertLayer(t, layered.Layers, "semantic_task_specific", true)
	if len(layered.Layers) != 7 {
		t.Fatalf("layer count=%d want 7: %+v", len(layered.Layers), layered.Layers)
	}
}

func TestDefaultVerificationLayersFailFileOpenWhenOutputDoesNotOpen(t *testing.T) {
	ir := compiler.WorkbookOperationIR{OperationFamily: "write_values", PreserveOriginal: true}
	result := VerificationResult{
		Pass:            false,
		OperationFamily: "write_values",
		OutputWorkbook:  filepath.Join(t.TempDir(), "missing.xlsx"),
		Reasons:         []string{"execution hash evidence is missing"},
	}

	layered := WithDefaultLayers(ir, result)
	assertLayer(t, layered.Layers, "file_opens", false)
	assertLayer(t, layered.Layers, "semantic_task_specific", false)
}

func assertLayer(t *testing.T, layers []LayerResult, name string, pass bool) {
	t.Helper()
	for _, layer := range layers {
		if layer.Name == name {
			if layer.Pass != pass {
				t.Fatalf("layer %s pass=%v want %v", name, layer.Pass, pass)
			}
			return
		}
	}
	t.Fatalf("layer %s missing in %+v", name, layers)
}

func writeMinimalWorkbook(t *testing.T, path string) {
	t.Helper()
	file := excelize.NewFile()
	defer func() { _ = file.Close() }()
	if err := file.SetCellValue("Sheet1", "A1", "ok"); err != nil {
		t.Fatalf("SetCellValue: %v", err)
	}
	if err := file.SaveAs(path); err != nil {
		t.Fatalf("SaveAs: %v", err)
	}
}
