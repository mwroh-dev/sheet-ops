package schema

import (
	"os"
	"path/filepath"
)

const packageRootEnv = "SHEET_OPS_PACKAGE_ROOT"

// ResolveRepoPath locates repository-relative runtime contracts across source,
// installed skill, and bundled binary layouts. Caller-derived source paths are
// returned only when they still exist on the current machine.
func ResolveRepoPath(callerFile string, sourceRootDepth int, rel ...string) string {
	relPath := filepath.Join(rel...)
	for _, candidate := range repoPathCandidates(callerFile, sourceRootDepth, relPath) {
		if fileExists(candidate) {
			return candidate
		}
	}
	return relPath
}

func repoPathCandidates(callerFile string, sourceRootDepth int, relPath string) []string {
	candidates := make([]string, 0, 7)
	if explicit := filepath.Clean(os.Getenv(packageRootEnv)); explicit != "." && explicit != "" {
		candidates = append(candidates, filepath.Join(explicit, relPath))
	}
	if workingDir, err := os.Getwd(); err == nil && workingDir != "" {
		candidates = append(candidates,
			filepath.Join(workingDir, ".codex", "skills", "sheet-ops", "agent-system", relPath),
			filepath.Join(workingDir, ".codex", "skills", "sheet-ops", relPath),
			filepath.Join(workingDir, relPath),
		)
	}
	if executablePath, err := os.Executable(); err == nil && executablePath != "" {
		candidates = append(candidates, filepath.Join(filepath.Dir(executablePath), "..", relPath))
	}
	if filepath.IsAbs(callerFile) {
		root := filepath.Dir(callerFile)
		for i := 0; i < sourceRootDepth; i++ {
			root = filepath.Dir(root)
		}
		candidates = append(candidates, filepath.Join(root, relPath))
	}
	return candidates
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
