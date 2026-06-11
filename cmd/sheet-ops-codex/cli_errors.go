package main

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"strings"
)

const (
	cliErrorInvalidUsage          = "invalid_usage"
	cliErrorMissingRequiredOption = "missing_required_option"
	cliErrorUnknownCommand        = "unknown_command"
	cliErrorInvalidData           = "invalid_json_or_schema"
	cliErrorInstallDependency     = "install_dependency_unavailable"
	cliErrorStateRootMismatch     = "state_root_mismatch"
	cliErrorRequestCheckpoint     = "request_checkpoint"
	cliErrorValidationBlocked     = "validation_blocked"
	cliErrorExecutionFailed       = "execution_failed"
	cliErrorVerificationFailed    = "verification_failed"
	cliErrorInternal              = "internal_error"

	cliExitOK            = 0
	cliExitGeneral       = 1
	cliExitUsage         = 64
	cliExitData          = 65
	cliExitUnavailable   = 69
	cliExitSoftware      = 70
	cliExitConfiguration = 78
)

const (
	cliErrorStatusEmitted  = "emitted"
	cliErrorStatusReserved = "reserved"
)

type cliError struct {
	Code              string   `json:"code"`
	Message           string   `json:"message"`
	Recoverable       bool     `json:"recoverable"`
	SuggestedCommands []string `json:"suggested_commands"`
	ExitCode          int      `json:"exit_code"`
	JSONEmitted       bool     `json:"-"`
}

func (err *cliError) Error() string {
	if err == nil {
		return ""
	}
	return err.Message
}

var _ error = (*cliError)(nil)

type cliErrorEnvelope struct {
	SchemaVersion string   `json:"schema_version"`
	OK            bool     `json:"ok"`
	Error         cliError `json:"error"`
}

type cliErrorContractPayload struct {
	JSONFailureShape string                   `json:"json_failure_shape"`
	Codes            []cliErrorCodeDescriptor `json:"codes"`
}

type cliErrorCodeDescriptor struct {
	Code        string `json:"code"`
	ExitCode    int    `json:"exit_code"`
	Recoverable bool   `json:"recoverable"`
	Status      string `json:"status"`
	Producer    string `json:"producer"`
	Meaning     string `json:"meaning"`
}

func newCLIError(code, message string, recoverable bool, exitCode int, suggestedCommands ...string) *cliError {
	return &cliError{
		Code:              code,
		Message:           message,
		Recoverable:       recoverable,
		SuggestedCommands: append([]string(nil), suggestedCommands...),
		ExitCode:          exitCode,
	}
}

func newStateRootMismatchError(message string) *cliError {
	return newCLIError(
		cliErrorStateRootMismatch,
		message,
		true,
		cliExitConfiguration,
		"preflight --json --input-file <workbook>",
		"capabilities --json",
	)
}

func newInvalidDataError(message string, suggestedCommands ...string) *cliError {
	if len(suggestedCommands) == 0 {
		suggestedCommands = []string{
			"schema command preview-request --json",
			"capabilities --json",
		}
	}
	return newCLIError(
		cliErrorInvalidData,
		message,
		true,
		cliExitData,
		suggestedCommands...,
	)
}

func classifyCLIError(err error) cliError {
	if err == nil {
		return cliError{Code: "", ExitCode: cliExitOK}
	}

	var typed *cliError
	if errors.As(err, &typed) {
		return *typed
	}

	if os.IsNotExist(err) {
		return *newCLIError(
			cliErrorInvalidUsage,
			strings.TrimSpace(err.Error()),
			true,
			cliExitUsage,
			"preflight --json --input-file <workbook>",
			"schema --json",
		)
	}

	message := strings.TrimSpace(err.Error())
	lower := strings.ToLower(message)
	switch {
	case strings.Contains(lower, "required flag(s)") && strings.Contains(lower, "not set"):
		return *newCLIError(
			cliErrorMissingRequiredOption,
			message,
			true,
			cliExitUsage,
			"prepare-use --help",
			"run-validated --help",
			"run-intent --help",
		)
	case strings.Contains(lower, "unknown command"):
		return *newCLIError(
			cliErrorUnknownCommand,
			message,
			true,
			cliExitUsage,
			"capabilities --json",
			"schema --json",
			"help",
		)
	default:
		return *newCLIError(cliErrorInternal, message, false, cliExitGeneral)
	}
}

