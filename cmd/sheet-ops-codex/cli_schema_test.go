package main

import (
	"bytes"
	"encoding/json"
	"slices"
	"strings"
	"testing"
)

type cliCapabilitiesDocument struct {
	SchemaVersion        string                      `json:"schema_version"`
	CLIName              string                      `json:"cli_name"`
	HumanWorkbookEntry   string                      `json:"human_workbook_entry"`
	MachineEntryCommands []string                    `json:"machine_entry_commands"`
	ReadOnlyCommands     []string                    `json:"read_only_commands"`
	CommandGroups        []cliCapabilityGroup        `json:"command_groups"`
	Commands             []cliCapabilityCommandBrief `json:"commands"`
}

type cliCapabilityGroup struct {
	Name          string   `json:"name"`
	Description   string   `json:"description"`
	SafeDiscovery bool     `json:"safe_discovery"`
	Commands      []string `json:"commands"`
}

type cliCapabilityCommandBrief struct {
	Name           string `json:"name"`
	Classification string `json:"classification"`
	Mutating       bool   `json:"mutating"`
	ReadOnly       bool   `json:"read_only"`
	Hidden         bool   `json:"hidden"`
}

type cliSchemaDocument struct {
	SchemaVersion string                 `json:"schema_version"`
	CLIName       string                 `json:"cli_name"`
	Commands      []cliCommandSchemaView `json:"commands"`
}

type cliCommandSchemaEnvelope struct {
	SchemaVersion string               `json:"schema_version"`
	CLIName       string               `json:"cli_name"`
	Command       cliCommandSchemaView `json:"command"`
}

type cliCommandSchemaView struct {
	Name             string                `json:"name"`
	Classification   string                `json:"classification"`
	IntendedCaller   string                `json:"intended_caller"`
	Description      string                `json:"description"`
	Usage            string                `json:"usage"`
	Options          []cliSchemaOptionView `json:"options"`
	OutputMode       string                `json:"output_mode"`
	SideEffects      []string              `json:"side_effects"`
	ReadArtifacts    []string              `json:"read_artifacts"`
	WrittenArtifacts []string              `json:"written_artifacts"`
	StateBehavior    string                `json:"state_behavior"`
	SafetyNotes      []string              `json:"safety_notes"`
	RelatedCommands  []string              `json:"related_commands"`
	Mutating         bool                  `json:"mutating"`
	ReadOnly         bool                  `json:"read_only"`
	Hidden           bool                  `json:"hidden"`
}

type cliSchemaOptionView struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func TestCapabilitiesJSONReportsSafeEntryBoundaries(t *testing.T) {
	t.Setenv(stateRootEnv, "/tmp/not-the-workbook-state-root")

	var doc cliCapabilitiesDocument
	stdout := executeCLIJSON(t, &doc, "capabilities", "--json")
	if strings.Contains(stdout, "\"schemaVersion\"") {
		t.Fatalf("capabilities JSON used schemaVersion key, want schema_version:\n%s", stdout)
	}

	if doc.SchemaVersion == "" {
		t.Fatalf("schema_version is empty")
	}
	if doc.CLIName != "sheet-ops-codex" {
		t.Fatalf("cli_name = %q, want sheet-ops-codex", doc.CLIName)
	}
	if doc.HumanWorkbookEntry != "sheet-ops" {
		t.Fatalf("human_workbook_entry = %q, want sheet-ops", doc.HumanWorkbookEntry)
	}
	if !slices.Equal(doc.MachineEntryCommands, []string{"capabilities", "schema"}) {
		t.Fatalf("machine_entry_commands = %v, want [capabilities schema]", doc.MachineEntryCommands)
	}
	for _, name := range []string{"capabilities", "schema", "help", "completion"} {
		if !slices.Contains(doc.ReadOnlyCommands, name) {
			t.Fatalf("read_only_commands missing %q: %v", name, doc.ReadOnlyCommands)
		}
	}

	agentGroup := findCapabilityGroup(t, doc.CommandGroups, "agent_contract")
	if !agentGroup.SafeDiscovery {
		t.Fatalf("agent_contract group safe_discovery = false")
	}
	for _, name := range []string{"prepare-use", "capabilities", "schema"} {
		if !slices.Contains(agentGroup.Commands, name) {
			t.Fatalf("agent_contract commands missing %q: %v", name, agentGroup.Commands)
		}
	}

	runValidated := findCapabilityCommand(t, doc.Commands, "run-validated")
	if runValidated.Classification != "internal_handoff" {
		t.Fatalf("run-validated classification = %q, want internal_handoff", runValidated.Classification)
	}
	if runValidated.ReadOnly {
		t.Fatalf("run-validated read_only = true, want false")
	}
	if !runValidated.Mutating {
		t.Fatalf("run-validated mutating = false, want true")
	}
}

