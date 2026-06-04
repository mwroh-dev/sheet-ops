package main

import (
	"encoding/json"
	"io"
)

type UseTerminalResult struct {
	Entry         string `json:"entry"`
	Status        string `json:"status"`
	ScenarioID    string `json:"scenario_id"`
	RunID         string `json:"run_id"`
	RuntimeRunID  string `json:"runtime_run_id,omitempty"`
	SessionID     string `json:"session_id"`
	TerminalState string `json:"terminal_state"`
	Message       string `json:"message"`
	OutputFile    string `json:"output_file"`
	EvidencePath  string `json:"evidence_path"`
}

func newFailedUseTerminalResult(
	scenarioID string,
	runID string,
	sessionID string,
	terminalState string,
	message string,
	evidencePath string,
	runtimeRunID ...string,
) UseTerminalResult {
	result := UseTerminalResult{
		Entry:         "use",
		Status:        "failed",
		ScenarioID:    scenarioID,
		RunID:         runID,
		SessionID:     sessionID,
		TerminalState: terminalState,
		Message:       message,
		OutputFile:    "",
		EvidencePath:  evidencePath,
	}
	if len(runtimeRunID) > 0 {
		result.RuntimeRunID = runtimeRunID[0]
	}
	return result
}

func newSucceededUseTerminalResult(
	scenarioID string,
	runID string,
	sessionID string,
	message string,
	outputFile string,
	evidencePath string,
	runtimeRunID ...string,
) UseTerminalResult {
	result := UseTerminalResult{
		Entry:         "use",
		Status:        "succeeded",
		ScenarioID:    scenarioID,
		RunID:         runID,
		SessionID:     sessionID,
		TerminalState: "TERMINAL_SUCCESS",
		Message:       message,
		OutputFile:    outputFile,
		EvidencePath:  evidencePath,
	}
	if len(runtimeRunID) > 0 {
		result.RuntimeRunID = runtimeRunID[0]
	}
	return result
}

func writeUseTerminalResultJSON(output io.Writer, result UseTerminalResult) error {
	encoder := json.NewEncoder(output)
	encoder.SetIndent("", "  ")
	return encoder.Encode(result)
}
