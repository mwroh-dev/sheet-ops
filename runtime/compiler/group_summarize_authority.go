package compiler

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mwroh/sheet-ops/runtime/taskspec"
)

type GroupSummarizeAuthority struct {
	InputWorkbook string
	SourceSheet   string
	Filters       []FilterSpec
	GroupBy       []string
	Metrics       []MetricSpec
}

func groupSummarizeAuthorityFromInspection(inputWorkbook, sourceSheet string, filters []FilterSpec, groupBy []string, metrics []MetricSpec) GroupSummarizeAuthority {
	return GroupSummarizeAuthority{
		InputWorkbook: inputWorkbook,
		SourceSheet:   sourceSheet,
		Filters:       cloneFilters(filters),
		GroupBy:       cloneStrings(groupBy),
		Metrics:       cloneMetrics(metrics),
	}
}

func groupSummarizeAuthorityFromTask(task taskspec.GroupSummarizeTask) GroupSummarizeAuthority {
	return GroupSummarizeAuthority{
		InputWorkbook: task.TaskSpec.InputWorkbook,
		SourceSheet:   task.SourceSheet,
		Filters:       cloneFilters(task.Filters),
		GroupBy:       cloneStrings(task.GroupBy),
		Metrics:       cloneMetrics(task.Metrics),
	}
}

func validateGroupSummarizeAuthority(task taskspec.GroupSummarizeTask, inspection WorkbookInspection) error {
	if err := validateSummaryTaskCompositionBoundary(task.TaskSpec); err != nil {
		return err
	}

	expected := inspection.Authority
	actual := groupSummarizeAuthorityFromTask(task)

	sameWorkbook, err := sameWorkbookIdentity(expected.InputWorkbook, actual.InputWorkbook)
	if err != nil {
		return err
	}
	if !sameWorkbook {
		return fmt.Errorf("input workbook mismatch between task and inspection")
	}
	if expected.SourceSheet != actual.SourceSheet {
		return fmt.Errorf("source sheet mismatch between task and inspection")
	}
	if !equalFilters(expected.Filters, actual.Filters) {
		return fmt.Errorf("filters mismatch between task and inspection")
	}
	if !equalStrings(expected.GroupBy, actual.GroupBy) {
		return fmt.Errorf("group_by mismatch between task and inspection")
	}
	if !equalMetrics(expected.Metrics, actual.Metrics) {
		return fmt.Errorf("metrics mismatch between task and inspection")
	}
	return nil
}

func validateSummaryTaskCompositionBoundary(spec taskspec.TaskSpec) error {
	if spec.ExecutionKind != taskspec.ExecutionKindComposition {
		return fmt.Errorf("group summary task execution_kind=%q want %q", spec.ExecutionKind, taskspec.ExecutionKindComposition)
	}
	if spec.CompositionKind != taskspec.CompositionKindGroupSummary {
		return fmt.Errorf("group summary task composition_kind=%q want %q", spec.CompositionKind, taskspec.CompositionKindGroupSummary)
	}
	if spec.Operation != taskspec.OperationCreateSummarySheet {
		return fmt.Errorf("group summary task operation=%q want %q", spec.Operation, taskspec.OperationCreateSummarySheet)
	}
	return nil
}

func equalFilters(left, right []FilterSpec) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func equalMetrics(left, right []MetricSpec) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func equalStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func sameWorkbookIdentity(left, right string) (bool, error) {
	leftInfo, leftErr := os.Stat(left)
	if leftErr != nil && !errors.Is(leftErr, os.ErrNotExist) {
		return false, leftErr
	}
	rightInfo, rightErr := os.Stat(right)
	if rightErr != nil && !errors.Is(rightErr, os.ErrNotExist) {
		return false, rightErr
	}
	if leftErr == nil && rightErr == nil && os.SameFile(leftInfo, rightInfo) {
		return true, nil
	}

	leftAbs, err := filepath.Abs(left)
	if err != nil {
		return false, err
	}
	rightAbs, err := filepath.Abs(right)
	if err != nil {
		return false, err
	}
	leftEval, err := filepath.EvalSymlinks(leftAbs)
	if err == nil {
		leftAbs = leftEval
	}
	rightEval, err := filepath.EvalSymlinks(rightAbs)
	if err == nil {
		rightAbs = rightEval
	}
	return filepath.Clean(leftAbs) == filepath.Clean(rightAbs), nil
}
