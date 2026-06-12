package main

import (
	"bytes"
	"encoding/json"
	"slices"
	"strings"
	"testing"
)

type agentGuideDocument struct {
	SchemaVersion      string            `json:"schema_version"`
	Command            string            `json:"command"`
	OK                 bool              `json:"ok"`
	HumanWorkbookEntry string            `json:"human_workbook_entry"`
	Principles         []string          `json:"principles"`
	Phases             []agentGuidePhase `json:"phases"`
	BacklogPolicy      string            `json:"backlog_policy"`
}

type agentGuidePhase struct {
	Name             string   `json:"name"`
	Lane             string   `json:"lane"`
	Goal             string   `json:"goal"`
	Commands         []string `json:"commands"`
	ExpectedEvidence []string `json:"expected_evidence"`
	StopConditions   []string `json:"stop_conditions"`
}

func TestAgentGuideJSONReportsOrderedAgentWorkflow(t *testing.T) {
	var doc agentGuideDocument
	executeCLIJSON(t, &doc, "agent-guide", "--json")

	if doc.SchemaVersion != cliContractSchemaVersion {
		t.Fatalf("schema_version = %q, want %q", doc.SchemaVersion, cliContractSchemaVersion)
	}
	if doc.Command != "agent-guide" {
		t.Fatalf("command = %q, want agent-guide", doc.Command)
	}
	if !doc.OK {
		t.Fatalf("ok = false, want true")
	}
	if doc.HumanWorkbookEntry != "sheet-ops" {
		t.Fatalf("human_workbook_entry = %q, want sheet-ops", doc.HumanWorkbookEntry)
	}
	for _, principle := range []string{"single_human_entry", "stdout_json_contract", "preview_is_not_dry_run", "verify_before_success_claim"} {
		if !slices.Contains(doc.Principles, principle) {
			t.Fatalf("principles missing %q: %v", principle, doc.Principles)
		}
	}

	wantPhases := []string{"discover", "preflight", "inspect_impact", "execute", "verify"}
	if len(doc.Phases) != len(wantPhases) {
		t.Fatalf("phase count = %d, want %d: %+v", len(doc.Phases), len(wantPhases), doc.Phases)
	}
	for idx, want := range wantPhases {
		if doc.Phases[idx].Name != want {
			t.Fatalf("phase[%d].name = %q, want %q", idx, doc.Phases[idx].Name, want)
		}
		if len(doc.Phases[idx].Commands) == 0 {
			t.Fatalf("phase[%d] commands is empty", idx)
		}
		if len(doc.Phases[idx].ExpectedEvidence) == 0 {
			t.Fatalf("phase[%d] expected_evidence is empty", idx)
		}
	}
	if !slices.Contains(doc.Phases[0].Commands, "sheet-ops-codex capabilities --json") {
		t.Fatalf("discover commands missing capabilities: %+v", doc.Phases[0].Commands)
	}
	if !slices.Contains(doc.Phases[2].Commands, "sheet-ops-codex preview-request --json --intent-file <path-or-stdin> --input-file <workbook> --output-file <workbook>") {
		t.Fatalf("inspect_impact commands missing preview-request shape: %+v", doc.Phases[2].Commands)
	}
	if doc.BacklogPolicy == "" {
		t.Fatalf("backlog_policy is empty")
	}
}

func TestAgentGuideRequiresJSONAndEmitsErrorEnvelope(t *testing.T) {
	root := newRootCommand()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&stderr)
	root.SetArgs([]string{"agent-guide"})

	err := root.Execute()
	if err == nil {
		t.Fatalf("Execute(agent-guide) returned nil, want --json error")
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	var envelope cliErrorEnvelope
	if decodeErr := json.Unmarshal(stdout.Bytes(), &envelope); decodeErr != nil {
		t.Fatalf("json.Unmarshal error envelope: %v\nstdout:\n%s", decodeErr, stdout.String())
	}
	if envelope.OK {
		t.Fatalf("envelope.ok = true, want false")
	}
	if envelope.Error.Code != cliErrorInvalidUsage {
		t.Fatalf("error.code = %q, want %q", envelope.Error.Code, cliErrorInvalidUsage)
	}
	if !strings.Contains(envelope.Error.Message, "agent-guide requires --json") {
		t.Fatalf("error.message = %q", envelope.Error.Message)
	}
}
