package bundle

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const caseLocalBundleRoot = ".codex/skills/sheet-ops/agent-system"

type InstallProof struct {
	BundleRoot        string `json:"bundle_root"`
	ManifestPath      string `json:"manifest_path"`
	ModelContractPath string `json:"model_contract_path"`
	Status            string `json:"status"`
}

func RequireInstallRoot(caseRoot string, bundleRoot string) error {
	info, err := os.Lstat(bundleRoot)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("case-local bundle is missing at %s", bundleRoot)
		}
		return fmt.Errorf("inspect case-local bundle %s: %w", bundleRoot, err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("case-local bundle root %s must not be a symlink", bundleRoot)
	}
	if !info.IsDir() {
		return fmt.Errorf("case-local bundle root %s must be a directory", bundleRoot)
	}

	resolvedCaseRoot, err := filepath.EvalSymlinks(caseRoot)
	if err != nil {
		return fmt.Errorf("resolve case root %s: %w", caseRoot, err)
	}
	resolvedBundleRoot, err := filepath.EvalSymlinks(bundleRoot)
	if err != nil {
		return fmt.Errorf("resolve case-local bundle root %s: %w", bundleRoot, err)
	}
	expectedBundleRoot := filepath.Clean(filepath.Join(resolvedCaseRoot, caseLocalBundleRoot))
	if resolvedBundleRoot != expectedBundleRoot {
		return fmt.Errorf("case-local bundle root %s resolves to %s; expected %s", bundleRoot, resolvedBundleRoot, expectedBundleRoot)
	}

	for _, entry := range []string{"bundle.manifest.json", "model-contract.json"} {
		entryPath := filepath.Join(bundleRoot, entry)
		entryInfo, err := os.Lstat(entryPath)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf("case-local bundle is incomplete at %s; missing %s", bundleRoot, entryPath)
			}
			return fmt.Errorf("inspect bundle entry %s: %w", entryPath, err)
		}
		if entryInfo.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("case-local bundle entry %s must not be a symlink", entryPath)
		}
		if entryInfo.IsDir() {
			return fmt.Errorf("case-local bundle entry %s must be a file", entryPath)
		}
	}

	return nil
}
