package useorchestrator

import (
	"os"
	"path/filepath"
	"testing"

	runtimeknowledge "github.com/mwroh/sheet-ops/runtime/knowledge"
	runtimeworkbookcase "github.com/mwroh/sheet-ops/runtime/workbookcase"
)

func TestMain(m *testing.M) {
	root, err := os.MkdirTemp("", "sheet-ops-useorchestrator-test-*")
	if err != nil {
		panic(err)
	}
	_ = os.Setenv(runtimeworkbookcase.ArtifactRootEnv, filepath.Join(root, "artifacts"))
	_ = os.Setenv(runtimeknowledge.KnowledgeRootEnv, filepath.Join(root, "knowledge"))
	code := m.Run()
	_ = os.RemoveAll(root)
	os.Exit(code)
}
