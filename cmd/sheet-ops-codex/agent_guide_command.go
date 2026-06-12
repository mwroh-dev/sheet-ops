package main

import (
	"encoding/json"
	"io"

	"github.com/spf13/cobra"
)

type agentGuidePayload struct {
	SchemaVersion      string           `json:"schema_version"`
	Command            string           `json:"command"`
	OK                 bool             `json:"ok"`
	HumanWorkbookEntry string           `json:"human_workbook_entry"`
	Principles         []string         `json:"principles"`
	Phases             []agentGuideStep `json:"phases"`
	BacklogPolicy      string           `json:"backlog_policy"`
}

type agentGuideStep struct {
	Name             string   `json:"name"`
	Lane             string   `json:"lane"`
	Goal             string   `json:"goal"`
	Commands         []string `json:"commands"`
	ExpectedEvidence []string `json:"expected_evidence"`
	StopConditions   []string `json:"stop_conditions"`
}

func newAgentGuideCommand() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "agent-guide",
		Short: "Emit the recommended agent workflow for Sheet Ops CLI use",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !jsonOutput {
				return writeCLIErrorJSONAndReturn(
					cmd.OutOrStdout(),
					newCLIError(cliErrorInvalidUsage, "agent-guide requires --json", true, cliExitUsage, "agent-guide --json"),
				)
			}
			return writeAgentGuideJSON(cmd.OutOrStdout(), buildAgentGuidePayload(cmd.Root().Name()))
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Emit machine-readable JSON on stdout")
	return cmd
}

func buildAgentGuidePayload(rootName string) agentGuidePayload {
	return agentGuidePayload{
		SchemaVersion:      cliContractSchemaVersion,
		Command:            "agent-guide",
		OK:                 true,
		HumanWorkbookEntry: "sheet-ops",
		Principles: []string{
			"single_human_entry",
			"stdout_json_contract",
			"preview_is_not_dry_run",
			"verify_before_success_claim",
		},
		Phases: []agentGuideStep{
			{
				Name: "discover",
				Lane: "read_only",
				Goal: "Find stable command contracts before invoking workbook paths.",
				Commands: []string{
					rootName + " capabilities --json",
					rootName + " schema --json",
					rootName + " schema command <name> --json",
				},
				ExpectedEvidence: []string{
					"schema_version is sheet-ops-cli/v1",
					"human_workbook_entry is sheet-ops",
					"target command is present and classified before use",
				},
				StopConditions: []string{
					"unknown_command",
					"missing command contract",
				},
			},
			{
				Name: "preflight",
				Lane: "read_only",
				Goal: "Check local readiness and state-root ownership before planning mutation.",
				Commands: []string{
					rootName + " preflight --json --input-file <workbook> --output-file <workbook>",
				},
				ExpectedEvidence: []string{
					"ok true or recoverable checks with suggested next action",
					"read_only true",
					"no .sheet-ops-state or workbook creation",
				},
				StopConditions: []string{
					"state_root_mismatch",
					"input workbook is unreadable",
				},
			},
			{
				Name: "inspect_impact",
				Lane: "read_only",
				Goal: "Inspect normalized-intent impact without calling runtime execution.",
				Commands: []string{
					rootName + " preview-request --json --intent-file <path-or-stdin> --input-file <workbook> --output-file <workbook>",
				},
				ExpectedEvidence: []string{
					"dry_run false",
					"would_mutate and mutation_summary describe planned impact",
					"fingerprints bind normalized intent and input workbook",
				},
				StopConditions: []string{
					"invalid_json_or_schema",
					"preview limitations do not cover requested claim",
				},
			},
			{
				Name: "execute",
				Lane: "mutating_internal",
				Goal: "Let the installed sheet-ops skill handoff own workbook mutation.",
				Commands: []string{
					"sheet-ops skill request",
					rootName + " run-intent --intent-file <path-or-stdin> --input-file <workbook> --output-file <workbook>",
				},
				ExpectedEvidence: []string{
					"public entry result or internal handoff envelope on stdout",
					"runtime artifacts are reported when execution starts",
				},
				StopConditions: []string{
					"request_checkpoint",
					"validation_blocked",
					"execution_failed",
				},
			},
			{
				Name: "verify",
				Lane: "read_only",
				Goal: "Use runtime evidence before making a workbook success claim.",
				Commands: []string{
					rootName + " evidence-summary --json --evidence-dir <dir>",
				},
				ExpectedEvidence: []string{
					"verification pass/fail is explicit",
					"executed output workbook hash is present for success claims",
					"next_actions identify repair or review follow-up",
				},
				StopConditions: []string{
					"verification_failed",
					"missing runtime evidence",
				},
			},
		},
		BacklogPolicy: "Record broader design gaps separately; fix immediately only when the gap invalidates current evidence or creates regression risk.",
	}
}

func writeAgentGuideJSON(output io.Writer, payload agentGuidePayload) error {
	encoder := json.NewEncoder(output)
	encoder.SetIndent("", "  ")
	return encoder.Encode(payload)
}
