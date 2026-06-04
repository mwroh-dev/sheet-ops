package requestcompiler

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type PersistedResult struct {
	Result
	WorkUnitID string
	RequestDir string
}

func (result PersistedResult) DecisionPath() string {
	return filepath.Join(result.RequestDir, "compiler_decision.json")
}

func (result PersistedResult) GeneratedRequestPath() string {
	if result.ValidatedExecutionRequest == nil {
		return ""
	}
	return filepath.Join(result.RequestDir, "validated_execution_request.json")
}

func RequestArtifactDir(workspaceRoot, workUnitID string) (string, error) {
	workUnitID = strings.TrimSpace(workUnitID)
	if workUnitID == "" {
		return "", fmt.Errorf("work unit id required")
	}
	if filepath.IsAbs(workUnitID) || workUnitID == "." || workUnitID == ".." || strings.ContainsAny(workUnitID, `/\`) {
		return "", fmt.Errorf("work unit id must stay within request artifact root")
	}
	if len(workUnitID) > 192 {
		return "", fmt.Errorf("work unit id exceeds length limit")
	}
	workRoot := filepath.Join(filepath.Clean(workspaceRoot), ".sheet-ops-state", "artifacts", "work")
	requestDir := filepath.Join(workRoot, workUnitID, "request-compiler")
	if !pathWithin(requestDir, workRoot) {
		return "", fmt.Errorf("work unit id must stay within request artifact root")
	}
	return requestDir, nil
}

func pathWithin(path, root string) bool {
	path = filepath.Clean(path)
	root = filepath.Clean(root)
	if path == root {
		return true
	}
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return false
	}
	return rel != "." && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

func CompileAndPersist(input Input) (PersistedResult, error) {
	workUnitID, err := NewWorkUnitID(input.ScenarioSlug, time.Now().UTC())
	if err != nil {
		return PersistedResult{}, err
	}
	return compileAndPersist(input, workUnitID)
}

func CompileAndPersistWithWorkUnitID(input Input, workUnitID string) (PersistedResult, error) {
	return compileAndPersist(input, workUnitID)
}

func compileAndPersist(input Input, workUnitID string) (PersistedResult, error) {
	intent, err := interpreter.Interpret(input)
	if err != nil {
		return PersistedResult{}, err
	}
	intent, memorySummary, err := applyDeterministicIntentSignals(input, intent)
	if err != nil {
		return PersistedResult{}, err
	}
	intent = normalizeIntent(intent)

	result, err := ValidateIntent(input, intent)
	if err != nil {
		return PersistedResult{}, err
	}
	result.MemoryMatchSummary = memorySummary

	requestDir, err := RequestArtifactDir(input.WorkspaceRoot, workUnitID)
	if err != nil {
		return PersistedResult{}, err
	}
	requestParentDir := filepath.Dir(requestDir)
	if err := os.MkdirAll(requestParentDir, 0o755); err != nil {
		return PersistedResult{}, err
	}
	stagingDir, err := os.MkdirTemp(requestParentDir, ".request-compiler-staging-*")
	if err != nil {
		return PersistedResult{}, err
	}
	persistSucceeded := false
	defer func() {
		if !persistSucceeded {
			_ = os.RemoveAll(stagingDir)
		}
	}()

	if shouldPersistSensitiveCompilerArtifacts() {
		if err := writeJSON(filepath.Join(stagingDir, "compiler_input.json"), compilerInputArtifact(input, workUnitID)); err != nil {
			return PersistedResult{}, err
		}
	}
	if shouldPersistSensitiveCompilerArtifacts() {
		if err := writeFileSync(filepath.Join(stagingDir, "report.md"), []byte(buildCompileReport(result))); err != nil {
			return PersistedResult{}, err
		}
	}
	if err := writeJSON(filepath.Join(stagingDir, "normalized_intent.json"), intent); err != nil {
		return PersistedResult{}, err
	}
	if err := writeJSON(filepath.Join(stagingDir, "compiler_decision.json"), compilerDecisionArtifact(result)); err != nil {
		return PersistedResult{}, err
	}
	if result.Validation.Status == "compiled" && result.ValidatedExecutionRequest != nil {
		validationResult := any(result.Validation)
		validatedExecutionRequest := any(result.ValidatedExecutionRequest)
		if !shouldPersistSensitiveCompilerArtifacts() {
			validationResult = sanitizeValidationResult(result.Validation)
			validatedExecutionRequest = sanitizeValidatedExecutionRequest(result.ValidatedExecutionRequest)
		}
		if err := writeJSON(filepath.Join(stagingDir, "validation_result.json"), validationResult); err != nil {
			return PersistedResult{}, err
		}
		if err := writeJSON(filepath.Join(stagingDir, "validated_execution_request.json"), validatedExecutionRequest); err != nil {
			return PersistedResult{}, err
		}
	}
	if err := syncDir(stagingDir); err != nil {
		return PersistedResult{}, err
	}
	if err := os.Rename(stagingDir, requestDir); err != nil {
		return PersistedResult{}, err
	}
	if err := syncDir(requestDir); err != nil {
		return PersistedResult{}, err
	}
	if err := syncDir(requestParentDir); err != nil {
		return PersistedResult{}, err
	}
	persistSucceeded = true

	return PersistedResult{
		Result:     result,
		WorkUnitID: workUnitID,
		RequestDir: requestDir,
	}, nil
}

func compilerInputArtifact(input Input, workUnitID string) map[string]any {
	requestSource := map[string]any{
		"kind": string(input.RequestSource.Kind),
	}
	switch input.RequestSource.Kind {
	case RequestSourcePromptFile:
		requestSource["path"] = input.RequestSource.Path
	case RequestSourceDirectText:
		requestSource["text"] = input.RequestSource.Text
	}

	workbooks := make([]map[string]any, 0, len(input.InputWorkbooks))
	for _, workbook := range input.InputWorkbooks {
		workbooks = append(workbooks, map[string]any{
			"path": workbook.Path,
			"role": workbook.Role,
		})
	}

	return map[string]any{
		"request_source": requestSource,
		"workspace": map[string]any{
			"root": input.WorkspaceRoot,
			"cwd":  input.WorkingDir,
		},
		"scenario": map[string]any{
			"slug":         input.ScenarioSlug,
			"work_unit_id": workUnitID,
		},
		"inputs": map[string]any{
			"workbooks": workbooks,
		},
		"requested_output": map[string]any{
			"output_file": input.OutputFile,
		},
	}
}

func compilerDecisionArtifact(result Result) map[string]any {
	decision := result.Decision
	artifact := map[string]any{
		"status": string(decision.Status),
	}
	if len(decision.Notes) > 0 {
		artifact["notes"] = append([]string(nil), decision.Notes...)
	}
	if decision.Status == StatusNeedsHumanCheckpoint {
		artifact["checkpoint"] = map[string]any{
			"kind":     decision.Checkpoint.Kind,
			"question": decision.Checkpoint.Question,
			"options":  append([]string(nil), decision.Checkpoint.Options...),
		}
	}
	if decision.Status == StatusBlocked {
		reasonCodes := append([]string(nil), decision.StructuralSignals...)
		if len(reasonCodes) == 0 {
			reasonCodes = []string{IntentMarkerUnsupportedRequest}
		}
		artifact["blocked"] = map[string]any{
			"reason_codes": reasonCodes,
		}
	}
	if summary := result.MemoryMatchSummary; summary != nil && summary.hasEntries() {
		artifact["memory"] = map[string]any{
			"matched_paragraph_count": summary.MatchedParagraphCount,
			"applied_defaults":        append([]string(nil), summary.AppliedDefaults...),
			"ignored_preferences":     append([]string(nil), summary.IgnoredPreferences...),
			"conflict_notes":          append([]string(nil), summary.ConflictNotes...),
		}
	}

	return artifact
}

func buildCompileReport(result Result) string {
	files := []string{
		"- `compiler_input.json`",
		"- `normalized_intent.json`",
		"- `compiler_decision.json`",
	}
	if result.Validation.Status == "compiled" && result.ValidatedExecutionRequest != nil {
		files = append(files, "- `validation_result.json`", "- `validated_execution_request.json`")
	}
	report := "# Request Compiler Artifacts\n\nStatus: " + string(result.Decision.Status) + "\n\n## Files\n" + joinArtifactLines(files) + "\n"
	if summary := result.MemoryMatchSummary; summary != nil && summary.hasEntries() {
		report += "\n## Memory Influence\n"
		report += fmt.Sprintf("Matched paragraphs: %d\n", summary.MatchedParagraphCount)
		if len(summary.AppliedDefaults) > 0 {
			report += "\nApplied defaults:\n" + prefixedLines(summary.AppliedDefaults) + "\n"
		}
		if len(summary.IgnoredPreferences) > 0 {
			report += "\nIgnored preferences:\n" + prefixedLines(summary.IgnoredPreferences) + "\n"
		}
		if len(summary.ConflictNotes) > 0 {
			report += "\nConflict notes:\n" + prefixedLines(summary.ConflictNotes) + "\n"
		}
	}
	return report
}

func writeJSON(path string, value any) error {
	raw, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return writeFileSync(path, append(raw, '\n'))
}

func joinArtifactLines(values []string) string {
	switch len(values) {
	case 0:
		return ""
	case 1:
		return values[0]
	}

	total := 0
	for _, value := range values {
		total += len(value)
	}
	total += len("\n") * (len(values) - 1)

	builder := make([]byte, 0, total)
	for index, value := range values {
		if index > 0 {
			builder = append(builder, '\n')
		}
		builder = append(builder, value...)
	}
	return string(builder)
}

func prefixedLines(values []string) string {
	lines := make([]string, 0, len(values))
	for _, value := range values {
		lines = append(lines, "- "+value)
	}
	return joinArtifactLines(lines)
}

func writeFileSync(path string, body []byte) (err error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer func() {
		closeErr := file.Close()
		if err == nil {
			err = closeErr
		}
	}()

	if _, err := file.Write(body); err != nil {
		return err
	}
	if err := file.Sync(); err != nil {
		return err
	}
	return nil
}

func syncDir(path string) (err error) {
	dir, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() {
		closeErr := dir.Close()
		if err == nil {
			err = closeErr
		}
	}()

	if err := dir.Sync(); err != nil {
		return err
	}
	return nil
}
