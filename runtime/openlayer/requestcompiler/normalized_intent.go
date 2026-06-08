package requestcompiler

import (
	"encoding/json"
	"os"
	"path/filepath"
	goruntime "runtime"

	runtimeschema "github.com/mwroh/sheet-ops/runtime/schema"
)

func LoadNormalizedIntent(path string) (NormalizedIntent, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return NormalizedIntent{}, err
	}

	return LoadNormalizedIntentFromBytes(raw)
}

func LoadNormalizedIntentFromBytes(raw []byte) (NormalizedIntent, error) {

	var document any
	if err := json.Unmarshal(raw, &document); err != nil {
		return NormalizedIntent{}, err
	}
	if err := validateNormalizedIntentDocument(document); err != nil {
		return NormalizedIntent{}, err
	}

	var intent NormalizedIntent
	if err := json.Unmarshal(raw, &intent); err != nil {
		return NormalizedIntent{}, err
	}
	intent = normalizeIntent(intent)
	return intent, nil
}

func validateNormalizedIntent(intent NormalizedIntent) error {
	intent = normalizeIntent(intent)
	return validateNormalizedIntentDocument(intent)
}

func validateNormalizedIntentDocument(value any) error {
	return runtimeschema.ValidateStruct(normalizedIntentSchemaPath(), value)
}

func normalizeIntent(intent NormalizedIntent) NormalizedIntent {
	if intent.SourceSheetCandidates == nil {
		intent.SourceSheetCandidates = []string{}
	}
	if intent.LookupSheetCandidates == nil {
		intent.LookupSheetCandidates = []string{}
	}
	if intent.GroupKeys == nil {
		intent.GroupKeys = []string{}
	}
	if intent.Aggregates == nil {
		intent.Aggregates = []AggregateIntent{}
	}
	if intent.CompositionCandidates == nil {
		intent.CompositionCandidates = []string{}
	}
	if intent.Summary.Filters == nil {
		intent.Summary.Filters = []FilterIntent{}
	}
	if intent.Summary.GroupBy == nil {
		intent.Summary.GroupBy = []string{}
	}
	if intent.Summary.Metrics == nil {
		intent.Summary.Metrics = []AggregateIntent{}
	}
	if intent.JoinLookup.IncludeSourceColumns == nil {
		intent.JoinLookup.IncludeSourceColumns = []string{}
	}
	if intent.JoinLookup.AppendLookupColumns == nil {
		intent.JoinLookup.AppendLookupColumns = []string{}
	}
	if intent.AppendRows.IncludeSourceColumns == nil {
		intent.AppendRows.IncludeSourceColumns = []string{}
	}
	if intent.AppendRows.Values == nil {
		intent.AppendRows.Values = []CellValue{}
	}
	if intent.ExtendFormulas.TargetRows == nil {
		intent.ExtendFormulas.TargetRows = []int{}
	}
	if intent.ExtendFormulas.FormulaColumns == nil {
		intent.ExtendFormulas.FormulaColumns = []string{}
	}
	if intent.AddDataValidation.ValidationRule.Ranges == nil {
		intent.AddDataValidation.ValidationRule.Ranges = []string{}
	}
	if intent.AddDataValidation.ValidationRule.AllowedValues == nil {
		intent.AddDataValidation.ValidationRule.AllowedValues = []string{}
	}
	if intent.ProtectFormulaCells.ProtectionRule.FormulaRanges == nil {
		intent.ProtectFormulaCells.ProtectionRule.FormulaRanges = []string{}
	}
	if intent.ProtectFormulaCells.ProtectionRule.InputRanges == nil {
		intent.ProtectFormulaCells.ProtectionRule.InputRanges = []string{}
	}
	if intent.NormalizeHeaders.HeaderMappings == nil {
		intent.NormalizeHeaders.HeaderMappings = []HeaderMapping{}
	}
	if intent.RollForwardPeriod.CarryForwardMappings == nil {
		intent.RollForwardPeriod.CarryForwardMappings = []CarryForwardMapping{}
	}
	if intent.Ambiguities != nil {
		intent.Ambiguity.Markers = append([]string(nil), intent.Ambiguities...)
	} else if intent.Ambiguity.Markers != nil {
		intent.Ambiguities = append([]string(nil), intent.Ambiguity.Markers...)
	}
	if intent.Ambiguity.Markers == nil {
		intent.Ambiguity.Markers = []string{}
	}
	if intent.Ambiguity.UnresolvedFields == nil {
		intent.Ambiguity.UnresolvedFields = []string{}
	}
	if intent.Ambiguity.CheckpointHints == nil {
		intent.Ambiguity.CheckpointHints = []string{}
	}
	if intent.Ambiguities == nil {
		intent.Ambiguities = append([]string(nil), intent.Ambiguity.Markers...)
	}
	return intent
}

func normalizedIntentSchemaPath() string {
	workingDir, _ := os.Getwd()
	_, file, _, ok := goruntime.Caller(0)
	if !ok {
		return filepath.Join(resolvePackageRoot(workingDir, ""), "agents", "request-compiler", "contract", "normalized_intent.schema.json")
	}
	return filepath.Join(resolvePackageRoot(workingDir, file), "agents", "request-compiler", "contract", "normalized_intent.schema.json")
}
