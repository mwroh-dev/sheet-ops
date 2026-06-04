package policy

import "github.com/mwroh/sheet-ops/runtime/taskspec"

type Decision struct {
	Allow            bool     `json:"allow"`
	PreserveOriginal bool     `json:"preserve_original"`
	ApprovalRequired bool     `json:"approval_required"`
	Reasons          []string `json:"reasons"`
}

func EvaluateOriginalPreservation(task taskspec.GroupSummarizeTask) Decision {
	reasons := []string{}
	allow := true

	if task.TaskSpec.OutputWorkbook == "" {
		allow = false
		reasons = append(reasons, "output file must not be empty")
	}
	if task.TaskSpec.InputWorkbook == task.TaskSpec.OutputWorkbook {
		allow = false
		reasons = append(reasons, "output file must differ from input file")
	}
	if !task.PreserveOriginal {
		allow = false
		reasons = append(reasons, "operation must preserve original workbook")
	}
	if allow {
		reasons = append(reasons, "new output file will be written and source workbook remains untouched")
	}

	return Decision{
		Allow:            allow,
		PreserveOriginal: task.PreserveOriginal,
		ApprovalRequired: false,
		Reasons:          reasons,
	}
}
