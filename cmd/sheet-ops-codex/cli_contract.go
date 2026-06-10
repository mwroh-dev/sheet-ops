package main

import (
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"slices"
	"strings"

	"github.com/spf13/cobra"
)

const (
	cliContractSchemaVersion              = "sheet-ops-cli/v1"
	cliClassificationAgentContract        = "agent_contract"
	cliClassificationBuiltinSupport       = "builtin_support"
	cliClassificationCompatibilityHidden  = "compatibility_internal"
	cliClassificationInternalHandoff      = "internal_handoff"
	cliClassificationMaintainerDiagnostic = "maintainer_diagnostic"
	cliClassificationPublicInstall        = "public_install"
)

type cliCommandContract struct {
	Name             string
	Classification   string
	IntendedCaller   string
	ShortDescription string
	Usage            string
	Options          []string
	OutputMode       string
	SideEffects      []string
	ReadArtifacts    []string
	WrittenArtifacts []string
	StateBehavior    string
	SafetyNotes      []string
	RelatedCommands  []string
	Mutating         bool
	ReadOnly         bool
	DryRunCapable    bool
	Hidden           bool
}

func sheetOpsCLIContracts(rootName string) map[string]cliCommandContract {
	return map[string]cliCommandContract{
		rootName: {
			Name:             rootName,
			Classification:   cliClassificationBuiltinSupport,
			IntendedCaller:   "Human operators, automation, or agents discovering available sheet-ops-codex surfaces.",
			ShortDescription: "Thin public CLI adapter for Sheet Ops use requests",
			Usage:            rootName + " [command]",
			Options: []string{
				"-h, --help: Show top-level help.",
			},
			OutputMode: "Human-readable help text on stdout.",
			SideEffects: []string{
				"None. Dispatches to a selected subcommand only.",
			},
			ReadArtifacts: []string{
				"SHEET_OPS_CLI_NAME environment variable when present.",
			},
			WrittenArtifacts: []string{
				"None.",
			},
			StateBehavior: "Does not mutate workbooks or .sheet-ops-state directly; behavior depends on the chosen subcommand.",
			SafetyNotes: []string{
				"Use the `sheet-ops` skill, not this CLI root, as the single human-facing workbook request entry.",
			},
			RelatedCommands: []string{
				"capabilities",
				"install-skill",
				"preflight",
				"preview-request",
				"prepare-use",
				"run-validated",
				"run-intent",
				"schema",
			},
			Mutating: false,
			ReadOnly: true,
		},
		"completion": {
			Name:             "completion",
			Classification:   cliClassificationBuiltinSupport,
			IntendedCaller:   "Shell integrators or automation setting up command-line completion.",
			ShortDescription: "Generate the autocompletion script for the specified shell",
			Usage:            rootName + " completion [bash|zsh|fish|powershell]",
			Options: []string{
				"[shell]: Target shell for the generated completion script.",
			},
			OutputMode: "Shell completion script on stdout.",
			SideEffects: []string{
				"None. The caller decides whether to save the generated script.",
			},
			ReadArtifacts: []string{
				"Registered Cobra command tree.",
			},
			WrittenArtifacts: []string{
				"None.",
			},
			StateBehavior: "Read-only. Does not touch workbook files or .sheet-ops-state.",
			SafetyNotes: []string{
				"Support surface only; it is not part of the workbook request path.",
			},
			RelatedCommands: []string{
				"help",
				rootName,
			},
			Mutating: false,
			ReadOnly: true,
		},
		"help": {
			Name:             "help",
			Classification:   cliClassificationBuiltinSupport,
			IntendedCaller:   "Human operators or automation inspecting the CLI contract.",
			ShortDescription: "Help about any command",
			Usage:            rootName + " help [command]",
			Options: []string{
				"[command]: Optional command name to describe.",
			},
			OutputMode: "Human-readable help text on stdout.",
			SideEffects: []string{
				"None.",
			},
			ReadArtifacts: []string{
				"Registered Cobra command tree and Sheet Ops CLI contract metadata.",
			},
			WrittenArtifacts: []string{
				"None.",
			},
			StateBehavior: "Read-only. Does not touch workbook files or .sheet-ops-state.",
			SafetyNotes: []string{
				"Help describes supported surfaces without teaching the internal bundled launcher as a public workbook entry.",
			},
			RelatedCommands: []string{
				rootName,
				"capabilities",
				"completion",
				"schema",
			},
			Mutating: false,
			ReadOnly: true,
		},
		"install-skill": {
			Name:             "install-skill",
			Classification:   cliClassificationPublicInstall,
			IntendedCaller:   "Human operators or automation installing the project-local Sheet Ops skill bundle.",
			ShortDescription: "Install the Sheet Ops Codex skill into a project",
			Usage:            rootName + " install-skill --project /absolute/path/to/workspace [--go-bin /absolute/path/to/go] [--global] [--yes]",
			Options: []string{
				"--project: Project directory for project-local install.",
				"--global: Install under $HOME/.codex instead of the current project.",
				"--go-bin: Go binary used for version checks and bundled builds.",
				"-y, --yes: Accept Go install or upgrade prompts.",
			},
			OutputMode: "Human-readable install status on stdout; prompts when Go is missing or outdated.",
			SideEffects: []string{
				"Writes the installed sheet-ops skill tree under .codex/skills/sheet-ops or $HOME/.codex/skills/sheet-ops.",
				"Builds bundled entrypoints inside the installed skill package.",
			},
			ReadArtifacts: []string{
				"Canonical package files listed by cmd/sheet-ops-codex/install_manifest.txt.",
				"Go executable and go version output.",
			},
			WrittenArtifacts: []string{
				"<project>/.codex/skills/sheet-ops/** or $HOME/.codex/skills/sheet-ops/**.",
			},
			StateBehavior: "Does not use workbook-case state roots; installation writes only to Codex skill directories.",
			SafetyNotes: []string{
				"Install surface only. It does not execute workbook requests.",
				"Keep the repository-local install flow as the canonical install path.",
			},
			RelatedCommands: []string{
				"prepare-use",
				"capabilities",
				"schema",
				"help",
			},
			Mutating: true,
		},
		"prepare-use": {
			Name:             "prepare-use",
			Classification:   cliClassificationAgentContract,
			IntendedCaller:   "The installed sheet-ops skill wrapper, CI, or another agent preparing a typed handoff.",
			ShortDescription: "Deterministically write a use envelope for the internal agent-system boundary",
			Usage:            rootName + " prepare-use --scenario-id <id> --request-ref <path> --request-kind <kind> --input-file <path> --output-file <path> --envelope-file <path>",
			Options: []string{
				"--scenario-id: Scenario identifier for the use envelope.",
				"--request-ref: Path to the request reference file.",
				"--request-kind: prompt_text, structured_use_request, or organism_execution_request.",
				"--input-file: Path to the input workbook.",
				"--output-file: Path to the output workbook.",
				"--envelope-file: Path to write the use-envelope JSON file.",
			},
			OutputMode: "Indented JSON result on stdout describing the prepared envelope.",
			SideEffects: []string{
				"Writes the requested use-envelope JSON file.",
				"Creates parent directories for the envelope file when needed.",
			},
			ReadArtifacts: []string{
				"Request reference file path and request-kind compatibility rules.",
				"Input and output workbook paths supplied by the caller.",
			},
			WrittenArtifacts: []string{
				"Use-envelope JSON file at --envelope-file.",
			},
			StateBehavior: "Does not write .sheet-ops-state. It prepares a typed handoff artifact only.",
			SafetyNotes: []string{
				"Agent-contract surface only. It does not execute workbook mutations.",
				"Keep `sheet-ops` as the single human-facing workbook request entry.",
			},
			RelatedCommands: []string{
				"capabilities",
				"preview-request",
				"run-validated",
				"run-intent",
				"schema",
			},
			Mutating: true,
		},
		"run-intent": {
			Name:             "run-intent",
			Classification:   cliClassificationMaintainerDiagnostic,
			IntendedCaller:   "Maintainers or diagnostic harnesses exercising normalized-intent compilation plus execution.",
			ShortDescription: "Validate a normalized intent document and run it",
			Usage:            rootName + " run-intent --intent-file <path> --input-file <path> --output-file <path> [--scenario-id <id>]",
			Options: []string{
				"--intent-file: Path to a normalized intent JSON file.",
				"--input-file: Path to the input workbook.",
				"--output-file: Path to the output workbook.",
				"--scenario-id: Optional scenario identifier override.",
			},
			OutputMode: "Public entry JSON envelope on stdout when compilation stops or execution completes.",
			SideEffects: []string{
				"May validate and persist request-compiler artifacts under the workbook-case state root.",
				"May orchestrate runtime execution and write the requested output workbook.",
			},
			ReadArtifacts: []string{
				"Normalized intent JSON file.",
				"Input workbook and workbook-case runtime configuration.",
			},
			WrittenArtifacts: []string{
				"Output workbook when execution proceeds.",
				"Request-compiler and runtime artifacts under <workspace>/.sheet-ops-state/artifacts/.",
			},
			StateBehavior: "Enforces workbook-case .sheet-ops-state ownership through SHEET_OPS_STATE_ROOT before compiler or runtime work proceeds.",
			SafetyNotes: []string{
				"Maintainer-only diagnostic surface. It is not the primary human-facing workbook entry.",
				"Blocked or needs_human_checkpoint outcomes require artifact inspection before retry.",
			},
			RelatedCommands: []string{
				"capabilities",
				"preview-request",
				"prepare-use",
				"run-validated",
				"schema",
			},
			Mutating: true,
		},
		"run-request": {
			Name:             "run-request",
			Classification:   cliClassificationCompatibilityHidden,
			IntendedCaller:   "Legacy compatibility wrappers that still emit the older run-request form.",
			ShortDescription: "Compatibility alias for run-validated",
			Usage:            rootName + " run-request --file <path>",
			Options: []string{
				"--file: Path to a validated execution request JSON file.",
			},
			OutputMode: "Runtime orchestration result JSON on stdout.",
			SideEffects: []string{
				"May execute workbook mutations described by the validated request.",
				"May write workbook-case runtime artifacts under .sheet-ops-state.",
			},
			ReadArtifacts: []string{
				"Validated execution request JSON file.",
				"Referenced input workbook and runtime assets.",
			},
			WrittenArtifacts: []string{
				"Output workbook requested by the validated execution request.",
				"Runtime artifacts under <workspace>/.sheet-ops-state/artifacts/.",
			},
			StateBehavior: "Uses the same workbook-case state-root guard and runtime defaults as run-validated.",
			SafetyNotes: []string{
				"Hidden compatibility surface. Do not teach it as a public workbook entry.",
			},
			RelatedCommands: []string{
				"capabilities",
				"preview-request",
				"run-validated",
				"prepare-use",
				"schema",
			},
			Mutating: true,
			Hidden:   true,
		},
		"run-validated": {
			Name:             "run-validated",
			Classification:   cliClassificationInternalHandoff,
			IntendedCaller:   "The installed sheet-ops skill handoff or an internal compatibility dispatcher with a validated request.",
			ShortDescription: "Run a validated execution request",
			Usage:            rootName + " run-validated --request <path>",
			Options: []string{
				"--request: Path to a validated execution request JSON file.",
			},
			OutputMode: "Runtime orchestration result JSON on stdout.",
			SideEffects: []string{
				"May execute workbook mutations described by the validated request.",
				"May write workbook-case runtime artifacts under .sheet-ops-state.",
			},
			ReadArtifacts: []string{
				"Validated execution request JSON file.",
				"Referenced input workbook and runtime assets.",
			},
			WrittenArtifacts: []string{
				"Output workbook requested by the validated execution request.",
				"Runtime artifacts under <workspace>/.sheet-ops-state/artifacts/.",
			},
			StateBehavior: "Enforces workbook-case .sheet-ops-state ownership through SHEET_OPS_STATE_ROOT and configures safe runtime defaults before orchestration.",
			SafetyNotes: []string{
				"Internal handoff surface only. Do not present it as a second human-facing workbook request entry.",
			},
			RelatedCommands: []string{
				"capabilities",
				"prepare-use",
				"preview-request",
				"run-request",
				"run-intent",
				"schema",
			},
			Mutating: true,
		},
	}
}

