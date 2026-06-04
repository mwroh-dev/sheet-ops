package main

import (
	"bufio"
	"embed"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

const (
	minimumGoMajor = 1
	minimumGoMinor = 25
	goInstallURL   = "https://go.dev/doc/install"
)

//go:embed skill_assets
var embeddedSkillAssets embed.FS

type installSkillOptions struct {
	projectDir       string
	global           bool
	yes              bool
	goBin            string
	sheetOpsCodexBin string
}

type goStatusKind string

const (
	goStatusReady    goStatusKind = "ready"
	goStatusMissing  goStatusKind = "missing"
	goStatusOutdated goStatusKind = "outdated"
)

type goStatus struct {
	kind    goStatusKind
	version string
	err     error
}

func newInstallSkillCommand() *cobra.Command {
	var options installSkillOptions

	cmd := &cobra.Command{
		Use:   "install-skill",
		Short: "Install the Sheet Ops Codex skill into a project",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := ensureGoReady(cmd, options); err != nil {
				return err
			}
			return installSkill(cmd.OutOrStdout(), options)
		},
	}

	cmd.Flags().StringVar(&options.projectDir, "project", ".", "Project directory for project-local install")
	cmd.Flags().BoolVar(&options.global, "global", false, "Install under $HOME/.codex instead of the project")
	cmd.Flags().BoolVarP(&options.yes, "yes", "y", false, "Accept prompts")
	cmd.Flags().StringVar(&options.goBin, "go-bin", "go", "Go binary used for version checks")
	cmd.Flags().StringVar(&options.sheetOpsCodexBin, "sheet-ops-codex-bin", "", "sheet-ops-codex binary path written into the installed wrapper")
	if err := cmd.Flags().MarkHidden("sheet-ops-codex-bin"); err != nil {
		panic(err)
	}

	return cmd
}

func ensureGoReady(cmd *cobra.Command, options installSkillOptions) error {
	status := inspectGo(options.goBin)
	switch status.kind {
	case goStatusReady:
		return nil
	case goStatusMissing:
		fmt.Fprintf(cmd.OutOrStdout(), "Go was not found. Sheet Ops requires Go 1.25 or newer.\n")
		if !confirm(cmd.InOrStdin(), cmd.OutOrStdout(), options.yes, "Open the official Go installation guide before continuing?") {
			return errors.New("Go installation declined")
		}
		return openGoInstallGuide()
	case goStatusOutdated:
		fmt.Fprintf(cmd.OutOrStdout(), "Go 1.25 or newer is required. Current version: %s.\n", status.version)
		if !confirm(cmd.InOrStdin(), cmd.OutOrStdout(), options.yes, "Open the official Go upgrade guide before continuing?") {
			return errors.New("Go upgrade declined")
		}
		return openGoInstallGuide()
	default:
		return fmt.Errorf("unknown Go status: %s", status.kind)
	}
}

func inspectGo(goBin string) goStatus {
	path, err := exec.LookPath(goBin)
	if err != nil {
		if _, statErr := os.Stat(goBin); statErr == nil {
			path = goBin
		} else {
			return goStatus{kind: goStatusMissing, err: err}
		}
	}

	output, err := exec.Command(path, "version").Output()
	if err != nil {
		return goStatus{kind: goStatusMissing, err: err}
	}

	version := parseGoVersion(string(output))
	if version == "" {
		return goStatus{kind: goStatusMissing, err: fmt.Errorf("could not parse go version output: %s", strings.TrimSpace(string(output)))}
	}
	if !goVersionAtLeast(version, minimumGoMajor, minimumGoMinor) {
		return goStatus{kind: goStatusOutdated, version: version}
	}
	return goStatus{kind: goStatusReady, version: version}
}

func parseGoVersion(output string) string {
	re := regexp.MustCompile(`go([0-9]+)\.([0-9]+)(?:\.[0-9]+)?`)
	match := re.FindStringSubmatch(output)
	if len(match) == 0 {
		return ""
	}
	return match[0]
}

func goVersionAtLeast(version string, major, minor int) bool {
	re := regexp.MustCompile(`^go([0-9]+)\.([0-9]+)`)
	match := re.FindStringSubmatch(version)
	if len(match) != 3 {
		return false
	}
	actualMajor, err := strconv.Atoi(match[1])
	if err != nil {
		return false
	}
	actualMinor, err := strconv.Atoi(match[2])
	if err != nil {
		return false
	}
	if actualMajor != major {
		return actualMajor > major
	}
	return actualMinor >= minor
}

func confirm(input io.Reader, output io.Writer, yes bool, prompt string) bool {
	if yes {
		fmt.Fprintln(output, prompt+" yes")
		return true
	}

	fmt.Fprintf(output, "%s [y/N] ", prompt)
	scanner := bufio.NewScanner(input)
	if !scanner.Scan() {
		return false
	}
	answer := strings.TrimSpace(strings.ToLower(scanner.Text()))
	return answer == "y" || answer == "yes"
}