func TestSchemaJSONReportsCommandContracts(t *testing.T) {
	var doc cliSchemaDocument
	stdout := executeCLIJSON(t, &doc, "schema", "--json")
	if strings.Contains(stdout, "\"schemaVersion\"") {
		t.Fatalf("schema JSON used schemaVersion key, want schema_version:\n%s", stdout)
	}

	if doc.SchemaVersion == "" {
		t.Fatalf("schema_version is empty")
	}
	if doc.CLIName != "sheet-ops-codex" {
		t.Fatalf("cli_name = %q, want sheet-ops-codex", doc.CLIName)
	}

	prepareUse := findCommandSchema(t, doc.Commands, "prepare-use")
	if prepareUse.Classification != "agent_contract" {
		t.Fatalf("prepare-use classification = %q, want agent_contract", prepareUse.Classification)
	}

	runRequest := findCommandSchema(t, doc.Commands, "run-request")
	if !runRequest.Hidden {
		t.Fatalf("run-request hidden = false, want true")
	}

	installSkill := findCommandSchema(t, doc.Commands, "install-skill")
	for _, option := range installSkill.Options {
		if strings.Contains(option.Name, "--sheet-ops-codex-bin") {
			t.Fatalf("install-skill schema exposed hidden flag --sheet-ops-codex-bin: %+v", option)
		}
	}

	for _, forbidden := range []string{"bash", "fish", "powershell", "zsh"} {
		for _, command := range doc.Commands {
			if command.Name == forbidden {
				t.Fatalf("schema unexpectedly exposed generated completion subcommand %q", forbidden)
			}
		}
	}
}

func TestSchemaCommandJSONReportsPrepareUseContract(t *testing.T) {
	var doc cliCommandSchemaEnvelope
	executeCLIJSON(t, &doc, "schema", "command", "prepare-use", "--json")

	if doc.SchemaVersion == "" {
		t.Fatalf("schema_version is empty")
	}
	if doc.Command.Name != "prepare-use" {
		t.Fatalf("command.name = %q, want prepare-use", doc.Command.Name)
	}
	if doc.Command.Classification != "agent_contract" {
		t.Fatalf("classification = %q, want agent_contract", doc.Command.Classification)
	}
	assertSchemaCommandFields(t, doc.Command)
	if doc.Command.Mutating != true {
		t.Fatalf("prepare-use mutating = false, want true")
	}
	if doc.Command.ReadOnly {
		t.Fatalf("prepare-use read_only = true, want false")
	}
}

func TestSchemaCommandJSONReportsInternalCommandContract(t *testing.T) {
	var doc cliCommandSchemaEnvelope
	executeCLIJSON(t, &doc, "schema", "command", "run-validated", "--json")

	if doc.Command.Name != "run-validated" {
		t.Fatalf("command.name = %q, want run-validated", doc.Command.Name)
	}
	if doc.Command.Classification != "internal_handoff" {
		t.Fatalf("classification = %q, want internal_handoff", doc.Command.Classification)
	}
	assertSchemaCommandFields(t, doc.Command)
	if doc.Command.ReadOnly {
		t.Fatalf("run-validated read_only = true, want false")
	}
	if !doc.Command.Mutating {
		t.Fatalf("run-validated mutating = false, want true")
	}
}

func executeCLIJSON(t *testing.T, target any, args ...string) string {
	t.Helper()

	root := newRootCommand()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&stderr)
	root.SetArgs(args)

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute(%v): %v\nstderr:\n%s\nstdout:\n%s", args, err, stderr.String(), stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("Execute(%v) wrote to stderr:\n%s", args, stderr.String())
	}
	if err := json.Unmarshal(stdout.Bytes(), target); err != nil {
		t.Fatalf("json.Unmarshal(%v): %v\nstdout:\n%s", args, err, stdout.String())
	}
	return stdout.String()
}

func findCapabilityGroup(t *testing.T, groups []cliCapabilityGroup, name string) cliCapabilityGroup {
	t.Helper()

	for _, group := range groups {
		if group.Name == name {
			return group
		}
	}
	t.Fatalf("missing capability group %q", name)
	return cliCapabilityGroup{}
}

func findCapabilityCommand(t *testing.T, commands []cliCapabilityCommandBrief, name string) cliCapabilityCommandBrief {
	t.Helper()

	for _, command := range commands {
		if command.Name == name {
			return command
		}
	}
	t.Fatalf("missing capabilities command %q", name)
	return cliCapabilityCommandBrief{}
}

func findCommandSchema(t *testing.T, commands []cliCommandSchemaView, name string) cliCommandSchemaView {
	t.Helper()

	for _, command := range commands {
		if command.Name == name {
			return command
		}
	}
	t.Fatalf("missing command schema for %q", name)
	return cliCommandSchemaView{}
}

func assertSchemaCommandFields(t *testing.T, command cliCommandSchemaView) {
	t.Helper()

	for fieldName, value := range map[string]string{
		"name":            command.Name,
		"classification":  command.Classification,
		"intended_caller": command.IntendedCaller,
		"description":     command.Description,
		"usage":           command.Usage,
		"output_mode":     command.OutputMode,
		"state_behavior":  command.StateBehavior,
	} {
		if strings.TrimSpace(value) == "" {
			t.Fatalf("%s is empty", fieldName)
		}
	}

	if len(command.Options) == 0 {
		t.Fatalf("options is empty")
	}
	if len(command.SideEffects) == 0 {
		t.Fatalf("side_effects is empty")
	}
	if len(command.ReadArtifacts) == 0 {
		t.Fatalf("read_artifacts is empty")
	}
	if len(command.WrittenArtifacts) == 0 {
		t.Fatalf("written_artifacts is empty")
	}
	if len(command.SafetyNotes) == 0 {
		t.Fatalf("safety_notes is empty")
	}
	if len(command.RelatedCommands) == 0 {
		t.Fatalf("related_commands is empty")
	}
}
