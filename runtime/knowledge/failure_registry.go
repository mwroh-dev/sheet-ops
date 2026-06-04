package knowledge

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	VerificationSemanticNamespace   = "knowledge/verification/semantic"
	defaultFailureRegistryThreshold = 2
)

type SemanticRecord struct {
	AgentNamespace   string   `json:"agent_namespace"`
	RecordID         string   `json:"record_id"`
	Version          int      `json:"version"`
	Replaces         string   `json:"replaces,omitempty"`
	Strategy         string   `json:"strategy"`
	UpdatedAfterRuns []string `json:"updated_after_runs,omitempty"`
	Summary          string   `json:"summary"`
	FailureClass     string   `json:"failure_class,omitempty"`
	DomainCode       string   `json:"domain_code,omitempty"`
	CorrectiveRule   string   `json:"corrective_rule,omitempty"`
	RepairHint       string   `json:"repair_hint,omitempty"`
	EvidenceRuns     []string `json:"evidence_runs,omitempty"`
	PromotionReason  string   `json:"promotion_reason,omitempty"`
}

type FailureRegistryPromotionOptions struct {
	Threshold int
}

type FailureRegistryPromotionResult struct {
	SourcePath   string
	RegistryPath string
	Promoted     []SemanticRecord
}

func PromoteFailureRegistry(options FailureRegistryPromotionOptions) (FailureRegistryPromotionResult, error) {
	sourcePath := filepath.Join(KnowledgeRoot(), "verification", "episodic", "failures", "records.jsonl")
	registryPath := filepath.Join(KnowledgeRoot(), "verification", "semantic", "records.jsonl")
	threshold := options.Threshold
	if threshold <= 0 {
		threshold = defaultFailureRegistryThreshold
	}

	episodes, err := readFailureRegistryEpisodes(sourcePath)
	if err != nil {
		return FailureRegistryPromotionResult{}, err
	}
	existing, err := readSemanticRecords(registryPath)
	if err != nil {
		return FailureRegistryPromotionResult{}, err
	}

	promoted := buildFailureRegistryPromotions(episodes, threshold)
	combined := mergeSemanticRecords(existing, promoted)
	if err := writeSemanticRegistry(registryPath, combined); err != nil {
		return FailureRegistryPromotionResult{}, err
	}

	return FailureRegistryPromotionResult{
		SourcePath:   sourcePath,
		RegistryPath: registryPath,
		Promoted:     promoted,
	}, nil
}

func readFailureRegistryEpisodes(path string) ([]EpisodicRecord, error) {
	records, err := readEpisodicRecords(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}

	schemaPath := repoJoin("contracts", "knowledge", "episodic_record.schema.json")
	for _, record := range records {
		if err := validateStruct(schemaPath, record); err != nil {
			return nil, err
		}
	}
	return records, nil
}

