package knowledge

import (
	"bufio"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

const maxRecentRoutingOperations = 3

type OrchestratorRoutingKnowledge struct {
	KnowledgePath    string
	RecordCount      int
	UsedFallback     bool
	FallbackReason   string
	RecentOperations []string
}

func LoadOrchestratorRoutingKnowledge() (OrchestratorRoutingKnowledge, error) {
	knowledgePath := filepath.Join(KnowledgeRoot(), "orchestrator", "episodic", "records.jsonl")
	records, err := readEpisodicRecords(knowledgePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return OrchestratorRoutingKnowledge{
				KnowledgePath:  knowledgePath,
				UsedFallback:   true,
				FallbackReason: "No orchestrator knowledge records found; routing will use request signals only.",
			}, nil
		}
		return OrchestratorRoutingKnowledge{}, err
	}
	if len(records) == 0 {
		return OrchestratorRoutingKnowledge{
			KnowledgePath:  knowledgePath,
			UsedFallback:   true,
			FallbackReason: "No orchestrator knowledge records found; routing will use request signals only.",
		}, nil
	}
	return OrchestratorRoutingKnowledge{
		KnowledgePath:    knowledgePath,
		RecordCount:      len(records),
		RecentOperations: recentRoutingOperations(records),
	}, nil
}

func readEpisodicRecords(path string) ([]EpisodicRecord, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	var records []EpisodicRecord
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var record EpisodicRecord
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

func recentRoutingOperations(records []EpisodicRecord) []string {
	var operations []string
	seen := make(map[string]struct{})
	for index := len(records) - 1; index >= 0; index-- {
		operation, err := recordOperation(records[index])
		if err != nil || operation == "" {
			continue
		}
		if _, ok := seen[operation]; ok {
			continue
		}
		seen[operation] = struct{}{}
		operations = append(operations, operation)
		if len(operations) == maxRecentRoutingOperations {
			break
		}
	}
	return operations
}

func recordOperation(record EpisodicRecord) (string, error) {
	if operation := strings.TrimSpace(record.Operation); operation != "" {
		return operation, nil
	}
	for _, path := range record.SourcePaths {
		operation, err := readOperationDocument(path)
		if err != nil {
			continue
		}
		if operation != "" {
			return operation, nil
		}
	}
	return "", nil
}

func readOperationDocument(path string) (string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	var document struct {
		Operation string `json:"operation"`
	}
	if err := json.Unmarshal(raw, &document); err != nil {
		return "", err
	}
	return strings.TrimSpace(document.Operation), nil
}