func sheetOpsCLIReadOnlyDiscoveryContracts(rootName string) map[string]cliCommandContract {
	return map[string]cliCommandContract{
		"capabilities": {
			Name:             "capabilities",
			Classification:   cliClassificationAgentContract,
			IntendedCaller:   "Automation, CI, or another agent discovering safe Sheet Ops CLI entry boundaries.",
			ShortDescription: "Emit machine-readable command groups and safe CLI entry boundaries",
			Usage:            rootName + " capabilities --json",
			Options: []string{
				"--json: Emit machine-readable JSON on stdout.",
			},
			OutputMode: "Machine-readable JSON on stdout.",
			SideEffects: []string{
				"None.",
			},
			ReadArtifacts: []string{
				"Registered Sheet Ops CLI command-contract metadata.",
			},
			WrittenArtifacts: []string{
				"None.",
			},
			StateBehavior: "Read-only. Does not invoke workbook runtime paths or touch .sheet-ops-state.",
			SafetyNotes: []string{
				"Discovery surface only. It exists so another agent can find safe entry boundaries without parsing help prose.",
			},
			RelatedCommands: []string{
				"schema",
				"preflight",
				"preview-request",
				"prepare-use",
				"help",
			},
			Mutating: false,
			ReadOnly: true,
		},
		"schema": {
			Name:             "schema",
			Classification:   cliClassificationAgentContract,
			IntendedCaller:   "Automation, CI, or another agent inspecting the Sheet Ops CLI contract in detail.",
			ShortDescription: "Emit machine-readable command schema for the CLI surface",
			Usage:            rootName + " schema [command <name>] --json",
			Options: []string{
				"--json: Emit machine-readable JSON on stdout.",
				"command <name>: Optional nested selector for a specific command contract.",
			},
			OutputMode: "Machine-readable JSON on stdout.",
			SideEffects: []string{
				"None.",
			},
			ReadArtifacts: []string{
				"Registered Sheet Ops CLI command-contract metadata.",
			},
			WrittenArtifacts: []string{
				"None.",
			},
			StateBehavior: "Read-only. Does not invoke workbook runtime paths or touch .sheet-ops-state.",
			SafetyNotes: []string{
				"Meta command only. It describes install, handoff, and diagnostic surfaces without executing them.",
				"Cobra-generated completion leaf commands and the hidden --sheet-ops-codex-bin install flag are intentionally excluded from the machine-readable schema because they are generated or internal plumbing rather than stable Sheet Ops command-contract surfaces.",
			},
			RelatedCommands: []string{
				"capabilities",
				"preflight",
				"preview-request",
				"prepare-use",
				"help",
			},
			Mutating: false,
			ReadOnly: true,
		},
		"preflight": {
			Name:             "preflight",
			Classification:   cliClassificationAgentContract,
			IntendedCaller:   "Automation, CI, or another agent checking whether Sheet Ops CLI surfaces are ready before mutation.",
			ShortDescription: "Run read-only readiness checks for Sheet Ops CLI execution",
			Usage:            rootName + " preflight --json [--project <path>] [--input-file <path>] [--output-file <path>] [--go-bin <path>]",
			Options: []string{
				"--json: Emit machine-readable JSON on stdout.",
				"--project: Project directory whose local Sheet Ops install/readiness should be checked.",
				"--input-file: Optional workbook path used to check readability and state-root ownership.",
				"--output-file: Optional output workbook path used to check nearest parent ancestor accessibility.",
				"--go-bin: Go binary used for readiness checks.",
			},
			OutputMode: "Machine-readable readiness JSON on stdout.",
			SideEffects: []string{
				"None. Does not create workbooks, envelopes, install trees, or .sheet-ops-state directories.",
			},
			ReadArtifacts: []string{
				"Project directory metadata.",
				"Go executable and go version output.",
				"Embedded package manifest metadata.",
				"Optional input workbook metadata.",
				"Optional output parent directory metadata.",
				"SHEET_OPS_STATE_ROOT when set.",
			},
			WrittenArtifacts: []string{
				"None.",
			},
			StateBehavior: "Read-only. Checks whether SHEET_OPS_STATE_ROOT matches the workbook-case .sheet-ops-state when an input workbook is provided, but does not create or mutate state roots.",
			SafetyNotes: []string{
				"Preflight is a diagnostic surface only; it is not a dry-run for workbook mutations.",
				"Mutating commands remain non-dry-run-capable until a truthful runtime planning mode exists.",
			},
			RelatedCommands: []string{
				"capabilities",
				"schema",
				"preview-request",
				"prepare-use",
				"run-validated",
				"install-skill",
			},
			Mutating:      false,
			ReadOnly:      true,
			DryRunCapable: false,
		},
		"preview-request": {
			Name:             "preview-request",
			Classification:   cliClassificationAgentContract,
			IntendedCaller:   "Automation, CI, or another agent inspecting planned workbook impact before choosing a mutating execution path.",
			ShortDescription: "Inspect normalized-intent impact without runtime execution",
			Usage:            rootName + " preview-request --json --intent-file <path> --input-file <path> --output-file <path> [--scenario-id <id>]",
			Options: []string{
				"--json: Emit machine-readable JSON on stdout.",
				"--intent-file: Path to a normalized intent JSON file.",
				"--input-file: Path to the input workbook.",
				"--output-file: Planned output workbook path.",
				"--scenario-id: Optional scenario identifier override.",
			},
			OutputMode: "Machine-readable preview JSON on stdout.",
			SideEffects: []string{
				"None. Does not create workbooks, request-compiler artifacts, runtime artifacts, or .sheet-ops-state directories.",
			},
			ReadArtifacts: []string{
				"Normalized intent JSON file.",
				"Input workbook metadata.",
				"SHEET_OPS_STATE_ROOT when set.",
			},
			WrittenArtifacts: []string{
				"None. Output workbook and state artifacts are reported as planned impact only.",
			},
			StateBehavior: "Read-only. Computes the workbook-case .sheet-ops-state path and validates state-root ownership without creating or mutating it.",
			SafetyNotes: []string{
				"Preview is not a dry-run and does not prove execution success.",
				"Use runtime verification from a later mutating command before claiming workbook success.",
			},
			RelatedCommands: []string{
				"capabilities",
				"schema",
				"preflight",
				"prepare-use",
				"run-intent",
			},
			Mutating:      false,
			ReadOnly:      true,
			DryRunCapable: false,
		},
	}
}

