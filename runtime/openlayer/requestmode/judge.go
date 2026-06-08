package requestmode

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	goruntime "runtime"
	"strings"

	runtimeschema "github.com/mwroh/sheet-ops/runtime/schema"
)

const (
	ModeStructuredUseRequest     = "structured_use_request"
	ModePromptText               = "prompt_text"
	ModeOrganismExecutionRequest = "organism_execution_request"
)

type Judgment struct {
	RequestMode string `json:"request_mode"`
	Reason      string `json:"reason"`
}

type RequestRef struct {
	Kind string `json:"kind"`
	Path string `json:"path"`
}

func JudgeRequestRef(ref RequestRef) (Judgment, error) {
	if err := runtimeschema.ValidateStruct(requestRefSchemaPath(), ref); err != nil {
		return Judgment{}, err
	}

	if ref.Kind == ModePromptText {
		if _, err := readRequestFile(ref.Path); err != nil {
			return Judgment{}, err
		}
		return Validate(Judgment{
			RequestMode: ModePromptText,
			Reason:      "request_ref kind explicitly keeps the request on the prompt_text path",
		})
	}

	raw, err := readRequestFile(ref.Path)
	if err != nil {
		return Judgment{}, err
	}

	var document any
	if err := json.Unmarshal(raw, &document); err != nil {
		return Judgment{}, err
	}

	switch ref.Kind {
	case ModeStructuredUseRequest:
		if err := runtimeschema.ValidateStruct(structuredUseRequestSchemaPath(), document); err != nil {
			return Judgment{}, err
		}
		return Validate(Judgment{
			RequestMode: ModeStructuredUseRequest,
			Reason:      "request_ref kind is structured_use_request and the file satisfies the structured_use_request contract",
		})
	case ModeOrganismExecutionRequest:
		if err := runtimeschema.ValidateStruct(organismExecutionRequestSchemaPath(), document); err != nil {
			return Judgment{}, err
		}
		return Validate(Judgment{
			RequestMode: ModeOrganismExecutionRequest,
			Reason:      "request_ref kind is organism_execution_request and the file satisfies the organism_execution_request contract",
		})
	default:
		return Judgment{}, fmt.Errorf("unsupported request_ref kind %q", ref.Kind)
	}
}

func readRequestFile(path string) ([]byte, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read request file %s: %w", path, err)
	}

	trimmed := strings.TrimSpace(string(raw))
	if trimmed == "" {
		return nil, fmt.Errorf("request file %s is empty", path)
	}

	return raw, nil
}

func Validate(j Judgment) (Judgment, error) {
	if err := runtimeschema.ValidateStruct(requestModeJudgmentSchemaPath(), j); err != nil {
		return Judgment{}, err
	}
	return j, nil
}

func requestModeJudgmentSchemaPath() string {
	_, file, _, ok := goruntime.Caller(0)
	if !ok {
		return filepath.Join("contracts", "requests", "request_mode_judgment.schema.json")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", "..", "contracts", "requests", "request_mode_judgment.schema.json"))
}

func requestRefSchemaPath() string {
	_, file, _, ok := goruntime.Caller(0)
	if !ok {
		return filepath.Join("contracts", "requests", "request_ref.schema.json")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", "..", "contracts", "requests", "request_ref.schema.json"))
}

func structuredUseRequestSchemaPath() string {
	_, file, _, ok := goruntime.Caller(0)
	if !ok {
		return filepath.Join("contracts", "requests", "structured_use_request.schema.json")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", "..", "contracts", "requests", "structured_use_request.schema.json"))
}

func organismExecutionRequestSchemaPath() string {
	_, file, _, ok := goruntime.Caller(0)
	if !ok {
		return filepath.Join("contracts", "requests", "organism_execution_request.schema.json")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", "..", "contracts", "requests", "organism_execution_request.schema.json"))
}
