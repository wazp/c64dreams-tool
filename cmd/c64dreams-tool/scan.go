package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newScanCmd(opts *options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "scan",
		Short: "Scan the C64 Dreams directory and report contents",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateOptions(opts); err != nil {
				return err
			}

			if opts.input == "" {
				return fmt.Errorf("--input is required")
			}

			fmt.Fprintf(
				cmd.OutOrStdout(),
				"scan target=%s region=%s group-by=%s max-name-len=%d dry-run=%v json=%v input=%s\n",
				string(opts.target),
				opts.region,
				opts.groupBy,
				opts.maxNameLen,
				opts.dryRun,
				opts.json,
				opts.input,
			)

			return nil
		},
	}

	return cmd
}