func sheetOpsCLIAllContracts(rootName string) map[string]cliCommandContract {
	contracts := sheetOpsCLIContracts(rootName)
	for name, contract := range sheetOpsCLIReadOnlyDiscoveryContracts(rootName) {
		contracts[name] = contract
	}
	return contracts
}

func sortedContractNames(contracts map[string]cliCommandContract) []string {
	names := slices.Collect(maps.Keys(contracts))
	slices.Sort(names)
	return names
}

func applyCLIContracts(rootCmd *cobra.Command) {
	rootCmd.InitDefaultHelpCmd()
	rootCmd.InitDefaultCompletionCmd()

	contracts := sheetOpsCLIAllContracts(rootCmd.Name())
	applyCLIContract(rootCmd, contracts[rootCmd.Name()])
	for _, command := range rootCmd.Commands() {
		if contract, ok := contracts[command.Name()]; ok {
			applyCLIContract(command, contract)
		}
	}

	defaultHelpFunc := rootCmd.HelpFunc()
	rootCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		if cmd == rootCmd {
			if err := renderRootHelp(cmd.OutOrStdout(), rootCmd, contracts); err != nil {
				fmt.Fprintln(cmd.ErrOrStderr(), err)
			}
			return
		}
		defaultHelpFunc(cmd, args)
	})
	rootCmd.Long = strings.Join([]string{
		"Use the `sheet-ops` skill for workbook requests.",
		"This CLI exposes install, agent-contract, internal handoff, and maintainer diagnostic surfaces without making the bundled launcher a second human-facing entry.",
	}, "\n\n")
}

