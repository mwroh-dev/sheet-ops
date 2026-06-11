package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

const (
	preflightStatusPass = "pass"
	preflightStatusFail = "fail"
	preflightStatusSkip = "skip"
)

type preflightOptions struct {
	ProjectDir string
	InputFile  string
	OutputFile string
	GoBin      string
}

type preflightResult struct {
	SchemaVersion string           `json:"schema_version"`
	Command       string           `json:"command"`
	OK            bool             `json:"ok"`
	ReadOnly      bool             `json:"read_only"`
	Checks        []preflightCheck `json:"checks"`
}

type preflightCheck struct {
	Name        string `json:"name"`
	Status      string `json:"status"`
	Recoverable bool   `json:"recoverable"`
	Message     string `json:"message"`
}

func newPreflightCommand() *cobra.Command {
	var jsonOutput bool
	options := preflightOptions{
		ProjectDir: ".",
		GoBin:      "go",
	}

	cmd := &cobra.Command{
		Use:   "preflight",
		Short: "Run read-only readiness checks for Sheet Ops CLI execution",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !jsonOutput {
				return newCLIError(cliErrorInvalidUsage, "preflight requires --json", true, cliExitUsage, "preflight --json")
			}
			return writePreflightJSON(cmd.OutOrStdout(), runPreflight(options))
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Emit machine-readable JSON on stdout")
	cmd.Flags().StringVar(&options.ProjectDir, "project", ".", "Project directory to check")
	cmd.Flags().StringVar(&options.InputFile, "input-file", "", "Optional input workbook path to check")
	cmd.Flags().StringVar(&options.OutputFile, "output-file", "", "Optional output workbook path to check")
	cmd.Flags().StringVar(&options.GoBin, "go-bin", "go", "Go binary used for readiness checks")
	return cmd
}

func runPreflight(options preflightOptions) preflightResult {
	checks := []preflightCheck{
		checkProjectDirectory(options.ProjectDir),
		checkGoRuntime(options.GoBin),
		checkPackageManifest(),
		checkStateRoot(options.InputFile),
		checkInputFile(options.InputFile),
		checkOutputParent(options.OutputFile),
	}
	ok := true
	for _, check := range checks {
		if check.Status == preflightStatusFail {
			ok = false
			break
		}
	}
	return preflightResult{
		SchemaVersion: cliContractSchemaVersion,
		Command:       "preflight",
		OK:            ok,
		ReadOnly:      true,
		Checks:        checks,
	}
}

func checkProjectDirectory(projectDir string) preflightCheck {
	path := strings.TrimSpace(projectDir)
	if path == "" {
		return failedPreflightCheck("project_directory", true, "--project is empty")
	}
	info, err := os.Stat(path)
	if err != nil {
		return failedPreflightCheck("project_directory", true, fmt.Sprintf("project directory is not accessible: %v", err))
	}
	if !info.IsDir() {
		return failedPreflightCheck("project_directory", true, "project path is not a directory")
	}
	return passedPreflightCheck("project_directory", "project directory is accessible")
}

func checkGoRuntime(goBin string) preflightCheck {
	status := inspectGo(goBin)
	switch status.kind {
	case goStatusReady:
		return passedPreflightCheck("go_runtime", fmt.Sprintf("Go runtime is ready: %s", status.version))
	case goStatusOutdated:
		return failedPreflightCheck("go_runtime", true, fmt.Sprintf("Go runtime is outdated: %s", status.version))
	default:
		return failedPreflightCheck("go_runtime", true, "Go runtime is missing or unavailable")
	}
}

func checkPackageManifest() preflightCheck {
	file, err := embeddedSkillAssets.Open("skill_assets/agent-system/bundle.manifest.json")
	if err != nil {
		return failedPreflightCheck("package_manifest", false, fmt.Sprintf("embedded bundle manifest is unavailable: %v", err))
	}
	if err := file.Close(); err != nil {
		return failedPreflightCheck("package_manifest", false, fmt.Sprintf("embedded bundle manifest could not be closed: %v", err))
	}
	return passedPreflightCheck("package_manifest", "embedded bundle manifest is available")
}

func checkStateRoot(inputFile string) preflightCheck {
	inputPath := cleanRequiredPath(inputFile)
	if inputPath == "" {
		return skippedPreflightCheck("state_root", "input-file not provided; state-root ownership was not checked")
	}
	expected := filepath.Join(resolveWorkspaceRoot(inputPath), ".sheet-ops-state")
	current := strings.TrimSpace(os.Getenv(stateRootEnv))
	if current == "" {
		return passedPreflightCheck("state_root", fmt.Sprintf("%s is unset and will default to %s", stateRootEnv, expected))
	}
	if samePath(current, expected) {
		return passedPreflightCheck("state_root", fmt.Sprintf("%s matches workbook-case state root", stateRootEnv))
	}
	return failedPreflightCheck("state_root", true, fmt.Sprintf("%s=%q does not match expected %q", stateRootEnv, filepath.Clean(current), expected))
}

func checkInputFile(inputFile string) preflightCheck {
	inputPath := cleanRequiredPath(inputFile)
	if inputPath == "" {
		return skippedPreflightCheck("input_file", "input-file not provided")
	}
	info, err := os.Stat(inputPath)
	if err != nil {
		return failedPreflightCheck("input_file", true, fmt.Sprintf("input file is not accessible: %v", err))
	}
	if info.IsDir() {
		return failedPreflightCheck("input_file", true, "input file path is a directory, expected a workbook file")
	}
	file, err := os.Open(inputPath)
	if err != nil {
		return failedPreflightCheck("input_file", true, fmt.Sprintf("input file is not readable: %v", err))
	}
	if err := file.Close(); err != nil {
		return failedPreflightCheck("input_file", true, fmt.Sprintf("input file could not be closed: %v", err))
	}
	return passedPreflightCheck("input_file", "input file is readable")
}

func checkOutputParent(outputFile string) preflightCheck {
	outputPath := cleanRequiredPath(outputFile)
	if outputPath == "" {
		return skippedPreflightCheck("output_parent", "output-file not provided")
	}
	parent := filepath.Dir(outputPath)
	existing, err := nearestExistingAncestor(parent)
	if err != nil {
		return failedPreflightCheck("output_parent", true, err.Error())
	}
	info, err := os.Stat(existing)
	if err != nil {
		return failedPreflightCheck("output_parent", true, fmt.Sprintf("output parent ancestor is not accessible: %v", err))
	}
	if !info.IsDir() {
		return failedPreflightCheck("output_parent", true, "output parent ancestor is not a directory")
	}
	return passedPreflightCheck("output_parent", fmt.Sprintf("nearest output parent ancestor is accessible: %s", existing))
}

func nearestExistingAncestor(path string) (string, error) {
	current := filepath.Clean(path)
	for {
		if _, err := os.Stat(current); err == nil {
			return current, nil
		} else if !os.IsNotExist(err) {
			return "", fmt.Errorf("inspect output parent ancestor %s: %v", current, err)
		}
		next := filepath.Dir(current)
		if next == current {
			return "", fmt.Errorf("no existing output parent ancestor for %s", path)
		}
		current = next
	}
}

func passedPreflightCheck(name, message string) preflightCheck {
	return preflightCheck{Name: name, Status: preflightStatusPass, Recoverable: false, Message: message}
}

func failedPreflightCheck(name string, recoverable bool, message string) preflightCheck {
	return preflightCheck{Name: name, Status: preflightStatusFail, Recoverable: recoverable, Message: message}
}

func skippedPreflightCheck(name, message string) preflightCheck {
	return preflightCheck{Name: name, Status: preflightStatusSkip, Recoverable: true, Message: message}
}

func writePreflightJSON(output io.Writer, result preflightResult) error {
	encoder := json.NewEncoder(output)
	encoder.SetIndent("", "  ")
	return encoder.Encode(result)
}
