package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	requestmode "github.com/mwroh/sheet-ops/runtime/openlayer/requestmode"
)

func TestLoadOrganismExecutionRequestValidatesEnvelopeBindings(t *testing.T) {
	tempDir := t.TempDir()
	requestPath := filepath.Join(tempDir, "organism-request.json")
	inputFile := filepath.Join(tempDir, "in.xlsx")
	outputFile := filepath.Join(tempDir, "out.xlsx")

	request := map[string]any{
		"scenario_id":  "invoice-organism",
		"request_text": "Build an invoice line billing template",
		"input_file":   inputFile,
		"output_file":  outputFile,
		"steps": []any{
			map[string]any{
				"atom_id":          "copy_period_sheet",
				"composition_kind": "period_copy",
				"source_sheet":     "Jan",
				"target_sheet":     "Feb",
			},
		},
	}
	raw, err := json.MarshalIndent(request, "", "  ")
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	if err := writeTestFile(requestPath, string(raw)); err != nil {
		t.Fatalf("write request: %v", err)
	}

	req, err := loadOrganismExecutionRequest(UseEnvelope{
		ScenarioID: "invoice-organism",
		Request: requestmode.RequestRef{
			Kind: requestmode.ModeOrganismExecutionRequest,
			Path: requestPath,
		},
		InputFile:  inputFile,
		OutputFile: outputFile,
	})
	if err != nil {
		t.Fatalf("loadOrganismExecutionRequest: %v", err)
	}
	if req.ScenarioID != "invoice-organism" {
		t.Fatalf("scenario_id=%q want invoice-organism", req.ScenarioID)
	}
	if len(req.Steps) != 1 || req.Steps[0].AtomID != "copy_period_sheet" {
		t.Fatalf("steps=%+v want copy_period_sheet", req.Steps)
	}
}

func writeTestFile(path string, value string) error {
	return os.WriteFile(path, []byte(value), 0o644)
}
