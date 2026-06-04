package execute

import (
	"github.com/mwroh/sheet-ops/runtime/compiler"
	"github.com/xuri/excelize/v2"
)

func RunGroupSummarizeFormulas(ir compiler.WorkbookOperationIR, inputWorkbook, outputWorkbook string) (ExecutionResult, error) {
	inspection, err := compiler.InspectWorkbook(inputWorkbook, ir.SourceSheet, ir.Filters, ir.GroupBy, ir.Metrics)
	if err != nil {
		return ExecutionResult{}, err
	}
	plan, err := compiler.BuildGroupSummarizePlan(inspection, ir)
	if err != nil {
		return ExecutionResult{}, err
	}
	return RunGroupSummarizePlan(plan, outputWorkbook)
}

func writeFormulaSummarySheet(file *excelize.File, plan compiler.GroupSummarizePlan) ([]string, error) {
	expectationByCell := make(map[string]compiler.FormulaExpectation, len(plan.FormulaExpectations))
	for _, expectation := range plan.FormulaExpectations {
		expectationByCell[expectation.Cell] = expectation
	}

	formulaCells := []string{}
	for rowIndex, row := range plan.SummaryRows {
		excelRow := rowIndex + 1
		for colIndex, value := range row {
			cell, _ := excelize.CoordinatesToCellName(colIndex+1, excelRow)
			expectation, hasFormula := expectationByCell[cell]
			if !hasFormula {
				if err := file.SetCellValue(plan.Operation.TargetSheet, cell, value); err != nil {
					return nil, err
				}
				continue
			}
			if err := file.SetCellFormula(plan.Operation.TargetSheet, cell, expectation.Render()); err != nil {
				return nil, err
			}
			formulaCells = append(formulaCells, cell)
		}
	}

	return formulaCells, nil
}
