package formula

import "testing"

func TestTranslateFormulaRowsMovesRelativeRows(t *testing.T) {
	got, err := TranslateFormulaRows("=B2*C2+$D$2+E$2+$F2+SUM(G2:G3)", 2)
	if err != nil {
		t.Fatalf("TranslateFormulaRows: %v", err)
	}
	want := "=B4*C4+$D$2+E$2+$F4+SUM(G4:G5)"
	if got != want {
		t.Fatalf("formula=%q want %q", got, want)
	}
}

func TestTranslateFormulaRowsRejectsInvalidTranslatedRow(t *testing.T) {
	if _, err := TranslateFormulaRows("=A1", -1); err == nil {
		t.Fatal("expected invalid translated row error")
	}
}
