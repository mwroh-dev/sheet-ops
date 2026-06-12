package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

type evidenceSummaryResult struct {
	SchemaVersion string                 `json:"schema_version"`
	Command       string                 `json:"command"`
	OK            bool                   `json:"ok"`
	ReadOnly      bool                   `json:"read_only"`
	EvidenceDir   string                 `json:"evidence_dir"`
	Status        string                 `json:"status"`
	Checks        []evidenceSummaryCheck `json:"checks"`
	NextActions   []string               `json:"next_actions"`
}

type evidenceSummaryCheck struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

func newEvidenceSummaryCommand() *cobra.Command {
	var jsonOutput bool
	var evidenceDir string

	cmd := &cobra.Command{
		Use:   "evidence-summary",
		Short: "Summarize runtime evidence for agent success checks",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !jsonOutput {
				return writeCLIErrorJSONAndReturn(
					cmd.OutOrStdout(),
					newCLIError(cliErrorInvalidUsage, "evidence-summary requires --json", true, cliExitUsage, "evidence-summary --json --evidence-dir <dir>"),
				)
			}
			result, err := summarizeEvidenceDir(evidenceDir)
			if err != nil {
				var cliErr *cliError
				if strings.TrimSpace(err.Error()) != "" {
					cliErr = newInvalidDataError(err.Error(), "evidence-summary --json --evidence-dir <dir>")
				} else {
					cliErr = newInvalidDataError("invalid evidence directory", "evidence-summary --json --evidence-dir <dir>")
				}
				return writeCLIErrorJSONAndReturn(cmd.OutOrStdout(), cliErr)
			}
			return writeEvidenceSummaryJSON(cmd.OutOrStdout(), result)
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Emit machine-readable JSON on stdout")
	cmd.Flags().StringVar(&evidenceDir, "evidence-dir", "", "Path to a runtime evidence directory")
	if err := cmd.MarkFlagRequired("evidence-dir"); err != nil {
		panic(err)
	}
	return cmd
}

func summarizeEvidenceDir(evidenceDir string) (evidenceSummaryResult, error) {
	trimmedDir := strings.TrimSpace(evidenceDir)
	if trimmedDir == "" {
		return evidenceSummaryResult{}, fmt.Errorf("evidence-dir path cannot be empty")
	}
	absDir, err := filepath.Abs(trimmedDir)
	if err != nil {
		return evidenceSummaryResult{}, err
	}
	info, err := os.Stat(absDir)
	if err != nil {
		return evidenceSummaryResult{}, err
	}
	if !info.IsDir() {
		return evidenceSummaryResult{}, fmt.Errorf("evidence-dir must be a directory")
	}

	verification, verificationPresent, err := readEvidenceJSON(filepath.Join(absDir, "verification.json"))
	if err != nil {
		return evidenceSummaryResult{}, err
	}
	execution, executionPresent, err := readEvidenceJSON(filepath.Join(absDir, "execution.json"))
	if err != nil {
		return evidenceSummaryResult{}, err
	}
	_, failurePresent, err := readEvidenceJSON(filepath.Join(absDir, "failure.json"))
	if err != nil {
		return evidenceSummaryResult{}, err
	}

	checks := make([]evidenceSummaryCheck, 0, 4)
	verificationStatus := "missing"
	verificationMessage := "verification.json is missing"
	verificationPass := false
	if verificationPresent {
		pass, ok := boolField(verification, "pass")
		switch {
		case ok && pass:
			verificationStatus = "pass"
			verificationMessage = "verification.json reports pass=true"
			verificationPass = true
		case ok:
			verificationStatus = "fail"
			verificationMessage = "verification.json reports pass=false"
		default:
			verificationStatus = "unknown"
			verificationMessage = "verification.json is present but has no boolean pass field"
		}
	}
	checks = append(checks, evidenceSummaryCheck{Name: "verification_json", Status: verificationStatus, Message: verificationMessage})

	if executionPresent {
		checks = append(checks, evidenceSummaryCheck{Name: "execution_json", Status: "present", Message: executionMessage(execution)})
	} else {
		checks = append(checks, evidenceSummaryCheck{Name: "execution_json", Status: "missing", Message: "execution.json is missing"})
	}

	if failurePresent {
		checks = append(checks, evidenceSummaryCheck{Name: "failure_json", Status: "present", Message: "failure.json is present"})
	} else {
		checks = append(checks, evidenceSummaryCheck{Name: "failure_json", Status: "absent", Message: "failure.json is absent"})
	}

	hashStatus := "missing"
	hashMessage := "output workbook sha256 is missing"
	if value, ok := stringField(verification, "output_workbook_sha256"); ok && isSHA256(value) {
		hashStatus = "present"
		hashMessage = "output_workbook_sha256 is present"
	}
	checks = append(checks, evidenceSummaryCheck{Name: "output_workbook_sha256", Status: hashStatus, Message: hashMessage})

	status, nextActions := evidenceSummaryStatus(verificationPresent, verificationPass, failurePresent, hashStatus == "present")
	return evidenceSummaryResult{
		SchemaVersion: cliContractSchemaVersion,
		Command:       "evidence-summary",
		OK:            true,
		ReadOnly:      true,
		EvidenceDir:   absDir,
		Status:        status,
		Checks:        checks,
		NextActions:   nextActions,
	}, nil
}

func readEvidenceJSON(path string) (map[string]any, bool, error) {
	raw, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return map[string]any{}, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	var document map[string]any
	if err := json.Unmarshal(raw, &document); err != nil {
		return nil, false, fmt.Errorf("parse %s: %w", path, err)
	}
	return document, true, nil
}

func boolField(document map[string]any, key string) (bool, bool) {
	value, ok := document[key]
	if !ok {
		return false, false
	}
	typed, ok := value.(bool)
	return typed, ok
}

func stringField(document map[string]any, key string) (string, bool) {
	value, ok := document[key]
	if !ok {
		return "", false
	}
	typed, ok := value.(string)
	return typed, ok
}

func executionMessage(execution map[string]any) string {
	if operation, ok := stringField(execution, "operation"); ok && strings.TrimSpace(operation) != "" {
		return "execution.json is present for operation " + operation
	}
	return "execution.json is present"
}

func isSHA256(value string) bool {
	if len(value) != 64 {
		return false
	}
	for _, r := range value {
		if (r < '0' || r > '9') && (r < 'a' || r > 'f') && (r < 'A' || r > 'F') {
			return false
		}
	}
	return true
}

func evidenceSummaryStatus(verificationPresent, verificationPass, failurePresent, hashPresent bool) (string, []string) {
	switch {
	case verificationPass && hashPresent && !failurePresent:
		return "verified_success", []string{"safe_to_report_success_with_output_workbook_sha256"}
	case verificationPresent && !verificationPass:
		return "verification_failed", []string{"inspect_failure_evidence_and_repair_advice"}
	case failurePresent:
		return "failure_evidence_present", []string{"inspect_failure_evidence_and_repair_advice"}
	case verificationPass && !hashPresent:
		return "incomplete_success_evidence", []string{"locate_output_workbook_sha256_before_success_claim"}
	default:
		return "incomplete", []string{"run_or_locate_runtime_verification_evidence"}
	}
}

func writeEvidenceSummaryJSON(output io.Writer, result evidenceSummaryResult) error {
	encoder := json.NewEncoder(output)
	encoder.SetIndent("", "  ")
	return encoder.Encode(result)
}
