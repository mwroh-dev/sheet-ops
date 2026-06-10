package main

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestInstallSkillInstallsProjectLocalSkillWhenGoIsReady(t *testing.T) {
	projectDir := t.TempDir()
	goDir := t.TempDir()
	goPath := writeFakeGo(t, goDir, "go version go1.25.10 darwin/arm64\n")

	cmd := newRootCommand()
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stdout)
	cmd.SetArgs([]string{
		"install-skill",
		"--project", projectDir,
		"--go-bin", goPath,
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("install-skill failed: %v\noutput=%s", err, stdout.String())
	}

	skillRoot := filepath.Join(projectDir, ".codex", "skills", "sheet-ops")
	assertFileContains(t, filepath.Join(skillRoot, "README.md"), "# Sheet Ops")
	assertFileContains(t, filepath.Join(skillRoot, "AGENTS.md"), "## Canonical Package Contract")
	assertFileContains(t, filepath.Join(skillRoot, "CLAUDE.md"), "## Canonical Package Contract")
	assertFileContains(t, filepath.Join(skillRoot, "install-skill.sh"), "install Go 1.25 or newer")
	assertFileContains(t, filepath.Join(skillRoot, "SKILL.md"), "name: sheet-ops")
	assertFileContains(t, filepath.Join(skillRoot, "references", "quickstart.md"), "skill-owned runtime handoff")
	assertFileContains(t, filepath.Join(skillRoot, "references", "capabilities.md"), "contracts/capabilities/records")
	assertFileContains(t, filepath.Join(skillRoot, "references", "request-examples.md"), "group_summarize")
	assertFileContains(t, filepath.Join(skillRoot, "references", "verification.md"), "Evidence is scenario-level proof")
	assertFileContains(t, filepath.Join(skillRoot, "references", "failure-repair.md"), "repair advice artifact")
	assertFileContains(t, filepath.Join(skillRoot, "references", "artifact-layout.md"), "artifacts/evidence")
	assertFileContains(t, filepath.Join(skillRoot, "scripts", "run-sheet-ops.sh"), "$PACKAGE_ROOT/bin/sheet-ops-codex")
	assertExistsLocal(t, filepath.Join(skillRoot, "bin", "sheet-ops-codex"))
	assertNotExistsLocal(t, filepath.Join(skillRoot, "platforms"))
}

func TestInstallSkillBundledCLIExposesAgentContract(t *testing.T) {
	if status := inspectGo("go"); status.kind != goStatusReady {
		t.Skipf("go runtime is not ready for bundled CLI smoke: %s %v", status.kind, status.err)
	}

	projectDir := t.TempDir()
	cmd := newRootCommand()
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stdout)
	cmd.SetArgs([]string{
		"install-skill",
		"--project", projectDir,
		"--go-bin", "go",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("install-skill failed: %v\noutput=%s", err, stdout.String())
	}

	installedCLI := filepath.Join(projectDir, ".codex", "skills", "sheet-ops", "bin", "sheet-ops-codex")
	assertExistsLocal(t, installedCLI)

	var capabilities cliCapabilitiesDocument
	runInstalledCLIJSON(t, installedCLI, &capabilities, "capabilities", "--json")
	if !containsString(capabilities.MachineEntryCommands, "schema") {
		t.Fatalf("machine_entry_commands = %v, want schema", capabilities.MachineEntryCommands)
	}
	agentGroup := findCapabilityGroup(t, capabilities.CommandGroups, "agent_contract")
	if !containsString(agentGroup.Commands, "preflight") {
		t.Fatalf("agent_contract commands = %v, want preflight", agentGroup.Commands)
	}

	var prepareUseSchema cliCommandSchemaEnvelope
	runInstalledCLIJSON(t, installedCLI, &prepareUseSchema, "schema", "command", "prepare-use", "--json")
	if prepareUseSchema.Command.Name != "prepare-use" {
		t.Fatalf("schema command name = %q, want prepare-use", prepareUseSchema.Command.Name)
	}
	if prepareUseSchema.Command.Classification != cliClassificationAgentContract {
		t.Fatalf("prepare-use classification = %q, want %q", prepareUseSchema.Command.Classification, cliClassificationAgentContract)
	}
	if !prepareUseSchema.Command.Mutating || prepareUseSchema.Command.ReadOnly {
		t.Fatalf("prepare-use schema has unsafe flags: %+v", prepareUseSchema.Command)
	}
	if !schemaOptionsInclude(prepareUseSchema.Command.Options, "--envelope-file") {
		t.Fatalf("prepare-use options = %+v, want --envelope-file", prepareUseSchema.Command.Options)
	}

	var preflightSchema cliCommandSchemaEnvelope
	runInstalledCLIJSON(t, installedCLI, &preflightSchema, "schema", "command", "preflight", "--json")
	if preflightSchema.Command.Name != "preflight" {
		t.Fatalf("schema command name = %q, want preflight", preflightSchema.Command.Name)
	}
	if !preflightSchema.Command.ReadOnly || preflightSchema.Command.Mutating || preflightSchema.Command.DryRunCapable {
		t.Fatalf("preflight schema has unsafe flags: %+v", preflightSchema.Command)
	}

	var preflight preflightDocumentView
	runInstalledCLIJSON(t, installedCLI, &preflight, "preflight", "--json", "--project", projectDir)
	if !preflight.ReadOnly {
		t.Fatalf("preflight read_only = false, want true")
	}
	if findPreflightCheck(t, preflight.Checks, "package_manifest").Status != preflightStatusPass {
		t.Fatalf("package_manifest check did not pass: %+v", preflight.Checks)
	}
}