func applyCLIContract(cmd *cobra.Command, contract cliCommandContract) {
	cmd.Short = contract.ShortDescription
	cmd.Annotations = map[string]string{
		"sheet_ops_classification":  contract.Classification,
		"sheet_ops_intended_caller": contract.IntendedCaller,
		"sheet_ops_output_mode":     contract.OutputMode,
	}
}

type cliCapabilitiesPayload struct {
	SchemaVersion        string                      `json:"schema_version"`
	CLIName              string                      `json:"cli_name"`
	HumanWorkbookEntry   string                      `json:"human_workbook_entry"`
	MachineEntryCommands []string                    `json:"machine_entry_commands"`
	RootCommand          cliCapabilityBriefPayload   `json:"root_command"`
	ReadOnlyCommands     []string                    `json:"read_only_commands"`
	CommandGroups        []cliCapabilityGroupPayload `json:"command_groups"`
	Commands             []cliCapabilityBriefPayload `json:"commands"`
	ErrorContract        cliErrorContractPayload     `json:"error_contract"`
}

type cliCapabilityGroupPayload struct {
	Name          string   `json:"name"`
	Description   string   `json:"description"`
	SafeDiscovery bool     `json:"safe_discovery"`
	Commands      []string `json:"commands"`
}

type cliCapabilityBriefPayload struct {
	Name           string `json:"name"`
	Classification string `json:"classification"`
	Mutating       bool   `json:"mutating"`
	ReadOnly       bool   `json:"read_only"`
	Hidden         bool   `json:"hidden"`
}

