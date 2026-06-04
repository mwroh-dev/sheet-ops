package main

import (
	"encoding/json"
	"errors"
	"fmt"

	runtimeformula "github.com/mwroh/sheet-ops/runtime/formula"
	"github.com/spf13/cobra"
)

func newTraceCommand() *cobra.Command {
	var cell string
	var asJSON bool

	cmd := &cobra.Command{
		Use:   "trace input.xlsx --cell Sheet!A1",
		Short: "Trace a workbook formula cell",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if cell == "" {
				return errors.New("--cell is required")
			}
			trace, err := runtimeformula.TraceWorkbookCell(args[0], cell)
			if err != nil {
				return err
			}
			if asJSON {
				encoder := json.NewEncoder(cmd.OutOrStdout())
				encoder.SetIndent("", "  ")
				return encoder.Encode(trace)
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "%s %s\n", trace.Cell, trace.Formula)
			return err
		},
	}

	cmd.Flags().StringVar(&cell, "cell", "", "Formula cell reference to trace, for example Summary!D5")
	cmd.Flags().BoolVar(&asJSON, "json", true, "Write trace as JSON")
	return cmd
}
