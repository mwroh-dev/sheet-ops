package requestmode

import (
	"os"
	"path/filepath"
	"testing"
)

func TestJudgeRequestRefAcceptsOrganismExecutionRequest(t *testing.T) {
	tempDir := t.TempDir()
	requestPath := filepath.Join(tempDir, "organism-request.json")
	if err := writeTestFile(requestPath, `{
  "scenario_id": "invoice-organism",
  "request_text": "Build an invoice line billing template",
  "input_file": "/tmp/in.xlsx",
  "output_file": "/tmp/out.xlsx",
  "steps": [
    {
      "atom_id": "copy_period_sheet",
      "composition_kind": "period_copy",
      "source_sheet": "Jan",
      "target_sheet": "Feb"
    }
  ]
}`); err != nil {
		t.Fatalf("write request: %v", err)
	}

	judgment, err := JudgeRequestRef(RequestRef{
		Kind: ModeOrganismExecutionRequest,
		Path: requestPath,
	})
	if err != nil {
		t.Fatalf("JudgeRequestRef: %v", err)
	}
	if judgment.RequestMode != ModeOrganismExecutionRequest {
		t.Fatalf("request_mode=%q want %q", judgment.RequestMode, ModeOrganismExecutionRequest)
	}
}

func writeTestFile(path string, value string) error {
	return os.WriteFile(path, []byte(value), 0o644)
}