type cliSchemaPayload struct {
	SchemaVersion string                    `json:"schema_version"`
	CLIName       string                    `json:"cli_name"`
	Commands      []cliCommandSchemaPayload `json:"commands"`
}

type cliCommandSchemaEnvelopePayload struct {
	SchemaVersion string                  `json:"schema_version"`
	CLIName       string                  `json:"cli_name"`
	Command       cliCommandSchemaPayload `json:"command"`
}

type cliCommandSchemaPayload struct {
	Name             string                   `json:"name"`
	Classification   string                   `json:"classification"`
	IntendedCaller   string                   `json:"intended_caller"`
	Description      string                   `json:"description"`
	Usage            string                   `json:"usage"`
	Options          []cliSchemaOptionPayload `json:"options"`
	OutputMode       string                   `json:"output_mode"`
	SideEffects      []string                 `json:"side_effects"`
	ReadArtifacts    []string                 `json:"read_artifacts"`
	WrittenArtifacts []string                 `json:"written_artifacts"`
	StateBehavior    string                   `json:"state_behavior"`
	SafetyNotes      []string                 `json:"safety_notes"`
	RelatedCommands  []string                 `json:"related_commands"`
	Mutating         bool                     `json:"mutating"`
	ReadOnly         bool                     `json:"read_only"`
	DryRunCapable    bool                     `json:"dry_run_capable"`
	Hidden           bool                     `json:"hidden"`
}

