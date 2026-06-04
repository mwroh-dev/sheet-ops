package pathid

import (
	"strings"
	"testing"
)

func TestValidateAcceptsPublicScenarioIdentifiers(t *testing.T) {
	for _, value := range []string{
		"02-join-lookup",
		"orders-summary-values",
		"codex-skill-cli-summary",
		"case_1",
		"case.1",
	} {
		got, err := Validate("scenario id", value)
		if err != nil {
			t.Fatalf("Validate(%q): %v", value, err)
		}
		if got != value {
			t.Fatalf("Validate(%q)=%q want unchanged", value, got)
		}
	}
}

func TestValidateRejectsUnsafePathIdentifiers(t *testing.T) {
	longValue := "a" + strings.Repeat("b", 128)
	for _, value := range []string{
		"",
		"   ",
		"../x",
		"x/y",
		`x\y`,
		"/tmp/x",
		".",
		"..",
		"-starts-with-symbol",
		longValue,
	} {
		if _, err := Validate("scenario id", value); err == nil {
			t.Fatalf("Validate(%q) error=nil want rejection", value)
		}
	}
}
