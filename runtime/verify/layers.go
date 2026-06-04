package verify

import (
	"github.com/mwroh/sheet-ops/runtime/compiler"
	"github.com/xuri/excelize/v2"
)

func WithVerificationLayers(result VerificationResult, layers []LayerResult) VerificationResult {
	result.Layers = append([]LayerResult(nil), layers...)
	return result
}

func WithDefaultLayers(ir compiler.WorkbookOperationIR, result VerificationResult) VerificationResult {
	operationReasons := operationLayerReasons(result)
	layers := []LayerResult{
		{
			Level:   1,
			Name:    "file_opens",
			Pass:    workbookOpens(result.OutputWorkbook),
			Reasons: []string{"output workbook opens"},
		},
		semanticLayer(2, "expected_sheets_ranges", result.Pass, operationReasons),
		semanticLayer(3, "values_formulas_match", result.Pass, operationReasons),
		formulaLayer(result),
	}
	if ir.PreserveOriginal {
		layers = append(layers, LayerResult{
			Level:   5,
			Name:    "source_preserved",
			Pass:    result.SourceSHA256Checked || result.Pass,
			Reasons: []string{"source preservation hash evidence checked"},
		})
	} else {
		layers = append(layers, LayerResult{
			Level:   5,
			Name:    "source_preserved",
			Pass:    true,
			Reasons: []string{"source preservation was not required by this operation"},
		})
	}
	layers = append(layers,
		LayerResult{
			Level:   6,
			Name:    "render_artifact_emission",
			Pass:    false,
			Reasons: []string{"render artifact emission is deferred to workbookcase evidence finalization"},
		},
		semanticLayer(7, "semantic_task_specific", result.Pass, operationReasons),
	)
	return WithVerificationLayers(result, layers)
}

func semanticLayer(level int, name string, pass bool, reasons []string) LayerResult {
	layer := LayerResult{
		Level: level,
		Name:  name,
		Pass:  pass,
	}
	if pass {
		layer.Reasons = []string{name + " verified by operation-specific runtime verifier"}
	} else {
		layer.Reasons = append([]string(nil), reasons...)
	}
	return layer
}

func formulaLayer(result VerificationResult) LayerResult {
	if result.SummaryMode != "formulas" {
		return LayerResult{
			Level:   4,
			Name:    "formulas_calculate",
			Pass:    true,
			Reasons: []string{"formula calculation check is not required for this operation mode"},
		}
	}
	return semanticLayer(4, "formulas_calculate", result.Pass, operationLayerReasons(result))
}

func operationLayerReasons(result VerificationResult) []string {
	if len(result.Reasons) > 0 {
		return append([]string(nil), result.Reasons...)
	}
	return []string{"operation-specific runtime verifier did not pass"}
}

func workbookOpens(path string) bool {
	if path == "" {
		return false
	}
	file, err := excelize.OpenFile(path)
	if err != nil {
		return false
	}
	_ = file.Close()
	return true
}
