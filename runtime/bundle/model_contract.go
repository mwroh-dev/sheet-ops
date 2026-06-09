package bundle

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	goruntime "runtime"
	"strings"

	runtimeschema "github.com/mwroh/sheet-ops/runtime/schema"
)

type RoleContract struct {
	Model           string `json:"model"`
	ReasoningEffort string `json:"reasoning_effort"`
	Carrier         string `json:"carrier"`
	Prompt          string `json:"prompt"`
	Fallback        string `json:"fallback"`
}

type ModelContract struct {
	Orchestrator    RoleContract `json:"orchestrator"`
	RequestCompiler RoleContract `json:"request_compiler"`
	ResultVerifier  RoleContract `json:"result_verifier"`
	RepairAdvisor   RoleContract `json:"repair_advisor"`
}

func LoadModelContract(path string) (ModelContract, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return ModelContract{}, err
	}

	return LoadModelContractFromBytes(raw)
}

func LoadModelContractFromBytes(raw []byte) (ModelContract, error) {
	var document any
	if err := json.Unmarshal(raw, &document); err != nil {
		return ModelContract{}, err
	}
	if err := validateModelContractDocument(document); err != nil {
		return ModelContract{}, err
	}

	var contract ModelContract
	if err := json.Unmarshal(raw, &contract); err != nil {
		return ModelContract{}, err
	}
	if err := validateModelContract(contract); err != nil {
		return ModelContract{}, err
	}

	return contract, nil
}

func validateModelContract(contract ModelContract) error {
	for role, spec := range map[string]RoleContract{
		"orchestrator":     contract.Orchestrator,
		"request_compiler": contract.RequestCompiler,
		"result_verifier":  contract.ResultVerifier,
		"repair_advisor":   contract.RepairAdvisor,
	} {
		if strings.TrimSpace(spec.ReasoningEffort) == "" || strings.TrimSpace(spec.Carrier) == "" || strings.TrimSpace(spec.Prompt) == "" {
			return fmt.Errorf("%s contract is incomplete", role)
		}
		if spec.Fallback != "forbidden" {
			return fmt.Errorf("%s fallback must be forbidden", role)
		}
	}

	return nil
}

func validateModelContractDocument(value any) error {
	return runtimeschema.ValidateStruct(modelContractSchemaPath(), value)
}

func modelContractSchemaPath() string {
	_, file, _, ok := goruntime.Caller(0)
	if !ok {
		return filepath.Join("contracts", "bundle", "model_contract.schema.json")
	}
	return runtimeschema.ResolveRepoPath(file, 2, "contracts", "bundle", "model_contract.schema.json")
}
