package requestpacker

import (
	"testing"

	requestmode "github.com/mwroh/sheet-ops/runtime/openlayer/requestmode"
)

func TestPackAcceptsOrganismExecutionRequestMode(t *testing.T) {
	envelope, err := Pack(Input{
		ScenarioID:  "invoice-organism",
		RequestPath: "/tmp/organism-request.json",
		InputFile:   "/tmp/in.xlsx",
		OutputFile:  "/tmp/out.xlsx",
		Judgment: requestmode.Judgment{
			RequestMode: requestmode.ModeOrganismExecutionRequest,
			Reason:      "organism request contract validated",
		},
	})
	if err != nil {
		t.Fatalf("Pack: %v", err)
	}
	if envelope.Request.Kind != requestmode.ModeOrganismExecutionRequest {
		t.Fatalf("request.kind=%q want %q", envelope.Request.Kind, requestmode.ModeOrganismExecutionRequest)
	}
}
