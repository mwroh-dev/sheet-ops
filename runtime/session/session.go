package session

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type RunSession struct {
	IDs
	ScenarioID string
	Paths
}

type sessionRecord struct {
	SessionID   string `json:"session_id"`
	RunID       string `json:"run_id"`
	TraceID     string `json:"trace_id"`
	ScenarioID  string `json:"scenario_id"`
	OpenedAt    string `json:"opened_at"`
	SessionRoot string `json:"session_root"`
	EvidenceDir string `json:"evidence_dir"`
}

func OpenSession(caseRoot, scenarioID string) (RunSession, error) {
	return OpenSessionWithIDs(caseRoot, scenarioID, NewIDs())
}

func OpenSessionWithIDs(caseRoot, scenarioID string, ids IDs) (RunSession, error) {
	caseRoot = strings.TrimSpace(caseRoot)
	scenarioID = strings.TrimSpace(scenarioID)
	ids.SessionID = strings.TrimSpace(ids.SessionID)
	ids.RunID = strings.TrimSpace(ids.RunID)
	ids.TraceID = strings.TrimSpace(ids.TraceID)
	if caseRoot == "" {
		return RunSession{}, errors.New("case root required")
	}
	if ids.TraceID == "" {
		return RunSession{}, errors.New("trace id required")
	}
	paths, err := BuildPaths(caseRoot, scenarioID, ids)
	if err != nil {
		return RunSession{}, err
	}

	runSession := RunSession{
		IDs:        ids,
		ScenarioID: scenarioID,
		Paths:      paths,
	}
	if err := ensureSessionDirs(runSession); err != nil {
		return RunSession{}, err
	}
	if err := writeJSON(runSession.SessionPath, sessionRecord{
		SessionID:   runSession.SessionID,
		RunID:       runSession.RunID,
		TraceID:     runSession.TraceID,
		ScenarioID:  runSession.ScenarioID,
		OpenedAt:    nowUTC().Format(time.RFC3339Nano),
		SessionRoot: runSession.SessionRoot,
		EvidenceDir: runSession.EvidenceDir,
	}); err != nil {
		return RunSession{}, err
	}
	if err := ensureFile(runSession.TransitionLogPath); err != nil {
		return RunSession{}, err
	}
	return runSession, nil
}

func ensureSessionDirs(runSession RunSession) error {
	for _, dir := range []string{
		runSession.SessionRoot,
		runSession.TelemetryDir,
		runSession.ProofsDir,
		runSession.OutputsDir,
		runSession.EvidenceDir,
	} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	return nil
}

func writeJSON(path string, value any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(raw, '\n'), 0o644)
}

func ensureFile(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	file, err := os.OpenFile(path, os.O_CREATE, 0o644)
	if err != nil {
		return err
	}
	return file.Close()
}
