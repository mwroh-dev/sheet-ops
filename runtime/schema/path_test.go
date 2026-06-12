package schema

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveRepoPathUsesExistingPackageRootEnv(t *testing.T) {
	root := t.TempDir()
	rel := filepath.Join("contracts", "requests", "request_ref.schema.json")
	path := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(path, []byte("{}\n"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	t.Setenv(packageRootEnv, root)

	got := ResolveRepoPath(filepath.Join(t.TempDir(), "runtime", "schema", "validate.go"), 2, "contracts", "requests", "request_ref.schema.json")
	if got != path {
		t.Fatalf("ResolveRepoPath()=%q want %q", got, path)
	}
}

func TestResolveRepoPathIgnoresBlankPackageRootEnv(t *testing.T) {
	t.Setenv(packageRootEnv, "   ")
	callerFile := filepath.Join(t.TempDir(), "runtime", "schema", "validate.go")

	got := ResolveRepoPath(callerFile, 2, "contracts", "requests", "missing.schema.json")

	if got != filepath.Join("contracts", "requests", "missing.schema.json") {
		t.Fatalf("ResolveRepoPath()=%q want repository-relative fallback", got)
	}
}

func TestResolveRepoPathAbsolutizesRelativePackageRootEnv(t *testing.T) {
	workingDir := t.TempDir()
	root := filepath.Join(workingDir, "package-root")
	rel := filepath.Join("contracts", "requests", "request_ref.schema.json")
	path := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(path, []byte("{}\n"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	oldWorkingDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	if err := os.Chdir(workingDir); err != nil {
		t.Fatalf("Chdir: %v", err)
	}
	defer func() {
		if err := os.Chdir(oldWorkingDir); err != nil {
			t.Fatalf("restore working dir: %v", err)
		}
	}()
	t.Setenv(packageRootEnv, "package-root")

	got := ResolveRepoPath(filepath.Join(t.TempDir(), "runtime", "schema", "validate.go"), 2, "contracts", "requests", "request_ref.schema.json")

	if canonicalPath(t, got) != canonicalPath(t, path) {
		t.Fatalf("ResolveRepoPath()=%q want %q", got, path)
	}
}

func TestResolveRepoPathDoesNotReturnMissingCallerDerivedSourcePath(t *testing.T) {
	callerFile := filepath.Join(t.TempDir(), "runtime", "schema", "validate.go")

	got := ResolveRepoPath(callerFile, 2, "contracts", "requests", "missing.schema.json")
	if filepath.IsAbs(got) {
		t.Fatalf("ResolveRepoPath()=%q returned missing absolute caller-derived path", got)
	}
	if got != filepath.Join("contracts", "requests", "missing.schema.json") {
		t.Fatalf("ResolveRepoPath()=%q want repository-relative fallback", got)
	}
}

func canonicalPath(t *testing.T, path string) string {
	t.Helper()

	canonical, err := filepath.EvalSymlinks(path)
	if err != nil {
		t.Fatalf("EvalSymlinks(%s): %v", path, err)
	}
	return canonical
}

func TestResolveRepoPathIgnoresTrimpathRelativeCallerFile(t *testing.T) {
	workingDir := t.TempDir()
	rel := filepath.Join("contracts", "requests", "request_ref.schema.json")
	callerDerivedPath := filepath.Join(workingDir, "trimpath", rel)
	if err := os.MkdirAll(filepath.Dir(callerDerivedPath), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(callerDerivedPath, []byte("{}\n"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	oldWorkingDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	if err := os.Chdir(workingDir); err != nil {
		t.Fatalf("Chdir: %v", err)
	}
	defer func() {
		if err := os.Chdir(oldWorkingDir); err != nil {
			t.Fatalf("restore working dir: %v", err)
		}
	}()

	got := ResolveRepoPath(filepath.Join("trimpath", "runtime", "schema", "validate.go"), 2, "contracts", "requests", "request_ref.schema.json")
	if got == callerDerivedPath || got == filepath.Join("trimpath", rel) {
		t.Fatalf("ResolveRepoPath()=%q used relative caller-derived source root", got)
	}
	if got != rel {
		t.Fatalf("ResolveRepoPath()=%q want repository-relative fallback", got)
	}
}
