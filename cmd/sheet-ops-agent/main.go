package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	if err := newRootCommand().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func newRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:           "sheet-ops-agent",
		Short:         "Case-local Sheet Ops agent entry",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	rootCmd.AddCommand(newUseCommand())
	rootCmd.AddCommand(newUseOpenCommand())
	rootCmd.AddCommand(newUseStructuredCommand())
	rootCmd.AddCommand(newRunValidatedCommand())
	rootCmd.AddCommand(newTraceCommand())
	return rootCmd
}
