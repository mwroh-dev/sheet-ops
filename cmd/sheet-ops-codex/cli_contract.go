package main

import (
	"fmt"
	"io"
	"maps"
	"slices"
	"strings"

	"github.com/spf13/cobra"
)

const (
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
				"install-skill",
				"prepare-use",
				"run-validated",
				"run-intent",
			},
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
				"completion",
			},
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
				"help",
			},
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
				"run-validated",
				"run-intent",
			},
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
				"prepare-use",
				"run-validated",
			},
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
				"run-validated",
				"prepare-use",
			},
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
				"prepare-use",
				"run-request",
				"run-intent",
			},
		},
	}
}

func sortedContractNames(contracts map[string]cliCommandContract) []string {
	names := slices.Collect(maps.Keys(contracts))
	slices.Sort(names)
	return names
}

func applyCLIContracts(rootCmd *cobra.Command) {
	rootCmd.InitDefaultHelpCmd()
	rootCmd.InitDefaultCompletionCmd()

	contracts := sheetOpsCLIContracts(rootCmd.Name())
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
	if err := renderContractSection(output, "Agent-contract surface:", contracts, []string{"prepare-use"}); err != nil {
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