func writeCLIErrorJSON(output io.Writer, err *cliError) error {
	if err == nil {
		return nil
	}
	encoder := json.NewEncoder(output)
	encoder.SetIndent("", "  ")
	return encoder.Encode(cliErrorEnvelope{
		SchemaVersion: cliContractSchemaVersion,
		OK:            false,
		Error:         *err,
	})
}

func writeCLIErrorJSONAndReturn(output io.Writer, err *cliError) error {
	if err == nil {
		return nil
	}
	if writeErr := writeCLIErrorJSON(output, err); writeErr != nil {
		return writeErr
	}
	err.JSONEmitted = true
	return err
}

func buildCLIErrorContractPayload() cliErrorContractPayload {
	return cliErrorContractPayload{
		JSONFailureShape: `{"ok":false,"error":{...}}`,
		Codes: []cliErrorCodeDescriptor{
			{
				Code:        cliErrorInvalidUsage,
				ExitCode:    cliExitUsage,
				Recoverable: true,
				Status:      cliErrorStatusEmitted,
				Producer:    "json_mode_guard",
				Meaning:     "The command form or required --json mode is invalid; inspect help or schema before retrying.",
			},
			{
				Code:        cliErrorMissingRequiredOption,
				ExitCode:    cliExitUsage,
				Recoverable: true,
				Status:      cliErrorStatusEmitted,
				Producer:    "cobra_required_flags",
				Meaning:     "A required option is missing; retry with the command-specific required flags.",
			},
			{
				Code:        cliErrorUnknownCommand,
				ExitCode:    cliExitUsage,
				Recoverable: true,
				Status:      cliErrorStatusEmitted,
				Producer:    "cobra_or_schema_lookup",
				Meaning:     "The command or schema target is unknown; rediscover commands through capabilities or schema.",
			},
			{
				Code:        cliErrorInvalidData,
				ExitCode:    cliExitData,
				Recoverable: true,
				Status:      cliErrorStatusEmitted,
				Producer:    "normalized_intent_loader",
				Meaning:     "Input JSON, request data, or schema validation failed before execution.",
			},
			{
				Code:        cliErrorInstallDependency,
				ExitCode:    cliExitUnavailable,
				Recoverable: true,
				Status:      cliErrorStatusReserved,
				Producer:    "install_dependency_check",
				Meaning:     "An install-time dependency such as Go or the bundled CLI binary is unavailable.",
			},
			{
				Code:        cliErrorStateRootMismatch,
				ExitCode:    cliExitConfiguration,
				Recoverable: true,
				Status:      cliErrorStatusEmitted,
				Producer:    "state_root_guard",
				Meaning:     "State-root configuration would place artifacts outside the workbook-case state root.",
			},
			{
				Code:        cliErrorRequestCheckpoint,
				ExitCode:    cliExitConfiguration,
				Recoverable: true,
				Status:      cliErrorStatusReserved,
				Producer:    "request_compiler_checkpoint",
				Meaning:     "The request compiler stopped for a human checkpoint or blocked decision.",
			},
			{
				Code:        cliErrorValidationBlocked,
				ExitCode:    cliExitData,
				Recoverable: true,
				Status:      cliErrorStatusReserved,
				Producer:    "runtime_validation",
				Meaning:     "Validation blocked execution; inspect generated validation artifacts before retrying.",
			},
			{
				Code:        cliErrorExecutionFailed,
				ExitCode:    cliExitSoftware,
				Recoverable: false,
				Status:      cliErrorStatusReserved,
				Producer:    "runtime_executor",
				Meaning:     "Runtime execution failed after a validated request reached the orchestrator.",
			},
			{
				Code:        cliErrorVerificationFailed,
				ExitCode:    cliExitSoftware,
				Recoverable: false,
				Status:      cliErrorStatusReserved,
				Producer:    "result_verifier",
				Meaning:     "Result verification failed after runtime execution and needs artifact review.",
			},
			{
				Code:        cliErrorInternal,
				ExitCode:    cliExitGeneral,
				Recoverable: false,
				Status:      cliErrorStatusEmitted,
				Producer:    "fallback_classifier",
				Meaning:     "An uncategorized internal CLI or runtime error occurred.",
			},
		},
	}
}
