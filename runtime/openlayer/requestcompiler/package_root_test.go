package requestcompiler

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolvePackageRootPrefersInstalledSkillTreeFromWorkspace(t *testing.T) {
	workspaceDir := t.TempDir()
	installedRoot := filepath.Join(workspaceDir, ".codex", "skills", "sheet-ops")
	mustWriteRequestCompilerFixture(t, installedRoot)

	callerFile := filepath.Join(t.TempDir(), "runtime", "openlayer", "requestcompiler", "interpreter.go")
	got := resolvePackageRoot(workspaceDir, callerFile)
	if got != installedRoot {
		t.Fatalf("resolvePackageRoot()=%q want %q", got, installedRoot)
	}
}

func TestResolvePackageRootPrefersExplicitInstalledPackageRootSignal(t *testing.T) {
	workspaceDir := t.TempDir()
	installedRoot := filepath.Join(t.TempDir(), ".codex", "skills", "sheet-ops")
	sourceRoot := t.TempDir()
	mustWriteRequestCompilerFixture(t, installedRoot)
	mustWriteRequestCompilerFixture(t, sourceRoot)
	t.Setenv("SHEET_OPS_PACKAGE_ROOT", installedRoot)

	callerFile := filepath.Join(sourceRoot, "runtime", "openlayer", "requestcompiler", "interpreter.go")
	got := resolvePackageRoot(workspaceDir, callerFile)
	if got != installedRoot {
		t.Fatalf("resolvePackageRoot()=%q want %q", got, installedRoot)
	}
}

func TestResolvePackageRootFallsBackToCallerDerivedSourceRoot(t *testing.T) {
	sourceRoot := t.TempDir()
	mustWriteRequestCompilerFixture(t, sourceRoot)

	callerFile := filepath.Join(sourceRoot, "runtime", "openlayer", "requestcompiler", "interpreter.go")
	got := resolvePackageRoot(t.TempDir(), callerFile)
	if got != sourceRoot {
		t.Fatalf("resolvePackageRoot()=%q want %q", got, sourceRoot)
	}
}

func TestResolvePackageRootAcceptsPackageRootWorkingDirectory(t *testing.T) {
	packageRoot := t.TempDir()
	mustWriteRequestCompilerFixture(t, packageRoot)

	callerFile := filepath.Join(t.TempDir(), "runtime", "openlayer", "requestcompiler", "interpreter.go")
	got := resolvePackageRoot(packageRoot, callerFile)
	if got != packageRoot {
		t.Fatalf("resolvePackageRoot()=%q want %q", got, packageRoot)
	}
}

func TestResolvePackageRootPrefersCallerDerivedSourceRootOverWorkspaceInstalledFallback(t *testing.T) {
	workspaceDir := t.TempDir()
	installedRoot := filepath.Join(workspaceDir, ".codex", "skills", "sheet-ops")
	sourceRoot := t.TempDir()
	mustWriteRequestCompilerFixture(t, installedRoot)
	mustWriteRequestCompilerFixture(t, sourceRoot)

	callerFile := filepath.Join(sourceRoot, "runtime", "openlayer", "requestcompiler", "interpreter.go")
	got := resolvePackageRoot(workspaceDir, callerFile)
	if got != sourceRoot {
		t.Fatalf("resolvePackageRoot()=%q want %q", got, sourceRoot)
	}
}

func TestResolvePackageRootDoesNotReturnMissingCallerDerivedSourceRoot(t *testing.T) {
	workspaceDir := t.TempDir()
	callerFile := filepath.Join(t.TempDir(), "runtime", "openlayer", "requestcompiler", "interpreter.go")

	got := resolvePackageRoot(workspaceDir, callerFile)
	if got != workspaceDir {
		t.Fatalf("resolvePackageRoot()=%q want working dir fallback %q", got, workspaceDir)
	}
}

func mustWriteRequestCompilerFixture(t *testing.T, root string) {
	t.Helper()

	for _, rel := range []string{
		filepath.Join("agents", "request-compiler", "agent.md"),
		filepath.Join("agents", "request-compiler", "prompt.md"),
		filepath.Join("agents", "request-compiler", "contract", "normalized_intent.schema.json"),
	} {
		path := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("MkdirAll(%s): %v", path, err)
		}
		if err := os.WriteFile(path, []byte("{}\n"), 0o644); err != nil {
			t.Fatalf("WriteFile(%s): %v", path, err)
		}
	}
}
