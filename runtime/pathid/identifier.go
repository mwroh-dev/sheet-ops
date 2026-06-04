package pathid

import (
	"fmt"
	"regexp"
	"strings"
)

const patternText = `^[A-Za-z0-9][A-Za-z0-9._-]{0,127}$`

var pattern = regexp.MustCompile(patternText)

func Validate(name, value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", fmt.Errorf("%s required", name)
	}
	if !pattern.MatchString(value) {
		return "", fmt.Errorf("%s must match %s", name, patternText)
	}
	return value, nil
}

func Pattern() string {
	return patternText
}
