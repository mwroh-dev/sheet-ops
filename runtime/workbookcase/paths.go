package workbookcase

import (
	"encoding/json"
	"os"
	"path/filepath"
	goruntime "runtime"
	"strings"
	"time"

	"github.com/mwroh/sheet-ops/runtime/pathid"
	runtimeschema "github.com/mwroh/sheet-ops/runtime/schema"
)

const ArtifactRootEnv = "SHEET_OPS_ARTIFACT_ROOT"

func NewRunIDs() RunIDs {
	stamp := time.Now().UTC().Format("20060102T150405.000000000Z")
	return RunIDs{
		RunID:   "run-" + stamp,
		TraceID: "trace-" + stamp,
	}
}

func artifactRoot() string {
	root := strings.TrimSpace(os.Getenv(ArtifactRootEnv))
	if root != "" {
		return filepath.Clean(root)
	}
	return repoJoin("artifacts")
}

func BuildRunPaths(scenarioID string, ids RunIDs) (RunPaths, error) {
	scenarioID, err := pathid.Validate("scenario id", scenarioID)
	if err != nil {
		return RunPaths{}, err
	}
	root := artifactRoot()
	telemetryDir := filepath.Join(root, "telemetry", "runs", ids.RunID)
	evidenceDir := filepath.Join(root, "evidence", "scenarios", scenarioID, "runs", ids.RunID)
	reportDir := filepath.Join(root, "reports")
	return RunPaths{
		TelemetryDir:           telemetryDir,
		EvidenceDir:            evidenceDir,
		ReportDir:              reportDir,
		RequestPath:            filepath.Join(evidenceDir, "request.md"),
		InspectionPath:         filepath.Join(evidenceDir, "input-summary.json"),
		PlanPath:               filepath.Join(evidenceDir, "plan.json"),
		PolicyPath:             filepath.Join(evidenceDir, "policy-decision.json"),
		TaskSpecPath:           filepath.Join(evidenceDir, "task-spec.json"),
		OperationIRPath:        filepath.Join(evidenceDir, "operation-ir.json"),
		EpisodePath:            filepath.Join(evidenceDir, "execution-episode.json"),
		ExecutionPath:          filepath.Join(evidenceDir, "execution-summary.json"),
		VerificationPath:       filepath.Join(evidenceDir, "verification-summary.json"),
		VerificationReviewPath: filepath.Join(evidenceDir, "verification-review.json"),
		RenderPath:             filepath.Join(evidenceDir, "render-artifact.json"),
		OutcomePath:            filepath.Join(evidenceDir, "outcome.json"),
		FailureEvidencePath:    filepath.Join(evidenceDir, "failure-evidence.json"),
		RepairAdvicePath:       filepath.Join(evidenceDir, "repair-advice.json"),
		EvidenceIndex:          filepath.Join(evidenceDir, "telemetry-links.json"),
		ReportPath:             filepath.Join(reportDir, ids.RunID+"-summary.md"),
	}, nil
}

func EnsureRunPaths(paths RunPaths) error {
	for _, dir := range []string{paths.TelemetryDir, paths.EvidenceDir, paths.ReportDir} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	return nil
}

func AppendTelemetry(paths RunPaths, category string, event TelemetryEvent) error {
	if !shouldPersistSensitiveWorkbookArtifacts() {
		return nil
	}
	return appendJSONL(filepath.Join(paths.TelemetryDir, category+".jsonl"), event)
}

func writeJSON(path string, value any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(sanitizeWorkbookArtifactValue(value), "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(raw, '\n'), 0o644)
}

func writeText(path, value string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(value), 0o644)
}

func appendJSONL(path string, value any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return err
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()
	if _, err := file.Write(append(raw, '\n')); err != nil {
		return err
	}
	return nil
}

func repoJoin(parts ...string) string {
	_, file, _, ok := goruntime.Caller(0)
	if !ok {
		return filepath.Join(parts...)
	}
	return runtimeschema.ResolveRepoPath(file, 2, parts...)
}

func fileExists(path string) bool {
	if path == "" {
		return false
	}
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}