func openGoInstallGuide() error {
	if exec.Command("open", goInstallURL).Run() == nil {
		return fmt.Errorf("install Go 1.25 or newer from %s, then rerun install-skill", goInstallURL)
	}
	return fmt.Errorf("install Go 1.25 or newer from %s, then rerun install-skill", goInstallURL)
}

func installSkill(output io.Writer, options installSkillOptions) error {
	targetRoot, err := resolveInstallRoot(options)
	if err != nil {
		return err
	}

	targetDir := filepath.Join(targetRoot, "skills", "sheet-ops")
	tempDir, err := os.MkdirTemp(filepath.Dir(targetDir), ".sheet-ops.tmp.")
	if err != nil {
		if mkdirErr := os.MkdirAll(filepath.Dir(targetDir), 0o755); mkdirErr != nil {
			return mkdirErr
		}
		tempDir, err = os.MkdirTemp(filepath.Dir(targetDir), ".sheet-ops.tmp.")
		if err != nil {
			return err
		}
	}
	defer os.RemoveAll(tempDir)

	if err := materializeCanonicalPackage(tempDir); err != nil {
		return err
	}
	if err := materializeCanonicalSkillEntry(tempDir); err != nil {
		return err
	}
	runtimeAdapterPath, err := resolveInstalledRuntimeAdapterPath(targetDir, options.sheetOpsCodexBin)
	if err != nil {
		return err
	}
	if err := writeEmbeddedSkillAssets(tempDir, runtimeAdapterPath); err != nil {
		return err
	}
	if err := os.RemoveAll(targetDir); err != nil {
		return err
	}
	if err := os.Rename(tempDir, targetDir); err != nil {
		return err
	}
	if err := buildBundledEntrypoints(targetDir, options.goBin); err != nil {
		return err
	}

	fmt.Fprintf(output, "installed sheet-ops skill to %s\n", targetDir)
	return nil
}

func resolveInstallRoot(options installSkillOptions) (string, error) {
	if options.global {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, ".codex"), nil
	}

	projectDir, err := filepath.Abs(options.projectDir)
	if err != nil {
		return "", err
	}
	return filepath.Join(projectDir, ".codex"), nil
}

var canonicalInstallManifestPath = filepath.Join("cmd", "sheet-ops-codex", "install_manifest.txt")

func materializeCanonicalPackage(targetDir string) error {
	sourceRoot, err := resolveSheetOpsSourceRoot()
	if err != nil {
		return err
	}

	relativePaths, err := loadCanonicalInstallPaths(sourceRoot)
	if err != nil {
		return err
	}
	for _, relativePath := range relativePaths {
		if err := copyPath(filepath.Join(sourceRoot, relativePath), filepath.Join(targetDir, relativePath)); err != nil {
			return fmt.Errorf("copy canonical package path %s: %w", relativePath, err)
		}
	}
	return nil
}

func resolveInstalledRuntimeAdapterPath(targetDir, explicit string) (string, error) {
	if explicit != "" {
		return filepath.Abs(explicit)
	}
	return "$PACKAGE_ROOT/bin/sheet-ops-codex", nil
}

func loadCanonicalInstallPaths(sourceRoot string) ([]string, error) {
	manifestPath := filepath.Join(sourceRoot, canonicalInstallManifestPath)
	body, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("read install manifest %s: %w", manifestPath, err)
	}

	var paths []string
	for _, line := range strings.Split(string(body), "\n") {
		entry := strings.TrimSpace(line)
		if entry == "" || strings.HasPrefix(entry, "#") {
			continue
		}
		paths = append(paths, filepath.FromSlash(entry))
	}
	if len(paths) == 0 {
		return nil, fmt.Errorf("install manifest %s did not contain any paths", manifestPath)
	}
	return paths, nil
}

func materializeCanonicalSkillEntry(targetDir string) error {
	sourceRoot, err := resolveSheetOpsSourceRoot()
	if err != nil {
		return err
	}

	skillEntrySourceRoot, err := resolveCanonicalSkillEntrySourceRoot(sourceRoot)
	if err != nil {
		return err
	}
	for _, mapping := range []struct {
		source string
		target string
	}{
		{source: filepath.Join(skillEntrySourceRoot, "SKILL.md"), target: "SKILL.md"},
		{source: filepath.Join(skillEntrySourceRoot, "references"), target: "references"},
	} {
		if err := copyPath(mapping.source, filepath.Join(targetDir, mapping.target)); err != nil {
			return fmt.Errorf("copy canonical skill entry %s: %w", mapping.source, err)
		}
	}
	return nil
}

func resolveCanonicalSkillEntrySourceRoot(sourceRoot string) (string, error) {
	candidates := []string{
		filepath.Join(sourceRoot, "skills", "sheet-ops"),
		sourceRoot,
	}
	for _, candidate := range candidates {
		skillPath := filepath.Join(candidate, "SKILL.md")
		referencesPath := filepath.Join(candidate, "references")
		if info, err := os.Stat(skillPath); err == nil && !info.IsDir() {
			if info, err := os.Stat(referencesPath); err == nil && info.IsDir() {
				return candidate, nil
			}
		}
	}
	return "", fmt.Errorf("could not resolve canonical skill entry under %s", sourceRoot)
}

