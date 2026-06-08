package formula

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

var rowReferencePattern = regexp.MustCompile(`(?i)(\$?[A-Z]{1,3})(\$?)([0-9]+)`)

func TranslateFormulaRows(formula string, rowDelta int) (string, error) {
	normalized := normalizeFormula(formula)
	var translateErr error
	translated := rowReferencePattern.ReplaceAllStringFunc(normalized, func(token string) string {
		if translateErr != nil {
			return token
		}
		start := strings.Index(normalized, token)
		if start >= 0 && isScientificNotationToken(normalized, start, start+len(token)) {
			return token
		}
		match := rowReferencePattern.FindStringSubmatch(token)
		if len(match) != 4 {
			return token
		}
		if match[2] == "$" {
			return token
		}
		row, err := strconv.Atoi(match[3])
		if err != nil {
			translateErr = err
			return token
		}
		translatedRow := row + rowDelta
		if translatedRow < 1 {
			translateErr = fmt.Errorf("translated row %d from %q is outside worksheet bounds", translatedRow, token)
			return token
		}
		return match[1] + match[2] + strconv.Itoa(translatedRow)
	})
	if translateErr != nil {
		return "", translateErr
	}
	return translated, nil
}

func IsFormulaWithUnsupportedExtensionSyntax(formula string) bool {
	masked := stripQuotedStrings(formula)
	return regexp.MustCompile(`[A-Za-z_][A-Za-z0-9_]*\[[^\]]+\]`).MatchString(masked)
}

func isScientificNotationToken(value string, start int, end int) bool {
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
