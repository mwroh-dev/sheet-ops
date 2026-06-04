package main

import (
	"fmt"
	"io"
	"strings"

	requestmode "github.com/mwroh/sheet-ops/runtime/openlayer/requestmode"
	"github.com/mwroh/sheet-ops/runtime/openlayer/useorchestrator"
	runtimesession "github.com/mwroh/sheet-ops/runtime/session"
	"github.com/spf13/cobra"
)

func newUseStructuredCommand() *cobra.Command {
	var envelopePath string

	cmd := &cobra.Command{
		Use:   "use-structured",
		Short: "Run a typed structured Sheet Ops use request without open-layer compilation",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUseCommandWithSetup(cmd.OutOrStdout(), envelopePath, func(output io.Writer, ctx useCommandContext) error {
				structuredRequest, err := loadStructuredUseRequest(ctx.Envelope)
				if err != nil {
					return writeRunFailureResult(output, ctx.RunSession, ctx.EnvelopePath, ctx.BundleRoot, ctx.Envelope, "REQUEST_INVALID", err.Error())
				}
				return executeUseStructured(output, ctx, structuredRequest)
			})
		},
	}

	cmd.Flags().StringVar(&envelopePath, "envelope", "", "Path to a use-envelope JSON file with request.kind=structured_use_request")
	if err := cmd.MarkFlagRequired("envelope"); err != nil {
		panic(err)
	}
	return cmd
}

func executeUseStructured(output io.Writer, ctx useCommandContext, req useorchestrator.UseRequest) error {
	if strings.TrimSpace(req.Operation) == "" {
		return writeRunFailureResult(output, ctx.RunSession, ctx.EnvelopePath, ctx.BundleRoot, ctx.Envelope, "REQUEST_INVALID", "structured request operation is required")
	}
	if _, err := runtimesession.WriteProof(ctx.RunSession, "structured-use-request.json", req); err != nil {
		return err
	}
	runtimeResult, err := runCaseLocalUseRequest(ctx.CaseRoot, req)
	if err != nil {
		return writeRuntimeFailureResult(output, ctx.RunSession, ctx.EnvelopePath, ctx.BundleRoot, ctx.Envelope, runtimeResult, "RUNTIME_FAILED", err.Error())
	}
	result, err := buildRunSuccessResult(ctx.RunSession, ctx.EnvelopePath, ctx.BundleRoot, ctx.Envelope, runtimeResult)
	if err != nil {
		return err
	}
	return writeUseTerminalResultJSON(output, result)
}

func runCaseLocalUseRequest(caseRoot string, req useorchestrator.UseRequest) (useorchestrator.RunResult, error) {
	restore, err := forceCaseLocalRuntimeRoots(caseRoot)
	if err != nil {
		return useorchestrator.RunResult{}, err
	}
	defer restore()
	return useorchestrator.Orchestrate(req)
}

func loadStructuredUseRequest(envelope UseEnvelope) (useorchestrator.UseRequest, error) {
	if envelope.Request.Kind != requestmode.ModeStructuredUseRequest {
		return useorchestrator.UseRequest{}, fmt.Errorf("use-structured requires request.kind %q, got %q", requestmode.ModeStructuredUseRequest, envelope.Request.Kind)
	}

	req, err := useorchestrator.LoadRequest(envelope.Request.Path)
	if err != nil {
		return useorchestrator.UseRequest{}, fmt.Errorf("load structured request %s: %w", envelope.Request.Path, err)
	}
	if strings.TrimSpace(req.ScenarioID) != envelope.ScenarioID {
		return useorchestrator.UseRequest{}, fmt.Errorf(
			"structured request scenario_id %q does not match envelope scenario_id %q",
			req.ScenarioID,
			envelope.ScenarioID,
		)
	}
	if strings.TrimSpace(req.Operation) == "" {
		return useorchestrator.UseRequest{}, fmt.Errorf("structured request operation is required")
	}
	if !samePath(req.InputFile, envelope.InputFile) {
		return useorchestrator.UseRequest{}, fmt.Errorf(
			"structured request input_file %q does not match envelope input_file %q",
			req.InputFile,
			envelope.InputFile,
		)
	}
	if !samePath(req.OutputFile, envelope.OutputFile) {
		return useorchestrator.UseRequest{}, fmt.Errorf(
			"structured request output_file %q does not match envelope output_file %q",
			req.OutputFile,
			envelope.OutputFile,
		)
	}
	return useorchestrator.NormalizeUseRequest(req), nil
}
