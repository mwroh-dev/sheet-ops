package knowledge

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	runtimeschema "github.com/mwroh/sheet-ops/runtime/schema"
)

const (
	KnowledgeRootEnv              = "SHEET_OPS_KNOWLEDGE_ROOT"
	OrchestratorEpisodicNamespace = "knowledge/orchestrator/episodic"
	VerificationFailureNamespace  = "knowledge/verification/episodic/failures"
)

type EpisodicRecord struct {
	AgentNamespace string   `json:"agent_namespace"`
	RunID          string   `json:"run_id"`
	RecordedAt     string   `json:"recorded_at"`
	Summary        string   `json:"summary"`
	ScenarioID     string   `json:"scenario_id,omitempty"`
	Operation      string   `json:"operation,omitempty"`
	Outcome        string   `json:"outcome,omitempty"`
	Phase          string   `json:"phase,omitempty"`
	FailureClass   string   `json:"failure_class,omitempty"`
	DomainCode     string   `json:"domain_code,omitempty"`
	CriticalStep   string   `json:"critical_step,omitempty"`
	RepairHint     string   `json:"repair_hint,omitempty"`
	SourcePaths    []string `json:"source_paths,omitempty"`
}

type OrchestratorAppendInput struct {
	RunID      string
	ScenarioID string
	Operation  string
	Outcome    string
}

type VerificationFailureAppendInput struct {
	RunID        string
	ScenarioID   string
	Operation    string
	Outcome      string
	Phase        string
	FailureClass string
	DomainCode   string
	CriticalStep string
	RepairHint   string
}

func AppendOrchestratorEpisodicRecord(input OrchestratorAppendInput) error {
	record := EpisodicRecord{
		AgentNamespace: OrchestratorEpisodicNamespace,
		RunID:          input.RunID,
		RecordedAt:     time.Now().UTC().Format(time.RFC3339Nano),
		Summary: fmt.Sprintf(
			"Recorded because closed verification completed for harness-managed run %s (scenario=%s, operation=%s, outcome=%s).",
			input.RunID,
			input.ScenarioID,
			input.Operation,
			input.Outcome,
		),
		ScenarioID: input.ScenarioID,
		Operation:  input.Operation,
		Outcome:    input.Outcome,
	}
	return appendEpisodicRecord(filepath.Join(KnowledgeRoot(), "orchestrator", "episodic", "records.jsonl"), record)
}

func AppendVerificationFailureEpisodicRecord(input VerificationFailureAppendInput) error {
	if strings.TrimSpace(input.Phase) != "verification" {
		return fmt.Errorf("verification failure episodic record requires verification phase, got %q", input.Phase)
	}
	if strings.TrimSpace(input.Outcome) != "fail" {
		return fmt.Errorf("verification failure episodic record requires fail outcome, got %q", input.Outcome)
	}
	record := EpisodicRecord{
		AgentNamespace: VerificationFailureNamespace,
		RunID:          input.RunID,
		RecordedAt:     time.Now().UTC().Format(time.RFC3339Nano),
		Summary: fmt.Sprintf(
			"Recorded after failure evidence was durably emitted for harness-managed run %s (scenario=%s, operation=%s, phase=%s, failure_class=%s, domain_code=%s).",
			input.RunID,
			input.ScenarioID,
			input.Operation,
			input.Phase,
			input.FailureClass,
			input.DomainCode,
		),
		ScenarioID:   input.ScenarioID,
		Operation:    input.Operation,
		Outcome:      input.Outcome,
		Phase:        input.Phase,
		FailureClass: input.FailureClass,
		DomainCode:   input.DomainCode,
		CriticalStep: input.CriticalStep,
		RepairHint:   input.RepairHint,
	}
	return appendEpisodicRecord(filepath.Join(KnowledgeRoot(), "verification", "episodic", "failures", "records.jsonl"), record)
}

func appendEpisodicRecord(path string, record EpisodicRecord) error {
	if err := validateStruct(repoJoin("contracts", "knowledge", "episodic_record.schema.json"), record); err != nil {
		return err
	}
	return appendJSONL(path, record)
}

func KnowledgeRoot() string {
	if root := strings.TrimSpace(os.Getenv(KnowledgeRootEnv)); root != "" {
		return filepath.Clean(root)
	}
	return repoJoin("knowledge")
}

func compactPaths(paths []string) []string {
	if len(paths) == 0 {
		return nil
	}
	out := make([]string, 0, len(paths))
	seen := make(map[string]struct{}, len(paths))
	for _, path := range paths {
		path = strings.TrimSpace(path)
		if path == "" {
			continue
		}
		if _, ok := seen[path]; ok {
			continue
		}
		seen[path] = struct{}{}
		out = append(out, path)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func validateStruct(schemaPath string, value any) error {
	return runtimeschema.ValidateStruct(schemaPath, value)
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
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return filepath.Join(parts...)
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
	all := append([]string{root}, parts...)
	return filepath.Join(all...)
}
