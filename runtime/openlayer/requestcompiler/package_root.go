package requestcompiler

import (
	"os"
	"path/filepath"
	"strings"
)

const packageRootEnv = "SHEET_OPS_PACKAGE_ROOT"

func resolvePackageRoot(workingDir, callerFile string) string {
	for _, candidate := range packageRootCandidates(workingDir, callerFile) {
		if hasRequestCompilerMaterials(candidate) {
			return candidate
		}
	}

	if callerFile == "" {
		return "."
	}
	if workingDir != "" {
		return filepath.Clean(workingDir)
	}
	return "."
}

func packageRootCandidates(workingDir, callerFile string) []string {
	candidates := make([]string, 0, 5)
	if explicit := strings.TrimSpace(os.Getenv(packageRootEnv)); explicit != "" {
		if absExplicit, err := filepath.Abs(explicit); err == nil {
			candidates = append(candidates, absExplicit)
		}
	}
	if callerFile != "" {
		candidates = append(candidates, filepath.Clean(filepath.Join(filepath.Dir(callerFile), "..", "..", "..")))
	}
	if workingDir != "" {
		candidates = append(candidates,
			filepath.Clean(filepath.Join(workingDir, ".codex", "skills", "sheet-ops")),
			filepath.Clean(workingDir),
		)
	}
	return candidates
}

func hasRequestCompilerMaterials(root string) bool {
	if root == "" {
		return false
	}
	for _, rel := range []string{
		filepath.Join("agents", "request-compiler", "agent.md"),
		filepath.Join("agents", "request-compiler", "prompt.md"),
		filepath.Join("agents", "request-compiler", "contract", "normalized_intent.schema.json"),
	} {
		info, err := os.Stat(filepath.Join(root, rel))
		if err != nil || info.IsDir() {
			return false
		}
	}
	return true
}
