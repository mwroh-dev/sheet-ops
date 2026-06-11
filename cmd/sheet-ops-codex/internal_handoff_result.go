package main

import (
	"encoding/json"
	"io"

	runtimeworkbookcase "github.com/mwroh/sheet-ops/runtime/workbookcase"
)

type InternalHandoffRunResult struct {
	SchemaVersion  string                 `json:"schema_version"`
	OK             bool                   `json:"ok"`
	Command        string                 `json:"command"`
	Status         string                 `json:"status"`
	Classification string                 `json:"classification"`
	Recoverable    bool                   `json:"recoverable"`
	Artifacts      []PublicResultArtifact `json:"artifacts"`
	NextActions    []string               `json:"next_actions"`
	Runtime        PublicRuntimeResult    `json:"runtime"`
}

func newInternalHandoffRunResult(command string, result runtimeworkbookcase.RunResult) InternalHandoffRunResult {
	publicRuntime := newPublicRuntimeResult(result)
	return InternalHandoffRunResult{
		SchemaVersion:  cliContractSchemaVersion,
		OK:             result.Verification.Pass,
		Command:        command,
		Status:         "executed",
		Classification: cliClassificationInternalHandoff,
		Recoverable:    false,
		Artifacts:      publicExecutionArtifacts(publicRuntime),
		NextActions: []string{
			"Inspect runtime.verification before claiming workbook success.",
			"Treat this as an internal handoff result, not a public workbook request entry.",
		},
		Runtime: publicRuntime,
	}
}

func writeInternalHandoffRunResultJSON(output io.Writer, result InternalHandoffRunResult) error {
	encoder := json.NewEncoder(output)
	encoder.SetIndent("", "  ")
	return encoder.Encode(result)
}
