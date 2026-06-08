package formula

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"github.com/xuri/excelize/v2"
)

type Trace struct {
	Cell                 string   `json:"cell"`
	Formula              string   `json:"formula"`
	Status               string   `json:"status"`
	Functions            []string `json:"functions,omitempty"`
	UnsupportedFunctions []string `json:"unsupported_functions,omitempty"`
	UnsupportedSyntax    []string `json:"unsupported_syntax,omitempty"`
	References           []string `json:"references,omitempty"`
}

var (
	errCellReferenceRequiresSheet = errors.New("cell reference must include a sheet name")
	functionPattern               = regexp.MustCompile(`(?i)\b([A-Z][A-Z0-9.]*)\s*\(`)
	referencePattern              = regexp.MustCompile(`(?i)(?:'((?:[^']|'')+)'|([A-Z_][A-Z0-9_ .]*))!\$?([A-Z]{1,3})(?:\$?([0-9]+))?(?::\$?([A-Z]{1,3})(?:\$?([0-9]+))?)?|\$?([A-Z]{1,3})(?:\$?([0-9]+))(?::\$?([A-Z]{1,3})(?:\$?([0-9]+))?)?|\$?([A-Z]{1,3}):\$?([A-Z]{1,3})`)
	supportedFunctions            = map[string]bool{
		"SUM":      true,
		"SUMIFS":   true,
		"COUNTIF":  true,
		"COUNTIFS": true,
		"XLOOKUP":  true,
		"VLOOKUP":  true,
		"INDEX":    true,
		"MATCH":    true,
	}
)

func TraceWorkbookCell(inputFile string, cellRef string) (Trace, error) {
	sheet, cell, err := splitCellReference(cellRef)
	if err != nil {
		return Trace{}, err
	}

	file, err := excelize.OpenFile(inputFile)
	if err != nil {
		return Trace{}, err
	}
	defer func() { _ = file.Close() }()

	formula, err := file.GetCellFormula(sheet, cell)
	if err != nil {
		return Trace{}, err
	}
	if formula == "" {
		return Trace{}, fmt.Errorf("%s has no formula", cellRef)
	}
	return TraceFormula(sheet, cell, formula), nil
}

func TraceFormula(sheet string, cell string, formula string) Trace {
	normalizedFormula := normalizeFormula(formula)
	functions, unsupported := traceFunctions(normalizedFormula)
	unsupportedSyntax := traceUnsupportedSyntax(normalizedFormula)
	status := "supported"
	if len(unsupported) > 0 || len(unsupportedSyntax) > 0 {
		status = "unsupported"
	}
	return Trace{
		Cell:                 sheet + "!" + cell,
		Formula:              normalizedFormula,
		Status:               status,
		Functions:            functions,
		UnsupportedFunctions: unsupported,
		UnsupportedSyntax:    unsupportedSyntax,
		References:           traceReferences(sheet, normalizedFormula),
	}
}

func traceFunctions(formula string) ([]string, []string) {
	seen := map[string]bool{}
	unsupportedSeen := map[string]bool{}
	functions := []string{}
	unsupported := []string{}
	for _, match := range functionPattern.FindAllStringSubmatch(stripQuotedStrings(formula), -1) {
		name := strings.ToUpper(match[1])
		if !seen[name] {
			seen[name] = true
			functions = append(functions, name)
		}
		if !supportedFunctions[name] && !unsupportedSeen[name] {
			unsupportedSeen[name] = true
			unsupported = append(unsupported, name)
		}
	}
	if len(functions) == 0 {
		functions = nil
	}
	if len(unsupported) == 0 {
		unsupported = nil
	}
	return functions, unsupported
}

func traceReferences(currentSheet string, formula string) []string {
	masked := stripQuotedStrings(formula)
	references := []string{}
	seen := map[string]bool{}
	for _, match := range referencePattern.FindAllStringSubmatchIndex(masked, -1) {
		if isScientificNotationReferenceToken(masked, match[0], match[1]) {
			continue
		}
		reference := normalizeReference(currentSheet, masked, match)
		if reference == "" || seen[reference] {
			continue
		}
		seen[reference] = true
		references = append(references, reference)
	}
	if len(references) == 0 {
		return nil
	}
	return references
}

func traceUnsupportedSyntax(formula string) []string {
	if regexp.MustCompile(`[A-Za-z_][A-Za-z0-9_]*\[[^\]]+\]`).MatchString(stripQuotedStrings(formula)) {
		return []string{"structured_reference"}
	}
	return nil
}

func isScientificNotationReferenceToken(value string, start int, end int) bool {
	token := value[start:end]
	if len(token) < 2 || strings.IndexFunc(token, unicode.IsDigit) == -1 {
		return false
	}
	if start == 0 || !unicode.IsDigit(rune(value[start-1])) {
		return false
	}
	if !strings.ContainsAny(token, "Ee") {
		return false
	}
	if end < len(value) {
		next := rune(value[end])
		if unicode.IsLetter(next) || unicode.IsDigit(next) || next == '_' {
			return false
		}
	}
	return true
}

func normalizeReference(currentSheet string, formula string, match []int) string {
	token := formula[match[0]:match[1]]
	token = strings.ReplaceAll(token, "$", "")
	if strings.Contains(token, "!") {
		parts := strings.SplitN(token, "!", 2)
		sheet := strings.Trim(parts[0], "'")
		sheet = strings.ReplaceAll(sheet, "''", "'")
		return sheet + "!" + strings.ToUpper(parts[1])
	}
	return currentSheet + "!" + strings.ToUpper(token)
}

func normalizeFormula(formula string) string {
	if strings.HasPrefix(formula, "=") {
		return formula
	}
	return "=" + formula
}

func splitCellReference(cellRef string) (string, string, error) {
	parts := strings.SplitN(cellRef, "!", 2)
	if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" || strings.TrimSpace(parts[1]) == "" {
		return "", "", errCellReferenceRequiresSheet
	}
	sheet := strings.Trim(parts[0], "'")
	sheet = strings.ReplaceAll(sheet, "''", "'")
	return sheet, strings.ReplaceAll(strings.ToUpper(parts[1]), "$", ""), nil
}

func stripQuotedStrings(value string) string {
	var builder strings.Builder
	inString := false
	for _, r := range value {
		if r == '"' {
			inString = !inString
			builder.WriteRune(' ')
			continue
		}
		if inString {
			if unicode.IsSpace(r) {
				builder.WriteRune(r)
			} else {
				builder.WriteRune(' ')
			}
			continue
		}
		builder.WriteRune(r)
	}
	return builder.String()
}