type cliSchemaOptionPayload struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func newCapabilitiesCommand() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "capabilities",
		Short: "Emit machine-readable command groups and safe CLI entry boundaries",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !jsonOutput {
				return newCLIError(cliErrorInvalidUsage, "capabilities requires --json", true, cliExitUsage, "capabilities --json")
			}
			return writeContractJSON(cmd.OutOrStdout(), buildCapabilitiesPayload(cmd.Root().Name()))
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Emit machine-readable JSON on stdout")
	return cmd
}

func newSchemaCommand() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "schema",
		Short: "Emit machine-readable command schema for the CLI surface",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !jsonOutput {
				return newCLIError(cliErrorInvalidUsage, "schema requires --json", true, cliExitUsage, "schema --json")
			}
			return writeContractJSON(cmd.OutOrStdout(), buildSchemaPayload(cmd.Root().Name()))
		},
	}

	cmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "Emit machine-readable JSON on stdout")
	cmd.AddCommand(&cobra.Command{
		Use:   "command <name>",
		Short: "Emit machine-readable schema for one command",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !jsonOutput {
				return newCLIError(cliErrorInvalidUsage, "schema command requires --json", true, cliExitUsage, "schema command <name> --json")
			}
			contracts := sheetOpsCLIAllContracts(cmd.Root().Name())
			contract, ok := contracts[args[0]]
			if !ok {
				return writeCLIErrorJSONAndReturn(
					cmd.OutOrStdout(),
					newCLIError(cliErrorUnknownCommand, fmt.Sprintf("unknown command %q", args[0]), true, cliExitUsage, "schema --json", "capabilities --json"),
				)
			}
			return writeContractJSON(cmd.OutOrStdout(), cliCommandSchemaEnvelopePayload{
				SchemaVersion: cliContractSchemaVersion,
				CLIName:       cmd.Root().Name(),
				Command:       buildCommandSchemaPayload(contract),
			})
		},
	})
	return cmd
}