func buildFailureRegistryPromotions(records []EpisodicRecord, threshold int) []SemanticRecord {
	type promotionGroup struct {
		FailureClass   string
		DomainCode     string
		RuleCandidate  string
		RepairHint     string
		EvidenceRunSet map[string]struct{}
	}

	groups := make(map[string]*promotionGroup)
	for _, record := range records {
		if strings.TrimSpace(record.AgentNamespace) != VerificationFailureNamespace {
			continue
		}
		runID := strings.TrimSpace(record.RunID)
		if runID == "" {
			continue
		}

		ruleCandidate := failureRegistryRuleCandidate(record)
		repairHint := normalizeSentenceFragment(record.RepairHint)
		key := failureRegistryGroupKey(record.FailureClass, record.DomainCode, ruleCandidate, repairHint)
		group, ok := groups[key]
		if !ok {
			group = &promotionGroup{
				FailureClass:   strings.TrimSpace(record.FailureClass),
				DomainCode:     strings.TrimSpace(record.DomainCode),
				RuleCandidate:  ruleCandidate,
				RepairHint:     repairHint,
				EvidenceRunSet: make(map[string]struct{}),
			}
			groups[key] = group
		}
		group.EvidenceRunSet[runID] = struct{}{}
	}

	keys := make([]string, 0, len(groups))
	for key, group := range groups {
		if len(group.EvidenceRunSet) >= threshold {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)

	promoted := make([]SemanticRecord, 0, len(keys))
	for _, key := range keys {
		group := groups[key]
		evidenceRuns := mapKeys(group.EvidenceRunSet)
		correctiveRule := failureRegistryCorrectiveRule(group.RuleCandidate, group.DomainCode, group.RepairHint)
		promoted = append(promoted, SemanticRecord{
			AgentNamespace:   VerificationSemanticNamespace,
			RecordID:         failureRegistryRecordID(group.FailureClass, group.DomainCode, group.RuleCandidate, group.RepairHint),
			Version:          1,
			Strategy:         "failure_registry",
			UpdatedAfterRuns: evidenceRuns,
			Summary:          correctiveRule,
			FailureClass:     group.FailureClass,
			DomainCode:       group.DomainCode,
			CorrectiveRule:   correctiveRule,
			RepairHint:       group.RepairHint,
			EvidenceRuns:     evidenceRuns,
			PromotionReason: fmt.Sprintf(
				"Promoted after %d independent runs with the same failure_class, domain_code, and corrective rule candidate.",
				len(evidenceRuns),
			),
		})
	}
	return promoted
}

func mergeSemanticRecords(existing, promoted []SemanticRecord) []SemanticRecord {
	combined := make([]SemanticRecord, 0, len(existing)+len(promoted))
	for _, record := range existing {
		if strings.TrimSpace(record.Strategy) == "failure_registry" {
			continue
		}
		combined = append(combined, record)
	}
	combined = append(combined, promoted...)
	sort.Slice(combined, func(i, j int) bool {
		return combined[i].RecordID < combined[j].RecordID
	})
	return combined
}

func writeSemanticRegistry(path string, records []SemanticRecord) error {
	if len(records) == 0 {
		if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
		return nil
	}

	schemaPath := repoJoin("contracts", "knowledge", "semantic_record.schema.json")
	for _, record := range records {
		if err := validateStruct(schemaPath, record); err != nil {
			return err
		}
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()

	encoder := json.NewEncoder(file)
	for _, record := range records {
		if err := encoder.Encode(record); err != nil {
			return err
		}
	}
	return nil
}

func readSemanticRecords(path string) ([]SemanticRecord, error) {
	file, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	defer func() { _ = file.Close() }()

	var records []SemanticRecord
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var record SemanticRecord
		if err := json.Unmarshal(scanner.Bytes(), &record); err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return records, nil
}

func failureRegistryRuleCandidate(record EpisodicRecord) string {
	step := normalizeSentenceFragment(record.CriticalStep)
	if step != "" {
		return step
	}
	return normalizeSentenceFragment(fmt.Sprintf("%s %s", record.FailureClass, record.DomainCode))
}

func failureRegistryGroupKey(failureClass, domainCode, ruleCandidate, repairHint string) string {
	return strings.Join([]string{
		strings.TrimSpace(failureClass),
		strings.TrimSpace(domainCode),
		strings.TrimSpace(ruleCandidate),
		strings.TrimSpace(repairHint),
	}, "\x1f")
}

func failureRegistryCorrectiveRule(ruleCandidate, domainCode, repairHint string) string {
	if strings.TrimSpace(repairHint) != "" {
		return fmt.Sprintf(
			"The critical step %q must complete without %s; repair by %s.",
			ruleCandidate,
			strings.TrimSpace(domainCode),
			strings.TrimSpace(repairHint),
		)
	}
	return fmt.Sprintf(
		"The critical step %q must complete without %s.",
		ruleCandidate,
		strings.TrimSpace(domainCode),
	)
}

func failureRegistryRecordID(failureClass, domainCode, ruleCandidate, repairHint string) string {
	parts := []string{
		"failure_registry",
		sanitizeRecordIDPart(failureClass),
		sanitizeRecordIDPart(domainCode),
		sanitizeRecordIDPart(ruleCandidate),
	}
	if strings.TrimSpace(repairHint) != "" {
		parts = append(parts, sanitizeRecordIDPart(repairHint))
	}
	return strings.Join(parts, "/")
}

func sanitizeRecordIDPart(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	var builder strings.Builder
	lastUnderscore := false
	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			builder.WriteRune(r)
			lastUnderscore = false
			continue
		}
		if lastUnderscore {
			continue
		}
		builder.WriteByte('_')
		lastUnderscore = true
	}
	out := strings.Trim(builder.String(), "_")
	if out == "" {
		return "unknown"
	}
	return out
}

func normalizeSentenceFragment(value string) string {
	value = strings.Join(strings.Fields(strings.TrimSpace(value)), " ")
	value = strings.TrimRight(value, ".!?")
	return value
}

func mapKeys(values map[string]struct{}) []string {
	out := make([]string, 0, len(values))
	for value := range values {
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}
