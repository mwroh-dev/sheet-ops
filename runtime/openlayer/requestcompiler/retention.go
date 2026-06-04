package requestcompiler

import (
	"path/filepath"

	runtimeconfig "github.com/mwroh/sheet-ops/runtime/runtimeconfig"
	runtimevalidate "github.com/mwroh/sheet-ops/runtime/validate"
)

func shouldPersistSensitiveCompilerArtifacts() bool {
	return !runtimeconfig.UsesRedactedArtifacts()
}

func sanitizeValidatedExecutionRequest(req *runtimevalidate.ValidatedExecutionRequest) *runtimevalidate.ValidatedExecutionRequest {
	if req == nil {
		return nil
	}
	clone := *req
	clone.RequestText = runtimeconfig.RedactedString(clone.RequestText)
	clone.InputFile = runtimeconfig.UseBasename(clone.InputFile)
	clone.OutputFile = runtimeconfig.UseBasename(clone.OutputFile)
	return &clone
}

func sanitizeValidationResult(result runtimevalidate.Result) runtimevalidate.Result {
	clone := result
	clone.ValidatedExecutionRequest = sanitizeValidatedExecutionRequest(result.ValidatedExecutionRequest)
	return clone
}

func sanitizeCompilerInputArtifact(input Input, workUnitID string) map[string]any {
	requestSource := map[string]any{
		"kind": string(input.RequestSource.Kind),
	}
	if input.RequestSource.Kind == RequestSourcePromptFile {
		requestSource["path"] = filepath.Base(input.RequestSource.Path)
	}
	if input.RequestSource.Kind == RequestSourceDirectText {
		requestSource["text"] = runtimeconfig.RedactedString(input.RequestSource.Text)
	}

	workbooks := make([]map[string]any, 0, len(input.InputWorkbooks))
	for _, workbook := range input.InputWorkbooks {
		workbooks = append(workbooks, map[string]any{
			"path": runtimeconfig.UseBasename(workbook.Path),
			"role": workbook.Role,
		})
	}

	return map[string]any{
		"request_source": requestSource,
		"workspace": map[string]any{
			"root": runtimeconfig.UseBasename(input.WorkspaceRoot),
			"cwd":  runtimeconfig.UseBasename(input.WorkingDir),
		},
		"scenario": map[string]any{
			"slug":         input.ScenarioSlug,
			"work_unit_id": workUnitID,
		},
		"inputs": map[string]any{
			"workbooks": workbooks,
		},
		"requested_output": map[string]any{
			"output_file": runtimeconfig.UseBasename(input.OutputFile),
		},
	}
}
