package requestpacker

import (
	"fmt"
	"path/filepath"
	goruntime "runtime"
	"strings"

	runtimerequestmode "github.com/mwroh/sheet-ops/runtime/openlayer/requestmode"
	runtimeschema "github.com/mwroh/sheet-ops/runtime/schema"
)

type RequestRef struct {
	Kind string `json:"kind"`
	Path string `json:"path"`
}

type UseEnvelopeV2 struct {
	ScenarioID string     `json:"scenario_id"`
	Request    RequestRef `json:"request"`
	InputFile  string     `json:"input_file"`
	OutputFile string     `json:"output_file"`
}

type Input struct {
	ScenarioID  string
	RequestPath string
	InputFile   string
	OutputFile  string
	Judgment    runtimerequestmode.Judgment
}

func Pack(input Input) (UseEnvelopeV2, error) {
	if strings.TrimSpace(input.ScenarioID) == "" {
		return UseEnvelopeV2{}, fmt.Errorf("scenario_id is required")
	}
	if strings.TrimSpace(input.RequestPath) == "" {
		return UseEnvelopeV2{}, fmt.Errorf("request_path is required")
	}
	if strings.TrimSpace(input.InputFile) == "" {
		return UseEnvelopeV2{}, fmt.Errorf("input_file is required")
	}
	if strings.TrimSpace(input.OutputFile) == "" {
		return UseEnvelopeV2{}, fmt.Errorf("output_file is required")
	}
	if _, err := runtimerequestmode.Validate(input.Judgment); err != nil {
		return UseEnvelopeV2{}, err
	}

	envelope := UseEnvelopeV2{
		ScenarioID: input.ScenarioID,
		Request: RequestRef{
			Kind: input.Judgment.RequestMode,
			Path: input.RequestPath,
		},
		InputFile:  input.InputFile,
		OutputFile: input.OutputFile,
	}

	if err := runtimeschema.ValidateStruct(useEnvelopeV2SchemaPath(), envelope); err != nil {
		return UseEnvelopeV2{}, err
	}

	return envelope, nil
}

func useEnvelopeV2SchemaPath() string {
	_, file, _, ok := goruntime.Caller(0)
	if !ok {
		return filepath.Join("contracts", "requests", "use_envelope_v2.schema.json")
	}
	return runtimeschema.ResolveRepoPath(file, 3, "contracts", "requests", "use_envelope_v2.schema.json")
}