func writeEmbeddedSkillAssets(targetDir, binaryPath string) error {
	return fs.WalkDir(embeddedSkillAssets, "skill_assets", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		relativePath, err := filepath.Rel("skill_assets", path)
		if err != nil {
			return err
		}
		if relativePath == "SKILL.md" || strings.HasPrefix(relativePath, "references/") {
			return nil
		}
		targetPath := filepath.Join(targetDir, relativePath)
		if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
			return err
		}

		content, err := embeddedSkillAssets.ReadFile(path)
		if err != nil {
			return err
		}
		if relativePath == filepath.Join("scripts", "run-sheet-ops.sh") {
			content = []byte(strings.ReplaceAll(string(content), "__SHEET_OPS_CODEX_BIN__", binaryPath))
		}

		mode := fs.FileMode(0o644)
		if strings.HasPrefix(relativePath, "scripts"+string(filepath.Separator)) || strings.HasPrefix(relativePath, "scripts/") ||
			strings.HasPrefix(relativePath, filepath.Join("agent-system", "scripts")+string(filepath.Separator)) ||
			strings.HasPrefix(relativePath, "agent-system/scripts/") {
			mode = 0o755
		}
		return os.WriteFile(targetPath, content, mode)
	})
}

func buildBundledEntrypoints(skillDir, goBin string) error {
	for _, entrypoint := range []struct {
		targetPath string
		packageDir string
		label      string
	}{
		{
			targetPath: filepath.Join(skillDir, "agent-system", "bin", "sheet-ops-agent"),
			packageDir: "./cmd/sheet-ops-agent",
			label:      "sheet-ops-agent",
		},
		{
			targetPath: filepath.Join(skillDir, "bin", "sheet-ops-codex"),
			packageDir: "./cmd/sheet-ops-codex",
			label:      "sheet-ops-codex",
		},
	} {
		if err := os.MkdirAll(filepath.Dir(entrypoint.targetPath), 0o755); err != nil {
			return err
		}
		cmd := exec.Command(goBin, "build", "-o", entrypoint.targetPath, entrypoint.packageDir)
		cmd.Dir = skillDir
		cmd.Env = os.Environ()
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("build bundled %s from %s: %w\n%s", entrypoint.label, skillDir, err, strings.TrimSpace(string(output)))
		}
		if err := os.Chmod(entrypoint.targetPath, 0o755); err != nil {
			return err
		}
	}
	return nil
}

func resolveSheetOpsSourceRoot() (string, error) {
	if explicit := os.Getenv("SHEET_OPS_REPO_ROOT"); explicit != "" {
		return validateSheetOpsSourceRoot(explicit)
	}

	_, filePath, _, ok := runtime.Caller(0)
	if !ok {
		return "", errors.New("could not resolve sheet-ops source root from runtime caller")
	}
	return validateSheetOpsSourceRoot(filepath.Clean(filepath.Join(filepath.Dir(filePath), "..", "..")))
}

func validateSheetOpsSourceRoot(candidate string) (string, error) {
	root, err := filepath.Abs(candidate)
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err != nil {
		return "", fmt.Errorf("sheet-ops source root %s is missing go.mod: %w", root, err)
	}
	if info, err := os.Stat(filepath.Join(root, "cmd", "sheet-ops-agent")); err != nil {
		return "", fmt.Errorf("sheet-ops source root %s is missing cmd/sheet-ops-agent: %w", root, err)
	} else if !info.IsDir() {
		return "", fmt.Errorf("sheet-ops source root %s has non-directory cmd/sheet-ops-agent", root)
	}
	return root, nil
}

func copyPath(sourcePath, targetPath string) error {
	info, err := os.Lstat(sourcePath)
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("refusing to copy symlink path %s", sourcePath)
	}

	if info.IsDir() {
		return filepath.WalkDir(sourcePath, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.Type()&os.ModeSymlink != 0 {
				return fmt.Errorf("refusing to copy symlink path %s", path)
			}

			relativePath, err := filepath.Rel(sourcePath, path)
			if err != nil {
				return err
			}
			if relativePath == "." {
				return os.MkdirAll(targetPath, 0o755)
			}

			destinationPath := filepath.Join(targetPath, relativePath)
			if d.IsDir() {
				return os.MkdirAll(destinationPath, 0o755)
			}
			return copyFile(path, destinationPath)
		})
	}

	return copyFile(sourcePath, targetPath)
}

func copyFile(sourcePath, targetPath string) error {
	info, err := os.Lstat(sourcePath)
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("refusing to copy symlink path %s", sourcePath)
	}

	content, err := os.ReadFile(sourcePath)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		return err
	}
	return os.WriteFile(targetPath, content, info.Mode().Perm())
}
