package requestcompiler

import (
	"errors"
	"os"
	"slices"
	"strings"

	runtimeinspect "github.com/mwroh/sheet-ops/runtime/inspect"
	runtimevalidate "github.com/mwroh/sheet-ops/runtime/validate"
	"github.com/xuri/excelize/v2"
)

type Result struct {
	Decision                  Decision
	Validation                runtimevalidate.Result
	ValidatedExecutionRequest *runtimevalidate.ValidatedExecutionRequest
	MemoryMatchSummary        *MemoryMatchSummary
}

func Compile(input Input) (Result, error) {
	intent, err := interpreter.Interpret(input)
	if err != nil {
		return Result{}, err
	}
	intent, memorySummary, err := applyDeterministicIntentSignals(input, intent)
	if err != nil {
		return Result{}, err
	}
	result, err := ValidateIntent(input, intent)
	if err != nil {
		return Result{}, err
	}
	result.MemoryMatchSummary = memorySummary
	return result, nil
}

func applyDeterministicIntentSignals(input Input, intent NormalizedIntent) (NormalizedIntent, *MemoryMatchSummary, error) {
	intent = normalizeIntent(intent)

	requestText, err := loadRequestText(input.RequestSource)
	if err != nil {
		return NormalizedIntent{}, nil, err
	}
	primaryInput, err := primaryWorkbookPath(input.InputWorkbooks)
	if err != nil {
		return NormalizedIntent{}, nil, err
	}
	facts, err := runtimeinspect.InspectWorkbookFacts(primaryInput)
	if err != nil {
		return NormalizedIntent{}, nil, err
	}
	intent = augmentIntentFromWorkbookFacts(requestText, facts, intent)
	memorySummary, intent, err := applyPreferenceMemory(input, requestText, facts, intent)
	if err != nil {
		return NormalizedIntent{}, nil, err
	}

	if !slices.Contains(intent.Ambiguity.Markers, IntentMarkerSourceTruthConflictRequested) {
		return intent, memorySummary, nil
	}

	intent.Ambiguity.Markers = slices.DeleteFunc(intent.Ambiguity.Markers, func(ambiguity string) bool {
		return ambiguity == IntentMarkerSourceTruthConflictRequested
	})

	hasSourceTruthConflict, err := workbookHasSourceTruthConflict(primaryInput)
	if err != nil {
		return NormalizedIntent{}, nil, err
	}
	if !hasSourceTruthConflict {
		intent.Ambiguities = append([]string(nil), intent.Ambiguity.Markers...)
		return intent, memorySummary, nil
	}

	intent.SourceSheetCandidates = appendUnique(intent.SourceSheetCandidates, "Orders_RAW", "Payments_RAW")
	intent.Ambiguity.Markers = appendUnique(intent.Ambiguity.Markers, IntentMarkerSourceTruthConflict)
	intent.Ambiguities = append([]string(nil), intent.Ambiguity.Markers...)
	return intent, memorySummary, nil
}

func loadRequestText(source RequestSource) (string, error) {
	switch source.Kind {
	case RequestSourceDirectText:
		return trimRequestText(source.Text), nil
	case RequestSourcePromptFile:
		return loadPromptFile(source.Path)
	default:
		return "", errors.New("unknown request source kind")
	}
}

func primaryWorkbookPath(inputs []WorkbookInput) (string, error) {
	for _, input := range inputs {
		if input.Role == "primary_input" {
			return input.Path, nil
		}
	}
	return "", errors.New("missing primary workbook")
}

func loadPromptFile(path string) (string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return trimRequestText(string(raw)), nil
}

func workbookHasSourceTruthConflict(path string) (bool, error) {
	file, err := excelize.OpenFile(path)
	if err != nil {
		return false, err
	}
	defer file.Close()

	orders, err := sheetAmountsByOrderNumber(file, "Orders_RAW")
	if err != nil {
		var sheetErr excelize.ErrSheetNotExist
		if errors.As(err, &sheetErr) {
			return false, nil
		}
		return false, err
	}
	payments, err := sheetAmountsByOrderNumber(file, "Payments_RAW")
	if err != nil {
		var sheetErr excelize.ErrSheetNotExist
		if errors.As(err, &sheetErr) {
			return false, nil
		}
		return false, err
	}

	for orderID, orderAmount := range orders {
		paymentAmount, ok := payments[orderID]
		if ok && paymentAmount != orderAmount {
			return true, nil
		}
	}
	return false, nil
}

func sheetAmountsByOrderNumber(file *excelize.File, sheet string) (map[string]string, error) {
	rows, err := file.GetRows(sheet)
	if err != nil {
		return nil, err
	}
	amounts := make(map[string]string, len(rows))
	for rowIndex, row := range rows {
		if rowIndex == 0 || len(row) < 2 {
			continue
		}
		amounts[row[0]] = row[1]
	}
	return amounts, nil
}

func trimRequestText(raw string) string {
	return strings.TrimSpace(raw)
}
