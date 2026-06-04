package session

import (
	"path/filepath"
	"strings"
	"time"

	"github.com/mwroh/sheet-ops/runtime/pathid"
)

const stateDirName = ".sheet-ops-state"

var nowUTC = func() time.Time {
	return time.Now().UTC()
}

type IDs struct {
	SessionID string
	RunID     string
	TraceID   string
}

type Paths struct {
	SessionRoot       string
	SessionPath       string
	TransitionLogPath string
	TelemetryDir      string
	ProofsDir         string
	OutputsDir        string
	EvidenceDir       string
}

func NewIDs() IDs {
	stamp := nowUTC().Format("20060102T150405.000000000Z")
	return IDs{
		SessionID: "session-" + stamp,
		RunID:     "run-" + stamp,
		TraceID:   "trace-" + stamp,
	}
}

func StateRoot(caseRoot string) string {
	return filepath.Join(filepath.Clean(strings.TrimSpace(caseRoot)), stateDirName)
}

func SessionRoot(caseRoot, sessionID string) string {
	return filepath.Join(StateRoot(caseRoot), "sessions", sessionID)
}

func EvidenceRoot(caseRoot, scenarioID, runID string) string {
	return filepath.Join(StateRoot(caseRoot), "evidence", scenarioID, runID)
}

func BuildPaths(caseRoot, scenarioID string, ids IDs) (Paths, error) {
	scenarioID, err := validatePathIdentifier("scenario id", scenarioID)
	if err != nil {
		return Paths{}, err
	}
	sessionID, err := validatePathIdentifier("session id", ids.SessionID)
	if err != nil {
		return Paths{}, err
	}
	runID, err := validatePathIdentifier("run id", ids.RunID)
	if err != nil {
		return Paths{}, err
	}

	sessionRoot := SessionRoot(caseRoot, sessionID)
	return Paths{
		SessionRoot:       sessionRoot,
		SessionPath:       filepath.Join(sessionRoot, "session.json"),
		TransitionLogPath: filepath.Join(sessionRoot, "transition-log.ndjson"),
		TelemetryDir:      filepath.Join(sessionRoot, "telemetry"),
		ProofsDir:         filepath.Join(sessionRoot, "proofs"),
		OutputsDir:        filepath.Join(sessionRoot, "outputs"),
		EvidenceDir:       EvidenceRoot(caseRoot, scenarioID, runID),
	}, nil
}

func validatePathIdentifier(name, value string) (string, error) {
	return pathid.Validate(name, value)
}
