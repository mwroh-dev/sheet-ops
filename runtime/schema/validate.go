package schema

import (
	"encoding/json"

	"github.com/santhosh-tekuri/jsonschema/v5"
)

func ValidateStruct(schemaPath string, value any) error {
	compiler := jsonschema.NewCompiler()
	schema, err := compiler.Compile(schemaPath)
	if err != nil {
		return err
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return err
	}
	var document any
	if err := json.Unmarshal(raw, &document); err != nil {
		return err
	}
	return schema.Validate(document)
}
