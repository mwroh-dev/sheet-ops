package validate

import runtimeinspect "github.com/mwroh/sheet-ops/runtime/inspect"

func (validator Validator) Validate(intent NormalizedIntent, decision CompilerDecision, facts runtimeinspect.WorkbookFacts) (Result, error) {
	admitted, result, err := AdmitIntent(intent, decision)
	if err != nil || result != nil {
		return derefResult(result), err
	}

	bound, result, err := BindFacts(admitted, facts)
	if err != nil || result != nil {
		return derefResult(result), err
	}

	return validator.AdmitExecution(bound)
}

func derefResult(result *Result) Result {
	if result == nil {
		return Result{}
	}
	return *result
}