func TestInstallSkillSkipsPromptWhenGoIsReady(t *testing.T) {
	projectDir := t.TempDir()
	goDir := t.TempDir()
	goPath := writeFakeGo(t, goDir, "go version go1.26.3 darwin/arm64\n")

	cmd := newRootCommand()
	cmd.SetIn(strings.NewReader(""))
	cmd.SetArgs([]string{
		"install-skill",
		"--project", projectDir,
		"--go-bin", goPath,
		"--sheet-ops-codex-bin", "/opt/sheet-ops/bin/sheet-ops-codex",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("install-skill should not prompt when Go is ready: %v", err)
	}
}

func runInstalledCLIJSON(t *testing.T, binary string, target any, args ...string) {
	t.Helper()

	cmd := exec.Command(binary, args...)
	stdout, err := cmd.Output()
	if err != nil {
		t.Fatalf("%s %s failed: %v", binary, strings.Join(args, " "), err)
	}
	if err := json.Unmarshal(stdout, target); err != nil {
		t.Fatalf("json.Unmarshal(%s %s): %v\nstdout:\n%s", binary, strings.Join(args, " "), err, stdout)
	}
}

func schemaOptionsInclude(options []cliSchemaOptionView, name string) bool {
	for _, option := range options {
		if option.Name == name {
			return true
		}
	}
	return false
}

func TestInstallSkillPromptsBeforeContinuingWhenGoIsOutdated(t *testing.T) {
	projectDir := t.TempDir()
	goDir := t.TempDir()
	goPath := writeFakeGo(t, goDir, "go version go1.24.9 darwin/arm64\n")

	cmd := newRootCommand()
	cmd.SetIn(strings.NewReader("n\n"))
	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetErr(&output)
	cmd.SetArgs([]string{
		"install-skill",
		"--project", projectDir,
		"--go-bin", goPath,
		"--sheet-ops-codex-bin", "/opt/sheet-ops/bin/sheet-ops-codex",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected outdated Go prompt rejection to fail")
	}
	if !strings.Contains(output.String(), "Go 1.25 or newer is required") {
		t.Fatalf("prompt output=%q, want Go requirement", output.String())
	}
	if !strings.Contains(err.Error(), "Go upgrade declined") {
		t.Fatalf("error=%q, want declined message", err.Error())
	}
	assertNotExistsLocal(t, filepath.Join(projectDir, ".codex", "skills", "sheet-ops"))
}

func TestInstallSkillPromptsBeforeContinuingWhenGoIsMissing(t *testing.T) {
	projectDir := t.TempDir()

	cmd := newRootCommand()
	cmd.SetIn(strings.NewReader("n\n"))
	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetErr(&output)
	cmd.SetArgs([]string{
		"install-skill",
		"--project", projectDir,
		"--go-bin", filepath.Join(t.TempDir(), "missing-go"),
		"--sheet-ops-codex-bin", "/opt/sheet-ops/bin/sheet-ops-codex",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected missing Go prompt rejection to fail")
	}
	if !strings.Contains(output.String(), "Go was not found") {
		t.Fatalf("prompt output=%q, want missing Go message", output.String())
	}
	if !strings.Contains(err.Error(), "Go installation declined") {
		t.Fatalf("error=%q, want declined message", err.Error())
	}
	assertNotExistsLocal(t, filepath.Join(projectDir, ".codex", "skills", "sheet-ops"))
}

func TestCopyPathRejectsSymlinkedFiles(t *testing.T) {
	sourceDir := t.TempDir()
	targetDir := t.TempDir()

	sourceFile := filepath.Join(sourceDir, "source.txt")
	if err := os.WriteFile(sourceFile, []byte("data"), 0o644); err != nil {
		t.Fatalf("WriteFile(source): %v", err)
	}
	symlinkPath := filepath.Join(sourceDir, "link.txt")
	if err := os.Symlink(sourceFile, symlinkPath); err != nil {
		t.Fatalf("Symlink(link.txt): %v", err)
	}

	if err := copyPath(symlinkPath, filepath.Join(targetDir, "copied.txt")); err == nil {
		t.Fatal("copyPath unexpectedly accepted symlinked file")
	}
}

func TestCopyPathRejectsSymlinkedDirectories(t *testing.T) {
	sourceDir := t.TempDir()
	targetDir := t.TempDir()

	realDir := filepath.Join(sourceDir, "real")
	if err := os.MkdirAll(realDir, 0o755); err != nil {
		t.Fatalf("MkdirAll(real): %v", err)
	}
	if err := os.WriteFile(filepath.Join(realDir, "source.txt"), []byte("data"), 0o644); err != nil {
		t.Fatalf("WriteFile(source): %v", err)
	}
	symlinkDir := filepath.Join(sourceDir, "link-dir")
	if err := os.Symlink(realDir, symlinkDir); err != nil {
		t.Fatalf("Symlink(link-dir): %v", err)
	}

	if err := copyPath(symlinkDir, filepath.Join(targetDir, "copied")); err == nil {
		t.Fatal("copyPath unexpectedly accepted symlinked directory")
	}
}

func writeFakeGo(t *testing.T, dir, versionOutput string) string {
	t.Helper()

	path := filepath.Join(dir, "go")
	script := "#!/usr/bin/env bash\n" +
		"set -euo pipefail\n" +
		"if [[ \"${1-}\" == \"version\" ]]; then\n" +
		"  printf '%s' " + shellQuote(versionOutput) + "\n" +
		"  exit 0\n" +
		"fi\n" +
		"if [[ \"${1-}\" == \"build\" && \"${2-}\" == \"-o\" && -n \"${3-}\" ]]; then\n" +
		"  mkdir -p \"$(dirname \"$3\")\"\n" +
		"  printf '#!/usr/bin/env bash\\nexit 0\\n' >\"$3\"\n" +
		"  chmod 755 \"$3\"\n" +
		"  exit 0\n" +
		"fi\n" +
		"printf 'unsupported fake go command: %s\\n' \"$*\" >&2\n" +
		"exit 1\n"
	if runtime.GOOS == "windows" {
		t.Fatal("writeFakeGo test helper is not implemented for windows")
	}
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatalf("WriteFile(fake go): %v", err)
	}
	return path
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'"
}

func assertFileContains(t *testing.T, path, needle string) {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%s): %v", path, err)
	}
	if !strings.Contains(string(content), needle) {
		t.Fatalf("%s missing %q:\n%s", path, needle, string(content))
	}
}

func assertNotExistsLocal(t *testing.T, path string) {
	t.Helper()

	if _, err := os.Stat(path); err == nil {
		t.Fatalf("expected path to be absent: %s", path)
	} else if !os.IsNotExist(err) {
		t.Fatalf("stat(%s): %v", path, err)
	}
}

func assertExistsLocal(t *testing.T, path string) {
	t.Helper()

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected path to exist: %s: %v", path, err)
	}
}