func buildCapabilitiesPayload(rootName string) cliCapabilitiesPayload {
	contracts := sheetOpsCLIAllContracts(rootName)
	groupNames := []string{
		cliClassificationPublicInstall,
		cliClassificationAgentContract,
		cliClassificationInternalHandoff,
		cliClassificationMaintainerDiagnostic,
		cliClassificationBuiltinSupport,
		cliClassificationCompatibilityHidden,
	}

	commandGroups := make([]cliCapabilityGroupPayload, 0, len(groupNames))
	for _, classification := range groupNames {
		names := make([]string, 0, len(contracts))
		for _, name := range sortedContractNames(contracts) {
			if name == rootName {
				continue
			}
			contract := contracts[name]
			if contract.Classification == classification {
				names = append(names, contract.Name)
			}
		}
		if len(names) == 0 {
			continue
		}
		commandGroups = append(commandGroups, cliCapabilityGroupPayload{
			Name:          classification,
			Description:   capabilityGroupDescription(classification),
			SafeDiscovery: true,
			Commands:      names,
		})
	}

	readOnlyCommands := make([]string, 0, len(contracts))
	commands := make([]cliCapabilityBriefPayload, 0, len(contracts))
	for _, name := range sortedContractNames(contracts) {
		contract := contracts[name]
		if name == rootName {
			continue
		}
		if contract.ReadOnly {
			readOnlyCommands = append(readOnlyCommands, contract.Name)
		}
		commands = append(commands, cliCapabilityBriefPayload{
			Name:           contract.Name,
			Classification: contract.Classification,
			Mutating:       contract.Mutating,
			ReadOnly:       contract.ReadOnly,
			Hidden:         contract.Hidden,
		})
	}

	return cliCapabilitiesPayload{
		SchemaVersion:        cliContractSchemaVersion,
		CLIName:              rootName,
		HumanWorkbookEntry:   "sheet-ops",
		MachineEntryCommands: []string{"capabilities", "schema"},
		RootCommand:          buildCapabilityBriefPayload(contracts[rootName]),
		ReadOnlyCommands:     readOnlyCommands,
		CommandGroups:        commandGroups,
		Commands:             commands,
		ErrorContract:        buildCLIErrorContractPayload(),
	}
}

func buildCapabilityBriefPayload(contract cliCommandContract) cliCapabilityBriefPayload {
	return cliCapabilityBriefPayload{
		Name:           contract.Name,
		Classification: contract.Classification,
		Mutating:       contract.Mutating,
		ReadOnly:       contract.ReadOnly,
		Hidden:         contract.Hidden,
	}
}

func buildSchemaPayload(rootName string) cliSchemaPayload {
	contracts := sheetOpsCLIAllContracts(rootName)
	commands := make([]cliCommandSchemaPayload, 0, len(contracts))
	for _, name := range sortedContractNames(contracts) {
		commands = append(commands, buildCommandSchemaPayload(contracts[name]))
	}
	return cliSchemaPayload{
		SchemaVersion: cliContractSchemaVersion,
		CLIName:       rootName,
		Commands:      commands,
	}
}

