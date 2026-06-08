package verify

import (
	"fmt"
	"strings"

	"github.com/mwroh/sheet-ops/runtime/compiler"
	"github.com/xuri/excelize/v2"
)

func VerifyAddDataValidation(ir compiler.WorkbookOperationIR, inputWorkbook, outputWorkbook, sourceSHA256Before, sourceSHA256After string) (VerificationResult, error) {
	if sourceSHA256Before == "" || sourceSHA256After == "" {
		return VerificationResult{Pass: false, OperationFamily: ir.OperationFamily, OutputWorkbook: outputWorkbook, SummarySheet: ir.SourceSheet, Reasons: []string{"execution hash evidence is missing"}}, nil
	}
	if sourceSHA256Before != sourceSHA256After {
		return VerificationResult{Pass: false, OperationFamily: ir.OperationFamily, OutputWorkbook: outputWorkbook, SummarySheet: ir.SourceSheet, Reasons: []string{"source workbook changed during execution"}}, nil
	}
	currentHash, err := hashFile(inputWorkbook)
	if err != nil {
		return VerificationResult{}, err
	}
	if currentHash != sourceSHA256Before {
		return VerificationResult{Pass: false, OperationFamily: ir.OperationFamily, OutputWorkbook: outputWorkbook, SummarySheet: ir.SourceSheet, Reasons: []string{"source workbook hash differs from execution evidence"}}, nil
	}
	outputHandle, err := excelize.OpenFile(outputWorkbook)
	if err != nil {
		return VerificationResult{}, err
	}
	defer func() { _ = outputHandle.Close() }()

	validations, err := outputHandle.GetDataValidations(ir.SourceSheet)
	if err != nil {
		return VerificationResult{}, err
	}
	written := []string{}
	for _, targetRange := range ir.ValidationRule.Ranges {
		found := false
		for _, validation := range validations {
			if validation == nil || validation.Sqref != targetRange || validation.Type != "list" {
				continue
			}
			if !sameAllowedValues(validation.Formula1, ir.ValidationRule.AllowedValues) {
				continue
			}
			found = true
			break
		}
		if !found {
			return VerificationResult{
				Pass:            false,
				OperationFamily: ir.OperationFamily,
				OutputWorkbook:  outputWorkbook,
				SummarySheet:    ir.SourceSheet,
				WrittenCells:    written,
				Reasons:         []string{fmt.Sprintf("missing list validation on %s", targetRange)},
			}, nil
		}
		written = append(written, targetRange)
	}
	return VerificationResult{
		Pass:                true,
		OperationFamily:     ir.OperationFamily,
		OutputWorkbook:      outputWorkbook,
		SummarySheet:        ir.SourceSheet,
		WrittenCells:        written,
		Reasons:             []string{"target ranges contain expected list data validation and source workbook is unchanged"},
		SourceSHA256Checked: true,
	}, nil
}

func sameAllowedValues(formula string, expected []string) bool {
	normalized := strings.Trim(formula, `"`)
	actual := strings.Split(normalized, ",")
	if len(actual) != len(expected) {
		return false
	}
	for i := range expected {
		if actual[i] != expected[i] {
			return false
		}
	}
	return true
}
