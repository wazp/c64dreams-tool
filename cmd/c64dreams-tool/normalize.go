package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newNormalizeCmd(opts *options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "normalize",
		Short: "Normalize the C64 Dreams collection for hardware targets",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateOptions(opts); err != nil {
				return err
			}

			if opts.input == "" {
				return fmt.Errorf("--input is required")
			}

			if opts.output == "" {
				return fmt.Errorf("--output is required")
			}

			if opts.sheet == "" {
				return fmt.Errorf("--sheet is required")
			}

			fmt.Fprintf(
				cmd.OutOrStdout(),
				"normalize target=%s region=%s group-by=%s max-name-len=%d dry-run=%v json=%v input=%s output=%s sheet=%s\n",
				string(opts.target),
				opts.region,
				opts.groupBy,
				opts.maxNameLen,
				opts.dryRun,
				opts.json,
				opts.input,
				opts.output,
				opts.sheet,
			)

			return nil
		},
	}

	return cmd
}
