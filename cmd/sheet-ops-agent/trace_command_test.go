package main

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/xuri/excelize/v2"
)

func TestTraceCommandWritesFormulaTraceJSON(t *testing.T) {
	inputFile := filepath.Join(t.TempDir(), "trace-command.xlsx")
	file := excelize.NewFile()
	defer func() { _ = file.Close() }()

	defaultSheet := file.GetSheetName(0)
	if err := file.SetSheetName(defaultSheet, "Orders"); err != nil {
		t.Fatalf("SetSheetName: %v", err)
	}
	if _, err := file.NewSheet("Summary"); err != nil {
		t.Fatalf("NewSheet: %v", err)
	}
	if err := file.SetCellFormula("Summary", "D5", "=SUM(Orders!D:D)"); err != nil {
		t.Fatalf("SetCellFormula: %v", err)
	}
	if err := file.SaveAs(inputFile); err != nil {
		t.Fatalf("SaveAs: %v", err)
	}

	cmd := newRootCommand()
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{"trace", inputFile, "--cell", "Summary!D5"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute trace: %v", err)
	}

	var document map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &document); err != nil {
		t.Fatalf("Unmarshal trace output: %v\n%s", err, stdout.String())
	}
	if document["cell"] != "Summary!D5" {
		t.Fatalf("cell=%v want Summary!D5", document["cell"])
	}
	if document["status"] != "supported" {
		t.Fatalf("status=%v want supported", document["status"])
	}
	references, ok := document["references"].([]any)
	if !ok || len(references) != 1 || references[0] != "Orders!D:D" {
		t.Fatalf("references=%v want [Orders!D:D]", document["references"])
	}
}
