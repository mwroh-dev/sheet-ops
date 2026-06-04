package workbookcase

import (
	"fmt"
	"os"
	"strings"

	runtimetaskspec "github.com/mwroh/sheet-ops/runtime/taskspec"
)

func resolveRequestText(taskSpec runtimetaskspec.TaskSpec) (string, error) {
	if strings.TrimSpace(taskSpec.RequestText) != "" {
		return taskSpec.RequestText, nil
	}
	if taskSpec.Source.Kind != "case_markdown" {
		return "", nil
	}

	raw, err := os.ReadFile(taskSpec.Source.Path)
	if err != nil {
		return "", fmt.Errorf("read case markdown request %q: %w", taskSpec.Source.Path, err)
	}
	return string(raw), nil
}
