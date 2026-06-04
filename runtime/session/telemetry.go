package session

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type TelemetryEvent struct {
	SessionID string         `json:"session_id,omitempty"`
	RunID     string         `json:"run_id"`
	TraceID   string         `json:"trace_id"`
	EventType string         `json:"event_type"`
	Timestamp string         `json:"timestamp"`
	Component string         `json:"component"`
	Payload   map[string]any `json:"payload,omitempty"`
}

func AppendTelemetry(runSession RunSession, name string, event TelemetryEvent) error {
	name, err := validatePathIdentifier("telemetry name", name)
	if err != nil {
		return err
	}
	return appendJSONL(filepath.Join(runSession.TelemetryDir, name), event)
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
