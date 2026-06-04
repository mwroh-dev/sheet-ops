package episode

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type ExecutionEpisode struct {
	TaskSpecPath     string `json:"task_spec_path"`
	OperationIRPath  string `json:"operation_ir_path"`
	ExecutionPath    string `json:"execution_path"`
	VerificationPath string `json:"verification_path"`
	Outcome          string `json:"outcome"`
}

type ExecutionArtifactPaths struct {
	TaskSpecPath     string
	OperationIRPath  string
	ExecutionPath    string
	VerificationPath string
}

func NewExecutionEpisode(taskSpecPath, operationIRPath, executionPath, verificationPath string, pass bool) ExecutionEpisode {
	outcome := "fail"
	if pass {
		outcome = "pass"
	}

	return ExecutionEpisode{
		TaskSpecPath:     taskSpecPath,
		OperationIRPath:  operationIRPath,
		ExecutionPath:    executionPath,
		VerificationPath: verificationPath,
		Outcome:          outcome,
	}
}

func NewExecutionEpisodeForPaths(paths ExecutionArtifactPaths, pass bool) ExecutionEpisode {
	return NewExecutionEpisode(
		paths.TaskSpecPath,
		paths.OperationIRPath,
		paths.ExecutionPath,
		paths.VerificationPath,
		pass,
	)
}

func WriteExecutionEpisode(path string, episode ExecutionEpisode) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(episode, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(raw, '\n'), 0o644)
}