func buildCommandSchemaPayload(contract cliCommandContract) cliCommandSchemaPayload {
	return cliCommandSchemaPayload{
		Name:             contract.Name,
		Classification:   contract.Classification,
		IntendedCaller:   contract.IntendedCaller,
		Description:      contract.ShortDescription,
		Usage:            contract.Usage,
		Options:          buildOptionPayloads(contract.Options),
		OutputMode:       contract.OutputMode,
		SideEffects:      append([]string(nil), contract.SideEffects...),
		ReadArtifacts:    append([]string(nil), contract.ReadArtifacts...),
		WrittenArtifacts: append([]string(nil), contract.WrittenArtifacts...),
		StateBehavior:    contract.StateBehavior,
		SafetyNotes:      append([]string(nil), contract.SafetyNotes...),
		RelatedCommands:  append([]string(nil), contract.RelatedCommands...),
		Mutating:         contract.Mutating,
		ReadOnly:         contract.ReadOnly,
		DryRunCapable:    contract.DryRunCapable,
		Hidden:           contract.Hidden,
	}
}

func buildOptionPayloads(options []string) []cliSchemaOptionPayload {
	payloads := make([]cliSchemaOptionPayload, 0, len(options))
	for _, option := range options {
		name, description, hasDescription := strings.Cut(option, ":")
		payload := cliSchemaOptionPayload{Name: strings.TrimSpace(name)}
		if hasDescription {
			payload.Description = strings.TrimSpace(description)
		}
		payloads = append(payloads, payload)
	}
	return payloads
}

func capabilityGroupDescription(classification string) string {
	switch classification {
	case cliClassificationPublicInstall:
		return "Installation surface for materializing the project-local Sheet Ops skill package."
	case cliClassificationAgentContract:
		return "Agent-facing discovery and typed handoff surfaces that do not create a second human workbook entry."
	case cliClassificationInternalHandoff:
		return "Internal validated-request execution surfaces owned by the installed sheet-ops skill handoff."
	case cliClassificationMaintainerDiagnostic:
		return "Maintainer-only diagnostic surfaces that may compile or execute workbook intents."
	case cliClassificationBuiltinSupport:
		return "Read-only help and completion support surfaces."
	case cliClassificationCompatibilityHidden:
		return "Hidden compatibility surfaces kept for internal or legacy callers."
	default:
		return "Sheet Ops CLI surface."
	}
}

func writeContractJSON(output io.Writer, value any) error {
	encoder := json.NewEncoder(output)
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}

func renderRootHelp(output io.Writer, rootCmd *cobra.Command, contracts map[string]cliCommandContract) error {
	rootContract := contracts[rootCmd.Name()]
	if _, err := fmt.Fprintln(output, rootContract.ShortDescription); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(output); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(output, "Use the `sheet-ops` skill for workbook requests."); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(output, "This CLI exposes install, agent-contract, internal handoff, and maintainer diagnostic surfaces."); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(output); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(output, "Usage:\n  %s\n\n", rootContract.Usage); err != nil {
		return err
	}

	if err := renderContractSection(output, "Install surface:", contracts, []string{"install-skill"}); err != nil {
		return err
	}
	if err := renderContractSection(output, "Agent-contract surface:", contracts, []string{"preflight", "preview-request", "prepare-use"}); err != nil {
		return err
	}
	if err := renderContractSection(output, "Internal handoff surface:", contracts, []string{"run-validated"}); err != nil {
		return err
	}
	if err := renderContractSection(output, "Maintainer diagnostic surface:", contracts, []string{"run-intent"}); err != nil {
		return err
	}
	if err := renderContractSection(output, "CLI support surface:", contracts, []string{"help", "completion"}); err != nil {
		return err
	}

	rootCmd.Flags().SetOutput(output)
	if _, err := fmt.Fprintln(output, "Flags:"); err != nil {
		return err
	}
	rootCmd.Flags().PrintDefaults()
	if _, err := fmt.Fprintln(output); err != nil {
		return err
	}
	_, err := fmt.Fprintf(output, "Use \"%s [command] --help\" for more information about a command.\n", rootCmd.Name())
	return err
}

func renderContractSection(output io.Writer, title string, contracts map[string]cliCommandContract, names []string) error {
	if _, err := fmt.Fprintln(output, title); err != nil {
		return err
	}
	for _, name := range names {
		contract := contracts[name]
		if _, err := fmt.Fprintf(output, "  %-14s %s\n", contract.Name, contract.ShortDescription); err != nil {
			return err
		}
	}
	_, err := fmt.Fprintln(output)
	return err
}
